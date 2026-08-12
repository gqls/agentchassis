# NOTES — finetuning.uk repair

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-03 — session start, orientation

Owner: "the finetuning site is looking terrible … fix it using the framework
(not locally) … run the audit checks … make sure the handlers are all
automatically picking up the items properly … check the framework catches
everything."

`finetuning.uk` = `1368e337-dd1d-4799-bbb3-8221a1b79bcc`, status `deployed`.
Live host returns HTTP 200, 49,707 bytes. So not an outage — a quality problem.

Open work items on the site: **61**, across 22 (item_type, status) groups. Two
generations, visible in the timestamps: a block all stamped
`2026-07-26 21:06:59.606963+00` at status `detected`, and older rows at
`unresolved` from April/May. Nothing has moved since 2026-07-27. Today is 08-03.

---

## 2026-08-03 — the visual defect, found by reading the served HTML

Stripped tags from the homepage and read the prose: it is fine. Well-written,
specific, no lorem, no obvious truncation. So "looking terrible" is not the copy.

Then listed the `<img>` tags, and there it was:

```
<img src="cpu"      alt="Automation & Workflow Department" class="member-photo">
<img src="network"  alt="Intelligent Agent Systems Department" class="member-photo">
<img src="database" alt="Data & Research Department" class="member-photo">
… eight in total on the homepage, eleven more on /about.html
```

Confirmed against HTTP rather than inferred — `/cpu`, `/network`, `/database`
all **404**. Each renders as a broken-image icon at `width:120px;height:120px`
in a circle, down the middle of the page. That is the "looking terrible".

Also confirmed 404: five `/assets/images/case-study-*.jpg`. Those five **were**
already filed as `image_url_404` work items. The nineteen broken icons were not
filed as anything.

---

## 2026-08-03 — MISSTEP: my first fleet measurement said "zero", and it was wrong

To size the problem I wrote a fleet-wide query for `<img src>` values with no
slash and no dot, using `'<img\b[^>]*\bsrc\s*=\s*"([^"]*)"'`.

**It returned 0 rows.** I had the live HTML in front of me showing the exact
shape, on a site inside the population, so I knew the answer could not be zero —
otherwise I would have written "the fleet is clean" and been confidently wrong.

**Cause: Postgres POSIX regex does not have `\b` as a word boundary.** `\b` is a
BACKSPACE character; the word-boundary escape is `\y`. The pattern was asking for
a literal backspace and matching nothing — **with no error and exit 0.** Re-run
with `[[:space:]]`:

```
ai-agent-orchestration.com   16 occurrences
finetuning.uk                15 occurrences
                             31 total, 2 sites, 1 component
```

The cheap check that would have caught it: **run the query against a row you
already know matches, first.** A census whose known-positive is absent is a
broken census, not a clean estate. Logged to `WRONG_CALLS.md`.

---

## 2026-08-03 — root cause: a photo component wearing a department's clothes

All 31 attribute to one component, `departments-grid`. Its template:

```
{{if .icon}}<img src="{{.icon}}" alt="{{.name}} Department" class="member-photo">{{end}}
```

and its `input_schema`:

```
"departments": { "items": { "icon": "string", "name": "string", … } }
```

The schema is RIGHT — `icon` is a name. The template is wrong. The surrounding
markup gives the history away: `team-section`, `team-member`, `member-photo`,
`member-title`, `member-bio`. It was forked from a staff-photo component and
repurposed, and the `<img src>` came with it.

The correct target is not a guess. The `features` component, **on the same
page**, already does:

```
{{if .icon}}<div class="feature-icon"><i data-lucide="{{.icon}}"></i></div>{{end}}
```

and it works. Checked all four affected pages load `lucide.min.js` and call
`lucide.createIcons()` — they do. So the replacement renders.

---

## 2026-08-03 — the framework gap, and a claim I nearly got wrong

`check_image_url_404` owns broken images. Two predicates:

- `imagePathRefPattern` — requires `/assets/images/<name>.<ext>`
- `emptyImgSrcPattern` — requires `""`, `'  '`, `"#"` or `'#'`

`src="cpu"` matches neither. **Structurally invisible.** Same run, same two
pages: five findings raised, nineteen broken images ignored.

