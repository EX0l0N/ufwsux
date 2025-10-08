package tokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const intervalSec = 300 // 5 min

func GenerateToken(targetHost, targetPort string, t time.Time) string {
	step := t.Unix() / intervalSec
	data := fmt.Sprintf("%s:%s:%d", targetHost, targetPort, step)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func ValidateToken(token, targetHost, targetPort string, t time.Time) bool {
	// accept current and previous interval to allow small clock skew
	for _, offset := range []int64{0, -1} {
		step := (t.Unix()/intervalSec + offset)
		data := fmt.Sprintf("%s:%s:%d", targetHost, targetPort, step)
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(data))
		expected := hex.EncodeToString(h.Sum(nil))
		if hmac.Equal([]byte(expected), []byte(token)) {
			return true
		}
	}
	return false
}
