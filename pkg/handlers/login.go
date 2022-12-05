package handlers

import (
    "auth/pkg/config"
    "auth/pkg/helpers"
)

import (
    "bytes"
    "crypto/sha1"
    "fmt"
    "golang.org/x/crypto/pbkdf2"
    "net/http"
    "time"
)

type LoginHandler struct {
    salt []byte
    secret []byte
    loginToKey map[string][]byte
}

const (
    KEY_ITER_COUNT = 4096
    KEY_SIZE = 32
)

func NewLoginHandler(cfg *config.Config) *LoginHandler {
    return &LoginHandler {
        salt: cfg.Salt,
        secret: cfg.Secret,
        loginToKey: cfg.LoginToKey,
    }
}

func (this *LoginHandler) ServeHTTP(responseWriter http.ResponseWriter, request *http.Request) {
    if request.Method != "POST" {
        responseWriter.WriteHeader(http.StatusNotFound)
        return
    }

    login, password, ok := request.BasicAuth()
    if !ok {
        responseWriter.WriteHeader(http.StatusBadRequest)
        fmt.Fprintf(responseWriter, "No basic auth headers\n")
        return
    }

    expectedKey, ok := this.loginToKey[login]
    if !ok {
        responseWriter.WriteHeader(http.StatusForbidden)
        fmt.Fprintf(responseWriter, "No such login\n")
        return
    }

    key := pbkdf2.Key([]byte(password), this.salt, KEY_ITER_COUNT, KEY_SIZE, sha1.New)

    if !bytes.Equal(key, expectedKey) {
        responseWriter.WriteHeader(http.StatusForbidden)
        fmt.Fprintf(responseWriter, "Wrong password\n")
        return
    }

    access_token, err := helpers.IssueToken(login, time.Minute, this.secret)
    if err != nil {
        responseWriter.WriteHeader(http.StatusInternalServerError)
        fmt.Printf("Failed to issue access token: %v\n", err)
        return
    }

    refresh_token, err := helpers.IssueToken(login, time.Hour, this.secret)
    if err != nil {
        responseWriter.WriteHeader(http.StatusInternalServerError)
        fmt.Printf("Failed to issue refresh token: %v\n", err)
        return
    }

    http.SetCookie(responseWriter, helpers.TokenToCookie(helpers.ACCESS, access_token))
    http.SetCookie(responseWriter, helpers.TokenToCookie(helpers.REFRESH, refresh_token))
}

