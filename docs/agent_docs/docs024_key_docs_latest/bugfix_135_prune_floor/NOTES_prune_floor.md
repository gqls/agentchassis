# NOTES — bugs_open/135 prune floor

Append-only, newest at the bottom. Missteps are the point of this file, not an
appendix.

---

## (1) 2026-07-31 — picking the bug, and why who-owns.py could not do it alone

`scripts/who-owns.py` printed **"OWNED or recently active"** for every candidate I
tried (033, 066, 071, 072, 084, 085, 093, 096, 098, 113, 115, 121, 123). That is
not the tool failing — its window is 14 days and this tree takes ~1,500 commits a
week, so *everything* is recently active. The verdict is only useful as a pointer
to the owning workstream; it cannot answer "is someone on it **now**".

What discriminated: **last commit touching the bug file** (`for f in bugs_open/*.md;
do git log -1 --format='%ad %h' --date=short -- "$f"; done | sort`) plus a read of
the working tree for in-flight WIP. The tree is what caught the real hazard —
`asset_lock_guard.go` **untracked** and `derive_card_asset_action.go`,
`derive_brand_head_assets_action.go`, `plan_sections_action.go`,
`diagnose_assemble_bundle_action.go` all dirty, i.e. a lane mid-fix on
`bugs_open/143`/`152`/`155` **which no git command would have shown me**. Memory
already says this ("who-owns reads COMMITS, so a session mid-fix is invisible");
today it was load-bearing, not decorative.

135 was chosen because the case file *hands it over explicitly*: split out of the
markdown-indexing plan, "pre-existing and independent … it belongs here, not as a
rider on someone else's change."

## (2) 2026-07-31 — grounding the bug rather than trusting it

The case file's figures are from 07-28. Re-measured 07-31:

```
 gqls/agentchassis | alias     | d98010e8… |    29
 gqls/agentchassis | func      | d98010e8… |  3048
 gqls/agentchassis | interface | d98010e8… |    33
 gqls/agentchassis | method    | d98010e8… |  1025
 gqls/agentchassis | struct    | d98010e8… |   857
```

4,992 rows, **one repo, one commit, five kinds**. Two things follow that the case
file could not have said: (a) arming a 0.5 floor today refuses nothing, because
every cohort ratio is ~1.0 on a healthy run; (b) `interface` (33) and `alias` (29)
are small enough that a genuine refactor *could* trip the floor on one of them —
which is exactly why the case file's insistence on a **resolvable** floor matters,
and it moved from "nice" to "load-bearing" once I saw the size of those cohorts.

Also read the CHECK constraint: `kind` is limited to eight Go kinds, so markdown
rows cannot exist yet. Worth knowing before writing anything that reasons about a
"doc" cohort — the guard protects a class that does not exist *yet*, and that is
deliberate, not an oversight.

## (3) 2026-07-31 — the case file's 016b quotation is a paraphrase, not a quote

The case file quotes 016b §9 as: *"the identical DELETE+INSERT rebuild pattern
destroyed a working A\* pathfinding game; recurred independently on a second site
months later."* I went to cite that verbatim in the council submission and
**could not find that sentence in 016b.** What is actually there (:147, :174,
:4555) is a longer account of `game-pathfinding` being discarded by a full rebuild
that regenerated from `plan_sections`, plus the generalisation "child rows behind a
DELETE+INSERT rebuild".

Same substance, different words. Not a defect in the case file — a paraphrase
doing its job — but I nearly passed it to a council seat inside quotation marks,
which would have been my error, not the case file's. **Grounded the submission on
the text I actually read instead.** The cheap check is one grep before quoting.

## (4) 2026-07-31 — MISSTEP: I diagnosed another session's uncommitted work as a HEAD defect

Recorded in full in `WRONG_CALLS.md`. Short version: `go test
./platform/orchestration/actions/` would not compile. I ran `git log -1` on the
failing **test** file, saw a 07-29 commit, saw `git diff HEAD` on that file was
empty, and concluded the package's tests had been broken at HEAD for two days by
that commit. I then "fixed" it.

Wrong. The type had been changed in the **action** file — dirty in the tree, from a
live session — and `asset_lock_guard.go`, which defines the new type, was
**untracked**. `git show HEAD:<action file>` shows `lockedBrandHeadKeys` returning
`map[string]bool` at HEAD, where the test compiles fine. My "fix" would have
*broken* HEAD if committed.

