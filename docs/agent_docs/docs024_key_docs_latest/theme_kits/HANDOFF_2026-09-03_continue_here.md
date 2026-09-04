# HANDOFF — theme kits lane, 2026-09-03 (session "theme kits")
## re-verified 2026-09-04 11:14Z — READ §0 FIRST, one status below has flipped

> ## 0. WHAT CHANGED OVERNIGHT, verified 2026-09-04 11:14Z
>
> **THE ROUND-4 GUARD IS NOW LIVE.** This file said "COMMITTED AND NOT LIVE" and that is
> now false. A roll happened overnight (`agent-chassis-ffc9ddff9-*`, a different
> ReplicaSet from the `85c4984f77-*` pods this file was written against, ~13 h old at
> re-check). Probed with both controls:
>
> | needle | result |
> |---|---|
> | **`classifierDesignIntentState`** (the round-4 guard) | **PRESENT** |
> | `apply_theme_kit` (positive control — Phase 1) | PRESENT |
> | `zzz_not_a_real_symbol_zzz` (negative control) | absent — *the probe discriminates* |
>
> **Everything else re-verified and UNCHANGED:** `to_regclass` returns both tables; **4**
> kits and **14** archetypes seeded; **adoption is still 0**; and **0** rows carry the new
> `design_intent_supersede_risk` field, which is exactly consistent with adoption 0 —
> nothing has applied a kit, so the guard has never had occasion to fire.
>
> **The lesson is the file's own, paid back within a day.** §7 told the next session to
> roll and re-probe. Doing that as the FIRST action changed the answer to the question the
> rest of this document is written around. **Re-run §1's four checks before believing any
> status here** — this file has now been stale in both directions inside 48 hours.

## IF YOU ARE PICKING THIS UP COLD — do these, in this order

1. **Re-run §1's four checks.** They are four INDEPENDENT facts and none implies
   another; this file has been stale in both directions inside 48 hours. Commands and
   their traps: `RUNBOOK_theme_kits.md` §1.
2. **Read §2 before planning anything.** Three of a kit's four dimensions cannot change
   how a site looks. That is measured, not argued, and it is why there is no obvious
   Phase 2.
3. **Do NOT adopt a kit onto any site.** The council's architecture seat made this a
   precondition, not a note — §4. Adoption is 0 and should stay 0 until an owner decision
   lands.
4. **The three owner decisions in §7 are the real blockers.** They are written for the
   owner in plain prose in `README_where_we_are.md`. Nothing in this lane needs building
   until one of them is answered.
5. **If you are about to touch `content_components`, run
   `grep -n 'content_components' LANDMINES.md` first.** §6 explains why: this lane wrote
   that landmine and then walked into it twice the next day.

**There is no urgent work here.** Nothing is broken, nothing is running, no site is
affected. The lane is finished until a decision arrives — if you came here looking for a
task, §7 is the honest answer and it is mostly "ask the owner".

---

**Supersedes `HANDOFF_2026-09-02_continue_here.md`.** That file is still accurate about
what was BUILT and is the fuller account of the eight pre-apply defects and the owner's
ruling; read it second. Read this first for state, and read §1 before believing anything.

**The lane now has its standing five** (`PLAN_2026-09-02_theme_kits.md`,
`RUNBOOK_theme_kits.md`, `NOTES_theme_kits.md`, `README_where_we_are.md`,
`SUMMARY_2026-09-03_theme_kits.md`). The design is migrated out of the approved plan file
at `/home/ant/.claude/plans/please-think-hard-about-starry-locket.md`, which remains the
source for corrections C1–C10.

---

## 1. STATE — four independent facts, none implying another (re-verified 2026-09-04)

| fact | value | how |
|---|---|---|
| binary | **LIVE**, `agent-chassis` `v1.0.1355` | `/proc/1/exe` capability probe, positive AND negative control |
| schema | **APPLIED** 2026-09-02, migrations 689 + 691 | `to_regclass` both tables; 4 kits, 14 fleet archetypes |
| adoption | **0** | `SELECT count(*) FROM site_specs WHERE aspect='theme_kit_adoption' AND is_current` |
| round-4 guard | **LIVE** as of 2026-09-04 (was committed-not-live when this file was written) | `/proc/1/exe` probe for `classifierDesignIntentState`, both controls — see §0 |

**Nothing has adopted a kit.** Every kit-conditional branch is live, reachable, and has
never run. **Cite this lane as "built and reachable", never as "working".** The four facts
are independent: the binary can carry code whose tables are absent (it degrades silently
by design), the tables can exist with nothing using them, and the guard shipped a day
after the rest. **Check all four.**

