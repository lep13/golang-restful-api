package main

import (
    "log"
    "net/http"
    "os"
    "time"

    "github.com/lep13/golang-restful-api/handlers"
    "github.com/joho/godotenv"
    "github.com/gorilla/mux"
)

func main() {
    err := godotenv.Load(".env")
    if err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    mongoURI := os.Getenv("MONGO_URI")
    if mongoURI == "" {
        log.Fatal("MongoDB URI is not set in environment variables")
    }

    handlers.Initialize(mongoURI)

    r := mux.NewRouter()
    r.HandleFunc("/users", handlers.CreateUser).Methods("POST")
    r.HandleFunc("/users", handlers.GetUsers).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT")
    r.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")

    port := os.Getenv("PORT")
    if port == "" {
        port = "3000"
    }

    srv := &http.Server{
        Handler:      r,
        Addr:         ":" + port,
        WriteTimeout: 15 * time.Second,
        ReadTimeout:  15 * time.Second,
    }

    log.Printf("Server is running on port %s...", port)
    log.Fatal(srv.ListenAndServe())
}
