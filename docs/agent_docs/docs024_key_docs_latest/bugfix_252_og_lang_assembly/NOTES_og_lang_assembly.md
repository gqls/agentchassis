# NOTES — bug 252 og/lang assembly

Append-only, newest at the BOTTOM. Evidence, commands, and every misstep.

---

## 2026-08-20 (a) — pickup, and the premise check that changed the design

**Ownership swept before touching anything.** `scripts/who-owns.py 252` returns the ambiguous-number
warning and names the *disk* 252's lane (closed 08-15) — so its "OWNED" verdict is about the other
case. For this slug: last commit on the bug file is `f666408ed` (08-11, the owner's option-3
decision); no open `site_work_items` match og/lang; nothing in the `needs_diagnosis` queue overlaps.
The LMC lane's `HANDOFF_2026-08-15b_continue_here.md` §4.2 has this queued behind "verify 251 is
live", never started.

**251, the stated blocker, is already discharged.** `61abbdbd0` added `preferredPageURL`
(`rerender_single_page_action.go`), council corr `33fb41cb` = APPROVED round 1 with one advisory.
Confirmed live at the artefact: `about.html` serves `<link rel="canonical" href="…/about.html">`.
So A is unblocked and og:url has an existing helper to agree with.

**The premise check that mattered.** The bug file says an assembled page carries og:2. Rather than
inherit that, I counted the stored heads:

```sql
SELECT count(*) AS heads,
  count(*) FILTER (WHERE rendered_html LIKE '%og:url%')   AS og_url,
  count(*) FILTER (WHERE rendered_html LIKE '%property="og:title" content=""%') AS blank_ph
FROM site_components WHERE slot_name='head';
--  24 | 22 | 4
```

22 heads carry og:url. That is the opposite of "the tags are missing", so I went to the artefact:

```
$ curl -s https://ai-agent-orchestration.com/about.html | grep -iE '<html|og:|rel="canonical"'
<html lang="en">
    <meta property="og:title" content="">          <- template, blank
    <meta property="og:description" content="">    <- template, blank
    <meta property="og:type" content="website">
    <meta property="og:site_name" content="AI Agent Orchestration">
  <meta property="og:title" content="AI Agent Orchestration">     <- injectBrandHeadTags, DUPLICATE
  <meta property="og:url" content="https://ai-agent-orchestration.com/">   <- the HOMEPAGE
<link rel="canonical" href="https://ai-agent-orchestration.com/about.html">
```

Two og:title tags, and an og:url that contradicts the canonical on the same page. `git log -S'og:url'`
dates the cause to `d3f73a724` (imagery I1, after the bug was filed). **This is why the bug file's own
fix candidate would have fixed almost nothing** — it proposed adding blank placeholders to fill, and
the sites that have them already have them shadowed by a filled duplicate. Design changed to
remove-then-inject as a result (PLAN D1).

**Two of my own grep errors, same root, both caught within minutes.**

1. I grepped `'lang="en"'` across the Go tree and concluded `rerender_single_page_action.go:670` and
   `rerender_pages_actions.go:527` had already been fixed — both were cited by the bug file and
   neither appeared. **They were there all along**: the emitters build the string in Go, so the
   source reads `lang=\"en\"` (escaped) and a literal `lang="en"` grep cannot see it.
   *The check:* when grepping for HTML that Go EMITS, search the escaped form too
   (`grep -rn 'lang=\\"\|lang="'`), or search a fragment that cannot contain the quote (`<html lang`).
   Cheap and decisive. Logged in `WRONG_CALLS.md`.
2. My first two DB queries failed on column names — `site_components` has `rendered_html` and
   `slot_name`, not `html` / `component_type`. CLAUDE.md says schema first (`\d <table>`); I
   guessed twice, then read it. No damage (a failed query is loud), but it is the same habit as (1).

**A stale premise found in a comment, worth recording for whoever fixes the checkers.**
`discovery_checks/check_site_structural_validity.go` (~:55, :1029) documents at length that og:url
is excluded from `head_essentials_missing` "because the shared `<head>` cannot carry a per-page
value", and that `verify_site.py`'s `OG_PER_PAGE` exemption (`…/loanandmortgagecalculator_couk/verify_site.py:71`)
treats it as an accepted loss. Both premises are what this lane is removing. They will become false
when this ships, and neither is a code path that will fail loudly — added to the close-out list.

