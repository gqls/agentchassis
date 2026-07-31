# HANDOFF — staged_component_build, 2026-07-31 (afternoon)

**Written because context ran long, at the owner's suggestion.** This is the cold-start
doc for a fresh thread on this lane. Everything below is verified as of 2026-07-31 ~09:00
UTC unless marked. **Re-run every figure** — three of them moved inside 24 hours while I
was working (fence count 23→25, landmine corpus 57→190, tool components 28→29).

> ## ⚠ CORRECTIONS, 2026-07-31 afternoon — read these BEFORE §4 and §5
>
> The next thread picked this up and two items below are now wrong. Both are corrected in
> place in `NOTES` (entries of 2026-07-31 afternoon) and in `features_open/027`.
>
> 1. **§4 item 1 is DONE.** `tool-review-council-simulator` has an **18-check fence live in
>    `doc_plans`** (row `ec711f24`), and the cluster acceptance run is **GREEN**:
>    correlation `cf6b6e34`, `complete` in **18s, 22 passed / 0 failed / 14 intentional
>    profile-gated skips**. `CHECK_naming_contract.sh` now reports **BROKEN B: 0**.
>    The instruments that made it possible are new and reusable — `scripts/try_fence.go` and
>    `scripts/prove_fence_can_fail.go` (register **TL-036**), plus RUNBOOK §8–§10.
>    **The first dispatch FAILED on the 120-second `runDeadline`** with an error that names
>    the browser and reads as infrastructure; size a fence for the pod, ~3-5s per evaluation
>    there against ~0.3s locally. **An offline harness proves a fence CORRECT, never that it
>    FITS.**
>
> 2. **§4 item 1's second finding and §5's "orphan" claim are REFUTED.**
>    `tool-arena-interface` is **NOT an orphan**. It is **live, deployed and serving** on
>    vonc.com under a page named `tool-arena` (`/tools/arena/index.html`,
>    `build_status=deployed`), and its markup is present in the served page. My own check
>    caused the error: it concluded *"no page at all"* from *"no page under the two names I
>    guessed"*, and its URL guess assumed a `<name>.html` convention vonc.com does not use.
>    **So the question is NOT "should this component exist" — it plainly should — but "which
>    of the two names should move".** The check now asks placement via `page_components`
>    before concluding absence, and prints the remedy. The rename itself is **still not
>    done and still needs a decision**: it is another site's live page, and the 07-31
>    precedent requires measuring blast radius first.
>
> §3 (the **component** Go gate, `subject_type='component'`) is **UNCHANGED and still the
> open item** — `doc_plans` holds 0 component rows and `docSubjectGateReason` has never been
> observed accepting the type. Everything in §3 still applies as written.

---

## 1. Read these, in this order, and do not re-derive them

| doc | why |
|---|---|
| `SUMMARY_2026-07-31_we_cut_the_ladder_down.md` | **start here.** Plain prose: what the lane is, and why the eight-stage ladder was cut to three |
| `PLAN_2026-07-30_staged_component_build.md` | decisions **D1–D8** and the phasing. **D8 is the scope cut. D3 is the correctness requirement.** |
| `RUNBOOK_staged_component_build.md` | the commands, each with the gotcha attached |
| `NOTES_staged_component_build.md` | append-only log; read the **last three entries** |
| `REPLY_2026-07-30_vendor_trust_checklist_build.md` | the other lane's forward run + my answer. **The most useful single document in the directory** |
| `PROPOSAL_2026-07-30_…` | the original argument, under a *superseded-in-scope* banner. Read for history, **do not build from it** |

`features_open/027` is the anchor; `features_open/028` is the rename-orphan deferral.
Concept register entry is **DOC-068**.

---

## 2. State — what is live, what is not

**LIVE and proven:**
- **`subject_type='component'` on the DB half.** Migration **273 APPLIED** 2026-07-31 by
  hand (`psql -f` + `--record-only`), council **APPROVED** (trail
  `e5673868-7c5b-489c-931a-7ba59b959b91`, r1 REVISE → r2 approved, 11 approve / 3 advisory).
  Both CHECKs allow `component`; `doc_notes` kept `landmine`.
