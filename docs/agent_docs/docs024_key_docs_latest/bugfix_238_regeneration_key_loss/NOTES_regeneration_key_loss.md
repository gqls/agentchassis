# NOTES — bugfix 238

Append-only, newest at the bottom. The missteps are the point, not an appendix.

## 2026-08-10/11 — session 1 (the whole lane, so far)

### Picking it up

`who-owns.py 238` returned **OWNED or recently active** — the finetuning lane
filed it and has commits within 24h. That is a lagging signal by construction, so
I read the owning lane's handoff instead of trusting the verdict either way, and
its §8.4 says in terms: *"`bugs_open/238` is the whole of the remaining imagery
work … Fix candidate 1 there … is the durable answer"*, addressed to a fresh
thread. Not competing. Also checked `git status --short` for in-flight Go work in
the bug's code paths (the LANDMINE about commit-blind ownership checks) — clean.

### The finding that changed the whole shape of the fix

The bug file says the generator "reproduces the ones that look like copy and
drops the ones that look like plumbing". **That is wrong**, and I nearly built
the wrong fix on it — a "preserve structural keys through the LLM round-trip"
design, which would have been machinery for a journey those keys never make.

The 11 lost keys are exactly the component's **non-`llm`-sourced** fields. The
LLM is *forbidden* from emitting them ("exactly the keys listed"). What actually
happened: the declared sources resolved nothing, `on_missing` defaulted to
`skip_field`, and `save_page_sections` replaced the row wholesale.

The tell I should have reached for first, and did not: **the lost set was
suspiciously exact.** Not "some images", not "the ones the model forgot" — every
key ending `_url` and nothing else. A model dropping things it finds boring does
not produce a clean partition of the schema by a field nobody showed it.

### The measurement that made it certain

`page_component_history` still holds the last good row (58 keys, the snapshot
`save_page_sections` took *immediately before deleting it*). Diffed against the
live 47, cross-referenced against `input_schema`: the 11-key diff is precisely
and only the non-llm fields, and **zero** llm fields were lost. That is
disconfirmable — a generator that "drops plumbing" would have lost some alt texts
too, or kept some URLs.

### Wider than filed

While verifying the served page I checked for anchors as well as images. The five
"Read case study" links and the section CTA were **gone** — not empty, gone,
because the template gates them with `{{if}}`. Three of the four checks I would
naturally write (count `<img>`, count `src=""`, grep the CSS class) see nothing
wrong. **A gated field fails more quietly than an ungated one**, which is the
opposite of the intuition that gating is the safe pattern.

### Misstep 1 — I wrote SQL against a schema I had truncated

`\d page_component_history` piped through `head -25`. The column list fitted; the
CHECK constraints did not. My manual snapshot INSERT used `op = 'UPDATE'`, which
`pch_op_check` rejects (`'overwrite' | 'delete' | NULL`), and the transaction
aborted. Cost: one round trip. **`\d` piped through `head` truncates exactly the
block you are about to write SQL against** — the constraints print last.

### Misstep 2 — the FK named the opposite of what its name implies

Second attempt failed too: `page_component_history.component_id` is a FK to
**`page_components(id)`** — the row — not to `content_components(id)`, even
though `page_components.component_id` means exactly that. And it is
`ON DELETE SET NULL`, and `save_page_sections` archives a row then deletes it, so
**every historic row reads NULL** — which is why the column looked unused and
tempted me to fill it with the wrong id in the first place.

Both are now a LANDMINES entry. Both were caught by the transaction rather than
by me, which is the argument for `BEGIN` + `DO/RAISE` over a bare `UPDATE`.

### Misstep 3 — a test that passed twice and was a coin flip

`plan_sections_structural_carry_test.go` passed on its first run and on a
targeted re-run, then failed in the full suite. Not flaky infrastructure: **`planSection`
iterates the schema's fields MAP**, so the order in which sources resolve — and
therefore the order the queries arrive in — is randomised per run. My ordered
sqlmock expectations were a coin flip. Fixed with
`MatchExpectationsInOrder(false)` and a comment saying why.

