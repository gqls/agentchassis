# RUNBOOK — gauntlet_dead_cta (every command that was hard to get right, with its gotcha)

## 1. Fire the feature-designer / implementer

```bash
# designer (work item is capability_gap:tools-api-gauntlet-debate, id 9ed684bc-…)
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/0NN_TRIGGER_feature_designer_v1.sh 9ed684bc-864a-4aa1-b17a-7ed061e08f2a
# implementer — ONLY via its orchestrator trigger, ONLY after council 'approved':
FEATURE_CORR=<approved-corr> ./docs/.../0NN_TRIGGER_feature_implementer_v1.sh
```
- **GOTCHA (fatal, repeated):** fire ONCE and wait. Ingest latency under load is
  MINUTES (observed 9 min). A refire creates a second implementer that races branch
  creation; both mutually die on E4. Treat a missing orchestration row as QUEUED for
  ≥10 min. Patient watcher: `scratchpad/fire_impl6.sh` shape.
- **GOTCHA:** E4 hard-refuses if `feat/<corr8>` exists — `git push origin --delete feat/<corr8>` first.
- **GOTCHA (bugs_open/071, fixed 2026-07-25 but know the signature):** a
  stage_commit await that expires while the COMMIT LANDED on the branch =
  produced-but-never-consumed response. Killer was `agent-job-cleanup` deleting
  live `job.*` topics every 10 min (guard label matched zero pods always). If it
  recurs: `kubectl -n ai-persona-system logs job/agent-job-cleanup-<tick>` must
  say "Live spawned workload … keeping", NOT "No running spawned pods".
- **GOTCHA (bugs_open/066):** the implementer runs in a SPAWNED pod using
  `agent_definitions.image_tag` — NOT the chassis deployment's image. After any
  chassis roll that the implementer needs:
  `UPDATE agent_definitions SET image_tag='<tag>' WHERE type IN ('feature-implementer','feature-implementer-orchestrator') …` (snapshot first).

## 2. Watch a run (never trust the wrapper; UTC/BST trap)

```sql
-- the implementer's OWN orchestration (wrapper completes early/independently):
SELECT orchestration_id, status, current_step FROM orchestration_states
WHERE owner_agent_type='feature-implementer' ORDER BY created_at DESC LIMIT 1;
-- designer verdicts by corr:
SELECT created_at, kind, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='<CORR>' ORDER BY created_at;
```
- DB times are UTC; local is BST (+1). A "10:45" orch at your 11:45 is the same moment.
- Refusal reason: `collected_data->>'__step_error'` on the agent's orchestration.
- Gate log: `collected_data->'stage_gate'->>'log'` (tail it; PASS/FAIL at the end).

## 3. Council gate submission (platform Go changes — the norm)

```bash
./docs/.../097_TRIGGER_council_review_v1.sh <submission.json>   # save SUBMISSION_CORR
RESUBMIT_CORR=<corr> ./docs/.../097_TRIGGER_council_review_v1.sh <same.json>  # same trail
```
- **GOTCHA:** `complete_invalid` with no council_report = an infra death (e.g.
  Anthropic endpoint timeout at a seat), NOT a judgement. Resubmit on the same trail.
- Verdict: doc_notes (categories ? 'council-gate') or diagnosis_artifacts by corr.

## 4. Owned-page content delivery (the P4 front-end path, proven 07-22)

- Template/js/schema edits: dollar-quoted UPDATE on `content_components` (backup first).
- Deliver: section-editor `content_edit` via the 086-style DIRECT orchestrator envelope
  — `scripts/deliver_gauntlet_section_edit.sh`. The bare 049b `action=orchestrate`
  envelope has silently failed to ingest here; don't use it.
- **GOTCHA:** `apply_section_edit` does NOT republish `js_content`. Follow with the
  assemble-only rerender: `scripts/republish_gauntlet_js.sh`.
- **GOTCHA:** section-editor leaves `pc.build_status='approved'` — set back to
  `'deployed'` (an assemble path may drop non-deployed components).
- Verify live cache-busted, matching strings the change CREATED (never generic CSS).

## 5. tools-api build & deploy — ISLAND, not the cluster (as done 2026-07-25)

> The cluster kustomize steps that stood here are SUPERSEDED: the exposure home
> is the island VM (Route B1) — the PR's kustomize manifests are unused.
> Full as-built record: `infra/island/RUNBOOK_island.md` §ENGINE LANDED.

