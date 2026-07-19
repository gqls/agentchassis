# PLAN — Claims verification

**Written retrospectively 2026-07-19.** It should have existed from 2026-07-16,
when the build started; the standing-five directive postdates the work. Everything
below is reconstructed from the decisions as they were actually taken, and the
corrections are recorded as corrections — not tidied into the original design.

Originating brief: `SPEC_claims_verification.md` (owner-approved 2026-07-16).
Running record: `NOTES_claims_verification.md`. Commands: `RUNBOOK_claims_verification.md`.

---

## The design in one line

Make evidence machine-readable per site, then check claims against it at three
points the platform already has hooks — build gate, post-deploy discovery, and the
writer's prompt — with every judgement call terminating at a human.

## Phasing (as built)

| Phase | What | State |
|---|---|---|
| V0 | `evidence_base` formalised as structured data, transcribed from the hand audit | LIVE 2026-07-16 |
| V1a | Build-time gate: banned claims (blocker) + unregistered numbers (error) | LIVE, in prod since v1.0.1128 |
| V1b | Post-deploy discovery check `unverified_claims` | LIVE, enabled on `quality-discovery-agent` |
| V2 | Writer whitelist injected into the page-content-writer prompt | LIVE 2026-07-17 (DB config, no image needed) |
| V3 | `claims-auditor` agent — LLM judgement lane for prose | LIVE 2026-07-18; not on a cadence (owner call) |
| V4 | Freshness: re-run SQL facts, regenerate whitelist, raise drift | Committed `06376bcbf`; **inert until an image ships** |

Build order was deliberate: cheapest and most certain first. V0 carries no code
risk; V1a is the smallest code change catching the worst class; V3 is the only
piece needing prompt-design iteration, so it went late.

## Decisions, and why

1. **One shared scan engine in `datahelpers`, consumed by both the gate and the
   discovery check.** Precedent: `ExtractHrefs`/`PageURLSet`, where the deploy gate
   and the phantom-link audit agree by one literal implementation. Two copies of a
   claim scanner would drift, and the drift would be invisible — a claim blocked at
   build but ignored post-deploy, or vice versa.

2. **Parse assertion text nodes, never raw HTML.** An email or number inside a
   `placeholder=` attribute, a `<code>` sample or a `<script>` body is not an
   assertion about the business. This is not hypothetical: the email checker once
   flagged `placeholder="jane@company.com"` and blocked every build of every page
   using the shared contact block. `mailto:` hrefs are the single attribute surface
   that *is* an assertion (published contact), so they are included explicitly.
   *Recorded boundary:* `alt`/`title` text is user-visible but not scanned —
   deliberately deferred to V3 rather than silently ignored.

3. **Banned claims are NOT injected into the writer's prompt.** Naming a
   fabrication to a language model puts it in context — don't-think-of-an-elephant.
   The whitelist (what you may say) goes in; the blacklist (what you may not) stays
   out and is enforced deterministically at the gate. Each mechanism does what it
   is good at.

4. **Severities: banned = blocker, unregistered number = error.** A banned pattern
   is a *known* falsehood placed on the list by a human, so blocking is safe.
   Number extraction has false positives by construction, and `error` already
   routes to human review rather than deploying. This answers the spec's open
   question 1 in favour of `error` over `warning` — the one false positive found in
   practice was an engine bug worth fixing, not review noise.

5. **Everything terminates at human review.** `claims_unverified` and
   `stale_evidence` items are created with `status='needs_human_review'` and the
   `human-review` pseudo-handler — there is no automated handler to attach, so
   there is no path by which the system rewrites a factual claim. This is the
   layer's governing rule expressed as an absence, which is stronger than a policy.

6. **Opt-in per site, on `evidence_base` presence.** Both checks and the prompt
   block no-op without it. A site that has not asked for this is untouched, and
   `banned_claims` starts empty elsewhere. Fleet-wide enforcement on data one site
   has would be the classic over-reach.

