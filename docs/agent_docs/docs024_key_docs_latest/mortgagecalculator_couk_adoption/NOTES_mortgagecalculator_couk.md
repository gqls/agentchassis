# NOTES — mortgagecalculator.co.uk adoption

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 2026-07-31 — session 1, picking up the handoff

### The handoff's stated blocker had already dissolved

Handoff §1 carries an explicit `[UNMEASURED]`: *"I could not enumerate the bucket.
The `b2` CLI is present on this machine but there are no B2 credentials
(`env | grep B2_` is empty; the keys live only as GitHub Actions secrets)."* It then
makes bucket enumeration item 1 of "what I would check first".

That is no longer true. `env | grep B2_` is still empty — so **the check the handoff
used would still report "no credentials"** — but `~/.config/b2/account_info` exists
(mtime 2026-07-31 21:56 local) and `b2 account get` authorises with
`listBuckets`/`listFiles`/`readFiles`. The credentials are in the CLI's sqlite
account store, not the environment.

**Lesson, and it is the generalisable one:** an absence measured through **one**
access path is not an absence. `env | grep B2_` answers "are the keys in the
environment", which is not the question "can this machine read the bucket". The
cheap check that settles it is the tool's own auth probe — `b2 account get` — not a
grep for how you assumed the tool is configured.

Result: the whole `[UNMEASURED]` was replaced with a measurement in about ten
seconds, and the handoff's item 1 was closed.

### Bucket reconciliation — the handoff's inventory was right

`b2 ls --recursive` → 34 entries = **29 real files + 5 `.bzEmpty`** placeholders.
23 HTML (14 top-level + 9 guides), `css/style.css`, `js/calculators.js`, 2 PNG,
1 XCF, `robots.txt`. Matches handoff §2's count of 23 HTML exactly.

`gemini/02` really is the byte source of truth: 28 of 29 bucket files
sha256-identical. The two asymmetries the handoff predicted both confirmed —
`robots.txt` bucket-only, `README.md` local-only (404 live).

### MISSTEP — I nearly committed Cloudflare's edge injection into the origin

I fetched `robots.txt` over HTTPS to fill the gap in the local tree. It came back
**2,327 bytes** with `x-amz-*` headers present, which is the usual tell that a file
is served from B2 rather than synthesised by Cloudflare. The handoff had already
characterised it as *"a real origin file (substantial, WordPress-style hardening
rules), **not** Cloudflare's Managed robots.txt block"* — so both the header
evidence and the prior thread agreed.

Both were wrong, and it is a **composite**. The served file is the origin file with
a `# BEGIN Cloudflare Managed content` … `# END Cloudflare Managed Content` block
**injected at the edge**. Pulling the same object straight out of the bucket gives
**491 bytes** and `grep -c "Cloudflare Managed"` → **0**.

Had I committed the fetched copy, the origin would permanently carry a hardcoded
copy of Cloudflare's managed block, which Cloudflare would then inject *again* on
every request — a duplicated directive set in a file whose whole job is to be
parsed strictly by crawlers.

**What caught it:** printing the whole file instead of the tail. **Why the handoff
missed it:** it ran `curl … | tail -5`, and the injected block sits at the *top*,
above the origin's own rules. A tail is a fine way to confirm a file is non-empty
and a bad way to establish what a file *is*.

**The transferable check:** when bytes are the deliverable, read them from the
origin store, not through a CDN. `x-amz-*` headers prove where the object was
fetched from; they prove nothing about whether the edge rewrote the body on the way
out. The 28/28 HTML matches are the control that makes this specific rather than
paranoid — Cloudflare rewrote `robots.txt` and nothing else.

### MISSTEP — a failed dry run printed as a clean one

First dry-run attempt used `--dryRun` (b2 CLI v3 spelling). This CLI is v4.7.0, so
it exited **2** with a usage dump. My pipeline then ran
`grep -i '^delete' … || echo "(none)"` over that dump and printed:

```
=== DELETIONS the sync would perform ===
(none)
=== UPLOADS the sync would perform ===
(none)
```

Which is exactly what a perfectly safe no-op sync would print, on the run that
verifies I am not about to delete a live site.

**What caught it:** I printed `=== EXIT ${PIPESTATUS[0]} ===` in the same block and
it read 2. **The lesson is not "check exit codes"** — it is that a `grep … || echo
"(none)"` idiom *manufactures* a reassuring answer out of a failure, because zero
matches and zero output are indistinguishable to it. The check and the failure mode
produce the same characters. Any "no findings" print needs a positive control in the
same block, and for a subprocess the exit status is the cheapest one available.

### The real dry run — 29 uploads, 35 deletes, and that is fine

With `--dry-run`, exit 0:

- **29 uploads** — every file, because `b2 sync` compares mtime and freshly-staged
  copies are newer than Jan-2026 bucket objects. `--skip-newer` skips only when the
  **destination** is newer, so it does not apply. The GitHub runner's fresh
  `actions/checkout` has the same property, so this is what every domain's deploy
  already does.
- **30 deletes marked `(old version)`** — B2 version pruning, each paired with a
  re-upload of byte-identical content. `index.html (old version)` twice, so the
  bucket held two superseded versions.
- **5 `.bzEmpty` deletes** — the placeholders.
- **0 live content files deleted without replacement.**

I want to be careful about how this gets summarised later, because "the sync is a
no-op" is the tempting phrasing and it is false. The true statement is narrower:
**the sync changes no served content.** 29/29 staged files are sha256-identical to
the bucket objects they replace, and `comm -23 bucket_files staged_files` is empty,
which is the direct evidence that `--delete` removes no content.

### Decision D1 recorded, and why I am writing it down at length

The owner chose `--fidelity high` after being shown the code. `high` is not a
softer `locked`; it is the *absence* of a setting, falling through to the recreate
path — new synthesised URLs for every page, every page regenerated by an LLM. I put
the choice in front of them with `apply_adoption_plan_action.go:426` and
`082_submit_domain_unified.sh:64-66` quoted, and they took `high`. That stands.

Recording it in PLAN as D1 because in three weeks the live site will have
`/tools/repayment/index.html` where it used to have `/repayment.html`, and the
question "was that intended?" needs an answer that is not archaeology.

**One thing I have carried across from the handoff and must NOT:** handoff §5d says
assert `needs_content_page` + `needs_tool_recreation` = **0**, "if either appears,
an LLM is about to rewrite working calculators — stop." That is the **`locked`**
assertion. Under `high` those work items are the intended output. Copying that check
across would have me halt the run at its first correct step. Noted in RUNBOOK §6.

### State at end of this entry

Deploy repo populated and committed locally (29 files, pathspec-scoped). **Not
pushed** — the push triggers the live sync, so it is the owner's call. Nothing has
touched the cluster or the live site yet.

---

## 2026-07-31 — session 1 continued: the deploy landed, and two more measurement traps

### The mirror push, and why it was rebased rather than merged

Before pushing I found **origin was 30+ commits ahead**, all `Rerender:` commits —
the platform actively committing into `gqls/sites` while I worked. That is handoff
§7's "two independent writers" claim, corroborated first-hand rather than inherited.

