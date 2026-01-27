---
title: "Phase 01: Cleanup Old Artifacts"
description: "Remove obsolete Tilt and serverless configuration files"
status: pending
priority: P1
effort: 30m
branch: feature/hybird-serverless-and-servers-infrastructure
tags: [cleanup, tilt, serverless]
created: 2026-01-28
---

# Phase 01: Cleanup Old Artifacts

## Context Links

- [Main Plan](plan.md)
- [Research Reports](../research/)
- [Brainstorm Report](../reports/brainstorm-260128-0037-docker-air-hot-reload-migration.md)

## Overview

**Priority:** P1 (must be done first)
**Current Status:** Pending
**Description:** Remove obsolete Tilt configuration files and old serverless setup artifacts that are no longer needed.

**Files to DELETE:**
- `Tiltfile.development` - Old Tilt dev config (superseded by Air)
- `infra/development/k8s/` - Old K8s manifests (if exists)

**Files to KEEP (unchanged):**
- `Tiltfile.production` - Production K8s setup (still needed)
- `docker-compose.yml` - Production Docker setup (unchanged)

---

## Key Insights

From research reports:
- **Tiltfile.development** was for serverless + K8s setup
- **Production Tiltfile** still needed for K8s deployments
- Removing old files reduces confusion and maintenance burden

---

## Requirements

### Functional Requirements
1. Delete `Tiltfile.development` from project root
2. Delete `infra/development/k8s/` directory if it exists
3. Verify `Tiltfile.production` remains intact
4. Clean up any references in documentation

### Non-Functional Requirements
- No impact on production Docker workflow
- Git history preserved
- No breaking changes to existing systems

---

## Architecture

**Before:**
```
project-root/
├── Tiltfile.development      ❌ DELETE
├── Tiltfile.production       ✅ KEEP
├── infra/
│   └── development/
│       └── k8s/              ❌ DELETE (if exists)
└── docker-compose.yml        ✅ KEEP
```

**After:**
```
project-root/
├── Tiltfile.production       ✅ KEEP
└── docker-compose.yml        ✅ KEEP
```

---

## Related Code Files

### Files to DELETE
- `E:\zuno-marketplace-api\Tiltfile.development`
- `E:\zuno-marketplace-api\infra\development\k8s\*` (if exists)

### Files to VERIFY (unchanged)
- `E:\zuno-marketplace-api\Tiltfile.production`
- `E:\zuno-marketplace-api\docker-compose.yml`
- `E:\zuno-marketplace-api\README.md` (check for Tilt references)

---

## Implementation Steps

### Step 1: Backup Verification (2 min)
```bash
# Verify git is clean (no uncommitted changes)
git status

# If changes exist, commit or stash them
git add .
git commit -m "Pre-cleanup commit" || git stash
```

**Success:** Git status shows clean working tree

**Rollback:** `git reset --hard HEAD` if needed

---

### Step 2: Check for K8s Directory (3 min)
```bash
# Check if infra/development/k8s exists
if [ -d "infra/development/k8s" ]; then
  echo "K8s directory exists, will delete"
  ls -la infra/development/k8s/
else
  echo "No K8s directory found, skipping"
fi
```

**Success:** Know whether k8s directory exists

**Rollback:** None (read-only check)

---

### Step 3: Delete Tiltfile.development (5 min)
```bash
# Delete the file
rm Tiltfile.development

# Verify deletion
ls Tiltfile* | grep -v production
# Should return empty (only Tiltfile.production exists)
```

**Success:** Only `Tiltfile.production` remains

**Rollback:** `git checkout Tiltfile.development`

---

### Step 4: Delete K8s Directory (5 min)
```bash
# If directory exists from Step 2
if [ -d "infra/development/k8s" ]; then
  rm -rf infra/development/k8s
  echo "Deleted infra/development/k8s/"
fi

# Clean up empty parent directories if needed
if [ -d "infra/development" ] && [ -z "$(ls -A infra/development)" ]; then
  rmdir infra/development
  echo "Removed empty infra/development/"
fi
```

**Success:** K8s directory removed

**Rollback:** `git checkout infra/development/k8s/`

---

### Step 5: Check Documentation References (10 min)
```bash
# Search for Tiltfile.development references
grep -r "Tiltfile.development" . --include="*.md" --exclude-dir=".git"

# Search for K8s references
grep -r "infra/development/k8s" . --include="*.md" --exclude-dir=".git"
```

If references found:
1. Update documentation to remove Tiltfile.development mentions
2. Keep Tiltfile.production references
3. Note: README.md updates will happen in Phase 07

**Success:** No outdated references remain (document for Phase 07)

**Rollback:** None (documentation updates in Phase 07)

---

### Step 6: Verify Production Files Intact (5 min)
```bash
# Verify Tiltfile.production exists
test -f Tiltfile.production && echo "✓ Tiltfile.production exists"

# Verify docker-compose.yml exists
test -f docker-compose.yml && echo "✓ docker-compose.yml exists"

# Verify production workflow still works
docker compose config >/dev/null 2>&1 && echo "✓ docker-compose.yml valid"
```

**Success:** All production files verified

**Rollback:** `git checkout Tiltfile.production docker-compose.yml`

---

## Todo List

- [ ] Backup: Verify git status is clean
- [ ] Check if `infra/development/k8s/` exists
- [ ] Delete `Tiltfile.development`
- [ ] Delete `infra/development/k8s/` (if exists)
- [ ] Search for documentation references
- [ ] Verify production files intact
- [ ] Test `docker compose config` validates
- [ ] Commit cleanup changes

---

## Success Criteria

✅ `Tiltfile.development` deleted
✅ `infra/development/k8s/` deleted (if existed)
✅ `Tiltfile.production` exists and unchanged
✅ `docker-compose.yml` validates successfully
✅ No broken references to deleted files
✅ Production workflow still works

**Validation Commands:**
```bash
# Verify deleted files
ls Tiltfile*  # Should show only Tiltfile.production
ls infra/development/k8s/ 2>/dev/null  # Should fail

# Verify production works
docker compose config  # Should succeed
```

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Accidentally delete Tiltfile.production | Low | High | Git checkout to restore |
| Delete infra/development with other files | Low | Medium | Check directory contents first |
| Documentation breaks with dead links | Medium | Low | Update in Phase 07 |
| Production workflow broken | Low | Critical | Verify with `docker compose config` |

---

## Security Considerations

None (cleanup only, no security impact)

---

## Next Steps

**After Phase 01:**
- Phase 02: Air Configuration (can start immediately)
- Phase 07: Update README to remove Tilt references

**Dependencies:**
- None (can start immediately)

**Follow-up Tasks:**
- Update README in Phase 07
- Update onboarding docs

---

## Rollback Plan

**Full Rollback:**
```bash
# Restore all deleted files
git checkout Tiltfile.development
git checkout infra/development/k8s/

# Verify restored
git status
```

**Partial Rollback:**
```bash
# Restore only Tiltfile.development
git checkout Tiltfile.development

# Restore only K8s directory
git checkout infra/development/k8s/
```
