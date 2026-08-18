# RUNBOOK — bugs_open/293, the whole-page shrink floor's axis

Every command here was hard to get right once. The gotcha is attached to the command, not kept
in someone's scrollback.

## Re-run the calibration (this is the thing `section_visible_text.go` says to do and nobody could)

Three exports, then the harness. The harness lives in the repo —
`platform/orchestration/actions/shrink_axis_calibration_test.go` — and SKIPS unless
`SHRINK_CALIBRATION_JSONL` is set, so `go test ./...` is unaffected.

```bash
D=docs/agent_docs/docs024_key_docs_latest/bugfix_293_whole_page_shrink_axis
S=<scratch>                     # 11 MB of live page HTML — never in the repo

$D/export_pairs.sh        $S/pairs_delete_v2.jsonl    # 1,079 EXACT rebuild pairs
$D/export_overwrite.sh    $S/pairs_overwrite.jsonl    #   263 section-editor pairs (positive control)
$D/export_intermediate.sh $S/pairs_intermediate.jsonl # 2,454 weak-join pairs (refusal HUNTING only)
```

⚠ **Chunk the export; do not `json_agg` it.** A single 2 MB row dies mid-stream with
`unexpected EOF` and the harness then fails on a parse error that looks like a bug in the harness
(`NOTES_shared_template_write.md`, 2026-08-17). All three scripts page 10–20 rows at a time.
`export_intermediate.sh` takes ~10 min and needs backgrounding.

⚠ **The tree may not build** because other lanes have WIP in it. Test against committed HEAD:

```bash
H=$S/headtree; rm -rf $H; mkdir -p $H; git archive HEAD | tar -x -C $H
cd $H && SHRINK_CALIBRATION_JSONL=$S/pairs_delete_v2.jsonl \
  go test ./platform/orchestration/actions/ -run 'TestShrinkAxis|TestMinimumSweep|TestPageTotal' -count=1 -v
```

## The controls, and why each one is there

| test | what it would catch |
|---|---|
| `TestCalibrationAxisMatchesGuard` | the harness measuring a THIRD axis nobody ships — it asserts the calibration's tag-stripped measure IS `strippedIncomingBySlot`'s, and that the two axes actually disagree on a CSS-for-prose swap. Runs in the normal suite, no export needed |
| `TestShrinkAxisBlindness` controls | a hollower that quietly did nothing (asserts 0 visible chars left), sections with no prose counted as "protected" (excluded from the denominator), and an empty population (fails rather than reporting a comforting zero) |
| `TestMinimumSweep`'s tail assertion | the sweep's replicated arithmetic drifting from `evaluateSectionShrink` — it must reproduce the shipped decision exactly at the shipped minimum |

## The join, and how to check it is still exact

```sql
-- Disconfirming control: NO live row may be older than the last archived delete of its own
-- (page, slot). A wrong join does not produce a zero here.
WITH lastdel AS (
  SELECT DISTINCT ON (page_id, slot_name) page_id, slot_name, created_at AS del_at
  FROM page_component_history WHERE op='delete' AND slot_name IS NOT NULL
  ORDER BY page_id, slot_name, created_at DESC)
SELECT count(*) FILTER (WHERE pc.created_at < d.del_at) AS must_be_zero,
       count(*) FILTER (WHERE pc.created_at < d.del_at + interval '5 seconds') AS same_transaction
FROM lastdel d JOIN page_components pc USING (page_id, slot_name);
```

⚠ **Never join on `(page_id, slot_name)` alone.** Slot names repeat on 14 pages, so it is a
cartesian product and it manufactures pairs that never existed (it manufactured a refusal on the
first run of this calibration). Check before trusting any pairing:

```sql
SELECT count(*) FROM (SELECT page_id, slot_name, count(*) n FROM page_components
                      WHERE slot_name IS NOT NULL GROUP BY 1,2 HAVING count(*)>1) x;  -- 14
```

⚠ **A long gap means the guard never judged that pair.** On the intermediate export, attribute every
refusal by `gap_s` before calling it a false refusal: a rebuild deletes and re-inserts inside one
transaction (<5 s for 1,109 of 1,123), so a slot that was absent for 1,700 s to 93 hours was a DROP,
and the guard skips drops (`!present → continue`). 7 of the 8 intermediate refusals are this.

## Was the guard even live? (the censoring question, which decides what the population means)

```bash
git log -1 --format='%ad %s' --date=short 2da3e08e5   # per-slot shrink guard: 2026-08-02
```
```sql
SELECT min(created_at)::date FROM page_component_history WHERE op IS NOT NULL;  -- 2026-08-09
```
The guard predates the archive, so **every archived pair is a write the live guard ALLOWED** — which
makes the population right for measuring false refusals and useless for measuring true catches. That
is why the true-catch argument comes from the constructed wipe, not from history.

