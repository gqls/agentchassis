# PLAN — bugfix_427_event_render

Started 2026-09-02, by the session renamed `bugs_open/427` to resume/drive the fix.
Scope: the two-thirds of bug 427 that are NOT `news_feed_ingestion`'s populator
(candidate #1) — the correction path (candidate #2) and the render target (the
missing third of the whole fix, not separately numbered in the bug file but
implied by "so the calendar shipped empty").

## 0. How this lane's scope was set

Bug 427 was filed jointly by `calendar_component` and `boxingonline.com`, then
handed off. Within the first hour of this session:

- `feed_lane` (owner-named session) independently opened `news_feed_ingestion`
  and claimed candidate #1 (the populator) — see that lane's own PLAN. Verified
  their design was sound, caught one real defect before it shipped (`kind:
  "event"` isn't in `EvidenceFact.Kind`'s closed vocabulary — see NOTES), and
  deferred to their build rather than duplicating it.
- `gap_planner` (owner-named session) resolved the separate diagnosis this bug's
  §6 pointed at (`d6d350ec…`, why the site planner drops `entity-page`/
  `entity-directory` roles) and filed `bugs_open/428`. That is candidate #3's
  territory — not this lane's build, cross-referenced only.
- This lane's own research (reading `refreshCitationFact`'s dispatch conditions
  directly) found that candidate #2's "does the announcing article still say
  this" half needs **zero new code** — any fact with `source.citation`
  (which every fact candidate #1 registers carries, regardless of `kind`)
  already gets re-verified daily by the existing citation-refresh arm.

So this lane's actual deliverable narrowed to two things, both self-contained
and buildable without waiting on candidate #1's facts to exist:

1. A real gap found by tracing what CONSUMES the new fields: `composeWriterBlock`
   only substitutes `{value}`; a `writer_line` using `{event_date}`/`{venue}`/
   `{participants}`/`{broadcaster}` would have shipped those tokens
   UNSUBSTITUTED into the writer's prompt. Fixed.
2. The render target: nothing turns a registered event fact into anything a
   visitor sees. Built as a new `query.upcoming_events` resolver, using the
   estate's existing query-source + dependency-class mechanism (RFC_052)
   rather than inventing a new render path.

## 1. Decision: `composeWriterBlock` gains a third bucket, not a rewrite

`refresh_evidence_base_action.go`'s writer-whitelist regenerator classifies
each fact with a `writer_line` into NUMBERS (has a numeric `value`, `{value}`
substituted) or CAPABILITIES (everything else, carried **verbatim**, no
substitution — that's the bucket's whole contract: "assert without inventing
numbers"). An event fact has no numeric `value`, so without a third path it
falls into CAPABILITIES and any `{event_date}`-style token in its `writer_line`
reaches the writer's prompt as a literal brace.

**Fix**: a third "SCHEDULED EVENTS" bucket, gated on a non-empty `event_date`
(same style of gate as the numeric-value branch), with its own substitution —
`{event_date}`/`{venue}`/`{broadcaster}`/`{participants}` — where an unstated
field renders `"TBC"` rather than a bare brace (an event's venue/broadcaster is
routinely unstated at announcement time; `{value}` never has this problem
because it's always populated by construction). Participants join with `", "`
not `" vs "` — a boxing match reads fine either way, but the mechanism has to
work for any vertical with dated events (a panel, a hearing), not just fights.

Shipped: `f865153f8`. Mutation-tested (reverting the substitution call to a
plain pass-through fails the new test on the unsubstituted-token assertion).
Submitted: `d0442d50-e383-477f-9ed8-19eaaeea3d93`.

## 2. Decision: candidate #2 needs no new dispatch arm

Traced directly: `refreshOneSiteEvidence`'s per-fact loop dispatches by which
key exists under `fact.source`, not by `kind`. Every fact `VerifyAndRegisterCitationsAction`
registers — including these event-shaped ones, per `news_feed_ingestion`'s own
committed extension of its field pass-through list — carries `source.citation`.
The existing `if _, has := src["citation"]; has { refreshCitationFact(...) }`
branch (unconditional on `kind`) already re-fetches the announcing article and
re-checks the quote, daily, via the existing `evidence-freshness` scheduled
task. **No `source.feed_item` marker, no new `refreshFeedItemFact` function.**
Building one would duplicate a check that already runs, and would be exactly
the "declared, never read" shape this whole bug is about.

**What is NOT covered, named as a residual rather than silently dropped**: a
correction published as a SEPARATE, later article (a postponement, a venue
change reported fresh rather than by editing the original page). That needs
same-event matching across feed items — a component that does not exist and
whose false-positive behaviour would need its own measurement first. Filed as
a follow-up, not built here. `content_feed_items.duplicate_of` is the column
that would eventually get a writer for this; it stays unwritten until then.

## 3. Decision: the render target is a `query.*` resolver, not a new action

The estate already answered "a component shows rows derived from a store that
changes" once, for news items, and the answer generalised again for
directories (RFC_052): a component's `input_schema` declares a field sourced
`"query.<name>"`, `queryresolve` resolves it into `page_components.content_data`
at plan/render time, and a producer that changes the underlying store queues a
`section_data_resolved` `page_rerender` for every consuming page via
`queryresolve.ConsumerPages`. `news_items.go`'s own header explains why the two
alternatives were rejected there and apply identically here:

- **HTML-patching `rendered_html` directly** — rejected by mechanism 003; a
  scoped rerender regenerates from `html_template` + stored fields and would
  wipe it.
- **A Go action writing a JSON file + client-side fetch** (the news feed's
  original design) — puts the page back in the "runtime fetch" bucket
  `experience_loop`'s own check flags this page for being in.

So: `query.upcoming_events`, reading a site's own `evidence_base` register
(site-scoped — confirmed `directory_entities`/`directory_claims` is a
different, global, cross-site registry for an unrelated product; pure name
collision with the "entity-directory" page ROLE, not a route for this). Facts
with a parseable, non-past `event_date` are selected, sorted ascending, and
projected with HTML-escaping (`text/template` does not auto-escape) — an
unparseable date is excluded and logged, never guessed. A new `DepEvidenceBase`
dependency class is declared whole-item-set (a new event changes which
fixtures exist, same reasoning as `latest_news`/`news_archive`), and
`queueEvidenceBasePageRerenders` (a direct cousin of `queueNewsPageRerenders`)
fires whenever `refresh_evidence_base` actually writes a changed register —
this is also candidate #2's PROPAGATION half: a human's correction via the
`stale_evidence` item needs this to actually reach the page.

Shipped: `da2ab0d44`. Both lockstep tests
(`TestSourceDependenciesMatchTheResolvers`,
`TestEveryRegisteredBaseDeclaresItsDependencies`) pass with the new base
registered. Mutation-tested: the past-date filter, the escaping, and the
dependency-class constant each caught a broken version when disabled.
Submitted: `08f56b7e-61e4-42d1-a3b6-13d700dd833c`.

## 4. What is explicitly NOT built here

- **The component that actually declares `source: "query.upcoming_events"`**
  on `/tools/fight-calendar/index.html`. This resolver returns an empty array
  on every site today (zero facts carry `event_date` fleet-wide as of
  2026-09-02) and no component names the source yet — there is nothing to
  place it on top of until candidate #1's workflow-config half (part 2, still
  gated on an image roll) produces real facts. Placing an empty list early
  would look like "fixing" the emptiness by hand, which it isn't.
- **Which existing component (if any) already accepts a query-sourced `items`
  array with event-shaped columns**, vs. a small new `event-list` component
  being needed — an open question, not assumed either way (see RUNBOOK for the
  check).
- **`entity_ids`/`duplicate_of`** — confirmed orphaned/unwritten; not this
  fix's job to populate (see NOTES for the reasoning, independently reached by
  both this lane and `news_feed_ingestion`).
- **Candidate #3** (`entity-directory` page role) — `bugs_open/428` now holds
  the diagnosed root cause (the planner reads `recommended_page_types`
  correctly and *deliberately* defers the roles, exercising a "final say"
  license the prompt grants it — not a wiring bug). Not this lane's fix.

## 5. Phasing from here

1. **Done**: writer-block substitution (§1), the resolver + producer hook (§3).
   Both awaiting council verdicts.
2. **Watching**: `news_feed_ingestion`'s part 2 (workflow-config wiring an
   image roll away) — once it lands and produces boxingonline's first fact,
   the render target has something real to point at.
3. **Next, once facts exist**: resolve the open component question, write the
   (likely small) migration placing the component on the shipped tool page
   (precedent: migration 267 for inserting a component into an already-shipped
   page), submit that separately (DB migration, council-scope on its own).
4. **Follow-ups, each its own bug**: cross-article correction matching (§2);
   `entity_ids`/`duplicate_of` — define a target or drop.

## 6. Status, 2026-09-02 (later same day) — see bugs_open/427 §10/§11 for the live account

Fresh chassis build (`ebf27c60377f`) confirmed live for `agent-chassis` and
`core-manager`, carries all of §1-3's code plus feed_lane's real facts. Council
REVISE on the resolver (compliance HIGH: no evidence gate before rendering
real-world scheduling claims) — fixed same session (citation-url+quote gate,
travelling disclaimer, settled `site_specs.evidence_base` naming, checked and
documented why the existing `evidence-chart` component doesn't fit) and
resubmitted on the same correlation. **This file is not being kept as the
live status record going forward — bugs_open/427 §10/§11 is; read there.**
