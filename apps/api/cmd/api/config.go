package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type configuration struct {
	env, databaseURL, frontendURL            string
	jwksURL, issuer, audience, cookieName    string
	notificationsURL, notificationsRecipient string
}

func loadConfig() (configuration, error) {
	read := func(name string) string { return strings.TrimSpace(os.Getenv(name)) }
	c := configuration{
		env: read("API_ENV"), databaseURL: read("DATABASE_URL"), frontendURL: read("FRONTEND_URL"),
		jwksURL: read("AUTH_JWKS_URL"), issuer: read("BOOKING_JWT_ISSUER"),
		audience: read("BOOKING_JWT_AUDIENCE"), cookieName: read("AUTH_COOKIE_NAME"),
		notificationsURL: read("NOTIFICACIONES_URL"), notificationsRecipient: read("NOTIFICACIONES_DEMO_DESTINATARIO"),
	}
	if err := c.validate(read("API_PORT"), read("TZ")); err != nil {
		return configuration{}, err
	}
	if password := read("PGPASSWORD"); password != "" && !configured(password) {
		return configuration{}, fmt.Errorf("PGPASSWORD requiere un valor local")
	}
	return c, nil
}

func configured(value string) bool {
	return value != "" && !strings.Contains(value, "replace-with-")
}

func (c configuration) validate(port, zone string) error {
	if port != "8080" {
		return fmt.Errorf("API_PORT debe ser 8080")
	}
	switch c.env {
	case "dev", "test", "prod":
	default:
		return fmt.Errorf("API_ENV debe ser dev, test o prod")
	}
	dsn, err := url.Parse(c.databaseURL)
	if err != nil || !configured(c.databaseURL) || dsn.Hostname() == "" ||
		(dsn.Scheme != "postgres" && dsn.Scheme != "postgresql") {
		return fmt.Errorf("DATABASE_URL debe ser una URL PostgreSQL válida")
	}
	if _, err := pgx.ParseConfig(c.databaseURL); err != nil {
		return fmt.Errorf("DATABASE_URL debe ser una URL PostgreSQL válida")
	}
	origin, err := url.Parse(c.frontendURL)
	if err != nil || !configured(c.frontendURL) || origin.Hostname() == "" ||
		(origin.Scheme != "https" && (c.env == "prod" || origin.Scheme != "http")) ||
		c.frontendURL != origin.Scheme+"://"+origin.Host {
		return fmt.Errorf("FRONTEND_URL debe ser un origen HTTP válido, HTTPS en prod")
	}
	if zone != "America/Santiago" {
		return fmt.Errorf("TZ debe ser America/Santiago")
	}
	if _, err := time.LoadLocation(zone); err != nil {
		return fmt.Errorf("no se pudo cargar TZ")
	}
	if c.env != "dev" || c.jwksURL != "" || c.issuer != "" || c.audience != "" || c.cookieName != "" {
		jwks, err := url.Parse(c.jwksURL)
		if err != nil || !configured(c.jwksURL) || jwks.Scheme != "https" || jwks.Hostname() == "" || jwks.User != nil || jwks.Fragment != "" {
			return fmt.Errorf("AUTH_JWKS_URL debe ser una URL HTTPS válida")
		}
		if !configured(c.issuer) {
			return fmt.Errorf("BOOKING_JWT_ISSUER es obligatorio con identidad configurada")
		}
		if !configured(c.audience) {
			return fmt.Errorf("BOOKING_JWT_AUDIENCE es obligatorio con identidad configurada")
		}
		if !configured(c.cookieName) || (&http.Cookie{Name: c.cookieName, Value: "validation"}).Valid() != nil {
			return fmt.Errorf("AUTH_COOKIE_NAME debe ser el nombre válido de la cookie de Identidad")
		}
	}
	if c.notificationsURL != "" || c.notificationsRecipient != "" {
		if c.env != "dev" || c.jwksURL != "" || !configured(c.notificationsURL) || !configured(c.notificationsRecipient) {
			return fmt.Errorf("configura NOTIFICACIONES_URL y NOTIFICACIONES_DEMO_DESTINATARIO juntos solo para el mock en dev")
		}
	}
	return nil
}
