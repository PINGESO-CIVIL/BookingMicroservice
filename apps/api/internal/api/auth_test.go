package api

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bookingms/internal/service"
	"bookingms/internal/store"

	"github.com/golang-jwt/jwt/v5"
)

func TestSessionCookie(t *testing.T) {
	const origin = "https://agenda.example"
	const cookieName = "test-session"
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	fixtures, err := store.NewFixtures()
	if err != nil {
		t.Fatal(err)
	}
	database, err := store.OpenPostgres("")
	if err != nil {
		t.Fatal(err)
	}
	authorize := jwtAuthorizer("identity", "agenda", func(*jwt.Token) (any, error) { return &privateKey.PublicKey, nil })
	handler := NewWithAuthorizer(service.New(fixtures, database), authorize, origin, cookieName)
	sign := func(permission string, lifetime time.Duration) string {
		t.Helper()
		claims := agendaClaims{RegisteredClaims: jwt.RegisteredClaims{
			Issuer: "identity", Audience: jwt.ClaimStrings{"agenda"}, Subject: "user-001",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(lifetime)),
		}, Permissions: []string{permission}}
		token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(privateKey)
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	valid := sign("reservas:leer", time.Minute)
	for _, tc := range []struct {
		name, method, origin, cookie, bearer string
		status                               int
	}{
		{"válida", http.MethodGet, origin, valid, "", 200},
		{"ausente", http.MethodGet, origin, "", "", 401},
		{"inválida", http.MethodGet, origin, "invalid", "", 401},
		{"vencida", http.MethodGet, origin, sign("reservas:leer", -time.Minute), "", 401},
		{"sin permiso", http.MethodGet, origin, sign("recursos:leer", time.Minute), "", 403},
		{"origen ajeno", http.MethodPost, "https://otro.example", valid, "", 403},
		{"bearer de demostración", http.MethodGet, origin, "", "demo-admin", 401},
		{"bearer entre servicios", http.MethodGet, "", "", valid, 200},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, "/api/v1/reservas", nil)
			request.Header.Set("Origin", tc.origin)
			request.Header.Set("X-Mock-Role", "admin")
			if tc.cookie != "" {
				request.AddCookie(&http.Cookie{Name: cookieName, Value: tc.cookie})
			}
			if tc.bearer != "" {
				request.Header.Set("Authorization", "Bearer "+tc.bearer)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tc.status {
				t.Fatalf("HTTP %d, se esperaba %d: %s", response.Code, tc.status, response.Body)
			}
			if allowed := response.Header().Get("Access-Control-Allow-Origin"); (allowed == origin) != (tc.origin == origin) {
				t.Fatalf("origen permitido inesperado: %q", allowed)
			}
		})
	}
}
