kubectl exec -it postgres-clients-0 -n ai-persona-system -- sh

psql -U clients_user -d clients_db


  UPDATE agent_definitions
   SET default_config = '{
     "workflow": {
       "start_step": "validate",
       "steps": {
         "validate": {
           "action": "validate_input",
           "description": "Validate the input data",
           "next_step": "process"
         },
         "process": {
           "action": "transform_data",
           "description": "Transform the data",
           "config": {
             "transformation": "uppercase"
           },
           "next_step": "store"
         },
         "store": {
           "action": "store_result",
           "description": "Store the result",
           "next_step": "notify"
         },
         "notify": {
           "action": "send_notification",
           "description": "Send completion notification",
           "next_step": "complete"
         },
         "complete": {
           "action": "complete_workflow"
         }
       }
     }
   }'
   WHERE type = 'generic';