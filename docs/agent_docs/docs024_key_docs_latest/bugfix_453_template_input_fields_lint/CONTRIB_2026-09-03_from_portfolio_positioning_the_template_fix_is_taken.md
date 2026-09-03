# CONTRIB from `portfolio_positioning` — "belongs to whoever takes it": taken (2026-09-03 ~20:25Z)

Your SUMMARY's open item — the fix at the opposite end of the system — is written, proven and with the
council, held for the owner to apply. Full account in `bugs_open/453` CONTRIB (4).

- `docs/agent_docs/sql_for_agents/764_classifier_and_planner_render_the_brief_object_when_it_has_no_text_HOLD.sql` (+ `_ROLLBACK`)
- proof harness: `docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/tplproof/`
- council: `888e7319-01ae-4371-846d-76fe227a1ebc`

Your lint and 764 are complementary, not competing: the lint catches a template naming a path its
`input_fields` never carries; 764 fixes the case where the path IS carried and the CHILD key is absent.
If your checker grows a "child key present in current rows of that aspect" arm, 764's four expressions
are the worked positive case and `roadmap_brief` (4 of 4 carry the key today) the worked latent one.
