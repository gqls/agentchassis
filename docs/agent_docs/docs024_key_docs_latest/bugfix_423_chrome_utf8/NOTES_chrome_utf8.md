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


## 2026-09-02 — council REVISE, and the seats earned it

Gated by guardian HIGH: I asserted the blast radius instead of enumerating it. Enumerating
it changed the design — 7 workflows, none with an `error_step`, so the escalation went
opt-in (`escalate_chrome_store_failure`, default OFF) with the reporting half left
unconditional because the silence is the actual bug.

bug_historian's census ask was the one that most improved the work, and it cut BOTH ways:
it found the one remaining swallow (:1411) **and** the measurement that says leave it alone
— 23 of 57 sites have no chrome rows, so reporting there would file ~69 items about sites
nobody ever built. **My instinct on reading the objection was to widen the fix; the
measurement I ran to justify widening is what stopped me.** That is the rule working in
the direction it is least often quoted for.

⚠ **Misstep of process, not of fact:** I could not run the package tests at all for a
while — two other sessions had `seed_content_sources_action.go` and
`apply_theme_kit_action.go` mid-refactor in the shared tree, so `go test ./actions/` failed
to build on THEIR code. Resolved without touching their files by pinning every peer-dirty
file in the package to HEAD with `go test -overlay` (28 files, listed in the run). Worth
knowing: on this tree a red package build is not evidence about your own change, and the
overlay is the non-invasive way to find out which.

Round 2 resubmitted on the same correlation so the trail accumulates.


## 2026-09-02 — council round 2 REVISE: both HIGHs were about the SUBMISSION, not the code

editquality HIGH: "neither path constructs the sentinel, so the gate is a no-op". Reading
the sketch, exactly right — I rewrote round 2's rationale and carried round 1's sketch over
unchanged, so it still showed a bare `fmt.Errorf`. The CODE was correct (`:1416`, `:1454`
both construct `&chromeStoreRefusedError{...}`), the SHOWN diff was not, and a reviewer can
only judge what is shown. **A rationale and its sketch are two claims and they can drift
apart in one edit.**

prior_art HIGH: the seven-workflow enumeration was the load-bearing absence claim for the
whole round-2 design and I gave only its conclusion. Query now in `grounded_in` verbatim.

Best find of the round, from reuse_agent: I never checked for an existing title-caser.
Censused now — the only in-house one (`harvest.go:405 titleCase`) turns out to BE one of the
eight sites I converted, `strings.Title` (10 uses, incl. `:1646` in `loadNavItems`) is
rune-based and not a member of this class, and `x/text` is indirect and lower-cases the
remainder so it fails the ASCII-parity test that licensed the one-pass conversion. Nothing
changed, but the round-1 claim was unearned.

bug_historian: `:1411` filed as **bugs_open/435** so the next reader inherits tracked work
rather than a measurement in a closed file; and `bugs_closed/054` now cited at the code as
the escalation precedent this REUSES rather than parallels.

architecture: declaring `ConfigKeys` for this action is right and is NOT in this change —
live steps carry six keys and a wrong list hard-fails every workflow that stamps the action,
so it is its own change. Same argument the guardian used against my round-1 escalation.

⚠ **Misstep, logged in WRONG_CALLS:** I published round 3 in the same shell call whose edit
had just failed an assertion, joined with `;`. The submission asserted a code citation that
did not exist for four minutes. `&&`, not `;`, behind anything that tells someone else.


## 2026-09-02 — round 3 APPROVED, and then I deleted the thing round 2 was about

bug_historian's medium advisory sent me back to my own caller and I found that the opt-in
arm could never gate what I thought it gated. `chromeUnserved` is appended to only when
`!chromeSlotHasStoredHTML`, so the arm's entire reach was "a slot with nothing to serve" —
the case 260 and 054 already ruled must fail.

**My round-2 reasoning error, stated so I do not repeat it:** I measured **7 workflows
with no error_step** and treated that as the blast radius. It is not. It bounds WHO is hit
when the arm fires; it says nothing about HOW OFTEN it can fire. The second number is ONE
row fleet-wide. I answered the wrong question with a real measurement, which is worse than
not measuring, because it *felt* rigorous — and it survived a HIGH objection, a redesign
and two rounds of review before a seat asked the question from the other end.

Gate deleted, sentinel deleted, key deleted, three tests deleted with it. Full suite green
under an overlay isolating 17 peer files. Strictly less machinery than round 3, and it
dissolved four advisories instead of deferring them.

⚠ Overlay gotcha found the hard way: my first re-run FAILED because the overlay pinned my
own DELETED test file back from HEAD (`git status` shows ` D`, and the script treated it
like any tracked peer file). A deleted file of your own must map to `""`, not to its HEAD
copy — otherwise you are testing a symbol you just removed.


## 2026-09-02 16:21Z — LIVE, and half 2 proven at the artefact

Fresh chassis build `v1.0.1354`. Probed the running binary rather than trusting the tag:
both added literals present, the DELETED emitter text absent (a removed-string control,
the strongest available), nonsense absent. `escalate_chrome_store_failure` also present —
which correctly says the build carries `cccb5ccd6`, the GATED version, and not my
uncommitted-at-the-time deletion.

Then the behaviour, which is the part a probe cannot tell you: dispatched `rerender-chrome`
for garden-tools.uk (`af0857d2`, publish receipt asserted via kafka-publish-lib rather than
hand-rolled kcat). Footer stored at 16:21:32Z — **2,427 bytes after ten days of NULL** —
digest matching, header and head stored in the same run so it was real work.

**The check that actually proves the mechanism rather than the absence of a crash:** the
offending label renders intact — `How We Assess Garden Tools — Our Methodology | Garden
Tools UK` — em-dash preserved, not mangled to U+FFFD and not dropped. A footer that stored
because the label had been silently discarded would have looked identical in every other
column.

Census after: unstored footers **2 → 1**, NULL/empty `rendered_html` rows **1 → 0**.

Chose NOT to re-render boxingonline: paid site, mid-delivery by another lane, and the
re-render replaces a hand-patch that is currently the only definition of its footer. Told
that lane; it is theirs to fire.
