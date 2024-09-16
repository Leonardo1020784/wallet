package kafka

import (
    "database/sql"
    "encoding/json"
    "log"

    ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type UserCreatedEvent struct {
    UserID    string `json:"user_id"`
    Email     string `json:"email"`
    CreatedAt string `json:"created_at"`
}

func StartUserCreatedConsumer(db *sql.DB, broker string) {
    config := &ckafka.ConfigMap{
        "bootstrap.servers": broker,
        "group.id":          "transaction-service",
        "auto.offset.reset": "earliest",
    }

    consumer, err := ckafka.NewConsumer(config)
    if err != nil {
        log.Fatalf("Failed to create Kafka consumer: %v", err)
    }
    defer consumer.Close()

    consumer.Subscribe("user-created", nil)

    for {
        // Correctly declare and initialize msg
        msg, err := consumer.ReadMessage(-1)
        if err != nil {
            log.Printf("Consumer error: %v", err)
            continue
        }

        var event UserCreatedEvent
        err = json.Unmarshal(msg.Value, &event)
        if err != nil {
            log.Printf("Error unmarshalling Kafka message: %v", err)
            continue
        }

        // Insert user into user_balance table
        _, err = db.Exec(`INSERT INTO user_balance (user_id, balance, created_at) VALUES ($1, 0.00, $2)`,
            event.UserID, event.CreatedAt)
        if err != nil {
            log.Printf("Error inserting user balance: %v", err)
        } else {
            log.Printf("User %s inserted with initial balance 0.00", event.UserID)
        }
    }
}
