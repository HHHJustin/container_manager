package handlers

import (
    "net/http"
    "os"
    "time"

    "github.com/gin-gonic/gin"
    jwt "github.com/golang-jwt/jwt/v5"
)

type loginDTO struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

// Login 回傳 JWT，使用者與密碼預設 admin/admin，可用環境變數覆蓋。
func Login(c *gin.Context) {
    var dto loginDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    user := getenv("AUTH_USER", "admin")
    pass := getenv("AUTH_PASS", "admin")
    if dto.Username != user || dto.Password != pass {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
        return
    }

    secret := os.Getenv("JWT_SECRET")
    if secret == "" { secret = "devsecret" }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub": dto.Username,
        "iat": time.Now().Unix(),
        "exp": time.Now().Add(2 * time.Hour).Unix(),
    })
    s, err := token.SignedString([]byte(secret))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "token sign failed"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"token": s})
}

func getenv(key, def string) string {
    if v := os.Getenv(key); v != "" { return v }
    return def
}


