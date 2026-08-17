package applemusic

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"
)

// TokenConfig contains the Apple developer credentials used to sign MusicKit
// developer tokens.
type TokenConfig struct {
	TeamID     string
	KeyID      string
	PrivateKey string
	TTL        time.Duration
}

// DeveloperToken signs a short-lived ES256 JWT for Apple Music API requests.
func DeveloperToken(now time.Time, cfg TokenConfig) (string, error) {
	if cfg.TeamID == "" || cfg.KeyID == "" || strings.TrimSpace(cfg.PrivateKey) == "" {
		return "", fmt.Errorf("applemusic: developer token: missing credentials")
	}
	ttl := cfg.TTL
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	key, err := parsePrivateKey(cfg.PrivateKey)
	if err != nil {
		return "", err
	}

	header := map[string]string{
		"alg": "ES256",
		"kid": cfg.KeyID,
		"typ": "JWT",
	}
	claims := map[string]any{
		"iss": cfg.TeamID,
		"iat": now.Unix(),
		"exp": now.Add(ttl).Unix(),
	}

	encodedHeader, err := encodeJSON(header)
	if err != nil {
		return "", fmt.Errorf("applemusic: developer token: header: %w", err)
	}
	encodedClaims, err := encodeJSON(claims)
	if err != nil {
		return "", fmt.Errorf("applemusic: developer token: claims: %w", err)
	}
	signingInput := encodedHeader + "." + encodedClaims
	digest := sha256.Sum256([]byte(signingInput))
	r, s, err := ecdsa.Sign(rand.Reader, key, digest[:])
	if err != nil {
		return "", fmt.Errorf("applemusic: developer token: sign: %w", err)
	}

	sig, err := joseSignature(r, s)
	if err != nil {
		return "", err
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func encodeJSON(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func parsePrivateKey(raw string) (*ecdsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(raw))
	if block == nil {
		return nil, fmt.Errorf("applemusic: developer token: private key is not PEM")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		if ecKey, ecErr := x509.ParseECPrivateKey(block.Bytes); ecErr == nil {
			return ecKey, nil
		}
		return nil, fmt.Errorf("applemusic: developer token: parse private key: %w", err)
	}
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("applemusic: developer token: private key must be ECDSA")
	}
	if ecKey.Curve != elliptic.P256() {
		return nil, fmt.Errorf("applemusic: developer token: private key must use P-256")
	}
	return ecKey, nil
}

func joseSignature(r, s *big.Int) ([]byte, error) {
	type ecdsaSignature struct {
		R, S *big.Int
	}
	if _, err := asn1.Marshal(ecdsaSignature{R: r, S: s}); err != nil {
		return nil, fmt.Errorf("applemusic: developer token: marshal signature: %w", err)
	}
	out := make([]byte, 64)
	r.FillBytes(out[:32])
	s.FillBytes(out[32:])
	return out, nil
}
