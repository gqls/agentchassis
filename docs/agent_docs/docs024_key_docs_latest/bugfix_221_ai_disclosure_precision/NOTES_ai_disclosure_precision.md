# NOTES — bugs_open/221, `as an ai` convicts legitimate copy about AI

Append-only, newest at the bottom. Technical log: what was tried, what the
system actually said, and every misstep.

---

## 2026-08-08 — session start: pick-up and validity re-check

**Why this bug.** The session was pointed at `bugs_open/116` as a possible
target. It is not one: `scripts/who-owns.py 116` names two ACTIVE owning
workstreams (`bugfix_179_deploy_path_override`, `bugfix_116_link_check_coverage`),
run 1 of its D3 programme fired **today** (2026-08-08, orchestration
`99d39725-1223-4cbb-b4c5-60e66483d734`), and the file's own status block says it
is **owner-gated and not a coding task**. Taking it would have been competing
with a live lane on a question the owner has reserved.

Swept every other file in `bugs_open/` through `who-owns.py`. **221 is the only
one with no owning workstream.** Two confirmations beyond the script, because
`who-owns` reads commits and is blind to a session mid-fix:

- The filing lane disowns it in writing —
  `loancalculator_couk/HANDOFF_2026-08-08_continue_here.md:28`: *"221 OPEN,
  belongs to webdesign.co.uk — same check, different mechanism"*, and `:71`
  lists it under what that lane will **not** solve.
- Grepped the live `.jsonl` transcripts of the 14 most recently active sessions
  for `metaCommentaryPatterns|checkMetaCommentary`. Four hit. Three are this
  session or incidental (a `ls bugs_open/` listing). The fourth
  (`8134dee6…`) is the loancalculator voice lane — its recent tool calls are
  all voice rewrites and CSS restoration, not this check.

### The bug is STILL LIVE at HEAD — verified by running the real code, not by SQL

SQL first, as a locator only (1,247 `page_components` rows on `pages.status='active'`):

```
 rows_scanned | as_an_ai | lang_model | refusal | positive_control
         1247 |        1 |          0 |       0 |              281
```

The positive control (`calculator`, 281) is what makes the 1 meaningful — a zero
here could have been a broken query rather than a clean fleet.

Then the load-bearing check, because a `LIKE` is not the validator: dumped the
stored `rendered_html` of all five rows matching **any** of the multi-word
patterns and ran the **real `checkMetaCommentary`** over those bytes in a
scratch test in-package:

```
4149fea8… (webdesign.co.uk/tools-index/ported-page)      -> 1 issue
    BLOCKER value="as an ai" location="calBusiness schema, as an AI-builder prompt"
6c1a1ac5… (loancalculator/tool-interest-rate-stress-test) -> 0 issues
993eda99… (loancalculator/index/tool-3)                   -> 0 issues
d54dd48e… (loancalculator/tool-consolidation)             -> 0 issues
f74f84f2… (loancalculator/tool-car-finance-calculator)    -> 0 issues
```

**Those four zeroes are a real negative control, not filler.** They are the
`input_schema`/`on_missing` rows — the exact bodies that convicted three pages
before `bugs_open/219`. They return 0 now because 219's scope fix works. So the
harness can distinguish a conviction from a pass, and the single remaining
conviction is the one 221 describes.

The convicted copy, in full context (a product description on a tools index):

```html
<a class="index-card" href="/tools/seo-injector/index.html">
  <h3>JSON-LD SEO Injector</h3>
  <p class="index-subtitle">LocalBusiness schema, as an AI-builder prompt</p>
  <p class="index-desc">Turn local business details into strict LocalBusiness JSON-LD wrapped in a prompt.</p>
</a>
```

### MISSTEP — `kubectl exec -i` inside a `while read` loop ate the loop's stdin

The first dump loop wrote **one** file and exited, silently and at exit 0. The
`-i` flag makes `kubectl` read stdin, and inside `while … read < manifest` the
stdin it drains is the manifest. It looks exactly like a one-row result set.
Fix: `kubectl exec -i … < /dev/null` inside the loop. Logged to `WRONG_CALLS.md`;
the check is in the RUNBOOK.

### Framework context gathered before planning

- The same class is open as `bugs_open/222` (the fabrication gate's regex
  convicts a comment that *denies* fabrication) and is **owned** by the
  mortgagecalculator lane. Do not fix it here. Shape this fix so it does not
  foreclose theirs.
- `bugs_open/221` itself rules the wider question out of scope: *"If a fixing
  thread wants to widen this to 'how should the fleet's blocker-severity string
  scans be governed', that is a 090 or an RFC, not this file."* Honoured — the
  change stays inside `checkMetaCommentary`.
