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

## 2026-08-20 (h) — BOTH HALVES PROVEN AT THE ARTEFACT, and I withdrew a wrong finding to get there

**The lang half, end to end.** Dispatched `rerender-pages` with `spec.refresh_site_components: true`
(corr `5dba9a92`) → 14 steps ran including `render_site_components` → stored head became
`<head lang="en-GB">` (`updated_at` 17:33:04) → page rerender (corr `e554bc8e`) →

```
$ curl -s https://ai-agent-orchestration.com/about.html | grep -oE '<html[^>]*>|<meta property="og:(title|description|url)"[^>]*>|<link rel="canonical"[^>]*>'
<html lang="en-GB">
<meta property="og:title" content="About | AI Agent Orchestration">
<meta property="og:description" content="Fix multi-agent systems failing in production: …">
<meta property="og:url" content="https://ai-agent-orchestration.com/about.html">
<link rel="canonical" href="https://ai-agent-orchestration.com/about.html">
```

**That is the whole fix, on the wire**: the document declares British English, and the page states its
own Open Graph identity in agreement with its canonical. The full chain is proven —
`site_specs.site_config.locale.lang` → `input_schema` `source: config.locale.lang` → the gated
`<head{{if .lang}}…>` → `headLangAttr` → `htmlDocumentOpen`.

### ⚠ MISSTEP, and it nearly became a bug filed against working code

Between (g) and here I wrote up **two agents as broken** — `rerender-chrome`, then `rerender-pages`
"reporting COMPLETED having skipped their own start step". **Both were fine. My dispatch envelope was
incomplete.** `049b_deploy_single_page.sh` sends five Kafka **headers** I omitted (`action=orchestrate`,
`sender_agent_type`, `sender_agent_id`, `responses_topic`, `timestamp`); I had `action` only in the JSON
**body**. Headers added, nothing else changed, and the same agent ran all 14 steps first try.

**Why the wrong story fitted:** `COMPLETED`, `error` NULL, no `__step_error`, and `collected_data`
holding exactly one key — `complete`, echoing my own request. Indistinguishable from a skipped start
step. I then read the live definition, found `start_step`/`force_rerender`/the step key all correct,
and treated that as confirming a platform bug — when a correct definition beside a no-op run should
have pointed **upstream of the workflow**, not into it.

**The check that settles it in one command, and which I had the材料 for the whole time:** diff
`collected_data`'s keys against a run of the same shape that WORKED. Fourteen step keys versus one.
**A broken workflow and an unrouted message differ in how many steps appear, never in the status.**

**And a second instance was not corroboration**, because both runs shared my method. The
discriminating experiment was to vary the envelope, not the agent. Retracted in FINDINGS §B4 (kept
visible, struck through, with the evidence left in place) and logged in `WRONG_CALLS.md`.

**What survives as a real finding — FINDINGS §B4b:** an orchestrate message missing its headers is
**accepted, given an orchestration row, and marked `COMPLETED` having run nothing**. That is a
dispatch-surface trap worth fixing, and it is not the agents I accused. The `rerender-chrome` claim is
**withdrawn entirely** — my run cannot speak to it, since it carried the same broken envelope.

## 2026-08-20 (i) — fleet CHROME fully repaired; PAGES are the remaining work, and I have sized it

Drove `rerender-chrome` across every drifted site (correct envelope this time), in foreground batches
— a `while read` loop with `kubectl run -i` inside it exits after ONE iteration, because `kubectl -i`
competes for the loop's stdin. Fixed by writing the payload to a file and redirecting (`< msg.json`)
so the loop keeps its own stdin. Worth remembering: the loop reported success and had silently done
1 of 20.

**Fleet chrome, before → after (all 24 head rows):**

| measure | before | after |
|---|---|---|
| fingerprint drifted | 22 | **0** |
| heads carrying `<head lang=…>` | 0 | **22** |
| heads still baking a homepage `og:url` | 22 | **0** |
| heads with a blank `og:title` placeholder | 4 | **0** |

