package ca

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"golang.org/x/crypto/ssh"
)

// loadCAKey to sign pubkeys from disk
func loadCAKey(keyfile string) ssh.Signer {
	// Parse CA Key
	caPrivateKey, err := ioutil.ReadFile(keyfile)
	if err != nil {
		log.Fatalln("Unable to read keyfile: " + keyfile)
	}
	caSigner, err := ssh.ParsePrivateKey(caPrivateKey)
	if err != nil {
		log.Fatalln("CA Private Key is invalid: " + keyfile)
	}
	return caSigner
}

// signPubkey with the provided signer
func signPubkey(caSigner ssh.Signer, sshKey string, ips []string, daysValid time.Duration) string {

	// Parse SSH Pubkey
	sshKeyBytes, err := base64.StdEncoding.DecodeString(sshKey)
	if err != nil {
		log.Println("[signPubkey] Attempted to sign SSH key we could not decode")
		return ""
	}
	out, err := ssh.ParsePublicKey([]byte(sshKeyBytes))
	if err != nil {
		log.Println("[signPubkey] Attempted to sign SSH key we could not parse")
		return ""
	}

	c := ssh.Certificate{
		Nonce:           make([]byte, 32),
		Key:             out,
		Serial:          0,
		CertType:        ssh.HostCert,
		KeyId:           "gcp-ssh-ca",
		ValidPrincipals: ips,
		ValidAfter:      uint64(time.Now().Unix()),
		ValidBefore:     uint64(time.Now().Add(daysValid).Unix()),
		SignatureKey:    caSigner.PublicKey(),
	}
	rand.Read(c.Nonce)

	// See the reference implementation here:
	// https://github.com/openssh/openssh-portable/blob/master/ssh-keygen.c#L1794
	c.SignCert(rand.Reader, caSigner)
	// Output cert can be inspected with: ssh-keygen -L -f keyToSign.pub

	return fmt.Sprintf("ssh-ed25519-cert-v01@openssh.com %v", base64.StdEncoding.EncodeToString(c.Marshal()))
}
