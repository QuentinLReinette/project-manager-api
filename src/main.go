package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Project Management API server...")

	ConnectDatabase()

	// run schema migrations (structs defined in models.go)
	log.Println("Running schema migrations...")
	err := DB.AutoMigrate(&User{}, &Project{}, &Task{})
	if err != nil {
		log.Fatal("Database migration failed: ", err)
	}
	log.Println("Database migration completed successfully.")

	router := http.NewServeMux()

	router.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "pong", "status": "running"}`))
	})

	serverAddr := ":8080"
	log.Printf("API engine ready and listening on port %s", serverAddr)

	err = http.ListenAndServe(serverAddr, router)
	if err != nil {
		log.Fatal("Failed to start HTTP server: ", err)
	}
}
