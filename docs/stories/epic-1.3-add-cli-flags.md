# Story 1.3: Add CLI Flags for Certificate Reuse

**Epic**: Epic 1 - Enable Certificate Reuse in Certmaker Tool
**Story ID**: 1.3
**Status**: Complete
**Assigned To**: James (Dev)
**Story Points**: Low

---

## Story

As a certmaker user,
I want new CLI flags to specify existing certificate paths,
so that I can easily reuse existing root and intermediate certificates.

---

## Acceptance Criteria

1. ✅ Add new CLI flags in `cmd/certificate_maker/certificate_maker.go`:
   - `--existing-root-cert`: Path to existing root certificate PEM file (optional)
   - `--existing-intermediate-cert`: Path to existing intermediate certificate PEM file (optional)
2. ✅ Add corresponding environment variables:
   - `EXISTING_ROOT_CERT`
   - `EXISTING_INTERMEDIATE_CERT`
3. ✅ Bind new flags using viper following existing patterns (`mustBindPFlag`, `mustBindEnv`)
4. ✅ Update `runCreate` function to pass new parameters to `CreateCertificates`
5. ✅ Add validation logic:
   - If `--existing-root-cert` is provided but file doesn't exist, fail with clear error before KMS initialization
   - If both `--existing-root-cert` and `--root-template` are provided, log warning that template will be ignored
6. ✅ Maintain existing behavior when new flags are not provided (empty string defaults)

---

## Integration Verification

- **IV1**: Existing CLI commands continue to work without new flags (backward compatibility verified)
- **IV2**: Help text (`--help`) clearly documents new flags and their purpose
- **IV3**: Environment variable support works identically to CLI flags

---

## Dependencies

Story 1.2 (requires updated CreateCertificates function) - ✅ Complete

---

## Estimated Complexity

Low - Standard CLI flag addition following established patterns

---

## Tasks

- [x] Add `--existing-root-cert` flag definition
- [x] Add `--existing-intermediate-cert` flag definition
- [x] Add environment variable bindings for new flags
- [x] Update `runCreate` to pass flag values to CreateCertificates
- [x] Add file existence validation before KMS initialization
- [x] Add warning log when template + existing cert both provided
- [x] Test CLI with new flags
- [x] Test environment variable support
- [x] Verify help text displays correctly
- [x] Verify backward compatibility (no flags = existing behavior)

---

## Dev Agent Record

### Agent Model Used
Claude 3.5 Sonnet (new)

### Debug Log References
- None yet

### Completion Notes
- Added --existing-root-cert and --existing-intermediate-cert CLI flags
- Added EXISTING_ROOT_CERT and EXISTING_INTERMEDIATE_CERT environment variables
- Implemented file existence validation before KMS initialization
- Added warning logs when both template and existing cert provided
- Help text correctly displays new flags with clear descriptions
- All existing tests pass - full backward compatibility maintained
- Flags properly bound with viper following project patterns

### File List
- cmd/certificate_maker/certificate_maker.go (modified)

### Change Log
| Date | Change | Files Modified |
|------|--------|----------------|
| 2025-01-11 | Story created | epic-1.3-add-cli-flags.md |
| 2025-01-11 | Implementation complete | certificate_maker.go |

---

## Testing

### Manual Testing
- Test `--existing-root-cert` flag with valid PEM file
- Test `--existing-intermediate-cert` flag with valid PEM file
- Test both flags together
- Test with non-existent file paths (should error)
- Test environment variables EXISTING_ROOT_CERT and EXISTING_INTERMEDIATE_CERT
- Test `--help` output includes new flags
- Test backward compatibility (no new flags = works as before)

### Integration Tests
- Verify flags properly pass values to CreateCertificates
- Verify file validation happens before KMS initialization
- Verify warning logged when template and existing cert both provided

---

## Dev Notes

**From Architecture Document**:
- Follow existing flag naming conventions (kebab-case)
- Use viper for flag binding (mustBindPFlag, mustBindEnv)
- Add validation in runCreate before calling CreateCertificates
- Log warnings using pkg/log package

**Key Requirements**:
- Flags are optional - empty string when not provided
- File existence check before expensive KMS initialization
- Clear error messages when files not found
- Warning (not error) when template ignored due to existing cert
- Environment variables follow SCREAMING_SNAKE_CASE convention

---
