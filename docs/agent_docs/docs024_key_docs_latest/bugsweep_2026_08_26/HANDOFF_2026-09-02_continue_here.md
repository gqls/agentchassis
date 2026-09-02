# HANDOFF — bug sweep, 2026-09-02 (continuing the 2026-08-26 session)

**Read this instead of `HANDOFF_2026-08-26_continue_here.md` for current state**; that file is
still correct about mechanism and still worth reading for its §6, but a week had passed and
three of its status lines had expired. Where it was wrong I say so rather than editing it.

---

## 1. WHAT THE 08-26 HANDOFF SAID, AND WHAT WAS ACTUALLY TRUE A WEEK ON

| its claim | state 2026-09-02 |
|---|---|
| 359 CLOSED, proven | **HOLDS, and is now stronger** — but its census INSTRUMENT was under-reporting (§2) |
| 404 round 3 owed | **WAS still owed** — exactly 2 verdicts existed, both r2-era. r3 and r4 now submitted (§3) |
| 407 owner decision owed + nav rebuild | **RULED and RUN.** Owner said YES; the rebuild had never happened in 7 days (§4) |
| "next bug: 338" | **still unowned** — `who-owns.py` shows no owning workstream; its ACTIVE verdict comes from the 08-26 handoff commits themselves |

---

## 2. 359 — the acceptance HOLDS; the instrument did not

**Acceptance, re-established and stronger than at close.** The detector filed **8** items on
08-26/27 and NOTHING since. Six days of silence is also what a stalled detector looks like, so it
needed a demand control: the check is still live in the agent's config, and resolving the 8 item
keys to pages gives **exactly** the 8 pages today's census finds serving. Coverage is complete;
the silence is correct dedup.

**The defect (fixed, `4e26b1063`).** `audit-archived-still-serving.sh`'s `verdict()` ended
`if 200 -> SERVING; else correctly-absent`, so a curl `000` — no HTTP answer at all — was counted
as ABSENCE. Its two controls are per-DOMAIN and cached, so neither can see a per-PAGE transport
failure.

**What caught it: running the census TWICE.** Two runs ten minutes apart disagreed, 7 vs 8. And
its own self-test row `000 404 200 -> rc 0` had **pinned the defect as intended behaviour**, so the
suite passed at full strength while the census mis-scored.

⚠ **`scripts/audit-archived-still-serving.sh` now REFUSES a persistent `000`** (rc 2,
`UNJUDGEABLE-target-transport`) and retries once first. Today's stable reading:
**8 serving / 52 absent / 1 unjudgeable** (`boxingonline.com /contact.html`, a catch-all origin —
pre-existing, correctly refused).

`DISCOVERY_CHECK_ERROR` over 7 days: 48 rows, only **2** naming the new check, one of those a
site erroring on all three checks at once. Fleet-wide background, predates this work.

---

## 3. 404 — r3 REVISE, r4 SUBMITTED, verdict UNREAD

**Do not read `orchestration_states` for a verdict** — it returned ZERO rows for this correlation,
which reads like "never submitted" and means the runs aged out. The verdict is an artefact:
`diagnosis_artifacts WHERE correlation_id=... AND kind='council_report'`.

**r3 was gated by `editquality` [HIGH], seconded by `prior_art_librarian` [HIGH], and they were
right.** r3's whole purpose was "the narrative claims what the sketches do not show" — and I
corrected edits 1, 4 and 8 while leaving r2's stale sketch on **edit 7, the one edit the r2 gating
HIGH was about**. Full entry in `WRONG_CALLS.md`. The transferable half: I *had* verified the guard
in the migration file, and that real check is what made the unchecked claim feel safe.
**Checking the artefact is not checking what you are showing the reviewer.**

**r4 carries the check that would have caught it**: a loop over `plan.edits` asserting each sketch
CONTAINS the strings its rationale claims, plus a negative control that superseded text is gone.
Nine PASS lines, under a second. Reuse it — `scratchpad/submission_404_r4.json` has the shape.

**Nine seats approved r3. No seat has yet objected to the DESIGN** — every gating objection across
three rounds has been submission accuracy.

