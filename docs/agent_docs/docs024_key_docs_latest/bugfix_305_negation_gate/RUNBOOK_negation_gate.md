# RUNBOOK — `bugfix_305_negation_gate`

Every command here was hard to get right once. The gotcha is attached to the command, not filed
somewhere else.

## 0. The regex rule that governs every query in this lane

**Postgres has no `\b`.** There `\b` is a *backspace character*; the word boundary is `\y`. A Go
pattern pasted into psql silently matches nothing and returns a confident zero
(`LANDMINES.md:4219`). Go's RE2 spells it `\b` and has no `\y`.

**So: prove the pattern before you quote the count.** One query, both arms:

```sql
SELECT (SELECT count(*) FROM regexp_matches(
          'not a demo, but a product', '\ynot (?:just )?[^.;:]{2,50},\s*but\y', 'gi')) AS must_be_1,
       (SELECT count(*) FROM regexp_matches(
          'we ship on Kubernetes and Kafka',   '\ynot (?:just )?[^.;:]{2,50},\s*but\y', 'gi')) AS must_be_0;
```

## 1. The distribution census (what the gate will cost)

`1,503` rows is one week of `page-content-writer`. Needs `SET statement_timeout` — the default kills it.
⚠ **Do not run this while a `090` diagnosis run is in flight** — `bugs_open/305 §4` records a loop's
data request dying with `statement_timeout` while its filer ran a heavy `llm_call_log` query.

```sql
SET statement_timeout='150s';
WITH c AS (
  SELECT id,
    (SELECT count(*) FROM regexp_matches(response_text, '[a-z\)"''],\s+(?:not|never)\s+(?:just\s+|merely\s+|simply\s+|only\s+)?[a-z]', 'gi')) AS x_not_y,
    (SELECT count(*) FROM regexp_matches(response_text, '\ynot (?:just |merely |simply |about )?[^.;:]{2,50},\s*but\y', 'gi')) AS not_x_but_y,
    (SELECT count(*) FROM regexp_matches(response_text, '\yrather than\y', 'gi')) AS rather_than,
    (SELECT count(*) FROM regexp_matches(response_text, '[.!?]\s+(?:It|This|That|They|These|We)\s+(?:doesn.t|does not|isn.t|is not|won.t|will not|can.t|cannot|aren.t|are not|don.t|do not)\y', 'g')) AS neg_reveal
  FROM llm_call_log
 WHERE agent_type='page-content-writer' AND success AND created_at >= '2026-08-13')
SELECT count(*) AS calls,
       count(*) FILTER (WHERE x_not_y>=1)                                   AS ge1_xny,
       count(*) FILTER (WHERE x_not_y>=2)                                   AS ge2_xny,
       count(*) FILTER (WHERE not_x_but_y>=1)                               AS ge1_nxby,
       count(*) FILTER (WHERE rather_than>=1)                               AS ge1_rt,
       count(*) FILTER (WHERE neg_reveal>=1)                                AS ge1_reveal,
       count(*) FILTER (WHERE x_not_y+not_x_but_y+rather_than >= 2)          AS ge2_family
  FROM c;
```

## 2. Verify at the ARTEFACT, never at `updated_at`

A rerender bumps `page_components.updated_at` without regenerating anything — that is what made the
complained-of copy look five days newer than it was (`bugs_open/305 §3`).

```sql
SELECT p.url, pc.slot_name,
       pc.content_data::text ~* '\w,\s+(not|never)\s+\w' AS x_not_y,
       pc.content_data::text ~* '\yrather than\y'         AS rather_than
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
 WHERE s.domain = 'ai-agent-orchestration.com'
   AND p.url ~ '(model-directory|adoption-tracker|protocol-tracker)'
 ORDER BY 1, 2;
```

⚠ `pages` has **no `slug` and no `path`** column — it is `name` and `url`. Two agents wasted a round
trip each on that.

## 3. Read the live writer workflow (before anchoring a migration on it)

`jsonb_pretty` minus the 12 KB prompt template, or the output is unreadable:

```sql
SELECT jsonb_pretty(
         (default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content}')
         #- '{config,prompt_template}')
  FROM agent_definitions
 WHERE type='page-content-writer' AND is_active
   AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

⚠ `jsonb - 'key'` fails with *"operator is not unique: unknown - unknown"* when both sides are
untyped literals in a `-Atc` one-liner. Use `#-` with a path array, as above.

## 4. Is the gate live? (three levels, in this order)

