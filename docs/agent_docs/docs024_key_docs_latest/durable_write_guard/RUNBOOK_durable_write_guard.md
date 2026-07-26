# RUNBOOK — durable-write completeness guard (bugs_open/021 INSTANCE 1)

Every query/command worth reusing, with its gotcha attached. DB access:

```
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db
```

## Where the guard lives / the simulation harness (from component_write_guard.go)

The guard was calibrated against `component_versions` transitions. To re-run that
simulation before changing any threshold (compare consecutive versions of each
component; a hard shrink that ends cleanly is a legitimate rewrite, one ending
mid-token is the shape to refuse):

```sql
WITH v AS (SELECT component_id, version_number, html_template AS cur,
       lead(html_template) OVER (PARTITION BY component_id ORDER BY version_number) AS nxt
     FROM component_versions)
SELECT c.name, length(cur), length(nxt),
       round(100.0*length(nxt)/length(cur)) AS pct,
       (right(rtrim(nxt),1)='>') AS ends_cleanly
FROM v JOIN content_components c ON c.id=v.component_id
WHERE nxt IS NOT NULL AND length(nxt) < length(cur)
ORDER BY 1.0*length(nxt)/length(cur) ASC;
```

## Recovery-table check (which durable fields have a history table)

```sql
-- history/version/snapshot tables
SELECT tablename FROM pg_tables WHERE schemaname='public'
  AND (tablename LIKE '%version%' OR tablename LIKE '%history%'
       OR tablename LIKE '%snapshot%') ORDER BY 1;
```
Findings 2026-07-21:
- `component_versions` — snapshots `html_template` (the 012 recovery table). GUARDED source.
- `page_component_history` — snapshots **`content_data` only**, NOT `rendered_html`.
  9,933 rows, latest 2026-07-21. So rendered_html has NO direct history table;
  recovery is by re-render from html_template.
- `site_snapshots.pages_snapshot` — captures pages.rendered_header/footer/head in
  JSONB, but the columns are always NULL so it captures/restores NULL.

## Target A verification — are pages.rendered_* really never written?

Grep proved no Go writer. Confirm against live data that they are empty
(a non-empty value would mean a writer exists somewhere I missed):

```sql
SELECT
  count(*) FILTER (WHERE rendered_header IS NOT NULL AND rendered_header<>'') AS hdr,
  count(*) FILTER (WHERE rendered_footer IS NOT NULL AND rendered_footer<>'') AS ftr,
  count(*) FILTER (WHERE rendered_head   IS NOT NULL AND rendered_head  <>'') AS head,
  count(*) AS total_pages
FROM pages;
```
Expect hdr=ftr=head=0. [RESULT: pending — run before final sign-off.]

## Target B verification — is rendered_html ever NOT reproducible from html_template?

The design says rendered_html is derived. The counter-claim (save_page_sections
:276-285) is that interactive tools "exist ONLY as rendered_html". Find any live
tool page_component whose rendered_html is materially larger than / diverges from
its component's html_template (i.e. rendered_html carries bytes the template does
not):

```sql
-- tool components: compare rendered_html length on the page vs the durable template
SELECT pc.page_id, cc.name, cc.function,
       length(pc.rendered_html) AS rh_len, length(cc.html_template) AS tmpl_len,
       round(100.0*length(pc.rendered_html)/NULLIF(length(cc.html_template),0)) AS pct
FROM page_components pc
JOIN content_components cc ON cc.id = pc.component_id
WHERE cc.function IN ('interactive_tool','tool','game')   -- refine to real tool functions
  AND pc.rendered_html IS NOT NULL
ORDER BY rh_len DESC
LIMIT 40;
```
Interpretation: if rh_len ≈ tmpl_len (+ the fixed tool-doc header) the render is
an identity over the template ⇒ recoverable ⇒ no separate guard needed. If rh_len
>> tmpl_len for real tools, rendered_html holds durable content the template does
not, and IS a genuine unguarded durable source. [RESULT: pending.]

## site_components.rendered_html — the chrome writers (open question)

If the chrome (header/footer/head, keyed site_id+slot_name) is written from a
whole LLM artifact, it shares the 012 shape and is a real INSTANCE-1 target. Find
the writers:
```
grep -rn "site_components" platform/ --include=*.go | grep -i "INSERT\|UPDATE\|rendered_html"
```
[RESULT: pending — being classified by agent 1.]

