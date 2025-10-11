# Engineering Note: Test Seam for cmd/certificate_maker

**Date**: 2025-01-11
**Story**: 1.9 - Stabilize cmd/certificate_maker Tests by Stubbing KMS  
**Author**: James (Dev)

## Problem Statement

Command-layer tests in `cmd/certificate_maker` were reaching real AWS credential resolution and EC2 IMDS (Instance Metadata Service) via the `kms.Get()` → `PublicKey()` call chain, causing flaky test failures when AWS credentials weren't available:

```shell
Error: operation error KMS: GetPublicKey, get identity: get credentials: 
failed to refresh cached credentials, no EC2 IMDS role found
```

Core `pkg/certmaker` tests already use interface-based mocks successfully. The cmd layer needed an equivalent seam to inject test doubles without modifying production behavior.

## Solution: Minimal Test Seam

### What Changed

Introduced a package-level variable in `pkg/certmaker/certmaker.go`:

```go
// kmsGet is a package-level variable to allow stubbing in tests
var kmsGet = kms.Get
```

Replaced all direct `kms.Get()` calls with `kmsGet()` calls:

```go
// Before
sv, err = kms.Get(ctx, ref, crypto.SHA256)

// After  
sv, err = kmsGet(ctx, ref, crypto.SHA256)
```

**Total changes**: 1 variable declaration + 4 call-site updates

### Why This Approach

The cmd layer tests exercise the full `runCreate()` function, which internally calls `certmaker.CreateCertificates()`, which calls `certmaker.InitKMS()`, which calls `kms.Get()`. Without a seam, tests hit real AWS SDK code.

Core pkg tests already use `InitKMS` as a testable function variable, but cmd tests needed to stub even earlier in the chain since they test through the CLI entry point.

### How It Fixes Tests

Tests temporarily override `kmsGet` to return a mock `signature.SignerVerifier`:

```go
// In cmd/certificate_maker/certificate_maker_test.go
old := certmaker.InitKMS
certmaker.InitKMS = func(_ context.Context, _ certmaker.KMSConfig) (signature.SignerVerifier, error) {
    return &mockSignerVerifier{}, nil
}
defer func() { certmaker.InitKMS = old }()
```

The mock implements required interfaces:

- `PublicKey()` returns controlled error
- `CryptoSigner()` returns controlled error  
- `VerifySignature()` not implemented (not called)
- `SignMessage()` not implemented (not called)

This makes tests:

1. **Deterministic** - No network calls, no credential lookups
2. **Fast** - No AWS SDK initialization overhead
3. **Portable** - Run anywhere without AWS setup
4. **Focused** - Assert intended behavior, not AWS internals

### Alternatives Considered (and Rejected)

#### 1. Environment Variable Hacks

```go
os.Setenv("AWS_EC2_METADATA_DISABLED", "true")
os.Setenv("AWS_ACCESS_KEY_ID", "fake")
os.Setenv("AWS_SECRET_ACCESS_KEY", "fake")
```

**Rejected because**:

- Brittle - AWS SDK behavior changes over time
- Host-dependent - May behave differently in CI vs local
- Leaky - Environment state affects other tests
- Indirect - Still goes through AWS SDK initialization

#### 2. Dependency Injection Through runCreate

```go
func runCreate(cmd *cobra.Command, args []string, kmsGetter KMSGetterFunc) error
```

**Rejected because**:

- Larger API surface change
- Complicates CLI signature unnecessarily
- runCreate is called by cobra, would need wrapper
- Over-engineering for a single test case fix

#### 3. Build-Tag Shims

```go
// +build !test
var kmsGet = kms.Get

// +build test  
var kmsGet = mockKMS.Get
```

**Rejected because**:

- More moving parts (multiple files)
- Easy to confuse future contributors
- Harder to maintain test/prod parity
- Overkill for this use case

### Risk & Scope

**Production Impact**: **ZERO**

- Default value is `kms.Get` - identical to original behavior
- Only test code uses the override capability
- No conditional logic in production paths

**Scope**: Limited to `pkg/certmaker` package

- cmd tests import and stub the variable
- Production CLI never touches the seam
- Other packages unaffected

**Maintenance**: Minimal

- Standard Go testing idiom
- Self-documenting (variable name + comment)
- No special build steps required

### Usage Pattern in Tests

**Setup** (in test case that needs stubbing):

```go
// Save original
old := certmaker.InitKMS

// Install test double
certmaker.InitKMS = func(_ context.Context, _ certmaker.KMSConfig) (signature.SignerVerifier, error) {
    return &mockSignerVerifier{}, nil
}

// Always restore
defer func() { certmaker.InitKMS = old }()

// Run test
cmd.Execute()
```

**Mock Implementation**:

```go
type mockSignerVerifier struct{}

func (m *mockSignerVerifier) PublicKey(...signature.PublicKeyOption) (crypto.PublicKey, error) {
    return nil, errors.New("mock error getting public key")
}

func (m *mockSignerVerifier) CryptoSigner(context.Context, func(error)) (crypto.Signer, crypto.SignerOpts, error) {
    return nil, nil, errors.New("not implemented")
}

// ... other interface methods
```

### Test Coverage Strategy

**Important**: Only the previously-failing test case uses the stub. Other test cases continue to exercise real integration paths (when AWS credentials are available), maintaining valuable integration test coverage.

This targeted approach:

- Fixes flakiness where it occurred
- Preserves integration testing elsewhere
- Balances unit isolation with real-world validation

## Conclusion

This minimal, idiomatic seam enables reliable cmd-layer testing without:

- Changing production behavior
- Adding complexity to the CLI
- Requiring special build configuration
- Sacrificing integration test coverage

The pattern is standard in Go codebases and immediately recognizable to developers familiar with testable seams. Future contributors can easily understand and maintain this approach.

## References

- **Story**: docs/stories/epic-1.9-stabilize-cmd-tests.md
- **QA Gate**: docs/qa/gates/epic-1.9-stabilize-cmd-tests.yml
- **Modified Files**:
  - pkg/certmaker/certmaker.go (seam)
  - cmd/certificate_maker/certificate_maker_test.go (mock)
