# NOTES — `bugs_open/136` (section-editor slug), sibling link repair

Append-only, newest at the bottom. Technical log: what was tried, what the system actually
said, and every misstep.

---

## 2026-08-02 ~10:20 BST — picking the bug, and the two I put back

Swept `bugs_open/` (60 files) against every `.jsonl` transcript touched in the last 5–10
hours, counting `bugs_open/NNN` mentions per session. That is the check CLAUDE.md asks for
and it earned its keep twice:

- **`bugs_open/155`** (deploy_image_asset resolves by purpose) read like the ideal pick —
  high severity, root cause CONFIRMED by the diagnosis loop, fix candidates written. Session
  `693556a1` has **119** hits on `bugs_open/155` and 113 on `resolveStorageURIFromAsset`. It
  is being fixed right now. `scripts/who-owns.py 155` said "OWNED or recently active" but
  named the *doc-review* and *143* workstreams — i.e. **the lagging-indicator failure mode
  the memory entry describes**: who-owns reads commits, and a session mid-fix has none yet.
  The transcript grep is what actually found the owner.
- **`bugs_open/093`** (stat audit, one guarded call site) is the same *family* as this bug
  and reads as unowned. Its own last update (2026-07-27 triage) says it is **not a code task
  any more**: candidate 1 was built and shipped in v1.0.1172, and the check it was built
  into is fired only by `improvement-sweep`, which has been disabled since 2026-05. It is
  blocked on `bugs_open/083`, which is hot in two live sessions. Left alone.

Landed on `bugs_open/136` **(section-editor slug)**. `136` is one of the ambiguous numbers;
`who-owns.py` matched on the number and surfaced the *other* 136's commits, which is exactly
why CLAUDE.md says resolve by slug and `git log` the file path.

## 2026-08-02 ~10:40 BST — is it still valid, and is the file right about its own targets

```
grep -rn "repairSectionLinks\|RepairPageLinks" --include=*.go platform/ internal/
```
→ three call sites, all pre-existing: the build gate, the outbound rerender seam, the
section save. **Still valid**; nothing has closed it since 2026-07-28.

Two corrections to the bug file's own account, both from reading the code it names:

1. **The report path's prose really is raw.** `renderReportSection` escapes every
   deterministic field with `html.EscapeString` (`esc(...)` on ~25 fields), which had me
   about to write "no anchor can reach this HTML". Then line 413:
   `fmt.Fprintf(&b, `<div class="report-prose"><h2>%s</h2>%s</div>`, esc(sec.heading), body)`
   — `body` is `summary_html` / `candidates_html` / `integration_html` /
   `vendor_questions_html`, LLM-authored, **unescaped**. The bug file is right and my
   first read was wrong; the escaping is dense enough to look total.
2. **The blog listing is not a member of the class.** The bug file lists it as a writer with
   no repair (true) and rates it "lower risk but not zero". It is lower than that: its only
   anchor comes from the template, and the live template has exactly one, `href="{{.url}}"`,
   filled from `pages.url`:

   ```sql
   SELECT function, (SELECT count(*) FROM regexp_matches(html_template,'<a[^>]+href','gi')) AS anchors,
          (SELECT string_agg(m[1],' | ') FROM regexp_matches(html_template,'href="([^"]*)"','g') m) AS hrefs
   FROM content_components WHERE function='content-listing' AND is_active;
   -- content-listing | content-listing | 1 | {{.url}}
   ```

   Since the repair index is built from `pages.url` under a *looser* predicate
   (`status NOT IN ('deleted','archived')`) than the listing's own eligibility floor, every
   href it emits is in the index by construction. Repair there can only no-op. Excluded.

   **Misstep worth recording:** I nearly added the call anyway "for symmetry". A no-op call
   on a live path is not free — it is a `pages` query per rebuild and a new failure surface —
   and "the bug file lists it" is not evidence. The query above took 20 seconds.

## 2026-08-02 ~10:55 BST — the census the bug file marks `[UNMEASURED]`

