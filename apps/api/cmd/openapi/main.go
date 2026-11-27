package main

import (
	"log"
	"os"

	"bookingms/internal/api"
)

func main() {
	spec, err := api.OpenAPIYAML()
	if err != nil {
		log.Fatal(err)
	}
	if len(os.Args) != 2 {
		log.Fatal("uso: openapi <archivo de salida>")
	}
	if err := os.WriteFile(os.Args[1], spec, 0644); err != nil {
		log.Fatal(err)
	}
}
