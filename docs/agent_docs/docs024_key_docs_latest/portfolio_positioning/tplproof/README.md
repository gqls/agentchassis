# tplproof — render proof for migration 764 (bugs_open/453)

Pulls nothing itself: put the two live templates beside it, then `go run .`.

```sh
for pair in "build-site-planner:plan_site" "domain-research-classifier:classify_and_extract"; do t=${pair%%:*}; s=${pair##*:}
  kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A \
    -c "SELECT default_config #>> '{workflow,steps,$s,config,prompt_template}' FROM agent_definitions WHERE type='$t' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;" > "$t.tpl"; done
go run .
```

It applies 764's two replacements to each template in memory, parses with `text/template` + the same
`toJSON` shape as `datahelpers`, and executes three contexts (brief with `text` / brief object without /
no brief), asserting on the **delta** of `<no value>` against the unmodified template (0 / −1 / 0) and on
a sentinel appearing only in the fixed object case — plus a CONTROL that the unmodified template still
reproduces the defect, so a passing harness is one that could have failed.

⚠ The harness context is thin: research inputs (`Search Results`, `Scraped Website Content`, the layout
library line) render `<no value>` in BOTH orig and fixed, which is why the assertion is a delta and
not an absolute. A first version asserted absolute zero inside the Mission block and failed for that
reason — recorded so nobody "fixes" the assertion back.

This proves the template engine, not the fleet: after applying 764, make it run once (steps in the
migration header).
