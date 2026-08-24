# NOTES — `bugs_open/381` inexpressive composition

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-08-24 — session opens, ownership checked first

`scripts/who-owns.py 381` → OWNED-or-active by `loanzy_uk_example_site`, but that lane's own
HANDOFF 24 §0b says **"UNOWNED"** and ranks 381 behind 376. `ListAgents` showed a live
`loanzy.uk` session; messaged it, and it confirmed it is the filing lane and expects someone else
to take the fix. **This lane owns the fix; that lane owns the account of the site.** The
who-owns verdict and the file's own text disagreed, and the file won — worth remembering that
`who-owns.py` reads commits and cannot read a bug file's status line.

Working tree checked for collisions: 93 dirty paths, **none** touching the planner, the writer,
`content_components`, or `plan_sections`. Re-checked before writing.

## 2026-08-24 — the bug re-validated, still live

22 sections on `garden-tools.uk`; not one built from a list/table-capable template; not one
rendering a `<ul>` or `<strong>`. Fleet-wide `[MEASURED 2026-08-24, 30d]`: 741 pages / 29 sites,
**327 (44%) with no list, table or `<strong>` anywhere**, and **1,863 of 1,980 section placements
(94%) used a prose-only template**. The bug is not site-specific and not stale.

## 2026-08-24 — ⚠ MISSTEP 1, mine, caught by a control: a PostgreSQL `\b` is BACKSPACE

First structure baseline used `~* '<(ul|ol)\b'` and returned **0 for every slot, including
`article-body`**. I nearly wrote that up as "no pass-through slot has ever carried a list", which
would have been a dramatic and completely false finding. In PostgreSQL regex `\b` is a backspace
character; the word-boundary operators are `\y`/`\m`/`\M`. Re-run with `'<(ul|ol)[\s>]'`:
`article-body` **116/153**. **What caught it: the number being 0 EVERYWHERE, including a slot I
had just read HTML out of by hand.** A uniform zero across a heterogeneous population is an
instrument reading, not a finding. Logged to `WRONG_CALLS.md`. Cheap check that would have caught
it first time: a positive control row that MUST match.

Also lost ~4 minutes to two unguarded `jsonb_each` sweeps over `content_components` that hung and
killed the exec at the 2-minute tool timeout. `SET statement_timeout` in every heredoc since.

## 2026-08-24 — ⚠ MISSTEP 2, mine, and it changed the fix: I had the lever wrong

My first plan said: the writer emits only `<p>` because `generic-text-block.content` is
`type: text` and RULE 9 forbids HTML in text fields; therefore **retype to `html`** and the writer
is freed. I wrote that into the plan and told four other lanes.

**It is wrong, and the disproof was already in the data I had.** `article-body.content` is ALSO
`type: text`, under the same RULE 9, rendered by the same engine — and it carries a list in
**76%** of instances against `generic-text-block`'s **7%**. The difference between them is not
the type. It is that `article-body.content` has an `llm_guidance` string saying *"Use h2 for main
sections, h3 for subsections, p for paragraphs, ul/ol for lists, blockquote for callouts"*, and
`generic-text-block.content` **has no guidance at all**.

So **`llm_guidance` is the lever and RULE 9 does not bind.** The retype still earns a place, but a
smaller one: the prompt prints the field as `` `content` (html, required) ``, so the type is the
**routing key** between RULE 9 and RULE 10 — retyping stops RULE 9 contradicting the guidance. It
is honesty, not force.

**What caught it:** reading `article-body`'s `input_schema` because I wanted to know whether the
blog writer had a conflicting rule before including it in the retype — i.e. a scope question, not
a hypothesis test. **The cheap check I skipped: before claiming a mechanism explains a difference,
find the case that has the SAME mechanism and the OPPOSITE outcome.** Two components, same type,
11-fold difference, sitting in the same table the whole time. Logged to `WRONG_CALLS.md`.

The filing lane independently flagged the same wrinkle from the other direction — that the
`text`-typed field *already* contains `<p>` today, so the type is evidently unenforced. Both
observations land on the same correction.

## 2026-08-24 — what the type actually does, traced rather than assumed