- **The Go half is in the running binary by build date** — pods `5c847465c4-*` built
  **2026-07-31 08:49:09 UTC**, after the Go commit `c659e312b` (07-30 19:28 UTC).
- **`CHECK_naming_contract.sh`** — runs, and correctly FAILS with 2 findings.
- **`VERIFY_273_before_apply.sh`** — runs, and its probe has been watched in **both**
  states (refused pre-migration, wrote-and-read-back post-migration).
- **`tool-review-council-simulator`** page renamed (fundamentallyai, page `e4f422e7`);
  live page byte-identical before/after (200 / 60,021 bytes).

**NOT verified, and it is the single next action:**
- **The GO gate has never been exercised with `subject_type='component'`.** Build date says
  the binary *can* carry the vocabulary; nothing proves it *does*. `doc_plans` currently
  holds **0** component rows, so nothing has ever gone through `docResolveSubject` for this
  type.

---

## 3. THE NEXT ACTION, with everything needed to do it

**Goal:** prove `docSubjectGateReason` no longer returns `unsupported subject_type` for
`component`, i.e. that the Go half genuinely shipped.

**Why it needs a dispatch and not a query:** `load_doc_context` takes `subject_type` from
**step config**, not input data (`load_doc_context_action.go:37-43` — `subject_type` is
Optional in the input spec, resolved by `docResolveSubject(config, …)`). So `psql` cannot
reach the gate; the probe in `VERIFY_273` deliberately says so.

**Measured 2026-07-31: exactly ONE active agent has a `load_doc_context` step**
(`tool-acceptance-agent`), and it is configured for `subject_type='tool'`. So you need a
scratch route.

**Recommended, smallest, reversible:** seed a scratch agent with one `load_doc_context`
step whose config sets `subject_type: component`, plus one component PLAN to read.

1. Write a component PLAN so there is something to load (this also re-proves the DB half):
   ```sql
   INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
   VALUES ('component','teaser-reveal-panel','# PLAN — probe','handoff-goproof','staged_component_build');
   ```
2. Seed the scratch agent (config-only, live immediately, **no image needed**). Follow the
   shape of `tool-acceptance-agent`'s `load_docs` step. **`snapshot_agent` first if you
   touch an existing row — do not.**
3. Dispatch it, then read the step output. **PASS** = the PLAN body comes back.
   **FAIL** = `unsupported subject_type "component"`, which would mean the build does *not*
   carry the Go half despite its date — the exact "necessary not sufficient" case the gate
   warns about.
4. **Clean up:** delete the scratch agent row and
   `DELETE FROM doc_plans WHERE source='handoff-goproof';`

**Then, and only then**, it is honest to say component travelling docs work.

---

## 4. Open items, ordered, with owners

