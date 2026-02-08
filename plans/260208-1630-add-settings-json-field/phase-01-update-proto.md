---
title: "Phase 1: Update Proto Definition"
description: "Add settings_json field to Collection message in proto file"
phase: 1
status: completed
priority: High
dependencies: []
---

# Phase 1: Update Proto Definition

## Context Links

- Parent Plan: [plan.md](./plan.md)
- Next Phase: [Phase 2: Update Repository](./phase-02-update-repository.md)
- Proto File: `proto/collection.proto`

## Overview

| Field | Value |
|-------|-------|
| **Priority** | High |
| **Status** | Pending |
| **Description** | Add settings_json field to Collection message in proto definition |
| **Estimated Effort** | 30 minutes |

## Key Insights

- Proto file defines the gRPC contract between services
- Need to add `settings_json` field to `Collection` message
- Field number must be unique within the message
- Use `string` type for JSON content

## Requirements

### Functional Requirements
- Add `settings_json` field to Collection message
- Field should be optional (default behavior in proto3)
- Use appropriate field number

### Non-Functional Requirements
- Follow proto naming conventions (snake_case)
- Add comment documenting field purpose
- Ensure backward compatibility

## Related Code Files

### Files to Modify
| File | Change |
|------|--------|
| `proto/collection.proto` | Add settings_json field to Collection message |

## Implementation Steps

1. **Open proto/collection.proto**
2. **Find Collection message definition**
3. **Add settings_json field**
   - Type: `string`
   - Name: `settings_json`
   - Field number: next available number
4. **Regenerate Go code from proto**
   - Run `make proto` or equivalent command

## Todo List

- [ ] Add `settings_json` field to Collection message in proto file
- [ ] Add comment documenting the field
- [ ] Regenerate Go protobuf code
- [ ] Verify generated code compiles

## Code Snippet

```protobuf
// proto/collection.proto

message Collection {
  // ... existing fields ...

  // Settings JSON stores rich metadata for collection display
  // Includes overview, utility accordions, etc.
  string settings_json = 45;  // Use next available number
}
```

## Success Criteria

- [ ] Proto file updated with settings_json field
- [ ] Generated Go code includes new field
- [ ] Code compiles without errors

## Next Steps

After completing this phase:
1. Proceed to [Phase 2: Update Repository](./phase-02-update-repository.md)
2. Update database repository to include field in queries
