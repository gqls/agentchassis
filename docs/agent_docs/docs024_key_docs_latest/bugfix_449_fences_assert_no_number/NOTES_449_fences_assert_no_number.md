# NOTES — bug 449, no fence the tool-generator writes ever asserts a number

Append-only, newest at the bottom. Technical log: evidence, commands, what the system
actually said, and every misstep.

Bug file:
`bugs_open/449_HANDOFF_2026-09-02_no_fence_the_tool_generator_writes_ever_asserts_a_number_so_a_calculator_that_computes_garbage_passes_acceptance.md`

---

## 2026-09-03 — session opens: is there an active thread, and is the bug still true?

### The thread was inactive, so this session resumed it

`scripts/who-owns.py 449` names `mortgagecalculator_couk_adoption` [ACTIVE, 32 commits/14d]
as the owner. That verdict is **lagging by construction** (it reads commits), so I checked
the clock rather than trusting the label:

- The lane's last commit of any kind: `bcef68058`, **2026-09-02 22:35:15 +0100**. The bug
  itself was filed 2 minutes earlier (`fd33fe4f9`, 22:33:22) and the lane went quiet
  immediately after.
- `git log --since='2026-09-03 00:00' -- .../mortgagecalculator_couk_adoption/` → **empty**.
- No commit today mentions 449 in its subject.
- Today's fleet restart wave is real and large — 40+ commits between 11:23 and 11:44 from a
  dozen lanes — and **the mcalc lane is not in it.**
- `git status` on `internal/adapters/browserrunner/run_checks_action.go` and on
  `platform/orchestration/actions/*criteria*` → **clean**, so no uncommitted session is
  mid-fix in the fence code either (the check `who-owns.py` cannot do).

The bug file's own status line agrees: *"Status: OPEN. Nothing changed, nothing dispatched."*

**Conclusion: inactive. Resumed here.**

### The bug is still valid — and it is getting WORSE while nobody watches

Two independent re-measurements, both first-hand today.

**(a) The cause is still in place.** Both fence-authoring agents still lack the vocabulary:

```sql
SELECT type,
       (default_config::text LIKE '%computed_values%') AS knows_computed_values,
       (default_config::text LIKE '%interaction%')     AS knows_interaction,
       updated_at
  FROM agent_definitions
 WHERE default_config::text LIKE '%```criteria%'
   AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```

```
        type        | knows_computed_values | knows_interaction |          updated_at
--------------------+-----------------------+-------------------+-------------------------------
 experience-planner | f                     | t                 | 2026-09-03 08:56:53.045885+00
 tool-generator     | f                     | t                 | 2026-09-03 08:56:53.045885+00
```

⚠ **`updated_at` of today, 08:56, is NOT a prompt change and must not be read as one.**
I checked before believing it: `SELECT count(*) FROM agent_definitions WHERE updated_at =
'2026-09-03 08:56:53.045885+00'` returns **208** — a bulk touch of every row at one second.
Had I quoted the timestamp as "the prompt was revised this morning" I would have been
confidently wrong about the one fact the whole bug rests on.

**(b) The census has grown by addition, exactly as the counting-date rule predicts.**
`bugs_open/449` §2 was measured 2026-09-02. Re-run today over `doc_plans` (`subject_type
='tool' AND is_current`, fence = the ` ```criteria ` block):

| author | fences | assert **no value at all** | uses `computed_values` | **drives inputs** (fill/select) | **drives inputs AND asserts nothing** |
|---|---|---|---|---|---|
| `tool-generator` | **186** (was 170) | **115** (was 107) | **0** | 91 | **55** |
| `operator:bugfix224-session` | 16 | 0 | 16 | 16 | 0 |
| `webdesign_couk_thread` | 14 | 4 | 0 | 6 | 0 |
| `operator:mortgagecalculator-lane-a4` | 8 | 0 | 8 | 8 | 0 |
| `operator:staged_component_build` | 8 | 0 | 6 | 7 | 0 |

**[MEASURED 2026-09-03.]** `+16 fences and +8 blind ones in about 24 hours`, and the newest
`tool-generator` fence carries `created_at` of **today**. The defect is not a standing
backlog to be tidied — it is a **live intake**, and every hour the fleet runs it mints more.

