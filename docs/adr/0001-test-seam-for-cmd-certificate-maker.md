# ADR-0001: Test Seam Pattern for cmd/certificate_maker KMS Stubbing

## Status

**ACCEPTED** - Implemented in Story 1.9

## Date

2025-01-11

## Context

### Problem

Command-layer tests in `cmd/certificate_maker` were experiencing flaky failures when AWS credentials were unavailable. The test suite was making real network calls to AWS KMS and EC2 Instance Metadata Service (IMDS), resulting in errors like:

```shell
Error: operation error KMS: GetPublicKey, get identity: get credentials: 
failed to refresh cached credentials, no EC2 IMDS role found
```

This caused:

- **Test Flakiness**: Tests passed in AWS environments but failed locally or in CI without credentials
- **Slow Test Execution**: Network calls and AWS SDK initialization added overhead
- **Poor Developer Experience**: Contributors couldn't run tests without AWS setup
- **CI/CD Instability**: Build pipelines failed inconsistently based on credential availability

### Requirements

1. Tests must run reliably without AWS credentials
2. No production behavior changes
3. Minimal code invasiveness
4. Maintainable and idiomatic Go
5. Preserve integration test value where appropriate

### Call Chain Analysis

The problematic path:

```
runCreate() 
  → certmaker.CreateCertificates()
    → certmaker.InitKMS()
      → kms.Get()  // ← Real AWS SDK call
        → aws.Config.LoadDefaultConfig()
          → EC2 IMDS metadata lookup
```

Core `pkg/certmaker` already had testable seams via the `InitKMS` function variable, but cmd-layer tests needed to stub even earlier in the chain.

## Decision

**We will introduce a minimal test seam by adding a package-level variable `kmsGet` in `pkg/certmaker/certmaker.go` that defaults to `kms.Get` but can be overridden in tests.**

### Implementation

**In `pkg/certmaker/certmaker.go`:**

```go
// kmsGet is a package-level variable to allow stubbing in tests
var kmsGet = kms.Get
```

**Replace all `kms.Get()` calls:**

```go
// Before
sv, err = kms.Get(ctx, ref, crypto.SHA256)

// After
sv, err = kmsGet(ctx, ref, crypto.SHA256)
```

**In `cmd/certificate_maker/certificate_maker_test.go`:**

```go
// Stub for problematic test case
old := certmaker.InitKMS
certmaker.InitKMS = func(_ context.Context, _ certmaker.KMSConfig) (signature.SignerVerifier, error) {
    return &mockSignerVerifier{}, nil
}
defer func() { certmaker.InitKMS = old }()
```

### Scope

- **Changed Lines**: 5 (1 variable declaration + 4 call-site updates)
- **Test Changes**: 1 mock type + 1 stub in failing test
- **Production Impact**: Zero (default behavior unchanged)

## Consequences

### Positive

1. **Test Reliability**
   - Tests run consistently without AWS credentials
   - No more flaky failures due to credential availability
   - Deterministic test behavior

2. **Developer Experience**
   - Contributors can run tests locally without AWS setup
   - Faster feedback loop (no network calls)
   - Lower barrier to entry for new contributors

3. **CI/CD Stability**
   - Build pipelines no longer depend on AWS credential configuration
   - Consistent test results across environments
   - Reduced troubleshooting time

4. **Minimal Invasiveness**
   - Single variable addition
   - Standard Go testing idiom
   - No architectural changes
   - Self-documenting with clear variable name

5. **Maintainability**
   - No special build configuration needed
   - Pattern is immediately recognizable to Go developers
   - Easy to understand and extend

### Negative

1. **Additional Indirection**
   - One more level of indirection for KMS calls
   - Developers must know about the seam for test modifications
   - **Mitigation**: Clear documentation in code comments and engineering notes

2. **Potential for Misuse**
   - Seam could theoretically be used in production code (though unlikely)
   - **Mitigation**: Variable is unexported, clear naming, documentation emphasizes test-only use

3. **Mock Maintenance**
   - Mock implementation must stay in sync with interface changes
   - **Mitigation**: Go compiler enforces interface compliance

### Neutral

1. **Integration Test Coverage**
   - Only failing test uses mock; others remain integration tests
   - Maintains valuable real-world validation
   - Balances isolation with integration testing

## Alternatives Considered

### Alternative 1: Environment Variable Hacks

**Approach**:

```go
os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
os.Setenv("AWS_ACCESS_KEY_ID", "fake")
os.Setenv("AWS_SECRET_ACCESS_KEY", "fake")
```

**Rejected Because**:

- **Brittle**: AWS SDK behavior changes over time
- **Host-Dependent**: Different behavior in CI vs local vs different OS
- **Leaky**: Environment state affects other tests
- **Indirect**: Still initializes AWS SDK, just with fake credentials
- **Maintenance Risk**: Breaking changes in SDK credential resolution