Grepped every Go branch on a field's declared type. Only two exist and neither is the component
field type: `ai_actions.go:1299` is the STEP's `output_format`, and the envelope guards key on the
LLM step's `{"result":…,"type":"text"}` output envelope, not on a schema field. The only real
reader is `DeclaredTypeSatisfied` — **default-TRUE, only `array`/`list` checked**
(`content_type_violations.go:262`). The `staged-component-build` lane flagged the consequence
before I could hit it: a typo'd type string behaves identically to a correct one, forever, with
nothing able to surface it. Hence the post-apply read-back asserting the literal `'html'`.

## 2026-08-24 — `content_shape` is a dead column with a live intent

Designed as exactly this axis (`sql_for_tables/005_content_components.sql:9118`, COMMENT names
`prose, structured_list, structured_card, key_value_pairs`), and: zero Go readers, omitted from
the birth INSERT (`store_generated_component_action.go:634`), NULL on 128/151, drifted to free
text (`series`, `sequence`, `mixed`), and **12 rows marked `structured_list` whose templates
contain no list markup**. This is the argument for deriving capability rather than declaring it,
and it answers TLIB-016's open `verify-later` ("does the selector Go code read these columns?" —
it does not).

## 2026-08-24 — the "34 sat available" premise, corrected with the filing lane

Enumerating the structural components rather than counting them: directories, trackers,
calculators, quizzes, spec sheets, one `pricing`, two carousels, and `site-footer` (chrome). **No
generic checklist, steps, comparison-table or calendar component exists.** So the bug's fix
candidate 1 is necessary but not sufficient — on its own it converts a blind choice into an
informed refusal. The filing lane verified this here, retracted the premise in the bug file, and
named the general error: **counting CAPABILITY without checking SUITABILITY.** `[MEASURED
2026-08-24]`

## 2026-08-24 — volume is on the gap planner, not the greenfield planner

`llm_call_log` 30d: `content-gap-planner` **749**, `build-site-planner` **27**, `site-planner`
**5**. The bug was found on a greenfield build; the fix's volume lands elsewhere. `site-planner`'s
menu query takes **no site parameter**, so the evidence-base row gate cannot be expressed there —
recorded as a known asymmetry rather than papered over with an invented param.

## 2026-08-24 — coordination, and one defect found in someone else's code by asking

Seven lanes briefed before writing a line of SQL. Five replied.

- **`bugs_open/380`** — numbers split 591–595 / 596–599. Both lanes ship anchored `replace()` with
  exact-count prechecks, never a base64 whole-prompt rewrite (which would silently revert whoever
  landed first). Their anchors: `No verified facts are registered…` ×2, rule 17,
  `{{if .site_specs.specs.evidence_base.writer_block}}`. Mine: the `load_components` query value,
  the listing line, a rule appended after rule 18, the rule-10 literal. Disjoint.
- **`news editorial`** — plain vocabulary only, and a hard request adopted: the guidance must
  **explicitly forbid `<img>`/`<figure>`/`<iframe>`**, because in-blob imagery is the loss class
  `inline_guide_imagery`/035 exist to retire and this change would be the enabling edit. Also: no
  furniture class names in guidance, because under 035 a pull-quote becomes a child component
  instance and a class name in guidance is a comment, not a control. They then MEASURED
  `data_sources` for me: empty on both `evidence-chart` and `evidence-timeseries`, one active
  component uses it fleet-wide — so the `requires-evidence-base` tag is the ONLY mechanism, not
  belt-and-braces.
- **`bugs_open/283`** — zero of their files touch `input_schema`; four `UPDATE … SET html_template`
  and no `content_components` INSERTs, so my id-scoped pre-state guard cannot be disturbed.
  Their caution, adopted: an `{{.InstanceID}}` token is a **binding, not content** — 140 of 297
  active templates carry it — so a template token must not read as a field.
- **`bugs_open/378`** — corrected 107 in my favour: the incumbency trap they warned about is a
  property of Path 2 (the `section_type` scored contest); this fix is Path 1, resolved by name,
  and never reaches that score. They added my kind-census to 107 as the shopping list it has
  never had.
