package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"simpleapi/internal/api/middlewares"
)

func main() {
	port := ":3000"

	cert := "cert.pem"
	key := "key.pem"

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
		Addr:      port,
		Handler:   middlewares.SecurityHeaders(mux),
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port:", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}
