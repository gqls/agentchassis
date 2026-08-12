# RUNBOOK — ai_site_selling_automation

Commands that were hard to get right, with their gotchas. Update HERE when one
changes.

## DB access

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Is the Anthropic account cap still in force? (one command, no cluster access)

```
curl -s -X POST https://preview.webdesign.uk/api/chat -H 'Content-Type: application/json' \
  -d '{"conversation_id":"","message":"how much does a website cost?"}'
```

A real answer → the cap has been lifted (LLM fleet is back; council gate
usable again). The "Thanks for your patience. Please reach us directly …"
contact line → still capped. Gotcha: HTTP status is 200 either way — the
fallback is fail-closed by design, so read the BODY, not the code. (Proven by
the webdesign lane 2026-08-10; see their HANDOFF_2026-08-10c §1 step 2.)

## Client rows, live

```
SELECT id, external_id, name, created_at FROM clients ORDER BY created_at;
```

As of 2026-08-10: 2 placeholder rows ("Default Client", "System Scheduler"),
1 network, 39 sites. Gotcha: today a customer's contact details land on the
SITE row (`sites.email/phone`, written by `082 --email`), not on clients —
until the columns migration in PLAN §2.1 ships, `clients` has no
customer-shaped columns at all.

## Submitting a build by hand (the only intake door)

```
scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh <domain> --email … --mission-file …
```

Gotchas (all from the start-here handoff §3.2): `client_id` is hardcoded
`demo_client` — the exact seam this lane exists to replace; `--fidelity` is
recorded but wired to nothing except `locked`; `hitl_mode=auto` synthesises
answers from classification defaults rather than merely skipping gates; and
dispatch itself is unreliable until `bugs_open/239` lands (queue starved —
the webdesign lane hand-drove every stage).

## Council submission: when a missing row IS a dropped dispatch (2026-08-11)

The standing rule "a missing orchestration row is almost always latency — do
not retry" assumes the publish SUCCEEDED. Distinguish by the kcat pod's exit:

- `% Delivery failed for message: Local: Message timed out` in the trigger
  output + `pod ... terminated (Error)` → the message NEVER reached the
  topic. Retry IS correct — with `RESUBMIT_CORR=<same corr>` so the printed
  correlation stays the one in your commit trailer. (Cause that night:
  kafka broker 0 was 0/1 NotReady; brokers 1-2 carried the retry.)
- kcat pod `Completed` + the trigger printing its full trailer-instructions
  tail → published; a missing row is queue latency, wait it out.

Prove the dispatch (not just the publish) by payload:
```
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```
Gotcha: don't pipe the 097 trigger through `tail`/`grep` when running it in
the background — the pipe buffers until EOF, so the correlation banner
(printed BEFORE the slow kcat step) stays invisible exactly when you want it.

## Verify the billing surface after an auth-service roll (owed post-roll; council debug_historian ask)

Ask the SERVICE, not git, and not `strings` (fleet practice, rewritten 08-11):

```
# 1. the pod runs a commit that contains the billing build
kubectl -n ai-persona-system logs -l app=auth-service --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor 1834bd3c0 <the stamp> && echo billing-commit-shipped

# 2. the billing mount line (startup, so it scrolls — check soon after the roll)
kubectl -n ai-persona-system logs -l app=auth-service --tail=300 | grep -m1 'billing'
#   "billing mounted without a payment provider" = mounted, keyless (expected until Stripe keys land)
#   "clients_database not configured"            = the config edit did NOT ship — stop and look
#   "clients database unreachable"               = mounted-code shipped but pool failed this run

# 3. the route answers (in-cluster; 401 proves the route EXISTS behind auth — a
#    missing route would 404)
kubectl -n ai-persona-system exec deploy/core-manager -- \
  curl -s -o /dev/null -w '%{http_code}' http://auth-service:8081/api/v1/admin/billing/settings   # expect 401
```

Gotcha: the degrade-to-unmounted design makes a silent non-deploy look like
success from outside — which is exactly why step 2/3 exist. Do all three.

```
# 4. (council round 2's addition — a mounted route is not a working WRITE path)
#    Issue a real voucher through the API with an admin JWT: expiry tomorrow,
#    recipient "post-roll acceptance". 201 + a WD-XXXXX-XXXXX code proves the
#    whole chain (route → service → clients_db insert). Vouchers are
#    single-use and expire, so the test artefact dies by itself — but note
#    its code here when you run it, so nobody mistakes it for an issued one.
curl -s -X POST -H "Authorization: Bearer $ADMIN_JWT" -H 'Content-Type: application/json' \
  http://auth-service:8081/api/v1/admin/billing/vouchers \
  -d '{"drops_price_to_pence":1000,"recipient_name":"post-roll acceptance","ttl_days":1}'
```

## Changing what a site is allowed to say, then changing what it says

Two separate jobs in a fixed order. The register first, always: the writer reads
it, so a rewrite dispatched against a stale register reproduces the old offer
very fluently. Worked end to end on webdesign.uk 2026-08-12.

### 1. Test the register BEFORE you write it (`cmd/claimscan`)

`claimscan` runs the platform's own scan engine over live data without
deploying anything. It is the only place a malformed register is visible: in
production, an unparseable `evidence_base` disarms the claims layer with no
error at all.

