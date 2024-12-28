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