## Post-roll verification (artefact level, with a demand control)

> **CORRECTED 2026-08-17 — do NOT use the log recipe that stood here.** It read
> `logs "$POD" --tail=3000 | grep -m1 'build provenance'`, and on this service that returns **another
> lane's payload**: the chassis logs whole council/diagnosis envelopes and those quote the phrase, so
> the first match in time order was a seat's SQL check timestamped two hours after the pod started. No
> `--tail` or `--since` size fixes a content collision. Read the stamp from the BINARY, where nothing
> can imitate it, and take the value NEXT TO THE MARKER rather than grepping for a sha you already
> believe — a bare-sha grep returns "absent" for a stamped binary too, whenever the stamp is a
> different commit, which is the normal case. Full trap in `LANDMINES.md`.

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
STAMP=$(kubectl -n ai-persona-system exec "$POD" -- \
  sh -c "grep -oa 'buildinfo.GitCommit=[0-9a-f-]*' /proc/1/exe | head -1" | cut -d= -f2)
echo "$STAMP"; git log -1 --format='%h %ad %s' --date=short "$STAMP"
git merge-base --is-ancestor <this-lane's-commit> "$STAMP" && echo "in the image"

# THREE controls, and the third is the one people skip:
kubectl -n ai-persona-system exec "$POD" -- grep -aq "gqls/agentchassis" /proc/1/exe \
  && echo "the probe can read this binary at all"                    # sanity
git merge-base --is-ancestor HEAD "$STAMP" || echo "HEAD absent — CONTROL PASSES"   # must be absent
# ⚠ pick an ABSENT control that postdates the stamp. A commit that PREDATES it is an ancestor and
#   "fails" for the right reason — I burned a check on exactly that on 2026-08-18.
```
⚠ **Ask every pod, not one.** `-l app=agent-chassis` is two pods of the **22** running this image, and
`save_page_sections` does not run in either: page builds run in ephemeral `agent-build-dispatch-loop-*`
pods that are gone minutes later, taking their logs with them. **A log-based proof of execution lives
only as long as the pod that produced it** — which is why the checks below are DB-side.
```sql
-- The refusal, when one happens: the message names the axis, so the row is greppable
SELECT created_at, spec->>'reason' FROM site_work_items
 WHERE item_type='needs_human_review' AND spec->>'reason' ILIKE '%VISIBLE text%'
 ORDER BY created_at DESC LIMIT 5;

-- DEMAND CONTROL — a zero above means nothing unless whole-page rebuilds are actually running
SELECT count(*) AS rebuild_writes, max(created_at) FROM page_component_history
 WHERE op='delete' AND created_at > now() - interval '24 hours';
```

## INDUCED REFUSAL — proven at the artefact 2026-08-18 on `v1.0.1309`, and safe in BOTH branches

The point of this recipe is that it cannot damage a page. **The payload is the page's own sections,
byte-for-byte**, so a refusal writes nothing and a *failure* to refuse writes back identical content.
Nothing else about it needs trusting.

⚠ **Do NOT induce by shrinking a real page's content on this path.** `save_page_sections` is
DELETE-then-INSERT: if the guard does not fire, the page is left holding whatever you sent. That is the
damage of `bugs_closed/285`, which is what this guard exists to prevent.

⚠ **Do NOT induce by raising a floor on a live `agent_definitions` row and waiting.** A refusal fails
the step, and none of the three build loops sets `continue_on_error`, so it can strand every page after
it in that loop. It also edits shared config other lanes depend on.

```bash
S=<scratch>; PAGE=<page uuid>; SITE=<site uuid>
# 1. SNAPSHOT first — this is both the payload and the before/after evidence.
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -At -c "
SELECT json_agg(row_to_json(t) ORDER BY t.position)::text FROM (
  SELECT id::text, slot_name, position, COALESCE(component_id::text,'') AS component_id,
         rendered_html, COALESCE(content_data,'{}'::jsonb) AS content_data, build_status
  FROM page_components WHERE page_id='$PAGE' ORDER BY position) t;" > $S/before.json
```
Build a ONE-LINE JSON body (python; `separators=(",",":")`) — the step config is the whole trick:

```
config.workflow = {start_step:"induce_save", steps:{
  induce_save:{action:"save_page_sections", next_step:"complete", config:{
     page_name_field:"input_data.page_name",         # override the default current_page.name
     site_id_field:"input_data.site_id",              # override site_record.site_id
     sections_metadata_field:"input_data.sections",   # naming this avoids the ambiguous-caller path
     page_total_text_floor:0.95 }},                   # see the NOTE below
  complete:{action:"complete_workflow"}}}
