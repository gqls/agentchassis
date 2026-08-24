# HANDOFF 2026-08-22 — `bugfix_342_absent_required`, continue here

> # ✅✅ LANE CLOSED. Bug closed 2026-08-23; lane wound up 2026-08-24.
> **Re-verified on `v1.0.1332` on 2026-08-24** (the fleet rolled twice since the fix landed — a
> "live on v1.0.NNNN" line expires, so this one is dated): all three mechanisms present on both
> replicas with a negative control, both switches still armed. Read-out:
> `SUMMARY_2026-08-24_lane_closed.md`. **Nothing here needs picking up.**
> `bugs_open/342` → `bugs_closed/342`.
> Fixed, live on **v1.0.1330**, and proven at the artefact on every mechanism it owns. The whole
> file below is now HISTORY — read it for the traps in §6 and the reasoning, not for state.
> **Successor: `bugs_closed/367` — ALSO CLOSED 2026-08-24** (migrations 574+576; the router now
> closes only on POSITIVE evidence of absence, and the population this lane's producer reaches
> PARKS with its facts instead. Not repaired — visible and honest. Lane:
> `docs024_key_docs_latest/bugfix_367_router_remit/`.) ~~(the router only sees deployed rows, so it closes as `stale` the
> population this producer uniquely reaches — filed with the evidence, and it corrects a claim
> 342 used to make).~~ Also still open elsewhere: `bugs_open/344` (a refused edit's driving item may
> read `complete` — UNEXERCISED, not verified benign), and the chrome refusal built and
> deliberately unarmed with its trigger named. Nothing here needs picking up.

> # ⏸ SUPERSEDED — STATE AS OF 2026-08-23 ~14:00Z — THE LANE IS BLOCKED ON ONE THING: A ROLL.
> Everything this lane owns is written, council-APPROVED and committed. **Nothing here needs a
> decision or more work.** The routability fix (`eb918bd58` 13:08, `23d2a577d` 13:34, trail
> `a0ef0b07`) is **NOT live**: the deployed image is **v1.0.1328, started 11:51Z**, i.e. it
> predates both commits, and the makefile's `v1.0.1329` is a bumped tag nothing is running.
> Verified at the artefact with a working-probe control — the new literal
> `capability_gap:required_fields_missing` is ABSENT while yesterday's live
> `refuse_absent_required_fields` is PRESENT in the same probe.
>
> **To finish this lane after the next `make release` (owner-run, whole-fleet):**
> 1. Re-probe the binary for `capability_gap:required_fields_missing` — and keep a second arm in
>    the same breath (a literal you know is live), because otherwise ABSENT cannot be told apart
>    from a broken probe. Check literal NOVELTY first: `git grep -c "<literal>" <commit>^`.
> 2. Confirm the armings survived the roll: editor refusal `true`, chrome record 7/7 (§8 queries).
>    A roll re-applies overlays, so "it was armed" has a shelf life.
> 3. Then the two closure checks in §4/§7.
>
> **Production while inert, measured 2026-08-23:** accrual is ZERO (still exactly 2 render-time
> items, newest 08-22 18:03), and **12 `section-editor` items completed with 0 refusals** — the
> demand control showing the armed refusal is not breaking healthy traffic.

> # ⚠ SUPERSEDED IN PART — 2026-08-23. Read this box before anything below it.
>
> This file said the lane was done bar one check. **It was not.** Running that check on 08-23
> surfaced a real defect in the ESCALATION half (shipped earlier, on trail `bb7f5d0e`): the
> `required_fields_missing` items it files were **unroutable — 2 of 2, a 100% failure rate**, and
> the first failure was **real production traffic** on `loans-application-tracker` at 13:32 on
> 08-22, hours before the canary. `required-fields-missing-handler` resolves the page by
> `spec->>'page_name'` and the component by `spec->>'slot_name'`; the producer supplied neither,
> and keyed on `<site_id>:<function>` where the sibling producer keys on `<page_id>:<slot_name>`,
> so the co-dedup claim was false too. **Both were one assumption — that reusing a TYPE meant
> inheriting its ROUTER. It does not.** Fixed at source in `eb918bd58`, pinned by
> `TestRequiredFieldsMissingItemsAreRoutable`, council trail `a0ef0b07`. **INERT until the next
> roll, so the bug does NOT close** — see §7, whose recommendation is now out of date by exactly
> this. Everything else below still holds: the refusal half is live, armed and canary-proven.