1. **P1a — the naming contract**, and it is not finished. `CHECK_naming_contract.sh` FAILS
   with two findings:
   - **`tool-review-council-simulator`** (ours) — has a PLAN, **no ```criteria fence**, so a
     run SKIPS and reads clean. **Remedy: author the fence.** Never invent a selector;
     watch every criterion pass by hand before writing it. This is the concrete next build.
   - **`tool-arena-interface`** — **no page under either name.** An orphaned component, a
     different defect. **Do NOT rename it.** Someone must decide whether it should exist.
2. **10 tools resolve to a page but have no PLAN at all** — the authoring backlog. Honest
   (nothing claims they were tested), so not a check failure. Tools born via
   `create_tool_component_action` are born compliant (`tool-relevant-alternative`, 07-31), so
   **this check is a backlog cleaner, not a permanent gate** — do not wire it into the birth
   path, where it would assert what the code already guarantees.
3. **S6 for components** — dispatch a component's fence to `browser-runner-adapter` as
   `tool-acceptance-agent` does for tools. Wiring, not construction. **Blocked in practice
   by `bugs_open/157`** (below).
4. **`features_open/028`** (rename orphaning) — filed, unowned. Candidate 2, a detector for
   travelling docs whose `subject_key` resolves to nothing, is the only one that finds
   orphans from *past* renames — a count nobody has measured, across all six subject types.

---

## 5. Do NOT do these

- **Do not rebuild the eight-stage ladder.** Owner cut it (D8). Two of the eight gates were
  wrong on first contact and S4 *would have blocked a correct build*.
- **Do not take `bugs_open/157`** (`has_visible_area` reports 0 for whole-number axes). It
  belongs to the **leopardess lane** — they filed it, hold the reproducer, and have the root
  cause at `playwright-go@v0.6100.0/js_handle.go:109-114`. I told them so in writing in the
  REPLY doc. Two threads on two lines is the waste.
- **Do not run `./scripts/migration/run-migrations.sh --apply`.** It takes **every** pending
  file. As of 07-31 those include `275_oufe_tool_relevant_alternative.sql` (**syntax error**),
  `274_…` (**contains its own ROLLBACK**), and 266/269 (**probe inconclusive**). Apply single
  files by hand + `--record-only`, per migration 270's precedent.
- **Do not renumber `273_fix_proposer_plan_repair_loop.sql`.** A duplicate 273 exists
  (another session's). Checked: it does not touch `doc_plans_subject_type_check`, the
  lockstep test still resolves to mine and passes, and it is not ours to renumber.
- **Do not roll the chassis to ship something.** Builds come from committed HEAD, so your
  commit ships on anyone's next roll — mine did, twice, without my intervention. A
  deliberate roll kills an in-flight council and imposes a ~300s dispatch blackout.
- **Do not adopt `features_open/015`.** Accepted decomposition: **015 = rung vocabulary,
  027 = gate mechanism, 026 = missing instrument.** Composable, not merged.

---

## 6. Landmines this lane paid for — the ones that cost real time

- **The migration is only half of a `subject_type` change.** `validDocSubjectTypes`
  (`doc_subjects_common.go`) is a second enforcement point. DDL alone **is** `bugs_open/064`,
  which exists because migration 184 did exactly that. **Grep the VALUE you are adding, not
  the table you are changing** — `git grep "experience-pattern"` returns the Go list, the
  migration and the four-point checklist in one command. `\d <table>` is silent about every
  gate in front of the database.
- **`doc_notes` and `doc_plans` do NOT share a vocabulary.** `doc_notes` also allows
  `landmine`. Rebuilding its CHECK from `doc_plans`' array orphans the landmine corpus
  (**190 rows** on 07-31, and it was 57 the day before).
- **A negative from a short grep marker is worthless.** Go compiles short string literals to
  immediate comparisons that never reach rodata: `grep -ac "selector_count"` returns **0** on
  a binary that fully supports it. Comments never reach a binary at all — I used two as
  markers and both read 0. Bare `component` matches **761** times. **When a change adds no
  unique string, date the build** (`stat -c %Y` vs the commit's `%ct`) and say out loud that
  it is necessary, not sufficient. These images have **no `strings`** — `grep -ac` the binary.
- **Backticks inside a double-quoted bash string are command substitution.** Writing
  ` '%```criteria%' ` inline stopped a script parsing. Put the pattern in a single-quoted
  variable.
- **`kubectl exec -i` inside a `while read` loop eats the loop's stdin** and the loop ends
  after one row — silently, **under-reporting**. Read into an array, or `</dev/null` the call.
- **`grep -c` exits 1 on zero matches**, so a `|| echo 0` fallback yields the two-line string
  `"0\n0"` and a string comparison against `"0"` fails. That produced a **false pass** in my
  own gate.

---

## 7. The one thing to carry into any new work here

**Nine instances in two days of a single class: a check that reports health it never
measured** — and the last three were inside the detectors I built to catch the class. The
worst was a column I named `has_fence` that only tested for a PLAN *row*, which promoted the
tool I had just renamed to "testable now" when its run would skip.

So the rule that survived the scope cut, and the only one I would defend without evidence:

> **Watch every branch of a check fail before quoting anything it says — and a check must not
> be able to say `ok` about something it did not measure.**

Printing the number is not enough if the same code path can print `?`.