- ~~No `LANDMINES.md` entry footprints `checkMetaCommentary`, `metaCommentaryPatterns`
  or `validate_page_content.go`'s meta scan. The nearest entry (line 1895)
  concerns the runtime-fill marker and explicitly lists `validate_page_content.go`
  as a **caller, not a test site**.~~

  > **CORRECTED 2026-08-08 (same session, ~2h later) — THIS WAS FALSE, and it was
  > false when I wrote it.** There *is* such an entry, and it is a good one:
  > `LANDMINES.md:6413`, *"A new pattern in `validate_page_content` is a BLOCKER by
  > default, and a blocker there means 'this page can never be rebuilt again'"* —
  > footprinting `metaCommentaryPatterns` and `checkMetaCommentary` by name, and
  > **already citing `bugs_open/221` as one of its two live instances.** It was
  > committed in `e4d620fca`, the same commit that filed 221, i.e. it existed
  > before this session started.
  >
  > **What caught it:** nothing I did deliberately. I only found it because
  > `landmines_lib.parse()` reported *two* entries matching `metaCommentaryPatterns`
  > when I expected one — I was checking my own append had synced, and the
  > duplicate fell out. Had I not run that check for an unrelated reason, I would
  > have shipped a second entry competing with the first.
  >
  > **The cause:** I ran `grep -n "…|checkMetaCommentary|…" LANDMINES.md | head -30`
  > over a **7,000-line** file. The 30 matches I read all came from the first 1,921
  > lines. The entry is at line **6,413** — *below the cut*. `head` does not say it
  > truncated, so a partial answer arrives looking exactly like a complete one, and
  > I wrote "no entry exists" from it.
  >
  > **The check, which costs nothing:** for an existence question, `grep -c` FIRST
  > (or `| tail`, or no pipe at all). A count cannot be truncated, and if it exceeds
  > the window you are about to read, you know your sample is a sample. **`head` is
  > for previewing output you already know the size of; it is never an answer to
  > "does this exist?"** Logged in `WRONG_CALLS.md`.
  >
  > **Consequence, already applied:** the duplicate entry I had appended was
  > **removed**, and only the genuinely-new material (the `Re` field and its unsafe
  > nil default; the "SQL `LIKE` is not this check" trap; the
  > `error_message`-vs-`context.issues` trap) was folded into the existing entry as
  > a dated addendum. One entry per trap — two would have drifted.

### Diagnosis provenance

No `090` run, and this is a **stated substitution** under the 2026-07-31 owner
ruling. This is a pattern-precision defect in one function: the conviction was
reproduced by executing the current code over the real stored bytes, the
matching substring was read in its rendered context, and the negative control
(four rows that must NOT convict) came out negative. There is no cross-cutting
structural claim here — the cross-cutting half of this check was `bugs_open/219`
(scope), which is fixed and whose fix these same four rows demonstrate working.

## 2026-08-08 — sizing the real-world damage, and a second misstep of my own

**Question:** has this check actually blocked anything, and has the `as an ai`
arm ever fired?

### MISSTEP — I searched `agent_error_log.error_message` and got 0, which would have read as "this check has never fired at all"

First query filtered `error_message ILIKE '%meta_commentary%'` (plus two other
spellings) over all history: **0 rows**. Written up carelessly that becomes
"the meta-commentary check has never blocked a page", which is **false** — it
blocked three page builds on loancalculator.co.uk on 2026-08-08, which is the
whole of `bugs_open/219`.

The messages carry no detail at all:

```
step validate_content failed: … content validation failed: 1 blockers, 0 errors
Validation produced 1 blocker(s) and 0 error(s); see context.issues for detail
```

The message literally *tells you* where the detail is, and I had searched the
one column that structurally cannot contain it. This is
`a-grep-proves-absence-only-for-its-spelling`, with the aggravation that the
right place was named in the text I was reading.

**The check that works, and is now the RUNBOOK's:** enumerate the jsonb rather
than pattern-match a rendered string —

```sql
SELECT iss->>'category' AS category, iss->>'value' AS value, count(*) AS hits,
       count(DISTINCT domain) AS domains, max(occurred_at)::date AS newest
FROM agent_error_log ael, jsonb_array_elements(ael.context->'issues') iss
WHERE ael.occurred_at > now() - interval '14 days'
GROUP BY 1,2 ORDER BY 3 DESC;
```

(Deliberately NOT `context::text LIKE '%"category":"meta_commentary"%'` — jsonb
renders a space after the colon, so that form matches nothing and returns a
confident zero.)

### What it says [MEASURED 2026-08-08, retention floor 2026-07-09, 19,493 rows]

```
 category        | value        | hits | domains | newest
 meta_commentary | input_schema |    6 |       1 | 2026-08-08
 …
 (no meta_commentary row with value 'as an ai', in the whole retained window)
```

Two findings, and they point opposite ways:

1. **The `as an ai` arm has never fired in production.** 221 is a **latent
   trap, not an active fire** — exactly as the bug file states, and now
   independently confirmed from the error log rather than from the page census.
   Nobody has requested a rebuild of webdesign.co.uk `tools-index` since the
   copy landed.