**Two facts about the fleet's own machinery that shaped the rollout, both confirmed in source:**
- `renderAndStoreSiteComponent`'s idempotence exit skips any head whose `rendered_html` is non-empty
  and whose `build_status <> 'pending'`. So **a Go change regenerates no stored head, ever.**
- `datahelpers/chrome_render_inputs.go` hashes the template and site_specs **by value**, and Go code
  is not an input. Hence: the og fix must work without a chrome rebuild (it does — it strips at
  assembly), and the lang fix gets its fleet propagation for free from the migrations.

## 2026-08-20 (b) — 090 dispatched, and a planning assumption that expired

`090_TRIGGER_needs_diagnosis_v1.sh` accepted the symptom (mechanism-only, no counts asserted, three
tables/symbols named). Intake corr `855be313-c96a-47d6-a92d-cdfa53de6b03`; the dispatch loop claimed
it and minted **run corr `af31ec22-5662-4798-91b9-b12132ebca70`** — that second one is the key the
artefacts are written under. Filed even though the mechanism is read directly from source, because
the durable claim here is cross-cutting (two producers, a shared artefact, a fingerprint that cannot
see the code) and CLAUDE.md's 2026-07-31 ruling puts the burden on the asserter.

> **CORRECTED, same session:** my plan recorded that the working tree does not compile and that both
> of my edit targets were held dirty by the bug-260 session, with a whole mitigation built around it
> (build from a clean `git archive HEAD` extraction; sequence the emitter edit behind a re-check).
> Re-checked at execution start: HEAD has moved to `2b3e61e8e`, that session has **committed**,
> `go build ./platform/orchestration/actions/` exits 0, and **both target files are clean**. The
> mitigation is no longer needed for correctness — the clean-extraction build is now belt-and-braces,
> not a workaround. Keeping the emitter edit in its own commit anyway: the file is one another lane
> touches often, and a one-line commit is the cheapest thing to cherry-pick around.
>
> *The check that caught it:* re-run `git log -1` + `git status --porcelain` + a build at the start of
> every phase, never trust the snapshot the plan was written against. On this tree a planning
> assumption about another session's tree state has a half-life measured in minutes.

**Migration numbering: 497 and 498 are BOTH already taken twice** (`497_escalation_owners_map…` and
`497_…_ROLLBACK`; `498_escalation_literal_markdown…` AND `498_schedule_meta_description_backfiller…`).
So the directory already contains a same-number collision from two lanes, and the highest number is
501. This lane takes **502** and **503**. Read the directory fresh before authoring — do not derive
the next number from a doc.

**Council scope widened under me, in my favour.** CLAUDE.md changed on disk during this session:
appliable migrations (`docs/agent_docs/sql_for_agents/NNN_name.sql`) are now IN council scope
(`bugs_open/314`), with the scope single-sourced in `scripts/council-scope.sh` and `DRY_RUN=1` on the
097 trigger available to test admission for free. So both migrations go through the gate with the Go,
in one submission, rather than the Go alone.

## 2026-08-20 (c) — the Go half, and the two things mutation testing caught

**Written:** `platform/orchestration/actions/head_assembly.go` (`spliceOpenGraph`, `headLangAttr`,
`htmlDocumentOpen`) + `head_assembly_test.go`; wired into `assemblePage` at two points
(`rerender_single_page_action.go` — the splice inside the stored-head branch, and the doctype write).

**The package would not compile in the working tree, and it was not my change.** Another session had
just added an untracked `platform/orchestration/actions/work_item_failure_ladder.go` declaring
`nullableJSON`, which already exists at HEAD in `hitl_persistence.go` — `redeclared in this block`,
whole package dead. Confirmed not mine (`grep -c nullableJSON` on both my files = 0; the offender is
`??` in `git status`, and HEAD has exactly one declaration). Worked around exactly as planned:
`git archive HEAD | tar -x` into the scratchpad, copy my three files over it, build and test there.
**Full package suite green in that extraction** (`ok … 2.539s`), so the change is clean against HEAD
and the breakage is theirs to land.

