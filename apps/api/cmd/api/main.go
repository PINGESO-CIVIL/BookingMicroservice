// mock-api serves demonstration fixtures only. It is not the production reservation service.
package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bookingms/internal/api"
	"bookingms/internal/service"
	"bookingms/internal/store"
)

func main() {
	address := os.Getenv("MOCK_ADDR")
	if address == "" {
		address = "127.0.0.1:4010"
	}
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		_, port, err := net.SplitHostPort(address)
		if err != nil {
			os.Exit(1)
		}
		client := &http.Client{Timeout: 3 * time.Second}
		response, err := client.Get("http://127.0.0.1:" + port + "/readyz")
		if err != nil {
			os.Exit(1)
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusOK {
			os.Exit(1)
		}
		return
	}
	fixtures, err := store.NewFixtures()
	if err != nil {
		log.Fatal(err)
	}
	database, err := store.OpenPostgres(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Print("Configuración inválida de PostgreSQL")
	}
	defer database.Close()
	docsDir := os.Getenv("API_DOCS_DIR")
	if docsDir == "" {
		docsDir = "/docs"
	}
	handler := api.New(service.New(fixtures, database))
	server := &http.Server{Addr: address, Handler: routes(handler, os.Getenv("API_ENV"), docsDir), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdown)
	}()
	log.Printf("BookingMS MOCK en http://%s — datos prefijados, sin persistencia ni autenticación real", address)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
