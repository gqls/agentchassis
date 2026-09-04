# HANDOFF — bugfix 463, a section index's new children — 2026-09-04, continue here

**Read this first, then `bugs_open/463_HANDOFF_2026-09-03_validate_site_plan_pass_c_silently_drops_every_new_child_page_of_a_section_index.md`.**
Lane dir: `docs/agent_docs/docs024_key_docs_latest/bugfix_463_section_children/`.

---

## 1. State in one paragraph

The fix is **written, committed, council-APPROVED, and LIVE in the running chassis since
2026-09-03 22:07:19Z**. It has **never been exercised** — zero plan runs have happened since the
roll, so it is proven in the binary and in tests but not at the artefact. **One re-plan of
gamedesign.uk closes it, and that re-plan is ENQUEUED** — `needs_briefing` `6430726e`, enqueued
2026-09-04 10:54:36Z by the `gamedesign.uk` lane, triaged and waiting on the fleet selector. They
will report the correlation and steps 1/2/3. Both lanes that were holding have been cleared and
have each re-run the liveness probe independently. Nothing is blocked on code or on this lane.

## 2. What was wrong (both halves — the second is not in the original bug report)

**A.** `reconcilePlanWithRealised`'s Pass C compared the FIRST PATH SEGMENT of a planned page
against a realised section index's stem. A child (`/articles/x.html`) and a collider
(`/articles.html`) both reduce to `articles`, so every newly planned child of a section index was
deleted — silently, orchestration COMPLETED, no finding. Pass A's union restores *realised* pages
immediately after, so the damage is invisible except on a hub that is empty **today**, which can
then never be filled. `[MEASURED 2026-09-03]` 53 of 78 hubs across 21 sites are in that state.

**B.** Fixing A alone would have changed **nothing on the served page**. Both write surfaces
discard the planner's URL and re-derive it from `CanonicalisePage`, whose leaf-role arms default
the directory (`blog-post` → `/blog/`). `ValidateRoles` never derived `parent_section`, and its
rule 5 rescues only `/tools/`, `/guides/`, `/games/`. `[MEASURED]` the live `plan_site` prompt
(32,191 chars) never contains the string `parent_section`, and 109 of 109 `blog-post` rows in
`site_plan_pages` have it empty. **`bugs_open/463` §5 says the bug is "NOT about `parent_section`"
— true of Pass C, false of the write path.** Corrected in place in that file.

## 3. THE ONE OUTSTANDING ACTION

Re-plan **gamedesign.uk** (9 pages, hub `articles-index` at `/articles/index.html`, 0 children).
It must be that site or another **under `bugs_open/467`'s 20-page cap** — on the 26 of 42 sites
already over it, 467 discards every net-new page immediately afterwards and the test proves
nothing. The `gamedesign.uk` lane owns the site and has offered to run it; it is their dispatch,
not this lane's.

Then assert **in this order**. The middle step is the one that matters — a Pass-C-only fix passes
the first and fails the second, and the served page cannot tell them apart:

```sql
-- 1. step boundary: proposed must equal survived
SELECT jsonb_array_length(collected_data->'plan_site'->'result'->'pages') AS proposed,
       jsonb_array_length(collected_data->'validate_plan'->'pages')       AS survived
  FROM orchestration_states WHERE correlation_id='<corr>';

-- 2. PLACEMENT - the assertion a half-fix fails
SELECT spp.name, spp.role, spp.url, spp.parent_section
  FROM site_plan_pages spp JOIN site_plans sp ON sp.id=spp.plan_id
  JOIN sites s ON s.id=sp.site_id
 WHERE s.domain='gamedesign.uk' AND spp.role='blog-post'
 ORDER BY sp.created_at DESC;
-- PASS: url LIKE '/articles/%' AND parent_section = 'articles'
-- FAIL: url LIKE '/blog/%'  -> half B did not take effect
```

