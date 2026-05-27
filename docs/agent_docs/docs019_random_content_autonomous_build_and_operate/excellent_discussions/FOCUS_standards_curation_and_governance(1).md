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

## 7. The coordinator role

The curators are peers; peers can't arbitrate each other, and some things belong to no single concern. A thin coordinator layer above the curators owns exactly those:

- the **concern taxonomy** itself (the top-level set — slow-moving, schema-like),
- the **`applies_to` vocabulary** (shared with routing),
- **cross-concern conflicts** when two curators' atoms disagree,
- (per §5) the **packager of cross-concern decision packages** for human confirmation, since these span multiple curators' areas.

**Resolved:** the coordinator both *arbitrates and frames* — it does not merely detect conflicts and route them. The objection to arbitrate-and-frame (whoever frames the choice shapes what the human confirms, and the human is poorly placed to second-guess the frame) is answered not by stripping the coordinator's framing role but by giving the user a dedicated advocate inside the framing process (§8). The check on framing power is a user-aligned agent that can contest the frame, not a human who can't.

---

## 8. Advocacy and the user-representative

A standing agent (or several — see §8.4) representing the user sits inside the framing process as the check on the coordinator's framing power. Early framework iterations already reached for this: an agent that dissects the user's input and helps frame decisions for the user. The intake orchestrator and the briefing questionnaire are the seed — one-shot user-input dissection at intake; the user-rep generalises that into a standing advocate. Reuse, not a new invention.

### 8.1 Two kinds of advocate
- **Curators advocate for concerns** — correctness, consistency, the standards.
- **The user-rep advocates for intent** — fidelity to what the user actually wants, and their longer-term interest.
- **The coordinator arbitrates between them and frames** the result for human confirmation.

Where a curator says "this violates the data-schema standard," the user-rep is the voice that can say "the user asked for this explicitly, so either the violation is intentional and we confirm it, or the standard shouldn't apply here — and possibly there's no conflict at all once you see what they were after."

### 8.2 Triaging claimed conflicts (the subtle-tension problem)
When the coordinator detects that two user inputs "conflict," there is not one situation but at least three, wanting different responses. The failure mode is collapsing all three into "choose A or B," which forces a false binary on the latter two — or worse, resolves them silently and wrongly.

1. **Real conflict the user hasn't noticed** → surfacing it is genuinely valuable; becomes a decision.
2. **Real tension the user has already implicitly reconciled** → the user isn't contradicting themselves; the model lacks the context to see how the two fit. There is a resolution the coordinator missed. → dissolve, don't escalate.
3. **Illusory conflict** — an artifact of imprecise wording or a shallow read. → dissolve, don't escalate.

So the user-rep's signature job is not only dissecting input up front but **triaging claimed conflicts before any reach the user as a decision.** Dissolve what's illusory or already-reconciled; escalate only genuine tradeoffs.

### 8.3 Clarify vs decide
The case the plain decision-package model (§5) has no slot for: when the model is unsure whether a conflict is even real, it goes to the user as a **clarification, not a decision** — "I think these might pull against each other for this reason, but I suspect I'm missing something — is this a real tension or am I misreading it?" Asking-to-understand and asking-to-decide are different acts. A forced choice freezes a possibly-wrong interpretation; a clarifying question defers to the actual intent. Same references-not-copies instinct, applied to framing: don't reify a surface contradiction into a decision before you understand the why underneath it.

### 8.4 "Or several" — stakeholder advocates
The reason for several is not copies of one role but **distinct stakeholders whose interests genuinely diverge**: the operator making the change, the client whose site it is, and the site's end-users (ties to the platform's own "what's best for the users of the site"). Each is an interest that can conflict with the others; the coordinator arbitrating among stakeholder-advocates as well as concern-curators is a clean generalisation. Caution: don't stand all of them up at once before the single operator-facing user-rep has proven it earns its keep.

### 8.5 Guardrails
An advocate for the user has two failure modes:
- **Putting words in the user's mouth** — confidently advocating an intent the user doesn't hold (hallucination wearing a sympathetic hat). Discipline: ground every claim about intent in actual inputs or history; when extrapolating, say so and ask rather than assert.
- **Yes-manning** — rationalising whatever the user said even when a curator's blocker is right and the user is about to break something. Discipline: the user-rep argues for intent but does **not** override a genuine blocker. It makes the case; the coordinator weighs it; the human confirms. An advocate, not a trump card.

---

## 9. Decision authority

### 9.1 Default: co-equal voices in the frame
When the user-rep and a curator genuinely disagree and the coordinator can't dissolve it, both cases are presented to the human as **co-equal voices** — the human sees "the architecture curator says X, your advocate says Y," not a single pre-synthesised recommendation. This keeps the disagreement visible, which is the whole point of giving the user an advocate, and prevents the coordinator re-concentrating the framing power the user-rep exists to check.

### 9.2 Abstention fallback: the advocate decides, bounded
If the user **chooses not to choose**, the advocate (user-rep) makes the decision — not arbitrarily, but reasoning from the codified, durable expression of intent: the mission documents, architectural best practices, the standards, and longer-term benefit, reliability, and robustness. Rationale: the advocate is the user's standing proxy, so when the live user is silent, the proxy acts on their behalf using the codified intent rather than the live preference. When the human is silent, the system leans on the user's longer-term interest, not on concern-correctness alone.

Two guardrails on the fallback:
- **Stakes-scaling (consistency with §5).** Distinguish "I trust you, decide" (delegation) from non-engagement (neglect). Both end in the advocate deciding for low-stakes calls; but neglect on a high-stakes decision should **escalate to the creator (§9.3), not auto-decide.**
- **No silent blocker override.** A genuine curator blocker plus user abstention is exactly the high-stakes case: it escalates rather than the advocate resolving it against the blocker. The "advocate is not a trump card" rule (§8.5) holds in the fallback too.

### 9.3 Creator override / veto
The framework creator holds a privileged authority over difficult or contentious decisions: at minimum a **veto** (negative authority — can block a bad or contentious outcome without being forced to author one), and optionally **final choice** (positive authority). The veto is the safer minimum, because it lets the creator stop a contentious decision without becoming the bottleneck for forcing every outcome. Reserved for difficult/contentious decisions and high-stakes abstention escalations — not invoked routinely, or it recreates the bottleneck the rest of the model avoids.

### 9.4 Three distinct "human" roles (to keep unambiguous)
- the **user** whose intent the advocate represents,
- the **confirmer** who decides among framed options at the gate (often the operator),
- the **creator** with override/veto.

These may be the same person in a self-development context, but they are distinct roles. The advocate represents the user's intent; the confirmer chooses among framed options; the creator can override. The abstention fallback (§9.2) is the confirmer declining, after which the advocate acts as the user's proxy, with the creator above all of it.

---

## 10. One-line state

Maintenance ownership is flat, by concern, not tied to the agent tree: one curator per concern (reusing the auditor pattern, doubling as the routing advisor), vigilance-and-drafting authority but no authority over rule meaning. The coordinator arbitrates and frames; its framing power is checked by a user-representative advocate (§8) that triages claimed conflicts (dissolve illusory/already-reconciled, clarify-not-decide when unsure, escalate only genuine tradeoffs) and advocates intent against the curators' concerns. Authority: co-equal voices to the human by default; on abstention the advocate decides bounded by mission/architecture/long-term reliability, with high-stakes abstention and blocker conflicts escalating; creator holds veto (or final choice) on contentious calls. Rule changes remain agent-reasoned and human-confirmed via decision packages, rigor scaled to stakes.
