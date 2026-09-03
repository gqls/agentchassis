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
## 2026-09-03 (afternoon) — council round 1 came back REVISE, and it was worth it

**Verdict: REVISE**, gated by `editquality`'s single HIGH. **`architecture` APPROVED**, which settles
the scope question I reasoned about in the PLAN: this is not architecture-scope. 8 approve, 3 object,
5 abstain.

**The HIGH found a real hole in my reasoning, and its conclusion for my code turned out false — and
those two facts are not in tension.** The objection: `insertDocNote` must pass a `subject_type`, and
`doc_plans` / `doc_notes` carry TWO SEPARATE CHECK constraints with different allowed sets; since I
made the insert **non-fatal**, a refused value would make P2 a permanent silent no-op with no signal.

Measured from `pg_constraint`:

```
doc_plans_subject_type_check: tool, pipeline, experience, action, experience-pattern, component   (6)
doc_notes_subject_type_check: those six PLUS landmine, decision                                   (8)
```

So `doc_notes` ⊇ `doc_plans` **today**, `'tool'` is in both, and the branch is gated on
`subjectType == "tool"` anyway. **My code was safe by luck of the value I happened to pass, and I had
not checked** — my own 8-point risk section never raised it, while citing the "two enforcement points"
shape elsewhere. That is the fair hit and the reviewer's remedy was right.

⚠ **The load-bearing fact is the SUPERSET relation, and it is not guaranteed.** Add a type to
`doc_plans` alone and this rule goes quiet. Recorded at the call site rather than in a lane doc.

⚠ **And a bound worth stating, because it is easy to overclaim the test I added:** `sqlmock` cannot
see a CHECK constraint — it accepts any string. So the test pins the value being **sent** and cannot
prove the database would take it. The `pg_constraint` read is what proves that. A test asserting "the
insert was attempted" would have passed in a world where every insert was refused.

### What the other objections were worth

| objection | answer | changed code? |
|---|---|---|
| `criteria_field` path unverified (low) | It is **explicitly configured** on the `judge` step, not a hopeful default — and `request_browser_run` SKIPS when it is empty, so any PASS existing proves the path is populated | no |
| reuse: proof-of-search, not inference (medium) | Exactly **3** Go types decode a fence's `checks`; the Tier-2 one does not decode `expect_values` **at all**; the browserrunner one is unexported in another tree; no exported fence decoder exists | yes — the search is now in the file header, with the residual stated |
| doc_notes category vocabulary (low) | Searched DB-side: `needs_criteria` (120) and `criteria` (1). `needs_criteria` means **no fence at all** — the opposite state | no |
| three-callers claim asserted (medium) | Re-verified **structurally** by walking workflow steps for `action='write_doc_plan'`, not by substring. Three, as claimed | no |
| who consumes the judge's result (low) | Exactly one agent; **zero** Go readers of `acceptance_verdict`, so no strict mapping to break | no |
| no lint against a future ad-hoc parser (low) | Accepted as a residual and stated in the file; the mitigation is that the two arms an ad-hoc parser gets wrong are pinned in BOTH the Go test and the detector's `--self-test` | no |

**Two of six changed the code. That ratio is the argument for the gate**, and matches this estate's
record (`a-revise-round-is-cheaper-than-the-defect-it-finds`).

Round 2 submitted on the same trail (`RESUBMIT_CORR`), so the artefacts accumulate in one place.

## 2026-09-03 — the detector is now SCHEDULED, and proven at the artefact

The first version was a bash script. **A detector nobody runs is this estate's documented failure
mode** — "detection works; schedule and dispatch do not" — so it was rewritten as Python on the
`listing-class-promise-check` pattern: ONE file with a dual-mode `_psql_argv`, so the hand-run
(`kubectl exec`) and the clock-run (direct psql) share the same query rather than being two
implementations.

**Proven end-to-end, at the artefact rather than at the apply output:**

- `kubectl get cronjob fence-value-assertion-check` reads back `40 7 * * *`, `suspend=false`.
- A manual `Job` from it: pod `state.terminated.exitCode` **0** (`|| [ $? -eq 1 ]` maps "findings",
  which is the normal state until the authoring fix lands, to success — only a 2 should fail the Job).
