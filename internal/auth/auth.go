package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/jwtauth/v5"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"golang.org/x/crypto/bcrypt"
)

func GetPasswordHash(password string) (string, error) {
	if strings.TrimSpace(password) == "" {
		return "", fmt.Errorf("empty password")
	}
	bytePassword := []byte(password)
	hashedPassword, err := bcrypt.GenerateFromPassword(bytePassword, bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func CheckPasswordHash(password, hash string) bool {
	byteHash := []byte(hash)
	bytePassword := []byte(password)
	err := bcrypt.CompareHashAndPassword(byteHash, bytePassword)
	return err == nil
}

func GenerateAuthToken(secretKey string) *jwtauth.JWTAuth {
	tokenAuth := jwtauth.New("HS256", []byte(secretKey), nil)
	return tokenAuth
}

func Verifier(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return jwtauth.Verify(ja, AuthorizationTokenFromHeader, jwtauth.TokenFromHeader, jwtauth.TokenFromCookie)
}

func Authenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			token, _, err := jwtauth.FromContext(r.Context())

			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if token == nil || jwt.Validate(token, ja.ValidateOptions()...) != nil {
				http.Error(w, "Invalid jwt token", http.StatusUnauthorized)
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func AuthorizationTokenFromHeader(r *http.Request) string {
	return r.Header.Get("Authorization")
}