The RUNBOOK has the commands and the traps for all three. The one worth repeating here:
`psql` through `kubectl exec` takes 1–3 minutes on this cluster, so put SQL in a file, pipe
it with `-f -`, and run it in the background rather than fighting a 120 s timeout.

---

## 2. THE HEADLINE — three of a kit's four dimensions cannot change how a site looks, and the fourth is reachable without kits

This is the finding the next session should act on, and it is now measured from four
directions rather than argued.

| dimension | does adopting a kit change anything? |
|---|---|
| palette | **No.** `render_css_from_spec` is spec-wins on all 8 core slots and `analyze_design` reads `design_intent`, never the composed palette row. Measured at the artefact: gamedesign.uk resolved a hand-chosen palette (`palette_source=mission_hint`, first time that rung ever fired fleet-wide) and served **none** of its eight core colours. **This is the owner's ruling working, not a defect** — the lever on served colour is the BRIEF. |
| page structure | **Barely.** 1,022 of 1,083 live pages (**94.4%**) match no exact `defaultSectionsForPage` output, and 5.6% is an UPPER bound because a planner can choose those lists unaided. The structure lever is the planner's prompt. |
| chrome | **No — found 2026-09-03.** All four kits pin `header-theme-chrome`/`footer-theme-chrome`, which is exactly the row `ResolveChromeComponent` returns for a site with NO pin at all (proven under the POOL predicate, so no tiebreak is involved). The pins are no-ops. |
| layout | **Yes — and it is the only one.** But two of the four kits name a layout the tag matcher would have picked anyway: `tool-portal-light` (14 sites reach it by tags) and `brochure-formal` (the resolver's hard fallback, so a kit there dresses the default up as a choice). |

**So the honest open question is not "what else should a kit bundle" — it is whether a kit
is the right vehicle at all**, or whether the entire value is in **layout reachability**,
which is cheaper and more direct. That question is put to the owner in
`README_where_we_are.md` and is not mine to answer. **Do not build Phase 2 before it is
answered.**

`soft-editorial` is the one kit worth keeping and the register now says why in the honest
form: [MEASURED 2026-09-03, `bugs_open/445`] it scores above zero on 27 of 33 sites but
only at **0.50, the same-scheme bonus ALONE with zero tag hits**, and is one of nine of
eighteen layouts no site's tags reach at all. **It is a deliberate route to an otherwise
unreachable layout — a workaround for a tag-vocabulary defect, not a design choice.**
`docs-sidebar` is **pre-positioned, not demanded**. **Do not curate by taste**;
`bugs_open/445` is building a fleet scorer, and a kit candidate should be simulated against
the live fleet before it is seeded. Adoption is 0, so reseeding is free.

---

## 3. THE DEFECT THE COUNCIL FOUND — a kit applied before classification loses palette AND typography. The REPORTING half is fixed and live; the ordering half is `bugs_open/438`'s, not this lane's

**Found by the council gate, round 2. Not by me.** Recorded with three costed remedies as
**`bugs_open/438` §6d** (a CONTRIB — 438 §6a-bis already owns the mechanism) and documented
in `apply_theme_kit_action.go`'s own header.

On the FRESH path (`082` with no `--from`), `domain-research-classifier` writes
`design_intent` **after** `apply_theme_kit` does, and `write_site_spec` supersedes the
current row after a deep merge in which **scalar keys are overwritten by the incoming
value**. `[VERIFIED 2026-09-03 by reading the file]` **there is no guard** — grep for
`classifier`/`domain-research` in the action finds only comments about the ruling, never a
predicate.

- **layout SURVIVES** (aspect `theme_kit_adoption`, which the classifier does not write).
- **palette is discarded** — moot for appearance, per §2.
- **TYPOGRAPHY IS DISCARDED, AND TYPOGRAPHY IS THE DIMENSION THAT RENDERS.** This is the
  one that costs something.
- ⚠ **`design_intent.<dim>.locked` does NOT protect against this.** It is read when
  `apply_theme_kit` writes; **nothing makes the classifier respect it** — and the key
  survives the deep merge while the values do not, so the row ends up **asserting a human
  pin over a classifier's values.** That is worse than having no pin. Do not recommend it
  as protection against this path.

**So a kit works on an ALREADY-CLASSIFIED site and is defeated on a new one — the inverse
of the owner's *"by default it can start with a theme."***

**THE GUARD IS BUILT — `b18091066`, and this paragraph used to say it was deferred.**
Council round 3 gated on exactly that deferral, and two rounds gating on the same missing
guard is the council saying the deferral *was* the defect. `classifierDesignIntentState()`
asks whether `domain-research-classifier` has ever written this site's `design_intent`; if
not, the apply records `design_intent_supersede_risk` in **three** places — a WARN naming
the mechanism, the `theme_kit_adoption` spec so it is durable and queryable, and the
action's result. Three-state string, never a bool, so a read failure cannot be recorded as
"no risk". Proven by two mutations, both red with the right message, restored green,
evidence in the test header; the predicate discriminates 38 of 39 live sites.

> ~~⚠ **THE GUARD IS COMMITTED AND NOT LIVE.**~~ **SUPERSEDED 2026-09-04 — it is LIVE
> now, see §0.** Kept because the *method* is the point: the council's `debug_historian`
> seat asked for a pod probe and was right to. I had recorded the guard as "inert because
> adoption is 0", which was true and hid a second, stronger reason — it was not in the
> image either. **On 2026-09-03 the probe said absent; on 2026-09-04, after a roll, it
> says PRESENT.** Two reasons for inertness, and only one of them has cleared.

**It REPORTS, it does not REFUSE, and that is the judgement most open to challenge.**
Layout survives on a different aspect and is the only dimension a kit moves, so refusing
the whole apply would throw away the part that works to protect the part that does not.

**What is still NOT fixed, and stays `bugs_open/438`'s:** the ordering itself. The
classifier still supersedes and the kit's typography is still lost on the fresh path — the
guard only makes the loss visible instead of silent. The other two candidates are
architecture-scope (make the classifier respect `locked`, which changes its write authority
over a shared aspect) or build on 438's own defect (write `mission.preferred_typography`,
which survives that path only by accident).
**The guard is LIVE but has never fired: adoption is 0, and `[VERIFIED 2026-09-04]` zero
`theme_kit_adoption` rows carry the `design_intent_supersede_risk` field.** So it is
reachable and unexercised — the same distinction this lane insists on everywhere else.

