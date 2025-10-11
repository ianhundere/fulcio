# Story 1.1: Implement Certificate Loading Infrastructure

**Epic**: Epic 1 - Enable Certificate Reuse in Certmaker Tool
**Story ID**: 1.1
**Status**: Complete
**Assigned To**: James (Dev)
**Story Points**: Medium

---

## Story

As a certmaker developer,
I want a robust certificate loading mechanism from PEM files,
so that existing certificates can be parsed, validated, and integrated into the certificate creation workflow.

---

## Acceptance Criteria

1. ✅ Create new file `pkg/certmaker/cert_loader.go` with certificate loading functions
2. ✅ Implement `LoadCertificateFromFile(path string) (*x509.Certificate, error)` function that:
   - Reads PEM-encoded certificate files
   - Parses and validates X.509 certificate structure
   - Returns clear errors for invalid files, missing files, or malformed certificates
3. ✅ Implement `ValidateCertificateKeyMatch(cert *x509.Certificate, signerVerifier signature.SignerVerifier) error` function that:
   - Extracts public key from certificate
   - Compares with public key from KMS signer
   - Returns clear error if keys don't match
4. ✅ Create comprehensive test file `pkg/certmaker/cert_loader_test.go` with:
   - Test cases for valid PEM certificates
   - Test cases for invalid/corrupted files
   - Test cases for missing files
   - Test cases for public key match validation (both matching and mismatching scenarios)
5. ✅ Achieve >85% code coverage for new certificate loading code

---

## Integration Verification

- **IV1**: Existing certmaker tests in `certmaker_test.go` continue to pass without modification
- **IV2**: New loading functions integrate with existing error handling patterns (`fmt.Errorf` wrapping)
- **IV3**: Certificate loading does not introduce performance regression (benchmark comparison)

---

## Dependencies

None - this is foundation work

---

## Estimated Complexity

Medium - New code but straightforward file I/O and validation logic

---

## Tasks

- [x] Create `pkg/certmaker/cert_loader.go` file with certificate loading functions
- [x] Implement `LoadCertificateFromFile()` function with PEM parsing and validation
- [x] Implement `ValidateCertificateKeyMatch()` function for public key comparison
- [x] Create `pkg/certmaker/cert_loader_test.go` with comprehensive test cases
- [x] Verify >85% code coverage for new code
- [x] Run existing certmaker tests to ensure no regression
- [x] Validate error handling follows project patterns

---

## Dev Agent Record

### Agent Model Used

Claude 3.5 Sonnet (new)

### Debug Log References

- None

### Completion Notes

- Implemented certificate loading with robust error handling (93.8% coverage)
- Implemented public key validation supporting RSA, ECDSA, and Ed25519 (100% coverage)
- All 23 test cases passing including edge cases and error conditions
- No regression - all existing certmaker tests pass
- Error handling follows project patterns (fmt.Errorf wrapping)
- Fixed Go typed nil issue in tests
- Fixed Ed25519 public key comparison logic

### File List

- pkg/certmaker/cert_loader.go (new)
- pkg/certmaker/cert_loader_test.go (new)

### Change Log

| Date | Change | Files Modified |
|------|--------|----------------|
| 2025-01-11 | Story created | epic-1.1-certificate-loading-infrastructure.md |
| 2025-01-11 | Implemented certificate loading infrastructure | cert_loader.go, cert_loader_test.go |

---

## Testing

### Unit Tests

- Test `LoadCertificateFromFile()` with valid PEM
- Test `LoadCertificateFromFile()` with invalid PEM
- Test `LoadCertificateFromFile()` with missing file
- Test `ValidateCertificateKeyMatch()` with matching keys
- Test `ValidateCertificateKeyMatch()` with mismatched keys

### Integration Tests

- Verify existing certmaker tests pass (regression check)
- Verify error messages follow existing patterns

---

## Dev Notes

**From Architecture Document**:

- Use `encoding/pem` for PEM decoding
- Use `crypto/x509` for certificate parsing
- Follow error wrapping pattern: `fmt.Errorf("context: %w", err)`
- Use existing logging via `github.com/sigstore/fulcio/pkg/log`
- Function naming: PascalCase for exported (e.g., `LoadCertificateFromFile`)
- Error variables: ErrPrefix convention (e.g., `ErrCertificateNotFound`)

**Key Requirements**:

- No new external dependencies
- Must integrate with existing `signature.SignerVerifier` interface
- Public key comparison must be provider-agnostic
- Clear, actionable error messages

---