input_data = {page_name, site_id, sections:[{stored_slot_name, rendered_html, component_id, content_data}, …]}
```
`stored_slot_name` is taken VERBATIM (`bugs_open/189`) — use it, or the save renames the slot and the
lock/decision matching misses the row it protects.

> **NOTE, changed 2026-08-18.** The original run used `page_total_text_floor: 1.5`, which refuses even
> identical content (100% kept < 150%) — perfectly safe, and it is how the refusal was first proven.
> That also exposed a missing clamp, now fixed: the floor is capped at 0.95 like its siblings, so
> **identical content is no longer refusable**. Repeatable recipe: keep the payload identical except
> drop ~10% of ONE slot's prose (e.g. delete a sentence), which puts the page total under 0.95. The
> worst case if the guard is broken is that the page loses one sentence, and you are holding the
> original in `before.json`.

⚠ **Two traps in BUILDING the reduced payload, both hit on 2026-08-18.**

1. **Pick a text run OUTSIDE `<style>`/`<script>`.** The longest run in a real section is usually CSS —
   728 of 1,062 live sections carry over half their tag-stripped "text" inside those blocks. A builder
   that truncates "the longest text run" removes 76 characters and moves the VISIBLE total by **zero**,
   producing a payload that is 100% kept and would be ALLOWED, i.e. WRITTEN. Exclude those spans first.
2. **Assert the arithmetic before dispatching, and refuse to send otherwise.** Compute the verdict the
   guard will compute — under both the clamped and unclamped floor — and abort unless BOTH are REFUSE:

```python
if not (after < before*0.95 and after < before*1.50):
    raise SystemExit("ABORT: both must refuse, or a non-firing guard could write. Not dispatching.")
```
That assertion, not care, is what stopped the bad payload above going out.

```bash
CORR=$(uuidgen); ORCH=$(uuidgen)
kubectl -n kafka run -i --rm kcat-293-induce-$(date +%s) --image=edenhill/kcat:1.7.1 --restart=Never -- \
  kcat -P -b personae-kafka-cluster-kafka-bootstrap.kafka.svc.cluster.local:9092 -t system.agent.generic.requests \
  -H correlation_id=$CORR -H orchestration_id=$ORCH -H message_type=request -H client_id=demo_client \
  -H action=process -H sender_agent_type=cli -H sender_agent_id=cli-user \
  -H responses_topic=system.agent.generic.responses < $S/induce293.json
```
⚠ **`kcat -P` exits 0 having sent nothing** — the orchestration row is the only proof of publication.

### The four checks, in the order that makes them meaningful

```sql
-- (a) it fired, in the shipped binary
SELECT status, current_step, error FROM orchestration_states WHERE orchestration_id='<ORCH>';
--    FAILED | induce_save | ...PAGE CONTENT REGRESSION REFUSED... 581 chars of VISIBLE text against 581
--    deployed across 3 sections (100% kept, floor 150%)...
```
**Read the numbers, not just the refusal.** 581 visible chars on a page holding 7,343 BYTES of HTML is
the whole finding: the retired axis would have counted the markup and seen no problem.

```sql
-- (c) the queue surface, with its OWN remedy sentence and not a sibling's
SELECT status, severity, summary, spec->>'fix' FROM site_work_items
 WHERE item_type='save_refused_incomplete' ORDER BY created_at DESC LIMIT 1;
```
**(b) THE ARTEFACT, which is the assertion** — re-run the snapshot query into `after.json` and compare
**the row ids as well as the bytes**: identical ids prove the DELETE never ran, where identical bytes
alone would also be consistent with a delete-and-reinsert of the same content.

**(d) THE ALLOW ARM, or a refusal-only test certifies a guard that refuses everything.** Do NOT
manufacture a successful whole-page save for this — it rewrites real rows. Take it from live traffic
instead: export the window's pairs and confirm the guard judged real writes and allowed them
(2026-08-18, `v1.0.1309`: 6 rebuild writes, 4 in scope on the live axis, 0 refused, and the only
refusal on the whole image was the induced one).

### Afterwards — CANCEL the synthetic row

It lands as `needs_human_review`, severity high, and it describes an event that did not happen. Leaving
it is a false alarm in an operator's queue.
```sql
UPDATE site_work_items SET status='cancelled',
  result = COALESCE(result,'{}'::jsonb) || jsonb_build_object('reason','SYNTHETIC - induced to prove …')
WHERE id='<the row>';
```
`cancelled` is in `workItemTerminalStatuses` (migration 157), so it releases the dedup slot — which is
what lets the induction be repeated.
