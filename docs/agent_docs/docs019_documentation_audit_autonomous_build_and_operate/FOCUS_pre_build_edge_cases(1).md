# FOCUS — Pre-Build Edge Cases and Reasoning Bugs

**Status:** a hard adversarial pass over the whole design before building, hunting for the things that don't show up as document-to-document inconsistencies but as "this breaks when X happens." Grouped by severity. Each item is the problem in plain terms plus how to handle it. None of these block starting; all are cheaper to settle now than to hit in code.

---

## 1. Things that would corrupt data or hang — settle in the contracts

### 1.1 Tree cycles would hang the compute-on-read walk
The objectives tree (`parent_id`), the version chains (`previous_version_id`), and `standards.related` are all self-referential. The effective-priority-profile walk goes root→leaf; if a cycle ever gets in (a bad confirm, a mistaken edit), the walk loops forever and the bundle builder hangs. Confirm-not-initiate doesn't save us — a human can confirm a cycle by mistake.
**Handling:** enforce acyclicity on write/confirm (reject an edit that would create a cycle), and make the walk cycle-detecting (bounded, fails loud) as a backstop. Cheap to add now, nasty to diagnose later.

### 1.2 Orphaned proposed rows / two proposals for one target
A proposal writes a layer row with `status = proposed` and emits a work-item. The work-item dedup (`item_key`) stops two *work-items* for one target — but nothing stops two *proposed layer rows* if two agents (or the same agent twice) propose for the same target before resolution. When one work-item resolves it points at one version; the other proposed row is orphaned, stuck at `proposed` forever.
**Handling:** one live proposal per target. A new proposal **replaces** the existing proposed row rather than adding a second, and expiring/superseding a work-item also resolves (deprecates) its proposed row. The "at most one live thing per target" rule that the work-items already have must extend to the proposed layer rows.

### 1.3 The confirmer's apply must be atomic and recoverable
The confirmer does several things to apply one confirmation: flip the row to `active`, deprecate the prior version, write the decision-log entry, emit the in-band change event. If it crashes between steps you get a half-applied state — an active row with no decision-log entry, or no in-band event so maintenance never learns the change happened.
**Handling:** the database writes go in **one transaction**; the in-band event uses an **outbox** (write the event to an outbox table in the same transaction, a relay publishes it), so a crash leaves a consistent state and a retry completes. The contract already says the apply is idempotent — this is what makes idempotency hold across a crash, not just a repeat.

### 1.4 Bundle assembly can take a torn read
The bundle reads standards, then objectives, then mechanical config. If a confirmation lands between those reads, the bundle mixes pre-change standards with post-change objectives — internally inconsistent context, and the logged hash describes a state that never coherently existed.
**Handling:** assemble the bundle from a **consistent snapshot** (one transaction / one point-in-time read across the layer tables).

---

## 2. Things that would make it painful or unstable — design the handling now

### 2.1 Cold-start + confirm-every + a big first config = a confirmation avalanche
Every new (tenant, capability) starts at `confirm_every`, and the first onboarding of our own repo produces a large batch of proposals at once — every convention atom, every objective node, every mechanical entry, plus the docs-vs-code drift findings. At confirm-every that is potentially hundreds of individual confirmations. Onboarding becomes a wall of clicking, which is exactly the cliff the progressive design was meant to avoid.
**Handling:** **batch confirmation** for the initial onboarding — confirm a whole proposed set with review, not item-by-item — while keeping per-item confirmation for later drift. The gate stays; the granularity adapts to the initial bulk.

### 2.2 De-graduation must ignore transient and infra failures
De-graduation can auto-apply on "severe evidence." A flaky test or an infra blip is not a real regression, but if it counts as evidence, one blip drops a capability to `confirm_every` and triggers a confirmation avalanche (2.1) for no real reason. The master plan's defect-vs-partition idea exists for exactly this but isn't wired to the de-graduation trigger.
**Handling:** the evidence that feeds de-graduation must pass the **defect-vs-partition filter first** — infrastructure-attributed and transient failures are quarantined out before they can lower trust. Wire the filter into the trust-ledger evidence path explicitly.

