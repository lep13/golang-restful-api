package app

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/lep13/golang-restful-api/db"
	"github.com/lep13/golang-restful-api/handlers"
)

// InitializeAndRunServer initializes and starts the HTTP server.
func InitializeAndRunServer() {
	// Load environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Get MongoDB URI from environment
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		log.Fatal("MongoDB URI is not set in environment variables")
	}

	// Initialize MongoDB connection
	mongoClient := db.ConnectDB(mongoURI)

	// Get the collection from the MongoDB client
	collection := db.GetCollection(mongoClient)

	// Wrap the collection and pass it to the handlers
	wrappedCollection := db.NewMongoCollectionWrapper(collection)
	handlers.Initialize(wrappedCollection)

	// Create a new router using gorilla/mux
	router := mux.NewRouter()

	// Define your routes and their handlers
	router.HandleFunc("/users", handlers.GetUsers).Methods("GET")
	router.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	router.HandleFunc("/users/{id}", handlers.GetUser).Methods("GET")
	router.HandleFunc("/users/{id}", handlers.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", handlers.DeleteUser).Methods("DELETE")
	router.HandleFunc("/health", handlers.HealthCheck).Methods("GET")

	// Get the port from environment or set a default one
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	// Start the HTTP server
	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server is running on port %s", port)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
