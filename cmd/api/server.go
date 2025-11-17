package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	mw "simpleapi/internal/api/middlewares"
)

type middlewareFunc func(http.Handler) http.Handler

var middlewares = []middlewareFunc{
	mw.Cors,
	mw.SecurityHeaders,
	mw.ResponseTimeMiddleware,
	mw.Compression,
}

func applyMiddleWares(mux http.Handler) http.Handler {
	var middleWare http.Handler = mux
	for _, mwFunc := range middlewares {
		middleWare = mwFunc(middleWare)
	}
	return middleWare
}

func main() {
	port := ":3000"

	cert := "cert/cert.pem"
	key := "cert/key.pem"

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w, "Hello root Route")
		w.Write([]byte("Hello root route"))
	})
	mux.HandleFunc("/teachers", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w, "Hello root Route")
		w.Write([]byte("Hello Teachers route"))
	})
	mux.HandleFunc("/students", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w, "Hello root Route")
		w.Write([]byte("Hello students route"))
	})

	mux.HandleFunc("/execs", func(w http.ResponseWriter, r *http.Request) {
		// fmt.Fprintf(w, "Hello root Route")
		w.Write([]byte("Hello Execs route"))
	})

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}
	server := &http.Server{
		Addr:    port,
		Handler: applyMiddleWares(mux),
		// Handler:   middlewares.Cors(mux),
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port:", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}
