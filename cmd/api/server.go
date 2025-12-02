package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	mw "simpleapi/internal/api/middlewares"
	"simpleapi/internal/api/router"
	"simpleapi/internal/repository/sqlconnect"
	"simpleapi/pkg/utils"
	"time"

	"github.com/joho/godotenv"
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
	err := godotenv.Load()
	if err != nil {
		return
	}

	_, err = sqlconnect.ConnectDb()
	if err != nil {
		fmt.Println("Error connecting to DB", err)
		return
	}
	port := os.Getenv("API_PORT")

	cert := "cert/cert.pem"
	key := "cert/key.pem"

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// secureMux := utils.ApplyMiddleWares(mux, middlewares)
	router := router.Router()
	secureMux := mw.SecurityHeaders(router)
	server := &http.Server{
		Addr:    port,
		Handler: secureMux,
		// Handler:   middlewares.Cors(mux),
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port:", port)
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting the server", err)
	}
}
