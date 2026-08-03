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
