package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"simpleapi/internal/api/handlers"
	mw "simpleapi/internal/api/middlewares"
	"simpleapi/pkg/utils"
	"time"
)

var rl = mw.NewRateLimiter(5, time.Minute)

var hppOptions = mw.HPPOptions{
	CheckQuery:              true,
	CheckBody:               true,
	CheckBodyForContentType: "application/x-ww-form-urlencoded",
	Whitelist:               []string{"sortBy", "sortOrder", "name", "age", "class"},
}

var hppMiddleware = mw.HppMiddleware(hppOptions)

var middlewares = []utils.MiddlewareFunc{
	hppMiddleware,
	mw.Compression,
	mw.SecurityHeaders,
	mw.ResponseTimeMiddleware,

	rl.Middleware,
	//Needs to be at the end
	mw.Cors,
}

func main() {
	port := ":3000"

	cert := "cert/cert.pem"
	key := "cert/key.pem"

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/teachers/", handlers.TeachersHandler)
	mux.HandleFunc("/students/", handlers.StudentHandler)

	mux.HandleFunc("/execs/", handlers.ExecsHandler)

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// secureMux := utils.ApplyMiddleWares(mux, middlewares)
	secureMux := mw.SecurityHeaders(mux)
	server := &http.Server{
		Addr:    port,
		Handler: secureMux,
		// Handler:   middlewares.Cors(mux),
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port:", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}