**MISSTEP AVOIDED, worth recording because I was one step from asserting it.**
`/blog.html` carries `<img src="">`, which the empty-src predicate SHOULD catch,
and there was no `empty_src` work item. I was about to write "the empty-src
branch is broken too". Then I noticed the stored summaries say *"Pages reference
unknown image X"* while current source says *"…but no active asset deploys to
that path"* — different wording, so **the running check is not the checked-out
check**. `git log` on the file: the empty-src branch arrived in `beff42809` on
**2026-07-31**; the last discovery run was **2026-07-26**. The branch had not
shipped yet. Nothing is broken there.

The lesson: the artefact in the database was written by a binary, and the binary
has a version. Comparing DB output against working-tree source is comparing two
different programs.

---

## 2026-08-03 — the dispatch finding, which is bigger than this site

Every open item has a `handler_agent` (`rerender-pages`, `page-build-handler`,
`image-build-handler`, `webdesign-agent`, `asset-deployer`, `tool-auditor`). So
handlers exist and are named. But **`attempt_count = 0`** on all of them,
including items detected eight days ago. Nothing has ever tried.

The chain, read in source rather than assumed:

- `LoadWorkItemsAction` claims `WHERE wi.status IN ('triaged','approved')`.
- Every open item on this site is `detected` or `unresolved`.
- The only promoter of `detected → triaged` is `triage_findings`, a step INSIDE
  the improvement-loop.
- The improvement-loop's only schedule, `improvement-sweep`, is `enabled=f`,
  last triggered **2026-05-02**.

Fleet-wide: **detected 204 / 10 sites · triaged 2 / 1 site · unresolved 235 / 8
sites.** Detection works; dispatch does not. This is not new — `WRONG_CALLS.md`
already carries two entries where a session tripped over it — but the scale is
worth stating plainly.

Note the `[unresolved after 2 attempts]` prefixes on old summaries are
HISTORICAL. The `attempt_count` column reads 0 for those same rows. The string
and the counter disagree; the counter is the one that reflects the current
dispatcher.

---

## 2026-08-03 — MISSTEP: I wrote a trailer I could not honour

Committed the checker fix with `Council-Submitted: pending` **before** running
the submission, so "pending" is not a resolvable correlation. Forward-only
forbids an amend, so it stays wrong in that commit.

The real correlation is **`cfc94d91-3d17-4f29-a370-2b91d1a59a6f`**, submitted
minutes later and recorded here and in `README_where_we_are.md`. `098` will not
auto-credit that commit, because the string it resolves is a literal `pending`.

The cheap check: **get the correlation first, then commit.** The trailer exists
precisely so you can commit before the VERDICT — it does not let you commit
before the SUBMISSION.

---

## 2026-08-03 — the existing tests caught a false positive in my own fix

First version of `bareTokenImgSrcPattern` excluded `/`, `.` and `:`. It did not
exclude `#`, so `<img src="#">` matched BOTH the empty-src predicate and mine,
and `TestImageURL404_EmptyImgSrcIsCountedAndReportedOnce` failed with 2 work
items where it wanted 1.

Correct outcome, and it is the right test that caught it: a `#` src is a
placeholder, which the empty-src shape already owns. Added `#` to the excluded
set. Then wrote the negative controls I should have written first — six
legitimate src shapes (rooted path, relative filename, relative path, absolute
URL, protocol-relative, data URI) that must all stay silent, because the cheapest
way to kill a false positive is to widen a pattern until it reports nothing, and
a check that reports nothing looks exactly like a check that found nothing.

---

## 2026-08-03 — repairs applied

1. **`departments-grid` template** — `sql_for_agents/293_…sql`, applied live,
   verified by a `DO/RAISE` block (a verify block of `SELECT`s cannot stop a
   `COMMIT`; `ON_ERROR_STOP` ignores a non-empty result). Now emits
   `<div class="member-icon"><i data-lucide="{{.icon}}"></i></div>`, with
   `.member-photo` replaced by a `.member-icon` badge of the same 120px geometry
   so the grid does not reflow. Fixes all 31, both sites, at source.

2. **`check_image_url_404`** — third predicate, own kind `bare_token_src`, own
   item key, severity high, 3 new tests. Committed `1985c0433`.
   **Pod-grepped both replicas: NEW 0 / CTRL 1 — committed and NOT live.**
   Inert until a chassis build. Stated, not hidden.

3. **`294_TRIGGER_improvement_loop_v1.sh`** — written because the manual trigger
   documented at `004_improvement_loop.md:265` (`./trigger-audit.sh`) **is not in
   the tree**; `find` returns nothing. Registered as IMP-050.