- **`305 negation gate`** — ⚠ **asking a question found a real defect in shipped code.** I asked
  whether their sentence scanner splits on tag boundaries, because `html` slots would start
  producing `<li>` and `<h3>`. It does — on `</p`, `<br`, `</li`, `</h`, `</div`, `</td` — but
  **`</th` was missing and is NOT covered by the `</h` arm** (`</th` differs at the third
  character). So a define-by-negation construction in a table HEADER cell produced a
  markup-bearing "sentence", and their repair splices over exactly that span — it would have
  replaced `</th><th>` with prose and broken the table. Fixed by that lane in `714789d7b`,
  mutation-proven, **Go, inert until the next roll.** **Timing consequence here: prefer landing
  the writer half after that roll.**

## 2026-08-24 — `report-dossier` dropped from the retype set

Its `body` field's own guidance reads *"Pre-rendered dossier HTML from create_report_page. Never
authored by an LLM and never assembled from a template."* It is `source: llm` in the schema and
is not written by an LLM — so it looked like a pass-through prose slot to my census and is not
one. **Four target fields, not five.** The census predicate (an `llm` text field rendered directly
inside a block container) cannot see that distinction; only the guidance text could.

## 2026-08-24 — council APPROVED round 1, and the sharpest objection was a real gap in my verification

`ca400ba6`, approved with 3 advisory objections, none high-severity. Acted on all of them rather
than banking the approval — notes on the two worth remembering.

**`bug_historian` (medium) asked the question I had not.** I had verified the renderer
(`text/template`, no escaping) and the type checker (`DeclaredTypeSatisfied`, default-TRUE) and
concluded the writer could safely emit `<ul>`. **I never asked what the slot's CONTAINER was.** If
any of the four templates wrapped `{{.content}}` in a literal `<p>`, then `<h3>`/`<ul>`/`<table>`
would nest block elements inside a paragraph — invalid HTML, repaired differently by each browser,
and **invisible at every point we look**: the DB row stays schema-valid, no check fails, only the
served page is wrong. That is 016b §9's most-repeated family (the `<style>`-in-the-wrong-half case).
**CLEARED `[MEASURED 2026-08-24]`: all four use a `<div>`.** But it was cleared by the reviewer's
question, not by my own checking, and the transferable lesson is narrow and useful: **I verified
every layer that would ERROR and none of the one that would silently malform.** A "can the writer
emit this?" question has two halves — is it permitted, and will its container accept it — and I
had only done the first.
⚠ Note `about-content` contains `<p>` elsewhere in its template, so the predicate must test the
**container of the slot**, not the presence of `<p>` anywhere. A naive `html_template LIKE '%<p%'`
would have produced a false positive and sent me chasing a non-problem.

**`prior_art_librarian` (medium) was right that I had argued the wrong thing.** I had written at
length that `content_shape` is dead — measurements, zero readers, the omitted birth INSERT — which
establishes that it does not WORK, not that it should not be REVIVED. Those are different claims
and only the second justifies building something new. The rejection is now argued explicitly in
591's header (a 128-row backfill, correcting 12 actively-wrong rows, a CHECK it never had, a Go
change to the birth INSERT, then hand-maintenance for ever against a column another lane is
rewriting this week). **"The existing thing is broken" is not the same as "the existing thing
should not be fixed", and a reuse objection is asking for the second.**

Also acted on: the `html`-implies-both-list-and-table premise is now stated (the vocabulary means
CAPABILITY, never intent); the function's COMMENT records that its output vocabulary is a shared
contract with no version tag, read by three prompts at once; and why the four guidance blocks are
not one shared string (each carries the clause that prevents its own component's defect —
`about-content` must not duplicate `highlights`, `illustrated-text-block` must not bypass its own
image fields, `article-body` keeps `<h2>` because it is a whole article body).

Guardian's two checks both came back clean and are recorded in 594's header: **no Go path branches
on a field type of `html`** (the only readers of a declared type are `DeclaredTypeSatisfied` and
`component_schema_fields.go`, which copies it into the writer's field spec — the routing effect
intended), and all four target agents have exactly one active row.

## 2026-08-24 — three coordination outcomes worth keeping