⚠ **One r3 objection is FALSE and r4 says so rather than complying:** `tooling_provenance` [low]
claimed migration 656's RAISEs cite `'655:'`. Measured: `'656:` **9** times, `'655:` **zero**.
Editing correct text to satisfy it would have introduced the defect it warned about.

**Owed next:** read the r4 verdict.
`RESUBMIT_CORR=f2e4ac2a-2bfc-4c82-ac99-d5fd7616edef`, `RUN_ORCH_ID=40639f27-fdca-4059-92bd-1a01d9f55f57`.
If APPROVED, the trailer is already resolvable — the r3/r4 commits carry `Council-Submitted:`, and
098 credits them at report time with no amend.

**Also corrected in r3/r4:** the 08-26 handoff answered guardian's "enumerate the callers" with
**"MEASURED: ZERO"**. That is true of realised ITEMS and does not answer the CALLER question,
because pre-fix the action DISCARDED the reason — a caller passing `template_changed` left an item
carrying no reason at all. Config-side there are **three** callers and one, `rerender-pages`,
passes `input_data.spec.reason` THROUGH. Named blind spot: a config scan cannot see a hand-run
kcat dispatch.

---

## 4. 407 — ⚖ OWNER RULED, rebuild RUN, served page VERIFIED

**The §B decision is CLOSED. The owner ruled 2026-09-02: a site's declaration DOES override the
three fleet-wide defaults** (`in_header`, `neverPrimaryTypes`, the child-URL bar). System and legal
pages stay non-overridable. **This is what the shipped code already does, so no code change.**

⚠ **The rebuild had NEVER RUN.** Migration 654 wrote the declaration at 22:25:04Z on 08-26; the
nav tables were last rebuilt at **20:43:23Z — an hour and three quarters EARLIER**. The declaration
sat live and unread for seven days and nothing surfaced it, because the row looked correct
throughout. **A migration applying is not the mechanism running.**

Dispatched via `kafka_publish_checked` (OPP-009), NOT the 149 lane's `TRIGGER_nav_rebuild.sh` —
that script's publish uses the racing `kubectl run -i` + stdin form which drops ~4 in 5 at exit 0.
Its pre-flight is still worth running separately (it confirmed zero nav rows a DELETE-and-rebuild
could not reproduce).

Result read from the step itself before inferring anything: `nav_declaration_source=site_config`,
`declared_missing`/`declared_ineligible`/`declared_flag_disagreed`/`declared_truncated_by_cap` all
null, `max_header_items_effective=9`.

**§3's evidence is discharged**: `protocol-tracker` and `adoption-tracker` — hardcoded into the
fleet tier-2 list to fix this once, and still absent when the bug was filed — are in the header by
the site's own declaration.

**Verified at the served page 16:25Z**, anchored on `<nav>`. Drain was PARTIAL and the sample
disagrees, which is the correct reading: `/`, `/services.html`, `/about.html`, `/tools.html` carry
the new header; `/model-directory.html` still carries the old eight (5 of 45 items complete).
**Re-sample rather than quoting that table.** If pages are still stale much later, the question is
the dispatch loop's claiming rate — the tables are already right and a second run only re-files items.

---

## 5. ADJACENT, FLAGGED NOT FIXED — all verified, none mine

- **`platform/livespec` is RED at committed HEAD.** `TestNoNewMigrationFileReadersOutsideTheAllowList`
  fails on `platform/orchestration/actions/write_audit_findings_origin_test.go` (405 lane,
  `ffa1707b3`). Both paths clean in the tree, so committed breakage, **unchanged for seven days**,
  on the shared drift-auditor allow-list. The test's own text names the two sanctioned remedies.
  The council's guardian seat has now flagged it in two consecutive rounds.
- **`_RELOCK` is an unclassified migration suffix** — the 097 trigger WARNs that
  `council-scope.sh` treats it as IN scope by the safe default. Classify it in
  `COUNCIL_SCOPE_NOT_THE_CHANGE_RE` if it is not the change (`bugs_open/314`).
