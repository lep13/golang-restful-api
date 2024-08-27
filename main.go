package main

import (
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "golang-restful-api/handlers"
)

func main() {
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }

    r := mux.NewRouter()
    r.HandleFunc("/users", handlers.CreateUser).Methods("POST")
    r.HandleFunc("/users", handlers.GetUsers).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
    r.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")
    r.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT")

    http.Handle("/", r)
    log.Fatal(http.ListenAndServe(":3000", nil))
}
