package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dustin-decker/security-validator/pkg/handlers"
)

// Run starts the API server
func Run() {
	fmt.Println("webhook starting up...")
	http.HandleFunc("/", handlers.ValidatingWebhook)

	s := &http.Server{
		Addr:           ":443",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	// log.Fatal(s.ListenAndServe())
	log.Fatal(s.ListenAndServeTLS("/certs/cert.pem", "/certs/key.pem"))
}