**Lane state: the refusal half is DONE and PROVEN; the ESCALATION half has a fix awaiting a roll
(see the box above), and check (c) is still waiting on ordinary traffic.** Read this file first; it is written so you do not have to re-derive
anything or re-run anything that has already been measured.

---

## 1. What the bug is, in one paragraph

Go's `text/template` runs with `missingkey=zero`, so a field the template references and the
content never supplied renders as **empty, with no error**. Page assembly then drops a
visually-empty section — so the content does not arrive broken, it does not arrive at all, and
nothing says so. It is the mechanism behind the recorded fleet-wide blanking of article bodies
(`bugs_closed/004`/`005`). Bug file:
`bugs_open/342_HANDOFF_2026-08-20_an_absent_required_field_still_renders_empty_and_silent_at_13_of_the_15_render_call_sites.md`

## 2. What is LIVE right now (all verified at the artefact, 2026-08-22, `agent-chassis` v1.0.1326)

| half | state | how it was proven |
|---|---|---|
| **Detection at the seam** (9 of 15 call sites pass a schema) | live since v1.0.1322 | prior lane; re-derived first-hand this lane, 9+6=15 closes |
| **Escalation** (live-page routes file `required_fields_missing`) | live since v1.0.1323 | binary probe, both replicas |
| **Editor REFUSAL** (`refuse_absent_required_fields` at `ApplySectionEditAction`'s one persist switch) | **live + ARMED (mig 551) + CANARY-PROVEN** | see §3 |
| **Chrome record** (`record_absent_required_fields`, 7 steps) | **live + ARMED (mig 550)** | 7/7 read from live rows, negative control |
| **Chrome REFUSAL** (same decision function, chrome store) | live in code, **deliberately UNARMED** | see §5 |

Council trail **`3626629a-f2bc-4089-9118-c1d6dd007807` — APPROVED at round 2.**

## 3. The canary — it passed, and you can re-run it

`./CANARY_342_editor_refusal.sh refuse` and `… control` (in this directory).
⚠ **Its two targets are deliberately NOT `deployed`. Never re-point it at a live row.**

- **refuse arm** (`0a1498b3`, tool-cta): step FAILED with *"refusing to persist — 2
  schema-required field(s) rendered empty (headline, trust_note)"*; stored `rendered_html`
  **byte-identical**, `updated_at` still **2026-07-17** (untouched); the
  `required_fields_missing` item **was filed anyway** (`detected`, both fields named).
- **control arm** (`9737d0d9`, use-cases-list): COMPLETED through `deploy_page`, `updated_at`
  moved to 18:05:04 — the write path ran. *An arm that merely stopped edits would have failed here.*

## 4. THE ONE OUTSTANDING ITEM — and it is a wait, not a task

The canary's third check (the DRIVING work item's terminal status under `bugs_open/344`) **could
not fire**: a CLI dispatch has no driving work item, and 0 trampled rows were measured in the
window. **It is UNEXERCISED, not "verified benign" — do not let anyone write that it passed.**

It will resolve itself in ordinary traffic. Measured 2026-08-22: **15 `section-editor`
orchestrations in the week of 08-17, 12 of them queue-driven (80%)**; 4 open items handled by
`section-editor`. So a real queue-driven edit will meet the refusal within days. One query
settles it:

```sql
SELECT wi.id, wi.item_type, wi.status, wi.completed_at, wi.retry_after,
       wi.retry_after > wi.completed_at AS trampled_344, left(wi.error, 200)
  FROM site_work_items wi
 WHERE wi.handler_agent = 'section-editor'
   AND (wi.error ILIKE '%refusing to persist%' OR wi.updated_at > '2026-08-22')
 ORDER BY wi.updated_at DESC LIMIT 10;
```

`status='complete'` **with** a refusal error ⇒ `bugs_open/344` confirmed on this route.
Parked `failed`/`needs_human_review` ⇒ 344 does not reach here and check (c) passes outright.
**Either result closes it.**

⚠ **Do NOT file a synthetic work item to force this.** It would put fabricated work on a live
queue to test a defect another lane already owns and has measured.

## 5. Things that look outstanding and are NOT — do not "fix" these