```bash
# build from the 086 branch (carries tools-api source + post-merge fixes);
# bump IMAGE_TAG in the makefile first, commit both:
make build-tools-api-ref
# ship WITHOUT registry creds on the island:
docker save docker.io/aqls/tools-api:<TAG> | gzip | \
  ssh root@toolsapisuk.vs.mythic-beasts.com 'gunzip | docker load'
# update image: tag in infra/island/docker-compose.yml, scp it to /opt/island/,
# then on the island: cd /opt/island && docker compose up -d tools-api
```
- **Migrations go to the ISLAND Postgres** (`docker exec -i island-postgres-1
  psql -U tools_api -d tools_api`), ledgered in its `island_migrations` table —
  NOT clients_db, NOT schema_migrations. The island also carries a minimal
  `sites` table (CORS lookup + FK target) seeded with real cluster site ids.
- **GOTCHA:** the island runs the code exactly as committed — a defect fixed
  in-repo is inert until you rebuild + save|load + `compose up -d` (same
  image-first rule as the cluster, different transport).
- **GOTCHA (Cloudflare):** origin 502 bodies are REPLACED by Cloudflare's own
  page — use 503 for "engine offline" so the JSON shape survives; verify error
  shapes from OUTSIDE, not just island-local.
- Smoke from outside:
  `curl -s -X POST https://tools.apis.uk/api/v1/tools/gauntlet/round -H 'Origin: https://vonc.com' -d '{}'`
  plus the matrix: denied origin → 403, missing round → 404, real-round
  /position with no key → 503, preflight OPTIONS → 204, `/anything` → 404.

## 6. Experience re-plan re-fire (ONLY after the API answers a smoke POST) — PROVEN 2026-07-25

```bash
./docs/agent_docs/sql_for_agents/092_TRIGGER_experience_plan.sh vonc.com vonc-spark-game "the Spark daily-provocation game"
```
- FIRST: carry the liveness evidence into the compose decisions block (the 197
  channel) — a small migration replacing the D1-REVISED text's API description with
  the real verified URL + a compact citation of an actual round-trip (see 207).
- Accept only `approved` + `abstained:0` + `reviewers:5`; the current `is_current`
  plan is REJECTED-do-not-build until then.
