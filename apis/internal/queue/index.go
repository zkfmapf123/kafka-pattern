package queue

import (
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaQueue struct {
	producer *kafka.Producer
	consumer *kafka.Consumer

	deliveryChanNum int
}

// kafka
func NewQueue(queue string, id string) (KafkaQueue, error) {

	kq := KafkaQueue{}

	// producer
	p, err := newProducer(id)
	if err != nil {
		return kq, err
	} else {
		kq.producer = p
	}

	// consumer
	c, err := newConsumer(id)
	if err != nil {
		return kq, err
	} else {
		kq.consumer = c
	}

	fmt.Println("kq producer : ", kq.producer)
	fmt.Println("kq consumer : ", kq.consumer)

	if queue == "kafka" {
		return kq, nil
	}

	return kq, fmt.Errorf("not Exists Queue : %s", queue)
}

func newProducer(producerId string) (*kafka.Producer, error) {

	// producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092", // update
		"client.id":         producerId,
		"acks":              "all",
	})

	return p, err

}

/*
auto.offset.reset
- smallest : 가장 오래된 오프셋부터 메시지 소비
- largest  : 가장 최신 오프셋부터 메시지 보시
- none 	   : 오프셋이 설정되지 않은 경우 오류 발생
*/
func newConsumer(groupId string) (*kafka.Consumer, error) {

	// consumer
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
		"group.id":          groupId,
		"auto.offset.reset": "smallest",
	})

	return c, err
}

func (k *KafkaQueue) SetChan(num int) {
	if num < 100 {
		num = 10000
	}

	k.deliveryChanNum = num
}
