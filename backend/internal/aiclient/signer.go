package aiclient

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
)

// canonical builds the exact string the Python service rebuilds and checks:
// "v1\n" + timestamp + "\n" + METHOD + "\n" + path_with_raw_query + "\n" + actor + "\n" + hex(SHA-256(body))
func canonical(timestamp int64, method, pathWithQuery, actor string, body []byte) string {
	sum := sha256.Sum256(body)
	return strings.Join([]string{
		"v1",
		strconv.FormatInt(timestamp, 10),
		method,
		pathWithQuery,
		actor,
		hex.EncodeToString(sum[:]),
	}, "\n")
}

// sign returns hex(HMAC-SHA256(secret, canonical)).
func sign(secret string, timestamp int64, method, pathWithQuery, actor string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(canonical(timestamp, method, pathWithQuery, actor, body)))
	return hex.EncodeToString(mac.Sum(nil))
}
