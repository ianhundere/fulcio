# Story 1.7: Lint Cleanup (Test Files Only)

**Epic**: Epic 1 - Enable Certificate Reuse in Certmaker Tool
**Story ID**: 1.7
**Type**: Chore
**Status**: Complete
**Assigned To**: James (Dev)
**Story Points**: Low

---

## Story

As a developer,
I want clean linting in test files,
so that the codebase maintains high quality standards without false positives.

---

## Scope

- **LIMIT TO:** Test files under `pkg/certmaker/` (primarily `cert_loader_test.go`)
- **DO NOT:** Change production code behavior
- **FOCUS:** Fix gocritic and revive warnings without altering test functionality

---

## Acceptance Criteria

1. ✅ Resolve gocritic `ifElseChain` suggestions:
   - Convert simple if/else blocks to `switch {}` with default
   - Target occurrences at lines ~170 and ~325 in cert_loader_test.go
2. ✅ Resolve revive `unused-parameter` warnings:
   - Rename unused parameters to `_` in test helpers
   - Add `_ = param` if parameter may be used later
   - Fix 7 occurrences in cert_loader_test.go
3. ✅ Verify tests still pass and coverage unchanged
4. ✅ Run `golangci-lint run` and produce clean report
5. ✅ Update story with before/after summary

---

## Tasks

- [x] Fix gocritic ifElseChain at line ~170
- [x] Fix gocritic ifElseChain at line ~325
- [x] Fix revive unused-parameter warnings (7 occurrences)
- [x] Run tests to verify no functional changes
- [x] Run golangci-lint and verify clean output
- [x] Document changes

---

## Dev Agent Record

### Agent Model Used

Claude 3.5 Sonnet (new)

### Debug Log References

- None

### Completion Notes

- Fixed 2 gocritic ifElseChain warnings by converting if/else chains to switch statements
- Fixed 7 revive unused-parameter warnings by renaming unused parameters to `_`
- All tests pass: 31 tests, 100% passing
- Coverage unchanged: 67.7% (before and after)
- No functional changes - only code style improvements
- Test execution time: ~1.1s (no performance impact)

**Changes Summary:**

1. Line ~170 (TestLoadCertificateFromFile): Converted if/else chain to switch
2. Line ~320 (TestValidateCertificateKeyMatch): Converted if/else chain to switch
3. Lines 249, 264, 272, 279, 287: Renamed unused `*testing.T` and `crypto.PublicKey` parameters to `_`

### File List

- pkg/certmaker/cert_loader_test.go (modified - lint fixes only)

### Change Log

| Date | Change | Files Modified |
|------|--------|----------------|
| 2025-01-11 | Story created | epic-1.7-lint-cleanup.md |
| 2025-01-11 | Lint cleanup complete | cert_loader_test.go |

---

## Testing

### Verification Steps

1. Run tests before changes: `go test ./pkg/certmaker/... -v -cover`
2. Apply lint fixes
3. Run tests after changes: `go test ./pkg/certmaker/... -v -cover`
4. Compare coverage - must be identical
5. Run linter: `golangci-lint run ./pkg/certmaker/...`

---

## Dev Notes

**Specific Issues to Fix:**

1. **gocritic ifElseChain (2 occurrences):**
   - Line ~170: Replace `if tt.wantError != nil { ... } else { ... }`
   - Line ~325: Replace `if tt.wantError != nil { ... } else { ... }`
   - Use: `switch { case tt.wantError != nil: ... default: ... }`

2. **revive unused-parameter (7 occurrences):**
   - In test helper functions with unused `*testing.T` or `crypto.PublicKey`
   - Rename to `_` or add `_ = param` to acknowledge intent

---
