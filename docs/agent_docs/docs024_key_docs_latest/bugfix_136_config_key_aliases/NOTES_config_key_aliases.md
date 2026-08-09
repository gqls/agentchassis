# NOTES — `bugfix_136_config_key_aliases` (append-only, newest at the bottom)

## 2026-08-08 — session 1: picking the bug

Triage of `bugs_open/` against 32 live `.jsonl` transcripts. The useful discriminator was
**not** `who-owns.py`, which returns "OWNED or recently active" for almost everything on this
estate because almost every workstream directory is active. Two things worked:

1. `git log -1 --format=%cd -- bugs_open/<file>` per bug, sorted — a bug whose file has not
   moved in ≥9 days is a real signal.
2. Grepping the **tail** of each recently-touched transcript (`tail -c 400000`) for
   `bugs_open/NNN` and taking the top few by count. A session's *current* focus is legible
   that way; a whole-file grep is not, because every session that ran `ls bugs_open` matches
   everything.

Rejected after checking: `093` (its own last update says it is no longer a code task — it is
blocked on `083`, and `083` is hot in two sessions) · `211` (`who-owns` OWNED, the 122 lane
committed to it today, and one session had 62 hits on it) · `085`, `181`, `185`, `189`, `203`
(fixed and live; only site-level verification owed) · `114`, `126`, `146` (owning lanes active
within 8 days) · `040` (lane dormant, but the bug is infrastructure and the metrics predate
any claim).

**A trap worth naming**: `b5a58a2b` showed 20 hits on `136`'s vocabulary, which read as a
second session on this bug. Every hit was the auto-memory file
`bugfix-136-domain-pipeline-rename.md` being loaded as context, not work. **Grep the hits,
do not count them** — a memory file mentioning a bug looks exactly like a session working it.

## 2026-08-08 — the bug was still valid, and worse than filed

