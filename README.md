# Kafka-in-go

## Infra (추후 MSK / Terraform)

[infra](./infra/docker-compose.yml)

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

## Producer 성능개선

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