---

## 2026-08-03 10:14 — the framework running

Fired the improvement-loop at the site. Orchestration
`9ae95283-cef1-483c-83ec-c0f5709d91e6`. Progression observed:
`spawn_design_discovery` → `call_design_audit`.

Discovery is producing: site items went **detected 26 → 87**,
**needs_human_review 14 → 59** within four minutes. So the discovery half of the
loop is healthy; what had been missing was only ever the promotion and dispatch
half. Continues below.

---

## 2026-08-03 — what this session did NOT do, and why

- **Did not re-enable `improvement-sweep`.** Off since 2026-05-02, and IMP-016
  records the pause as deliberate. Re-enabling it would start promoting and
  dispatching across ten sites — a fleet decision, taken as a side effect of a
  one-site task. Fired the loop at one site instead. The fleet question is put to
  the owner in `PLAN`/`README` rather than answered here.
- **Did not hand-edit `rendered_html`.** The owner asked for the framework
  explicitly, and a hand-edit is reverted by the next rerender anyway.
- **Did not build/roll the chassis** to make the checker fix live — see
  README for the recommendation and why it is the owner's call while another
  session is rolling images.

---

## 2026-08-03 10:29 — the repair LANDED on /about.html, proven at the artefact

The improvement-loop completed at 10:24. It promoted **128 items detected →
triaged** — the first time anything on this site has been dispatchable in months —
and `build-dispatch-loop` began claiming immediately.

**But the queued rerenders would not have applied the template fix**, and that is
the most useful thing in this file. Checked the pending items' reasons, as the
`page_rerender`/`spec.reason` landmine prescribes:

```
(NULL — re-staples STORED html)   42
cta_links_stale                   26
```

and per page carrying `departments-grid`:

```
/index.html   cta_links_stale   <- regenerates: the fix WILL land
/index.html   (NULL)
/about.html   (NULL)            <- re-staples: the fix would NEVER land
```

`rerender_single_page_action.go` says it in its own header, line 4: *"Simple
concatenation - no template re-rendering"*. So `/about.html` — eleven of the
nineteen broken images — was queued for a rerender that would have completed,
reported success, and preserved exactly the defect it was supposed to remove.
`bugs_open/140` is the same story on 08-02: backlog drained 294 → 0, six pages
COMPLETED, not one section updated.

Queued `294_…sql`: one `page_rerender` for about with
`spec.reason='section_data_resolved'`, priority 20, `triaged`. Result:

```
claimed   10:29:10  by build-dispatch-loop
complete  10:29:56           (46 seconds)
/about.html page_component updated_at 10:29:40 · lucide=t · bare_img=f
```

Verified at the live artefact, not at the status:

```
curl -sS -L https://finetuning.uk/about.html | grep -oE '<i data-lucide="[a-z-]+"|<img src="[a-z-]+"'
  → 10 × <i data-lucide="…">   (cpu, database, download-cloud, globe, layers,
                                lock, map, network, settings, workflow)
  → 0  × <img src="…">
```

**Eleven broken images gone from the served page.** `/index.html` still pending —
it carries a `cta_links_stale` item at priority 35, which DOES take the
regenerating branch, so it needs no intervention, only its turn in the queue.

---

## 2026-08-03 — council round 1: REVISE, and the objection was right

Six of thirteen seats raised the same thing, gated high by `tooling_provenance`:
a standing landmine keys `check_image_url_404.go` together with
`check_placeholder_image_in_use.go` — *"Two discovery checks already own 'a page
renders the fallback image path' — extending one silently competes with the
other"* — and **my submission showed no sign of having read it. I had not.**

I had run `grep LANDMINES` for the paths I was editing at the start of the
session and this entry did not surface, because I grepped for the SYMBOL and the
FILE I was changing, and the entry's discriminating text is about the *other*
file. Worth noting as its own small lesson: an overlap landmine is keyed to a
PAIR, and you will only find it by the half you are not editing.

Ran the check the landmine itself prescribes —
`grep -l "assets/images" platform/orchestration/actions/discovery_checks/*.go` —
four files, all headers read. The answer is structural rather than a judgement:
`placeholder_image_in_use` matches two **literal** paths,
`/assets/images/hero.jpg` and `/assets/images/logo.png`. Both contain `/` and
`.`; the new character class is `[^"/.:#\s]+`, which excludes both. **No input
exists on which both fire.** The other two checks in that space do not read
rendered HTML at all.

