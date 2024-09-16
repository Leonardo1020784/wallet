package kafka

import (
    "encoding/json"
    "fmt"
    "log"
    "os"

    ckafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
    "wallet/user-service/internal/models"
)

func ProduceUserCreatedEvent(user models.User) error {
    // Create the Kafka producer
    p, err := ckafka.NewProducer(&ckafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_BROKER")})
    if err != nil {
        return fmt.Errorf("failed to create producer: %v", err)
    }
    defer p.Close()

    // Marshal the user data to JSON
    userPayload, err := json.Marshal(user)
    if err != nil {
        return fmt.Errorf("failed to marshal user: %v", err)
    }

    // Define the Kafka topic
    topic := "user-created"

    // Create the Kafka message
    message := &ckafka.Message{
        TopicPartition: ckafka.TopicPartition{Topic: &topic, Partition: ckafka.PartitionAny},
        Value:          userPayload,
    }

    // Produce the message to Kafka
    err = p.Produce(message, nil)
    if err != nil {
        log.Printf("Failed to produce message: %v", err)
        return err
    }
    log.Printf("Produced message to topic %s", topic)
    // Wait for message delivery
    p.Flush(15 * 1000)

    return nil
}
