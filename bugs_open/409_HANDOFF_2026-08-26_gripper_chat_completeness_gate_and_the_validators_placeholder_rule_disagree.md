# 409 — gripper intake: the chat's completeness gate accepts what the validator's placeholder rule will refuse to publish — and the visitor's word "travel" binds into a field that means jaw span

Filed 2026-08-26 by the gripper-dossier ("AI page 3") lane, from the pilot's first two
LIVE production requests (both driven by this lane as the visitor, so every link is
first-hand: chat JSON, stored spec, validator issue row, orchestration error).
Two findings, one seam: **what the chat accepts and what the pipeline downstream will
tolerate are decided by two mechanisms that have never been introduced to each other.**

## Finding 1 — vague-but-complete: `complete:true` on a value the prose cannot state honestly

**Evidence chain (request `6dac176b-93d2-423a-8fa3-0fa2be6611c9`, 2026-08-26 09:16–09:21Z):**
1. Visitor names the mounting but not the flange standard. The chat records
   `mounting: "Gantry, ISO 9409 flange (standard not yet specified)"` — a faithful
   record, per its "record only what the visitor states" rule — and returns
   **`complete: true`** (the field is non-null; `Missing()` checks nulls only,
   `internal/tools-api/gripper/spec.go`).
2. The report-builder's prose writer, honestly, writes: *"; that detail needs to be
   confirmed against the VG10's flange before ordering."*
3. `validate_page_content` blocks: `placeholder_text / "to be confirmed" / severity
   blocker` (agent_error_log `context.issues`, orchestration
   `86fd2cfd-3858-4503-a52c-e7cf862fd39b`). Workflow fails out; failure sidecar
   published; visitor gets the apology email.

**The defect is the seam, not either mechanism.** The chat's completeness gate is
null-based; the validator's honesty gate is phrase-based; a visitor who leaves ANY
detail vague therefore gets a FAILED report (apology email) instead of either (a) a
follow-up question pinning the detail, or (b) a published report carrying an honest
caveat. Every vague visitor is a guaranteed failure that spends the full build
(~3–8 min + LLM calls) before failing.

**Family:** `bugs_closed/377` (placeholder pattern "your company" convicting ordinary
B2B prose) and `bugs_open/218` (placeholder scan convicting JavaScript) — the same
detector's precision problem. NEW here: the upstream gate that *manufactures* the
convicting prose is ours, and it is reachable by any ordinarily-imprecise visitor.

**Fix candidates, ordered by what closes the door** (per the estate rule):
1. **Make vagueness unrepresentable in a complete spec**: the chat's field guidance
   gains a per-field "a value containing a hedge ('not yet specified', 'TBC',
   'unknown') is NOT a recorded value — ask once more or record null". Closes the
   door at the cheap end, before any build spends.
2. Teach the prose prompt the validator's banned phrases (express caveats as
   engineering advice — "confirm X before ordering" is fine wording that does not
   trip the current pattern; "needs to be confirmed" is not). Fragile: two lists to
   keep in step.
3. Whitelist honest-caveat phrasing in the validator for report pages. Weakens the
   guard that 377 shows already errs the other way; least preferred.

## Finding 2 — the visitor's word "travel" binds into `travel_mm`, which means JAW SPAN

`travel_mm`'s guidance (spec.go:70): *"the jaw opening the gripper must span to pick
the part … usually the part's width or diameter"*. In robotics vernacular "travel"
is ARM MOTION. Evidence, same morning, same inputs, two paths:

- **Session 1** (`14cc5bc5…`): visitor volunteers "Pick-and-place travel is about
  300 mm" → chat binds `travel_mm: 300` with **no clarifying question** →
  `complete: true`. For a 120×80×40 mm part a 300 mm jaw span is nonsense — this
  spec would have scored grippers against a wrong premise; only Finding 1's blocker
  stopped the report shipping.
- **Session 2** (`2ffa5871…`): the model itself asked "which dimension does the jaw
  span?" → bound `travel_mm: 80` (the width) → and then correctly REFUSED the
  visitor's attempt to "correct" it to 300, explaining the semantics. (That refusal
  is GOOD behaviour and worth protecting — this lane initially misread it as a
  merge bug; refuted by reading spec.go:70 before filing. NOTES 2026-08-26.)

