output "topic_names" {
  description = "The names of the created Kafka topics."
  value       = [for topic in kafka_topic.topics : topic.name]
}