What caught it: the *second* breakage appearing the moment I fixed the first
(duplicate `equalStrings`). Two independent compile errors in one package within
minutes is a tree-state smell, not two coincidental stale commits. I reverted my
edit to that file, built a clean tree from `git archive HEAD` with only my own
files overlaid, and the whole `actions` package went green — which is also the only
statement about my own change that is worth anything on a shared tree.

The cheap check I skipped: **`git status <file>` and `git show HEAD:<file>` before
attributing a compile error to a commit.** A compile error names the file it
*fails* in, not the file that *changed*.

## (5) 2026-07-31 — a test of mine was wrong about my own code, and it was right to be

`TestEvaluatePruneFloorRefusesATruncatedRun` asserted `kind=method` would sort
worst-first. It is `kind=func`: 1160/3048 = 38.06%, 400/1025 = 39.02%. I had
eyeballed "400 of 1025 looks worse". The code was right and the test was wrong;
fixed the expectation and wrote the three ratios into the test as a comment so the
next reader does not redo the arithmetic in their head. Cheap, but it is the class
of thing that becomes a "flaky test" if you fix it by loosening the assertion.

## (6) 2026-07-31 — what I deliberately did NOT do

- **Did not convert the three sibling call sites** (`populate_nav_tables`,
  `link_registry`, `save_page_sections`). Each needs its own cohorts measured, and
  two are another lane's live territory this week. They are named in
  `prune_floor.go`'s header and in CTXA-025's open-review-question instead. If a
  council seat objects that a shared rule with one consumer is speculative
  generality, that is a fair objection and the answer is in the grep: three
  identical DELETE shapes, one of which has already destroyed live content once.
- **Did not use the analyser's `file_count`** for the whole-repo signal (the case
  file's uncosted candidate 4). It has nowhere to be stored for a
  previous-run comparison, so it would need a schema change; `count(DISTINCT path)`
  is the same signal against data already in the table.
- **Did not make the doc_notes suppression window configurable.** One more knob on
  a surface with no automated consumer. Hardcoded 24h, said so in the comment.

## (7) 2026-07-31 — [UNVERIFIED at time of writing] the induction has not run yet

The guard is committed and the tests prove the rule. **The refusal branch has not
been seen to fire in production**, because Go is inert until an image rolls. Until
the induction in the RUNBOOK has run, the honest state is "shipped, unproven in
situ" — the case file is explicit that a green healthy run proves only inertness,
and CTXA-025's status line says the same. Updated below when it has run.

## (8) 2026-07-31 — the SQL, PREPAREd against the live schema before it shipped

`go build` cannot parse SQL, so all three new statements were PREPAREd and
EXECUTEd against `clients_db` (not asserted from a read):

| statement | at the current commit | at a bogus commit_sha |
|---|---|---|
| kind cohorts | func 3048/3048, method 1025/1025, struct 857/857, interface 33/33, alias 29/29 | 0 of each |
| path cohort | 592 / 592 | 0 / 592 |
| doc_notes suppression probe | `f` (no recent refusal note) | — |

The bogus-commit column is the useful one: it is exactly what a total-loss run
measures, and every cohort reads 0%, which is what the floor refuses on. So the
measurement half is proven against real data *before* the roll — what is still
unproven is the wiring that acts on it.

**A figure, and a CORRECTION to what I first wrote about it.** Distinct paths are
**592** today.

> **CORRECTED, same session, before anyone else read it:** I first wrote here — and,
> worse, in the round-2 council submission's `grounded_in` — that 592 "corrects a
> figure carried in an older register entry: … not 530." **That is wrong twice
> over.** 530 appears nowhere in the register: it is a number *I* invented as a
> test fixture in `prune_floor_test.go`, and I then read my own fixture back as
> though it were a live measurement. The register's actual figure is CTXA-011's
> evidence line, `"4,155 symbols; 499 distinct paths"`, from a reindex months ago —
> which 592 does supersede, but that is a stale-evidence line doing its job, not a
> figure anyone carried forward wrongly.
>
> **What caught it:** writing "corrected in CTXA-025's text" and then going to make
> that edit — CTXA-025 states no path count at all, so there was nothing to correct,
> which is when the claim fell apart. The cheap check is the one I skipped:
> `grep` for the number you say you are correcting **before** saying so. The
> submission is already in flight and cannot be amended (forward-only), so if a
> seat objects to that line it will be objecting to a real error of mine, and the
> answer is this paragraph.

