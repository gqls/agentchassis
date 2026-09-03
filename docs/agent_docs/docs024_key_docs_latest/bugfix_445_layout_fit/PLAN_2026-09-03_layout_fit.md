# PLAN — bugs_open/445, layout fit

**Created 2026-09-03, LATE — after Phases 0 and 1 had already shipped.** CLAUDE.md requires the
standing five at the START of a workstream; this lane ran a day on a plan file outside the repo
(`/home/ant/.claude/plans/snazzy-spinning-allen.md`) because a chassis build was imminent and
committing the Go change before it took priority. Recording the deviation rather than
back-dating it.

## The problem, in one paragraph

445 was filed as "the library has no archetype for content-hub-with-tools". True, and a symptom.
The mechanism is that **the estate could not see a library gap of this shape by construction**:
the growth signal fired only when the *total* score was zero library-wide, and the category /
description / same-scheme bonuses are added independently of tag matching, so a layout matching
none of a site's tags still scored above zero. Full evidence in `bugs_open/445` §8.

## Phasing, and why in this order

| phase | what | state |
|---|---|---|
| 0 | Classifier prompt: stop promising a mechanism that does not exist; stop offering layout names as tag examples; prefer FORM over industry words | **DONE** — migration 735, applied + verified at the live row |
| 1 | Measure fit, record it, fire the signal on it | **DONE** — commit `76db94fc7`, inert until the next chassis roll |
| 2 | `internal/cronchecks` — the shared harness `RFC_024` asks for | not started |
| 3 | `cmd/layout-fit-check` — fleet sweep, clusters not sites | not started |
| 4 | Reachability guard for new layouts | not started |
| 5 | The archetype itself (owner ruling: this thread's) | not started |

**Decisions taken (owner, 2026-09-03):** fix forward only — no re-composition of the seven live
`magazine-grid` sites; build `internal/cronchecks` first rather than adding a tenth un-harnessed
cron check; fix the prompt now rather than sequencing it behind the roll.

**Deviation from the approved plan, deliberate:** the plan had the `layoutmatch` package
extraction landing *first*, with Phase 1 on top of it. A chassis build was announced mid-session
and `make build-*` builds from committed HEAD, so Phase 1 shipped **in place** and the
extraction is deferred to Phase 2/3, which are the only things that need it. Moving a package
under a deadline on a tree this many sessions share risks breaking HEAD for everyone — which
happened to another lane this week.

## Design decisions and their reasons

**Coverage is IDF-weighted, not a plain term count.** Two design passes disagreed (0.25
unweighted vs 0.50 weighted). Weighted wins because it is the same quantity the matcher already
scores with; an unweighted count makes a site's junk tags weigh exactly as much as its
distinguishing ones. The denominator deliberately includes terms **no** layout carries — a term
the library cannot serve is precisely what must count against the fit, not be quietly dropped.

**The threshold is 0.50, and it was chosen not invented.** The measured coverage distribution
has two empty intervals: (10%, 15%) and (38%, 62%). One tag on a ten-term site moves coverage
~8-10 points, so a cut in the narrow band flips on ordinary classifier variance; only the wide
band is stable. 0.50 is also the figure migration 103's own worked example names.
**Pre-registered disconfirmation** is in `bugs_open/445` §8e — if compositions now land inside
the 38-62% band, the cut was a 33-site artefact.

**Selection is untouched.** No site's layout changes; only what is recorded and whether a parked
review item is filed. This is the whole risk argument for shipping it in one roll, and
`TestOneAttractorTagIsAWeakFit` asserts it (magazine-grid still wins for designblog's shape).

**The unit of a finding should be a CLUSTER, not a site** (Phase 3). Seven sites leaning on one
attractor tag is one design task; seven per-site items are a queue nobody drains.

## Corrections to the originating brief

- **445 §6 candidate 1 ("add an archetype") is necessary and NOT sufficient.** Simulated: the
  archetype breaks the 7-site cluster but leaves designblog.co.uk and apis.uk still winning on a
  single tag at 6-8%. §8g.
- **445 §2's rule-outs both survive**, and I extended the first: the description path §2 did not
  check is also zero.
- **445 §4's concentration is not "what the fleet builds"** — it is four attractor strings.