### Alternative 2: Dependency Injection Through CLI

**Approach**:

```go
func runCreate(cmd *cobra.Command, args []string, 
               kmsGetter func(context.Context, string, crypto.Hash) (signature.SignerVerifier, error)) error
```

**Rejected Because**:

- **API Churn**: Changes function signature unnecessarily
- **Complexity**: runCreate is called by Cobra framework, would need wrapper
- **Over-Engineering**: Too much machinery for single test fix
- **Production Impact**: Complicates production code for test-only need

### Alternative 3: Build-Tag Conditional Compilation

**Approach**:

```go
// +build !test
// certmaker.go
var kmsGet = kms.Get

// +build test
// certmaker_test.go
var kmsGet = mockKMS.Get
```

**Rejected Because**:

- **Complexity**: Multiple files to maintain
- **Confusion**: Hard to reason about which version is active
- **Build Overhead**: Requires special build tags
- **Maintenance Burden**: Changes require touching multiple files
- **Overkill**: Too much machinery for simple problem

### Alternative 4: Full Interface Extraction

**Approach**:
Create `KMSProvider` interface and inject implementation.

**Rejected Because**:

- **Over-Engineering**: Large refactoring for single test issue
- **Unnecessary Abstraction**: No other implementations needed
- **Delayed Value**: Benefits don't justify immediate cost
- **Note**: Could reconsider if multiple KMS implementations needed

## Implementation Notes

### Test Usage Pattern

```go
func TestSomething(t *testing.T) {
    // Save original
    old := certmaker.InitKMS
    
    // Install test double
    certmaker.InitKMS = func(_ context.Context, _ certmaker.KMSConfig) (signature.SignerVerifier, error) {
        return &mockSignerVerifier{}, nil
    }
    
    // Always restore (prevents test pollution)
    defer func() { certmaker.InitKMS = old }()
    
    // Run test
    // ...
}
```

### Mock Implementation Requirements

The mock must implement:

- `PublicKey(...signature.PublicKeyOption) (crypto.PublicKey, error)`
- `CryptoSigner(context.Context, func(error)) (crypto.Signer, crypto.SignerOpts, error)`
- `VerifySignature(io.Reader, io.Reader, ...signature.VerifyOption) error`
- `SignMessage(io.Reader, ...signature.SignOption) ([]byte, error)`

### Production Guarantee

The variable initialization `var kmsGet = kms.Get` ensures production code always uses the real implementation. Tests explicitly override via assignment.

## Validation

### Testing

- ✅ All pkg/certmaker tests pass (1.287s)
- ✅ All cmd/certificate_maker tests pass (5.372s)
- ✅ Previously failing test now passes consistently
- ✅ No regressions in other tests
- ✅ Tests run without AWS credentials

### Quality Gates

- ✅ golangci-lint: 0 issues
- ✅ Code review: APPROVED
- ✅ QA review: PASS (HIGH confidence)

### Verification

```bash
# Tests pass without AWS credentials
unset AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN
go test ./pkg/certmaker/... ./cmd/certificate_maker/... -v
# Result: PASS
```

## References

- **Story**: [docs/stories/epic-1.9-stabilize-cmd-tests.md](../stories/epic-1.9-stabilize-cmd-tests.md)
- **Engineering Note**: [docs/engineering-notes/test-seam-cmd-certificate-maker.md](../engineering-notes/test-seam-cmd-certificate-maker.md)
- **QA Gate**: [docs/qa/gates/epic-1.9-stabilize-cmd-tests.yml](../qa/gates/epic-1.9-stabilize-cmd-tests.yml)
- **Code Changes**:
  - [pkg/certmaker/certmaker.go](../../pkg/certmaker/certmaker.go)
  - [cmd/certificate_maker/certificate_maker_test.go](../../cmd/certificate_maker/certificate_maker_test.go)

## Related Patterns

- **Test Seam Pattern**: Michael Feathers, "Working Effectively with Legacy Code"
- **Dependency Injection**: Lightweight version for testing
- **Test Double**: Using mocks to replace external dependencies

## Future Considerations

### If This Approach Becomes Insufficient

Consider full interface extraction if:

1. Multiple KMS implementations are needed (e.g., local dev KMS)
2. More extensive stubbing required across many tests
3. Production code needs pluggable KMS backends

### Maintenance

- Keep mock in sync with interface changes (compiler enforces)
- Document any new usage of seam in test code
- Consider extracting mock to shared test package if reused widely

## Decision Makers

- **Proposed By**: James (Dev)
- **Reviewed By**: Quinn (QA)
- **Date**: 2025-01-11
- **Status**: ACCEPTED and Implemented

---

**Supersedes**: None (First ADR)  
**Superseded By**: None  
**Last Updated**: 2025-01-11
