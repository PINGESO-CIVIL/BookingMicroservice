package notificaciones

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"time"

	"bookingms/internal/service"

	openapiTypes "github.com/oapi-codegen/runtime/types"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.8.2 -generate types,client -include-tags Notificaciones -package notificaciones -o client.gen.go ../../../../../contracts/externos/notificaciones.yaml

type Sender struct {
	client       *Client
	destinatario openapiTypes.Email
}

func NewSender(baseURL, recipient string) (*Sender, error) {
	endpoint, err := url.Parse(baseURL)
	if err != nil || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") ||
		endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
		return nil, errors.New("NOTIFICACIONES_URL debe ser una URL HTTP interna válida")
	}
	address, err := mail.ParseAddress(recipient)
	if err != nil || address.Address != recipient {
		return nil, errors.New("NOTIFICACIONES_DEMO_DESTINATARIO debe ser un correo válido")
	}
	client, err := NewClient(baseURL, WithHTTPClient(&http.Client{
		Timeout:       5 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
	}))
	if err != nil {
		return nil, err
	}
	return &Sender{client: client, destinatario: openapiTypes.Email(recipient)}, nil
}

func (s *Sender) Confirm(ctx context.Context, confirmation service.Confirmation) error {
	key := openapiTypes.UUID(confirmation.IdempotencyKey)
	response, err := s.client.PostApiV1Notificaciones(ctx, &PostApiV1NotificacionesParams{
		IdempotencyKey: &key,
		XRequestId:     &confirmation.RequestID,
	}, NotificacionRequest{
		Destinatario: s.destinatario,
		Asunto:       "Confirmación de reserva",
		Cuerpo:       fmt.Sprintf("Su reserva para el bloque %s a %s ha sido confirmada.", confirmation.StartAt, confirmation.EndAt),
	}, func(_ context.Context, req *http.Request) error {
		req.Header.Set("Authorization", confirmation.Authorization)
		return nil
	})
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("Notificaciones respondió HTTP %d", response.StatusCode)
	}
	var result NotificacionResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, 64<<10)).Decode(&result); err != nil {
		return fmt.Errorf("respuesta inválida de Notificaciones: %w", err)
	}
	if result.Estado == nil || (*result.Estado != Pendiente && *result.Estado != EnviadoMock && *result.Estado != Enviado) {
		return errors.New("Notificaciones no confirmó la solicitud de envío")
	}
	return nil
}
