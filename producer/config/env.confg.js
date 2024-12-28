import dotenv from "dotenv";
dotenv.config();

export const getConfig = () => {
  return {
    PORT: process.env.PORT,
    APP_NAME: process.env.APP_NAME,
    KAFKA_TOPIC: process.env.KAFKA_TOPIC,
    KAFKA_BROKERS: process.env.KAFKA_BROKERS,
    KAFKA_CLUSTER_NAME: process.env.KAFKA_CLUSTER_NAME,
  };
};
