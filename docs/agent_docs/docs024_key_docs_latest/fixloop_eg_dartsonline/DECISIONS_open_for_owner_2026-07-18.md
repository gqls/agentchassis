# Open decisions for the owner — diagnosis/fix tool, 2026-07-18

*Two standing decisions surfaced by this week's work. Each is a genuine owner
call (a cost/quality trade, or whether to spend effort). Options + my
recommendation for each; nothing is actioned.*

---

## Decision 1 — roster-wide model: move the proposer + reviewers to Sonnet 5?

**Now:** `diagnose-agent` runs `claude-sonnet-5` (proven this week). The
`fix-proposer` and all 9 council reviewers still run `claude-sonnet-4-6`.

**For moving:** Sonnet 5 is materially better at code review and bug-finding —
directly the reviewers' job (the bug-historian's value this week was judgement;
better judgement compounds). Pricing is near-parity (Sonnet 5 intro $2/$10 per
MTok vs 4.6 $3/$15), so cost is not a barrier.

**The trap that MUST be handled together with the model swap** (learned on
diagnose-agent this week): on Sonnet 5, omitting the `thinking` field runs
**adaptive thinking ON** (4.6 ran it off), and thinking spends from the SAME
`max_tokens` budget as the output. The reviewer steps are set to
`max_tokens: 3000` (verdicts observed at 275–1290 tokens — comfortable at 4.6).
Move to Sonnet 5 *without* raising that, and a reviewer that thinks 2k tokens has
~1k left for its verdict JSON → a **truncated verdict** — the exact BUG A class,
inside the reviewer of the BUG A fix. Also: Sonnet 5's tokenizer runs ~30%
heavier.

**Plumbing note (good news):** `fix-proposer` has **no root `ai_service`**, so the
per-step `ai_service` blocks are read cleanly — BUG B (root shadows step) does
NOT bite here; model + max_tokens both live on the step and take effect.

**Recommendation: YES, but as ONE correct migration, not a bare model swap.**
The migration must, together: (a) set `claude-sonnet-5` on the proposer + reviewer
steps; (b) raise reviewer `max_tokens` 3000 → **≥8000**; (c) apply to **BOTH**
councils (`fix-proposer` AND `council-gate`) in one go per the standing rule; (d)
be **patch-style/idempotent**, never a whole-object re-seed (the config-clobber
finding). Timing: fine to do now — the trap is fully known and cheap to avoid,
and the reviewers are actively used, so better review is high-value. If you'd
rather not add config churn while BUG A/B fixes are landing, defer until those
PRs merge; no urgency either way.

---

## Decision 2 — build the CI guard the bug-historian keeps flagging?

**The residual** (raised by the historian on BUG A, twice): the fix decodes
`stop_reason` in both current provider adapters, but **nothing prevents a FUTURE
third provider being added without the guard** — "an architecture-level
observation, not a blocker for this change." The same generic-mechanism warning
underlies the whole silent-truncation class.

**Shape if built:** a Go test (or a `discovery_check`) asserting that every
provider client's `GenerateText`-equivalent decodes its stop/finish signal and
fails loud on truncation — so a new adapter that skips it fails CI, not
production. Small (one test), preventive (pays off only when a provider is
added).

**Recommendation: YES, but fold it into the BUG A fix PR (the `/bugs_open/008`
thread) — do not make it a separate workstream.** The guard should ship in the
same PR as the mechanism it guards, so the class is closed and re-closed in one
place. It is a natural, well-scoped **feature-builder pilot** if that thread
wants a first real target instead. Low urgency; it prevents a regression that
requires someone to add a provider first. Action: flag it to the 008 thread as a
"bundle this test into your PR" note (I can do that if you approve).

---

## If you want me to action either
- **D1:** I'll write the patch-style dual-council migration (model + max_tokens,
  both councils) and prove it on a re-grade run before it's called done.
- **D2:** I'll add the note to `/bugs_open/008` so the fixing thread bundles the
  guard test into the BUG A PR.
Both are owner-go; each spends nothing until you say so (D1's proof run spends
credits).