Two further objections, both fair, both answered rather than argued:

- `prior_art_librarian`: my false-positive case leaned on `storage.DeployedWebPath`,
  which other landmines flag as silently wrong. **Misdirected — but round 1's
  wording caused it, so the wording is fixed.** The bare-token branch never calls
  that helper; `loadDeployedAssetPaths` runs only in the `len(references) > 0`
  branch. A bare word has no extension to get wrong. (That landmine also carries a
  status banner: FIXED AND LIVE at HEAD, `bugs_closed/168`.)
- `prior_art_librarian`: "detected items are not promoted" was asserted without a
  lookup. It **was** measured — 204/10 sites vs 2/1 — I simply did not put the
  query in the submission. Now cited.

Resubmitted round 2 under `RESUBMIT_CORR=cfc94d91-…` so the trail accumulates.
The overlap analysis went **into the file**, not just the submission: the
landmine's failure mode is "you reach for whichever of the two files you found
first", so the answer belongs where the next author lands.

`bug_historian` also made a broader point I have NOT acted on: this closes one
shape of a generic problem — a schema-declared `string` rendered into a
structural HTML sink with no bind-time validation that it is a resolvable
reference. Correct, and the architecture seat approved this as a point fix on the
same reasoning. Widening a detector bug fix into a generic sink-type check is
exactly the scope the guardian seat vetoes. Recorded as a follow-up.

---

## 2026-08-03 — my LANDMINES edit was swept into another session's commit

Appended the Postgres-`\b` landmine, ran `landmines-sync.py --apply`, then
committed by pathspec — and got "1 file changed", the SQL only. `git status` on
LANDMINES.md: clean. `git show HEAD:…LANDMINES.md | grep -c BACKSPACE`: **1**.

Another session had committed the file, with my append inside it, between my
write and my commit — `478187ba0`, a webdesign.uk shopfront commit. This is the
**same-file passenger** case CLAUDE.md names: a pathspec commit protects you from
carrying other sessions' files, and cannot stop another session carrying yours.
Nothing lost, forward-only holds, and the entry is at HEAD. Recording it because
the tell was confusing for a moment — "my commit dropped a file" and "someone
else already committed my change" look identical at `git commit` and are
distinguished only by reading HEAD.

---

## 2026-08-03 — REFINEMENT to my own claim: the LLM layer DID see it; the structural layer did not

I have been writing "a live broken image with no finding anywhere", including in
the council submission. **Checked rather than left standing**, because it is the
load-bearing claim of the whole checker fix:

```sql
SELECT s.domain, wi.item_type, wi.created_at, left(wi.summary,100)
FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
WHERE wi.created_at < '2026-08-03 10:14'
  AND (summary ILIKE '%src="cpu"%' OR summary ILIKE '%bare word%'
       OR summary ILIKE '%broken img%' OR summary ILIKE '%member-photo%');
-- 0 rows, fleet-wide, all history
```

So the claim was TRUE as stated and as submitted. **But today's run changes the
picture going forward, and the refinement is worth more than the claim was.** At
10:16 the `design-audit` agent filed, unprompted:

> *"The team-section uses a broken img src value of `src="cpu"` — a bare word
> rather than a valid image path — which will render as a broken image icon,
> creating a significant visual layout defect with unpredictable sizing in the
> team grid that breaks alignment across cards."*

That is the same diagnosis I reached, reached independently, by an LLM auditor
reading the page. Two things follow.

**First, it corroborates the diagnosis from a source that shares none of my
reasoning** — I read the template and the schema; it looked at the rendering. Not
proof, but the strongest independent agreement available here.

**Second, it locates the gap more precisely than I did.** The framework was never
totally blind to this: its **LLM audit layer** could see it. What could not see it
is the **structural layer** — the deterministic `discovery_checks` that run cheaply
every cycle, need no tokens, and are what "does the framework catch this?" should
mean for a defect class that recurs. An LLM noticing a thing once, on the run
where someone happened to fire the loop, is not detection you can rely on; a
predicate is. So the fix is still the right one, and I should have said "no
STRUCTURAL finding" rather than "no finding anywhere".

Recording it as a refinement rather than a retraction because the submitted claim
was accurate for the period it described — but the more careful phrasing is the
one that would have survived this check, and it is the one to use from here.

