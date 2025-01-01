package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"github.com/gofiber/fiber/v2"
)

var (
	// SERVER
	PORT = os.Getenv("PORT")

	// KAFKA
	KAFKA_TOPICS         = os.Getenv("KAFKA_TOPICS")
	KAFKA_CONSUMER_GROUP = os.Getenv("KAFKA_CONSUMER_GROUP")
	KAFKA_BROKERS        = os.Getenv("KAFKA_BROKERS")
	
	// RETRY (Exponential)
	KAFKA_RETRY_COUNT = os.Getenv("KAFKA_RETRY_COUNT") // 최대 재시도 횟수
	KAFKA_BACKOFF = os.Getenv("KAFKA_BACKOFF") // 재시도 간격
)

func main() {

	app := fiber.New()
	kafkaConfig := NewKafka()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello world")
	})

	// consumer
	go kafkaConfig.ConsumeBatch()

	app.Listen(fmt.Sprintf(":%s", PORT))
}

type kafkaConn struct {
	consumer sarama.ConsumerGroup
}

func NewKafka() kafkaConn {

	kafkaBrokers := strings.Split(KAFKA_BROKERS, ",")

	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V3_6_0_0

	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // 최신 메시지부터 수동커밋...

	consumerGroup, err := sarama.NewConsumerGroup(kafkaBrokers, KAFKA_CONSUMER_GROUP, config)
	if err != nil {
		panic(err)
	}

	return kafkaConn{
		consumer: consumerGroup,
	}
}

type BatchListener struct{}

// Cleanup implements sarama.ConsumerGroupHandler.
func (b *BatchListener) Cleanup(sarama.ConsumerGroupSession) error {
	log.Panicln("[Cleanup] 파티션 재할당")
	return nil
}

// Setup implements sarama.ConsumerGroupHandler.
func (b *BatchListener) Setup(sarama.ConsumerGroupSession) error {
	log.Println("[Setup] 파티션 할당")
	return nil
}

func messageProcessor(session sarama.ConsumerGroupSession, batch []*sarama.ConsumerMessage) {
	for _, msg := range batch {
		log.Println("Topic : ", msg.Topic, "Values : ", string(msg.Value))
		session.MarkMessage(msg, "")
	}

	log.Println("Commit... ", len(batch))
}

// ConsumeClaim implements sarama.ConsumerGroupHandler.
var (
	_BATCH_SIZE = os.Getenv("BATCH_SIZE")
)

// 메시지 처리
func (b *BatchListener) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {	

	BATCH_SIZE ,_ := strconv.Atoi(_BATCH_SIZE)
	BATCH_TIMEOUT := 5 * time.Second
	
	var batch []*sarama.ConsumerMessage
	batchTimer := time.NewTimer(BATCH_TIMEOUT)

	for {
		select {
			case msg := <- claim.Messages():
				batch = append(batch, msg)
				if len(batch) >= BATCH_SIZE {
					messageProcessor(session, batch)	
					batch = nil
					batchTimer.Reset(BATCH_TIMEOUT)
				}
			
			case <-batchTimer.C:
				if len(batch) >0 {
					messageProcessor(session, batch)
					batch = nil
				}
				batchTimer.Reset(BATCH_TIMEOUT)
		}
	}
}

func (k kafkaConn) ConsumeBatch() {

	topics := strings.Split(KAFKA_TOPICS, ",")
	go func() {
		for {
			if err := k.consumer.Consume(context.TODO(), topics, &BatchListener{}); err != nil {
				panic(err)
			}
		}
	}()
}

func (k kafkaConn) Close() error {
	log.Panicln("Closing kafka consumer...")
	return k.consumer.Close()
}
