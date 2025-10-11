# Fulcio Brownfield Enhancement PRD

**Certmaker Tool: Issue Leaf Certificates from Existing Root/Intermediate Certificates**

---

## Intro Project Analysis and Context

### Existing Project Overview

#### Analysis Source

- **IDE-based fresh analysis** of Fulcio repository

#### Current Project State

Fulcio is a free-to-use certificate authority for issuing code signing certificates for OpenID Connect (OIDC) identities. It operates in General Availability with a 99.5% availability SLO and follows semver rules for API stability.

The certmaker tool (`pkg/certmaker` and `cmd/certificate_maker`) is a certificate creation utility that supports creating root, intermediate, and leaf certificates using multiple KMS providers (AWS, GCP, Azure, HashiCorp Vault). Currently, the tool generates complete certificate chains from scratch but lacks the ability to issue new leaf certificates from existing intermediate or root certificates.

**Primary Purpose**: Certificate authority service for code signing with short-lived (10-minute) certificates  
**Technology Stack**: Go-based service with Protocol Buffers API
**Current Certmaker Functionality**: Creates complete certificate chains (root → intermediate → leaf) but cannot reuse existing certificates

### Available Documentation Analysis

#### Available Documentation

- [x] Tech Stack Documentation - Go modules, KMS integrations
- [x] Source Tree/Architecture - Standard Go project structure
- [x] Coding Standards - Go coding conventions, error handling patterns
- [x] API Documentation - gRPC/HTTP API (fulcio.proto)
- [ ] External API Documentation
- [ ] UX/UI Guidelines - N/A (CLI tool)
- [x] Technical Debt Documentation - Issue #2070 identifies current limitation
- [x] Certificate Specification - docs/certificate-specification.md
- [x] Certificate Maker Documentation - docs/certificate-maker.md

**Note**: The project has comprehensive documentation covering certificate specifications, HSM support, OIDC integration, and security model. The certmaker tool itself is well-documented but requires enhancement for the certificate reuse use case.

### Enhancement Scope Definition

#### Enhancement Type

- [x] Major Feature Modification
- [ ] New Feature Addition
- [ ] Integration with New Systems
- [ ] Performance/Scalability Improvements
- [ ] UI/UX Overhaul
- [ ] Technology Stack Upgrade
- [ ] Bug Fix and Stability Improvements

#### Enhancement Description

Refactor and extend the certmaker tool to support issuing new leaf certificates from existing intermediate or root certificates, instead of requiring the generation of an entirely new certificate chain. This addresses Fulcio issue #2070 and enables users to maintain stable root/intermediate certificates while issuing new leaf certificates as needed.

#### Impact Assessment

- [x] Moderate Impact (some existing code changes)
- [ ] Minimal Impact (isolated additions)
- [ ] Significant Impact (substantial existing code changes)
- [ ] Major Impact (architectural changes required)

**Impact Details**: The enhancement requires modifications to the certmaker tool's workflow and certificate loading logic, but does not affect Fulcio's core server functionality, API, or CT log integration. Changes are isolated to the certmaker package and CLI.

### Goals and Background Context

#### Goals

- Enable certmaker to load and use existing root/intermediate certificates for signing new leaf certificates
- Maintain backward compatibility with existing certmaker workflows (creating full chains from scratch)
- Support all existing KMS providers (AWS, GCP, Azure, HashiCorp Vault) with certificate reuse
- Provide clear CLI flags and documentation for certificate reuse scenarios
- Ensure issued certificates meet Fulcio's certificate specification requirements

#### Background Context

The current certmaker implementation requires users to generate complete certificate chains (root → intermediate → leaf) even when they only need to issue a new leaf certificate. This creates operational overhead and inconsistency, as users must manage multiple root/intermediate certificates when they should be able to maintain a stable certificate hierarchy.

Issue #2070 highlights this limitation: users want to reuse existing root and intermediate certificates to sign new leaf certificates without regenerating the entire chain. This is a common operational pattern in certificate management where root and intermediate certificates have longer lifetimes and should remain stable, while leaf certificates are frequently rotated.

This enhancement fits within Fulcio's existing architecture as an improvement to the certmaker utility tool, supporting more flexible certificate management workflows without changing core Fulcio server behavior.