- **The row it wrote is on record**: 241 fences graded, **58 driving-and-blind fleet-wide**, 13 new in
  the window. A Completed pod that wrote nothing would have been a silent no-op, so the note is the
  assertion, not the exit code.
- `--self-test` pins the deployed `base/check.py` against the repo file; **that guard was
  mutation-proved** (append one line to the copy → self-test FAILS).

⚠ **The count moved 186 → 187 while this was being built**, which is the case for the CronJob in one
line: the thing it watches changes by a route no commit can carry.

---
## 2026-09-03 (~16:10Z) — P1 IS PROVEN IN PRODUCTION; P2 is unexercised for a reason no one had looked at

Picked the lane up from `HANDOFF_2026-09-03_continue_here.md` and ran its §3 first-fire checks. The
answer is good, and getting it raised one correction and one genuinely new structural finding.

### P1 — PROVEN, and proven on the discriminating arm, not just on a string match

`[MEASURED 2026-09-03 16:04Z]` The newest `acceptance-run` note is
**`tool-idea-stage-identifier`, 14:00:07.683828+00**, and it carries:

> `Scope of this verdict: ⚠ LIVENESS ONLY — this fence asserts no value of any kind, so the verdict
> says the page loads and responds and says NOTHING about what it computes (bugs_open/449)`

The four notes below it (08:47Z and earlier, all pre-roll) carry `carries_scope = f`. So the line
arrived with the roll and is firing.

⚠ **A printed line is not a working classifier, and this is exactly the `[MEASURED]`-but-not-
disconfirmable shape** the index warns about — if `criteriaAssertionPhrase` always returned
`LIVENESS ONLY`, the note above would look identical. So I checked the arm could have come out
otherwise, at two levels:

- **In source**, `criteriaAssertionPhrase` (`criteria_value_assertions.go:226-258`) has four
  outcomes — `exact`, `pattern`, `none`+`DrivesInputs`, `none` — and the unit tests already pin the
  negative direction (`criteria_value_assertions_test.go:179` asserts an `exact` fence does **not**
  contain `LIVENESS ONLY`).
- **On this subject**, the sub-branch was the right one. The note printed the *"page loads and
  responds"* variant, which is the **not**-`DrivesInputs` arm; so the fence must have no `fill` and
  no `select`. It does not:

```
subject_key=tool-idea-stage-identifier  created_by=tool-generator  created_at=2026-08-05
has_fill=f  has_select=f  has_expect_values=f  has_text_matches=f
```

Grade `none`, drives nothing → the *"page loads and responds"* wording is correct on every axis.
**P1 is live, exercised, and right.** (Note the fence itself dates from 2026-08-05 — this is a
standing-stock fence re-run today, not a new one.)

### CORRECTION — the roll was 13:28Z, not "~13:55Z"

The handoff's §3 said *"expect carries_scope = t on anything created after ~2026-09-03 13:55Z"*.
Measured at the pods, both `agent-chassis` replicas started **13:28:18Z / 13:28:43Z** on
`v1.0.1359`. The conclusion is unchanged either way, but the figure was wrong and every later
"pre-roll / post-roll" cut in this lane keys on it.

### P2 — zero notes, and the handoff's two-state discriminator could not see why

`[MEASURED 2026-09-03 16:04Z]` `fence_asserts_no_value` notes: **0 rows**. Demand control: **0**
tool PLANs written since 13:28:18Z (newest anywhere is 12:35:59Z, pre-roll). So by the handoff's
rule this is "not yet exercised, nothing is wrong" — and that is the right conclusion.

⚠ **But the handoff's §3 offered TWO states ("wait" / "the rule is broken") and there is a THIRD,
which is the one we are actually in.** A post-roll generator run *did* happen —
`3f5cb558-…` created **16:04:42Z** — and it wrote no PLAN because it died upstream:

```
step save_tool failed: failed to execute action create_tool_component:
tool birth refused (instance scope): script is not mechanically provable —
it declares into global scope and/or 7 binding(s) would dangle
```

