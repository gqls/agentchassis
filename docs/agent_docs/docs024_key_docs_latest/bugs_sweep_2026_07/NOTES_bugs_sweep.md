# NOTES — bugs sweep (append-only, newest at the bottom)

The technical log. **The missteps are the point**, not an appendix.

## 2026-07-27 — session 1

**Triage.** Ran `who-owns.py` over all 42 open bugs. Its VERDICT line said "OWNED or
recently active" for **42 of 42** — saturated and useless on a tree doing ~1,500
commits/week. The `likely OWNING workstream(s)` section discriminates: 18 unowned.
*My first extraction of that section was wrong in a way that silently reported "no owner"
for every bug* — the header is followed by a blank line when a bug IS owned, so a
`sed '/likely OWNING/,/^$/p'` range prints nothing. Read raw output before trusting a parser.

**080 (gap-planner canonicalisation).** Routed `applyNewPage` through `CanonicalisePage`.
Council caught that my submission said "the fourth surface" when my own grep output said
five. *I had run the command that answered the count and then wrote the number from the
bug file's prose.*

**095 (empty assembly reports COMPLETED).** Measured the blast radius before choosing the
rule: a blanket "fail when sections are planned" would have broken 17 pages, 5 legitimately
never built. Narrowed to "component rows exist and none contributed" — 0 instances, so no
fleet breakage. *Then the figure expired: 0 at 18:05, 1 at 18:35 (oufe.com/
tool-recovery-waterfall), while the council was still reading the submission that rested
on it.* A council seat objected to something else entirely and re-running the query to
answer them is what surfaced it. A wrong objection bought a real correction.

**109 (render-context allowlists).** Made `renderCtxToMap` derive from struct json tags.
The decisive check was that `EvidenceBase`… no — that `RenderContext` is **marshalled
nowhere**, which is what makes promoting inert tags to the authoritative declaration free.
Council asked for three checks I had not run (marshal absence verified wider, no existing
tag-to-map helper, exactly one production caller); all three held.

**103 (build brief as meta description).** Found a **second call site the bug file does not
name** by grepping every writer of the column instead of trusting the filed one. Also the
live count was 17, not the filed 15 — the census threshold was 400 chars; at 320 two more
genuine briefs appear.

**091 (reports a write that did not happen).** Same discipline found **two more sites**
with the identical shape. Checked before changing: no `conditional` step in any active
agent definition branches on the field.

**MISSTEP — I rolled the chassis on top of my own in-flight council run.** It died at
19:22:29; the pod I replaced went down at 19:22:02. I then reported it as "running" four
times across 70 minutes, because `current_step` and `status` say "running" forever on a
dead run. Only `updated_at` distinguishes working from wedged.

**MISSTEP — inside the write-up of that, I misidentified a second "collateral" run.** It
had advanced since and was not even mine; I had matched on a step name that sounded right.
Same error already logged that day as "I measured the table whose NAME matched the
concept". Corrected in place within 20 minutes.

## 2026-07-28 — session 1 continued

**MISSTEP — I called a settled convention "an owner decision".** The news URL shape is
fixed by `page_canonical.go`'s section-index family, shipped and council-approved, with
`relojistas.com` as the live worked example. I had **read and quoted that file's header as
evidence for my own fix** and still did not notice it answered the question I was calling
open. The answer was inside the evidence I had already collected.

**103 completed.** House voice landed 19:35, so the owner's hold turned out to change the
work rather than delay it — my staged copy broke the voice rules twice. Rewrote, then
backfilled 17 rows and re-rendered 17 pages. **The read-only STEP 1 earned its place:**
`pages.title` carries a nav suffix, so 11 of 17 would have published
"…ROI Estimator **| Tools**, free to run…".

**081 — stopped building.** Candidate 2's discriminator cannot be written: the structural
signal (`sections @> ["news-listing"]`) returns 4 rows fleet-wide and one is the *catalog*
index, byte-identical in shape to a real news page. Recorded the falsification rather than
shipping a predicate that would break a live page.

**094.** Council's sharpest objection was one I had only flagged: `page_id` arrives via
`ExtractActionInputs` Strategy 2's "aggressive recursive search", so a stale one could
resolve a different page of the *same* site, where before a missing `page_name` failed
loudly. Added resolution logging — does not prevent it, makes it attributable. Then drove
the branch end to end on v1.0.1182 and confirmed `page_name` resolved from `page_id`.

**097 diagnosis — three attempts.** #1 died in my own roll. #2 FAILED at `spawn_diagnoser`
(a live `bugs_open/029` instance, pod idle from birth for 63 min). #3 ran all five
iterations and returned **UNVERIFIABLE**, defeated by `bugs_open/108`: the code index
answered "0 rows … this is not an unanswered question" for `RepairPageLinks`, a symbol that
exists and is on the path under investigation. The index holds one commit, 970 behind HEAD.

**Open thread I could not settle:** `rerender_page_sections` returned `escalated: true`
while filing no work item and mutating nothing. Marked `[UNVERIFIED]` — plausible for a
tool-widget section, but it is the `bugs_open/091` shape.
