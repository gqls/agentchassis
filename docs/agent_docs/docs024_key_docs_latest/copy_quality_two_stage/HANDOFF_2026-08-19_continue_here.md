> # ⚠ SUPERSEDED 2026-08-20 — read `HANDOFF_2026-08-20_continue_here.md` first.
> Its state lines are stale in one way that matters: it says `bugs_open/327`'s platform fix is not
> this lane's to make and awaits an owner. **This lane took it, and it is LIVE on `v1.0.1319`.**

# HANDOFF 2026-08-19 (evening) — continue here

**Lane:** `copy_quality_two_stage`. **Supersedes `HANDOFF_2026-08-18_continue_here.md`**, whose
state lines are stale but whose peer-lane exchanges and apply recipe still hold.

> ## ▶ START HERE, IN THIS ORDER
> 1. **`SUMMARY_2026-08-19_the_fault_was_never_in_the_writer.md`** — 5 minutes, plain prose. Still
>    current on the writer/brief question. It does not know about `bugs_open/327` (below).
> 2. **`bugs_open/327`** — the defect found today. Read its §4 before anything else in it: it is
>    **NOT** the cause of the owner's complaint, and the tidy story that merges it with
>    `bugs_open/305` is wrong.
> 3. **This file's "Next work"**. Everything above it is context.
>
> **One-line state:** stage 2 is built, live and unchanged; the detector the last session
> specified is **built, run fleet-wide and registered (CQ-025)**; a second, larger defect in the
> same field is filed as `327`; four consumer lanes have been told; nothing is in flight from me
> except one diagnosis run; no decision is waiting on the owner.
>
> ⚠ **Re-verify before trusting anything dated here.** Chassis was **`v1.0.1315`** at 15:20Z
> (both replicas; the `build provenance` line had already scrolled out of `--limit-bytes=900000`,
> so the sha is unconfirmed — `copy-editor` is config-only, so no roll can affect it either way).
> Migrations **473/474 applied 10:34Z**, and **466, 489, 488, 491 landed after them the same day**.

## What changed today (2026-08-19, evening session)

### 1. The detector exists: `audit_writer_brief.py` (register **CQ-025**)

The previous session's stated build was *"`count_negation_tells.py` counts whole documents, so it
is the wrong tool pointed at a spec until it learns to read `formatted`."* It is done, and in a
better shape than that specified, because the first thing it does is **stop guessing which fields
the writer reads**:

```sql
SELECT DISTINCT m[1] FROM agent_definitions ad,
LATERAL regexp_matches(ad.default_config::text, '\{\{[^}]*site_specs[^}]*\}\}', 'g') m
WHERE ad.type='page-content-writer' AND ad.is_active
  AND COALESCE(ad.is_snapshot,false)=false AND ad.deleted_at IS NULL;
```

**Five fields, and only five:** `content_direction.formatted`, `identity.key_differentiators`,
`identity.target_audience`, `evidence_base.writer_block`, `design_intent.imagery_direction`. The
tool derives that at runtime, so a config change moves it. This **confirms** the landmine we filed
for `finetuning_uk_service` and **extends** it — `evidence_base.writer_block` is in the surface and
nobody here had looked at it; on `webdesign.uk` it carries **31 of that site's 45 tells**.

It reports surface · silent drops · tells over visible text only · supplied phrases, and
`--transfer "<phrase>"` settles the question that matters. **It reproduced yesterday's figures
independently** — `in days, not months` 1,369 prompts / 409 responses; `not a sales process`
**0 prompts** / 35 responses (present in output, in no prompt, therefore not transfer).
`--self-test` runs 15 controls, both arms; three mutation probes were caught.

**Exactly ONE mandated supplied phrase fleet-wide:** `ai-agent-orchestration.com`'s
`content_direction.emphasis` **orders** the tagline into *"the homepage hero, services page hero,
site footer, and meta descriptions"*. The owner objected to a hero; the brief commanded it.

### 2. `bugs_open/327` — the brief the writer reads is rebuilt from the PARTIAL, before the merge

Found while reading `FormatContentDirection` to write the tool's parser.
`site_spec_actions.go:212` formats the incoming partial; `:247` merges. `formatted` is a **string**,
so `siteSpecDeepMerge` takes its scalar-overwrite arm and the short one wins. **A careful, narrow
correction to a brief deletes the rest of it from the writer's view** — silently, with the document
still complete and the write logging success.

Fired on four sites on 2026-04-18 (`domain-research-classifier` → `build-site-planner`, nine
minutes apart); `finetuning.uk` recovered on 08-12; **three still serve a fragment**:
`robot-hands.com` (14 keys dropped, incl. ~3,900 chars of hero-copy rules that have never reached
a page), `leopardessconsulting.co.uk` (12), `ai-agent-orchestration.com` (12, running on 5 of 18
keys since April). Fleet: 8 of 25 sites drop at least one key.

**`090` filed** (`8be5f6e9-d0b3-43f7-9ee4-dee2432dd8b1`, orch `6073488a-…`). ⚠ **Iteration 1
returned `UNVERIFIABLE` and I published it as "the verdict" — it was not**; the run went back to
`assemble_bundle` and continued. **Read the FINAL verdict; do not quote 327's iteration-1 section.**
Its four named gaps are closed first-hand in the file regardless of outcome.

### 3. Four consumer lanes told, per the owner ruling of 2026-07-29 §3

