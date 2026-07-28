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

## 2026-07-28 — session 2 ("bug sweep continuation 2")

**Target selection.** The handoff's item 1 (108) is now OWNED — the architecture-review
thread has eight commits on it and names "108 candidate 1" as its own next job, so the
sweep must not take it; 097 stays blocked behind it. 091 candidate 1 belongs to
`work_item_completion_integrity`. Took **127** instead: fleet-wide (5 sites), confirmed
unowned (`who-owns` `likely OWNING` section prints `(none identified)`; only the filing
commit touches it), and webdesign.co.uk's News page is blocked on it.

**127 fixed (candidates 1+3, `723a10259`).** The filed mechanism held exactly at HEAD.
Two things the bug file did not know:

- **Firecrawl (the live primary) returns news dates as relative text** ("3 months ago"),
  verified against docs.firecrawl.dev — while `WriteFeedItemsAction` parses RFC3339 only.
  Without a date normaliser at the provider boundary the fix would pass every unit test
  and still fail the bug's own acceptance check (`source_published_at` populated). The
  acceptance check was the thing that surfaced this — reading the verifier before writing
  the fix paid for itself again.
- **The downstream age filter is already written and dead**: `WriteFeedItemsAction` skips
  items >30 days old or >1 day future, which today never fires because `published_at`
  always arrives empty. Once dates flow, feed item counts will legitimately DROP — do not
  read that as a regression post-roll.

Checked before building: all three providers registered available in the running adapter
pod (so DuckDuckGo declining news is safe — both keyed providers precede it), and
external API params verified by fetching both providers' docs rather than from memory
(ScrapingBee `search_type=news`/`news_results`; Firecrawl `sources`/`tbs`/`data.news`).
`[UNVERIFIED]` residual: ScrapingBee's news field names are documented-not-witnessed — no
live call with a real key was made; failure mode is loud (zero results → fallback), not
mislabelling.

Tests: 14 across three new files, green against a `git archive HEAD` overlay; the full
actions suite is green there too (the 07-21 `discovery_checks` RED is gone). Council
submission `a7ae8ce8-ef40-4503-be8a-972ebe1b0973`, verdict pending at commit time —
committed without trailer per standing practice.

**127 APPROVED round 1 and adapter ROLLED (v1.0.1185).** Pod-grep 1/1/0/2 on the four
markers; smoke tests witnessed Firecrawl serving a real `sources:["news"]`+`tbs` request
(3 results, 1.7s) and the DDG decline→fall-through path. The response-produce error in
the smoke test is the header-less hand-sent message hitting the non-existent default
reply topic — production routes via `reply_to_topic` headers; not a defect. Production
check pending on the ~13:50 `content-feed-refresh` (watcher armed). Chassis `time_range`
rides the next chassis roll, deliberately not forced.

**097: re-checked before taking it — another thread ("bugs thread 2") had re-fired the
diagnosis at ~11:45** the moment 108 closed, and it returned **CONFIRMED in ~12 min**:
the 07-25 learning-center rebuild ran through the BULK rerender path
(`rerenderSinglePage`), which deploys independently of the `validate_page_content` gate
where `RepairPageLinks` lives — the repair was never on that path. Their lane; their
working tree now shows the rerender/validate files mid-edit. Left alone.

**109 completed in code (`f78cf8125`, council corr `1d082754` pending).** The remaining
three maps now derive from the struct declaration; two divergence closures stated and
transcription-tested (fallback renderer painted DEFAULT colours + empty metadata for
restored contexts — twelve fields with `current_page`'s exact latent gap; `contextToMap`
lacked `logo_url`). Per-page trio stays excluded. Full suite green on a HEAD overlay.
Inert until a chassis roll.

---

## 2026-07-28, second session ("bugs thread 2") — 108 defect A: freshness by commit identity

Took `bugs_open/108` per the handoff's next-actions order. Ownership check first:
`who-owns` names `architecture_review` (36 commits/14d), whose HANDOFF_2026-07-28 §3
item 1 is exactly this fix and whose design direction was already settled — store the
ref the indexer fetched, key the verdict on the commit rather than the row clock. This
session implements THEIR design under the bugs-sweep remit; contribution goes into the
bug file, not a competing lane.