```bash
S=/tmp/claims                                          # anywhere writable
mkdir -p $S && go build -o $S/claimscan ./cmd/claimscan

# the corpus: every component of one site. locked_at IS NULL matters — a locked
# component is not the writer's to change, so it should not be in the sample.
kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -At -c \
  "SELECT p.name || E'\t' || COALESCE(pc.slot_name,'') || E'\t' ||
          replace(encode(convert_to(pc.rendered_html,'UTF8'),'base64'), E'\n','') ||
          E'\t' || COALESCE(p.page_type,'')
   FROM page_components pc JOIN pages p ON p.id = pc.page_id
   WHERE p.site_id = '<site>' AND pc.rendered_html <> '' AND pc.locked_at IS NULL" > $S/c.tsv

# the CURRENT register, as the control
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
  "SELECT data::text FROM site_specs WHERE site_id='<site>' AND aspect='evidence_base' AND is_current" > $S/old.json

$S/claimscan -evidence $S/old.json -components $S/c.tsv | tail -3   # control
$S/claimscan -evidence $S/candidate.json -components $S/c.tsv       # candidate
```

**Gotcha 1, and it is silent in production:** `EvidenceFact.Value` is
`*float64`. A fact with `"value": "after_approval"` fails the WHOLE document's
unmarshal, `ParseEvidenceBase` returns nil, and every claims check for that
site switches off. `claimscan` prints `parse evidence: … cannot unmarshal
string into Go struct field EvidenceFact.facts.value of type float64` — nothing
else ever will. Non-numeric attributes go in the `claim` prose or `source`.

**Gotcha 2:** compare the candidate against the CONTROL, not against zero. A
new ban set that finds what the old one found is decoration, and "0 findings"
from a register that failed to parse looks exactly like "0 findings" from a
clean site. On webdesign.uk the delta was 3 → 36 over the same 25 components.

**Gotcha 3:** scan your intended replacement wording too, as a one-row TSV, and
pass `-show-suppressed`. Bare `no` is deliberately not a negation cue, so a
`refund` ban makes *"there is no refund"* a page-blocking finding while
*"we do not offer refunds"* is suppressed. Put the required phrasing in
`writer_block` AND in the ban's `reason`.

### 2. Write the register by SUPERSEDING, never in place

`site_specs` carries its own history. An in-place `UPDATE` destroys the previous
register with no trace — that is how webdesign.uk's pre-deposit register was
lost. Copy `pinned` onto the new row or it loses its overwrite protection.
Full worked example, with its assertions: `SQL_2026-08-12_evidence_base_149.sql`.

### 3. Dispatch the rewrite as `content_rewrite` items with `mode=edit_live`

```sql
-- the shape apply_gap_plan_action.go writes; page-build-handler needs no
-- special path. spec.mode='edit_live' is LOAD-BEARING (bugs_open/178): without
-- it the writer never sees the page's current prose and fabricates a
-- replacement, measured elsewhere at 4,439 -> 1,806 chars.
INSERT INTO site_work_items
  (site_id, source, item_type, severity, summary, spec, page_id, priority,
   handler_agent, status, created_by, item_key, pipeline, triaged_at)
SELECT p.site_id, '<lane>', 'content_rewrite', 'high', '<summary>',
       jsonb_build_object('page_name', p.name, 'mode', 'edit_live',
                          'content_guidance', '<the brief>'),
       p.id, 20, 'page-build-handler', 'triaged', '<lane>',
       'copy_migration_<x>_' || p.name, 'build', now()
FROM pages p WHERE p.site_id='<site>' AND p.status='active' AND p.name IN (...);
```

**Filter on `p.status='active'`.** An archived page can share a url with a live
one (`index-rejected-v1-20260806` and `index` both carry `/index.html`), and
handing the archived one to the writer rewrites history.

**Queue depth is not the fleet backlog.** `find_dispatchable_site` orders by
`wi.created_at ASC` and takes one SITE per firing, skipping any site with a
`claimed` item. Count what is actually dispatchable before predicting a wait:
```sql
SELECT s.domain, count(*) FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.status IN ('triaged','approved') AND w.attempt_count < w.max_attempts
 GROUP BY s.domain ORDER BY 2 DESC;
```
Measured 2026-08-12: 35 items over 2 sites, and a new item was claimed in 87
seconds — not the "hours" the raw 722-item backlog figure implies.

### 4. Verify at the artefact, and allow for the sync

Check `last-modified`, not the body — the served HTML carries per-request
variation, so two fetches of the SAME file differ by md5.

```bash
curl -sSI https://preview.webdesign.uk/<page>.html | grep -i last-modified
```

**The box pulls on a timer, so `build_status='deployed'` leads the artefact.**
Measured 2026-08-12: faq deployed 16:43:49Z → served 16:45:03Z (74s); index
deployed 16:50:44Z → served 16:55:29Z (285s). Same mechanism, 4× the delay, so
it is a timer period and not a fixed lag. **Checking immediately reads as a
failed deploy** — and the estate's own "trust the artefact, not the status"
rule is what sends you to look. Confirm the content reached the repo first,
which separates "the deploy failed" from "the box has not pulled yet":

```bash
gh api repos/gqls/vm-sites/contents/<domain>/<page>.html --jq '.content' | base64 -d | grep -c '<your new string>'
```

### 5. If the page carries links, gate them mechanically

A prose instruction to preserve a SET is not followed (measured: 5 of 13 guide
links dropped). Declare the set on the page itself, then assert it:

```sql
UPDATE pages SET content_direction = COALESCE(content_direction,'{}'::jsonb)
     || jsonb_build_object('required_links', jsonb_build_array('/a.html','/b.html'))
 WHERE site_id='<site>' AND name='<page>';
```
```bash
G=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk/gate_page_links.py
python3 $G --domain <domain>              # must pass BEFORE (else the set is wrong)
python3 $G --domain <domain> --self-test  # must FAIL, or the gate is inert
# ... rewrite ...
python3 $G --domain <domain>              # the pass that means something
```
It reads `pages.content_direction` (the page-level column), not the `site_specs`
row of the same name — so it does not collide with a fleet voice pass editing
specs. Undeclared `required_links` makes it pass vacuously.