The file says: *"No count of phantom links actually shipped through these paths … Do not
state a volume until that is measured."* Measured now, fleet-wide, against stored
`page_components.rendered_html` (query in `RUNBOOK_sibling_link_repair.md`). It mirrors
`NormalizePagePath` in SQL — lowercase, strip `#`/`?` tail, strip `index.html`, trim
trailing `/` — and applies `ClassifyLinkScope`'s exclusions (external, `mailto:`/`tel:`/
`javascript:`, `#`-fragment, asset extensions).

**30 distinct (page, href) rows on 7 sites**: 14 `rewrite` (a real `.html` target exists,
the writer just omitted the extension) and 16 `UNLINK` (no such page — a live 404). None of
the carrying components has a `data-runtime-fill` marker, so the whole-input exemption
would not spare any of them. Buckets: `page_type='content'` 14 hrefs / 6 components,
`landing` 2 / 1, `tool` 1 / 1 (by page type of the *carrying* page).

**What this number is and is not.** It is the standing stock of unrepaired internal links in
storage. It is **[NOT ATTRIBUTABLE]** to a writer — a stored href cannot say who wrote it,
and 079's fix is live, so some of this predates it. Anyone quoting "30" as "the sibling
writers produced 30 phantom links" would be inventing the attribution. What it *does*
establish is that the class is live rather than theoretical, and that 8 of the 30 sit in
tool slots (`tool-cta`, `tool-ai-readiness-quiz`, …) — which is why the tool writers are
named as the ranked next candidate.

**Exposure of the path I am actually fixing, stated honestly:** `orchestration_states` (2469
rows, retained from 2026-07-13) contains **zero** runs owned by `section-editor` or
`tool-improver`, and there are **zero** `pages` rows with `page_type='report'`. So both call
sites I am guarding are *reachable and documented* but have not run in the retained window.
This is prevention, not a bleed being stopped, and the bug file should say so.

## 2026-08-02 ~11:05 BST — why the swap branch's UPDATE moves to the caller

`ApplySectionEditAction` has two persist sites, not one: `content_edit` returns its HTML and
the caller writes it (`updatePageComponentAfterEdit`, line 354), while `component_swap`
writes its own row *inside* `applyComponentSwap` (line 798) and then also returns the HTML.
A repair placed at either point leaves the other unguarded — which is this bug's own shape,
one level in. Moving the swap's UPDATE out to the caller gives one repair call before one
persist `switch`.

The locked-component semantics are preserved exactly: `errComponentLocked` was already
handled for both branches at the call site (the swap propagated it out of
`applyComponentSwap` into the `errors.Is` check at line 338), and after the move both
branches return through the same check.

## 2026-08-02 ~11:40 BST — the fix is committed (`66998d300`), and the census figures in it are WRONG

Committed with `Council-Submitted: 0275f9c2-035f-4c9e-8a50-83836dfeffd9` (submitted
~11:20, verdict pending — CLAUDE.md's trailer for exactly this case, so 098 credits the
commit automatically once approval lands).

> **CORRECTION — the numbers in that commit message and in the § census entry above are
> wrong.** I wrote "30 distinct (page, href) … 7 sites … 14 rewritable, 16 unlinkable"
> and "8 of the 30 in tool slots". The aggregate says:
>
> | | href occurrences | components | pages | sites |
> |---|---|---|---|---|
> | rewrite | 18 | 8 | — | 5 |
> | UNLINK | 17 | 8 | — | 5 |
> | **total** | **35** | **13** | **13** | **6** |
> | of which tool-shaped | 7 | 5 | 5 | 3 |
>
> Where the wrong numbers came from: my detail query ended `GROUP BY 1,2,3,4,5` (domain,
> page, slot, href, action) and printed `(30 rows)`. I read that footer as "30 links" — it
> is 30 *groups*, and repeats collapse into one. The "7 sites" was worse: I counted
> distinct domain names by eye down the listing and got it wrong by one.
>
> **What caught it:** writing `RUNBOOK_sibling_link_repair.md`. Making the query re-runnable
> meant giving it `count(DISTINCT …)`, and the aggregate contradicted prose I had already
> committed. Logged in `WRONG_CALLS.md` (2026-08-02, "I read a census off a LISTING").
>
> **What does not change:** the direction of the finding, the fix, and the tool-writer
> ranking. 7 of 35 in tool slots is still the reason those two files are the next candidate,
> and 17 unlinkable hrefs are still live 404s on real pages.