### Change Log

| Date | Version | Description | Author |
|------|---------|-------------|--------|
| 2025-01-11 | 1.0 | Initial brownfield PRD for certmaker enhancement | BMad Master |

---

## Requirements

### Functional

1. **FR1**: The certmaker tool shall accept existing root certificate file paths via a new CLI flag (`--existing-root-cert`) to load and use pre-existing root certificates instead of generating new ones.

2. **FR2**: The certmaker tool shall accept existing intermediate certificate file paths via a new CLI flag (`--existing-intermediate-cert`) to load and use pre-existing intermediate certificates.

3. **FR3**: When an existing root or intermediate certificate is provided, the tool shall extract and validate the certificate's public key and use the corresponding KMS key for signing operations.

4. **FR4**: The tool shall support a hybrid mode where users can provide existing root certificates while generating new intermediate certificates, or provide both existing root and intermediate certificates while generating only new leaf certificates.

5. **FR5**: The tool shall validate that loaded existing certificates match the KMS key IDs provided (public key verification) and fail with a clear error message if there's a mismatch.

6. **FR6**: When generating leaf certificates from existing intermediate/root certificates, the tool shall correctly set the Issuer field and AuthorityKeyId extension to match the signing certificate.

7. **FR7**: The tool shall maintain backward compatibility - existing workflows that generate full chains from scratch shall continue to work without modification.

8. **FR8**: Certificate loading functionality shall support PEM-encoded certificate files as input format.

9. **FR9**: The tool shall provide clear error messages when certificate files cannot be loaded, parsed, or validated against the provided KMS keys.

10. **FR10**: All generated certificates (whether from existing chains or new chains) shall meet Fulcio's certificate specification requirements as defined in `docs/certificate-specification.md`.

### Non Functional

1. **NFR1**: The enhancement shall maintain the existing certmaker tool's performance characteristics - certificate generation shall complete within the same time bounds as current implementation.

2. **NFR2**: The refactored code shall maintain test coverage at or above current levels (>80% for pkg/certmaker package).

3. **NFR3**: The implementation shall follow existing Go coding standards and conventions used throughout the Fulcio codebase, including error handling patterns and logging practices.

4. **NFR4**: CLI flag additions shall follow the existing naming conventions and parameter patterns established in cmd/certificate_maker/certificate_maker.go.

5. **NFR5**: Documentation updates shall maintain consistency with the existing style and structure of docs/certificate-maker.md.

6. **NFR6**: The implementation shall not introduce new external dependencies beyond those already used in the Fulcio project.

7. **NFR7**: Error messages and logging shall provide sufficient detail for troubleshooting while maintaining security (no exposure of private key material).

### Compatibility Requirements

1. **CR1: Existing API Compatibility** - All existing CLI flags, environment variables, and command structures shall continue to function exactly as before. No breaking changes to the existing certmaker interface.

2. **CR2: KMS Provider Compatibility** - The enhancement shall work with all currently supported KMS providers (AWS KMS, Google Cloud KMS, Azure Key Vault, HashiCorp Vault) without requiring provider-specific modifications.

3. **CR3: Certificate Format Compatibility** - Output certificates shall maintain the same PEM-encoded format and structure as currently generated certificates, ensuring compatibility with existing certificate consumers.

4. **CR4: Template Compatibility** - The existing embedded templates (root-template.json, intermediate-template.json, leaf-template.json) shall continue to work, and custom user-provided templates shall remain compatible.

---

## Technical Constraints and Integration Requirements

### Existing Technology Stack

**Languages**: Go 1.24.6
**Frameworks**:

- cobra/viper for CLI and configuration management
- sigstore/sigstore for KMS abstractions and signing operations
- go.step.sm/crypto/x509util for X.509 certificate template processing

**Key Dependencies**:

- github.com/sigstore/sigstore/pkg/signature/kms - KMS provider abstraction layer
- crypto/x509 - Standard Go X.509 certificate handling
- encoding/pem - PEM encoding/decoding

**KMS Providers**:

- AWS KMS (`github.com/sigstore/sigstore/pkg/signature/kms/aws`)
- Google Cloud KMS (`github.com/sigstore/sigstore/pkg/signature/kms/gcp`)
- Azure Key Vault (`github.com/sigstore/sigstore/pkg/signature/kms/azure`)
- HashiCorp Vault (`github.com/sigstore/sigstore/pkg/signature/kms/hashivault`)

**Testing**: Go standard testing with testify for assertions

**Build System**: Makefile with target `cert-maker` for building the binary

### Integration Approach

**Database Integration Strategy**: N/A - certmaker is a standalone CLI tool with no database dependencies

**API Integration Strategy**: N/A - certmaker does not expose APIs; it's a command-line utility

**Frontend Integration Strategy**: N/A - CLI-only tool

**Certificate Loading and Validation Integration**:

- Add new functions in `pkg/certmaker/certmaker.go` to load and parse PEM-encoded certificates from files
- Integrate certificate loading into the existing `CreateCertificates` function workflow
- Use existing `signature.SignerVerifier` interface for public key extraction and validation
- Leverage Go's `crypto/x509` package for certificate parsing and validation

**Testing Integration Strategy**:

- Extend existing test suite in `pkg/certmaker/certmaker_test.go` with new test cases for certificate loading
- Add test cases for hybrid modes (existing root + new intermediate, etc.)
- Mock file I/O operations for unit tests using `os.ReadFile` test helpers
- Validate certificate chain construction with existing certificates in integration tests

### Code Organization and Standards

**File Structure Approach**:

```
pkg/certmaker/
├── certmaker.go          # Main certificate creation logic (modify existing)
├── certmaker_test.go     # Tests (extend existing)
├── cert_loader.go        # NEW: Certificate loading functions
├── cert_loader_test.go   # NEW: Certificate loading tests
├── template.go           # Existing template handling (minimal changes)
├── template_test.go      # Existing template tests
└── templates/            # Existing embedded templates (no changes)
```

**Naming Conventions**:

- New CLI flags: `--existing-root-cert`, `--existing-intermediate-cert` (kebab-case, consistent with existing flags)
- New functions: `LoadCertificateFromFile`, `ValidateCertificateKeyMatch` (PascalCase for exported, camelCase for private)
- Error variables: `ErrCertificateNotFound`, `ErrKeyMismatch` (ErrPrefix convention)

**Coding Standards**:

- Follow existing error wrapping pattern: `fmt.Errorf("context: %w", err)`
- Use existing logging via `github.com/sigstore/fulcio/pkg/log` package
- Maintain consistency with existing KMS initialization pattern in `InitKMS` function
- Follow existing flag binding pattern with viper in `cmd/certificate_maker/certificate_maker.go`

**Documentation Standards**:

- GoDoc comments for all exported functions and types
- Update `docs/certificate-maker.md` with new CLI flags and usage examples
- Add inline code comments for complex certificate validation logic
- Include examples for each certificate reuse scenario in documentation

### Deployment and Operations

**Build Process Integration**:

- No changes to existing Makefile - the `cert-maker` target will build the enhanced binary
- Binary remains statically linked with embedded templates
- No new build dependencies required

**Deployment Strategy**:

- Certmaker is distributed as a standalone binary via GitHub releases
- Users download and run locally or in CI/CD pipelines
- No server-side deployment concerns

**Monitoring and Logging**:

- Use existing `log.Logger` from `pkg/log` for operational logging
- Log certificate loading operations at INFO level
- Log validation failures at ERROR level with actionable messages
- No metrics/telemetry - tool runs locally

**Configuration Management**:

- Maintain existing viper-based configuration with environment variables
- Add new flags to existing flag structure in `cmd/certificate_maker/certificate_maker.go`
- Support both CLI flags and environment variables for new options (e.g., `EXISTING_ROOT_CERT`)

### Risk Assessment and Mitigation

**Technical Risks**:

1. **Risk**: Certificate/key mismatch causing signing failures
   - **Mitigation**: Implement robust public key comparison before any signing operations

2. **Risk**: Invalid or corrupted certificate files causing parse failures
   - **Mitigation**: Comprehensive PEM parsing with clear error messages and validation

