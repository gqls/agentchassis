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

### Council verdict A — APPROVED, and the objection that corrected me

`bd38df2e`, round 1, 12 reviewers, 8 advisory objections, none high-severity.

**`editquality` caught a real error in my own framing, not in the code.** I had
named ai-agent-orchestration.com `/index.html` as "the honest acceptance case"
for the carry. It cannot be: **the carry reads the CURRENT deployed row, not
`page_component_history`**, and that row is already stripped — so it yields a
`STRUCTURAL_KEY_CARRY_MISS` and no repair. **The fix is prospective only.** It
protects rows that still have their keys and remediates none of the damaged ones,
including the one in the bug's title. Had I left it, a correct null result on that
page would have read as a working fix — the exact shape of a check that cannot
come out false.

Corrected in the register (visibly, struck through) and in the bug file, and the
real population measured: **3 deployed `case-studies-grid` rows still hold
`card1_image_url`** — aao `/enterprise-reference-deployment.html`, finetuning
`/index.html` (repaired), leopardess `/who-we-help.html`. aao is the cleanest
test: 0 spec aspects, 0 current-plan imagery, so its sources cannot resolve and
the carry is the only thing that could preserve those values.

**`debug_historian` asked a question I should have asked myself**: is
`build_status='deployed'` load-bearing on `page_components`, given the fleet's
history of status columns that lie (`site_components.build_status` is `'rendered'`
and never `'deployed'`)? **Measured: deployed 1138, pending 171, approved 23,
removed 1.** It holds. But I had scoped the query on an assumption and shipped it
untested — the cost of being right was zero and the cost of being wrong would have
been a silently inert fix.

**Three seats — `reuse_agent`, `constitution`, `prior_art_librarian` — each
independently asked whether `load_current_section_content_action.go` already
provides this.** I had answered that during exploration (gated on
`mode == "edit_live"`; selects `rendered_html` only, never `content_data`) and
then **left it out of the submission**. Three seats spent effort rediscovering it.
**A deferral or a rejected alternative is only credited if it is stated** — the
answer is now in the register entry, since the code comment does not carry it
either.

**`guardian` raised a failure mode I had not considered**: the carry re-supplies
whatever the stored row holds without checking it is *correct*, so a previously
fabricated value (the `bugs_open/203` phantom `/contact.html` class) becomes
self-perpetuating while its source stays unresolvable. Bounded — the value was
already being served, and live resolution wins the moment the source works — but
genuinely new, and recorded rather than argued away.

**`architecture` returned `needs_rfc`**, on the ground that a new resolution tier
on a two-consumer shared function plus a deliberate cross-path semantic change is
RFC-shaped, and that this compensates for the REPLACE/MERGE split rather than
closing it. **Recorded, not rebutted — a scope objection is not answered by more
measurements.** It did not block, and review here is after-the-fact by design.

Lesson for the next submission: I wrote a long "risks" block and still omitted the
three things reviewers actually asked for — the alternative I had already ruled
out, the scoping value I had assumed, and the limit of what the fix can reach.
**The risks block was full of what I had thought about; the gaps were in what I
had checked and not written down.**

### Council verdict B — REVISE, and the objection was the one I had argued around

`98852baa`, 13 reviewers, gated by one high-severity objection from
`bug_historian`: **I guarded one call site and called the class handled.**

That is `bugs_closed/021` / `bugs_open/093`'s shape, and it is the seat's whole
remit. My submission *named* `RenderTemplateWithMap` as "exempt and named" and
stopped there — which is the move that reads as diligence and isn't one, because
naming one exempt sibling implies the others were checked. They had not been.

**Answered by audit, not argument. The score is 3 guarded of 11.** Unguarded:
`assembleComponents`, `applyContentEdit`, `applyComponentSwap`, `RenderHeader`,
`RenderFooter`, `RenderHead`, the legacy head template, and
`RenderTemplateWithMap`. The list is now in `dead_url_guard.go`'s own header, so
it is read where the guard is read rather than in a verdict artifact nobody opens.

I still did not widen to all eight, and that is a defensible answer only because
it is now a *stated* one with the reasoning attached — the three chrome renderers
have their own response one layer up, the two editor paths edit a row a human
holds, and making `RenderTemplate` itself the reporting form fleet-wide changes
the primitive every render flows through.