**Also confirmed by the same run: the empty-src branch works.** It filed
*"16 `<img>` tags render with no image source (empty or '#' src)"*. That is the
branch I nearly declared broken on 07-26 evidence before noticing the running
binary predated it. It shipped 07-31, this is its first run on this site, and it
fired correctly. The earlier note stands: nothing was wrong there.

## What else the fresh audit found (answering "see what you can see wrong")

The run filed 157 new items. By weight:

| type | n | high | what it is |
|---|---|---|---|
| `page_rerender` | 70 | 28 | mostly misdirected CTAs — copy names a different page than the link goes to |
| `required_fields_missing` | 25 | 0 | components missing schema-required fields |
| `cta_names_unknown_destination` | 18 | 0 | a CTA whose label names a destination that does not exist |
| `phantom_internal_link` | 17 | **17** | links to pages that are not there |
| `empty_section` | 7 | 0 | the long-standing ones, re-detected |
| `image_url_404` | 6 | 0 | incl. the 16 empty-src tags |
| `claims_unverified` | 4 | 0 | stats on /about not in the register |
| `needs_design_review` | 2 | **2** | body font is Merriweather (serif) where the style collection specifies otherwise |
| `cta_improvement` | 1 | **1** | the only visible conversion path is one high-commitment CTA |

The phantom links and misdirected CTAs are the substantial content problem behind
the visual one, and they now have handlers and are queued — which is the first
time that has been true on this site since April.

---

## 2026-08-03 — CLASS FIX COMPLETE, and the closing zero was checked against a control

All four pages that mount `departments-grid`, verified at the served HTML:

```
https://finetuning.uk/                        broken_img=0   icons=18
https://finetuning.uk/about.html              broken_img=0   icons=11
https://ai-agent-orchestration.com/           broken_img=0   icons=17
https://ai-agent-orchestration.com/about.html broken_img=0   icons=8
```

Fleet census: **31 → 0**, both sites, page and chrome surfaces.

**The zero was NOT taken at face value**, because a zero-row census is precisely
what produced this lane's first wrong call this morning — a Go regex carried into
Postgres, where `\b` is a backspace, matching nothing at exit 0 and reading as
"the fleet is clean". Having written a landmine saying *"a census whose
known-positive is absent is a broken census, not a clean estate"*, accepting an
unchecked zero would have been the same error with the sign flipped. So:

```sql
-- same extraction, filter relaxed
total img src extracted              156
of those, bare tokens (the finding)    0
sanity: srcs containing a dot        133
```

The extraction is demonstrably working on live data. The zero is the estate, not
the query.

**Note the shape of that check, because it generalises**: when the finding
population drops to zero there is no known-positive left to test against, so the
control has to come from *relaxing the predicate* rather than *narrowing the
population*. "Does this query find anything at all?" is answerable even when the
thing it looks for is gone.

**Timing, for the record.** The last page (`ai-agent-orchestration.com/about.html`)
landed roughly 15 minutes after `296` queued it — it waited behind finetuning.uk's
115-item queue on a shared dispatcher, which is exactly the behaviour `295`'s
header predicted and priced. `SUMMARY_2026-08-03b` was written while it was still
in flight and now carries a visible dated update rather than an edit: the series
is the record, and what a summary believed at the time is part of it.

---

## 2026-08-08 — Thunder keys VERIFIED working; the roll shipped the checker; queue fully drained

**Thunder API, both keys, functionally proven** (service lane ask, recorded here
because the fresh-thread handoff cites it):

- Local `~/.config/thundercompute/token` → `GET /v1/instances/list` **HTTP 200**
  (`{}`, zero instances — matches all-23-decommissioned). **Negative controls:**
  invalid token → 401, no token → 401. The 200 discriminates.
- Cluster `THUNDER_COMPUTE_API_KEY`, exercised from **inside the adapter pod**
  via its own env (never printed): 200-class. Same-string question remains
  unanswerable (secret read is classifier-blocked) and now immaterial — both
  authenticate.

**The chassis roll happened** (owner said "later when quiet"; between 08-05 and
08-08 it did). Pod-grep, both replicas, with control:
`image_url_404:bare-token-src` = **1** · `empty-src` control = **1**. The
bare-token checker is LIVE fleet-wide. New pod hash 67ddcc695f (was 7fcc74d89d).

