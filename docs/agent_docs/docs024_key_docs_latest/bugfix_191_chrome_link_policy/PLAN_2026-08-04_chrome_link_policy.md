# PLAN — bugs_open/191: one chrome link-target policy

**Started 2026-08-04.** Single-session lane, picked up because `191` was filed
**UNOWNED** and `scripts/who-owns.py 191` identified no owning workstream.

## The problem, in one sentence

`RenderSiteComponentsAction` builds a header's nav items and its CTA button in one
run, from one `pages` table, and validates them with **two different eligibility
predicates** — so chrome ships a 404 button that the nav rendered beside it had
already filtered out.

## Why this bug and not another (the selection, recorded because it is reusable)

55 files in `bugs_open/`. The fleet runs ~30 concurrent sessions, so "not being
worked on" is the binding constraint, not "interesting". Method:

1. Tally `bugs_open/NNN` mentions across every session transcript modified in the
   last 5–10 hours (`~/.claude/projects/*/*.jsonl`). High counts (098 at 14, 153 at
   10, 192 at 9) are contended; single mentions usually are not.
2. Cross-check the low-contention candidates with `scripts/who-owns.py`.
3. **Then check the FILE, not the bug** — several uncontended bugs live in files
   half the fleet is editing. `render_site_components_action.go` showed 7 sessions;
   `run_checks_action.go` showed 2 with 29 and 15 hits.

**Decision, and it was a close one:** 191 was taken *despite* its file being hot,
because (a) the file was **clean in `git status`** at the time — nobody had
uncommitted edits in it — and (b) the change is small and additive in that file
(one helper call swapped at two lines), so the same-file passenger window is
minutes, not hours. The alternative candidates (188 `run_checks_action.go`,
194 `save_page_sections_action.go`) had *both* a contended bug and a contended file.

## The design, and the two decisions inside it

Fix candidate 2 from the bug file — *"one named predicate, one resolver"*, the shape
`bugs_closed/118` applied to component eligibility, applied here to **link targets**.

### Decision 1 — what actually gets extracted

The obvious reading is "give the CTA the nav's predicate". That is not enough, and
the reason is the structural half of the defect: **the escapes were inline.**
`applyNavVisibility` disables the deployment filter when the lookup errors and when
the site has zero deployed pages, and both branches sat inside the nav function, so
no other caller could reach the policy they encode. The CTA's author reached for the
nearest other helper because there was nothing else to reach for.

So the unit extracted is not the predicate (that already had one definition,
`datahelpers.NeverDeployedPagePredicate`, and this change adds no new spelling of
it). It is **the decision**: set + escapes + the per-URL test, as
`ChromeLinkPolicy`.

### Decision 2 — what the CTA does on a first build, and this CHANGES a behaviour

The bug file's own candidate 1 says the CTA should render **no button** when the
site has zero deployed pages, calling that "the correct outcome". **We did the
opposite**, deliberately.

The nav's reason for going unfiltered on a first build is written at
`nav_tables.go:194-202`: chrome is **idempotence-gated**, so what the first build
writes may never be re-rendered, and a nav emptied then would persist indefinitely.
That argument applies to the button *identically*. A buttonless header on a site
about to deploy 25 planned pages is permanently wrong; a planned-target button is
wrong only for the window in which the platform has already ruled the fully
unfiltered **nav** acceptable.

Answering the freeze one way for the list and the other way for the button beside it
would be a miniature of this very bug, inside its own fix.

The same argument carries the error case, which is a **real behaviour change**: the
CTA previously vanished on a lookup error. That was never a considered policy — it
was a side effect of `loadResolverPageSet` returning an empty set on error. It is
now unfiltered, like the nav. This is named explicitly in the council submission's
`risks` block, because it is the one thing in this change a reviewer should be
invited to disagree with.

> **The measurement that could have overturned Decision 2, and its result.**
> The plan predicted the escape would be taken only by genuinely young sites. The
> first query said **19 of 38 sites have zero shipped pages** — which looks exactly
> like the disconfirming answer. It was the wrong question. Split by whether the
> site has any pages at all: **18 have no pages whatsoever** (chrome renders nothing
> either way), **19 are strict**, and exactly **1** (`webdesign.uk`) has pages with
> none shipped and therefore actually takes the escape. See `NOTES` — this is
> logged in `WRONG_CALLS.md` as a near-miss, because the unsplit figure was one
> sentence away from being written down as fact.

### What was deliberately NOT changed

`loadResolverPageSet` keeps its loose predicate for its two **page-content**
callers. The two have different repair economics: a content CTA is re-resolved on
every render and rerender, and a batch build legitimately links siblings that deploy
minutes later, so a not-yet-shipped target self-corrects. Tightening it would strip
CTAs from most of every fresh build and would redefine the `unresolved_cta` HITL
signal (LNK-012) — a separate, fleet-wide, measured change, not a passenger on a
scope fix.

What the change does instead is convert the remaining difference **from drift into a
recorded decision**: a doc comment on the function saying which set is which and
why, plus a source-scanning allow-list naming each entitled caller *with its reason*.

## Phasing

1. Validate the bug still reproduces (code + DB + wire). — done
2. `fable` drafts the plan against the real code. — done
3. Run fable's own blast-radius commands rather than trusting its counts. — done
4. Implement; prove every guard by **mutation**, not by green ticks. — done
5. Register the seam (LNK-030) + LANDMINES, in the same commit as the code.
6. Council gate, commit with `Council-Submitted`, close the bug.
7. **Owed after the next roll:** pod-grep both replicas, re-run `nav-updater` on
   `mortgagecalculator.co.uk`, re-run the corrected SQL, curl the survivors.
