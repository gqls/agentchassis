# NOTES — bugs_open/423 (append-only, newest at the bottom)

## 2026-09-02 — session opens: is there a thread, and is the bug still real?

`scripts/who-owns.py 423` returned **OWNED or recently active**, but on inspection every
commit it counted was a **FILING** commit from the delivery lane (`6ef08c695`, `a0ad26473`),
plus same-day webdesign NOTES commits that only cite the bug. The filer's own NOTES line
settles it: *"Until a lane takes 423, the hand-patch serves — named at review."* So the
**fix** thread was unclaimed. Resumed here.

⚠ Worth stating because I nearly stopped: `who-owns.py` reads COMMITS and cannot tell
"filed this" from "is fixing this". Its VERDICT block says OWNED for both.

**Still valid, and still firing today.** `site_components` for boxingonline.com:
`footer / build_status='pending' / updated_at 2026-09-02 11:22:20Z / digest_ok=false`,
while `head` and `header` are `rendered` with matching digests from 11:23. So a render ran
today, stored two slots, and dropped the footer — the exact 08-31 shape, 2 days on.

## 2026-09-02 — the wrong half of the pipeline (misstep, logged in WRONG_CALLS)

Followed the 08-31 addendum's un-discharged `maskNonMarkup` reading and opened
`markup_spans.go`. The addendum's code read is **correct** — and useless to me, because it
is scoped to *"between RenderTemplate and the store"*, and the cut is **before**
`RenderTemplate`, in an input built at `:125`. Inheriting a conclusion inherits its bounds.
Cost ~10 minutes. What actually found it: a census of byte-indexed slices across the whole
path, which put `render_site_components_action.go:1622` in the output on the first run.

## 2026-09-02 — the cause, proven by execution rather than argued

`buildServicesHTML` (`:1622`, footer services column, called at `:125`):
`words[i] = strings.ToUpper(w[:1]) + w[1:]` after `strings.Fields`. A standalone em-dash
becomes its own WORD, so `w[:1]` cuts a 3-byte rune after 1 byte. Ran it:

```
"—dash"  -> ef bf bd 80 94 64 61 73 68  valid=false
"“quote" -> ef bf bd 80 9c ...          valid=false
"École"  -> ef bf bd 89 ...             valid=false
```

**`0x80` — the live pod capture's byte, exactly.** The site's trigger is
`pages.title` = "Boxing Quiz **—** Test Your Knowledge | Tools" (`tool-boxing-trivia-quiz`,
active, in_header, nav_order 200, so inside the query's `LIMIT 6`).

## 2026-09-02 — the census that could have come out otherwise, and a SECOND casualty

Ran it both ways (full SQL in the RUNBOOK). Sites with a multi-byte-initial word in a
services-column label within `LIMIT 6`: **2**. Sites whose footer is not `rendered`: **the
same 2**. Zero false positives, zero false negatives; every other footer on the fleet is
`rendered` with a matching digest.

The second is **garden-tools.uk**, failing since **2026-08-23** on
"How We Assess Garden Tools **—** Our Methodology | Garden Tools UK" — **ten days, unknown
to anyone**, which is half 1's cost measured rather than asserted. Its `rendered_html` is
**NULL**: nothing has ever been stored for that slot, so it is not "stale", it is absent.

## 2026-09-02 — it is a CLASS, and the estate had already half-fixed it

8 call sites of `strings.ToUpper(x[:1]) + x[1:]` in non-test Go. And `SafeCut` — a
rune-safe truncation primitive — **already existed**, added 2026-07-20 for `bugs_open/027`
§4b. The estate fixed the TRUNCATION shape of this class six weeks ago and never looked for
the CASING shape. That is the argument for one shared primitive over a ninth hand-roll, and
it is the argument against the one-line fix.

## 2026-09-02 — near-miss: `gofmt -w` would have stolen another lane's line

`gofmt -l` flagged the file after my edits. Reflex: `gofmt -w`. Checked first —
`git show HEAD:<path>` is **already** unformatted, from `effc3a090` (the noted_rebuild lane's
`cta_override_rejected` work). Formatting it would have carried their line into my pathspec
commit as a same-file passenger. Left alone; told that lane.

## 2026-09-02 — mutation runs (the tests are red without the fix, and I checked)

All three reverted in a single shell call with a `trap` so the shared tree is never left
mutated between calls:

1. `UpperFirst` → the byte idiom: **2 tests fail**, and the failure output prints
   `ef bf bd 80 94` — the production bytes, reproduced by the test harness.
2. `SafeCut(summary,247)` → `summary[:247]`: the sqlmock expectation goes unmet.
3. `InvalidUTF8At` → always clean: its test fails.

Full suite green after revert (`datahelpers` ok, `actions` ok, 5.8s).

## 2026-09-02 — submitted, committed, NOT yet live

Council `dc62975f-9d38-4b3c-9174-330307b9df95` (`Council-Submitted:`, per the rule that a
trailer you cannot yet have read must not claim a verdict). Registered **STY-059**.
**Go, so INERT until the next chassis roll** — both casualties stay broken until then, and
no claim in this file about the fix WORKING has been made, only about the fix EXISTING.