The two right-hand columns are mine, not the bug's, and they are the sharper cut:
**55 fences DRIVE INPUTS with `fill`/`select` and then assert no value of any kind.** The
fence itself declares "this tool takes input"; it then declines to check what came out. That
subset needs no classifier to identify — the evidence is inside the fence — which matters,
because a guarantee conditional on a classifier inherits the classifier's gaps.

### The cause, read out of the prompt rather than inferred

I dumped `agent_definitions.default_config` for both agents and read
`workflow.steps.compose_plan.config.prompt_template` (`tool-generator`, 2,766 chars).
It enumerates a **closed** vocabulary — four mandatory checks (`selector_exists` boots,
`no_console_errors`, `page_status_ok`, `no_horizontal_overflow` mobile-fit) plus "ONE
interaction check ONLY if you can copy real ids", and it says in terms:

> "No other check type exists for interactions — never emit `"type":"click"` or
> `"type":"fill"` as a check type."

`computed_values` appears nowhere in either agent's config. So the 0-of-186 is not a
modelling failure and not a hard case: **the type is never a candidate.** The prompt also
caps the whole PLAN at "under 3000 characters", which is a real constraint on any fix that
adds text to it.

> **On the diagnosis loop (090).** Not run, deliberately, and the CLAUDE.md norm of
> 2026-07-31 asks me to say why rather than omit it silently. That norm binds a session
> *filing* a cross-cutting root cause. This one is already filed by another lane, and the
> cause is not an inference from greps — it is the literal absence of a token in a prompt I
> read in full, plus a `write_doc_plan` seam I enumerated by grepping every Go writer of
> `doc_plans`. Both are self-evidencing: they could have come out otherwise and did not.
> The structural claim I am adding (single write door) is verified below, not asserted.

### The seam, enumerated rather than assumed

```
$ grep -rn "INSERT INTO doc_plans\|UPDATE doc_plans" --include='*.go' . | grep -v _test
platform/orchestration/actions/write_doc_plan_action.go:125   UPDATE doc_plans   (supersede)
platform/orchestration/actions/write_doc_plan_action.go:136   INSERT INTO doc_plans
platform/orchestration/datahelpers/travelling_docs_rekey.go:52 UPDATE doc_plans SET subject_key   (rekey only, never body)
docs/agent_docs/.../travelling_docs/write_doc_plan_action.go   (a doc artefact, not compiled)
```

So **`write_doc_plan_action.go` is the only production Go writer of a PLAN body**, and
exactly three live agents reach it:

```sql
SELECT type FROM agent_definitions WHERE default_config::text LIKE '%write_doc_plan%'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
 → experience-planner | experience-register-writer | tool-generator
```

⚠ **But operator scripts bypass it.** `install_fences.py` in the mcalc lane writes
`doc_plans` rows over `psql` directly. So the door governs every *generated* fence and no
hand-installed one — which is the right way round (the hand-installed ones are the 16+8+6
that already carry `computed_values`), but it means the door is not a total guarantee and
must not be described as one.

### The design tension that makes fix candidate 1 unsafe as written

I read `runComputedValues`'s docstring in `internal/adapters/browserrunner/run_checks_action.go`
in full. It says the values

> "are not authored by hand and are not judged for correctness here. They are CAPTURED from
> the tool while it is known good (`toolgolden.py --emit-criteria`) and then defended … a
> golden captured from an already-wrong tool pins the wrong answer."

So `computed_values` is a **regression/pinning** check by construction, not a correctness
check. At generation time the tool is newborn: there is no known-good state to capture, and
the generator is handed `{{.generated_html}}` — the tool's own code — so any expectation it
derives shares a failure mode with the implementation it is meant to police. The estate has
shipped that mistake twice already (`bugs_open/224`, `bugs_open/225`: an expired £625k FTB
SDLT cap certified green for sixteen months).