2. **The check itself is not theoretical.** The same function's `input_schema`
   arm blocked **6** builds on 1 domain, most recently the day of filing. This
   check demonstrably stops page builds when it matches. So the consequence
   claimed for 221 — "that page cannot be rebuilt" — is not an inference from
   reading the severity constant; the same code path has already done it to
   somebody else's pages, six times, this week.

**[ASSUMED] Not established:** that no page was *silently* abandoned earlier
than the retention floor (2026-07-09). The window is 30 days of rows but the
check's scope changed on 2026-08-08 (219), so pre-219 counts measure a
different function and are not comparable.

### The exposed surface, because "one page" understates it [MEASURED 2026-08-08]

```sql
SELECT count(*) FILTER (WHERE pc.rendered_html ~* '\mAI\M')            AS rows_mentioning_ai,
       count(DISTINCT s.domain) FILTER (WHERE pc.rendered_html ~* '\mAI\M') AS domains_mentioning_ai,
       count(*) FILTER (WHERE pc.rendered_html ~* 'as an AI')          AS as_an_ai
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE p.status='active';
```
```
 rows_mentioning_ai | domains_mentioning_ai | as_an_ai
                283 |                    10 |        1
```

**283 component rows across 10 domains already use "AI" as a word** (word-boundary
match, so `said`/`plain` do not count). One of them has so far landed on the
five-character sequence that convicts. The fix is therefore not "unblock one
page": it is removing a trap that 283 rows of existing copy, and every future
sentence written about AI tooling on ten sites, can step on. The estate's own
subject matter is what loads the mine.

### Where the hit sits in the extractor's output — checked, not assumed

Ran `ExtractAssertionText + headProseBlocks` over the offending artefact: **154
blocks**, and the conviction is block 134, which is exactly the paragraph text
and nothing else:

```
block[134] (len 45): "LocalBusiness schema, as an AI-builder prompt"
```

Useful for the fix design: blocks are **per-element prose units**, not the whole
document, so a pattern needing first-person context has a sentence-sized window
to work in — the disclosure form (`As an AI, I cannot…`) and the noun-phrase form
(`as an AI-builder prompt`) arrive as separate blocks, not as one haystack.

## 2026-08-08 — the fix, and a third misstep (a check that measured nothing and said so with a number)

Design in `PLAN_2026-08-08_ai_disclosure_precision.md`; prepared with fable,
then every load-bearing claim in it re-verified here before implementing (one
production call site, `regexp` already imported, the extractor's block shape).

**Sequencing that earned its keep:** the test went in FIRST and was run at HEAD.
7/7 must-not-block cases failed, 7/7 must-block cases passed. That is worth more
than the green run afterwards — it proves the test can fail, and it proves the
change is a pure narrowing rather than a rewrite of what blocking means.

Committed `61c8cc6ff`. Council submission `377a0488-214e-4e5c-bd3d-66343d34d9b2`
(`Council-Submitted:` trailer — the code is on the shared branch, per the
2026-07-20 rule that a coherent change is not held for a verdict).

### MISSTEP — `sed -n '/^var .../,/^}/p'` stopped at `}{`, so my "nothing was dropped" check compared 15 patterns against ZERO and I nearly read that as a diff

Before committing I wanted to assert that no pattern had been silently lost when
the table was rewritten from positional to keyed literals. The check:

```bash
comm -3 <(git show HEAD:…  | sed -n '/^var metaCommentaryPatterns/,/^}/p' | grep -oP '…') \
        <(…now…)
```

It printed all 15 patterns as "only in NOW", which reads exactly like *the
entire pattern set is new and HEAD's is gone*. Alarming, and false.

The range terminator `/^}/` matched the line **`}{`** — the closing brace of the
anonymous struct *type*, immediately followed by the opening brace of the
literal. So the HEAD-side range was four lines long, ended before a single
pattern, and yielded **zero**. The NOW side used `/^}$/` (anchored), which
skips `}{` correctly, so only one side was broken — and the asymmetry is what
manufactured the scary output.

**Two lessons, and the second is the general one:**

1. Anchor a sed range terminator (`/^}$/`), because `}{`, `})` and `},` all
   start with `}` in Go.
2. **A comparison where one side silently yields nothing does not report "I
   measured nothing" — it reports a maximal difference.** An empty operand and a
   total mismatch are the same output. The cheap check, which I only ran because
   the result looked wrong: **print each side's COUNT before diffing them**
   (`HEAD count: 15 / NOW count: 15`). A zero on either side then names itself
   instead of masquerading as a finding.

This is the `check-answers-the-question-you-encoded` class, with the aggravation
that the failure direction was *alarming* rather than reassuring — which is the
lucky case. The same broken extraction, had it been on the NOW side, would have
printed a clean empty diff and I would have committed believing the set was
verified. **A silent-empty operand is dangerous in whichever direction it falls;
it just only gets caught in one of them.**

Corrected check, now in use — both counts printed, then `diff`:
```
HEAD count: 15 / NOW count: 15
IDENTICAL — no pattern added, dropped or altered
```