**A question found a defect in someone else's shipped code.** Asking the `305` lane whether their
sentence scanner splits on tag boundaries — because `html` slots would start producing `<li>` —
surfaced that `</th` was missing from their boundary list while `</td` was present. A
define-by-negation construction in a table HEADER cell produced a markup-bearing "sentence", and
their repair splices over exactly that span, so it would have replaced the cell tags with prose and
broken the table. Fixed by them in `714789d7b`, mutation-proven. **Nobody had a symptom.** The
general form: when your change alters the SHAPE of what a downstream consumer sees, ask that
consumer what it assumes about the shape — the answer is cheap and it is occasionally a bug.

**An ID collision I created and did not notice.** My register entry was `CQ-028`, which the 277
lane already holds; the `367` lane spotted it, deliberately did not renumber it (my entry named
sources it could not see), and left a flag. Renumbered to `CQ-031` here and in every
cross-reference, and the now-stale flag replaced with a one-line trace — **a resolved warning left
in place is the "stale status line makes the correct action look premature" shape.**

**And a thing I could not settle, recorded as unsettled.** The `editorial` lane asked whether my
lane had written the `requires-evidence-base` tags. It had not — my migrations only READ the tag,
and nothing of mine has touched the live DB outside a rolled-back transaction. But in checking I
found something more useful than the answer: **`content_components` has NO `updated_at` trigger**
(the only trigger is `trg_cc_refuse_null_section_type`), so `updated_at` moves only when a writer
sets it explicitly, and there is no schema history for the table. **"Who changed this component
config, and when" is unanswerable in general**, not just for those two rows. That lane has since
filed it as a landmine with the trigger census credited. I offered a hypothesis about their
measurement being wrong; **they refuted it with better evidence than I had** (a raw column print,
no predicate involved) and found the likelier explanation themselves. Recorded because my
hypothesis was the confident one and it was wrong.

## 2026-08-24 — arm A APPLIED and live; arm B held mechanically, not by a note

**Applied 591/592/593 by hand and recorded them.** Not via `run-migrations.sh --apply`: the runner
has **no directory or file scope**, and `600_claims_audit_rotation.sql` was pending from another
lane at the time, so `--apply` would have shipped someone else's change under my action. Recorded
afterwards with `--record-only`, which is the documented path for an out-of-band apply.
⚠ The runner's default dry run re-executes every pending file inside a doomed transaction and
outran a 300s timeout — read the ledger (`schema_migrations`) and diff it against the non-sidecar
files on disk instead, or pass `--no-probe`.

**The check that mattered, and it nearly wasn't done.** I had verified the evidence-base gate was
*present* in the live query text. That is not a check — it cannot distinguish a working gate from
one that filters everything, or nothing. So I stripped the clause out of the live query text and
ran both versions against two sites:

| site | with gate | without gate | excluded |
|---|---|---|---|
| evidence-LESS (`garden-tools.uk`) | 149 | 151 | **exactly the 2 tagged components** ✓ |
| evidence-BEARING | 151 | 151 | **nothing** ✓ |

Both directions. `[VERIFIED 2026-08-24]` **A gate tested only on the site it is supposed to filter
proves nothing** — this is the same shape as the positive-control lesson from this morning's `\b`
misstep, arriving a second time in one session in a different costume.

Also verified rather than assumed: each menu query **executes** bound as the chassis binds it
(149/149/151 rows). A query that parses but fails at runtime would kill the whole planner step, and
nothing in the migration's own verify block could have caught that — the migration only checks the
query's *text*.

