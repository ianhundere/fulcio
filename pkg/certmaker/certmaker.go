// Copyright 2024 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

// Package certmaker implements a certificate creation utility for Fulcio.
// It supports creating root, intermediate, and leaf certs using (AWS, GCP, Azure, HashiVault).
package certmaker

import (
	"bytes"
	"context"
	"crypto"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sigstore/sigstore/pkg/signature"
	"github.com/sigstore/sigstore/pkg/signature/kms"
	"github.com/sigstore/sigstore/pkg/signature/options"

	// Initialize AWS KMS provider
	_ "github.com/sigstore/sigstore/pkg/signature/kms/aws"
	// Initialize Azure KMS provider
	_ "github.com/sigstore/sigstore/pkg/signature/kms/azure"
	// Initialize GCP KMS provider
	_ "github.com/sigstore/sigstore/pkg/signature/kms/gcp"
	// Initialize HashiVault KMS provider
	_ "github.com/sigstore/sigstore/pkg/signature/kms/hashivault"
	"go.step.sm/crypto/x509util"
)

type signerWrapper struct {
	signature.SignerVerifier
}

func (s signerWrapper) Public() crypto.PublicKey {
	key, _ := s.PublicKey()
	return key
}

func (s signerWrapper) Sign(_ io.Reader, digest []byte, _ crypto.SignerOpts) ([]byte, error) {
	return s.SignMessage(bytes.NewReader(digest), options.WithDigest(digest))
}

// KMSConfig holds config for KMS providers.
type KMSConfig struct {
	Type              string
	Region            string
	RootKeyID         string
	IntermediateKeyID string
	LeafKeyID         string
	Options           map[string]string
}

// InitKMS initializes KMS provider based on the given config, KMSConfig.
var InitKMS = func(ctx context.Context, config KMSConfig) (signature.SignerVerifier, error) {
	if err := ValidateKMSConfig(config); err != nil {
		return nil, fmt.Errorf("invalid KMS configuration: %w", err)
	}

	// Falls back to LeafKeyID if root is not set
	keyID := config.RootKeyID
	if keyID == "" {
		keyID = config.LeafKeyID
	}

	var sv signature.SignerVerifier
	var err error

	switch config.Type {
	case "awskms":
		ref := fmt.Sprintf("awskms:///%s", keyID)
		if config.Region != "" {
			os.Setenv("AWS_REGION", config.Region)
		}
		sv, err = kms.Get(ctx, ref, crypto.SHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS KMS: %w", err)
		}

	case "gcpkms":
		ref := fmt.Sprintf("gcpkms://%s", keyID)
		sv, err = kms.Get(ctx, ref, crypto.SHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize GCP KMS: %w", err)
		}

	case "azurekms":
		keyURI := keyID
		if strings.HasPrefix(keyID, "azurekms:name=") {
			nameStart := strings.Index(keyID, "name=") + 5
			vaultIndex := strings.Index(keyID, ";vault=")
			if vaultIndex != -1 {
				keyName := strings.TrimSpace(keyID[nameStart:vaultIndex])
				vaultName := strings.TrimSpace(keyID[vaultIndex+7:])
				keyURI = fmt.Sprintf("azurekms://%s.vault.azure.net/%s", vaultName, keyName)
			}
		}
		if config.Options != nil && config.Options["tenant-id"] != "" {
			os.Setenv("AZURE_TENANT_ID", config.Options["tenant-id"])
			os.Setenv("AZURE_ADDITIONALLY_ALLOWED_TENANTS", "*")
		}
		os.Setenv("AZURE_AUTHORITY_HOST", "https://login.microsoftonline.com/")

		sv, err = kms.Get(ctx, keyURI, crypto.SHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Azure KMS: %w", err)
		}

	case "hashivault":
		keyURI := fmt.Sprintf("hashivault://%s", keyID)
		if config.Options != nil {
			if token := config.Options["token"]; token != "" {
				os.Setenv("VAULT_TOKEN", token)
			}
			if addr := config.Options["address"]; addr != "" {
				os.Setenv("VAULT_ADDR", addr)
			}
		}

		sv, err = kms.Get(ctx, keyURI, crypto.SHA256)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize HashiVault KMS: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported KMS type: %s", config.Type)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get KMS signer: %w", err)
	}
	if sv == nil {
		return nil, fmt.Errorf("KMS returned nil signer")
	}

	return sv, nil
}