---

## 4. Council gate — APPROVED at round 4, after three revises that each found something real

**Trail correlation `bed139b2-f512-436a-9ba8-ff2fbfade8ef`** (use this — it is the key the
artefacts are written under).

| round | verdict | what it found |
|---|---|---|
| 1 | `revise` 2026-09-02 21:43Z | the rationale claimed a typography guard the sketch never showed. Correct: a reviewer judges the submission, not the repository. |
| 2 | `revise` 2026-09-03 15:32Z | **the §3 defect.** The best output of the whole review. |
| 3 | `revise` 2026-09-03 15:56Z | gated by **`bug_historian`**: I had diagnosed the defect, accepted it and shipped nothing. Fair. **This is what caused the guard to be built.** |
| 4 | **`approved`** 2026-09-03 16:19Z | *"approved with 7 advisory objection(s) — none high-severity"*, 3 abstained. Carried the guard. |

**So `Council-Reviewed: bed139b2-f512-436a-9ba8-ff2fbfade8ef` is now legitimate — an
approved verdict has been read.** Earlier commits carry `Council-Submitted:` and 098
credits them automatically now the correlation has approved; forward-only forbids amending
them and none is needed.

⚠ **THE APPROVAL IS NOT A CLEAN BILL. Two architecture-seat objections read as a GATE on
Phase 2, and this lane agrees with both:**

- *"All four seeded kits pin chrome identical to the unpinned default — the chrome
  dimension of a kit is currently a no-op. **Shipping more kits or adopters before this is
  addressed overstates what a kit does.**"*
- *"Palette cannot reach the served stylesheet under the current render-overlay precedence
  — `theme_kits.palette_id` is **structurally decorative** … it should **block further
  palette-bearing kit adoption** until the precedence is fixed or the capability is
  explicitly dropped from the contract."*

**So: do not adopt a kit onto any site until the contract states what a kit actually
delivers.** That is the council reaching §2's conclusion independently and turning it into
a precondition.

⚠ **READING A VERDICT ON A RESUBMITTED CORRELATION IS A TRAP, and another lane landmined
it the same day** (`a7352e2ca`). A resubmit reuses the correlation, so the query the
trigger itself prints — `… ORDER BY created_at DESC LIMIT 1` — can hand you the OLD
verdict, and it reads as your revision being rejected. **COUNT THE ROWS INSTEAD:**
`SELECT count(*) … AND kind='council_report'` must equal the number of rounds you have
submitted, so 3 rows after a fourth submission means round 4 has not landed. That is how
the four verdicts above were read; do not use `LIMIT 1` on this table across a
resubmission. `grep -n 'RESUBMITTED council round' LANDMINES.md`.