- **`boxingonline.com` answers 200 on an invented URL** — a catch-all origin, so every archived
  page on it is unjudgeable by the 359 census. Pre-existing.
- **`boxingonline.com`'s `header_slots` was edited at 11:21Z today** by an active session
  (fight-calendar moved out of the slots and into a header CTA). Live work — do not touch.

---

## 6. THE FIRST COMMANDS FOR WHOEVER PICKS THIS UP

```bash
cd /home/ant/projects/agentchassis
git log --oneline -25
# 1. the 404 r4 verdict — artefacts, NOT orchestration_states
#    diagnosis_artifacts WHERE correlation_id='f2e4ac2a-...' AND kind='council_report'
# 2. 407 drain: re-sample five pages; do NOT re-dispatch nav-updater
# 3. 359 census (now refuses a transport failure instead of scoring it absent)
scripts/audit-archived-still-serving.sh
# 4. next bug: 338, voice-gate density rules — still unowned
```

~~**338 is still the next one to take**~~ — **DONE 2026-09-02 (later the same day), see below.**
`who-owns.py` printed OWNED for it, but the commits driving that verdict were this handoff's own
and there was no owning workstream directory — the prediction in §1 held exactly.

---

## 7. WHAT THE NEXT SESSION PICKS UP (appended 2026-09-02, after 338)

**Item 1 of §6 is DISCHARGED: 404's round-4 verdict is APPROVED** (`f2e4ac2a-…`, 16:33:30Z,
*"approved with 3 advisory objection(s) — none high-severity"*; `editquality`, `bug_historian`
and `debug_historian` objecting, all advisory). The r3/r4 commits carry `Council-Submitted:`,
so 098 credits them at report time with no amend. **Owed: read those three objections** — both
this round and 338's approved *while still finding real defects*, which is the case for reading
them rather than stopping at the verdict.

**338 is FIXED IN CODE and stays OPEN.** `425398a01` — the gate's per-page COUNTS and
per-sentence SHARES no longer gate a single value; the per-hit patterns and the em-dash RATE
still do. Council **APPROVED** (`106802fc-…`). Registered **CQ-035**. It is a Go change, so it
is **inert until the next chassis roll** — the artefact check (the two blank pages on
`leopardessconsulting.co.uk` and `oufe.com` filling) is owed after that, and is the only thing
standing between 338 and closure.
⚠ **Do not "fix" 338 again by raising a site's thresholds** — that disables the rule for the
site's PAGES too. And ⚠ **`bugs_open/338` §4's own remedy is WRONG and is corrected in the
file**: it says to drop `em_dash_density` and hand-roll a flat test; the rate already reduces to
"contains an em dash" below 333 words, and a flat test would re-gate the seven sites that
switched the rule off.

**NEW: `bugs_open/442`, found by the council reviewing the 338 fix.** A refused meta description
returns a nil error, so the step COMPLETEs and nothing downstream asserts on it — and the one
human-facing `result_message` names four refusal reasons while omitting all three copy-gate ones.
Unowned. Config-only candidate 3 (fix the message) is ~10 minutes and closes the misleading half.

**Still adjacent, still not fixed, verified at HEAD 2026-09-02:**
- `platform/livespec` red (§5) — **unchanged**.
- **A SECOND red of the same class:** `TestNoHandSpelledTombstonePredicate` fails at committed
  HEAD on `check_unrendered_page_imagery.go:156/197/202/207` (`a87746b77`, the 114/IMG-077 lane).
- `_RELOCK` unclassified migration suffix (§5) — **unchanged**, still WARNed by the 097 trigger.
- `WII-035` is a **duplicate row id** in the concept index (lines 415/416), predating this work.

**Process notes worth carrying** (full detail in `NOTES_bugsweep.md`): this tree swept my edits
into other sessions' commits **twice in one session**, both times on the fleet-wide append-only
ledgers (`LANDMINES.md`, `WRONG_CALLS.md`) — nothing was lost, but it reads as a failed edit, so
verify with `git show HEAD:<file>` rather than `git log -- <file>`. And a register entry
committed by pathspec without `000_concept_index.md` lands entry-without-row; the `pattern-check`
advisory catches it at commit time.
