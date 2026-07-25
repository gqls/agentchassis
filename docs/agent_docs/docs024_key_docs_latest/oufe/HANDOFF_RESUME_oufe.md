# HANDOFF / RESUME — oufe.com + oxenunity.com

Cold-start entry point for this workstream. Written 2026-07-25.
Read `PLAN_2026-07-25_oufe.md` first for what the site IS and why; this file is
only "what state is it in and what do I do next".

## State at handoff

| Thing | State |
|---|---|
| Workstream docs | **Done** — standing five + briefs + disclaimer draft |
| oxenunity.com page | **Built and in B2**, unreachable at its domain (see owner list) |
| oufe.com Cloudflare wiring | **Already done** — nothing needed, it serves the moment content lands |
| oufe.com site row | **Created** — `a0d7f1ae-f37e-4ea5-b30c-9012d1d14f39` |
| oufe.com evidence_base | **Seeded** — 18 banned patterns, **0 facts** |
| oufe.com imagery_style_guide | **Seeded** |
| oufe.com Tier-3 submission | **CONSUMED — cascade running.** corr `e916f41b-a534-4b12-883f-411312ee7ad8`. Both briefs persisted (3,696 / 2,828 chars); `needs_domain_research` triaged |
| Thames Water evidence (V5) | **Not started** |
| Waterfall tool | **Authored + arithmetic verified**, insertion SQL prepared, **not applied** |
| Legal pages | **Not built**; wording drafted, needs owner approval |

## First thing to check — where the cascade got to

```sql
SELECT ss.aspect, ss.source_agent, ss.is_current
  FROM site_specs ss JOIN sites s ON s.id=ss.site_id
 WHERE s.domain='oufe.com' ORDER BY ss.created_at;

SELECT wi.item_type, wi.status, wi.handler_agent, LEFT(COALESCE(wi.error,''),100)
  FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
 WHERE s.domain='oufe.com' ORDER BY wi.priority;

SELECT name, page_type, build_status, jsonb_array_length(sections) AS n_sections
  FROM pages WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com');
```
Expected order as it fills: `classification` → `strategy` → `briefing` →
`site_plan` / `design_intent` / `resolved_composition`, then pages.

**If it looks stalled, check the queue before concluding anything:**
```bash
kubectl -n kafka exec personae-kafka-cluster-combined-pool-prod-0 -- \
  bin/kafka-consumer-groups.sh --bootstrap-server localhost:9092 \
  --describe --group generic-requests-group
```
The original submission sat unconsumed for ~28 minutes behind another session's
20KB council-gate run — a named-cause instance of `bugs_open/030` head-of-line
blocking, resolved the instant that council completed. **Never re-fire on an
absent row**; a duplicate queues behind the same blockage and then does the work
twice. Note also that `processed_messages` is *not* a reliable
was-it-consumed oracle (it records a narrower path) — consumer-group lag is.

## Then, in order

1. **Watch the cascade.** classifier → strategist → briefing → planner → pages.
   Blocker reasons at `validate_page_content` are **not recoverable from the DB** —
   watch the chassis log live or you will not learn why a page died.
2. **Turn the news feed off** once `classification` lands: deep-merge
   `content_features.news_feed.recommended = false` (supersede-then-insert). This
   is deliberate (PLAN §C7), not an oversight — the classifier will read the site
   as `finance` and seed generic market-news keywords.
3. **Thames Water evidence, before any dossier prose.** Fire `evidence-researcher`
   at the current state of the restructuring (judgments via the National Archives
   caselaw service, Ofwat publications, company releases). This is **V5's first
   real end-to-end exercise** — it was activated on v1.0.1140 and its blocker
   (bug 047) closed, but the smoke run was never repeated. Record what happens
   either way; a failure here is a platform finding worth more than the facts.
   Fallback: manual research, hand-authored citation facts.
4. **Then the dossier content.** V2 gives the writer only the verified facts.
   Sections with no verified fact say so. Watch `bugs_open/073`: a component with
   a required stat field plus an honest empty value fails the whole page build —
   prefer components that do not demand numbers.
5. **Legal pages** via the migration-182 pattern once the owner approves wording:
   hand-written content, `rebuild_policy='owned'` + permanent component lock,
   `rendered_html` written in the same migration. Also defuses `bugs_open/053`.
6. **The tool**: apply `PREPARED_tool_insert.sql` — but only after the site plan
   exists, or the reconciler can clobber the page. Read its header first; it
   carries three landmines in the comments.

## Owner list (blocking "live", not blocking building)

1. **Move oxenunity.com to Cloudflare** and bind the portfolio-sites Worker route
   to `oxenunity.com/*` and `*.oxenunity.com/*`. It is at porkbun today, serving a
   302 to a parking page. The built page is already in B2 waiting.
   **oufe.com needs nothing** — already wired, verified.
2. **Approve the disclaimer wording** (`DRAFT_disclaimer_for_owner_approval.md`).
3. **Confirm the contact email** (`oufe@contactforsales.com` is seeded).
4. **Respond to the challenge in PLAN §C1** — the owner ruled "radar first, it is
   lowest risk"; this workstream argues it is the highest risk first move and has
   built the dossier-plus-tool path instead. He has not yet answered that.

## Rails that must not be relaxed

- **No figure enters a brief, a spec, an identity or a content_direction.** Only
  the evidence register, with a source. A number in a spec is a *given* and beats
  every writer-side rule — that is how invented figures were written back over
  correct ones on another site with all guards live (043).
- **A clean claims report on this site means almost nothing.** The deterministic
  number scan has no finance vocabulary and excludes currency outright. The real
  layers are the writer whitelist, the banned patterns, V5 citations, and a human
  reading it.
- **Never publish a figure about a real company without a source URL and a
  capture date** — the vetcomparison rails, inherited whole.
