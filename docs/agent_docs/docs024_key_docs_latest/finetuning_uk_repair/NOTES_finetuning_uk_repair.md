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
