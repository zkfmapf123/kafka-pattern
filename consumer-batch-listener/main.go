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
	IS_STATIC_MEMBERSHIP = os.Getenv("KAFKA_IS_STATIC_MEMBERSHIP")
	
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

	groupId := ""
	hostname, _ := os.Hostname()

	if IS_STATIC_MEMBERSHIP == "true" {
		groupId = fmt.Sprintf("consumer-%s", hostname)
	}else{
		groupId = hostname
	}

	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // 최신 메시지부터 수동커밋...

	// static membership
	// config.Consumer.Group.InstanceId = "consumer-"
	
	
	consumerGroup, err := sarama.NewConsumerGroup(kafkaBrokers, groupId, config)
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

func messageProcessor(session sarama.ConsumerGroupSession, batch []*sarama.ConsumerMessage) error {
	fmt.Println("messageProcessor")

	// return fmt.Errorf("exception...")

	for _, msg := range batch {
		
		log.Println("Topic : ", msg.Topic, "Values : ", string(msg.Value), "Partition : ", msg.Partition, "Offset : ", msg.Offset)
		session.MarkMessage(msg, "")
	}

	log.Println("Commit... ", len(batch))
	return nil
}

// ConsumeClaim implements sarama.ConsumerGroupHandler.
var (
	_BATCH_SIZE = os.Getenv("BATCH_SIZE")
)

// 메시지 처리
// 재시도로직 추가
func (b *BatchListener) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {	
	log.Println("consumerClaim")

	BATCH_SIZE ,_ := strconv.Atoi(_BATCH_SIZE)
	BATCH_TIMEOUT := 5 * time.Second

	RETRY_COUNT := 1
	MAX_RETRY_COUNT,_ := strconv.Atoi(KAFKA_RETRY_COUNT)
	RETRY_BACKOFF, _  := strconv.Atoi(KAFKA_BACKOFF)
	
	var batch []*sarama.ConsumerMessage
	batchTimer := time.NewTimer(BATCH_TIMEOUT)

	for {
		select {
			case msg := <- claim.Messages():
				batch = append(batch, msg)
				if len(batch) >= BATCH_SIZE {
					// Message Failed
					if err := messageProcessor(session, batch); err!= nil {
						mustRetry(RETRY_COUNT, MAX_RETRY_COUNT, RETRY_BACKOFF)
						RETRY_COUNT++	
					}else{
						batch = nil
						batchTimer.Reset(BATCH_TIMEOUT)
					}

				}
			
			case <-batchTimer.C:
				if len(batch) >0 {
					// BatchTime Failed
					if err := messageProcessor(session, batch); err != nil {
						mustRetry(RETRY_COUNT, MAX_RETRY_COUNT, RETRY_BACKOFF)
						RETRY_COUNT++
					}else{
						batch = nil
					}

				}
				batchTimer.Reset(BATCH_TIMEOUT)
		}
	}
}

func mustRetry(retryCount int, maxRetryCount int, backoff int) {
	
	// exit
	if retryCount == maxRetryCount {
		log.Fatalf("Retry 최대한도... %d", retryCount)
	}

	retryBackoffDuration := retryCount * backoff
	log.Printf("재시도 횟수 : %d 재시도 시간 : %ds",retryCount, retryBackoffDuration )
	time.Sleep(time.Duration(retryBackoffDuration))
}

func (k kafkaConn) ConsumeBatch() {
	log.Println("consumerbatch")

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