### 2.3 Partial config is normal during onboarding — degrade, don't fail loud
The "fail loudly, don't tolerate the unexpected" rule and the "onboarding is never fully done" reality collide. During onboarding the config is legitimately incomplete — some areas have no why-chain, some change-types match no standards yet. The bundle builder must not treat that as an error.
**Handling:** distinguish **legitimately-absent (pending onboarding)** from **malformed (a real fault)**. The first degrades gracefully (use what's there, say what's pending); only the second fails loud. The no-fallbacks rule applies to malformed data, not to not-yet-authored data.

### 2.4 The reuse-search index is derived state that goes stale silently
Reuse-search runs over the pgvector index of code and definitions. When code or definitions change, the index is stale until re-indexed — and a stale index doesn't fail, it quietly suggests reusing something that has moved or been deleted. Silent wrong suggestions are worse than a loud failure.
**Handling:** re-index on the change layer (a code/definition change triggers re-indexing of what changed), and stamp the index freshness so reuse-search can flag "index last built at T" rather than imply currency it doesn't have.

---

## 3. A conceptual conflation to fix: config changes vs deliverables

The contracts gate **config** changes (standards, objectives, mechanical) through the work-items + central confirmer. But the tool's actual **deliverables** — generated code, and in this data-defined system, edits to workflows and agent definitions (jsonb on `agent_definitions`) — are a *different* thing gated by *different* machinery: the trust ledger + the reliability cascade + verification. These have been allowed to blur.

Two distinct gated paths, and the plan should name them separately:
- **Config path:** changes to the tool's own knowledge of the codebase → `config_work_items` → central confirmer → active config.
- **Deliverable path:** the tool's outputs (code, workflow/agent edits) → cascade (reuse/generate/verify) → trust-ledger gate → applied to the codebase, committed, in-band change event.

The decision log spans both, but the gates are not the same gate. Editing a workflow is a deliverable, gated by the capability's trust level, **not** by the config confirmer. Making this explicit avoids building one gate where two are needed.

---

## 4. Correctness refinements

### 4.1 The in-band guard must not suppress legitimate downstream effects
The guard says a confirmer-apply in-band event doesn't re-trigger maintenance on the entry just confirmed. Correct — but confirming a *new convention* should still trigger a code audit against it (does existing code comply with the new rule?). That is a different target (the code, not the convention). The guard must exempt **re-validating the confirmed entry itself**, not **downstream effects of it**.
**Handling:** scope the guard precisely to the confirmed target, and let genuine downstream triggers (audit code against a newly-active standard) fire.

### 4.2 Probe with `LIMIT N+1`, not `count(*)`
The multipass probe assumes counting is cheaper than fetching. For a query with a complex filter, `count(*)` is as expensive as the query — the probe doesn't save anything.
**Handling:** probe by fetching `LIMIT N+1` and checking whether the cap was hit (got N+1 → "more than N", reduce to sample/aggregate/pointer) rather than counting the whole result. Cheap, and it answers the only question the gate actually asks.

### 4.3 Log what the generator *saw*, not what we *assembled*
The generator consumes the rendered text form; provenance is taken from the structured form. If rendering drops or truncates something, the log overstates the context the generator actually had — and a later audit ("did this decision see the right standards?") would be wrong.
**Handling:** derive provenance from the **rendered** bundle (what was actually handed over), or prove the rendering is lossless against the structure. The audit must reflect what the generator saw.

### 4.4 Don't force a "premise" on mechanical decisions, and don't log the logging
The decision log's `premise` field fits judgement decisions ("chose X because…") but not mechanical ones — a bundle assembly has a rule-trace ("these standards matched these change-types"), not a reasoned premise. Forcing a premise there produces noise. And writing a log entry is a side-effect of a decision, not itself a decision — logging must not be recursive.
**Handling:** allow two flavours in the log — judgement (premise) and mechanical (rule-trace) — so each records what it actually has; and state plainly that logging is not itself a logged event.