// CreateCertificates creates certificates using the provided KMS and templates.
// It creates 3 certificates (root -> intermediate -> leaf) if intermediateKeyID is provided,
// otherwise creates just 2 certs (root -> leaf).
func CreateCertificates(sv signature.SignerVerifier, config KMSConfig,
	rootTemplatePath, leafTemplatePath string,
	rootCertPath, leafCertPath string,
	intermediateKeyID, intermediateTemplatePath, intermediateCertPath string) error {

	// Create root cert
	rootTmpl, err := ParseTemplate(rootTemplatePath, nil)
	if err != nil {
		return fmt.Errorf("error parsing root template: %w", err)
	}

	// Get public key from signer
	rootPubKey, err := sv.PublicKey()
	if err != nil {
		return fmt.Errorf("error getting root public key: %w", err)
	}

	rootCert, err := x509util.CreateCertificate(rootTmpl, rootTmpl, rootPubKey, signerWrapper{sv})
	if err != nil {
		return fmt.Errorf("error creating root certificate: %w", err)
	}

	if err := WriteCertificateToFile(rootCert, rootCertPath); err != nil {
		return fmt.Errorf("error writing root certificate: %w", err)
	}

	var signingCert *x509.Certificate
	var signingKey crypto.Signer

	if intermediateKeyID != "" {
		// Create intermediate cert if key ID is provided
		intermediateTmpl, err := ParseTemplate(intermediateTemplatePath, rootCert)
		if err != nil {
			return fmt.Errorf("error parsing intermediate template: %w", err)
		}

		// Initialize new KMS for intermediate key
		intermediateConfig := config
		intermediateConfig.RootKeyID = intermediateKeyID
		intermediateSV, err := InitKMS(context.Background(), intermediateConfig)
		if err != nil {
			return fmt.Errorf("error initializing intermediate KMS: %w", err)
		}

		intermediatePubKey, err := intermediateSV.PublicKey()
		if err != nil {
			return fmt.Errorf("error getting intermediate public key: %w", err)
		}

		intermediateCert, err := x509util.CreateCertificate(intermediateTmpl, rootCert, intermediatePubKey, signerWrapper{sv})
		if err != nil {
			return fmt.Errorf("error creating intermediate certificate: %w", err)
		}

		if err := WriteCertificateToFile(intermediateCert, intermediateCertPath); err != nil {
			return fmt.Errorf("error writing intermediate certificate: %w", err)
		}

		signingCert = intermediateCert
		signingKey = signerWrapper{intermediateSV}
	} else {
		signingCert = rootCert
		signingKey = signerWrapper{sv}
	}

	// Create leaf cert
	leafTmpl, err := ParseTemplate(leafTemplatePath, signingCert)
	if err != nil {
		return fmt.Errorf("error parsing leaf template: %w", err)
	}

	// Initialize new KMS for leaf key
	leafConfig := config
	leafConfig.RootKeyID = config.LeafKeyID
	leafSV, err := InitKMS(context.Background(), leafConfig)
	if err != nil {
		return fmt.Errorf("error initializing leaf KMS: %w", err)
	}

	leafPubKey, err := leafSV.PublicKey()
	if err != nil {
		return fmt.Errorf("error getting leaf public key: %w", err)
	}

	leafCert, err := x509util.CreateCertificate(leafTmpl, signingCert, leafPubKey, signingKey)
	if err != nil {
		return fmt.Errorf("error creating leaf certificate: %w", err)
	}

	if err := WriteCertificateToFile(leafCert, leafCertPath); err != nil {
		return fmt.Errorf("error writing leaf certificate: %w", err)
	}

	return nil
}

// WriteCertificateToFile writes an X.509 certificate to a PEM-encoded file
func WriteCertificateToFile(cert *x509.Certificate, filename string) error {
	certPEM := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filename, err)
	}
	defer file.Close()

	if err := pem.Encode(file, certPEM); err != nil {
		return fmt.Errorf("failed to write certificate to file %s: %w", filename, err)
	}

	// Determine cert type
	certType := "root"
	if !cert.IsCA {
		certType = "leaf"
	} else if cert.MaxPathLen == 0 {
		certType = "intermediate"
	}

	fmt.Printf("Your %s certificate has been saved in %s.\n", certType, filename)
	return nil
}

