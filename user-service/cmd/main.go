package main

import (
    "database/sql"
    _ "github.com/lib/pq"
    "fmt"
    "log"
    "net/http"
    "os"
    "github.com/gorilla/mux"
    "github.com/nats-io/nats.go"
    "wallet/user-service/internal/handlers"
)

func main() {
    // Connect to PostgreSQL
    dbURL := os.Getenv("POSTGRES_URL")
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal("Failed to connect to the database:", err)
    }
    defer db.Close()

    // Connect to NATS
    natsURL := os.Getenv("NATS_SERVER")
    nc, err := nats.Connect(natsURL)
    if err != nil {
        log.Fatal("Failed to connect to NATS:", err)
    }
    log.Println("Connected to NATS successfully")
    defer nc.Close()

    // Setup routes
    r := mux.NewRouter()
    r.HandleFunc("/createUser", handlers.CreateUser(db)).Methods("POST")
    r.HandleFunc("/getBalance", handlers.GetBalance(nc)).Methods("GET")

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    fmt.Printf("User service is running on port %s...\n", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}