`site_ai_agent_orchestration` (+ **the shrink-floor reply we owed them** — my gate had the same
defect and was made to **discriminate**, not relaxed: a shrink passes only if every removed figure
and link is still reachable elsewhere on the page, under 25% fails outright, and the discriminator
is mechanical so the agent cannot talk past its own gate) · `robot_hands` · `leopardessconsulting`
· `portfolio_positioning` (their pilot `remortgagecalculator.uk` **confirmed at 19**, still fleet
worst, and warned that the partial-write trap sits on the fix they want).

### 4. Two wrong calls, both mine, both logged

- **A per-site figure attributed to the wrong site**, off my own fleet report, already committed
  (`leopardessconsulting.co.uk` for what is `webdesign.uk`). **Check: quote a per-site figure only
  from a single-site run.** A `--fleet` report is for ranking, not quoting.
- **An intermediate diagnosis step published as a verdict** (above). **Check: read
  `current_step`, not the outcome field.**

Both in `WRONG_CALLS.md` with the cheap check; NOTES corrected in place.

## Next work, in the order that closes doors

1. ~~**Read the FINAL `090` verdict**~~ **DONE, same session — `CONFIRMED` at iteration 3**
   (orch `6073488a-…` `COMPLETED`). Four static citations matching §1's reading. ⚠ Two things
   came out of checking its evidence rather than its outcome, both now in `327`: it **refuted**
   my adoption-path claim (correct — that path never merges), and it **misread one state
   citation**, which surfaced a **second defect**: `formatted` is regenerated in a **random key
   order** on every write (Go map iteration, nothing sorts). Consequence that lands on the brief
   corrections we and `portfolio_positioning` are about to make: **a text diff of two briefs
   reports ~100% changed whether or not anything did** — verify by label presence and key
   content, never by diffing the rendered brief.
2. **`327`'s platform fix is NOT this lane's to make** — shared seam, council + roll. What is open
   is who takes it. The one-line invariant: compute `formatted` from `merged`, never from the
   incoming partial (both `site_spec_actions.go` and `apply_adoption_plan_action.go:280`). ⚠ The
   missing unit test is the interesting part: `FormatContentDirection` is tested on the map it is
   handed, which is the wrong scope — the defect is in **which map it is given**.
3. **The three fragment briefs are a CONTENT decision, not a repair.** Backfilling restores
   ~10,000 chars of instruction and changes what every future page says; on
   `ai-agent-orchestration.com` the restored `example_phrases` are themselves written in the
   construction the owner objected to. Read the diff. Not ours to run — told the owning lanes.
4. **A fourth run on `ai-agent-orchestration.com/index`** — unchanged from the last handoff, and
   still the cheapest open question about stage 2 itself: does a second pass find the two
   remaining restatements, or re-propose what it already did? ⚠ Two things have moved under it:
   that site's components were re-rendered (run 3's parked proposal already dangles by id), and
   that lane is actively working the site. Resolve by `(page_name, slot_name)`, and check their
   NOTES first.
5. **Dispatch** — wiring `content-quality-auditor` findings to `copy-editor`. Unchanged; held
   behind (4) and a human canary.
6. **Should the detector be SCHEDULED?** It is an observation a human runs, and an unrun detector
   goes stale — the estate's own lesson. The shaped option is the `optional-key-budget-check`
   pattern: a daily CronJob writing **one `doc_notes` row per run, including on clean results**,
   so a missing row means the job did not run rather than "nothing is wrong". That is a platform
   change this lane has deliberately not made. **Owner/architecture call, not a session's.**

## Standing cautions (fresh first; the 08-18 list still applies)

- **`splitlines()` is not `split("\n")`.** It also breaks on `\r`, `\x0b`, `\x0c`, `\x1c-\x1e`
  and ` `, all of which occur in authored spec prose. Parsing psql output that way **dropped
  three sites from a 25-site run and truncated a fourth to 2 chars, silently**. The tell was a
  number that had changed since a run ten minutes earlier. Let Postgres do the encoding
  (`jsonb_agg`) and there is no delimiter to get wrong.
- **An apostrophe is not a quote mark.** `['"]` reads `the client's own voice, not a long form`
  as a quoted phrase.
- **Never size a brief with `length(data::text)`.** On the worst site that is 4× what the writer
  sees. Establish the consumer's surface first, with a known-present phrase as the positive
  control — this is the rule that the withdrawn 08-19 census cost us.
- Everything under "Standing cautions" in `HANDOFF_2026-08-18_continue_here.md` still holds:
  dangling `page_component_id`, the duplicated migration number 462, served-page checks testing
  propagation, `kcat -P` exiting 0 having sent nothing.

## The five living docs

- **PLAN** — §11 delivery + corrections. *(Not touched today; nothing in the plan changed.)*
- **NOTES** — evidence log; today's entry is the tail, with one in-place correction.
- **README_where_we_are** — the owner's log; today's entry covers both findings in plain prose.
- **SUMMARY series** — 08-12 · 08-14 · 08-15 · 08-17 · **08-19 (newest)**. ⚠ **No new summary
  today, deliberately** — the 08-19 one was written this morning and the five headings would
  repeat it. The next genuine inflection is `327` being fixed or the briefs being corrected;
  write it then.
- **this HANDOFF.**

**Tooling this lane owns:** `gate_stage2_edit.py` (grades one proposal; `--self-test` MUST fail) ·
**`audit_writer_brief.py`** (the writer-visible brief; `--self-test`, `--fleet`, `--transfer`) ·
`count_negation_tells.py` (an OBSERVATION over a rendered page, never a gate) ·
`loanandmortgagecalculator_couk/gate_page_links.py` + its `acceptance/` baselines.

**Migrations this lane owns:** `447` (seed `copy-editor`) · `462` (3-edit budget + 32k cap). Both
have `_ROLLBACK` files. **No migration was written today.**
