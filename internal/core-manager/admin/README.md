New Endpoints in Swagger:
The documentation will now include:
Agent Definitions:

POST /api/v1/admin/agent-definitions - Create new agent type
GET /api/v1/admin/agent-definitions/{type}/topics/verify - Verify Kafka topics
POST /api/v1/admin/agent-definitions/{type}/topics/recreate - Recreate topics

Agent Instances:

GET /api/v1/admin/agent-instances - List all instances
GET /api/v1/admin/agent-instances/{agent_id} - Get instance details
PUT /api/v1/admin/agent-instances/{agent_id}/status - Toggle status
POST /api/v1/admin/agent-instances/{agent_id}/restart - Restart agent

All endpoints will show they require Bearer token authentication and admin role.