All four head families verified individually at the stored artefact:
`relojistas.com` → **`<head lang="es-ES">`** (the owner's non-English ruling, working),
`leopardessconsulting.co.uk` (head-seo-standard) → `en-GB`,
`noted.co.uk` (Document Head) → `en-GB`,
`webdesign.co.uk` → unchanged fragment, no `<head>` tag to carry it — **expected, documented, not a
failure.**

`rerender-chrome` deliberately touches no pages, so it cost 22 chrome renders and zero page churn.

### What is NOT done, measured rather than assumed

Spot-checked two live pages on non-canary sites:

```
https://noted.co.uk/index.html            -> <html lang="en">  og:url = https://noted.co.uk/
https://leopardessconsulting.co.uk/about.html -> <html lang="en">  og:url = https://leopardessconsulting.co.uk/
```

**Stored heads are fixed; deployed pages are not, and will not be until each re-assembles.** That is
~698 pages (700 minus the two canaries). The defect is no longer *reproducible* — any rerender now
produces a correct page, proven — but the *damage* persists on every page that has not re-rendered.

**So the lane cannot close yet**, and the remaining decision is a real trade-off rather than a task:
forcing the pages means ~698 `page_rerender` items into a queue whose drain half is `bugs_open/083`,
which is other lanes' risk too — and seed `351`'s own header warns that route is "two orders of
magnitude more churn than the goal". Letting natural rebuild traffic carry it is free and slow.
Recorded for the owner rather than chosen unilaterally.

## 2026-08-21 — OWNER RULING: do not force the pages. And the decay measured, which changes what "let rebuilds carry it" means

**OWNER RULING 2026-08-21: do NOT force the ~698 page rerenders — let natural rebuilds carry it.**
Recorded as the decision it is: the queue-saturation risk (`bugs_open/083`'s drain half, other lanes'
exposure) is judged to outweigh the speed. No `page_rerender` wave will be dispatched by this lane.

**v1.0.1321** (rolled 2026-08-20T19:51Z) re-probed on both replicas: `spliceOpenGraph`,
`headLangAttr` and `injectCanonicalLink` all PRESENT, `zzImpossible99` absent. So the fix **survived a
roll** — worth stating, because a later build cut from an earlier commit would have silently removed it.

### The decay, measured — and the metric validated at the artefact before I quote it

Signal: a page carries the fix iff `GREATEST(deployed_at, last_built_at) > its site's head
updated_at`. **Validated 3 of 3 against served bytes, in BOTH directions** before use — two pages the
DB called stale serve `lang="en"`, one it called fresh serves `lang="es-ES"`. Without that check the
number was not usable: my own dispatch window (08-20 17:00–18:00, 212 pages) sits right where a
bookkeeping-only `deployed_at` bump would have inflated it.

| measure | value |
|---|---|
| assembled pages fleet-wide | **722** (was 700 on 08-19 — the estate grows) |
| carrying the fix | **252 (34.9%)** |
| still serving the old head | **470** |
| **sites at 0% carried** | **13 of 26** |
| natural rebuild rate, excluding my own dispatches | **~14 pages in 15h (~1/hour), bursty** |

**The 13 zero-percent sites are the finding, not the 34.9%.** finetuning.uk (49 pages),
loancalculator.co.uk (43), leopardessconsulting.co.uk (40), mortgagecalculator.co.uk (30), idea.uk
(24), loancash.co.uk (22), lendzy.co.uk (20), loanzy.uk (14), noted.co.uk (12), oufe.com (11),
webdesign.uk (8), cookly.uk (5), remortgagecalculator.uk (4). At ~1 page/hour fleet-wide and bursty,
**"let rebuilds carry it" means weeks for active sites and effectively never for quiet ones** — which
is precisely the caveat attached to this option when it was first offered, now with numbers on it.

**So the honest framing for the close-out decision: the DEFECT is fixed (not reproducible — any
rebuild produces a correct page, proven at three artefacts), and the DAMAGE is 470 pages with no
scheduled end.** Those are different questions and the bug file should say so rather than let "fixed
and live" imply the fleet is clean.

**Not a reason to reopen the forcing decision** — it is the owner's, it is recorded, and the cheap
half (chrome, 22 renders, zero page churn) is already done. But it does mean a residual worth naming
in the close-out rather than a rounding error.

## 2026-08-21 (b) — the owner's four decisions, executed

**1. 252 CLOSED + tracking item.** Moved to `bugs_closed/` (both paths named on the commit; verified
at HEAD with `git ls-tree`, one line, no copy). `bugs_open/346` filed for the residual — 502 of 727
pages, **12 real sites at zero**. It is a tick-list, not a defect: it heals free whenever a lane next
touches one of those sites, and it carries the two dispatch traps that cost this lane time.

⚠ **A metric trap worth carrying forward, recorded in 346.** Comparing a page's rebuild time against
`site_components.updated_at` **re-classifies already-fixed pages every time chrome re-renders for any
reason** — my reading went 252 fixed → 217 an hour later with no page having regressed. Pin the clock
to when the first corrected chrome landed instead.

**2. webdesign.co.uk FIXED LIVE** — `bugs_closed/347`, migration `529`. Its head component was a bare
fragment with no `<head>` element; wrapped with the same gated-lang contract `507` gave the shared
templates, md5-guarded, with an assertion pinning the hand-authored contents so a wrap cannot quietly
replace the body. Proven at the artefact: an assembled guides page went from `<html lang="en">` with no
head element to `<html lang="en-GB">` plus `<head lang="en-GB">`.

Two verification traps, both of which nearly gave me a wrong read: my first check showed the OLD bytes
because the run was still `AWAITING_RESPONSES` at `deploy_page` — indistinguishable from a failed fix;
and `webdesign.co.uk/about.html` **already had** a `<head>` element because it is not an assembled
page, so verifying there would have shown "already fine" and hidden the defect completely.

**3. `site-locale-unset-check` LIVE** — registered **SEO-006**, daily 07:15 UTC, deployed and its
first manual run CLEAN with a `doc_notes` row. It reports two shapes; **B is the one worth having**: a
site whose locale IS set and whose head template has no `{{if .lang}}` gate to render it, which looks
*finished* while only the served page disagrees — 347's exact shape.

**It caught a real case before it was deployed.** I tested it by running its predicate against live
data rather than trusting a clean unit run, and it surfaced `buytoletcalculator.uk`, created that same
day, unset (migration `530`, marked `[EVIDENCE-THIN]` — no content yet to judge). That is the honest
way to test a check: run it against production and see whether it finds something you did not already
know.

**4. 322 item 4 FIXED** — the guard, not the tag. My write-up of this decision was ambiguous ("I
removed the offending tag; the guard is untouched") and the owner's reply could have meant either, so
I asked rather than guessed — the two readings led to materially different work on a shared renderer.
Per-tag idempotence now; commit `c2f050036`, `Council-Submitted: 54c660f8`.

> **MISSTEP: I committed platform code BEFORE submitting it.** `c2f050036` therefore carries no
> trailer, and forward-only forbids an amend — so the `098` coverage report **will list it as
> un-reviewed for ever**, even though it was reviewed. The trailer is a join key written at commit
> time; there is no way to attach one afterwards. *The check:* submit first — `097` prints
> `SUBMISSION_CORR` in seconds — then commit with `Council-Submitted:`. I did exactly this correctly
> for the 252 work earlier the same week and then didn't, because this fix felt like a small follow-on.
> Cost is bookkeeping only, and it is recorded here and in the 322 file so the trail survives the
> report's blind spot.

## 2026-08-21 (c) — 322 item 4 drew a REVISE, and the round paid for itself twice

Round 1 (corr `54c660f8`): **REVISE, gating objection from `editquality`**, plus objections from
`bug_historian`, `prior_art_librarian`, `guardian` and `debug_historian`. Two were **real defects**;
the rest were things I had asserted instead of shown. Revised and resubmitted on the same trail
(`RESUBMIT_CORR`), which is the point of the mechanism — the trail accumulates.

### The two real ones

**1. `bug_historian`: "'documented' in a comment is not a fail-loud guard."** My no-`</head>` branch
declined **silently**, with a comment explaining the divergence. That is *the exact shape this change
exists to remove* — a silent skip is how webdesign.co.uk lost every brand tag on 117 pages while every
caller reported success. I had written a careful comment about a silent failure in the middle of a fix
for silent failures. It now logs a `Warn` naming the domain and the consequence.

**2. `editquality`: the favicon comment contradicted the code.** The code writes **two** `rel="icon"`
when the head declares none (derived PNG + the site logo as a secondary, so a mark resolves before
`favicon.png` is committed) and **none** when one is authored. That asymmetry is deliberate and
pre-existing. My comment said "a second one must never be appended" without "beside an **authored**
one" — which reads as a rule the code breaks. **The comment was wrong, not the code**, and a reader
who "fixed" the code to match it would have removed a real fallback.

### The four I answered with checks I should have run before submitting

All three **high**-severity objections reduced to one question — *does the OTHER head producer see this
function's output?* — and the answer is **no**, provable in one read: `RenderHead`
(`component_library.go:2017`) resolves via `ResolveChromeComponent` and falls back to
`RenderFallbackHead`. **It never reads `site_components.rendered_html`.** So a page built through
`AssemblePageAction` gets neither these tags nor a stranded site-level `og:title` — the population
`prior_art_librarian` feared cannot exist. I had *asserted* this in the 252 round and never re-quoted
the source here.

Also now measured rather than estimated:
- **`injectBrandHeadTags` has exactly ONE caller** (`renderAndStoreSiteComponent`, gated `slot=="head"`).
- **The needle-gate count `debug_historian` asked for: exactly TWO stored heads are short of brand
  tags** — webdesign.co.uk (the motivating case, 117 pages) and loanandmortgagecalculator.co.uk
  (missing `og:image` only). A small, named repair population rather than an unbounded one.
- `guardian`'s question — does `spliceOpenGraph` tolerate a non-empty pre-existing `og:title`? — was
  **already tested**: it strips unconditionally, and `TestSpliceOpenGraphSelfHealsDuplicatedBrandTags`
  runs it against a fixture carrying a FILLED site-level `og:title`.

### The lesson, and it is not "the council was fussy"

**Every one of the four answerable objections was a claim I could have cited and instead asserted.**
The evidence existed — some of it produced by me, in this same lane, two days earlier. Grounding a
resubmission cost one query each. **A `grounded_in` entry is cheaper than a review round, and a review
round is far cheaper than the defect it finds** — and this round did find two.

## 2026-08-21 (d) — round 2 APPROVED, and its three advisories were all worth acting on

**Verdict: APPROVED — 3 advisory objections, none high-severity, 5 of 13 seats abstained.** Trailer
`Council-Reviewed: 54c660f8-1e05-4b88-9910-0d1427b1d805`. Round 1's two gating defects are fixed; the
three that remained are advisory, and none of them was noise.

### Acted on in code

**`bug_historian`: "a log line is not a durable fail-loud signal here."** Quantified in their own
words — *"chassis pod log retention is ~90 seconds and `orchestration_states.error` is a near-empty
sink versus `agent_error_log`'s actual volume. A Warn nobody durably records ... is symptom-patched,
not fail-loud — it satisfies the letter of round 1's objection without giving any downstream process a
row to act on."*

**They are right, and it is round 1's lesson one level deeper: I answered "your skip is silent" by
adding a log, and stopped.** `injectBrandHeadTags` now returns `(head, declinedReason)` and the caller
writes an `agent_error_log` row (`BRAND_HEAD_TAGS_DECLINED`) carrying domain, consequence and remedy.
The reason is *returned* rather than recorded in place because the function has no DB handle and its
one caller does — widening its dependencies to reach the error log would be the larger change.
Mutation-proven both ways: an empty reason on decline FAILS (that is the silent skip returning), and a
spurious reason on the healthy path FAILS (that would write a row on every render).

### Answered by measurement rather than change

**`bug_historian`: are the sibling injectors the same shape?** **No, and the distinction is worth
keeping.** `injectCanonicalLink`, `injectPageJSONLD`, `injectRobotsNoindex` and `injectComponentCSS`
are each **single-tag** emitters that check for *their own* tag before adding it — that is correct
idempotence. The wholesale shape is only a defect when **one marker gates MANY tags**, and
`injectBrandHeadTags` was the only multi-tag emitter in the family. So there is no unaudited sibling
carrying this bug.

**`guardian` (low): could `declaresHeadTag` false-negative and append duplicates fleet-wide?** Measured
across all 24 stored heads: **23 use double-quoted `property=`/`name=`, and there are ZERO
single-quoted, ZERO spaced-equals (`property = "…"`) and ZERO content-before-property forms.** The two
quote styles cover 100% of real authored variety, so the fleet-wide duplicate-append risk does not
exist today. (The 24th head is webdesign.co.uk, which carries no og tags at all.)

### Carried forward, not closed — both are real

**`guardian` (medium): how do the two short heads actually get re-rendered?** Correctly identified:
`renderAndStoreSiteComponent`'s `!force` idempotence exit means **code alone regenerates nothing** —
the same fact that shaped 252's whole design. The mechanism is a `rerender-chrome` dispatch per site
**after the next roll**, exactly as used for the 22 heads earlier today. **Owed: webdesign.co.uk and
loanandmortgagecalculator.co.uk once a build carrying `declaresHeadTag` is live and probed.** Noted in
`bugs_open/346` as the tick-list's sibling.

**`debug_historian` (medium): the "2 of 24 stored heads short" count is scoped to the
`render_site_components`-driven population, not the fleet.** Also right, and I did present it as the
whole gap. Pages built through `AssemblePageAction` → `RenderHead` → `RenderFallbackHead` get **no
brand tags at all** and this fix cannot reach them. I could not bound that population honestly: the
three agent types carrying `assemble_page` (`pageflow-builder`, `page-rebuild`,
`site-work-orchestrator`) are all `is_active`, and `orchestration_states` shows **zero runs for any of
them** — but that table retains ~24h, so **that is a weak negative, not proof the path is dead**
`[UNMEASURED beyond 24h]`. It wants a ticket rather than a claim.

**`prior_art_librarian` noted its approval rests on my greps rather than an independent check** — fair,
and the two claims are cheaply re-verifiable: `RenderHead` at `component_library.go:2017`
(`ResolveChromeComponent` → `RenderFallbackHead`, no `site_components` read), and
`injectBrandHeadTags`' single caller at `render_site_components_action.go:1139`.
