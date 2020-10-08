package main

import (
	"context"
	"fmt"
	"os"
	"github.com/gin-gonic/gin"
	kafka "github.com/segmentio/kafka-go"
)

var (
	// kafka
	kafkaBrokerUrl string
	kafkaTopic     string
	hostname       string
)

func init() {
	kafkaBrokerUrl = fmt.Sprintf("%s:9092",os.Getenv("KAFKA_ADDR"))
	kafkaTopic = os.Getenv("KAFKA_TOPIC")
}

func main() {
	// make a new reader that consumes
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{kafkaBrokerUrl},
		Topic:    kafkaTopic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	fmt.Printf("%+v",reader)
	r := gin.Default()
	consumer(reader)
	r.GET("/status", func(c *gin.Context) {
		hostname, err := os.Hostname()
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error": err,
			})
		}
		c.IndentedJSON(200, gin.H{
			"message": fmt.Sprintf("consumer:%s is OK", hostname),
			"stats":   reader.Stats(),
		})

	})
	r.Run(":8081")

}

func consumer(r *kafka.Reader) {
	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	if err := r.Close(); err != nil {
		fmt.Println("failed to close reader:", err)
	}
}

func getLatestConsumer(r *kafka.Reader) string {
	m, err := r.ReadMessage(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	return fmt.Sprintf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
}