// ValidateKMSConfig ensures all required KMS configuration parameters are present
func ValidateKMSConfig(config KMSConfig) error {
	if config.Type == "" {
		return fmt.Errorf("KMS type cannot be empty")
	}
	if config.RootKeyID == "" && config.LeafKeyID == "" {
		return fmt.Errorf("at least one of RootKeyID or LeafKeyID must be specified")
	}

	switch config.Type {
	case "awskms":
		// AWS KMS validation
		if config.Region == "" {
			return fmt.Errorf("region is required for AWS KMS")
		}
		validateAWSKeyID := func(keyID, keyType string) error {
			if keyID == "" {
				return nil
			}
			switch {
			case strings.HasPrefix(keyID, "arn:aws:kms:"):
				parts := strings.Split(keyID, ":")
				if len(parts) < 6 {
					return fmt.Errorf("invalid AWS KMS ARN format for %s", keyType)
				}
				if parts[3] != config.Region {
					return fmt.Errorf("region in ARN (%s) does not match configured region (%s)", parts[3], config.Region)
				}
			case strings.HasPrefix(keyID, "alias/"):
				if strings.TrimPrefix(keyID, "alias/") == "" {
					return fmt.Errorf("alias name cannot be empty for %s", keyType)
				}
			default:
				return fmt.Errorf("awskms %s must start with 'arn:aws:kms:' or 'alias/'", keyType)
			}
			return nil
		}
		if err := validateAWSKeyID(config.RootKeyID, "RootKeyID"); err != nil {
			return err
		}
		if err := validateAWSKeyID(config.IntermediateKeyID, "IntermediateKeyID"); err != nil {
			return err
		}
		if err := validateAWSKeyID(config.LeafKeyID, "LeafKeyID"); err != nil {
			return err
		}

	case "gcpkms":
		// GCP KMS validation
		validateGCPKeyID := func(keyID, keyType string) error {
			if keyID == "" {
				return nil
			}
			requiredComponents := []struct {
				component string
				message   string
			}{
				{"projects/", "must start with 'projects/'"},
				{"/locations/", "must contain '/locations/'"},
				{"/keyRings/", "must contain '/keyRings/'"},
				{"/cryptoKeys/", "must contain '/cryptoKeys/'"},
				{"/cryptoKeyVersions/", "must contain '/cryptoKeyVersions/'"},
			}
			for _, req := range requiredComponents {
				if !strings.Contains(keyID, req.component) {
					return fmt.Errorf("gcpkms %s %s", keyType, req.message)
				}
			}
			return nil
		}
		if err := validateGCPKeyID(config.RootKeyID, "RootKeyID"); err != nil {
			return err
		}
		if err := validateGCPKeyID(config.IntermediateKeyID, "IntermediateKeyID"); err != nil {
			return err
		}
		if err := validateGCPKeyID(config.LeafKeyID, "LeafKeyID"); err != nil {
			return err
		}

	case "azurekms":
		// Azure KMS validation
		if config.Options == nil {
			return fmt.Errorf("options map is required for Azure KMS")
		}
		if config.Options["tenant-id"] == "" {
			return fmt.Errorf("tenant-id is required for Azure KMS")
		}
		validateAzureKeyID := func(keyID, keyType string) error {
			if keyID == "" {
				return nil
			}
			if !strings.HasPrefix(keyID, "azurekms:name=") {
				return fmt.Errorf("azurekms %s must start with 'azurekms:name='", keyType)
			}
			nameStart := strings.Index(keyID, "name=") + 5
			vaultIndex := strings.Index(keyID, ";vault=")
			if vaultIndex == -1 {
				return fmt.Errorf("azurekms %s must contain ';vault=' parameter", keyType)
			}
			if strings.TrimSpace(keyID[nameStart:vaultIndex]) == "" {
				return fmt.Errorf("key name cannot be empty for %s", keyType)
			}
			if strings.TrimSpace(keyID[vaultIndex+7:]) == "" {
				return fmt.Errorf("vault name cannot be empty for %s", keyType)
			}
			return nil
		}
		if err := validateAzureKeyID(config.RootKeyID, "RootKeyID"); err != nil {
			return err
		}
		if err := validateAzureKeyID(config.IntermediateKeyID, "IntermediateKeyID"); err != nil {
			return err
		}
		if err := validateAzureKeyID(config.LeafKeyID, "LeafKeyID"); err != nil {
			return err
		}

	case "hashivault":
		// HashiVault KMS validation
		if config.Options == nil {
			return fmt.Errorf("options map is required for HashiVault KMS")
		}
		if config.Options["address"] == "" {
			return fmt.Errorf("address is required for HashiVault KMS")
		}
		if config.Options["token"] == "" {
			return fmt.Errorf("token is required for HashiVault KMS")
		}
		validateHashiVaultKeyID := func(keyID, keyType string) error {
			if keyID == "" {
				return nil
			}
			parts := strings.Split(keyID, "/")
			if len(parts) < 3 {
				return fmt.Errorf("hashivault %s must be in format: transit/keys/keyname", keyType)
			}
			if parts[0] != "transit" || parts[1] != "keys" {
				return fmt.Errorf("hashivault %s must start with 'transit/keys/'", keyType)
			}
			if parts[2] == "" {
				return fmt.Errorf("key name cannot be empty for %s", keyType)
			}
			return nil
		}
		if err := validateHashiVaultKeyID(config.RootKeyID, "RootKeyID"); err != nil {
			return err
		}
		if err := validateHashiVaultKeyID(config.IntermediateKeyID, "IntermediateKeyID"); err != nil {
			return err
		}
		if err := validateHashiVaultKeyID(config.LeafKeyID, "LeafKeyID"); err != nil {
			return err
		}

	default:
		return fmt.Errorf("unsupported KMS type: %s", config.Type)
	}

	return nil
}

// ValidateTemplatePath checks if the template file exists, has a .json extension,
// and contains valid JSON content.
func ValidateTemplatePath(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("template not found at %s: %w", path, err)
	}
	if !strings.HasSuffix(path, ".json") {
		return fmt.Errorf("template file must have .json extension: %s", path)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading template file: %w", err)
	}
	var js json.RawMessage
	if err := json.Unmarshal(content, &js); err != nil {
		return fmt.Errorf("invalid JSON in template file: %w", err)
	}

	return nil
}