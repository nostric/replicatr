package nip04

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/Hubmakerlabs/replicatr/pkg/ec"
	"github.com/Hubmakerlabs/replicatr/pkg/hex"
	log2 "github.com/Hubmakerlabs/replicatr/pkg/log"
)

var log = log2.GetStd()

// ComputeSharedSecret returns a shared secret key used to encrypt messages.
// The private and public keys should be hex encoded.
// Uses the Diffie-Hellman key exchange (ECDH) (RFC 4753).
func ComputeSharedSecret(pub string, sec string) (sharedSecret []byte, e error) {
	var secBytes []byte
	secBytes, e = hex.Dec(sec)
	if e != nil {
		return nil, fmt.Errorf("error decoding sender private key: %w", e)
	}
	secKey, _ := btcec.PrivKeyFromBytes(secBytes)
	// adding 02 to signal that this is a compressed public key (33 bytes)
	var pubBytes []byte
	if pubBytes, e = hex.Dec("02" + pub); log.Fail(e) {
		return nil, fmt.Errorf("error decoding hex string of receiver public key '%s': %w", "02"+pub, e)
	}
	var pubKey *btcec.PublicKey
	if pubKey, e = btcec.ParsePubKey(pubBytes); log.Fail(e) {
		return nil, fmt.Errorf("error parsing receiver public key '%s': %w", "02"+pub, e)
	}
	return btcec.GenerateSharedSecret(secKey, pubKey), nil
}

// Encrypt encrypts message with key using aes-256-cbc.
// key should be the shared secret generated by ComputeSharedSecret.
// Returns: base64(encrypted_bytes) + "?iv=" + base64(initialization_vector).
func Encrypt(message string, key []byte) (string, error) {
	// block size is 16 bytes
	iv := make([]byte, 16)
	// can probably use a less expensive lib since IV has to only be unique; not perfectly random; math/rand?
	if _, e := rand.Read(iv); e != nil {
		return "", fmt.Errorf("error creating initization vector: %w", e)
	}

	// automatically picks aes-256 based on key length (32 bytes)
	block, e := aes.NewCipher(key)
	if e != nil {
		return "", fmt.Errorf("error creating block cipher: %w", e)
	}
	mode := cipher.NewCBCEncrypter(block, iv)

	plaintext := []byte(message)

	// add padding
	base := len(plaintext)

	// this will be a number between 1 and 16 (including), never 0
	padding := block.BlockSize() - base%block.BlockSize()

	// encode the padding in all the padding bytes themselves
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)

	paddedMsgBytes := append(plaintext, padtext...)

	ciphertext := make([]byte, len(paddedMsgBytes))
	mode.CryptBlocks(ciphertext, paddedMsgBytes)

	return base64.StdEncoding.EncodeToString(ciphertext) + "?iv=" + base64.StdEncoding.EncodeToString(iv), nil
}

// Decrypt decrypts a content string using the shared secret key.
// The inverse operation to message -> Encrypt(message, key).
func Decrypt(content string, key []byte) (string, error) {
	parts := strings.Split(content, "?iv=")
	if len(parts) < 2 {
		return "", fmt.Errorf("error parsing encrypted message: no initialization vector")
	}

	ciphertext, e := base64.StdEncoding.DecodeString(parts[0])
	if e != nil {
		return "", fmt.Errorf("error decoding ciphertext from base64: %w", e)
	}

	iv, e := base64.StdEncoding.DecodeString(parts[1])
	if e != nil {
		return "", fmt.Errorf("error decoding iv from base64: %w", e)
	}

	block, e := aes.NewCipher(key)
	if e != nil {
		return "", fmt.Errorf("error creating block cipher: %w", e)
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	// remove padding
	var (
		message      = string(plaintext)
		plaintextLen = len(plaintext)
	)
	if plaintextLen > 0 {
		padding := int(plaintext[plaintextLen-1]) // the padding amount is encoded in the padding bytes themselves
		if padding > plaintextLen {
			return "", fmt.Errorf("invalid padding amount: %d", padding)
		}
		message = string(plaintext[0 : plaintextLen-padding])
	}

	return message, nil
}
