# HANDOFF — staged_component_build, 2026-07-31b (evening)

**This supersedes `HANDOFF_2026-07-31_continue_here.md`** for everything except its §3 (the
component Go gate), which is reproduced below unchanged because it is still the open item.
That file's §4 item 1 and §5 "orphan" claim are DONE and REFUTED respectively — it carries a
corrections banner, but read this file instead.

**Every figure below was measured on 2026-07-31 between 12:30 and 15:40 UTC. Re-run all of
them.** Four moved *while I was working*: fence count 23→25, landmine corpus 57→190, tool
components 28→29→**30** (a new tool appeared ten minutes before one of my runs), and the arena
page's served size 31,431→32,553 bytes. This is not incidental — one of them nearly produced a
false claim, and the byte one is now a recorded landmine.

---

## 1. Read these, in this order

| doc | why |
|---|---|
| `README_where_we_are.md`, last three entries | plain prose, the owner's log — fastest way in |
| `PLAN_2026-07-30_staged_component_build.md` | **D8** is the scope cut (8 gates → 3), **D3** the correctness requirement, **D5′** the two-enforcement-point lesson |
| `NOTES_staged_component_build.md`, last three entries | the technical log, including every misstep — the missteps are the point |
| `RUNBOOK_staged_component_build.md` §8–§11 | the four new procedures, each with its gotcha attached |
| `SUMMARY_2026-07-31_we_cut_the_ladder_down.md` | why the ladder is three gates and not eight |
| `REPLY_2026-07-30_vendor_trust_checklist_build.md` | the other lane's forward run — still the most useful single document here |

`features_open/027` is the anchor. Register: **DOC-068** (`subject_type='component'`),
**TL-036** (the two fence instruments). `PROPOSAL_2026-07-30_…` is history under a
superseded-in-scope banner — read it, do not build from it.

---

## 2. State — what is live and proven, what is not

**DONE and proven at the artefact:**

- **P1a, the three-way naming contract, is CLOSED.** `CHECK_naming_contract.sh` returns
  **PASS** — BROKEN A **0**, BROKEN B **0** — across 30 canonical tool components
  (12 testable now / 10 authoring backlog / 8 neither; 12+10+8=30 reconciled). First pass
  since it was written.
- **`tool-review-council-simulator` has an 18-check fence live in `doc_plans`** (row
  `ec711f24`, body 20,202 bytes) and the **cluster acceptance run is GREEN**: correlation
  `cf6b6e34-3c28-41db-8adf-ee7550bc4224`, `complete` in **18s, 22 passed / 0 failed / 14
  skipped**, every skip verified `not run on profile mobile` and none `not implemented`.
- **Two reusable instruments**, `scripts/try_fence.go` and `scripts/prove_fence_can_fail.go`
  (register TL-036). 17 mutants / 17 caught / **18 of 18 checks watched red** against an
  all-green baseline.
- **`subject_type='component'` DB half is LIVE** — migration 273 applied by hand 07-31, both
  CHECKs allow `component`, `doc_notes` kept `landmine`.
- **vonc.com's arena page renamed** so its tool is addressable — `pages.name` **and**
  `site_plan_pages.name`, one transaction, scoped by ID. Served page byte-identical, md5
  `4a2d2030e2f6d2630f6497f68705a067` both sides.

**NOT verified, and it is the single next action — unchanged from the previous handoff:**

- **The Go gate has never been exercised with `subject_type='component'`.** The binary carries
  the vocabulary *by build date only*; `doc_plans` holds **0** component rows, so nothing has
  ever passed through `docResolveSubject` for this type and `docSubjectGateReason` has never
  been observed accepting it. **Necessary, not sufficient.**

---

## 3. THE NEXT ACTION, with everything needed to do it

*(Carried forward verbatim in substance from the previous handoff — still correct, still open.)*

**Goal:** prove `docSubjectGateReason` no longer returns `unsupported subject_type` for
`component` — i.e. that the Go half genuinely shipped, rather than merely being in a binary
built after the commit.

