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
- No `LANDMINES.md` entry footprints `checkMetaCommentary`, `metaCommentaryPatterns`
  or `validate_page_content.go`'s meta scan. The nearest entry (line 1895)
  concerns the runtime-fill marker and explicitly lists `validate_page_content.go`
  as a **caller, not a test site**.

### Diagnosis provenance

No `090` run, and this is a **stated substitution** under the 2026-07-31 owner
ruling. This is a pattern-precision defect in one function: the conviction was
reproduced by executing the current code over the real stored bytes, the
matching substring was read in its rendered context, and the negative control
(four rows that must NOT convict) came out negative. There is no cross-cutting
structural claim here — the cross-cutting half of this check was `bugs_open/219`
(scope), which is fixed and whose fix these same four rows demonstrate working.