Re-measured before planning. The three named instances all still present in the live audit
(4 UNKNOWN KEYS, three of them this bug's). Two figures had moved since 2026-07-28:
`target_domain` is carried by **1** step, not 3 (migration 286 removed two under RFC 006), and
`item_domain` by 9, not 7.

Then the finding that changed the priority. §2a's `[MEASURED] Nothing is mislabelled today`
is **false as of 2026-08-04**: `content_duplication` and `page_canonical_collision` — both of
which propagate `dctx.Pipeline` — have joined `completeness-discovery-agent`, and four rows
are filed under `design` on an agent whose config says `content`. The bug file predicted this
exact trigger in the sentence immediately after the measurement. **A figure that depends on
which checks are registered where has a shelf life of days here**, and §2a did not mark it.

## 2026-08-08 — the framework finding, which is the whole shape of the fix

Reading `action_inputs.go` to see whether `ActionInputSpec.Deprecated` could carry these
renames: **it cannot, and would be worse than nothing.** Strategy 3 does
`ExtractNestedField(collectedData, config[oldKey])` — the old key's *value* is a dot-path.
That is a reference alias. A literal setting placed there resolves nothing, takes the
default, and — because `UnknownConfigKeys` recognises `Deprecated` keys on purpose — silences
the detector as well.

This answers the question the bug file asked and could not answer: *"They wrote it on one
action and not on the other two."* There was no honest declaration available, so two of three
authors correctly did nothing. The per-action shim was never the fix; it was the symptom.

## 2026-08-08 — missteps

1. **`grep 'config\["'` to enumerate an action's config reads.** Missed `priority`, which
   arrives via `GetIntField`. Caught by the planner before it reached a doc. Logged in
   `WRONG_CALLS.md`; the check is in the LANDMINE that shipped with the fix. I made this
   bug's own mistake — a key read in a way the obvious search cannot see — while writing its
   fix.
2. **A mutation that did not compile is not a mutation proof.** My first attempt at killing
   `TestResolveConfigSettingPrecedence` (`if canonical == key` → `if false`) left `old`
   unused, so the package failed to build and the test "failed" for the wrong reason. A build
   failure proves nothing about a test: it proves the compiler works. Redone as
   `canonical == key && false`, which compiles and makes the rule inert — and the test then
   failed on exactly the row that exists for the rule.
3. **`go build` at HEAD vs `go test` in the tree is not a comparison.** Chasing whether a
   `thunder` failure was mine, I built the package from `git archive HEAD` and it passed —
   which looked like proof I had broken it. `go build` does not run vet, and the failure was
   a vet diagnostic. `go vet` at HEAD reproduces it identically. Match the instrument before
   attributing the difference.

## 2026-08-08 — what shipped, and what was deliberately left

Committed `3f93456fd`, 13 files, `Council-Submitted: 433de2c0-682f-4d8d-8c48-28637309f1ba`.
Five guards, each killed by a deliberate mutation and restored:

| mutation | test that died |
|---|---|
| alias loop made inert (`&& false`) | `TestResolveConfigSettingPrecedence` — the alias-only row |
| precedence swapped (alias beats canonical) | same test — the both-set row |
| `DeprecatedConfigKeysInUse` gated on `checksConfig()` | `TestDeprecatedConfigKeysInUseIndependentOfOptIn` |
| alias loop removed from `UnknownConfigKeys` | `TestUnknownConfigKeysRecognisesDeprecatedConfigKeys` |
| triage action body reverted to the direct config read, **spec left declaring the alias** | `TestTriageDetectedItemsHonoursTargetDomain` (sqlmock, asserts the value reaches the UPDATE) |

That last one is the one worth keeping: it is the only test in the set that catches a
declaration which has stopped describing the code, which is this bug's own shape.

**Live verification that needed no roll**, run against production `agent_definitions`:
UNKNOWN KEYS **4 → 1**. The survivor, `plan_sections: domain`, is left on purpose —
`page-build-handler` is hot with several sessions today and the UNKNOWN line is the honest
record until its owner deletes the key.

Not done, all recorded on the bug file: the four mislabelled rows are not repaired (two sit
in `detected`, a queue with no consumer — `bugs_open/083`); `summary_template` is still
biting and is **not** an alias case (aliasing it to `summary` would ship a raw
`{{.input_data.topic}}` to a human reviewer); `create_work_item`'s full opt-in; every
definition edit.

## 2026-08-08 — council APPROVED, and the objections were better than the verdict

`433de2c0` — approved, 2 advisory objections, none high-severity, 12 seats, ~8 min. Full
trail in `bugs_open/136` §7. The `architecture` seat confirmed the scope call explicitly:
additive/inert/opt-in field + same-commit LANDMINE and register entry = **normal gate, not an
RFC**, per the 2026-07-29 owner ruling. The registration is what earns that, not the size of
the diff.

Three seats independently found things I had to go and fix or answer:

- **`guardian`**: the consumer search was "a grep over Go, not over every consumer that might
  exist in a definition". True — a `query_database` step filtering `pipeline='design'` would
  not appear in a Go grep. Closed both surfaces; nothing names `design` or `content`.
- **`debug_historian`**: the pod-grep recipe was positive-only and used `-l app=agent-chassis`.
  **Measured: that selector returns 2 pods and 25 RUNNING pods carry the image (34 including
  non-Running) — 8% coverage.** RUNBOOK recipe replaced: enumerate by IMAGE, and carry a
  positive control (`Using deprecated config pattern`, Strategy 3's long-live warn) plus an
  invented negative. Pre-roll baseline banked: `new=0 pos_control=1 neg_control=0`.
- **`reuse_agent`**: "confirm no other ad-hoc alias shim exists that should have converged."
  It does. `resolveAgentTypeForSpawn` (`spawn_actions.go:3154-3163`) hand-rolls
  `group_type` → `agent_type`, a literal-setting alias, this exact class. **Measured before
  deciding: `group_type` is set by ZERO live steps**, so it guards a shape nobody writes.
  Recorded as a convergence candidate with no live exposure rather than pulled in.

**`editquality` was right about the submission and wrong about the world**, and the gap is
worth naming: it objected that the parity test, the behaviour-preserving test, the LANDMINE
and the register entry are *"claimed in the rationale but absent from the edits list — if they
exist they should be edits; if not, the risk mitigations are fictional."* All four shipped in
`3f93456fd`. The cause is the schema's **8-edit cap**: I spent all eight slots on code and
described the rest in prose. **A reviewer can only see the edits list. A mitigation named only
in the rationale reads as fiction, and it is not the reviewer's job to assume otherwise.**
Next submission: spend a slot on the test file, or say in terms "shipped in the same commit,
not listed, cap reached".

## 2026-08-08 — the second misstep, and it is the same shape as the first

Logged in `WRONG_CALLS.md`. I claimed the four mislabelled rows were doing measurable damage,
citing `countDispatchableWorkItems`' `WHERE ... AND pipeline = $2`. **That query has one
caller and it always passes `"build"`** — so it cannot distinguish `design` from `content` and
cannot be evidence about confusing them. Every live pipeline-filtering consumer, Go and
definition alike, names `build`, `reports` or `diagnose`. The rows are mislabelled; the
demonstrated cost today is nil.

What caught it: going to answer the *opposite* question — whether anything would BREAK when
the fix moved those rows. The enumeration that proves a change safe is the same enumeration
that grades the harm claim, and I had done it in only one direction.

**Both of today's missteps are one shape**: I trusted the code in front of me over the
enumeration of what reaches it. A `config["` grep hid a key read through a helper; a `$2` in a
predicate hid a caller that always passes a constant. **A parameterised filter tells you what a
query CAN discriminate, never what it DOES.** Corrected in place at `bugs_open/136` §6, which
is the version to quote — §1's "the harm is not cosmetic" is superseded and says so.

Both corrections went in *after* the council had my overstated version. The `guardian` seat
found the same gap independently, from the submission alone, which is the argument for
submitting before you are certain rather than after.

## 2026-08-08 (late evening, cold-start session) — §9's "where do these executions log?" ANSWERED, and the answer kills log-sweeping as a witness

The executions log exactly where they should: **stdout of an ephemeral per-agent pod**
(`agent-<type>-<agent_id[:8]>-<hash>` — the tool-improver row of 21:18:16Z was filed from
`agent-tool-improver-3f5db2ad-pqjn9`, which was still Running when I looked). The sweep in §9
found nothing because **two independent mechanisms destroy the evidence within minutes**:

1. **Container log rotation eats the lines within seconds.** That pod's retrievable log
   *starts at 21:18:21.613Z* — five seconds AFTER the work-item row landed (21:18:16.17). The
   chassis's `DEBUGaa` state dumps are so large (223 retrievable lines ≈ tens of MB) that the
   rotation window during an active orchestration is measured in seconds. The
   `create_rerender_item` execution WAS logged and was already rotated away by the first
   opportunity to grep. A `--since=3h` sweep is theatre against a <1-minute retention.
2. **`agent-job-cleanup` deletes Completed agent pods within minutes** (a CronJob, visible
   completing every few minutes), so for short-lived carriers the pod itself is gone. Measured
   live: the domain cascade ran classifier→exemplar→strategist→briefing 22:09→22:17Z, filing
   four carrier rows, and at 22:18:13Z **zero** of those pods existed.

So §9's option 1 (find the logs) is answered and closed: the logs exist, and no after-the-fact
sweep can reach them. A 5s polling watcher over carrier pods is armed this session as a
belt-and-braces (`alias_witness_watcher.sh` in the session scratchpad), but rotation can
outrun even that during a dump-heavy step.

## 2026-08-08 (late evening) — the deterministic witness, designed and fired BEFORE reading the result

§9's option 2 (edit a live carrier to `item_domain: "content"`) touches another lane's agent
and risks a real row landing outside `pipeline='build'`, which `countDispatchableWorkItems`
and the stale-item reaper DO filter on. Instead: a **one-shot agent definition I own**
(`alias-witness-136`, category `test`), whose single `create_work_item` step carries ONLY the
deprecated `item_domain` key with the NON-default value `content`, filing an item that is
**born `status='cancelled'`** — outside `idx_swi_dedup`'s partial index, outside every
dispatcher's status predicate, and NOT `detected` (which `triage_detect_items` launders to
`pipeline='build'` — 090 trigger header, note 4). Anchored to `system.internal`, the 090
pseudo-site. No LLM steps, no credits.

**The disconfirming observation, written down before firing** (the rule this lane's third
wrong call bought): row `pipeline='content'` → alias honoured at runtime; row
`pipeline='build'` → the alias fell through to the default in production — each outcome
refutes the other, unlike every observation available from the nine live carriers, whose
configured value equals the default.

Fired 2026-08-08 ~22:40Z, `WITNESS_CORR=6e89cad7-29e1-4a9d-9d9c-c0a66b79d0cb`, publish
confirmed by the `PUBLISH_OK` marker pattern (payload in the container COMMAND — the kcat
stdin trap). Dispatch queues behind the shared generic lane: budget ~30 min. Cleanup owed
after the read: deactivate the definition; the witness row stays (it is the evidence and is
inert).

## 2026-08-08 (night) — the witness landed: pipeline='content'. WITNESSED.

Row `1d46cf49` at 22:25:39Z: `pipeline='content'`, `status='cancelled'`,
`item_key=alias_witness_136_eac60db8`. The config carried only `item_domain: "content"`;
the default is `"build"`; the row reads `content`. The alias is honoured at runtime in
production on v1.0.1268. Orchestration `8a46f229` COMPLETED on
`agent-chassis-778b7c77c7-rd27g` — the generic lane runs on the long-lived chassis
deployment pods, not an ephemeral pod, which is worth knowing for the next dispatch trace.
Dispatch→row ~3 minutes (quiet lane; budget 30 anyway).

The corroborating warn line was already unreachable ~2 minutes later on that same still-
running pod — measured the retention: oldest retrievable line was **0.4 seconds old**. That
number closes §9's log question permanently and became a LANDMINES entry.

Cleanup done: definition `alias-witness-136` deactivated + soft-deleted 22:27:46Z
(`remaining_active=0` verified). The witness row stays — born-terminal, no consumer, it is
the evidence. Bug file §11 written. The one thing the 2026-08-08 evening handoff said was
owed is delivered.

**Provenance footnote on the LANDMINES entry (2026-08-08, minutes later):** the "chassis pod
retrievable log <1s" entry reached HEAD as a same-file passenger in ANOTHER session's commit
`1eae32644` (fix(226), 22:30:58Z) — it landed between my append and my commit, so my own
commit `745ed93d8` says "New LANDMINES entry" while carrying zero LANDMINES diff. The entry
is live at HEAD and synced to doc_notes; only the attribution is odd. This is the
`a-pathspec-commit-still-takes-a-same-file-passenger` mechanism running in the direction
that costs nothing.

## 2026-08-09 — owner: "we can fix those deferred items now" — handoff written, facts re-measured

`HANDOFF_2026-08-09_deferred_items.md` is the cold start. Re-measurement while writing it:
the four mislabelled rows are now **THREE** — the 2026-08-03 `capability_gap` (`detected`)
no longer exists in `site_work_items` at all (not repaired; gone; cause not investigated —
the handoff says re-derive, not trust). Post-roll mislabels from
`completeness-discovery-agent`: **0** (fix holding). Deprecated-key carriers unchanged
(item_domain 9 / check_domain 3 / target_domain 1); `item_pipeline` already carried by 2
steps (claims-auditor, site-work-orchestrator). `spec_fields` 1 carrier, `domain`-on-
`create_work_item` 1 carrier (claims-auditor — the Strategy-1 landmine applies, adjudicate
before convicting). Next free migration number: **347**. Note `created_by`, not `source`,
is the discriminator for the mislabelled rows — `source='discovery'` matches the legitimate
design agent's rows too.

## 2026-08-09 (afternoon) — items A/B/C shipped, item D shipped, and the lane's own census query was wrong

Owner instruction was "we can fix those deferred items now". A, B and C are live
(migration **349**, applied by hand and recorded); D's data half is live (**350**) and its
code half is committed and submitted to the council. Item E is deliberately skipped, with
a reason that is not the one the handoff assumed. Migration numbering: the handoff said
"next free 347" and by the time I got there 347 and 348 were taken by other threads — I
used 349/350. Re-derive that number, never carry it forward from a doc.

**Acceptance, banked.** `./scripts/audit-config-keys.sh` before: `UNKNOWN KEYS:
plan_sections: domain` and three DEPRECATED families. After: **`UNKNOWN KEYS: none`**,
**`DEPRECATED KEYS: none`**, exit 0. That is the whole lane's headline number reached.

### The misstep that matters: the RUNBOOK census I trusted was blind, and it looked complete

The handoff's item C table said **13 live carriers** of the three old key names. It was
built from the RUNBOOK's census SQL, which walks `->'workflow'->'steps'`. The real number
is **19**. Six live inside a loop step's `sub_workflow.steps`:

```
component-quality-auditor  create_regen_items > sub_workflow > steps > create_work_item
internal-linker            create_items_loop  > sub_workflow > steps > create_rewrite_item
tool-auditor               create_items_loop  > sub_workflow > steps > create_improve_item
tool-auditor               create_items_loop  > sub_workflow > steps > create_review_item
tool-suggester             create_items_loop  > sub_workflow > steps > create_library_item
tool-suggester             create_items_loop  > sub_workflow > steps > create_novel_item
```

A 32% undercount, with no error and no empty result to hint at it. What caught it was not
suspicion of the query — it was carrying a **positive control** into a text-level scan for
an unrelated reason:

```sql
SELECT count(*) FILTER (WHERE default_config::text ~ 'item_domain')   AS old,   -- 12 defs
       count(*) FILTER (WHERE default_config::text ~ 'item_pipeline') AS newsp  -- 2 defs
FROM agent_definitions WHERE deleted_at IS NULL AND COALESCE(is_snapshot,false)=false AND is_active;
```

12 definitions carried `item_domain`; the step-level census had found 8 agents. **Two
instruments disagreeing is what made the blindness visible**, and only one of them could
be right about a count of definitions.

The sharp part: this exact failure is already named in the tree. `validation.WalkSteps`
exists *because of it* — its doc comment says the question "which (action, key) pairs are
live?" *"used to be answered by a hand-written `->'workflow'->'steps'` query, which sees
only the top level — the offline audit and the runtime validator were therefore blind in
the same direction and agreeing with each other (bugs_open/144)"*. The audit was fixed on
2026-07-29 and prints its own coverage line saying so (`68 pairs inside loop
sub-workflows, 25 of which exist ONLY there`). **This lane's RUNBOOK kept the bug that the
platform had already fixed**, and I read the audit's coverage banner in my own baseline
output before I ran the blind query. RUNBOOK corrected in place, with the recursive and
text-scan versions and a third gotcha.

Consequence for the migration: 349's rename is driven by a **depth-recursive walk**, not by
a hand-written list, so it cannot inherit the blindness. Guard asserts 19; verify re-walks
from scratch and asserts 0 remain.

### The near-miss: renaming a seed can break a LATER migration that deletes the same key

I renamed `item_domain` → `item_pipeline` across 22 seed files by grep. Six of those lines
were in `051_build_dispatch_loop.sql`, which seeds `build-dispatch-loop` with an
`item_domain` filter on `load_next_item` and `check_remaining`. **`052_build_pipeline_trigger.sql`
deletes exactly those keys by name** (`(config) - 'item_domain'`, guarded by
`... ? 'item_domain'`) because the filter was a defect — it meant design/content items were
never dispatched. Renaming the seed would have left 052 matching nothing on a replay, and a
`item_pipeline: "build"` filter would have survived into a rebuilt definition, silently
restoring the bug 052 fixed. **Reverted those six lines; 051 is untouched.** The general
shape: a seed is not a standalone description of intent, it is one frame of a chain, and
the next frame may be keyed on the exact spelling you are "tidying".

### Adjudicating the create_work_item keys (item D), and a fourth dead key nobody had listed

`domain` on claims-auditor is the key the bug file warns not to convict by grep, because the
action really does call `inputs.Get("domain")` at `:163`. Read `ExtractActionInputs` end to
end instead: Strategies 0, 2, 4 and the nested-object pass all iterate
`spec.Required ∪ spec.Optional`; Strategy 1 iterates `config["input_fields"]`; Strategy 3
iterates `spec.Deprecated`. **A key in none of those three sets is resolved by nothing.**
`domain` is in none, and the step sets no `input_fields`. Dead. Also never exercised:
`item_key LIKE 'claims_llm%'` → 0 rows, against **4861** rows carrying an underscore
item_key as the positive control.

Worth recording because it changes the fix: the author's intent (an item_key named for the
domain) is **already reachable** — `item_key_suffix_field` is read straight from config and
resolved with `ExtractNestedFieldString`, no spec involvement at all. And the alternative
"just declare `domain` in the spec" is not a declaration but a fleet-wide behaviour change:
the nested-object pass at `:544-568` would then resolve `site_record.domain` for **every**
`create_work_item` step, re-shaping `item_key` from `<prefix>_<siteid8>` to
`<prefix>_<domain>` everywhere and breaking dedup continuity against existing rows on
`idx_swi_dedup`.

