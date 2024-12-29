# Kafak Essential

## Infra / Container

- [kafka docker-compose.yml 코드](./infra/docker-compose.yml)
- [api docker-compose.yml 코드](./docker-compose.yml)

## Producer Code (nodejs)

```javascript
import { Kafka, Partitioners } from "kafkajs";
import { getConfig } from "./env.confg.js";

class KafkaConfig {
  _kafkaConn = undefined;

  constructor(clientId, brokers) {
    if (!this._kafkaConn) {
      console.log("kafak connect...");
      this._kafkaConn = new Kafka({
        clientId,
        brokers: brokers.split(","),
        // requestTimeout: 10000,
        // retry: 5,
      });
    } else {
      console.log("kafka is already connect...");
    }
  }

  async isExistTopic(topic) {
    const admin = this._kafkaConn.admin();

    try {
      await admin.connect();

      const topics = await admin.listTopics();
      if (topics.includes(topic)) {
        return true;
      }

      return false;
    } catch (e) {
      console.error(e);
    }
  }

  async producer(value) {
    const topic = getConfig().KAFKA_TOPIC;
    const p = await this._kafkaConn.producer({
      createPartitioner: Partitioners.LegacyPartitioner,
    });

    try {
      // const isExistsTopic = await this.isExistTopic(topic);
      // console.log(isExistsTopic);

      await p.connect();
      await p.send({
        topic,
        messages: [{ key: value.id, value: JSON.stringify(value) }],
      });
    } catch (e) {
      console.error(e);
    } finally {
      await p.disconnect();
    }
  }
}

export const kafkaConn = new KafkaConfig(
  getConfig().KAFKA_CLUSTER_NAME,
  getConfig().KAFKA_BROKERS
);
```

## Consumer Code (Golang)

```golang
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
				fmt.Println("[consume] topic : ",topic, string(msg.Value))
			}
		}(pConsumer)
	}
}

func (k kafkaConn) Close() error {
	log.Panicln("Closing kafka consumer...")
	return k.consumer.Close()
}
```

## Consumer 성능개선 (BatchListener)

```golang

```

## Consumer

![simple consumer](./public/consumer.drawio.png)

## Consumer (Batch Listener)

![batch listener](./public/batch-consumer.drawio.png)

## Pub / Sub Application

![app](./public/app.drawio.png)

## Producer use lambda

![1](./public/1.png)

[Lambda Code](./producer-lambda/main.go)

## Issue

- EC2에 구성된 카프카 local에서 접근 시, Connection Error

```json

// Error
 [cause]: KafkaJSConnectionError: Connection error:
      at Socket.onError (/Users/idong-gyu/dev/study/kafka-in-go/producer/node_modules/kafkajs/src/network/connection.js:210:23)
      at Socket.emit (node:events:518:28)
      at emitErrorNT (node:internal/streams/destroy:169:8)
      at emitErrorCloseNT (node:internal/streams/destroy:128:3)
      at process.processTicksAndRejections (node:internal/process/task_queues:82:21) {

// 해결방법

// KAFAK_CFG_ADVERTISED_LISTENERS 의 EXTERNAL 주소가 localhost로 되어있었음
 - KAFKA_CFG_ADVERTISED_LISTENERS=CLIENT://kafka1:19092,EXTERNAL://localhost:9092

// Public IP 주소로 바꿔주자...
 - KAFKA_CFG_ADVERTISED_LISTENERS=CLIENT://kafka1:19092,EXTERNAL://<EC2 퍼블릭 IP>:9092
```

## Reference

- <a href="https://kafka.js.org/docs/getting-started">Kafakjs</a>
