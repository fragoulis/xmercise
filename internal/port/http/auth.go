package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTConfig contains bearer token validation settings.
type JWTConfig struct {
	Secret   string
	Issuer   string
	Audience string
}

// JWTMiddleware authenticates bearer tokens with HS256.
type JWTMiddleware struct {
	config JWTConfig
}

// NewJWTMiddleware creates JWT authentication middleware.
func NewJWTMiddleware(config JWTConfig) *JWTMiddleware {
	return &JWTMiddleware{
		config: config,
	}
}

// Handler rejects requests without a valid bearer token.
func (m *JWTMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Fields(r.Header.Get("Authorization"))
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || !m.valid(parts[1]) {
			writeUnauthorized(w)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (m *JWTMiddleware) valid(rawToken string) bool {
	claims := &jwt.RegisteredClaims{}
	options := []jwt.ParserOption{
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	}
	if m.config.Issuer != "" {
		options = append(options, jwt.WithIssuer(m.config.Issuer))
	}
	if m.config.Audience != "" {
		options = append(options, jwt.WithAudience(m.config.Audience))
	}

	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (any, error) {
		return []byte(m.config.Secret), nil
	}, options...)
	return err == nil && token.Valid && claims.Subject != "" && claims.IssuedAt != nil
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(Error{
		Status: http.StatusUnauthorized,
		Title:  "Unauthorized",
	})
}
