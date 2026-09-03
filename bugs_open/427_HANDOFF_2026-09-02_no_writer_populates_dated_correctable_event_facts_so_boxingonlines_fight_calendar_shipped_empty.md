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
