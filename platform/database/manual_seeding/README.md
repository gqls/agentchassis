cat platform/database/manual_seeding/001_test_schema.sql | kubectl exec -i postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db
cat platform/database/manual_seeding/002_test_data.sql | kubectl exec -i postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db
cat platform/database/manual_seeding/003_human_tasks.sql | kubectl exec -i postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db



cat platform/database/manual_seeding/004_test_cleanup.sql | kubectl exec -i postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db
