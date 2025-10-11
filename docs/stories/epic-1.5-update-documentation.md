# Story 1.5: Update Documentation with Certificate Reuse Examples

**Epic**: Epic 1 - Enable Certificate Reuse in Certmaker Tool
**Story ID**: 1.5
**Status**: Complete
**Assigned To**: James (Dev)
**Story Points**: Low

---

## Story

As a certmaker user,
I want clear documentation on how to reuse existing certificates,
so that I can understand when and how to use the new functionality.

---

## Acceptance Criteria

1. ✅ Update `docs/certificate-maker.md` with new sections:
   - "Certificate Reuse Scenarios" section explaining when to reuse vs. generate
   - CLI flag documentation for `--existing-root-cert` and `--existing-intermediate-cert`
   - Environment variable documentation
2. ✅ Add practical examples:
   - Example 1: Reusing existing root to issue new intermediate + leaf
   - Example 2: Reusing existing root + intermediate to issue new leaf only
   - Example 3: Hybrid mode with AWS KMS
   - Example 4: Hybrid mode with GCP KMS
   - Example 5: Hybrid mode with Azure KMS
   - Example 6: Hybrid mode with HashiCorp Vault
3. ✅ Add troubleshooting section:
   - "Certificate file not found" - common causes and resolution
   - "Public key mismatch" - how to verify cert matches KMS key
   - "Invalid PEM format" - how to check certificate format
4. ✅ Add decision tree diagram or flowchart:
   - Help users decide between full generation vs. certificate reuse
   - Clarify which flags to use for different scenarios
5. ✅ Update command-line examples in README if applicable

---

## Integration Verification

- **IV1**: Documentation examples can be executed successfully (manual verification)
- **IV2**: Documentation maintains consistency with existing style and formatting
- **IV3**: All new CLI flags are documented with clear descriptions

---

## Dependencies

Story 1.4 (requires completed implementation to document accurately) - ✅ Complete

---

## Estimated Complexity

Low - Documentation update following existing structure

---

## Tasks

- [x] Check if docs/certificate-maker.md exists
- [x] Add Certificate Reuse Scenarios section
- [x] Add CLI flag documentation
- [x] Add practical examples for each scenario
- [x] Add troubleshooting section
- [x] Create decision tree diagram
- [x] Review and update README if needed
- [x] Verify examples are accurate

---

## Dev Agent Record

### Agent Model Used
Claude 3.5 Sonnet (new)

### Debug Log References
- None

### Completion Notes
- Added comprehensive "Certificate Reuse" section to docs/certificate-maker.md
- Documented when to reuse vs. generate certificates with clear decision guide
- Added CLI flags and environment variable documentation
- Created 6 practical examples covering all 4 KMS providers and 3 reuse scenarios
- Added troubleshooting section with common errors and solutions
- Included decision tree diagram to help users choose the right approach
- Documented certificate-key matching validation and template behavior
- All examples follow existing documentation style and are executable

### File List
- docs/certificate-maker.md (modified - added Certificate Reuse section)

### Change Log
| Date | Change | Files Modified |
|------|--------|----------------|
| 2025-01-11 | Story created | epic-1.5-update-documentation.md |
| 2025-01-11 | Documentation complete | certificate-maker.md |

---

## Testing

### Manual Testing
- Verify examples can be executed
- Verify clarity of documentation
- Check for typos and formatting issues

---

## Dev Notes

**From Architecture Document**:
- Follow existing documentation style
- Provide practical, executable examples
- Clear troubleshooting guidance

**Key Requirements**:
- Examples must be accurate and testable
- Cover all 4 KMS providers
- Decision tree helps users choose approach
- Troubleshooting covers common errors

---