**Misstep 1 — I wrote tests against a shared fixture's assumed contents.** Three tests failed on
first run asserting an `og:description` that never appeared. Cause: `canonicalTestPage()`
(`inject_canonical_link_test.go:29`) carries **no `MetaDesc`** — deliberately, because the canonical
tests want that shape and most of the fleet has no meta description. I had assumed it did. Fixed by
adding a local `ogTestPage()` helper rather than mutating a fixture four other tests read.
*The check:* read a shared fixture before asserting against its contents — it is another test's
input, not yours, and its emptiness may be the point.

**Misstep 2 — a mutation reported PASS because the mutation never applied.** The `\b`-removal
mutation on the head-lang regex came back green. The `sed` escaping had silently failed and the file
was unchanged. Applied via Python (with an `assert s != before`) it fails correctly:
`headLangAttr = "fr", want ""` — the `\b` is load-bearing, without it `<header lang="fr">` matches.
*The check:* `grep` the mutated line and confirm it changed before reading the verdict, or the run is
telling you about code you never touched. A false PASS from a failed mutation looks exactly like a
test that does not discriminate.

**Misstep 3, the expensive one — my ordering claim was backwards, and my own test could not fail.**
Full account in `WRONG_CALLS.md` and PLAN D3's correction block. Short version: reversing the call
order inside `Test…DoNotCollide` left it passing, because the fixture (copied from a live head)
carries the exact blank `<meta name="description" content="">` and so drives
`spliceMetaDescription` down its *targeted* path, where the fallback hazard cannot fire at all. A
discriminating fixture (blank og placeholders, **no** blank description tag) showed that og-splice-first
strips the blanks it owns, promotes `og:image` to "first blank", and the page description lands in
og:image — a tag outside my strip set, so nothing cleans it. **Order swapped in the shipped code**;
`TestSpliceOpenGraphRunsAfterTheDescriptionSpliceForAReason` now fails on the swap, verified.

**Mutation ledger — every test proven load-bearing by a mutation actually run:**