(read from `collected_data->>'__step_error'`; the `error` column is **NULL** and `execution_path`
is `[]` — the bugfix-099 trap, a failed step presenting as `COMPLETED`.)

**The step order is the mechanism, and it is load-bearing for how P2 ever gets proven.** From the
live `tool-generator` row, the chain is:

```
ensure_site_record → load_brand_context → generate_tool_html → load_site_page_names
  → suggest_related_pages → save_tool → compose_plan → write_plan → index_plan → enqueue_rerender
```

`save_tool` runs **before** `compose_plan → write_plan`. So **every generator run refused at birth
writes no PLAN at all, and P2 cannot fire on it** — no note, and no PLAN for the demand control to
count either. A control that counts only the *writer's output* cannot tell "nothing was attempted"
from "attempts died upstream of the writer". Logged in `WRONG_CALLS.md`.

**Is that a standing blocker? No — measured, not assumed.** The refusal is not new: the gate landed
in `tool_birth_instance_scope.go` on **2026-08-21/23** (`e186a2bd3`, `2817f6661`, `b1a9fe7d4`,
`0e6c62168` — the RFC_032 lane), eleven days before the roll, and **19 of 19** generator runs in the
preceding 72 h cleared it and completed. Exactly one run has been refused, and it is script-specific
(that tool's generated script declared into global scope). **So P2 is unexercised, not broken, and
not blocked** — the next generator run that produces a provable script will exercise the door,
provided its fence drives inputs.

⚠ **And P2 needs a narrower first fire than P1 did.** The door is gated on
`DrivesButAssertsNothing()` (`write_doc_plan_action.go:219`) — the fence must drive inputs **and**
assert nothing. P1's first-fire subject drives nothing, so it could never have tripped P2. On the
refreshed census the qualifying rate is **55 / 187** of `tool-generator`'s fences, so roughly one
run in three should do it.

### The census, re-run (the handoff says not to trust the old one, and it had moved again)

`[MEASURED 2026-09-03 16:04Z]`, `is_current` fences, per author:

| created_by | fences | assert_no_value | uses_computed_values | drives_inputs | drives_but_asserts_nothing |
|---|---|---|---|---|---|
| `tool-generator` | **187** | 116 | **0** | 91 | **55** |
| `operator:bugfix224-session` | 16 | 0 | 16 | 16 | 0 |
| `webdesign_couk_thread` | 14 | 4 | 0 | 6 | 0 |
| `operator:mortgagecalculator-…-701-rekey` | 8 | 0 | 8 | 8 | 0 |
| `operator:staged_component_build` | 8 | 0 | 6 | 7 | 0 |
| (7 smaller authors) | 8 | 4 | 0 | 4 | 3 |

**The sharpest single line in this lane, and it is now exact: `tool-generator` has authored 187
fences and `uses_computed_values` = ZERO.** Every one of the estate's ~~38~~ **30** value-asserting fences was
> **CORRECTED 2026-09-03 ~16:4x — the count was 30, not 38, and it was MY ARITHMETIC, not a stale
> census.** 16 (`bugfix224-session`) + 8 (`mortgagecalculator-…-701-rekey`) + 6
> (`staged_component_build`) = **30**. I took `staged_component_build`'s *fences* figure (8) from
> the adjacent column instead of its *uses_computed_values* (6), and mis-added on top. Caught by a
> second query run for a different purpose — `count(*) FILTER (WHERE body LIKE '%expect_values%')`
> grouped by `subject_type` returned **30** and disagreed with the total I had already published in
> four places. **The conclusion is untouched: `tool-generator` accounts for ZERO of them.** The
> number was decoration on that finding and I still got it wrong. `WRONG_CALLS.md` has it.

written by an operator or a lane, never by the agent. That is the bug stated as a census rather than
as an argument, and it is what P4 exists to change.

`max(created_at)` for `tool-generator` is **12:35:59Z today** — still a live intake, not a backlog.

## 2026-09-03 (~16:15Z) — P4's blockers: TWO DISSOLVED. The mcalc lane had already answered; the handoff said "unanswered" because nobody re-read the file

The handoff's §5/§6 record three questions to the `mortgagecalculator_couk_adoption` lane and three
to `loancalculator`, all **unanswered**. That was true when written and is **wrong now for mcalc**:
they replied *in place*, by appending a `# REPLY, 2026-09-03` section to the bottom of the CONTRIB
this lane put in their directory —
`docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/CONTRIB_2026-09-03_from_the_449_lane_I_am_taking_the_FRAMEWORK_half_you_keep_the_site_half.md`
(the reply starts at line 113). Their own `HANDOFF_2026-09-03_continue_here.md:20` points at it:
*"Reply and the full division: `CONTRIB_2026-09-03_from_the_449_lane_…`"*.

⚠ **The lesson, and it is cheap: a CONTRIB you send is a file someone else can APPEND TO, so the
reply arrives with no notification and no new filename.** This lane checked for *new* files in their
directory (`ls -lt`) and the reply was invisible to that, because the file it lives in kept its
original 12:44 mtime relative to their 12:45 handoff. **Re-read the CONTRIB you sent; do not look
for a new one.** Logged in `WRONG_CALLS.md`.

### Blocker 1 — 441's landing order: DISSOLVED, and the answer inverts the plan's caution

Their answer, quoted: **"`441`'s fix is not imminent and nothing is scheduled. I filed it; I am not
building it. … Treat '441 lands first' as unavailable."** And then the part that matters:

- **A fence written AT BIRTH is safe.** The generator emits selectors from the template it has just
  written, and the tool renders from that same template — `ScopeToolBirthTemplate`'s contract is that
  a tool carries its template verbatim as `rendered_html`. There is no window in which they disagree.
- **Backfilling existing tools is where the 441 risk lives** — and they verified the mechanism this
  lane described rather than repeating it: `runComputedValues` does `page.Count(sel) == 0` →
  `problems` → **fail**, it does not skip.

So: **ship the authoring fix for NEW fences (P4); do NOT backfill the 55 standing blind ones.** That
splits what this lane had been treating as one blocked change into one unblocked half and one
deferred half.

⚠ **Their caveat, which is worth carrying:** a fence correct at birth is **not** safe for ever — on
their own site, migration 701 adopted 11 tools with bare ids, the instance-scope sweep converted them
at 07:40 today, and five re-renders published new ids at 08:46–08:49, **breaking five fences**. That
is 441's problem, not 449's, and it does not change the shipping order — but P4 will make 441 *more*
visible, which they read as a feature and so do I.

### Blocker 3 — `no_auto_fix`: DISSOLVED. Tier 2 never evaluates `computed_values`

**I did not take this on their word — the whole risk of P4 turns on it, so I read the arm.**
`evaluateStaticCriteria` (`platform/orchestration/actions/discovery_checks/check_tool_acceptance.go`)
switches on `ch.Type` with arms for exactly `selector_exists`, `selector_count`, `interaction`,
`asset_loads`, `page_status_ok`, `attribute_absent`, `attribute_matches` — and its outer arm is:

```go
default:
    skip(ch.ID, ch.Type+" is not statically checkable (Tier 4)")
```

`computed_values` is not an arm, so it **falls to that default and is SKIPPED at Tier 2**. A skip is
neither a pass nor a fail, so adding a value assertion to a fence **cannot** arm Tier 2 to dispatch
`tool-improver`. The LANDMINE's other half is confirmed too, in the very next line of that file
(`:94`): *"Built-in shell checks — always run, independent of the criteria."* Those three fire
regardless of fence content, so the shared-component exposure of `bugs_closed/285` is **pre-existing
and orthogonal** — installing any PLAN at all switches it on, and `computed_values` widens it by
nothing.

**Action anyway, and their reason is better than the risk argument:** set `no_auto_fix: true` on any
generated fence carrying a value assertion — not to close the Tier-2 path, which it cannot, but
because *"the only way an automated rewriter can turn a red arithmetic fence green is by changing the
numbers on a page quoting tax and consumer credit."* An arithmetic failure means the maths or the law
moved; that is a human's call.

### Blocker 2 — where an expected value may COME FROM: still open at `loancalculator`, but no longer the thing holding P4 up

No reply from that lane (its newest file is this lane's own CONTRIB at 14:41). But the answer is
already legible from two sources this lane can read directly:

1. **`runComputedValues`' own contract**, which I read in full rather than citing
   (`internal/adapters/browserrunner/run_checks_action.go:790-808`) — and it is unambiguous that
   this type is a **regression** check, not a birth check:

   > *"The values are not authored by hand and are not judged for correctness here. They are CAPTURED
   > from the tool while it is known good (`toolgolden.py --emit-criteria`) and then defended: this
   > check's claim is 'the arithmetic has not moved since it was captured' … It follows that a golden
   > captured from an already-wrong tool pins the wrong answer — **the capture script therefore
   > refuses to emit for a tool whose outputs do not react to its inputs**, and the capture is only as
   > good as the state it was taken in."*

   It also states why the fail-not-skip design is safe there and would not be at Tier 2: *"this runs
   post-settle in a real browser, so 'no element matches' means the element genuinely is not there."*

2. **The mcalc lane's inherited warning**, which lands in the same place: *"a generated
   `computed_values` check must not pin whatever the tool printed at birth … the expectation needs a
   source other than the tool — otherwise the fix pins today's bugs as tomorrow's specification."*
   Their `verify_criteria.py` re-derives at three labelled strengths (DEFINITION / REGISTER /
   CONVENTION), and it **refused them** on exactly this: it reports *"NOT VERIFIED (no independent
   model): fact-finder, portfolio"*, which is why `tool-portfolio` still has no fence.

**So the design conclusion for P4 is settled even without `loancalculator`'s reply, and it is the
handoff's own §5 sentence, now evidenced twice over: the REFUSAL ARM IS THE LOAD-BEARING HALF.** If
the generator cannot derive an expectation from something that is not the tool it was just shown, it
must emit **no** `computed_values` check and say so in Dependencies. What is still genuinely open —
and is a question about generalisability, not a blocker — is whether the three-strength taxonomy
survives outside a domain with published formulae (mortgages) or a legal register (SDLT).

## 2026-09-03 (~16:4x) — P4 groundwork: the shape is pinned, and P4 is BIGGER than "two prompts"

Read everything P4 has to edit before writing any of it. Four findings, two of which resize the phase.

### The cause, now quotable verbatim rather than described

`tool-generator`'s `{workflow,steps,compose_plan,config,prompt_template}` is **2,766 characters / 2,782 bytes, 45
lines** `[MEASURED 16:2xZ]`. The whole fence vocabulary is one paragraph (line 34), and it is a
**closed enumeration**: four mandatory checks (`selector_exists`, `no_console_errors`,
`page_status_ok`, `no_horizontal_overflow`), then *"Add ONE interaction check ONLY if you can copy
real ids or classes from the HTML above"*, then the sentence that shuts the door:

> *"No other check type exists for interactions — never emit `"type":"click"` or `"type":"fill"` as
> a check type."*

`computed_values` is never named anywhere in the row. **That sentence is also the natural verbatim
anchor for the migration's pre-guard** — distinctive, one occurrence, and it is exactly the claim
the change has to amend.

### FINDING 1 — `experience-planner` carries the fence in THREE steps, not one

`[MEASURED 16:3xZ]` Steps whose `prompt_template` contains a ` ```criteria ` block:
**`compose`, `recompose`, `reframe`** (its five `review_*` steps do not). The old plan said "the
equivalent on `experience-planner`", singular. **It is three JSON paths on that row**, and a
migration that edits one leaves two authoring the old vocabulary — the `099`-mirror failure mode, one
level down.

### FINDING 2 — and `experience-planner` is a THREE-FENCE population, so it is not where the value is

`[MEASURED 16:3xZ]` `is_current` plans carrying a fence, by subject:

| subject_type | plans | with_fence | with_value_assertion |
|---|---|---|---|
| `tool` | 241 | 241 | **30** |
| `component` | **55** | 55 | **0** |
| `experience` | **3** | 3 | 0 |
| `experience-pattern` / `action` / `pipeline` | 14 | 0 | 0 |

**`experience-planner` has authored 3 fences, ever.** So the cost/benefit inside P4 is lopsided:
`tool-generator` is 187 of the 241 and every blind driving fence; `experience-planner` is three
documents. **Recommend: ship `tool-generator` first, alone, and treat `experience-planner`'s three
paths as a separate follow-on** — it triples the anchor surface for ~1% of the population, and
`732`-shaped guards are per-anchor.

⚠ **And the population P1/P2/P3 cannot see at all: `component`, 55 fences, ZERO value assertions.**
Every one is `operator:staged_component_build`. P2's door is gated on `subjectType == "tool"`
(`write_doc_plan_action.go:218`) and the sweep and the Tier-4 verdict line are tool-scoped too, so
**a component fence that asserts nothing is invisible to all three halves of this lane's shipped
work.** Not in scope for 449 as filed, and I am not widening it here — but it is the same defect in a
population 30% the size of the one we are fixing, and nobody is counting it. Belongs in §5 as a new
candidate.

### FINDING 3 — the JSON shape, pinned from the struct AND from a live worked example

From `criteriaCheck` / `criteriaStep` (`run_checks_action.go:221-246`) and confirmed against a real
`operator:staged_component_build` fence:

```json
{ "id": "<kebab-id>", "type": "computed_values", "profiles": ["desktop"],
  "steps": [ { "action": "fill", "selector": "#volume", "value": "5000" } ],
  "expect_values": { ".result-card.highlight .result-value": "$2,000.00" } }
```

- `expect_values` is a **map** selector → the exact text that selector must read after `steps` run
  (per-element and order-free; the runner sorts keys so a bounded failure message is stable).
- Step actions are `fill` (with `value`), `click`, `select` (with `value`), `reload`.
- Text comparison is `collapseSpace` on both sides — **whitespace-insensitive, everything else
  exact**. So `"$2,000.00"` must carry its currency symbol, separators and 2dp exactly as rendered.
- `no_auto_fix` / `no_auto_fix_reason` are **top-level fence keys**, not per-check
  (`acceptanceFenceFlags`, `tool_acceptance_actions.go`).

**That worked example is also the proof the derivation is doable at birth without reading the tool's
output:** 5000 × (3.85 − 3.45) = **$2,000.00**, margin/unit **$0.40**, annual = 2000 × 52 =
**$104,000.00**. All three follow from the spec's arithmetic. Nothing there was copied off the page.

### FINDING 4 — the 3000-character cap is a real constraint but NOT a blocker; there is headroom

`compose_plan`'s step config `[MEASURED 16:3xZ]`: `model=claude-sonnet-5`, **`max_tokens: 4000`**,
`input_fields: [input_data, site_record, generated_html]`, `output_format: text`. The prompt's last
line instructs *"Keep the whole document under 3000 characters"*.

A `computed_values` check in the shape above costs **~350-450 characters** of the output document.
3,000 characters is ≈750-1,000 tokens against a 4,000-token ceiling, so **the instructed cap can go
to ~3,500 without coming near truncation** — which is the right move, because leaving it at 3,000
makes the model trade the new assertion against prose it was also told to write, and the handoff's
warning ("it trades a value assertion for a truncated document") is exactly that failure.
⚠ I am **not** adding a template variable, so the `input_fields` landmine does not apply here.

### The design rule I will actually write, and why it is narrower than "teach the type"

The generator's only inputs are the spec (`function`/`name`/`description`) and
`{{.generated_html}}` — **the very artefact whose correctness is in question.** It has no register,
no formula source, no site facts. So a naive "emit `computed_values`" instruction can only produce a
value read off the tool, which is the pinning failure `runComputedValues`' own docstring warns about.

The one source that *is* independent and *is* present: **the model's own knowledge of a published
formula**, applied to inputs it chooses. That splits the tools cleanly:

- **Derivable (DEFINITION):** the tool computes a published, checkable rule — annuity repayment,
  compound interest, VAT at 20%, BMI, unit conversion, the fuel-margin arithmetic above. The
  generator picks the inputs, works the arithmetic itself, and the expectation never touches the
  page. This is where a wrong divisor or a dropped rate conversion gets caught.
- **Not derivable (must REFUSE):** an arbitrary scoring heuristic where "correct" is definitionally
  whatever the code says — `tool-idea-stage-identifier`, `tool-process-automation-scorer`. There is
  no independent oracle, so a `computed_values` check could only pin the implementation.

**So the instruction is conditional and its default is refusal**, and it must require the arithmetic
be shown in `## Dependencies` so a reviewer can see which of the two cases the generator thought it
was in. That is the refusal arm the council and both peer lanes converged on, expressed in the only
terms this prompt actually has available.

> **CORRECTED 2026-09-03 ~17:0x — "2,783 bytes" above was two errors in one small figure, and the
> units one is the transferable half.** The live value is **2,766 characters / 2,782 bytes**
> (`length()` vs `octet_length()`; the prompt carries **8** em-dashes, 3 bytes each, so bytes exceed
> characters by 16). My 2,783 came from `wc -c` on a `psql -At` dump, which also counts the trailing
> newline psql adds. **Why it is worth correcting rather than shrugging at: the cap this migration
> moves is stated in CHARACTERS** ("under 3600 characters"), and a PLAN document is full of em-dashes,
> £ signs and ⚠ — so anyone checking cap compliance in bytes will read ~10% high and "fix" a document
> that was never over. This is `MEMORY.md`'s own byte-vs-char trap arriving in a new place. Caught by
> the round-trip test printing `length()` = 2766 against my published 2783.

## 2026-09-03 (~17:0x) — council round 1 REVISE, and the loancalculator lane answered. Both improved the file

Two things landed within minutes of each other, and folding them into one round-2 revision was luck
of timing rather than planning.

### Round 1: REVISE, gated by editquality — and the defect was in my SUBMISSION, not my file

> *"The sketch's actual UPDATE statement substitutes the anchor text with the literal
> `$new$<<<the anchor sentence, then the block below>>>$new$` — a placeholder marker, not the real
> prompt text ... applying it inserts a broken string into the live prompt rather than teaching
> computed_values at all."* — editquality, **HIGH**

**The objection was right and the file was fine.** The shipped migration always carried the real
dollar-quoted text — that is what the 2766→5046 round trip measured — but I had put a placeholder in
the *sketch* and moved the real text to trailing `--` comments to keep the submission short. The
seat could not distinguish "documentation shorthand" from "a migration that writes a literal
placeholder into the live prompt", and it was right not to guess.

⚠ **The RUNBOOK already says "reviewers judge the sketch; it is the only view of your code they
get", and I read that line before submitting and still economised on exactly the part under review.**
Fixed structurally rather than by retyping: **the round-2 sketch is the file itself, sliced from
`BEGIN;` to `COMMIT;`.** Sketch and artefact now cannot diverge. That also answers debug_historian's
LOW objection that no `BEGIN`/`COMMIT` was visible.

### The objection that changed the CODE: no `snapshot_agent()` — raised independently by TWO seats

`tooling_provenance` and `debug_historian` both MEDIUM: mutating `agent_definitions.default_config`
with no pre-update snapshot, "a byte-swap rollback file is not equivalent to a queryable pre-image
row". **Correct, and it is the house convention** (`731`, `734`, `741` all call it).

Added: `PERFORM snapshot_agent('tool-generator', '749_...: pre-update')` in the pre-guard, **after**
the idempotency `RETURN` so a re-run does not mint a pointless snapshot.

⚠ **And I nearly recorded it working on a count that said it was not.** My first check counted
`agent_definitions WHERE is_snapshot=true` and got **0** while the NOTICE said "Snapshot captured".
Rather than shrug at the contradiction I read `pg_get_functiondef(snapshot_agent)`: **it copies the
row into a SEPARATE TABLE, `agent_definitions_backup`**, and does not touch `agent_definitions`.
Measured properly, in the rolled-back transaction:

| | backup rows for `tool-generator` |
|---|---|
| before | **0** |
| after first apply | **1** |
| after second apply | **1** (idempotent — no pointless snapshot) |

and the stored pre-image reads back at **length 2766**, exactly the pre-change prompt. **A recovery
point you cannot query is not a recovery point** — the count that disagreed with the notice was the
useful signal, not noise.

### The objection I had EARNED: containment checked one function and generalised

`prior_art_librarian`, MEDIUM: my "Tier 2 cannot dispatch `tool-improver`" claim rested on
`evaluateStaticCriteria` alone, and a landmine names a second route
(`component_level='tool'` → `check_tool_health` → `improve_tool`) that is **not** gated by
`no_auto_fix`.

**This is `an-objection-naming-one-file-is-naming-a-category` exactly, and I walked into it.** The
claim survives — but only after enumerating every producer of an `improve_tool` item:

| producer | reads the fence? | effect of adding `computed_values` |
|---|---|---|
| `check_tool_acceptance.go:274` (Tier 2) | yes | none — no `computed_values` arm, falls to `default: skip`; its 3 shell checks run "independent of the criteria" |
| `check_tool_health.go:308` | **NO** — `grep 'criteria\|doc_plans'` = **0 hits** | none — it cannot see a fence |
| `refresh_evidence_fact_drift.go:415` | yes | none — honours `no_auto_fix`, routes to `fact_drift_review` |
| `tool_acceptance_actions.go` (Tier 4) | yes | none — `no_auto_fix` fence "never reaches tool-improver at all" (`:38`) |
| `confirm_work_item_handler.go:98` | n/a | human-initiated, not automatic |

**Zero rows — now by enumeration rather than by one grep.** The seat was right that the argument was
unsound even though the conclusion was true.

### The objection I could NOT close, and did not pretend to

`bug_historian`, MEDIUM: the refusal arm is prose in a prompt, with no code-side check that the
asserted value is genuinely derivable — *"nothing stops the same model that just built a heuristic
scorer from convincing itself the score is a published formula"*. **Accepted as stated.** A prompt
cannot detect that. What round 2 does instead is (a) make the honest path reachable — see below —
and (b) name the enforcement point as a follow-on: a detector comparing a newly authored
`expect_values` against what the tool itself renders, flagging the case where they MATCH while
`## Dependencies` names no rule and shows no working. That is mechanically checkable, and it is
deliberately not in this migration.

### The peer lane's answer, which changed the prompt text

The `loancalculator_couk` lane replied in `bugs_open/449` §§A–C (asked 2026-09-03; they explicitly
did not take the bug). Two findings changed the file:

1. **"An input vector is PART OF THE EXPECTATION'S IDENTITY."** Their bug 385 established that stored
   and composed bytes are **different name-spaces on this exact pipeline** — and my generator reads
   *draft* markup while the checker drives the *deployed* artefact. Added: **write the inputs as
   LITERAL values, never "the page default", never a reference the checker resolves later.**
2. **The format/locale trap**: fill `300000`, the page displays `300,000`. Since the comparison
   collapses whitespace and is otherwise exact, **a correct derivation can fail on presentation.**
   Added: **DERIVE THE VALUE FROM THE RULE; READ THE FORMAT OFF THE CODE.** These are separable — the
   value is an arithmetic fact about the world, the format is a presentation fact about the code —
   and without the split a model that cannot predict the format must either refuse or read the whole
   value off the page. **The second is precisely `bug_historian`'s failure mode, so this change is
   load-bearing for that objection too, not cosmetic.**
3. **Not taken, filed instead:** their **relational assertions** rung (monotonicity, sign, bounds),
   derivable from a tool's PURPOSE with no known-good state, which would catch the
   0%-APR-computes-nothing class that reactivity passes. It is a better answer than refusal for many
   of the tools this prompt now teaches to refuse — **but it needs a NEW check type in the runner**,
   so it is a follow-on with its own footprint, not something to smuggle into a prompt migration.

They also warn the label vocabulary must be **single-sourced** (one shared enum, not three lanes'
private strings) or we re-mint their `LOCK-009` drift, and that **REGISTER** should mean "derives
from a fact id in the site's `evidence_base`" — which this step cannot reach anyway
(`input_fields` are `input_data`, `site_record`, `generated_html`), so 749 licenses **DEFINITION
only**. Both recorded for the follow-on.

**Round 2 resubmitted on the same correlation** (`RESUBMIT_CORR=dda64bd1-…`) so the trail
accumulates. Re-tested after every change: **2766 → 6013 → (re-apply: `UPDATE 0`) → 2766.**
