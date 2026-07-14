## Context paragraph (paste into a new chat)

I'm working on a multi-agent Go/Kafka/Postgres system that plans and builds multipage
websites from a domain name — every agent is an orchestrator owning a workflow of one or
more steps that call Go actions, children reply to the parent's responses_topic, and sites
deploy git → GitHub Actions → Backblaze B2. Kubernetes namespace is ai-persona-system
(Kafka in namespace kafka); the reference site is idea.uk (site_id
1244516d-014d-421c-88c6-090bb1e9552a); SQL runs via
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`.
Working rules: Go not Python, British English, keep workflows simple and put complexity in
Go action code, don't build sub-workflows in SQL (spawn sub-agents instead), reuse and adapt
existing functions before writing new ones, always read the DB schema (`\d table`) before
writing SQL, prefer structural fixes over quick patches, never trust a 0-row result without
checking the query, never use logger.Debug, and work in reasonable step sizes. Recent work
(all deployed and verified): closed a scheme-to-components bug where idea.uk resolved a light
palette but rendered dark, by completing the library's paired "on-colour" variable convention
across components; restored a dropped "brief-explanation" section and generated its
illustrations; and fixed presigned B2 image URLs that would have expired by localising all
asset URLs to repo paths. A Go batch is deployed (scheme-aware fallback chrome, deploy-time
URL localisation, plan_sections section-scope resolution) and a further batch is delivered
awaiting apply/confirm (a re-aimed CSS fixer that classifies what a section paints rather
than reading an is_dark_section flag, plus creator-prompt and contract rewrites teaching the
same painting standard). The immediate open task is to build three catalogued-but-uncomposed
pages on idea.uk (news-index, guides-index, tool-audience-check) whose nav links 404 — they
have page rows but no site_plan_sections, so the fix is to re-run the existing
build-site-planner route (which safely unions already-built pages with the new ones via
normaliseRealisedToPlanPage) rather than hand-writing plan rows. A separate new task is
removing contact-form spam and starting an IP block list. There is a detailed handoff doc and
a running runbook/notes; please continue from the handoff.
