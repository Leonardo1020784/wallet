package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "github.com/nats-io/nats.go"
    _ "github.com/lib/pq"
    "wallet/transaction-service/internal/handlers"
    "wallet/transaction-service/internal/kafka"
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

    // Start Kafka consumer to listen for user-created events
    go kafka.StartUserCreatedConsumer(db, os.Getenv("KAFKA_BROKER"))

    // Subscribe to NATS "balance.get" topic
    _, err = nc.Subscribe("balance.get", handlers.GetBalance(db))
    if err != nil {
        log.Fatal("Failed to subscribe to balance.get: ", err)
    }

    // Setup HTTP routes
    r := mux.NewRouter()
    r.HandleFunc("/addMoney", handlers.AddMoney(db)).Methods("POST")
    r.HandleFunc("/transferMoney", handlers.TransferMoney(db)).Methods("POST")


    // Start HTTP server
    port := "8081"
    fmt.Printf("Transaction service running on port %s...\n", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}
