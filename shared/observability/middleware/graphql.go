package middleware

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/getsentry/sentry-go"
)

// GraphQLFieldMiddleware creates Sentry spans for each GraphQL field resolution
func GraphQLFieldMiddleware() graphql.FieldMiddleware {
	return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
		// Get field context
		fc := graphql.GetFieldContext(ctx)
		if fc == nil {
			return next(ctx)
		}

		// Create span for this field
		parentTypeName := ""
		if fc.Field.ObjectDefinition != nil {
			parentTypeName = fc.Field.ObjectDefinition.Name
		}
		span := sentry.StartSpan(ctx,
			fc.Field.Field.Name,
			sentry.WithOpName("graphql.resolve"),
			sentry.WithDescription(parentTypeName+"."+fc.Field.Name),
		)

		// Add GraphQL context data
		span.SetData("graphql.field.name", fc.Field.Field.Name)
		span.SetData("graphql.field.type", fc.Field.Definition.Type.String())
		span.SetData("graphql.parent_type", parentTypeName)

		// Add operation info if available
		if rc := graphql.GetOperationContext(ctx); rc != nil {
			span.SetData("graphql.operation.name", rc.OperationName)
			if rc.Operation != nil {
				opType := "query"
				switch rc.Operation.Operation {
				case "query":
					opType = "query"
				case "mutation":
					opType = "mutation"
				case "subscription":
					opType = "subscription"
				}
				span.SetData("graphql.operation.type", opType)
			}

			// Get query variables (sanitized)
			if rc.Variables != nil {
				span.SetData("graphql.variables.count", len(rc.Variables))
			}
		}

		defer func() {
			if r := recover(); r != nil {
				span.SetData("graphql.panic", r)
				span.Status = sentry.SpanStatusInternalError
				span.Finish()
				panic(r)
			}
		}()

		// Resolve field
		result, err := next(span.Context())

		if err != nil {
			span.SetData("graphql.error", err.Error())
			span.Status = sentry.SpanStatusInternalError
		} else {
			span.Status = sentry.SpanStatusOK
		}

		span.Finish()

		return result, err
	}
}

// GraphQLResponseMiddleware captures operation-level metrics
// (Optional - use for query-level insights)
func GraphQLResponseMiddleware() graphql.ResponseMiddleware {
	return func(ctx context.Context, next graphql.ResponseHandler) *graphql.Response {
		// Get operation context
		rc := graphql.GetOperationContext(ctx)
		if rc == nil {
			return next(ctx)
		}

		// Start operation-level span
		opDesc := "query"
		if rc.Operation != nil {
			switch rc.Operation.Operation {
			case "query":
				opDesc = "query"
			case "mutation":
				opDesc = "mutation"
			case "subscription":
				opDesc = "subscription"
			}
		}
		span := sentry.StartSpan(ctx,
			rc.OperationName,
			sentry.WithOpName("graphql.operation"),
			sentry.WithDescription(opDesc),
		)

		span.SetData("graphql.operation.name", rc.OperationName)
		span.SetData("graphql.operation.type", opDesc)

		// Execute operation
		response := next(span.Context())

		// Record errors
		if response != nil && len(response.Errors) > 0 {
			span.SetData("graphql.errors.count", len(response.Errors))
			span.Status = sentry.SpanStatusInternalError
		} else {
			span.Status = sentry.SpanStatusOK
		}

		span.Finish()

		return response
	}
}