**Arm B is held by RENAME, not by a note.** `594`/`595` are now `*_HOLD.sql` so `SIDECAR_RE`
excludes them from the runner. The reasoning is the point: **a documented "do not apply yet" does
not survive another session's `--apply`**, and this tree has many sessions. The release condition
(the 305 lane's `714789d7b` live in the chassis) is in both headers, and I have asked that lane to
ping me when it rolls.

**⚠ And I could not confirm that fix's status — both obvious methods failed, which is worth
recording because they are the methods anyone would reach for.**
- `grep -a 714789d7b /proc/1/exe` on the chassis pod: **useless here.** The 40-zeros control came
  back **PRESENT** (Go's internal digit table matches it), so the probe cannot discriminate — the
  existing LANDMINES entry warns about exactly this and I walked into it anyway before reading my
  own control.
- Pod `.status.startTime`: dates the **ROLL, not the IMAGE.** Ours started 15:39:22Z and 15:39:53Z,
  about an hour *after* the fix was committed at 14:39:30Z — which feels like evidence and is not.
  A tag can be built before a commit and rolled after it.
- The honest check is a freshly-started pod's `build provenance` line plus
  `git merge-base --is-ancestor`. Ours had already scrolled past `--since-time`.

**Not weakened instead of held.** Dropping the `<table>` clause from the guidance would have removed
the dependency entirely. I did not, because 0 content tables across 741 fleet pages is part of what
this fix exists to restore, the council approved the vocabulary as written, and waiting costs
nothing — the defect has been live for months and nothing degrades while two files sit held. That
reasoning is in the header so a later reader does not "simplify" the hold away.

**What is live is honest about itself:** `generic-text-block` still reads `[prose only]` in the
planner menu, because arm B has not landed. That is correct, not a bug — and it is why both the
loanzy lane and this file now warn that a build in this window measures arm A alone (which
components get CHOSEN), not markup volume.

## 2026-08-24 — ⚠ CORRECTED by the 305 lane: my reason for the hold was wrong, and the obvious way to verify my own change is a FALSE PASS

Two corrections from that lane after my release, both revising premises **I** supplied.

**(1) `594`/`595` did NOT make the `</th` defect reachable.** I said they were "what first let the
writer emit a `<table>` into these slots". `[MEASURED 2026-08-24 by that lane]` `<th>` markup has
been reaching `page_components` since **10 August** (2 / 14 / 1 by week, **17 total**), and an
inspected instance reads `rendered_has_real_table=true`, `escaped=false`. **The render path was
already passing markup through at `type: text`.** My change raises the rate; it did not open the
path. Their fix was closing a fortnight-old *live* defect, not a prospective one — and the
fortnight did zero damage, because 0 of 76 `<th>` cells fleet-wide have ever carried a
define-by-negation construction (header cells are labels, not prose).
**How I got it wrong: I reasoned from my own change's intent instead of measuring the population.**
I knew `generic-text-block` showed 1 table in 173 instances — I had that number in my own baseline
— and read it as noise rather than as evidence the path was already open. The lane then took my
framing at face value, so my unmeasured premise propagated into *their* reasoning too. **An
unmeasured premise handed to a peer comes back as agreement, not as a check** — the same shape
`WRONG_CALLS.md` records from the vetcomparison/agritec pair.

**(2) The one that would have cost something: verifying this change at `rendered_html` is a check
that cannot fail.** The obvious next verification — *"look at a page and confirm a table appears"* —
would have shown a table **either way**, because that slot rendered real tables before the retype.
A clean pass, proving nothing, exactly like this morning's 40-zeros control and this morning's `\b`
that matched nothing. **Third instance of the same disease in one session, in a third costume.**

**The demand control that DOES discriminate** — and, luckily, the one I actually ran: the planner's
capability distribution moving **96/2 → 93/6** (`component_expresses` returning
`{html-block,list,table}` where it returned `{}`), plus each field's declared type and
`llm_guidance` read back literally. That is a value that could only change *because of this
migration*. Recorded in both migration headers and in the RUNBOOK so the next reader does not
"improve" the verification by going to the served page.

**What still stands:** that a hold must be ENFORCED (`_HOLD.sql`) rather than documented, because
another session's `--apply` has no file scope. Only my claim about what the change made reachable
was wrong — the hold cost an hour and was harmless.

**Lane status:** `bugs_open/305` is now CLOSED → `bugs_closed/305_HANDOFF_2026-08-18_v2_voice_does_not_suppress_define_by_negation.md`;
their closing handoff is `docs/agent_docs/docs024_key_docs_latest/bugfix_305_negation_gate/HANDOFF_2026-08-24b_continue_here.md`,
and it carries the `</th` boundary set plus the "ping me if RULE 10 widens" note, so that survives
their session.
