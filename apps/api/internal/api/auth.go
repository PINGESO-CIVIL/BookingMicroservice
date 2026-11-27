package api

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"bookingms/internal/service"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

type Authorizer func(authorization, permission string) (string, *service.Fault)

type agendaClaims struct {
	jwt.RegisteredClaims
	Permissions []string `json:"permissions"`
}

// NewJWTAuthorizer validates tokens issued by the configured identity service.
func NewJWTAuthorizer(ctx context.Context, issuer, audience, jwksURL string) (Authorizer, error) {
	endpoint, err := url.Parse(jwksURL)
	if err != nil || endpoint.Host == "" || endpoint.Scheme != "https" || endpoint.User != nil || strings.TrimSpace(issuer) == "" || strings.TrimSpace(audience) == "" {
		return nil, errors.New("configura el emisor, la audiencia y una URL HTTPS de JWKS")
	}
	failMissingKeys := false
	keys, err := keyfunc.NewDefaultOverrideCtx(ctx, []string{jwksURL}, keyfunc.Override{
		HTTPTimeout: 5 * time.Second, RateLimitWaitMax: 5 * time.Second,
		NoErrorReturnFirstHTTPReq: &failMissingKeys,
	})
	if err != nil {
		return nil, err
	}
	parser := jwt.NewParser(jwt.WithIssuer(issuer), jwt.WithAudience(audience), jwt.WithExpirationRequired(),
		jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
	return func(header, permission string) (string, *service.Fault) {
		raw, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || raw == "" {
			return "", &service.Fault{Example: "Unauthenticated"}
		}
		var claims agendaClaims
		token, err := parser.ParseWithClaims(raw, &claims, keys.Keyfunc)
		if err != nil || !token.Valid || claims.Subject == "" {
			return "", &service.Fault{Example: "Unauthenticated"}
		}
		if !slices.Contains(claims.Permissions, permission) {
			return "", &service.Fault{Example: "Forbidden"}
		}
		return claims.Subject, nil
	}, nil
}

func operationPermission(op operation) string {
	if op.name == "listAdminReservations" {
		return "reservas:administrar"
	}
	if op.admin {
		return "recursos:administrar"
	}
	switch op.name {
	case "listResources", "getResourceAvailability":
		return "recursos:leer"
	case "createReservation", "updateReservation", "cancelReservation":
		return "reservas:operar"
	default:
		return "reservas:leer"
	}
}
