# Run the migration
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db < platform/database/migrations/006_enhance_orchestrator_state.sql

# Update your generic agent with a test workflow
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
  "UPDATE agent_definitions
   SET default_config = '{
     \"workflow\": {
       \"start_step\": \"validate\",
       \"steps\": {
         \"validate\": {
           \"action\": \"validate_input\",
           \"description\": \"Validate input\",
           \"next_step\": \"transform\"
         },
         \"transform\": {
           \"action\": \"transform_data\",
           \"description\": \"Transform to uppercase\",
           \"config\": {
             \"transformation\": \"uppercase\"
           },
           \"next_step\": \"notify\"
         },
         \"notify\": {
           \"action\": \"send_notification\",
           \"description\": \"Send notification\",
           \"next_step\": \"complete\"
         },
         \"complete\": {
           \"action\": \"complete_workflow\"
         }
       }
     }
   }'
   WHERE type = 'generic';"

# Send a test message
# Your test code from before...

# Check the execution
kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
  "SELECT
     correlation_id,
     status,
     current_step,
     execution_metadata->>'completed_steps' as completed,
     execution_metadata->>'total_steps' as total
   FROM orchestrator_state
   ORDER BY created_at DESC
   LIMIT 5;"