**Therefore "just teach both agents the type" is not sufficient on its own**, and the bug
says so itself in its own ⚠ ("an emitted value is not an expected value"). Whatever ships
has to say what an expectation's SOURCE is, and let the platform tell an independently
derived value from one captured off the page. The machinery for the strong version already
exists — `verify_criteria.py` re-derives from DEFINITION / REGISTER / CONVENTION at three
labelled strengths — but it lives in one lane's directory, not in the framework.

### Prior art found, so it is not rebuilt

- `platform/orchestration/actions/experience_criteria.go` — `ValidateExperienceCriteria`,
  rules P1–P11, each traceable to a live failure. **Only production caller:**
  `write_experience_pattern_action.go`, the experience-PATTERN register. Tool fences have
  never been through it.
- `write_doc_plan_action.go` already carries the precedent for validating a tool fence at
  the door: for `subjectType == "tool"` it refuses a malformed `facts` declaration, sharing
  `criteriaFactsFromValue` with P11 rather than re-spelling the rule (2026-08-24,
  `bugs_open/288` defect A). Its comment records that P11 "had never once seen a tool fence"
  — the same gap, one rule along.
- `check_tool_acceptance.go` (Tier 2) already holds the honesty discipline this bug wants
  extended: *"passes → a finding only (Tier-2 pass must never be read as 'the tool works')"*.
  It runs on a schedule over every eligible tool, so it can see the standing 115 with no
  backfill, and it has a `needs_criteria` doc_note path with a 30-day per-subject cooldown.

---
## 2026-09-03 — neighbouring threads, and two constraints found before planning

### Nobody else is in the files this fix would touch

Checked rather than assumed, because on this tree a dirty file is somebody's plan:

| file | last touched |
|---|---|
| `write_doc_plan_action.go` | 2026-08-24 (`995b5fbbe`, the 288 facts-validation door) |
| `experience_criteria.go` | 2026-08-22 (`b32aa9cd9`, contrast_ratio) |
| `check_tool_acceptance.go` | 2026-08-25 (`f44451494`) |
| `run_checks_action.go` (browserrunner) | 2026-08-22 (`b32aa9cd9`) |

All four also clean in `git status`. So no live collision in code.

### ⚠ But a lane changed the tool-generator PROMPT four hours ago

`docs/agent_docs/sql_for_agents/732_tool_prompts_learn_the_paired_ink_rule.sql`, committed
**2026-09-03 12:10** (`0325ddebb`) for `bugs_open/458` — the tool prompt was never taught the
paired-ink rule. It edits the same `agent_definitions` ROW I need.

**It does not collide, and I checked the path rather than the row.** 732 anchors on
`{workflow,steps,generate_tool_html,config,prompt_template}` and
`{workflow,steps,improve_tool,…}`. The fence vocabulary lives at
`{workflow,steps,compose_plan,config,prompt_template}` — a different key. Two surgical
`jsonb_set`s on disjoint paths compose in either order.

**Had I checked only "does another migration touch tool-generator", I would have concluded
'yes, wait for it' and stalled for nothing.** The row is not the unit; the JSON path is.

