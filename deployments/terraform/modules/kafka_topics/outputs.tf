output "job_name" {
  value       = kubernetes_job.kafka_system_topics.metadata[0].name
  description = "Name of the system topics initialization job."
}