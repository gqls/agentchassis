output "job_name" {
  description = "Name of the system topics initialization job."
  value       = module.kafka_topics.job_name
}