---
title: "Add settings_json Field to Collection GraphQL Schema"
description: "Expose the settings_json database column in GraphQL API to support rich collection metadata"
status: completed
priority: P1
effort: 4h
branch: feature/add-collection-settings-json
tags: [graphql, collection, metadata, api]
created: 2026-02-08
---

# Add settings_json Field to Collection GraphQL Schema

## Overview

Expose the existing `settings_json` JSONB column from the `collections` table in the GraphQL API. This field will store rich collection metadata for the launchpad UI, including overview content and utility accordions.

## Background

The `collections` table already has a `settings_json` column (defined in migration `000003_add_collections_tables.up.sql`), but it's not exposed in the GraphQL schema. The frontend needs this field to display rich content in collection accordions.

## Architecture

```
Database (PostgreSQL)
├── collections table
│   └── settings_json (JSONB) - EXISTS but not exposed
│
GraphQL Schema (NEW)
├── Collection type
│   └── settingsJson: JSON - NEW FIELD
│
Proto (gRPC)
├── Collection message
│   └── settings_json: string - NEW FIELD
```

## Implementation Phases

| Phase | Name | Status | Priority |
|-------|------|--------|----------|
| 1 | [Update Proto Definition](./phase-01-update-proto.md) | Pending | High |
| 2 | [Update Database Repository](./phase-02-update-repository.md) | Pending | High |
| 3 | [Update GraphQL Schema](./phase-03-update-graphql.md) | Pending | High |
| 4 | [Update Mappers](./phase-04-update-mappers.md) | Pending | High |
| 5 | [Testing & Validation](./phase-05-testing.md) | Pending | Medium |

## Proposed JSON Structure

The `settings_json` field will store:

```json
{
  "overview": {
    "description": "Rich text description...",
    "roleInGameplay": {
      "title": "Role in Gameplay",
      "items": ["Item 1", "Item 2"]
    },
    "rpgProgression": {
      "title": "RPG Progression",
      "items": ["Item 1", "Item 2"]
    },
    "flexibleUsage": {
      "title": "Flexible Usage",
      "items": ["Item 1", "Item 2"]
    },
    "ecosystem": "Ecosystem description..."
  },
  "utility": {
    "title": "Utility",
    "items": [
      { "label": "Utility 1", "description": "Description..." },
      { "label": "Utility 2", "description": "Description..." }
    ]
  }
}
```

## Files to Modify

| File | Change |
|------|--------|
| `proto/collection.proto` | Add `settings_json` field to Collection message |
| `services/collection-service/internal/models/collection.go` | Add SettingsJSON field |
| `services/collection-service/internal/repository/collection_repository.go` | Include in queries |
| `services/graphql-gateway/graph/schemas/collection.graphqls` | Add settingsJson field |
| `services/graphql-gateway/internal/mapper/collection_mapper.go` | Map proto to GraphQL |

## Success Criteria

- [ ] `settingsJson` field available in GraphQL Collection type
- [ ] Field returns valid JSON when data exists
- [ ] Field returns null when data is empty
- [ ] Existing queries continue to work
- [ ] No breaking changes to API

## Dependencies

- None - field already exists in database

## Notes

- This is a **read-only** field for now
- Frontend will handle JSON parsing and validation
- No migration needed - column already exists
