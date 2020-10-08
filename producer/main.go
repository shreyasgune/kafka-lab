package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	kafka "github.com/segmentio/kafka-go"
)

var (
	kafkaTopic     string
	key            string
	value          string
)

func init() {
	kafkaTopic = os.Getenv("KAFKA_TOPIC")
	listTopics()
}

func main() {
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{fmt.Sprintf("%s:9092",os.Getenv("KAFKA_ADDR"))},
		Topic:    kafkaTopic,
		Balancer: &kafka.LeastBytes{},
	})
	fmt.Printf("%+v", w)
	r := gin.Default()
	r.POST("/data", func(c *gin.Context) {
		key = c.Query("key")
		value = c.Query("value")
		err := producer(key, value, w )
		if err != "" {
			c.AbortWithStatusJSON(400, gin.H{
				"error": err,
			})
		}  else {
			c.IndentedJSON(200, gin.H{
				"status": "Request has been pushed to the broker",
				 key:      value,
			})
		}
		
	})
	r.GET("/status", func(c *gin.Context) {
		hostname, err := os.Hostname()
		if err != nil {
			c.AbortWithStatusJSON(400, gin.H{
				"error": err,
			})
		}
		c.IndentedJSON(200, gin.H{
			"message": fmt.Sprintf("consumer:%s is OK", hostname),
			"stats":   w.Stats(),
		})
	})
	r.Run(":8080")

}

func producer(key string, value string, w *kafka.Writer) string {
	err := w.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		},
	)
	if err != nil {
		fmt.Println("failed to write messages:", err)
		return err.Error()
	}

	if err := w.Close(); err != nil {
		fmt.Println("failed to close writer:", err)
		return err.Error()
	}
	fmt.Printf("%s:%s pushed to broker", key, value)
	return ""
}

func createTopic() {
	// to create topics
	conn, err := kafka.DialLeader(context.Background(), "tcp", fmt.Sprintf("%s:9092",os.Getenv("KAFKA_ADDR")), kafkaTopic, 0)
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()
	topicConfigs := []kafka.TopicConfig{
		kafka.TopicConfig{
			Topic:             kafkaTopic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}

	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		panic(err.Error())
	}
}

func listTopics() {
	conn, err := kafka.Dial("tcp", fmt.Sprintf("%s:9092",os.Getenv("KAFKA_ADDR")))
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		panic(err.Error())
	}

	m := map[string]struct{}{}

	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}
	for k := range m {
		fmt.Println(k)
	}
	if len(m) == 0{
		createTopic()
	}
}
