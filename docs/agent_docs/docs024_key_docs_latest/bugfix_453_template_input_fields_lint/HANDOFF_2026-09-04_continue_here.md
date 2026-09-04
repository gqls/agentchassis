# HANDOFF — 2026-09-04 — `bugs_open/453`: one half LIVE, one half APPROVED-but-unscheduled, and a council REVISE outstanding

**Read this first, then `NOTES_template_input_fields_lint.md` (newest at the bottom).**
Supersedes nothing — this is the lane's first handoff.

**Bug:** `bugs_open/453_HANDOFF_2026-09-03_a_template_variable_input_fields_never_names_renders_empty_with_no_error_and_nothing_lints_the_pair.md`
**Lane dir:** `docs/agent_docs/docs024_key_docs_latest/bugfix_453_template_input_fields_lint/`
**Register:** **WFA-024** (the config lint) · **PRC-003** (the render fix)
**Councils:** `54abc24b` **APPROVED** (WFA-024) · `8384acb0` **REVISE** (PRC-003, gating HIGH — *not actioned*)
**Diagnosis:** `92309b45` **UNVERIFIABLE** (iteration-cap; not refuted)

---

## 1. What this lane is

`bugs_open/453` says a prompt template can name a variable nothing supplies, and it renders as
**nothing** — no error, no verdict. Four prior catches, all by fixtures somebody happened to
write. The seam has **three shapes** and this lane built a fix for two of them.

| shape | what it is | state |
|---|---|---|
| 1 | no `input_fields` at all → randomised recursive resolution | **open**, reported as context only |
| 2 | root missing from `input_fields` → resolves on no row, ever | **WFA-024 built + APPROVED**, ⚠ not scheduled |
| 3 | root present, sub-field absent in the DATA → `<no value>` | **PRC-003 LIVE + PROVEN**, ⚠ council REVISE outstanding |

---

## 2. WHAT IS LIVE, and how it was proven (not inferred)

**PRC-003 is live on `v1.0.1360`, carried by `681b0ee65`.** Two independent proofs:

**(a) Binary probe** — four controls through ONE `tr '\0' '\n' | grep -Fc` pipeline over
`/proc/1/exe` (BusyBox `grep` reports false absences on that file; this is the LANDMINE'd form):

| probe | result | meaning |
|---|---|---|
| `PROMPT RENDERED A STAND-IN INSIDE AN AUTHORITATIVE BLOCK` | 1 | new code present |
| `placeholders stripped before send` | 1 | new code present |
| `TEMPLATE RENDERED WITH MISSING DATA` | **0** | **old code gone** — the discriminating control |
| `MASTER EXTRACTOR START` | 1 | must-be-present control ✓ |
| `ZZZ_THIS_STRING_MUST_NOT_EXIST_ZZZ` | 0 | must-be-absent control ✓ |

**(b) Behaviour at the durable record** — prompts carrying the literal token:
**50 of 153 (32.7%) before the roll → 0 of 447 after.** The six survivors between 22:09 and
22:19:13Z were old pods finishing in-flight work.

⚠ **The pod-log route could NOT prove it, and reads exactly like failure.** A 6-hour window
returned **zero** `MASTER EXTRACTOR START` — no demand — while `llm_call_log` showed **526**
calls in the same period. `kubectl logs -l app=` reads one pod of N. **Always take the demand
control before believing a clean grep here.**

---

## 3. ⚠ THE MOST IMPORTANT THING ON THIS PAGE — the fix removed the durable evidence trail

`llm_call_log.prompt_rendered` stores `renderedPrompt`, which is what `RenderPromptTemplate`
**returns** — i.e. **post-strip**. So:

- the token no longer appears in the durable record;
- **the 437 lane's 65% measurement is no longer reproducible by anyone**;
- whether a hole occurred is now visible **only in ephemeral pod logs**.

That is precisely the council's gating objection, and the roll has made it empirical:

> `[MEASURED 2026-09-04]` of **448** post-roll `page-content-writer` calls, **267** carry an
> empty `Location:` line **inside the DO-NOT-INVENT block** — roughly 267 Error-level events
> since the roll, with **no durable record of a single one**.

*(Upper bound: an empty line is also what a present-but-empty string renders, so the signature
cannot separate the two. Before the roll the token itself was the exact measure. That is
what was lost.)*

**This is the next session's first job.** See §5.

---

## 4. What is built and where

| thing | path | state |
|---|---|---|
| Config lint | `cmd/config-key-audit/templateinputfields.go` + `scripts/audit-template-input-fields.sh` | built, **council APPROVED** |
| Its shared contracts | `platform/orchestration/{datahelpers,actions}/template_context_contract.go` | live (came with the roll) |
| CronJob for the lint | `deployments/kustomize/services/template-input-field-check/`, `build/docker/backend/template-input-field-check.dockerfile`, makefile targets | committed, **NOT deployed** |
| Render fix | `platform/orchestration/datahelpers/template_missing_values.go` + the block in `RenderPromptTemplate` | **LIVE v1.0.1360** |
| Live re-measurement probe | `platform/orchestration/datahelpers/live_sample_probe_test.go` | skipped unless `NOVAL_SAMPLE=<dump>` |

**Commits:** `4aaf64aee` (lint) · `681b0ee65` (render fix) · `6c60c3bc2`, `71c7ce40b`,
`3bd383310`, `ed3dee48a` (docs, corrections, council actions) · the CronJob commit · plus
register/landmine updates.

---

## 5. WHAT IS LEFT — ordered by what closes the door

### 5.1 BLOCKING for `bugs_open/453` closure

**(a) Action the council's HIGH objection: give the escalation a DURABLE surface.** `8384acb0`
returned **REVISE** on exactly this, and §3 shows the roll made it worse rather than merely
unaddressed. The cheapest honest route — **the durable record already exists, what is missing
is a reader**: a scheduled sweep over `llm_call_log` for the post-strip signature, reporting to
`doc_notes`, on the CronJob pattern this lane already built. Alternatives considered and
rejected in NOTES: an `agent_error_log` row per call (~1,453/day) and a work item filed from
inside a hot render path (no DB handle at that layer). **Then resubmit with
`RESUBMIT_CORR=8384acb0-bd99-45cd-9fbb-bd7e4add59f4`.**

**(b) Deploy the WFA-024 CronJob.** Manifests, dockerfile, makefile targets and release-list
entries are committed and `make check-release-coverage` passes. It is **read-only by
construction** — the owner asked, and it was verified at the source: the only two statements
reachable are the fleet-export `SELECT` and the `doc_notes` `INSERT`. Needs a fleet apply,
which is the owner's. Until then the lint is hand-run and this estate's record is that
unscheduled detection decays.

### 5.2 OWNER DECISIONS still outstanding (from the 2026-09-03 exchange)

- **The `research_result` block on `page-content-writer`** — wire the research up or delete the
  dead prompt block. WFA-024's only conviction. Live damage zero (the gate `needs_research` is
  carried by **0 of 554** `content_components`), but it is one word from costing real money.
