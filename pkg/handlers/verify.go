package handlers

import (
    "auth/pkg/config"
    "auth/pkg/helpers"
)

import (
    "fmt"
    "github.com/golang-jwt/jwt/v4"
    "net/http"
    "time"
)

type VerifyHandler struct {
    secret []byte
}

func NewVerifyHandler(cfg *config.Config) *VerifyHandler {
    return &VerifyHandler {
        secret: cfg.Secret,
    }
}

func getTokenFromRequest(request *http.Request, name string) string {
    cookie, err := request.Cookie(name)
    if err != nil {
        return ""
    }

    return cookie.Value
}

func verifyToken(tokenString string, secret []byte) (string, bool) {
    claims := jwt.StandardClaims{}
    token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) { return secret, nil })
    if err != nil {
        return "", false
    }

    ok := token.Valid && (claims.Valid() == nil)
    login := claims.Issuer
    return login, ok
}


func (this *VerifyHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
    if request.Method != "POST" {
        responseWriter.WriteHeader(http.StatusNotFound)
        return
    }

    access_token := getTokenFromRequest(request, helpers.ACCESS)
    refresh_token := getTokenFromRequest(request, helpers.REFRESH)

    if login, ok := verifyToken(access_token, this.secret); ok {
        fmt.Fprintf(responseWriter, "%s\n", login)
        return
    }

    if login, ok := verifyToken(refresh_token, this.secret); ok {
        access_token, err := helpers.IssueToken(login, time.Minute, this.secret)
        if err != nil {
            responseWriter.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintf(responseWriter, "Failed to re-issue access_token: %v\n", err)
            return
        }

        refresh_token, err := helpers.IssueToken(login, time.Hour, this.secret)
        if err != nil {
            responseWriter.WriteHeader(http.StatusInternalServerError)
            fmt.Fprintf(responseWriter, "Failed to re-issue refresh_token: %v\n", err)
            return
        }

        http.SetCookie(responseWriter, helpers.TokenToCookie(helpers.ACCESS, access_token))
        http.SetCookie(responseWriter, helpers.TokenToCookie(helpers.REFRESH, refresh_token))
        fmt.Fprintf(responseWriter, "%s\n", login)
        return
    }

    responseWriter.WriteHeader(http.StatusForbidden)
    fmt.Fprintf(responseWriter, "Invalid tokens\n")
}
