# HANDOFF — theme kits lane, 2026-09-03 (session "theme kits")

**Supersedes `HANDOFF_2026-09-02_continue_here.md`.** That file is still accurate about
what was BUILT and is the fuller account of the eight pre-apply defects and the owner's
ruling; read it second. Read this first for state, and read §1 before believing anything.

**The lane now has its standing five** (`PLAN_2026-09-02_theme_kits.md`,
`RUNBOOK_theme_kits.md`, `NOTES_theme_kits.md`, `README_where_we_are.md`,
`SUMMARY_2026-09-03_theme_kits.md`). The design is migrated out of the approved plan file
at `/home/ant/.claude/plans/please-think-hard-about-starry-locket.md`, which remains the
source for corrections C1–C10.

---

## 1. STATE — three independent facts, all verified today

| fact | value | how |
|---|---|---|
| binary | **LIVE**, `agent-chassis` `v1.0.1355` | `/proc/1/exe` capability probe, positive AND negative control |
| schema | **APPLIED** 2026-09-02, migrations 689 + 691 | `to_regclass` both tables; 4 kits, 14 fleet archetypes |
| adoption | **0** | `SELECT count(*) FROM site_specs WHERE aspect='theme_kit_adoption' AND is_current` |

**Nothing has adopted a kit.** Every kit-conditional branch is live, reachable, and has
never run. **Cite this lane as "built and reachable", never as "working".**

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

## 3. ⚠ THE DEFECT TO FIX FIRST — a kit applied before classification loses palette AND typography, silently

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

**My recommended remedy, and why it is not done:** refuse or warn when the site has no
classifier-written `design_intent` yet. It changes no cascade and converts a silent loss
into a reported one. It is still a behaviour change to a live shared action, so it owes its
own council round rather than a same-day commit. The other two candidates are
architecture-scope (make the classifier respect `locked` — it changes the classifier's
write authority over a shared aspect) or build on 438's own defect (write
`mission.preferred_typography`, which survives that path only by accident).

---

## 4. Council gate — a resubmit is IN FLIGHT, do not claim approval

**Trail correlation `bed139b2-f512-436a-9ba8-ff2fbfade8ef`** (use this — it is the key the
artefacts are written under).

| round | verdict | what it found |
|---|---|---|
| 1 | `revise` 2026-09-02 21:43Z | the rationale claimed a typography guard the sketch never showed. Correct: a reviewer judges the submission, not the repository. |
| 2 | `revise` 2026-09-03 15:32Z | **the §3 defect.** The best output of the whole review. |
| 3 | **UNREAD — in flight** | accepts the objection, withdraws a false retraction, corrects a second evidence claim. Run correlation `8e6f2aa8-ceae-4d22-a543-a47196f57193`. |

Every commit carries `Council-Submitted:`, which asserts nothing and is credited
automatically if the correlation approves. **Do NOT write `Council-Reviewed:` until you
have read an approved verdict** — 098 buckets that as MISMATCH. Resolve with the queries in
the RUNBOOK §4.

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

---

## 6. ⚠ CALIBRATION — read this before trusting anything I wrote

**Five errors today, and every one was a right conclusion resting on a wrong reason.** The
full list is in `NOTES_theme_kits.md` and `WRONG_CALLS.md`. The two that should change how
you work:

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
- **Read round 3's verdict** and act on it. Nothing may be recorded as approved until then.
- **The §3 remedy**, as its own council round.
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
