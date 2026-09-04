# RUNBOOK — bugs_open/450 tool page shells

Every command here was needed at least once and had a gotcha attached. Change it HERE when it
changes.

## 1. The shell census (is the bug still live, and where)

**⚠ THE SOURCE OF TRUTH IS `toolShellPredicateFor` IN `platform/orchestration/actions/owned_page_guard.go`, NOT THIS BLOCK.**
The SQL below is a **copy** of that function's `NOT EXISTS`, and it is a copy on purpose: the
question "which pages does the guard refuse?" can only be answered by the guard's own predicate.
**Before trusting this query, diff it against that function** — it is ~10 lines and pasting it
costs less than writing a version of it.

That rule was earned three times in one day, twice by me and once by the `bugs_open/427` lane, who
put it best: *when the thing being measured IS a mechanism, copy its predicate — do not paraphrase
it.* Every paraphrase produced a **floor delivered in the voice of a total**, and in each case
nothing in the result could have revealed the error — the numbers all looked reasonable. The
specific trap here is `cc.is_active`: leave it out and 9 pages across 5 sites, whose tool component
exists but is deactivated, silently read as "has a tool" while the guard refuses them.

**Use the GUARD'S OWN predicate, split by publication state.** The version first written here (and
in the bug file) was a floor twice over — see the warning below.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tA <<'EOF'
SELECT s.domain,
       count(*) FILTER (WHERE p.deployed_at IS NOT NULL) AS shipped,
       count(*) FILTER (WHERE p.deployed_at IS NULL)     AS never_shipped
  FROM pages p JOIN sites s ON s.id=p.site_id
 WHERE p.page_type='tool' AND p.status='active'
   AND NOT EXISTS (SELECT 1 FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
                   WHERE pc.page_id=p.id AND pc.build_status<>'removed'
                     AND cc.component_level='tool' AND cc.is_active)
 GROUP BY 1 ORDER BY 2 DESC, 3 DESC;
EOF
```

**67 pages / 16 sites `[MEASURED 2026-09-03 ~12:0xZ]`** — of which **48 are already
`rebuild_policy='owned'`** and were refused before this lane existed, so the guard's genuinely NEW
population is **19**.

⚠ **THE FIRST VERSION OF THIS QUERY WAS WRONG IN TWO DIRECTIONS, and both are general traps:**

1. **`deployed_at IS NOT NULL` cannot see a page that never shipped** — which is the sectionless
   fork (bugs_open/450 §7), i.e. the census excluded the very variant it exists to measure. +4.
2. **It did not test `cc.is_active` while the FIX does.** A page whose only tool component is
   inactive read as "has a tool" to the census and as a shell to the guard. +9, across four sites
   that did not appear in the old census at all.

**The lesson, not the numbers: run your FIX's predicate as the census.** A fix and its denominator
disagreeing is invisible while both look reasonable. Re-running the old query reproduced its
number and read as confirmation — a census reproduces the *question you encoded*, so a repeat
certifies nothing.

⚠ **It is a repair-INITIATED count, not repair-COMPLETED** (portfolio_positioning lane, 2026-09-03).
Attaching a tool component removes a page immediately, while the public keeps seeing prose until
the rerender drains — seotools left the census at 10:27Z with **0 of 7** pages published. A later
reader will see "seotools: clean" on a site serving seven prose pages. **Acceptance is the served
body (§3), never this census.**

⚠ Still an upper bound for the adopted estate: a ported tool can live inline in a non-tool-level
component, so it counts as a shell here and is not one. The mechanism is proven only where
`page_component_history` names the writer (query 2).

## 2. Who wrote a suspected shell

```sql
SELECT p.name, w.item_type, count(*) FROM page_component_history h
  JOIN pages p ON p.id=h.page_id
  LEFT JOIN site_work_items w ON w.id=h.source_item_id
 WHERE p.site_id=:site AND p.name LIKE 'tool-%' GROUP BY 1,2;
```

`unbuilt_internal_link` in the `item_type` column is this bug. ⚠ `site_work_items` is a rolling
window — a closed row is archived out of it, so the LEFT JOIN yields NULL for old writes and
undercounts. Absence of the item type is not absence of the mechanism.

## 3. At the body — a tool is a FORM, never a size

```bash
for t in <slugs>; do
  b=$(curl -s "https://<domain>/tools/$t/?cb=$RANDOM")
  printf "%-34s forms=%d inputs=%d selects=%d\n" "$t" \
    "$(grep -o '<form' <<<"$b"|wc -l)" "$(grep -o '<input' <<<"$b"|wc -l)" "$(grep -o '<select' <<<"$b"|wc -l)"
