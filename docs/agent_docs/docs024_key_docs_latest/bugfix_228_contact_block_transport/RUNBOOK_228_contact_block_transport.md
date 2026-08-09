# RUNBOOK — contact-block transport (bugs_open/228)

Commands that were hard to get right, with their gotchas. Update in place when
one changes.

## Finding the bug (before assuming it's still valid — re-run these first)

```
curl -s https://robot-hands.com/tools/assets/contact-block.js | wc -c
curl -s https://robot-hands.com/tools/assets/contact-block.js \
  | grep -cE 'fetch\(|XMLHttpRequest|sendBeacon|form\.submit\(|action *='
# expect 2100 / 0 if the bug is still unfixed on the live asset
curl -s https://robot-hands.com/contact.html | grep -o '<form[^>]*>'
# expect the cb-form tag with no action=/method= if still unfixed
```

Placement-drift check (do this BEFORE trusting any page list, including this
one's — the filing lane found 5 such rows fleet-wide on 2026-08-08):
```
curl -s https://<site>/<page>.html | grep -c 'data-component="contact-block"\|cb-contact-form'
```

## Consumer enumeration (re-run before any future change to this mechanism)

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT function FROM content_components WHERE is_active AND html_template LIKE '%form_action%';
"
```
2026-08-08 result: exactly one row, `contact-form`.

## Building and shipping the Go change

```
make build-agent-chassis
# prints "Building agent-chassis from committed ref HEAD = <sha>" -- confirm
# the sha matches your commit before trusting the build.
docker push docker.io/aqls/agent-chassis:$IMAGE_TAG
```

**Gotcha:** `IMAGE_TAG` in the makefile is a single shared default other
sessions edit locally and don't always commit. Check the *deployed* tag
(`kubectl get deploy -o jsonpath=...`) AND `git diff -- makefile` before
picking a new one, and go one past the highest number seen anywhere (deployed
or staged), not just +1 from whatever the file currently says.

## Deploy — do NOT run a single-service kubectl apply yourself

See [[releases-are-whole-fleet-make-release]] (memory). A single-service
`kubectl apply -k` was blocked by a permission classifier on 2026-08-03. Ask
the owner to run the whole-fleet release; do the verification yourself.

## Pod-verifying the Go change (mandatory before any DB change below)

```
for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o name); do
  kubectl -n ai-persona-system exec "${pod#pod/}" -- sh -c \
    'strings /app/agent-chassis | grep -c "seeded empty form_action for sanitiser"'
  # negative control -- deliberate misspelling, must also print 0:
  kubectl -n ai-persona-system exec "${pod#pod/}" -- sh -c \
    'strings /app/agent-chassis | grep -c "seeded empty form_axction for sanitiser"'
done
```
Both replicas must show **1** (or more) on the positive grep, **0** on the
negative, before proceeding. A roll happening is not evidence — this grep is
the gate ([[imperative-kubectl-scale-is-undone-by-the-next-deploy]] and
`bugs_open/153` both apply here: a same-tag rebuild, or a build kicked off
before your commit landed, ships a stale binary with a green rollout).

## The pending DB change (DO NOT RUN until the pod-verify above passes)

Save the current row first — it is live config with no git history, and this
is the only rollback:
```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
\copy (SELECT html_template, js_content FROM content_components WHERE function='contact-block') TO 'contact_block_before_228_fix.tsv'
"
```
(or a plain `SELECT` piped to a file if `\copy` isn't available from this
exec form — check `\d content_components` and confirm the column names first.)

Template: surgical `replace()`, not a full re-send of the 400+-line template
(safer, smaller diff, no transcription risk):
```sql
UPDATE content_components
SET html_template = replace(
  html_template,
  '<form class="cb-form" id="cb-contact-form" novalidate aria-label="{{.form_heading}}">',
  '<form class="cb-form" id="cb-contact-form" action="{{.form_action}}" method="POST" novalidate aria-label="{{.form_heading}}">'
)
WHERE function = 'contact-block';
```
Verify the replace actually matched (PostgreSQL's `replace()` is a silent
no-op on zero matches — check the row changed):
```sql
SELECT html_template LIKE '%action="{{.form_action}}"%' FROM content_components WHERE function='contact-block';
-- must be true after the UPDATE
```

JS: full replacement (see `js_content_after_228_fix.js` in this directory for
the exact text) via dollar-quoting to avoid escaping the single quotes in the
JS itself:
```sql
UPDATE content_components
SET js_content = $CB228$
<paste the new js_content_after_228_fix.js content here verbatim>
$CB228$
WHERE function = 'contact-block';
```

## Deploying the content_components change to the 2 live pages

Dispatch a page-rerender (same handler `check_contact_form_undeliverable.go`'s
own auto-remediation uses, reason `section_data_resolved`) for:
- `robot-hands.com` / `/contact.html`
- `leopardessconsulting.co.uk` / `/ai-readiness-quiz.html`

**Do NOT rerender `finetuning.uk/case-studies.html`** — placement-drift row,
component absent from the served page; a rerender could materialise it there.

Respect the ~300s no-dispatch window after any chassis pod restart
(CLAUDE.md).

## Verifying the live fix

```
curl -s https://robot-hands.com/contact.html | grep -o '<form[^>]*>'
# expect action="mailto:robot-hands@contactforsales.com?subject=robot-hands.com enquiry"
curl -s https://robot-hands.com/tools/assets/contact-block.js \
  | grep -cE 'setTimeout|has been sent'
# expect 0 (fabrication removed)
curl -s https://robot-hands.com/tools/assets/contact-block.js \
  | grep -c 'Opening your email client'
# expect 1 (new honest status string, and confirms the NEW asset was fetched)
```

Re-run the bug file's own census query and confirm `contact-block` now
reports `form_has_action = true` with no other row changed.
