package applemusic

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDeveloperTokenSignsES256JWT(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	rawKey := string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))

	token, err := DeveloperToken(time.Unix(100, 0), TokenConfig{
		TeamID:     "TEAMID123",
		KeyID:      "KEYID123",
		PrivateKey: rawKey,
		TTL:        time.Hour,
	})

	require.NoError(t, err)
	assert.Len(t, strings.Split(token, "."), 3)
	assert.NotContains(t, token, rawKey)
	assert.NotContains(t, token, "TEAMID123")
}

func TestDeveloperTokenRequiresCredentials(t *testing.T) {
	_, err := DeveloperToken(time.Now(), TokenConfig{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing credentials")
}
