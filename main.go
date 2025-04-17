package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/abstract-foundation/zksync-external-node-sidecar/clients"
	"github.com/abstract-foundation/zksync-external-node-sidecar/config"
	"github.com/gorilla/mux"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.Use(loggingMiddleware)
	externalNode := clients.NewZksyncExternalNodeClient()
	router.HandleFunc("/en/readiness", externalNode.HealthCheck).Methods(http.MethodGet)
	router.HandleFunc("/en/liveness", externalNode.HealthCheck).Methods(http.MethodGet)

	srv := &http.Server{
		Handler:      router,
		Addr:         cfg.Server.BindAddr,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	fmt.Println("Starting server:", cfg.Server.BindAddr)
	log.Fatal(srv.ListenAndServe())
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf(
			"%s %s",
			r.Method,
			r.RequestURI,
		)
		next.ServeHTTP(w, r)
	})
}
