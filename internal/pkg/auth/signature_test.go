package auth_test

import (
	"testing"
	"time"

	. "github.com/nepeta70/ride-hailing/internal/pkg/auth"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestCanonicalPayload(t *testing.T) {
	tests := []struct {
		name       string
		senderID   string
		senderType string
		requestID  string
		timestamp  string
		expected   string
	}{
		{
			name:       "joins fields with pipe delimiter",
			senderID:   "user-1",
			senderType: "rider",
			requestID:  "req-1",
			timestamp:  "2026-01-01T00:00:00Z",
			expected:   "user-1|rider|req-1|2026-01-01T00:00:00Z",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, CanonicalPayload(tt.senderID, tt.senderType, tt.requestID, tt.timestamp))
		})
	}
}

func TestSignAndVerify(t *testing.T) {
	secret := "shared-hmac-secret"
	payload := CanonicalPayload("sender", "rider", "request", time.Now().UTC().Format(time.RFC3339))

	tests := []struct {
		name      string
		secret    string
		payload   string
		signature string
		expected  bool
	}{
		{
			name:      "valid signature",
			secret:    secret,
			payload:   payload,
			signature: Sign(secret, payload),
			expected:  true,
		},
		{
			name:      "invalid signature",
			secret:    secret,
			payload:   payload,
			signature: "deadbeef",
			expected:  false,
		},
		{
			name:      "tampered payload",
			secret:    secret,
			payload:   payload + "-tampered",
			signature: Sign(secret, payload),
			expected:  false,
		},
		{
			name:      "missing signature",
			secret:    secret,
			payload:   payload,
			signature: "",
			expected:  false,
		},
		{
			name:      "missing secret",
			secret:    "",
			payload:   payload,
			signature: Sign(secret, payload),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Verify(tt.secret, tt.payload, tt.signature))
		})
	}
}

func TestAttachSignature(t *testing.T) {
	secret := "shared-hmac-secret"
	md := metadata.New(map[string]string{
		"sender-id":   "11111111-1111-1111-1111-111111111111",
		"sender-type": "matching",
		"request-id":  "22222222-2222-2222-2222-222222222222",
		"timestamp":   "2026-06-24T12:00:00Z",
	})

	signed := AttachSignature(md, secret)
	signature := signed.Get(MetadataSignatureKey)

	assert.Len(t, signature, 1)
	payload := CanonicalPayload(
		"11111111-1111-1111-1111-111111111111",
		"matching",
		"22222222-2222-2222-2222-222222222222",
		"2026-06-24T12:00:00Z",
	)
	assert.True(t, Verify(secret, payload, signature[0]))
}
