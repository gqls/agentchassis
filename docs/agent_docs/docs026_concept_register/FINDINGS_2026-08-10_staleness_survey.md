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

---

# UPDATE 2026-08-12 — version lag is built and visible, and the premise needed narrowing first

*Appended, not rewritten: the 08-10 survey above stands. This settles the first of its
three open signals and reports what surveying it changed about the plan.*

## The handoff said "make version lag visible — it needs no prose parsing". Half right.

Version lag is cleanly **extractable**. What a citation *means* is not, and that had to be
measured before building anything:

| classifying 315 citations by the words immediately before them | |
|---|---|
| **unclassified** | **244 (77%)** |
| SHIPPED-IN (permanent) | 47 |
| CONFIG-ROW VALUE (expiring) | 12 |
| VERIFIED-ON (expiring) | 9 |
| HISTORICAL/OLD (permanent) | 3 |

Two opposite things wear identical clothes, and no pattern separates them:

- `"deployed in chassis v1.0.1029"` — a permanent historical fact. **Never expires.**
- `"both replicas of v1.0.1218 return X"` — a verification pinned to a version. **Expires.**

So raw lag would have flagged 111 items of which most are permanent facts — the exact
"report nobody reads" failure the 08-10 design conclusion warns about, arrived at from a
different direction. **The survey did not confirm the plan; it resized it.**

## What works instead, and it was already in the register: its own FIELD vocabulary

`status:` and `status-evidence:` are claims about the **current state of the world** by the
register's own convention. `what:`, `why it exists:`, `sources:` are description and history.
Keying on the field needs no language understanding at all:

```
273 citations / 156 entries        → all fields
206 citations / 139 entries        → status: + status-evidence: only   (111 are 50+ behind)
   67 excluded (24%) — the key visibly does work, which is its own control
```

`status-evidence` is markedly staler than `status` — **median lag 103 versus 28**. Entries get
their status line updated; their evidence is not re-verified.

## The one class that CAN be called expired mechanically, and why

A citation quoting a container image tag directly. The reason is a fact about the fleet, not
about the prose: **all 187 live `agent_definitions` rows carry the live tag** (`v1.0.1290` on
2026-08-12, uniformly — control: 187 rows have a tag, so a zero for any other tag is absence,
not a blind predicate). A tag read off a live row therefore dates the *observation* and expires
on the next release. [The uniformity is MEASURED; that the release rewrites them is INFERRED.]

**Two entries corrected from that list, and the pair is the whole lesson:**

- **`SYS-077` — STALE.** Claimed the HITL agent definition "still references image
  `docker.io/aqls/agent-chassis:v1.0.407`", 883 versions behind. The row exists
  (`simple-content-writer-with-approval`, created 2025-11-03, active) and its `image_tag` is
  **`v1.0.1290`**. Corrected in place.
- **`HITL-020` — NOT stale.** Cites the *same* `v1.0.407`, but about what the **seed file**
  says — and `docs/humanintheloop/hitl_agent_definition.sql` still says exactly that,
  unchanged since 2025-11-03. A permanent fact about a repo artefact.

**Same version, same day, opposite verdicts, and only the artefact tells you which.** That is
why `DOC-077` names the class and refuses to judge it, printing the one-line check instead.

**A bonus answer:** `HITL-020`'s own `verify-later` asked "whether these definitions are loaded
in current DB". They are — both halves. ⚠ **And the obvious queries say they are not:** the
group is stored under the display name **"Content Approval with HITL"**, so
`WHERE name = 'content-approval-hitl'` returns 0; and the agent's type contains no "hitl" at
all, so a `%hitl%` search misses it too. This session made that exact mistake first and
reported "0 rows, not in live config" before the `ILIKE` control corrected it.

## Built: `DOC-077` — `scripts/report-register-version-lag.py`

Read-only, ~0.3s, cluster-optional (falls back to `makefile IMAGE_TAG` and **says which source
it used**, because a wrong live version biases every lag the same way and nothing inside the
report could reveal it). **Deliberately not scheduled and not a checker** — it reports "this
entry's evidence has expired", never "this entry is wrong".

```bash
scripts/report-register-version-lag.py               # summary + the 6 image-tag hits
scripts/report-register-version-lag.py --worklist    # + the oldest current-state citations
```

**Two mistakes inside building it, both instructive:**

1. **A proximity window is not an adjacency test.** "`image` within 12 characters of a version"
   read `"inert until an image roll … v1.0.1276"` and `"make deploy-… IMAGE_TAG=v1.0.1190"` as
   live image evidence — **2 of 7 precision**. Stripping the punctuation between the version and
   what precedes it, then requiring an image token immediately before, is **9 of 9**.
