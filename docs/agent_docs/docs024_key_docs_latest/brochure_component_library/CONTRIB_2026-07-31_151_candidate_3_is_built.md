# CONTRIB — your bug 151's candidate 3 is built. Candidate 1 is still yours.

From the `gauntlet_dead_cta` lane, 2026-07-31. **Appended as a contribution; nothing
of yours edited.** `who-owns.py 151` names you as the owner (ACTIVE, 113 commits/14d),
which is why this is a note rather than a change to your files.

## What landed, and why it was built outside your lane

The owner asked us to run your duplication census against vonc.com, then asked
whether it could be wired in as a checker with a handler. That is your candidate 3
verbatim:

> **A post-build fact-repetition census as a gate**, analogous to the claims gate
> … it is the only candidate here that also protects the 9 already-deployed sites
> while (1) is being built.

Committed (one commit, `feat(151 cand 3)`), council-submitted
`da3f2d9b-ae6f-492d-ad3b-748323b66367`:

| piece | what it is |
|---|---|
| `discovery_checks/check_content_duplication.go` | the census, permanent. Registers a verifier too |
| `actions/remove_duplicate_page_sections_action.go` | the deterministic repair |
| `datahelpers.NormaliseSectionText` / `SectionTokenSet` | **shared** — call these rather than writing a third normaliser |
| `sql_for_agents/269_*.sql` | the `deduplicate-sections` handler, seeded and live |
| `gauntlet_dead_cta/scripts/dedup_census.py` | the by-hand version, site-agnostic, for ad-hoc runs |

## What it deliberately does NOT do — this is the part that matters to you

**Owner ruling 2026-07-31: the handler gets authority only over the deterministic
half.** The check splits its population:

- **in-remit** — same page, content-*identical* sections → dispatched, deleted,
  page re-rendered. No LLM anywhere in that path.
- **residue** — near-duplicate and cross-page copy → **one `capability_gap` per
  site, with no handler**, whose spec names *your* candidate 1 as the structural
  fix and carries `do_not_auto_rewrite: true`.

So the residue is now *queued and counted* rather than invisible, and nothing will
try to rewrite it. When candidate 1 lands, those capability_gap items are the
population it should clear.

The reasoning, because it is a constraint on candidate 1 too: rewriting sibling
copy means deciding what each page is *for*, and a similarity score says nothing
about meaning. On 2026-07-29 an LLM rewrite on gamesdesign.co.uk turned "built
**for** designers" into "built **by** a designer" and generated supporting detail
to justify the invention — a fabricated human credential, live for ~44 minutes
(`bugs_open/149` § C1, now WITNESSED). Any rewrite path needs the claims gate in
front of it.

## Three findings that change your bug's own text

**1. Your method has a FLOOR, and it is not documented in 151.** Measured on vonc:
4 approved facts against fundamentallyai's 9, so of 55 sections **none** can reach
a 3-fact overlap. The fact half cannot fire at all there. *A clean fact census is
indistinguishable from a clean site* — and vonc was the worse of the two sites,
with a page rendering every section twice. The check now reports
`fact_census_blind: true` with the pool size below a floor of 6 rather than
returning quietly clean. Already appended to 151.

**2. The discriminator is content identity, NEVER slot name.** Fleet-wide, 17
duplicate `(page_id, slot_name)` groups exist and **11 are legitimate** — repeated
slots with differing content, `generic-text-block` ×2–3 on five sites. A unique
index or a slot-keyed rule would delete real sections on those five. There is a
test asserting that false positive stays unflagged; please keep it if you refactor
around here.

**3. There is a second, unrelated duplication defect** that your bug's framing does
not cover: *duplicate rows*, as opposed to independently-generated near-duplicate
copy. vonc's about page had 12 `page_components` rows that were 6 identical pairs
and painted twice for two days. Filed as `bugs_open/156` by another thread, cause
`[UNRECOVERABLE]` (the run's `collected_data` aged out). Worth knowing it exists so
151 does not get blamed for it — different cause, different fix, same symptom in
any census output.

## What you need from us / what to watch

- **The check is INERT.** A discovery check runs only when an agent's workflow
  config names it, and none does. Opting it in requires: a chassis image carrying
  `remove_duplicate_page_sections`, pod-grepped with a control, **then** adding
  `content_duplication` to a discovery agent's check list. We have not done that
  and will not without a word — the first site it runs against will start deleting
  rows, and fundamentallyai is yours.
- **If you want it on fundamentallyai first**, say so and we will coordinate the
  order rather than both enabling it.
- `220_claimed_item_timeout_generic_evidence.sql` gained `content_duplication` in
  its exclusion list. If you touch that migration, keep the entry — without it the
  timeout sweep auto-completes these items on orchestration evidence and bypasses
  the verifier.

Nothing here is a request. If you would rather own the checker too, take it — it is
your bug, and we would rather hand it over than have two implementations.
