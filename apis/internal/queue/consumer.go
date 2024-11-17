package queue

import (
	"fmt"
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func (k *KafkaQueue) Consumer(topics []string) ([]kafka.Event, error) {

	answers := []kafka.Event{}

	err := k.consumer.SubscribeTopics(topics, nil)
	if err != nil {
		return answers, err
	}

	run := true
	for run {
		ev := k.consumer.Poll(100)
		switch e := ev.(type) {
		case *kafka.Message:
			// processing
			fmt.Println("value : ", string(e.Value))
			answers = append(answers, e)

		case kafka.Error:
			// error
			// log.Println(e.Error())
			run = false

		default:
			// default
			log.Println("Default : ", e.String())
		}
	}

	k.consumer.Close()
	return answers, nil
}
