# Certmaker Tool Architecture Document

**Exclusive Focus: Certificate Maker Utility for Certificate Chain Generation**

**Related Issue**: [#2070 - Using certmaker to issue a new leaf cert from existing intermediate/root certs](https://github.com/sigstore/fulcio/issues/2070)

---

## Introduction

This document captures the CURRENT STATE of the certmaker tool architecture and implementation. Certmaker is a standalone CLI utility for creating certificate chains (root → intermediate → leaf) using KMS providers. This document focuses EXCLUSIVELY on certmaker components and ignores all Fulcio server functionality.

### Document Scope

**IN SCOPE**:

- `cmd/certificate_maker/` - CLI application
- `pkg/certmaker/` - Certificate creation logic
- KMS provider integration (`github.com/sigstore/sigstore/pkg/signature/kms`)
- Certificate template processing (`go.step.sm/crypto/x509util`)
- Enhancement for certificate reuse (Issue #2070)

**OUT OF SCOPE** (Not covered in this document):

- Fulcio server components (`cmd/app/`, `pkg/server/`)
- OIDC authentication handlers (`pkg/identity/`, `pkg/oauthflow/`)
- Certificate transparency log integration (`pkg/ctl/`)
- All other Fulcio server packages

### Change Log

| Date       | Version | Description                              | Author      |
|------------|---------|------------------------------------------|-------------|
| 2025-01-11 | 1.0     | Certmaker-focused architecture document  | Winston (Architect) |

---

## Quick Reference - Certmaker Components

### Critical Files

**CLI Layer** (`cmd/certificate_maker/`):

- `certificate_maker.go` - Main entry point, CLI flags, configuration binding
- `certificate_maker_test.go` - CLI integration tests

**Core Package** (`pkg/certmaker/`):

- `certmaker.go` - Certificate creation workflow, KMS initialization
- `certmaker_test.go` - Unit and integration tests with mock KMS
- `template.go` - Template parsing and certificate generation
- `template_test.go` - Template processing tests
- `templates/` - Embedded JSON templates (root, intermediate, leaf)

**External Dependencies**:

- `github.com/sigstore/sigstore/pkg/signature/kms` - KMS abstraction
- `go.step.sm/crypto/x509util` - Certificate template engine
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management

### Enhancement Impact (Issue #2070)

**Files to Modify**:

1. `pkg/certmaker/certmaker.go` - Add certificate loading capability
2. `cmd/certificate_maker/certificate_maker.go` - Add CLI flags

**New Files**:

1. `pkg/certmaker/cert_loader.go` - Certificate file loading functions
2. `pkg/certmaker/cert_loader_test.go` - Loading tests

---

## Certmaker Architecture

### Component Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    CERTMAKER TOOL                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  CLI Layer (cmd/certificate_maker)                    │  │
│  │  - Flag parsing (cobra)                               │  │
│  │  - Configuration (viper)                              │  │
│  │  - Command execution                                  │  │
│  └───────────────────┬──────────────────────────────────┘  │
│                      │                                       │
│                      ▼                                       │
│  ┌──────────────────────────────────────────────────────┐  │
│  │  Core Logic (pkg/certmaker)                          │  │
│  │                                                        │  │
│  │  ┌──────────────────────────────────────────────┐   │  │
│  │  │ CreateCertificates() - Main Workflow         │   │  │
│  │  │  - KMS initialization                         │   │  │
│  │  │  - Template processing                        │   │  │
│  │  │  - Certificate generation                     │   │  │
│  │  │  - File output (PEM)                          │   │  │
│  │  └──────────────────────────────────────────────┘   │  │
│  │                                                        │  │
│  │  ┌──────────────────────────────────────────────┐   │  │
│  │  │ Template Processing (template.go)            │   │  │
│  │  │  - JSON template loading                      │   │  │
│  │  │  - x509util integration                       │   │  │
│  │  │  - Certificate structure creation             │   │  │
│  │  └──────────────────────────────────────────────┘   │  │
│  └───────────┬──────────────────┬───────────────────────┘  │
│              │                  │                           │
│              ▼                  ▼                           │
│  ┌─────────────────┐  ┌──────────────────────────────┐    │
│  │  KMS Providers  │  │  Certificate Templates       │    │
│  │  (external)     │  │  (embedded)                  │    │
│  │  - AWS KMS      │  │  - root-template.json        │    │
│  │  - GCP KMS      │  │  - intermediate-template.json│    │
│  │  - Azure KV     │  │  - leaf-template.json        │    │
│  │  - HashiVault   │  │                              │    │
│  └─────────────────┘  └──────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Technology Stack

| Component           | Technology                                | Version  | Purpose                                    |
|---------------------|-------------------------------------------|----------|--------------------------------------------|
| Runtime             | Go                                        | 1.24.6   | Programming language                       |
| CLI Framework       | github.com/spf13/cobra                    | v1.10.1  | Command structure and flag parsing         |
| Configuration       | github.com/spf13/viper                    | v1.20.1  | Config binding (flags ↔ env vars)          |
| KMS Abstraction     | sigstore/sigstore                         | v1.9.6   | Unified KMS interface                      |
| AWS KMS             | sigstore/pkg/signature/kms/aws            | v1.9.5   | AWS KMS provider                           |
| GCP KMS             | sigstore/pkg/signature/kms/gcp            | v1.9.6   | Google Cloud KMS provider                  |
| Azure KMS           | sigstore/pkg/signature/kms/azure          | v1.9.5   | Azure Key Vault provider                   |
| HashiVault          | sigstore/pkg/signature/kms/hashivault     | v1.9.5   | HashiCorp Vault provider                   |
| Cert Templates      | go.step.sm/crypto                         | v0.70.0  | x509util - Template to certificate         |
| Testing             | github.com/stretchr/testify               | v1.11.1  | Test assertions and mocking                |
| Logging             | go.uber.org/zap                           | v1.27.0  | Structured logging                         |

---

## Source Tree - Certmaker Components Only

```
fulcio/
├── cmd/
│   └── certificate_maker/              # Certmaker CLI application
│       ├── certificate_maker.go        # Main: flags, viper binding, runCreate()
│       └── certificate_maker_test.go   # CLI-level integration tests
│
├── pkg/
│   ├── certmaker/                      # Core certmaker logic
│   │   ├── certmaker.go                # CreateCertificates(), InitKMS(), validation
│   │   ├── certmaker_test.go           # Mock-based unit/integration tests
│   │   ├── template.go                 # ParseTemplate(), template handling
│   │   ├── template_test.go            # Template parsing tests
│   │   └── templates/                  # Embedded JSON templates
│   │       ├── root-template.json
│   │       ├── intermediate-template.json
│   │       └── leaf-template.json
│   │
│   └── log/                            # Shared logging (used by certmaker)
│       └── log.go                      # zap logger configuration
│
├── docs/
│   ├── certificate-maker.md            # User documentation
│   ├── certificate-specification.md    # Certificate requirements
│   ├── prd.md                          # Enhancement PRD
│   └── architecture.md                 # This document
│
├── go.mod                              # Go module dependencies
└── Makefile                            # Build target: cert-maker
```

---

## Current Certmaker Workflow

### Certificate Generation Process

**Entry Point**: User executes `./certificate-maker create [common-name]`

**Flow**:

```
1. CLI Initialization (cmd/certificate_maker/certificate_maker.go)
   ├─ Parse CLI flags (cobra)
   ├─ Bind environment variables (viper)
   ├─ Validate KMS configuration
   └─ Call runCreate()

2. Main Workflow (pkg/certmaker/certmaker.go::CreateCertificates)
   │
   ├─ ROOT CERTIFICATE
   │  ├─ Initialize KMS signer for root key
   │  ├─ Get root public key from KMS
   │  ├─ Load template (embedded or file)
   │  ├─ Parse template → x509.Certificate
   │  ├─ Create self-signed root cert
   │  └─ Write to PEM file
   │
   ├─ INTERMEDIATE CERTIFICATE (optional)
   │  ├─ Initialize KMS signer for intermediate key
   │  ├─ Get intermediate public key
   │  ├─ Load intermediate template
   │  ├─ Parse template → x509.Certificate
   │  ├─ Sign with root signer
   │  └─ Write to PEM file
   │
   └─ LEAF CERTIFICATE (optional)
      ├─ Initialize KMS signer for leaf key
      ├─ Get leaf public key
      ├─ Load leaf template
      ├─ Parse template → x509.Certificate
      ├─ Sign with intermediate (or root if no intermediate)
      └─ Write to PEM file

3. Output
   └─ PEM-encoded certificate files written to specified paths
```

### Current Limitations (Issue #2070)

**Problem**: Certificate generation is always "all or nothing"

**Missing Capabilities**:

- ❌ Cannot load existing root certificate to sign new intermediate
- ❌ Cannot load existing intermediate certificate to sign new leaf
- ❌ Cannot reuse stable certificates while rotating others

**Why This Matters**:

- Root certificates typically have 10-year lifetimes
- Intermediate certificates typically have 5-year lifetimes  
- Leaf certificates may need frequent rotation
- Current implementation forces regeneration of entire chain

**Root Cause**: `CreateCertificates()` has no mechanism to:

1. Accept file paths to existing certificates
2. Load and parse PEM certificates
3. Validate loaded certificates against KMS keys
4. Skip generation when existing cert provided

---

## Key Components Deep Dive

### 1. CLI Layer (cmd/certificate_maker)

**Responsibilities**:

- Command-line interface using cobra
- Configuration management with viper
- Flag-to-environment variable binding
- Calling core certificate creation logic

**Main Function**: `runCreate()`

**Key Operations**:

```go
// Build KMS configuration from flags/env vars
config := certmaker.KMSConfig{
    Type:  viper.GetString("kms-type"),
    KeyID: viper.GetString("root-key-id"),
    // ... other config
}

// Call core creation logic
return certmaker.CreateCertificates(
    config,
    viper.GetString("root-template"),
    viper.GetString("leaf-template"),
    // ... other parameters
)
```

**Configuration Sources** (priority order):

1. CLI flags (e.g., `--kms-type awskms`)
2. Environment variables (e.g., `KMS_TYPE=awskms`)
3. Default values

**Viper Binding Pattern**:

```go
mustBindPFlag("root-key-id", createCmd.Flags().Lookup("root-key-id"))
mustBindEnv("root-key-id", "KMS_ROOT_KEY_ID")
```

### 2. Core Logic (pkg/certmaker/certmaker.go)

**Main Function**: `CreateCertificates()`

**Signature** (current):

```go
func CreateCertificates(
    config KMSConfig,
    rootTemplatePath, leafTemplatePath string,
    rootCertPath, leafCertPath string,
    intermediateKeyID, intermediateTemplatePath, intermediateCertPath string,
    leafKeyID string,
    rootLifetime, intermediateLifetime, leafLifetime time.Duration,
) error
```

**Core Operations**:

1. **KMS Initialization**:

```go
func InitKMS(ctx context.Context, config KMSConfig) (signature.SignerVerifier, error) {
    switch config.Type {
    case "awskms":
        ref := fmt.Sprintf("awskms:///%s", config.KeyID)
        return kms.Get(ctx, ref, crypto.SHA256)
    case "gcpkms":
        ref := fmt.Sprintf("gcpkms://%s", config.KeyID)
        return kms.Get(ctx, ref, crypto.SHA256)
    // ... other providers
    }
}
```

2. **Certificate Creation**:

```go
// Get public key from KMS
rootPubKey, err := sv.PublicKey()

// Get crypto.Signer for signing operations
cryptoSV := sv.(CryptoSignerVerifier)
rootSigner, _, err := cryptoSV.CryptoSigner(ctx, nil)

// Parse template and create certificate
rootCert, err := x509util.CreateCertificate(
    rootTmpl,      // Parsed template
    rootTmpl,      // Parent (self for root)
    rootPubKey,    // Public key
    rootSigner,    // Signer
)
```

3. **File Output**:

```go
func WriteCertificateToFile(cert *x509.Certificate, filename string) error {
    block := &pem.Block{
        Type:  "CERTIFICATE",
        Bytes: cert.Raw,
    }
    return pem.Encode(file, block)
}
```

**Configuration Validation**:

```go
func ValidateKMSConfig(config KMSConfig) error {
    // Validates KMS type, key ID format, required options
    // Provider-specific validation (AWS region, Azure tenant, etc.)
}
```

### 3. Template Processing (pkg/certmaker/template.go)

**Purpose**: Convert JSON templates to x509.Certificate structures

**Main Function**:

```go
func ParseTemplate(
    input interface{},           // Template content (string or []byte)
    parent *x509.Certificate,    // Parent cert (for Issuer, AuthorityKeyId)
    notAfter time.Time,          // Certificate expiration
    publicKey crypto.PublicKey,  // Public key to embed
    commonName string,           // Subject CN
) (*x509.Certificate, error)
```

**Template Format** (JSON with go-template support):

```json
{
  "subject": {
    "commonName": "{{ .Subject.CommonName }}"
  },
  "keyUsage": ["certSign", "crlSign"],
  "basicConstraints": {
    "isCA": true,
    "maxPathLen": 1
  }
}
```

**Processing Steps**:

1. Load template (from file or embedded)
2. Create base x509.Certificate with public key and validity
3. Use `x509util.NewCertificateFromX509()` to parse template
4. Set Issuer and AuthorityKeyId if parent cert provided
5. Return populated certificate structure

**Embedded Templates**:

- Compiled into binary via `//go:embed`
- Defaults for root, intermediate, leaf
- Can be overridden with custom templates via CLI flags

### 4. KMS Integration

**Abstraction Layer**: `github.com/sigstore/sigstore/pkg/signature/kms`

**Key Interface**: `signature.SignerVerifier`

```go
type SignerVerifier interface {
    PublicKey() (crypto.PublicKey, error)
    SignMessage(io.Reader, ...SignOption) ([]byte, error)
    VerifySignature(signature, message io.Reader, ...VerifyOption) error
}
```

**Extended Interface** (for signing):

```go
type CryptoSignerVerifier interface {
    SignerVerifier
    CryptoSigner(context.Context, func(error)) (crypto.Signer, crypto.SignerOpts, error)
}
```

**Provider Initialization Patterns**:

**AWS KMS**:

```go
// Key ID: arn:aws:kms:region:account:key/id or alias/name
// Requires: AWS_REGION environment variable
ref := "awskms:///arn:aws:kms:us-west-2:123456789012:key/abc-123"
sv, err := kms.Get(ctx, ref, crypto.SHA256)
```

**GCP KMS**:

```go
// Key ID: Full resource path with version
ref := "gcpkms://projects/proj/locations/loc/keyRings/ring/cryptoKeys/key/cryptoKeyVersions/1"
sv, err := kms.Get(ctx, ref, crypto.SHA256)
```

**Azure Key Vault**:

```go
// Key ID: azurekms:name=X;vault=Y format
// Requires: AZURE_TENANT_ID environment variable
ref := "azurekms://vault-name.vault.azure.net/key-name"
sv, err := kms.Get(ctx, ref, crypto.SHA256)
```

**HashiCorp Vault**:

```go
// Key ID: Simple key name (not full transit path)
// Requires: VAULT_TOKEN, VAULT_ADDR environment variables
ref := "hashivault://my-key"
sv, err := kms.Get(ctx, ref, crypto.SHA256)
```

**Common Operations**:

```go
// Get public key
pubKey, err := sv.PublicKey()

// Get crypto.Signer for certificate creation
signer, opts, err := sv.(CryptoSignerVerifier).CryptoSigner(ctx, nil)

// Sign certificate
cert, err := x509util.CreateCertificate(tmpl, parent, pubKey, signer)
```

---

## Enhancement Design (Issue #2070)

### Proposed Changes

**1. New Certificate Loading Module** (`pkg/certmaker/cert_loader.go`):

```go
// LoadCertificateFromFile reads PEM certificate from file
func LoadCertificateFromFile(path string) (*x509.Certificate, error) {
    // Read file
    // Decode PEM block
    // Parse X.509 certificate
    // Return certificate or error
}

// ValidateCertificateKeyMatch verifies cert matches KMS key
func ValidateCertificateKeyMatch(
    cert *x509.Certificate,
    sv signature.SignerVerifier,
) error {
    // Get public key from certificate
    // Get public key from KMS
    // Compare keys
    // Return error if mismatch
}
```

**2. Modified CreateCertificates Signature**:

```go
func CreateCertificates(
    config KMSConfig,
    rootTemplatePath, leafTemplatePath string,
    rootCertPath, leafCertPath string,
    intermediateKeyID, intermediateTemplatePath, intermediateCertPath string,
    leafKeyID string,
    rootLifetime, intermediateLifetime, leafLifetime time.Duration,
    existingRootCertPath string,          // NEW
    existingIntermediateCertPath string,  // NEW
) error
```

**3. Conditional Certificate Handling**:

```go
// Root certificate - load or generate
var rootCert *x509.Certificate
if existingRootCertPath != "" {
    // Load existing certificate
    rootCert, err = LoadCertificateFromFile(existingRootCertPath)
    // Validate against KMS key
    err = ValidateCertificateKeyMatch(rootCert, rootSV)
} else {
    // Generate new certificate (existing logic)
    rootCert, err = generateRootCertificate(...)
}

// Similar pattern for intermediate
```

**4. New CLI Flags** (`cmd/certificate_maker/certificate_maker.go`):

```go
// Flag definitions
createCmd.Flags().String("existing-root-cert", "", 
    "Path to existing root certificate PEM file")
createCmd.Flags().String("existing-intermediate-cert", "", 
    "Path to existing intermediate certificate PEM file")

// Environment variable binding
mustBindEnv("existing-root-cert", "EXISTING_ROOT_CERT")
mustBindEnv("existing-intermediate-cert", "EXISTING_INTERMEDIATE_CERT")
```

### Hybrid Modes Supported

1. **Existing Root + New Intermediate + New Leaf**:
   - Load root certificate from file
   - Validate root matches root KMS key
   - Generate new intermediate (signed by loaded root)
   - Generate new leaf (signed by new intermediate)

2. **Existing Root + Existing Intermediate + New Leaf**:
   - Load root and intermediate certificates
   - Validate both against KMS keys
   - Generate new leaf (signed by loaded intermediate)

3. **Existing Root + New Leaf** (direct signing):
   - Load root certificate
   - Validate root matches root KMS key
   - Generate new leaf (signed directly by root, no intermediate)

4. **Full Generation** (backward compatible):
   - No existing cert flags provided
   - Generate all certificates (current behavior)

### Integration Points

**Certificate Loading**:

- Uses standard Go `encoding/pem` for PEM decoding
- Uses `crypto/x509` for certificate parsing
- No new external dependencies

**KMS Validation**:

- Reuses existing `signature.SignerVerifier.PublicKey()` method
- Public key comparison using Go's crypto package
- Provider-agnostic validation

**Template System**:

- Templates become optional for loaded certificates
- Certificate already contains Subject, validity, etc.
- Templates only used for newly generated certificates

**Backward Compatibility**:

- Empty string for new parameters = existing behavior
- All existing tests pass with new signature
- No breaking changes to CLI or API

---

## Development Standards

### Build Process

**Build certmaker binary**:

```bash
make cert-maker
```

**Output**: `./certificate-maker` executable with embedded templates

### Testing Approach

**Test Structure**:

```
pkg/certmaker/
├── certmaker_test.go       # Main integration tests
├── template_test.go        # Template parsing tests  
└── cert_loader_test.go     # NEW: Certificate loading tests
```

**Mock Strategy**:

```go
// mockSignerVerifier implements signature.SignerVerifier and CryptoSignerVerifier
type mockSignerVerifier struct {
    key crypto.Signer
    // ... mock methods
}
```

**Test Coverage Requirements**:

- Existing code: Maintain >80% coverage
- New code: Achieve >85% coverage
- Edge cases: Invalid PEM, missing files, key mismatches

**Running Tests**:

```bash
# All certmaker tests
go test ./pkg/certmaker/... -v

# With coverage
go test ./pkg/certmaker/... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Code Standards

**Error Handling**:

```go
if err != nil {
    return fmt.Errorf("loading certificate from %s: %w", path, err)
}
```

**Logging**:

```go
import "github.com/sigstore/fulcio/pkg/log"

log.Logger.Info("Certificate loaded", 
    zap.String("path", path),
    zap.String("subject", cert.Subject.String()))
```

**Function Documentation**:

```go
// LoadCertificateFromFile reads a PEM-encoded X.509 certificate from
// the specified file path and returns the parsed certificate.
//
// Returns an error if the file doesn't exist, cannot be read, contains
// invalid PEM data, or the certificate structure is malformed.
func LoadCertificateFromFile(path string) (*x509.Certificate, error)
```

---

## Certificate Chain Validation

### Verification Commands

**Verify leaf against intermediate**:

```bash
openssl verify -CAfile intermediate.pem leaf.pem
```

**Verify full chain**:

```bash
cat intermediate.pem root.pem > chain.pem
openssl verify -CAfile chain.pem leaf.pem
```

**Inspect certificate details**:

```bash
openssl x509 -in cert.pem -text -noout
```

### Expected Certificate Structure

**Root Certificate**:

- Issuer == Subject (self-signed)
- IsCA: true
- KeyUsage: Certificate Sign, CRL Sign
- BasicConstraints: CA:TRUE, pathlen:1

**Intermediate Certificate**:

- Issuer: Root Subject
- IsCA: true
- KeyUsage: Certificate Sign, CRL Sign
- BasicConstraints: CA:TRUE, pathlen:0
- AuthorityKeyId: Root's SubjectKeyId

**Leaf Certificate**:

- Issuer: Intermediate (or Root if no intermediate)
- IsCA: false
- KeyUsage: Digital Signature
- ExtKeyUsage: Code Signing
- BasicConstraints: CA:FALSE
- AuthorityKeyId: Signer's SubjectKeyId

---

## Operational Considerations

### Configuration Management

**Environment Variables** (all optional):

```bash
# KMS configuration
KMS_TYPE=awskms
AWS_REGION=us-west-2
KMS_ROOT_KEY_ID=arn:aws:kms:us-west-2:123456789012:key/abc-123

# Certificate output paths
ROOT_CERT=root.pem
INTERMEDIATE_CERT=intermediate.pem
LEAF_CERT=leaf.pem

# NEW: Existing certificate paths
EXISTING_ROOT_CERT=/path/to/existing/root.pem
EXISTING_INTERMEDIATE_CERT=/path/to/existing/intermediate.pem
```

**CLI Flags Override Environment Variables**:

```bash
./certificate-maker create "CN" \
  --kms-type awskms \
  --existing-root-cert /path/to/root.pem \
  --leaf-key-id $LEAF_KEY_ID
```

### Security Considerations

**Certificate Files**:

- Certificates are PUBLIC information (not secrets)
- No special file permissions required for certificate files
- Private keys NEVER leave KMS - only public keys in certificates

**Validation**:

- Always validate loaded certificates against KMS public keys
- Detect and reject mismatched certificates before any signing
- Clear error messages guide users to resolution

**Logging**:

- Log certificate operations (INFO level)
- Never log private key material
- Log validation failures (ERROR level) with actionable messages

---

## Appendix - Common Operations

### Creating a Complete Chain (Current)

```bash
./certificate-maker create "https://fulcio.example.com" \
  --kms-type awskms \
  --aws-region us-west-2 \
  --root-key-id "arn:aws:kms:us-west-2:123:key/root" \
  --intermediate-key-id "arn:aws:kms:us-west-2:123:key/intermediate" \
  --leaf-key-id "arn:aws:kms:us-west-2:123:key/leaf" \
  --root-cert root.pem \
  --intermediate-cert intermediate.pem \
  --leaf-cert leaf.pem
```

### Reusing Root to Issue New Leaf (Proposed)

```bash
./certificate-maker create "https://fulcio.example.com" \
  --kms-type awskms \
  --aws-region us-west-2 \
  --existing-root-cert root.pem \
  --root-key-id "arn:aws:kms:us-west-2:123:key/root" \
  --leaf-key-id "arn:aws:kms:us-west-2:123:key/new-leaf" \
  --leaf-cert new-leaf.pem
```

### Debugging Tips

**Enable verbose logging**:

```go
log.ConfigureLogger("dev")  // Development mode with debug logs
```

**Common Errors**:

| Error Message | Cause | Resolution |
|---------------|-------|------------|
| "Invalid KMS configuration" | Incorrect key ID format | Check provider-specific key ID syntax |
| "Public key mismatch" | Certificate doesn't match KMS key | Verify correct certificate file provided |
| "Template parsing error" | Invalid JSON in template | Validate JSON syntax with `jq` |
| "Certificate not found" | File path incorrect | Check file exists at specified path |

---
