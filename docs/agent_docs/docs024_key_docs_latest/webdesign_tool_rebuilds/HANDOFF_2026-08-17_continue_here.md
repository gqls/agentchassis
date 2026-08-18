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

## State (all `[MEASURED 2026-08-18 07:45Z]`)

| | |
|---|---|
| chassis | `v1.0.1307`, stamp `a6d1c53c0` — a REAL roll. (`v1.0.1305` was NOT: same tag ⇒ cached image. Verify per roll, with a control that MUST match.) |
| **LIVE + PROVEN (5)** | `tool-aspect-ratio` · `tool-markdown-tables` · `tool-html-minifier` · `tool-svg-optimizer` · `tool-sri-generator` |
| owner has approved | aspect-ratio, markdown-tables. The other three are graded PASS and awaiting his look. |
| **parked on RFC_036 (2)** | `tool-ab-test-calculator` · `tool-meme-generator` — do NOT unblock by deactivating their library templates; both have live forks elsewhere. |
| failed once | `tool-ab-test-calculator` (#2) — the fleet-wide unique index. Cause understood, designed out of the recipe. |
| remaining | **~55** |

Three of the five replaced a tool that was **measurably broken in production** (two had their
comment-stripping call swallowed by its own comment; the minifier also corrupted `<pre>`/`<script>`
content). Reading the live `<script>` before writing the brief is what found all three.

## Next actions, in order

1. **Pick the next tool from the 55** with the RUNBOOK's "Scope the batch correctly" query
   (`p.name LIKE 'tool-%'`, NOT `p.url LIKE '/tools/%'`), smallest ported body first, `ext_scripts=0`
   before tackling the 13 external-script ones.
2. **Run the six-step recipe above. Do not file a rebuild you cannot attend** — see the retire race.
3. **The owner reviews served pages**, one at a time by his 2026-08-17 ruling (he relaxed it once, to
   review three together, on request).
4. **RFC_036 stays open.** Owner direction: keep the library-and-fork model, other sites fork ⇒
   option 2 (a rebuild records `forked_from`), NOT option 1. Nobody has built it.
5. Rich apps last and one at a time (owner ruling 2026-08-16 put them in scope as reimplementations).

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
