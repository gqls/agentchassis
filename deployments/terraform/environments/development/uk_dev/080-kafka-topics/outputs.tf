output "topic_names" {
  description = "The names of the created Kafka topics for the development environment."
  value       = [for topic in kafka_topic.topics_dev : topic.name]
}