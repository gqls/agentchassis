# 093 — the stat-claim audit is generic but has exactly ONE guarded call site; the re-render path publishes unchecked

**Filed 2026-07-26** by the bugfix-043 thread, **at the council gate's insistence and to its
credit**: this was named in my own submission's `risks` block as a known gap and I closed
`043` without filing it. Seat `bug_historian` gated the review on precisely that —
*"documents this gap accurately but does not propose closing it or even filing it as a
tracked follow-up"* (submission `569241fb`, severity **high**). It was right, and the
pattern it matched is a real one in `016b`: a mechanism made generic, then guarded at one
call site while the others stay open.

**Status:** OPEN, not started. Cause is structural and fully known — this is a scope gap,
not a mystery, so there is nothing to diagnose.
**Belongs to:** the fabricated-stats lane, `bugs_closed/043_…generated_page_copy_invents_
quantitative_claims.md` (resolve **by slug** — `043` is one of the ambiguous numbers).

---

## The gap

`datahelpers.ExtractStatClaims` / `ScanStatClaims` / `LintStatUnits` are generic: give them
a component name and a `content_data` map and they will audit the figures in it. Live in
`v1.0.1170`. But they are invoked from exactly one place —
`validate_page_content` check 9 (`platform/orchestration/actions/validate_page_content_stats.go`),
which runs only on the **build** path (`page-build-handler` → `validate_content` →
`save_sections`).

**The re-render path never runs that gate.** `rerender_page_sections_action.go` renders
from stored `content_data` with no LLM and no validation step
(`docs/agent_docs/sql_for_agents/034_page_rerender_agent.sql` has no `validate_content`).
So:

- a figure that was already stored, however it got there, is re-published on every
  re-render without ever being compared to the site's register;
- `043`'s own update (b) class — a junk unit suffix persisted into `content_data` — is
  **specifically** unreachable by a build-time gate, because the suffix is a `source:static`
  fallback that gets resolved and persisted, and a page carrying one may be re-rendered for
  years without a writer pass.

This is not hypothetical for this lane. `043` recorded four pages across three sites
carrying persisted junk suffixes, and the counter-correction in `bugs_closed/073` measured
`aao/index` being **re-rendered three times in one morning** while the writer path was
blocked — i.e. the fabrication republishing itself indefinitely through exactly this route.

## Why it is not simply "add the gate to the re-render path"

The re-render path is deliberately cheap and LLM-free; bolting the full
`validate_page_content` onto it would drag in seven other checks, several of which
(placeholder scan, meta-commentary, contamination) are about freshly *generated* prose and
have nothing to say about a carried-forward render. It would also give the re-render path a
new way to fail, on content it did not author — the same trap as `073`, where a page became
unbuildable for telling the truth.

## Fix candidates (unranked)

1. **Extend the post-deploy discovery check instead.**
   `discovery_checks/check_unverified_claims.go` already sweeps deployed pages and already
   reads `page_components`; adding a `content_data` pass there covers the re-render path,
   hand-edits, and every page that predates the gate, in one place. **It reuses the existing
   `claims_unverified` item type, so it incurs no `verifier_coverage_test.go` obligation** —
   the same reason check 9 was cheap. Cheapest real coverage, and the natural second call
   site. Note it detects after deploy rather than preventing.
2. **A minimal stat-only gate on the re-render path** — call `LintStatUnits` and
   `ScanStatClaims` from `rerender_page_sections_action.go` and route findings to a work
   item rather than failing the render. Prevents rather than detects, but adds a failure
   surface to a path chosen for having none. If taken, findings must NOT block the render.
3. **Do it at the source instead**: `LintStatUnits` needs no evidence base, so a one-off
   fleet sweep over stored `content_data` would clear the persisted junk suffixes that make
   the re-render path dangerous, after which (1) alone is sufficient upkeep.

**Prefer (1) + (3).** (3) is a query and an UPDATE and closes the known instances; (1)
stops the class recurring without giving a deliberately-simple path a new way to fail.

## How to verify a fix

Do **not** grade this on a build. Take a page with a persisted junk suffix (`043`'s update
(b) lists four), re-render it *without* a writer pass, and confirm the finding is raised:

```sql
-- pages whose stored content_data carries a magnitude marker followed by a unit
SELECT s.domain, p.name, e.k, e.v
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id,
LATERAL jsonb_each_text(pc.content_data) e(k,v)
WHERE e.k ~ '_suffix$' AND e.v <> '' ORDER BY 1,2;
```

Then confirm a `claims_unverified` (or equivalent) item exists for that page afterwards.
A green build proves nothing here — **the build path is the one that already works.**

## Related

- `bugs_closed/043` — the parent case; its § "Final verification" is explicit that one page
  on one site is not the fleet sweep.
- `bugs_closed/073` — where the re-render-vs-rebuild distinction was measured, after one
  thread mistook a re-render for a build and another corrected it.
