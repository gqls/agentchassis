-H responses_topic=system.responses.generic <<EOF
{
"action": "orchestrate",
"config": {
"group_type": "mvp-site-builder"
},
"input_data": {
"domain": "boxing-tickets.com",
"objective": "affiliate-sales",
"model": "PAS"
}
}
EOF

agent_group_definition
{
"start_step": "generate_build_plan",
"steps": {
"generate_build_plan": {
"action": "execute_llm_prompt",
"description": "Create the 'Build Plan' from a behavioral model",
"config": {
"prompt_template": "You are a Chief Marketing Strategist. A client wants a new site for '{{.domain}}' with the objective '{{.objective}}'. Your task is to generate a simple, generic JSON 'Build Plan' based on the '{{.model}}' behavioral model. The plan should only contain an array of sections with a 'semantic_purpose' tag.",
"input_fields": ["input_data.domain", "input_data.objective", "input_data.model"],
"ai_service": {
"provider": "anthropic",
"model": "claude-haiku-4-5-20251001",
"api_key_env_var": "ANTHROPIC_API_KEY"
},
"output_field": "build_plan_json"
},
"next_step": "assemble_template"
},
"assemble_template": {
"action": "assemble_from_library",
"description": "Build the site template using 'Intelligent Fallback'",
"config": {
"build_plan_field": "build_plan_json",
"in_house_components_table": "in_house_components"
},
"next_step": "populate_content",
"output_field": "html_template"
},
"populate_content": {
"action": "call_agent",
"description": "Call the content-creator-agent to fill the template",
"config": {
"agent_type": "content-creator-agent",
"timeout_seconds": 300,
"input_fields": ["html_template", "input_data"]
},
"next_step": "deploy_site",
"output_field": "final_html"
},
"deploy_site": {
"action": "commit_to_git",
"description": "Commit the final site to a new Git repo",
"config": {
"repo_name_field": "input_data.domain",
"files_to_commit": {
"index.html": "final_html"
}
},
"next_step": "complete_workflow"
},
"complete_workflow": {
"action": "complete_workflow"
}
}
}


---


New DB Table: in_house_components

    This is our "In-House Forge."

    Columns: id, name (e.g., "Generic Headline"), function (e.g., "problem_statement"), html (the component code).

    MVP Data: You would manually insert 3-4 "Base Fallback" components (e.g., generic-headline, generic-text-block).


New Action: assemble_from_library (for the Site Architect)

    This is a new Go function for an agent.

    It parses the build_plan_json from the previous step.

    It loops through the sections. For each semantic_purpose (e.g., "problem_statement"):

        It runs a SQL query: SELECT html FROM in_house_components WHERE function = 'problem_statement'.

        This is the "Intelligent Fallback": If it finds no match, it runs: SELECT html FROM in_house_components WHERE function = 'generic-text-block'.

    It stitches all the found HTML strings together and outputs the final html_template.

New Action/Adapter: commit_to_git (for the Deployer)

    This is a new adapter that can talk to the GitHub/GitLab API.

    It takes the final_html as input.

    It creates a new repository and commits the index.html file. This triggers your CI/CD or provides a handoff point.