7. **`context_terms` on facts (schema addition, not in the original spec).** A
   `gte` fact with a large value would otherwise blanket-support every smaller
   number on the site — "we support 12 clients" must not pass merely because 12 is
   less than the orchestration-records count. Non-exact tolerances now require a
   context-term match; a non-exact fact without terms degrades to exact, never to
   blanket support.

8. **V4 whitelist regeneration splits ownership: humans own the words, the machine
   owns the numbers.** Each fact carries a human-authored `writer_line` with a
   `{value}` placeholder, so audit caveats survive regeneration verbatim. A fact
   with no `writer_line` is omitted rather than auto-phrased — silence is the safe
   default for a claim nobody has worded. Regeneration is opt-in per site
   (`writer_block_managed`).

9. **V4 writes by compare-and-swap.** The pass rewrites a whole human-owned record
   from a copy read moments earlier — a lost-update hazard against a human editing
   the same row. The supersede is keyed on the row id read; a losing pass writes
   nothing and reports it. A lost refresh costs one scheduled tick, a lost human
   edit costs trust. *(Prompted by another thread's council objection about
   whole-config rewrites; the hazard is the same class in a different table.)*

10. **V4 treats its own SQL as untrusted.** The queries live in a JSONB column.
    Single statement, must begin with SELECT, no data-modifying keyword anywhere,
    executed in a read-only transaction with a statement timeout, single numeric
    scalar. The guard is a pure function so it is testable without a database —
    the first test attempt panicked on a nil handle, which was the design telling
    me the guard was in the wrong place.

## Corrections to the originating spec

Recorded, not absorbed — each of these says the spec was wrong or incomplete.

> **CORRECTION 2026-07-16 — §8's baseline was false.** The spec states "re-run V1b
> against the live leopardess site and expect zero findings (the site is currently
> clean — that's the baseline)". It was not clean. The first scan found nine
> banned-claim occurrences across four pages. Caught by running the scan rather
> than trusting the premise.

> **CORRECTION 2026-07-16 — the root cause was upstream of the pages.** The spec
> frames the problem as generation leaking past prompt rules. The deeper cause was
> that the *specs feeding the writer* still carried the fabrications — including a
> `writing_rules` entry instructing the writer to cite "least-privilege IAM
> policies" when discussing security. The writer was obeying, not hallucinating.
> Cleaning pages without cleaning specs would have regressed on the next rebuild.

> **CORRECTION 2026-07-17 — B3 is not a V3-only case.** The spec grades
> "2,767 Awards Won" a hard case for the LLM lane, to be documented if missed. It
> is caught deterministically at V1 by the banned *label* pattern, while the number
> lane correctly stays silent because 2,767 is registered. The lesson generalises:
> a false claim wearing a true number is often a banned-wording problem, not a
> numeric one.

> **CORRECTION 2026-07-17 — `checkpoint_for_review` was not usable.** The spec's
> HITL routing assumed the documented checkpoint action. It had never been
> registered, so any workflow naming it failed validation ("requires a topic").
> V3 uses `create_work_item` instead; the missing registry entry was added.
> A header comment's "Registration:" block is an intention, not a fact.

## What is deliberately NOT built

- **No auto-rewrite of factual content, ever** — including the tempting case where
  a live number merely exceeds the published one. That still changes copy, so it is
  human-gated (spec open question 3, answered in favour of human-gated).
- **No fleet-wide banned list.** Per-site until at least two sites have evidence
  bases (spec open question 2, unchanged).
- **No cadence for V3.** One LLM call per site per pass is a cost decision for the
  owner, not a default to assume.

## Open, for the owner

1. **One image build** activates V4; then apply the seed and switch on the daily
   pass. Nothing else is waiting on code.
2. **Cadence for the claims-auditor** — schedule it, or leave it manual.
3. **vetcomparison as the second evidence base**, with price claims as its
   deterministic lane (its rebuild already requires claim-licensing).
