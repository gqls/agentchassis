# 123 — content-creator's copy can never reach the claims assessor, and it fabricates statistics

**Filed** 2026-07-27 by the `gemini_content_provider` workstream · **Status** OPEN,
unowned · **Severity** HIGH if any content-creator output is published; **the
fabrication is confirmed live, not theoretical** · **Owner call needed before blog
or social output is published anywhere** · **Family** `bugs_closed/043`
(fabricated stats), `bugs_open/102` (the claims layer is `page_type`-blind)

---

## The observed fabrication

A live `content-creator-agent` generation on `gemini-pro-latest`, 2026-07-27,
212 words, topic "Why agent systems need to recover from their own errors". Verbatim:

> *"Industry data shows that large language models experience hallucination rates
> between 3% and 10% depending on the task."*

No source. Invented range. Attributed to "industry data", which is the shape that
makes it read as sourced. Nothing in the pipeline objected, because nothing in the
pipeline looks.

This surfaced by accident: I was verifying that the house voice had reached the
agent, read the output to check the prose rules, and the statistic was simply
sitting in it. **The voice check passed. The copy was still not publishable.**

## Why it can never reach the assessor as things stand

This is not "somebody forgot to call the checker". The checker **cannot be pointed
at this output**, because its input contract cannot be satisfied.

`UnverifiedClaimsCheck` is a *discovery check*
(`platform/orchestration/actions/discovery_checks/check_unverified_claims.go`). It
runs against a `DiscoveryCheckContext`, whose fields are (`registry.go:41-50`):

```go
type DiscoveryCheckContext struct {
	Ctx       context.Context
	DB        *sql.DB
	TX        *sql.Tx
	SiteID    uuid.UUID   // <- required
	Pipeline  string
	AgentType string
	BatchID   uuid.UUID
	Logger    *zap.Logger
}
```

and it resolves the evidence base by site:

```go
SELECT data FROM site_specs WHERE site_id = $1 AND aspect = 'evidence_base' AND is_current = true
```

**content-creator's output has no site, no page row, and no evidence base.** It is
free-standing text produced from a Kafka request and returned on
`system.agent.content-creator.responses`. There is no `site_id` to pass, so the
checker has nothing to scan *against* even if it were called.

Confirmed by grep: `internal/agents/contentcreator/agent.go` contains **zero**
occurrences of `validate`, `claims`, `evidence` or `fabricat`. The agent has no
validation of any kind.

## Why this went unnoticed until now

The whole fabrication apparatus — `043`'s evidence_base, the stat-field rules, the
prose checkers, migration 201's scalar-prompt rule — was built on the **site/page**
path. Everything it protects is a page belonging to a site. `content-creator` is a
standalone service on a different path, so it inherited none of it, and no coverage
report notices because coverage is computed per site.

It also does not write to `llm_call_log`, so it appears in no per-agent usage query
either (that is how it stayed invisible in the 30-day content-agent census that
prompted the house-voice work).

**Two independent invisibilities on the same agent.** Each on its own is
survivable; together they mean a content producer that no audit surface can see.

## What is NOT claimed

I have not established that content-creator output is currently published anywhere.
It returns text on a Kafka topic; whether a caller publishes it is a property of the
caller, and I did not trace them. **That question should be answered before
anything else here** — it decides whether this is urgent or merely important, and
it is a cheap check (grep the consumers of
`system.agent.content-creator.responses`).

## Fix candidates, ordered by what closes the door

**1 — Give the assessor an input shape it can accept for free-standing text.**
The check's dependency on `SiteID` is really a dependency on *an evidence base to
scan against*. A variant that takes `(text, evidenceBase)` directly would let any
producer call it, with the caller supplying whichever evidence base applies. This
makes the checker reusable rather than page-bound, and is the only candidate that
helps the *next* non-page content producer too.

**2 — Route content-creator output through a site context.** Require a `site_id` on
the request, load that site's evidence base, run the existing check. Smaller change,
but it forces every caller to have a site, which blog/social generation may not.

**3 — A text-level fabrication guard with no evidence base.** Refuse or flag
figures presented as sourced ("industry data shows", "studies find", "% of") when no
citation is present. Weaker than 1 or 2 because it cannot tell a true cited figure
from an invented one — but it is the only candidate that works when there is no
evidence base at all, which is content-creator's normal state.

**Not a candidate: adding "do not invent statistics" to the house voice block.**
The prompt already carried exactly that instruction in the darts brief earlier the
same day, and it held there — but a prompt rule is a request, not a gate, and 043
exists because prompt rules were not enough. The voice block governs how copy
*reads*; it must not be made to carry a truth guarantee it cannot enforce.

## How to verify a fix

Re-run the same request that produced the fabrication (topic above), and confirm the
"3% and 10%" claim is either removed, cited, or raised as a work item. **Then induce
the failing branch**: feed a prompt that invites a statistic and confirm the guard
fires — a clean generation proves nothing, since most generations do not invent
figures.

## Pointers

`docs/agent_docs/docs024_key_docs_latest/gemini_content_provider/` (how it
surfaced) · `bugs_closed/043` · `bugs_open/102` · `fabricated_stats_043/`
