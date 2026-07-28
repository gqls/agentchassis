# Bugs sweep — CONTINUE HERE (written 2026-07-28, end of the first session)

**What this thread is.** The owner pointed a session at `/bugs_open/` with: work through
the remaining bugs, prefer a framework fix over a per-case one, check with the diagnosis
loop first and the council later, don't collide with other threads, commit what should
ride the next chassis build, close and move tickets when done.

**Read this file first, then `NOTES_bugs_sweep.md` for the misstep log.**

---

## STATE — where the backlog stands

**100 closed, 47 open** (the open count *rose* during the session: other threads file
faster than one thread closes).

### Closed by this thread
| bug | what it was | proof it is done |
|---|---|---|
| **095** | empty page assembly reported COMPLETED | live v1.0.1177; no residual, fleet count back to 0 |
| **103** | tool pages published their build brief as the public meta description | 17 rows backfilled, 17 pages re-rendered, **17/17 verified on served HTML** |
| **105** | `EvidenceFact.Kind` declared and read by nothing | live v1.0.1180, verified by a **removed-symbol** control |
| **094** | single-page script's `section_data_resolved` branch could not run | live v1.0.1182, **branch actually driven end to end** |

### Fixed and live, still open for a stated residual
| bug | residual — this is the next action |
|---|---|
| **080** | robot-hands.com still serves BOTH `/news.html` and `/news/index.html`. **The shape is DECIDED** (see below); the *execution* is blocked on 081 |
| **091** | candidate 2 (honest reporting) live at 3 sites. **Candidate 1 — the dropped finding itself — is owed** and belongs to `work_item_completion_integrity` |
| **109** | **1 of 4 maps** derived. `mergeIntoRenderContextEnhanced`, `mergeIntoRenderContext`, `contextToInterfaceMap` still hand-maintained |

### Advanced by evidence, deliberately no code
| bug | finding |
|---|---|
| **081** | **candidate 2 is BLOCKED, not merely hard** — see below |
| **097** | diagnosis returned `UNVERIFIABLE`; blocker is now **108**, not 029 |
| **108** | now has a **measured cost** — it defeated the 097 diagnosis |
| **029** | fresh instance contributed with an exact timeline |

---

## THE THREE FINDINGS THAT SHOULD SURVIVE THIS SESSION

### 1. The news URL shape is DECIDED — do not re-open it
`/news.html` vs `/news/index.html` is settled by the **section-index family convention**:
`page_canonical.go` (doc 029 Phase 0, extended by `bugs_closed/015`) fixes it as
**`name=<section>-index`, `url=/<section>/index.html`, `page_type=<flavour>`**.
`relojistas.com` is the live worked example and the **only** conforming site in the fleet.
I wrongly called this "an owner decision" in a bug file and a commit; the owner corrected
it. Logged in `WRONG_CALLS.md`.

### 2. `bugs_open/081` candidate 2 cannot be built as specified
Its predicate needs to identify "a deployed page already doing the news job under the wrong
type". The structural, language-independent signal is *carries the `news-listing`
component* — and fleet-wide that returns 4 rows, one of which is
`robot-hands.com/gripper-catalog-index`, the **catalog** index that embeds a news feed.
Their section shapes are **byte-identical** (`["news-listing"]`), and so is a correctly
typed news page. `classification.content_features.news_feed` names no target page either.
**Shipping the predicate would offer the catalog index for re-typing and break a live
page.** Fix the signal (record the intended page at creation) or take candidate 3.

### 3. `bugs_open/108` is upstream of the diagnosis loop's usefulness
`code_symbols` holds **one** `commit_sha` (`e19aa5d`, 2026-07-24), now **970 commits**
behind HEAD. It answered *"0 rows … The query was RUN and found nothing; this is not an
unanswered question"* for `RepairPageLinks` — which exists at `link_repair.go:139` and is
called at `validate_page_content.go:357`. **Any `content`/`symbol` answer about code added
since 07-24 is a false zero, phrased as a positive absence claim.** Fix 108 before
spending another diagnosis run that depends on the code tier.

---

## NEXT ACTIONS, in the order I would take them

1. **`bugs_open/108`** — highest leverage. It is unowned, and until it is fixed the
   diagnosis loop's code tier actively misleads. Re-index, then re-fire 097.
2. **`bugs_open/097`** — after 108. Its question is unchanged: which of three mechanisms
   was on the 07-25 build path. Do NOT pick between its fix candidates first.
3. **`bugs_open/091` candidate 1** — coordinate with `work_item_completion_integrity`; the
   council also flagged that `apply_gap_plan`'s `ON CONFLICT DO NOTHING` on
   `site_work_items` is itself a dedup-rule violation.
4. **`bugs_open/109`** — the remaining 3 of 4 maps.
5. Untouched and unowned: **100, 104, 111, 114, 115, 117, 118**.

---

## HOW TO WORK THIS BACKLOG (what actually paid off)

- **`who-owns.py`'s VERDICT line is useless here** — it said "OWNED or recently active" for
  **42 of 42** bugs, because ~1,500 commits/week touches everything. The discriminating
  part is its `=== likely OWNING workstream(s) ===` section, which prints
  `(none identified)`. That found 18 genuinely unowned bugs.
- **Grep for a SECOND call site before fixing the filed one.** It paid off twice: 103 had
  an unnamed second site (`create_tool_component_action.go:265`), and 091 had two.
- **Every council submission earned its cost.** Six ran, all APPROVED, and objections
  changed the code four times — including catching me shipping an accessor with no reader
  inside the fix for "a field with no reader".
- **Test against `git archive HEAD` + your files overlaid.** The shared tree was broken by
  other sessions repeatedly (`agentenv`, `NewSagaCoordinator`, `pattern-check.py`).
- **Verify with a marker the change CREATED, plus a control that must read 0.**

## TRAPS THIS SESSION PAID FOR (full accounts in `WRONG_CALLS.md`)

- **A chassis roll kills every orchestration mid-step**, including your own. Check
  `orchestration_states WHERE status NOT IN ('COMPLETED','FAILED','CANCELLED')` BEFORE
  rolling. CLAUDE.md warns about dispatching *after* a restart; the inverse is unwritten.
- **`status` cannot tell you a run is alive — only `updated_at` can.** I reported a dead
  council run as "still deliberating" four times over 70 minutes.
- **A zero ages worse than any other figure.** "0 live instances fleet-wide" was true at
  18:05 and false at 18:35, inside one council round.
- **Use `git commit -F -` with a quoted heredoc for every message longer than one plain
  sentence.** Backticks in `-m` are command substitution; it ate a word again.
- **A hung pod is evidence on a clock** — reaped by job cleanup and rolls. Capture
  `-o yaml` and logs when you find one; don't point a later reader at a name.
