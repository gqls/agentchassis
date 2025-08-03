kubectl exec -it postgres-clients-0 -n ai-persona-system -- psql -U clients_user -d clients_db -c \
  "UPDATE agent_definitions
   SET default_config = '{
     \"workflow\": {
       \"start_step\": \"validate\",
       \"steps\": {
         \"validate\": {
           \"action\": \"validate_input\",
           \"description\": \"Validate the input\",
           \"next_step\": \"search_web\"
         },
         \"search_web\": {
           \"action\": \"web_search\",
           \"description\": \"Search for information\",
           \"topic\": \"system.agent.web-search.process\",
           \"next_step\": \"generate_content\"
         },
         \"generate_content\": {
           \"action\": \"generate_content\",
           \"description\": \"Generate content based on search results\",
           \"topic\": \"system.agent.content-creator.process\",
           \"next_step\": \"transform\"
         },
         \"transform\": {
           \"action\": \"transform_data\",
           \"description\": \"Transform the result\",
           \"config\": {
             \"transformation\": \"uppercase\"
           },
           \"next_step\": \"notify\"
         },
         \"notify\": {
           \"action\": \"send_notification\",
           \"description\": \"Send final notification\",
           \"next_step\": \"complete\"
         },
         \"complete\": {
           \"action\": \"complete_workflow\"
         }
       }
     }
   }'
   WHERE type = 'generic';"