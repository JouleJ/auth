package helpers

import (
    "github.com/golang-jwt/jwt/v4"
    "net/http"
    "time"
)

const (
    ACCESS  = "access_token"
    REFRESH = "refresh_token"
)

func IssueToken(issuer string, duration time.Duration, secret []byte) (string, error) {
    claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
        Issuer: issuer,
        ExpiresAt: time.Now().Add(duration).Unix(),
    })

    return claims.SignedString(secret)
}

func TokenToCookie(name, token string) *http.Cookie {
    cookie := http.Cookie{}
    cookie.Name = name
    cookie.Value = token
    return &cookie
}
