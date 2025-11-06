package middleware

import (
	"net/http"
	"os"
	"strings"
    "time"

	"github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v5"
)

// Auth 驗證 Authorization: Bearer <token>，預設 token 為 devtoken（可由環境變數 AUTH_TOKEN 覆蓋）。
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
        header := c.GetHeader("Authorization")
        parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing bearer token"})
			return
		}

        // 僅支援 JWT
        secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
        if secret == "" { secret = "devsecret" }
        token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, jwt.ErrTokenSignatureInvalid
            }
            return []byte(secret), nil
        }, jwt.WithLeeway(1*time.Minute))
        if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}
		c.Next()
	}
}

func getenvDefault(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
