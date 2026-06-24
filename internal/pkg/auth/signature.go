package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"

	"google.golang.org/grpc/metadata"
)

const MetadataSignatureKey = "signature"

// CanonicalPayload builds the deterministic string signed by inter-service callers.
func CanonicalPayload(senderID, senderType, requestID, timestamp string) string {
	return senderID + "|" + senderType + "|" + requestID + "|" + timestamp
}

// Sign computes an HMAC-SHA256 hex digest for the given payload and secret.
func Sign(secret, payload string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify checks the signature using constant-time comparison.
func Verify(secret, payload, signature string) bool {
	if secret == "" || payload == "" || signature == "" {
		return false
	}
	expected := Sign(secret, payload)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}

// AttachSignature signs canonical metadata fields and adds the signature header.
func AttachSignature(md metadata.MD, secret string) metadata.MD {
	payload := CanonicalPayload(
		metadataValue(md, "sender-id"),
		metadataValue(md, "sender-type"),
		metadataValue(md, "request-id"),
		metadataValue(md, "timestamp"),
	)
	md.Set(MetadataSignatureKey, Sign(secret, payload))
	return md
}

func metadataValue(md metadata.MD, key string) string {
	if vals := md.Get(key); len(vals) > 0 {
		return vals[0]
	}
	return ""
}
