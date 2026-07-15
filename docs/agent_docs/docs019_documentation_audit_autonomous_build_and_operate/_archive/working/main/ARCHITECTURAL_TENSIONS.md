# Architectural Tensions — a catalogue

A living list of recurring, genre-level design tensions in the platform —
distinct from `016_debugging_guide`, which is symptom → cause → fix for specific
incidents. This document is one level up: it names the *class* of problem that
keeps generating those incidents, so we recognise the next instance as a member
of a known family rather than debugging it from scratch.

Each entry states the tension, the principle in conflict, the instances observed,
the resolution principle, and status. An entry graduates from "observed" to
"resolved" only when the resolution principle is actually enforced in code, not
just agreed.

---

## How to read an entry

- **The pattern** — the recurring shape, in one or two sentences.
- **Why it recurs** — the structural reason it keeps happening, so we stop
  treating each occurrence as a one-off.
- **Instances observed** — concrete sightings, with dates/sites, linked to the
  debug-guide entry where the incident is recorded.
- **The reliability principle** — the design rule that resolves the tension.
- **Open design question** — where the resolution isn't yet decided.
- **Status** — observed / partially-addressed / resolved.

---

## Tension #1 — Trusting LLM free-text structure as truth (infer-and-repair) vs deriving structure deterministically

**Status: observed; resolution principle agreed in direction, not yet enforced.**

### The pattern

The pipeline repeatedly takes a *structural* decision — what type a page is,
where its URL goes, what its role is — from the LLM's free-text output, then
tries to *repair* that output after the fact with heuristics. Every repair is a
guess about what the LLM meant. When the LLM emits a shape the heuristics didn't
anticipate, the guess misses, and the system silently produces wrong structure
(usually by defaulting to `content`, which flattens section hubs into ordinary
pages).

The defining symptom: **a page looks built, but its structure is wrong**, and
nothing errored. Silent structural corruption is the worst failure mode because
there's no signal to investigate until you inspect the output by hand.

### Why it recurs

Three compounding reasons:

1. **Structure is taken as a label, not derived from a relationship.** Role
   (`section-index` / `content` / `tool` / `blog-post`) is something the LLM is
   asked to *assert*. But a label is exactly the kind of thing LLMs are
   unreliable at — it's an arbitrary taxonomy choice, not a fact about the
   content. In the 2026-05-23 run the same model typed `guides-index` as
   `blog-index` (right) but `games-index` and `tools-index` as `content`
   (wrong) — inconsistent labelling of three structurally identical pages, in
   one response.

2. **Repair reads signals the LLM doesn't reliably emit.** The validator
   (`ValidateRoles`) infers section-index status from the URL pattern
   (`/tools/index.html`) or from a child page declaring `ParentSection: "tools"`.
   But this planner emitted **no URL and no parent_section on any page** — so all
   the structural rules whiffed and every section index fell through to its raw
   `content` label. The validator wasn't wrong; it was starved of the signals it
   was built to read. (Its own tests all supply URLs/parents, which is why they
   pass while production fails.)

3. **The repair vocabulary is hard-coded to one vertical.** `nestedRoleFromURL`
   recognises only `tools` / `guides` / `games` as section directories. A vet
   site (`treatments`, `conditions`), a recipe site (`cuisines`), a property site
   (`listings`) — none match, all fall through to `content`, all flatten. The
   heuristic is silently game-design-specific in a platform meant to build any
   vertical from a domain.

The deeper reason all three recur: **the system trusts the LLM for the thing the
LLM is worst at (consistent formal labelling) and then tries to recover with
heuristics that need other things the LLM also didn't reliably provide.** Adding
another heuristic moves the next failure to the next unanticipated shape; it
doesn't change the structure of the problem.

### Instances observed