**Why it needs a dispatch and not a query:** `load_doc_context` takes `subject_type` from
**step config**, not input data (`load_doc_context_action.go:37-43`, resolved by
`docResolveSubject(config, …)`). So `psql` cannot reach the gate — writing a component PLAN
through `psql` proves the DB CHECK and nothing about Go.

**Measured 07-31: exactly ONE active agent has a `load_doc_context` step**
(`tool-acceptance-agent`) and it is configured for `subject_type='tool'`. So you need a
scratch route.

**Recommended, smallest, reversible:**

1. Write a component PLAN so there is something to load (this also re-proves the DB half):
   ```sql
   INSERT INTO doc_plans (subject_type, subject_key, body, source, created_by)
   VALUES ('component','teaser-reveal-panel','# PLAN — probe','handoff-goproof','staged_component_build');
   ```
2. Seed a scratch agent with one `load_doc_context` step whose config sets
   `subject_type: component`. Config-only, live immediately, **no image needed**. Follow the
   shape of `tool-acceptance-agent`'s `load_docs` step. **`snapshot_agent` first if you touch
   an existing row — do not touch one.**
3. Dispatch it, then read the step output. **PASS** = the PLAN body comes back.
   **FAIL** = `unsupported subject_type "component"`, which would mean the build does not
   carry the Go half despite its date.
4. **Clean up:** delete the scratch agent row and
   `DELETE FROM doc_plans WHERE source='handoff-goproof';`

**Then, and only then**, it is honest to say component travelling docs work.

---

## 4. Open items, ordered

1. **The Go gate proof** (§3). Nothing else in this lane is blocked on it, but DOC-068's status
   stays "DB half live, Go gate unexercised" until it is done, and that wording is deliberate.
2. **S6 for components.** Dispatch a component's fence to `browser-runner-adapter` the way
   `tool-acceptance-agent` does for tools. **Wiring, not construction** — and now cheaper than
   it was this morning, because `try_fence.go` lets you author and prove a component fence
   before any dispatch path exists. Gate: a deliberately broken component makes it go red.
3. **The authoring backlog: 10 tools resolve to a page but have NO PLAN at all.** Honest, not
   a check failure. **Do NOT wire the naming check into the tool-birth path** — tools born via
   `create_tool_component_action` are born compliant (three in a row now:
   `tool-relevant-alternative`, `tool-gripper-safety-factor-calculator`), so the check is a
   backlog cleaner and would be asserting there what the code already guarantees.
4. **`features_open/028`** (rename orphaning) — filed, unowned. Its candidate 2, a detector for
   travelling docs whose `subject_key` resolves to nothing, is the only one that finds orphans
   from *past* renames — a count nobody has measured, across all six subject types. **Note the
   arena rename did NOT touch a `subject_key`**, so it created no new orphan.
5. **`has_visible_area` checks are owed to every fence once `bugs_open/157` closes.** The type
   is live in the running binary and still wrong (integer axes read 0). The owner has taken 157
   in a separate thread — **do not duplicate it.**

---

## 5. Do NOT do these

- **Do not rebuild the eight-stage ladder.** Owner cut it (D8). Two of the eight gates were
  wrong on first contact and S4 *would have blocked a correct build*.
- **Do not take `bugs_open/157`.** The owner is doing it in another thread. It also belongs to
  the leopardess lane, who hold the reproducer and the root cause
  (`playwright-go@v0.6100.0/js_handle.go:109-114`).
- **Do not fire an acceptance run at the arena tool.** It is the `gauntlet_dead_cta` lane's
  decision now (CONTRIB filed in their directory + pointer in their cold-start §4). A failing
  verdict files an `improve_tool` item routed to an automated `tool-improver` against an
  `owned` page — the wrong way to settle whether the fence is stale or the tool is incomplete.
- **Do not run `./scripts/migration/run-migrations.sh --apply`.** It takes **every** pending
  file; as of 07-31 those include `275_oufe_tool_relevant_alternative.sql` (**syntax error**),
  `274_…` (**contains its own ROLLBACK**) and 266/269 (**probe inconclusive**). Apply single
  files by hand + `--record-only`, per 270/273's precedent.
