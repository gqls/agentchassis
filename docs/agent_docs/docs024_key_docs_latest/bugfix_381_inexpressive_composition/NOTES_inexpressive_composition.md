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
