https://claude.ai/chat/6cfec7c0-d5b7-472c-b728-ebf1411a3171

Where We Are — The Big Picture

We have a working agent orchestration framework that builds multipage websites from domain names. The framework uses Kubernetes, Kafka, and Postgres with a hierarchical agent architecture where orchestrators spawn and call specialised child agents. The agent chassis is a single Go binary driven by SQL workflow definitions.
The strategic direction is to build websites that compete on genuine domain expertise by leveraging deep, vertical-specific knowledge bases. Rather than producing generic AI content, each domain is routed to a specialised vertical (veterinary, energy, mortgage, etc.) that maintains its own deep knowledge, research strategy, and content patterns. Knowledge compounds across domains within a vertical — the tenth site benefits from everything the first nine taught the system.
The architecture for this is fully designed. The RAG pipeline that underpins it has been patched and is ready for deployment. We are at the boundary between design and implementation.