**Queue state**: 259 complete · 0 triaged/claimed — the 08-03 promotion has fully
drained. Remaining: 85 needs_human_review, 25 unresolved, 13 failed, 11 blocked
(the case-study `image_url_404` flags), 4 wont_fix.

**Still open, unchanged**: the five case-study images 404 (re-checked today).
Next actions and their traps are in
`../finetuning/HANDOFF_2026-08-08_continue_here.md` §3 — that file is now the
cold-start for both lanes.

---

## 2026-08-09 — Phase 1 fired: the five case-study images queued and dispatched

Picked up from `../finetuning/HANDOFF_2026-08-08_continue_here.md` §3 task A,
after verifying the unrelated B2 credential fix had shipped on v1.0.1274
(`bugs_open/233`).

**State on arrival, re-measured rather than carried forward:** all five
`/assets/images/case-study-*.jpg` still **404**, positive control
`/index.html` = 200, so the check discriminates.

**The extension trap is RESOLVED, and the answer is that the framework already
handles it.** The plan flagged `[UNVERIFIED] that the generated extension will
be .jpg` as the live risk that "fails silently in the direction of looking
successful". Settled three ways rather than by reading one of them:

1. `platform/storage/url_helpers.go` — `ImagePurposes["content_hero"]` is
   `{1600, 900, 85, "jpg"}`; `DeployedAssetPath` takes the extension from
   `GetImageConfig(purpose)` and the filename from
   `AssetKeyFilename(assetKey, dotExt)`, which is
   `strings.ReplaceAll(assetKey, "_", "-") + ext`.
2. **Live, on this site, with negative controls** — three assets that already
   went through this exact chain:
   `content-hero-ai-agent-roi-estimator.jpg` 200,
   `content-hero-llm-cost-calculator.jpg` 200, `hero-case-studies.jpg` 200;
   the same asset as `.png` → **404**, and with underscores kept → **404**.
   So both halves of the derivation are confirmed, not just the happy path.
3. **The decisive one — a completed run's own record.** The 08-03
   `llm-cost-calculator` item's `result` shows the generated original in S3 is
   a **`.png`** (`…/796d3589-….png`) while `deploy_result.data.file_path` is
   **`/assets/images/content-hero-llm-cost-calculator.jpg`**. The generator
   emitting PNG was the exact failure the plan feared; the deploy leg converts.
   Reading only the generator would have produced a confident wrong answer.

**What was queued.** Five `needs_imagery` rows, `handler_agent =
image-build-handler`, `purpose = content_hero` (the jpg config),
`asset_key = case-study-<slug>` — so `DeployedWebPath` yields exactly the five
paths **both** surfaces already reference, per the plan's resolved item 3.
Prompts derive from the site's OWN copy: each card's title, its excerpt, and
its existing `card*_image_alt`, which already carries the art direction
("abstract geometric network diagram…", "calm atmospheric geometry…"). Nothing
invented — the framework wrote that copy and the images should match it.

**Why hand-queued at all** (it looks like bypassing the framework and is not):
`check_image_url_404` is flag-only by design, so its eleven items sit `blocked`
with "No handler_agent set" and **no discovery pass will ever raise these**.
Same item type, same handler, same dispatcher as the ten that succeeded on
08-03 — only the raising step is manual, which is precisely what the plan's
Phase 1 specifies.

**Two mechanism facts worth keeping** (both read from the live rows, not assumed):
- `status='detected'` is NOT claimable. `load_work_items` claims
  `status IN ('triaged','approved') AND attempt_count < max_attempts AND
  approval_mode='auto'`, so the rows were written `triaged` directly.
- `build-dispatch-loop`'s `load_items` step sets **`max_items: 5`** and no
  pipeline/handler filter, ordered `priority ASC, created_at ASC`. Five items
  and no other `triaged` row on the site = one clean batch. Had there been six,
  one would have been left behind silently.

**Dispatched** via the standing per-site trigger; both pre-flights passed
(youngest chassis pod 8822s old; 0 claimed items).
`ORCH_ID=18b299ff-59d0-4676-b969-650f38ded505`,
correlation `80076b13-b8fe-4d15-9d2e-03f9b87e4b44`.

**Verification is at the served URL, not the item status** — a `complete`
`needs_imagery` row is not a serving image, and this lane has the 08-03
precedent for exactly that gap.

