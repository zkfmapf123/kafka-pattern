package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/IBM/sarama"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/hashicorp/go-uuid"
)

var (
	_bootstrap   = os.Getenv("BROKERS")
	_factorCount = os.Getenv("FACTOR_COUNT")
)

type APIGwEventParameter struct {
	RequestParameter requestParameter `json:"requestContext"`
}

type requestParameter struct {
	AccountId      string `json:"accountId"`
	RequestId      string `json:"requestId"` // 추후 이놈으로 토픽 생각 중
	Stage          string `json:"stage"`
	PathParameters struct {
		Topic string `json:"topic"`
	} `json:"pathParameters"`
	Body string `json:"body"`
}

type kafkaParameter struct {
	// acks optsion
	Acks          int  `json:"acks"` // 1, 0, -1
	ReturnSuccess bool `json:"returnSuccess"`
	ReturnError   bool `json:"returnErrors"`

	// tls options
}

type kafkaMessageParams struct {
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"` // -1 자동으로 설정
	Value     string `json:"value"`

	Key string
}

func createParameter() kafkaParameter {
	return kafkaParameter{
		Acks:          -1,
		ReturnSuccess: true,
		ReturnError:   true,
	}
}

func mustSerialize(e requestParameter) (kafkaMessageParams, []string, int) {

	// marshal
	var kafkaMessage kafkaMessageParams
	err := json.Unmarshal([]byte(e.Body), &kafkaMessage)
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Parameter : ", e)
	log.Printf("Topic : %s Partition : %d Value : %s\n", kafkaMessage.Topic, kafkaMessage.Partition, kafkaMessage.Value)

	bootstraps := strings.Split(_bootstrap, ",")
	factorCount, _ := strconv.Atoi(_factorCount)

	return kafkaMessage, bootstraps, factorCount
}

func HandleRequest(ctx context.Context, e requestParameter) error {

	// make variables
	kafkaMessage, bootstraps, factorCount := mustSerialize(e)

	// 1. set producer parameter
	kp, config := createParameter(), sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.RequiredAcks(kp.Acks)
	config.Producer.Return.Successes = kp.ReturnSuccess
	config.Producer.Return.Errors = kp.ReturnError

	// 2. producer
	producer, err := NewProducer(bootstraps, config)
	if err != nil {
		log.Fatalf("producer Error : %s", err.Error())
	}

	defer producer.Close()

	// 3. create topic
	uuid, _ := uuid.GenerateUUID()
	params := kafkaMessageParams{
		Topic:     kafkaMessage.Topic,
		Partition: kafkaMessage.Partition,
		Value:     kafkaMessage.Value,
		Key:       uuid,
	}

	topicErr := CreateTopic(bootstraps, config, params.Topic, int(params.Partition), factorCount)
	if err != nil {
		log.Fatalln(topicErr)
	}

	// 4. sendMessages
	parition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic:     params.Topic,
		Partition: params.Partition,
		Key:       sarama.StringEncoder(params.Key),
		Value:     sarama.StringEncoder(params.Value),
	})

	if err != nil {
		log.Fatalf("Send Message Error : %s\n", err.Error())
	} else {
		log.Printf("partition : %d offset : %d key : %s\n", parition, offset, uuid)
	}

	return nil
}

func NewProducer(bootstraps []string, config *sarama.Config) (sarama.SyncProducer, error) {
	producer, err := sarama.NewSyncProducer(bootstraps, config)
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func CreateTopic(bootstraps []string, config *sarama.Config, topic string, partitionCount int, replicationFactor int) error {

	client, err := sarama.NewClient(bootstraps, config)
	if err != nil {
		return err
	}

	defer client.Close()

	admin, err := sarama.NewClusterAdminFromClient(client)
	if err != nil {
		return err
	}

	defer admin.Close()

	details := sarama.TopicDetail{
		NumPartitions:     int32(partitionCount),
		ReplicationFactor: int16(replicationFactor),
	}

	err = admin.CreateTopic(topic, &details, false)
	if err != nil {
		return err
	}

	log.Printf("Topic %s created successfully\n", topic)
	return nil
}

func main() {
	lambda.Start(HandleRequest)
}