3. **Risk**: Breaking changes to existing workflows
   - **Mitigation**: Extensive regression testing, new flags are purely additive

**Integration Risks**:

1. **Risk**: KMS provider differences in public key handling
   - **Mitigation**: Use common `signature.SignerVerifier` interface, test with all providers

2. **Risk**: Template processing incompatibility with loaded certificates
   - **Mitigation**: Reuse existing template parsing logic, validate Issuer/AuthorityKeyId setting

**Deployment Risks**:

1. **Risk**: User confusion about when to use new flags vs. old workflow
   - **Mitigation**: Clear documentation with decision tree and examples for each scenario

2. **Risk**: Security concerns about certificate file handling
   - **Mitigation**: Follow Go best practices for file permissions, clear documentation on securing cert files

**Mitigation Strategies**:

- Comprehensive test coverage including edge cases (expired certs, wrong key types, etc.)
- Clear error messages that guide users to resolution
- Documentation includes troubleshooting section
- Maintain backward compatibility as primary constraint

---

## Epic and Story Structure

### Epic Approach

**Epic Structure Decision**: Single comprehensive epic for certmaker enhancement

**Rationale**: This enhancement is a focused modification to a single tool (certmaker) with a clear, cohesive goal - enabling certificate reuse. All work items are tightly related and build upon each other sequentially. Splitting into multiple epics would create artificial boundaries and dependencies. The single epic approach ensures:

- Clear context for all implementation work
- Logical progression from infrastructure to features to documentation
- Easier tracking of the overall enhancement completion
- Cohesive integration testing of all components together

---

## Epic 1: Enable Certificate Reuse in Certmaker Tool

**Epic Goal**: Extend the certmaker tool to support issuing new leaf certificates from existing root/intermediate certificates while maintaining full backward compatibility with existing workflows and all KMS providers.

**Integration Requirements**:

- Must integrate seamlessly with existing `CreateCertificates` function workflow
- Must support all four KMS providers (AWS, GCP, Azure, HashiCorp Vault) without provider-specific code paths
- Must maintain existing CLI interface and add new flags non-disruptively
- Must preserve existing certificate generation workflows for users who don't use the new functionality

### Story 1.1: Implement Certificate Loading Infrastructure

As a certmaker developer,
I want a robust certificate loading mechanism from PEM files,
so that existing certificates can be parsed, validated, and integrated into the certificate creation workflow.

**Acceptance Criteria**:

1. Create new file `pkg/certmaker/cert_loader.go` with certificate loading functions
2. Implement `LoadCertificateFromFile(path string) (*x509.Certificate, error)` function that:
   - Reads PEM-encoded certificate files
   - Parses and validates X.509 certificate structure
   - Returns clear errors for invalid files, missing files, or malformed certificates
3. Implement `ValidateCertificateKeyMatch(cert *x509.Certificate, signerVerifier signature.SignerVerifier) error` function that:
   - Extracts public key from certificate
   - Compares with public key from KMS signer
   - Returns clear error if keys don't match
4. Create comprehensive test file `pkg/certmaker/cert_loader_test.go` with:
   - Test cases for valid PEM certificates
   - Test cases for invalid/corrupted files
   - Test cases for missing files
   - Test cases for public key match validation (both matching and mismatching scenarios)
5. Achieve >85% code coverage for new certificate loading code

**Integration Verification**:

- IV1: Existing certmaker tests in `certmaker_test.go` continue to pass without modification
- IV2: New loading functions integrate with existing error handling patterns (`fmt.Errorf` wrapping)
- IV3: Certificate loading does not introduce performance regression (benchmark comparison)

**Dependencies**: None - this is foundation work

**Estimated Complexity**: Medium - New code but straightforward file I/O and validation logic

---

### Story 1.2: Refactor CreateCertificates to Support Certificate Reuse

As a certmaker developer,
I want the CreateCertificates function to conditionally load or generate certificates,
so that existing certificates can be reused while maintaining backward compatibility.

**Acceptance Criteria**:

1. Modify `CreateCertificates` function signature to accept optional existing certificate paths:

   ```go
   func CreateCertificates(config KMSConfig,
       rootTemplatePath, leafTemplatePath string,
       rootCertPath, leafCertPath string,
       intermediateKeyID, intermediateTemplatePath, intermediateCertPath string,
       leafKeyID string,
       rootLifetime, intermediateLifetime, leafLifetime time.Duration,
       existingRootCertPath, existingIntermediateCertPath string) error
   ```

2. Implement conditional logic:
   - If `existingRootCertPath` is provided: Load certificate using `LoadCertificateFromFile`, validate against root KMS key, skip root generation
   - If `existingRootCertPath` is empty: Generate root certificate as current behavior
   - Same pattern for intermediate certificate with `existingIntermediateCertPath`
3. Ensure signing certificate selection logic works correctly:
   - When using existing root + generating intermediate: Use loaded root cert for signing
   - When using existing intermediate: Use loaded intermediate for signing leaf
4. Validate that Issuer and AuthorityKeyId are correctly set when using loaded certificates
5. Update all existing tests to pass empty strings for new parameters (maintains compatibility)
6. Add new tests for hybrid scenarios:
   - Existing root + new intermediate + new leaf
   - Existing root + existing intermediate + new leaf
   - Existing root only + new leaf (direct signing)

**Integration Verification**:

- IV1: All existing certmaker test cases pass with new function signature (backward compatible)
- IV2: Certificate chains created with loaded certificates validate correctly using `openssl verify`
- IV3: KMS signing operations work identically whether cert is generated or loaded

**Dependencies**: Story 1.1 (requires certificate loading functions)

**Estimated Complexity**: High - Core workflow modification requiring careful testing

---

### Story 1.3: Add CLI Flags for Certificate Reuse

As a certmaker user,
I want new CLI flags to specify existing certificate paths,
so that I can easily reuse existing root and intermediate certificates.

**Acceptance Criteria**:

1. Add new CLI flags in `cmd/certificate_maker/certificate_maker.go`:
   - `--existing-root-cert`: Path to existing root certificate PEM file (optional)
   - `--existing-intermediate-cert`: Path to existing intermediate certificate PEM file (optional)
2. Add corresponding environment variables:
   - `EXISTING_ROOT_CERT`
   - `EXISTING_INTERMEDIATE_CERT`
3. Bind new flags using viper following existing patterns (`mustBindPFlag`, `mustBindEnv`)
4. Update `runCreate` function to pass new parameters to `CreateCertificates`
5. Add validation logic:
   - If `--existing-root-cert` is provided but file doesn't exist, fail with clear error before KMS initialization
   - If both `--existing-root-cert` and `--root-template` are provided, log warning that template will be ignored
6. Maintain existing behavior when new flags are not provided (empty string defaults)

**Integration Verification**:

- IV1: Existing CLI commands continue to work without new flags (backward compatibility verified)
- IV2: Help text (`--help`) clearly documents new flags and their purpose
- IV3: Environment variable support works identically to CLI flags

**Dependencies**: Story 1.2 (requires updated CreateCertificates function)

**Estimated Complexity**: Low - Standard CLI flag addition following established patterns

---

### Story 1.4: Comprehensive Integration Testing Across KMS Providers

As a certmaker developer,
I want comprehensive integration tests across all KMS providers,
so that certificate reuse works reliably regardless of which KMS backend is used.

**Acceptance Criteria**:

1. Extend `certmaker_test.go` with integration test scenarios:
   - Test certificate reuse with mocked AWS KMS provider
   - Test certificate reuse with mocked GCP KMS provider  
   - Test certificate reuse with mocked Azure KMS provider
   - Test certificate reuse with mocked HashiCorp Vault provider
2. Each KMS provider test must cover:
   - Loading existing root cert and generating intermediate + leaf
   - Loading existing root + intermediate and generating only leaf
   - Validation failure scenario (mismatched keys)
3. Add error path testing:
   - Certificate file not found
   - Invalid PEM format
   - Certificate expired
   - Public key mismatch with KMS key
4. Verify certificate chain integrity:
   - Generated certificate chains validate using `x509.Verify`
   - Issuer fields correctly reference loaded certificates
   - AuthorityKeyId extensions properly set
5. Test maintains >80% overall code coverage for pkg/certmaker package

**Integration Verification**:

- IV1: All existing integration tests continue to pass (regression check)
- IV2: New tests exercise both old workflow (generate all) and new workflow (reuse certs)
- IV3: Test execution time remains within acceptable bounds (< 30 seconds total)

**Dependencies**: Story 1.3 (requires complete implementation to test end-to-end)

**Estimated Complexity**: Medium - Comprehensive testing but using existing test infrastructure

---

### Story 1.5: Update Documentation with Certificate Reuse Examples

As a certmaker user,
I want clear documentation on how to reuse existing certificates,
so that I can understand when and how to use the new functionality.

**Acceptance Criteria**:

1. Update `docs/certificate-maker.md` with new sections:
   - "Certificate Reuse Scenarios" section explaining when to reuse vs. generate
   - CLI flag documentation for `--existing-root-cert` and `--existing-intermediate-cert`
   - Environment variable documentation
2. Add practical examples:
   - Example 1: Reusing existing root to issue new intermediate + leaf
   - Example 2: Reusing existing root + intermediate to issue new leaf only
   - Example 3: Hybrid mode with AWS KMS
   - Example 4: Hybrid mode with GCP KMS
   - Example 5: Hybrid mode with Azure KMS
   - Example 6: Hybrid mode with HashiCorp Vault
3. Add troubleshooting section:
   - "Certificate file not found" - common causes and resolution
   - "Public key mismatch" - how to verify cert matches KMS key
   - "Invalid PEM format" - how to check certificate format
4. Add decision tree diagram or flowchart:
   - Help users decide between full generation vs. certificate reuse
   - Clarify which flags to use for different scenarios
5. Update command-line examples in README if applicable

**Integration Verification**:

- IV1: Documentation examples can be executed successfully (manual verification)
- IV2: Documentation maintains consistency with existing style and formatting
- IV3: All new CLI flags are documented with clear descriptions

**Dependencies**: Story 1.4 (requires completed implementation to document accurately)

**Estimated Complexity**: Low - Documentation update following existing structure

---

### Story 1.6: Validation and Backward Compatibility Testing

As a certmaker developer,
I want comprehensive validation that existing workflows are not broken,
so that current users can upgrade without any disruption.

**Acceptance Criteria**:

1. Create backward compatibility test suite:
   - Test all existing certmaker examples from documentation still work
   - Test with all KMS providers using existing flags only
   - Test with custom templates (ensure template compatibility)
   - Test with different certificate lifetimes
2. Validate that:
   - Default behavior (no new flags) is identical to previous version
   - Output certificate format and structure unchanged
   - Certificate chains validate identically to before
   - Performance characteristics maintained (benchmark comparison)
3. Test edge cases:
   - Very short lifetime certificates (1 hour)
   - Very long lifetime certificates (20 years)
   - All supported key algorithms (ECDSA P-256, RSA, Ed25519 if supported)
4. Cross-version validation:
   - Certificates generated by old version can be loaded by new version
   - New version generates certificates compatible with existing Fulcio server
5. Create regression test script that can be run before each release

**Integration Verification**:

- IV1: All existing certmaker workflows produce identical output (certificate comparison)
- IV2: No performance regression in certificate generation (benchmark within 5% of baseline)
- IV3: Existing Fulcio integration points remain functional (if applicable)

**Dependencies**: Story 1.5 (final validation before completion)

**Estimated Complexity**: Medium - Thorough testing but primarily using existing test infrastructure

---

### Epic Completion Criteria

**All stories completed AND:**

1. ✅ Test coverage for pkg/certmaker package at or above 80%
2. ✅ All existing certmaker tests pass without modification
3. ✅ Documentation updated and reviewed
4. ✅ Backward compatibility validated through regression testing
5. ✅ All four KMS providers tested with certificate reuse
6. ✅ Certificate chains generated with existing certificates validate correctly
7. ✅ Clear error messages implemented for all failure scenarios
8. ✅ Performance maintained within acceptable bounds

**Rollback Plan**: If critical issues discovered post-implementation:

- New flags are optional, so users can simply not use them
- Existing workflows unchanged, so rollback is implicit (don't use new features)
- If code rollback needed, stories are isolated enough to revert individually

---
