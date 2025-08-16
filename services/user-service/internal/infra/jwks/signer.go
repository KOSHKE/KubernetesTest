package jwks

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"time"

	"user-service/internal/ports/token"

	"github.com/golang-jwt/jwt/v5"
)

// RSAKeySigner issues RS256 tokens and exposes JWKS
type RSAKeySigner struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	keyID      string
	issuer     string
	audience   string
	ttl        time.Duration
}

type Option func(*RSAKeySigner)

func WithIssuer(iss string) Option     { return func(s *RSAKeySigner) { s.issuer = iss } }
func WithAudience(aud string) Option   { return func(s *RSAKeySigner) { s.audience = aud } }
func WithTTL(ttl time.Duration) Option { return func(s *RSAKeySigner) { s.ttl = ttl } }

// NewRSAKeySigner generates a new RSA key pair and configures the signer
func NewRSAKeySigner(opts ...Option) (*RSAKeySigner, error) {
	pk, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	signer := &RSAKeySigner{
		privateKey: pk,
		publicKey:  &pk.PublicKey,
		ttl:        24 * time.Hour,
	}
	for _, o := range opts {
		o(signer)
	}
	signer.keyID = computeKeyID(signer.publicKey)
	return signer, nil
}

// GenerateAccessToken issues a JWT with RS256 and header kid
func (s *RSAKeySigner) GenerateAccessToken(subject token.Subject) (string, error) {
	if s.privateKey == nil {
		return "", errors.New("private key not initialized")
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":     subject.GetID(),
		"user_id": subject.GetID(),
		"email":   subject.GetEmail(),
		"iss":     s.issuer,
		"aud":     s.audience,
		"iat":     now.Unix(),
		"exp":     now.Add(s.ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = s.keyID
	return token.SignedString(s.privateKey)
}

// JWKS returns the JSON Web Key Set for the current public key
func (s *RSAKeySigner) JWKS() ([]byte, error) {
	if s.publicKey == nil {
		return nil, errors.New("public key not initialized")
	}
	jwk := map[string]string{
		"kty": "RSA",
		"alg": "RS256",
		"use": "sig",
		"kid": s.keyID,
		"n":   base64RawURLUInt(s.publicKey.N),
		"e":   base64RawURLUInt(big.NewInt(int64(s.publicKey.E))),
	}
	set := map[string]any{"keys": []any{jwk}}
	return json.Marshal(set)
}

// base64RawURLUInt encodes big integer as base64url without padding
func base64RawURLUInt(n *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(n.Bytes())
}

// computeKeyID builds a kid from sha256 of modulus N and exponent E (then shortened via sha1 for compactness)
func computeKeyID(pub *rsa.PublicKey) string {
	h := sha256.New()
	h.Write(pub.N.Bytes())
	h.Write(big.NewInt(int64(pub.E)).Bytes())
	sum := h.Sum(nil)
	// shorten to 20 bytes (sha1 of sha256) for readability
	h2 := sha1.New()
	h2.Write(sum)
	return base64.RawURLEncoding.EncodeToString(h2.Sum(nil))
}