⚠ **`scripts/verify-head-builds.sh --test` is RED at HEAD, and NOT for this lane.**
`[VERIFIED 2026-09-03]` two failures, both other lanes', both proven not ours (each
mechanism appears **0** times in all three files this lane touched):

- `render_seam_one_spelling_test.go` — UNDECLARED template executor
  `renderFailWorkItemMessage`, introduced by `83407cd37` (the `bugs_open/440` lane's
  phase 3, "BUILT and HELD"). Their guard firing exactly as designed: a new executor is a
  new dialect and must be declared with its language.
- `check_stylesheet_gutted_test.go` — `canonicalCSSTokens` declares four tokens the check
  does not police (`--color-accent-ink`, `--color-accent-text`, `--color-cta-bg-ink`,
  `--color-primary-ink`).

**This lane's own tests pass at HEAD** (`go test -run 'ClassifierDesignIntentState|SupersedeRiskConstants'`
→ ok), and the **build** half of `verify-head-builds.sh` is green. Do not spend time
debugging these, and do not patch another lane's guard to make your run go green — the
first one is asking its author a question only they can answer.

---

## 5. What was committed today

| commit | what |
|---|---|
| `28aeb4ca0` | §3a(ii) fixed: the kit layout arm recorded a candidate that was never scored — now an empty slice, which the consumer omits |
| `a113fe055` | DES-085: stale status corrected (it read "not applied, not rolled" for a day after both became true), plus the chrome no-op finding |
| `0b1dcc62c` | the standing five, created late and saying so, + a WRONG_CALLS entry |
| `cd84cdd5a` | **withdrew a retraction that was itself wrong** — see §6 |
| `58152c5be` | chrome eligibility has TWO predicates; my figure was right for pins only |
| `51cb87dfe` | the unowned seam question put to the owner |
| `e28df777a` | LANDMINES: fourth sighting of the name/function trap, and the retraction direction it did not cover |
| `c03280b20` | 438 §6d CONTRIB + `apply_theme_kit`'s header (it documented `fill_gaps` as the default when the shipped default is `start`) |
| `4b1b075bf`, `e8f08cc80`, `08286e12d` | NOTES, register and the owner-facing account of §3 |
| `efb9a490b`, `b19637c2f` | this handoff, and the 09-02 one marked superseded at the top |
| `d33a21300` | NOTES: the backtick trap ate a word from a commit message; `gofmt -l` cannot fail when chained with `&&` |
| **`b18091066`** | **the round-4 guard + its mutation-proven test** — the remedy the council gated on twice |
| `c0545a3f0` | 438 §6d-bis: candidate 3 is BUILT, superseding §6d's "not fixed, deliberately" |
| `49a8732dc`, `11117a9a3`, `c92e0a396`, `512d5b60a` | handoff/NOTES/register carrying round 3 + the guard, and the RED-at-HEAD attribution |
| **`d1e51dd38`** | **APPROVED recorded, three of the approval's objections answered, the false "one writer" claim corrected** — first commit legitimately carrying `Council-Reviewed:` |

