# HANDOFF — webdesign tool rebuilds. START HERE. Written 2026-08-17 18:35Z. Supersedes `HANDOFF_2026-08-16_continue_here.md`.

Read: this file → `PLAN_2026-08-15_…` (design + THREE owner rulings + two corrections) →
`RUNBOOK_…` (every command, incl. the two filing gates) → `NOTES_…` (evidence + missteps, newest at
the bottom) → `SUMMARY_2026-08-16_…` → `architecture_review/RFC_036_…` (the parked question).

## The recipe. It is PROVEN — three tools live, one failed for a reason now designed out.

1. **Read the LIVE tool's `<script>`** and write the spec from its behaviour. Never from its page copy.
   Describe *intent* where the ported version is defective (3 of 5 so far were), but do not add features.
2. **File with BOTH gates as pre-asserts inside the transaction** (RUNBOOK "Scope the batch correctly"
   + "Before REFILING"). They are near-complements; passing one says nothing about the other:
   - `idx_cc_tool_function_unique` — **fleet-wide** on `function`, `WHERE component_level='tool' AND
     forked_from IS NULL AND is_active`. **No `site_id`. Forks exempt.**
   - the generator's `already_exists` probe — **per-site**, joins `page_components`, **no
     `build_status` filter**, so a `removed` fork still blocks.
3. **Grade the RUN, not the item** — a failed build reports `complete` with `error` NULL.
   Want `current_step='complete'`, `create_result.page_adopted='true'`, no `already_exists`,
   no `__step_error`.
4. **Grade the COMPONENT before retiring anything.** `{{\.` count **0** (the real gate) · every
   control has a literal label · has `<script>`, no `<script src=`. **Read the JS.**
   ⚠ `visible_chars > 300` is a SMELL, not a gate — a minifier legitimately scored 243 (NOTES 18:1xZ).
5. **Record the revert handle, then retire IMMEDIATELY.** `page_components` row id + `length` + `md5`
   *before* filing; then the guarded `build_status='removed'` UPDATE (exactly 1 row) — and the md5
   after must equal the md5 before. **There is NO `page_component_history` archive row for a retire**
   (the trigger is `AFTER UPDATE OF rendered_html`); the surviving row IS the handle.
   **The race is real: the generator queues its own assemble-only rerender and the margin has been as
   little as ~2 minutes.** Retire the moment the component grades — not after writing it up.
6. **Grade at the served page**, `http=200` asserted FIRST (`/tools/<x>/` is a 404 that passes every
   cleanliness check), with a negative and a positive control in the same breath.

## State (all `[MEASURED 2026-08-17 18:35Z]`)

| | |
|---|---|
| chassis | **`v1.0.1307`**, pods 17:05Z, stamp **`a6d1c53c0`** — a REAL roll. (`v1.0.1305` was NOT: same tag ⇒ cached image, 267 commits inert. Verify per roll.) |
| **live + proven, owner-approved** | `tool-aspect-ratio` · `tool-markdown-tables` |
| **retired, awaiting rerender** | `tool-html-minifier` — rerender `b137a7ce-dc4b-428f-b359-ed1a52bf521d`, `triaged` |
| **building** | `tool-svg-optimizer` — add_tool `7ced2d32-4d63-4b88-b98f-e5e2eeb7d847`, `triaged` |
| **next to file (#6)** | `tool-sri-generator` — spec researched, see below. Blocked only by the serial throttle. |
| **parked on RFC_036** | `tool-ab-test-calculator` · `tool-meme-generator` |
| remaining after those | ~55 |

## Next actions, in order

1. **Grade `tool-html-minifier` at the served page** once `b137a7ce` completes (baseline taken:
   `ported-page` 1, `htmlInput` 2, before). Then it joins the owner-review set.
2. **Grade `tool-svg-optimizer`** through steps 3–6 when `7ced2d32` completes. Revert handle already
   recorded: slot `665075ab-591c-47e8-af95-faab9f48b73d`, 5,095 chars, md5 `be0f5c3530636eddb04e03c82141d8a8`.
3. **File #6 `tool-sri-generator`** the moment no `add_tool` is open (serial throttle).
   Page `211c3abc-d036-40a1-a5cf-ff6708efaba4`, `/tools/sri-generator/index.html`.
   Revert handle: slot **`7d4b69db-66a0-4c81-ba4e-6e27ab09fc49`, 4,752 chars, md5 `16332add36f84bddc5da40d8aa5d59c3`**.
   **Live behaviour, already read:** one textarea (`inputCode`) for pasted file content; `TextEncoder`
   → `crypto.subtle.digest('SHA-384', …)` → base64 via `btoa(String.fromCharCode.apply(null, …))` →
   emits `integrity="sha384-<b64>" crossorigin="anonymous"`. Defects to specify as fixed: `alert("Copied!")`
   and inline-`onclick` binding (use listeners + inline "Copied" feedback), and empty input early-returns
   leaving **stale output** (should clear). `crypto.subtle` requires a secure context — the page is HTTPS,
   so fine; ask for a graceful message if unavailable. Keep SHA-384 only; do not add an algorithm picker.
4. **The owner reviews #4/#5/#6 together** (he said so). Cadence otherwise remains one-at-a-time.
5. **RFC_036 stays open.** Owner direction: keep the library-and-fork model, other sites fork.
   ⇒ option 2 (a rebuild records `forked_from`), NOT option 1. **Do NOT deactivate `8c9a6e06`
   (ab-test) or `6ae53f32` (meme-generator) to unblock those two** — both have live forks elsewhere,
   and deactivating is exactly what makes a tool un-forkable, which contradicts the direction.

## Traps this lane has paid for (each cost a real cycle)

- **A written precondition saying "must be empty" is a STOP, not an input to a judgement.** I ran it,
  got 3 rows, reasoned past it with the *other* gate's logic, and lost a build. Freshly-read code for
  the wrong mechanism is more dangerous than ignorance. (`WRONG_CALLS.md` 2026-08-17.)
- **A negative control alone cannot license a negative finding.** Something must MATCH in the same run.
- **Do not sweep a commit window against `/proc/1/exe`** — ~20–30 s per grep; budget ~3 per exec,
  180 s timeout, and read other lanes' commit messages for the stamp first.
- **Elapsed time comes from the row (`now() - created_at`), never from the session clock.** I announced
  a 19-hour stall on a row one minute old.
- **A categorical census used to prove an ABSENCE must not carry a `LIMIT`.**
- 3 of the 5 tools examined so far were **worse than their page suggested** — two had their
  comment-stripping call swallowed by its own comment (censused: exactly those two, class closed).

## Not this lane's, but sits underneath it

`bugs_open/289` — `build-dispatch-loop` state doubles per iteration. Alive and running today
(measured), but every item here goes through it. If the queue stops, look there before anything else.
