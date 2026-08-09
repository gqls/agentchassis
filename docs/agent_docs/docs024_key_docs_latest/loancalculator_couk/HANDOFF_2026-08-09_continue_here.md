# HANDOFF — loancalculator.co.uk · **the `bugs_open/227` PLATFORM thread** · continue here (written 2026-08-08 late, for 08-09)

> ⚠ **SCOPE, added 2026-08-09 — this directory has two live handoffs dated 08-09.**
> **This file is the `bugs_open/227` job** (experience-planner writes another site's plan):
> the fix is written, dry-run proven and **NOT applied**. Everything below is still owed
> and still accurate about 227.
> **The site's COPY/VOICE thread is `HANDOFF_2026-08-09b_continue_here.md`** — note the
> **`b`**. They are different jobs; neither supersedes the other.

**Supersedes `HANDOFF_2026-08-08_continue_here.md`.** That file is correct about the
voice rollout and can be read for its §2/§3/§6 (the tools, the CSS trap, the method
notes), which are unchanged and still worth reading. It is **incomplete** in one way:
it was written before the last two things this lane did, and its "NOTHING IS OWED"
banner is now wrong.

> **CORRECTED 2026-08-09 (morning) — "Nothing is owed on the SITE" was WRONG when written.**
> This file (`a5c8bea7e`, 23:47) and `HANDOFF_2026-08-08b_continue_here.md` (`f0305cc50`,
> 23:33) were committed **14 minutes apart by concurrent sessions**, and both claim to
> supersede `HANDOFF_2026-08-08`. 08-08b's §2 recorded one outstanding item — `index`/
> `prose-2` still carrying the register the owner struck — which this file's author had no
> way to see. **08-08b was right.** That item is now closed (corr `26648f55`); the banner
> below is accurate again, and everything this file says about **227 is unaffected**.
>
> Two things to carry forward: the chassis line below is stale (fleet is on **v1.0.1270**,
> pods rolled 08:49Z 08-09), and 08-08b §3's per-section remedy now has a **correction** —
> the conditional phrasing was followed on 08-09 and leaked anyway; locking the untargeted
> siblings is what held. Read 08-08b §3 before writing any rewrite guidance.

```
site         loancalculator.co.uk   0162cde4-633e-45e9-8ca6-87a6b2fe1d26
chassis      v1.0.1270 (releases are whole-fleet; the owner runs make release)
voice H      26 of 26 active pages · 26/26 HTTP 200 · toolgolden 11/11 exact
the site     DONE as of 2026-08-09 morning (see correction above — it was not, on 08-08)
the job      bugs_open/227 — a platform bug this lane found and owns
```

## Read order for a cold start

This file → `bugs_open/227_HANDOFF_…` (read its **top correction block first** — the
body below it is superseded in three places) → the header of
`docs/agent_docs/sql_for_agents/345_experience_planner_site_brief_becomes_data.sql`,
which is where the design and its reasons live.

---

# THE JOB: apply and prove `bugs_open/227`

## What 227 is, in one paragraph

`experience-planner` is a fleet-shared agent with one active row. Its prompts hardcode
**vonc.com's** diagnosis, decisions, data contract and tool names. So a plan composed for
any other site describes vonc's pages. It went unnoticed for three weeks because the agent
had only ever been run on the site it is hardcoded for: of 61 experience plans all-history,
59 are vonc's. This lane found it on 08-08 by running the planner on
`debt-difficulty-help` to judge a page-ordering question, and getting back a confident,
detailed plan about a game site.

## What is already done — and what it cost to find

**The fix is written and dry-run proven. It is NOT applied.**

`docs/agent_docs/sql_for_agents/345_experience_planner_site_brief_becomes_data.sql`
(+ `_ROLLBACK.sql`). Config only — **live on apply, no image, no roll.**

It does five things: adds a `load_brief` step that reads the site's OWN brief from
`doc_notes` (category `experience-brief`), chains it in ahead of `compose`, adds
`experience_brief` to compose's `input_fields`, rewrites `compose` generically around
`{{.experience_brief.text}}`, surgically de-contaminates the four other prompts, and —
in the same transaction — moves vonc's brief verbatim into the new channel so the
59-plan site keeps its owner rulings.

**Three findings from writing it, each of which changed the fix. All three are in the
bug file's correction block:**

1. **It is FIVE prompts, not the one the bug file names.** 48 hits across `compose`,
   `review_feasibility`, `review_honesty`, `review_mvp`, `reframe`. My first census was
   **case-sensitive** and missed two whole steps, because "Gauntlet" is capitalised in
   both — it read 37 hits, three steps, and both missing steps hold a veto. If you
   re-measure anything here, measure `lower()`.
2. **Three of the four council seats hold vonc's criteria as their general rule**, and
   `reframe` — the step that rewrites a vetoed plan — does too. So the bug file's "the
   review layer is not the defect here" is too generous, and, more practically: **a
   correct post-fix plan can still be objected to by a seat hunting for a feed and a
   timer this site never had.** That failure would look exactly like "the fix did not
   work". This is why 345 patches the seats and not just `compose`.
3. **The hardcoded premise is stale as well as wrong-site, inside a hard veto.**
   `review_honesty` asserts "vonc's evidence_base has ZERO facts". True on 2026-07-18;
   **false since 2026-08-08 08:58Z**, when vonc's `evidence_base` spec gained 4 facts. A
   vonc plan run today is told by its own anti-fabrication seat that four verified facts
   do not exist.

**What "dry-run proven" means, exactly.** The file was run against the live database with
its final `COMMIT` swapped for `ROLLBACK`: both guards passed, the snapshot was captured,
the brief inserted, `UPDATE 1`, and the verify `DO` block raised nothing — which means all
five `replace()` calls matched, the whole-row census came back clean, and the chain is
wired. Then it rolled back, and I re-queried to prove it left no trace (0 `load_brief`
steps, 0 brief notes, 0 backup rows, row still contaminated). **It has never been run
against a live orchestration.** No plan has been composed with it.

## What is left, in order

### 1. Apply it

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- \
  psql -U clients_user -d clients_db -v ON_ERROR_STOP=1 -f - \
  < docs/agent_docs/sql_for_agents/345_experience_planner_site_brief_becomes_data.sql
```

Expect: `DO`, `DO`, `BEGIN`, a snapshot NOTICE, `INSERT 0 1`, `UPDATE 1`, `DO`, `COMMIT`.

**If the drift guard raises, do not force it.** It pins all five prompts by md5 against
what the file was composed against at 2026-08-08 22:45Z. A raise means another session
changed a prompt since — re-read the five and recompose the `replace()` strings. This is
not a theoretical guard: **the row was bulk-updated at 22:01:02.606329Z along with 186
others in a single statement, with no snapshot taken.** I could not identify the mechanism
(it is not `UpdateUsageCount` — `usage_count` is still 0 — and not migration 338). If you
have spare curiosity, that is worth ten minutes: something mass-rewrites agent definitions
and leaves no trail.

### 2. Prove it behaviourally — and the positive control is the whole point

```bash
./092_TRIGGER_experience_plan.sh loancalculator.co.uk debt-difficulty-help \
  "getting help when you cannot keep up with a loan repayment"
```

Save the `CORRELATION_ID`. Budget ~30 min: dispatch queues behind the fleet.

Then assert **both directions** at `llm_call_log.prompt_rendered` — the only record of
what the model was actually handed, and the field that would have caught this bug on day
one:

```sql
SELECT step_name,
       prompt_rendered ~* 'provocation|gauntlet'  AS leaked,        -- expect FALSE
       prompt_rendered LIKE '%no brief on file%'  AS got_sentinel   -- expect TRUE
  FROM llm_call_log WHERE correlation_id = '<CID>' AND step_name = 'compose';
```

**Then run the same two assertions for `vonc-spark-game` and require the OPPOSITE
answer** (`leaked` TRUE — its brief legitimately contains the word — and `got_sentinel`
FALSE). Without that control, a fix that silently loads **no** brief at all passes the
loancalculator half perfectly. The most likely way this migration looks applied and does
nothing is `experience_brief` missing from compose's `input_fields`, which renders the
template empty and errors nothing; 345 sets it and the verify block checks it, but the
control is what proves it at runtime.

And the outcome that matters:

```sql
SELECT is_current, body ~* 'provocation|gauntlet' AS still_wrong, length(body)
  FROM doc_plans WHERE subject_type='experience' AND subject_key='debt-difficulty-help'
 ORDER BY created_at DESC LIMIT 1;
```

**Expect a council round to be spent.** Read the verdict before reading it as failure —
see finding 2 above: if a seat objects that the plan has no feed or no timed round, that
is the *old* criteria talking and it means a `replace()` did not take, not that the plan
is bad.

### 3. The second, separable defect — NOT fixed by 345

`persist_plan` runs immediately after `compose` (`compose → persist_plan →
review_journeys → …`), so the plan is written `is_current=true` **before the council sees
it**, and nothing demotes it when the verdict is rejected. In the 08-08 run the council
vetoed the plan at 18:25 and the reframed version was persisted as current 8 seconds
before the run ended `complete_refused`. **A council-rejected, fabricated plan was the
plan of record** until it was demoted by hand.

Two routes, and the choice is real:

- **(a) config only** — rewire so persist happens on the approved path. But
  `complete_escalated.output_fields` lists `plan_persisted`, and the escalation path is
  *meant* to surface a plan for a human to choose from, so that coupling has to be
  answered first.
- **(b) a Go seam** — give `write_doc_plan` an `is_current` / `set_current_when` config
  field (`platform/orchestration/actions/write_doc_plan_action.go:104` INSERTs and relies
  on `doc_plans.is_current DEFAULT true`). That is a shared-mechanism change: council gate
  **and** a concept-register entry in the same commit.

Deliberately left out of 345 — it is independent, and folding a Go seam into a config fix
is the scope-veto shape `CLAUDE.md` warns about.

**Current state, verified:** both `debt-difficulty-help` rows still exist, both still
carry vonc text, **0 are current**. The hand-demotion holds. Do not delete them — they
are the failing evidence, and a fix should be proven against the real bytes rather than an
invented example (that is what made the 219 fix trustworthy). A hand-demoted row is
distinguishable from a properly superseded one: `superseded_at IS NULL AND NOT is_current`.

### 4. When it is live and proven

- Register the new channel: the `experience-brief` `doc_notes` category is **a delivery
  path another workstream could use and would not know exists** — that is the concept
  register's stated bar. `docs026_concept_register/register/<category>.md`.
- The council gate does not apply (DB config is refused client-side; scope is `platform/`,
  `internal/`, `pkg/`).
- 227 stays in `bugs_open/` when done — owner direction 08-06, evidence inside.

## Two things that are NOT owed, so nobody re-opens them

- **The site itself is finished.** 26/26 pages in voice H, all serving, calculators
  identical to `GOLDEN_2026-08-08_voice_h_complete.json`.
- **The "should we trim the expanded copy?" question is DECIDED — keep it** (owner,
  08-08 evening). Not "probably keep"; decided. `HANDOFF_2026-08-08` §5 records what the
  ruling covers.
- `debt-help-uk` was reordered on 08-08 to lead with the free charities (owner ruling,
  executed as a `content_rewrite` through the framework with bespoke guidance stating the
  *reason*, not the order). All nine facts checked on the wire. Done; not owed.

## The one method note worth carrying out of this session

**A worked example beats the instruction around it — that is now four for four in this
lane**, and 227 is the fourth and worst: here it beat *the actual data*. The run held
loancalculator's real page list in `collected_data` and wrote vonc's plan anyway, because
the prompt named vonc's surfaces in the imperative immediately after naming the right
site. The lane's other three (spec rules, spec exemplars, pinned per-item guidance) are
in `NOTES`. The corollary the fix leans on: the generic instruction "if there is no brief,
do not import surfaces from anywhere else" is **not** expected to be load-bearing on its
own — what makes it work is that there is no longer a worked example to beat it.