## Build / deploy (from CLAUDE.md — for when/if code ships)

- `make build-agent-chassis` builds from committed HEAD. Commit the task first.
- Bump `IMAGE_TAG` (makefile ~line 16) every build.
- Verify against the running pod, never git:
  `kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<symbol I CREATED>"'`
  Grep a literal the change CREATED, plus a positive control — not one it merely uses.

## The scratch one-step probe (INSTANCE 1's harness, generalised 2026-07-25)

Exercises ANY `IsLocal` action against the live binary in seconds, with a payload
you control — the only way to induce a fault the real dispatch path would
intercept upstream. Used for INSTANCE 1 (`create_tool_component`, 07-23) and
INSTANCE 2 (`complete_work_item`, 07-25).

1. Seed a one-step `agent_definitions` row (`is_active=true`,
   `is_snapshot=false`, `image_tag` = the LIVE tag), whose workflow is
   `{start_step: <your action>, next: complete_workflow}` and whose step config
   reads its inputs as `"<field>": "input_data.<field>"` — dot paths resolve
   against `collected_data` (`resolveConfigPath`).
2. Fire the 091 kcat envelope at `system.agent.generic.requests` with
   `{"action":"orchestrate","config":{"agent_type":"<scratch type>"},"input_data":{…}}`.
3. Read the outcome from `orchestration_states` + whatever the action writes.
4. **Clean up**: the scratch `agent_definitions` row, the `orchestration_states`
   (+ audit) rows, any `agent_error_log` rows (else the immune-system sweep
   triages a deliberate test as a real failure), and the fixtures. Leak-check with
   a `UNION ALL` of counts and require **0 on every line** — a single `SELECT` per
   table invites you to stop after the first one that looks clean.

**Fixture gotchas paid for (2026-07-25), so the first `INSERT` works:**
- `site_work_items.pipeline` is **NOT NULL** with no default (`design` is a safe
  value for a design-domain check). `\d site_work_items` before writing the
  insert — schema first, per CLAUDE.md, and it costs a round trip otherwise.
- `created_by` is NOT NULL too. Use a distinctive literal
  (`'scratch-021-check'`) — it is what every cleanup and leak-check query keys on,
  and it beats trying to remember the UUIDs.
- Give the fixture its own `item_key`, not the check's real one: `idx_swi_dedup`
  is UNIQUE on `(site_id, item_key)` for any non-terminal status, so a real open
  item on that site will otherwise reject your insert with a 23505.
- Use literal, recognisable UUIDs (`00000021-0000-4000-8000-…`) rather than
  `gen_random_uuid()`. You will be pasting them into a dozen ad-hoc queries.

**Containing a fixture that the live fleet can see.** A `site_work_items` probe
fixture has to sit on a REAL site (`site_id` is NOT NULL and the predicate needs
it), and a completion refusal releases the claim — status `triaged`, `claimed_by`
NULL — which makes it dispatchable to a real handler on a real production site.
Give the fixture `handler_agent = '<something>-nonexistent-agent'` and a distinct
`item_key`: the dispatch loop picked mine up within 5 seconds of the refusal and
parked it `blocked` ("Handler agent not registered: …") instead of doing anything
to robot-hands.com. Cheap, and it is the difference between a contained test and
an unplanned production edit. Delete the fixture as soon as you have read the
outcome — don't leave it over a break.

**The payload MUST be ONE line — `kcat -P` publishes one message PER LINE, and a
fragmented payload fails GREEN** (contributed 2026-07-26 by the `bugs_open/015`
thread, which lost a probe round to it). A pretty-printed or heredoc-wrapped JSON
`input_data` is published as several messages; with `-c 1` only the *first*
fragment is sent, which is invalid JSON. The chassis does **not** error visibly:
it creates an `orchestration_states` row carrying your `orchestration_id`, with
`input_data: null`, a fallback no-op `agent_config` ("No-op — scheduled task
pre_query already did the work"), **its own `orchestration_name`** (yours is
discarded), and marks it `COMPLETED`. So a poll keyed on *"is there a terminal
row for my orchestration_id"* returns success, the fixtures are untouched — and
that reads exactly like a probe whose refusal branch worked correctly. It is
indistinguishable from a real pass unless you open `collected_data`.