```bash
# a) the binary says what built it — per SERVICE, and the line SCROLLS
kubectl -n ai-persona-system logs -l app=agent-chassis --tail=300 | grep -m1 'build provenance'
git merge-base --is-ancestor <my-commit> <the-sha-from-that-line> && echo SHIPPED
```

```sql
-- b) the marker: did the gate actually run, and what did it do?
SELECT count(*) FILTER (WHERE collected_data::text LIKE '%__copy_gate%') AS runs_with_marker,
       count(*) AS orchestrations
  FROM orchestration_states
 WHERE created_at > '<roll timestamp>' AND collected_data::text LIKE '%page-content-writer%';

-- c) the retry rows. NOTE: the marker is present on SUCCESSFUL retries too, so filter failures on
--    success=false, never on a non-empty error_message (the bugs_open/119 precedent).
SELECT count(*) FROM llm_call_log WHERE error_message LIKE 'RETRY (bugs_open/305%';
```

⚠ **An empty `build provenance` grep means "scrolled out of range", not "unstamped"** — on a busy
service the line is gone within hours. Fall back to the binary probe with a control:
`kubectl -n ai-persona-system exec <pod> -- grep -aq "<expected-sha>" /proc/1/exe` and in the same
breath a sha that must be ABSENT. Never `strings` (not in the image).

## 5. Migration

`497` is written as `_HOLD` on purpose: it rewires the writer's step chain, so it must be applied
**after** the image carrying `rewrite_negations` is live, or the chain names a step the binary cannot
run. `SIDECAR_RE` excludes `_HOLD.sql` from the auto-apply sweep and still lists it.

```bash
./scripts/migration/run-migrations.sh              # DRY RUN — always first, per session
./scripts/migration/run-migrations.sh --apply      # takes EVERY pending file, not just mine
```

⚠ `--apply` applies **every** pending migration in the directory, not the one you are thinking about.
Read the dry-run list before running it, and expect other sessions' files in there. `_HOLD` is
excluded by `SIDECAR_RE` (`scripts/migration/run-migrations.sh:65`) and still listed, which is the
whole point of the suffix: it holds an ordering-critical file without hiding it. Renaming `497` out of
`_HOLD` is the deliberate act that says "the image is live".

## 6. Tests

```bash
go test ./platform/orchestration/datahelpers/ -run 'Negation|Voice|Strawman' -v
go test ./platform/orchestration/actions/ -run 'Negation|CopyGate|RewriteNegations' -v
go build ./... && go vet ./platform/... ./cmd/...
```


## 7. THE CANARY THAT MATTERS — run the shipped scanner over the owner's own three pages

This is the demand control for the whole gate, and it needs no roll: it runs the real
`datahelpers` functions over the real `page_components.content_data` with the real brief as the
exemption corpus. **It found a defect the unit tests could not** (the two-sentence reveal was being
attributed to the clean sentence before it), so run it again after any change to the shapes.

⚠ Write it into a SCRATCH copy of the tree, never into the repo: any `.go` file under the module root
joins the build, and a throwaway in `docs/` would break `go build ./...` for everyone. **`cmd/gatecanary`
is deliberately NOT a real command and must never become one** — it exists for the length of one
verification and is thrown away with the scratch tree. The pattern checker flags the path as a proposed
new capability surface; this paragraph is the answer to that flag.

> **⚠ CORRECTED 2026-08-24 — the recipe that stood here extracted the WHOLE TREE (459 MB measured)
> and never deleted it.** CLAUDE.md now names this class directly: 73 documents spell out
> `git archive HEAD | tar`, and **66 of them never delete anything** — which is why this machine keeps
> hitting `link: mapping output file failed: no space left on device`. This lane's own copy was one of
> the 66; I left a 459 MB tree behind this morning before reaping it. **The canary needs three
> packages, not the estate.** Targeted extract measured at **1.7 MB — 287x smaller — and it reproduces
> the full-tree result exactly** (`TOTAL 3 | exempt 1 | repairable 2`, verified 2026-08-24).