- **The `Location:` template gate.** The right fix for the 267 is one line —
  `{{if .reviewed_brief.headquarters}}Location: …{{end}}` — so an absent field prints nothing
  instead of a bare label. A migration on the fleet's highest-volume writer; wants its own round.
- **The other ten template seams** (§5.3) — file as their own bug, or widen this lane?

### 5.3 KNOWN-OPEN, recorded not fixed

- **The class is wider than the two seams patched.** `[MEASURED 2026-09-03]` **twelve** Go
  template executions outside tests/vendor/SQL; **two** guarded. The ten include
  `git_deployer_actions.go:657` — the LANDMINE'd `git_commit` case where any field outside
  `{domain, file_count, filename}` renders the token and the commit succeeds anyway.
- **Shape 1** — convicting it needs a rule about when recursive search is acceptable. Nobody has
  written one.
- **Council LOW** — `ScanMissingValues` and `missingBareFields` are two detectors of one
  judgement with nothing holding them in lockstep. A cross-reference test pinning the divergence
  *with its reason* (the `render_seam_absent_required_test.go` shape) is owed.
- **19 `declared_unread`** advisories. Three spot-checked, all genuine waste, no damage.

---

## 6. Traps this lane paid for — read before touching these

- **`orchestration_states` retains TWO DAYS.** I read a `count(*)` with no date predicate as
  "all time" and told the owner an agent had *"never run, not once, ever"*. It last ran
  **2026-01-18**. Ask a table its own span before reading a zero as history. (`WRONG_CALLS`)
- **A measurement answers the question you ENCODED.** My first estimate of the new Error's rate
  said **0 of 300**; it tested only the block before the FIRST occurrence, and the dangerous one
  is later in the same prompt. Running the real code over live prompts: **74 of 80 (92%)**.
- **Back up mutation targets BY PATH, not basename.** `template_context_contract.go` exists in
  two packages; a basename backup restored one package's file into the other, and it surfaced as
  a build failure I first read as a mutation result.
- **A dotted `input_fields` entry supplies its LAST segment** — `{{.a.b}}` is dead while
  `{{.b}}` works. Two LANDMINES entries from this lane.
- **Injected template roots are PER ACTION, never a union** — the first sizing pass convicted
  both live vision steps (2 false positives in 12).
- **Submit to the council BEFORE committing**, so the correlation exists to name in the trailer.
  `4aaf64aee` is the commit the council actually reviewed and it carries no trailer, for ever.

---

## 7. How to pick this up

```bash
# the lint, against the live fleet (read-only)
scripts/audit-template-input-fields.sh

# re-measure the Error rate after any template gate lands
#   (dump prompt_rendered rows to JSON first — see RUNBOOK)
NOVAL_SAMPLE=<dump.json> go test ./platform/orchestration/datahelpers/ -run TestLiveSampleEscalationRate -v

# the council round to resume
RESUBMIT_CORR=8384acb0-bd99-45cd-9fbb-bd7e4add59f4 \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh <submission.json>
```

⚠ `go test ./cmd/config-key-audit/` fails `TestShippedRegistryIsSelfConsistent` at HEAD for
**`bugfix_450`'s** registry entry, not this lane's — proven on a clean `cc572ea14` extract.
Scope your `-run`.