3. **Then** the served hub. If 1 and 2 pass but the hub still renders empty, that is
   **`bugs_open/457`** (`rebuild_blog_listing`, another lane's, in flight) — not this bug.

## 4. Decisions waiting on the owner

> **OWNER RULINGS 2026-09-04 — three of the four below are now DECIDED. Read these first;
> the numbered items keep their original wording as the record of what was asked.**
>
> 1. **`bugs_open/467` — ruled: a re-plan may add up to TEN new pages.** The cap is to bound
>    what a re-plan **adds**, not what a site may **contain**. Note for whoever implements it:
>    `max_pages` is read from step config (`v3_site_actions.go:3834`, defaulting to 20 only
>    when absent), so it is DB config and live immediately — but the ADD-budget is a code
>    change, because today `truncatePreservingRealised` derives the net-new budget as
>    `maxPages - len(keep)` and returns `keep` alone once `len(keep) >= maxPages`. Do not
>    "implement" this by raising `max_pages`: that raises the site ceiling, which is the
>    thing the owner did NOT ask for.
> 2. **463's closing bar — ruled: HOLD IT OPEN** until §3 passes at the artefact, as this
>    lane recommended. "Fixed AND live" is met and is deliberately not being used here.
> 3. **Scheduling the gamedesign re-plan — ruled: leave it in the natural queue.** Do not
>    spend a build cycle expediting `6430726e`. The `gamedesign.uk` lane has been told.
> 4. **`468`/`460` — NOT decided, and deliberately widened.** The owner's framing is that the
>    chosen directory paths ("blog, blogs, guides") and the near-identical filenames for
>    duplicate entries are *"a constant cause of bugs"*, and he asked for it to be looked at
>    properly rather than patched. See §9 for the census and the correspondence. **Do not fix
>    468 by threading one more field through one more producer until the vocabulary question
>    is answered** — that is the shape that produced 58 directories.


1. **`bugs_open/467` — the 20-page cap.** `truncatePreservingRealised` discards *every* net-new
   page once a site's preserved set reaches `max_pages` (20), not just the excess. `[MEASURED]`
   **26 of 42 sites are already past it**, largest 151. So on most of the estate a re-plan can
   currently add nothing. The engineering fix is easy — cap what a re-plan **adds** rather than
   what a site may **contain** — but *how many new pages a re-plan may add* is a product decision.
   Filed, unowned, not started.
2. **Closing bar for 463.** CLAUDE.md's bar is "fixed AND live", which is now met. This lane
   recommends **holding it open** until §3 passes, precisely because 463's own lesson is that a
   plausible fix can change nothing at the artefact. Owner's call.
3. **Whether to schedule the gamedesign re-plan** (it costs a build cycle) or wait for the site's
   next natural rebuild.
4. **`bugs_open/468`** (`create_blog_posts` never threads `ParentSection`) and **`bugs_open/460`**
   (its producer, `blog-content-planner`, dormant since 2026-04-24). Reviving that producer is the
   `feed lane`'s preferred route for designblog's `/the-design-feed/`. Unowned.

## 5. Traps — read before touching anything here

- **Check liveness at `service_binary_capabilities`, and filter by POD.** ⚠ *Corrected
  2026-09-04 — an earlier version of this handoff said "do not confirm this deploy by commit
  sha", and that was wrong.* The sha route works fine; I simply never found the stamp and so
  tested shas I had guessed. The authoritative query:
  ```sql
  SELECT pod_name, git_commit, started_at FROM service_binary_capabilities
   WHERE kind='build' AND pod_name LIKE 'agent-chassis-%' ORDER BY started_at DESC;
  ```
  then `git merge-base --is-ancestor <your commit> <stamp>`. Filter by `pod_name`, not by the
  `service` column — that column also carries rows for other pods on the same image
  (`agent-landmine-verifier-*` etc.) which may have started at a different time and could be
  running something else.
  ⚠ **That table is a TWO-HOUR WINDOW, not a history** (`LANDMINES.md`, "…IS A TWO-HOUR WINDOW…";
  `RetentionWindow` in `platform/buildcapability/buildcapability.go`). It answers *what is running
  now*. It **cannot** date a past event — and it fails silently, because a surviving pod's
  `started_at` can precede your event and read as proof that binary served it. **This matters for
  the verification below:** when the re-plan lands, do NOT reach for this table to ask "was the fix
  live when it ran". Corroborate with something that is not pruned —
  `kubectl -n ai-persona-system get rs -l app=agent-chassis --sort-by=.metadata.creationTimestamp`,
  which is how the 22:07:19Z roll time in §1 was established.
  What genuinely is NOT evidence here: the **image tag** (unchanged, `v1.0.1360` either side)
  and the **`build provenance` log line** (scrolled; reading the whole log OOM-kills the tool).
- **The capability probe is the fallback when your change adds a literal** — and it needs a
  present-control or it is unreadable:
  ```bash
  POD=$(kubectl -n ai-persona-system get pods -o name | grep -m1 agent-chassis | cut -d/ -f2)
  kubectl -n ai-persona-system exec $POD -- grep -aq "dropped flat page colliding with realised section index" /proc/1/exe  # present-control: MUST hit
  kubectl -n ai-persona-system exec $POD -- grep -aq "path collides with realised section index" /proc/1/exe                 # the fix: hits => live
  ```
  Both methods were run here and agree.
- **An earlier "9b540c2e6 is not live" reading EXPIRED, it was not refuted.** It was correct when
  taken (before the 22:07Z roll) and left two lanes holding for twelve hours. A liveness check on
  this estate has a shelf life of exactly one roll.
- **`/guides/`, `/tools/`, `/games/` are not neutral test fixtures** — `ValidateRoles` rule 5
  retypes pages there before the new derivation is reached, so a fixture in one of them measures
  rule 5. Use `/articles/`.
- **When modelling a Go helper in SQL, count its branches first.** `sectionStemOf` has two; a
  one-branch predicate gave me a plausible wrong answer (`WRONG_CALLS.md`, 2026-09-03).
- **HEAD may be red for reasons that are not yours.** It was when this lane started, from another
  lane's `83407cd37`. Run `scripts/verify-head-builds.sh --test ./platform/orchestration/...`
  against HEAD **alone** before attributing a failure to your change.
- ⚠ **`CLAUDE.md`'s "Ask the service what it is running" section is STALE relative to the landmine
  corpus, and following it is what cost this lane an hour.** It prescribes the `build provenance`
  log line first (which scrolls) and a binary probe as fallback, and never names
  `service_binary_capabilities` — which `LANDMINES.md` documents in nine places, including the
  retention caveat above and the rule that INERT code must be verified by ancestry rather than by
  literal. The corpus had the answer; the file every session loads unasked did not, and the
  SessionStart hook only surfaces landmines for files already dirty in the tree. Raised with the
  owner as a proposed CLAUDE.md edit; not made unilaterally.
  > **CORRECTED 2026-09-04 — DONE, the owner approved it and the edit is committed (`6a27fc59a`).**
  > `CLAUDE.md` now leads with the `service_binary_capabilities` query (verified against the live
  > table before it was written in), says to filter by `pod_name` not `service`, carries the
  > two-hour `RetentionWindow` caveat with the ReplicaSet list as the un-pruned corroboration,
  > demotes the log line and the binary probe to fallbacks, and adds the 440 lane's inert-code
  > rule (verify by ANCESTRY, never by literal). **So do not re-raise this, and do not read the
  > paragraph above as describing the file you have loaded** — it describes the file as it stood
  > before that commit.
- Full prospective entry: `LANDMINES.md`, "Comparing pages by their FIRST PATH SEGMENT…".

## 6. Everything committed

| commit | what |
|---|---|
| `9b540c2e6` | both code halves + mutation-proven tests (`Council-Submitted:` trailer) |
| `244651c03` | phantom-hub + mirror tests |
| `d4604b17b` | 463 §5 corrected in place, §9 added |
| `1d034bf8d` | 463 §10 — the phantom hub |
| `829dc0514` | 463 §9 cross-linked to `bugs_open/468` |
| `03ac14338` | `bugs_open/467` filed |
| `dff46487a` | `WRONG_CALLS.md` — my two-branch census error |
| `1999a4863` | `LANDMINES.md` entry (synced + verifier armed) |
| plus | the standing five in this directory |

Council correlation `9f6c6374-1b76-4094-9b4c-e04808d8428c` — round 1 REVISE, round 2 **APPROVED**
2026-09-03 17:26:38Z. The commit carries `Council-Submitted:`, so `098` credits it automatically;
no amend is needed and forward-only forbids one.

## 7. A finding worth knowing beyond this bug (§10 of the bug file)

`sectionStemOf` treats **any** non-root URL ending `/index.html` as a section index whatever its
`page_type` — and that is `CanonicalisePage`'s **default** shape for a tool, guide or game. So an
ordinary realised tool page registered as a phantom hub claiming the stem `tools`, and under the
old rule a newly planned sibling collided with it. **Mechanism proven by mutation; damage
deliberately NOT claimed** — 365 such rows on 39 sites is the population where it *can* fire, and
when I tried to measure losses the strong reading collapsed (110 of 171 post-Pass-C tool plan rows
are restorations, tools are also minted outside the plan path, and first plans skip Pass C). The
same fix covers it.

## 8. Peer lanes

- **`gamedesign.uk`** — filed 463, owns the site, holding a re-plan pending clearance. **Tell them
  it is live.**
- **`designblog.co.uk`** — same, via gamedesign; also corrected their own migration 732 rationale
  after this lane's finding that `isSectionIndexType` exempts a proposed hub from Pass C.
- **`428`** — shipped `recommended_type_reconciliation.go`, the stage-level detector; took a
  blocking caveat that `dropped_in_validation` for blog-post under a section prefix should now
  fall to zero, with the always-written audit row as the demand control. Do not read that zero as
  their detector going quiet.
- **`bugs_open/444`** — its gate is the guard in series; its `section_children` filings should now
  fall for the same reason. Not reachable by name when last tried; the interlock is written into
  463 §4 and §9.
- **`feed lane`** — filed `bugs_open/468`. **`portfolio_positioning`** — watching
  copyonline.co.uk (released 2026-09-03 15:49Z, no plan yet) as a possible live instance.

## 9. The directory-vocabulary question (owner-opened, 2026-09-04) — census + who holds what

The owner widened decision 4 above from "how do we fix 468" to "why do the directory paths keep
causing bugs". This section is the evidence gathered for it; it is **not** a proposal, and per
CLAUDE.md a shared directory vocabulary is architecture-scope, so it should end in an RFC.

**`[MEASURED 2026-09-04, live DB]`** — re-run before quoting; a census goes stale by ADDITION:

| what | figure |
|---|---|
| distinct first-path-segment directories, estate-wide | **56** active / 58 all statuses |
| …existing on exactly ONE site | **45** active / 47 all statuses |
| …containing exactly ONE page | **37** active / 39 all statuses |
| the head, i.e. the hardcoded role defaults | `tools` 40 sites · `guides` 39 · `blog` 21 |
| flat/nested twins (`/X.html` + `/X/index.html`, one site) | **7 pages on 3 sites** |
| tool + companion-guide twins | **38 pages on 10 sites** — 26 under `/guides/`, 12 under `/blog/` |

Near-synonym clusters in the tail: `guides`/`guias`/`buying-guides` · `articles`/`insights`/
`news`/`noticias` · `directory`/`brands`/`brand-directory`/`uk-studios-directory` · `report`/
`reports`. `homegarden.uk` additionally carries twelve month-name section indexes at site root.

**The mechanism, read from source (`platform/orchestration/datahelpers/page_canonical.go:129`).**
`CanonicalisePage` hardcodes a default directory per role — `tool`→`tools`, `guide`→`guides`,
`game`→`games`, `blog-post`→`blog`, `entity-page`→`entities` — and the **only** override is
`ParentSection`, which is `normaliseSlug` of arbitrary LLM text with **no allow-list and no
site-level registry**. `content`, `landing` and the section-index family deliberately refuse
`ParentSection` altogether. So a page's directory is either a Go literal or unconstrained model
output, with nothing in between. That is the whole explanation for the 58.

**The flat/nested twin is 463's own collider.** `sectionStemOf` reduces `/news.html` and
`/news/index.html` to the same stem — which is exactly the comparison that made Pass C delete
children. Fixing the vocabulary and fixing 463 are the same subject.

**Who holds the pieces** (correspondence sent 2026-09-04; replies not yet in at time of writing):
- **`bugs_open/241`, owned by the `loancalculator_couk` lane** (`scripts/who-owns.py 241`,
  ACTIVE, 38 commits/14d) — the flat-vs-nested half. Its own text says the representational
  half (`FlatURLs`) is committed and **the plumbing half is not**. Asked whether that still
  holds and whether they formed a view on registry-vs-allow-list.
- **`feed lane`** — filed `468`; asked whether a per-site registry would close it or whether
  `create_blog_posts` still needs `ParentSection` threaded regardless, and whether reviving
  `blog-content-planner` (`460`) is still their preferred route.
- **`site design planner`** — asked the three planner-side questions: whether `pages.site_area_id`
  is already a declared-section seam the planner reads or writes; whether the model should be
  told the site's existing directories or never choose at all; and whether `content`/`landing`
  refusing `ParentSection` is load-bearing.
- **`gamedesign.uk`** — told the two rulings; their re-plan is now also the reference case for
  what correct placement looks like.

~~**Open question this lane could not answer from source alone:** whether `pages.site_area_id` /
`site_areas` already constitutes a per-site section registry. If it does, this is much smaller
than an RFC. `[UNVERIFIED]` — do not assert either way until the planner lane answers or someone
reads the writers.~~

> **ANSWERED AND REFUTED 2026-09-04, by the `site-design-planner` lane — there is NO registry.**
> `[MEASURED]` `site_areas` holds **2 rows fleet-wide**, both `name='main'`,
> `url_prefix='/'`; `pages.site_area_id` is **NULL on all 1,362 pages**, zero exceptions; and
> the only Go reference to the symbol is a passthrough `SELECT` in
> `rerender_single_page_action.go:623` — nothing populates it and nothing acts on it.
> **So a declared-section registry is NEW CONSTRUCTION, not the extension of a working
> mechanism.** That raises the cost and makes the RFC route more clearly correct, not less.
> Do not let a future reader frame this as "site_areas already does half of it".

> **ROUTING CORRECTION, same exchange.** `site-design-planner` is **composition resolution
> only** — layout, palette, typography — and has never touched a page's URL, directory or
> role. The owning agent for `plan_site` / `page_canonical.go` is **`build-site-planner`**.
> This lane misrouted on the name; recorded so the next thread does not repeat it. Q2 (should
> the model be told the site's directories, or never choose) and Q3 (is `content`/`landing`
> refusing `ParentSection` load-bearing) were re-put to the `boxingonline.com` session as the
> live lane for `bugs_open/427`.

### 9a. Corrections and additions from the `feed lane`, 2026-09-04 — all independently re-run here

**CENSUS CORRECTED — my figures omitted a `status` filter.** The `feed lane` got 56/45/37 against
my 58/47/39 and identified the predicate difference exactly: dropping `status='active'`
reproduces mine byte for byte. **Re-run here and confirmed: 56 / 45 / 37 active.** So 2
directories and 2 single-page directories exist only on non-active rows. The table above now
carries both. Recorded because "close but not equal" between two lanes is what becomes a phantom
disagreement three documents later.

**THE `460` AXIS IS WRONG — "revive or replace" may both be false.** `check_empty_blog.go`, the
driver `460` names for the dormant `blog-content-planner`, gates on
`(page_type = 'blog-index' OR name = 'blog')` (`discovery_checks/check_empty_blog.go:30`,
verified here verbatim) and then counts `blog-post` rows. **`[MEASURED 2026-09-04, re-run by this
lane]`:**

| active listing hubs | count | visible to the gate? |
|---|---|---|
| `section-index` | **61** (28 sites) | **no** |
| `news-index` | **10** (10 sites) | **no** |
| `entity-directory` | **7** (7 sites) | **no** — *not in the feed lane's list; found here* |
| `blog-index` | **4** (4 sites) | yes |

**Sites with a blog hub: 4. Sites where the check would fire today: 0.** So the driver can see
**4 of 82** active listing hubs — about 5% — and every one of those is already served. **A
producer whose driver is scoped to a page type the estate stopped building is CORRECTLY IDLE,
and "ran 13 times then stopped" is what a satisfied backlog looks like as well as what a fault
looks like.** Both give silence on both instruments from the same day, which is the coincidence
that made `460` read as a fault.

⚠ **Nobody has run the separator, and it is cheap:** point the check at a site that genuinely
qualifies — a `blog-index` hub with zero `blog-post` rows — and see whether an item appears. An
item means the mechanism is alive and the population was simply empty; silence means the fault is
real and upstream of the agent. **Until someone runs it, do not write a cause into `460`** — the
feed lane deliberately contributed these as measurements with the inference labelled a candidate,
citing the 2026-07-31 ruling that fires on whoever asserts a cross-cutting root cause. Do not
undo that discipline by promoting it here.

**REGISTRY AND TRANSPORT ARE ORTHOGONAL, NOT SEQUENTIAL** (the feed lane's sharpening of `468`,
and it is the right correction). A registry decides **which strings are legal**; it cannot move a
value into a struct that has no field for it. `create_blog_posts` calls `CanonicalisePage` with a
two-field literal — `Role` and `Slug` — and reads nothing about a section from anywhere: not the
work item spec, not config, not the triggering page. **With a perfect registry in place the
article still lands in `/blog/`.** Both are needed; neither substitutes for the other.

**`468` is NECESSARY AND NOT SUFFICIENT for both of the feed lane's own cases.** Because
`check_empty_blog` can never fire for a `section-index` or a `news-index`, fixing 468 alone would
still leave nothing driving the producer for designblog's `/the-design-feed/` or advertise's
`/news/`. Worth stating plainly in the RFC: **there is no "just thread the field and my page
fills" argument available, including from the lane that filed 468.**

**TWO FRAMINGS FOR THE RFC, both from the feed lane and both adopted here:**

1. **`tools`/`guides`/`blog` are not a vocabulary anyone chose — they are FALLBACKS that became a
   convention** because nothing else was available. That is why the head is so concentrated and
   the tail is free text. The 58 directories are downstream of there being no vocabulary *and* of
   three hardcoded per-role defaults doing the work a declaration should do.
2. **Do NOT read the single-page tail as uniformly junk.** `/the-design-feed/` is single-page
   *because it has no children* — which is the very bug `444` and `463` are about. **A directory
   empty because its producer is missing is indistinguishable at the data layer from one that is
   a typo**, and this census cannot separate them. The tail is a mixed population.

**Ownership unchanged:** `468` stays filed and unowned, `460` stays unowned, the RFC is this
lane's. The feed lane explicitly declined to take any of it.