**The no-op check found a key the handoff did not know about.** Before setting
`CheckConfig: true` I enumerated every live `create_work_item` step's config keys, at all
depths, minus the proposed recognised set. Three keys came back: `domain`, `spec_fields`,
and **`spec`** — the last on 3 steps across improvement-loop and deduplicate-sections, and
read by nothing (the action builds its spec from `spec_data`/`spec_paths`/`spec_literal`).
Unlike the other two it has a live consequence: `051_build_dispatch_loop.sql:823` maps
`pending.first_item.spec.refresh_site_components` and `033_rerender_pages_action.sql:1107`
gates on `input_data.spec.refresh_site_components == true`, so improvement-loop's
`{"spec": {"refresh_site_components": true}}` never reaches either. **16 of 16
`improvement_rerender_*` rows carry `spec = {}`**, the most recent filed today; positive
control, 4972 rows fleet-wide DO carry a non-empty spec. `bugs_closed/024` established
`spec_literal`/`spec_paths` as the correct spelling (migration 180) and these three steps
never migrated.

Deliberately NOT swept into 350: two of the three have never filed a row and are safe to
translate, but `improvement-loop.insert_rerender_item` files ~2/day and turning its flag
back on would start triggering full site-component reassembly — which interacts with
`bugs_open/226` (chrome rebuild silently discards hand-patched content). That is an owner
call, not a tidy-up. Filed to the diagnosis loop instead:
**090 run `be967639-d195-444a-b9c3-ef1445ff7ae1`**. The opt-in therefore ships knowing it
will REPORT one key on three steps, and that report is the detector working — the
alternative, declaring `spec` in `ConfigKeys` to keep the number at zero, is precisely the
`bugs_closed/101` failure this bug's §5.4 forbids.

