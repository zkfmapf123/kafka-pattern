import express from "express";
import { v4 as uuid } from "uuid";
import { getConfig } from "./config/env.confg.js";
import "./config/kafka.config.js";
import { kafkaConn } from "./config/kafka.config.js";

const app = express();
app.use(express.json());
app.get("/ping", (req, res) => res.status(200).send("hello"));
app.post("/create", async (req, res) => {
  const { level, name } = req.body;

  const event = "create";
  const id = uuid();

  const values = {
    id,
    event,
    level,
    name,
  };

  try {
    await kafkaConn.producer(values);

    console.log(
      `[Success] Producer Topic : ${getConfig().KAFKA_TOPIC} username : ${name}`
    );
  } catch (e) {
    console.error(e);
    return res.status(500).json("not create");
  }

  return res.status(200).json("create");
});

app.listen(getConfig().PORT, () => {
  console.log(`connect to ${getConfig().APP_NAME}:${getConfig().PORT} connect`);
});
