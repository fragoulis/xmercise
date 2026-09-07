package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	httpadapter "github.com/fragoulis/xmercise/internal/port/http"
)

func TestJWTMiddleware(t *testing.T) {
	t.Parallel()

	const secret = "test-secret"
	middleware := httpadapter.NewJWTMiddleware(httpadapter.JWTConfig{
		Secret: secret,
	})
	handler := middleware.Handler(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	t.Run("accepts a valid token", func(t *testing.T) {
		t.Parallel()

		token := signedToken(t, secret, jwt.RegisteredClaims{
			Subject:   "user-1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})
		request := httptest.NewRequest(http.MethodGet, "/v1/companies/id", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
		}
	})

	t.Run("rejects a token without a subject", func(t *testing.T) {
		t.Parallel()

		token := signedToken(t, secret, jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		})
		request := httptest.NewRequest(http.MethodGet, "/v1/companies/id", nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()

		handler.ServeHTTP(response, request)

		if response.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", response.Code, http.StatusUnauthorized)
		}
	})
}

func signedToken(t *testing.T, secret string, claims jwt.RegisteredClaims) string {
	t.Helper()

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}