Two more things from the commit itself, both worth the next thread's time:

1. **The `logged-model-output` pattern-check finding on `create_report_page_action.go:428`
   is a false positive, and not mine.** Line 428 is
   `fmt.Fprintf(&b, `<div class="report-prose">…%s</div>`, esc(sec.heading), body)` — a
   write into a `strings.Builder` that IS the page, not a log call. It was already there
   (`5eb433e47`); it surfaced only because my edit put the file in the changed set. Left
   alone deliberately: "don't embed the prose" would break the report. If that check gets
   tightened, `fmt.Fprintf(&builder, …)` is the shape to exclude.
2. **`HEAD` compiles from a clean archive**, not just in this tree:
   `git archive HEAD | tar -x -C <tmp> && go build ./...` → clean. On a shared tree a green
   local build can be green because of somebody else's uncommitted work.

## 2026-08-02 ~12:10 BST — APPROVED at round 1, and the objection that was a genuine catch

`0275f9c2-035f-4c9e-8a50-83836dfeffd9` → **approved**, *"approved with 5 advisory
objection(s) — none high-severity"*, 14 seats reporting (4 abstained). Dispositions are
tabulated in the bug file; three are worth recording here as method rather than outcome.

**1. The `guidelines` seat caught a vacuous negative, and it was right.** It flagged, from
the PLAN alone — it never saw the test — that `TestRepairComponentHTML_ConfigLeverOffDoesNothingAtAll`
might prove inertness via `mock.ExpectationsWereMet()` with no expectations registered.
That is exactly what it did. Two independent reasons it could not fail:

- with nothing registered, `ExpectationsWereMet()` returns nil unconditionally — it reports
  on expectations that WERE registered, not on calls that were not expected;
- the other assert (`got != in`) could not catch it either, because the fail-open path also
  returns the input unchanged. A lever-less version would have returned identical HTML.

Fixed by inverting the assertion: **register the call that must not happen, then require it
to go UNMATCHED.** Mutation C (`if false` in place of the lever check) now fails it on both
lines. The same treatment went on the clean-component test's "and silent" claim.

**And the mutation that DIDN'T fail, which taught me more than the ones that did.**
Mutation D removed the `len(repairs) == 0` early return in the seam and the clean-component
test still passed. My first read was "the test is weak". Wrong: `writeLinkRepairLog` has the
identical guard internally, so behaviour was unchanged and a passing test was *correct*.
Mutation D' removed the inner guard instead — still passed, because the outer one now
short-circuits. Only mutation E, removing **both**, fails the test. The two guards are in
**series**, not alternatives, so no single mutation can disprove the silence claim.
**A mutation that passes is not automatically a weak test — check whether you actually
changed the behaviour before you believe your own mutation.**

**2. Two seats asked for lookups behind assertions, and both were one command away.**
`applyComponentSwap` has exactly one caller (`section_editor_actions.go:331`; the other
grep hits are the definition and five log/comment strings) — that had been asserted in the
submission as "grep confirms", which is not evidence, it is a claim that a grep exists.
And `adopt_verbatim.go` first wrote `page_components` in `e6a8bb63b`, **2026-07-30** —
which confirms "a new writer appeared between the filing (07-28) and the fix" as a dated
fact rather than a story I liked.

**3. The `editquality` seat found the weakest link in the allow-list, from a landmine.** It
noticed `adopt_verbatim.go` was excluded on prose while the blog listing was excluded on
measurement, and cited the `--fidelity high` landmine as reason to distrust the assumption.
The answer holds, but it needed citing rather than asserting: the file writes
`content.RawHTML` (`:514`, `:533`) and stores `sha256(RawHTML)` in `content_data` (`:487`),
so repairing would invalidate its own hash, and it is reachable only under
`fidelity == fidelityLocked` (`apply_adoption_plan_action.go:426`, a strict binary). The
landmine is about which PATH runs, not what that file writes. Allow-list comment rewritten
with the line numbers so the next reader does not have to re-derive it.

