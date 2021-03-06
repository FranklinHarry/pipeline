package secret

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"

	secretTypes "github.com/banzaicloud/pipeline/pkg/secret"
	"golang.org/x/crypto/ssh"
)

// SSHKeyPair struct to store SSH key data
type SSHKeyPair struct {
	User                 string `json:"user,omitempty"`
	Identifier           string `json:"identifier,omitempty"`
	PublicKeyData        string `json:"publicKeyData,omitempty"`
	PublicKeyFingerprint string `json:"publicKeyFingerprint,omitempty"`
	PrivateKeyData       string `json:"PrivateKeyData,omitempty"`
}

// NewSSHKeyPair constructs a SSH Key from the values stored
// in the given secret
func NewSSHKeyPair(s *SecretItemResponse) *SSHKeyPair {
	return &SSHKeyPair{
		User:                 s.Values[secretTypes.User],
		Identifier:           s.Values[secretTypes.Identifier],
		PublicKeyData:        s.Values[secretTypes.PublicKeyData],
		PublicKeyFingerprint: s.Values[secretTypes.PublicKeyFingerprint],
		PrivateKeyData:       s.Values[secretTypes.PrivateKeyData],
	}
}

// StoreSSHKeyPair to store SSH Key to Bank Vaults
func StoreSSHKeyPair(key *SSHKeyPair, organizationID uint, clusterID uint, clusterName string) (secretID string, err error) {
	log.Info("Store SSH Key to Bank Vaults")
	var createSecretRequest CreateSecretRequest
	createSecretRequest.Type = secretTypes.SSHSecretType
	createSecretRequest.Name = fmt.Sprint("ssh-cluster-", clusterID)
	createSecretRequest.Tags = []string{"cluster:" + clusterName}

	createSecretRequest.Values = map[string]string{
		secretTypes.User:                 key.User,
		secretTypes.Identifier:           key.Identifier,
		secretTypes.PublicKeyData:        key.PublicKeyData,
		secretTypes.PublicKeyFingerprint: key.PublicKeyFingerprint,
		secretTypes.PrivateKeyData:       key.PrivateKeyData,
	}

	secretID, err = Store.Store(organizationID, &createSecretRequest)

	if err != nil {
		log.Errorf("Error during store: %s", err.Error())
		return "", err
	}

	log.Info("SSH Key stored.")
	return
}

// GenerateSSHKeyPair for Generate new SSH Key pair
func GenerateSSHKeyPair() (*SSHKeyPair, error) {
	log.Info("Generate new SSH key")

	key := new(SSHKeyPair)

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Errorf("PrivateKey generator failed reason: %s", err.Error())
		return key, err
	}

	privateKeyPEM := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)}
	keyBuff := new(bytes.Buffer)
	if err := pem.Encode(keyBuff, privateKeyPEM); err != nil {
		log.Errorf("PrivateKey generator failed reason: %s", err.Error())
		return key, err
	}
	key.PrivateKeyData = keyBuff.String()
	log.Debug("Private key generated.")

	pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
	if err != nil {
		log.Errorf("PublicKey generator failed reason: %s", err.Error())
		return key, err
	}
	log.Debug("Public key generated.")

	key.PublicKeyData = fmt.Sprintf("%s %s \n", strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(pub)), "\n"), "no-reply@banzaicloud.com")

	key.PublicKeyFingerprint = ssh.FingerprintSHA256(pub)
	log.Info("SSH key generated.")

	return key, nil
}
