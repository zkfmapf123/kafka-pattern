package queue

import "github.com/confluentinc/confluent-kafka-go/kafka"

func (k *KafkaQueue) Producer(topic string, value string) error {

	deliveryCh := make(chan kafka.Event, k.deliveryChanNum)
	err := k.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          []byte(value),
	},
		deliveryCh,
	)
	return err
}