I rebased my single unpushed commit rather than merging, for a specific reason. The
workflow computes its target from `git diff --name-only HEAD~1 HEAD`. On a **merge**
commit, `HEAD~1` is the first parent — my own commit — so the diff would show only
the *other* domains' rerenders and **my domain would be absent from the changed
list**. The sync for `mortgagecalculator.co.uk` would silently not run and the job
would still go green. (Same shape as the fleet lesson "a GREEN run ships NOTHING if
your push became a merge".) Rebasing keeps my commit as the tip, so `HEAD~1` is the
previous origin tip and the diff is exactly my 29 files.

**Result, verified in the run log rather than assumed:** `Changed domains:
mortgagecalculator.co.uk`, sync executed, and all 29 live files sha256-identical
before and after. The bucket is now 29 files with 0 `.bzEmpty`, 1:1 with the repo —
the same shape as the sibling. **The outage hazard from handoff §1 is closed.**

### MISSTEP — my local `grep` is ugrep, and it silently disagreed with the workflow

After pushing the link fix I previewed what the workflow would compute:

```bash
git diff --name-only HEAD~1 HEAD | grep -E '^[^/]+\.[^/]+/' | cut -d'/' -f1 | sort -u
```

It printed **nothing**. Taken at face value that is alarming: an empty `CHANGED`
makes `deploy-to-b2.yml` fall through to `ls -d */`, i.e. **sync every domain in the
repo**. I was one step from "reporting" a deploy-pipeline defect that does not exist.

`git diff` alone returned `mortgagecalculator.co.uk/index.html` correctly, so the
fault was in the pipe. `type grep` explains it: in this session **`grep` is a shell
function wrapping `ugrep 7.5.0`**, not GNU grep. It returns **exit 1, zero matches**
for `^[^/]+\.[^/]+/` against a string GNU grep matches:

```
printf 'a.b/c\n'                            | grep -E '^[^/]+\.[^/]+/'   -> exit 1
printf 'mortgagecalculator.co.uk/index.html\n' | grep -E '^[^/]+\.[^/]+/' -> exit 1
printf 'mortgagecalculator.co.uk/index.html\n' | grep -E '^.+\.[^/]+/'    -> MATCHES
```

So a greedy negated class that must backtrack across the `\.` fails, while the
equivalent with `.+` succeeds. **The runner uses real GNU grep and computed the
domain correctly both times** — confirmed in both run logs, so the workflow was never
at fault.

**What makes this dangerous rather than annoying:** it fails the way a *true
negative* looks. Zero matches, exit 1, no error, no stderr. Every instinct says
"the pattern didn't match because the thing isn't there". Use `command grep` when
reproducing behaviour that runs elsewhere, and confirm against the real system's own
log rather than a local re-implementation of its logic. Landmine written.

**Also caught here:** `git pull --rebase` failed with "cannot pull with rebase: You
have unstaged changes" *inside a chained block*, and because the block had no `set
-e` the following `git commit` and `git push` ran anyway. They happened to be
correct — origin had not moved — but the pull silently not happening is exactly the
kind of thing that produces a merge next time. Chained git blocks need `&&`, not
`;`.

### MISSTEP — my own pre-flight query could not tell the two domains apart

The pre-flight asked "is another lane working this domain?" as:

```sql
WHERE spec::text ILIKE '%mortgagecalculator.co.uk%'
```

It returned **41 `page_rerender` rows in `triaged`** — which reads exactly like
another lane mid-adoption on our domain, and 41 is a plausible page count.

It is a substring. **`loanandmortgagecalculator.co.uk` contains
`mortgagecalculator.co.uk`.** Every one of those 41 rows is the sibling lane's, and
41 is precisely the count its own handoff reports ("Mine caught 41 in one second").
The same flaw hit the orchestration query: both "recent adoption runs for this
domain" were the sibling's, and reading `input_data.fidelity` on them returns
`locked` — the sibling's setting, which I could have mistaken for evidence about our
run.

Re-measured by joining `sites` and matching `domain =` exactly: our domain has
**0 sites rows, 0 orchestration runs, 0 work items**. Clean.

**Why this one is worth writing down even though I caught it:** the wrong answer was
not empty, it was *populated and plausible*. An absence would have made me look
harder; a confident 41 invited me to act on it. This is the family the index already
names — "your measurement answers the question you ENCODED" — and the specific
lesson is that **on this platform, domain names nest**: `loancalculator.co.uk`,
`mortgagecalculator.co.uk` and `loanandmortgagecalculator.co.uk` are three sites
where two of the names contain a third. `ILIKE '%domain%'` is never safe here. Join
on `site_id`, or match `=`.

### Held off dispatching — a chassis roll was in flight

Pre-flight found two replicasets live and one pod not ready: another session was
mid-roll. Latest pod start `23:10:17Z`, so the ~300s no-dispatch window runs to
about `23:15:20Z`. Waited rather than firing into it, since that failure mode is a
silently dropped spawn with no error to read afterwards.

---

## 2026-07-31 — the adoption ran, and the queue is HELD

Both orchestrations `COMPLETED`. **23 pages, 25 work items, 5 specs.**

### The crawl found 23 URLs, not the predicted 20 or 22 — and one is a 404

Both link fixes worked, confirmed in the crawl payload: `guides/buy-to-let.html` and
`guides/your-mortgage-scorecard.html` are both present, and they were the two
orphans. `404.html` was correctly not reached.

The 23rd is **`/guides/index.html`, which does not exist** — it is the target of the
6 guides' broken `Home` links (defect #1, deferred by owner decision D7). Firecrawl
followed those links and captured the B2 error body as page content:

```
statusCode: 404, title: (empty)
markdown: ```json { "error": "B2 returned error",
  "objectKey": "mortgagecalculator.co.uk/guides/index.html",
  "status": 404, ... "Key not found" } ```
```

**So a deferred cosmetic defect turned into adoption input.** That is the transferable
point: a broken internal link is not only a user-facing 404, it is a *content source*
for anything that crawls the site. Deferring it looked free and was not.

**But the outcome was benign, and I should record why rather than claim a save I did
not make.** The analyser did not treat the error blob as content. It planned
`guides-index` / `section-index` / title "Mortgage Guides | MortgageCalculator.co.uk"
— i.e. it inferred that a guides index *ought* to exist. `CanonicalisePage` maps
`section-index` with slug `guides-index` to **`/guides/index.html`**, the very URL
that 404s. So building it **fixes defect #1 as a side effect**. Lucky, not designed:
had the analyser echoed the error text instead, this would have been a junk page.

### The URL map, measured from `pages` rather than derived

My prediction from reading `CanonicalisePage` matched the created rows exactly. The
homepage is the sole survivor of its own URL:

| old (live now) | new (planned) |
|---|---|
| `/index.html` | **`/index.html` — unchanged, so it is OVERWRITTEN in place** |
| `/repayment.html` | `/tools/repayment/index.html` |
| `/simple.html`, `/stamp-duty.html`, +9 more | `/tools/<slug>/index.html` |
| `/fact-finder.html` | `/games/fact-finder/index.html` (classified `game`) |
| `/investor.html` | `/investor/index.html` (classified `section-index`) |
| `/guides/<x>.html` × 9 | `/guides/<x>/index.html` |
| *(404 today)* | `/guides/index.html` — newly created |

All 23 are `build_status='planned'`, `rebuild_policy='generic'` — **not `owned`**,
which is the recreate path's signature and the opposite of what `locked` produces.

### Work items, and why I held them

25 items, all `triaged` on creation: 12 `needs_tool_recreation` → tool-recreation-
handler, 11 `needs_content_page` → page-build-handler, 1 `needs_domain_research` →
classifier, 1 `needs_rerender` → rerender-pages.

`build-pipeline-trigger` runs at `interval_seconds=120`, so that queue was ~2 minutes
from starting an unreviewed LLM rewrite of the whole site, ending in a rerender that
overwrites the live homepage.

**Held 24 of 25 to `deferred`; left `needs_domain_research` running.** Reasoning:

- `deferred` is verified **not** in `workItemTerminalStatuses`
  (`work_items_common.go:37-44` — complete/failed/verified/rejected/wont_fix/
  unresolved/cancelled), so the rows keep their `idx_swi_dedup` slot and nothing can
  create duplicates behind them. Release is a plain `UPDATE` back to `triaged`.
- The classifier is research, not publishing, and it **supersedes** the identity spec
  — so it has to run *before* the positioning work, not after.
- Holding is one reversible statement; letting the tick fire is not. Under time
  pressure the action that preserves the owner's options is the correct default.

### The identity came back better than the handoff predicted — but still not narrow

Handoff §6 warned the sibling's auto-identity was *"UK consumers researching loans,
mortgages, car finance, and debt management"* — generic and cross-contaminated. Ours:

> UK homeowners, first-time buyers, property investors, and anyone seeking mortgage
> advice or calculations

That is **mortgage-only** — no loans, no car finance — so the contamination the
handoff feared did not happen. It is still broad ("anyone seeking mortgage advice")
and, critically, **says nothing about what this site is NOT**. No `divergence_rule`.
The classifier will supersede this anyway, so the positioning work happens after it
lands, not now.

---

## 2026-08-03 — the classifier had already failed, and it was a platform defect, not our site

Picked the lane back up expecting to read the classifier's output and start the
positioning work. It was `failed`, 3 attempts of 3, at 2026-08-02 13:41 UTC.

```
step classify_and_extract failed: ... response truncated:
stop_reason=max_tokens (output_tokens=6000 reached the configured cap,
26179 chars recovered)
```

Filed as **`bugs_open/183`**. The full evidence is there; what belongs here is the
reasoning, including two theories of mine that the measurements killed.

### It writes nothing when it fails — checked, not assumed

`classify_and_extract` is step 6 of 15 and precedes all four `write_*_spec` steps.
So the adoption-seeded specs were untouched. Confirmed at the rows rather than by
reading the step graph: every `site_specs` row for this site still carried the
adoption timestamps (`23:21:49` / `23:23:08` on 07-31), none from 08-02.

### Theory 1: "adopted sites overrun because they echo the adopted specs back" — REFUTED

This felt obviously right. The prompt has a big `{{if .site_specs.specs.site_archetype}}`
"Adoption Reference" block that tells the model to *preserve* the adopted
`content_direction` — voice, `writing_rules`, `things_to_avoid`, `example_phrases`.
More to reproduce ⇒ longer output ⇒ truncation. Ours is an adoption, so the story
closed neatly.

`llm_call_log` stores `prompt_rendered`, so it was one query instead of an argument:

```sql
SELECT (prompt_rendered LIKE '%Adoption Reference%') AS is_adoption, count(*),
       count(*) FILTER (WHERE error_message ILIKE '%stop_reason=max_tokens%') AS truncated
  FROM llm_call_log WHERE step_name='classify_and_extract' AND prompt_rendered IS NOT NULL
 GROUP BY 1;
--  f | 20 | 2   (10.0%)
--  t | 34 | 3   ( 8.8%)
```

Adopted sites truncate slightly **less** often. The theory was wrong, and it was
wrong in the comfortable direction — it would have made this "a quirk of adoption",
i.e. our problem and nobody else's, when it is in fact every site's problem.

### Theory 2: "another session is editing the classifier" — REFUTED, and this one was nearly costly

`agent_definitions.updated_at` for `domain-research-classifier` read
`2026-08-02 22:08` — four hours old when I looked. On this tree that reads as an
active lane, and the right response to an active lane is to back off.

It was a **bulk sweep: 184 rows share that minute**. `version` did not move either,
so a swept row and a hand-edited row look identical. And it postdates the failures
by ~8 hours, so it could not have caused them.

The generalisable bit, now in `LANDMINES.md`: **grep transcripts for the STEP name,
not the agent type.** `domain-research-classifier` matched 9 live sessions — every
fleet census lists it. `classify_and_extract` matched **0**. The specific string is
the one that carries information.

### What the numbers actually say

54 calls since 2026-04-02. **Zero truncations until 08-02; five of six that day.**
Cap 6000 and model `claude-sonnet-4-6` constant throughout (`model_resolved` too —
not an alias drift). Over the 49 successes: mean 4592, **p95 5551 (92.5% of cap),
max 5642 (94.0%)**.

So this was never a regression. It is a step that has been running two hundred
tokens under its ceiling for four months. And **6000 is the only step at that cap in
the entire fleet** — the modes are 8000 (47 steps) and 16000 (20). It emits one of
the largest documents in the estate on the lowest cap above 4000.

> **[UNEXPLAINED] — and left that way deliberately.** I could not find what tipped
> it on 08-02. Cap, model and prompt were unchanged and the one structural theory
> is refuted above. The honest claim is the *margin*, which needs no trigger. I have
> written that into `183` explicitly so the next reader does not invent one.

### Fix, and why not the other fixes

Raised the cap to **16000** in the live row (DB config — live on write, no image).
Two checks before believing it:

1. **`bugs_open/009`'s shadowing interaction** (flagged at `016b:759`): a **root**
   `ai_service` block makes step-level `max_tokens` dead config. This agent has
   **no** root block → the step value is live.
2. Corroboration that does not depend on my reading the JSON right: every
   pre-change `llm_call_log` row recorded `max_tokens=6000`, exactly the step-level
   value. If a root block were shadowing it, the log would have shown something else.

I did **not** add a `repairTruncatedJSON` salvage path, though one exists and is
right for the councils. Here it would be actively harmful: the repair keeps a prefix
ending at a complete value, so trailing fields go **silently absent**. `design_intent`
is the last of the four sections, and its 8-slot `palette.reference_values` is what
the composition pipeline actually reads. Salvage would produce a spec set missing
exactly the mandatory part and mark the item `complete`.

`platform/aiservice/truncation.go:26-29` says the cap raise is not a class fix, and
it is right — `experience-planner/compose` truncated at 32,000. I have recorded the
structural fix (split the step, one bounded generation per spec) as candidate 3 in
`183` rather than pretending 16000 settles it.

### The site lock did NOT hold the site — and I only found out because I checked

> **CORRECTION to my own decision D4, written the same hour.** I switched from
> deferring items one by one to `sites.locked_at`, on the reasoning that a chain
> beats hand-holding at a 120-second tick. The reasoning was right. **The mechanism
> does not work.**

Locked at 23:21:35. Fresh `build-dispatch-loop` orchestrations at **23:23:13,
23:25:44, 23:28:13**, and by the time I looked the chain had run four handlers deep —
vertical research, strategy, briefing, and `build-site-planner` mid-flight.

Three predicates, none agreeing:

| where | question | checks the lock? |
|---|---|---|
| `scheduled_tasks.build-pipeline-trigger.pre_query` | "fire at all?" | yes — **but it is a fleet-wide `HAVING COUNT(*)>0`**, so it never scopes to a site |
| `agent_definitions...find_dispatchable_site` | **"which site?"** | **NO** |
| `load_work_items` (Go) | "which items?" | yes (`load_work_item_actions.go:126-138`) — reached too late |

The middle one is the one that chooses, and it has no lock clause.

**This is already written up and never applied.** `213_dispatch_gate_matches_dispatcher.sql`
adds exactly this clause and names the divergence in its own header. It also assessed
the gap as *"Inert today (0 of 32 sites locked, ever)"* — which was true, and which I
falsified simply by being the first to use the feature. **A dormant gap is inert
because nobody has used the feature, not because the feature is safe.** `schema_migrations`
has no 213 row; the migration belongs to the active `bugs_open/029` dispatch-gate
lane, so I did **not** apply it as a side effect of an adoption task.

**What I did instead**, and why it is not just the same hand-holding: a 15-second
auto-defer loop against a 120-second tick. That is a control I own, scoped to one
site and one transition. It earned its keep within minutes — `build-site-planner`
finished and emitted **19 items at once**, including **3 `needs_page` and 1
`needs_rerender`**, the two types that can reach the live site. All deferred before
any tick could pick them up.

**Verified at the artefact, not at the queue:** all 29 live files fetched and hashed
against the repo — **28 identical, 1 differing (`robots.txt`)**, and that one differs
by exactly the Cloudflare Managed block documented in RUNBOOK §2 before any of this
started (491 origin vs 2327 served). Nothing this session changed the live site.

Final held state: 43 `deferred`, 11 `owned_page_review` at `needs_human_review`,
5 research items `complete`, 26 pages all still `build_status='planned'`.

**The lesson I want to keep**: I verified `deferred` was safe by reading
`workItemTerminalStatuses` in the source, then swapped to a *different* mechanism
without giving it the same treatment — I read that `locked_at` was checked in
`LoadWorkItemsAction` and stopped there, satisfied. One grep of the *gate* would have
shown it. **Checking one reader of a flag is not checking the flag**; the question is
always which reader decides, and that is rarely the one you find first.

## 2026-08-03 (later) — owner-directed fixes, and the single-page trial

### Three fixes on the owner's instruction ("increase that budget, fix broken things")

1. **`classify_and_extract` 16000 → 32000.** 16000 was already proven in production
   (one run at 6590 output tokens); 32000 is ~5x the observed maximum and matches the
   fleet's next tier. No root `ai_service` block, so the step value is live.

2. **`sites.locked_at` now actually holds a site.** Added the missing predicate to
   `find_dispatchable_site`. **Proved it by discrimination, which is stronger than
   the before/after I first reached for** — run the gate's own SQL twice, once with
   the clause and once without, against the same live data:

   | query | picks |
   |---|---|
   | old (no lock clause) | **`mortgagecalculator.co.uk`** — our site, first in line |
   | live (with clause) | `vetcomparison.uk` — ours correctly skipped |

   That is the counterfactual made explicit: without the fix our site would have been
   building at that moment. A plain "it didn't dispatch" could not have shown that,
   because **a quiet queue has two causes** — I also sampled
   `scheduled_tasks.last_triggered_at` (09:12:52 → 09:15:22) to prove the gate looked.

3. **Six live guides had a broken header logo link** (`href="index.html"` from inside
   `/guides/` → `/guides/index.html` → 404). The line below it already used
   `../index.html` and the logo's own `img src` already used `../images/`, so this was
   one missing `../`, not a design choice. Fixed, pushed, deploy named
   `Changed domains: mortgagecalculator.co.uk`, all six verified at the wire, and the
   whole-site check still reads **28 identical / 1 differing** (`robots.txt`, Cloudflare).

### The single-page trial — and my own ordering mistake

Built `/guides/first-time-buyer/index.html`. It went `planned → deployed`, served 200
at the new URL, and the old `/guides/first-time-buyer.html` kept serving. The homepage
was never dispatchable.

**What I got wrong: I built a page before the site had a stylesheet.** The page
references `/assets/css/styles.css` → **404**, and carries no `<header>`, `<nav>` or
`<footer>`. My first reading was "the rebuild produces unstyled orphan pages" — which
would have been a serious finding and is **false**. The comparison that corrected it:
the sibling `loancalculator.co.uk/guides/hidden-loan-fees.html` has nav and footer and
its `/assets/css/style.css` resolves 200. The pipeline can do this.

The cause was mine. Among the 19 items I auto-deferred were:

| item | summary |
|---|---|
| `needs_composition` | Resolve palette/layout/typography composition for the site |
| `needs_design` | **Generate site stylesheet** |

I held back the stylesheet and then built a page that needs it. **The correct order is
composition → design → pages**, and "release one page first" has to mean one page
*after* the site's design exists, not before.

Two of the page's three links (`/tools/affordability/index.html`,
`/scorecard-simulator.html`) are also 404 — and those are **not** defects either:
both are `build_status='planned'` rows, i.e. forward references to pages this build
has not reached yet. I nearly filed a hallucinated-link bug; the `pages` table
settled it in one query.

> **The lesson, and it is the same one as the lock:** I twice built a confident
> negative reading out of a partial system, and both times the fix was to find the
> *comparison* — a working sibling, a counterfactual query — rather than to look
> harder at the broken thing on its own.

### What IS a genuine defect: `bugs_open/184`

Literal `**Decision Engine**` in the hero copy — markdown emphasis reaching the
visitor as asterisks. Not an ordering artefact, and not ours alone: 3 components on
3 unrelated sites and 3 slot types. Every existing check passes it (valid HTML,
complete component, `deployed`); it was found only because a human read the prose.

### State at handoff

Site **locked, and the lock now demonstrably works** (`gate_says: NOT SELECTABLE —
held`). 42 items deferred, 11 `needs_human_review`, 6 complete. The homepage item is
`deferred`. Live site: 29 files, unchanged except the six intentional link fixes.

---

## 2026-08-03 ~11:00–11:10 UTC — the ordering canary PASSED, and chrome came from a table nobody had looked at

### What I set out to do

HANDOFF §9.3: re-run the first-time-buyer guide, which had CSS but no
`<header>`/`<nav>`/`<footer>`, and see whether it comes back WITH chrome. If yes,
composition → design → pages is confirmed end to end.

### Result: PASSED

| check | before | after |
|---|---|---|
| served bytes | 8,854 | **20,550** |
| `<header>` / `<nav>` / `<footer>` | 0 / 0 / 0 | **1 / 1 / 1** |
| `/assets/css/styles.css` | 200 | 200 |
| live at | — | **11:06:07 UTC** (deploy run `30808020578`, commit `8f921c5f8`) |

**Live site integrity re-verified after the change**: every file byte-identical to
the repo except `robots.txt` (Cloudflare, expected) and the trial page itself
(my deliberate change, mid-propagation at the time of the check).

### The thing worth knowing: where chrome actually lives

I started by reading `pages.rendered_header` / `rendered_footer` / `rendered_head`
and found them **empty for all 26 pages**. The obvious reading was "that's the bug".
It is not — **those three columns are empty for all 562 pages FLEET-WIDE**, on sites
whose served pages plainly have nav. They are vestigial. Only
`discovery_checks/check_missing_structure.go` still reads them.

Chrome comes from **`site_components`** (`slot_name` in header/footer/head), and our
site had **zero rows** there while `loancalculator.co.uk` had three. That was the
whole defect.

> **The census is what saved me.** One site with empty columns looks like a bug; the
> whole fleet with empty columns is a dead column. Same query, opposite conclusion —
> and the only difference was not putting a `WHERE domain=` on it.
> `[VERIFIED]` — `SELECT count(*) FILTER (WHERE length(rendered_header)>0) … GROUP BY domain` → 0 everywhere.

### Why the site was stuck, and the fix

`nav-updater`'s live workflow is
`populate_nav_tables → render_site_components → create_rerender_items → get_pages_for_rerender`.
We had **14 `site_nav_items` rows and 0 `site_components`** — stalled exactly between
steps 1 and 2. Ran the documented bypass:

```
./docs/agent_docs/docs024_key_docs_latest/bugfix_149_nav_membership/TRIGGER_nav_rebuild.sh mortgagecalculator.co.uk
```

COMPLETED first poll (not the 7–9 min the memory note warns about). Produced
header 2,125 B · head 8,635 B · footer 987 B, all `rendered`.

Then the single page, assemble-only (**no `reason`** = no LLM, authored copy untouched):

```
./docs/agent_docs/docs024_key_docs_latest/cta_link_integrity/scripts/049b_deploy_single_page.sh \
  849383e6-f1a9-4437-b42e-19a7ccc93c5f 62b5978e-4271-4589-8e00-4baebfc0447c mortgagecalculator.co.uk
```

### The safety check I did BEFORE running any of it

The nav rebuild ends by filing a `page_rerender` item **per page** — 26 of them,
including the homepage, the one page that overwrites live content. `get_pages_for_rerender`
filters on **`p.status`**, not `p.build_status`, and all 26 of our pages are
`status='active'` — so the homepage IS in scope. What makes it safe is one branch
further down: `rerender_single_page_action.go:565` returns empty for a page with
**zero `page_components`**, and `:168-209` turns that into `skipped:true` with no
deploy. Only `guide-first-time-buyer` has component rows (3); the other 25 have none.

**`status` vs `build_status` on `pages` is a genuine trap** — same table, both plausible,
and only one of them is what the rerender selector reads.

I still deferred all 26 afterwards rather than relying on that (0 armed restored).

**I also verified the site lock at its source rather than trusting the handoff's
reconstruction of it.** The handoff's `gate_says` query hardcodes `s.locked_at IS NULL`,
so it proves the *query* respects the lock, not the *gate*. The real gate is a SQL
string in `agent_definitions`, and it does carry the clause:
```sql
SELECT default_config->'workflow'->'steps'->'find_dispatchable_site'->'config'->>'query'
  FROM agent_definitions WHERE type='build-pipeline-trigger' AND is_active …;
-- → "… JOIN sites s ON s.id = wi.site_id WHERE s.locked_at IS NULL AND …"
```

### A one-link nav is CORRECT, not broken

The rebuilt header carries only **Home**, though there are 14 nav rows. That is
deliberate: `GetNavItems(..., NavFetchableOnly, ...)` drops items whose target has
never been deployed, because chrome ships on every page and a dead nav item is a
site-wide 404 (the `bugs_open/049` fix). `loadFetchablePageSet` always injects the
site root, so Home survives; everything else is correctly withheld until built.
**Note the cliff**: with **0** deployed pages the filter disables itself and ships the
full 14-item nav; we have exactly 1, so the filter is active. The nav grows as pages
ship, and `nav-updater` runs with `force_rerender:true`, so re-running it refreshes chrome.

### FINDING: the header CTA is validated by a DIFFERENT predicate from the nav

In the **same function**, ~70 lines apart:

- nav → `loadFetchablePageSet` (`nav_tables.go:258`):
  `status NOT IN ('deleted','archived') AND NOT (NeverDeployedPagePredicate)`
- header CTA → `loadResolverPageSet` (`resolve_internal_links_action.go:486`):
  `status NOT IN ('deleted','archived')` — **no deployment predicate**

So the nav is filtered to deployed pages while the CTA guard
(`render_site_components_action.go:172,182`) accepts a page that has never been built.
Ours took the fallback branch and picked `/tools/stamp-duty/index.html`, a `planned`
page. **Live consequence, confirmed at the wire: `HTTP 404`** on the "Get Started"
button — and chrome ships on every page, so this scales with the batch.

> **CORRECTED IN THE SAME SESSION, before it reached the owner.** I first measured
> this from the DB, got **2 of 14 sites** (`lendzy.co.uk` → `/tools/price-cap-checker/index.html`,
> never deployed) and was about to report a fleet-wide defect. **Checked at the wire
> and lendzy's target returns `HTTP 200`** — and its served homepage carries no
> `header-cta` at all. `deployed_at IS NULL` over-reports: it means "no recorded
> deploy", not "does not serve". So the code asymmetry is real and the confirmed live
> instance is **exactly one site — ours**. `[VERIFIED at the wire, not from the DB]`
> The cheap check that caught it was a single `curl -o /dev/null -w '%{http_code}'`.
> This is the same trap as `bugs_open/098` in the opposite direction.

Also still present, unchanged: **`bugs_open/184`** — literal `**Decision Engine**`
asterisks in the hero. Assemble-only re-ships stored HTML, so a rerender cannot fix it.
And `/assets/images/favicon.png` **404s** (referenced twice from the head component).

---

## 2026-08-04 ~21:40–21:50 UTC — the homepage was rebuilt over the live original, and the lock did not stop it

### What happened

Returning to run the next batch, I found the state materially changed by other sessions:

| | 08-03 handoff | 08-04 21:40 |
|---|---|---|
| chassis | v1.0.1238 | **v1.0.1251** |
| deployed pages | 1 | **2 — `index` had joined it** |
| armed items | 0 | **21** |
| `site_components` | 11:01, header 2,125 B | **re-rendered 19:41, header 2,052 B** |

`/index.html` was rerendered and deployed at **19:45:55 UTC** (`gqls/sites` `fe6b81926`),
replacing the 11,125-byte original with a 27,546-byte framework rebuild — the one page the
owner had reserved, because it is the only URL that overwrites live content.

### The correction that matters: THE LOCK NEVER LAPSED

`locked_at` has been continuously set since 2026-08-03 10:30:34, by this lane, and the §3
gate query still answers `NOT SELECTABLE — held`. **The lock held and the page changed
anyway**, because `s.locked_at IS NULL` lives only in `find_dispatchable_site` — the
**work-item dispatch** gate. A direct `orchestrate` publish to Kafka never reads it.

That is not an obscure path: it is exactly what `TRIGGER_nav_rebuild.sh` and
`049b_deploy_single_page.sh` do, and **what I used all through 08-03 precisely because it
bypasses the lock.** I used the bypass, documented that I was using it, and still wrote
"nothing is queued that can move without you" — true of the queue, false of the site. The
`bugs_open/191` lane needed a live site to verify their fix; ours was the reproduction I
had named in the bug file.

> **`[WRONG CALL]` — "Site locked, 0 armed, nothing can move." Two true measurements, one
> false conclusion.** Both readings were correct and neither covered the direct-dispatch
> door. The cheap check I never ran: **`git log` the artefact**, not the lock —
> `git log --format='%h %ci %s' -- mortgagecalculator.co.uk/index.html`, which named the
> rerender in one line. Logged in `WRONG_CALLS.md`; landmine appended for the lock itself.

### Assessing the damage before reacting — and one false alarm of my own

First integrity sweep reported **all 33 files differing**. That was **my check breaking, not
the site**: the session scratchpad had been relocated by another lane's commit, `curl -o`
failed silently, and every `sha256sum` compared against a missing file. `curl -sf` plus a
fetch-failure branch fixed it. **A comparison against a file that does not exist reports
"differs" for everything — which reads exactly like catastrophe.**

Real state, once measured properly: **33 files, 1 differing (`robots.txt`, Cloudflare).**
Nothing but the homepage had changed.

Then the functional question. The rebuild's markers looked alarming (`css/style.css` →
`assets/css/styles.css`, 8 × `site-header`), but the honest read is narrower:

- **No calculator was lost.** The old homepage had **0** `<input>` and **0** `id=` —
  verified with two independent tools, because a mortgage site's homepage scoring zero
  form fields is exactly the sort of number that is usually a broken grep. It was a
  **28-link landing page**; the calculators are separate files (`repayment.html`,
  `simple.html`, …), untouched and still 200.
- **The rebuild was technically clean** — and it is the proof that `191`'s fix works:
  **no `header-cta` element at all**, because with no deployable target the gated template
  now renders no button instead of a 404. Zero literal markdown. v1.0.1251 carries
  `LoadChromeLinkPolicy` (pod-grep: 2).
- **The real cost was navigation: 28 internal links → 4.** The front door stopped pointing
  at the calculators, because the platform correctly refuses to link unbuilt pages. Nothing
  was broken; the site was just **rebuilt ahead of its own content**, which is the same
  ordering lesson as 08-03 one level up.

### Resolution

Owner chose restore. Put back `825a36994` — the last owner-approved state (import + the two
deliberate crawl-link fixes; the 08-03 logo fixes touched guides, never `index.html`).
Committed `59e4eb9ae`, pushed **rebased, never merged** (a merge makes the deploy's
`git diff HEAD~1 HEAD` drop the domain while still going green). Live in ~40 s: 11,125
bytes, 28 links, all five calculator pages 200, whole-site sweep back to 1 of 33.

Owner also chose to defer **only** the two site-wide `needs_rerender` items (priority 99
and 30, "26 pages missing header/footer") and leave the 19 audit findings armed for
whoever owns them. Verified afterwards that **no armed item targets the homepage** — the
one whose `item_key` matched `%index%` is the guide's link to
`/tools/affordability/index.html`, i.e. the target URL, not the page. Checked rather than
assumed.

### `bugs_open/191` — filed by this lane 08-03, FIXED by another session 08-04

Three commits and three council rounds (`d32692882`, `6ae203679`, `007814ff1`), live in
v1.0.1251. They took **fix candidate 2**, the structural one — a shared
`LoadChromeLinkPolicy` replacing `loadResolverPageSet` — not my smaller candidate 1, and
handled the first-build escape the bug file flagged. **I went to implement candidate 1 and
found it already done**; `git log` on the file was the whole check, and running it first
is the cheapest step in this workflow.

---

## 2026-08-05 ~11:20–13:05 UTC — three guides live; the arithmetic checker EXISTS, was falsely refusing ratio tools, and now certifies all 12 originals

### Guides batch (owner-approved): DONE

`guide-remortgaging`, `guide-buy-to-let`, `guide-negative-equity` — armed, built through
the real queue (unlock + auto-defer backstop from RUNBOOK §10c), all three live with
chrome (hdr=1, ftr=1, ~19.6–20.6 KB), all three ORIGINAL file-form URLs still 200. Site
re-locked. The queue claims by PRIORITY within a site once selected — `literal_markdown`
(pri 5) went before our guides (25–27) — not the fleet FIFO, which chooses the SITE.

Homepage secured first (owner choice): its 4 rebuilt `page_components` deleted (backed up
in scratchpad + rebuilt HTML in git `fe6b81926`), restoring zero-components ⇒
assembles-to-nothing ⇒ `skipped:true` protection **by construction**.

### The owner asked: "is the framework in control, or is content authored and tools fixed?"

Measured answer: **the framework controls almost nothing yet.** 23 of 26 pages are
`planned` rows serving nothing. Everything a visitor uses is the hand-authored original —
homepage, 9 guides at old URLs, and all 12 calculators driven by `js/calculators.js`
(3,622 bytes, 5 functions). Owner then chose **complete adoption, with an arithmetic
checker created FIRST** and an explicit instruction to search for prior art.

### The prior-art search PAID — the checker exists (TL-038), do not build a duplicate

- **`computed_values`** (Tier-4 check type, browser-runner): drives a tool and asserts the
  EXACT text of every output. Council-APPROVED round 2, 8 tests, proven able to fail.
  **Verified live in the current adapter: `INSTALL_GATE.sh` PASSED 2026-08-05.**
- **`toolgolden.py`** (loancalculator lane): captures a working tool's answers; `--compare`
  proves a rewrite didn't move them; `--emit-criteria` hands enforcement to the platform.
- Enforcement path: fence in the tool's PLAN (`doc_plans` row, `subject_type='tool'`,
  `subject_key`=pages.name) → `load_doc_context` → `doc_context.criteria_json` →
  `tool-acceptance-agent` on the normal schedule. **Live precedent:** `tool-loan-vs-savings`
  has a computed_values fence installed (authored 2026-08-05 by `staged_component_build`).

### FALSE CONVICTION found and fixed: uniform vectors cannot see a ratio

First capture refused `investor.html`: *"reacts, but output is identical for every input
value — arithmetic ignores its inputs."* **The tool is correct.** Its two calculators are
pure ratios (yield = rent×12/price, LTV = loan/price) and the harness scales ALL fields by
one factor per vector — a quotient is scale-invariant, so the output CANNOT vary. The
harness's question ("does output depend on input?") was unanswerable for ratio arithmetic
as posed.

Fix (in their instrument, contributed not forked): fourth vector `asym` — per-field
deterministic factors `[1.7, 0.6, 2.3, 1.1, 3.1, 0.45]` cycled by document order; gate B
now also diffs defaults↔asym; presence guards keep pre-asym goldens loading, comparing
(with a printed NOTE) and emitting.

**Non-regression proven the way TL-038's own landmine demands** (re-capture the corpus and
diff, because "a single green run cannot see" a drive-heuristic change):
`--compare GOLDEN_2026-08-03b` over all 11 loancalculator pages → **11/11 MATCHES**.
Then our re-capture: investor vary 0→1 (certified), portfolio vary 4→8 (richer), **golden
written for 12/12** → `acceptance/GOLDEN_2026-08-05_original_tools.json`.

### Emission: 1 of 12 — and the 11 refusals ARE the rewrite contract

`--emit-criteria`: only `fact-finder` emitted (4 checks, 12 assertions). All 11 others
refused with one uniform finding: **every calculate button on the original site has no
`id`** — capture can hold the element, but a fence must NAME it. (Same shape as
loancalculator's 3 refusals.) So the recreation contract is: **preserve every input/output
id verbatim; give every button an id.** Then `--compare` proves arithmetic parity against
the golden, fences emit from the id-complete rebuilds, and the platform enforces the
answers forever.

Coverage caveat, recorded not hidden: the harness presses ONE button per page, so
investor.html's golden covers the yield calculator only — the LTV half sits behind the
second button and is uncovered. `[KNOWN GAP]`

### Still open (tasks #4/#5)

How the 12 fences get installed for the RECREATED tools (precedent format known; check the
`staged_component_build` lane before writing PLAN rows — they are actively authoring
these), whether tool-recreation's prompt needs the id-contract stated, then arm the 12
recreations and verify each against its fence.

---

## 2026-08-07 → 08-08 — the three zero-output recreations DIAGNOSED; validator fixed (committed, inert); portfolio was held by the fabrication gate

State re-verified at session start (08-07): queue drained (0 triaged/claimed), site
LOCKED (21:19 08-05 lock note intact), 9 rebuilt tools live at the wire with chrome,
3 confirmed 404, originals intact — the §10f sweep flagged `data/latest-news.json`
which was only a stale local checkout (77 commits behind; the news feed auto-commits
daily ~08:10 UTC; `git pull --rebase` cleared it). The `**` hits on rebuilt tools are
JSDoc `/**` blocks in inline JS, NOT bugs_open/184.

**Why 3 of 12 recreations produced nothing (items `complete`, 0 components, 404):**

1. **tool-overpayment + game-fact-finder — convicted by a validator false positive.**
   `validate_tool` (action `validate_page_content`) failed each with "1 blockers":
   `checkPlaceholderPatterns` substring-matches `[name` against the WHOLE HTML
   including inline `<script>`, hitting `fields[name]` (overpayment, 12:21:33) and
   `([name, val]) =>` destructuring (fact-finder, 12:59:20). A third conviction the
   same day on idea.uk (`querySelector('input[name=...')`). Evidence: `agent_error_log`
   `step_name='validate_content'`, `context.issues` type `placeholder_text`. The
   recreate_tool LLM calls had SUCCEEDED (9–12k output tokens, not cut) — the finished
   tools were discarded.
2. **tool-portfolio — held by the bug-020 fabrication gate** (`needs_human_review`
   item 12:55:38 "appears to INVENT data"). Which signals fired is UNRECOVERABLE:
   orchestration row purged, pod logs rotated with the 08-06 roll. Original
   portfolio.html is self-contained (no fetch/XHR, users type their own properties),
   so likely a false positive (Tier A declaration match, or `dataSourceIsExternal`'s
   loose reading of the analysis), but that is [INFERRED] — a fresh run will
   regenerate the signals; read `check_fabrication` output from the orchestration row
   BEFORE it is purged.
3. **Secondary defect found while answering the 090's open question:** the LIVE
   `tool-recreation-handler` row carries `validate_tool`'s `error_step` INSIDE
   `config` where the engine never reads it (`processor.go:433` reads step level);
   the seed has it at step level (= "validation advisory, save anyway"). So a
   validation failure discards the recreation and the item still completes. Filed as
   defect B in `bugs_open/218`; needs a deliberate decision, NOT hot-fixed.

**Process trail:** 090 filed BEFORE asserting (intake `0de6e0e4`, run `86721efd`) —
verdict **UNVERIFIABLE / iteration-cap**, sub-claims confirmed, two named missing
evidence items supplied first-hand in `bugs_open/218` (the ruling's stated-substitution
hatch, stated there). Fix committed `201350e23` (strip script/style bodies from the
placeholder scan only; tests carry the three convicted snippets + guard-survival cases;
mutation run proves the tests bite). Council submission `a9ffed15` (Council-Submitted
trailer on the commit) — **verdict unread as of this writing, read it**. §9 pattern
added to 016b; index row 218 added.

**Chassis rolled TWICE during this work (v1.0.1262→1263, pods 08-08 08:54 UTC) — both
predate `201350e23`, so the validator fix is NOT live.** Wait for a post-`201350e23`
roll; prove at the pod: `strings /app/agent-chassis | grep -c stripScriptAndStyle` ≥1
every replica, and a negative control.

**Wrong turn logged:** first diagnosis query filtered `orchestration_states` by
`created_at > '2026-08-05'` + regex and returned only unrelated feed runs — read as
"runs purged". The refined check (whole 09:00–14:00 window, no regex) showed the
purge was real for that day, but the first query could not have distinguished
"purged" from "my filter missed them": the count you keep is not a census.

**Also noted, unowned:** the three dead items' `site_work_items.result` payloads
describe the WRONG artefacts (overpayment's = a stamp-duty calculator; fact-finder's
= a legal-disclaimer page proposal). In bugs_open/218's tail. Do not trust `result`
on this path until someone looks.

### 2026-08-08 later — council round 1: REVISE, and the objection was RIGHT

> **CORRECTED 2026-08-08:** the entry above says the fix "strips script/style
> bodies". That was round 1 (`201350e23`) and it drew a REVISE: `reuse_agent` +
> `prior_art_librarian` caught that `datahelpers.ExtractAssertionText` — the
> claims checks' prose scope, called TWO LINES below my edit — already solved
> "read prose, not markup", with a real HTML parse that also excludes
> code/pre/head/attributes. Round 2 (`b75f36601` + gofmt `f51ac6af8`) reuses it
> and deletes the stripper. Wrong call logged in `WRONG_CALLS.md` (the cheap
> check: grep datahelpers before writing a helper). One reuse trap found and
> pinned: `<no value>` parses away as markup, so it stays on the raw document
> or the pattern goes silently inert. Resubmitted under the same trail
> `a9ffed15`; **round-2 verdict unread as of this writing.** Consumers named
> from live definitions: page-build-handler, content-reviewer,
> tool-recreation-handler, report-builder. bug_historian's gating point (failed
> validation → silent complete, no escalation) = 218 defect B, related to 034/040.

### 2026-08-08 later still — round 2 also REVISE; round 3 in; defect B routed

Round 2 objections were about the PLAN RECORD, not the code (unchanged since
`f51ac6af8`): (a) my "verification edit" was mislabelled config_change AND stepped
into two documented landmines — label-selector pod coverage and grep -c printing
nothing on zero — both now fixed in 218's verify block; (b) "no evidence the other
three consumers don't rely on attribute/code-context detection" — answered by
MEASUREMENT: all-history census = 46 convictions, 43 prose (preserved), 3 = the JS
false positives, zero attribute/code true positives ever; (c) bug_historian gated
again on "filed ≠ routed" for defect B — answered by routing it: 090 intake
`315f7f88`, dispatch-loop run `c56b691d`. First 090 attempt FAILED with invalid
JSON — escaped double-quotes in the symptom text reach the script's dollar-quoted
JSON as literal backslashes; write symptoms with NO quote characters (this cost one
correlation id, `741bf434`, which has no row). Round 3 submitted under the same
trail; verdict unread.

### 2026-08-08 evening — round 3: 9/11 approve, still REVISE; the head escape was REAL and is fixed; round 4 in

Round 3's editquality objection was a genuine code catch, not plan-record noise:
narrowing to ExtractAssertionText silently dropped <head> — a placeholder in
<title> or a meta description (visitor-visible prose the OLD scan covered) escaped.
Fixed in `35889819c`: headProseBlocks scans title + description/og/twitter meta
content alongside body blocks; JSON-LD deliberately unread (code-shaped — the very
collision class being fixed); mutation dropping it fails 6 cases. Census re-checked:
zero of the 46 historical convictions were head-context, so nothing was lost in the
window. The gating seat escalated to "routing defect B is not a fix" — answered in
round 4 with the process fact: defect B is under ACTIVE diagnosis (`c56b691d`,
status diagnosing), the 090 coverage rule forbids a second thread patching a target
mid-diagnosis, and the right fix is a design decision (save-anyway vs
cannot-complete) that the diagnosis run should land, not a rider on this patch.
Lesson worth keeping: **a reviewer pool with fresh eyes each round keeps finding
real things — round 3's "new objection" was the best catch of the three rounds.**

### 2026-08-08 close — round 4 APPROVED

"Approved with 2 advisory objections, none high-severity." Trail: 3× REVISE
(reuse → plan-record+census+routing → head escape), each answered same day, final
code `35889819c`. The Council-Submitted trailers on all four code commits resolve
to this approval at 098 time — no amends, forward-only held throughout. Defect B's
diagnosis (`c56b691d`) still `diagnosing` at close; its verdict is the next
session's first read alongside the roll check.

### 2026-08-08 afternoon — fix confirmed LIVE; defect-B diagnosis died, its refutation verified first-hand

**Validator fix is live.** Chassis rolled to v1.0.1264 at 13:08 UTC (both
replicas). Pod-grep, same exec, both pods: `headProseBlocks` (round-4 ADDED
symbol) = 2, `stripScriptAndStyle` (round-1 symbol round 2 REMOVED) = 0. Both
tells correct on both replicas — the binary carries the full round-4 fix, not
round 1. (The handoff's suggested positive control `ExtractAssertionText` was
weak — it pre-exists in datahelpers with other callers; `headProseBlocks` is
unique to `35889819c`.)

**Defect-B 090 run (`c56b691d`) produced NO verdict** — work item `8f460338`
`failed` at `call_diagnoser` 10:27 UTC, `result={}`, five bundles then
iteration cap. But its final bundle's hypothesis-under-test REFUTED the filed
mechanism, and this session verified the refutation first-hand (all three
legs): `coordinator.go:3529-3537` falls back to `step.Config["error_step"]`
(nested key IS read); live row has `next_step` = `error_step` =
`save_sections` (paths converge — no divergent fail path exists);
the real discard is `save_sections` reading `validation_result.clean_html`
(success-only field) and `save_page_sections_action.go:321-330` reporting the
empty input as SUCCESS (`skipped: true, sections_saved: 0`) → happy-path
completion with 0 components. `bugs_open/218` defect B corrected in place
(mechanism refuted, phenomenon stands, fix candidates re-ranked: candidate 2
"restore step-level error_step" is a no-op); WRONG_CALLS entry appended
(cite-the-arm shape, plus: converged next/error steps make routing claims
unobservable downstream).

**Consequence for the re-runs:** with the validator fixed, overpayment and
fact-finder should PASS validation and save normally. If a re-run fails
validation for a REAL reason, expect the same silent discard (item complete,
0 components) — that door is still open until defect B's design call lands.

### 2026-08-08 evening — the three re-runs LANDED; 12/12 tools live

Three fresh `needs_tool_recreation` items filed at `triaged` (cloned spec from
the terminal-complete rows; dedup index permits key reuse once the old row is
terminal): `aaaa8861` tool-overpayment, `eac0c3bb` game-fact-finder,
`c21a1b32` tool-portfolio. Unlock window 15:34–15:57 UTC; §10c backstop every
15s (foreground-tested first) — **deferred nothing the whole window** (78 ticks
of `UPDATE 0`); killed the moment the batch settled (§10g); site re-locked
15:57.

Outcomes, all evidence captured live to
`acceptance/EVIDENCE_2026-08-08_rerun_3tools_orchestration_capture.jsonl`
(44 snapshots; the orchestration rows purge ~a day, this file is the durable
copy):

- **tool-overpayment**: validate_tool PASSED (`validation_issues: []`,
  clean_html present) — the 218 defect-A fix works on the exact case that
  motivated it. Then `deploy_page` FAILED with `CHILD_ORCHESTRATION_FAILED`
  ("workflow completed but its result could not be delivered to the parent
  (failed_transient)") → `complete_error`. **The known spawn→call handshake
  race, and the child had actually deployed** — the page was on the wire and
  byte-identical to the repo. Item reads `complete`; nothing to re-run.
- **game-fact-finder**: full happy path. Validation clean, fabrication check
  `fabricated: false` with no signals. 4 components. `build_status` even went
  `deployed` for once.
- **tool-portfolio**: full happy path — **the 08-05 fabrication conviction did
  NOT reproduce**: `{"fabricated": false, "signals": null, "tier": "",
  "detail": ""}` — a clean pass, not a borderline. Whether the 08-05 conviction
  was a true positive is now unknowable (its signals purged before anyone read
  them), but the artefact standing today passed the gate. The 08-05
  `needs_human_review` item (`aca92097`) is now MOOT — its subject artefact was
  discarded and has been superseded by this clean run; closing it is an owner
  call, flagged in README.

**Wire verification**: all three 200 at the FULL `/index.html` form
(32,888B / 17,679B / 35,998B), chrome present (header/nav/footer), correct
tool on each page, zero cross-wiring. §10f sweep across the whole domain:
exactly one line, `robots.txt` — originals intact.

**Misstep (cheap, self-caught, but the shape matters):** my first wire-check
used bare directory URLs (`/tools/overpayment/`) and read 404 — as did
`/tools/repayment/`, one of the 9 verified live 08-07, which briefly read as a
site-wide regression. **This host does not resolve directory URLs to
index.html anywhere except the root.** The RUNBOOK sweep (§10f) was never
wrong — it fetches full file paths; the bare-URL form was my invention. Check
added to RUNBOOK §10f. The tell that unpicked it: `/` served 11,125B — the
original homepage byte count — so the site could not be "down".

Remaining on this lane (unchanged from handoff §2): the id-alignment batch
(08-05 §7 path a), fences to `staged_component_build`, arithmetic verification
still **0 of 12 proven** — now 12/12 candidates live to verify against.

### 2026-08-08 late — id-alignment batch: 9/12 saved+live; the comparator finally speaks, and it says FORMULA divergence

Mechanism verified BEFORE spending: the `recreate_tool` prompt renders
`spec.interactive_features[]` under "Mandatory Behaviour Requirements …
OVERRIDE" — so 12 fresh items each carried an "ID Contract" entry naming that
tool's golden ids verbatim (generated from the golden; inputs = `controls`
keys, outputs = tracked ids minus controls) plus the button rule (first
button, inline onclick — that is what toolgolden actually presses).
Window ~16:10–16:58 UTC, §10c backstop (deferred nothing), monitor captured
per-tick evidence → `acceptance/EVIDENCE… (idalign jsonl in scratchpad;
summarised here)`. All 12 items `complete` in ~45 min.

**Saved + deployed (9):** repayment, stamp-duty, overpayment, simple,
bridging-loan, equity-release, investor-index, fee-analyser, rate-forecaster —
fresh components 16:12–16:52, §10f sweep clean (robots.txt only).
**ID compliance proved by the comparator itself: the one-sided "wholesale
rename" noise is GONE for these 9** — golden ids now resolve on both sides.

**Refused/discarded (3), each keeping its previous live version — nothing lost:**
- `tool-affordability`: SHRINK REFUSED (hero 14,907→6,495 chars, floor 50%,
  `bugs_open/178` guard). Payload preserved; id-complete 5/5.
- `game-fact-finder`: prune REFUSED (re-confirmed 1 of 4 stored sections).
  Payload preserved; id-complete 19/19.
- `tool-portfolio`: fabrication gate `tier:declaration` on the payload's ONLY
  matching line — `// In-memory portfolio store (no fabricated data — starts
  empty)` — **a FALSE POSITIVE, judged and evidenced** (the handoff's morning
  instruction "judge true/false positive THEN" is hereby answered). Filed as
  **`bugs_open/222`** (regex at `check_tool_fabrication_action.go:91` has no
  negation awareness; the recreate prompt's own Data Integrity section
  manufactures the echo). New review item `3d11e960` filed by the gate; the
  morning's `aca92097` now has a strong same-class prior. Payload preserved;
  id-complete 15/15.

**Comparator (`acceptance/COMPARE_2026-08-08_id_aligned_9of12.txt`), the real
milestone — arithmetic is now COMPARABLE, and it diverges on FORMULA, not ids:**
- Judge `after_press` only: `after_input` is systematically incomparable —
  several rebuilds live-compute while the originals hold stale display from
  the previous vector, so after_input diffs are behaviour, not arithmetic.
  (Corollary found en route: the 08-05 §7 claim "several originals computed
  live on input" is WRONG against the golden — only fact-finder does.)
- REAL after_press divergences: **repayment** (£1,390 vs £1,169.18 monthly on
  defaults — the original does not compute textbook monthly amortisation),
  **overpayment** (£12,949 vs £24,505 interest saved), **bridging-loan**
  (£20,225 vs £19,180.99 interest), **investor** (yield 5.76% vs 1200% — the
  rebuild reads rent as annual where the original reads monthly, or v.v.),
  **rate-forecaster** (£1,390 vs £1,111.66 — same repayment-formula class),
  **simple** (half-rate vector only: £765 vs £739.94; defaults/double match).
- **stamp-duty: arithmetic is SELF-CONSISTENT** — £7,500 standard vs golden
  £2,500 FTB-relief are BOTH correct for their buyer type; the rebuild
  reordered/renamed the `buyerType` select options so the driver lands on a
  different type. Fix is OPTION-SET alignment (text+values+order verbatim),
  not maths. The long-standing "£0 after press" mystery dissolves the same
  way: the old rebuild's select order made the driven selection a no-op.
- **fee-analyser: shows '—' even after press** despite an inline-onclick
  first button (contract honoured, verified at the wire) — likely its input
  validation rejects a driven value; needs a local drive with console.

**Next batch contract (formula alignment, ALL 12 in one window):** ids
verbatim (proven) + select options verbatim (text, values, order) + "copy the
CALCULATION LOGIC verbatim from the original source — same formula, same
rounding, same units; reference-only reading invites reimplementation and the
comparator catches it" + results populate only on press (except fact-finder's
live score) + portfolio's comment-style clause (`bugs_open/222` workaround) +
delete affordability/fact-finder components first (per §10e precedent) so the
shrink/prune guards compare against nothing.

---

## 2026-08-08 — cross-lane notice from the bugfix_210 lane (not this lane's author)

**A new mechanism can silently swallow your re-recreation dispatches, and the swallow reads
like ordinary dedup.** bugs_open/210's fix (committed 2026-08-08, inert until the next roll;
register PBP-038) parks a page after 3 content-failed generic builds behind an OPEN
`page_build_failed` item that holds the same `(site_id, 'needs_page:<page>')` dedup slot your
`needs_tool_recreation` items use (e.g. `needs_page:tool-overpayment`). While a park is open,
your emitter's insert returns "already covered" and no item is created. Check before
diagnosing your dispatcher:
`SELECT item_type, status, spec->>'skip_reason' FROM site_work_items WHERE site_id='<site>' AND item_key='needs_page:<page>' AND status='needs_human_review';`
A hit means the generic pipeline is repeatedly failing on that page — close the park (or fix
the cause; a successful deploy auto-closes it) and your dispatch works again. Full entry:
LANDMINES.md § "An `insertWorkItem` false return on a `needs_page:<name>` key may be a PARKED
page". — bugfix_210 lane

> **CORRECTED 2026-08-08 (same night, owner question exposed it):** the entry
> above's central claim — "6 tools diverge on FORMULA … rebuilds write textbook
> maths where the originals compute their own" — is **WRONG**, and the proposed
> "copy the original's calculation verbatim" contract with it. What caught it:
> the owner asked "explain why it's all different", and the first hand check
> (golden repayment £1,390 IS textbook-correct; hand-driving the rebuilt page
> with the same inputs returns £1,389.58) contradicted the story. The real
> mechanism, verified per tool below: `toolgolden.py` DRIVE_JS derives every
> driven value by SCALING THE PAGE'S OWN markup `value` attributes (and drives
> a fixed 1000 into fields with none) — it goldens a page against ITSELF, so
> `compare_rebuilt.py` drove the original with the original's defaults and the
> rebuild with the rebuild's different/absent defaults. Full account + per-tool
> arithmetic: `WRONG_CALLS.md` 2026-08-08 (differential-test entry) and the
> 08-08b handoff. **Zero demonstrated arithmetic defects in any rebuilt tool.**
> Surviving REAL findings: (1) bridging-loan — identical defaults, genuinely
> different interest model (original: retained-interest gross-up
> `gross = net/(1 − fee% − monthlyRate%×months)`, the standard bridging quote
> structure; rebuild: a compound variant) — a correctness judgement for the
> improvement loop, not a copy-the-original fix; (2) stamp-duty select-option
> ORDER (both sides' SDLT arithmetic verified correct for what each actually
> selected); (3) the comparator itself cannot prove input-equality until it
> REPLAYS the golden's recorded absolute fill plan (sel/action/value are
> already recorded in the golden for exactly this purpose) instead of
> re-deriving a drive from the page under test.

### 2026-08-08 night — OWNER RULING (verbatim intent, recorded same hour)

The owner, on reading the formula-divergence report: **(1) do NOT copy an
original's calculation method if it is wrong — improve every tool to the best
it can be; the experience/tool loops own that improvement. (2) The arithmetic
checker's job is to prove results don't differ (on identical inputs) and to
catch wrong results. (3) The site need not stay locked — especially not to
preserve tools reporting wrong results. (4) All content and tools are to be
controlled from the framework so they can be improved.** Consequences: the
byte-frozen "originals are the contract" posture ENDS for tools (the golden
remains the drive-plan source and a regression reference, not an arithmetic
authority); the site lock is RELEASED (done, this session); fidelity-to-wrong
is out, correctness is the bar.

### 2026-08-08 night (later) — OWNER RULING addendum: both-right → supply BOTH

Owner, verbatim intent, on the bridging-loan class (two models, each defensible):
**"If the two calculators are 'right' in different ways then we can explain it
and supply both calculators for each task — maybe as a separate, but well
flagged and signposted page (for those that are interested or need one or the
other)."** So the routing for a genuine model divergence is now three-way, not
two-way: rebuild wrong → fix; original wrong → improve past it; **both right in
different ways → keep the primary tool, and supply the alternative model as a
separate, clearly signposted page explaining when each applies.** For
bridging-loan specifically: retained-interest gross-up (the structure lenders
quote) and the compound-interest variant are both legitimate answers to
different questions — candidates for exactly this treatment. Goes through the
framework as everything does (ruling §0.4); the improvement loop owns the pages.

## 2026-08-08 (late night) — replay comparator built; all 9 id-aligned tools JUDGED on identical inputs

**The checker fix (handoff §3.1) is done and the answer is in.** `compare_rebuilt.py`
rewritten: it now REPLAYS the golden's recorded fill plan (`sel`/`action`/`value`)
into the rebuilt page — fills set the literal recorded value and read it back,
selects go BY VALUE never index, checkables are set not toggled. Press still uses
toolgolden's heuristic (no original press button carries an id — `pressed.sel` is
null on all 12). Verdicts: VERIFIED / DIVERGED / NEEDS-JUDGEMENT / DOMAIN-DIFF
(one side refused an input the other accepts — validation, not arithmetic) /
REPLAY-FAIL (an input did not take; tool NOT judged). Rounding-equal = within half
a unit of the coarser side's displayed precision. Ids whose text hits the 200-char
fingerprint slice are listed for eyeball, never machine-judged (the two sides
truncate at different points of e.g. an amortization table). Harness validated
first on repayment per the handoff's instruction: it reproduces the 08-08 hand
drive exactly (rebuilt £1,389.58 vs golden £1,390 on 250000/4.5/25, rounding-equal).
Report of record: `acceptance/COMPARE_2026-08-08_replay_absolute_inputs.txt`.

**Per-tool judgement (identical inputs, evidence inline):**

- **simple — VERIFIED.** All 4 vectors rounding-equal (£1,111.66 vs £1,112 etc.).
  The old "half diverges" was the derive-from-defaults artefact; absolute replay
  of 12.5 years passes — the rebuilt simple tool ACCEPTS fractional years.
- **repayment — VERIFIED where both answer + a domain difference.** defaults/
  double rounding-equal on every display id AND on the visible amortization rows
  (£11,136.70 vs £11,137…). asym/half (terms 57.5/12.5y): rebuild refuses —
  `Number.isInteger(years)` in its validation (curl-verified). Original accepts
  fractional terms. Stricter domain, defensible; not an arithmetic defect.
- **overpayment — VERIFIED in substance.** Sole diff across all vectors:
  `dispYearsEarlier` golden `0` (years) vs rebuilt `6 months` — same fact,
  rebuild reports finer units.
- **investor — VERIFIED in substance; golden's 0% is a HARNESS ARTEFACT.** Yield
  ids agree identically. `ltvResult`: golden 0% on every vector because the
  original page has TWO calculators with TWO buttons and toolgolden's PRESS_JS
  presses only the FIRST ("Calculate Yield") — the LTV section was never pressed
  during capture. Replayed inputs ltvLoan 225000 / ltvPrice 300000 = 75.0% =
  exactly what the rebuilt shows (asym 247500/690000 = 35.87% → 35.9% ✓).
- **equity-release — mostly verified; same single-press artefact + one real
  table difference.** debt10/20/30 golden £0s: the original's "Project Future
  Debt" is a SECOND button never pressed by capture (read in its HTML). Rebuilt
  projections are penny-exact compound (100000×1.065^y = 187,714/352,365/661,437).
  Real difference: max-LTV at 65 — original step table `>=65 → 0.31` (£124,000),
  rebuilt linear `0.20+(age−55)×0.01` → 0.30 (£120,000). Both self-described
  industry approximations; the original's own comment says "65: ~30%" while its
  code uses 0.31. Improvement-loop judgement; not wrong-vs-right.
- **stamp-duty — REBUILD RIGHT, ORIGINAL WRONG in the FTB £500–625k window; plus
  a spec gap.** Replay first hit honest REPLAY-FAIL: rebuilt renamed option
  VALUES (`ftb`→`firstTime`, `next`→`homeMover`) — the id contract pinned element
  ids, not option values. Hand-replay with the value mapped: defaults/double/half
  (£350k/£700k/£175k FTB) match EXACTLY. asym £595k FTB: golden £14,750 vs
  rebuilt £19,750. Post-April-2025 rules: FTB relief nil to £300k, 5% to £500k,
  LOST ENTIRELY above £500k → correct = standard rates = £19,750 = rebuilt. The
  original implements a no-regime hybrid (300k nil + 625k cap; its own comments
  hedge "rules vary… avoid under-quoting" then under-quote by £5,000).
- **rate-forecaster — BOTH RIGHT, DIFFERENT MODELS (ruling §0.5 class).**
  Original = 3-phase rate path: years 1–2 at rate1, years 3–5 at rate2 on
  remaining balance/term, year 6+ at rate3 on the balance after year 5 (read in
  `calcForecast`; reproduced to the penny: 1389.58/1525.78/1286.39 vs golden
  1390/1526/1286). Rebuild = each rate from day one over the full term
  (textbook-exact: 1535.22/1251.56). The original's model is the more
  product-realistic "forecast"; the rebuild's is a rate comparison. Candidate
  for the both-calculators treatment.
- **fee-analyser — BOTH RIGHT, DIFFERENT DEFINITIONS (ruling §0.5 class).**
  Original `tcTotal` = total repayments over the deal + fees (25y amortization:
  1076.77×24+999 = 26,841.44 = golden exactly). Rebuild = simple interest + fees
  (200000×4.19%×2+999 = 17,759 exactly). "Cash out the door" vs "cost excluding
  principal you keep as equity" — a definitional split worth explaining to users.
- **bridging-loan — BOTH RIGHT, DIFFERENT MODELS (known since 08-08 evening,
  now quantified on all 4 vectors).** Retained-interest gross-up vs compound
  variant; each internally consistent (fee = 2% of each side's own gross).

**Scoreboard per ruling §0.2: on identical inputs, ZERO rebuilt tools compute a
wrong number. ONE ORIGINAL does** (stamp-duty FTB £500–625k). Three tools split
on legitimate model/definition grounds → the §0.5 both-calculators treatment.
Two harness artefacts found and understood (single-press blindness; 200-char
truncation). One spec gap (option VALUES belong in the id contract — stamp-duty
re-file should pin them; emitted criteria also select by value, so fences hit
the same wall until aligned).

## 2026-08-09 — supply-both items FILED + legislation watch answered by SEEDING, not building

**Improvement-loop routing (owner "yes, put them through", 08-09): five items
filed and armed (`triaged`), insertion order add_tool first so FIFO builds the
companion pages early.** Row identity verified after insert:
`0dc7a786` add_tool `tool-bridging-compound` · `0c529013` add_tool
`tool-rate-scenarios` · `c9f810a3` recreation `tool-bridging-loan`
(retained-interest gross-up primary) · `df5c5935` recreation
`tool-rate-forecaster` (3-phase path primary) · `ba68c674` recreation
`tool-fee-analyser` (both cost figures, one page, new ids tcTerm/tcOutlay).
Each recreation spec embeds the model FORMULA and a worked check the
implementation must reproduce (bridging 200000/1.0/10/2.0 → G=227272.73;
forecaster 250000/25/4.5/5.5/3.5 → 1389.58/1525.78/1286.39; fee-analyser
200000/4.19/2y/999/25y → outlay 26841.44, true cost 17384.79) — a wrong model
now fails a stated check instead of reading as plausible. ID contracts copied
verbatim from the 08-08 batch. Cross-links via the framework's own cross-link
items (add_tool `related_pages`), NOT hardcoded URLs (bugs_open/029's lesson).
Dedup slots were free (prior `needs_page:` holders all `complete`; no
`page_build_failed` parks — bugfix_210 check done). `add_tool` path is LIVE:
12 complete fleet-wide; the 2 deferred rows on this site were parked by
triage, not a dead pipeline.

**The owner's legislation question — the scheduled task ALREADY EXISTS; what
was missing was this site's enrolment.** `scheduled_tasks.evidence-freshness`:
daily, enabled, ran 08-09; drives `refresh_evidence_base` which (V4) re-runs
`source.sql` facts mechanically and (V5) re-verifies CITATION facts by
re-fetching the source URL and matching the stored verbatim quote
(`evidence_citations.go` → `datahelpers.QuoteFoundInText`, normalising curly
punctuation/thousands/whitespace/case on both sides); drift → `stale_evidence`
/ citation_lost work items; `staleness_days` forces periodic re-attestation.
The fact schema already carries legislation (leopardessconsulting cites
legislation.gov.uk). mortgagecalculator.co.uk simply had NO evidence base row.

**Seeded it (site_specs aspect `evidence_base`, pinned, is_current): 4 SDLT
facts** citing the GOV.UK residential-rates page with quotes extracted
PROGRAMMATICALLY from the fetched HTML (never retyped — the emission-rewrite
trap): `sdlt-standard-bands` (12% top), `sdlt-ftb-nil-rate` (300000),
`sdlt-ftb-relief-cap` (500000 — THE fact the original stamp-duty tool
violates), `sdlt-additional-surcharge` (5). All carry `writer_line`s;
`writer_block_managed: true`; governing_rule states "a calculator is a claim
about legislation". **CHECK NEXT SWEEP (within ~24h): a `citation_lost` on day
one means my extraction differs from `VisibleTextFromHTML`'s, NOT moved
legislation** — fix the quote, don't believe the alarm. `[UNVERIFIED]` until
that first sweep passes: quote-normalisation parity between my python
extraction and the Go extractor is designed-for but not yet observed.

**Still open on this thread (handoff updated):** (1) tools-vs-facts acceptance
— nothing yet connects an evidence fact to the JS constants inside a tool; the
real fix is an oracle-from-the-register acceptance check (platform seam →
council when built; the loanandmortgagecalculator lane's oracle.py is the
worked pattern). (2) The published "current rules" page the owner floated —
right vehicle is a framework-built guide page whose numbers come from
writer_lines; BLOCKED on confirming the page-row creation path for a new guide
page on this site before filing (add_tool creates tool pages; guides arrived
with the adoption).

## 2026-08-09 (afternoon) — measuring the facts→tools seam before designing it

Design written up as `PLAN_2026-08-09_facts_into_tool_acceptance.md`. **No code
this session.** What follows is the evidence it rests on; each item names the
check, and where a check could not have come out otherwise I say so.

**The five improvement-loop items are all `complete`** — built 11:08–11:19Z,
verified by row id (`c9f810a3`, `df5c5935`, `ba68c674`, `0dc7a786`, `0c529013`).
The site now has 14 tool pages. **OWED: re-run the replay comparator** — the
08-08c handoff's follow-up, now unblocked. Nothing yet confirms the rebuilds
landed the agreed models rather than merely reporting success (016b: `complete`
is not proof the work happened).

**The register's first sweep has NOT run over our facts.** `scheduled_tasks`
`evidence-freshness`: enabled, 86400s, `last_completed_at = 2026-08-09
08:58:22Z` — i.e. **before** the ~12:30 seed. Zero `stale_evidence`/`citation`
items for this site. Due ~08:58Z 08-10. `[MEASURED]` The RUNBOOK §11 check is
still owed and the day-one gotcha still applies.

**Fact shape, enumerated rather than read off a seed** (there is no `.sql` in
the repo for these facts — they were seeded direct against the live row, so a
repo grep would have found nothing and told me nothing):
`{id, kind, unit, claim, value, source.citation{url,quote,publisher,title,
accessed,published}, verified_at, writer_line, staleness_days}`; top level
`{facts, governing_rule, writer_block_managed}`. `[MEASURED via
jsonb_object_keys]` — a path read would not have seen a shape change underneath
it, which is why the keys were enumerated.

**The tool agents are blind to the register.** `page-content-writer` and
`build-site-planner` reference `evidence_base`; `tool-generator`,
`tool-deployer`, `tool-recreation-handler`, `tool-improver`, `tool-suggester`
and `tool-acceptance-agent` do not. `[MEASURED]` — **disconfirmable: the same
query returned true for two of the eight**, so a blanket false was not baked in.

**…and yet `tool-recreation-handler` already loads them.** Its `load_site_specs`
step calls `read_site_spec` **with no `aspect` in config**, and that mode
returns *all* current aspects keyed by aspect name
(`site_spec_actions.go:457-490`). So `{{.site_specs.specs.evidence_base.facts}}`
— `build-site-planner`'s own template path — already resolves in its context.
The facts arrive and are never shown. This is PBP-037's exact finding recurring
on the tool path, and it makes the highest-value first move a **prompt seed with
no Go and no image roll**.

**Structural: this site's twelve recreated tools have no `doc_plans` PLAN**, so
no criteria, so no Tier 2 and no Tier 4 — and **zero `acceptance_run` /
`improve_tool` / `audit_tool` / `acceptance_stuck` items have ever existed for
this site.** `[MEASURED]` The two companions built this morning DO have PLANs
(created by `tool-generator` 11:17/11:19Z). This is `TL-032` biting as written.

**`doc_plans` has no `site_id` column** — `UNIQUE (subject_type, subject_key)
WHERE is_current`, fleet-global. `[MEASURED: \d doc_plans]` So a fact id (which
is per-site) cannot be resolved against the PLAN; it must be resolved against
the site of the page being driven. Today `mortgages-stamp-duty`
(loanandmortgage) and our `tool-stamp-duty` are the same calculator under two
keys and do not collide — **that is luck, not design**, and 0 collisions
fleet-wide today does not license depending on it.

**LANDMINE (not yet filed to LANDMINES.md — see below): never round-trip
`evidence_base` through the typed struct.** `EvidenceBase`/`EvidenceFact` in
`datahelpers/claims.go` do not model `citation`, `writer_line`, `unit`,
`staleness_days` or `writer_block`. Both live write paths
(`refresh_evidence_base_action.go:683`, `evidence_citations.go:350`) marshal
`map[string]interface{}` — which is *why* those keys survive. A new consumer
that parses typed and writes back would silently delete every citation on the
site, and the sweep would then report the facts as unsourced rather than as
damaged.

**Concurrent-lane state, re-measured rather than read from the commit.**
loanandmortgagecalculator's `5dbd47653` (14:25Z) says its fences were "NOT
installed". `doc_plans` says 9 `mortgages-*` PLANs carrying `computed_values`,
`created_by = operator:bugfix224-session`, written **14:33–14:40Z** — after the
commit. Both are true statements about different moments; the commit is not the
current state. Fleet: 19 of 59 current tool PLANs carry `computed_values`.
This is the "a record goes stale faster than its reader can tell" case with a
half-life of fifteen minutes.

**The design's load-bearing borrowing:** PBP-037's settled semantics — *the
assignment pins WHICH facts, never their values*. Anything that pins a value
into an artefact re-creates the golden trap that `run_checks_action.go:775-781`
already names in the code that does it.

**Owed follow-ups from this session:** (1) re-run the replay comparator;
(2) check the 08-10 sweep; (3) file the typed-struct landmine to `LANDMINES.md`
+ `--apply` the sync; (4) the twelve missing tool PLANs.

## 2026-08-10 — the sweep proved itself, the rebuilds landed, and Piece 1 is live

Cold start for this lane is now `HANDOFF_2026-08-10_continue_here.md`.

**A2 DONE — the legislation watch is PROVEN, not merely armed.** Sweep ran
09:02:33Z. All four SDLT facts: `verified_at` AND `source.citation.accessed`
both moved `2026-08-09` → `2026-08-10`; zero `stale_evidence`/`citation` items.
`[MEASURED]` — and this is the check that could have come out otherwise: four
`citation_lost` items was the predicted day-one failure. **It also closes
08-09's `[UNVERIFIED]`:** my python quote extraction and Go's
`VisibleTextFromHTML` agree on all four quotes. Per RUNBOOK §11 the proof is
`verified_at` moving on OUR facts, never the task's own `last_completed_at`,
which covers the fleet. **The day-one gotcha is now spent — the next
`citation_lost` here is a real signal.**

**A5 DONE — comparator re-run:
`acceptance/COMPARE_2026-08-10_after_supply_both_builds.txt`.** Verdicts in the
handoff §1(b). The three supply-both rebuilds landed:
- bridging-loan **VERIFIED** outright (16 rounding-equal).
- rate-forecaster: defaults drive to **1,389.58 / 1,525.78 / 1,286.39** — the
  spec's worked check to the penny, so the 3-phase model landed. Its lone
  DOMAIN-DIFF is the `double` vector, which is a **50-year term**; driven
  directly, the rebuild answers *"Please enter a term of 40 years or less."* and
  computes correctly at 40. A stated cap, same class as repayment's fractional-
  term refusal. `[MEASURED — drove the live page at 25y, 50y and 40y]`
- fee-analyser: `tcTotal` **£17,384.79** (= the spec's worked check exactly) and
  `tcOutlay` **£26,841.44** (= the original to the penny). `[MEASURED — drove the
  live page at the golden's defaults via CDP]`

**MISSTEP AVOIDED, and it is a new trap: I nearly read fee-analyser's DIVERGED as
a defect.** `compare_rebuilt.py` judges only ids present on BOTH sides. A rebuild
specified to ADD an output is therefore **structurally guaranteed** to read
DIVERGED: the id that agrees with the original (`tcOutlay`) is new and invisible
to the comparison, and the id that gets judged (`tcTotal`) is the one we
deliberately changed. The verdict is a property of the comparator's design, not
evidence about the tool. **Drive the new ids directly before believing DIVERGED
on any tool whose spec added outputs.** Also filed to the handoff §5.

**A3 PART DONE — migration `366` applied**: `tool-recreation-handler`'s
`recreate_tool` prompt now carries a "Verified facts — these OVERRIDE the
original tool AND the specification" section injecting
`{{.site_specs.specs.evidence_base.writer_block}}`. Snapshot `8701375f`,
`UPDATE 1`, guard passed, recorded in the ledger.

Three things worth carrying forward from doing it:

1. **`--apply` takes EVERY pending file — 11 others were pending**, one of which
   (`324`) refuses by design because on an older binary it deploys the wrong
   asset bytes. Scoped with `MIGRATIONS_DIR=<dir with only my file>`, md5 checked
   against the repo file first so the ledger's checksum is the real one.
2. **My own guard refused my own file** — I asserted the `writer_block` reference
   appeared once; it appears twice (`{{if}}` + interpolation). The guard was
   right and the EXPECTATION was wrong. Fixed to `= 2`, not loosened to `>= 1`,
   because the exact count is the double-application check.
3. **The no-op case was the one that could have broken six sites.** A malformed
   template, or a chained access through a missing map key, would break tool
   recreation fleet-wide. So the LIVE prompt was pulled from the DB and parsed +
   executed through the same engine and funcMap as
   `datahelpers.RenderPromptTemplate` across four shapes: register+block →
   renders; register without block → else; **no `evidence_base` aspect at all →
   else, no error, no `<no value>`**; empty specs → else (its lone `<no value>`
   is the pre-existing `identity.industry` line, not this section).

**366's effect on a real rebuild is UNPROVEN and must not be written up as a
win.** A prompt change with no observed output is a claim. The proof is to
re-file one recreation and read the generated JS for £500,000 rather than
£625,000 — next action 1 in the handoff. Note also that the code comment 366
asks for beside each registered constant is a **trace for a human reader**; it
must never become the machine declaration of Piece 2, because a comment enforces
nothing and a source-scanning consumer would make every comment load-bearing.

### 2026-08-10 evening — 366 PROVEN on a real rebuild, A1 done, and the register turns out to cut both ways

**The handoff's proposed proof was not disconfirmable, and I nearly ran it anyway.**
Next action 1 read: *re-file one recreation and read the generated JS for
£500,000 rather than £625,000.* I dumped the **existing** component first
(`page_components` `9bf28c5e`, built 08-08, i.e. BEFORE 366) and it already
contained `const FTB_RELIEF_LIMIT = 500000;` and the correct band table. So that
test returns £500,000 whether or not 366 exists — it measures the model's memory
of SDLT, not the register. `[MEASURED — the pre-366 artefact, read in full]`

**What discriminates is ATTRIBUTION, not the number.** 366's prompt asks for the
fact's wording *beside the constant, in a code comment*. That is a thing the
register can cause and the model's own knowledge cannot. So the test became: do
the register's composed `writer_line`s appear in the artefact? With the pre-366
build as the control, since it is the same tool, same agent, and a spec
identical but for one id-contract clause.

| register writer_line | pre-366 | post-366 |
|---|---|---|
| Standard residential SDLT is banded: nothing up to £125,000, … | 0 | 1 |
| First-time buyers pay no SDLT up to £300,000, then 5% … | 0 | 1 |
| Above £500,000 first-time buyer relief disappears entirely … | 0 | 1 |
| An additional residential property usually costs 5 percentage points … | 0 | 1 |
| *positive control* `Stamp Duty` | 7 | 3 |

**The first run of that table said 0 and 0, and it was my measurement that was
broken, not the change.** The generator wraps a long `writer_line` across two
`//` lines, so a verbatim search finds neither side. Strip comment markers per
line, then collapse whitespace, and it resolves. **A verbatim match against
generated source is a claim about the generator's line width** — worth
remembering before reading a 0 as an absence. The positive control is what
stopped me publishing the first table: `Stamp Duty` matching in both files
proved the search could fire on either.

**Item mechanics, and one that will bite the next person.** Filed
`49bbd08b` at `triaged`; it read **`complete` 52 seconds later**. It was not.
The orchestration ended at `complete_error`, `__step_error.failed_step =
analyze_tool`, message *"You have reached your specified API usage limits. You
will regain access on 2026-09-01"* — the fleet-wide Anthropic cap, failing
14:51–17:02Z. `result.response` held the **site record**, an early step's
output, which is what a truncated run looks like from the item. `page_components`
was untouched, so it was a clean no-op. **Recovery measured, not assumed: last
failure 17:02:12Z, then 70 successful calls in the 18:00 hour across 3 agent
types** — the stated 2026-09-01 reset did NOT hold, and the fleet-wide LANDMINES
entry has been corrected accordingly (it currently tells every lane its council
obligation is unsatisfiable for three weeks; it is not).

Attempt 2 (`e0a64199`, 18:19) ran properly: claimed in ~70s, component saved in
~4 min, deployed, item complete in 5m05s.

**Two results from that build.**

1. **Option VALUES landed exactly** — `next` / `ftb` / `additional`, in the
   original's order, `next` selected on load. That clears handoff action 4:
   stamp-duty is no longer **REPLAY-FAIL**. The comparator now judges it and
   returns **DIVERGED with the ORIGINAL wrong**: at £595k FTB golden `£14,750`
   vs rebuilt `£19,750`, and the defaults vector (£350k FTB) agrees at `£2,500`.
   Report: `acceptance/COMPARE_2026-08-10b_stamp_duty_option_values_aligned.txt`.
2. **The rebuild DROPPED the £40,000 additional-property surcharge floor.**
   `SURCHARGE_FLOOR = 40000` appears twice in the pre-366 build and **zero**
   times after. That is true law, correctly implemented before. It went because
   366's own section says *"Do NOT state a rule that is not in the register"* and
   the register held four SDLT facts, none of them the floor. **Nothing failed.
   The tool simply became wrong below £40,000, silently.**

> **This is the finding of the day and it is not the one I went looking for:
> the register is load-bearing in BOTH directions. What it omits can be deleted
> from a rebuilt tool.** A partial register is not a neutral one — and every
> register is partial. Filed fleet-wide to `LANDMINES.md` with the prospective
> check (enumerate the constants the current tool encodes, ask which the register
> carries, register the gaps BEFORE filing the rebuild).

**A1 DONE — 4 facts → 13, one per band edge and per rate.**
`evidence/SEED_2026-08-10_sdlt_facts_per_threshold.sql` (+ its generators, now in
the repo — the PLAN recorded their absence as a gap). Standard bands: 125k/2%,
250k/5%, 925k/10%, 1.5m/12% as separate scalar facts; FTB nil band, FTB 5% rate,
relief cap; surcharge rate; **and the £40,000 floor, cited to the higher-rates
guidance page** — registered precisely because the rebuild had just shown what
omitting it costs. Retired `sdlt-standard-bands` (bands in prose) and
`sdlt-ftb-nil-rate` (two rules in one claim); checked first that neither id is
referenced by `doc_plans`, `site_work_items` or `page_components` — only this
lane's own docs. `pinned` carried forward (CLM-001: a replacement row defaults
to false and silently loses human-owned status).

**Quotes were lifted by the REAL Go extractor, not by python.** `evidence/quotecheck`
is a scratch module that `replace`s the repo and calls
`datahelpers.VisibleTextFromHTML` + `QuoteFoundInText` directly, so the day-one
`citation_lost` class (my extraction vs the sweep's) cannot arise. Quotes come
out of the dumped text **by regex, never retyped**, with `.` standing in for the
currency symbol and the curly apostrophe so nothing non-ASCII is typed at all.
All 13 verified against the live GOV.UK pages. **And the check was induced red
first**: asking for `Up to £126,000 Zero` returns `NOTFOUND` and exit 2 in the
same run as the real ones. Thirteen FOUNDs mean nothing until one NOTFOUND shows
the check can fail.

**Then the induced proof, run forward.** Re-filed the recreation a third time
(`f7016d32`) with a spec **byte-identical to attempt 2** — diffed as parsed JSON
before firing — so the register was the only changed input. The result:

- `const ADDITIONAL_THRESHOLD = 40000;` is **back**, carrying the register's new
  writer_line as its comment, and **read at the branch** (`if (price >=
  ADDITIONAL_THRESHOLD)`), not merely declared. A declared-but-unread constant
  would have been the other way to fail this test.
- **All ten** granular writer_lines now appear as comments beside their
  constants, including the five band lines that did not exist in the register
  four hours earlier. The generator even titles the block *"Rate bands (verified
  fact register; wording beside each constant)"*.
- Arithmetic unchanged and correct: £19,750 at £595k FTB, £2,500 at £350k FTB.
  Live on the wire at `/tools/stamp-duty/index.html` (25,741 B, 200).
  `acceptance/COMPARE_2026-08-10c_stamp_duty_register_driven.txt`.

So the chain **register → prompt → generated JavaScript** is now demonstrated end
to end, in the direction that matters (change the register, the tool changes),
without lying to the register to do it. `[MEASURED — but n=1 on a
non-deterministic generator: this evidences the mechanism, it does not prove it.
The honest claim is that the register was the only changed input.]`

**Misstep, minor, logged to `WRONG_CALLS.md`:** I ran `landmines-sync.py --apply`
and then `landmines-verify-dispatch.sh`. The dispatcher runs the sync itself and
computes "new or changed" by diffing against the rows the sync already wrote — so
my direct `--apply` **consumed the signal**, and the dispatcher exited 0 saying
"nothing needs verification", which reads exactly like "all fine". Run the
consumer, not the producer. The two new entries are synced to `doc_notes` but
have not been through the landmine-verifier.

---

## 2026-08-10, third session — A4: the twelve tool PLANs, and the four things the plan had wrong about them

Picked up `HANDOFF_2026-08-10b` §3 action 1: "create tool PLANs for the twelve
recreated tools … the single biggest blocker". It is done for eight of them, and
**four of the assumptions I inherited were wrong.** Each was wrong in the same
direction — the work looked smaller and safer than it was — so they are recorded
before the result.

### 1. The subject key is NOT the page name, and a PLAN under the wrong key fails SILENTLY and for ever

Both tiers derive the key themselves
(`discovery_checks/tool_eligibility.go`, `toolSubjectKeyExpr`):

```
CASE WHEN cc.component_level='tool' THEN cc.function
     ELSE regexp_replace(p.name,'^tool-','') END
```

Our recreated pages carry a **section** component, so `tool-stamp-duty` is keyed
**`stamp-duty`** — not `tool-stamp-duty`. Had I written the PLANs under the page
name, Tier 2 would have gone on recording `needs_criteria` and Tier 4 would have
gone on emitting nothing: **indistinguishable from having written no PLAN at
all.** No error, no log line, no row anywhere saying "there is a plan but I
cannot see it".

**This is not inference — the platform had already written the answer down.**
`doc_notes` for this site carries `needs_criteria` notes under subject keys
`simple` and `stamp-duty` ("tool_acceptance sweep found no current PLAN criteria
fence (has_plan=false)"). The sweep was looking for exactly the keys I ended up
using, and had been for days. I found this after choosing the key from the Go
source, which is the only reason I know the source and the live system agree.

### 2. Three of the "twelve" are not ladder-eligible at all

`toolEligibilityWhere` admits a component only if it is `component_level='tool'`,
OR it is the **sole** component on a `page_type='tool'` page. Measured:

| tool | why it is out |
|---|---|
| `tool-affordability` | **two** components (hero + generic-text-block) — fails the sole-component clause |
| `game-fact-finder` | `page_type='game'` |
| `investor-index` | `page_type='section-index'` |

So the population is **nine**, not twelve. A PLAN for the other three would be a
row that reads like coverage and is never loaded. The query that establishes this
could have come out otherwise: it returned rows for the nine and nothing for the
three, in one pass.

### 3. Installing a PLAN turns Tier 2 ON — and Tier 2 can fail a page the fence says nothing about

The neighbouring lane's `install_fences.py` states the guard I was about to
inherit: *"With only computed_values in the fence, Tier 2 finds nothing it can
fail, so it can never raise improve_tool for these pages."* **That is incomplete,
and the gap matters here more than it did there.**

`check_tool_acceptance.go:478-500` appends **three built-in shell failures**
outside the criteria loop entirely — `shell-doc-header`,
`shell-template-residue`, `shell-dead-controls`. They run on any tool with a
parseable fence, whatever the fence contains. Any one of them creates an
`improve_tool` item carrying `spec.component_id`, and for these pages that id is
the **shared `hero` component: 252 pages across 18 sites** (measured) — wider
than the ~154-page ported-page shell that lane was protecting.

So before installing anything I ran the three checks against all twelve live
pages **using the platform's own functions** — a scratch Go module that
`replace`s the repo and calls `content.ToolDocOpen` and
`datahelpers.DeadControlAnchorsOutsideRuntimeFill` directly. Re-implementing
`DeadControlAnchorsOutsideRuntimeFill` in python would have been a claim about
its behaviour, and it carries a per-anchor runtime-fill exemption
(`bugs_open/137`) that a re-implementation would very likely get wrong.

**Twelve PASS, and then the red was induced** — a fixture carrying all three
defects fires all three in the same run as a real page passing. Twelve greens
from a checker that has never gone red are not evidence.

`[MEASURED 2026-08-10 — and it is a fact about TODAY, not a guarantee. A future
copy edit that leaves a dead anchor on one of these pages hands the fleet's
shared hero to an automated rewriter. The thing actually holding that off is
`no_auto_fix: true` on the Tier-4 side plus these twelve passes on the Tier-2
side; there is no structural guard.]`

### 4. "Zero acceptance runs have ever happened on this site" is now false

Two `acceptance_run` items completed on 2026-08-09 20:56–21:04 for the two
generator-built companions, both PASSED at Tier 4. True when the PLAN wrote it
on 08-09; overtaken the same evening. Re-measured rather than carried forward,
per this file's own rule — including from this handoff.

### What was actually built

`acceptance/verify_criteria.py` (new), `acceptance/install_fences.py` (new),
`acceptance/criteria/*.criteria.json` (nine emitted).

**Emit → re-derive → install**, and the middle step is the one that matters. An
emitted value is only "expected" because the tool prints it; pinning one that
nothing else reproduces is F3 from the PLAN's own table, which
`run_checks_action.go:775-781` states in the code that does it. So every value is
recomputed from a source that is not this page's script, at one of three
strengths, reported separately and never flattened:

- **DEFINITION** (56 assertions) — the published formula, via the neighbouring
  lane's `oracles.py`. Reused, not re-written: it was authored from the
  definitions, and a second copy is a second thing to keep right.
- **REGISTER** (4) — stamp-duty, and this is the lane's whole point: the bands
  are built from **this site's 13 registered SDLT facts**, each a scalar with its
  own verbatim GOV.UK quote, re-verified daily. Not from `oracles.py`'s
  hard-coded band table, which would be a second hand-typed copy of the law.
- **CONVENTION** (20) — the tool's own design choice (rate-forecaster's 24/36
  phase split, read from its script; fee-analyser's definition of "total cost").
  Weaker, and labelled so: it catches a rewrite that moves the arithmetic, not a
  convention that was wrong to begin with.

**80 of 80 agree.** Anything not re-derived was **dropped, not pinned** — that
rule replaces the neighbouring lane's substring container heuristic and does the
same job better: containers, prose breakdowns and echoed inputs all fall out of
it automatically. 41 assertions dropped across the eight tools.

**The register mutation is the control, and it is the best evidence here.**
`verify_criteria.py --mutate sdlt-ftb-relief-cap=625000` — the SUPERSEDED
pre-April-2025 cap — makes the £595k FTB vector expect **£14,750**: the original
tool's wrong figure, the £5,000 under-quote, reproduced exactly by putting the
expired rule back into the register. That single run establishes what 80
agreements cannot: the register is genuinely the source of the expectation, not
decoration beside it. A second control (`sdlt-standard-rate-250k-925k=6`) fires
on two vectors.

### Two tools install nothing, and that is the correct outcome

- **`portfolio`** — toolgolden derives its vectors by scaling the page's own
  defaults, and this form has none, so it drove `#mortgageTerm` to 1000 / 2000 /
  500 / 450 years. The tool refused all four. **Every emitted assertion is the
  validation message** "Please enter a remaining term between 1 and 40 years."
  A fence built from that would certify an error message and call it a
  calculator — F3 wearing a different hat. It falls out of the "only re-derived
  assertions" rule rather than needing a special case.
- **`fact-finder`** — not ladder-eligible (§2).

### A misstep of my own, and it wore the costume of the defect it was written to find

`verify_criteria.py`'s first run reported rate-forecaster's `#diff2` wrong by
**£1,923.22** — a number large enough to look like a real arithmetic fault. It
was my parser. The page renders a fall as `<U+2212 MINUS SIGN>£961.61`: the sign
sits **outside** the currency symbol, so a `re.search(r"-?\d…")` — which requires
the sign to be adjacent to the first digit — matches at the `9` and returns
**+961.61**. I had even written a comment about the U+2212 trap while walking
straight into the adjacency one beside it.

Caught only because the oracle disagreed. Had the same parser read both sides it
would have agreed silently and pinned a sign error into the acceptance record.
**Strip the noise; do not scan past it** — `re.sub(r"[^0-9.\-]", "", s)`, which
is what the neighbouring lane's `num()` did all along. Logged to `WRONG_CALLS.md`.

### One assertion was modelled, disagreed, and was then DROPPED rather than argued into agreement

`#saveTime` ("3 years 6 months"). My model disagreed with the page by **exactly
one month on three of four vectors, always one month more.** That pattern is not
an arithmetic fault: both sides run the same textbook amortisation and part
company only on **when a balance counts as cleared** — the page stops at
`remaining > 0.005` (half a penny), `oracles.amortise` at `1e-9`. A residual
between those thresholds ends the schedule a month earlier on one side.

Nothing published settles a sub-penny residual, so asserting either number would
pin **my** convention as the tool's law. That is exactly the move
`PLAN_2026-08-09` §5.4 forbids and that the neighbouring lane logged to
`WRONG_CALLS` on 08-09 (six "mismatches" that were its own rounding convention).
The arithmetic is defended by `#saveInterest`, which agrees to the penny. The
reasoning is written into the file where the assertion would have gone, so the
next person can settle the threshold and pin it honestly rather than rediscover
the ambiguity.

### Installed, and verified at the artefact rather than at the status

Eight rows in `doc_plans`, keys `bridging-loan`, `equity-release`, `fee-analyser`,
`overpayment`, `rate-forecaster`, `repayment`, `simple`, `stamp-duty`.
80 assertions, all `computed_values`, `profiles: ["desktop"]`,
`no_auto_fix: true`.

`fence_pos > 0` and a `LIKE '%computed_values%'` prove nothing about content, so
every fence was **read back out of the database, parsed, and compared to its
source file**: 80 of 80 byte-identical, including the 68 assertions carrying
non-ASCII and specifically the U+2212 in `rate-forecaster/computes-asym/#diff2`
(confirmed by code point, not by eye).

### The blocker this uncovered: 7 of the 8 can never be swept

`check_tool_acceptance_due.go` gates on `PageHasShippedPredicateFor` =
`NOT (deployed_at IS NULL AND build_status <> 'deployed')`. Measured:

| page | build_status | deployed_at | sweepable |
|---|---|---|---|
| tool-simple | deployed | 2026-08-09 | **yes** |
| the other seven | needs_rebuild | **NULL** | **no** |

All seven serve HTTP 200 — I fetched every one of them. They were built and
deployed; `deployed_at` was simply never stamped. This is not a new discovery:
`datahelpers/links.go:304-308` records the same measurement from 08-09 and names
**"nine mortgagecalculator.co.uk pages, almost all `build_status =
'needs_rebuild'`"** as its worked example. Our seven are inside that nine.

So installing the fences turns the ladder on for **one** tool automatically. I
have NOT stamped `deployed_at` or flipped `build_status` to make the others
sweepable: both would assert a deploy event I did not observe, on rows parked in
a queue (`needs_rebuild`) that MEMORY records as dead and that another lane may
own. That is a decision to take deliberately, not a side effect of finishing A4.
Recorded in the handoff as the top open item.

### Proving a fence actually executes

A PLAN with no run is a claim, not a result — the correction the previous handoff
earned. So one `acceptance_run` was filed by hand for **stamp-duty** (the
register-driven one), following the due sweep's own item shape.

### Both new landmines came back NEEDS_HUMAN_REVIEW — and the gap is in the verifier's index, not the entries

Fired `trigger-landmine-verifier.sh` for both new entries (running
`landmines-verify-dispatch.sh` would have been the *correct* consumer, but the
previous session's misstep — running `--apply` first and consuming the diff —
had already happened here, so the two triggers were fired directly from the
`NEEDS_VERIFICATION:` lines the sync printed).

Both verdicts confirm the core mechanism and then stop:

- *"`ToolDocOpen` returned 0 rows (index cannot represent const/var kinds)"*
- *"`tool_eligibility.go` (and its symbols `toolSubjectKeyExpr`,
  `toolEligibilityWhere`) returned 0 rows in both path and symbol searches"*

**Checked, and the verifier is right about itself:**

```sql
SELECT kind, count(*) FROM code_symbols GROUP BY 1 ORDER BY 2 DESC;
-- func 3653 | method 1119 | struct 987 | alias 42 | interface 36
SELECT count(*) FROM code_symbols WHERE path LIKE '%tool_eligibility%';        -- 0
SELECT count(*) FROM code_symbols WHERE path LIKE '%check_tool_acceptance.go%';-- 21
SELECT count(*) FROM code_symbols
 WHERE symbol IN ('toolSubjectKeyExpr','toolEligibilityWhere','ToolDocOpen');  -- 0
```

**`code_symbols` carries five kinds and `const` is not among them.**
`tool_eligibility.go` declares *only* constants, so **the whole file is absent
from the index** — 0 rows, against 21 for the sibling file that happens to
contain functions. It is not stale and it is not mis-pinned; the file is
structurally unrepresentable.

Two things follow, and the second is the one that matters beyond this lane:

1. **A landmine whose footprint is a Go `const` cannot currently be verified.**
   Both of mine are: the subject-key rule *is* `toolSubjectKeyExpr`, and the
   Tier-2 sentinel *is* `content.ToolDocOpen`. The verifier behaved correctly —
   it reported NEEDS_HUMAN_REVIEW rather than passing on an absence, which is the
   right way round (a `0 rows` that reads as "confirmed absent" is the failure
   mode this estate has logged repeatedly). But the entries most worth verifying
   are often exactly the ones anchored to a constant, because a constant is what
   a shared rule gets written as.
2. **The gap is the index, not the verifier**, so it is wider than landmines:
   anything reading `code_symbols` — `diagnose_code_lookup`, the council seats'
   read-only checks — is blind to every Go constant in the estate, and to any
   file that contains nothing else. `tool_eligibility.go` is a live example of
   the second case, and it encodes a rule several lanes need.

`[MEASURED 2026-08-10, live DB. Disconfirmable: the same query returns 21 rows
for the neighbouring file, so a blanket zero was not baked into the check.]`
Not filed as a bug from this lane — it belongs to whoever owns the code index and
RFC_005's verification path, and I have not checked whether it is already known.

### All eight fences driven, all eight PASSED — 19:05–19:16Z

`stamp-duty`, `bridging-loan`, `equity-release`, `fee-analyser`, `overpayment`,
`rate-forecaster`, `repayment`, `simple` — **4/4 checks each on desktop**, mobile
skipped by design, **zero `acceptance-fail` notes**. Runs were filed by hand
because the due sweep cannot see seven of the eight pages (the `deployed_at`
blocker above).

Three of these are worth more than "green":

- **`rate-forecaster` is the encoding proof.** Its `computes-asym` vector asserts
  `<U+2212>£961.61` as EXACT text, and `computed_values` permits whitespace and
  nothing else. Passing means the character survived python → JSON →
  dollar-quoted psql → `doc_plans` → the fence extractor → the Kafka envelope →
  headless Chromium → the comparison. The DB round-trip I checked earlier covers
  two of those hops; this covers all of them.
- **`equity-release` passed where the neighbouring lane's equivalent failed
  today.** `mortgages-equity-release` FAILED at 03:28 with `#dispAge reads "130",
  expected "65"` — state bleeding between vectors, because the runner opens ONE
  page per (url, profile) and runs every check against it with no reload. Ours
  drives `#erAge` absolutely in every vector, so each check sets its own state and
  cannot inherit the previous one's. Not foresight on my part: it falls out of
  emitting every driven input per check. Worth knowing as the reason, though,
  because a future fence that omits an input from a later vector re-acquires the
  bug.
- **`overpayment` is the thinnest fence — 1 assertion per vector** — and that is
  the honest consequence of dropping `#saveTime` and everything else the model
  could not reproduce. A fence of four assertions that all mean something beats
  nineteen that include a duration I could not derive and a prose panel that
  fails on a copy edit.

**Nothing was armed by installing them.** Re-measured after the runs: zero
`improve_tool` and zero `acceptance_stuck` items fleet-wide in the surrounding
three hours, and the only work items anywhere naming the shared `hero` component
`23f95f00-f293-466e-b43a-81791ea0fc6c` are these eight acceptance runs. That is
the §3 risk checked *after* the fact as well as before it — the before-check was
the twelve-page shell sweep, and they agree.

**And the check type is not inert:** 41 `acceptance-fail` notes fleet-wide
against 118 runs, including the `mortgages-equity-release` one above whose detail
is the exact `computed value(s) diverge` arm. So "can this fence ever go red?" is
answered by fleet evidence rather than by deliberately corrupting one of our own
live PLANs — which would have cost an `acceptance_stuck` item and proved
something already demonstrated this morning.