2. **The display can convict a correct detector.** The first draft printed each hit's *line
   head*, truncated. A line can carry several versions, so the text shown was often a different
   citation from the one that matched — and **three correct hits read as false positives**. I was
   about to loosen a 9-of-9 detector on the strength of its own misleading output. **Show the
   evidence you tested, windowed on the match.**

## Where this leaves the three signals

| signal | state |
|---|---|
| **version lag** | **DONE — visible, tooled (`DOC-077`), premise narrowed, 2 entries corrected** |
| unresolvable `sources:` citations (96) | open — worth testing whether the same field key applies |
| moved bug references (156) | open, and still ⚠ ONE-DIRECTIONAL (owner ruled 08-06 a fixed bug STAYS in `bugs_open/`, so a non-moved bug proves nothing) |
| features awaiting a non-roll condition (5) | open |

**The transferable question for the remaining two:** does the signal have a key that does not
require reading prose? Version lag only became trustworthy when it stopped trying to understand
sentences and started keying on structure the register already maintains.

---

# UPDATE 2026-08-12b — signals 2 and 3 closed, and the answer to the carried question is NO

*Appended, not rewritten. The 08-10 survey and the 08-12 version-lag update above both
stand. This settles the last two open signals, and the useful result is a **negative** one
about the key that made the first signal work.*

## The carried question, answered

`DOC-077`'s own `verify-later` asked whether the two remaining signals "turn out to have a
comparable field-keyed shape, or whether version lag was the only one with a clean
mechanical key". **Version lag was the only one.**

The field key does not transfer. Unresolved rates by field:

| field | citations | unresolved |
|---|---|---|
| `sources:` | 4,595 | 19% |
| `what:` | 609 | 25% |
| `verify-later:` | 716 | 10% |
| `status-evidence:` | 622 | 23% |
| `relations:` | 544 | 15% |
| `status:` | 102 | 12% |

No break anywhere, and the reason is structural rather than statistical: **a citation's
field says nothing about whether its target moved.** Version lag worked because
`status:`/`status-evidence:` are current-state claims *by convention*, so the field
predicted whether the number could expire. Nothing about `sources:` predicts whether a
file was renamed on 2026-08-04.

**What is total and mechanical here is a different structural key: what git can say about
the cited target** — at HEAD / moved-but-present / deleted / never existed. It reads no
prose, it classifies every citation, and the middle two verdicts **name their own repair**.

**And the field key comes back, as SEVERITY rather than as a filter.** An unresolvable path
means something different depending on where it sits: in `sources:` it is a grounding claim
nobody can open; in `verify-later:` it is a to-do named wrongly, which is mild by design.
That ordering is what the report prints, and it is why the four dead citations below are
listed in that order.

## The measurement — the shape inverts

`scripts/report-register-citation-rot.py` (`DOC-078`), against HEAD `b9b32ba92`:

| verdict | citations | what it means |
|---|---|---|
| `AT-HEAD` + `AT-HEAD-DIR` | **5,883 (75%)** | resolves as written |
| `MOVED-AT-HEAD` | 286 | the file is at HEAD under another path — **printed** |
| `BUG-MOVED` | 316 | signal 3, and ⚠ still ONE-DIRECTIONAL |
| `DELETED` (+2 dirs, +3 moved-then-deleted) | 194 | recoverable through git |
| `MOVED-AMBIGUOUS` | 769 | the file exists; the citation gives only a **bare filename** |
| `UNJUDGED-DIRSHAPE` / `NEVER-UNROOTED` | 39 / 306 | declared unjudgeable, not counted as defects |
| **`NEVER-REPO-PATH`** | **4** | **no file, ever, under that name** |

**Four.** In 7,793 path citations across 1,767 entries, four name a repo-rooted file that
has never existed — and three of those sit in `verify-later:`.

| entry | field | cited | what is actually there |
|---|---|---|---|
| `ADP-018` | `sources:` | `bugs_open/158_HANDOFF_2026-08-01_reply_drop_sizing.md` | `bugs_closed/158_HANDOFF_2026-07-30_eight_reply_sites_still_drop_silently…` — right number, wrong directory, wrong date, wrong slug |
| `VET-006` | `verify-later:` | `platform/orchestration/actions/med_export_json_action.go` | `vet_med_export_action.go` |
| `SYS-004` | `verify-later:` | `platform/orchestration/sweeper.go` | no such file; the sweeper is a CronJob (`bugs-open-staleness-sweep`) |
| `HITL-017` | `verify-later:` | `internal/gateway/hitl_handler.go` | no HITL Go file at HEAD at all |