### My own false claim, caught by running the mutation instead of asserting it

I wrote on the new behaviour test: *"MUTATION PROOF: set CheckConfig back to false and
`checked` goes false here."* I then ran that mutation and **it passed**. `checksConfig()` is
`s.CheckConfig || len(s.ConfigKeys) > 0` — a guard in **series**, so a non-empty
`ConfigKeys` already carries the opt-in and `CheckConfig` is redundant today. The claim was
false the moment I typed it, and nothing but running it would have said so. Corrected in
place, and a fourth test now isolates the second signal by registering a copy of the spec
with `ConfigKeys` emptied under a probe name (with the both-signals-off control, so it
cannot pass against a `checksConfig()` that returned true unconditionally). Logged in
`WRONG_CALLS.md`.

### Item E: skipped, and not for the reason the handoff gave

The handoff calls it "a small code change with no behavioural exposure" that could ride D's
council round. Zero live carriers re-verified today (`group_type` 0, `group_type_field` 0;
positive control, `agent_type_field` 5). But it is **not** a one-liner: `spawn_agent` has no
`ActionInputSpec` at all, so converging means creating one; `agent_type` is a **framework**
key (`frameworkStepConfigKeys`), so declaring it in an action spec would misstate ownership;
and `group_type` (literal) and `group_type_field` (path-valued) would land in *different*
alias fields — `DeprecatedConfigKeys` and `Deprecated` respectively — which is the very
distinction this bug's landmine is about. Recorded as still-open with that correction.