- **Do not renumber the duplicate `273_fix_proposer_plan_repair_loop.sql`.** Checked: it does
  not touch `doc_plans_subject_type_check`, the lockstep test still resolves to mine and
  passes, and it is not ours to renumber.
- **Do not roll the chassis to ship anything.** Builds come from committed HEAD, so your commit
  ships on anyone's next roll — mine did, three times, without my intervention. A deliberate
  roll kills an in-flight council (there were two running at 14:51Z today) and imposes a ~300s
  dispatch blackout.
- **Do not adopt `features_open/015`.** Accepted decomposition: **015 = rung vocabulary,
  027 = gate mechanism, 026 = missing instrument.** Composable, not merged.

---

## 6. Landmines this lane paid for — all now in `LANDMINES.md`, footprinted

The ones that cost real time today, shortest useful form:

- **A fence can be correct, fast locally, and still FAIL in the cluster on the 120s
  `runDeadline`** — and the error names the *browser* (`browser open failed … context deadline
  exceeded`), so it reads as infra. 36 evaluations = 10.6s locally (×3) but FAILED at 133s
  in-cluster; ~3-5s per evaluation there. Gate to desktop every check whose answer is
  profile-independent. **An offline harness proves CORRECTNESS, never FITNESS.**
- **Renaming `pages.name` silently removes the page from `check_sectionless_pages`**, which
  joins `site_plan_pages.name = pages.name`. Move both name-side rows, then **re-run the
  detector's own join**. Generally: grep the COLUMN as a **join key**, not just the table.
- **`selector_count` does NOT assert a count** — same case arm as `selector_exists`, passes on
  `n>0`, no expected-count field, while printing a number that reads like an assertion. Assert
  counts through text the tool itself renders.
- **Prose naming the criteria fence in backticks HIJACKS extraction** — both extractors take
  the FIRST such marker and read to the next triple-backtick. Assert it appears exactly once.
- **`has_visible_area` is now IN the running pod, which makes it more dangerous, not less** —
  present ≠ correct while 157 is open. Grep `/bugs_open/` for a check type as well as the pod.
- **A served-page byte baseline goes stale in minutes** — take it in the same minute as the
  change, keep the md5, and diff before attributing a difference to yourself.
- **"No page found under the name I expected" is not "no page"**, and grepping the served HTML
  for a component's `function` returns 0 for any component that emits no `data-component`. Ask
  placement via `page_components`.
- **A negative from a short grep marker is worthless** (Go compiles short literals to immediate
  comparisons); when a change adds no unique string, date the build and say out loud that it is
  necessary, not sufficient. These images have no `strings` — `grep -ac` the binary.
- **`kubectl exec -i` inside a `while read` loop eats the loop's stdin** and the loop ends after
  one row, silently under-reporting. Read into an array, or `</dev/null` the call.

---

## 7. The one thing to carry into any new work here

**Twelve instances in three days of a single class: a check that reports health it never
measured** — and five of them were inside detectors built to catch the class. Today added two
genuinely new shapes to it, and the distinction is worth keeping:

- a check whose **logic or name** is narrower than its claim (`has_fence` tested for a PLAN
  *row*; `threshold-lever-updates-live` could not isolate the event it named) — findable by
  reading the code;
- a check whose **environment** differs from production (`try_fence.go` passing on a machine an
  order of magnitude faster than the pod, with no model of the deadline) — **not** findable by
  reading anything;
- a check whose **conclusion is wider than its measurement** ("no page under the two names I
  guessed" reported as "no page at all") — findable only by asking what it actually queried.

So the rule that survived the scope cut, in its strongest current form:

> **Watch every branch of a check fail before quoting anything it says; make sure it cannot say
> `ok` about something it did not measure; and run it once where it will really run.**

Printing the number is not enough if the same code path can print `?`. Passing on your laptop
is not enough if the pod has a deadline. And a conclusion is not a measurement.