---

## 5. Operational safety: the tool can break the platform it runs on

The tool runs on the chassis it modifies. A change it makes to the chassis could, in the worst case, break the chassis badly enough that the tool can't run to fix it. This is the self-hosting trap.
**Handling:** recovery must not depend on the tool. Rollback to a known-good state must be runnable **externally** — a path that does not route through the agents/orchestrator the tool uses. The build/operate asymmetry (rollback is the safe direction) already points here; the requirement is that the rollback path has no dependency on the thing it's rolling back.

---

## 6. Prerequisites to do before (or at the very start of) code

- **Author the governed vocabularies first.** The concern taxonomy (for `standards.concern`) and the priority dimensions (for the profiles) are fixed vocabularies the conventions and intent agents classify *into*. They must be seeded/authored before those agents can run. Currently assumed, not called out.
- **The first constitution is hand-authored.** The tool that would help write it isn't built yet, so the initial constitution is written by hand from 001/003 + preferences. Expected, worth stating so it isn't treated as a chicken-and-egg blocker.
- **Verify the remaining reuse interfaces** (schema-before-SQL, generalised to interfaces): the orchestrator/spawn API, the adapter interfaces, the vector-store API, and how the `check_*.go` validators are invoked. We verified `site_work_items` and `agent_definitions`; these are the other things the build leans on and assumes.
- **"Us" is a real tenant row**, not a sentinel or special case — so the single-tenant path exercises the same code as the multi-tenant one (dogfood the isolation), and there is no separate "no tenant" branch to maintain.

---

## 7. One-line state

The design holds up, but a careful pass found: five data-integrity or hang risks to settle in the contracts (tree cycles, orphaned proposals, confirmer atomicity, torn reads, one-live-proposal-per-target); four stability/usability risks to design for (confirmation avalanche, de-graduation on blips, partial-config degradation, stale reuse index); one conceptual fix (config changes and deliverables are different gated paths); four correctness refinements (guard precision, probe by limit, log-what-was-seen, premise-vs-rule-trace); one operational rule (external rollback for self-hosting); and four prerequisites (author the vocabularies, hand-write the first constitution, verify the remaining interfaces, make "us" a real tenant). All are cheaper to handle now than after they're built on.

---

# Second pass — higher-order findings (final look)

The fixes above are applied. This second look checks they're coherent and steps back to question what the earlier passes assumed. It found two small seams the fixes opened, two real gaps, and one strategic point about build order.

## 8. Build order — validate the premise with a thin slice before building the infrastructure

This is the most important item here. The entire design rests on one unproven premise: **that an assembled bundle actually beats the current paste-and-rot workflow.** Six contracts, a queue, a ledger, a log, and an onboarding system are all in service of that premise — but none of them test it. Building them all before assembling a single useful bundle is a lot of work staked on an assumption.

The cheapest test is a **thin vertical slice** that skips almost all the infrastructure:
- hand-write a minimal constitution as a flat markdown file (not the `standards` table);
- build the Go analyser (signatures + reuse-search) and the schema extractor;
- assemble **one** bundle for **one** real task, render it to text, paste it into a chat by hand.

No `config_work_items`, no trust ledger, no decision log, no change layer, no confirmer. If the assembled bundle visibly helps on a handful of real tasks, *then* build the infrastructure to automate and scale it — and the contracts are the right destination for that. If it doesn't help, far better to learn it now. The contracts being well-specified means they'll still be there; sequencing the slice first de-risks the whole effort.

## 9. The loop runs on outcomes, but outcomes aren't sourced (real gap)

The trust ratchet and the improvement loop graduate or de-graduate on evidence — "recent success rate, last verification outcome." But the contracts log **decisions and inputs**, not **outcomes**. The `decision_log` records what was decided and what was in hand; `evidence_summary` is "populated by the maintenance agent" — from what raw signal? Nothing currently records *did the deliverable succeed*: did it pass verification, did it get reverted, did a human have to fix it afterward.

