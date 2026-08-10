# FINDINGS — is the register still TRUE? The staleness survey, 2026-08-10

*Handoff item 3. The handoff calls staleness "a design question, not a chore" and
says it deserves its own session. It does — so this is deliberately **the survey
and not the mechanism**, because this lane's own doctrine is that the survey
usually resizes the feature, and here it did: the obvious design (parse the
`status:` field) is the one the measurement rules out.*

**What this answers:** how stale is the register, measured rather than asserted.
**What it does not:** it does not re-verify 1,818 entries. Every signal below is a
claim an entry makes that **the repo or the cluster can contradict on its own** —
none of them read the prose for truth.

---

## The four signals, and what each is worth

### 1. Status claims that were built to expire — the real class, and the smallest

**44** of 1,818 entries carry a status matching "not live" language. Of those,
**6 have already been corrected in place** (`~~built, inert~~ → LIVE`) — which is
the convention working, and had to be separated out before any count meant
anything.

Of the **38** remaining, all read by hand (see §"why by hand" below):

| what they actually are | count | worth |
|---|---|---|
| **"inert until the next chassis roll", and the roll has happened** | **~20** | **the finding** |
| conditional on something OTHER than a roll (a migration/seed naming it, a fence installing it) | 5 | expires on a different clock; needs its own check |
| written 2026-08-10, 0–3 rolls ago | 6 | still current |
| false positives of the regex — currently accurate | 6 | see below; these are why the regex cannot be trusted |

**PROVEN AT THE ARTEFACT, not inferred.** Two of the oldest, pod-grepped on **both**
replicas of the running chassis `v1.0.1280`, each with a nonsense-symbol negative
control returning 0 in the same exec:

- **`FIX-055`** — status said *"built, shipped to HEAD, **NOT yet live**"*. Both
  replicas carry `hasGatingObjection` (1) and `gatesOnlyBecauseTruncated` (2).
  **The status was false, and had been for some part of 13 days and 22 rolls.**
  Corrected in place.
- **`SCR-002`** — status said *"built, inert until the chassis image rolls"*. The
  image rolled ~23 rolls ago; `fetch_provenance.go` is in both replicas. But its
  own evidence line says vet collection has been off since 2026-03-18, so it is
  **live and unexercised** — a different claim, on a different clock. Corrected,
  and the two claims separated. **The old wording conflated them, and that
  conflation is itself a design input**: "inert until an image rolls" expires by
  itself; "nothing drives it" does not.

The remaining ~18 are listed in §"the worklist" and are **not** corrected here,
because correcting one honestly costs a pod-grep against a symbol that lane chose,
and some cannot be grepped at all (`DOC-073`'s case: a control-flow change with no
new string literal has no marker — `WFA-012` says so itself). **Writing "live" on
an entry I had not proved would manufacture exactly the false confidence this
survey exists to measure.**

### 2. Version citations — the cleanest mechanical signal in the register

**129** entries cite a chassis version. The fleet is on **v1.0.1280**.

| lag | entries |
|---|---|
| 0–4 versions behind (current) | 9 |
| 5–19 behind | 15 |
| 20–49 behind | 25 |
| **50+ behind** | **80** |

The extremes: `SYS-077` and `HITL-020` cite **v1.0.407** — **873 versions ago**.
`ADP-015` cites v1.0.424, `RES-002` v1.0.575.

This does not say those entries are wrong. It says **their evidence is from a
platform that has been rebuilt 873 times since**, and a reader — or a council seat
quoting the entry — has no way to see that from the entry itself. It is
zero-ambiguity to compute and needs no prose parsing at all.

### 3. Citations that no longer resolve

**96 of 2,611** judgeable path citations in `sources:` lines (**3.7%**) point at
files that do not exist at HEAD. Where they point: `WM/` (21), `docs019` (21),
`docs` (15), `bugs_open` (12), `ED/` (6) — i.e. overwhelmingly the numbered-docs
tree deleted on 2026-08-04, which the handoff already records as 43 known lines
resolving only through git. The true figure is larger than the 43 that were known.

> **A caution about this number, because the first version of it was wrong.** A
> naïve path regex returned **187**. Sampling 22 showed a large minority were
> **artefacts of my own extraction**, not defects in the register: tails of brace
> notation (`{PLAN,NOTES}_x.md` → `_x.md`, `check_site_unreachable{,_test}.go` →
> `_test.go`) and abbreviated citations containing a literal ellipsis
> (`docs021.../025_….md`). 92 tokens excluded on those grounds. **The 187 was
> publishable-looking and wrong**; it is recorded here because the next person to
> write this check will hit the same thing.

