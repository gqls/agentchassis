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