Two clean runs did not establish stability, and I had already half-believed them.
The general form is in memory as *"2 clean runs cannot establish STABILITY"*; the
specific trigger worth remembering is **production code iterating a map ⇒ any
ordered mock expectation is probabilistic**.

### Misstep 4 — an inherited claim about the result payload

I wrote, in a doc comment, that the new result keys are "absent from the payload
entirely when nothing happened" — copying the shape of the neighbouring
`source_aliases_used` comment. **A nil typed map inside `map[string]interface{}`
is not `nil`**, and a nil map marshals to `null`, not to absence. My own test
caught it before it shipped (it asserted the wrong thing and failed), and I
corrected the comment to say `null`. Small, but it is the "inherited a claim
without re-deriving it" shape that `WRONG_CALLS` 2026-08-10 records one lane over.

### What I did NOT do, and why

- **No `090` diagnosis run.** The standing rule (owner, 2026-07-31) is that a
  `bugs_open/` file asserting a cross-cutting root cause is not filed until it has
  been through the loop *or the session states plainly why it substituted
  equivalent first-hand verification*. This is that statement: I am not filing a
  new cross-cutting claim, I am **correcting an existing file's claim** with the
  primary artefact — `page_component_history`'s own before/after rows, the
  component's `input_schema`, and the four code sites read end to end. The
  disconfirming result was available and did not occur (any llm field in the lost
  set would have refuted me).
- **No sixth guard in `save_page_sections`.** `bugs_open/178` left a standing
  instruction that a further floor there is the trigger for a unified
  content-loss detector as its own submission. The carry is not a floor and is
  not in that function.
- **Did not widen `sectionHasImageField`** — measured benefit zero, and it writes
  live-path data.
- **Did not re-scope `image_url_404`'s site-wide empty-src key**, though it is a
  real defect (finetuning's `blocked` row has held the fleet slot since 08-03).
  The tally is site-wide by construction — its query does not even select the
  page — so per-page keying is a restructure with its own volume question, on a
  check that cannot currently fire. Recorded as a follow-up rather than half-done.

### Repair, and the order it had to go in

Data first (`378`), then a **no-LLM** re-render (`379`, `reason:
section_data_resolved`). A regenerating rebuild is the operation that caused the
bug, and `needs_page` has a poor record on this site (5 failed / 4 wont_fix / 2
rejected against 20 complete). The repair regenerates nothing, so it could run
before the code fix and did.

Applied by hand rather than through `run-migrations.sh --apply`: the runner takes
**every** pending file, and other threads have several. Recorded afterwards with
`--record-only`.

### Council

Two submissions, deliberately separate: `bd38df2e` (prevention) and `98852baa`
(detection + the predicate widenings). Bundling them is the breadth RFC_016
slice 1 was rejected for, and they have genuinely different blast radii — one
adds a resolution source and can only re-supply bytes the page already served;
the other can refuse a live rebuild. Committed with `Council-Submitted:` on both,
per the trailer rule, because holding code for a verdict is not available on a
shared HEAD.

### My own LANDMINES entries were swept into another session's commit

`5c3322aa8` ("correct(168 sweep + 244)") carried my two new LANDMINES entries.
Nothing lost — they are at HEAD — and forward-only holds. Worth recording because
it is the *documented* hazard working exactly as documented: a pathspec commit
protects you from other sessions' files, and cannot protect your own edits inside
a file someone else commits. The 9 lines still uncommitted in that file when I
looked belong to the `diagnosis_schema_visibility` lane; I left them.

### CLAUDE.md changed under me, mid-session

The deploy-verification recipe was rewritten on 2026-08-11 (another lane's work,
`bugs_open/249`): binaries now carry a build-provenance stamp, and `strings` is
explicitly deprecated — it produced three confidently wrong readings in one day.
I had already written `verify-later` lines using the old `strings` recipe into
both register entries and into `380_..._HOLD.sql`. Rewrote all three, and marked
the PBP-039 one as a same-day correction rather than silently editing it, so the
next reader can see the recipe moved rather than assuming I chose badly.