Three cheap defences, all worth having:
- Build the payload with `jq -cn …` and assert it is one line before firing
  (`[ "$(printf '%s' "$PAYLOAD" | wc -l)" -eq 0 ]`).
- **Grade the probe on `collected_data-><your step name>`, never on
  `status='COMPLETED'`.** The step key is absent entirely when the action never
  ran; that absence is the discriminator.
- **Clean up by `orchestration_id`, never by `orchestration_name`.** The chassis
  assigns its own `generic-orchestrate-MMDD-HHMM` name and discards the one you
  set in the header, so a name-keyed cleanup
  (`WHERE orchestration_name LIKE 'scratch…%'`) matches NOTHING and silently
  leaves every probe row behind. `orchestration_id` is the header that survives.

  > **CORRECTED within the hour, 2026-07-26.** I first wrote this bullet as "if
  > the chassis replaced your `orchestration_name`, your message did not parse" —
  > i.e. offered the name as a discriminator for the fragmented-payload failure.
  > **That is wrong.** The very next probe parsed perfectly, ran the action and
  > returned its result, and its name was *also* replaced
  > (`generic-orchestrate-0726-1359`). The name is replaced unconditionally, so it
  > separates nothing. Caught by reading the field on a known-GOOD run instead of
  > only on the known-bad one — which is the general lesson: **a discriminator
  > claimed from the failing case alone is not a discriminator until the passing
  > case disagrees with it.** The `collected_data-><step>` check above is the real
  > one, and it was confirmed against both.

**A fired probe with no `orchestration_states` row is almost always QUEUED, not
lost.** Cost this session ~20 minutes and one wrong inference: nothing happened
for 9 minutes, so I blamed the `kubectl run -i` stdin race, added `-c 1` and
re-fired — then *both* messages ran, 10 minutes later, seconds apart. The `-c 1`
changed nothing; the consumer was simply wedged. Diagnose in this order, and do
not re-fire until step 2 says the message is missing.

**Run the script that already answers this** — `scripts/dispatch-queue-depth.sh`
(added 2026-07-25 by the `bugs_open/030` thread) prints consumer position, queue
depth and an explicit *"QUEUED, not lost — do NOT re-fire"* verdict. `097` and
`090` call it automatically; **`091` and `092` do not**, so a probe copied from
`091` (like this harness) has to run it by hand. The two manual checks below are
what it automates — keep them for when you need the raw values.

1. **Is the message on the topic?** If it is, the produce worked and the trigger
   is not your problem:

```
kubectl -n kafka run -i --rm kcat-read-$(date +%s) --image=edenhill/kcat:1.7.1 \
  --restart=Never -- kcat -C -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 \
  -t system.agent.generic.requests -o -6 -e -q
```

2. **Has the CONSUMER read it?** One consumer, serialised (`bugs_open/030`), so a
   16-seat council in flight freezes every other dispatch behind it for its whole
   duration — on 2026-07-25 that was ~15 minutes with `CURRENT-OFFSET` completely
   frozen and `LAG` climbing:

```
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```
A frozen `CURRENT-OFFSET` with rising `LAG` is the stall; it drains on its own.
(Ignore `liveness-test-throwaway-group` — an abandoned group with permanent lag.)

## Know the expected verdict BEFORE firing a live verifier probe

A probe only discriminates if you know what it *should* say. For INSTANCE 2:
dump the verifier's population and run a **verbatim copy** of the shipped
predicate over it locally (stdlib-only `go run` in a scratch dir — do NOT add a
throwaway test to the shared tree):

```sql
SELECT row_to_json(t) FROM (
  SELECT s.domain, p.name AS page, COALESCE(pc.slot_name,'') AS slot,
         COALESCE(pc.rendered_html,'') AS html
  FROM page_components pc JOIN pages p ON pc.page_id=p.id JOIN sites s ON s.id=p.site_id
  WHERE pc.locked_at IS NULL
    AND pc.rendered_html ~ 'background(-color)?:\s*#[0-9a-fA-F]{3,8}'
    AND pc.rendered_html LIKE '%<style%') t;
```
`row_to_json` escapes newlines, so each component is exactly one line — a plain
scanner reads it. This is what turned "fire and see" into a discriminator pair:
one site where the transform still bites (must REFUSE) and one where the
detector matches but the transform does not (must PASS).