**Four seats objected that my censuses were unverifiable by them** — the tables
are outside their schema access. Fair, and the fix is to attach the numbers, not
to insist. All three re-run and attached: 26/12/2/2 by declared type, exactly one
live `render_component` agent, and both squatting `empty-src` rows (finetuning
`blocked` count 16 since 08-03, aao `detected` count 19). **Every one reproduced
exactly.** The lesson is not that the numbers were wrong — it is that a number a
reviewer cannot re-derive is an assertion to them however carefully I measured it.

**`debug_historian` found a real hole in the migration**: it updates by `type`
with no row tie-breaker, and four agent types here carry two active rows of which
only the higher version loads. `page-content-writer` has exactly one today, so it
does not bite — but the HOLD file now refuses rather than half-applying if a
second appears before the flag is lifted. Cheap, and I should have written it
that way first: the landmine exists precisely because "by type" looks total.

**`reuse_agent` asked for a shared emit helper instead of a sibling.** Not done,
with the reasoning recorded: the primitive that carries the semantics
(`insertWorkItem`, and its dedup/two-strike behaviour) is already shared by both
emitters; what differs is the item's shape and its response, which is per-surface.
The `architecture` seat named the right trigger — **a third dead-control emitter
is when this consolidates**, and that is now written where the next author will
be standing.

**Conceded outright:** record-only on the rerender path is worth more than
discarding the finding and less than detection, and I should stop describing it
as "the fleet's only live detection" while the rotations are paused. Nobody is
watching the queue it writes to.

The pattern across BOTH verdicts is the same and worth naming: my risks blocks
were full of what I had *thought about*, and every objection that landed was
about something I had *checked and not written down*, or *assumed and not
checked*. Length is not coverage.

### Council round 2 — REVISE again, and this time it found a real defect

`98852baa` round 2. Gated by `editquality` (HIGH) on the migration's `jsonb_set`
path being **asserted rather than checked** — and with `create_missing := true`,
a wrong path inserts a new branch and reports success, arming nothing while every
reader says "armed". I checked: **the path was right.** But the objection was
about process, not luck, and it was correct — so the file now asserts each path
resolves to a `render_component` step *before* any write, which makes mis-nesting
structurally impossible instead of verified-once-by-hand.

**The find of the whole lane came from `debug_historian`, and it came via my
METHOD rather than my number.** My coverage census was
`default_config::text LIKE '%render_component%'`. The seat pointed out that `_`
is a SQL wildcard — a fair but minor point. Re-measuring properly exposed the
real problem: **that query counts AGENTS (answer: 1), and the question was how
many STEPS. The answer is TWO** — `render_section` *and* `render_from_template`,
both inside page-content-writer.

So **"one config key is full live coverage" was false**, and my migration would
have armed half the render path while the register, the file header and the
coverage report all said "armed". Nothing in the surrounding prose — and there
was a lot of it — could have caught that. **The wrong instrument concealed the
number, and I quoted the number three times.**

Fixed: the migration arms both, and the verify block now **counts** every
`render_component` step and demands they all carry the flag, so a future third
step fails loudly rather than shipping unguarded.

**And I induced the failure to prove the block is not decorative.** Mutated the
file to arm only `render_section` and ran it: it raised
`1 of 2 render_component step(s) armed` and rolled back. A verify block that has
never failed is a claim, not a control.

**Three seats independently said the record-only emit should be gated** —
unconditional DB writes on a shared repair path, while the refusal one file over
was already opt-in. They were right and the inconsistency was inside my own
patch. Now behind `record_dead_url_controls`, default OFF.

**`bug_historian` also checked one of my exemptions and it was false.** I claimed
the chrome renderers "already have a dead-control response one layer up". True
for `render_site_components_action.go`; **false for `rerender_pages_actions.go:532`,
which renders the head template with its own bare `RenderTemplate` and never
routes through that guard.** Corrected in the header rather than dropped.

**Not resubmitting a third time on the remainder.** 8 of 11 call sites stay
unguarded; making `RenderTemplate` itself the reporting form is a change to the
primitive every render flows through, and that is RFC-shaped. The seat's own
phrase — "a documented remainder, not a closed exposure" — is the honest label,
and re-arguing it would be answering a scope judgement with more evidence, which
this repo has already ruled does not work.

**The lane's real lesson, across three verdicts.** Every single objection that
landed was one of two shapes: something I had **checked and not written down**, or
something I had **assumed and not checked**. Not one was about the design. My
risks blocks got longer each round and caught none of them — because they
enumerated what I had considered, and the failures were all in what I had
measured badly or not at all.