**You can log every input and decision and still have no feedback signal.** Before Phase 2 (where the ratchet runs), define the **outcome record** — the result of a deliverable (verification pass/fail, reverted, human-corrected, accepted-as-is) — as the raw signal the maintenance agent aggregates into `evidence_summary`. Without it the ratchet has nothing real to move on.

## 10. Define "the bundle helped" (real gap, smaller)

Related: the success metric for a bundle is fuzzier than build/test-pass. It looks like fewer correction rounds, fewer convention violations in generated output, less manual context-gathering before the task can start. Name it early, because §8's thin-slice test and §9's ratchet both need this signal, and it won't define itself.

## 11. Does `bundle_assembly` belong in the decision log? (seam the fix opened)

Adding `rule_trace` let mechanical assemblies sit in `decision_log` — but the doc still calls itself "the published-reasoning log, one row per decision carrying the premise (the reasoning)," and bundle assemblies (one per task, no reasoning) would **dominate it by volume**, diluting a reasoning log into mostly-not-reasoning. The §140 summary now contradicts the `rule_trace` addition.

A cleaner model: a bundle's provenance is the **`inputs_used` of whatever decision consumed it** — in Phase 2+ that's the generation/action decision (the bundle is its input, not a decision of its own); in Phase 1 (human consumes the bundle) it's a lightweight **task-attached snapshot**, not a reasoning entry. That keeps `decision_log` for reasoned decisions and puts bundle provenance where it belongs. Decide this before building the log — it changes the table's shape and purpose.

## 12. The work-item's `target_ref` pins a version, which the "replace" fix makes stale (seam the fix opened)

`target_ref` is `{table, id, version}`. But the one-live-proposal fix says a new proposal **replaces** the proposed row (a new version), which leaves the existing work-item pointing at a superseded version. For a *proposed* row that can be replaced, pinning a version is wrong.
**Handling:** the work-item references the **identity** (`{table, id}`) for proposed rows, and the confirmer applies **the current live proposed row** for that identity — not a pinned version. (Version still matters when deprecating the prior *active* row.) Small, but it has to be consistent or the confirmer applies the wrong version.

## 13. Blast radius should set the trust ceiling (make explicit)

The catalog sets a capability's ceiling from verifiability + containment. Containment is the one that matters most here and should be stated as a seeding principle: **the more a capability can break, the lower its ceiling — regardless of how verifiable it is.** A capability that edits the chassis itself has low containment (it can break everything, including the tool) and should have a low ceiling — never fully autonomous. A capability that edits one product site has high containment and can graduate higher. The mechanism exists; the principle ("blast radius caps trust") should be explicit so the catalog is seeded that way.

## 14. Recursive self-improvement is a residual risk, not a solved one (name it)

The tool improves the chassis that includes the tool. A self-improvement that passes verification could still degrade the tool's own judgement in ways the checks don't catch — gradual drift, not a crash. This isn't fully solvable. It's *managed* by conservative early trust levels, the human in the loop, external rollback (§5), and the low ceiling for chassis-touching capabilities (§13). Worth naming as an accepted residual risk rather than assuming the safeguards close it completely.

## 15. Two smaller notes

- **Reuse-search over definitions needs a searchable text projection.** Embedding a workflow (jsonb) or a tool row for similarity means deciding *what text* represents it. A real implementation decision, not automatic. (Bundle-contract open item.)
- **Batch confirmation must respect `depends_on` order.** Approving a batch is fine, but the confirmer still applies in dependency order — the batch is approval granularity, not apply granularity.

## 16. One-line state (second pass)

The fixes are coherent except two small seams now flagged (`target_ref` identity vs version; the `decision_log` summary wording vs `rule_trace`). The real gaps are the unsourced **outcome signal** (the loop's feedback) and the undefined **success metric** — both needed before Phase 2. The strategic point dominates: **validate the premise with a thin slice before building the six-contract infrastructure.** The design is the right destination; the slice is the cheap way to make sure it's worth reaching.
