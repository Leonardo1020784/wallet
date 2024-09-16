package handlers

import (
    "database/sql"
    "encoding/json"
    "log"
    "net/http"
    "time"
    "wallet/user-service/internal/kafka"
    "wallet/user-service/internal/models"
    _ "github.com/lib/pq"
    "github.com/nats-io/nats.go"

)

func GetBalance(nc *nats.Conn) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        email := r.URL.Query().Get("email")
        if email == "" {
            http.Error(w, "Email is required", http.StatusBadRequest)
            return
        }

        log.Printf("Requesting balance for email: %s", email)

        // Request balance from transaction-service using NATS
        msg, err := nc.Request("balance.get", []byte(email), 5*time.Second)
        if err != nil {
            log.Printf("Error making NATS request: %v", err)
            http.Error(w, "Failed to fetch balance", http.StatusInternalServerError)
            return
        }

        log.Printf("Received response from transaction service: %s", string(msg.Data))

        // Parse NATS response
        var balanceResponse models.BalanceResponse
        err = json.Unmarshal(msg.Data, &balanceResponse)
        if err != nil {
            log.Printf("Error unmarshalling NATS response: %v", err)
            http.Error(w, "Invalid response from transaction service", http.StatusInternalServerError)
            return
        }

        // Send balance back to client
        json.NewEncoder(w).Encode(balanceResponse)
    }
}


func CreateUser(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        var user models.User
        err := json.NewDecoder(r.Body).Decode(&user)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
            return
        }

        user.CreatedAt = time.Now()

        // Insert user into the database
        query := `INSERT INTO users (user_id, email, created_at) VALUES ($1, $2, $3)`
        _, err = db.Exec(query, user.UserID, user.Email, user.CreatedAt)
        if err != nil {
            log.Printf("Failed to insert user into the database: %v", err)
            http.Error(w, "Failed to create user", http.StatusInternalServerError)
            return
        }

        // Produce event to Kafka
        err = kafka.ProduceUserCreatedEvent(user)
        if err != nil {
            log.Printf("Error producing Kafka event: %v", err)
            http.Error(w, "Could not produce event", http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(user)
    }
}

