---
title: "Phase 4: Update Mappers"
description: "Map settings_json from proto to GraphQL in collection mapper"
phase: 4
status: completed
priority: High
dependencies: ["phase-03-update-graphql"]
---

# Phase 4: Update Mappers

## Context Links

- Parent Plan: [plan.md](./plan.md)
- Previous Phase: [Phase 3: Update GraphQL](./phase-03-update-graphql.md)
- Next Phase: [Phase 5: Testing](./phase-05-testing.md)
- Mapper File: `services/graphql-gateway/internal/mapper/collection_mapper.go`

## Overview

| Field | Value |
|-------|-------|
| **Priority** | High |
| **Status** | Pending |
| **Description** | Map settings_json from proto to GraphQL in collection mapper |
| **Estimated Effort** | 30 minutes |

## Key Insights

- Mapper converts proto Collection to GraphQL Collection model
- Proto field: `settings_json` (snake_case)
- GraphQL field: `SettingsJson` (PascalCase in Go)
- Simple string-to-string mapping

## Requirements

### Functional Requirements
- Map `settings_json` from proto to `SettingsJson` in GraphQL model
- Handle empty/null strings gracefully
- No transformation needed (JSON string passthrough)

### Non-Functional Requirements
- Maintain existing mapper patterns
- Add appropriate nil checks

## Related Code Files

### Files to Modify
| File | Change |
|------|--------|
| `services/graphql-gateway/internal/mapper/collection_mapper.go` | Add settings_json mapping |

## Implementation Steps

1. **Open collection_mapper.go**
2. **Find ProtoToGraphQLCollection function**
3. **Add settings_json mapping**
   - Map proto.SettingsJson to model.SettingsJson
   - Handle empty string case

## Todo List

- [ ] Add settings_json mapping in ProtoToGraphQLCollection
- [ ] Handle empty/null strings
- [ ] Verify mapping compiles
- [ ] Test with sample data

## Code Snippet

```go
// services/graphql-gateway/internal/mapper/collection_mapper.go

func ProtoToGraphQLCollection(proto *pb.Collection) *model.Collection {
    if proto == nil {
        return nil
    }

    return &model.Collection{
        // ... existing mappings ...

        SettingsJson: proto.SettingsJson,  // ADD THIS
    }
}
```

## Success Criteria

- [ ] Mapper includes settings_json mapping
- [ ] GraphQL queries return settingsJson field
- [ ] Empty/null values handled correctly
- [ ] End-to-end flow works

## Next Steps

After completing this phase:
1. Proceed to [Phase 5: Testing](./phase-05-testing.md)
2. Test the complete implementation