- **LLM inventing/mislabelling page roles.** `analyze_site` (adoption) typed
  `games-index`/`tools-index` as `content`; `build-site-planner`'s `plan_site`
  faithfully kept `content`; `CanonicalisePage` then flattened `games-index` →
  `games` at `/games.html` and stripped guide prefixes. Two LLM touchpoints, same
  unreliable label, no deterministic correction that survived. Recorded in
  `016_debugging_guide` ("Adoption faithfulness: … WriteSitePlanAction strips
  identity for content/blog_post types"). gamesdesign.co.uk, 2026-05-23.

- **`ValidateRoles` hard-coded directory vocabulary.** `tools`/`guides`/`games`
  baked into `nestedRoleFromURL`; no path for other verticals. Latent — will fire
  on the first non-game-design adopted site with section hubs.

- **`CanonicalisePage` selective prefix re-add.** Re-adds the type prefix for
  `tool`/`game`/`guide`/section-index roles but not for `content`/`blog_post`, so
  a page mistyped into those two roles loses any `-index`/`guide-` identity its
  name carried. This is the repair layer *amplifying* the upstream mislabel
  rather than catching it.

(Whether the missing-`index.html`-on-deploy report is an instance: **ruled out.**
`index` was typed `landing`, planned, and deployed correctly per the DB. If a
file is missing from git/s3 it's a deploy-adapter issue, a different stage.)

### The reliability principle

**Constrain the source and derive structure deterministically from the LLM's
*reliable* signals; never trust its formal labels, and fail loudly when signals
conflict or are absent rather than defaulting silently.**

Concretely, three moves, in order of leverage:

1. **Stop reading the formal `page_type` label as truth.** Derive role from the
   signal the LLM is actually reliable at: **naming**. The model named
   `games-index`, `tools-index`, `guides-index` correctly and consistently even
   while mislabelling their types. A name ending in `-index` is a section hub —
   and this signal is **vertical-agnostic** (`treatments-index`, `cuisines-index`
   work identically), unlike the hard-coded directory list. Derivation from a
   reliable signal beats repair of an unreliable one.

2. **Schema-constrain generation to eliminate *form* errors** — but understand
   its limit. Structured-output / tool-use with a closed `page_type` enum and
   required fields makes "missing URL", "missing parent", and "invented role
   string" into hard, retryable failures instead of silent fall-through. **It
   does not fix correctness** — a closed enum still lets the model pick `content`
   for a hub. Schema guarantees well-formed, not right. So schema is necessary
   (it removes the missing-field class entirely) but not sufficient; it must sit
   *under* deterministic derivation, not replace it.

3. **Make the repair layer fail loud.** When the deterministic derivation can't
   confidently classify a page (no name signal, no relationship, conflicting
   signals), the answer must be "flag it / re-prompt / surface it", never
   "default to content". A heuristic that always produces *an* answer hides its
   own failures. A layer that can say "I don't know" is more reliable than four
   heuristics that are always confident and sometimes wrong.

The unifying idea: **separate what must be reliable (structure) from what the LLM
produces (content + naming), make the reliable part a deterministic projection of
the LLM's most-reliable output, and treat any fallback heuristic firing as a
monitored signal that the structured path failed — not as a normal code path.**

### Open design question — should the LLM emit an explicit page tree (parent pointers)?

This was the specific fork. **Recommendation: no — do not make a free parent/child
tree the primary structural contract.** Reasoning, because "complex = bad" alone
isn't the argument:

LLM reliability is tiered. They are reliable at *naming* and *semantic intent*
(the model named the hubs correctly and gave `games-index` a `game-list` section).
They are unreliable at *formal taxonomy labels* (the `content` vs `blog-index`
inconsistency). They are *least* reliable at *fields they must remember to
populate consistently across every item* — and in this run the model omitted
`url` and `parent_section` on **every** page. A free parent-pointer tree lands
squarely in that third, worst category: a relational field that must be present
and correct on every page. Requiring it fights the model's specific weakness, and
a hallucinated parent is no better than a hallucinated label.

The better contract leans on the tier the model is reliable at:

- **Let the naming convention encode the hierarchy.** `<section>-index` marks a
  hub; the set of hubs *defines* the site's sections, vertical-agnostically. A
  leaf's section is derivable from its name/path against that set. The model
  already produces these names reliably; we make role a deterministic function of
  them rather than trusting a separate label.
- **Where an explicit relationship is genuinely needed, ask for the *minimal,
  constrained* one** — a leaf's `section` chosen from the **closed set of
  hub pages**, not a free tree. A constrained choice over an enumerated set has a
  far smaller hallucination surface than free parent pointers, and it degrades to
  "derive from name" when omitted.
- **Role is then derived**: page is a hub (name/`-index`, or others point to it) →
  section-index; belongs to a hub → leaf of that section; neither → content /
  landing. No formal label trusted at any point.

So the long-term route is *not* "ask the LLM for more structure" and *not* "add
another heuristic." It is: invert the contract so the deterministic layer
*assigns* role/URL from canonical naming, schema-constrain generation to kill
form errors, and fail loud on ambiguity. This is more work than a fourth
validator rule, but it's the difference between "fixes gamesdesign" and "reliable
across verticals we haven't seen" — and silent flattening on every new vertical
is the cost of not doing it.

### Status and next step

Observed and diagnosed; resolution direction agreed. Not yet enforced. The
minimal honest first step that moves toward the principle without a big-bang
rewrite: **add a name-suffix (`-index`) derivation to `ValidateRoles` as a
structural rule, and de-hard-code `nestedRoleFromURL` to read the site's actual
section set rather than `tools`/`guides`/`games`** — both deterministic, both
vertical-agnostic, both reading reliable signals. The schema-constrained
generation and loud-fallback pieces are the larger follow-on. Sequence and scope
TBD; this entry is the rationale, not the implementation plan.

---

## Tension #2 — (reserved)

Future entries follow the same template. Candidates noticed but not yet written
up: the lock-model coherence debt (two patterns, implicit hard/soft, half-built
expiry — see `PLAN_lock_coherence.md`); the convergence/canonicaliser running in
two places (validate_plan preserves, write_site_plan then re-derives and can
undo the preservation); inferring pipeline behaviour from intermediate signals
(see the debug guide's tempo-trap entry — arguably a debugging tension rather
than an architectural one).