done
```

⚠ **Always probe a known-real tool in the same run as the control** — advertise.co.uk's three read
1 form / up to 11 inputs. Size, status, headline and a 200 all pass on a prose shell; only the form
count discriminates. The cache-buster is not optional (the CDN will serve you a stale body and it
looks like a result).

## 4. The 090 verdict for this bug

```sql
SELECT result FROM site_work_items
 WHERE spec->>'dispatch_correlation_id' LIKE '96e97dc4%' AND item_type='needs_diagnosis';
```

⚠ **NOT in `doc_notes`** — the `doc_notes` query returns nothing for this run and reads as "no
verdict". The verdict lives in the item's `result`. Output is ~45 KB; pipe it through
`python3 -c "import json,sys; print(json.load(sys.stdin)['response']['response']['conclusion'])"`
rather than reading it raw.

## 5. Ownership, before routing anything at a bug

```bash
python3 scripts/who-owns.py <number|slug>
```

⚠ It reads COMMITS, so a session mid-fix is invisible; check the tree too, and re-run at each
phase boundary (the answer goes stale in minutes on this tree). For 450 it names the FILING lane —
which owns the instance, not the class. Read the lane's handoff before concluding it owns the fix.

## 6. Build and prove the change

```bash
scripts/verify-head-builds.sh --with <file> [--test]      # BEFORE committing
scripts/verify-head-builds.sh [targets]                   # AFTER committing
go test ./platform/orchestration/actions/ -run 'ToolShell|GenericBuild|OwnedPage'
```

⚠ Never hand-roll `git archive HEAD | tar` — that recipe is why the machine runs out of space.
⚠ `/tmp` is a 16 GB tmpfs; a full one reports as `link: mapping output file failed: no space left
on device`, which reads like a compiler fault and is not one.

## 7. Prove it shipped (after the next roll)

```bash
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <commit-sha> <stamped-sha> && echo SHIPPED
```

⚠ The provenance line is a STARTUP line and scrolls out of reach on a busy service — an empty
result means "not in range", not "unstamped". Fall back to the binary probe with **both** controls
(a sha that must be present and one that must be absent):

```bash
kubectl -n ai-persona-system exec <pod> -- grep -aq "<expected-sha>" /proc/1/exe
```

Never `strings` (absent from the image) and never a discovery grep for "some 40-hex string" (it
matches Go's internal digit table and returns the same wrong answer on every service).
⚠ Read the stamp of the **service you mean** — one release can straddle several commits.

## 8. The demand control (a post-fix zero is not evidence)

After the roll, the fix's *positive* signal — not the absence of new shells:

```sql
-- a queued item at a shell target should now terminate wont_fix with a receipt
SELECT w.item_type, w.status, w.result->>'reason'
  FROM site_work_items w JOIN pages p ON p.id = w.page_id
 WHERE w.item_type='unbuilt_internal_link' AND p.page_type='tool'
 ORDER BY w.updated_at DESC LIMIT 20;
-- and the receipt itself
SELECT summary, spec->>'refusal_class' FROM site_work_items
 WHERE item_type='owned_page_review' AND spec->>'refusal_class'='tool_pending'
 ORDER BY created_at DESC LIMIT 10;
```

Positive control in the same window: a **non**-shell page building normally. Without it, a
zero means "nothing tried", not "the guard held".

## 8b. ⚠ Do NOT verify this fix with a re-render (until 9831e9ab4 rolls)

Flagged by the `bugs_open/427`/`454` lane 2026-09-03, and it would have cost this lane a day if
we had reached for it: since 2026-09-02 **every light re-render renders the page's own stored
`content_data` back at itself** with no freshly resolved data (`classifyStoredSection` dropped its
section plan). The run reports clean, the `rerendered` count is healthy, and nothing was
delivered. Both live agent-chassis builds still carried the defect at 09:54Z.

So a check of the form "re-render the page and read the result to see whether the fix took" is
reading a mirror. The verification in §8 deliberately reads **work-item terminal status +
receipt + the served body**, never a re-render, and must stay that way until a chassis image
carrying `9831e9ab4` is live. Full account: `bugs_open/454`.

## 9. Council submission

```bash
DRY_RUN=1 ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <sub.json>
./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <sub.json>
```

Save the printed `SUBMISSION_CORR`; budget ~30 minutes (dispatch queues behind the fleet — 29
minutes publish→start is normal). Find the run by payload, never by the printed id:

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = '<SUBMISSION_CORR>';
```