- **GOTCHA (real, cost real rounds):** `compose`'s `input_fields` are only
  `[experience_context, input_data]`, and `load_context`'s query pulls LIVE SITE
  STATE (pages/components/work items/loader source) — neither reads any prior
  `council_report` or rejected `doc_plans` history. **A bare re-fire after an
  escalation starts genuinely blind** and will likely rediscover the same gaps at
  real cost. If a run escalates (`complete_escalated`, hit `max_rounds` without
  approval — this is a DESIGNED circuit-breaker, its own description says "the
  disagreement IS the round-boundary decision menu"), read the last few
  `council_report` bodies, write a small migration folding the SPECIFIC named
  objections into the compose Decisions channel (same replace()-on-an-anchor
  pattern as 197/207; see 209 for a worked example), THEN re-fire. This converged
  a previously-escalated plan in one round.
- **GOTCHA (real, found live):** all 5 reviewer seats (`review_journeys`,
  `review_feasibility`, `review_honesty`, `review_mvp`, `review_contracts`) share
  ONE `max_tokens` (8000 as of the pre-208 seed) for a compact-JSON-but-reasoning
  call — a large enough round can truncate any of them mid-verdict
  (`stop_reason=max_tokens`). Same defect class as `bugs_closed/067` (whole-artifact
  re-emitters undersized) but on reviewer seats; that bug's own addendum warned
  reviewers were left unswept. If it recurs, bump all 5 together (see 208), not
  just the one that failed — check `default_config->'workflow'->'steps'->'review_*'->'config'->'ai_service'->>'max_tokens'` for each.
- **Applying an owned migration WITHOUT sweeping other threads' pending ones:**
  `run-migrations.sh` (default) PROBEs every pending file safely, but its
  `--apply` runs the WHOLE pending batch — including other threads' unreviewed
  migrations. To apply only your own: probe first (`bash scripts/migration/
  run-migrations.sh`, no flags — dry-run, confirms your file is clean), then
  apply it BY HAND (`kubectl … psql -v ON_ERROR_STOP=1 < your_file.sql`, the
  file's own `BEGIN;`/`COMMIT;` + `DO $$ … RAISE EXCEPTION` guards do the real
  verification), then ledger it (`bash scripts/migration/run-migrations.sh
  --record-only <file> --note '<what you verified>'`). Used for 207/208/209 this
  session, each cleanly, none touching the other threads' concurrently-pending
  files (198/203×2/204/206×2/208-the-other-one).
- **Common paren mistake in a `jsonb_set(…, to_jsonb(replace(…)), true)` migration:**
  `to_jsonb(replace(X, old, new))` needs its OWN closing `)` before the
  `, true)` that closes `jsonb_set` — writing only one `)` after `$NEW$…$NEW$`
  closes `replace()` but leaves `to_jsonb()` unclosed, producing `syntax error at
  or near "WHERE"` at probe time (harmless — the probe transaction rolls back;
  fix and re-probe). Hit this twice (207, 209) before internalising it.

## 7. Island exposure (P3 — DONE, supersedes the bastion/WireGuard plan)

The owner-drafted bastion+WireGuard plan (`infra/README_bastion_exposure.md`) was
never built — a concurrent thread re-decided the route on 2026-07-24 to **Route
B1, a standalone VM** ("the island"), stronger isolation (production cluster
appears nowhere in the public path) and already live. As-built runbook + every
command: `infra/island/RUNBOOK_island.md`. `README_bastion_exposure.md` and the
WireGuard drafts are historical only — do not build from them.

## 8. P4 front-end delivery — the full sequence, with what bit (2026-07-26)

Sources, harnesses and the pre-change backups all live in `p4_sources/`.

```bash
SC=docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/p4_sources

# 0. Feed. Regenerate + publish to the deploy repo (NOT the DB — there is no
#    provocations table; the file is committed into gqls/sites and synced to B2).
python3 $SC/build_provocations.py > provocations.json
SHA=$(gh api repos/gqls/sites/contents/vonc.com/data/provocations.json --jq '.sha')
# PUT with {message, content:<base64>, sha}; payload on STDIN via `gh api --input -`
#   (argv blows ARG_MAX on anything large — see scripts/webdesign_publish_assets.sh)

# 1. Component rows. Dollar-quoted, wrapped in BEGIN/COMMIT with a DO $$ guard
#    that RAISEs unless the new markers are present — a silent no-op UPDATE is
#    otherwise indistinguishable from success.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 < update_gauntlet.sql

# 2. Deliver the section (generalised; takes the component + a field_updates file)
./docs/.../scripts/deliver_section_edit.sh <page_component_id> <field_updates.json>

# 3. apply_section_edit does NOT republish js_content — assemble-only rerender:
./docs/.../scripts/republish_gauntlet_js.sh

# 4. section-editor leaves pc.build_status='approved' — put it back:
UPDATE page_components SET build_status='deployed' WHERE id='1048b344-...';

# 5. Snippets bundle (the archive loader ships in assets/js/snippets.js, a
#    DIFFERENT pipeline from the component's tools/assets/<function>.js):
./scripts/initial_messages/210_vonc_trigger/083_trigger-asset-renderer-vonc.sh
```

- **GOTCHA (cost a dispatch, 2026-07-26):** a spawn fired within ~300 s of a
  chassis pod restart is silently dropped — no orchestration row, no error. The
  chassis had restarted at 14:57:24Z because the DB was crash-looping
  (`bugs_open/082`); the 15:02 dispatch produced nothing. **Before blaming queue
  latency, check the chassis pod's `startedAt`** — that distinguishes "dropped"
  from "queued", and only one of them is safe to re-fire:
  `kubectl -n ai-persona-system get pod -l app=agent-chassis -o jsonpath='{.items[0].status.containerStatuses[0].state.running.startedAt}'`
- **GOTCHA:** `git add` of a workstream `build/` directory is silently refused —
  the repo's `.gitignore` catches `build` anywhere. Name it something else
  (`p4_sources/`); `git check-ignore -v <path>` tells you before you commit.

## 9. Driving a component end-to-end locally before delivering it

`p4_sources/drive_gauntlet.py` and `drive_archive.py` (python venv + playwright
chromium; `pip install` needs `python3 -m venv` — the system python is
PEP-668-managed and refuses). They render the real template with the real
`field_updates`, serve the real JS at the path the live page uses, and drive it
in Chromium against the LIVE API. This caught two defects that no
selector-existence check would have: a closed detail region reading as
populated, and a dead `href="#"` left on the hidden clone-source.

- **GOTCHA:** the browser's CORS **preflight** cannot be re-stamped by Playwright
  routing, so a localhost page gets 403 on `OPTIONS`. Proxy the call
  server-side with `Origin: https://vonc.com` and fulfil the route.
- **GOTCHA:** a 403 from `tools.apis.uk` has TWO different senders. The API's own
  is `{"error":"origin not allowed"}`; Cloudflare's is the plain-text
  `error code: 1010`, which it returns for a bare `Python-urllib` fingerprint.
  Send a browser `User-Agent` from any script that calls the API.
- **GOTCHA (shapes an acceptance run):** `browserrunner/run_checks_action.go:200`
  waits `stepDelay = 300ms` between an interaction step and its assertion, and
  `Text()` is Playwright `InnerText()`. Two consequences: (a) any check asserting
  on AI output will fail, because `/position` and `/defend` measure 8–18 s; and
  (b) `innerText` on a `display:none` element falls back to `textContent`, so a
  hidden-but-populated region reads as non-empty and a check can pass without the
  interaction having done anything.

## 10. Delivering a component that has NO template variables (the Arena path, proven 2026-07-27)

The §4/§8 path (`apply_section_edit` + `field_updates`) **does not work here** and
will refuse you: the Arena's `html_template` has zero `{{ }}` variables and its
`page_components.content_data` is `{}`. `deliver_section_edit.sh` rejects an empty
field_updates object by design. Use this instead.

**The load-bearing fact:** `rerender_single_page` assembles the page from
**`page_components.rendered_html`** (`rerender_single_page_action.go:163, 232, 511`),
*not* from `content_components.html_template`. Write only the template and the
live page will not change; write only `rendered_html` and the component library
silently diverges from what is served. **Write both, in one transaction.**

```bash
# 1. Back up BOTH columns byte-exactly. base64 through the wire so nothing is
#    mangled; psql -A -t alone will not preserve an arbitrary HTML blob.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
 "SELECT encode(convert_to(html_template,'UTF8'),'base64') FROM content_components WHERE id='<cc_id>';" \
 | tr -d '\n' | base64 -d > backups/backup_<name>_$(date +%F).html

# 2. CHECK THE BYTES. A backup taken while Postgres is crash-looping comes back
#    EMPTY and looks like a success. Compare against octet_length, NOT length:
#    length() counts CHARACTERS, so a UTF-8 blob reports fewer than its bytes
#    (the Arena: 38,632 chars / 38,704 bytes — both correct, neither is corruption).
wc -c backups/backup_<name>_*.html
```

```bash
# 3. Write both columns in ONE transaction, base64 so quoting cannot bite.
B64=$(base64 -w0 new_component.html)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -c "
BEGIN;
UPDATE content_components SET html_template = convert_from(decode('${B64}','base64'),'UTF8'), updated_at=NOW() WHERE id='<cc_id>';
UPDATE page_components  SET rendered_html = convert_from(decode('${B64}','base64'),'UTF8'), updated_at=NOW() WHERE id='<pc_id>';
COMMIT;"
```

```bash
# 4. Fire the assemble-only rerender. rerender_single_page_action.go has NO
#    rebuild_policy check, so this is safe on an `owned` page — unlike a generic
#    rerender, which is hard-refused (bugs_closed/024).
./docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/rerender_arena_vonc.sh
```

**Gotchas, each of which cost time:**

- **Keep `<style>` ahead of `<script>`** in the template. A `<style>` block placed
  after the script is silently deleted (`bugfix 072`).
- **Do not reuse `210_vonc_trigger/085_*.sh` as a publish template.** It uses
  `kubectl run -i --rm … -- kcat -P <<JSON`, which races stdin attachment: kcat
  sees EOF, publishes NOTHING and exits 0. `scripts/rerender_arena_vonc.sh` has
  the hardened form (payload in the container COMMAND, `--command` to beat the
  kcat ENTRYPOINT, `&& echo PUBLISH_OK`).
- **A missing orchestration row is not a dropped dispatch.** The discriminating
  check is consumer lag, not elapsed time:
  ```bash
  kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
    bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
    --describe --group generic-requests-group
  ```
  Non-zero LAG with a consumer attached = **queued, do not re-fire**. Measured
  2026-07-27: lag 2, row appeared 4.5 min later, behind a `review_guardian` run.
- **Do not read "orchestrations are completing" as "the lane is moving".**
  `check_endpoint_health` runs every ~90s on its own lane (since 030 closed) and
  will happily show COMPLETED rows while the generic request lane is blocked.
- **This path DOES maintain `pages.deployed_at` and `build_status`** (Arena:
  both `deployed`, stamped 13:08:55). That is a difference from the
  `apply_section_edit` path, which left `tool-gauntlet` at `needs_rebuild` with an
  11-day-stale `deployed_at` — so do not generalise one to the other.

**Verify on the served page, never the row**, and always with a negative control:

```bash
curl -s -A 'Mozilla/5.0 … Chrome/126.0.0.0 …' "https://<domain>/<path>?cb=$(date +%s)" -o live.html
grep -c 'THING_THAT_MUST_BE_GONE' live.html   # expect 0
grep -c 'THING_THAT_MUST_REMAIN' live.html    # expect >0, else the zeros above are vacuous
```

A browser UA is required: Cloudflare 403s a bare `Python-urllib`/curl-default
fingerprint with plain-text `error code: 1010`, which is NOT the origin check.
This applies to `vonc.com` itself, not just `tools.apis.uk`.

## 11. CORRECTION to §10 — `rendered_html` is NOT always a copy of `html_template`

**§10 says "write the same new string to both columns". That is only safe for a
component with NO template variables.** It was written from the Arena, which has
none and an empty `content_data`. Following it on the **gauntlet**, which has 27
`{{.placeholders}}` and a populated `content_data`, **shipped a live page showing
raw `{{.hero_title_plain}}` to the owner** (2026-07-28).

**The rule:**

- `content_components.html_template` holds the template, **with** `{{.vars}}`.
- `page_components.rendered_html` holds that template **with the vars already
  substituted from that page_component's `content_data`**.

They are byte-identical **only** when the template has no variables.

**Check before copying — one query answers it:**

```sql
SELECT (SELECT count(*) FROM regexp_matches(cc.html_template,'\{\{\.[a-zA-Z_]+\}\}','g')) AS vars,
       (SELECT count(*) FROM jsonb_object_keys(COALESCE(pc.content_data,'{}'::jsonb))) AS data_keys
FROM content_components cc JOIN page_components pc ON pc.component_id = cc.id
WHERE cc.id = '<component>';
-- vars = 0  -> §10's copy-both-columns path is safe
-- vars > 0  -> you MUST substitute; copying the template blanks the page's copy
```

**Substituting (this is the repair that fixed it):**

```bash
# pull content_data, substitute, assert nothing is left unrendered
python3 - template.html content_data.json out.html <<'PY'
import json,re,sys
t=open(sys.argv[1]).read(); d=json.load(open(sys.argv[2]))
missing=[n for n in set(re.findall(r'\{\{\.([a-zA-Z_]+)\}\}',t)) if n not in d]
assert not missing, f"content_data is missing: {missing}"
for k,v in d.items(): t=t.replace("{{."+k+"}}", "" if v is None else str(v))
left=re.findall(r'\{\{\.[a-zA-Z_]+\}\}',t)
assert not left, f"still unrendered: {sorted(set(left))}"
open(sys.argv[3],'w').write(t)
PY
```

**Then always verify the SERVED page, not the row:**

```bash
curl -s -A 'Mozilla/5.0 … Chrome/126' "https://<domain>/<path>?cb=$(date +%s)" | grep -c '{{\.'   # expect 0
```

**Why this got through:** every check I ran was on the change I had *made* — the
type sizes, brace balance, line counts — and all of them passed. Nothing asserted
that the page still rendered its own content. **A diff proves what you changed; it
cannot tell you what you destroyed.** Grep the served page for `{{.` after ANY
component delivery — it costs one command and it is the difference between
shipping a page and shipping a template.

## 12. Delivering a change to the gauntlet component (proven 2026-07-29, the ledger)

The gauntlet has **27 `{{.vars}}` and populated `content_data`**, so §11's rule
applies: `html_template` and `rendered_html` are NOT the same string. Build the
rendered copy by substituting, never by copying the template.

**One guarded transaction, all three columns.** The `WHERE updated_at = <the
value you read>` clauses are the concurrency guard — another session's write
between your read and your write zeroes the row counts instead of silently
overwriting them. The `DO` block is the anti-no-op guard: a `UPDATE 0` and a
successful write are otherwise indistinguishable in the output.

```bash
# ids (stable): cc=5da50747-7936-4b8f-a66d-c1ea98919c75  pc=1048b344-f1fa-44ea-b936-951bc7eafc59
# read updated_at FIRST and paste it into both WHERE clauses.
T=$(base64 -w0 new_template.html); J=$(base64 -w0 new.js); R=$(base64 -w0 new_rendered.html)
# BEGIN; UPDATE content_components SET html_template=…, js_content=…, updated_at=NOW()
#   WHERE id='<cc>' AND updated_at='<read value>';
# UPDATE page_components SET rendered_html=…, updated_at=NOW()
#   WHERE id='<pc>' AND updated_at='<read value>';
# DO $$ … RAISE EXCEPTION unless the new markers are present AND rendered_html
#   NOT LIKE '%{{.%' … $$; COMMIT;
```

**Then the assemble-only rerender — use the gauntlet script, not the old one:**

```bash
./docs/agent_docs/docs024_key_docs_latest/gauntlet_dead_cta/scripts/rerender_gauntlet_vonc.sh
```

- `scripts/republish_gauntlet_js.sh` still carries the **racy `kubectl run -i`
  heredoc** form that publishes nothing at exit 0 (~4 in 5 lost). The new
  script is the hardened form (payload in the container COMMAND, `--command`,
  `&& echo PUBLISH_OK`). Prefer it; it republishes `js_content` too.
- It **double-fires** (measured twice now: 07-28 E+F, 07-29 ledger — two
  orchestrations under one correlation, both COMPLETED). Assemble-only is
  idempotent so it is harmless here; count orchestrations **by payload**.

**Verify on the served page — and mind the URL shape:**

```bash
UA='Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36'
curl -s -A "$UA" "https://vonc.com/tools/gauntlet/index.html?cb=$(date +%s)" -o live.html
grep -c '{{\.' live.html            # expect 0
grep -c 'YOUR_NEW_MARKER' live.html # expect >0, else the zero above is vacuous
```

- **GOTCHA (cost a verifier run):** the page serves **only** at
  `/tools/gauntlet/index.html`. `/tools/gauntlet` and `/tools/gauntlet/` are
  both **404** — there is no directory index. Probe the variants before reading
  a 404 as a failed deploy.
- Full behavioural verifier (13 checks incl. a real round on the live page):
  `p4_sources/verify_live_ledger_2026-07-29.py`. It also asserts the **served
  JS is byte-identical** to what was written to `js_content` — the check that
  catches a rerender that redeployed HTML but not the asset.

## 13. Swapping the island engine and PROVING it is the new binary (2026-07-29)

```bash
ISL=root@toolsapisuk.vs.mythic-beasts.com
make build-tools-api-ref IMAGE_TAG=v1.0.XXXX     # pass the tag on the CLI, do NOT
                                                 # edit the makefile — another session's
                                                 # uncommitted IMAGE_TAG bump lives there
docker save docker.io/aqls/tools-api:v1.0.XXXX | gzip | ssh $ISL 'gunzip | docker load'
ssh $ISL 'cd /opt/island && cp docker-compose.yml docker-compose.yml.bak-<old>-pre<new> \
  && sed -i "s|tools-api:v1.0.<old>|tools-api:v1.0.<new>|" docker-compose.yml \
  && docker compose up -d tools-api'
```

**Identity — three checks, because a tag proves nothing:**

```bash
ssh $ISL 'cd /opt/island && docker compose ps --format "{{.Name}} {{.Image}} {{.RunningFor}}"'
ssh $ISL 'docker inspect aqls/tools-api:v1.0.XXXX --format "{{.Id}} {{.Created}}"'  # CreatedAt must be NOW, not a retag
ssh $ISL 'docker exec island-tools-api-1 sha256sum /tools-api'                       # must equal the local binary
# local side: C=$(docker create aqls/tools-api:v1.0.XXXX); docker cp $C:/tools-api ./b; docker rm $C; sha256sum ./b
```

- **Image IDs are NOT portable across `save|load`** (07-28 landmine) — compare
  the **binary hash**, which is.
- **Mirror the tag back into the repo compose.** Found 07-29: repo said 1178
  while live ran 1193 — an owner hand-swap that was never mirrored. The repo
  copy is not authoritative; `diff` it against the live file before editing.
- For a change with **no static marker** (a literal in an options map), the
  verification is behavioural: N consecutive real calls with the failure
  absent, plus the armed log still silent.
