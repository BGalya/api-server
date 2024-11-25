package main

import (
	api_sec "f5.com/ha/pkg"
	"log"
	"net/http"
)

func main() {
	// Create the log file in main
	err := api_sec.CreateLogFile()
	if err != nil {
		log.Fatalf("Error creating log file: %v", err)
	}

	http.HandleFunc("/register", api_sec.Register)
	http.HandleFunc("/login", api_sec.Login)

	http.HandleFunc("/accounts", api_sec.Auth(api_sec.AccountsHandler))
	http.HandleFunc("/balance", api_sec.Auth(api_sec.BalanceHandler))

	port := "8080"
	log.Printf("Server is running on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Could not start server: %s\n", err)
	}
}
