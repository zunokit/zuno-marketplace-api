---
title: "Phase 3: Update GraphQL Schema"
description: "Add settingsJson field to GraphQL Collection type"
phase: 3
status: completed
priority: High
dependencies: ["phase-02-update-repository"]
---

# Phase 3: Update GraphQL Schema

## Context Links

- Parent Plan: [plan.md](./plan.md)
- Previous Phase: [Phase 2: Update Repository](./phase-02-update-repository.md)
- Next Phase: [Phase 4: Update Mappers](./phase-04-update-mappers.md)
- Schema File: `services/graphql-gateway/graph/schemas/collection.graphqls`

## Overview

| Field | Value |
|-------|-------|
| **Priority** | High |
| **Status** | Pending |
| **Description** | Add settingsJson field to GraphQL Collection type |
| **Estimated Effort** | 30 minutes |

## Key Insights

- GraphQL schema uses camelCase (settingsJson)
- Proto uses snake_case (settings_json)
- Need to regenerate GraphQL code after schema change
- Field should be nullable (String type)

## Requirements

### Functional Requirements
- Add `settingsJson` field to GraphQL Collection type
- Field type: `String` (nullable)
- Add appropriate documentation comment

### Non-Functional Requirements
- Follow GraphQL naming conventions (camelCase)
- Maintain backward compatibility
- Regenerate Go code from schema

## Related Code Files

### Files to Modify
| File | Change |
|------|--------|
| `services/graphql-gateway/graph/schemas/collection.graphqls` | Add settingsJson field |

## Implementation Steps

1. **Update GraphQL schema**
   - Open `collection.graphqls`
   - Find Collection type definition
   - Add `settingsJson: String` field

2. **Regenerate GraphQL code**
   - Run `go generate ./...` or `make generate`
   - Verify generated code includes new field

## Todo List

- [ ] Add `settingsJson: String` field to Collection type in schema
- [ ] Add documentation comment for the field
- [ ] Regenerate GraphQL Go code
- [ ] Verify generated models include new field

## Code Snippet

```graphql
# services/graphql-gateway/graph/schemas/collection.graphqls

type Collection {
  # ... existing fields ...

  # Settings JSON stores rich metadata for collection display
  # Includes overview content, utility accordions, etc.
  settingsJson: String
}
```

## Success Criteria

- [ ] GraphQL schema includes settingsJson field
- [ ] Generated Go models include new field
- [ ] Schema compiles without errors
- [ ] Field appears in GraphQL introspection

## Next Steps

After completing this phase:
1. Proceed to [Phase 4: Update Mappers](./phase-04-update-mappers.md)
2. Update proto-to-GraphQL mapper