### CORRECTION, same session — I banked a clean audit and then quoted it after changing the code

I recorded `UNKNOWN KEYS: none / DEPRECATED KEYS: none / exit 0` as the lane's headline, and
wrote in three documents that the `spec` key "will be reported once the image rolls". **The
timing is wrong and the headline is stale.** `audit-config-keys.sh` runs
`go run ./cmd/config-key-audit`: it reads the **source at HEAD** and joins it against **live
DB config**. Committing item D therefore changed the report immediately — no roll:

```
UNKNOWN KEYS: create_work_item: spec      DEPRECATED KEYS: none      exit 1
```

Only the RUNTIME validator's warnings wait for the image. I conflated the offline audit with
the deployed binary, which is the same category error as reading a deploy from git — the one
this estate warns about constantly, committed here in the direction nobody watches for
(claiming a *better* state than is true, from a run that really did happen).

**What caught it:** re-running the audit one final time after committing, rather than citing
the run banked ninety minutes earlier. **A banked pass is evidence about the moment it was
taken.** The three figures in the middle row of §12's table are the honest lane-attributable
result; the bottom row is today's truth.

Practical consequence worth flagging to anyone wiring this up: **the script exits 1 today**,
correctly (`1 = unknown keys found`), and will until `bugs_open/234` is decided.