| # | mutation | tests that failed |
|---|---|---|
| M1 | strip narrowed to blank-only og tags (the bug file's original fix candidate) | SelfHeals, OmitsDescription, Idempotent, SkipsAnOgURL, ordering |
| M2 | inject without stripping | SelfHeals, OmitsDescription, Idempotent, SkipsAnOgURL, ordering |
| M3 | og:url from raw `page.URL` instead of `preferredPageURL` | AgreesWithCanonicalAndJSONLD |
| M4 | emit og:description even when empty | OmitsDescription |
| M5 | drop `\b` from the head-lang regex | HeadLangAttrReadsTheHeadAndNotTheHeader |
| M6 | change the lang default away from `en` | HTMLDocumentOpenDefaultsToTodaysBytes |
| M7 | drop the og:url eligibility guards | SkipsAnOgURL |
| M8 | swap the two splice calls | RunsAfterTheDescriptionSpliceForAReason |

**Note M1's first row.** The mutation that reproduces *the fix candidate written in the bug file* —
fill blank placeholders only — fails five tests. That is the clearest statement of why the design
changed shape, and it is worth putting in the council submission rather than arguing it in prose.

## 2026-08-20 (d) — the 090 verdict: UNVERIFIABLE, and the reason is worth more than a CONFIRMED would have been

Run corr `af31ec22-5662-4798-91b9-b12132ebca70` completed at 09:24Z. **Verdict: UNVERIFIABLE,
`stopped_by: iteration-cap`** — "Diagnosis NOT confirmed. Best-effort trail attached for a human;
no fix proposed."

**Read it honestly: this is NOT a refutation, and it does not touch the premise.** The loop's own
`still needed` text says what it could not do, and both blockers are structural rather than
evidential:

1. **`pages.rendered_head` returned 0 rows across every attempt, fleet-wide.** Those three `pages`
   columns are VESTIGIAL — already a documented landmine, and the subject of `bugs_closed/270`. The
   loop reached for the obvious per-page head store and there has never been anything in it.
2. **Every `site_components.rendered_html` row it fetched was TRUNCATED before the tail near
   `</head>`** — precisely where `injectBrandHeadTags`' block sits. So it fetched the one row holding
   the evidence and got the half that does not contain it.

Its conclusion is therefore: the static mechanism is confirmed from source (its citations quote
`b.WriteString("  <meta property=\"og:title\" cont…` and it enumerated 18 symbols including
`injectBrandHeadTags`, `assemblePage`, `InjectHead`, `RenderHead`), but nothing in its bundle shows
the mechanism *having produced* a duplicated or homepage-rooted tag on a real page.

**That gap is exactly what I had already closed before filing, by the instrument the loop does not
have.** Two `curl`s against `ai-agent-orchestration.com/about.html` show the duplicated `og:title`
and the homepage `og:url` beside a correct per-page canonical. That evidence is in the bug file, the
commit messages, register SEO-005 and the council submission's `grounded_in` — it does not depend on
this run. **So: premise unchanged, no claim weakened, and no second run spent** (spending one would
buy the same answer, since the blockers are not iteration-count problems).

**What the run DID earn, and it is not nothing:** it independently re-derived the mechanism from the
same functions without my framing, which is the half it is good at. And its failure mode is a
transferable trap, now filed as a LANDMINES entry footprinted on `090_TRIGGER…`, `pages.rendered_head`
and `site_components.rendered_html`: **the loop cannot see served bytes, so for a defect that lives in
deployed markup it returns UNVERIFIABLE, which reads exactly like "your claim is doubtful".** The
dangerous next move is to weaken a claim you have artefact evidence for because a DB-shaped tool could
not reach the artefact. The 090 authoring guidance ("point at the tables where the evidence lives")
quietly assumes there IS a table; for this class there is not.

## 2026-08-20 (e) — the migration-number collision I had just documented, then walked into

Committed 502/503; both had been taken while I was writing them (bug-260's
`arm_mistyped_llm_fields`, and `service_binary_capabilities`), and 504–506 had gone too. Renumbered
forward to **507/508** — a `git mv` plus a commit naming **both** old and new paths, because a
pathspec commit that names only one side ships a COPY and leaves the old files at HEAD. Verified with
`git ls-tree -r --name-only HEAD` (4 files, not 8), not with `ls`.

Two details worth keeping. The renumbering sed used `\b502\b`, which deliberately does NOT touch the
component UUID `aec98dbe-76b7-4e13-9641-e5b6ba2502aa` (no word boundary inside `ba2502aa`) or the
template md5s; I re-counted all five literals after the rename rather than assuming. And the INSERT's
`created_by` marker moved with it (`migration-503-locale-lang` → `migration-508-locale-lang`) because
the ROLLBACK matches rows on that exact string — a renumber that missed it would leave a rollback
that silently deletes nothing.

*The check, now in the RUNBOOK:* **read the migration directory at the moment you COMMIT, not when you
start authoring.** Note NOTES (b) had already recorded that 497 and 498 are each doubled — I wrote the
hazard down and still picked my number from a read that was an hour stale by the time it mattered. The
directory is shared mutable state exactly like the working tree.

**Also, in the harmless direction:** my `WRONG_CALLS.md` entry and my `000_concept_index.md` row were
already at HEAD by the time I named them on my own commit — swept in by another session's commit
(`0c33ff690`) between my write and my commit. Nothing lost, forward-only holds, and it is the
same-file passenger trap firing the way round that costs nothing. Worth noting only because the
inverse — my commit taking THEIR half-finished edit — is the same mechanism.

## 2026-08-20 (f) — council APPROVED round 1, and the deployed build does NOT carry the fix

**Council corr `3b6712d4-4565-4bfe-87f6-c47ecefd6a93`: APPROVED, 5 advisory objections, none
high-severity, round 1.** 12 seats reviewed, 6 abstained. Four seats asked for **checks rather than
changes**; all were run.

| seat | asked | answer |
|---|---|---|
| `guardian`, `prior_art_librarian` | "only one consumer of the stored head" is asserted, not checked — enumerate readers of `site_components.rendered_html` before deleting from it fleet-wide | **Claim HOLDS, now checked.** The one candidate a loose grep surfaced, `save_sections_component_floor.go:169`, reads **`page_components`**, not `site_components` — my grep matched on `slot_name`+`rendered_html`, which both tables have. Every other hit is `fix_component_template_action.go` (maintenance), a comment, or `page_components`. |
| `guardian` | the archive-write side effect of changing those bytes | **Real, and benign.** `trg_site_component_archive` fires `AFTER UPDATE OF rendered_html … WHEN (new IS DISTINCT FROM old)` — verified in `pg_trigger`. One history row per site on the next chrome render (24), and it PRESERVES the pre-change head. That is bugs_open/226's archive doing its job. |
| `guidelines` | does removing the two og template lines leave a dangling `required:true`? | **No — and the reason is a finding.** `head-seo-standard` declares **no `og_title`/`og_description`/`og_image`/`site_name` fields at all**. Live list: `accent_color, background_color, canonical_url, description, font_url, gtm_container_id, primary_color, secondary_color, structured_data, text_color, theme_css, title`. So `{{if .og_title}}…{{else}}{{.title}}{{end}}` has ALWAYS taken the else branch, and `{{.title}}` is empty at a site-level render — **that is the direct cause of the blank og:title on 4 sites**, and I had inferred it rather than proven it until now. `canonical_url` is declared but nothing sets it. Three dead branches; 507 removes the two that emit a blank. |
| `debug_historian` | "the binary is proven running" is named as the HOLD release criterion but the mechanism is unspecified, and a tag or git state proves nothing | **Agreed and specified.** Pod-grep for `spliceOpenGraph` with both controls, on every replica. Written into 507's header. Result below. |

Two objections I am **not** closing, both fair and both recorded rather than argued away:
- `bug_historian`: the untouched wholesale idempotency guard means **any future per-page tag added to
  `injectBrandHeadTags` reproduces bug 252's exact shape** — og:url was the symptom, that guard is the
  mechanism. It is `bugs_open/322` item 4; the seat's framing is better than that file's and has been
  carried into `FINDINGS_2026-08-20_errors_caught.md` §C2.
- `architecture`: this is the **fourth** fix to land on one head producer only, and a fifth should not
  be routine. A binding escalation threshold is now in SEO-005: **a fifth instance raises an RFC
  rather than taking a fifth patch.**

### The deploy check — and it says no

A fresh chassis build was deployed today. **It does not carry this fix**, and the tag would have told
me the opposite:

```
pods: agent-chassis-86b95b967b-{2fqm5,jwdb5}  image v1.0.1319  started 2026-08-20T10:18Z
my Go commit 4abcd55a4: 2026-08-20T14:03:00Z    <- ~4 hours AFTER the build was cut
```

Timestamps are an argument, not evidence, so I probed the binary (2026-08-20 14:35Z), all three arms:

| symbol | expected | result |
|---|---|---|
| `spliceOpenGraph` | present iff my fix shipped | **absent** |
| `injectCanonicalLink` | must be PRESENT (positive control) | PRESENT |
| `zzQuiteImpossibleSymbol42` | must be absent (negative control) | absent |

The positive control is the load-bearing arm: it proves the probe can find a symbol in this binary, so
"absent" means absent rather than "the probe is blind". **Both migrations therefore stay HELD, and no
canary is possible yet** — applying them now is precisely the trap 507's header describes, and it
would leave the fleet re-stamped and still wrong.

⚠ **Also, the recommended provenance command fired the documented landmine on me.**
`kubectl logs -l app=agent-chassis --tail=N | grep 'build provenance'` returned **2.4MB of another
lane's landmine corpus** — the chassis logs whole council/diagnosis payloads, and those payloads quote
the phrase. The symbol probe has no such failure mode. Recorded in
`FINDINGS_2026-08-20_errors_caught.md` §B2 as a fix to the recipe in CLAUDE.md, not just a trap to
know about.

## 2026-08-20 (g) — LIVE on v1.0.1320, og half PROVEN at the artefact, migrations applied

**Binary probed on BOTH replicas of v1.0.1320 (started 16:09Z, my Go commit 14:03Z):**
`spliceOpenGraph` PRESENT, `headLangAttr` PRESENT, `htmlDocumentOpen` PRESENT, positive control
`injectCanonicalLink` PRESENT, fabricated negative control absent. Five arms, both pods.

### The og half is proven, in both directions

Assemble-only rerenders (`049b_deploy_single_page.sh`, direct route — the `spawn_agent`→`call_agent`
wrapper is the one that hangs), verified by the orchestration row rather than `kcat`'s exit code, then
read at the served bytes.

`/about.html` — corr `a4913050`:

| | before | after |
|---|---|---|
| `og:title` | **two** — `""` and `"AI Agent Orchestration"` | **one** — `"About \| AI Agent Orchestration"` |
| `og:description` | **two** — `""` and the site tagline | **one** — the page's own |
| `og:url` | `https://ai-agent-orchestration.com/` (**the homepage**) | `…/about.html` |
| `canonical` | `…/about.html` | `…/about.html` — **now agrees** |
| blank `content=""` og tags | 2 | **0** |
| `og:type`/`og:site_name`/`og:image` | present | **byte-preserved** |

`/index.html` — corr `1e35e7e4`, and this is the **discriminating control**: `og:url` came out as the
bare `https://ai-agent-orchestration.com/`, **not** `/index.html`, matching its canonical. So
`preferredPageURL`'s root normalisation carried into og:url, and the pair proves the value is no longer
constant-per-site *and* that the root case does not regress. `og:title` also came out
HTML-escaped (`Kafka &amp; Postgres`), so `htmlEscapeAttr` is doing its job on real copy.

### Migrations applied — and 508's guard earned its place immediately

Backups taken first (council advisory). **507: `UPDATE 1`, `UPDATE 1`, three DO guards passed.**

**508 ABORTED on its first attempt, correctly:**
```
ERROR: real sites exist that this migration does not name: indoorplanters.co.uk
```
That site was created **the same day I authored the file**. The guard I added because a blanket rule
would have been wrong is what stopped a brand-new site silently keeping `en`. It has no identity spec
and no content yet, so unlike every other row its evidence is the `.co.uk` registration plus estate
context — added as `en-GB` and **marked `[EVIDENCE-THIN]` in the file**, with a note to re-check when
it has copy. Re-applied: **15 merged, 11 inserted, all three assertions passed** (25 en-GB, 1 es-ES,
`analytics.gtm_container_id` still on 14 rows).

**Owner ruling recorded (2026-08-20): non-English sites must NOT be en-GB, and this generalises to
future language sites** — the explicit-domain list, not TLD derivation, is the mechanism that
implements it.

Both files renamed to drop `_HOLD` (the runner refuses to record an uppercase-suffixed sidecar —
`446_asset_retraction_agent` is the precedent) and recorded `--record-only`. Their original banners
are kept as the RECORD of why they were held, annotated with what released the hold.

### Propagation is armed, and its blast radius is measured rather than assumed

Recomputed the chrome fingerprint's `template` digest against every stored `render_inputs`:

| slot | drifted | total |
|---|---|---|
| **head** | **22** | 24 |
| footer | 0 | 24 |
| header | 0 | 24 |

Exactly the two shared head templates, nothing else. So `StaleSiteComponentsCheck` will file
`stale_chrome` per site on its next discovery run → `rerender-pages` with
`refresh_site_components:true` → chrome re-render plus per-page fan-out. That is the owner-approved
wave rollout, and it is now demonstrably armed rather than hoped for.

### A side-quest defect found and NOT chased: `rerender-chrome` skips its own start step

Dispatched it (corr `83d980c2`) to force one site's chrome ahead of the schedule. It returned
`complete | COMPLETED` in seconds, the stored head was untouched (`updated_at` still 2026-08-18), and
`collected_data` holds **no `site_components_result` and no `render_site_components` key** — only
`complete`, whose result is a verbatim echo of the request. No `__step_error`, no `error`.

The live definition is correct on inspection: `start_step: render_site_components`, `force_rerender:
true`, `next_step: complete`, `output_field: site_components_result`. So the agent ran
`complete_workflow` as its first and only step and reported success. **A COMPLETED orchestration that
did nothing** — the platform's own "trust the artefact, not the status" rule, met in the wild.

Not chased: it belongs to the 226/351 lane and my propagation path does not need it. Filed in
`FINDINGS_2026-08-20_errors_caught.md` §B4 so it is not lost. ⚠ Anyone reaching for `rerender-chrome`
as a lever should verify it wrote something before believing it.