**Two things the bug file did not know, found before writing code:**

- **The "no ref parameter to point" correction is itself superseded.** The
  `code-index-refresh` task's `input_data` really is `{repo, owner, language}` — but its
  `pre_query` (a column the bug file never read) derives `ref` from the most recent
  orchestration whose `input_data.ref` matches `^[0-9]{3}_`, so the cadence dispatch DID
  carry `ref=086_experience_loop` on both recent runs (orch `59f41f80`, `a9b48ceb`,
  verified in `collected_data`). The ref reaches the indexer today and is discarded at
  the upsert — which is why the fix is "store it", not "plumb it".
- **`analyse_repo_local` pins its fetch to the index's own commit by DEFAULT**
  (`pin_to_index_commit` default true). Had that been live on the indexer lane, even a
  push could never advance the index — checked the live `agent_definitions` row before
  assuming: the code-indexer sets `pin_to_index_commit: false` explicitly ("the indexer
  DEFINES the index commit"). No third defect; recorded here so nobody re-walks the scare.

**Built (candidate 1 + the banner half of 4):** `GitHubSource.CommitInfo` (one read-only
GET `/commits/{sha}` in the spawned pod — resolves full sha + committer date);
`analyse_repo_local` exports `commit_time` best-effort (failure → absent → NULL → banner
UNKNOWN, never an invented date); `index_code_symbols` stores `ref` + `commit_time`
(plain assignment in the upsert, same doctrine as `body`); `freshnessBanner` verdict now
keys STALE on COMMIT age (`codeIndexCommitStaleAfter`, 48h) and renders NULL commit_time
as loud-UNKNOWN — FRESH is unprovable without evidence about the commit. The
missed-refresh branch on `updated_at` survives as pipeline-liveness. Both lanes' call
sites updated; banner states the pushed-tip mirror fact in every branch. Migration 250
applied (columns nullable, all 4,535 rows NULL until the first post-roll reindex — the
VERIFY script's check 4 encodes the induced fault as SQL). Unit tests: the exact live
failure (refreshed 2h ago, commit 4d old) is now the first test case and must render
STALE; green against `git archive HEAD` + overlay.

**Deliberately NOT done:** pushing the branch (the only thing that makes the index
CURRENT — owner call, stated in the bug file); prompt-side edits to the CODE INDEX
LIMITS paragraph (config-side, owning workstream's call, and it errs safe as written).

Council submission `b5285973-9038-47a1-b5d9-4a9696fb1eb3`.

**108 CLOSED (both defects fixed AND live) — moved to `bugs_closed/`.** Council
`b5285973` APPROVED round 1 (9 reviewers, 0 unreadable; 3 advisory objections, each
answered with evidence before commit — the exhaustive-caller grep, the reuse search for
an existing commits-endpoint helper, and the token-presence-by-runtime-witness). Commit
`87d0bcf97`; mig 250 applied; live on v1.0.1184; VERIFY green 4,992/4,992;
`internal/tools-api` 0→29, `RepairPageLinks` 0→1. Full trail in the case file.

Three things this close surfaced, priced for the next session:
- **The branch got PUSHED mid-session** (origin at `d98010e8b`, 19 behind) — the
  MEMORY.md "955/1,003 commits behind" warning is obsolete; the index now genuinely
  describes today's tree AND says what it describes.
- **`bugs_open/129` filed** (spawned child adopts the parent's orchestration row and
  silently declines — child-side logs; 2-of-3 failure rate on the index lane today;
  diagnosis `dcde1ed9` dispatched). Cost this close ~an hour of retries.
- **`TRIGGER_code_indexer_v2.sh` was committed non-executable** — fixed (`d59852fb7`).
  A `./`-invocation died on Permission denied and the failure was invisible through a
  grep filter that matched neither that phrase nor anything: when a wrapper filters a
  script's output, put an unfiltered `tail` fallback in, or check `$?` explicitly.

NEXT per the handoff order: **097** — its blocker is gone twice over (honest banner +
RepairPageLinks actually indexed). Re-fire its diagnosis with the unchanged question:
which of the three mechanisms was on the 07-25 build path.
