topic doesn't exist for agent - see tests/agents

kubectl exec -it -n ai-persona-system postgres-clients-0 -- psql -U postgres
kubectl -n ai-persona-system exec -it postgres-clients-0 -- psql -U clients_user -d clients_db

\l                      # List databases
\du                     # List users
\c clients_db           # Try to connect to clients_db
\dt                     # List tables if connection works
\q                      # Exit