The live figure matters for one reason only: the path cohort is now the largest
denominator in the guard (592 paths vs 33 for the smallest kind), so it is the
cohort *least* likely to trip on ordinary churn and the most trustworthy signal
that a run genuinely did not see the repo.

## (9) 2026-07-31 16:11 — council round 1: REVISE, gating objection from `reuse_agent` (HIGH)

13 reviewers, 4 abstained, 0 unreadable. `editquality` **approved** with two low
objections. The gating one is worth quoting because it is right:

> "Landmine entries for both `diagnose_code_lookup` and `index_code_symbols`
> reference a prior fix dated 2026-07-27 … specifically about the reader side of
> this exact mechanism. The plan's grounded_in … never names migration 243 or
> council 18fe4035, and gives no indication anyone checked whether the
> freshness/staleness problem this edit solves (mixedCommitNote) was already
> addressed, partially addressed, or explicitly deferred by that prior council
> decision."

**It is a fair hit.** I had read that code — `bodyCoverageNote` is right above the
function I extended — and I still submitted without showing the check. The seat
could not tell the difference between "checked and complementary" and "never
looked", and its founding incident is precisely a reinvented mechanism.

**Checked properly, four ways, and the answer is COMPLEMENTARY:**

1. **What 18fe4035 actually did on the reader side**, from its own trail note
   (`doc_notes`, `subject_key='diagnose_code_lookup'`, 07-27 "later the same
   day"): body-OR-declaration content search, `[body]`/`[decl]` marking, excerpt
   around the match, empty-result-as-an-answer, and the degrade-window note. **The
   note never mentions commit identity or staleness.** Its own list of what it left
   owed is explicit (council-gate has no code_lookup step; markdown unreachable
   pending the `kind` CHECK relaxation) and a mixed-commit index is not on it.
2. **The freshness half is a different lineage** — `bugs_open/059` + 108 defect A,
   commit `87d0bcf97`, migration 250 — and is structurally *single-row*
   (`ORDER BY updated_at DESC LIMIT 1`). A one-row read cannot express "and 400
   other rows describe an older commit". Different question, not a second opinion.
3. **Nothing in the tree measures commit spread.** `grep -rn 'DISTINCT
   commit_sha\|DISTINCT COALESCE(commit_sha' --include=*.go platform/ internal/
   pkg/` → exactly one line, and it is mine.
4. **Never considered, so neither addressed nor deferred.** `grep -rni
   'spans.*commit|multiple commit|mixed commit|two commit_sha'` over
   `architecture_review/`, migrations 243/244 and `bugs_closed/108` → zero hits.

And the point that actually favours the plan, from the same note the seat says I
ignored: *"codeIndexScope … is used by BOTH lanes … One judgement, not a sibling
copy that drifts (016b §9)."* My change extends **that** helper and wires **both**
lanes in one commit. The reuse discipline was there; the *evidence* was not.

**Round 2 changes no code.** An evidence objection is answered with evidence;
churning the fix to look responsive would be the wrong move and would invalidate
the tests. Resubmitted on the same correlation with the four checks above in
`grounded_in`, plus the prior-art search for the prune rule itself (which I *had*
run — `grep -rn 'prune_floor|floor_ratio|reconcil'` and the `DELETE FROM` census —
and also failed to show), plus explicit answers to editquality's two lows
(fail-closed on a transient measurement error is deliberate because the cost is
asymmetric; the 24h note window is deliberately not a knob).

**The transferable lesson, and it is not about this bug:** I ran the right
searches and reported the conclusions instead of the searches. A reviewer cannot
audit a conclusion. *Put the grep in the submission, not the answer to the grep.*

## (10) 2026-07-31 — routing note: no 016b §9 entry, deliberately

CLAUDE.md's own routing says §9 is for "how the SYSTEM fails (you have a
symptom)", and LANDMINES.md is for "what will mislead you when you TOUCH
something (no symptom yet)". `135` is explicitly a latent defect with **no
symptom** — nothing has failed — so the prospective check belongs in LANDMINES.md
(appended, and synced into `doc_notes` with `landmines-sync.py --apply`), and the
callable mechanism in the register (CTXA-025). Writing a §9 entry as well would
put a symptom-shaped account of something that has never happened into the file
sessions read *after* they have a symptom. Recorded because "add the pattern to
016b §9" is the reflex, and here the reflex is wrong.