⚠ **One of this lane's writes is NOT in a commit of its own.** The `WRONG_CALLS.md` entry
about the false "one writer" claim was **swept into another session's commit
(`32c776765`, the 449 lane's)** before mine ran — the shared-tree hazard CLAUDE.md
describes. **Nothing was lost**; the entry is in HEAD, just under someone else's message.
Do not go looking for it under a theme-kits commit.

---

## 6. ⚠ CALIBRATION — read this before trusting anything I wrote

**Seven errors in one session. FOUR were durable claims, and every one of the four was a
right conclusion resting on a wrong reason.** Stated as a list rather than a running count,
because I revised the count three times while writing and that is itself the tell:

| # | the claim | what it actually was |
|---|---|---|
| 1 | `header-theme-chrome` "does not exist in any state" | **a retraction of a TRUE claim** — I filtered `function` and concluded about `name` |
| 2 | `contact-hero` "has ZERO rows" | one row, `name='contact-hero'`, `function='hero-contact'` — found by a round-2 **reviewer** |
| 3 | "3 chrome-eligible rows" | right for **pins**, wrong as a general claim — there are two predicates differing by one clause |
| 4 | "the ONE writer of `sites.style_collection_id`" | **two** writers — found by the council's `prior_art_librarian`, which predicted the error class from my own record |

Plus three mechanical misreads where my *check* was broken rather than my claim: a backtick
in a `-m "…"` executed and ate a word from a commit message; `gofmt -l` chained with `&&`
reads as a permanent failure because it exits 0 either way; and
`verify-head-builds.sh agent-chassis` reported "HEAD does NOT build" when the real fault
was my passing a service name where it wants a Go package path.

The two that should change how you work:

1. **I retracted a TRUE claim by querying the wrong column.** `content_components` has
   **both** `name` and `function`, holding near-identical vocabularies by design. I ran
   `WHERE function LIKE '%theme-chrome%'`, got 0 rows, and published "these components do
   not exist in any state" into the register and a live council submission. They are `name`
   values. **A retraction reads as "someone went and checked", so it outranks the assertion
   it replaces and the next reader stops there.** What caught it was grepping for the
   claim's propagation before warning another lane: 70 files name those components and
   migration 339 carries `RAISE EXCEPTION` drift guards on updating them. **A component
   that does not exist does not need guards against being overwritten.**
2. **This lane wrote the landmine for that exact trap the day before, and its founding case
   is the very pair I then got wrong** (`contact-hero`/`hero-contact`) — which a round-2
   reviewer found in a second claim in the same submission. **The `SessionStart` hook
   matches PATHS and that footprint is a table and a column, so nothing surfaces it.**
   `grep -n 'content_components' LANDMINES.md` is part of opening that table.

**The rule I would want carried forward:** when the conclusion keeps surviving while your
reasons for it keep failing, the conclusion is coming from somewhere other than the
evidence you are citing — go and find where. And **select both columns; never filter on one
and conclude about the other.**

The one thing that went right is worth copying: having found a client's forked header
chrome-eligible and alphabetically ahead of the default, I was one step from filing "every
unpinned site resolves to a client's forked header" as a live fleet defect. **I read
`ResolveChromeComponent` instead of inferring from the rows, and it is already handled and
documented.** Three earlier errors that day came from asserting a mechanism from row data.

---

## 7. OPEN — owner decisions, and what is owed

**Owner decisions** (all three are in `README_where_we_are.md` in plain prose):
1. **`bugs_open/438`: retire or build?** Still open. Both lanes agree the capability does
   not exist and neither will choose. **Note §2: building it would still not put a colour
   on a site.** My recommendation is retire.
2. **Is a kit the right vehicle at all**, given §2? Or is the value entirely in layout
   reachability? **Do not build Phase 2 before this is answered.**
3. **The seam question**, raised by gamedesign.uk and previously filed by nobody: should
   `resolved_composition` — schema-validated, with an enforced lineage enum — describe core
   colours the public never sees? Three options costed in the README; my recommendation is
   to stop recording core colours there and say in the record what it no longer claims.
   **Not filed as an RFC deliberately** — RFC_059 was the structural version of this seam
   and the owner withdrew it, so reopening it as an RFC would relitigate a settled decision.

**Owed by this lane:**
- ~~Read round 4's verdict~~ **DONE — `approved` 16:19Z.** Four rounds on one
  correlation: revise, revise, revise, approved, and **every revise found something real**
  — the missing typography sketch, then the supersede defect, then the fact that I had
  deferred its fix. Three of the approval's own objections were answered on the spot (pod
  verification, the false "one writer" precedent, the unevidenced `locked` claim); the
  rest are in `NOTES_theme_kits.md`.
- ~~**Roll the binary** so the round-4 guard is actually live, then re-probe the pod~~
  **DONE — a roll landed overnight and the re-probe confirms the guard is LIVE
  (2026-09-04, both controls). See §0.** What is still owed on it is a *behavioural*
  observation: adoption is 0, so the guard has never fired and `[VERIFIED 2026-09-04]` no
  `theme_kit_adoption` row carries `design_intent_supersede_risk`. **The first kit ever
  applied is the test** — check that field on the resulting adoption row.
- ~~**The §3 remedy**, as its own council round.~~ **DONE** — built as `b18091066` and
  carried in round 4. What remains is 438's, not this lane's: the ordering itself.
- **A ping to `portfolio_positioning` and `vetcomparison`** with the chrome experiment's
  outcome — **not yet due**: that experiment runs at their remake №5, held behind
  `bugs_open/444`. Their recipe is §5 of
  `docs024_key_docs_latest/portfolio_positioning/RUNBOOK_remake_release.md`. **Their CONTRIB
  is still accurate** — I checked it today against the live data, including its "`site-header`
  has 2 eligible rows, hardcode the resolved UUID" line, which is right.

**Cross-lane state is unchanged** from the 2026-09-02 handoff §6, except that
`bugs_open/445` shipped migration 736 (a 19th layout, `content-hub-tools`) and committed a
fit measurement into `resolve_composition_layout_action.go`. Not a conflict. **Their
DES-086 blind spot still stands and is deliberate: a kit site's layout fit is UNMEASURED,
because my rung returns before their matcher.** They designed their evidence against my
`candidates` being empty, which `28aeb4ca0` now makes true.
