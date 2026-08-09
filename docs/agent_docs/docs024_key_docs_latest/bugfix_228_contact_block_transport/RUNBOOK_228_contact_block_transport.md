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

## The pending DB change (DO NOT RUN until the pod-verify above passes, FLEET-WIDE)

**Superseded 2026-08-09 after council round 2** (correlation
`46f87e4c-05fc-4a5c-bd6a-93a073b63253`) — the inline SQL that used to be here
was exactly the NEEDLE-GATE SQL SURGERY shape `debug_historian` correctly
objected to (prose-described replace(), no needle-count guard, no backup, no
RETURNING, no separate rollback file). Use the scripts instead — they generate
a fresh backup + auto-derived rollback file every run, guard on an exact
needle count before writing, and require reading a `RETURNING` postcondition
before you `COMMIT`:

```
./apply_228_contact_block_fix.sh
```

It writes `BACKUP_<ts>_contact_block_before_fix.sql.{html_template,js_content}.txt`,
`ROLLBACK_<ts>_contact_block.sql` (generated FROM that backup, not hand-typed),
and `APPLIED_<ts>_contact_block.sql` (the actual UPDATE, with a `RETURNING`
clause proving all four target properties before you commit) — review the
`RETURNING` row, then `COMMIT;` or `ROLLBACK;` by hand in a follow-up
`psql -c`.

Pod-verify precondition is now fleet-wide, not 2-pod (see the landmine
`` `-l app=agent-chassis` returns 2 pods; 41 run that binary `` —
`prior_art_librarian`'s round-2 objection): enumerate by IMAGE across every
pod, Job and Deployment alike:
```
kubectl -n ai-persona-system get pods -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.ownerReferences[0].kind}{"\t"}{.spec.containers[0].image}{"\n"}{end}' | grep agent-chassis
```
Every row must be on the pod-verified-positive tag before running the script above.

## Deploying the content_components change to the 2 live pages

**Superseded 2026-08-09** — a DB row change does NOT propagate to
already-rendered pages on its own (`render_guardian`'s round-2 HIGH
objection: "this closes the DB row but can leave the live pages serving the
old fabricated-success markup — the bugs_open/024 false-green class"). Run
the dispatch script, which also prints the deployed-page verification curls:
```
./dispatch_228_rerenders.sh
```
It fires `page-rerender` with `reason=section_data_resolved` directly at the
two live pages (`robot-hands.com/contact`, `leopardessconsulting.co.uk/ai-readiness-quiz`)
— confirmed both have non-NULL `content_data` on their contact-block section
(2026-08-09 query), so this stays in the light re-render path and does NOT
escalate to full LLM content regeneration.

**Full blast radius, measured 2026-08-09 — exactly 3 `page_components` rows
use `contact-block`, none locked:** the two above, plus the already-documented
`finetuning.uk/case-studies` drift row. **Do NOT dispatch that third one** —
placement-drift, component absent from the served page; a rerender could
materialise it there, a behaviour change outside this fix's scope.

Respect the ~300s no-dispatch window after any chassis pod restart
(CLAUDE.md).

## Verifying the live fix

**Superseded 2026-08-09 after council round 3** (`prior_art_librarian`, HIGH
gating) — a bare curl right after dispatch lands on two documented
LANDMINES.md traps ("verifying straight after firing a rerender shows a 404
or stale copy" and "curl|grep twice against the same URL during a deploy
reports a regression that never happened"). Use `verify_228_deployed_page.sh`
instead — it checks `pages.deployed_at` before trusting a fetch, cache-busts,
and greps ONE saved response for every property rather than re-curling. Exact
invocations are in `dispatch_228_rerenders.sh`'s trailing output.

Re-run the bug file's own census query and confirm `contact-block` now
reports `form_has_action = true` with no other row changed.

## Council round 3, medium objections — investigated, not code changes

- **`reuse_agent`: should this use `content_components.forked_from` instead
  of in-place mutation?** Read `fork_theme_from_site_action.go`: forking
  exists to let ONE site diverge from a shared library asset (a per-site
  customisation promoted or split out), and creates a new row plus repoints
  that site's own reference. This fix wants the OPPOSITE property — both
  consuming sites should get IDENTICALLY the same repair, not a per-site
  fork. The `finetuning.uk` drift row isn't a content-divergence case forking
  would fix either: that page's serving pipeline doesn't render the
  component in its build output at all (a placement issue), so a forked row
  would sit there just as unreferenced as the shared one does today. In-place
  mutation of the one shared row is correct here; forking would add
  `page_components` repointing blast radius for no benefit.
- **`prior_art_librarian`: does `sql_for_agents/NNN_slug.sql` +
  `_ROLLBACK.sql` already cover this shape?** That house convention exists
  for `agent_definitions.default_config`-style JSON config patches, where the
  OLD value is a known literal at write time, so a rollback file can be
  hand-written statically alongside the apply file (e.g.
  `340_unbuilt_link_dispatch_authoritative_page_id_ROLLBACK.sql` just removes
  the keys the paired apply file added). `content_components.html_template`/
  `js_content` have no git-tracked source of truth — the only correct
  rollback content is whatever the row held immediately before THIS run, which
  cannot be known until runtime. `apply_228_contact_block_fix.sh`'s dynamic
  backup-then-generate-rollback step is a deliberate, evidenced departure
  from the static convention, not an oversight of it — a static rollback file
  for this specific mutation would risk restoring stale content if the live
  row had drifted between planning and execution.
- **`guardian`: does the html_template/js_content change risk the
  quote-in-`<script>`-block landmine?** Grepped the current template: its
  only `<script>` tag is `<script src="/tools/assets/contact-block.js">` —
  external reference, no inline substitution. `form_action`'s value is always
  either `""` or a server-built `mailto:<email>?subject=<domain> enquiry`
  (never containing a literal `"`). The landmine's footprint (a schema value
  substituted INTO an inline `<script>` block) does not apply to this edit.

## Council submission gotcha: a DB/config edit's `file` field must be a real repo path

The server-side validator (`diagnose_persist_fix_plan`) rejects any edit whose
`file` is not "repo-relative with no traversal or whitespace" — this is
**not** one of the 097 script's own documented "three type traps" and is not
caught by the client-side `jq` checks, so a descriptive non-path string (e.g.
`"content_components (function='contact-block', columns ...)"`) passes local
validation and then fails server-side as `complete_invalid` with no verdict
row, indistinguishable from "still queued" until you check
`agent_error_log WHERE orchestration_id=...`. For a DB/config-only edit that
has no corresponding source file, point `file` at a real repo file that
documents/carries the change instead (this RUNBOOK, or a saved copy of the
literal new content like `js_content_after_228_fix.js`) — never at a
descriptive label.
