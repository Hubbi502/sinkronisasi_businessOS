package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

// VerifySignature returns a Gin middleware that validates HMAC-SHA256 signatures.
// The expected signature is read from the X-Signature header.
func VerifySignature(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		signature := c.GetHeader("X-Signature")
		if signature == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing X-Signature header",
			})
			return
		}

		// Read body and replace it so downstream handlers can read it again
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": "failed to read request body",
			})
			return
		}
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

		// Compute HMAC-SHA256
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedSignature := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "invalid signature",
			})
			return
		}

		c.Next()
	}
}
