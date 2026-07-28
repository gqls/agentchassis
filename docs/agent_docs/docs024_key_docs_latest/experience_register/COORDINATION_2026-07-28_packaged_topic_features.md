# Coordination — `features_open/001` packaged topic features ↔ the experience register

**Written 2026-07-28** at the owner's prompt, after reading the oufe session
(`b026618c-a3c6-42eb-bab5-ee0d8175b8ae`) and `features_open/001_FEATURE_packaged_topic_features.md`.
**Advisory. This workstream does not own 001** — it belongs with news-feed-pooling / oufe. This
file exists so neither side builds the other's half twice.

## The short answer: related, in two specific places, and NOT the same thing

The owner's closing message in the oufe session asks for:

> "extract from the content … pertinent points and then extend the article with maybe different
> pages for each major point … branch/link off to a deeper exploration that may have historical
> data graphs and tools and factual commentary that may be updated regularly … a sort of deepthink
> workflow that works with the **experience workflows** and the tool checking workflows and tool
> builder workflows and graphing tools"

He names the connection himself. It is real, but it is narrower and more useful than "they are both
library-plus-fork".

## 1. What he described is a level-3 micro-journey we have already scoped

A main article → its pertinent points → a page per point → each with tools and graphs → linking
back and sideways **is** hub-and-spoke, and `hub-spoke-index` is already listed in
`design/taxonomy_seed.md` §"Level 3 — micro-journey candidates (HARVEST-PENDING)":

| candidate | shape | where observed |
|---|---|---|
| `hub-spoke-index` | section-index page → navigate child pages → navigate back/sideways | live `parent_section` hubs |

So the *behaviour* half of what 001 needs is a register entry that does not exist yet but is on the
list. The tool half is **already harvested**: `MJ-002 timed-remote-challenge-loop` is a real
turn-based tool session against a live backend (the vonc gauntlet), including the clause that the
clock and progress markers advance **only** on a 200 — which is exactly the honesty rule a
"what happens to stakeholder income if I change this input" tool needs.

**What the register contributes:** the dossier's cross-links become *declared* rather than
incidental. A hub that promises spokes of a named role, checked. That matters here more than
usually, because a living dossier is mostly links, and links are what rot — `bugs_open/071`, and
the four dead carousel destinations this workstream found by hand on 2026-07-26.

## 2. 001's hardest stated problem is one the register has already implemented

001 names its own worst risk, twice:

> "when the substrate updates ('Hormuz situation changed'), every derived angle is now potentially
> stale or contradicted. That fan-out needs to be counted and recorded at update time, not
> discovered later."

and proposes the fix as a distinction between **substrate-only** updates (facts change, narrative
does not) and **narrative** updates (the story changed; angles must be regenerated).

That is the same problem as: *an approved base entry's contract changes — is the approval still
good?* It is solved and live in `write_experience_pattern_action.go`:

- every column is classified **contract / selection / cosmetic / system**, and a column classified
  into none of them **fails the build** (`TestExperiencePatternColumns_EveryColumnIsClassified`);
- a change to a *contract* field demotes an approved entry to `draft` and logs which fields
  changed; a change to a *cosmetic* field does not;
- comparison is on canonical JSON, so key order cannot cause a spurious demotion — which matters,
  because a warning that fires spuriously is one people learn to click past.

001's "substrate-only vs narrative" is exactly this classification, applied to facts instead of
clauses. **Whoever builds 001 should copy the mechanism, not re-derive it** — in particular the
part that is easy to miss: the classification must be *compulsory*, or the list silently goes stale
and a changed fact keeps a stamp saying the angle was checked against it. That failure has now
happened four times in one file in this workstream alone (see `WRONG_CALLS.md`, 2026-07-28).

## 3. Where they must NOT be merged

The register holds **behaviour**: what a control does, where it leads, what must never be presented
as a control. 001 holds **editorial content**: facts, citations, an angle per audience.

Putting prose in `experience_patterns` would break the property the whole design rests on — that a
base entry is site-agnostic and carries no site-specific values (`bugs_closed/045`: a static value
baked into a shared base re-applies on every render and cannot be overridden per site). An *angle*
is by definition site-specific. It belongs in its own substrate/angle tables.

The correct seam: **a dossier page's behaviour comes from the register; its content does not.**

## 4. Two concrete things 001 can use today

- **`evidence-chart` exists.** The oufe session was asked to check whether a chart renderer had
  been built; it has (`content_components.evidence-chart`, section_type `evidence-chart`,
  `usage_count = 0` — built, live, unused). Owner: the brochure component library workstream.
  001's "historical data graphs" do not need a new renderer.
- **The harness gap is measured and ranked** (`SUMMARY_2026-07-28`). If 001's dossiers are mostly
  cross-links, the capability that most limits checking them is **attribute assertion** — 7 blocked
  clauses across 7 of the register's 9 entries — followed by **cross-page status** (a url on
  `page_status_ok`). That is a shared dependency, not this workstream's alone.

## What this workstream is NOT doing

Not building 001, not designing the substrate/angle tables, not claiming ownership of the topic
lifecycle. If 001 gets scheduled, the ask of this workstream is one entry —
`hub-spoke-index` — harvested from a live hub the way the other nine were, plus the binding work
that makes its spoke links checkable.

`[UNVERIFIED]` I have not read the news-feed-pooling docs or `features_open/002`; this note is
written from `001` itself and the oufe session transcript. Anyone acting on it should read those
before treating the mapping above as complete.
