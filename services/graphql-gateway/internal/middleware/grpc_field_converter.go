package middleware

import (
	"context"
	"reflect"
	"strings"
	"unicode"

	"google.golang.org/grpc"
)

// GrpcFieldConverter returns a gRPC client interceptor that converts camelCase fields to snake_case
func GrpcFieldConverter() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// Convert request fields before sending
		convertFields(req)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// convertFields recursively converts all string fields from camelCase to snake_case
func convertFields(v interface{}) {
	if v == nil {
		return
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() != reflect.Struct {
		return
	}

	rt := rv.Type()
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		fieldType := rt.Field(i)

		// Skip unexported fields
		if !field.CanSet() {
			continue
		}

		switch field.Kind() {
		case reflect.String:
			// Check if this is a sort field or other field that needs conversion
			fieldName := fieldType.Name
			if strings.Contains(strings.ToLower(fieldName), "sort") ||
				strings.Contains(strings.ToLower(fieldName), "field") ||
				strings.Contains(strings.ToLower(fieldName), "by") {
				converted := CamelToSnakeCase(field.String())
				field.SetString(converted)
			}

		case reflect.Ptr:
			if !field.IsNil() {
				convertFields(field.Interface())
			}

		case reflect.Struct:
			convertFields(field.Addr().Interface())
		}
	}
}

// CamelToSnakeCase converts camelCase to snake_case
func CamelToSnakeCase(s string) string {
	if s == "" {
		return s
	}

	var result strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				result.WriteByte('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