### 4. Bug references that have moved — a one-directional signal

**156** entries cite `bugs_open/NNN` for a bug now filed under `bugs_closed/`. The
citation misses, and it usually means the entry's premise moved.

⚠ **One-directional on purpose.** The owner ruled on 2026-08-06 that **a finished
bug stays in `bugs_open/`**, so a bug that has *not* moved proves nothing at all.
This signal can only ever find a subset, and a checker built on it must say so or
it will read as a clean bill of health.

### Not a signal: `verify-later:`

**1,754 of 1,818** entries carry one. It is a template field, so its presence
measures nothing — counting it would produce a large, alarming, meaningless
number. Recorded here so nobody reaches for it as a staleness metric. *Whether any
was ever answered* is a real question and is not answered here.

---

## Why the statuses had to be read by hand — and what that rules out

The obvious mechanism is to parse `status:`. **The measurement says do not.**

The status vocabulary is bimodal: **1,635 of 1,818** entries use one of seven bare
words (`deployed` 867, `aspirational` 289, `partial` 258, `superseded` 85,
`abandoned` 67, `convention` 49, `unknown` 20), and the rest are free prose that
encodes history, corrections, partial states and ordered preconditions.

My regex said **38**. Reading them said **~20**. The overcount was not sloppiness in
the pattern — it was the field doing things no pattern can classify:

- **`WFA-006`: "runtime-inert BY DESIGN".** A permanent property that reads exactly
  like an expiring one. A checker that flags this is wrong for ever.
- **`VONC-011`: "deployed — UPDATED 2026-08-02, was `built, not live`".** The stale
  claim is quoted *inside* the correction. Same shape as the frozen-log trap the
  drift check already had to be head-bounded for — **a watcher crying wolf about
  its own archive**, one level down.
- **`CLC-013`, `STY-056`, `WFA-009`, `CGV-031`: half live, half not.** One entry,
  two statuses, on two clocks. There is no single answer to grade.
- **`PBP-037`: "INERT end to end until three things happen IN ORDER".** A
  precondition chain, not a state.

**Design conclusion, evidence-backed:** a staleness checker must key on things
with no prose ambiguity — **a version number, a file path, a bug id, a date** —
and must report **"this entry's evidence has expired"**, never **"this entry is
wrong"**. That is the same bar the drift check already holds itself to ("nothing
here is a claim that an entry is WRONG"), and it is the reason that check is
trusted enough to be read.

The one exception worth building anyway is §1, and it does **not** need the prose:
pair the **entry's own commit date** (git has it) against the **roll clock**
(`IMAGE_TAG` bump commits in the makefile — 107 in the last 14 days). An entry
whose status contains a roll-conditional phrase and predates N rolls is a
**candidate**, handed to a human or to its lane with the pod-grep already
suggested. Candidate, not verdict.

---

## The worklist — entries whose stated condition has been met

Not corrected here, deliberately. Each needs one pod-grep against a symbol its own
lane chose. Sorted oldest first; "rolls" counts `IMAGE_TAG` bumps since the entry
was written.

| entry | file | written | rolls since |
|---|---|---|---|
| `CLM-016` | claims-verification.md | 2026-07-28 | 22 |
| `CLM-017` | claims-verification.md | 2026-07-29 | 22 |
| `CLC-013` (pin half) | component-lifecycle.md | 2026-07-31 | 19 |
| `TL-038` | tool-lifecycle.md | 2026-07-31 | 18 |
| `WFA-005` | workflow-authoring.md | 2026-08-01 | 17 |
| `VONC-011` (per-category half) | vonc.md | 2026-08-05 | ~7 |
| `CLM-019` | claims-verification.md | 2026-08-03 | 15 |
| `PBP-029` | page-build-pipeline.md | 2026-08-04 | 10 |
| `WFA-009` (Go half) | workflow-authoring.md | 2026-08-04 | 9 |
| `RSH-005` | resilience-self-heal.md | 2026-08-04 | 8 |
| `TL-040` | tool-lifecycle.md | 2026-08-05 | 7 |
| `PBP-037` | page-build-pipeline.md | 2026-08-06 | 6 |
| `CQ-020` | content-quality.md | 2026-08-08 | 5 |
| `WII-012` | work-item-integrity.md | 2026-08-08 | 5 |
| `CGV-031` (guards half) | content-governance.md | 2026-08-08 | 5 |
| `CLM-020` | claims-verification.md | 2026-08-08 | 5 |
| `CQ-021` | content-quality.md | 2026-08-09 | 4 |
| `WFA-012` | workflow-authoring.md | 2026-08-09 | 4 |
| `IMG-069` | imagery.md | 2026-08-09 | 4 |
| `STY-056` (Go half) | styling-render-pipeline.md | 2026-08-09 | 4 |

