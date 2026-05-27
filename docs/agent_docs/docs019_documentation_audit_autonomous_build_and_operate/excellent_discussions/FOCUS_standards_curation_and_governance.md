# FOCUS — Standards Curation & Governance (Who Maintains the Docs)

**Status:** exploratory design. Covers who creates and maintains the best-practice atoms, and how rule changes are confirmed. The coordinator role (§7) is flagged as likely-yes and still to be discussed. Amends the governance section of `FOCUS_best_practice_doc_tree.md` (§4 there) — this doc is the detailed source; that one points here.

Companions: `FOCUS_best_practice_doc_tree.md` (atom structure), `FOCUS_mediator_routing_model.md` (the advisor role curators also play), `FOCUS_doc_tree_adoption.md` (the careful first path).

---

## 1. Two different relationships to a doc

Kept apart so ownership doesn't leak into the volatile-coupling problem:

- **Consumption** — docs as runtime context, resolved by reference at spawn (constitution + objective node + tag-matched concern atoms). Never *owned* by an agent. This is the view, assembled fresh.
- **Maintenance** — docs as artifacts that need authoring and upkeep. This genuinely needs an owner. This document is about maintenance only.

"An agent responsible for a doc part" is the maintenance side, not the consumption side.

---

## 2. The unit of maintenance ownership is the concern — flat, not a tree

Ownership maps to the **concern**, not to a node in the agent tree, and the set of owners is flat (one per top-level concern, ~8–10 owners).

- Not tied to the agent tree: the agent tree is the volatile, fractal, runtime structure. Anchoring stable knowledge ownership to the most unstable structure is the coupling to avoid.
- Not a maintenance tree mirroring the objective hierarchy: that reintroduces the duplication the horizontal concern tree exists to remove — a fractal of owners each curating overlapping copies of the same cross-cutting rule. Concerns are horizontal, so their ownership is horizontal.

---

## 3. Concern curators

One curator (steward) agent per top-level concern, owning that concern's atoms.

- **Reuses the auditor pattern.** A concern curator is the existing auditor pattern (content-quality-auditor, component-quality-auditor) pointed at the standards instead of at sites.
- **Curation and advice are two faces of one owner.** The agent that advises on a concern at routing time (the routing FOCUS) is the natural one to maintain that concern's atoms, because it sees every violation. Same owner, two roles — keeps the count down.

---

## 4. What a curator does — and does not

**Does (vigilance + drafting + mechanical health):**
- Watches its concern's signal: violations logged by its advisor role, validator failures tagged to its atoms.
- Notices when an atom is wrong, missing, redundant, or routed badly.
- Drafts proposals: new standard, severity change, deprecation, merge of near-duplicates.
- Runs STEP-ZERO-for-standards before proposing (no duplicate, no contradiction in the same concern + overlapping `applies_to`).
- Maintains mechanical health: `applies_to` tags still match real change types; every `check` points at a validator that exists; deprecated atoms have their `referenced-by` cleaned up.

**Does not:**
- Hold write authority over a rule's *meaning*. It owns vigilance and drafting; the human owns the rule (see §5). This line is what stops "agent responsible for a doc part" collapsing into "agent rewrites the standards."

---

## 5. The HITL model: confirm, not initiate

**Amendment to earlier framing.** Rules are not authored by humans from scratch and are not changed unilaterally by agents. The decision's *reasoning* is agent-led; the human *confirms*.

**Why.** The complexity of these decisions often exceeds what a person can originate or fully reason through unaided. Expecting humans to author rules from a blank page mis-places the burden. Expecting agents to change rules alone removes accountability. The workable split: agents carry the analysis; humans confirm an informed, well-framed choice.

**The curator's output is a decision package, not a raw proposal.** To confirm responsibly without holding all the complexity, the human gets:
- a clear **summary** of what would change and why now,
- an **explanation** of the reasoning, the tradeoffs, and what is affected (the `referenced-by` impact),
- a small set of explicit **choices** — genuinely including reject and defer, not just accept.

Producing this package well *is* the curator's job. A proposal without a clear summary, surfaced tradeoffs, and real options has not been done.

**Guardrail against rubber-stamping.** Confirm-not-initiate fails if it becomes one-click assent. Mitigations:
- the choices must be genuine, and the summary must surface real risk, not sell the change;
- rigor scales with stakes — a `blocker` rule, an irreversible deprecation, or a change with a large `referenced-by` set demands more than a quick confirm (e.g. an explicit second look, or the coordinator's sign-off as well);
- low-stakes mechanical fixes (a stale `applies_to` tag, a dead `check` link) can be lighter-weight, since they don't change a rule's meaning.

**On acceptance:** writes a new `version`, flips the old to `deprecated` (never deletes); spawn-fresh agents read the current `active` version, so no stale rule persists in a long-running process.

---

## 6. Straddle resolution (the cross-concern seam)

Cross-cutting rules have several legitimate claimants (a rule about a Go action writing to the DB is code-style *and* data-schema *and* maybe messaging). Ambiguous ownership produces gaps (everyone assumes someone else owns it) or collisions (two curators drafting conflicting edits).

Resolved by the atom's fields:
- `applies_to` is **plural** — the rule can govern several change types.
- `concern` is **singular** — exactly one owning concern, even when `applies_to` is broad. This is the ownership key.
- The home concern for a straddling rule is a human call made **once at authoring time** and recorded on the atom — not negotiated by curators per change.

---

## 7. The coordinator role (likely-yes, to discuss)

The curators are peers; peers can't arbitrate each other, and some things belong to no single concern. A thin coordinator layer above the curators is the candidate owner for exactly those:

- the **concern taxonomy** itself (the top-level set — slow-moving, schema-like),
- the **`applies_to` vocabulary** (shared with routing),
- **cross-concern conflicts** when two curators' atoms disagree,
- (per §5) plausibly the **packager of cross-concern decision packages** for human confirmation, since these span multiple curators' areas.

This is the one place a sliver of hierarchy may earn its place, precisely because the seams between concerns need an owner and peers can't resolve each other. **Open for discussion:** its exact scope, and whether cross-concern conflicts route through it or straight to a human.

---

## 8. One-line state

Maintenance ownership is flat, by concern, not tied to the agent tree: one curator per concern (reusing the auditor pattern, doubling as the routing advisor), with vigilance-and-drafting authority but no authority over rule meaning. Rule changes are agent-reasoned and human-confirmed via a summary/explanation/choices decision package, with rigor scaled to stakes to avoid rubber-stamping. Single-`concern` ownership resolves straddles. A coordinator over the curators (taxonomy, vocabulary, cross-concern conflicts, decision packaging) is the next thing to pin down.