- **The chrome refusal is unarmed on purpose.** Zero rows can trigger it (re-measured at approval:
  candidate_pairs 0, rows_missing 0, while 813 required-field pairs and 72 chrome rows both exist
  — the JOIN is what is empty). Arming it now would arm an unexercisable refusal. **Flip trigger:
  the first `capability_gap` item whose `spec->>'finding_type'` is `required_fields_missing` (⚠ CORRECTED 2026-08-23 — this trigger used to name a `required_fields_missing` item with `surface='site_component'`, which the capability_gap rework means will now NEVER be filed, so the old trigger could not fire).** Not on a schedule.
- **The five no-schema non-tool components owe NOTHING.** Verified per component: three have no
  `{{.field}}` placeholders at all; two gate every reference with `{{if .x}}…{{else}}fallback{{end}}`,
  which `missingBareFields` deliberately does not report — confirmed at the artefact (no empty
  `<button>` or heading tags). **Do not author schemas for them.**
- **The six unwired call sites are correct as they are.** Each read at its own call site: two take
  raw template strings with no component row; the legacy head render loads only `html_template`
  *and its action is in no `GlobalActionRegistry` entry*; `RenderTemplateWithMap` is a different
  executor; the two audit probes remove fields on purpose.
- **Refusal at the SEAM is out of scope by owner ruling** (2026-08-02 §2 — new authority over
  content that renders successfully today at sites that never asked). The `bug_historian` seat
  noted (advisory, correctly) that this leaves a generic root cause patched per call site. That is
  the standing argument, not a settled one; if a future ruling licenses a seam-level default, this
  is the note that says why we did not take it.

## 6. Traps this lane hit — read before touching the same things

- **A binary carries ONE sha — its own build — not its ancestors.** Probing `/proc/1/exe` for your
  own commit sha returns OUT and means nothing.
- **A probe arm is only a control if the literal is NEW.** `refusing to persist` already existed
  in three other files, so that arm would have said PRESENT whatever shipped. Check with
  `git grep -c "<literal>" <commit>^` *before* probing. (`WRONG_CALLS.md`, 2026-08-22.)
- **A zero over zero candidates is not a finding** — count the candidate pairs a JOIN-based census
  tested, in the same statement. (Same day, same file.)
- **A count aggregated over two populations can be satisfied by one** — my canary monitor said
  "both arms terminal" when the refuse arm's two rows alone met `count(*) = 2`. Use
  `count(DISTINCT correlation_id)`.
- **A no-schema census over `content_components` over-counts ~20×** — 95 of 100 are tools,
  schema-less by design. `LANDMINES.md` entry, verifier returned STILL_VALID.
- **`section-editor.apply_edit` has no `error_step`** — hence the 344 interaction above. Do not
  "fix" it by adding one without reading 344's candidates first.

## 7. Should the bug be CLOSED?

**Not by a session, and not yet — but it is close, and here is the honest state.** The estate's
bar is *fixed AND live*. The defect's silence is fixed and live; the refusal is armed, live and
proven on the path that writes straight to a serving page; every other path either already
refuses before rendering (build, rerender) or cannot carry the check by construction. What keeps
the file open is §4's one unexercised check and the deliberate decision in §5 not to refuse at the
seam — i.e. the *class* is still patched per call site by design, and the bug's own title is about
those call sites. ~~**Recommendation: close it once §4's query returns a verdict**~~ **SUPERSEDED 2026-08-23 —
closure now needs BOTH that verdict AND the routability fix live and proven (a filed item that
the router actually classifies). Recommendation: close it once a `required_fields_missing` item
filed by the render-time producer reaches a NON-`failed` disposition**, recording which
of the two outcomes it was, and cite the closure as "live on v1.0.1326 as at <date>" rather than
bare "live".

## 8. The lane's files

`PLAN_2026-08-22_342_residuals.md` · `NOTES_342_absent_required.md` (append-only, newest at the
bottom — the canary and residual findings are at the end) · `README_where_we_are.md` (owner's
plain-prose log) · `SUMMARY_2026-08-22_the_refusal_half.md` · `RUNBOOK_342_absent_required.md`
(the guarded census shapes) · `CANARY_342_editor_refusal.sh` · `submission_342_refusal_half.json`.
Migrations: `sql_for_agents/550_bugfix_342_arm_chrome_absent_required_record.sql` and
`551_bugfix_342_arm_editor_absent_required_refusal.sql` (both applied and ledgered, both with
`_ROLLBACK` sidecars). Register: `STY-057` in `docs026_concept_register/register/styling-render-pipeline.md`.