Separate clock, not a roll — do not lump these in: `CQ-019` (awaits migration 303),
`PLAN-047` (awaits seed 306), `PBP-025` (awaits a `run_checks` array naming it),
`TL-038` (roll **and** a criteria fence), `TL-040` (a live fence).

⚠ **`WFA-012` cannot be settled by pod-grep at all** — its change is control flow
with no new string literal, and `ExtractNestedField` predates it and greps 8 times
either way. That is `DOC-073`'s case: **a positive control that cannot fail**. Its
lane owns the proof.

---

## Headline

The register is **complete and self-consistent, and its evidence is ageing
unevenly**. The two halves agree exactly (1,818/1,818 as of this survey). But 80
entries cite a platform version 50+ rebuilds old, ~20 assert "not live" about code
that has since shipped, and 96 citations lead nowhere.

**None of that was visible from inside the register**, and that is the actual
finding: every mechanism this lane has built so far — coverage, drift, and now the
authoring gate — asks whether the register agrees **with itself**. Nothing has ever
asked whether it agrees with **the platform**, and the survey took an afternoon.

The cheapest first move is not a checker at all. It is to make the **version lag**
computable and visible — 129 entries already carry the number, and the fleet's
current version is one `kubectl` call.

---

## ADDENDUM, same evening — the worklist is settled, and by a mechanism that did not exist when the survey was written

**A fresh chassis rolled: `v1.0.1283`. `BLD-019`'s build provenance shipped with
it, and it changes how this whole class is verified.**

The running binary now carries the commit it was built from. Read back from
**both** replicas:

```
strings /app/agent-chassis | grep -oE '^[0-9a-f]{40}(-tree)?$'
  -> d3c09cc746e563b6339831cfb69576eb52135c43     (identical on both; no -tree suffix, so a clean committed build)
```

That retires the per-entry pod-grep this document called for four hours ago. The
question "is this entry's code live?" is now one exact command:

```bash
git merge-base --is-ancestor <the entry's own commit> d3c09cc746e563b6339831cfb69576eb52135c43
```

**Controlled before use, because an ancestry test that always says yes says
nothing.** Positive: `FIX-055`'s `3a59b5012` → IN, agreeing with the pod-grep that
proved it independently this afternoon. Negative: `3ac87646a`, an off-branch merge
commit → NOT IN. The test can return false.

**Result: every commit cited by a worklist entry is IN the image.** The build was
made from *exactly current HEAD*, so the roll-conditional half of §1 is settled
wholesale rather than one entry at a time. **19 entries annotated in place** with a
dated correction that states precisely what is proven — the Go code is in the
running binary — and explicitly declines to claim the feature is *exercised*,
which is a separate condition on a separate clock (`CQ-019` awaits migration 303,
`PLAN-047` seed 306, `PBP-025` a `run_checks` array, `TL-038`/`TL-040` a live fence).

### The new finding, and it is an authoring rule

**13 of the 29 entries examined cite NO commit sha at all.** For those, provenance
can only infer inclusion from the entry's date — sound, but not verifiable, and it
degrades to exactly the guesswork the stamp exists to remove. Their annotation says
so rather than hiding it.

> **So: an entry whose status is conditional on a roll must NAME ITS COMMIT.**
> It costs nine characters at authoring time and converts an unfalsifiable status
> into a one-command check for ever after. This is the cheapest thing in this
> document and probably the most valuable — it is a candidate for the authoring
> gate (OPP-006) rather than for a watcher, on the same argument: put the check
> where the error is made.

### What this does NOT settle

`WFA-012` remains unsettleable by pod-grep (control flow, no new string literal)
— but it cites two commits, both IN, so **provenance settles what grepping could
not.** That is the clearest single demonstration of why `BLD-019` matters:
`DOC-073`'s positive-control-that-cannot-fail is a dead end for marker hunting and
a non-problem for provenance.

The other three signals — version lag (80 entries 50+ behind), unresolvable
citations (96), moved bug references (156) — are **untouched by this roll** and
remain the open work.
