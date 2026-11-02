package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalid = errors.New("invalid or tampered value")
	ErrExpired = errors.New("expired")
)

// Codec seals values into cookies with AES-256-GCM, so the browser carries the session and no replica needs a
// store. Sessions are revoked only by expiry.
type Codec struct {
	// keys[0] seals; every key opens, so a previous key keeps sessions valid through a rotation.
	keys []cipher.AEAD
}

// NewCodec takes base64-encoded 32-byte keys; previous may be empty.
func NewCodec(primary, previous string) (*Codec, error) {
	c := &Codec{}
	for i, k := range []string{primary, previous} {
		if k == "" && i > 0 {
			continue
		}
		raw, err := base64.StdEncoding.DecodeString(k)
		if err != nil || len(raw) != 32 {
			return nil, errors.New("must be 32 bytes, base64-encoded (openssl rand -base64 32)")
		}
		if weak(raw) {
			return nil, errors.New("is a placeholder (one repeated byte); generate one with openssl rand -base64 32")
		}
		block, err := aes.NewCipher(raw)
		if err != nil {
			return nil, fmt.Errorf("create cipher: %w", err)
		}
		aead, err := cipher.NewGCM(block)
		if err != nil {
			return nil, fmt.Errorf("create cipher: %w", err)
		}
		c.keys = append(c.keys, aead)
	}
	return c, nil
}

type envelope struct {
	Expires int64           `json:"e"`
	Value   json.RawMessage `json:"v"`
}

// Seal encrypts v until expires. name is authenticated with it, so a value sealed for one cookie is rejected by
// another.
func (c *Codec) Seal(name string, v any, expires time.Time) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}
	plain, err := json.Marshal(envelope{Expires: expires.Unix(), Value: raw})
	if err != nil {
		return "", fmt.Errorf("encode: %w", err)
	}
	aead := c.keys[0]
	nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(plain)+aead.Overhead())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(aead.Seal(nonce, nonce, plain, []byte(name))), nil
}

func (c *Codec) Open(name, sealed string, now time.Time, v any) error {
	data, err := base64.RawURLEncoding.DecodeString(sealed)
	if err != nil {
		return ErrInvalid
	}
	for _, aead := range c.keys {
		if len(data) < aead.NonceSize() {
			return ErrInvalid
		}
		plain, err := aead.Open(nil, data[:aead.NonceSize()], data[aead.NonceSize():], []byte(name))
		if err != nil {
			continue
		}
		var env envelope
		if json.Unmarshal(plain, &env) != nil {
			return ErrInvalid
		}
		if !now.Before(time.Unix(env.Expires, 0)) {
			return ErrExpired
		}
		if json.Unmarshal(env.Value, v) != nil {
			return ErrInvalid
		}
		return nil
	}
	return ErrInvalid
}

// weak catches placeholder keys such as all zeros, which decode to valid 32-byte keys.
func weak(key []byte) bool {
	for _, b := range key[1:] {
		if b != key[0] {
			return false
		}
	}
	return true
}