732 is also the template to copy, and it is the pattern CLAUDE.md asks for: a pre-guard that
counts the verbatim anchor and `RAISE EXCEPTION`s if it has moved ("re-read it and re-anchor
rather than overwriting a prompt this migration has not seen"), an idempotency arm, then a
post-verify in a `DO` block that raises — **not a block of bare `SELECT`s, which cannot stop
the `COMMIT`.**

### ⚠ `write_doc_plan` is TWO optional keys from the RFC_022 budget

```
$ ./scripts/audit-optional-key-budget.sh
      8 optional keys   3 carriers  write_doc_plan
  budget: 10 — 0 shared action(s) over it
```

So the door has **8 of its 10**. A fix adding one optional key to `write_doc_plan`'s input
spec leaves it at 9 and passes; **two puts it AT the budget** and owes one review of the
accumulated surface, recorded in `architecture_review/optional_key_budget_acks.json`. That
is a design constraint on the plan, not an afterthought — and it is exactly the accumulation
signal RFC_022 exists to raise. Prefer a fix that adds **zero or one**.

### Landmines already banked against this machinery — read before authoring anything

`grep 'criteria' LANDMINES.md` returns a lot, and four bear directly:

1. **Tier 2 ignores `no_auto_fix`** (LANDMINES §8626). A fence of only `computed_values` is
   inert to Tier 2 — but Tier 2 appends three built-in shell failures *outside* the criteria
   loop, so it can still raise `improve_tool` and aim an automated rewriter at a SHARED
   component. Adding one innocuous `page_status_ok` to a correctness fence is what arms that.
   **Directly relevant:** any fix of mine that tells the generator to emit `computed_values`
   alongside the existing four health checks creates exactly this combination.
2. **An unknown key in a criteria fence is dropped in silence** (§8989), so a check can
   assert less than it appears to and never say so. `selector_count`'s `expect_count` is the
   worked case — a real-sounding key nothing decodes. Any new field I add must be added to
   the runner's decode struct AND to `experienceCheckTypeFields`, or it is decoration.
3. **`computed_values` reads a `display:none` subtree perfectly well** (§9303) — so it pins
   the arithmetic and says nothing about whether a visitor could see it. This is the precise
   reason the bug's §4 wants correctness *and* health in one fence, not correctness instead.
4. **Prose naming the fence in backticks hijacks extraction** (§1271) — the extractor takes
   the FIRST ` ```criteria `, so a sentence mentioning it in a PLAN body silently becomes the
   fence. Relevant because my prompt change adds prose about fences to a prompt that writes
   PLAN bodies.

### The other lanes this reaches, for CONTRIBs once the plan is settled

- `mortgagecalculator_couk_adoption` — filed 449 and owns 441 (fences naming pre-conversion
  ids). ACTIVE but quiet since 22:35 last night.
- `loancalculator_couk` — **ACTIVE, committed 11:50 today.** Owns `toolgolden.py`, the
  `--emit-criteria` capture that `computed_values` fences are built from, and the LANDMINE
  about it. Any framework rule about where an expected value may come from is theirs too.
- `staged_component_build` — the existence proof: 6 of its 8 fences carry `computed_values`
  AND `interaction`. Its practice is the template being generalised.
- the `458` lane (migration 732, committed 12:10 today) — sharing the tool-generator prompt.
- the boxingonline/designblog rulings lane — `scripts/audit-experience-promises.py` RULE B
  ("a tool page with nothing the reader can use or read") is the same shape one level out:
  *the page keeps the letter of its build and breaks the promise it makes.*

---
## 2026-09-03 — the fable run failed, and the plan is therefore un-second-opinioned

The owner asked for the plan to be prepared **using fable**. I dispatched a
`claude-fable-5-1` agent with a full briefing — the census, the prompt text, the seam
enumeration, the design tension, the CLAUDE.md rulings that bind it, and the §6 verification
bar. It ran ~30 minutes and **terminated on a session rate limit** (HTTP 429, "You've hit
your session limit · resets 4:10pm Europe/London"), having emitted one line: *"I'll start by
reading the bug file and the key seam files in parallel."*

**So the plan in `PLAN_2026-09-03_449_fences_assert_no_number.md` is mine, not fable's**, and
the head of that file says so. This is recorded rather than glossed because the *reason* the
owner asked for fable — an independent read of a design decision that inverts the obvious
ordering — is exactly the input the document is missing. Re-running it after 16:10 is cheap.
If fable's plan differs materially, the difference goes in as a correction here and in the
PLAN, not as a replacement.

⚠ **Do not retry a fable dispatch before 16:10 BST** — the limit is per session and a retry
costs a round for a guaranteed 429.

## 2026-09-03 — the ordering decision, and why the obvious order is wrong

The cheap, live-on-apply, cause-closing move is to teach `tool-generator` the check type.
Every instinct says do that first. **The plan puts it LAST**, and the reasoning is the part
worth keeping:

1. A pinned value the generator could not derive independently is **worse than no value**.
   Silent blindness is uninformative; a defended wrong number is believed. That is the
   `bugs_open/224`/`225` class — sixteen months of a green certificate on an expired tax cap.
2. `bugs_open/441` is, in the mcalc lane's own re-framing today, a **live generator of stale
   fences**. `runComputedValues` fails rather than skips when its element is missing — by
   design, and rightly. So value assertions authored while 441 is live would fail for the
   wrong reason and aim `tool-improver` at arithmetic that was never wrong.

So the honest-record phases (P1 verdict, P2 door note, P3 report) ship first: they are
strictly additive, they touch all 186 fences at once with no backfill, and none of them can
create a false failure. Only then does the authoring change land.

**[UNVERIFIED]** that 441's fix lands before the next generator run — asked of the mcalc lane
in `CONTRIB_2026-09-03_from_the_449_lane_…`, not yet answered. If it lands after, P4 must be
gated rather than merely sequenced.

---
## 2026-09-03 (afternoon) — P1+P2+P3 shipped; two missteps, one of them my own control

### What landed

| | commit | live? |
|---|---|---|
| P1 verdict states its own scope | `0b9a5c9e1` | **inert until the roll** (Go) |
| P2 door records a fence born blind | `0b9a5c9e1` | **inert until the roll** (Go) |
| P3 standing report, windowed | `23c8a7d71` | live now (script) |
| register TP-009 + index row | `e27aa00bb` | n/a |
| LANDMINE + verifier armed | `d58658d31` | synced to `doc_notes` |
| bug file §8 | `e9ef673a5` | n/a |

Council: `Council-Submitted: 8745ad9e-1802-4e08-a9b0-eb493cd11243`, dispatched 11:5x, seen at
`review_prior_art` then `review_tooling_provenance`. **Verdict not read at time of writing** — the
trailer asserts nothing, so 098 credits the commit automatically once approved. Still owed: READ it
and act on a REVISE.

⚠ **The owner told us mid-session that a fresh chassis was building within the hour**, which is why
P1/P2 were written and committed before the docs were finished rather than after. On this tree
`make build-*` builds from committed HEAD, so the difference between committing at 12:47 and
committing at 14:00 is the difference between shipping today and waiting for the next roll.

### The verification, and why "the tests pass" was not enough

All three mutations were RUN:

| mutation | expected red | result |
|---|---|---|
| delete the judge's `Scope of this verdict:` line | liveness-only + pattern tests | **RED** |
| delete the door's 449 block | door records test | **RED** |
| credit `computed_values` by TYPE, not by having expectations | empty-`expect_values` subtest | **RED** |

Then restored and proved restored with `diff -q` against pre-mutation copies — not by eye.

⚠ **The package is ALREADY RED from other lanes' dirty files, and non-deterministically so.** Two
full-package runs with my change absent failed on *different* tests
(`TestValidateTemplateDataStillReportsAGenuineAbsence`, then `TestOneAttractorTagIsAWeakFit`).
**That is the confound, and running the dirty tree again would never have resolved it.**
`scripts/verify-head-builds.sh --with <4 files> --test` against HEAD `48bd6c5b6` returns **`ok`**
for `platform/orchestration/actions` — which is the whole point of that script and I should have
reached for it first instead of arguing from which files were dirty.

It also surfaced a genuine red at HEAD that is nobody's fault but wants an owner:
`discovery_checks/TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens`, from the 458 lane's
`0325ddebb` (12:10 today) — four paired-ink tokens added to `canonicalCSSTokens` and not to
`rendererGuaranteedTokens`. Reported into `bugs_open/458` §11 with the two things they cannot see
from where they stand: their own §9 verification runs `-run TokenAudit`, which does not select that
test, and the failure is in a **different package** from the one they edited.

### Misstep 1 — a `cd` in one Bash call re-pointed every later relative path

Logged in `WRONG_CALLS.md`. Caught only because I named a file and got ENOENT, which I first read as
*another session has moved it*. The quiet form — `grep -rn … .` from the wrong directory — returns
zero matches, exit 0, no message, and I had used exactly that shape to establish a **negative**
(that `write_doc_plan_action.go` is the only Go writer of `doc_plans`).

### Misstep 2 — my own demand control was pinned to a name that changed the same morning

Also logged. `audit-fence-value-assertions.sh` had a built-in demand control (right instinct: a
census returns the same comfortable number whether the corpus is clean or the query is blind), and I
implemented it as two hard-coded `created_by` values on the reasoning that a *specific* control is
stronger. **`created_by` is a free-text lane label, not an identity.** The mcalc lane re-keyed its
eight fences hours earlier — `operator:mortgagecalculator-lane-a4` had become
`operator:mortgagecalculator-lane-2026-09-03-701-rekey` — and my control passed only on the other
name. It **fails closed**, which is precisely why I would not have questioned it: a false exit 2
reads as "be careful", and repeated, it teaches everyone to ignore the control.

Rewritten over the PROPERTY ("some author still shows a non-zero `uses_computed_values`") and it now
**prints which author satisfied it**, so a reader sees the evidence rather than the word PASSED.

### The number moved again while I was working

First run of the report: `tool-generator` **116** blind, not the 115 I measured at 11:4x. And
**10 blind fences created in the last 24 hours**. That is the strongest single argument for P1: the
standing stock is a per-site repair job, but the intake is a framework defect and it is open.

---
## 2026-09-03 13:3x — THE ROLL HAPPENED AND DID NOT CARRY MY CHANGE. P1+P2 are still inert.

The owner said mid-session that a fresh chassis was building. It shipped: `agent-chassis` moved
**v1.0.1356 → v1.0.1358**, pods restarted ~13:07 BST. My P1+P2 commit `0b9a5c9e1` landed 12:47 and
is in HEAD's ancestry, so the arithmetic looked right and I was one step from writing "it shipped".

**It did not.** `[MEASURED 2026-09-03 13:3x]`

| probe | result | |
|---|---|---|
| `fence_asserts_no_value` | **absent** | under test — my P2 doc_note category |
| `liveness_only` | **absent** | under test — my P1 `verdict_scope` value |
| `Tier-4 acceptance PASSED` | **PRESENT** | control + — must be present, and is |
| `zzz_invented_string_449_check` | absent | control − — must be absent, and is |

Both controls sat on **opposite sides** of the answer, so the probe could discriminate and the two
absences mean what they say: **v1.0.1358 was cut from a commit earlier than 12:47.** P1 and P2 are
committed, reviewed-in-flight, and **not running**. They need the *next* roll.

⚠ **Whoever builds next must bump `IMAGE_TAG`** (makefile ~line 22, currently `v1.0.1358` — now
deployed). A same-tag rebuild serves the node's cached binary and every downstream check reads green.

### The method here is NOT the one CLAUDE.md prescribes, and that matters

CLAUDE.md says to ask the service what it is running:
`kubectl logs -l app=<service> --tail=300 | grep -m1 'build provenance'`, warning that an empty
result means "not in range, it is a startup line and it scrolls".

**There is no such line.** `LANDMINES.md` records it `[MEASURED 2026-08-25]`: `grep -rn 'build
provenance'` over the Go source returns **zero**, and a grep over a whole 4.6 MB pod log returns
nothing either. **The documented failure mode perfectly explains the real one**, so you conclude you
need more scrollback and go looking for it. I ran the prescribed command, got a 2.7 MB result of
unrelated JSON, and only found the correction because the landmine's own text was echoed *inside*
those logs — agents read it out of `doc_notes` — and the phrase caught my eye.

**That is luck, not method.** The method that would have worked is the standing one I had already
been reminded of at session start: **grep LANDMINES for the symbol you are about to trust**, before
running the command, not after it returns something confusing.

The correct check is the landmine's: **probe the CAPABILITY, not the commit, with a control on BOTH
sides in the same breath.** Its second layer is the sharper warning — the fallback binary probe
"produces an ANSWER": three absent results read exactly like "the fix did not ship", and if every
control is on the same side of the answer the set cannot discriminate at all. Mine were not, which
is the only reason today's two absences are evidence rather than three uninterpretable readings.

**The general form, and it is the one to carry:** *a roll is not evidence your fix shipped* — but
the sharper half is that **an image tag advancing, a pod restarting, and your commit being an
ancestor of HEAD are three true facts that jointly imply nothing.** All three held here and the
binary still does not carry the change.

---
