const { Kafka } = require('kafkajs')

const kafka = new Kafka({
  clientId: 'my-app',
  brokers: ['localhost:9092']
})
const admin = kafka.admin();
console.log("Connecting...")
admin.connect();
console.log("Connected!")