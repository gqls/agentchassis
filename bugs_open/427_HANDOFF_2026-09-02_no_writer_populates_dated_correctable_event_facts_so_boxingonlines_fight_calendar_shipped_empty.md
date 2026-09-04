# 427 — nothing on the estate turns a confirmed real-world event into a dated, correctable fact, so a "calendar" tool page has a hero banner, an essay about itself, and zero fixtures

Filed 2026-09-02, jointly by the `calendar` session (diagnosis, fleet-wide `evidence_base`
census, the three-writers read) and `boxingonline.com` / `site_delivery_and_editor`
(the motivating site, the fleet-key count that started the thread, the owner's ruling).
Reached from two directions in the same day, independently: the calendar lane arrived by
asking "why is boxingonline's calendar tool empty", the delivery lane arrived by asking
"why did the site not get what its own strategy spec asked for" — both land on the same
missing mechanism. **Status: OPEN, unowned** — the feed lanes that would build the fix
(`bugfix_316`, `news_feed_pooling`, `bugfix_410`) have no live session as of today.

**Severity: MEDIUM-HIGH.** Blocks delivery of a paid site (owner ruling: fixed before
delivery, see §6) and the underlying gap — nothing populates `evidence_base` with dated
facts — is fleet-wide, not specific to this site (§3).

## 0. First-hand verification, stated per CLAUDE.md's 2026-07-31 ruling

This asserts a structural, cross-cutting root cause without a `090` diagnosis-loop run.
Substituting: every count below was queried live against the DB in this session today: 2026-09-02
(§3, §4), or inherited from the sibling `boxingonline.com` session's own live measurement
today. One figure carried from an earlier session (the news feed's "three zeros",
2026-08-19) was **deliberately re-queried fresh rather than cited stale** — and it had, in
fact, moved: see §4's correction. The writer audit (§5) is a direct read of the Go source
and its own header comments, not inferred from behaviour.

## 1. The symptom

`boxingonline.com`'s paid brief's own words, quoted by the customer: *"The calendar will
be populated with real upcoming events to start with."* What shipped,
`/tools/fight-calendar/index.html`: `[MEASURED 2026-09-02]` exactly two components,
`hero-tool` (badge "Fight calendar", headline "Every upcoming boxing fight, listed and
dated") and `generic-text-block` (2,000+ characters opening "The calendar above pulls
together the fights worth building your weekend around… in one place"). **There is no
third component and no fixture data anywhere on the page** — no date, venue, fighter or
broadcaster field exists. The prose describes a calendar that isn't there.

Independently confirmed by `experience_loop`'s nightly page-behaviour check: this is the
**only** tool page fleet-wide, out of 320, classified as serving no control, no inline
data and no runtime fetch (four others were never built at all, in a separate bucket).

The owner's own diagnosis, verbatim, via the `boxingonline.com` session: *"The research
agent should have researched what's on and that is what should have appeared on this page
and the calendar."* His related ruling on the site's fighter-comparison tool: it "should
contain detailed, fact checked information that prefills the form for the comparisons for
the boxers" — the same shape of gap, one tool over.

## 2. What this is NOT

- **Not `bugs_closed/381`'s defect class.** 381 was a planner choosing prose-only
  components when list/table-capable ones existed. This page's components aren't
  under-expressive — `hero-tool` + `generic-text-block` are exactly what a tool page's
  hero-plus-explainer shape calls for. The gap is upstream: nothing ever handed either
  component a fixture to render.
- **Not a job for `period-calendar` (VIZ-017, this lane's own component).** It refuses
  dates and numbers by design (605's rule 1: a period is a recurring NAME, never a date).
  A fight calendar is one-off, dated, real-world events — the opposite shape. Placing it
  here would be wrong, not merely insufficient.
- **Not a `news_editorial_features` job.** That lane was asked first (name-match on
  "the tools need real data") and declined correctly: they own editorial FEATURE pages
  built from `content_feed_items`, not feed ingestion, normalisation, or relevance, and
  had never touched this site.

## 3. Root cause, end 1: `site_specs.aspect='evidence_base'` has no populator, fleet-wide

`evidence_base` is the estate's shared, per-site, structured-and-cited fact store —
`facts[]`, `allowed_entities[]`, `charts[]` — read by the writer and by the claims-gating
pipeline. `[MEASURED 2026-09-02]`:

```sql
SELECT count(*) FROM sites;                                            -- 54
SELECT count(DISTINCT site_id) FROM site_specs
 WHERE aspect='evidence_base' AND is_current;                          -- 20
```

**34 of 54 sites (63%) have no current `evidence_base` row at all.** Of the 20 that do,
the fact-count distribution is skewed and mostly not "a little thin" but "essentially
none":

| facts in the array | sites |
|---|---|
| 0 | 3 |
| 1–3 | 2 |
| 4–10 | 5 |
| 11+ | 10 |

`boxingonline.com` sits in the thin end: 2 rows, 3 facts total, 0 `allowed_entities`.

> **Correction, 2026-09-02 (resuming session, cross-checked against `boxingonline.com`'s
> own live re-measurement, identical query shapes).** The 3-facts figure above is now
> stale — it moved *after* this bug was filed, for a reason worth recording rather than
> silently re-querying past. `[MEASURED 2026-09-02]`:
> ```
> source       | created_by                       | created_at           | is_current | n_facts
> order_intake | seed_build_queue                 | 2026-08-31 12:21:12  | f          | 2
> operator     | site_delivery_and_editor-session | 2026-08-31 15:54:48  | t          | 1
> ```
> The seeded row was superseded the same day this bug's underlying gap was being
> diagnosed, by the `site_delivery_and_editor` lane acting on the owner's privacy ruling
> (`bugs_open/420`): one of the two seeded facts was the text "Enquiries reach
> aaa@designconsultancy.co.uk" — the owner's own billing address, registered as a
> publishable business claim — and was removed because the owner ruled his personal
> address off the site entirely. **This makes the case stronger, not weaker**: of the two
> facts this site was ever seeded with, one was a misregistered contact detail, not a
> fact about the fight calendar at all. The live corpus is one fact, about the business
> name, and it still contains zero dated event facts — the root cause in §3 is unaffected
> by which of {3, 1} is quoted, but the smaller number is the current one and the larger
> one should stop circulating.

**The combined figure is the one that actually settles §3's question, and it was derived
independently twice — once here, once by `boxingonline.com`, matching exactly on
re-check** `[MEASURED 2026-09-02, both sessions, identical query shapes]`: folding the
34 no-row sites in with the 20-that-have-a-row distribution above, **37 of 54 sites (69%)
hold ZERO facts, and 42 of 54 (78%) hold five or fewer.** Only 12 sites carry a corpus
worth the name, and exactly one exceeds 50 facts.

That is the decisive form of the backfill-vs-pipeline question this section opened with.
63% with no row could still be explained away as "those sites are young, or the pipeline
just hasn't reached them yet." 78% at five facts or fewer, concentrated almost entirely in
a single outlier site, cannot be — it is the shape of a mechanism that has never routinely
run, with a handful of hand-populated exceptions (the dartsonline PDC facts, migration
`494`, are one of them). **`boxingonline.com` is not the anomaly here. It is the typical
case, and the typical case is empty.**

> **Correction to a figure already in circulation on this bug.** The `boxingonline.com`
> session's own message describing this bug quoted "facts (444 sites)" from a fleet key
> census. `[MEASURED 2026-09-02]` **444 is a ROW count, not a site count** — it is
> `SELECT count(*) FROM site_specs WHERE data ? 'facts'`, which includes every historical
> and superseded row, not distinct sites: `SELECT count(DISTINCT site_id) FROM site_specs
> WHERE aspect='evidence_base' AND data ? 'facts'` returns **20**. Worth fixing at the
> source before it propagates further — it reads as "444 sites have facts", which is 22×
> the true figure, and would have made this look like a boxingonline-specific gap rather
> than the fleet-wide one it is.

**The writer-side effect this produces on `boxingonline.com` specifically**, traced by
the sibling session and consistent with `evidence_base` holding almost nothing: the site's
`about.html` renders editorial-POLICY prose (*"How we cover it — … A preview that says a
fight 'could be great' tells the reader nothing…"*) that paraphrases the research spec's
own `lessons.avoid[]` list rather than following it — a rule ABOUT the writing, present in
context, emitted AS the writing (filed separately,
`copy_quality_two_stage/CONTRIB_2026-08-31_…`). An empty evidence base does not cause that
mechanism, but it is the same symptom family: the writer reaching for whatever is actually
in its context when the thing it should be grounded in (facts, in one case; a fixture
list, in this one) isn't there.

**Why: read the three writers, don't infer from the empty rows.**

1. `seedCustomerIdentity` (`seed_build_queue_action.go:316`) — seeds a **minimal register**
   at build-queue time, `INSERT ... WHERE NOT EXISTS a current row` (guarded,
   insert-if-absent only; deliberately never merges or supersedes — council `7e3dd082`
   guards this on purpose, see the action's own header). Guarantees a row can exist; does
   not populate it with anything beyond minimal.
2. `refresh_evidence_base` (`refresh_evidence_base_action.go`, registered in
   `registry.go:662`) — **refreshes EXISTING facts only**: staleness checks, drift
   detection, re-verifying citation values against tolerance. It has no code path that
   invents a new fact from nothing; its own name says so and the code confirms it.
3. `VerifyAndRegisterCitationsAction` (`evidence_citations.go:181`) — the one genuine
   automated writer of NEW facts: registers a `citation` fact "in the site's evidence_base
   (created if the site has none)" when a citation is verified. **Opportunistic**: it
   registers whatever the writer happened to cite, not a proactive "go find what's
   currently scheduled" step.

**None of the three is, or is meant to be, "research what's currently happening and write
it down as a dated, correctable fact."** (1) is a stub. (2) only maintains what's already
there. (3) only captures citations content already contains. The owner's diagnosis — "the
research agent should have researched what's on" — names a step that does not exist in any
of the three: `research-agent`/`evidence-researcher`/`content-researcher` do one-time,
point-in-time landscape research (competitor analysis, vertical conventions), and nothing
downstream of them turns "what's currently on" into a fact this store, or any store,
retains.

`[UNMEASURED, left to the fixing thread]`: whether `VerifyAndRegisterCitationsAction`
fired at all during boxingonline's build and simply had nothing offered to it, or never
ran on this build path. The three-writer read above is sufficient to explain the
STRUCTURAL gap regardless of which combination fired here — establishing which one is a
call-graph trace this filing did not do.

## 4. Root cause, end 2: the news feed has no path from an item to a structured event

`content_feed_items` is where a promoter's fight announcement, or its equivalent for any
vertical, would first arrive as data — `feed-triage` already tags real volume with topics,
credibility and source-tier. But two fields declared for exactly this purpose are written
by nothing:

```sql
SELECT count(*) AS total,
       count(*) FILTER (WHERE entity_ids        IS NOT NULL) AS entity_ids_set,
       count(*) FILTER (WHERE duplicate_of      IS NOT NULL) AS duplicate_of_set,
       count(*) FILTER (WHERE published_page_id IS NOT NULL) AS published_page_id_set,
       count(*) FILTER (WHERE relevance_score   IS NOT NULL) AS relevance_score_set_control
FROM content_feed_items;
```
`[MEASURED 2026-09-02]` → **14,013 total | entity_ids 0 | duplicate_of 0 |
published_page_id 15 | relevance_score 12,281 (control — the search shape works)**.

> **Correction to the inherited figure.** The `news_editorial_features` lane measured this
> as "three zeros" on 2026-08-19 (`entity_ids`, `duplicate_of`, `published_page_id` all 0
> of 10,855). Re-run fresh today rather than cited stale: `entity_ids` and `duplicate_of`
> are **still** genuinely 0 out of a now-larger 14,013 — that half holds. `published_page_id`
> has **moved to 15** since 08-19. Traced: all 15 belong to exactly two source clusters —
> the darts PDC calendar-density feature and a robotics-industry feature — the two
> hand-built `NEWS-020` feature pages (`news_editorial_features` lane's own deliberate,
> manually-authored work), not a general automated news-item-to-page pipeline. The
> structural claim is unaffected — there is still no automated path from an arbitrary
> confirmed-fight item to a structured record — but the number itself would have been
> wrong to carry forward unchecked, and it is exactly the kind of addition-by-small-amount
> that this estate's own memory warns reads as "still true" when it silently isn't.

**So even where a confirmed-fight article lands in the feed** (and the extraction pipeline
that tags topics/credibility runs on it at volume), **there is no step that turns "this
article confirms a fight" into a structured, dated, correctable record** — the shape
`entity_ids` was declared for and never given.

**Method note, flagged by `boxingonline.com` as worth stating rather than leaving
implicit.** Re-running the inherited "three zeros" instead of citing them, finding one had
moved, tracing why, and recording the movement anyway is a live instance of this estate's
own standing rule that a census goes stale by ADDITION and reads as current forever — a
bug arguing "do not trust an inherited figure" (as §3's `444`→`20` correction also does)
would be self-refuting if it then inherited one without checking. Both corrections in this
file were caught the same way: re-derive before citing, not after being contradicted.

## 5. Why this is one bug, not two

Both ends are the same missing capability seen from either side. `evidence_base` is the
STORE with 444 historical rows and (correctly counted) 20 sites actually using it — what's
missing is a writer that puts DATED, CORRECTABLE event facts into it. The news feed is the
most plausible SOURCE for those facts — a promoter's confirmed-fight announcement arrives
there first, as data, already — but nothing extracts a structured record from it. Fixing
either end alone leaves the other stub: a populator with nothing well-shaped to populate
from, or an extractor with nowhere durable to write. File and fix together, or the same
question returns in a month under a different site's name.

## 6. Constraints on the fix, from the owner and from the site

- **Owner ruling (via `boxingonline.com`, 2026-09-02): fix before this site is delivered.**
  There is real time — not an emergency — but the cut-line is explicit, and it makes this
  bug's status sharper than "open": it is **currently blocking a paid deliverable**, not
  queued behind one. `boxingonline.com` has told the owner this bug is where the one
  unstaffed piece (the feed lanes) now lives.
- **Do not invent a parallel store.** `evidence_base` already has 444 historical rows and
  is read by the claims-gating pipeline and the writer; a second corpus for "dated facts"
  specifically would just relocate this exact bug.
- **The output shape wants to be closer to `entity-directory`, not a content component.**
  `boxingonline.com`'s own `strategy` spec (`recommended_page_types`) already asked for
  exactly this: *"An event directory — one page per major upcoming fight — gives each bout
  a permanent URL with full details: fighters, date, venue, broadcast, undercard, and a
  brief preview."* The planner emitted neither `entity-directory` nor `entity-page` for
  this site — diagnosis correlation `d6d350ec-e16b-4792-9282-ca5155369791` asked why.

  > **Update, 2026-09-02 (resuming session): that diagnosis has since completed —
  > `status: UNVERIFIABLE`, `stopped_by: iteration-cap`, "no fix proposed."** Read at the
  > artefact (`site_work_items.result`, not the item's `status='complete'`, which is the
  > item's lifecycle, not its verdict). What it DID confirm, citation-backed:
  > boxingonline.com's strategy names both `entity-page` and `entity-directory` with full
  > per-type reasoning; the site's current plan carries neither role; other sites' current
  > plans do carry both, so the roles are not globally unproducible; and
  > `page_role_validator.go`'s `ValidateRoles`/`normaliseRole` actively recognise and
  > preserve both role names rather than dropping them, so a downstream validation drop is
  > ruled out. What it did NOT reach before hitting the iteration cap: whether
  > `recommended_page_types` is wired into the `build-site-planner` plan-writing step's
  > prompt/input at all — `WriteSitePlanAction` only consumes the LLM's already-produced
  > plan, never reads `recommended_page_types` itself, so the gap is equally consistent
  > with the planning LLM never being handed that reasoning in the first place. A sibling
  > diagnosis on `bugs_open/419` (a planner zero-section page) hit the same iteration cap
  > the same day. **Fix candidate #3 stays blocked on that gap closing, owned by whoever
  > takes the planner** — do not duplicate the diagnosis, and do not treat UNVERIFIABLE as
  > either a confirmation or a refutation.
- **Dates get corrected, not just added.** The research spec's own `lessons.avoid[]`
  already names this: *"Stale calendar entries — a wrong fight date actively harms
  readers."* Whatever writes a fixture must also be able to revisit and correct it —
  `refresh_evidence_base`'s staleness/drift machinery is the closest existing analogue and
  is worth reusing rather than re-deriving, even though it currently only refreshes facts
  it's handed, never creates them.

## 7. Fix candidates — named, not decided

1. **A new extraction step, downstream of `feed-triage`, that fills `entity_ids` (or a
   comparable typed field) for a confirmed-event item** — date, venue, participants,
   broadcaster where stated — and a corresponding writer that registers it as a dated
   `evidence_base` fact (extending `VerifyAndRegisterCitationsAction`'s "create if the site
   has none" pattern to a proactive rather than opportunistic trigger).
2. **A revisit/correction path reusing `refresh_evidence_base`'s staleness machinery**,
   extended to accept a fact that can be superseded by a later, more specific news item
   (a postponed date, a changed venue) rather than only re-verifying a numeric tolerance.
3. **The `entity-directory` page role, once §6's diagnosis run reports**, as the render
   target for the resulting fixtures — one page per event, matching what boxingonline's
   own strategy already specified.

None of these is this lane's build — extraction and directory-page machinery are
news-ingestion and site-planner territory respectively, not `calendar_component`'s. Filed
here as the shared diagnosis both approaching lanes converged on.

## 8. Cross-references

- `docs/agent_docs/docs024_key_docs_latest/calendar_component/` — this lane's own docs;
  PLAN §4 carries the same finding from the calendar side.
- `docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/OWNER_REVIEW_2026-08-31_boxingonline_what_he_found_and_what_each_finding_actually_is.md`
  — the owner's original review; §1, §6, §7, §8 all touch this gap from different angles.
- `docs/agent_docs/docs024_key_docs_latest/site_delivery_and_editor/COMPARISON_2026-08-31_boxingonline_ours_vs_the_other_builders_and_why_theirs_looks_better.md`
  — line 176 (`design_intent.layout_preference`'s fixture-row ask), line 144-146
  (`entity-directory` spec text), item 7 ("give the tools real data").
- `docs/agent_docs/docs024_key_docs_latest/news_editorial_features/PLAN_2026-08-19_news_editorial_features.md`
  §3 — the original "three zeros" measurement, corrected in §4 above.
- `bugs_open/206` — entity-directory/section-index page roles have no real builder; related
  but distinct (206 is about page-role→builder mapping in general, not about where
  fixture data would come from).
- Diagnosis run in progress, correlation `d6d350ec-e16b-4792-9282-ca5155369791` — why the
  planner drops page roles the strategy names. Read before assuming a shared cause with §6.

## 9. Status update, 2026-09-02 (same day, resuming session)

Within hours of filing, three sessions picked up different pieces independently,
each verified before building rather than duplicating another's work. Status as of
this update — **check each lane's own docs for anything past this point**, this is a
snapshot:

- **Candidate #1 (populate), owned by `news_feed_ingestion`** (opened by the
  owner-named "feed lane" session; see that lane's PLAN):
  part 1 committed (`a7a134af7`) — new load/mark actions plus an extension of
  `VerifyAndRegisterCitationsAction`'s field pass-through list, reusing its
  verification and CAS-write unchanged. Corrected before shipping: candidates carry
  `kind="entity"`, not `"event"` — the latter isn't in `EvidenceFact.Kind`'s closed
  vocabulary (`datahelpers/claims.go`, `bugs_open/105`) and would silently demote to
  `"metric"` while tripping a per-build warning. Submitted for council review
  (`SUBMISSION_CORR=4849c95f-2594-48e6-87b9-acee6341b0f8`). Part 2 (wiring the
  extraction into `feed-triage`'s live workflow config) deliberately not done yet —
  needs the image built and rolled first.
- **Candidate #2 (correct), resolved with far less new code than the fix
  candidates anticipated** (`bugfix_427_event_render`, this session): tracing
  `refreshOneSiteEvidence`'s actual dispatch condition (keyed on `fact.source`
  contents, not `kind`) found that ANY fact carrying `source.citation` — which
  every fact candidate #1 registers does, regardless of `kind` — is **already**
  re-verified daily by the existing, unmodified citation-refresh arm. No new
  `source.feed_item` marker, no new dispatch branch. What tracing the actual
  CONSUMERS of the new fields did find, and fixed: `composeWriterBlock` only
  substitutes `{value}`, so a `writer_line` using `{event_date}`/`{venue}`/
  `{participants}`/`{broadcaster}` would have shipped those tokens unsubstituted
  into the writer's prompt. Fixed, mutation-tested, committed (`f865153f8`),
  submitted (`SUBMISSION_CORR=d0442d50-e383-477f-9ed8-19eaaeea3d93`). **Residual,
  named not dropped**: a correction published as a separate, later article (not an
  edit to the original) needs same-event matching across feed items — unbuilt,
  follow-up bug territory, `content_feed_items.duplicate_of` is the column that
  would eventually get a writer for it.
- **The render target** (`bugfix_427_event_render`, this session) — the piece
  none of the original fix candidates named directly: even a populated,
  correctable fact was never going to reach the page without something to put it
  there. Built as `query.upcoming_events`, a new resolver using the estate's
  existing query-source + dependency-class mechanism (RFC_052) rather than a new
  action, HTML-patching, or a client-fetched JSON file — same shape as
  `latest_news`/`news_archive`. Plus `queueEvidenceBasePageRerenders`, the
  propagation half that makes a human's correction (or the daily citation
  re-verification) actually reach a consuming page. Committed (`da2ab0d44`),
  submitted (`SUBMISSION_CORR=08f56b7e-61e4-42d1-a3b6-13d700dd833c`). **Not yet
  built**: the component that actually declares the new source on
  `/tools/fight-calendar/index.html` — deliberately deferred until candidate #1's
  part 2 produces a real fact to point at; placing an empty list now would look
  like a fix without being one. Registered:
  `docs026_concept_register/register/page-build-pipeline.md` PBP-048.
- **Candidate #3 (entity-directory page role) — root cause now DIAGNOSED, filed
  separately.** `gap_planner` resolved the `d6d350ec…` diagnosis this bug's §6
  pointed at: **not a wiring drop.** `recommended_page_types` IS in the
  planner's prompt and the model reads it correctly — the planning LLM
  *deliberately* defers `entity-page`/`entity-directory` at launch, exercising a
  "you have final say on architecture" license the prompt itself grants, then
  states its reasoning in `strategy_notes`. Fleet-sampled: ~76% omission rate
  when a strategy names these roles (33 sampled, 25 cleanly evaluable, 19
  omitted at least one). 13 `site_work_items` rows already detect this shape
  fleet-wide and nothing dispatches them. Filed as
  `bugs_open/428_HANDOFF_2026-09-02_site_planner_llm_knowingly_defers_strategy_named_entity_roles_citing_its_own_final_say.md`
  — read that file, not this one, for anything touching the planner's behaviour.
  Nothing built here depends on 428 landing: `query.upcoming_events` renders onto
  the *existing* tool page, not a future `entity-page`/`entity-directory` page.

**Update, 2026-09-02 (same day): candidate #1 part 2 is now also done — the
whole populate-side mechanism is LIVE.** `news_feed_ingestion`: image
`v1.0.1352` built from committed HEAD, pushed, rolled, verified at the binary
(sha positive+negative control on `/proc/1/exe`, both live pods — not the roll
status). Council-approved submission wired into `feed-triage`'s live workflow
config (`apply_scores → load_for_event_extraction → check_for_events →
extract_event_facts → register_event_facts → mark_event_extracted →
complete`). **Two real defects caught by testing rather than review, both
fixed before either could misbehave in production**: migration 684 was
written and verified on paper but never actually applied — the first live
dispatch failed cleanly on `column cfi.event_extracted_at does not exist`,
fixed by hand-apply + `run-migrations.sh --record-only`; and `check_has_items`
was found to short-circuit straight to `complete` whenever a cycle's ingest
queue was empty, bypassing the whole extraction chain regardless of an
unextracted backlog — restructured so both entry paths converge on the loader,
which gates on its own live count instead. Re-dispatched against
boxingonline.com end to end: **6 new dated event facts registered in its
`evidence_base`**, each verified live against its source article, real fight
results with venues and participants named, broadcaster correctly left blank
where no source stated one. 49 `content_feed_items` rows (29-item pre-existing
backlog + 20 fresh) now carry `event_extracted_at`. So `query.upcoming_events`
(candidate #2/render-target work, above) now has real facts to point at —
placing the component that declares the new source is no longer blocked on
"nothing to render." Full detail: `docs024_key_docs_latest/news_feed_ingestion/`
(PLAN status line, NOTES for the two mistakes' full account).

## 10. Status update, 2026-09-02 (later same day) — fresh build verified, REVISE round closed

**Fresh chassis build confirmed at the artefact, both directions** (per this repo's own
landmine against trusting a roll without checking): `service_binary_capabilities` /
core-manager's own startup log both show `git_commit =
ebf27c60377f984fd2847a1d5d88ff87ae01ebf7` on every current `agent-chassis` pod and on
`core-manager`, uniformly (no split-tag). `git merge-base --is-ancestor` confirms this
commit descends from every commit this session and `news_feed_ingestion` made through
today, including the composeWriterBlock fix, the `query.upcoming_events` resolver, and
feed_lane's event-extraction work. **So the render-target code, the six real boxingonline
facts, and the writer-block fix are all live in the running fleet right now.**

**`query.upcoming_events`'s council round (08f56b7e) came back REVISE, not approved.**
Read in full, not skimmed: compliance's HIGH objection was real and correct — the
resolver rendered real-world scheduling claims (fight dates, venues, named people) gated
only on "the date parses and is in the future", with no check that the underlying fact
was actually evidenced. Fixed same session (commit `987ed3b3b`): a fact now renders only
if its citation carries BOTH a `url` and a `quote` — the same bar
`VerifyAndRegisterCitationsAction` already requires to *register* a fact, enforced again
here so the resolver does not depend on that being the only write path forever. Also
fixed: every rendered item now carries a constant disclaimer ("schedule details can
change..."); the dependency-class literal is now `site_specs.evidence_base` (was the
bare, self-flagged-as-unsettled `evidence_base`); `content_components`' existing
`evidence-chart` function was checked directly (not argued from silence) and confirmed
a different mechanism (numeric chart series via `site_specs.*`, not a per-item dated
list via `query.*`); and the CAS-guard placement guardian asked about is now documented
inline. Resubmitted on the same correlation (`RESUBMIT_CORR`), verdict pending again.

**Decisions this lane now needs from a person, not from a session:**

1. **Build and deploy Phase B** — the actual component + migration that puts
   boxingonline's 6 real facts on `/tools/fight-calendar/index.html`. Everything it
   depends on is live (resolver, real data). This is real, scoped, undecided-only-in-the-
   sense-of-"has anyone started it yet" work — see §11 below for exactly what's left.
2. **Whether to release any of `bugs_open/428`'s 13 record-mode verdicts.** The tool to do
   it safely now exists (backend live in this same build; see bug 428's own status), but
   *using* it — deciding boxingonline's own row (`e3c2b440-…`) or any of the other 12 is
   worth turning into real work — is a per-row human judgement call this session
   deliberately did not make for you.
3. **Whether `news_feed_ingestion`'s extraction prompt should run beyond boxingonline.com.**
   It is vertical-agnostic by design but currently wired/tested against one site only;
   rolling it fleet-wide is a cost/quality decision for that lane's owner, named here
   because it directly determines how many OTHER sites' entity-directory gaps (bug 428 §3's
   76%-omission sample) could ever have real data to show.

## 11. What is left before this lane can close

- [ ] **Phase B (the actual fix a visitor sees).** Resolve the open questions in
  `bugfix_427_event_render/PLAN`'s §5 R1-R3 (does an existing component already accept a
  query-sourced items array with event-shaped columns, or is a small new `event-list`
  component needed; which path minted the tool page and which sections-source it reads
  from). Write the migration placing the component (precedent: migration 267). The
  eventual template MUST `{{if}}`-guard every optional field (venue/broadcaster/
  participants) — Go's `missingkey=zero` renders an unguarded absent key as silent empty
  text, flagged by council `bug_historian` as this platform's most recent recurring defect
  class if unpatched. Submit that migration to council separately (DB migrations are
  council-scope on their own).
- [ ] **Read the pending council verdict** on the resubmitted `08f56b7e` correlation and
  act on it (approved → nothing further; another REVISE → address and resubmit again).
- [ ] **A human decision on releasing any of bug 428's 13 verdicts** (§10.2) — not this
  lane's code to write, but the calendar page's own completeness may depend on whether
  boxingonline's entity-directory gap (bug 428) ever gets acted on.
- [ ] **Verify on boxingonline** once Phase B ships: `experience_loop`'s nightly check
  should reclassify `/tools/fight-calendar/index.html` out of "no control, no inline data,
  no runtime fetch" (this bug's own original measurement instrument) — that is the
  closing signal for this bug, not a code diff.
- Not this lane's job, tracked elsewhere: `bugs_open/428`'s own remaining items (see that
  file's own §11); `content_feed_items.duplicate_of`/cross-article correction (named
  residual, no owner yet).

## 12. Status update, 2026-09-02 (later same day) — round 3 APPROVED; Phase B started, split in two

**`query.upcoming_events`'s council submission (`08f56b7e`) is now APPROVED**, after two
REVISE rounds, both genuinely load-bearing, both fixed rather than argued against:

- **Round 2** (`bug_historian` HIGH): the resolver's own "omit rather than invent" policy
  for a missing field had no visibility mechanism — an incomplete or unevidenced fixture
  would render (or fail to render) with nothing durable recording it. Fixed with a NEW
  discovery check, `check_event_fixture_completeness` (commit `d6a952249`), kept
  deliberately OUT of the resolver (a query.* resolver must stay a pure read) — it
  classifies every event-dated fact as complete/incomplete/unevidenced and files one
  `capability_gap` item per site naming which, retracting it the moment a site is clean.
  Registered but **inert until named in a live `run_checks.config.checks` array** — a
  separate, low-risk operational step, not done in this round.
- Round 3 fixed the round-2 objections' own stale submission text (the "risks" section
  had kept saying the naming question was "open" and blast radius was "nil" after both
  were already resolved/stale in the code — a lesson in itself: updating the CODE is not
  the same as updating the SUBMISSION reviewers actually read).

**Phase B (the actual visible fix) is started, deliberately split in two:**

- **Step 1 — SHIPPED, live**: a new `event-list` content_components row (migration `712`,
  commit `33b19ba63`), query-sourced from `query.upcoming_events`, every optional field
  independently `{{if}}`-guarded — validated by actually EXECUTING the template through
  Go's `text/template` engine against three cases (empty / full / missing-optional-fields)
  before writing the migration, not just reading it. **Attached to NO page** — zero live
  effect today, pure library addition. Submitted for council review
  (`ff91e666-608d-4b26-9c41-d97d23a21437`).
- **Step 2 — NOT done, and here is exactly why, named rather than guessed past**:
  attaching `event-list` to boxingonline.com's live fight-calendar page. The framework's
  own designed mechanism for "a page's declared sections name a component that exists but
  isn't built" (`check_unresolved_sections`, confirmed live in
  `completeness-discovery-agent`) marks the page `build_status='needs_rebuild'`, which
  `get_pages_to_build_actions.go` then routes through the **FULL `page-build-handler`
  pipeline** — the same path that plans and writes a page from scratch, not the scoped,
  no-LLM `page-rerender` path that only refreshes an already-present component's
  query-sourced field. This session did not verify that a `needs_rebuild` pass on an
  **already-deployed page with two existing, approved sections** (hero-tool,
  generic-text-block) carries that content forward untouched rather than re-planning or
  re-writing it. Given the target is a **paid customer's live page**, this was treated as
  a real decision rather than an assumption to make silently.

**Decision needed to finish this bug**: either (a) verify the carry-forward guarantee
first (read `plan_sections_action.go`'s handling of an already-deployed generic page, or
test against a non-customer site first), then run step 2, or (b) accept the risk knowingly
and run it — the exact SQL is written out, commented, at the foot of migration `712`.
Neither was decided unilaterally this session.

**Also held back this session, same reasoning**: deploying the admin-dashboard frontend
for `bugs_open/428`'s release surface (`make admin-dashboard`) — low risk, but a real
production deploy action, flagged rather than run on a generic "carry on" instruction.

## 13. Status update, 2026-09-03 — component attached and deployed; one open defect found

**Owner said to continue.** Resolved decision #1 above WITHOUT needing to answer the
carry-forward question at all: re-read `08f56b7e`'s pending council round first and found
`prior_art_librarian` had raised a HIGH gating objection pointing at exactly the narrower
mechanism (`apply_section_edit`, `component_swap`) migration 712's own rationale had asserted
didn't exist. Verified the objection was right by reading the actual Go — `component_swap`
repoints ONE existing `page_components` row, touches no other section, and never goes near
`check_unresolved_sections`/`needs_rebuild`/the full replan pipeline. Used it: dispatched
directly (kafka envelope carrying section-editor's own live workflow), repointing the
`generic-text-block` row (never actually served — confirmed by curling the live preview
first) onto `event-list`. **Deployed and confirmed at the artefact**: git commit `007b3a7a1`,
GitHub Actions "Sync to B2" run `33672753667` shows the real `upload
tools/fight-calendar/index.html` + Cloudflare cache purge, not just a green job status.
Fixed the `pages.sections` drift this left (migration `719`, applied by hand + recorded,
verified before/after) — otherwise `check_unresolved_sections`'s next sweep would have
re-marked the page `needs_rebuild` and reopened the exact risk this whole approach avoided.

**One real defect found, not fixed: `query.upcoming_events`'s `items` field never
populates.** Reproduced 3× (including once under a chassis build 650 commits newer than
this bug's own fix, confirmed by `git merge-base --is-ancestor`, ruling out a stale-binary
explanation). One evidence_base fact genuinely qualifies (Canelo Alvarez vs Christian Mbilli,
2026-10-31, citation complete) — the resolver's own future-date filter correctly excludes the
other five (past results/historic), that part works. Each light-rerender (`page-rerender`,
`reason=section_data_resolved`) reports success with `escalated=false` and a real `rerendered`
count, yet `page_components.content_data` for the event-list row comes back byte-identical to
before every time — still the OLD `generic-text-block` fields, never `items`/`headline`/
`empty_text`. Could not catch either `plan_sections_action.go`'s query.* branch or
`queryresolve/upcoming_events.go`'s own logging firing, across three careful live-log
captures (one with `kubectl logs -f` started BEFORE dispatch). **Root cause not established
— full reproduction recipe and two untried next steps (distinguish carry-vs-fresh-render;
control-test against a known-working query-sourced component) are in
`docs024_key_docs_latest/bugfix_427_event_render/NOTES_bugfix_427_event_render.md`'s
2026-09-03 entry.** This is exactly the class CLAUDE.md's diagnosis-before-debugging section
says should go through `090_TRIGGER_needs_diagnosis` rather than get one more guess from a
session that has already spent a long time on it.

**Admin-dashboard (bug 428's item) — also done this session, cross-referenced there.**
Built, verified the image content, `docker push` correctly refused by the auto-mode
classifier and handed to the user; live now at `v1.0.1356`, re-verified at the artefact. Full
account: `bugs_open/428` §12.

**What's left before this lane can close, updated:**
- [ ] Diagnose and fix the `query.upcoming_events` items-not-populating defect (above).
- [ ] Resubmit the `ff91e666` council round (still REVISE) with the `component_swap` answer
  to `prior_art_librarian`'s objection — the migration itself doesn't need new code, the
  submission text needs updating to say what was actually done and why it satisfies the
  objection.
- [ ] Once items populate: re-verify `experience_loop`'s nightly check reclassifies
  `/tools/fight-calendar/index.html` out of "no control, no inline data, no runtime
  fetch" — this bug's own original measurement instrument and the real closing signal.
- Not this lane's job, tracked elsewhere: bug 428's own remaining item (a human uses the
  release surface on a real verdict).

## 14. Status update, 2026-09-03 (later same day) — the open defect is diagnosed and fixed, and it was NOT this bug's code

**§13's "one real defect found, not fixed" now has a root cause, and it lives nowhere near
this lane.** Filed as its own case: **`bugs_open/454`** — *the light re-render computes a
section plan and drops it, so every page is rendered from its own stored data*. Fix
committed `9831e9ab4`; council submission `075cfedd-aef0-4230-b4f1-909ecf68959d`.

**In one sentence:** `classifyStoredSection` in `rerender_page_sections_action.go` calls
`planSection`, uses the result to decide the row can render, then returns without setting
`c.plan` — a struct field that is **read at exactly one line in the repository and written
at none** — so `renderPlannedSection` gets a zero value, and every light re-render since
2026-09-02 has composed `base ⊕ stored content_data` with `plan.ResolvedData == nil` and
persisted the stored map unchanged.

That is the whole of §13's symptom. `items` was never going to populate, on this page or
any other, no matter what this lane did to the component, the schema, the resolver or the
register. Introduced by `94f81cc60` (2026-09-02 12:27 BST, an extraction commit in the
`features_open/035` decomposition) — which is **hours before** the event-list attachment
work of §12/§13, so this lane never saw the mechanism work even once.

**Two things §13 asserted that are now corrected, visibly rather than by editing them away:**

> **CORRECTED 2026-09-03.** §13 said "Could not catch either `plan_sections_action.go`'s
> query.\* branch or `queryresolve/upcoming_events.go`'s own logging firing, across three
> careful live-log captures". Under the real cause the resolver **is** still called —
> `planSection` runs in full, and only its *result* is discarded — so
> `logger.Info("queryresolve: resolved upcoming_events", …)`, which fires unconditionally on
> every call, must have been emitted. The pod-log capture was the faulty instrument, not the
> code. Do not carry "the query resolver never ran" forward.

> **CORRECTED 2026-09-03.** §13 framed the open question as carry-vs-fresh-render and named
> distinguishing the two as the first next step. The framing was too narrow: the row was
> **freshly rendered from stale inputs**, a third state neither option covers — and one that
> produces exactly the byte-identical output §13 had (correctly) said could not discriminate
> between its two. The *action* it named was still the right one: reading
> `classifyStoredSection` line by line is what found this, within minutes.

**On the `090_TRIGGER_needs_diagnosis` run §13 nominated: not run, deliberately, and the
reason is stated in `bugs_open/454` §7** rather than left as a silent omission. Short
version: the claim turned out to be a property of the source text — a field read once and
assigned nowhere, settled by one grep and then demonstrated by a failing test — so there was
no hypothesis for the loop to refute.

**What this changes for this lane's remaining work:**

- [ ] **~~Diagnose and fix the `query.upcoming_events` items-not-populating defect~~ — DONE,
  as `bugs_open/454`. But this lane stays OPEN and the box stays unticked**, because the fix
  is Go and is therefore inert until a chassis image carrying `9831e9ab4` is rebuilt and
  rolled. Both live `agent-chassis` builds still carried the defect at 2026-09-03 09:54 UTC.
  When it rolls: re-dispatch the page-rerender (recipe in the RUNBOOK), then read the
  artefact — `content_data->'items'` non-empty and `rendered_html` no longer 1,813 bytes —
  and then curl the served page.
- [ ] Resubmit the `ff91e666` council round (unchanged from §13).
- [ ] Re-verify `experience_loop`'s nightly reclassification (unchanged from §13) — this
  remains the real closing signal, and it cannot fire until the roll above.

**The one genuinely good piece of news for this lane:** nothing built here was wrong. The
component, its schema, the evidence gate, the resolver, the register fact and the
`component_swap` attachment were all correct and all in place; they were being fed into a
render path that had stopped delivering fresh data to *every* component on the estate three
hours earlier. When the chassis rolls, this page should populate with no further work.

---

# ADDENDUM 2026-09-03 10:3xZ — the WRITER end now works on this site; the READER end does not

**Added by the boxingonline review session.** The bug's title says "no writer populates dated,
correctable event facts". **On boxingonline that half is now FALSE.** Measured at the live DB.

## What changed

```sql
SELECT jsonb_array_length(data->'facts'), source, created_by, created_at
  FROM site_specs WHERE site_id='d2aa5206-…' AND aspect='evidence_base' ORDER BY created_at;
```
| facts | source | created_by | created_at |
|---|---|---|---|
| 2 | order_intake | seed_build_queue | 2026-08-31 12:21 (superseded) |
| 1 | operator | site_delivery_and_editor-session | 2026-08-31 15:54 (the email-claim removal) |
| **7** | **research** | **evidence-researcher** | **2026-09-02 12:41** |
| **7** | **scheduled** | **evidence-refresher** | **2026-09-03 09:11** |

**⚠ THE "ONE FACT" FIGURE IN §3 IS STALE.** I supplied it and I am correcting it here rather than
leaving it to be quoted: this site went 1 → 7, and a `scheduled` refresher has since re-run.

**And they are real, dated, cited facts** — `source.citation` carrying a URL and a verbatim quote:

- Filip Hrgovic stopped Moses Itauma in round 9 for the vacant IBF heavyweight title, 30 Aug 2026
- Chantelle Cameron beat Mikaela Mayer by decision, unified light-middleweight
- Mayer vs Cameron headlined MVPW-06 on 29 August at the bp pulse arena
- **Canelo Alvarez is scheduled to fight Christian Mbilli on October 31** ← a FORWARD-LOOKING,
  DATED FIXTURE, which is precisely the shape this file says nothing produces
- Adrian Rueda won the vacant WBO I-C super lightweight title, 30 Aug 2026
- Wladimir Klitschko fought Ross Puritty, 5 December 1998, Kyiv

## What has NOT changed — the reader end

Measured at the served site in the same minute:

```
/tools/fight-calendar/index.html   inputs=0   inline data arrays=0   fetch(=0     (unchanged)
Canelo/Mbilli occurrences: /news/index.html 3 · index, articles-index, calendar, about: 0
```
and the 3 on the news page are the FEED's own items, not the fact corpus. The comparator is
unchanged: 18 manual inputs, no shipped data.

**So the corpus is populated and nothing consumes it.** The two ends of this bug have separated:
acquisition works on at least one site; there is no path from `evidence_base` to a tool.
**Whoever picks this up should restate the title before building** — building the writer again
would be building the half that already works.

## Explicitly NOT established

- **Why** `evidence-researcher` ran here — deliberate dispatch or normal build flow. Unknown.
- **Whether the fleet picture moved** from the §3 baseline (20 of 54 sites with a row; 42 of 54 at
  ≤5 facts). That baseline is now at least one site out of date and it sizes the remaining job.
  **Re-measure before planning against it** — the same staleness that made the "one fact" figure
  wrong, one level up.

## A narrow related defect, recorded so it is not lost

The seeded `business_name` fact carries `source.attested_by` with **no date**, and the attestation
staleness checker treats undated as beyond its 180-day cadence — so it parked a `stale_attestation`
work item at `needs_human_review` on a three-day-old site
(`0cdddb6f-f05f-4c7a-bd7f-375f807da73b`, "1 attested fact(s) due for human re-look").
Fleet-wide it is a **singleton: 146 attested facts, exactly 1 undated**, and it is this one — so
order-intake's seeder writes an attestation that can never be satisfied without a human. Narrow,
cheap, and it puts a spurious item in front of the owner at his pre-delivery review.

## 15. Status update, 2026-09-03 — `ff91e666` resubmitted (round 2), and a defect in this lane's OWN migration 719

**Round 2 of `ff91e666` is submitted** (dispatch correlation `c46cf6c2`, trail still keyed on
`ff91e666`). It answers the gating objection with what actually happened rather than with a
better argument, and it **widens the edit list from one migration to three**, because the
objection was about a deferral and the deferral no longer exists.

- **The gating objection (`prior_art_librarian`, HIGH) was right, and is now answered by
  events.** Round 1 justified holding Phase B on the claim that only the full
  page-build-handler pipeline could attach the component. That was a load-bearing absence
  claim made without checking the section-editor family. The seat's own hypothesis — *"if
  section-editor can attach a new component to sections, the stated blast-radius concern is
  moot and Phase B could ship far sooner and smaller than planned"* — is exactly what
  happened: `apply_section_edit` with `edit_type=component_swap`, read in
  `ApplySectionEditAction` before use, attached it the same day with no rebuild.
- **The three unanswered objections are now MEASURED at the live row, not argued.**
  `[MEASURED 2026-09-03]` the component's `input_schema` top-level keys are exactly
  `{notes, fields}` — native dialect, so `chk_input_schema_no_legacy_dialect` is satisfied
  (and was at INSERT); `component_level='section'`, deliberately not `'tool'`, so nothing
  here pulls this page into `check_tool_health` or Tier-2 acceptance gating; and the
  `query.*`-sourced component census re-reads as **33 rows / 30 active**, which corroborates
  round 1's "32" as that day's all-rows count taken before this row existed.
- **Migrations `719` and `727` are in the edit list because a migration IS the running
  system.** Submitting the library-only INSERT while its two live-data siblings went
  unreviewed would have been reviewing the safe third of the work.

### The defect in 719, found while writing the resubmission — not by any detector

`719`'s UPDATE rebuilt the array with `jsonb_agg(DISTINCT x)` and **no `ORDER BY`**, which
does not preserve input order. So:

| | value |
|---|---|
| before 719 | `["hero-tool", "generic-text-block", "advertising"]` |
| after 719 | `["advertising", "hero-tool", "event-list"]` |
| 719's own stated intent ("replaces the array entry") | `["hero-tool", "event-list", "advertising"]` |

**`pages.sections` is order-bearing BY INDEX**, not by membership —
`save_page_sections_action.go:1979` says so outright, `adopt_fragment_section.go` replaces a
section with `planned[Position-1]`, and `section_editor_actions.go`'s
`loadPageComponentBySlotRO` carries a match arm on `p.sections->(pc.position - 1)`. After
719, `page_components` position 1 (`hero-tool`) indexed to `"advertising"` and position 2
(`event-list`) to `"hero-tool"`.

**Bounded honestly: this was NOT live damage.** That section-editor arm is gated on
`pc.slot_name IS NULL OR pc.slot_name = ''`, and `[MEASURED 2026-09-03]` both rows on this
page carry non-empty slot names. It was a **latent** misalignment that goes live the moment a
build leaves a slot name empty — and silent in both directions, because a wrong-but-present
name matches nothing rather than erroring.

**Fixed by migration `727`** (`727_boxingonline_fight_calendar_sections_restore_position_order.sql`),
applied by hand and recorded `--record-only`. Two `RAISE` pre-checks (the array is exactly the
post-719 value; the live composition really is `hero-tool@1`, `event-list@2`), and a verify
block that asserts **the index alignment itself** rather than the array's literal value — so
it would also catch a future drift this migration did not cause. **Rehearsed under
`BEGIN`/`ROLLBACK`** (clean; post-rollback control unchanged) and **induced-failure-proven**:
writing a deliberately wrong order made the verify block `RAISE` and abort with the row
untouched. Live after apply: `["hero-tool", "event-list", "advertising"]` against a live
composition of `hero-tool@1, event-list@2`.

**Named and deliberately NOT fixed:** `"advertising"` is declared in that array and has no
`page_components` row on this page (`content_components` `ad_zone_inline`,
`function='advertising'`, `component_level='section'`). It predates 719 and predates this
lane, so 727 left it exactly as found. If `check_unresolved_sections` is flagging this page,
that entry is why — a separate decision from the ordering fix.

**Not censused, and stated as a gap rather than left implied:** whether other pages
fleet-wide carry a `pages.sections` whose entries do not index onto their own
`page_components` positions. 727 restores ONE page. If 719's aggregate pattern has been
copied elsewhere, that is a fleet question this lane has not answered.

**719's header is left unedited on purpose.** It is an applied migration and the runner's
drift guard hashes it; its now-refuted paragraph about the items defect ("No log line from
either function appeared…") is corrected in §14 above and in `bugs_open/454` §5 instead.

## 16. Status update, 2026-09-03 — `ff91e666` round 2 came back REVISE, and every gating objection was right

`ff91e666` round 2: **REVISE** at 11:11:24Z, `decided_by: "gating objection from guardian"`,
5 abstained, no truncation. Seven seats approved (`guidelines`, `diagnosis_guardian`,
`render_guardian`, `constitution`, `mission`, `prior_art_librarian`, `architecture`); five
objected. **Not one of the objections was wrong**, and two of them found live conditions this
lane had named and then failed to act on. Recorded here before round 3 because the findings
outlive the round.

### 16.1 The census three seats asked for — run, and it does not say what a tidy story would

`guardian` (HIGH/LOW), `bug_historian` (MEDIUM) and `reuse_agent` (MEDIUM) all pushed on the
same gap: 727 fixed one page and the plan admitted no fleet census. `bug_historian` was
explicit that *"approve is reachable if the fleet census check comes back clean."*

`[MEASURED 2026-09-03]` over live `page_components` on active pages, positions indexable into
`pages.sections`:

| | count |
|---|---|
| indexable live rows fleet-wide | **2,719** |
| aligned (index names the row's own slot or function) | **2,610** |
| **misaligned** | **109** |
| pages carrying at least one misalignment | **68** |
| sites | **21** |

**It did not come back clean, and the honest reading is narrower than the number.**
Disaggregated, because the total conflates two different defects: **95** of the 109 have their
own name present ELSEWHERE in the array — an ordering/offset shape — while **14** are absent
from it entirely, which is a declared-vs-realised divergence and a different bug. And **72 of
the 109 sit on pages whose declared entry count differs from their live row count**, which
offsets indices by construction and has nothing to do with any reordering transform.

**So: I am NOT attributing the 109 to 719's idiom, and nobody reading this should.** The
causal share is unmeasured. What the census establishes is that the misalignment *class* is
fleet-wide and mostly pre-existing, not that this lane caused it.

**The containment fact, which is the one that answers the guardian: 0 of the 109 have an empty
`slot_name`.** `section_editor_actions.go`'s positional match arm is gated on
`pc.slot_name IS NULL OR pc.slot_name = ''`, so it cannot currently fire on a single one of
them. Fleet-wide the misalignment is LATENT, exactly as it was on this page — and it goes live
the moment a build leaves a slot name empty.

### 16.2 `bug_historian` asked whether the anti-pattern was guarded against reuse. It is not, and it is already reused five times

`[MEASURED 2026-09-03]` five other migrations rebuild `pages.sections` with
`jsonb_agg(DISTINCT x)` and no `ORDER BY`: **248, 252, 255, 266, 267**, each appending a slot.
**`267`'s own header recommends the idiom** — *"both statements are naturally idempotent:
NOT EXISTS on the slot, and `jsonb_agg(DISTINCT)` on sections"* — which is true of membership
and false of order. So the trap is written into this repo as good practice, which is exactly
why 719 reached for it.

Now carried as a **LANDMINES entry** (footprinted on `pages.sections`, the idiom, and the three
positional readers), with the check being: write the literal array in position order, gate the
UPDATE on the exact prior value, and verify the JOIN rather than the value. Nothing lints for
it; that entry is the only guard.

### 16.3 `guardian`'s second MEDIUM was live, not hypothetical — and 727's header was wrong to leave it

> *"The 'advertising' array entry with no page_components row is left in place… Author names it
> but does not close it or explain why it is safe to leave."*

727's header called it "pre-existing, left exactly as found". That was true and it **was not a
justification**. Checked against `check_unresolved_sections.go:36-56`'s actual predicate, all
four arms held: page `active`, `build_status='deployed'`, a live non-forked component matches
the name (`ad_zone_inline`, `function='advertising'`), and no `page_components` row joins to it.
**The next sweep would have marked this page `needs_rebuild`** and routed it into the
full-rebuild pipeline this entire correlation exists to avoid.

**Closed by migration `728`** — remove the declaration rather than realise it, because
`[MEASURED 2026-09-03]` there are **ZERO** `page_components` rows fleet-wide joining to
`function='advertising'`, across every site and every non-removed status. The component exists
in the library and nothing has ever placed one. Three active pages declare it, all on
boxingonline.com: `index` (already `needs_rebuild`), `cruiserweight-boxings-best-kept-secret`,
and this page. Fleet population armed by the same predicate today: **18 pages across 3 sites**.

728 is guarded by two `RAISE` pre-checks (exact pre-state; and a refusal if an advertising row
has appeared since the census, in which case dropping the declaration would be wrong), and its
verify block asserts **both** that the page no longer arms the detector and that 727's index
alignment survived. Rehearsed under `BEGIN`/`ROLLBACK` and induced-failure-proven: leaving the
entry in makes the verify `RAISE`. Live after apply: `["hero-tool", "event-list"]`, armed 0,
alignment intact.

`[NOT ESTABLISHED]` and deliberately not asserted: whether a rebuild would realise an
advertising row at all. If it would not, those pages are on a permanent re-arm treadmill
(marked → rebuilt → still unresolved → marked). I have not read the build path far enough to
claim it.

**Scope stated rather than silently narrowed:** 728 touches ONE page. The other two boxingonline
pages are named and untouched — `index` is already `needs_rebuild` so its door has already
opened, and `cruiserweight-…` is a content page this lane has no business re-planning.

### 16.4 The objections I am answering with words rather than code

- **`reuse_agent` MEDIUM ×2** — 719 and 727 hand-write SQL against `pages.sections` instead of
  going through `save_page_sections_action.go`, the typed writer, or `ReconcileSitePlanAction`.
  **The seat is right that this was never checked, and the order-loss defect is the direct
  consequence it predicts.** Round 3 will say so plainly rather than defend it.
- **`debug_historian` MEDIUM/LOW** — no pre-mutation dump before production jsonb surgery. 727
  and 728 were each rehearsed under `BEGIN`/`ROLLBACK` and each carry a hand rollback statement,
  but a rehearsal is not a backup and the seat is right that the protocol asks for one.
- **`editquality` LOW ×2** — edit 2 changes nothing functional and should have been labelled a
  review-context annotation, not a `modify`; and 727's exact-match pre-check makes it
  non-idempotent against other drift, which **was intentional** (this migration corrects one
  specific prior value and must refuse anything else) and should have said so.

## 17. Status update, 2026-09-03 — the chassis rolled, the fixture DOES resolve, and the page is now blocked one step later

**The roll happened (chassis `d0252fd4d`, image `v1.0.1358`, 12:18Z) and it carries the 454
fix.** Dispatched the re-render this lane has been waiting two days to run. Result, from the
run's own `sections_metadata`:

- **`event-list` resolved its fixture.** `content_data` gained `items`, **n = 1**, and
  `rendered_html` went **1,813 → 2,498 bytes**. The control taken minutes earlier read
  `content, heading`, 0 items, 1,813 bytes — the same value this lane measured across three
  dispatches yesterday.
- **`hero-tool` regained `hero_url` and `background_image`** in the same pass, which nobody was
  looking for. That is `planSection`'s authoritative hero aliasing, lost to the same defect.

**So every claim this lane made about its own work is now confirmed at the artefact.** The
resolver reads the register correctly, the evidence gate admits the one qualifying fact and
excludes the five that do not, the component's guarded template renders the fixture rather than
the empty state, and the `component_swap` attachment put it in the right place. §14's conclusion
— "nothing built here was wrong" — was right.

**But the page has NOT changed yet, and will not until a separate decision is taken.** The save
was refused:

> `OWNED_PAGE_GUARD: page tool-fight-calendar is page_type=tool with no tool component: a
> generic section save would publish prose about a tool that is not there.`

That is `bugs_open/450`'s guard, live in the same image (`587666be8`). **Its predicate genuinely
holds on this page** — `page_type='tool'`, and both components are `component_level='section'`
(`hero-tool` is a section named for a tool, not a `component_level='tool'` row). This lane is
not claiming the guard is wrong, and is not routing around it.

`[MEASURED 2026-09-03]` the refusal reaches **58** active tool pages on **12** sites, **53** of
them on **9** sites already serving deployed components. Raised with the `450` lane with that
measurement — they had explicitly asked to be told if their tool-shell arm blocked a 454-driven
re-render. **The scope decision is theirs; this file will not fork a second account of it.**

**What this changes for this lane, honestly:** §14 said the page "fills in on its own once 454
ships". That is now **wrong** and is corrected here rather than left standing. 454 shipped, the
data resolves, and the page still shows the empty state, because the last step before the
artefact refuses. The remaining work is not this lane's code.

- [x] ~~Diagnose and fix the `query.upcoming_events` items-not-populating defect~~ — **DONE and
  VERIFIED LIVE** (`bugs_open/454` §12). The re-render half needs no further work.
- [ ] **The save must be allowed to complete.** Blocked on `bugs_open/450`'s scope decision, not
  on anything here. When it lifts, re-dispatch (recipe in the RUNBOOK) and read the served page,
  not the job status.
- [ ] `ff91e666` **round 3** in flight (dispatch `b1a2cf68`) — verdict not yet landed.
- [ ] Re-verify `experience_loop`'s nightly reclassification once the page actually changes.
  Still the real closing signal, and still gated on the save.

## 18. Status update, 2026-09-03 — the blocker has a committed fix; it rides the next roll

**`bugs_open/450` removed the tool arm from `save_page_sections` (`29b40e8bc`, 13:32 BST).**
Verified at the source rather than taken on report: `save_page_sections_action.go:210` now reads

```go
if refused, class, _ := pageRefusesGenericBuild(...); refused && class == refusalOwned {
```

so only the **owned** arm fires at that seam and `refusalToolPending` no longer does. Surgical —
one condition, not a removed guard. The tool arm still fires at its other three call sites
(`load_page_record_action.go:259`, `multipage_actions.go:43`,
`rerender_page_sections_action.go:1205`), so 450's own protection is intact, and migration 164's
verbatim-tool protection is untouched.

**This unblocks 427's last step — but not until the next chassis roll.** The live chassis
(`d0252fd4d`) still carries the refusing version, so the fight-calendar save will keep failing
until an image containing `29b40e8bc` ships. Nothing further to do here before then.

**A lever exists and is NOT this lane's to pull:** `DISABLE_TOOL_SHELL_REFUSAL`
(`owned_page_guard.go:96`) disarms the arm fleet-wide with no build. The 450 lane has put it to
the owner as their decision. Recorded here so nobody in this lane reaches for it.

**Census reconciliation, because their number and mine disagree and the difference is
instructive.** They measured 67 matched / 54 serving / 10 sites; I measured 58 / 53 / 9. Re-run
today to find the encoding difference rather than assert one of us was right:

| predicate | pages | sites |
|---|---|---|
| `status='active'` + a `build_status='deployed'` row (mine) | **53** | 9 |
| `status='active'` + any non-`removed` row (theirs) | **54** | 9 |
| any status + any non-`removed` row | 56 | **10** |
| `status='active'` matched at all (mine) | 58 | 12 |
| any status matched at all | 64 | — |

So **53 vs 54 is `deployed` versus any-non-`removed`** — one page holds a live component row that
is not yet `deployed`. They adopted that split; the more inclusive form is the honest one for
"would be refused a repair", since a live-but-not-yet-deployed row is a page mid-maintenance.

> **RECONCILED 2026-09-03 (the 450 lane), and the correction is MINE to own: my predicate was not
> the guard's.** The 58-vs-67 gap is `cc.is_active`. `toolShellPredicateFor`
> (`owned_page_guard.go:160-168`) carries `AND cc_g.is_active = true`; my census asked only
> whether a `component_level='tool'` row existed at all. Verified first-hand: the gap is **exactly
> 9 pages across 5 sites** that hold a tool component which is INACTIVE — ai-agent-orchestration
> ×2, finetuning.uk ×3, gaswholesalers ×1, leopardessconsulting ×1, robot-hands ×2. Running the
> guard's exact predicate (`status='active'`, `page_type='tool'`, `NOT EXISTS` a not-removed
> `component_level='tool'` row with `is_active`) reads **66 pages / 15 sites**, against **57 /
> 11** for my looser form — the same 9-page gap.
>
> ⚠ **THAT FIGURE IS A SNAPSHOT OF A DRAINING POPULATION, not a constant, and my first
> annotation of it was wrong twice.** It was stamped "~14:00" when the database clock read
> **12:00–12:40 UTC** — I wrote BST as though it were UTC, on a number I had just finished
> logging a wrong call about. Corrected: `[MEASURED 2026-09-03 12:00–12:40 UTC]`, **stable at
> 66/15 across two readings 40 minutes apart**. And the set is being actively drained: the
> portfolio lane's repairs attach tool components as they land — `[MEASURED 2026-09-03]` **50 in
> the trailing 12 hours**, 13 in the 12:00 UTC hour alone — and each one that lands on a
> refused page removes it. The 450 lane read **67/16** slightly earlier and watched a page leave
> between two of their own readings. **So expect this number to fall, and re-read it rather than
> quoting it.** My two readings did not move, which bounds the drain rate over that window
> without disproving it — most of those 50 attachments are landing on pages that already had a
> tool, or on pages that are not `page_type='tool'`.
>
> **Theirs is the operative number and mine answered a different question.** "How many pages does
> this guard act on" is 66-67; "how many pages have a tool component of any kind" is 57-58. The
> lesson is the one this estate keeps paying for: **I measured a guard's reach with a predicate
> that was not the guard's**, so my figure was a floor and read as a total. When the thing being
> measured IS a mechanism, copy its predicate rather than paraphrasing it.
>
> None of it changes the decision — every encoding said the arm refused several times more repair
> than harm — and the 450 lane notes the same `is_active` seam was found from the opposite
> direction by the portfolio lane four hours earlier, which they read as an argument for keeping
> it (it is also what `create_tool_component` probes with).

## 19. Round 3 APPROVED — and checking the objection I had conceded twice refuted its premise AND found that migrations 719/727/728 are TRANSIENT

`ff91e666` round 3: **APPROVED** at 2026-09-03 12:41:09Z,
`decided_by: "approved with 6 advisory objection(s) — none high-severity"`, 4 abstained, no
truncation. Seven seats approved; six objected, none above MEDIUM.

**Five separate seats** (`reuse_agent`, `guardian`, `constitution`, `prior_art_librarian`,
`architecture`) repeated one objection: that 719/727/728 should have gone through
`save_page_sections_action.go`, *"the platform's typed writer for exactly this field"*, or
`ReconcileSitePlanAction`. Round 2 and round 3 both recorded it as an objection I was not
answering. **This time I checked it, and the check refuted its premise — and then found
something much worse.**

### 19.1 The named remedy does not exist

`[MEASURED 2026-09-03]`, read at the source:

- **`save_page_sections_action.go` contains no `UPDATE pages` at all** and never writes the
  `sections` column. Its own first line says what it owns: *"saves rendered HTML sections to
  **page_components** table"*. Its `sections` locals are `[]SectionData` — page_components rows.
  It is the typed writer for a **different table**.
- **`ReconcileSitePlanAction` is site-scoped** (`target_site_id`), reads `site_plan_sections`,
  and emits rebuild decisions. Its own comment states the comparison is *"deliberately NOT
  plan-to-`pages.sections`"*. It cannot make a one-page section-array correction.

The real writers of the column are `apply_gap_plan_action.go` (×2),
`load_page_sections_from_spec_action.go`, `ensure_page_section_layout_action.go`, and
`site_db_actions.go`'s upsert. **None takes an arbitrary array for one page**; each is bound to
its own workflow.

> **The correction is mine to own twice over.** `reuse_agent` named the file in round 2; I quoted
> it respectfully in round 3's `grounded_in` **without grepping it**, and four more seats then
> objected on my own quotation. I manufactured the consensus I was conceding to. This is
> MEMORY's *"a CITATION is not a READ — quote the deciding ARM"* and *"an objection naming one
> file is naming a CATEGORY"*, and I had both lessons in front of me.

### 19.2 But the seats' INSTINCT was right, for a reason none of them stated

> **CORRECTED 2026-09-03 (later, session "427" on resumption) — the CONCLUSION below is
> RIGHT and the MECHANISM is WRONG, and the difference changes what you must guard.**
>
> `sync_pages` is **not** what reverts these migrations, and a re-plan is **safe**. Before
> `sync_pages` runs, `ValidateSitePlanAction` → `reconcilePlanWithRealised`
> (`v3_site_actions.go:7701-7724`) **snaps** a `deployed`/`needs_rebuild` page's realised
> `pages.sections` back **onto** the plan proposal — so a re-plan launders the cache
> FORWARD into the next plan's authority. (`realisedPageCompositionIsPreserved`, `:7883`,
> is the `deployed`/`needs_rebuild` test; the live `load_existing_pages` step supplies
> `p.sections` and `p.build_status`, confirmed in `agent_definitions`.)
>
> What actually reverts is the **page BUILD**: `load_page_sections_from_spec_action.go`
> reads tier 1 `site_plan_sections` for the current plan (`:142-148`) and **syncs it DOWN
> over `pages.sections`** (`:558-570`). **No re-plan is required.** This is the trap
> migration `154`'s own header documents from 2026-07-15, after `153` made exactly the
> mistake 719/727/728 made.
>
> **Why this mattered enough to correct rather than annotate:** "no re-plan is scheduled,
> so we are safe" is a reasonable inference from the paragraph below, and it is false. It
> also mis-sized the escalation — §19.3 concluded the fix was blocked on an owner ruling,
> when the remedy was a single-page migration with a precedent already in the tree.
>
> **RESOLVED**: migration `750` (2026-09-03) corrects the current plan's rows to
> `[hero-tool, event-list]`. Applied, verified, artefact byte-identical. See §21.
> Logged in `WRONG_CALLS.md`, together with this session repeating the same shape an hour
> later while correcting it.



Chasing the real writer found this, and it is the most consequential thing in this bug file:

**`pages.sections` is a CACHE. `site_plan_sections` is the authority — and the plan still names
the old composition.** `site_db_actions.go:1276` (`sync_pages`) writes
`sections = EXCLUDED.sections` whenever the incoming plan's proposal is non-empty; the
non-empty→empty transition is the only one it intercepts.

`[MEASURED 2026-09-03]` the current plan for this page (`site_plans bba66eda`, 2026-08-31) is:

| ordering | component_name |
|---|---|
| 0 | `hero-tool` |
| 1 | **`generic-text-block`** ← the slot `719` swapped away |
| 2 | **`advertising`** ← the entry `728` removed |

**So the next `sync_pages` run for boxingonline.com overwrites `pages.sections` back to
`["hero-tool","generic-text-block","advertising"]` in one write — undoing 719, 727 AND 728
together, and re-arming `check_unresolved_sections` on both names.** Three guarded, rehearsed,
induced-failure-proven migrations, all transient, because every one of them wrote the cache and
none wrote the authority.

**This is the same class as `bugs_open/443`** (*"`create_blog_posts` writes `pages.sections`, the
cache, never `site_plan_sections`, the authority"*) — the 450 lane flagged 443 as required
reading and they were right for a reason none of us had connected.

### 19.3 What I did NOT do, and why

> **NARROWED 2026-09-03 (later, session "427") — the premise is real; it does not reach
> this page, and that is checkable rather than arguable.**
>
> Per-plan immutability genuinely does underwrite `decideEmit` — but only for the
> **plan-to-plan** comparison, and only where a *superseded* plan's rows are involved.
> For this page none of that applies:
> - `decideEmit` returns `skip_built` on `BuiltFromPlanVersion == planID`
>   (`reconcile_site_plan_action.go:612-614`) **before it compares any section list**;
> - `[MEASURED 2026-09-03]` boxingonline.com has **exactly one** `site_plans` row, which is
>   both `is_current` and this page's `built_from_plan_version` — there is no superseded
>   plan whose history could be falsified for a cohort;
> - `site_plan_sections` is keyed per `(plan_id, page_name)`, so no other page's verdict moves.
>
> After the correction `built_from_plan_version` becomes a **true** statement about what was
> served, where before it was false. So the fourth migration this section declined to write
> was the right call for the *general* case and the wrong call for *this* one.
>
> The general question — may a non-planner action mutate the current plan's rows — remains
> open and is now going to architecture review on its own merits, with a typed action
> proposed so no lane hand-writes a three-store correction again.



**I did not write a fourth migration against `site_plan_sections`.** That table is relied upon as
**immutable per plan** — `reconcile_site_plan_action.go`'s `decideEmit` comment says
`site_plan_sections` *"is per-plan and immutable, so it can"* tell an unchanged page from a
re-composed one. Mutating it would silently break the one mechanism that distinguishes those, on
every page of the site, to fix one page's array. That is a bigger decision than this lane should
take at the end of a session, and it belongs in front of the council or the owner rather than in
a hand-applied migration.

**So this is filed, not fixed**, and it is now this lane's top open item:

- [ ] **719/727/728 are transient.** Either the plan's proposal for `tool-fight-calendar` must be
  corrected (needs a decision about `site_plan_sections`' immutability), or a re-plan of
  boxingonline.com must be expected to revert all three. **Do not close 427 believing the
  section array is durable.** Cross-reference `bugs_open/443`.
- The council advisories that stand un-actioned, recorded rather than dressed up: the
  `jsonb_agg(DISTINCT)` anti-pattern still has only a LANDMINES entry and no lint (four seats);
  `bug_historian`'s point that 18 pages carry the identical `advertising` arming condition and
  this lane filed the number without filing the work.

### 19.4 The `sync_pages` write has a GUARD, and stating it changes the severity — correction from the `bugs_open/384` lane

§19.2 said `sync_pages` *"writes `sections = EXCLUDED.sections` whenever the incoming plan's
proposal is non-empty"*. That is true and incomplete, and the missing half changes what a reader
should do about it. The write is a four-arm `CASE` (`site_db_actions.go:1277-1283`), confirmed
here against the source:

```sql
sections = CASE
  WHEN $13::bool THEN EXCLUDED.sections                                  -- allow_empty_sections
  WHEN COALESCE(jsonb_array_length(EXCLUDED.sections), 0) > 0 THEN EXCLUDED.sections
  WHEN COALESCE(jsonb_array_length(pages.sections), 0) = 0 THEN EXCLUDED.sections
  ELSE pages.sections
END
```

**The non-empty → empty transition is intercepted** unless the caller passes
`allow_empty_sections`. That guard is `bugs_open/204`'s, added after a measured incident in which
one replan emptied **41 of 45 live pages** and queued 20 `needs_page` items against them;
deliberate emptying now travels through the `recompose_pages` release instead.

**So the severity of §19.2 is "your write is lost", NOT "a live page is silently emptied"** — and
those are different things a reader would act on differently. My migrations still lose, because a
non-empty plan proposal takes arm 2; that conclusion is unchanged. But nobody reading this should
conclude that `sync_pages` can blank a populated page's section list.

**Two further nuances worth carrying:**

- **Arm 3 protects populated pages only.** A page whose `sections` is ALREADY empty takes the
  incoming proposal unconditionally. The guard is not "the plan can never win", it is "the plan
  cannot win by emptying something that has content".
- **The read side has a real dependency on this column being non-empty**, which is why the guard
  matters beyond bookkeeping: `query.blog_posts` resolves through `resolvePagesWhereType(...,
  listedOnly=true)`, whose floor (`ListedPageEligibilitySQL`, `queryresolve.go:469`) requires
  `p.deployed_at IS NOT NULL AND jsonb_typeof(p.sections)='array' AND jsonb_array_length(p.sections) > 0`.
  The same literal is shared with `discovery_checks.ContentImageMissingCheck`, so an emptied
  array would drop a page out of its listing AND stop the imagery sweep giving it a card — silent
  in both directions at once. (Established by the `384` lane, checked both directions; recorded
  here because it is the reason the guard is load-bearing rather than tidy.)

## 20. Status update, 2026-09-03 15:10 UTC — the page is REPAIRED and DEPLOYED, and 454 is CLOSED

**Chassis rolled a second time** (`v1.0.1359`, commit `3043885191b20a0e9b83594b2002e8805fbe95ec`)
carrying both `9831e9ab4` and `29b40e8bc`. Re-dispatched `page-rerender` immediately
(correlation `53f08444-1c00-4265-a641-d4d32eedf8d0`) — result: **`COMPLETED`**, through
`save_sections`, `render_page` and `deploy_page`, no `__step_error`.

**The fixture is on the page:**

| | before | after |
|---|---|---|
| `event-list` items | 0 | **1** |
| `event-list` rendered_html | 1,813 B | **2,498 B** |

**And the deploy is real**, traced past the DB row: `deploy_result.response.data.commit_sha =
0cc6da28b4fc18e59ff9df1a995ce3cc943bc094`; GitHub Actions run `33771117580` completed
**success**, its own "Sync to B2" step showing `delete tools/fight-calendar/index.html (old
version)` then `upload tools/fight-calendar/index.html`. `portfolio-sites/boxingonline.com`
now carries the fight fixture.

**`bugs_open/454` is CLOSED** (moved to `bugs_closed/`, same commit round). Full closure record
there, including two corrections from the `components` lane after independent corroboration
(their own two-day-old `bugs_open/425` regression was this same mechanism) and one enumerated
exception the fix cannot reach (`component_id NULL` rows, `bugs_open/457`).

**The public domain and the preview subdomain still show the empty state, and that is
EXPECTED, not a new problem.** `sites.handed_over_at IS NULL` for boxingonline.com — it is
pre-handover and not DNS-live at `boxingonline.com` at all. The preview
(`boxingonline.ugg2.com`) is served by `site-publisher`, a separate reconciliation pipeline on
its own tick (established 2026-09-02/03; checked again today via `WebFetch`, still the empty
state, not chased — same call as before, and not this lane's job).

**What is left before this lane can actually close, updated:**

- [x] ~~Diagnose and fix the `query.upcoming_events` items-not-populating defect~~ — **DONE**,
  `bugs_open/454` **CLOSED**.
- [x] ~~The save must be allowed to complete~~ — **DONE**, `bugs_open/450`'s guard fix (`29b40e8bc`)
  is live; this session's own dispatch proved it.
- [ ] **`719`/`727`/`728` are TRANSIENT** (§19.2–19.4) — the next `sync_pages` for
  boxingonline.com will revert `pages.sections` to the stale plan and undo all three. **This is
  now the lane's top open item.** Needs a council/owner decision about `site_plan_sections`
  before a durable fix can be written; not a migration.
- [ ] `ff91e666` is **APPROVED** (round 3, 12:41:09Z) — no further council action needed.
- [ ] Re-verify `experience_loop`'s nightly reclassification of
  `/tools/fight-calendar/index.html` out of "no calendar mechanism at all". **This is now the
  real remaining closing signal for 427 itself** — everything upstream of it is proven.
- [ ] `bugs_open/460` (why `blog-content-planner` stopped) — unowned, unrelated to closing 427.

## 21. Status update, 2026-09-03 evening (session "427", lane resumed) — the fix is DURABLE, and the closing signal cannot clear as things stand

Picked up from `HANDOFF_2026-09-03d_continue_here.md` ("final consolidated handoff for a
new chat"). The prior session was messaged, confirmed it was not still on it, and
independently re-verified both findings below at the source before agreeing.

### 21.1 The top open item is CLOSED, and its stated mechanism was wrong

§19.2's *"719/727/728 are TRANSIENT"* was right; its `sync_pages` mechanism was not. See the
CORRECTED block there. The reverting path is the tier-1 loader sync-down on any page
**build**, which needs no re-plan.

**Migration `750`** (committed, council `b290bef5`) renames the current plan's ordering-1
row `generic-text-block` → `event-list` and deletes the ordering-2 `advertising` row. Applied
by hand 2026-09-03 after an **induced-failure run** (a needle no write produces; the verify
`RAISE`d and the transaction rolled back with tier 1 unchanged).

`[MEASURED 2026-09-03, post-apply]` the loader's own query returns `hero-tool, event-list`;
its sync-down guard would update 0 rows; and the **artefact is byte-identical** —
`hero-tool`/3,859 B, `event-list`/2,498 B, `pages.updated_at` still `15:10:36.708975+00`.
That last is the point: this removes a latent revert and must not change output.

**Shaped as an in-place rename, NOT migration 154's delete-renumber-insert**, because
`ordering` is a positional join key for four consumers — `assigned_fact_ids` (where `'[]'`
and `NULL` are *different instructions* to the section writer), `subject`,
`page_components.position`, and `site_plan_imagery.scope_ref`, which for section scope is
literally `'<page>:<ordinal>'` (`[MEASURED]` live: `index:1`, `index:2`, `about:2`). 154 is
the wrong template today and the first one a reader finds; now a `LANDMINES.md` entry.

**Residual, stated:** this defends against the tier-1 loader. It does not make the page
immune to a genuine re-plan minting a *new* `site_plans` row, which no per-plan migration can
reach. That is the framework fix's half.

### 21.2 ⚠ The stated closing signal CANNOT clear — the page is still not a tool

§20 named the nightly `experience_loop` reclassification as *"the real remaining closing
signal for 427 itself"*. **It has already run twice since the fix and still flags the page.**

`[MEASURED 2026-09-03 15:14:47Z]`, i.e. four minutes *after* the deploy:
> `Rule B (tool page with nothing usable): 1`
> `[B] boxingonline.com/tools/fight-calendar/index.html: 6358 chars rendered, no control, no inline data, no runtime fetch — a page about a tool, not a tool.`

Verified at the artefact rather than taken from the check: both `page_components` are
`component_level='section'`, and neither `rendered_html` contains a `<button>`, `<input>`,
`<select>` or `<script>`. The `event-list` component itself is `render_mode='template'` with
no script and no control in its `html_template` — **a static section by design**. A static
`event-list` on a `page_type='tool'` page can never satisfy Rule B.

So the fixture is real and durable, and the page is still not a tool. **Owner decision
2026-09-03: build a real calendar mechanism** (rather than reclassify the page as the event
directory §6 discusses). That is now 427's remaining work.

**⚠ ORDERING, and it is load-bearing:** the revert had not fired only because
`page-build-handler`'s `load_page_record` carries `refuse_owned_page: true` and this page
satisfies the tool-shell predicate (`page_type='tool'`, **zero** `component_level='tool'`
components). **That refusal self-clears the moment a real tool component lands.** Building the
mechanism first would have armed the very revert 750 just removed. 750 shipped first,
deliberately.

### 21.3 Fleet work done alongside

- **Migration `753`** (committed, council `ca720d44`) cleared the `section_source_drift`
  backlog — six open items, oldest 2026-07-28. Closed **by predicate re-derived at apply
  time**, not by an id list, so `apis.uk/index` (a live divergence owned by another lane) was
  excluded by the data. Every receipt records a `direction`.
- **`bugs_open/469` filed**: 3 of 4 closed items were `authority_won` — the cache was
  *overwritten*. `robot-hands.com/gripper-catalog` has lost `gripper-spec-sheet`, **the very
  component migration 154 was written to rescue in July**, and `idea.uk/guides-index` lost
  `guide-list`. The loss has completed, twice, unremarked.
- **`apis.uk` lane notified** and fixed their own page the same hour (their migration `754`),
  using `750` as the template. They contributed a finding folded in above: IMG-075's section
  bindings key on `scope_ref` ordinals, so a renumbering correction breaks imagery where a
  rename cannot. Independently re-verified here before it was written down.
- Two `LANDMINES.md` entries and a `WRONG_CALLS.md` entry (the wrong-mechanism pair —
  including this session repeating the same shape an hour later while correcting it).

### 21.4 What is left

- [ ] **Build the calendar mechanism** (owner decision) — through the tool pipeline, never by
      hand. Closing signal: Rule B stops naming the page, read at `doc_notes`.
- [ ] **The framework fix** — an RFC plus a typed action so no lane hand-writes a three-store
      composition correction again. Architecture-scope: it narrows the immutability guarantee
      `reconcile_site_plan_action.go:596-601` rests on.
- [ ] `bugs_open/469` — whether the two destroyed components should be restored. Needs a human
      who knows what those pages are for.
- [x] ~~719/727/728 are TRANSIENT~~ — **DONE**, migration `750`.
- [x] ~~the `section_source_drift` backlog~~ — **DONE**, migration `753`.

## 22. Status update, 2026-09-03 evening — the tool pipeline built the MECHANISM correctly and FABRICATED its data, which is this bug's own root cause demonstrating itself

Per the owner's decision (§21.2), the calendar mechanism was dispatched through the framework
— `tool-generator`, correlation `77bbab58-950d-4076-af9a-e7c54b7ea0e1`, published through
`scripts/kafka-publish-lib.sh` with the receipt asserted (never hand-rolled `kcat`, which
exits 0 on publishing nothing).

### 22.1 What went right

The run `COMPLETED`, and — verified at the artefact, not from the status:

- a `content_components` row `tool-fight-calendar-boxingonline-com`, **`component_level='tool'`**,
  16,098-byte template, **with a `<script>` and with controls**. That is exactly what Rule B
  wants and what `event-list` structurally could not provide.
- **no duplicate page.** `UpsertPageForRole` attached to the existing
  `tool-fight-calendar` (`4b74ff1f`) as designed — its `Refresh: []string{}` is empty on
  purpose — and a companion guide page was created, consistent with the site's other four
  tools.

### 22.2 What went wrong, and why it is this bug and not a new one

`[MEASURED 2026-09-03]` the tool ships **12 inline fixtures that are fabricated and stale**:

- **Fabricated.** Real, named people with specific claims: `Canelo Alvarez vs Jermell Charlo,
  2025-05-03, T-Mobile Arena, DAZN`; `Tyson Fury vs Oleksandr Usyk, 2025-06-21, Kingdom
  Arena, Riyadh`; `Deontay Wilder vs Anthony Joshua`. **Nothing verified any of it against
  `evidence_base`.**
- **Stale.** Every date runs `2024-01-13` → `2025-12-06`. Today is **2026-09-03**. So under
  the tool's own "past fights drop out of the default view" behaviour, **the default view
  renders EMPTY** — the precise symptom this bug was filed about, reproduced by the fix for it.

**This is §3/§4's root cause restated, not a separate defect.** Nothing on the estate turns a
confirmed event into a dated fact, so when a generator needs fixtures it invents them. The
research spec's own `lessons.avoid[]` already names the harm: *"Stale calendar entries — a
wrong fight date actively harms readers."* Related but distinct: `bugs_open/449` (no fence the
tool generator writes ever asserts a number) is the numeric sibling of the same missing
control.

### 22.3 What I did about it

**Nothing fabricated has been deployed.** `pages.deployed_at` is still
`2026-09-03 15:10:36.708975+00` — the verified state from §20. The queued
`page_rerender` for this page (`f0fe578b`, `triaged`) was **cancelled**, with the reason
recorded in its `result`, so the invented fixtures cannot reach `portfolio-sites/` while this
is decided. The tool component was **kept**: the mechanism is right and only its data is
wrong, so deleting it would throw away the good half.

**This is the owner's call, and it is a real one:**

1. **Wire the tool's fixture array to real dated facts** — i.e. actually build what this bug
   asks for, and re-dispatch. The tool's brief already specified that the array be shaped so a
   later pipeline can regenerate it without changing the interface, so this is the intended path.
2. **Ship the mechanism with an empty/placeholder array** and an honest "fixtures to follow"
   state, so the page is a real tool with no false claims.
3. **Hand-curate a small set of genuinely verified upcoming fights** into `evidence_base` first,
   then regenerate.

**Option 1 is the only one that closes 427 rather than moving it**, but it is also the one
that needs the missing mechanism to exist. Recommend 2 as the interim if the site must be
delivered before that lands — an empty calendar that says so is not a false claim; twelve
invented ones are.

**Do not simply re-dispatch `tool-generator` and hope.** It produced this output from a brief
that explicitly asked for real upcoming events; the generator has no source of truth to draw
on, which is the whole of this bug.

### 22.4 Containment, and the near miss that proves one cancellation is not containment

**Cancelling the queued rerender was NOT sufficient, and I was wrong to treat it as such.**
`rerender_single_page_action.go:978-984` assembles **every** `page_components` row that is
`NotRemoved`, `ORDER BY position` — it does **not** filter by `pages.sections` membership. So
the fabricated component was assembly-eligible regardless of which work item fired, and four
further `page_rerender` items were queued on that site the same minute.

And a rebuild did fire, at **17:39:23**, from the `needs_content_page` items the toolgen run
itself created. It renumbered the tool to position 3, rewrote `hero-tool` (3,859→3,601 B) and
`event-list` (2,498→2,548 B), and **nulled the tool row's `component_id`** — `bugs_open/457`'s
class, which is why a containment `UPDATE` joined through `content_components` matched zero
rows. `pages.deployed_at` then moved to **17:39:50**.

> **CORRECTED 2026-09-03 (later) — THIS IS WRONG, AND IT IS THE MOST IMPORTANT CORRECTION IN
> THIS FILE. THE FABRICATED FIXTURES SHIPPED.**
>
> `[MEASURED 2026-09-03]`, `gh api repos/gqls/sites/contents/boxingonline.com/tools/
> fight-calendar/index.html` → **72,945 bytes**, containing `Jermell Charlo`, `Tyson Fury`,
> `Deontay Wilder` and `var FIGHTS`. Commit **`6cb1cdb1`**, 17:39:47Z, *"Rerender:
> tools/fight-calendar/index.html"*; GitHub Actions run **33785878793 "Deploy to B2",
> completed success, 17:39:50Z** — and still the newest commit on that path at the time the
> paragraph below was written.
>
> **The error:** the check below queried `boxingonline.ugg2.com`, the **preview**, served by
> `site-publisher` on its own reconciliation tick. It was returning **57,987 bytes** — the OLD,
> clean commit `0cc6da28` from 15:10Z. It was stale, and stale looked like success. The deploy
> target is a different surface, and it is the one **this file's own §13 and §20 already used
> as proof of deployment** (`deploy_result.response.data.commit_sha` + the "Deploy to B2"
> Actions run). The right instrument was in this file twice and I used a different one.
>
> **What actually happened:** the `needs_content_page` item the tool run itself raised
> (`create_tool_component_action.go:576`) drove page-build-handler → `save_page_sections`
> overwrite at 17:39:22Z (all three slots deleted and re-inserted, tool renumbered 2→3,
> `component_id` nulled) → deploy at 17:39:50Z. Cancelling `f0fe578b` removed one trigger of
> several, which §22.4 itself says — and then verified the consequence on the wrong surface.
>
> **Remediation:** an assemble-only `page_rerender` was filed (`1af13106`, priority 5). The
> tool placement is `build_status='removed'`, so assembly yields `hero-tool` + `event-list`
> only and overwrites the artefact. **Do not treat this bug as contained until `gh api` on
> that path returns ~57,987 bytes with `Charlo`=0 and `Mbilli`=1, and the "Deploy to B2" run
> for that sha is green.** Logged in `WRONG_CALLS.md`.
>
> **Duration of exposure:** live from 17:39:50Z until that rerender deploys. The site is
> pre-handover and not DNS-live at `boxingonline.com`, which limits who could see it — but it
> was published, and the owner had been told it was not.

**Verified at the artefact, not inferred: nothing fabricated was published.** `WebFetch` of
`https://boxingonline.ugg2.com/tools/fight-calendar/index.html` (the real publish target —
`sites.publish_target='b2worker'`, `publish_project='boxingonline.ugg2.com'`; the customer
domain is pre-handover and not DNS-live) names **only** Canelo Alvarez and Christian Mbilli,
dated 2026-10-31, with the citation headline. It does **not** contain Charlo, Fury, Usyk,
Wilder, Joshua or Inoue. The served page is the verified `event-list` output.

**Containment now, by store rather than by trigger:**

| store | state |
|---|---|
| `page_components` `0791ab91` | `build_status='removed'` → excluded by `NotRemoved`, so no assembly can pick it up |
| `content_components` `e5e8fa33` | `is_active=false` → cannot be re-rendered, re-forked, or `deploy_tool_to_site`'d onto another site |
| queued `page_rerender` `f0fe578b` | `cancelled`, reason in `result` |

`[MEASURED 2026-09-03]` `SELECT count(*) FROM content_components WHERE html_template ILIKE
'%Jermell Charlo%' AND is_active` → **0**.

The component is kept, not deleted: the *mechanism* is right (interactive, controls,
countdown — exactly what Rule B wants) and only its *data* is invented. It is the shell to
rewrite against `query.upcoming_events`, not something to throw away.

**The transferable lesson, for 016b §9:** on a page whose components are assembled by
`NotRemoved` rather than by the declared section list, **cancelling a dispatch is not
containment — it removes one trigger from many.** Contain at the STORE (`build_status`,
`is_active`), then re-query every store for the offending content and require the count to be
zero. And re-read the row identifiers immediately before writing: a concurrent framework
rebuild renumbered the position and nulled the `component_id` this containment had planned to
join on, inside two minutes.

### 22.5 Why nothing stopped it: the claims perimeter is structurally blind to `<script>`, and a tool's data is ALWAYS in a script

Asked what *should* have caught twelve invented fights about real named people on a paid
site, and traced it. The answer is that the estate's claims machinery could not see them, by
construction — not that it looked and was fooled.

**The shared implementation.** `datahelpers/claims.go` provides `ExtractAssertionText` /
`ScanBannedClaims` / `ScanUnregisteredNumbers`, and both claim surfaces use that one
implementation so they "agree by one literal implementation on what counts as an asserted
claim" — the build-time deploy gate (`validate_page_content` check 8) and the post-deploy
audit (`discovery_checks/check_unverified_claims.go`).

**The blind spot.** `claims.go:500` lists the non-assertion contexts:

```go
"script": true, "style": true, "noscript": true, "template": true,
```

`ExtractAssertionText` walks text nodes *outside* those. `check_unverified_claims`'s own
header says it scans "assertion TEXT NODES only, never raw HTML/attributes", with a landmine
about `placeholder="jane@…"` being an example rather than a claim — a correct rule for
attributes that happens to exclude script bodies too.

**Every one of the twelve fabricated fights lived in `var FIGHTS = [...]` inside `<script>`.**
So they were invisible to the gate and to the audit simultaneously, and both were working as
designed.

**This generalises, and that is the point.** A `component_level='tool'` component ships its
data *as code*, in a script block, by definition. So **no tool's data has ever been inside
the claims perimeter** — not this one, and not the estate's other tools. The two scans that
do run (banned claims, unregistered numbers) would not have caught it anyway: "Tyson Fury" is
not on any banned list, and `"2025-06-21"` is not an unregistered *number* in the stat sense.

**Consequences worth stating separately:**

- It is not enough to make this tool read `query.upcoming_events`. A future generator can put
  a literal array back, and nothing will object. The gate has to see script-borne data, or
  the write path has to refuse a tool that carries its own factual data.
- `bugs_open/449` ("no fence the tool generator writes ever asserts a number") is the
  numeric sibling of exactly this hole, one layer down. The two want one fence, not two.
- The honest scope claim: this is a **fleet-wide** blind spot in the claims perimeter, not a
  boxingonline defect. It should be measured across every tool component before anyone sizes
  the fix — do not assume this tool is the only one carrying invented facts.

**Scope, measured rather than assumed** `[MEASURED 2026-09-03]`. Fleet-wide there are **333
active `component_level='tool'` components**. **31** contain an ISO-date literal in their
`html_template`; **0** carry the fabricated shape (a `date:`-keyed object array) — that one
was unique, and is now `is_active=false`. Sampling the 31: they are overwhelmingly **dates in
HTML comments** — provenance and change annotations such as `(loancalculator.co.uk/tools/
interest-rate-stress-test.html, 2026-07-31)` and `/* THE VERDICT — FIXED 2026-08-03 */` — not
assertions. One (`tool-blue-carbon-estimator`) carries a visible *"Figures captured
2026-08-31"* line, which is a provenance claim and sits in a `<p>`, i.e. **inside** the
claims perimeter already.

So the honest reading is: **the fabrication looks like a first occurrence, not a widespread
pattern — but the hole that let it through is fleet-wide.** The checker should be sized to
close the hole, not to chase 31 comments. A detector keying on "any date literal in a tool
template" would be ~31/31 false positives on today's data; one keying on *structured
factual data* (an array of objects with date + named-entity fields) would have fired on
exactly one component: this one.

### 22.6 The fabrication is OFF the deploy target — verified with both controls

`[MEASURED 2026-09-03 19:13Z]`, at the surface the deploy actually writes, using the same
instrument §13/§20 use as proof of deployment.

Assemble-only `page_rerender` `1af13106` ran `triaged → claimed → complete` in ~63 seconds.
Because the tool placement is `build_status='removed'`, assembly yielded `hero-tool` +
`event-list` only.

| | before (`6cb1cdb1`, 17:39:47Z) | after (`f2cb3378`, 19:13:14Z) |
|---|---|---|
| bytes | 72,945 | **57,779** |
| `Jermell Charlo` / `Tyson Fury` / `Deontay Wilder` / `Naoya Inoue` | present | **0 / 0 / 0 / 0** |
| `var FIGHTS` | 1 | **0** |
| `Mbilli` (the one VERIFIED fact) | 1 | **1** |
| `Canelo` | — | **2** |

**Both directions checked, deliberately.** The negative controls prove the fabrication is
gone; the **positive** controls prove the page is not merely blank — the one evidenced fixture
survives. A measurement that only checked absence would read identically if the rerender had
emptied the page, which is the failure this bug is about.

Deploy receipt: GitHub Actions run **33795145748**, `f2cb3378`, *"Deploy to B2"*,
**completed / success**, 19:13:17Z. `f2cb3378` is the newest commit on that path.

**Exposure window: 17:39:50Z → 19:13:14Z, about 1 hour 33 minutes.** The site is pre-handover
and not DNS-live at `boxingonline.com`, so the audience was whoever reads the deploy repo or
the B2 bucket directly — but it was published, and for most of that window I was asserting it
had not been.

**Still true and unchanged:** the tool component (`e5e8fa33`) remains `is_active=false` and its
placement `removed`, so nothing can re-ship it; and `[MEASURED]` zero active stores hold the
fabricated text. The *mechanism* is still the right shape and is still the shell to rewrite
against `query.upcoming_events` — see §23.

## 23. The verified-facts plan, corrected by review — owner requirement: "we can't have invented fights, ever"

The draft plan below was put through a `Plan` review instructed to critique rather than agree.
**It found the draft would have caused a second incident**, so the corrections are recorded
before the plan, not after it.

### 23.1 What the review overturned

- **Tools ARE data-bound, but only on one path.** At birth and on the tool-birth deploy the
  template ships verbatim. On a **sections-branch** rerender (`reason` ∈ `image_landed`,
  `section_data_resolved`, `cta_links_stale`, `template_changed`, `literal_markdown`),
  `rerenderFlatSections` walks every stored row — tool rows included — through
  `planSection` → `RenderTemplate`, and **no `component_level` predicate exists anywhere in
  that path**. `queryresolve.ConsumerPages` has no level filter either, so a tool declaring
  `query.upcoming_events` *would* receive the daily evidence-refresh rerender.
  `content_components.data_sources` has **zero Go readers** — `input_schema` is the mechanism.
- **⚠ The draft's "just give the tool an `input_schema`" step was a latent full-rebuild
  trigger.** `rerender_page_sections_action.go:428-460` escalates any non-self-contained row
  with empty `content_data` to `page-build-handler` and returns *before rendering*. A schema
  makes the tool non-self-contained (`isSelfContainedSection` requires an empty schema); birth
  writes `content_data='{}'`. So the first such rerender would have triggered **a full LLM
  rebuild of a paid page** — which is exactly what produced the 17:39 incident. Seeding
  `content_data` non-empty at birth is not a detail; it is what makes the change safe.
- **The JSON-in-`<script>` idea was wrong twice over.** `RenderTemplate`'s FuncMap has **no
  JSON encoder**, and the resolver HTML-escapes strings (right for text, wrong inside JSON).
  Worse, it would keep the fixtures **outside** the claims perimeter — §22.5's whole point.
  The correct shape renders fixtures as **visible elements carrying `data-*` attributes**
  (`data-date`, `data-fact-id`, …) with JS reading `dataset`: inside the perimeter, and a
  no-JS fallback for free. `event-list`'s template is the idiom to mirror.
- **The brief I wrote invited the fabrication.** Orchestration `77bbab58`'s spec says verbatim
  *"The fixture data ships inline in the component as a JSON array"*, and the
  `generate_tool_html` prompt (5,115 chars) has 22 rules about colours, ids and IIFEs and
  **nothing** about provenance, `evidence_base` or not inventing facts. **Any fix that starts
  with "re-dispatch" without changing the prompt reproduces this.**
- **Re-dispatch currently fails loudly anyway**, because of my own containment: the
  already-exists probe requires `cc.is_active`, so it misses → the CREATE arm runs → `INSERT`
  dies on `content_components_name_key`. Worth knowing before anyone tries.

### 23.2 The three checker layers

1. **Arm `check_event_fixture_completeness`** — built, council-reviewed, and armed on **zero**
   of five discovery agents. Its commit `d6a952249` **is** an ancestor of the live `v1.0.1359`,
   so arming cannot fail the run step. `[MEASURED 2026-09-03]` blast radius: 12 sites carry
   dated facts, 64 dated facts fleet-wide, **0 unevidenced** — so it fires ~nothing today and
   establishes the baseline. **And it retracts**: it is one of only three `capability_gap`
   producers that populate `CheckResult.Resolved` (`:160`), which matters because
   `capability_gap` is the estate's least-closed type — `[MEASURED]` **334 filed, 1 closed**
   across `site_work_items` UNION its archive. Arming a *retracting* check into that pile is
   safe; arming a non-retracting one would not be.
2. **Refuse literal facts at tool BIRTH** — in `create_tool_component_action.go` beside
   `HasToolDocHeader`, before the already-exists branch so both arms are covered. Narrow shape:
   inside `<script>`, an array literal of objects carrying an ISO-date-valued key and ≥2
   name-like string fields, where the declared schema has no `query.*` source. `[MEASURED
   2026-09-03]` 333 active tools, 31 with a date literal (**all comments/provenance**), **0**
   with this shape — so the measured false-positive rate is **zero** and the owner's "never" is
   satisfiable now. **Sequencing: the generator must be able to express the good shape first**,
   or the refusal makes calendars unbuildable.
3. **The checker the requirement actually implies, on the artefact:** every `data-fact-id` in a
   served component must resolve to a current `evidence_base` fact carrying citation url AND
   quote. That closes "invented fights" *whatever produced them* — generator, hand-migration,
   or tool-improver — and `query.upcoming_events` already emits `fact_id`, so the data side is
   free. This is the one that would have caught the incident.

**Explicitly NOT in scope here:** widening `ExtractAssertionText` to script bodies. It changes
what the shared claims gate guarantees → architecture-scope under the 2026-07-29 ruling. Route
as its own RFC citing §22.5. And `bugs_open/449` is the **numeric sibling**, not the same fence
— share its REGISTER vocabulary (`data-fact-id` is exactly "derives from a fact id"), cross-file,
do not extend it (`scripts/who-owns.py 449` first; its lane is active).

### 23.3 Status

Not built. Awaiting the owner's go-ahead, and layer 2 depends on layer 1 of the *generator*
change landing first. The verified record is ready: `CIT-5b2cc9894bfc475f`, `event_date`
2026-10-31, `participants`, full citation, `verified_at` — one real fight, which is the honest
content for this page today.

## 24. Status update, 2026-09-04 — a PRE-EXISTING instance of §22's exact root cause found on two OTHER tools, and it would evade all three §23.2 checkers as currently scoped

Filed by the `calendar` session as its own bug, **`bugs_open/482`**, after a report relayed by
`boxingonline.com` — verified independently, corrected on one figure, then cross-checked
against this file only after filing (a `grep`/re-read gap on my part, worth naming rather
than hiding: 427 had grown to 1,600+ lines since I last read it in full, and I filed before
re-reading the tail). **Not a duplicate — genuinely a different pair of tools** —
`tool-fight-countdown` and `tool-fighter-comparator`, both pre-existing (built long before
§22's `tool-generator` dispatch), not the fight-calendar tool this section is about. But it
is the **identical mechanism**: a generator with nothing real to draw on fabricated
plausible-sounding dated content, and nobody had caught it until now.

**Why it matters here rather than only in 482: it evades all three of §23.2's checkers, and
that is worth knowing before any of them are built.**

- **Layer 2 (birth-time refusal) would NOT have caught it.** Its stated shape is *"an array
  literal of objects carrying an **ISO-date-valued key**"*. `tool-fight-countdown`'s fixtures
  are `{ year: 2025, month: 5, day: 14, ... }` — three separate numeric fields, constructed
  into a date via `Date.UTC(fight.year, fight.month, fight.day, ...)`, never an ISO string
  anywhere in the component. `[MEASURED 2026-09-04]` a plain `grep -oE
  '202[0-9]-[0-9]{2}-[0-9]{2}'` over its 13,475-byte `rendered_html` returns **zero** hits —
  confirmed directly, not inferred. **So the "0 matches, zero false positives" figure in
  §23.2 is correct for its own definition and still has a live blind spot**: a
  year/month/day numeric triplet is the same fabrication in a shape the stated pattern does
  not look for. Widen it before relying on "satisfiable now, zero false positives" as the
  reason to ship layer 2 as scoped.
- **Layer 3 (`data-fact-id` resolution) would NOT have caught it either, for a different
  reason.** `tool-fight-countdown` carries no `data-*` attributes at all — its fixtures are
  bare JS object literals, not rendered as elements with a `data-fact-id`. A checker that
  validates every `data-fact-id` resolves to a real fact never looks at this component in
  the first place; it is silently out of scope rather than silently passing. The same is
  true of `tool-fighter-comparator`, which has no fixtures to validate (0 options, 0
  selects) — a different failure the checker was never aimed at either.
- **Layer 1 (`check_event_fixture_completeness`) is about the wrong side of the seam** — it
  checks *facts declared without evidence*, not *tool content with no fact behind it*. Not
  applicable to either shape found here.

**So `bugs_open/482`'s finding is not "one more site with the same bug" — it is a working
counter-example against the specific detection shapes §23.2 proposes**, found by reading
one pre-existing tool's actual JS rather than by running any of the three checkers (none of
which existed yet to run). Whoever builds §23's plan should read `482` first: at minimum,
layer 2's pattern needs to cover numeric year/month/day date construction alongside ISO
strings, and a **census pass over already-built tools** is needed regardless of which
layers ship — birth-time refusal only stops new fabrication, and at least one violation
(now two, counting the comparator's emptiness as the sibling failure mode) already shipped
before any checker existed to refuse it.

Cross-referenced both ways. Not this section's job to re-plan §23 — recorded so the plan is
built against the fuller picture, not the one available when it was drafted.