**Note for whoever runs Phase 2/3:** another thread fired discovery and a design
review at this site earlier today (runs at 13:51–14:44 UTC, all COMPLETED, none
in flight when I dispatched). Two new `needs_imagery` items appeared at 13:51
for `model-approach-selector` and `tool-automation-savings-estimator` — those
are content heroes for other pages, not case studies, and are unrelated to
these five.

### 2026-08-09, later — the run COMPLETED and fixed nothing, and one claim above is now CORRECTED

> **CORRECTED 2026-08-09.** Earlier in this entry I wrote, as a checked mechanism fact:
> *"Five items and no other `triaged` row on the site = one clean batch."*
> **That was true when measured and false by the time it mattered — my own dispatch
> destroyed it.** Full account in `WRONG_CALLS.md` (2026-08-09).

`ORCH 18b299ff` ran 48 minutes, reported `complete / COMPLETED`, and left all five
items `triaged` with `attempt_count = 0`. Nothing failed. It dispatched a batch of
five — a priority-5 `deactivated_component`, a `needs_design_review` and three
others — **none of them mine.**

**The mechanism I missed sits inside the loop I fired.** `improvement-loop` runs
`triage_findings` *before* `spawn_dispatch`/`call_dispatch`, and that step promotes
`detected → triaged` in bulk across the site. It moved ~95 findings in this run:

```
priority 35 → 20 items (19 page_rerender + 1 content_rewrite)
priority 60 →  8 · priority 65 → 5 · priority 80 → 46 page_rerender
priority 90 →  my 5 (+2 other needs_imagery, +3 needs_page)
```

`load_work_items` is `ORDER BY priority ASC, created_at ASC LIMIT max_items(5)`, so
the five case-study items went from *the only claimable rows on the site* to
*positions ~80–84 of a 95-item queue that drains five per run* — during the very run
meant to serve them. At five per firing that is ~16 further loop runs.

**What caught it: the watcher's authority was the served URL.** It held at
`SERVED: 0/5` across 30 polls while the orchestration said COMPLETED. Both the
orchestration status and the item status would have read as success — the run *did*
succeed, at other work. This is the lane's own "trust the artefact, not the status"
rule paying for itself a second time.

**Fix applied:** the five re-prioritised to `priority = 1` — exactly one `max_items`
batch — with an in-transaction assertion that **zero** claimable items sit ahead of
them. `1` rather than something softer because triage mints low numbers itself (the
batch that displaced them carried a priority-5 item), so any middling value can be
outranked again by the next run's own promotions. SQL:
`scratchpad/reprioritise_case_study_imagery.sql`.

**Not queue-jumping, correct ordering:** 65 of the items ahead were `page_rerender`,
and a rerender that runs *before* the images exist re-staples pages that still lack
them and has to run again.

**Also confirmed while diagnosing this** (`LANDMINES.md`, 2026-08-08,
webdesign_uk_build_service): hand-dispatching `build-dispatch-loop` with a bare
`action=orchestrate` to skip the expensive loop **reports COMPLETED and processes
nothing**. So re-firing the full 294 trigger is the supported route, not a
convenience — the cheap-looking shortcut is a documented no-op.

---

## 2026-08-12 — OWNER RULING on the lane, the design audit RUN, and the first finding verified down

### The lane, as the owner set it (2026-08-12)

**This lane owns the DESIGN of finetuning.uk** — the site that was broken when the
work started. **Everything else goes to `finetuning_uk_service/`**, which owns the
paid fine-tuning service backend (Thunder, the run, the bundle) and has its own
standing five. A third lane, `finetuning/`, holds the older service-thinking plus
the site history; its `HANDOFF_2026-08-08` now opens with a banner naming all
three. I planned a day of the service lane's Phase 0 before discovering it existed
— `WRONG_CALLS.md` 2026-08-12; the check that would have caught it in a minute is
`grep finetuning MEMORY_workstreams.md`.

### Task B (the visual-designer pass) was genuinely DUE, and here is why

The handoff gated it on "once the images are real". Two facts made it due rather
than merely unblocked, both measured today:

- `/index.html` was **deployed 2026-08-12 03:34:52Z** — the `bugs_open/238` repair,
  by that lane. The homepage now carries all five `case-study-*.jpg` and **zero**
  empty `src` (served-page check, with a fabricated sixth filename 404ing as the
  control).
