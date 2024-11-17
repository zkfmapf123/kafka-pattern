package main

import (
	"fmt"
	"internal/queue"
)

const (
	TOPIC_CCOUNTS = 1
	CH_COUNT      = 10000
)

func main() {

	// set kafka
	kafkaQueue, err := queue.NewQueue("kafka", "leedonggyu-queue")
	if err != nil {
		panic(err)
	}

	// producer kafka
	topics := []string{}
	for i := 0; i < TOPIC_CCOUNTS; i++ {

		topic, value := fmt.Sprintf("topic-%d", i), fmt.Sprintf("value-%d", i)

		topics = append(topics, topic)
		fmt.Println("producer : ", value)

		err := kafkaQueue.Producer(topic, value)
		if err != nil {
			panic(err)
		}
	}

	// consumers
	cArr, err := kafkaQueue.Consumer(topics)
	if err != nil {
		panic(err)
	}

	fmt.Println(cArr)

}