**4. Filed RFC 008** for the mandatory-write-seam question, because four seats converged on
"advisory is the wrong ceiling" and `bug_historian` asked for a ticket by name. The RFC
states the case AGAINST as well: two of the ten writers must never repair, so a mandatory
seam needs an opt-out parameter, which is an allow-list wearing a type signature. What
settles it is a measurement nobody has taken — whether advisory `pattern-check` findings
are read and acted on at all.

## 2026-08-02 ~15:30 BST — LIVE on v1.0.1229, closed, and the census re-run found a different bug

**The roll happened (another session's build) and the fix is live.** Both replicas, one exec
each: new marker 0→1, `Component link repair SKIPPED` 0→1, **negative control
`failed to update page_component for swap` 1→0**, positive control 1→1. The negative control
is the one that made this evidence: the same pattern returned 1 on v1.0.1228 an hour before,
so a 0 now cannot be a mis-spelled grep. `bugs_open/136` → `bugs_closed/136`.

**The live-path regression check the pod-grep cannot give you.** This change edited
`rerender_link_repair.go`, a live proven path. Post-roll it wrote **21
`CONTENT_LINK_REPAIR_DETAIL` rows** (`page-rerender` / `render_page` / `rerender_page`,
newest 14:25Z) — correctly labelled, nothing blank. But the honest boundary: those rows come
from `writeLinkRepairLog`, which I did **not** touch. The function I *extracted*,
`writeLinkRepairSkipLog`, has **zero** rows all-window because it only fires on a failed
page-index read. So "the refactor is proven by live traffic" would be an overclaim, and the
bug file says so explicitly.

### The census went 35 → 36, and the new row was not a link

`vetcomparison.uk` / `tool-cma-obligation-checker`, `href="' + q.link + '"`. That is a
**JavaScript string concatenation** that builds an anchor at runtime, sitting inside a
`<script>` block in the stored HTML. My census regex extracts `href="…"` from anywhere in
the document and cannot tell markup from a string literal.

**And then the important part.** I ran the shipping `RepairPageLinks` over those exact bytes:

```
repairs=1  action=unlink href=""
IN : ' <a href="' + q.link + '" target="_blank" rel="noopener">See guide section</a>.</p>'
OUT: ' See guide section.</p>'
```

The href capture `[^"']*` cannot cross the `'` immediately after `href="`, so it captures
EMPTY, takes the `LinkScopeEmpty` arm, and deletes the anchor from the JavaScript. The output
is still valid JS and still reads sensibly — the visitor just cannot click. Filed as
**`bugs_open/180`**, with the probe, the measured exposure (1 component, 1 site, not
protected by any runtime-fill marker on its page) and the honest caveat that the exposure
query catches only ONE SPELLING (a template literal `href="${url}"` would take the *phantom*
arm instead and be unlinked too).

**Three lessons I want the next thread to have:**

1. **My own census figure is an UPPER BOUND, not a count.** It includes at least one
   non-anchor. The correction is not that the number is wrong by one — it is that a regex
   over `href="…"` answers "how many href-shaped byte sequences are there", which is a
   different question from the one I asked. Same family as the `(30 rows)` mistake earlier
   today: twice in one day I let a query's answer wear the noun I wanted.
2. **The ranked "next candidate" flipped.** This file and the bug both said: wire the seam
   into the tool-markup writers next. That is now **wrong until 180 is fixed** — tool markup
   is precisely where JS-built anchors live, so wiring the repair there would delete working
   links from tools. 180 first, then the tool writers.
3. **A probe beats a fetch.** My first instinct was to curl the live page to see whether the
   outbound rerender had already corrupted it; the URL 404'd and the check stalled. Running
   the actual function over the actual bytes took two minutes and gave a deterministic
   answer instead of an observation I would have had to interpret.