```bash
SP=<your scratch dir>/cy ; REPO=/home/ant/projects/agentchassis
trap 'rm -rf "$SP"' EXIT            # <-- the half the 66 documents omit
rm -rf $SP && mkdir -p $SP/gotmp

# Only what the scanner needs, from COMMITTED HEAD (never the working tree, which
# is the union of every session's WIP). If the import set grows, the go build
# below names what is missing — add it and re-run; one round was enough today.
git -C $REPO archive HEAD go.mod go.sum \
    platform/orchestration/datahelpers pkg/models platform/orchestration/types \
  | tar -x -C $SP

mkdir -p $SP/cmd/gatecanary && cat > $SP/cmd/gatecanary/main.go <<'GO'
package main
// SCRATCH-ONLY. Per component, per string field, per hit:
//   datahelpers.WalkContentStrings -> ScanDefineByNegation -> NegationExempt(hit, brief)
//   + IsHeadlineField(path)  (a headline hit is repaired regardless of budget)
// Import is platform/orchestration/datahelpers -- NOT platform/datahelpers.
// The brief corpus is every string in the site's current site_specs rows,
// flattened recursively (679 strings on aiao, 2026-08-24).
GO

cd $SP && TMPDIR=$SP/gotmp go run ./cmd/gatecanary <components.json> <brief.json>
```

⚠ **`TMPDIR` must point at disk, not `/tmp`** — `/tmp` is a 16 GB **tmpfs (RAM)**, and a full one
presents as `link: mapping output file failed: no space left on device`, which reads like a compiler
fault and is not one.

⚠ **Getting the two inputs is where the time goes.** `content_direction` is a column on **`pages`**,
not `sites`; the site brief lives in **`site_specs`** keyed on **`aspect`** (not `spec_type`), and you
want `WHERE is_current`. Dump both with `jsonb_pretty(jsonb_object_agg(...))` via
`kubectl exec … psql`, so the Go program needs no DB connectivity at all:

```sql
-- brief.json
SELECT jsonb_pretty(jsonb_object_agg(ss.aspect, ss.data)) FROM site_specs ss
  JOIN sites s ON s.id=ss.site_id WHERE s.domain='<domain>' AND ss.is_current;
-- components.json
SELECT jsonb_pretty(jsonb_agg(jsonb_build_object('url',p.url,'slot',pc.slot_name,
       'updated_at',pc.updated_at,'content_data',pc.content_data)))
  FROM pages p JOIN sites s ON s.id=p.site_id JOIN page_components pc ON pc.page_id=p.id
 WHERE s.domain='<domain>' AND p.url ~ '(model-directory|adoption-tracker|protocol-tracker)';
```

**Two readings, and the DIFFERENCE is the point.** The first is the baseline this lane started from;
the second is where the gate has got the pages to. ⚠ **Do not read the baseline as "what working looks
like" any more** — it was, on 2026-08-20, and quoting it today would report a fixed gate as broken.

`[MEASURED 2026-08-20]` — the baseline, before the gate reached these pages:
```
TOTAL 7 | exempt (brief-supplied or regulatory) 1 | repairable 6, of which headline-class 6
```

`[MEASURED 2026-08-24]` — after the gate, the ceiling fix (`569`) and the site's own rebuilds:
```
/adoption-tracker.html     TOTAL 1 | exempt 1 | repairable 0
/protocol-tracker.html     TOTAL 2 | exempt 0 | repairable 2
ALL THREE                  TOTAL 3 | exempt 1 | repairable 2
```
`model-directory` does not appear because it has **ZERO** hits — it is the page that carried both
sentences the owner quoted. `protocol-tracker`'s 2 are the whole remaining residual of this bug, and
they are waiting on a rerender that is **already queued** behind another lane's `claims_unverified`
item; they are NOT evidence of a gate defect.

- both sentences the owner quoted are REPAIRABLE — *"The registry shows you what's possible, not what
  survives production."* (`x_not_y`, headline) and *"It doesn't tell you how they hold up under real
  Kafka throughput…"* (`negative_reveal`, subheadline);
- the canonical tagline on `adoption-tracker`'s hero is **`exempt:brief_supplied_sentence`** — the
  designed behaviour, and the reason this fix does not clean those pages.

If the exempt count goes to 0, the brief changed (or the exemption broke) — check which before
celebrating. If the reveal's sentence comes back as *"A model directory tells you which agents
exist."*, the attribution regression is back.

## 8. THE POST-548 PERSISTENCE CHECK — one query, two-sided, and it defeats the truncation trap

Added 2026-08-22, after `548` pointed `render_section` at `copy_gate.result`. This is the query that
answers "did the repair reach the page?" **without** reading the marker's `from`/`to`, which truncate
at 160 chars and share their opening (§20/§22 both wasted time on that).

It compares the two **durable** content fields and then asks which of them the stored component
matches. Substitute the domain and page:

```sql
WITH run AS (
  SELECT collected_data cd FROM orchestration_states
   WHERE collected_data->'input_data'->'current_page'->>'name'='<page>'
     AND collected_data->'input_data'->'site_record'->>'domain'='<domain>'
     AND status='COMPLETED' ORDER BY created_at DESC LIMIT 1),
gate AS (SELECT k, v FROM run, LATERAL jsonb_each(run.cd) e(k,v)
          WHERE k ~ '^copy_gate_[0-9]+$' AND v->>'status'='repaired')
SELECT gate.k,
  (gate.v->'result' <> (SELECT cd->replace(gate.k,'copy_gate','generated_content')->'result' FROM run)) AS gate_changed_something,
  (SELECT bool_or(pc.content_data->>'content' = (SELECT cd->replace(gate.k,'copy_gate','generated_content')->'result'->>'content' FROM run))
     FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
    WHERE s.domain='<domain>' AND p.name='<page>') AS stored_matches_PRE_repair,
  (SELECT bool_or(pc.content_data->>'content' = gate.v->'result'->>'content')
     FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
    WHERE s.domain='<domain>' AND p.name='<page>') AS stored_matches_POST_repair
FROM gate;
```

**Reading it.** `gate_changed_something = false` → the run proves nothing about persistence (every
rewrite was rejected); pick another page. Otherwise the two `stored_matches_*` columns are the
answer, and asking BOTH is the point — a single "is the rewrite present?" test cannot distinguish
"lost" from "never attempted".

- PRE=true, POST=false → the §22 defect: repair made and thrown away.
- PRE=false, POST=true → **the fix works at the artefact.** This is what 548 is for.
- both false → the component was rewritten by something else after the save; go and find it before
  drawing any conclusion about the gate.

**Measured 2026-08-22 as a NEGATIVE CONTROL** on `loanzy.uk/tool-loan-repayment-calculator`, built
09:10Z — before 548 applied at 09:20:25Z — and whose save was ACCEPTED:
`gate_changed_something=true, stored_matches_PRE_repair=true, stored_matches_POST_repair=false`.
Same pipeline, same morning, save accepted, repair lost. That is the control the post-548 run has to
invert. ⚠ Use `created_at`, not `updated_at`, to place a run either side of the migration.

**If the save was REFUSED there is nothing to read.** Check the parent before blaming the gate:

```sql
SELECT p.status, p.current_step, left(p.collected_data->'__step_error'->>'message',300)
  FROM orchestration_states c JOIN orchestration_states p ON p.orchestration_id=c.parent_orchestration_id
 WHERE c.collected_data->'input_data'->'current_page'->>'name'='<page>' ORDER BY c.created_at DESC LIMIT 1;
```
A parent at `complete_error` with `save_page_sections … COMPONENT FLOOR REFUSED` is `bugs_open/253`
(framework_rewrite slug), not this gate — and "nothing was written" means the whole page, so every
section's repair is invisible, not just the refused slot.

## 9. THE RECONCILIATION CENSUS — and the one thing that makes it read wrong

Added 2026-08-23. §26's fix is judged by this query, so it must be segmented by `status` or it lies.

```sql
WITH m AS (
  SELECT os.updated_at, e.key, e.val->>'status' AS status,
         (e.val->>'targets')::int AS targets,
         COALESCE(jsonb_array_length(e.val->'rewritten'),0) AS rw,
         COALESCE(jsonb_array_length(e.val->'rejected'),0)  AS rj
  FROM orchestration_states os, LATERAL jsonb_each(os.collected_data) AS e(key,val)
  WHERE e.key LIKE 'copy\_gate%' AND jsonb_typeof(e.val)='object'
    AND (e.val->>'targets') IS NOT NULL AND (e.val->>'targets')::int > 0)
SELECT status, count(*) AS markers,
       count(*) FILTER (WHERE targets <> rw + rj) AS not_reconciling,
       count(*) FILTER (WHERE rw + rj = 0)        AS account_for_none,
       count(*) FILTER (WHERE targets < rw + rj)  AS over_counted,
       sum(targets) AS targets
FROM m GROUP BY status;
```

⚠ **`LIKE 'copy\_gate%'`** — the underscore is a wildcard in `LIKE`, so an unescaped `copy_gate%`
also matches things like `copyXgate…`. Escape it.

⚠ **Segment by `status`, and do NOT expect zero overall.** `targets == rw + rj` holds only for
`status='repaired'`. A `repair_unavailable` marker accounts for **none** of its targets by design —
the five early returns in `runNegationRepair` (lines 454, 458, 540, 559, 570) all precede the
`unansweredTargetRejections` call at 665. Read its `error` field for which one fired. **Do not tune
this query until the total reads zero** — that hides the ceiling failures, which are the expensive
half (§27).