So the binding is path-dependent: the clarifying-question path lands the right
value, the visitor-volunteers path lands the wrong one, and both read `complete`.
**Fix candidate**: the guidance already says "ask for it if the geometry does not
make it plain" — extend it to "if the visitor uses the word 'travel'/'stroke'/'reach'
for a distance, confirm which distance before recording", or rename the asked-for
concept in the question the model poses (it never says "travel" to the visitor; the
trap is only on volunteered values).

## How to verify any fix

Replay both transcripts (in this file's evidence + NOTES 2026-08-26) against the
changed prompt: session 1's volunteered "300 mm travel" must NOT bind 300 without a
question; a hedged mounting must NOT reach `complete: true` (option 1) or must
produce publishable prose (options 2/3). Then one live E2E each way: vague →
follow-up question; precise → emailed report (the 613916a7 baseline).

## Status

Both requests terminal and correct-by-design at the visitor surface (apology + real
dossier respectively). No workaround needed for launch-to-nobody (page unlinked);
**worth fixing before the widget goes on the site** — the widget invites exactly the
vague first message that hits Finding 1.

---

## FIX BUILT 2026-08-26 (same day) — chat side; NOT yet live; 409 stays OPEN until the live replay passes

Fix candidate 1 implemented at `Normalise`'s single choke point (`coerce` KindText
rejects hedge-phrased values — covers the chat AND plain-form doors), plus the
matching prompt rule, plus travel_mm/mounting guidance for finding 2. Tests added;
the guard is MUTATION-PROVED (removing `containsHedge` fails
`TestNormaliseRejectsHedgedTextValues` on every hedged value, including the exact
live-failure string). Council: `Council-Submitted: 70083c99-c299-4b35-a868-1583d3355396`.

**Inert until the island image rolls** (owner-run swap). Close criteria: replay the
session-1 shapes live — (a) hedged mounting → the assistant ASKS AGAIN, spec stays
incomplete, no build spent; (b) volunteered "300 mm travel" → clarifying question,
not a bind; (c) the 613916a7 happy-path baseline still passes. The cluster-side
prose/validator seam is deliberately untouched — if hedges can no longer enter the
spec, the prose has nothing to hedge about; if a live run still manufactures
"to be confirmed" from a clean spec, that is a NEW finding against the prose
prompt, filed separately.

## COUNCIL ROUND 1 = REVISE (same day) — the objection was RIGHT, and round 2 answers it in code

`editquality`, gating, on edit 2: finding 2 got prompt/guidance coverage only,
against the submission's own "a prompt line is not a control". Correct — and the
REVISE round earned its cost exactly as the estate's memory says it does.

**Round 2** (commit `0419ca584`, resubmitted on the same corr `70083c99`):
`reconcile()` — a cross-field pass at the END of `Normalise`, which every real
spec path exits through (chat's SQL-merge RETURNING, session rescan,
submit-from-session, plain-form inline), so there is no second call site to
forget. Drops `travel_mm` > **1.5×** the largest number stated in
`part_geometry` (the live mis-bind was 2.5×). Fails OPEN on absent/numberless
geometry — the guard catches a contradiction between stated facts, it does not
demand facts. `Merge` mirrors it. Boundary pinned by tests (180 keeps, 181
drops); mutation-proved. Image **v1.0.1343** (`3abb46509`).

## LIVE REPLAYS on v1.0.1342 (guidance half): (a) and (b) PASS

Session `21e67276…`, the exact session-1 input shape (vague flange AND
volunteered "300 mm travel" in one message):
- (a) mounting recorded as the CLEAN partial fact `"gantry, ISO 9409 flange"` —
  no hedge phrase; **PASS**.
- (b) `travel_mm` NOT bound; the assistant asked the discriminating question
  verbatim ("is that the robot's movement, or the distance the jaws must
  open?"); **PASS**.
- (c) baseline: same session completed correctly (travel 80, full flange),
  submitted, in flight at write time.

**Still owed before CLOSE**: the v1.0.1343 swap (code guard live), one
behavioural probe of the code guard (geometry + travel 300 in ONE message →
travel absent from the recorded spec), (c) reaching `emailed`, and the round-2
council verdict read.
