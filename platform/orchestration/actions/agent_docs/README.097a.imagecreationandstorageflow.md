Generic Agent
↓ spawns
Image-Generator Agent (dynamic)
↓ sends to
system.adapter.image-generator.requests (Kafka)
↓ consumed by
Image Generator Adapter (3 replicas)
↓ calls
Stability AI API → Generates image
↓ uploads to
Backblaze B2 Storage
↓ responds to
Parent Agent's Response Topic
↓ received by
Image-Generator Agent → Generic Agent