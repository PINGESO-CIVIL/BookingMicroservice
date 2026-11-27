package api

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"bookingms/internal/service"
	"bookingms/internal/store"

	"github.com/golang-jwt/jwt/v5"
)

func TestMatrizAcceso(t *testing.T) {
	const origin = "https://agenda.example"
	const cookieName = "agenda-session"

	fixtureJSON, err := os.ReadFile("../store/fixtures.json")
	if err != nil {
		t.Fatal(err)
	}
	var f struct {
		Examples map[string]json.RawMessage `json:"examples"`
	}
	if err := json.Unmarshal(fixtureJSON, &f); err != nil {
		t.Fatal(err)
	}

	rsaKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	signers := []struct {
		name    string
		keyFunc jwt.Keyfunc
		sign    func(perms ...string) string
	}{
		{
			name:    "RS256",
			keyFunc: func(*jwt.Token) (any, error) { return &rsaKey.PublicKey, nil },
			sign: func(perms ...string) string {
				claims := agendaClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    "identity",
						Audience:  jwt.ClaimStrings{"agenda"},
						Subject:   "user-001",
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
					},
					Permissions: perms,
				}
				token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(rsaKey)
				if err != nil {
					t.Fatal(err)
				}
				return token
			},
		},
		{
			name:    "EdDSA",
			keyFunc: func(*jwt.Token) (any, error) { return edPub, nil },
			sign: func(perms ...string) string {
				claims := agendaClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    "identity",
						Audience:  jwt.ClaimStrings{"agenda"},
						Subject:   "user-001",
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
					},
					Permissions: perms,
				}
				token, err := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims).SignedString(edPriv)
				if err != nil {
					t.Fatal(err)
				}
				return token
			},
		},
	}

	endpoints := []struct {
		method     string
		path       string
		example    string
		permission string
		okStatus   int
	}{
		{"GET", "/api/v1/reservas", "", "reservas:leer", 200},
		{"GET", "/api/v1/admin/reservas", "", "reservas:administrar", 200},
		{"POST", "/api/v1/reservas", "CreateRequest", "reservas:operar", 201},
		{"GET", "/api/v1/reservas/00000000-0000-4000-8000-200000000001", "", "reservas:leer", 200},
		{"PATCH", "/api/v1/reservas/00000000-0000-4000-8000-200000000001", "UpdateRequest", "reservas:operar", 200},
		{"POST", "/api/v1/reservas/00000000-0000-4000-8000-200000000001/cancelaciones", "CancelRequest", "reservas:operar", 200},
		{"GET", "/api/v1/recursos", "", "recursos:leer", 200},
		{"POST", "/api/v1/recursos", "CreateResourceRequest", "recursos:administrar", 201},
		{"GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", "", "recursos:administrar", 200},
		{"PATCH", "/api/v1/recursos/00000000-0000-4000-8000-100000000001", "UpdateResourceRequest", "recursos:administrar", 200},
		{"PUT", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/horarios", "UpdateResourceScheduleRequest", "recursos:administrar", 200},
		{"POST", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/bloqueos", "CreateResourceClosureRequest", "recursos:administrar", 201},
		{"DELETE", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/bloqueos/00000000-0000-4000-8000-300000000001", "", "recursos:administrar", 204},
		{"GET", "/api/v1/recursos/00000000-0000-4000-8000-100000000001/disponibilidades?from=2026-11-10T03:00:00Z&to=2026-11-11T03:00:00Z", "", "recursos:leer", 200},
	}

	fixtures, err := store.NewFixtures()
	if err != nil {
		t.Fatal(err)
	}
	database, err := store.OpenPostgres("")
	if err != nil {
		t.Fatal(err)
	}

	for _, signer := range signers {
		t.Run(signer.name, func(t *testing.T) {
			authorize := jwtAuthorizer("identity", "agenda", signer.keyFunc)
			h := NewWithAuthorizer(service.New(fixtures, database), authorize, origin, cookieName)

			for _, ep := range endpoints {
				body := string(f.Examples[ep.example])
				cases := []struct {
					label  string
					cookie string
					want   int
				}{
					{"sin_sesion", "", http.StatusUnauthorized},
					{"permiso_insuficiente", signer.sign("permiso:ajeno"), http.StatusForbidden},
					{"autorizado", signer.sign(ep.permission), ep.okStatus},
				}
				for _, tc := range cases {
					t.Run(ep.method+"_"+ep.path+"_"+tc.label, func(t *testing.T) {
						req := httptest.NewRequest(ep.method, ep.path, strings.NewReader(body))
						req.Header.Set("Origin", origin)
						req.Header.Set("Content-Type", "application/json")
						req.Header.Set("Idempotency-Key", "matriz-acceso-key")
						if tc.cookie != "" {
							req.AddCookie(&http.Cookie{Name: cookieName, Value: tc.cookie})
						}
						rec := httptest.NewRecorder()
						h.ServeHTTP(rec, req)
						if rec.Code != tc.want {
							t.Fatalf("%s %s (%s): HTTP %d, esperado %d: %s", ep.method, ep.path, tc.label, rec.Code, tc.want, rec.Body)
						}
					})
				}
			}
		})
	}
}
