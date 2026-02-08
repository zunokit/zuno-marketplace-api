---
title: "Phase 2: Update Database Repository"
description: "Include settings_json in repository queries and model"
phase: 2
status: completed
priority: High
dependencies: ["phase-01-update-proto"]
---

# Phase 2: Update Database Repository

## Context Links

- Parent Plan: [plan.md](./plan.md)
- Previous Phase: [Phase 1: Update Proto](./phase-01-update-proto.md)
- Next Phase: [Phase 3: Update GraphQL](./phase-03-update-graphql.md)
- Model File: `services/collection-service/internal/models/collection.go`
- Repository File: `services/collection-service/internal/repository/collection_repository.go`

## Overview

| Field | Value |
|-------|-------|
| **Priority** | High |
| **Status** | Pending |
| **Description** | Include settings_json in repository queries and model |
| **Estimated Effort** | 1 hour |

## Key Insights

- Collection model already has SettingsJSON field (JSONB type)
- Repository queries need to include this field in SELECT statements
- Converter needs to map between database model and proto message

## Requirements

### Functional Requirements
- Include `settings_json` in repository SELECT queries
- Map database JSONB to proto string in converter
- Handle null/empty JSON gracefully

### Non-Functional Requirements
- Maintain existing query performance
- No breaking changes to existing queries

## Related Code Files

### Files to Modify
| File | Change |
|------|--------|
| `services/collection-service/internal/repository/collection_repository.go` | Add settings_json to SELECT queries |
| `services/collection-service/internal/converter/collection_converter.go` | Add mapping for settings_json |

## Implementation Steps

1. **Update repository queries**
   - Find SELECT statements for collections
   - Add `settings_json` to column list

2. **Update converter**
   - Map `SettingsJSON` (JSONB) to proto `SettingsJson` (string)
   - Handle null values

## Todo List

- [ ] Update `GetCollection` query to include settings_json
- [ ] Update `ListCollections` query to include settings_json
- [ ] Update converter to map SettingsJSON field
- [ ] Handle null/empty JSON in converter
- [ ] Test queries return correct data

## Code Snippets

### Repository Query Update
```go
// services/collection-service/internal/repository/collection_repository.go

// Update SELECT statements to include settings_json:
const selectCollectionColumns = `
    id, slug, user_id, name, symbol, description, category,
    contract_address, chain_id, token_standard, deployer_address,
    deployed_block, status, deployed_at, index_status,
    image_url, banner_url, featured_image_url, website_url,
    base_uri, max_supply, mint_price_allowlist, mint_price_public,
    mint_start_time, allowlist_stage_end, mint_limit_per_wallet,
    royalty_fee_bps, royalty_recipient, total_supply, total_minted,
    is_verified, is_hidden, source, social_links_json, tags_json,
    settings_json,  // ADD THIS
    created_at, updated_at
`
```

### Converter Update
```go
// services/collection-service/internal/converter/collection_converter.go

func ModelToProto(collection *models.Collection) *pb.Collection {
    // ... existing mappings ...

    if collection.SettingsJSON != nil {
        settingsJSON, _ := json.Marshal(collection.SettingsJSON)
        protoCollection.SettingsJson = string(settingsJSON)
    }

    return protoCollection
}
```

## Success Criteria

- [ ] Repository queries include settings_json
- [ ] Converter maps field correctly
- [ ] Null values handled gracefully
- [ ] gRPC responses include settings_json when present

## Next Steps

After completing this phase:
1. Proceed to [Phase 3: Update GraphQL](./phase-03-update-graphql.md)
2. Update GraphQL schema and mappers