⚠ **`over_counted` is NOT a regression — it is the hallucination case, and it FIRED.** A replacement
matching no target appends a `no_such_sentence` rejection with no target behind it, pushing `rw + rj`
above `targets`. **CORRECTED 2026-08-24**: this section previously said `no_such_sentence` had never
fired and that a non-zero reading would be "new information". It fired the day the accounting fix went
live — 1 marker of 122 (`targets=5, rewritten=4, rejected=2`; all five targets accounted, plus one
invented sentence correctly logged). **So expect a small non-zero `over_counted`, and do NOT chase it.**
The invariant to test is:

```
targets == rewritten + rejected - count(reason='no_such_sentence')     -- status='repaired' only
```

⚠ **Never close the gap by loosening `matchTarget`.** Making every replacement find a target would
splice rewrites into copy the model was not describing — a silent content defect.
`TestReconciliationExcludesHallucinatedReplacements` fails if anyone tries (mutation-proven).

## 10. IS ANY STEP HITTING ITS OUTPUT CEILING? (fleet-wide, and it found §27)

The estate's own durable rule is the detector: **`output_tokens == max_tokens` means the completion
was CUT, not finished.** `llm_call_log.max_tokens` is fed from `__sent_max_tokens`, i.e. the ceiling
**APPLIED**, not the one requested — which is the distinction `d6fc76dde` was about.

```sql
SELECT agent_type, step_name, max_tokens, count(*) AS calls,
       count(*) FILTER (WHERE output_tokens >= max_tokens) AS cut,
       round(100.0*count(*) FILTER (WHERE output_tokens >= max_tokens)/count(*),1) AS pct_cut
FROM llm_call_log
WHERE created_at > now() - interval '3 days'
  AND max_tokens IS NOT NULL AND output_tokens IS NOT NULL AND provider='anthropic'
GROUP BY 1,2,3 HAVING count(*) FILTER (WHERE output_tokens >= max_tokens) > 0
ORDER BY pct_cut DESC, cut DESC;
```

⚠ There is **no `finish_reason` column** on `llm_call_log` — reaching for one is the first thing that
fails. ⚠ Pin `provider='anthropic'`: the comment at `rewrite_negations_action.go:~503` records that
gemini's `max_tokens` bookkeeping differs (`bugs_open/110`), so a mixed-provider census compares two
different things.

⚠ **The step name is the LOOP-EXPANDED one** — `process_sections_loop_iter_1_rewrite_negations`, not
`rewrite_negations`. A query filtering on the bare step name returns nothing and reads as "no
truncation". Filed fleet-wide 2026-08-23 (`LANDMINES.md`, "A truncation/cost census over
`llm_call_log` keyed on the step name you CONFIGURED returns zero rows") — the pre-existing entry for
this trap names pod logs and `orchestration_states`, **not** `llm_call_log`, so it is not findable
from the table you are querying.

## 11. READING THIS STEP'S CONFIG AT ALL — it is NOT a top-level step

The `rewrite_negations` step lives inside a **sub-workflow**, so every top-level
`default_config->'workflow'->'steps'` query returns 0 rows and reads as "the step does not exist".
This is a known fleet-wide trap — `LANDMINES.md` documents it and gives the general idiom, a
**recursive walk** that finds a step at any depth and is what to reach for when you do not already
know the path:

```sql
FROM agent_definitions ad,
     LATERAL jsonb_path_query(ad.default_config, '$.**.steps') AS steps,
     LATERAL jsonb_each(steps) AS s(key,value)
WHERE s.value->>'action' = 'rewrite_negations'
```

The explicit-path version below is faster and asserts the path is where this lane thinks it is —
use it for THIS step, and the walk above when auditing an action across agents.

```sql
SELECT e.k AS step, e.s->'config'->'ai_service'->>'max_tokens' AS max_tokens,
       e.s->'config'->'ai_service'->>'model' AS model
FROM agent_definitions a,
     LATERAL jsonb_each(a.default_config #> '{workflow,steps,process_sections_loop,config,sub_workflow,steps}') AS e(k,s)
WHERE a.type='page-content-writer' AND a.is_active
  AND COALESCE(a.is_snapshot,false)=false AND a.deleted_at IS NULL
  AND e.s->'config'->'ai_service' IS NOT NULL ORDER BY e.k;
```

Both steps should read `16000` / `claude-sonnet-5` after `569`. `generate_content` is the **anchor**
the repair ceiling was chosen against — if it moves, re-read `569`'s rationale.
