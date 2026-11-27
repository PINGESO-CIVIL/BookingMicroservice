// mock-api serves demonstration fixtures only. It is not the production reservation service.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	_ "time/tzdata"

	"bookingms/internal/api"
	"bookingms/internal/service"
	"bookingms/internal/store"
	"bookingms/internal/store/notificaciones"
)

const serviceName = "bookingms"

var version = "dev"

func main() {
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://127.0.0.1:8080/readyz")
		if err != nil {
			os.Exit(1)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	config, err := loadConfig()
	if err != nil {
		logger.Error("configuración inválida", "servicio", serviceName, "version", version, "error", err.Error())
		os.Exit(1)
	}
	fixtures, err := store.NewFixtures()
	if err != nil {
		logger.Error("no se pudieron cargar los fixtures", "servicio", serviceName, "version", version, "error", err.Error())
		os.Exit(1)
	}
	database, err := store.OpenPostgres(config.databaseURL)
	if err != nil {
		logger.Error("DATABASE_URL no permite configurar PostgreSQL", "servicio", serviceName, "version", version)
		os.Exit(1)
	}
	defer database.Close()
	docsDir := os.Getenv("API_DOCS_DIR")
	if docsDir == "" {
		docsDir = "/docs"
	}
	svc := service.New(fixtures, database)
	if config.notificationsURL != "" {
		sender, err := notificaciones.NewSender(config.notificationsURL, config.notificationsRecipient)
		if err != nil {
			logger.Error("configuración inválida de Notificaciones", "servicio", serviceName, "version", version, "error", err.Error())
			os.Exit(1)
		}
		svc.SetNotifier(sender)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	var authorizer api.Authorizer
	if config.jwksURL != "" {
		authorizer, err = api.NewJWTAuthorizer(ctx, config.issuer, config.audience, config.jwksURL)
		if err != nil {
			logger.Error("AUTH_JWKS_URL no permite inicializar las llaves de Identidad", "servicio", serviceName, "version", version)
			os.Exit(1)
		}
	}
	handler := api.NewWithAuthorizer(svc, authorizer, config.frontendURL, config.cookieName)
	server := &http.Server{
		Addr:              ":8080",
		Handler:           routes(handler, config.env, docsDir, logger),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	logger.Info("servicio iniciado", "servicio", serviceName, "version", version, "entorno", config.env, "direccion", ":8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("error al detener el servidor HTTP", "servicio", serviceName, "version", version, "error", err.Error())
		os.Exit(1)
	}
}