- The **last design audit ran 2026-08-11 13:21** — *fourteen hours before* that
  deploy — so every finding it holds was formed against the broken homepage. 98
  `page_rerender` items also completed overnight. The audit was stale against the
  artefact.

### What I fired, and why not 294

`295_TRIGGER_design_audit_detect_only_v1.sh` (new; register **IMP-054**), which
dispatches `design-audit-agent` directly instead of the full improvement-loop.
**294 would have run the audit AND `call_dispatch`**, and `call_dispatch` runs the
content-REGENERATING handlers — the exact sequence that on 08-09 dropped the
homepage's image URLs (`bugs_open/238`). We wanted the report, not a rewrite.

Detect-only is a verified property, not an assumption: the live `design-audit-agent`
workflow is six steps with **no triage**, neither child auditor contains one, and
`write_audit_findings_action.go:677` hardcodes status `'detected'` while the
dispatcher claims only `('triaged','approved')`. ⚠ The near-identical
`081b_design_audit_agent_robot_hands.sh` claims in its comments that the agent
triages and dispatch picks items up on the next 30s tick — **stale, from an older
definition.** Copy its envelope, not its expectations.

**Run:** ORCH `62889bbf-c9ac-4077-a74e-3f446b285a8a`, COMPLETED in ~60s from a
baseline of **0 detected** items. Both auditors ran (`call_visual_auditor` and
`call_content_auditor` both present in `collected_data`, each with a result).
Visual: 5 findings → 4 items. Content: 5 findings → 1 item (4 deduped). **5 new
`detected` items**, nothing claimable.

| item_type | sev | the claim |
|---|---|---|
| `cta_improvement` | high | only conversion path is "Book a Discovery Call"; no intermediate CTA |
| `dark_section_audit` | high | `case-studies-grid-section` defines its own inline `--section-*` vars in a `<style>` instead of inheriting shared dark-section tokens |
| `needs_design_review` | high | hero uses hardcoded `rgba` overlay + hardcoded `--hero-btn-ink` hex instead of variables |
| `needs_design_review` | high | body font is Merriweather serif in the theme, but the style collection specifies a system sans stack |
| `spacing_fix` | med | CTA section padding uses `var(--spacing-section, 5rem 2rem)`, a two-value fallback |

### ⚠ The font finding is REAL but its stated EFFECT is wrong — verify before promoting

Worth writing down because it is the whole argument for detect-only. The claim is
"body font **is** Merriweather serif". Checked at the served artefact:

- the style collection (`professional-dark`) does specify
  `-apple-system, BlinkMacSystemFont, Segoe UI, Roboto, sans-serif`;
- the served HTML **does** contain `body { font-family: "Merriweather","Georgia",serif; }`
  — so the conflict is real, and my first pass (grepping only the external
  stylesheet, which has **0** Merriweather) wrongly read it as a hallucination;
- **but it does not apply.** The served page has exactly **one** `body {`
  declaration, in an inline `<style>` at lines 55–137. The external
  `/assets/css/styles.css` loads at line **142 — after it** — and declares
  `body { font-family: var(--font-body) }` with
  `--font-body: 'Inter','DM Sans',system-ui,-apple-system,sans-serif`. Equal
  specificity, later wins, and **no `body` font rule appears after line 142**.

So the page renders **sans**, roughly as the collection intends. The defect is a
**dead conflicting declaration** that will bite whoever reorders the stylesheets —
not a visible font error. **A naive repair is actively dangerous here:** an agent
told "the body font is wrong" could edit the *winning* external rule and change a
page that currently looks right.

> **[REASONED, not browser-measured]** — this is a complete cascade analysis
> (one inline rule, one external rule, known order, no later override), not
> `getComputedStyle`. `LANDMINES.md` warns that a CSS literal present in source
> may never be applied; that warning is what prompted this check, and the same
> caveat applies to my own conclusion. If a repair is ever dispatched here,
> measure computed style first.

### Where this leaves it

Five findings sit `detected`. Nothing will act on them until someone promotes one
with an explicit `UPDATE … status='triaged'`. **Recommended order once the owner
chooses:** the two token/hardcoding findings (`dark_section_audit`,
`needs_design_review` hero) are genuine structural drift and the safest to repair;
`spacing_fix` is small; the font one should be repaired as *delete the dead
declaration*, phrased so no agent touches the external rule; `cta_improvement` is
a business decision about the funnel, not a design defect.
