# RFC 007 — chrome eligibility needs a package both sides can import, and the guard-scan count is the meter

**Filed:** 2026-08-01 by the `bugfix_170_chrome_pin_eligibility` lane.
**Status:** open, not urgent. Raised because **five seats converged on it in one
council round** (`21bac2a2-2b46-4883-894f-19d7ec5e5b45`, APPROVED — all seven
objections advisory, none high), and the architecture seat asked for it by name.

## The ask, in the architecture seat's own words

> Recommend: after this ships, an RFC to move `chromeComponentLevels` + the two SQL
> predicates into a small package importable from both `actions` and
> `discovery_checks`, retiring the lockstep test entirely rather than hardening it
> again next time.

## Why it is worth a paper rather than a shrug

The round approved the fix and then, from six different angles, said the same thing
about the *shape* of it. That convergence is the signal — not any one objection:

| seat | what it said |
|---|---|
| `architecture` (medium) | fourth hand-maintained guard over one vocabulary; the plan "documents the root cause — `discovery_checks` cannot import `actions` — but does not propose fixing the import direction, only adding another test to detect drift from it" |
| `reuse_agent` (medium) | a landmine already says chrome selection has TWO guard scans with disjoint blind spots "and that the fix is to unify them, not multiply them" |
| `editquality` (medium) | a third regex scan "compounds that exact problem" |
| `debug_historian` (low) | "worth the reviewer/owner being told this is now a THIRD independent scan set over chrome eligibility, not a second" |
| `prior_art_librarian` (medium) | the plan does not show it read that landmine before adding a third and fourth scan |
| `guardian` (low) | the lockstep "is a workaround for a dependency-direction problem, not a fix to it" |

**They are right, and the lane accepts it.** The 170 fix needed a predicate in
`discovery_checks` and could not import the one in `actions`, because `actions`
imports `discovery_checks` to register the checks — the dependency only runs one
way. So the predicate was hand-typed and a lockstep test added to stop it drifting.
That is a correct local decision and a bad global trend.

## The meter, stated precisely

The thing to notice is not "there are several test files". It is that
**each new consumer of chrome eligibility now costs one more hand-maintained
regex guard instead of one import**:

| date | consumer added | guard it cost |
|---|---|---|
| 2026-07-31 (`118`) | assignment call sites | `TestNoChromeSelectionHandTypesItsOwnLookup` — scans SQL, **skips `component_library.go`** |
| 2026-07-31 (`167`) | build path | `TestNoBuildPathResolvesChromeByPlainFunctionLookup` — scans Go calls, **covers** `component_library.go` (added precisely because the first was blind there) |
| 2026-08-01 (`170`) | pin, at three consumers | `TestNoConsumerDereferencesAChromePinUnguarded` — scans Go **and** SQL |
| 2026-08-01 (`170`) | detector, cross-package | `TestChromePinPredicateMatchesTheActionsPackage` — parses the **other package's source text** |

The second exists *because* the first was blind. The fourth exists *only* because of
the import direction. And this lane's own NOTES records that its first versions of
both new guards **passed for the wrong reason** — one could not distinguish a correct
fix from an unguarded read, the other matched `'footer'` where the file used it as a
slot name. Both were caught and fixed, but `debug_historian` put the right reading on
it: that is evidence the class is fragile, not evidence it is now solid.

## What is actually proposed

A small package — say `platform/orchestration/chrome` — holding **only**:

- `ComponentLevels` (today's `chromeComponentLevels`)
- `EligibleSQL(alias)` — the POOL predicate
- `PinEligibleSQL(alias)` — the PIN predicate, and the comment explaining why it
  omits `forked_from IS NULL`

Imported by `actions` and `discovery_checks` alike. It has no dependencies of its own
(three string constants and two functions), so it cannot recreate a cycle.

**What that buys, concretely:** `check_chrome_pin_lockstep_test.go` is **deleted**,
not hardened — there is nothing left to drift. The remaining scans keep their jobs
(they guard *call shape*, which a shared constant cannot), but they stop being the
only thing standing between two copies of one rule.

## What this RFC does NOT claim

- **Not urgent.** Nothing is broken. The lockstep is mutation-proven in both
  directions and green. This is debt with a visible meter, not a defect.
- **It does not collapse the three call-shape scans into one**, and that should be
  considered separately. They have genuinely different jobs and the 167 lane's note
  is explicit that merging them would cost the 118 scan the `component_library.go`
  exemption that makes it precise. **"Unify the scans" and "share the predicate" are
  two proposals; only the second is uncontroversial**, which is why this RFC asks
  only for the second.
- **[UNMEASURED]** whether any other package hand-copies a chrome predicate. The
  lane's scan covers `actions` and its `discovery_checks` subpackage only. Worth one
  fleet grep before acting: `grep -rn "component_level IN" --include=*.go .`

## Related

- Council `21bac2a2-2b46-4883-894f-19d7ec5e5b45` (this round, APPROVED, the full
  objection text quoted above)
- `bugs_open/170` + `docs024_key_docs_latest/bugfix_170_chrome_pin_eligibility/`
- Concept register **CLC-013** (the seam, all four extensions)
- `LANDMINES.md` — the entry the seats cite ("chrome selection now has TWO guard
  scans with disjoint blind spots and a shared vocabulary"), and the 170 entry
  ("a `style_collections` chrome PIN is not a per-site assignment")
- `bugs_closed/118`, `bugs_closed/166`, `bugs_closed/167`