`ADP-018`'s is the one worth a lane's attention: it is precise-looking and wrong in three
of its four parts, which is exactly the shape `CLAUDE.md` warns about when it says a bare
bug number is ambiguous and you must resolve by slug.

**So the 96 was not wrong — it answered a different question.** The 08-10 figure asked
"does this path exist at HEAD?", which is the right question if you want to know what a
reader can click. Asking instead "can git still find this file?" moves almost all of it:
the unresolved mass is (a) a house style of abbreviating citations to a suffix or a bare
filename, and (b) the numbered-docs tree moved on 2026-08-04, one rename from resolving.
**The register's citations are not rotting. They are abbreviated, and one directory move
made a lot of the abbreviations stop working at once.**

## Three missteps, and the first two are the same mistake

### 1. My own normalisation manufactured the headline

The `(N)` suffix in citations is an **extraction-unit id** in some places
(`PLAN_tool_widget_clobber(9).md` — the file is `PLAN_tool_widget_clobber.md`) and
**genuinely part of the filename** in others (`002e_concept_spark(6).md` and
`016b_debugging_guide_merged(3).md` are what those files are called on disk). I stripped it
unconditionally. That converted correct citations into unresolvable ones and produced
**27 of 34 "never existed" findings**, which arrived sorted by frequency and therefore
*led the draft*:

> ~~15 entries cite `docs/016b_debugging_guide_merged.md` — the estate's most-read
> document — at a path that has never existed.~~ **FALSE.** They cite
> `docs/016b_debugging_guide_merged(3).md`, git has exactly that path, and it was moved on
> 2026-08-04. The citation was right when written.

It was the most alarming number in the report and it was an artefact of the instrument. I
had already stated it in chat before the check caught it.

**What caught it:** looking up what the file is actually called before writing down the
repair. **The cheap check that would have:** never report "no file, ever, under that name"
without printing the near-miss git *does* have — an absence claim with no near-miss shown
is unfalsifiable from the output.

### 2. Then the fix had the same shape one level down

Resolving as-cited *and* stripped, I kept the **last** variant's verdict instead of the
**best** one. So a citation that resolved exactly as written was overwritten by the
stripped form's failure, and the same 15 entries read as "never existed" for a second time,
now with the correct logic underneath. Two different bugs, one wrong number, and the number
never moved between them — which is precisely why a stable figure is not a corroborated one.

### 3. `git rev-list --objects --all` cannot enumerate paths

The first control I wrote — "every path at HEAD must appear in the set of paths that ever
existed" — **failed**, with 791 of 9,301 HEAD paths missing. The cause is not history: the
command dedups **by object**, so content-identical files share one blob and only one of
their paths is ever printed. **791 of 791 of the missing were content-identical duplicates.**
A path-existence check built on it reports live files as never having existed. Use
`git log --all --no-renames --pretty=format: --name-only`. In `LANDMINES.md`, because
nothing about the wrong output looks wrong.

**All three were caught by controls, not by reading the output** — and the two that
mattered were caught by a control I only wrote because this lane's own doctrine says a
clean result means nothing unless a dirty one was reachable.

## Where this leaves the survey

| signal | state |
|---|---|
| status claims built to expire | settled 2026-08-10 via `BLD-019` build provenance |
| version lag | **CLOSED 2026-08-12** — `DOC-077` |
| unresolvable `sources:` citations | **CLOSED 2026-08-12b** — `DOC-078`; 4 dead, the rest abbreviation and one tree move |
| moved bug references | **CLOSED 2026-08-12b** — `DOC-078` measures it (316 citations); ⚠ still one-directional, and the report says so every run |
| features awaiting a non-roll condition | 5 — unchanged (`CQ-019`, `PLAN-047`, `PBP-025`, `TL-038`, `TL-040`) |

**What is deliberately NOT done:** none of the 801 repairable citations were rewritten. An
automated 801-line rewrite across 111 files is the change no reviewer can check, and the
value of doing it by hand is low — the citations were correct when written and the reader
can find the file. Whether anyone repairs them at all is `DOC-078`'s own `verify-later`,
and if the answer in a month is "nobody did", the real fix is a `sources:` convention at
authoring time (cite a path, not a bare filename), which is the same conclusion the sha
rule reached: **put the check where the error is made.**
