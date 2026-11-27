package main

import (
	"strings"
	"testing"
)

func TestConfiguration(t *testing.T) {
	base := configuration{
		env: "prod", databaseURL: "postgres://agenda@postgres:5432/agenda",
		frontendURL: "https://agenda.example", jwksURL: "https://identity.example/jwks",
		issuer: "https://identity.example", audience: "agenda", cookieName: "test-session",
	}
	if err := base.validate("8080", "America/Santiago"); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, port, zone string
		change           func(*configuration)
	}{
		{"API_PORT", "", "America/Santiago", func(*configuration) {}},
		{"TZ", "8080", "", func(*configuration) {}},
		{"API_ENV", "8080", "America/Santiago", func(c *configuration) { c.env = "" }},
		{"DATABASE_URL", "8080", "America/Santiago", func(c *configuration) { c.databaseURL = "postgres://replace-with-secret@postgres/agenda" }},
		{"FRONTEND_URL", "8080", "America/Santiago", func(c *configuration) { c.frontendURL = "https://agenda.example/path" }},
		{"AUTH_JWKS_URL", "8080", "America/Santiago", func(c *configuration) { c.jwksURL = "" }},
		{"AUTH_COOKIE_NAME", "8080", "America/Santiago", func(c *configuration) { c.cookieName = "" }},
		{"BOOKING_JWT_ISSUER", "8080", "America/Santiago", func(c *configuration) { c.issuer = "" }},
		{"BOOKING_JWT_AUDIENCE", "8080", "America/Santiago", func(c *configuration) { c.audience = "" }},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := base
			test.change(&c)
			err := c.validate(test.port, test.zone)
			if err == nil || !strings.Contains(err.Error(), test.name) || strings.Contains(err.Error(), "replace-with-secret") {
				t.Fatalf("error inesperado: %v", err)
			}
		})
	}
}