⚠ Do not submit within ~300 s of a chassis roll, and expect a roll to kill an in-flight council run.

## 10. Applying migration 729 (the plan-side gate) — NOT YET, and the precondition is not the usual one

**Do not apply it because the guards pass.** They do; that is not the gate on applying it. The
replacement prompt text tells the planner *"validation holds back tool pages whose tool does not
exist"*, and that is **FALSE until a chassis carrying `5e6fee47b` rolls**. Applying early ships a
prompt asserting a validation that is not running. The KEY alone would be order-safe (old binaries
ignore it); the SENTENCE is not, so both wait.

**Preconditions, in order:**

1. Council verdict on corr `4e7497ed-62ed-4426-a814-8361754c2352` read (REVISE → revise first).
2. A chassis image carrying `5e6fee47b` is LIVE — checked at the artefact, per §7 above:
   `git merge-base --is-ancestor 5e6fee47b <the service's build-provenance stamp>`.
3. Then, and only then:
   ```bash
   kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
     psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 \
     < docs/agent_docs/sql_for_agents/729_planner_tool_source_gate.sql
   ```
   ⚠ **This file only.** Never an unscoped migration-runner `--apply`: it takes EVERY pending file.

**Rehearsing a change to it first** (this is how 720's lane found its own guard arithmetic was
wrong, and it is worth repeating for any edit):

**Use the SELF-BASELINING form. Do not hardcode a "before" md5** — the first version of this
recipe did, and within an hour the live template had moved (another lane landed a prompt
migration), so the recorded value read as a failure when nothing was wrong. Capture the baseline
inside the same transaction and let the query answer `byte_exact` for you:

```bash
sed -e 's/^BEGIN;$//' -e 's/^COMMIT;$//' docs/agent_docs/sql_for_agents/729_planner_tool_source_gate.sql          > /tmp/729_body.sql
sed -e 's/^BEGIN;$//' -e 's/^COMMIT;$//' docs/agent_docs/sql_for_agents/729_planner_tool_source_gate_ROLLBACK.sql > /tmp/729_rb_body.sql
{ echo "BEGIN;"
  echo "CREATE TEMP TABLE _base AS SELECT md5(default_config #>> '{workflow,steps,plan_site,config,prompt_template}') AS before_md5
          FROM agent_definitions WHERE type='build-site-planner' AND is_active
           AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;"
  cat /tmp/729_body.sql          # apply
  cat /tmp/729_rb_body.sql       # then unwind
  echo "SELECT b.before_md5 = md5(a.default_config #>> '{workflow,steps,plan_site,config,prompt_template}') AS byte_exact
          FROM _base b, agent_definitions a
         WHERE a.type='build-site-planner' AND a.is_active
           AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL;"
  echo "ROLLBACK;"
} | kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -v ON_ERROR_STOP=1
```

Proven 2026-09-03 (twice, at two different baselines): every guard and both verify blocks pass,
`byte_exact = t`, `enforce_tool_sources` ends ABSENT and 720's `enforce_listing_sources` stays
`true`. Drop the `_rb_body` line to rehearse the apply alone.

**The neighbour checks are the point of the verify blocks.** 729 refuses if 720's listing
sentence, 720's flag, 433's directory rule, 718's imagery surface **or 640's rule 17**
(`may also carry a "subject"`) has gone missing; the ROLLBACK asserts the same. Rule 17 was added
at the `bugs_open/443` lane's request and confirmed against the live row before hardcoding —
their `REPEATED_COMPONENT_BUILT_WITHOUT_SUBJECT` detector is live-firing, and its fire-rate is
only interpretable if that sentence is still in the prompt. **If you add a surface to this row
worth defending, add its literal here too**: a migration that quietly ate a neighbour's sentence
would otherwise look like a clean apply and produce a wrong conclusion in someone else's lane.

**Unwinding:** `729_ROLLBACK` first, then `720_ROLLBACK` if you also want the listing gate gone.
While 729 is applied, 720_ROLLBACK's anchor does not appear verbatim and it refuses by its own
design — that is correct behaviour, not a broken file.

