package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/IBM/sarama"
	"github.com/gofiber/fiber/v2"
)

var (
	// SERVER
	PORT = os.Getenv("PORT")
	
	// KAFKA
	KAFKA_TOPICS = os.Getenv("KAFKA_TOPICS")
	KAFKA_CONSUMER_GROUP = os.Getenv("KAFKA_CONSUMER_GROUP")
	KAFKA_BROKERS = os.Getenv("KAFKA_BROKERS")
)

func main() {

	app := fiber.New()
	kafkaConfig := NewKafka()

	app.Get("/",func (c *fiber.Ctx)  error{
		return c.SendString("Hello world")
	})

	// consumer
	for _, topic := range strings.Split(KAFKA_TOPICS, ",") {
		go kafkaConfig.Consume(topic)
	}

	app.Listen(fmt.Sprintf(":%s", PORT))
}

type kafkaConn struct {
	consumer sarama.Consumer
}

func NewKafka() kafkaConn {

	kafkaBrokers := strings.Split(KAFKA_BROKERS, ",")

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V3_6_0_0

	consumer,err := sarama.NewConsumer(kafkaBrokers, config)
	if err != nil {
		panic(err)
	}

	return kafkaConn{
		consumer : consumer,
	}	
}

func (k kafkaConn) Consume(topic string)  {
	
	partitionList, err := k.consumer.Partitions(topic)
	if err != nil {
		panic(err)
	}

	for _, partition := range partitionList {
		pConsumer, err := k.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			panic(err)
		}

		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				fmt.Println("[consume] topic : ",topic, msg)
			}
		}(pConsumer)
	}
}

func (k kafkaConn) Close() error {
	log.Panicln("Closing kafka consumer...")
	return k.consumer.Close()
}