## 11. Proving the plan-side gate actually fired (after it is applied)

The gate's positive signal, not the absence of shells:

```sql
-- held pages: one deferred capability_gap per held tool page
SELECT s.domain, w.summary, w.spec->>'builder_needed', w.created_at
  FROM site_work_items w JOIN sites s ON s.id = w.site_id
 WHERE w.item_type='capability_gap' AND w.item_key LIKE 'capability_gap:tool:%'
 ORDER BY w.created_at DESC LIMIT 20;
-- and its durable audit row
SELECT created_at, message FROM agent_error_log
 WHERE error_code='TOOL_PAGE_HELD_NO_TOOL_SOURCE' ORDER BY created_at DESC LIMIT 10;
```

⚠ **A capability_gap here is NOT a defect report about the tool pipeline.** Holding a planner tool
stub starves nothing (bugs_open/450 §7). Treat a rising count as the gate working, and only
investigate if a HELD page's tool later turns out to have existed — which would mean the census or
the candidate-name list is wrong, and is the failure mode to watch for.

**The control that makes the above mean something:** a site whose tools DO exist must plan its tool
pages normally and file NO gap. Without that in the same window, zero gaps is indistinguishable
from the gate never running.

## §11 — the orphaned-`component_id` census, and the repair (bugs_open/479)

**The census. Do NOT join `content_components` on `component_id` to find these** — that is the
whole point: with the id NULL the join drops the row, so a page serving a 20 KB tool reads as
having no tool component at all. Match on `slot_name`.

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c "
SELECT s.domain, p.name AS page, pc.slot_name, length(pc.rendered_html) AS bytes, pc.created_at
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.component_id IS NULL AND pc.build_status <> 'removed'
 ORDER BY pc.created_at DESC;"
```

`[MEASURED 2026-09-04]` 17 rows / 7 sites. **Date every quote of this number** — it grows by
addition and it grew 11 → 17 in eleven days.

**The exposure population (what Layer 2 would preload), and how to get the predicate RIGHT.**
`interactiveHTMLSQL` assembles its SQL in Go from two marker slices; transcribing it by hand is
how §1's rule gets broken. Print it from the function instead:

```bash
cat > platform/orchestration/actions/zz_tmp_print_test.go <<'GO'
package actions
import ("fmt";"testing")
func TestZZTmpPrint(t *testing.T){ fmt.Println(interactiveHTMLSQL("pc.rendered_html")) }
GO
go test ./platform/orchestration/actions/ -run TestZZTmpPrint -v | sed -n '/ILIKE/p'
rm -f platform/orchestration/actions/zz_tmp_print_test.go   # ⚠ shared tree — delete it, or another session commits it
```

Then `WHERE pc.build_status='deployed' AND <that boolean>` → `[MEASURED 2026-09-04]` **378 rows /
371 pages**. That is an UPPER BOUND on exposure, not a prediction: the slot must also fail to
match the incoming section set.

**The repair:** `REPAIR_479_reattach_orphaned_tool_component_ids.sql` in this directory. It ships
with `ROLLBACK` on the last line — **rehearse, read the NOTICEs, then change that one word to
`COMMIT`.** Rehearsed 2026-09-04: `UPDATE 5`, both guards passed, bytes untouched.

Three things about it that were each paid for:

- **it refuses rather than guesses.** A row resolving to more than one free candidate aborts the
  whole transaction. `[MEASURED 2026-09-04]` 3 of the 5 tool rows have 2–3 active same-function
  components (tools are forked across sites under compounded names), so the "one active match on
  `cc.function`" recipe in the 450 handoff is NOT safe as written.
- **it compares the rows it TOUCHES, before against after** — never a population digest. The
  portfolio lane's first verify used `rendered_html_digest IS DISTINCT FROM md5(rendered_html)`
  and falsely refused: `IS DISTINCT FROM` convicts a NULL digest, and a NULL digest is normal
  (206 of 3,220 rows fleet-wide).
- **every check is `DO`/`RAISE`.** A verify block of bare `SELECT`s cannot stop a `COMMIT` —
  `ON_ERROR_STOP` ignores a non-empty result set.

**No re-render afterwards** (§8b). The served bytes are already correct; the repair touches
`component_id` only, and a re-render is the thing these rows are being protected FROM.
