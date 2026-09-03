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

### Post-roll check (chassis v1.0.1283, pods 21:43Z) and the batch-8 visibility landmine measured against our fences

The fresh build changes nothing for this lane — everything shipped today is
config, and `git log` on the five acceptance-path files shows no commit in the
last day. All eight runs completed before the roll.

The staged_component_build lane's finding (`68b7d78da`) that **`computed_values`
reads a `display:none` subtree** — `InnerText` falls back to `textContent` on an
unrendered element, so a tool that computes correctly and never shows the visitor
anything passes a values-only fence — was measured against our eight tools
directly rather than assumed either way:

- results containers on all eight pages are **visible from load and update in
  place**; no reveal-class JS (`classList.add('visible')` etc.) on any of them;
- the only `display:none` rules are chrome (`.header-cta` in the mobile media
  query, same two rules on every page);
- the single `hidden` attribute in the population is `#tcError` on fee-analyser —
  the error line, not a result.

So the visibility gap is **empty on this site today**: there is no reveal step
for a mutant to break. `[MEASURED 2026-08-10 late — and it is a property of the
CURRENT builds. A future recreation that introduces a hidden-until-submit panel
re-acquires the gap silently, because the fence stays green either way; the
handoff carries the check to add when that happens.]`

---

## 2026-08-11 — two owner decisions executed/routed, a concurrent lane found IN THIS DIRECTORY, and its finding extends to our tools

**Owner decisions (this morning, via the session):**

1. **The seven unstamped pages: stamp `deployed_at` ONLY.** Executed 12:57:35Z —
   `UPDATE pages SET deployed_at=now() … RETURNING` confirmed exactly the seven
   fenced tool pages, `build_status` untouched. **The narrow variant is the
   point:** the Tier-4 due sweep's predicate is satisfied by `deployed_at` alone,
   while Tier 2's built-in shell checks (the 252-page shared-hero rewriter path)
   gate on `build_status='deployed'` and stay OFF. `[The timestamp records the
   DECISION, not an observed deploy — the pages have served since 08-08/09; any
   reader of pages.deployed_at on this site should know these seven are
   owner-authorised inferences, 2026-08-11.]` No surprise sweep follows: all
   eight tools hold an `acceptance-run` note < 7 days old, so the due check's
   cooldown skips them until ~08-17.
2. **Equity-release max cash: MATCH THE ORIGINAL's table** (£120k, not the
   rebuild's £124k — lender policy, and the rebuild's table is the generator's
   invention). **[CORRECTED 2026-08-11 afternoon, entry below: these two figures
   are SWAPPED — the original gives £124k, the rebuild £120k. The routed action
   is unaffected.]** **NOT executed this session, and the obvious mechanism is a
   trap:** an `improve_tool` item would carry `component_id` = the shared hero
   and hand a 252-page component to tool-improver. The safe route is §12's
   `needs_tool_recreation` re-file with the original's age→percentage table
   extracted from the live original (`/equity-release.html`) and pinned in the
   spec's id contract. After the rebuild: re-emit → re-verify → re-install the
   fence for that one tool (a regenerated page may legitimately reformat), then
   re-run its acceptance.
3. (Resolved without us: the stamp-duty ORIGINAL was already patched by the
   owner on 08-09 during bugs_open/225 — re-verified on the wire today,
   `grep -c 625000` = 0 on the flat page. Handoff action 5 closed.)

**A concurrent lane is active and its plan lives in this directory.**
`PLAN_2026-08-11_decompose_into_framework.md` (bugfix_210 lane, committed 11:15
and 13:57 BST — the second within a minute of our stamp). Read it before any
site-wide action. Its coordination analysis holds: their pages are `index`,
`guide-*` and four never-built; ours are `tool-*` — **disjoint, safe in
parallel, PROVIDED NEITHER RUNS A SITE-WIDE REPLAN OR RERENDER.** Our stamp
touched only the seven tool rows and no build_status, so their `index` warning
(never flip it out of needs_rebuild before the port) is untouched.

**And their afternoon finding extends to our tools — measured, not assumed:**
the live homepage links **all eleven tools at FLAT paths** (`repayment.html`,
`stamp-duty.html`, …), contains **zero** `tools/` hrefs, and both forms serve
200. So the twelve rebuilt tool pages are in the `bugs_open/114` shape at page
level: correct, fenced, acceptance-passed — **and unreachable by any link a
visitor follows.** Visitors get the ORIGINALS (only stamp-duty among them was
patched). Consequences:

- The port lane's deploy-path decision (flat vs directory-form) must cover
  TOOLS as well as guides — one site-wide decision, not two lane-local ones.
- **Our fences survive a path migration untouched**: `request_browser_run`
  resolves the URL from `pages` by NAME at run time, and no fence stores a URL.
- What does NOT survive: `acceptance/compare_rebuilt.py`'s `MAPPING` dict and
  the golden replay both hardcode dir-form URLs — one-line-per-tool update when
  the paths move.

---

## 2026-08-11 (afternoon) — equity-release re-file EXECUTED; the decision text's two figures were SWAPPED; the 366 prompt read before filing

**The routed action from this morning's decision 2 is done**: work item
`97f4d0ab-bd28-481e-9e31-c2f45a2c4b2f`, `needs_tool_recreation`,
`item_key needs_page:tool-equity-release`, filed `triaged` at 14:24:31Z. Spec =
the 08-08 item's two features verbatim (calculator + id contract) **plus a third
contract pinning the original's age→LTV step table**, with the worked example
(£400,000 at 65 → £124,000), the no-linear-formula prohibition, the roll-up
formula, the minimum-age rule and the original's defaults all stated explicitly.
Read back from the row after insert: 3 features, £ signs intact.

> **CORRECTED 2026-08-11 (afternoon): the morning entry above and HANDOFF 10c §4
> recorded the two figures attached to the wrong sides** — they call £120k "the
> original" and £124k "the rebuild". Re-derived from both artefacts before
> filing:
>
> - **Original** (`/equity-release.html`; bucket and live sha256-identical,
>   `0befb538…`): a STEP table — `>=85: 0.52 · >=80: 0.47 · >=75: 0.42 ·
>   >=70: 0.36 · >=65: 0.31 · >=60: 0.25 · else 0.20`. At 65 on £400,000 →
>   **£124,000**. Its own comment says "65: ~30%" while the code uses 0.31 —
>   the page is internally inconsistent, which is likely how the swap started.
> - **Rebuild** (component `cfa17203…`, `maxLtvForAge`): LINEAR
>   `0.20 + (age−55) × 0.01` clamped to 0.55 → 0.30 at 65 → **£120,000**. The
>   installed fence corroborates: `computes-defaults` pins `#erMaxCash` £90,000
>   at 65/£300,000 = 0.30, and it PASSED on 08-10.
>
> What caught it: re-deriving the figure from each side's artefact before
> repeating it. The routed ACTION was unambiguous either way — "pin the table
> extracted from the live original" — and that is what was executed. **If the
> owner's intent was the £120,000 FIGURE rather than the original's table, the
> counter-action is one cheap re-file keeping the linear table** — flagged in
> README_where_we_are for the owner. Logged in `WRONG_CALLS.md`.

**The CLM-021 landmine was read against this re-file before filing** (this is
equity-release's FIRST rebuild under migration 366 — the 08-08 build predates
it, and the register carries zero equity-release facts):

- The live `recreate_tool` prompt routes unregistered thresholds to the SPEC:
  *"Do NOT state a rule that is not in the register. If the tool needs a
  threshold that is not listed, implement what the specification says…"* — and
  `interactive_features` land in "Mandatory Behaviour Requirements", which
  *"OVERRIDE anything implied by the original source code or the functional
  specification"*. The stamp-duty floor was deleted because it lived only in
  the reference-only original source; a table pinned as a spec contract is on
  the protected side of that line.
- **Handoff §5 action 4's constants sweep, discharged for THIS tool**: the
  current component encodes the LTV rule, min-age 55, the compound projection
  (N = 10/20/30) and input defaults. All are conventions/industry averages, none
  is a citable published rule, so none belongs in the evidence register (whose
  daily sweep needs a verbatim quote from an official source) — all four are now
  pinned in the spec contract instead. The sweep for the OTHER tools before
  their next rebuild remains open.

**Coordination checked before filing**: no non-terminal `needs_tool_recreation`
/ `improve_tool` on this page (and the dedup index would have rejected the
INSERT if one had appeared in between); recently-active transcripts grepped for
the symbols — the only other mentions are bugs 223/224 context and an enum
listing, nobody mid-flight on this action.

### The item then sat 80+ minutes untouched — fleet-queue starvation, not a filing fault — and was dispatched by hand

The 08-08/08-10 experience ("picked up in minutes") did not repeat: 50 minutes
of monitoring showed no orchestration row while the fleet completed 64 rerenders
on other sites. Cause read from the live `build-pipeline-trigger` definition,
not inferred: `find_dispatchable_site` picks **one site per 120s tick, ordered
by the globally oldest dispatchable item** (`ORDER BY created_at ASC LIMIT 1`).
Measured at ~15:30Z: 7 sites, **~273 dispatchable items older than ours** (81
on ai-agent-orchestration.com dating to 07-24). Everything else checked clean:
item dispatchable, site unlocked, no claimed item on the site, trigger firing
(last 15:21:56). Also noted on the way: the per-site pickup
(`load_work_item_actions.go:681`) orders `priority ASC` — **lower number first**
— so the copied priority 14 was harmless, and §12's example priority 8 is
actually ahead of it.

Bypass: the `081b_trigger_dispatch_gamesdesign.sh` precedent, adapted —
one `orchestrate` message pinning `build-dispatch-loop` to this site
(correlation `5125e6b6-2ce4-40ce-af1c-adbea1560f72`). Item `claimed` and
handler spawning within a minute. RUNBOOK §15 now carries the mechanism, the
three pre-checks and the kcat caveat. `[Starvation figures MEASURED 2026-08-11
~15:30Z — a fact about that afternoon's queue, not a property of the site.]`
Filed fleet-wide to `LANDMINES.md` (synced; verifier fired, correlation
`795585c6…` — NB the sync `--apply`-consumed-the-diff misstep from 08-10 was
repeated here before firing the trigger directly).

### Rebuild landed, fence rebuilt, acceptance PASSED 4/4 — the decision is closed end to end

**Rebuild verified at the artefact, then at the wire** (~15:28Z, run ~9 min
claim→complete): NEW component `539e851f…` (replaces `cfa17203…`), 14,343 chars,
`deployed`. The step table is present branch-for-branch with the contract's own
comments beside each band; original defaults restored (400000/65/100000/6.5);
all 10 contract ids exactly once; sub-55 refuses; projection formula unchanged;
**no reveal pattern** (results render on load and update in place — the batch-8
visibility gap stays EMPTY on this rebuild, `calculate()` runs once on load).
Served page: dir-form URL carries `age >= 65) return 0.31` (1 hit) and zero
traces of `maxLtvForAge|0.55` (control).

**Fence chain (§14), all green with reds induced:**

- **Emit**: 4 checks, 24 assertions. toolgolden's scaling landed the vectors on
  age 120 (→ the 85+ band, £416,000 on £800k) and ages 33/39 — **two of four
  vectors exercise the pinned minimum-age refusal**, worth having since the
  refusal is contract behaviour (unlike portfolio, where refusals were ALL the
  emit had).
- **Model**: `m_equity_release` now models `#erMaxCash` (step table, CONV) and
  the refusal markers (`N/A`, U+2014 — string-compare, terse markers pinned
  rather than the prose sentence, which dies on a copy edit). The swapped
  "rebuild 124k vs original 120k" comment — the likely origin of the morning's
  wrong call — corrected in place. Full run: **84/84 agree** (was 80; equity
  now contributes 16), `#dispAge`/`#limitResult` correctly fall out.
- **Red control**: the NEW model against the OLD linear page's criteria (from
  git `e211b596f`) — **3 MISMATCHes exactly where the two tables differ**
  (65: 0.31 vs 0.30; 95: 0.52 vs 0.55-cap, both directions), agreement at 55
  where the tables genuinely coincide, debts untouched. The control could have
  come out otherwise and didn't.
- **Tier-2 shell checks**: rebuilt the scratch Go module (the 08-10 one did not
  survive); new page passes all three, the fixture fires all three in the same
  run.
- **Install**: `--only equity-release --apply` — 16 pinned / 8 dropped, new
  `doc_plans` row current 15:33:12Z (£124,000 in, £90,000 gone, refusals in),
  08-10 row superseded.
- **Acceptance**: run item `67594cfc…` (second by-hand dispatch also needed,
  correlation `42ca7dbc…`) — **Tier-4 PASSED, 4/4 on desktop**, mobile skipped
  by design, zero `improve_tool`/`acceptance_stuck` fleet-wide after.

**And the run's render critique found a real defect the fence cannot see**: the
Calculate button label renders near-invisible (light-on-light) on the new
rebuild. The vision-finding mechanism (shipped this morning by the
staged_component_build session) **filed it automatically**:
`vision_finding` → `needs_human_review`, 15:35:58. Not this lane's fix — it sits
with the contrast machinery / owner review; noted here so nobody re-discovers
it. Evidence screenshot in the acceptance note
(`acceptance-evidence/…/equity-release/d81357a6…_desktop.png`).

> **CORRECTED same day (below): "not this lane's fix" lasted two hours** — the
> owner directed this lane to fix it, and the mechanism turned out to be ours
> (the tool generator's own CSS idiom), not the contrast lane's 382 class.

## 2026-08-11 (evening) — the ghost button diagnosed to a CSS self-cycle, fixed by migration 393 on THREE pages, redeployed, re-probed 15.39:1

**Owner direction (in chat): fix the button in this lane; notify the contrast
machinery lane.**

**Diagnosis — computed styles, not source reading** (probe script in scratch;
recipe: CDP `Runtime.evaluate` → `getComputedStyle`). The generator wrote the
theme bridge as a SELF-REFERENCE:

```css
.tool-page { --primary-color: var(--primary-color, #0b2545); ... }
```

A custom property referencing itself is a dependency cycle → invalid at
computed-value time; **the fallback cannot rescue its own cycle, and `:root`'s
perfectly good definition is not consulted** (`:root` has
`--primary-color: #b59230`; `.tool-page` computed **empty**). Probe results,
before:

| page | button bg (computed) | label contrast | bridge lines |
|---|---|---|---|
| equity-release | transparent | **1.05:1** | 9 |
| stamp-duty | transparent | **1.05:1** | 8 |
| rate-forecaster | transparent | **1.05:1** | 7 |
| simple (control) | `#0a2540` | 15.54:1 | 0 (literal idiom) |

So the vision finding's page was one of THREE — the whole bridge block
(primary/accent/bg/panel/border/text/muted) inoperative on each. Same
generator, two idioms; only the literal one can work. `content_data` is NULL on
tool components — `rendered_html` is the stored source, so no rerender path
regenerates this; the class fix belongs to the generator prompt (A3 territory)
and the spec contracts.

**Fix — migration `393` (+ ROLLBACK), the 382 shape**: backup to
`migration_backups`, `regexp_replace` `--x: var(--x, <lit>)` → `--x: <lit>`
(backreference in the PATTERN asserts self-reference; two-name bridges
untouched), DO/RAISE verify incl. tool-simple as no-op control. **Both RAISE
guards induced first in rolled-back txs** (389's discipline). Applied 18:13:55Z:
self-refs 9/8/7 → 0, simple untouched. Recorded in `schema_migrations`
(`record-only`, per today's estate practice — the runner is blocked by other
threads' pending files, as 391 found). **Literals, deliberately not a
re-bridge**: the site token is `#b59230` gold, whose white pairing is the
contrast lane's open 2.95:1 finding — inheriting it would trade 1.05:1 for
2.95:1.

**Redeploy + verify**: §10b assemble-only deploys ×3 (corr `b7b91228…`,
`354af892…`, `aa0b4d28…`); all three served pages carry the literal and zero
self-refs (stamp-duty lagged one poll — edge cache, settled in <1 min).
Re-probe: **all three buttons `#0b2545` on white, 15.39:1** (control run,
same probe, same session as the 1.05 readings).

**Notified**: `staged_component_build/CONTRIB_2026-08-11_from_mortgagecalculator_ghost_buttons_self_cycle.md`
— distinct class from their 382, one-regex suggestion for their palette-contract
check (`(--[A-Za-z-]+): *var\( *\1[,)]`). Fleet-wide LANDMINES entry added +
synced + verifier fired (`5b2e812b…`).

**Acceptance re-runs ×3 filed** (the pages changed): `e4473518` (equity),
`0fd75a19` (stamp-duty), `0aae22cd` (rate-forecaster), plus the dispatch nudge
(corr `6aa19208…`). Verdicts pending as of this entry; vision_finding
resolution waits on them.

---

## 2026-08-11 (evening, 2) — the homepage "AI slop": NOT the model. A blind audit commissioned it and the site's own voice spec mandated it

**Owner, in chat:** revert the homepage content writer to Gemini ("the copy has
regressed considerably back to AI slop"); then, on the same thread, *"customer
focused in the customer's voice and not all this short too-clever titles like
'Tools that do the bank's maths for you' … much less competitive, softer,
friendlier"*, and *"it keeps changing and it seems to get worse"*.
**Then: "ok, don't change the model then."** The model was NOT changed. Good —
the research says it was never the cause.

### What actually happened, with times

1. **12:31Z** migration `389` (another lane, owner-decided) re-enabled the
   `improvement-sweep`, cost-watched.
2. **17:41:49Z** it swept this site. `design-audit` filed `content_rewrite` +
   `cta_improvement` on `index`.
3. **17:51:19Z** all four homepage components were **CREATED** (not updated) and
   deployed at 17:51:58 — the framework built its own homepage for the first
   time, over an adopted page that had served since July.
4. **~17:54Z** the sweep was disabled again (`enabled=f`), by whoever was
   cost-watching. So the churn engine is OFF as of this entry.
5. **~20:00Z** the owner reads the new homepage and calls it slop. Same page,
   about two hours old.

### The brief the writer was given — read it before blaming any model

`spec` on the 17:41 `content_rewrite` item, verbatim:

- `category`: `differentiation`
- `description`: *"**With no retrievable content**, there is no evidence of any
  differentiator explaining why a user should use this calculator over
  **MoneySavingExpert, Which**, or the dozens of competing UK mortgage
  calculator tools."*
- `acceptance_test`: *"Homepage contains a written value proposition … that
  references a specific feature or benefit **not shared by all generic mortgage
  calculators**"*

**So the copy was commissioned to be competitive.** "See What the Bank's Decision
Engine Sees Before You Apply" is that acceptance test being passed. **Swapping
the model would have produced differently-worded competitive copy, because the
brief and its acceptance test require a competitive claim.** `[MEASURED — the
spec is quoted from the row, not inferred.]`

### Why the auditor said "no retrievable content" — it reads the DATABASE, not the site

`content-quality-auditor`, step `load_page_content`:

```sql
SELECT p.name, LEFT(string_agg(pc.rendered_html,' '),1000) AS content_sample
FROM pages p JOIN page_components pc ON pc.page_id=p.id
WHERE p.site_id=$1 AND p.name IN ('index','about','services','contact') …
```

Before 17:51 the `index` page had **zero `page_components`** (all four rows are
`created_at 17:51:19`). The homepage existed and served fine — it was the
adopted original, outside the framework's tables. So the query returned no rows,
and "no content" became "no differentiation", which became the brief.
**On any ADOPTED site, a page that serves perfectly reads as empty to this
audit** — this lane's whole business is adopted sites. Two further blindnesses in
the same query: `LEFT(...,1000)` judges a site's differentiation from 1,000
characters of raw HTML including tags, and `load_brief` reads
`site_specs.aspect='site_plan'`, which **this site does not have** — so it ran
with `target_audience`, `tone` and `key_messages` all empty and had nothing to
judge by except competitors. `[MEASURED: the load_brief query returns
(none)|(none)|(none) for this site.]`
**Not asserted as a platform root cause here — that wants `090`** (see below).

### And the voice the owner objects to was WRITTEN DOWN, in this site's own spec

`content_direction` (extracted from the ORIGINAL site on 08-02, so the framework
was faithfully reproducing the original author's register):

- `voice.emotional_tone`: *"Challenging … the reader is made to feel slightly
  exposed … **galvanising rather than reassuring** … bad news is never softened."*
- `writing_rules`: *"**Use the lender's voice ('we')** … to simulate an insider
  perspective"* — the exact inverse of "the customer's voice".
- `writing_rules`: *"Use quotation marks around coined institutional terms and
  dramatic labels: 'Flight Risk', 'Knockout Rules', **'The Inheritance
  Destroyer'**"* — the too-clever titles, mandated.
- `things_to_avoid`: *"**Do not write in a reassuring or apologetic tone**"* —
  the owner's request was explicitly forbidden.
- `example_phrases.would_never_say` listed warm phrasing ("We're here to help
  you… with confidence and peace of mind") as banned.
- `writing_rules`: *"**Emoji are used** in navigation cards and homepage feature
  blocks"*.

**The copy is a faithful execution of a spec nobody had re-read since adoption.**
Owner ruling §1 (correctness beats fidelity) licenses moving off the original's
voice, so that is what was done.

### What was changed (config, live immediately, site-scoped, reversible)

`content_direction`, `identity.tone`, `strategy.tone` superseded + re-inserted
(RUNBOOK §8: supersede as a SEPARATE statement). Rewrote only the voice-bearing
keys — `voice`, `heading_style`, `sentence_style`, `cta_style`,
`example_phrases`, `things_to_avoid`, `things_to_emulate`, `writing_rules`,
`persuasion_approach`, `terminology`. **Everything else carried across
untouched**, asserted in the generator: all 5 `compliance_rules`, both
cross-site scope rules (secured lending only; unsecured belongs to
loancalculator.co.uk), `content_depth`, `paragraph_style`,
`formatting_conventions`.

The new rules are **not my taste** — they are borrowed from this estate's own
owner-driven work of the same day:
- sentence ceiling 20 words, average 15, 3+-syllable words under 12% —
  `provocation_readability.go:52-56` (the readability rail shipped today from
  *"readable by a 5 year old … perhaps use ASD-STE100"*);
- "no colon-joined slogan headlines, no ALL-CAPS eyebrows" — house style prompt
  v3 rule 7 (`travelling_docs/pitch_pdf_source/`);
- contractions in ordinary sentences — same prompt, rule 6.
Plus, from the owner's words: never the lender's voice; no coined labels; no
comparison with other sites (a `persuasion_approach.competitive_framing` key
that says so explicitly, so the next differentiation brief has something to
collide with); no emoji; no urgency; headings plain or the reader's own question.
`would_never_say` now quotes **the live headings verbatim**, because the previous
spec had warmth on that list and the strongest available signal is the real
offender, checkable against the page.

**The `formatted` blob is the only field the writer reads** (RUNBOOK §8), so it
was regenerated — and **verified line-for-line against the platform's own
`datahelpers.FormatContentDirection`** in a scratch Go module (138 lines each,
identical after sorting; Go map order is random so order is not a property).
**The comparison was proven able to fail**: perturbing one value in the same run
made it mismatch. In-transaction `DO`/`RAISE` guards asserted the new rules
reached `formatted`, the scope rule survived, and the old instructions
("Inheritance Destroyer", "galvanising", the lender's-voice rule) are gone.

### THE FINDING THAT STOPS THIS BEING FINISHED: half the homepage's words are not the homepage's

The cards are not written copy. `tool-list.items[].title` and
`guide-list.items[].title` in `content_data` hold **each target page's own
title**, rendered verbatim:

- `"Stamp Duty Calculator 2026 — UK SDLT Rates | MortgageCalculator.co.uk"`
- `"Buy-to-Let Guide | The Investor's Reality Check"`
- `"First Time Buyer Guide | The Unvarnished Truth"`

So a perfect homepage rewrite under the new spec **still leaves an SEO title
string with a pipe and the domain name as a card heading**, and still leaves the
clever guide labels — they belong to the other pages. A plain `nav_label`
("Stamp Duty", "Buy-to-Let") is already present in the same rows, but which field
the card renders is a SHARED-component decision (the 252-page `hero`/list family)
and is not a lane-local edit. **Owner decision needed** — see README.

### Deliberately NOT done, and why

- **No rewrite fired yet.** The words live in `content_data`, so a `page_rerender`
  would re-render the same words; changing them needs the writer path
  (`content_rewrite` → `page-build-handler`). **That is the mechanism
  `bugs_open/253` found DESTRUCTIVE today** on the sibling site's homepage: a
  framework rewrite of a homepage prose block kept 84% of the words and **0% of
  the layout classes** (`card` 18→0, `tool-grid` 3→0, `hero` 1→0), and the shrink
  guard passed it because it measures text volume and is blind to markup. Firing
  it here tonight risks trading bad copy for a broken homepage. It should be
  driven once, watched, and checked for component classes — not fired at 8pm on a
  Tuesday because the copy is embarrassing.
- **No `site_plan` aspect invented** to un-blind the auditor: I have not checked
  what else consumes that aspect, and guessing its shape is how a spec that
  "looks applied" steers nothing.
- **Model untouched**, per the owner's later message.

### Owed

1. **A `090` diagnosis run on the audit-blindness class** — "a content audit
   reads `page_components` and reports *no retrievable content* for an adopted
   page that serves HTTP 200, then commissions a competitive rewrite of it".
   Durable, cross-cutting, and about a mechanism outside the symptom: CLAUDE.md's
   own criteria say file it rather than assert it. Distinct from 253 (that is
   about markup loss during the rewrite; this is about the brief that orders the
   rewrite).
2. **Drive the homepage rewrite once, with a component-class check** before and
   after, per 253's fix candidate 1.
3. **The card-title decision** (owner).
4. If the `improvement-sweep` is re-enabled, this site will be swept again — with
   the new spec in place, which is the point, but the differentiation brief will
   still say "differentiate or die". The spec now contradicts it explicitly; which
   wins is **unmeasured**.

### 19:06Z — the churn was still running, and one item is now PARKED

The 17:41 sweep filed a THIRD item nobody had noticed: `needs_content_planning`
(`content-gap-planner`), premise *"No FAQ or guidance content is detectable"* —
false in the same way as the others (this site has four linked guide pages). It
ran 19:00–19:06 and its output was a **new `content_rewrite` on `index`,
priority 35** — i.e. queued AHEAD of everything else and about to rewrite the
homepage again, unattended, through the path `bugs_open/253` measured today as
keeping 84% of the words and **0% of the layout classes**.

**Parked it** (`status='deferred'`, reason appended to `created_by`, flip back to
`triaged` to release): `d1cd9757-7e70-4e76-895f-36033d1be2be`. Not cancelled —
the owner is mid-decision on the voice, and the brief is worth re-reading once
the card-title question is settled.

`sites.locked_at` would have stopped all of it in one switch (RUNBOOK §9) and was
**deliberately NOT used**: owner ruling §3 in force says this site stays
unlocked. Parking one row is the narrow version of the same intent.

**Queue after the park — no further copy rewrite is pending:** `page_rerender` ×16
(re-renders from `content_data`, so it reproduces the existing words rather than
writing new ones), `needs_internal_links` ×4, `needs_imagery` ×2,
`acceptance_run` ×3 (ours, priority 90, behind the rerenders), `needs_rerender`
×1, `audit_tool` ×2. **So the homepage copy is now stable** — the sweep is
disabled, and nothing queued will rewrite it. That is the state the owner asked
for while the voice is settled.

**Also learned about the dispatcher** (extends RUNBOOK §15): the site-selection
query skips any site holding a `claimed` item, and the per-site pickup is
`priority ASC`. Two nudges appeared to do nothing because the planning item held
the site's single-flight slot for 80 minutes; the third worked. **A nudge that
"does nothing" may mean the site is busy, not that the publish dropped** — check
for a `claimed` row before re-firing.

### 19:30–20:00Z — titles and homepage copy rewritten; the owner redirected me MID-FLIGHT and the first pass was wrong

**Owner decision 1:** *"we can change those pages titles and accept the effect on
Google."*
**Owner correction, minutes later — and it caught a real error of mine:** *"The
titles don't have to be 'plain', they still need to have character, just not so
much, not so competitive, not so forceful, not so bold, not so trying to be
clever. More — actual clever but subtle, effective, informative, benefit led for
the user, not so much to do with 'the bank', our capabilities etc but more
focused on what they are trying to achieve by visiting this site."*

**My first pass over-corrected into flat and generic** — "Mortgage fee
calculator", "Remortgaging explained", "Negative equity explained". I had read
the earlier instruction ("not so clever") as "plain", and *plain* is a different
target from *benefit-led with subtle character*. Nothing had shipped when the
correction arrived (titles were in `pages` but `rendered_html` was still stale and
the live site untouched), so the cost was one wasted pass. **The tell I missed:
"benefit led" was already in the owner's first message about the voice spec and I
applied it to the body copy but not to the titles.**

Second pass — each title answers *what am I here to find out*:

| page | before | after |
|---|---|---|
| tool-stamp-duty | Stamp Duty Calculator 2026 — UK SDLT Rates \| MortgageCalculator.co.uk | What stamp duty will cost you |
| tool-affordability | Mortgage Affordability Calculator \| How Much Can I Borrow? | How much you could borrow |
| tool-fee-analyser | Mortgage Fee Analyser \| True Cost Calculator | Which deal works out cheaper |
| tool-overpayment | Mortgage Overpayment Calculator \| Calculate Interest Savings | What overpaying could save you |
| guide-negative-equity | Negative Equity Guide \| The Mortgage Prisoner Trap | If your home is worth less than your loan |
| guide-remortgaging | Remortgaging Guide \| Stick or Twist? | When switching your mortgage pays off |
| guide-mortgage-scorecard | The Secret Scorecard \| How Banks Grade You | Where you stand before you apply |
| guide-how-banks-decide | How Banks Decide: The Underwriter's Guide | Getting your application ready |
| about-index | About MortgageCalculator.co.uk — The UK's Authority on Mortgage Finance | About us |

31 titles, all ≤60 chars, sentence case, no pipe, no domain name — asserted in
the generator, and the "no pipe" guard was **induced first** (it fired on 30 rows
against the pre-change state). `about-index`'s old title was also an unevidenced
superlative claim, which the new spec bans anyway.

**`pages.title` is doing two jobs**, which is the whole reason the cards looked
like that: it is the `<title>` tag AND the visible card heading, because
`tool-list`/`guide-list` render `items[].title` verbatim. The card items hold a
FROZEN copy (both components have `data_sources` NULL, so nothing re-resolves
them), so the same transaction re-pointed every card label at its target page's
`title` **by SQL join, not by retyping** — the card now says exactly what a
rebuild would resolve, and a guard asserts card label = page title for every item.

**15 homepage copy fields** rewritten in `content_data` (hero, both section
headings and intros, the closing CTA), guarded by a check that no
`bank's`/`the bank`/`Decision Engine`/`Scorecard Simulator`/`stress-test`/
`won't tell you` string survives anywhere in the homepage's `content_data`:

- hero: *"See What the Bank's Decision Engine Sees Before You Apply"* →
  **"Know your numbers before you talk to a lender"**
- tool section: *"Tools That Do the Bank's Maths for You"* (the owner's own
  example) → **"The numbers you came to work out"**
- guide section: *"Guides for what the bank won't tell you"* →
  **"Help with the decision you're facing"**
- closing CTA: *"See What the Bank Sees Before You Apply"* → **"Start with your
  own figures"**
- CTAs: *"Run the Scorecard Simulator"* → **"Work out your payments"**

**KEPT deliberately:** `guide-list.cta_heading` = "Not sure which guide applies to
you?" and its subtext. Already addressed to the reader's situation; changing
compliant text would be churn for its own sake.

**Deploy route — and why not the obvious one.** `rendered_html` was stale (it
holds the old words), so an assemble-only deploy would have shipped the old copy;
`content_data` edits are invisible until a re-render. Did NOT fire a content
rewrite (`bugs_open/253`: 84% of words, 0% of layout classes). Instead **released
the `page_rerender` item that already existed for `index`** — `deferred` since
08-03, so a fresh INSERT was refused by `idx_swi_dedup` (correctly: the dedup
index covers non-terminal rows, and `deferred` is one). Flipped it to `triaged`,
priority 40, reason appended to `created_by`. Claimed within a minute of the
dispatch nudge.

**Both lanes brought in, as the owner directed:**
`vigilant_designer_offer_analysis/CONTRIB_2026-08-11_from_mortgagecalculator_the_offer_question_arrived_as_a_copy_complaint.md`.
Their B4 (the offer analyser itself) is **not built yet** and the design critic is
at trial, so they could not be dispatched to write this — the CONTRIB hands them
the graded case instead, and names the seam: **a site's offer is currently
asserted by whichever checker speaks first.** Their own
`missing_conversion_path` finding on THIS site was promoted at 17:43Z, two
minutes after the `content_rewrite` that produced the copy the owner rejected —
the same question, answered by two mechanisms that do not talk.

`migration_backups` holds every previous value under
`titles_2026-08-11b_benefit_led_titles` and `homepage_copy_2026-08-11_benefit_led`.

### 19:39–19:45Z — LIVE and verified, and the 253 check passed with a real baseline

**The assemble-only rerender was the wrong tool and said `complete` anyway.**
Item `c0ab25e1` finished with `rendered_html` untouched: an assemble job
concatenates existing component HTML and re-reads `pages.title`, so the **`<title>`
updated and the body did not**. A `complete` work item is not a repaired artefact
— caught only because the monitor asserted on the HTML, not on the status.

**The right tool was `apply_section_edit`** (`section_editor_actions.go:229`
updates `rendered_html`; `buildRenderContextFromDB` reads `content_data` from the
DB). Fired per slot via the 130 trigger's shape, `edit_type=content_edit`, four
correlations (`778e011d`, `5ebbeee2`, `8103e980`, `d17de0c2`). `items[]` was
deliberately NOT sent — the render context reads it from the DB, so the refreshed
card titles arrived without pushing a large array through Kafka. All four slots
re-rendered 19:39:28–19:39:55 and deployed themselves.

**The `bugs_open/253` check, done properly.** My first attempt counted
`class="card` and `tool-grid` in `rendered_html` and got **0 and 0** — which
looks exactly like the flattening 253 describes. It was **my needles being wrong
for this template** (it uses `tl-card`, `guide-card`), and I only knew because I
baselined instead of reading the zeros. Same measurement both sides, live page
before tonight's changes vs live page now:

| | before | after |
|---|---|---|
| distinct classes | 48 | **48** |
| total class attributes | 88 | **88** |
| `href=` | 33 | **33** |
| `tl-card` / `guide-card` | 6 / 4 | **6 / 4** |
| bytes | 35,676 | 35,589 |
| **classes that DECREASED** | — | **0** |

`[MEASURED 2026-08-11 19:45Z. Disconfirmable: the same script reports per-class
losses, and it was written before the result was known — it is the check that
would have printed "<-- LOST" beside any class 253 predicts.]` The 87 fewer bytes
are the new copy being shorter, not markup going missing.

**Live copy now** (`grep -c "Decision Engine"` = **0** on the served page):

- H1 "Know your numbers before you talk to a lender"
- H2 "The numbers you came to work out"
- H2 "Help with the decision you're facing"
- H2 "Start with your own figures"
- `<title>` "Work out what your mortgage will cost"
- cards: "What stamp duty will cost you", "How much you could borrow", "Which deal
  works out cheaper", "If your home is worth less than your loan", …

**Still owed on this thread:** the other 30 pages' `<title>` tags are updated in
`pages` but only reach their served HTML on each page's next assemble — the
homepage got one because it was the page being re-rendered. Their card labels on
the homepage are already correct (those come from `pages.title` via the refreshed
`items`). A per-page assemble pass (§10b) or the queued `page_rerender` backlog
will carry them; **the guide/tool pages themselves still serve their old
`<title>` until then.**

### 20:1xZ — the staccato pass, corrected: register blended, spec rules fixed at the source, LIVE

**Owner rejected the first warm pass on five specific grounds**, all fair: "for the
figures that matter" is an LLM-ism; "The numbers you came to work out" is *rudely
assumptive*; "Help with the decision you're facing" is *outwardly presumptive*;
"Start with your own figures" is *a direct rude order*; and the whole thing is
*"staccato. Firing short phrases/clauses at me like a machine gun."* Plus: England
not USA, so sentences are more thoughtful and titles more clever than direct; and
on the old "Decision Engine" line — nobody calls it that, so borrowing the term
would have to buy humour or an angle, and it bought nothing. **Users are here for
information and to work out their best mortgage options — take a HOLISTIC view,
price may not be all they want.**

**The cause was my own spec**, logged in `WRONG_CALLS.md`: I set `sentence_style`
from the readability rail's ASD-STE100 thresholds (20 max / 15 avg / one idea per
sentence). Those exist for safety-critical technical instructions read by
non-native speakers. **A 20-word ceiling plus one-idea-per-sentence forbids the
subordinate clause, which is the thing that makes English sound considered rather
than barked.** Research afterwards confirmed the direction: British copy runs to
longer sentences and more complex grammar than American, with understatement and
a claim made once ("Americans like to be sold to, Brits like to be persuaded").

**Research done properly this time** (owner: "research all sorts of copy styles"):
Nationwide (the antidote to presumption is the inclusive conditional — *"Whether
you're a first time buyer or looking for a better deal…"* — which covers the cases
instead of asserting which one the reader is); Which? (nominal headings —
"First-time buyers", "Home movers" — cannot presume because they never address the
reader); MoneyHelper/GOV.UK (public-guidance impartiality). Four registers were
put to the owner — building-society warmth, broadsheet explainer, quiet editorial,
reference/almanac — and the answer was **"a mix"**.

> **A visible error of mine in that presentation:** one option description
> contained a corrupted token (a Cyrillic word for "never" in place of the English).
> Flagged to the owner rather than left to be found, in a set of samples about
> careful writing.

**Spec fixed at the source first** (so the next writer cannot rebuild the
staccato): `sentence_style` now asks for considered sentences of 25–40 words
carried by subordinate clauses and ordinary connectives, with at most one short
sentence per section; `heading_style` gains an explicit `never` (no imperative
headings, no presuming the reader's situation, no borrowed insider terms) and a
`how_to_avoid_presuming` clause; `things_to_avoid` gains five rules including the
named LLM fillers ("the figures that matter", "everything you need to know", "at a
glance", "cut through the noise" …) and "do not lead on price alone". Guard
asserted the old ceiling is gone from `formatted` and the cross-site scope rule
survived.

**Live copy** (all four slots re-rendered via `apply_section_edit`, corr
`00517a8f`/`9b70c36f`/`45eb08f6`/`b168cb2d`):

| | before tonight | now |
|---|---|---|
| H1 | See What the Bank's Decision Engine Sees Before You Apply | There's usually more to a mortgage than the rate |
| tools | Tools That Do the Bank's Maths for You | Calculators for the parts that are hard to picture |
| guides | Guides for what the bank won't tell you | Reading round the decision |
| closing | See What the Bank Sees Before You Apply | If you'd like somewhere to start |

**253 layout check against the pre-change baseline: 48 distinct classes and 88
class attributes both sides, 33 links both sides, ZERO classes decreased.** Bytes
35,676 → 35,915. `[MEASURED 2026-08-11, live page both sides, same script.]`
Banned-phrase count on the served page: `Decision Engine` 0, `figures that matter`
0, `numbers you came` 0, `decision you're facing` 0, `Start with your own` 0.

### NOT DONE — the card DESCRIPTIONS are still in the old voice, and the check found it

The titles are fixed; the blurb under each card is not. Those are the target
pages' **`meta_description`** values, rendered verbatim — the same double-duty
structure as the titles, and the same fix. Still live on the homepage tonight:

- *"Understand what negative equity means, how it traps homeowners, and what
  options are available to escape the **mortgage prisoner trap**."*
- *"A **no-nonsense** guide to buy-to-let mortgages…"*
- *"**Everything you need to know** about remortgaging…"* — which is on the LLM-ism
  list I banned in the spec two hours earlier
- *"Find out how much you can borrow with **our UK mortgage affordability
  calculator**…"* — capability-led, the thing the owner explicitly did not want

Ten are visible on the homepage (6 tool cards, 4 guide cards); ~30 exist. They are
a genuine writing pass, not a mechanical edit, and deserve their own round rather
than a hurried one. **Next session: this is the top item.**

### 20:4xZ — the density correction, and THREE over-corrections of mine in one evening

**Owner, on the copy I had just shipped:** *"'a single number can't settle. There's
no sign-up, and nothing here is selling you a deal.' this bit is llm-speak and
horrible — no one talks about 'selling you a deal', no one says 'a single number
can't settle'."* Both replaced:

> "…and the guides go into the parts that aren't just arithmetic. Everything's
> free to use, and you don't need to sign up for any of it."

The *fact* (no sign-up) was worth keeping; the brochure-voice wording was not.
Guides heading → **"Reading round the subject"** (the owner took the version I had
flagged against myself, which is the argument for flagging your own doubts).

**THE FINDING, and it corrects my rule rather than the copy.** Owner: *"Funnily
enough, with the rest of the copy being more gentle, now 'Help with the decision
you're facing' doesn't sound so intrusive because it's a one off in the whole site
and not part of a constant barrage."*

**Presumption is a DENSITY property, not a property of the sentence.** The exact
heading I had condemned as "outwardly presumptive" reads fine once. What made it
grate was every heading in turn telling the reader what they came for, what they
were deciding and what to do next. My spec had banned the device outright, and an
absolute ban is what produced the flatness that followed. Rewritten as
`heading_style.presuming_is_about_density`: at most one such heading per page, and
rarely; elsewhere name the thing, observe something true, or use the inclusive
conditional.

**And the card descriptions I flagged as the "top item" are FINE — the owner
reviewed all four and said so.** "mortgage prisoner trap", "no-nonsense",
"Everything you need to know", "our UK mortgage affordability calculator": I had
called them defects **because they matched a ban-list I wrote two hours earlier**,
not because anyone reading the page would object. That pass is cancelled. The
filler list is now demoted in the spec to *a smell rather than a crime*, with an
explicit instruction not to hunt existing pages for listed words, and a note that
the owner accepted these on 2026-08-11.

**Three over-corrections in one evening, all the same shape:**

1. Borrowed ASD-STE100 ceilings → staccato (`WRONG_CALLS`).
2. Absolute ban on presumption → flat, characterless headings.
3. Mechanical ban-list application → sound copy reported as defective.

Each time I turned a valid observation into a hard rule and then let the rule
write the copy. **A style rule is a prompt for judgement, not a substitute for it,
and on this site the rules now say so in as many words.**

**The one rule I would keep over any list**, added tonight: *do not write a
sentence no one would say out loud.* Both rejected phrases are grammatical,
on-message, and things no person says. No banned-word list would have caught
either; reading them aloud catches both instantly. The owner's two examples are in
the spec as the worked case.

### 21:0xZ — two rules of English the spec had wrong, both found by the owner reading the live page

**Owner:** *"'and the guides go into the parts that aren't just arithmetic' could
be less negative … also when using words like arithmetic — which aren't common
even though it might be correct — we can't lead into it with a casual 'aren't', it
would be more usual to say 'are not' in this case."*

Both are rules, not preferences, and both were wrong in my spec:

1. **A contraction must match the register of the words beside it.** My
   `voice.formality` said contractions were "welcome and **preferred** in ordinary
   sentences" — too blanket, and it is what produced the clash. An uncommon or
   formal word lifts the register of its clause; a casual contraction next to it
   jars. Either use the everyday word and contract freely, or keep the less common
   word and write the full form, but **never mix the two in one clause**. The
   owner's own pair is now the worked example in the spec.
2. **Say what a thing IS, not what it is not.** A negative definition makes the
   reader do subtraction and reads colder, because it withholds. Added to
   `things_to_emulate`.

Copy: *"…the parts that **are more judgement than arithmetic**."* — positive
(names the thing as judgement), and no contraction beside the formal word.

**A mechanical check now exists for rule 1** in the emit script: no contraction
may sit in the same clause as a word from a formal list. It would have caught
this one before it shipped. `[The list is short and hand-written, so it catches
the shape rather than every instance.]`

**Fourth correction of the evening where the defect was in a RULE I wrote, not in
the sentence** — and every one was caught by the owner reading the live page,
never by the rule itself. The rules improve because the page keeps testing them;
the spec cannot test itself.

---

## 2026-08-12 — three owner observations, traced to three framework defects

Owner on the live homepage: **hero image gone · top nav says just "Home" · cards
have no imagery.** Then, explicitly: *"The point is not to do that manually but to
figure out why the framework didn't do it."* Also flagged as bad copy: *"and you
don't need to sign up for any of it."*

Full write-up in `HANDOFF_2026-08-11_continue_here.md` §11. Evidence trail here.

### The hero — generated, deployed, filed under a template placeholder

The work item's own `result` carries both halves:

```
"hero_url":  "/assets/images/hero.jpg"                    <- what the page asks for
"file_path": "/assets/images/input-data.asset-key.jpg"     <- where the bytes went
```

`input-data.asset-key` is the slugified form of the literal `input_data.asset_key`
— a dotted-path INPUT shipped as a value. Wire test, the decisive one:

```bash
curl -sS -o /dev/null -w '%{http_code} %{size_download} %{content_type}\n' \
  https://mortgagecalculator.co.uk/assets/images/input-data.asset-key.jpg
# 200 68984 image/jpeg          <- the hero image, live, at the wrong URL
curl … /assets/images/hero.jpg
# 404 1151 text/html            <- what the served CSS references
```

**The control matters here**: both URLs in the same breath, one that must be present
and one that must be absent. A 404 alone would only have said "no hero"; the pair
says "the hero exists and the path is wrong", which is a different bug with a
different fix.

`storage.DeployedAssetPath` (`url_helpers.go:317`) returns `hero.jpg` only when
`assetKey == "" || assetKey == purpose`. Given a literal it builds a filename from
it, correctly. The shared derivation (`bugs_open/168`) is sound — its input was a
template that never resolved. Every step reported success, so the item closed
`complete`.

⚠ `image-build-handler` and `asset-deployer` were both updated **2026-08-11
21:52:40Z**, nine hours after the bad deploy, and the config now reads
`"asset_key?": "input_data.spec.asset_key"` — optional marker, different path shape
from the literal that failed (`input_data.asset_key`). **[UNVERIFIED]** whether
that is the fix: `schema_migrations` has no row after 20:00Z on 08-11, so I could
not attribute the change. Resisted the temptation to call it fixed.

### The nav — the data was never the problem

```sql
-- 16 rows: 5 primary, 11 utility, all active, all with a page_id
SELECT g.group_key, i.label, i.url FROM site_nav_items i
JOIN site_nav_groups g ON g.id=i.group_id WHERE i.site_id=<sid>;
```

**MISSTEP, recorded because it nearly ended the investigation early:** my first
hypothesis was "the nav tables are empty and `GetNavItems` fell back to the
hardcoded Home/About/Services/Contact default" (`multipage_actions.go:368`). The
query above refuted it in one shot — 16 rows. Had I trusted the hypothesis and
gone looking for *why the tables were empty*, I would have spent the session in
`PopulateNavTablesAction`, which is entirely innocent here.

What actually discriminated: the **footer carries 16 links and the header carries
1**, same page, same tables. Header renders `primary` only; footer renders
primary+utility. A single-cause theory cannot produce that asymmetry.

Then the per-item join to `pages`:

| primary item | url | build_status | deployed_at |
|---|---|---|---|
| Home | /index.html | deployed | 08-11 20:28 |
| Guides | /guides/index.html | planned | — |
| Investor Tools | /investor/index.html | needs_rebuild | — |
| About | /about/index.html | deployed | 08-11 **19:38** |
| Scorecard | /scorecard-simulator.html | planned | — |

Three never deployed → `ChromeLinkPolicy` drops them rather than ship a site-wide
404, which is right. About deployed at 19:38; the stored header chrome was written
**18:06** — 92 minutes earlier. `loadFetchablePageSet` always injects the site
root, which is why "Home" survives and the bar isn't empty.

**The finding worth more than this site:** chrome is written once behind the
idempotence gate (`render_site_components_action.go:656`), and the only repair
channel, `markStaleChromeLinkSlot`, fires when stored chrome holds a link the
policy **refuses**. A nav *missing* an item has no offending href, so nothing marks
the slot stale. **The nav thins and never thickens.** `chrome_link_policy.go:15-18`
names this very site for the opposite case (a dead CTA); the omission direction has
no channel.

### The cards — right items, wrong status

Every card item holds `"image": ""` — field present, template willing, value never
filled. Zero `<img>` in any card.

```
needs_imagery:section:index:1:icon_stamp_duty        priority 98  deferred
needs_imagery:section:index:1:icon_affordability     priority 98  deferred
needs_imagery:section:index:1:icon_repayment         priority 98  deferred
needs_imagery:section:index:1:icon_scorecard         priority 98  deferred
```

Created by `build-site-planner` **2026-08-02 23:30:20Z**, all 13 `needs_imagery`
rows set `deferred` at **23:31:32.884181Z** — 72 seconds later, one batch,
`handled_by` NULL, `attempt_count` **0**. Never attempted.

Dispatch claims `status IN ('triaged','approved')`
(`claim_work_item_action.go:102`); `TriageDetectedItemsAction` promotes
`detected` → `triaged` only. **Nothing promotes `deferred`.** And `deferred` is not
in `workItemTerminalStatuses`, so it still occupies the `idx_swi_dedup` slot — the
finding cannot even be re-filed under the same `item_key`. Undispatchable and
un-refilable at once.

**[UNVERIFIED]** what deferred them: no Go path writes `deferred` for these types
(only migration `389`, for `contrast_failure`, on 08-11). Four bulk batches exist
(07-31, 08-02, 08-03, 08-05); a hand-park at adoption is the obvious guess and I
did not establish it.

### Misstep 2 — I generalised a card defect from one card

Wrote that "every card description is empty" after reading the first `tl-card` in
the HTML. Counting says **1 of 6** (`tl-card-desc: 6 found, empty=1`) — stamp-duty
only, whose page has no `meta_description`; 9 of 31 pages have none. Corrected in
the handoff where the claim was made. The cheap check that would have caught it is
the one I eventually ran: count all matches, don't read the first.

### Found while looking, not yet named by the owner

**The hero CTA is bare text.** `hero-content` holds `Work out your payments` as a
raw text node — no `<a>`, no `href`. The hero has no working call to action. This
is NOT the §4 ghost-button case (a poisoned CSS var on a real button); here there
is no link element at all.

### Shape of all three

Detection worked every time. The right item was filed every time. The artefact
never arrived — once because the deploy path was an unresolved template, once
because the repair channel only runs one way, once because the item sits where
dispatch cannot select it. **No missing mechanism anywhere.** Three candidates for
`090` are listed in handoff §11.5, alongside §2's still-owed audit-blindness run.

---

## 2026-08-13 — brand-asset census, and a handoff cut mid-task at the owner's instruction

Owner: *"carry on. And also the site currently has no hero image or logo etc."* Then,
mid-turn: fresh chassis build deployed; update docs; hand off if token load is high.
It was — `HANDOFF_2026-08-13_continue_here.md` is the continuation state. Facts
gathered this morning before the cut:

**Every brand asset 404s** (measured in one pass, 2026-08-13):
hero.jpg, logo.png, logo.jpg, favicon.png, /favicon.ico, og-card.png — all 404.
`sites.logo_url` NULL, `logo_text` NULL, `brand_assets = {}`. So "etc." was right:
it is not just the hero — the site serves NO brand imagery at all, while
`needs_brand_head_assets` items for favicon and og_card are `complete` TWICE
(08-05, 08-11). A `complete` work item is not a deployed artefact, again — the
next session should read those items' `result` for the committed file_path.

**The asset rows moved under us** — the owner has been curating: `477838e3` is now
`rejected`, `d6ead260` (the placeholder-path deploy from §11.1) and `9e94250d` are
`superseded`, and TWO heroes are `active` (`0e11c818`, `2e2bea17`, both generated
19:10–19:11Z on 08-11). Any redeploy should ship an ACTIVE row, not d6ead260 —
yesterday's §11.1 trace names d6ead260 and is now stale on that point.

**Chassis v1.0.1294** rolled 09:48/09:49Z both replicas. The provenance startup
line is already out of `--tail=300` on agent-chassis (busy service — the LANDMINES
rotation case, confirmed again). Tag read from the pod spec instead; next session
uses the OCI-label fallback if it needs the commit.

**Not done, in order** (full detail in the handoff §2): migration 399 cutting
248's rung 2 (READ the three-rung ladder code first — my paraphrase is unverified);
maybe 400 (dispatcher purpose mapping); hero live via redeploy-existing or
detection-refile; logo — find what "the old existing logo" is before generating
(owner rejected two generated candidates on 08-11); favicon/og-card diagnosis;
router assignment (site-first, then fleet); card icons only after confirming the
render path consumes them.

## 2026-08-14 (session claude-mcalc-brand) — re-verified the 08-13 handoff, four premises had moved; logo ingested; hero/brand-head deploys in flight

Re-verified before acting (the handoff is a snapshot; this tree moves in hours):

> **CORRECTED 2026-08-14:** handoff §2.1/2.2 — migrations 399/400 are NOT to be written.
> Rung 2 was cut IN GO by the 248 owning lane (`930ace3bd`, 08-12): rung deleted, dotted
> asset_key class-guarded, `assetRowIdentity` makes the asset ROW the authority for
> purpose/asset_key. Live since v1.0.1297 (08-13 22:29Z), verified per-service via OCI
> `image.revision` (`3b0ea20ff`) + `merge-base --is-ancestor` with a negative control, on
> agent-chassis, build-dispatch-loop AND image-generator-adapter; re-verified v1.0.1299.
> Migration 401 (image-build-handler maps asset_id) applied. Numbers 399–406 taken by
> other threads. Bug 248 round 2 resubmitted by its owner 08-14 — nothing left here.

> **CORRECTED 2026-08-14:** handoff §2.3(a) — the "TWO active heroes" were
> `hero_about`/`hero_contact` (page variants), not homepage candidates; every
> `asset_key='hero'` row was rejected/superseded. The premise then expired same-day:
> `claude-session-248-hero-retest-20260814` generated + deployed a THIRD hero
> (`3b0cac59`) to the CORRECT path at 16:55Z as the 248 fix's wire-proof. hero.jpg 200.

> **CORRECTED 2026-08-14:** handoff §2.4 "two generated logo candidates" is paraphrase
> drift (bugfix_210 NOTES §16/§18 is explicit): the owner reviewed generated HERO images
> that read as logos. "The old existing logo" = the original gold roundel
> `images/full-logo.png` in gqls/sites (812×844 PNG, sha `db6ce1aa…`, byte-identical to
> `icon-logo.png`), serving 200 at `/images/full-logo.png`, never in the assets table.
> The original guide pages still render it; only the rebuilt homepage chrome is text-only.

**Why the framework "didn't do it" (owner ask 1, answered):** `site-discovery-rotation-design`
is DISABLED — the deliberate 08-10 cost pause, whose rationale (every generated asset wasted
on the placeholder path) died with the rung-2 fix. Detection cannot re-file anything while
it is off. Second layer: `placeholder_image_in_use:hero` carried 2 strikes (the two 08-11
`complete` rows; strikes = complete+failed in a 7-day window, `load_work_item_actions.go:1276-1284`)
— expire ~08-18. Third layer (favicon/og-card): the two `needs_brand_head_assets` items were
FALSELY completed (attempt_count=0, foreign content-planner JSON) — evidence contributed to
`bugs_open/213` §D; and `derive_brand_head_assets` is logo-gated (no active logo row existed,
ever). Card icons: promoting the 4 deferred icon items is bugs_open/114-class waste — no
per-array-element source resolution exists (`plan_sections_action.go:2076`), so nothing can
write a generated icon into the frozen `items[].image`. Parked with that stated reason.

**Owner decisions today (via question):** hero = reinstate hero_v2 (`9e94250d`) — confirmed
AGAIN after being told v3 went live at 16:55Z (v3 embeds a wordmark that fights the page's
own text overlay; v2 is text-free); discovery returns as a ONE-SHOT for this site only;
fleet rotation re-enable left open for the owner.

**Done this session:**
- Original logo INGESTED through the framework: `scripts/amend-asset.sh` → staging
  `a8976eb4` ingested → assets row `e766370e` (asset_key=logo, purpose=logo, ACTIVE,
  origin uploaded/operator-supplied, 12,325 B, 812×844 — size+dimensions+action-side
  sha256 all match the source; bucket URL is private so the row metadata is the byte proof).
- **Caught bugs_open/213 §D / 274 live on my own item:** the amend item completed with a
  SUBSTITUTED content-planner payload and attempt_count still 0 — child succeeded
  (artefact persisted), its reply hit 274's cannot-deliver at 20:15:55Z, parent completed
  the item anyway with a foreign result. Correlation `aec9d3ed` survives for tracing.
  Contributed to 213 §D (with a same-evening visible correction of my own wrong
  attempt_count inference — also in WRONG_CALLS). **Consequence for THIS lane: item
  statuses/results are untrustworthy while 274 fires — verify every deploy at the URL.**
- Hero curation per owner: `3b0cac59` superseded (kept), `9e94250d` ACTIVE. NOTE:
  `idx_assets_site_asset_key_unique` — one active row per (site, asset_key); supersede
  before activate or the swap aborts.
- Three items filed 21:34:15Z, all `triaged`: `deploy_amended:logo` (deploy e766370e →
  logo.png), `needs_brand_head_assets:derive_20260814` (spec {"mode":"brand_head"} — the
  LIVE check_mode conditional already tests input_data.spec.mode; the 07-11 seed file does
  not — read live config, not seeds), `undeployed_asset:9e94250d…` (redeploy hero_v2 →
  hero.jpg). **In flight at the time of writing — wire-verify at the four URLs.**

**Still owed (tasks 4-6):** router assignment site-scoped (397 header SQL is fleet-wide,
add the site predicate; NO SQL exists for the unsatisfiable arm — write it; routers'
summary_from is a template-string defect, cosmetic); one-shot design discovery AFTER
assets land (mirror oneshot-design-discovery-rh-20260730: topic system.agent.scheduled.requests,
interval 86400, timeout 900, input_data {domain, site_id}, disable after firing); the bad
copy line via apply_section_edit content_edit on the hero slot (never content_rewrite).

### 2026-08-14 (same session, outcomes) — everything landed; wire-verified at the URLs

- **21:39:07Z logo.png 200** (deployer serves it at 385×400 from the 812×844 original — resize
  is the deploy render, byte-source verified at ingest). **21:39:39Z favicon.png 200 (64×64
  roundel), og-card.png 200 (1200×630 roundel on light ground), hero.jpg bytes changed
  96,755 → 68,984 = hero_v2.** All four looked at as images, not just statuses. Head references
  `<link rel="icon" href="/assets/images/favicon.png">` and the og:image URL. The placeholder
  litter `input-data.asset-key.jpg` still serves, unreferenced — 248's backlog-drain owns it.
- **Routers wire-proven on live items** (IMG-071 condition met): all 3 `image_url_404` →
  escalate branch (asset_facts.backing=has_asset resolved correctly through the object
  flatten); 2 `image_source_unsatisfiable` → mappable/hero → filed needs_imagery, and item_key
  DEDUP capped both runs at ONE filed item (`from_unsatisfiable:` key is domain+source-scoped).
  **But their premises had expired 3 minutes earlier** (this session's own deploys) — so: the
  1 filed generation request cancelled with reason; the 3 spurious escalations cancelled with
  reason; the remaining 15 stale unsatisfiable items closed `cancelled` with the evidence
  inline (site_assets.hero resolves; the 21:46Z discovery run did not re-file the class).
  ⚠ **Lesson for the fleet's ~73: route findings only after a fresh discovery pass on that
  site — a pre-fix finding routed post-fix files noise and (dedup-capped) spend.** Fleet
  assignment deliberately NOT done; recommend per-site discovery-then-route.
- **One-shot design discovery fired 21:46:17Z, COMPLETED in 6s, and the negative result is
  the proof:** zero re-filings of needs_hero_image / needs_logo / needs_brand_head_assets /
  image_url_404. It filed forward work only: ~10 `needs_imagery:page:tool-*:content_hero_*`
  (page-scope tool heroes — the CONSUMABLE class, page-plan joins exist), `stale_chrome`
  rerender, `deactivated_head`, tool evaluations, and `undeployed_asset:e766370e` (nothing
  references logo.png yet — redeploy is a harmless same-bytes write). **Expect ~10 paid
  tool-page hero generations to flow via triage in the coming hours — cancel them if unwanted.**
  Row disabled after firing (`enabled=f`), matching the rh/quality precedent.
- **Copy live:** homepage subheadline now ends "It's all free, and there's nothing to sign up
  for." (section_edit → section-editor, same pass; verified on the served page).
- Deferred logo-generation item `needs_imagery:site:-:logo` cancelled (owner prefers the
  original; executed).
- Still open on this lane: header shows the text wordmark, not the roundel `<img>` (chrome
  work, parked behind the thinning bug — `stale_chrome` item now detected may move it);
  the 30 stale titles; card icons (parked, 114-class, needs a component-field change);
  fleet router assignment (above); fleet design-rotation re-enable (owner cost call).

## 2026-08-16 (session claude-mcalc-brand, resumed) — hero items ran; 274 fix verified live; a NEW records defect found and filed (287); the header now shows the roundel

Owner instructions this turn: let the hero items run; check the 213/274 fix is in the live
chassis (both CLOSED 08-15); update docs; carry on; hand off if token load is high.

- **The 10 tool-page hero items ran** (08-15 19:31–19:43Z): 10/10 `complete`, 10 active
  `content_hero_tool_*` assets. Not eyeballed — page-scope assets on tool pages, the owner
  said let them run. `[UNVERIFIED that each tool page renders its own; check one]`.
- **274 fix IS in the live chassis** (v1.0.1303, rolled 08-15 18:45Z): OCI `image.revision`
  `5e075a6f9…`; both `919cc6976` (envelope) and `3ba384c63` (WFA-014 park carry) are ancestors,
  control ok. **Behaviour holds:** 0 cannot-deliver rows against 859 child completions since
  the roll. The 213/274 mechanism that substituted a content-planner payload into my logo item
  cannot recur.
- **BUT a different substitution shape appeared WITH that roll and dominates since —
  filed as `bugs_open/287`.** Items complete with the SPAWN RECORD (`{role,topics,agent_id,
  agent_type}` = `handler_spawned`) as `result`, not the handler's reply, while the child
  completes and the parent's own `handler_result` ends up correct. 0 instances in any hour
  before 10:14Z 08-15; ~270 vs 70 correct since 18:46Z. Work is done, record is wrong. One of
  them is this site's `undeployed_asset:e766370e` (logo redeploy, 19:12Z) — logo.png was and is
  200. `090` filed: run `fb7ae3bc-e9bf-4a96-b540-d593b91bc79c` (⚠ trigger warned local HEAD is
  853 commits ahead of origin, which the loop reads — it may not see WFA-014).
- **Live site today:** logo/favicon/og-card/hero all 200 (bytes unchanged from 08-14). **The
  header now renders the roundel as `<img src="/assets/images/logo.png">`** and the nav has
  About — the `stale_chrome` rerender the 21:46Z discovery filed ran 08-15 18:40Z and consumed
  the logo. The chrome-thinning "parked" item resolved itself with the asset present; the
  header-image scope note from 08-14 is closed. `sites.logo_url` is still NULL — the chrome
  reads the asset, not the column `[INFERRED from the render; column unread]`.
- **My hero copy line was overwritten** 08-15 18:50Z by `site-review_content_rewrite_index`
  (the improvement loop, unasked). New line: *"No sign-up, no upsell, and no personal data
  collected to use any calculator."* — a sentence a person would say; layout classes intact
  (253's failure mode did not bite here). Not chased; recorded. The owner's §0.4 ask is
  satisfied either way.
- Discovery items from 21:46Z all `complete`; one `content_rewrite` (tool-bridging-compound)
  in `needs_human_review`.

**Handoff for a fresh chat is `HANDOFF_2026-08-16_continue_here.md`.**

## 2026-08-16 (afternoon, fresh session) — the stale-title item was already DONE; and the site is serving eight broken internal links nothing is watching for

Picked up `HANDOFF_2026-08-16_continue_here.md` §4. Worked the two items it left as
actionable and measured both before acting. One was already finished; the other turned up
a live defect the handoff does not mention.

### 1. §4.5 "the 30 stale `<title>`s" is CLOSED — carried forward unmeasured for five days

[MEASURED 2026-08-16 15:47Z, live HTTP against every deployed page] **27 of 27 deployed
pages serve a `<title>` byte-identical to their `pages.title` row.** Zero stale. The
08-11 finding ("the other 30 pages still SERVE their old `<title>`") was true when
written and was overtaken by the 08-15/08-16 rerender traffic — 24 of the 27 pages were
re-deployed 08-15 08:39Z–08-16 11:28Z for other reasons, and an assemble re-reads
`pages.title` (08-11 §7's own mechanism, working as documented).

The check that would have come out otherwise: it compares the SERVED `<title>` to the DB
row per page, so any page whose row had moved without a rerender fails it. The loop is in
this session's scrollback; it is one `curl | grep -o '<title>[^<]*'` per row of
`SELECT url,title FROM pages WHERE build_status='deployed'`.

**The misstep worth recording is the handoff's, and I nearly repeated it:** the item was
listed as *"mechanical, unchanged"* — a status inherited from 08-11 and never re-measured
across five days of heavy rerender traffic on this exact site. Had I acted on it I would
have fired 30 rerenders to change nothing. **A carried-forward "unchanged" on a site with
a live improvement loop is a claim about the past, not a measurement.**

### 2. NEW, live, and unfiled: four dead internal-link targets, eight link instances, seven pages

[MEASURED 2026-08-16 15:52Z, live HTTP; every 404 re-probed 3× before being believed]

| target | live | link instances | on pages | page row |
|---|---|---|---|---|
| `/scorecard-simulator.html` | 404 | 4 | first-time-buyer, how-banks-decide, disclaimer, tool-affordability | `planned`, **`in_header=t`, `in_footer=t`** |
| `/guides/mortgage-scorecard/index.html` | 404 | 2 | how-banks-decide, games/fact-finder | `planned` |
| `/guides/lender-restrictions/index.html` | 404 | 1 | how-banks-decide | `planned` |
| `/tools/rate-forecaster/` | 404 | 1 | tools/rate-scenarios | target EXISTS and is deployed — the href is directory-form (§4) |

⚠ **A single-shot 404 inside a fast scan is not evidence.** My first sweep also reported
`/games/fact-finder/index.html` as 404; re-probed 3× it is 200 every time, and the page
serves its full body. One transient inside a ~28-URL burst. Everything in the table above
survived triple probing; fact-finder did not, and is fine.

### 3. Why the framework did not catch them — the §2 answer needs a second rotation added

The handoff §2 layer 1 says detection was off, and names `site-discovery-rotation-design`.
**`site-discovery-rotation-completeness` is ALSO `enabled=false`, last triggered
2026-08-10 17:40Z** — and it is the one that owns link integrity: `phantom_internal_links`
(DB-derived) and `dead_internal_link_live` (live-probing), among 42 checks.

[MEASURED] `site_discovery_rotation` for this site: `completeness-discovery-agent`
last selected **08-09 20:56Z** — 6.8 days. Every one of the eight links above was authored
or re-deployed after that pass except the `/scorecard-simulator.html` pair, and **that pair
WAS detected**: two `unbuilt_internal_link` items filed 08-09 20:56Z, both sitting in
`needs_human_review` ever since. So on this site the class divides cleanly:

- **detected and parked** (2 links) — the framework did its job and nobody answered it;
- **undetected** (6 links) — filed by nothing because the detector has not run since.

**This is NOT a new finding about the rotation and I am not filing one.** The disabled
state is already recorded by `bugfix_203_phantom_cta_cleanup/NOTES`,
`vision_finding_revalidator/HANDOFF_2026-08-11_pre_plan.md` and `bugs_closed/270`, which
hit it as a blocker on 08-16 and worked around it with a direct per-site dispatch. What is
new is the *cost of it on this site*, measured above.

### 4. ~~NEW CLASS, fleet-wide~~ **NOT NEW** — on the B2 route every directory-form URL is a live 404, and no DB-derived check can ever see it

> **CORRECTED 2026-08-16, same session, ~40 minutes after I wrote it.** I called this a new class and > it is not: `LANDMINES.md` has carried it since before today — *"A `/section/` URL 404s on every > B2-hosted site, and a local server hides it"* — with the worker lines, the second-order trap, and > **`mortgagecalculator.co.uk/guides/` already named in its confirmed list.** What caught it: I went > looking for the FIX location, grepped for `worker.js`, and landed in the middle of the entry I had > just failed to find. **My prior-art grep was `trailing slash|directory-form|directory form` — three > phrasings of the SYMPTOM, none of which appears in an entry written about `/section/` URLs.** The > footprint symbols (`worker.js`, `NormalizePagePath`) would have found it first try, which is exactly > what MEMORY's `grep-landmines-for-your-symbols` says and what I did not do. Logged in `WRONG_CALLS.md`.
>
> **What survives as genuinely new** is narrower and is now an ADDENDUM to that entry rather than a rival > to it: (a) the five DB-derived mechanisms that collapse the two URL forms via `NormalizePagePath`, which > makes the existing entry's closing rule — *"if a checker disagrees with production, change the checker"* > — read differently once you notice PRODUCTION does the forbidden normalisation itself; (b) the git-route > contrast (`relojistas.com/noticias/` 200), which shows the platform already serves this correctly on one > of its two routes; and (c) the fleet census below. The measurements in this section stand; only the > word NEW was wrong.


The `/tools/rate-forecaster/` row above is not a bad link in the ordinary sense — the
target page exists, is deployed, and serves 200 at `/tools/rate-forecaster/index.html`.
The href is written in directory form, and **this site serves nothing in directory form**:

[MEASURED 2026-08-16 15:55Z] `mortgagecalculator.co.uk` — `/guides/` 404, `/about/` 404,
`/tools/repayment/` 404, `/investor/` 404, `/` 200. Same on two more `github_repo=''`
(B2-route) sites: `gaswholesalers.com/tools/supplier-comparison-calculator/` 404 vs
`…/index.html` 200; `leopardessconsulting.co.uk/tools/automation-savings-estimator/` 404
vs `…/index.html` 200. **Contrast — the git route serves it correctly:**
`relojistas.com/noticias/` 200 AND `/noticias/index.html` 200 (`github_repo='vm-sites'`).
So this is a property of the hosting route, not of the link, and the platform demonstrably
can serve it right on one of its two routes.

**Why no DB-derived check will ever flag it:** `datahelpers.NormalizePagePath`
(`links.go:215-227`) strips `index.html` and then trailing `/`, so `/tools/rate-forecaster/`
and `/tools/rate-forecaster/index.html` collapse to the same key. `PageURLSet.Contains`
therefore returns true, and every consumer of that set agrees — the build-time resolver
(`loadResolverPageSet`), the deploy gate (`validate_page_content`), `RepairPageLinks`, and
`check_phantom_internal_links`. One normalisation, five mechanisms, all blind together, all
correct on the git route and all wrong on B2. `dead_internal_link_live`
(`check_site_structural_validity.go`) is the one check that would catch it, because it
probes the href as written — and it lives in the disabled rotation of §3.

[MEASURED, fleet-wide] Only **two** directory-form internal hrefs exist in any deployed
`page_components.rendered_html` today: this one, and `relojistas.com /glosario → /noticias/`
— which is on the git route and serves 200. **So the live damage of the whole class today
is one link, on this site.** Recorded as a landmine rather than filed as a bug: the
authoring path emits this shape rarely, and the acute cost is one href.

⚠ The wider exposure is not internal links at all — it is every human, external site or
search result that reaches a B2 site with a trailing slash. That is unmeasurable from here
and is left as a stated exposure, not a claim. Canonicals are NOT part of it: the pages
emit `…/index.html` canonicals, matching what is served [MEASURED: three pages].

### 5. Acting on it — owner chose "build the pages"; two built and live, the third is blocked by `bugs_open/260`

Put the fork to the owner with the measurements (build the three planned pages vs retarget the
links at pages that exist). **Owner chose: build the three pages.** Rationale recorded at the time:
the hrefs are already correct, so building fixes the links *without editing a line of the copy this
lane spent weeks tuning* — which on a site whose improvement loop rewrites hero copy unasked is the
material advantage, not a stylistic one.

**Nothing was hand-built and nothing was hand-armed beyond status.** The site's own queue already
held the right items, deferred since 07-31/08-02:

| item | id | outcome |
|---|---|---|
| `needs_page:guide-mortgage-scorecard` | `7fd27e59` | armed `deferred`→`triaged` 16:14:55Z → **complete 16:26:21Z**, page deployed 16:26:13Z |
| `needs_page:guide-lender-restrictions` | `dde1c0fc` | armed with it → **complete 16:31:50Z** |
| `gap_plan_add_scorecard-simulator` | `0c65f9fa` | **already in flight when I arrived** (claimed 16:10:14Z by `build-dispatch-loop`, unprompted) → `needs_human_review` 16:18:01Z |

[MEASURED] Both new pages serve 200 with correct chrome, `<title>` matching `pages.title`,
`…/index.html` canonicals, no template leak: `lender-restrictions` 22,486 B / 444 words,
`mortgage-scorecard` 22,919 B / 526 words. Copy is in the house voice
(*"Some things narrow your choice of lender before the rate ever comes into it"*).

**Nothing else on the site was armed** — the armed set was empty when I started (`triaged`/
`approved`/`detected` = 0, one `claimed`), so §10c's backstop was unnecessary and no other lane's
items were touched.

#### 5a. The scorecard-simulator build is refused by `bugs_open/260`, and I contributed the instance

`validate_content` → **20 blockers, 0 errors**, all `unrendered_template`/`unrendered_template_block`
on the `mechanism-flow` component: `{{if .eyebrow}}<span class="mech-flow__eyebrow">Before the
decision</span>{{end}}` — directives intact, field values substituted, which is 260 §1's verbatim
fingerprint. Item parked, nothing written (260 §4's gate holds). Contributed to 260 §10 with a
census isolating the defect from the other 8 issue types under the same `error_code`: **11 events,
4 domains, 10 work items, 08-11→08-16**, five of them since 260 was filed, and this site is 6 of the
11. 260 is actively owned (council trail 08-14) — contributed, did not compete.

#### 5b. ⚠ The count of live 404s went 8 → 7, NOT 8 → 1, and the reason is worth knowing

I expected building the pages to fix 7 of 8. It fixed 3. **The two pages the framework just wrote
each link to `/scorecard-simulator.html` themselves** — because the site's design intent names the
Scorecard Simulator as an expected page, so every writer that reads the intent links to it. Live
instances of that dead href went **4 → 6**.

[MEASURED 2026-08-16 16:33Z, full re-audit of all 29 deployed pages, 28 distinct internal links]

| target | before | after |
|---|---|---|
| `/guides/mortgage-scorecard/index.html` | 2 dead | **200, 2 links resolve** |
| `/guides/lender-restrictions/index.html` | 1 dead | **200, 1 link resolves** |
| `/scorecard-simulator.html` | 4 dead | **6 dead** (2 added by the new pages) |
| `/tools/rate-forecaster/` | 1 dead | 1 dead (directory-form; not hand-fixed, see §4) |

So the honest read: **the queue that `260` buries is self-fuelling on this site.** Every page the
framework writes adds another link to the one page it cannot build. That is a stronger argument for
260's fix than anything in my §10c contribution, and it is measured rather than argued.

⚠ **A 404 immediately after a deploy may be the CDN's cached negative, not the truth.** The
`mortgage-scorecard` page returned 404 on the first probe 40s after `deployed_at`, and 200 with a
`?cb=` buster in the same second (`cache-control: public, max-age=300` on the worker's 404 arm).
`how-banks-decide` did the same during the re-audit and is 200 on three plain probes since. **Probe
twice, second one cache-busted, before believing a post-deploy 404** — otherwise you will file a
deploy failure that never happened.

## 2026-08-17 (morning) — owner directed three things: the rotation is ON, the `.xcf` is GONE, and the worker fix was explained rather than applied

Same lane, next session. The owner read `HANDOFF_2026-08-16b` §5 and answered its three open
decisions. Two were actions and are executed; one was an explanation.

### 1. `site-discovery-rotation-completeness` — ENABLED 11:31Z

**Measured the blast radius BEFORE flipping it, because this exact task carries the
demand-control trap** (LANDMINES: never `enabled=true` a scheduled task to test it — its
`pre_query`'s recency predicate can select 0 rows and read exactly like a clean run). Read the
`pre_query` first:

- `LIMIT 1` — **one site per tick**, `interval_seconds=3600`.
- Eligible = `status IN ('active','deployed')` AND last selected > 7 days ago.
- **Skips any site holding a `claimed` build item** (`pipeline='build'`) — so it cannot pile onto
  a site already working.
- Stamps `site_discovery_rotation` in the same statement, so it round-robins rather than
  re-picking the same site.

[MEASURED before enabling] **22 of 23 active sites due**, one never checked at all
(`remortgagecalculator.uk`), the rest last selected 08-09. So the pre_query could not have
returned 0 — the demand control was satisfied in advance, not asserted afterwards.

[MEASURED after enabling] **First tick 11:32:14Z, `sites_due` 22 → 21**, and
`orchestration_states` shows `completeness-discovery-agent` COMPLETED at 11:32:18 — 4s.
So trigger → agent → completion is proven end to end.

⚠ **The first pass is NOT evidence about the checks, and I nearly recorded it as if it were.**
It drew `remortgagecalculator.uk`, which has **0 pages** — `count(*)` on `pages` for that site
is zero, deployed or otherwise. A discovery pass over nothing returns zero findings whatever
the checks do or do not detect, which is the same could-not-have-come-out-otherwise shape
`WRONG_CALLS.md` keeps recording. **The first informative tick is ~12:32Z (`robot-hands.com`,
last checked 08-09).** Judge the rotation there, not here.

⚠ **The thing to watch is downstream of the rotation, not the rotation.** Findings insert at
`detected`; `detected-item-promoter` is live on a 15-minute cadence and moves them to `triaged`,
where handlers dispatch. One completeness pass on `leopardessconsulting.co.uk` filed ~77 items
(`bugs_closed/270`, 08-16) at a time when they were inert. They are not inert now. Twenty-two
sites of that is real work over ~a day. Reversal is one statement and it is in the handoff §0a.1.

### 2. The `.xcf` — removed from BOTH the bucket and the deploy source

Flagged three times since 07-31, never actioned; 175,232 B of GIMP master publicly served.

**The part that mattered and was nearly missed: the bucket copy is downstream of a local
directory.** `b2 ls` showed it re-uploaded **2026-08-16 20:41:52Z** alongside `full-logo.png` and
`icon-logo.png` — i.e. after this lane's previous session. Deleting only the bucket would have
been undone by the next sync. Traced the source to
`~/projects/sites/mortgagecalculator.co.uk/images/` (the authoritative deploy dir per LANDMINES'
`sites` vs `sites2` entry), removed it there too, committed in **that** repo as `7c9078f20`.

- `b2 rm --dry-run --versions` first → exactly 1 match, exit 0 (RUNBOOK §1's flag gotcha: `--dry-run`,
  not `--dryRun`, and check the exit code because a failed run's piped output reads as a clean no-op).
- Then `b2 rm --versions`, exit 0. Bucket listing after: the two PNGs, no `.xcf`.
- **Verified gone:** 404 on three cache-busted probes and one plain (probing twice with a buster is
  the 08-16 CDN negative-cache lesson applied in the opposite direction).
- **Verified harmless:** 0 of 29 live pages reference it; 0 rows in `page_components`,
  `site_components` or `assets`. Deleting it broke no link.
- **Four byte-identical copies retained**, all sha256 `78a635bb…`: the deploy repo's own history
  (`65d06ef4e`), `~/projects/domains/mortgagecalculator.co.uk/images/`,
  `~/Downloads/mortgagecalculator/`, and `/home/ant/mortgagecalculator_asset_backup/` +
  `SHA256SUMS.txt`.

**One RUNBOOK claim narrowed by evidence:** §2 warns that `curl` does not return origin bytes
because Cloudflare injects into the response. The curl copy taken 08-16 and the B2 copy taken
today are **sha256-identical**, so that injection is specific to `robots.txt` (as §2's own
evidence actually said: 28 of 28 non-robots files matched) and does not touch binaries. The habit
of taking bytes from B2 before an irreversible delete is still right; the reason is
belt-and-braces, not a live hazard for this file type.

### 3. The worker fix was EXPLAINED, not applied

`scripts/cloudflare/worker.js:9-12` — the three-line directory-index rewrite from
`HANDOFF_2026-08-16b` §4. Owner asked for the explanation; nothing was changed in the repo or at
Cloudflare. It remains a 36-zone shared-serving change → owner + council.

### 4. The `facts` declaration seeded — all 13 ids — and the "one config step" was not one step

Owner: *"seed it for real"*, answering the `register_guards_code_phase_b` ask at the top of
`HANDOFF_2026-08-16b`.

**All 13, not the 225 pair.** Their CONTRIB offered the smaller option. [READ] The honest
declaration is 13: `verify_criteria.py:load_register_bands()` hard-requires every one of the 13
(`need` list, lines 139-145) and `banded()` consumes each in the model (lines 159-187), so the
fence encodes 13 facts. Declaring 2 would understate the tool's dependencies to buy a smaller
burst. [MEASURED] all 13 present with values in this site's current `evidence_base` — the
declaration is live, not inert.

**⚠ `install_fences.py --only stamp-duty --apply` REFUSED, and quietly.**
```
SKIP     stamp-duty         not ladder-eligible on this site — a PLAN here would never be read
```
[MEASURED] `tool-stamp-duty` is **2 components, 0 at `component_level='tool'`** → fails the
sole-component clause of the ladder's eligibility predicate. Following the CONTRIB literally would
have produced a clean-looking run, no error, and no `facts` key.

**The guard's premise is what CLM-022 made stale, so I narrowed the guard rather than the rule.**
Its refusal is justified in its own docstring as *"a PLAN here would never be read"* — false since
Piece 3, which resolves declaring PLANs by the name rule Tier 4 uses, deliberately NOT by the
eligibility predicate (their CONTRIB point 2 says why: that predicate misses exactly these tools).
New `--allow-ineligible` requires **both** a `facts` declaration in the criteria doc **and** an
existing current `doc_plans` row, so the subject key is **inherited** rather than constructed from
a page name — rule 1's silent-permanent-failure mode. Verified it still refuses without the flag.

**Verified at the artefact, not the exit code.** New body `400657e0…` vs superseded `c3eaf877…`:
`diff` is **the 15-line facts block and nothing else** — 4 checks, 4 assertions, whole prose body
byte-identical. Fence re-parsed out of the stored body: 13 unique non-empty strings,
`no_auto_fix` still true.

**⚠ Their `dryrun_fact_drift.sh` uses the `kubectl run -i … kcat -P <<JSON` stdin-race form**
(LANDMINES: ~4 publishes lost in 5, exit 0, no receipt either way). Used the safe base64-into-
`--command` form with a `PUBLISH_OK` receipt instead; both dry runs landed first time. Told them.

#### 4a. The induced proof is STAGED, NOT RUN — and the clean run proves nothing on its own

- **Code live at the binary**, both controls in one exec on `agent-chassis-5657f446c7-q7b82`:
  `fact_drift_review` **2**, `stale_attestation` (positive) **5**, an impossible string
  (negative) **0**.
- **Steady-state dry run clean**: corr `5763c238-1faf-4e6e-9cf1-a5f2e4e56130`, COMPLETED
  12:09:50Z, `dry_run:true`, 1 site, **no `fact_drift` key**, 13 facts all `outcome: fresh`.
- **Step 2 (supersede `sdlt-ftb-relief-cap` 500000→550000, dry-run, restore) was REFUSED by the
  session permission layer** — it rewrites a live tax figure on a public site. Reasonable refusal;
  not worked around. Both directions were staged first (`scratchpad/mutate.sql`, `restore.sql`);
  the restore flips `is_current` back onto the ORIGINAL row `2303a6f7…` rather than re-inserting a
  copy, so it restores exactly rather than approximately.
- [MEASURED after the refusal] register untouched: `2303a6f7…` still current at **500000**, and
  `count(*) WHERE created_by='mcalc-lane-induced-test'` = **0**.

**So CLM-022 on this site is DECLARED and READ, not PROVEN TO FIRE, and nothing in our docs says
otherwise.** A clean steady-state dry run is equally consistent with a live mechanism and an inert
one — the could-not-have-come-out-otherwise shape `WRONG_CALLS.md` exists to catch. The one result
that would discriminate is the one still owed.

### 5. The rotation's second tick answered §1's open caveat — the checks fire, and they found real damage immediately

[MEASURED 12:33Z] Tick 2 (12:32:43Z) drew `robot-hands.com` — **35 deployed pages of 41**, so
unlike tick 1 this could have come out either way. It filed **87 items**: `head_essentials_missing`
29, `page_rerender` 20, `undeployed_asset` 17, `needs_internal_links` 6, **`dead_internal_link_live`
4**, `literal_markdown` 4, `needs_content_image` 2, `canonical_mismatch` 1, plus
`deactivated_component`, `nav_drift`, `orphan_blog_posts`.

**The four `dead_internal_link_live` findings are real 404s**, every one linked from that site's
`/tools.html`: `gripper-cycle-time-estimator.html`, `gripper-payload-calculator.html`,
`privacy.html`, `terms.html`. Not trailing-slash cases — ordinary dead links, sitting live.

Two things follow. **(a)** The caveat recorded at §1 was worth recording and is now discharged
honestly: tick 1's zero proved nothing, tick 2's 87 proves the checks work. **(b)** This site was
not a special case. The week `site-discovery-rotation-completeness` spent disabled was hiding real
broken links across the estate, and the first real site swept had four. Twenty-one sites remain
in the backlog at roughly one an hour.

## 2026-08-17 (afternoon) — the induced proof RAN. The fan-out is live; the `value_drift` arm is still unproven, and cannot be proven on a freshly-seeded fence

Owner gave permission for the one write, and told us a fresh chassis build had rolled.

**Re-verified the code on the NEW binary first** — a fresh build is exactly when a null result
would be misread as "the mechanism is broken". Both replicas (`agent-chassis-5bd56bdd9b-6sb8t`,
`-jzmns`), same exec, three greps: `fact_drift_review` **2**, positive control
`stale_attestation` **5**, an impossible string **0**. Pods 91 min old, so past the ~300s
post-restart dispatch hole.

### The run

Window open **14 seconds** (16:17:36→16:17:50Z). Mutate → dry-run → restore, with the restore in
a `trap … EXIT` so it fires on any exit path including a failed publish or a hung wait. Restore
flips `is_current` back onto the ORIGINAL row `2303a6f7…` rather than re-inserting a copy.

[MEASURED] Register after: original row current, `sdlt-ftb-relief-cap` **500000**, `pinned` **t**
(carried, CLM-001). Induced row exists but `is_current=false` — 0 current. **0
`fact_drift_review` work items exist fleet-wide**, so the dry run wrote nothing, as designed.

### The result — baseline vs induced, the only comparison that discriminates

| run | `results[0].fact_drift` | kind | route | `new_value` for `sdlt-ftb-relief-cap` |
|---|---|---|---|---|
| baseline (register 500000) | **13** | `unreconciled_declaration` | `fact_drift_review` | **500000** |
| induced (register 550000) | **13** | `unreconciled_declaration` | `fact_drift_review` | **550000** |

**What this PROVES.** The fan-out resolves the declaration on a tool the acceptance ladder cannot
see (`tool-stamp-duty`, 2 components, 0 tool-level) — the exact worry that made the other lane
avoid keying on `toolEligibilityWhere`; it resolves the right page (`page_id 3d7d0d72…`,
`page_name tool-stamp-duty`); it routes every one of the 13 to `fact_drift_review`, correct for a
`no_auto_fix` fence; and **it reads the register AT CHECK TIME** — the two runs differ in exactly
the one value I changed, and nothing else.

**What this does NOT prove, and the reason is structural.** Both runs report
`kind: unreconciled_declaration`, `reason: never_reconciled` — **not** `kind: value_drift`, which
is what the other lane's RUNBOOK step 3 says to expect. On a freshly-seeded declaration every
(fact, tool) pair is in the never-reconciled state and that arm takes precedence, so **the
`value_drift` arm is unreachable by induction until a REAL (non-dry) sweep has recorded the 13
baselines.** Their step 3 expectation cannot be met on day one of a seeding, and a lane following
it literally would read a correct result as a failure. Told them.

### ⚠ THE TRAP, and I walked into it and said so out loud before catching it

`fact_drift` is **per-site, nested in `results[]`** — `refresh_result->'results'->0->'fact_drift'`.
The top-level `refresh_result` has no such key, and its `total_drifted` counts **citation** drift
(the daily re-fetch), NOT fact drift: both my runs reported `total_drifted: 0` while carrying 13
fact_drift entries each.

So the obvious query — `SELECT (collected_data->'refresh_result') ? 'fact_drift'` — returns **f**
on a run that fired thirteen times, and `total_drifted` reads **0** beside it, which corroborates
the wrong answer. I read exactly that, and reported "the induction did not fire" to the owner
before dumping the full payload and finding all 13. **Check at
`results[N].fact_drift`, and never read `total_drifted` as the fact-drift count.**
Logged in `WRONG_CALLS.md`.

### The rotation's fleet impact, measured rather than predicted [16:25Z, ~5h after enabling]

I warned the owner to expect "real work arriving over the next day" on the strength of one site's
~77 items. Closing that with a measurement rather than leaving it as a forecast.

**Swept, one per hour as designed:** `remortgagecalculator.uk` 11:32, `robot-hands.com` 12:32,
`cookly.uk` 13:33, `loancalculator.co.uk` 14:33, `idea.uk` 15:34. **17 sites still due**, so the
backlog drains through tomorrow morning.

**Filed by discovery since 11:31Z, fleet-wide: 301 items across 8 sites**, of which **155 have
already moved past `detected`** — i.e. `detected-item-promoter` is doing exactly what I flagged
it would. Biggest: `robot-hands.com` 115 (81 promoted), `gamesdesign.co.uk` 91 (55),
`loancalculator.co.uk` 48 (6), `idea.uk` 36 (7).

⚠ `gamesdesign.co.uk` is **not** in the completeness rotation's swept list above — those 91 come
from another discovery agent on its own schedule. **Do not attribute the whole 301 to this
morning's decision**; the attributable share is the five sites named, and the honest figure for
"what enabling the rotation caused" is 208 items across those five, 95 promoted. The forecast was
the right order of magnitude; the attribution needed a second look.

#### Precision note on "re-verified on the NEW binary" [added 16:30Z]

Prompted by the MEMORY line added today — *a "fresh build" can ship no new code; a same-tag
rebuild serves the cached image*. Tightening my own wording before it is quoted back:

**What I actually established:** the pods are a new replicaset (`5bd56bdd9b`, started 14:42–14:43Z,
vs `5657f446c7` before) running image tag **`v1.0.1305`**, and `fact_drift_review` is present in
the binary they are running, with a positive control (`stale_attestation` 5) and a negative
control (impossible string 0) in the same exec.

**What I did NOT establish:** which commit built `v1.0.1305`. The `build provenance` startup line
had already scrolled out of `--tail=6000` on both replicas — CLAUDE.md warns it is a startup line
with a short shelf life on a busy service, and that an empty result there means "out of range",
not "unstamped". So I cannot name the commit, and **the same symbol counts (2 / 5) were true of
the PRE-roll binary too**, which means my check could not have distinguished a new image from a
cached one.

**Does it matter here? No, and this is why:** the test did not depend on the build being new. It
needed the fact-drift code to be PRESENT in whatever is running, which is exactly what the control
pair establishes — and then the mechanism demonstrably fired and tracked a live value change. A
positive result is indifferent to provenance. It would have mattered had the result been null:
then "the code is not in this build" and "the mechanism is broken" would have been
indistinguishable, and I had no evidence to separate them. **Worth knowing for next time: verify
provenance BEFORE the run, not after, because it is only cheap while the pod is young.**

## 2026-08-17 (evening) — CLM-022 is PROVEN END TO END: the `value_drift` arm fired with the right old→new values

Owner rolled another chassis and said carry on. Both halves of the proof are now closed.

### 0. The build actually shipped this time — checked at the digest, not at the pod

Today's MEMORY line (*a "fresh build" can ship no new code*) applied directly, and the check is
cheap, so I did it BEFORE the run rather than after (the lesson from this morning's precision note):

- tag moved **v1.0.1305 → v1.0.1307**, new replicaset `6d6d7b9996`;
- **digests match** — local `docker inspect aqls/agent-chassis:v1.0.1307` and the running
  `imageID` are both `sha256:8339bdbd7999…` on both replicas. No exec, no controls to get wrong;
- built from **`a6d1c53c0`** (image label `org.opencontainers.image.revision`, which survives when
  the `build provenance` log line has scrolled — it had, on both pods, at `--tail=4000`);
- `git merge-base --is-ancestor 989addb1c a6d1c53c0` → **YES**, so CLM-022's Go is aboard;
  16 commits in local HEAD are not in it.

**`org.opencontainers.image.revision` on the image is a better provenance source than the startup
log line** for anything older than a few minutes — same answer, no shelf life.

### 1. The real sweep — 13 baselines written, and the spec NOT touched

Fired `refresh_evidence_base` with `dry_run:false`, scoped by `site_id` (the site's own daily
mechanism, on demand). [MEASURED] `fact_drift` 13 entries, every `outcome: inserted` →
**13 `fact_drift_review` items now exist**, all `low` / priority **60** / `needs_human_review`,
13 distinct `item_key`s, no handler — exactly the burst the CONTRIB predicted.

⚠ **`writer_block_action` came back `unchanged`, not `regenerated`** — and the md5 of
`data->>'writer_block'` is **`73c1d35f…` before and after**, so nothing in the spec moved. The
earlier DRY run had reported `regenerated`, which I had flagged as a possible pending spec write.
**A dry run's `writer_block_action` is a plan, not a prediction of what a real run will do** — do
not read it as pending work.

### 2. The `value_drift` arm — induced, and it is right

Second induction (window ~15s, 18:18:0x→18:18:23Z, same trap-guarded restore):

```json
{"kind": "value_drift", "route": "fact_drift_review", "reason": "not_a_fork",
 "fact_id": "sdlt-ftb-relief-cap", "old_value": 500000, "new_value": 550000,
 "detail": "registered value moved 500,000 → 550,000; stamp-duty declares it encodes this fact",
 "page_name": "tool-stamp-duty", "subject_key": "stamp-duty", "outcome": "dry_run"}
```

**And the self-quieting is proven in the same three runs**, which is the part no single run shows:

| run | entries | kinds |
|---|---|---|
| dry, pre-baseline | **13** | `unreconciled_declaration` |
| REAL sweep | **13** | `unreconciled_declaration` (all `inserted`) |
| dry, post-baseline, one fact moved | **1** | **`value_drift`** |

13 → 13 → **1**. The baselines took, and only the fact that actually moved reports.
[MEASURED after] 13 items total, **0 added by the dry run**; 1 current register row, **0** test
rows; relief cap back at **500000**, `pinned` carried.

⚠ **`reason` is `not_a_fork`, where the other lane's RUNBOOK predicts `no_auto_fix`.** Both are
sufficient conditions for this route (their CONTRIB point 1 says so); the code records the one it
evaluated first. The routing is correct — only their expected string is off. Told them.

### 3. A pinned row id expired mid-session, exactly like a `git show HEAD:` baseline

The first re-run of the induction **failed**: `mutate.sql` pinned `2303a6f7…`, which stopped being
current the moment my own real sweep superseded it. `UPDATE 0`, then the INSERT hit
`idx_site_specs_current` → transaction aborted, nothing written (verified: 1 current row, no test
rows). Same shape as MEMORY's *a baseline that reads HEAD expires when you commit* — **and I
caused the expiry myself, two commands earlier.**

Both scripts now resolve the current row dynamically and the restore reads its target back out of
the test row's own `notes` (`"… restores to <uuid>"`), so it stays exact without a hardcoded id,
and it is idempotent. Kept in `scratchpad/{mutate,restore,induce,dryrun_safe,realrun_safe}.sh`.

### 4. The 13 items' question ANSWERED with evidence — green plus an induced red

Each of the 13 says *"stamp-duty declares it encodes this fact … Confirm the tool computes from
that figure."* This lane owns the tool that answers exactly that, so I ran it rather than
eyeballing:

```
python3 verify_criteria.py stamp-duty
  register: 13 SDLT facts loaded live from site_specs
  4 pinned value(s) recomputed: 4 agree, 0 MISMATCH
     of the agreements: 0 DEFINITION, 4 REGISTER, 0 CONVENTION
```

**And the red, because 4 agreements prove nothing on their own** (the script's own docstring says
so — a run that read the wrong column or fell back to hard-coded bands would print the same 4):

```
python3 verify_criteria.py --mutate sdlt-ftb-relief-cap=625000 stamp-duty
  MUTATED sdlt-ftb-relief-cap: 500000 -> 625000  (this run MUST report a mismatch)
  MISMATCH  stamp-duty  computes-asym  #sdltResult  pinned £19,750  REGISTER 14750.0000  delta +5000.0000
  4 pinned value(s) recomputed: 3 agree, 1 MISMATCH
```

Exactly one assertion moves, by exactly the perturbation. **So the tool demonstrably computes from
the register**, and the 13 items' question is answered in both directions. `#breakdown` remains
NOT re-derived and is reported as such, unchanged.

**I did NOT close the 13 items, and here is the reasoning rather than an omission.** Closing is
SAFE — `factDriftLastItemQuery` (`refresh_evidence_fact_drift.go:275-278`) selects on
`site_id` + `spec->>'check'='fact_drift'` with **no status filter**, so a resolved item still
serves as the baseline and the mechanism keeps working. What stopped me is that
`fact_drift_review` is handler-less with no documented resolution path, so I would be inventing
closure semantics for a new type; and `bugs_open/033` says that review queue has no working
surface, which makes closing them tidiness rather than function. **The evidence above is the
answer whether or not the rows are marked** — it is here, dated, with its red. Owner's call.

## 2026-08-18 — the rotation finished the whole backlog overnight, and the fleet picture it produced

[MEASURED 2026-08-18 ~12:35Z] `site-discovery-rotation-completeness`: **0 sites still due, all 23
swept**, last tick 10:42Z. It drained in ~23 hours, which is what the `LIMIT 1` + hourly interval
predicted before it was enabled — the one forecast this lane made that needed no correction.

**1,258 items filed, 639 already promoted past `detected`, across 23 sites.** By type:
`head_essentials_missing` 570, `page_rerender` 336, `page_component_status_drift` 62,
`unbuilt_internal_link` 41, `required_fields_missing` 30, `undeployed_asset` 26,
`needs_internal_links` 19, `orphan_blog_posts` 18, `canonical_mismatch` 18,
**`dead_internal_link_live` 16 across 11 sites**, `literal_markdown` 14, `content_rewrite` 14.

**The link-integrity total is 80 findings across the estate** — 41 unbuilt, 18 canonical
mismatches, 16 live-probed dead links on **11 different sites**, 5 phantoms. This site's eight
were not a local mess; they were a sample of a fleet-wide blind spot that had been unwatched since
2026-08-10.

⚠ **`head_essentials_missing` at 570 is 45% of everything filed and wants a human eye before
anyone acts on it.** A check that fires on nearly half the estate is either finding something
systemic or is mis-calibrated, and this lane has no basis to say which — flagged, not diagnosed.
It is the shape 016b §9 warns about ("a check that is always red is a check nobody reads"), so
whoever picks it up should ask that question first rather than draining the queue.

⚠ **The SUMMARY written this morning is already stale in its last section** — it says the checker
"will work through the estate over the next day", and it had already finished. Left unedited (the
series is the record and each entry is what we believed at that milestone); recorded here instead.
**A forecast written about a mechanism you switched on yesterday can expire before the ink dries.**

## 2026-08-18 (afternoon) — the directory-URL fix is SHIPPED (DGH-012), and the deploy taught two traps worth more than the fix

Owner approved it. Three lines in `scripts/cloudflare/worker.js`; live 12:23:10Z.

### Why it was safe, established BEFORE touching anything

1. **One-directional by measurement:** `b2 ls --recursive` over the whole bucket → **ZERO keys
   ending in `/`**. So the rewrite can only convert a 404 into a 200; no object that served before
   can stop serving. That is the difference between "I think this is safe" and knowing it.
2. **The repo copy was byte-identical to what was deployed** (downloaded the live script, stripped
   the multipart envelope, diffed against `git show HEAD:` — 5,499 bytes both sides). So nobody had
   dashboard-edited it, my deploy would change only my lines, and HEAD was an exact rollback
   artefact.
3. **A behavioural baseline captured with the same probe I would re-run after** — 11 URLs across
   three B2 sites, a git-route control, the health endpoint and a genuine miss.

### The two traps, either of which could have cost an outage or a false alarm

⚠ **`node --check` PASSES a syntactically broken `worker.js`.** I broke a copy deliberately (one
`)` removed) and `node --check` exited **0** on it. ESM syntax in a `.js` file makes the check a
no-op. As `.mjs` it correctly exited 1 with the SyntaxError at the right line. **My first
"SYNTAX OK" was worthless and I only know that because I ran the control.** A syntax error here is
a 36-zone outage.
> And a second-order slip inside that same check: my first control run reported `bad exit=0`
> because I wrote `docker … | head -3; echo $?` — **`$?` was `head`'s status, not docker's.** Same
> shape as the RUNBOOK's `${PIPESTATUS[0]}` warning for `b2`. Both files then "passed", which is
> what made me look harder rather than accept it.

⚠ **The PUT response returns `result.bindings: []` on a completely successful deploy.** That is
the exact signature of the credential-stripping outage `scripts/cloudflare/README.md` warns about,
and for about thirty seconds I believed I had caused it. **The sites were fine** — confirmed by
probing immediately, then by the `/settings` endpoint (both `plain_text` B2 bindings present,
`compatibility_date` and `observability` preserved). **Never read the PUT response as evidence
about bindings.**

Also: the README's hand-typed metadata is lossy — it omits `observability` and
`compatibility_flags`, which are live on this worker. The deploy script builds the metadata from
`~/.cloudflare/portfolio-sites-router.settings.json` and **refuses to PUT unless both B2 bindings
are present and non-empty**.

### The result, measured both sides with one probe

| URL | before | after |
|---|---|---|
| `mortgagecalculator.co.uk/guides/` | 404 | **200** |
| `mortgagecalculator.co.uk/tools/repayment/` | 404 | **200** |
| `gaswholesalers.com/tools/supplier-comparison-calculator/` | 404 | **200** |
| `leopardessconsulting.co.uk/tools/automation-savings-estimator/` | 404 | **200** |
| every `…/index.html`, the roots, `/worker-health` | 200 | 200 |
| a genuine miss | 404 + site's own 404 page | 404 + site's own 404 page |
| `relojistas.com/noticias/` (git route, control) | 200 | 200 |

`/guides/` and `/guides/index.html` return the same `<title>`, so it serves the real page and not
a soft-404; the 404 body leaks no bucket internals (`bugs_open/132`'s guarantee intact); and the
re-exported live script is byte-identical to the repo copy.

**On this site it closed the last non-260 broken link without touching a word of copy** —
`/tools/rate-forecaster/` resolves, and a re-audit of all 30 internal links leaves exactly one
dead target (`/scorecard-simulator.html`, which is 260's).

⚠ **Fleet-wide it resolved 1 of the 16 open `dead_internal_link_live` findings, not more.** A
second (`gaswholesalers.com/fuel-pricing-framework.html`) went 200 the same day because somebody
built the page — **not my change, and not credited to it.** The internal-link population was
always tiny (2 hrefs fleet-wide); the value is in the URLs nobody can enumerate — typed, shared
and inbound links — and that share is stated as unmeasurable rather than estimated.

### Governance, stated rather than skipped

Registered as **DGH-012 in the same commit that shipped it** (the ordering-exemption's condition
2), with its landmines and one open review question: canonicals emit `…/index.html` and now both
forms serve, so there is a duplicate-content question nobody has ruled on. **NOT sent to the
council gate — it refuses paths outside `platform/`/`internal/`/`pkg/` client-side**, so
`scripts/cloudflare/` cannot be submitted; that is a property of the gate, not an exemption I took.
Both LANDMINES entries describing the old behaviour were **corrected in place, not deleted** — the
local-server lesson and the still-404ing slashless form outlive the fix. The `pattern-check` hook
then caught that DGH-012 had no row in `000_concept_index.md`; added in a follow-up commit.

---

## 2026-08-21 — state check against the 08-18 handoff: the blocker CLEARED itself, and the site is down to ONE dead target

Picked this lane up cold. Three days had passed and **1,391 commits** had landed on the tree, none
of them this lane's. So before doing anything I re-measured every figure the 08-18 handoff carried.
Four of them had moved, and one of the four is the thing the whole lane was waiting on.

### What moved — measured 2026-08-21 ~10:20–10:30Z

| the handoff said | today | how measured |
|---|---|---|
| `bugs_open/260` blocks the third page, **actively owned elsewhere** | **`bugs_closed/260` — CLOSED 08-20, fixed AND live** | `ls bugs_closed/260*`; `git log` on the file path |
| 27 pages | **32 pages** (29 deployed, 1 `needs_rebuild`, 2 `planned`) | `pages` count by `build_status` |
| dead internal links: **one** (`/scorecard-simulator.html`) | **still exactly one**, and now the ONLY non-200 target of 33 | full live audit, below |
| the 4→6 self-fuelling link count | **6, unchanged** — no new page has been built since 08-16 | live grep across all 29 deployed pages |
| 13 `fact_drift_review` items open | **13, untouched since 08-17 18:28Z** — still the owner's call | status count |

### The live link audit — 29 pages fetched, 33 distinct internal targets, ONE dead

Scripted (`scratchpad/linkaudit.sh`): fetch every `build_status='deployed'` page, extract every
non-absolute `href`, dedupe to distinct targets, resolve each **once** cache-busted.

- 1,030 raw internal hrefs → **33 distinct targets**
- **32 return 200. One returns 404: `/scorecard-simulator.html`** — probed three times,
  cache-busted each time, 404 every time (the handoff's own trap: a single 404 in a fast scan is
  not evidence).
- Six live pages each carry exactly one anchor to it: `/disclaimer.html`,
  `/guides/first-time-buyer/`, `/guides/how-banks-decide/`, `/guides/lender-restrictions/`,
  `/guides/mortgage-scorecard/`, `/tools/affordability/`.

So the site's entire remaining product defect is **one page that cannot build, advertised from six
pages that can.**

### The platform DID detect all six — and that is a correction to how `bugs_open/328` frames it

328 (filed 08-19 by the `loanzy_uk_example_site` lane, **OPEN and UNOWNED**) says: *"Nothing tells
the pages that link to it."* On this site that is **not quite right, and the difference matters for
the fix.** There are **seven `unbuilt_internal_link` items** on the blocked page, each naming the
linking component by page and each quoting the same href:

```
unbuilt_internal_link in page_component (<page>:<component>):
  href "/scorecard-simulator.html" points at a page that has never been deployed
```

Detection exists, it is per-linking-page, and it is accurate. **What is missing is a route:** all
seven sit at `needs_human_review` with no handler, so the information is produced and then parked.
That is a different fix from "build the information" — and it is cheaper.

**Six of the seven match the live site exactly. The seventh is stale.** `8a230338` names
`contact-index`, and the link is in neither the served page nor the stored content:

| check | result |
|---|---|
| served `/contact/index.html` grep `scorecard-simulator` | **0** |
| `page_components.content_data LIKE '%scorecard-simulator.html%'` fleet-query on this site | **6 pages — `contact-index` NOT among them** |

So the condition was repaired (contact-index's stored copy lost the link, and the page has sat at
`needs_rebuild` since) and **the item outlived it**. Nothing re-checks these, so a parked
`unbuilt_internal_link` is not evidence the link is still there — [UNVERIFIED] whether that is
general or particular to this row; I only measured this site.

⚠ **Watch the stored-vs-served split here.** `contact-index` is `build_status='needs_rebuild'`
with `deployed_at` NULL, and it still serves a healthy 1,267-word page from its previous build.
"Not deployed" in the `pages` row does not mean "not serving".

### 260's fix is aboard the running chassis — proven at the binary, four controls

The handoff's standing instruction was *do not re-fire item `0c65f9fa`, it is 260's live test
case*. That reservation expired when 260 closed. Before acting on it I checked the fix is actually
running, using 260's own literals (the startup provenance line was unreadable — it is buried
inside a single 1.8 MB JSON log line on this service):

```
PRESENT  (want PRESENT) refusing to emit output that was not executed     <- added by the fix
ABSENT   (want ABSENT ) Go template execution failed, using regex fallback <- deleted by the fix
PRESENT  (want PRESENT) orchestration                                     <- must-be-present control
ABSENT   (want ABSENT ) zzz-not-a-real-literal-qqq                        <- must-be-absent control
```

`agent-chassis` **v1.0.1321**, both replicas on that tag, pods 14 h old. All four behaved, so the
grep discriminates and the reading is real rather than a probe that matches everything.

### Re-fired the blocked build at 10:28:49Z

Pre-state pinned first (`scratchpad/prestate_0c65f9fa.txt`): item `needs_human_review`,
`attempt_count 0`, error `step validate_content failed: … 20 blockers, 0 errors`, page
`build_status='planned'`, **0 `page_components`** — that last one is 260's signature, nothing
broken ever reached stored content.

Dispatch conditions checked rather than assumed:

- site **unlocked**, `locked_by` NULL;
- **nothing else armed on this site** (0 rows at `triaged`/`approved`), so arming one item drags
  nothing in and no backstop is needed;
- fleet queue nearly empty — 2 armed items fleet-wide, oldest 08-21 00:22Z, so this item (created
  08-16) is the globally oldest and dispatches on the next 120 s tick;
- chassis pods 14 h old, so well clear of the ~300 s post-restart silent-drop window;
- `/scorecard-simulator.html` is file-form and currently 404, so a build **cannot overwrite live
  content** (§10d's check).

Armed `0c65f9fa` alone — the exact item that failed under 260 — because re-running the recorded
failure on the fixed binary is the most informative single reading available: either the page
builds and six dead links die at once, or it fails **naming one field**, which is the behaviour
change 260 claims and is actionable content repair either way.

### Other things that had happened on this site, and what they turned out to be

- **Two `instance_scope_conversion` items completed 08-19 21:15** (RFC_034's lane), summarised
  `tool-bridging-compound`/`tool-rate-scenarios` **[SERVING BROKEN on 1 page(s)]** — and the
  fix result is `{"fixed": false, "unprefixed_before": 0, "reason": "no unprefixed bindings found
  and nothing changed — already sound"}`. Both pages serve 200 with **zero** `{{` template leaks
  and 1,633 / 1,795 words. So "SERVING BROKEN" in the summary and "already sound" in the result
  disagree, and the artefact agrees with the result. Not this lane's to chase — noted for RFC_034.
- **Nine pages had `updated_at` bumped to 08-19 20:42:1x** by that same sweep, without content
  changing.
- **A completed item's `result` holds a plan for a page that does not exist.** `28c506f1`
  (`item_key gap_plan_new_disclaimer_…`, `completed_at` **2026-08-11** 18:18Z) has
  `updated_at` **08-19 11:54Z** and a `result` of `{"approach":"new_page","new_page":{"name":
  "guide-mortgage-affordability", …}}` — a full gap plan for a *different* page. There is no
  `guide-mortgage-affordability` page row and no follow-on work item naming it, so **that plan
  went nowhere.** `apply_gap_plan_action.go` is documented as *"new_page: creates page record +
  needs_content_page work item"*, so the shape suggests a planner wrote back to a stale
  `work_item_id` and the apply half never ran. **[UNVERIFIED as a mechanism]** — I could not find
  the run: `orchestration_states` has no row referencing the item id (that table prunes). Recorded
  here as an observation, not filed as a bug, because I cannot yet name the writer.
- Two long-parked items were re-touched 08-20 16:01Z without changing status (`07bc64cd`
  needs_section_data — contact-info wants a real business email; `e781118c` required_fields_missing
  — `tool-simple` hero has no `headline`). Both are owner-input items.

### 2026-08-21 (later) — the re-fire's result, and the second defect it exposed

**Attempt 1 failed, and the failure is exactly the behaviour change `bugs_closed/260` claimed.**
Where the pre-state error was `step validate_content failed: … 20 blockers, 0 errors`, the new one
names the field:

```
step process_sections_loop_iter_1_render_section failed:
  failed to execute action render_component: component "mechanism-flow":
  content does not match the declared field type(s) —
  steps[2].branches: declared array (items: object), got string;
  steps[3].branches: declared array (items: object), got string;
  refusing to render (bugs_open/260)
```

Same item, same route, pinned pre-state — so this is a clean A/B, not two observations.

**The component schema is not at fault.** `content_components.input_schema` for `mechanism-flow`
declares `steps[].branches` as `{type: array, items: {type: object, required: [body],
properties: {body, label}}}`, with the description *"a decision point: two or more outcomes,
rendered side by side"*. Well-formed and unambiguous. The writer produced **prose** for two of the
steps' `branches`. So this is the **writer** half of 260 — which 260's own closure assigns to
`copy_quality_two_stage` by the owner's 2026-08-12 split — and not a schema defect to fix here.

Attempt 2 armed 10:38:43Z to test whether the mistyping is **reliable or stochastic** on this
component. Two of roughly four steps carried it, so the writer gets the shape right some of the
time; if attempt 2 mistypes the same two indices, that is a much stronger finding than one run.

### The second defect: the failed build reported `complete`

Attempt 1 ended `status='complete'`, `completed_at` stamped, **0 `page_components`**, page still
`build_status='planned'`, URL still 404. The orchestration chain shows why: the render step's
orchestration **FAILED**, its parent reached `complete_error`, and the outer sagas reported
`complete`/COMPLETED.

**This is a known SHAPE with a NEW cause, and that is the point.** `bugs_closed/028` closed
"page-build no-op reports complete" on 2026-07-25 by adding a guard step; a sibling fix added a
second. Both are in the live `page-build-handler` workflow today and **their own descriptions state
the intent**:

| guard step | its description |
|---|---|
| `mark_no_ready_sections` | *"park the work item visibly instead of letting the dispatch loop stamp it complete"* |
| `mark_writer_skipped` | *"park the work item visibly instead of letting the dispatch loop stamp it complete"* |

There is **no equivalent for a render refusal**, and the routing table shows no path for one —
read from the live `agent_definitions`, both agents:

- `page-content-writer`: `process_sections_loop -> (none)` (the loop that owns
  `..._iter_N_render_section`)
- `page-build-handler`: `spawn_content_writer -> (none)`; `validate_content -> mark_needs_review`

So the **pre-260 path was routed and visible** (`validate_content` → `mark_needs_review` →
`needs_human_review`, which is why the pre-state was parked in a queue) and the **post-260 path is
unrouted** (render refusal → child FAILED → no `error_step` anywhere → success-labelled complete).
`CompleteWorkItemAction`'s guard in `load_work_item_actions.go` only preserves a status a handler
**deliberately set** — its `WHERE status NOT IN (…)` cannot help, because nothing flagged the item.

**Net effect: 260's fix moved the failure earlier, from a routed step to an unrouted one, and the
same defect that used to park visibly now terminates as `complete` at `attempt_count 1` of 3.**
The excellent named diagnosis lands in a terminal item's `error` column where nothing looks. The
per-cause guard pattern does not cover a new cause — which is `bugs_open/328`'s own argument
(*"the route needs one rule, not one rule per cause"*) arriving on a different route.

**Filed through the diagnosis loop rather than asserted**, because it is cross-cutting and the
cause sits outside the symptom: intake correlation `47a4d1d5-5fa3-4940-8d95-f431d5896cb2`, **run
correlation `0b498cf8-73ac-4d34-9a14-89a84f4e7b7a`** — use the run correlation for the artifacts.
The diagnosis queue was empty beforehand (no duplicate) and `grep` of both bug dirs found the shape
(028) but not this cause. **Verdict not read yet — do not repeat the mechanism above as settled
until it is.** If CONFIRMED it belongs to the 260 lane and/or a new bug file, not here.

### My own wrong call this session, recorded where it was made

The attempt-2 watcher **reported a terminal state within 30 s of arming, and the build had not even
been claimed.** I had deliberately kept attempt 1's error text in the column so attempt 2 could be
compared against it — and that retained text contains the word *"failed"*, which was one of my
watcher's terminal patterns. So the watcher matched **history, not the event**.

The check: **a watcher must key on something that cannot hold a previous run's value** — a
timestamp compared against a pinned instant (`completed_at > <armed-at>`), an incremented counter,
or a hash of the field rather than a substring of it. The rewritten watcher keys on
`completed_at > '2026-08-21 10:38:43'` and `attempts=2`. Deliberately preserving evidence and
keying a detector on that same field are individually correct and jointly a false positive.

### 2026-08-21 (11:10Z) — attempt 2's result decides the writer question, and the 090 never reached the question

**Attempt 2 ran and was refused for the same reason at DIFFERENT step indices.** That difference is
the whole finding:

| | component | field | occurrences | indices |
|---|---|---|---|---|
| attempt 1 | `mechanism-flow` | `steps[].branches` | 2 | **2, 3** |
| attempt 2 | `mechanism-flow` | `steps[].branches` | 2 | **1, 2** |

Confirmed a genuinely fresh run rather than a retained error, by three independent markers:
`attempt_count` 1 → **2**, `completed_at` **11:06:10Z** (after the 10:38:43Z arming), and
`md5(error)` **`24859342` → `62b415f3`**. I checked those *because* my own watcher had just been
fooled by retained text — the check that catches it is the check the earlier mistake taught.

**So the writer mistypes this field RELIABLY, not stochastically.** Both runs, same component, same
field, same count; only the positions move. **A third retry has no reason to succeed and I have not
fired one.** This is `bugs_closed/260`'s **writer half**, which 260's closure assigns to
`copy_quality_two_stage` by the owner's 2026-08-12 split — it is not this lane's to fix and not a
schema defect (the schema is well-formed, checked above).

**The page therefore stays 404 and the six dead links stay live.** That is the honest state: the
blocker moved from "the renderer corrupts it silently" to "the writer produces the wrong shape,
reliably, and we now know exactly which field" — real progress, and not a fix.

### The 090 returned no verdict, for infrastructure reasons

```
verdict      FAILED  AI endpoint unavailable: provider=anthropic model=claude-sonnet-5 … status 400
mark_failed  FAILED  failed to apply work item failure ladder:
                     ERROR: could not determine data type of parameter $4 (SQLSTATE 42P18)
```

Two separate failures, neither about the claim. The first is the documented fleet-wide Anthropic
budget-window shape (RUNBOOK §, "it presents as a work item that completed"). **The second is a
platform defect in its own right: when a run fails, the machinery that RECORDS the failure also
fails** — an untyped `$4` in the failure ladder. Recorded in `bugs_open/348` §5; not chased.

**I did not re-fire.** Memory carries a prior lane's finding that a second 090 firing on a failing
loop buys nothing (*"it is the diagnosis LOOP, not this symptom; stopped at two firings"*), and the
first-hand evidence here is now stronger than one loop run: two A/B runs, both live routing tables,
and the guard code read at the deciding arm.

### Filed `bugs_open/348`

The routing gap is filed as its own bug — OPEN, UNOWNED — **with the 090 substitution stated in a
banner at the top**, per the owner's 2026-07-31 ruling. Fleet census run rather than deferred:
**exactly one occurrence fleet-wide, which is this item**, and the file says plainly that twenty
hours of exposure cannot distinguish "rare" from "not yet fired", so the number must be re-run
before anyone sizes it. The comparator that does carry volume (124 items parked visibly by the
`validate_content` path since 08-01) is recorded with an explicit warning **not** to quote it as
the blast radius — only the mistyped-field subset shifts, and that subset is [INFERRED].

### 2026-08-21 (12:00Z) — ⚠ CORRECTION: my "second defect" mechanism is REFUTED, and `bugs_open/344` owns it

The `bugfix_307_terminal_write_contract` lane read my row within the hour
(`CONTRIB_2026-08-21_from_the_307_lane_your_item_was_flagged_and_then_overwritten.md`) and I
re-measured. **They are right.**

> ~~"the post-260 path is unrouted (render refusal → child FAILED → no `error_step` anywhere →
> success-labelled complete)"~~ and ~~"`CompleteWorkItemAction`'s guard cannot help, because
> nothing flagged the item"~~ — **both false.**

**A failure write DID reach the item.** Re-measured on my own row:

| column | reading |
|---|---|
| `attempt_count` | 0 → **1** → **2** — a ladder consumed an attempt each run |
| `retry_after` | stamped **+30 m** each run (`2026-08-21 12:06:07` after attempt 2) |
| `page-build-handler.call_content_writer.error_step` | **`mark_item_failed`** — routed |

The child's refusal is unrouted *inside* `page-content-writer`, but the child FAILING makes the
parent's **`call_content_writer`** step error, and that step **is** routed to `mark_item_failed`,
which ran the failure ladder (WII-024). **The dispatch loop's `mark_complete` then overwrote the
re-triaged row ~2 s later**, because `triaged` is not in the completion guard's excluded list.
**The flag was written and trampled, not absent.** Fingerprint: **`retry_after > completed_at` on
a `complete` row.** Filed and owned as **`bugs_open/344`**; my row is its named natural-damage case.

**What survives:** the observation (two builds composed nothing and reported `complete`, 0
`page_components`), the A/B on the writer mistyping (untouched), and the argument that a per-cause
guard pattern does not survive a new cause. `bugs_open/348` is corrected in place with a banner
rather than rewritten, and its §3 routing table is left unedited as the evidence.

**How I got it wrong — and it is worse than a missed check.** I dumped the whole routing table and
**`call_content_writer -> mark_item_failed` was four rows above the `spawn_content_writer -> (none)`
I quoted.** The refuting line was in my own output. I selected the row that fitted the theory and
never asked *which parent step actually fired* — the orchestration named the CHILD step and I
treated that as the whole answer. And `attempt_count` was in the `SELECT` I ran repeatedly, going
0→1→2 in front of me, while I wrote the sentence it refutes. **An absence claim ("nothing wrote to
X") is a query, not a reading.** `WRONG_CALLS.md` 2026-08-21 (second entry).

**And the 090's failure was not neutral.** I substituted first-hand verification when the loop
broke — permitted, and I declared it — but the loop is exactly what would have caught this, and I
recorded my substitution as though the evidence were equivalent. It was not. A REFUTED verdict here
would have cost one run.

### Also corrected: the stamp-duty config question is NOT outstanding

I told the owner twice today that *"the other lane's stamp-duty config question has been waiting
since 08-16"*. **It was answered on 08-17** — the owner said *"seed it for real"*, option 1, all 13
ids, `doc_plans` `400657e0…` installed 12:06:56Z. **The 13 `fact_drift_review` items ARE that
answer's consequence.** I carried the line forward from the README's 08-17 entry, which was written
before the answer landed — in the very session where I re-measured everything else and warned that
a carried-forward status is a claim about the past. The 08-18 handoff §0, which I read at the start
of this session, states the owner's ruling plainly. **Three decisions are open, not four.**

## 2026-08-21 (afternoon) — owner asked for the contact page to be REWORDED; state re-checked first, then driven through the framework

### State check before acting (33 commits since my last, 7dd00dc4c..HEAD)

| checked | finding |
|---|---|
| `contact-index` | still `build_status='needs_rebuild'`, `deployed_at` NULL, serving fine from its previous build |
| my scorecard item `0c65f9fa` | untouched in substance — still parked by me, my `handled_by` intact; `updated_at` bumped 13:46 by a sweep (the known unreapable-item landmine) |
| site lock | unlocked; nothing else armed |
| **NEW item `ee6f837e`** (12:59) | `missing_conversion_path` — *"lead_generation site … has no working conversion path: page 'index' is nav-reachable but carries no form"*. **A second detector reaching the same fact from a different direction.** Not actioned: it asks for a FORM on the homepage, which is a revenue-model question, not a wording one |
| **`238` lane shipped "every dead contact control on the fleet is gone"** (`cf24ea645`) | **Adjacent, NOT competing — checked before assuming.** Their fix removes empty `mailto:`/`tel:` *controls*; the fleet census went `href="tel:"` 5→0, `mailto:` 1→0. **This site was never in it.** Our page has zero such controls — it is their `gamesdesign` case ("renders nothing at all… which is the case that proves the gate"). Their layer is the markup; ours is the **prose**, which still invites contact. Different defect, same page |

### Why the copy is the defect — measured, not inferred

The page has exactly two components and **both invite contact**, four times between them:

- `hero-contact.headline` — *"There's a place here for questions the tools and guides don't answer"*
- `hero-contact.subheadline` — *"…tell us here."*
- `generic-text-block.content` — *"if you write to us…"*, *"questions people send in"*, *"that's worth
  telling us. Say what you entered and what you saw, and we'll take a look."*

Against, measured on the served page: **0 mailto links, 0 `@` in the text, 0 `<form>`, 0 phone.**

### Followed the framework rule rather than writing the words

Owner ruling 2026-08-06 — *"we want it all to be done through the framework, so we don't want you
writing things yourself"*. So this is a `content_rewrite` carrying a **brief**, not copy.

**The check that memory says is the wasted-run trap — `llm_fields` FIRST — was run before filing:**

| component | LLM-authored | NOT llm |
|---|---|---|
| `hero-contact` | `headline`, `subheadline` | — |
| `generic-text-block` | `content`, `heading` | — |
| `contact-info` | `intro_text`, `section_title` | **`address`, `email`, `hours`, `phone`** |

All four fields I need are writer-owned, so the ask is well-formed. **And the last row independently
confirms what I told the owner about item `07bc64cd`:** `email` is sourced from
`site_specs.identity`, **not** from the LLM — so no prompt can invent one and that item genuinely
requires a human. That was an inference this morning; it is a measurement now.

**Spec keys verified against the readers, not guessed** (`bugs_open/271` is the bug for getting this
wrong): `suggestion` is *"the key page-build-handler reads"* — the comment appears at four separate
call sites — and `mode: "edit_live"` opts the item into `load_current_section_content`, which
attaches each ready section's current `rendered_html` **so the writer EDITS instead of regenerating**
(`bugs_open/178`'s protection). That is exactly a reword rather than a rebuild.

**Before-state pinned per slot** (the canary the memory prescribes, after a prior run cost a page):

| position | function | len | md5 |
|---|---|---|---|
| 1 | `hero-contact` | 286 | `71fe39a8c3456aa01faefa0c0c8a2cea` |
| 2 | `generic-text-block` | 960 | `c5ed08d91f3e80769cbce9fa6f4f2fbe` |

Item **`e31ba039`** filed and armed 15:5xZ, `priority 30`, `page_id` set on the row (the mode needs
it). The brief states the measured evidence, changes the STANCE rather than the subject, keeps the
honest "figures not advice" framing and the two working links, and forbids inventing any contact
route — including a form, which is the one `ee6f837e` would have wanted.

**Watching the component md5s, not the item status** — this morning proved an item can report
`complete` having composed nothing, so the artefact is the signal.

### 2026-08-21 (16:0xZ) — the reword is LIVE, and getting it dispatched corrected a LANDMINE

**Outcome: both components reworded, deployed, and verified at the served URL.**

| slot | before md5 | after md5 |
|---|---|---|
| `hero-contact` | `71fe39a8…` | **`7d08a7e8…`** |
| `generic-text-block` | `c5ed08d9…` | **`5f12a3f8…`** |

**Verified at the artefact, not the item status** (this morning's lesson): the served page carries
the new copy, and a promise-phrase sweep on the served HTML returns **0 for all seven**:
*"tell us here"*, *"write to us"*, *"send in"*, *"we'll take a look"*, *"get back to you"*,
*"drop us a line"*, *"contact us today"*. `mailto` **0**, `<form>` **0**, and
**zero email-shaped strings in the raw HTML** — nothing was invented.

The writer also found a destination I had not thought to name: it points a reader whose figure
looks wrong at `/disclaimer.html` (*"the assumptions behind our figures"*), which exists, is
deployed and is exactly the right page. **That is the argument for briefing the framework rather
than writing the words** — I would have shipped the two links I already knew about.

⚠ **A false alarm of my own making, recorded because it nearly became a finding.** My served-page
watcher reported `at_signs=3` where the old page had 0, which reads as "an email address was
invented". It was not: the watcher stripped TAGS with `sed 's/<[^>]*>/ /g'`, which leaves the TEXT
CONTENT of `<style>` blocks — the three `@` were `@media` rules. **Stripping tags is not stripping
CSS.** The disconfirming check took one command: a regex for an email-shaped string over the raw
HTML returned `[]`.

### The dispatch: RUNBOOK §15 and a LANDMINE disagreed, and the landmine was the stale one

The item would have starved — my site was **last of nine**, behind 62 dispatchable items on
`webdesign.co.uk` (oldest 11:45Z) and 15 on `loanzy.uk`, because site selection is globally
oldest-first and my item was minutes old. That is §15's scenario, and its three preconditions all
held (item dispatchable; site unlocked with 0 claimed and exactly 1 armed — mine, so blast radius
was one item; trigger alive, last fired 47 s earlier).

**But `LANDMINES.md` said the §15 technique does not work** — that orchestrating
`build-dispatch-loop` directly *"reports COMPLETED and processes NOTHING"*. Rather than pick a side
I ran it and applied the landmine's own check, which is the right check. **It worked:** item
`triaged → claimed`, `claim_result: true`, `handler_spawned: true`, full chain through
`deploy_page`. **Sequence across the estate: 08-08 no-op · 08-11 works (§15's own record) · 08-21
works.** The landmine was written 08-08 and the runbook's 08-11 success was never folded back into
it, so a session reading LANDMINES alone would avoid a technique the runbook documents as working.
**Corrected in place** (25 insertions, 0 deletions, single hunk, against a base re-verified unmoved
immediately before the edit — the shared-ledger trap), verifier armed via
`landmines-verify-dispatch.sh` (correlation `43275492…`). **Cause not claimed:** `bugs_closed/239`
is the obvious suspect but its trigger was `source`+`spec` co-occurring and neither payload had
that. **[UNVERIFIED.]**

**And `edit_live` was verified to have engaged, not merely requested:** the plan carried
`edit_live_meta {applied: true, matched: 2, fallback_matched: 0}`, so the writer received both
sections' current content and edited rather than regenerated (`bugs_open/178`'s protection).

### What is left on this page, and it is the owner's call, not a defect

The `<title>` is still **"Contact us"** and the footer `nav_label` still **"Contact"**. Those are
`pages` columns, not LLM content fields, so they were untouched by design — changing them is a
page-record edit, and the nav label appears in every page's footer. A page titled "Contact us" that
says there is no contact route is defensible (it is the page people look for, and it now answers
honestly) but it is a judgement I have deliberately **not** made unasked. **Flagged, not changed.**

Item `07bc64cd` (`needs_section_data`, wanting a business email) is **left open**: the owner asked
for a reword rather than supplying an address, but has not said "never", and closing it would
assert a decision they did not make.

## 2026-08-24 — owner asked for the not-financial-advice disclaimer on EVERY page; used STY-051 rather than touching 32 pages

### The mechanism already existed and is purpose-built for exactly this

`site-footer`'s `input_schema` declares two config-sourced fields, and the first is this request
almost verbatim:

| field | source | description (verbatim) |
|---|---|---|
| `compliance_lines` | `config.chrome.compliance_lines` | *"Every-page compliance/mission lines: an ARRAY of plain-text strings rendered as a gated block in the footer chrome on every page"* — STY-051 |
| `footer_note` | `config.chrome.footer_note` | per-site disclosure band — STY-052 |

⚠ **Neither is an LLM field**, so the framework's writer cannot author them — the value is config,
living in `site_specs` aspect `site_config` under `chrome.*`. That is *why* the wording went to the
owner rather than to a brief: there is no generator to ask, and a financial-advice disclaimer is a
legal statement that is the owner's to make. **Owner chose the site's own voice** (option 1 of 3),
reusing the phrasing the framework itself wrote for the contact page on 08-21, and **declined an
FCA/regulated line** — deliberately, because neither of us has verified the site's regulatory
status and a wrong claim about it is worse than no claim.

⚠ **`compliance_lines` must be an ARRAY of strings.** The schema says a non-array *"degrades the
whole template to the regex fallback renderer"*. The template is
`{{range .compliance_lines}}<p>{{.}}</p>{{end}}`, so each element becomes its own `<p>` — two
elements, two tidy centred lines.

### The write, and the index that has bitten this lane before

`site_specs` has `idx_site_specs_current UNIQUE (site_id, aspect) WHERE is_current`. This lane
previously aborted on exactly that index by naming a pinned row id that a later sweep had
superseded. So: **one statement, resolving the current row dynamically**, superseding and inserting
in a single CTE so two current rows never coexist —

```sql
WITH old AS (UPDATE site_specs SET is_current=false, superseded_at=now()
             WHERE site_id=… AND aspect='site_config' AND is_current
             RETURNING site_id, aspect, data)
INSERT INTO site_specs (…) SELECT site_id, aspect,
  data || jsonb_build_object('chrome',
    COALESCE(data->'chrome','{}'::jsonb) || jsonb_build_object('compliance_lines', …)) …
FROM old;
```

`data || …` and the `COALESCE(data->'chrome', …)` merge are deliberate: the pre-existing
`{"locale":{"lang":"en-GB"}}` survived, and any future `footer_note` would too. Verified in the
`RETURNING`: both `chrome` and `locale` present.

### Chrome re-render — asserted the STAMP, not the run status

Dispatched `rerender-chrome` (STY-055; `config.agent_type`, `input_data={site_id,domain}`),
correlation `3ae42823-09aa-4346-8477-8bc2f8a27577`. **STY-055's own landmine says a run at a
locked site completes GREEN having stamped nothing**, so locks were checked first (all three slots
`locked_at` NULL) and the assertion was the artefact:

| slot | md5 before | md5 after | length |
|---|---|---|---|
| footer | `32a3c879…` | **`d1a2600e…`** | 2,722 → 3,139 |

and the stored footer now literally contains
`<div class="footer-compliance"><p>This site works out figures rather than giving financial
advice.</p><p>Any decision about your own borrowing needs a lender or a broker who can look at your
full circumstances.</p></div>`.

### The canary earned its keep: the page got 2.4× BIGGER, and my change is ~400 bytes of it

Memory's rule — *never size a re-render by YOUR change; canary TWO pages* — and it was right.
Chrome is baked byte-for-byte into each deployed page (verified: the stored footer's opening
fragment is present in the served homepage), so all 29 deployed pages need re-rendering.

Canaried the **oldest** (`guides-index`, last deployed 08-16) via the assemble-only single-page
route (no `reason` ⇒ no LLM, authored copy untouched):

| | before | after |
|---|---|---|
| bytes | 25,173 | **59,961** |
| visible words | 315 | 351 |
| `<head>` | 9,110 | **43,960** |
| inline CSS | 16,454 (8 blocks) | 50,204 (7 blocks) |

**Visible-text diff is clean** — every delta is nav/footer link text (the nav rebuilds from
currently-deployed pages, and more are deployed now than on 08-16) plus the disclaimer itself. **No
body copy changed.** But +34,788 bytes for +36 words needed explaining before repeating it 28 times.

**The cause: a single 34,530-byte `tool-portal-light` LAYOUT stylesheet (240 rules) now inlined in
the head chrome.** The stored `head` slot went **8,628 → 43,102**.

**Is that damage I caused, or catch-up?** The disconfirming control was one query — head-chrome
size across the fleet, ordered by recency (**measured 2026-08-24**):

| rendered | site | head bytes |
|---|---|---|
| 08-24 (mine) | mortgagecalculator.co.uk | 43,102 |
| 08-23 | garden-tools.uk | 48,891 |
| 08-23 | dartsonline.com | 62,038 |
| 08-22 | robot-hands.com | 59,800 |
| 08-22 | cookly.uk | 43,153 |
| 08-21 | fundamentallyai.com | 43,521 |
| **08-20** | **webdesign.uk** | **8,580** |
| **08-20** | **vonc.com** | **9,335** |

**The small heads are exactly the sites last rendered on 08-20 — which is what this site was.** So
the inflation is this site adopting the current design system, not something the disclaimer
introduced. ⚠ **One counter-example, stated rather than smoothed over:** `loancalculator.co.uk` is
8,561 bytes and was rendered 08-23. So "recency alone predicts head size" is **not** exactly true —
[UNVERIFIED] why; probably a different layout/theme. The pattern is strong, not universal.

**Proceeded** because stopping after one page is the worst state available: one page on the new
design system and 28 on the old is a visibly inconsistent site.

### A hazard found on the way, which matters to the NEXT person more than to me

**Three deployed pages have a section with NULL `content_data`** — `tool-bridging-compound`,
`tool-rate-scenarios`, `tool-simple` (one section each, **as of 2026-08-24**). Per
`049b_deploy_single_page.sh`'s own header, if you pass a `reason` (e.g. `section_data_resolved`)
and **any** section has NULL `content_data`, *"the whole page escalates to the content writer and
the copy IS regenerated"*. **Assemble-only (no reason) does not escalate**, which is what was used
here — but anyone re-rendering these three pages *with* a reason would silently have their copy
rewritten. Two of the three are the pages the RFC_034 lane's `instance_scope_conversion` items
called *"SERVING BROKEN"* on 08-19 and then closed having changed nothing, and the third is
`e781118c`'s missing-headline page. **That cluster is worth someone's attention; it is not this
lane's to chase today.**

### Also worth recording: assemble-only DOES pick up new chrome

The script's header warns that assemble-only *"stitches the STORED rendered_html and therefore
cannot pick up a component template change"* — true of **component templates**, and it reads as
though it covers chrome too. **It does not: chrome is injected at assembly, so the new footer
landed via the no-`reason` route with zero LLM calls.** Tested rather than assumed, on one page,
before committing to 28.

### Near-miss: I built the page list from `build_status='deployed'` and would have missed the CONTACT page

**My own landmine, recorded in this file three days ago, caught me.** The 08-21 entry says:

> ⚠ **Watch the stored-vs-served split here.** `contact-index` is `build_status='needs_rebuild'`
> with `deployed_at` NULL, and it still serves a healthy 1,267-word page from its previous build.
> "Not deployed" in the `pages` row does not mean "not serving".

The propagation list was `WHERE build_status='deployed'` → **29 pages**. `contact-index` is
`needs_rebuild` with `deployed_at` NULL, so it was excluded — while serving **HTTP 200**. It would
have been the single page on the site without the disclaimer, and it is the page most obviously
about what the site will and will not do for you.

Caught before completion, not after: re-probed the row, confirmed 200 with 2 components and **0**
NULL `content_data` sections (so assemble-only is safe on it), dispatched it, and added it to the
verification set so the completion check covers **30 serving pages**, not the 29 the `pages` table
calls deployed.

**The check, generalised:** on this site the set of pages that SERVE is not the set the `pages`
table calls `deployed` — it is a superset. Any fan-out that must reach "every live page" should be
built from a **probe** or from `deployed ∪ needs_rebuild`, never from `build_status='deployed'`
alone. The two `planned` pages (`scorecard-simulator`, `guide-market-structure`) are genuinely not
live and need nothing — they will pick the footer up when they are first built, because the chrome
is now in the store.

### Result: 30/30 serving pages carry the disclaimer — and the two canaries disagreed, informatively

**Coverage confirmed twice, independently** (2026-08-24): a rigorous pass (3 attempts per page,
25 s timeout) and a separate hand check (2 attempts) both returned **30/30**. 28 pages dispatched
assemble-only, **0 dispatch failures, 0 FAILED orchestrations** (23+ COMPLETED on this site in the
window).

**The two-canary rule paid off exactly as memory says it does — they disagreed:**

| | `guides-index` (STALE, last deployed 08-16) | `guide-lender-restrictions` (FRESH, 08-23) |
|---|---|---|
| bytes | 25,173 → 59,961 | 22,064 → 56,955 |
| `<head>` | 9,110 → 43,960 | 9,587 → 44,061 |
| visible words | 315 → 351 | 451 → **480** |
| visible-text changes | nav/footer links **+** the disclaimer | **the disclaimer, and nothing else** |

**That difference is the finding.** The fresh page's ONLY visible change is the two disclaimer
lines (+29 words). The stale page additionally picked up eight days of nav changes — because its
nav was eight days old, not because of anything I did. **Had I canaried only the stale one I would
have reported "re-rendering also churns the nav" as a property of this change; had I canaried only
the fresh one I would have reported "nothing changes but the disclaimer" and been surprised later.**
Both are true, of different pages.

And both took the **same** ~34.5 KB head, which settles what that is: uniform chrome, independent
of page age — consistent with the fleet control rather than with anything page-specific.

### Verification discipline: a single probe under-reported, and I nearly explained it wrongly

`investor-index` read as MISSING the disclaimer on one probe, minutes after I had watched it render
correctly. **I said out loud that the cause was my `--max-time 8` truncating a now-71 KB response —
and then could not reproduce it**: the same page returns complete at `--max-time 2`. So the
timeout explanation is **withdrawn**; it was a single transient failed fetch, cause not established.

The remedy does not depend on the cause and this lane already had it written down: *"a single 404
in a fast scan is not evidence"* (08-18 handoff §4). Six consecutive probes showed the page fine.
**Measured proof that single-probe checks under-report here:** at the same instant, the
single-probe watcher reported **28/30** while the 3-attempt pass reported **30/30**. The weaker
check is conservative (it can produce false ABSENCES, never false presences), so it is safe to run —
but its number must never be the one quoted.

That is the third checker-of-mine to give a confident wrong reading in this lane in four days
(08-21: a terminal-state watcher matching retained error text; 08-21: an `@`-count matching CSS
`@media`; today: this). All three were caught within a minute, all three by re-reading the whole
line rather than the verdict. **The pattern worth carrying: my watchers fail toward FALSE ALARM,
not false calm — which is the safe direction, and is worth preserving deliberately when writing
the next one.**

### Final state

- **30 serving pages, 30 carrying the disclaimer.** The `pages` table calls 29 of them `deployed`;
  `contact-index` serves while flagged `needs_rebuild`.
- Two pages are genuinely not live (`scorecard-simulator`, `guide-market-structure`, both
  `planned`). They need nothing — the chrome is in the store, so they inherit the disclaimer the
  first time they are built.
- Site unlocked, nothing armed by this lane, no work items left open by this task.

### ⚠ CORRECTION 2026-08-24 — the "three pages" hazard is ONE page, and I took it from a doc comment instead of the code

I wrote earlier today, and repeated it in the handoff and to the owner:

> ~~"**Three deployed pages have a section with NULL `content_data`** … anyone re-rendering these
> three pages *with* a reason would silently have their copy rewritten."~~

**Wrong for two of the three.** The source of the error is exactly the thing my own memory warns
about: I took the rule from `049b_deploy_single_page.sh`'s **header comment** —

> *"if ANY section has NULL content_data the whole page escalates to the content writer and the
> copy IS regenerated"*

— and never opened the action it describes. The comment is a fair summary of the **general** rule
and omits the **exemption**, which is in the code and is deliberate
(`rerender_page_sections_action.go:~400`):

```go
// A self-contained TOOL section legitimately has no content_data: a tool
// is complete HTML with no LLM-authored fields, so content_data={} is
// its correct shape, not the missing-content defect this pre-check
// exists to catch.
if comp, _, ok := resolveComponent(s); ok && isSelfContainedSection(comp) { continue }
```

`isSelfContainedSection` (`:1361`) is two conditions: **empty `input_schema` AND
`component_level == 'tool'`**. Applied to the three (measured 2026-08-24):

| page | component | `component_level` | `input_schema` | verdict |
|---|---|---|---|---|
| `tool-rate-scenarios` | `tool-rate-scenarios` | `tool` | NULL | **EXEMPT — correct shape, no escalation** |
| `tool-bridging-compound` | `tool-bridging-compound` | `tool` | NULL | **EXEMPT — correct shape, no escalation** |
| `tool-simple` | `hero` | `section` | 1,300 chars | **WOULD ESCALATE — copy regenerated** |

**So the hazard is ONE page — `tool-simple` — and it is a HERO, not a tool.** For the two tool
pages, NULL `content_data` is not a defect at all; it is what a self-contained tool is supposed to
look like, and the code says so in as many words.

**The fleet control I ran actually pointed at this and I misread it.** NULL `content_data` is
15.9% of `tool-*` components (36/226) against 1.3% everywhere else (24/1807) — **12× commoner**.
I recorded that as "a known-ish pattern, still a minority" and treated the minority as the signal.
The right reading was that the tool population has a *different rule*, which is what a 12× gap
usually means. A rate difference that large is a prompt to go and find the branch, not to average
over it.

**And the ONE real case is not new** — `tool-simple`'s hero is the same component as work item
`e781118c` ("Component 'hero' on page tool-simple is missing 1 schema-required value field(s):
headline"), which I looked at on 08-21 and reported to the owner as *"probably not a real problem"*
because the live page has a proper `<h1>`. Both readings can hold: the page LOOKS right, and the
stored record is incomplete in a way that would bite a reason-driven re-render. That is worth
saying plainly rather than picking one.

**Corrected in the handoff too.** Chat correction given to the owner the same turn.

## CONTRIB 2026-08-25 from the `deferred_work_item_park` lane (`bugs_open/396`) — your `[UNVERIFIED] what deferred them` is ANSWERED, and the answer is this lane's own backstop

Your NOTES around `:2844` say:

> **[UNVERIFIED]** what deferred them: no Go path writes `deferred` for these types (only migration
> `389`, for `contrast_failure`, on 08-11). Four bulk batches exist (07-31, 08-02, 08-03, 08-05); a
> hand-park at adoption is the obvious guess and I did not establish it.

**It was the hand-park, and the recipe is in this lane's own handoff.**
`HANDOFF_2026-08-03_continue_here.md:81-90` carries it verbatim:

```sql
-- every 15s, defer anything dispatchable that is NOT the items you are running
UPDATE site_work_items SET status='deferred', updated_at=NOW()
 WHERE site_id=(SELECT id FROM sites WHERE domain='mortgagecalculator.co.uk')
   AND status IN ('triaged','approved') AND id NOT IN (<your item ids>);
```

⚠ **That statement sets `status` and `updated_at` and NOTHING ELSE — which is exactly why these rows
carry no provenance stamp**, while other lanes' hand-parks (`loancalculator_rebuild_thread`,
`apis-uk-bees-lane`) wrote `result.deferred_by` / `deferred_reason` / `deferred_from_status` and are
still fully attributable a fortnight on.

Corroboration from your own NOTES, which you had already written: `:305` *"Held 24 of 25 to
`deferred`"*, `:462-473` *"auto-defer loop against a 120-second tick … Final held state: 43
`deferred`"*, `:526` *"The cause was mine. Among the 19 items I auto-deferred were …"*. Today's
surviving residue on this site is **38 rows** [MEASURED 2026-08-25]: 12 `needs_imagery` +
12 `page_rerender` + 5 `needs_content_page` + 3 `needs_page` + 3 `needs_rerender` + 2 `add_tool` +
1 `needs_content_planning`. The `3 needs_page` + `1 needs_rerender` match `:462-473`'s named counts
exactly.

**Nothing here is a criticism of the hold** — three of the four fleet-wide hand-parks were
owner-directed, and yours is documented better than most. The finding is that **the platform has no
park verb**, so every lane that needs one improvises the same `UPDATE`, and only the lanes that
think of it leave a stamp. That is now `bugs_open/396`'s primary fix candidate (§6a).

**Two things you may want to act on:**

1. **These 38 rows are still parked, and still hold their `idx_swi_dedup` slots.** One of them —
   `page_rerender` on `/guides/mortgage-scorecard/index.html`, parked 08-03 — blocked
   `bugs_open/328`'s dispatch on 2026-08-25 with a `23505` that reads as *"already queued"*. It was
   re-armed to `triaged` and **completed in 2 minutes**, and the page now serves correctly. **If the
   adoption's hold window is over, the rest are probably owed the same release** — that is your
   call, not mine, and I have touched only the one row 328 needed.
2. **The `[UNVERIFIED]` marker at `:2844` is worth correcting in place**, since a later reader (me)
   took it at face value and re-derived the whole question from scratch. Logged as my wrong call,
   not yours: an uncertainty marker is a claim about what the writer checked, and I should have
   treated it as a lead to close rather than a boundary to respect.

Full account, the corrected population split and the fix candidates:
`bugs_open/396_HANDOFF_2026-08-25_…md` and
`docs/agent_docs/docs024_key_docs_latest/deferred_work_item_park/`.

## 2026-08-26 — owner reported missing images; ran the framework's design discovery, and "missing" turned out to be THREE different things

### Measured the symptom before triggering anything

Full sweep of all **42** deployed pages (site has grown 32 → 46 pages, 42 deployed, since 08-24):
**76 image references, 9 distinct image URLs, and every one returns HTTP 200** with real bytes.
**Nothing referenced is broken.** So "missing" had to mean images that should be on a page and are
not — which is a different query, and a different fix.

| measure | value (2026-08-26) |
|---|---|
| deployed pages with a content image | 28 |
| deployed pages with **no** content image | **14 — and all 14 are TOOL pages** |
| active assets on the site | 28 |
| distinct images actually referenced | 9 (3 of them chrome: logo, favicon, og-card) |

**12 `content_hero` images and 6 of 11 tool `card` images exist, serve 200 (115–142 KB), and are
referenced by zero pages.**

### Triggered `design-discovery-agent` — it owns all eight image checks

Established that first rather than guessing: the live agent definitions put
`content_image_missing`, `undeployed_assets`, `unfulfilled_imagery_plan`, `image_url_404`,
`asset_reference_404`, `placeholder_image_in_use`, `image_source_unsatisfiable` and
`unfulfilled_image_prompt` **all on `design-discovery-agent`**, none on the other four.

Dispatched scoped to this site (correlation `1abf5acd-9099-4b6d-ba5b-854af2584884`), pre-state
pinned at 438 items so new rows are attributable. Run COMPLETED; **38 new items**.

### What it found — and the three distinct causes

**(1) REAL BREAKAGE the check found and I had not: two tool pages have dead calculators.**
`asset_reference_404` × 2. Verified at the artefact, 2 probes each:

| script | referenced by | status |
|---|---|---|
| `/tools/assets/tool-btl-investor.js` | `/tools/btl-investor/index.html` | **404** |
| `/tools/assets/tool-equity-release.js` | `/tools/equity-release/index.html` | **404** |

Controls: the other two referenced scripts (`/assets/js/snippets.js`, and the lender-directory
listing) both **200**, so the path and the probe are sound.
⚠ **My first control was worthless and I nearly reported it as a finding:** I curled
`/tools/assets/tool-repayment.js`, got 404, and briefly took it as evidence the whole path was
down. **That URL was one I composed** — this lane already filed `bugs_open/387` on exactly that
trap ("a COMPOSED url's 404"). The honest control is a script the pages actually reference.

**(2) Imagery PLANNED but never generated — 12 items DEFERRED since 2026-08-02**, untouched:
7 page heroes (`hero_home`, `hero_about`, `hero_guides`, `hero_tools`, `hero_investor`,
`hero_contact`, `hero_scorecard`), 4 section icons, 1 infographic. All from `build-site-planner`,
all deferred the same day they were filed. Per the inbound CONTRIB of 08-26 from the 307 lane,
**this lane's own 15-second auto-defer backstop deferred them** — the §10c backstop that defers
everything not on your list. So the site has never had its planned page heroes or section icons.

**(3) Imagery GENERATED and never referenced** — the 12 content heroes + 6 cards above. New
`needs_imagery` items were filed for the **7 newest tool pages** (btl-investor, credit-health-check,
deposit-tracker, overpayment-priority, rate-stress-test, remortgage-savings, stamp-duty), which have
no hero asset at all yet.

### The check's own 12 `undeployed_asset` items: 11 mislabelled, 1 a false positive

All 12 named URLs serve **200** on two probes each. So "generated but not deployed to site" is
false for every one of them:

- **11 `content_hero`** — the observation (0 deployed page components reference them) is **true**;
  the wording is not. They are deployed and unreferenced, which is `bugs_open/114`.
- **1 `logo`** — a **false positive**, and of the exact class `bugs_closed/142` was fixed for:
  measured **0** occurrences in deployed `page_components` and **2** in `site_components`, and the
  predicate reads `page_components` only, excluding brand-head purposes through a hardcoded
  two-entry map (`favicon`, `og_card`) that does not contain `logo`. Evidence filed into
  `bugs_closed/142`.

⚠ **A theory I refuted before writing it down.** I first supposed the predicate could never match
because `purpose='content_hero'` (underscore) while files are `content-hero-…` (hyphen). **In SQL
`LIKE`, `_` is a wildcard — it matches.** Checked in SQL rather than reasoned about; and commit
`6c01755dc` is a *previous* session reaching and retracting the identical conclusion. Two sessions
caught by one underscore.

### Also filed by the run, outside the image question

6 `audit_tool`, 6 `improve_tool` (three Tier-2 acceptance failures on the newest tools), 3
`acceptance_run`, 1 `deactivated_component` (*"Site component head points to deactivated component
'Document Head'"* — worth a look, the head is chrome on every page) and 1 `needs_rerender`
(chrome stale). Left for the owner's decision; the three armed items already on the site belong to
another lane's tool work (`tool-deployer`, `tool-improver`, `tool-auditor`) and were not touched.

### State note

`site-discovery-rotation-design` is now **enabled** (10,800 s). The 08-18 handoff recorded it as
`enabled=false` and *"the owner's separate call"* — that has changed since, not by this lane.

## 2026-09-02 — state check for a fresh handoff: BOTH of the owner's targets need re-aiming

### The dead calculators are already fixed — by another lane, in the week we were away

Measured today: `tool-equity-release.js` **200** (was 404); `tool-btl-investor.js` still 404 but
**referenced by nothing** — 0 occurrences in the live page, 0 in stored `page_components`. Both tool
pages now carry ~6 KB of **inline** JS with real inputs and buttons. Item `d5131e25` closed
`complete` 08-27; **`a7c5d5ab` is still `detected` and is STALE.**

**Third stale-open item on this site in two weeks** (after the `unbuilt_internal_link` on 08-21).
The check stands: *a parked item is not evidence its condition survives* — re-probe before acting.

What is actually unresolved: 2 `improve_tool` **failed** (08-27) on
`step load_tool failed: … query_database: query param path 'input_da…'` — an **infrastructure**
failure, so `tool-deposit-tracker` and `tool-remortgage-savings` have acceptance **unverified**,
not failed. That distinction is the whole finding.

### "Wire up the existing images" would have changed nothing — and three pages prove it

The obvious job is to point each tool hero at its own `content-hero-tool-*.jpg`. **That is already
done on three pages and makes no difference:**

| page | `content_data.background_image` | renders a background? |
|---|---|---|
| `tool-equity-release` | its own content-hero | **NO** |
| `tool-overpayment` | its own content-hero | **NO** |
| `tool-simple` | its own content-hero | **NO** |
| 7 other `tool-*` pages | `/assets/images/hero.jpg` | **NO** |
| **8 `tool-*-guide` pages — same component, same field** | `/assets/images/hero.jpg` | **YES** |

So it is not the data and not the template: `html_template` **does** emit
`{{if or .hero_url .background_image}}…url('{{or .hero_url .background_image}}')`, and the value
sits in `content_data` while being **absent from `rendered_html`**. The value is not reaching the
render context, and only on this page type.

⚠ `background_image` is `source: site_assets.hero` — **resolver-populated, not an LLM field**.
MEMORY `the-framework-writes-the-content-not-you` records a wasted run from asking a writer to set a
resolver-owned URL; check the populator before editing anything.

**Had I acted on the owner's instruction literally I would have wired 11 more pages and reported
success, and 14 tool pages would still show no image.** The guide-vs-tool control is what stopped
that — same component, one renders, one does not.

### A fake regression I generated and caught

Re-running the image audit with a **different extractor** (no CSS `url()`) gave *"6 pages with
images, 36 without"* against last week's 28/14 — which reads as a catastrophic regression. Re-run
with the **original** extractor: **28/14, identical.** Distinct images referenced did rise 9 → 12.

**The check: when you re-measure to COMPARE, re-run the original query, not a new one that answers
the same question differently.** A changed instrument and a changed world are indistinguishable in
the number alone. (And `snippets.js` gave one 404 in 13 probes — fine; third transient
false-negative on this site in two weeks.)

### Handoff written

`docs/agent_docs/docs024_key_docs_latest/mortgagecalculator_couk_adoption/HANDOFF_2026-09-02_continue_here.md`,
superseding the 08-21 file. Leads with both re-aimings, because a fresh session acting on either
instruction as stated would spend a day and move nothing.

## 2026-09-02 (b) — "verify the tools, fold in the images": both re-aimings in the 09-02 handoff were themselves wrong

The handoff written this morning (`HANDOFF_2026-09-02_continue_here.md`) re-aimed both of the
owner's asks. Working them, **both re-aimings turned out to be wrong**, in the same way and for the
same reason: they reasoned from `content_data` and the `hero` template without ever asking what the
position-1 row on a tool page actually *is*.

### §2 is refuted: the tool pages do not have a hero that fails to render. They have no hero at all.

The handoff's §2 says the `hero` component renders a background on `tool-*-guide` pages and not on
`tool-*` pages, "same component, same field, same value shape", and sets the next session to
diffing the render path. **There is no render-path difference to find.** [MEASURED 2026-09-02]

| page | slot | component | `rendered_html` length | what the bytes actually are |
|---|---|---|---|---|
| `tool-simple` | `hero` | `hero` | **9,590 B** | `<div class="tool-page"><div class="tool-header"><h1>Simple Mortgage Calculator</h1>…` — **the calculator** |
| `tool-equity-release` | `hero` | `hero` | **14,164 B** | the calculator |
| `tool-equity-release-guide` | `hero` | `hero` | **3,267 B** | `<section class="hero" data-component="hero" style="background-image: …url('/assets/images/hero.jpg')…` |

A hero band is ~3.2 KB. A row of 9.5–22 KB under the `hero` identity is not a hero that failed to
render an image; it is **a whole working tool stored under the shared `hero` component's identity**.
`content_data.background_image` sits on those rows because it is the `hero` schema's field, and it
is inert because **nothing on the page ever renders a hero**. The `rendered_html` was never produced
by the `hero` template at all.

⚠ **This makes the handoff's suggested next step actively dangerous.** "Diff the render path" leads
to "make the hero slot render", and re-rendering that slot through the `hero` template would
**replace a working calculator with a 3 KB title band** on ten pages. That is exactly the damage
`bugs_open/357` exists to prevent and migration `701` exists to make impossible — the CONTRIB of
2026-09-02 from the 357 lane, sitting unread in this directory, says so in its first paragraph.
**I did not read that CONTRIB until after I had measured the row lengths; reading it first would
have saved the measurement.** Read the inbound CONTRIBs before believing the outbound handoff.

### The real composition census — three generations of tool page on one site

[MEASURED 2026-09-02] 18 tool pages (`page_type='tool'`, excluding the 9 `*-guide` pages):

- **10 "adopted"** — one `hero`-identity row holding the calculator, plus a `generic-text-block`:
  affordability, bridging-loan, equity-release, fee-analyser, overpayment, portfolio,
  rate-forecaster, repayment, simple, stamp-duty. **No imagery.**
- **4 "bare"** — a single tool-level component and nothing else: bridging-compound, rate-scenarios,
  deposit-tracker, remortgage-savings. **No imagery.** (deposit-tracker and remortgage-savings start
  at `position=2` — position 1 is absent, not merely different.)
- **4 "native"** — the generator's four-section shape (`hero-tool` + `tool-guide-intro` + the tool +
  `tool-cta`): btl-investor, credit-health-check, overpayment-priority, rate-stress-test.
  **These do carry imagery.**

10 + 4 + 4 = 18, and 14 carry none — which reconciles exactly with the 28-with/14-without page
census the handoff quotes. The 14 are not a render bug. They are pages whose composition contains
no component that can hold a picture.

### Two false findings I generated with careless extractors, both caught in the same session

1. **"All four native pages render a background image."** I tested
   `rendered_html LIKE '%background-image%'` and got `t` for the four `hero-tool` rows. That matched
   the string `background-image` **inside a `<style>` block**, not an image. The `hero-tool` CSS sets
   `background:var(--color-primary,#1a1f36)` — a solid colour.
2. **"`tool-cta` renders the stamp-duty card on every tool page."** I used SQL `substring(… from
   '<img[^>]+src="([^"]+)"')`, which returns **only the first match**. `tool-cta` renders **6
   distinct** card thumbnails (it is a list of other tools); the first happens to be stamp-duty on
   every page. There is no defect.

**The check both needed: count the matches before reading one.** `regexp_matches(…,'g')` with a
`count(DISTINCT)` would have refused to produce either claim. This is the handoff's own §3 lesson
(a changed instrument and a changed world are indistinguishable in the number alone) firing twice
more, one level down — not at the level of "which query", but at the level of "does this regex
return one row or all of them".

### The images: `hero-tool` has an image in its template and no image field in its schema

`plan_sections_action.go:2846` gates the per-page hero on **`sectionHasImageField(fieldsRaw)`** —
the component's `input_schema` must declare a field of `type: image` or `image_url`. Only then is
the resolved page hero written into `resolved_data` under the aliases `hero_url` /
`background_image`. The comment immediately above says what happens otherwise, in the code's own
words: *"without it, `{{or .hero_url .background_image}}` picks the site-wide value and **every page
shows the same image**."*

[MEASURED 2026-09-02, live DB]

| component | declares an image-typed field? | declares `background_image`? |
|---|---|---|
| `hero` | **yes** | yes (`type: image`, `source: site_assets.hero`, `fallback: /assets/images/hero.jpg`) |
| `hero-tool` | **no** | no |

So `hero-tool` emits `url('{{or .hero_url .background_image}}')` in its template while its schema
switches the resolver that fills those keys off. The fleet-wide consequence, and the control that
makes it mean something:

| component | instances | carrying a per-page `content-hero-*` image | sites |
|---|---|---|---|
| `hero-tool` (no image field) | 69 | **0** | 21 |
| `hero` (declares the field) | 632 | **72** | 36 |

**0 of 69 versus 72 of 632.** The measurement could have come out otherwise and did not.
**54 of those 69 pages, across 21 sites, have their own `content_hero_*` asset already generated,
active, and structurally unshowable.** That is a named mechanism for `bugs_open/114`, contributed
there rather than filed new (114 is OPEN, owned and actively worked today).

The resolver itself is not at fault and needs no change: `ensureAssets` already prefers, in order,
the planner's page hero → the Lane B `ContentHeroKey` content hero → the site brand hero. **All 18
of this site's tool pages have an active asset at exactly their `ContentHeroKey`** (`content_hero_`
+ page name with hyphens underscored) — verified row by row. The resolver would find every one of
them. `hero-tool` never asks it.

⚠ The one `hero-tool` instance fleet-wide that does show a per-page image
(`leopardessconsulting.co.uk/tool-automation-savings-estimator`) has `background_image` written
directly into `content_data`. So the **template** works; only the resolver path is switched off.
That is a tempting per-page workaround and it is the wrong fix — hand-setting a resolver-owned URL
is MEMORY `the-framework-writes-the-content-not-you`.

### The tools: 18/18 serve, and 9/18 are outside the verification ladder entirely

`scripts/probe-page-url.sh mortgagecalculator.co.uk <18 tool pages>` — **all 200**, invented-URL
control 404, known-good sibling 200. Fetched all 18 and checked every literal JS binding against the
page's own ids: **0 dangling bindings, 0 template residue (`<no value>`, `{{`) on all 18.**

⚠ **That check has a stated blind spot and must not be read as "the tools work":** it only sees
bindings written as literal strings. `btl-investor` and `fee-analyser` report **0 literal bindings**
— they bind entirely through variables, which is precisely the class `bugs_closed/324` was filed for
("reads clean on every id check and every binding through a variable dangles"). A real verdict needs
a browser, and the platform has one.

**It cannot reach half of them.** Running the platform's own eligibility predicate
(`discovery_checks/tool_eligibility.go`, `toolEligibilityWhere`) against this site returns **9 of
18** pages. Eligible: the 8 with a `component_level='tool'` component, plus `tool-simple` under the
sole-component clause. **Invisible to Tier 2 and Tier 4 alike:** affordability, bridging-loan,
equity-release, fee-analyser, overpayment, portfolio, rate-forecaster, repayment, stamp-duty — the
adopted ten minus simple. They are multi-component (the calculator + a `generic-text-block`) with no
tool-level component, so they satisfy neither clause.

**Seven of those nine still have a current PLAN with a criteria fence** (installed 2026-08-10/11/17
under keys `bridging-loan`, `equity-release`, `fee-analyser`, `overpayment`, `rate-forecaster`,
`repayment`, `stamp-duty`). Those fences are orphaned: nothing loads them. The RUNBOOK §14 warning
"a PLAN under the wrong key produces no error" fired here in a form §14 does not describe — the key
did not change, **the page's shape did**, when a second component was added.

### Why the two `improve_tool` failures are NOT what the handoff says they are

The handoff records both as failing "on INFRASTRUCTURE, not on the tool", concluding acceptance is
"unverified, not failed". **That conflates two different runs.** Read the row: the `summary` is the
*acceptance* verdict, the `error` is the *fixer's* failure.

- `error` (the fixer): `step load_tool failed: … query_database: query param path
  'input_data.spec.page_id' resolved to nil` — infrastructure, correctly identified.
- `summary` (the acceptance run that filed the item): `calculate-shows-results@desktop: step 2
  (select #tracker-deposit-percent) failed: playwright: timeout … waiting for
  locator('#tracker-deposit-percent')`. **That acceptance ran and failed, on a concrete selector.**

And the selector is real — it is the criteria that are stale:

```
html_template  : id="{{.InstanceID}}-tracker-deposit-percent"
rendered_html  : id="c-tool-deposit-tracker-tracker-deposit-percent"
criteria fence : {"action":"select","selector":"#tracker-deposit-percent"}
```

`bugs_closed/283` converted every interactive component to instance-scoped ids. Nothing updated the
acceptance criteria, and **neither checker knows about the prefix**:
`check_tool_acceptance.go:anchorPresent` tests `strings.Contains(html, 'id="'+id+'"')` — an exact
match. So the ladder reports a missing anchor for an element that is present.

Fleet-wide, and disconfirmable — it could have shown these were genuinely invented selectors:
**187 acceptance-fail notes in 45 days name an absent anchor; 134 of them (72%) name an element that
EXISTS in that tool's own template under the `{{.InstanceID}}-` prefix, across 99 distinct tools.**
Control: acceptance is not simply broken — 178 passing `acceptance-run` notes over 127 tools in the
same window.

Filed to the diagnosis loop before asserting the root cause, per CLAUDE.md's cross-cutting rule:
`090` intake `0c852424`, run correlation **`7177c2d6-fe22-40c4-b9bc-b53f93ec59c9`**, work item
`f49713ae`. No prior filing: the `needs_diagnosis` queue was empty and neither `/bugs_open/` nor
`/bugs_closed/` carries the mechanism (283's five CONTINUE_HERE files never mention criteria or
acceptance at all).

### Migration 701 will orphan eight more fences, including its own pilot's

701 retypes the adopted rows to `component_level='tool'` with `new_function` = `tool-<slug>`. The
ladder's subject key is `cc.function` when the component is tool-level, so post-701 the key for
these pages moves from `<slug>` to `tool-<slug>` — and all eight of this site's current PLANs are
keyed `<slug>`. **`tool-simple`, 701's designated pilot, is the one page whose fence works today.**
Net, 701 is good for verification (it makes nine invisible tools eligible); it needs the eight
`doc_plans` rows re-keyed in the same change. CONTRIB written to the 357 lane.

### One stale item closed, with a demand control that first caught my own blind query

Item `a7c5d5ab` (`/tools/assets/tool-btl-investor.js` returns 404) — **cancelled**, condition gone.
The script is still 404 but nothing references it: 0 of 18 live pages, 0 stored `page_components`.

⚠ **My first control was worthless and I nearly recorded the zero on it.** I controlled with
`snippets.js` — referenced by all 42 pages — and the stored-component query returned **0 for it
too**, because `snippets.js` is injected by the chrome assembler and never appears in
`page_components`. A control that returns the same zero as the target proves nothing. The honest
demand control is `/tools/assets/mortgage-lender-directory-listing.js`, which **is** stored in a
component and returns **1**. This is the lane's fourth composed-or-blind-probe near-miss
(`bugs_open/387`'s family) and the second in this file where the *control*, not the target, was the
defective half.

### The 090 came back UNVERIFIABLE — and naming its own two gaps is what made the case airtight

`f49713ae` completed **`UNVERIFIABLE`**, `stopped_by: iteration-cap`. **Not a refutation** — it ran
out of iterations having failed to reach two facts, and it named both precisely:

1. it could not tell **which** component row serves the failing URL
   (`garden-tools.uk/tools/watch-service-interval-calculator/index.html`), because two rows share the
   function and the bundle carried no `site_id`/`created_at` to separate them;
2. **`check_tool_acceptance.go` was not in its bundle at all** — so it had no code evidence that the
   selector resolution is literal rather than instance-aware.

Both were cheap to close by hand and closing them produced better evidence than the loop would have:

- **Gap 2** — `anchorPresent:565-567` is `strings.Contains(html, 'id="'+id+'"')`. Exact match, no
  prefix awareness. Read, quoted in the bug file.
- **Gap 1** — one function, two active rows, ONE criteria fence keyed on `function`. Curled both,
  200 each: `garden-tools.uk` serves `id="c-tool-watch-service-interval-calculator-calc-btn"`
  (fence's `#calc-btn` matches **nothing**), `relojistas.com` serves `id="calc-btn"` (matches).
  **Same fence, opposite verdicts, and the passing site is the positive control** that rules out
  "the fence is malformed" and "the checker is simply broken".

That second finding is worth more than the confirmation would have been: **a fence keyed on
`function` cannot be right for a function whose rows differ in scoping.** [MEASURED 2026-09-02]
**10 tool functions are split** (≥1 converted and ≥1 unconverted active row); **6 of those hold a
current criteria fence** and are therefore unsatisfiable by construction. So candidate 2 — "re-emit
the fences from the live pages", which the lane's own `toolgolden.py` does automatically — **cannot
be the fix**, only a follow-up. Only a scope-aware checker satisfies a split function.

Filed as `bugs_open/441`, which states in §2 that the loop returned UNVERIFIABLE and that the filing
rests on first-hand verification instead, per the owner ruling of 2026-07-31.

⚠ **Do not read this as "the 090 was useless."** It cost one run and it is the reason the bug file
carries a quoted function and a two-site control rather than an assertion. **A loop that stops and
tells you what it could not see is doing its job** — and on this occasion the seeding gap it
reported (a symptom naming symbols in a file the bundle does not include) is itself worth someone's
attention.

### Two housekeeping facts from the commit itself, both worth recording

**1. Three of my nine files were swept into other sessions' commits between my write and my
commit.** I named nine paths on `git commit`; six landed under my message (`d6b612a5f`). The other
three — `bugs_open/114`, `016b`, `WRONG_CALLS.md` — were already committed by the lanes that own
them (`26242df2c`, `d0c8ca9c3`, `01276a88e`) because we edited the same files in the same window.
**Nothing was lost:** all three are at HEAD, verified with `git show HEAD:<path> | grep -c`, not by
looking at the working tree (which cannot distinguish "committed" from "still dirty"). This is
CLAUDE.md's stated same-file passenger case — a pathspec commit protects you from *others'* files,
never from a co-edit of *one* file — and the remedy is the one the rule gives: finish, commit the
remainder, say so. ⚠ **The check that matters here is `git show HEAD:<path>`, not `git status`.**
A clean status on a file you just edited means *someone committed it*, and that someone may not be
you.

**2. The `bugfix_114_imagery_wiring` lane reached my §1 conclusion independently, on the same day,
and supplied the cause I did not have.** Their CONTRIB
(`CONTRIB_2026-09-02_from_114_your_render_path_diff_is_bug_357_not_a_divergence.md`) says do not diff
the render path, it is `bugs_open/357` — with the same measurement (tool `hero` rows 9.5–14.5 KB of
`<div class="tool-page">`, guide heroes 3.3 KB of real hero markup). **Verified their citation
rather than accepting it:** `platform/orchestration/actions/adopt_fragment_section.go:14-15` states
it in the file's own header — a fragment is stored under the sentinel name `"section"` (*identity
unknown*) and **that sentinel is then replaced by `planned[Position-1]` from `pages.sections`**, so
the row inherits whatever identity the plan had at that position. That is the mechanism; I had the
symptom and the danger but not this.

Two corrections to my own reading follow from theirs, and one thing of mine survives intact:

- **The 12 content-hero images are NOT waste.** I had them filed mentally under "generated and
  unreferenced". The event-driven card derive (IMG-073) is live and 193/193 fleet-wide, and a
  content hero on a page that cannot render it is **still the card source** — our tool cards exist,
  are entity-linked and serve 200. Do not let anyone cite this site as evidence that the generation
  was wasted.
- **The `undeployed_asset` mislabelling I recorded at §4 of the morning handoff generalises to
  1,651 rows fleet-wide**, born at `unresolved` by the recurrence brake. There is a new LANDMINES
  entry on why draining them is the trap.
- **My `hero-tool` finding is NOT their finding and is not covered by 357.** Theirs explains the
  **ten adopted** pages (no hero exists). Mine explains the **four native** pages (a hero exists,
  declares no image-typed field, so `sectionHasImageField` gates the resolver off). Disjoint sets,
  disjoint mechanisms, and 357's constructive adoption fixes only the first. The CONTRIB I filed
  into `bugs_open/114` stands.

## 2026-09-02 (c) — the site divides perfectly in two, and exactly one tool is verifiable today

Continuing after the token expired (see the block below). Two cluster-free measurements, and
together they are the sharpest statement of where this site stands.

### Every existing fence is VALID — against the pages it was written for

I tested all 48 `#id` selectors in the lane's 9 stored criteria files
(`acceptance/criteria/*.criteria.json`) against the live pages I had already fetched.
**0 missing, on all 9.** So the morning assumption that this site's fences are stale is wrong: they
are correct, and 7 of them are simply never read.

### The reason is one property, and it splits the 18 pages cleanly

`grep`ping the fetched pages for `id="c-tool-…"` versus bare ids:

| shape | pages | eligible for the ladder? | fence state |
|---|---|---|---|
| **instance-scoped** (`id="c-tool-…"`) | 8 — bridging-compound, btl-investor, credit-health-check, deposit-tracker, overpayment-priority, rate-scenarios, rate-stress-test, remortgage-savings | **YES** (all have `component_level='tool'`) | **STALE — `bugs_open/441`** |
| **bare ids** | 10 — affordability, bridging-loan, equity-release, fee-analyser, overpayment, portfolio, rate-forecaster, repayment, simple, stamp-duty | **NO for 9** (multi-component, no tool-level row) | **VALID** |

The two axes are the same axis. The instance-scope conversion only ever touched components with a
`content_components` row — which is exactly the eligibility criterion — so **every tool that the
ladder can see has a stale fence, and every tool with a good fence is invisible to the ladder.**

**`tool-simple` is the sole exception**: bare ids (so its fence is valid) *and* eligible (sole
component on a `page_type='tool'` page, key `simple`). **It is the one tool on this site that can be
verified today.** [UNMEASURED — needs the cluster] whether its last Tier-2/Tier-4 run actually
passed; the fence being satisfiable is necessary, not sufficient.

⚠ **And `tool-simple` is migration 701's designated pilot.** 701 gives it a `component_level='tool'`
row, which moves its ladder key from `simple` to `tool-simple` and orphans the fence — so **the one
verifiable tool on this site is the one 701 is about to make unverifiable**, and the pilot will look
completely healthy while it happens. Strengthened the CONTRIB to `bugs_open/357` accordingly; this
is the concrete case for the re-key `UPDATE` landing in the same transaction.

### The second failure in the two stuck items is now diagnosed from source: `bugs_open/448`

The `improve_tool` items for deposit-tracker and remortgage-savings died at
`query param path 'input_data.spec.page_id' resolved to nil`. Cause, read at HEAD:
`JudgeAcceptanceResultsAction` (`tool_acceptance_actions.go:867-874`) **re-derives** `page_id` with

```sql
LEFT JOIN pages p ON p.id = pc.page_id AND p.site_id = $2::uuid   -- site filter on the JOIN, not the WHERE
... WHERE cc.function = $1 AND cc.is_active LIMIT 1               -- and no ORDER BY
```

so a function with rows on more than one site yields `p.id` NULL → `''` → the conditional at :994
omits `spec.page_id` → the fixer's first step has nothing to read. `tool-deposit-tracker` has
exactly that shape: two active rows, one of them the loanandmortgagecalculator fork.

**The reliable code already exists 300 lines below** — `routePortedAcceptanceFailure:1213` reads
`input_data.spec.page_id` from the run item it was handed, which `check_tool_acceptance_due` always
writes. The file uses the sound path for the PORTED route (a human queue with no automated fixer)
and the unsound one for the route that feeds `tool-improver`. That is backwards from where
reliability matters.

⚠ **Filed with the 090 NOT run, and the bug file says so** — the cluster token expired, so no
dispatch was possible. Two blast-radius queries are left in `448` §5 marked `[UNMEASURED]`. **Do not
quote a size for 448 until they are run.**

### ⚠ Blocked: kubeconfig token expired mid-session

`kubectl` returns `Unauthorized` fleet-wide (`kubectl version` confirms *"the server has asked for
the client to provide credentials"*). That is the known 3-day expiry; the owner refreshes it. Work
that stopped at that line, in the order it should resume:

1. Confirm `tool-simple`'s last acceptance verdict — the one tool that should pass today.
2. Run `448` §5's two queries and put real numbers in the bug file.
3. Re-emit fences for the 8 scoped tools from the live pages (`toolgolden.py --emit-criteria`
   drives the deployed page, so it picks the `c-tool-…` ids up automatically), `verify_criteria.py`
   to 0 MISMATCH **plus the mutation test exiting 1**, then `install_fences.py --apply` and fire
   Tier-4 runs. ⚠ This is a *lane-level* workaround, NOT `441`'s fix: it re-breaks at the next
   conversion and cannot satisfy a split function. None of this site's 8 is split
   `[UNMEASURED — the query is in the (c) block above, needs the cluster]`, so it is safe here and
   nowhere in general.

## 2026-09-02 (d) — token refreshed; the real scoreboard, and TWO corrections to my own claims from three hours ago

### CORRECTION 1 — "every ladder-visible tool has a stale fence" was WRONG. Only two are stale.

`## 2026-09-02 (c)` above says the site splits cleanly and *"every tool the ladder can see has a
stale fence"*. **That does not survive testing against the platform's own anchor rule.** I inferred
staleness from "the page is instance-scoped"; staleness actually requires the FENCE to name a
scoped id, and most of these fences anchor on classes or on wrapper ids the conversion never
touched.

Re-tested by reimplementing `selectorAnchor` (`^\s*([#.]?[A-Za-z][A-Za-z0-9_-]*)`) and
`anchorPresent` exactly, over every check's `selector`, `steps[].selector` and `expect.selector`:

| fence | anchors | absent | verdict |
|---|---|---|---|
| `tool-bridging-compound` | 3 | **0** | satisfiable |
| `tool-overpayment-priority` | 3 | **0** | satisfiable |
| `tool-rate-scenarios` | 3 | **0** | satisfiable |
| `tool-deposit-tracker` | 9 | **8** | STALE — `441` |
| `tool-remortgage-savings` | 9 | **7** | STALE — `441` |

**So `441` blocks exactly TWO tools on this site, not eight** — and they are precisely the two whose
`improve_tool` failed. ⚠ **My first version of this test was also wrong**, in the opposite
direction: I matched the whole selector string as an id, so `#bridgeForm button[type=submit]`
counted as "missing" and bridging-compound read as stale. **Implement the platform's rule, do not
approximate it** — the anchor is the LEFTMOST token, and that one detail moved a tool between
buckets.

Three of the nine eligible tools (`tool-btl-investor`, `tool-credit-health-check`,
`tool-rate-stress-test`) have **no fence at all** — `needs_criteria` notes, no PLAN. They are
unverified for want of a fence, which is a different problem from `441` and cheaper to fix.

### CORRECTION 2 — the split in `bugs_open/441` is at the RENDERING, not the template

Filed as 10 split functions / 6 with a fence, measured over `html_template`. The fence is judged
against the **deployed page**, so the right surface is `page_components.rendered_html`. Re-measured:
**214 tool functions with placements, 16 split at the rendering, 8 of those holding a fence** (was
10/6). And the cause is different from what I filed: `tool-credit-health-check` and
`tool-rate-stress-test` have **every active row scoped** and still serve bare ids on
`loancalculator.co.uk`, because those two placements were last rendered **2026-08-02** and
**2026-08-09** — before the conversion. **A converted template does not convert the pages already
built from it.** Corrected visibly in the bug file; it makes candidate 1 stronger, since you cannot
fix a stale rendering by converting a template.

### The actual verification scoreboard, from the verdict record

| tool | last PASS | last FAIL | state |
|---|---|---|---|
| `simple` | 08-26 | — | **PASSING** (desktop; 4 mobile checks SKIPPED) |
| `tool-overpayment-priority` | 08-27 00:46 | 08-26 19:11 | **PASSING** |
| `tool-rate-scenarios` | 08-26 11:58 | 08-25 23:53 | **PASSING** |
| `tool-bridging-compound` | 08-09 | 08-26 11:57 | **stale FAIL** — see below |
| `tool-deposit-tracker` | — | 08-27 | FAILING — `441` |
| `tool-remortgage-savings` | — | 08-27 | FAILING — `441` |
| `tool-btl-investor` / `tool-credit-health-check` / `tool-rate-stress-test` | — | — | no fence |
| the 9 adopted pages | — | — | ineligible (needs `701`) |

### `tool-bridging-compound`'s failure is STALE — it was repaired the same day and never re-run

Verdict `2026-08-26 11:57:52`: *"expected element `#results` absent after interaction"*, desktop and
mobile. But `#results` **is** in the served page today, and the expect test is
`page.Count(selector) == 0` — a DOM count, **not** a visibility test (`run_checks_action.go:750`),
so a hidden-but-present div would pass. The explanation is chronology, not mechanism:

```
11:57:52  Tier-4 FAILS, raises improve_tool f80d3397
12:21:03  f80d3397 completes  (the repair)
23:07:24  page_components re-rendered          <- the fix reaches the page
   —      no acceptance run since
```

So the tool was fixed nine hours after it failed and **the failing verdict is the newest one on
record**. ⚠ **This is a general shape worth carrying: a FAIL outlives its own repair whenever the
fixer does not re-run the check.** Anyone reading the verdict record — including me, three hours
ago — sees a failing tool. Fired a fresh run to settle it: `acceptance_run` **`21b2d81d`**,
dispatched 21:27, coverage-checked first (no open acceptance work on that target).

⚠ Also worth noting from the `simple` pass: **4 mobile checks were SKIPPED** and only desktop ran.
A "PASSED" verdict here means *passed on desktop*. `[UNMEASURED]` why the mobile profile skipped.

### The deepest answer to "verify the tools": where the ladder DOES run, it is not checking the sums

Chasing why `simple`'s pass reported *"4 skipped: computes-defaults@mobile, …"* led somewhere much
larger than a skipped profile. The skip itself is benign — `simple`'s fence declares doc-level
`profiles: ["desktop","mobile"]` while every check is pinned `profiles: ["desktop"]`, so the mobile
pass has nothing applicable. But reading the fence to find that out exposed the real gap.

**The two fence families on this site test disjoint things, and neither is complete:**

| family | check types | asserts | blind to |
|---|---|---|---|
| lane-authored (10, installed 08-10..08-17) | `computed_values` ×4 | **the numbers are right** | whether the page loads, errors, or fits a phone |
| `tool-generator` (5) | `interaction`, `no_console_errors`, `no_horizontal_overflow`, `page_status_ok`, `selector_exists` | **the tool is alive and responds** | **whether any number it prints is correct** |

Fleet-wide, and this is `bugs_open/449` (filed): **`tool-generator` has written 170 current fences;
107 of them (63%) assert no expected value of any kind, and 0 use `computed_values`** — a check type
the runner implements (`run_checks_action.go:708`, `:809`) and that works, since `simple` passes
four of them. Cause, in one place: neither fence-authoring agent's `default_config` mentions the
type. `tool-generator` and `experience-planner` both know `interaction` and `selector_exists`;
neither knows `computed_values`. **The type is absent from the prompt, so it is never a candidate.**

⚠ **My first version of that finding was too strong and I corrected it before filing.** I wrote
"zero generated fences assert a computed value" — true of the check *type*, false of the behaviour:
**63 of the 170 use `interaction.expect.text_matches`**, and some patterns are real values
(`\$1000\.00`, `40.0%`). The defensible number is the 107 that assert nothing. And
`operator:staged_component_build` is the existence proof that both can live in one fence — 6 of its
8 carry `computed_values` *and* `interaction`.

**What this does to our three "PASSING" verdicts — say it this way to the owner:**

- `tool-overpayment-priority`, `tool-rate-scenarios`: their fences contain no value assertion at all.
  PASS = *the page loads and something appears when you click*. **Nothing about the arithmetic.**
- `simple`: PASS = *four sums are right, on desktop*. No boot, console, status or mobile check.

**Not one tool on this site is verified for both correctness and health.** That is the honest answer
to the owner's question, and it is a better answer than a green tick would have been.

### Run dispatched

`acceptance_run` `21b2d81d` — claimed by `build-dispatch-loop` at 21:32:29, 5m38s after insert
(RUNBOOK §14 says ~3 min to claim, ~30 min to complete; 5–6 min is within normal queue latency, not
a fault). Verdict lands as a `doc_notes` row, **not** on the work item — read the note, not the
status.

### VERDICT: `tool-bridging-compound` PASSES — the stale-fail diagnosis is confirmed at the artefact

`acceptance_run` `21b2d81d`: inserted 21:27:00, claimed 21:32:29, **complete 21:33:17** — 6m17s
end to end, of which the run itself was 48 seconds. (RUNBOOK §14's "~30 minutes created→complete"
is the neighbouring lane's figure under load; note the real spread rather than treating 30 min as
the expectation.)

**doc_note `f6d333ef` — Tier-4 acceptance PASSED, all 9 checks, desktop AND mobile:**
`boots@desktop, status@desktop, calculate-shows-results@desktop, console@desktop, boots@mobile,
status@mobile, mobile-fit@mobile, calculate-shows-results@mobile, console@mobile` (1 skipped,
`mobile-fit@desktop`, correctly — it is pinned to mobile).

**`calculate-shows-results` is the exact check that failed on 08-26 on both profiles.** It now
passes on both. So the chronology diagnosis holds: repaired 12:21, page rebuilt 23:07, and the
failing verdict simply outlived its own fix by never being re-run. **This is the confirming
instance of that shape** — I predicted the pass from the timestamps before firing the run, and the
run agreed.

The accompanying `render-critique` (`24466871`) returned **FINDINGS: none** on both profiles. It
noted that several footer links appear under both "Quick Links" and "Explore" and judged it
intentional site structure; `[UNCHECKED]` whether that duplication is wanted — it is site chrome,
not this tool, and out of scope here, but someone should decide.

### Scoreboard after the run

| state | tools |
|---|---|
| **PASSING** | `simple`, `tool-overpayment-priority`, `tool-rate-scenarios`, **`tool-bridging-compound`** (4) |
| FAILING — `441` stale fence, fixer blocked by `448` | `tool-deposit-tracker`, `tool-remortgage-savings` (2) |
| no fence at all (`needs_criteria`) | `tool-btl-investor`, `tool-credit-health-check`, `tool-rate-stress-test` (3) |
| ineligible for the ladder — needs `701` | the 9 adopted pages |

⚠ **All four passes are still subject to `bugs_open/449`** — `bridging-compound`'s fence is the best
on the site (9 checks, both profiles, boot + status + console + overflow + interaction) and it
**still asserts no number**. It proves the calculator responds; it does not prove the calculator is
right.

## 2026-09-03 — 701 landed overnight; both predictions came true, and one effect nobody predicted

Migration `701` was applied by the owner ~22:00Z on 09-02 (`bugs_closed/357`, population 0).
Owner instruction this morning: **hand the imagery job to the 114 lane and carry on.** Done — see
§"Imagery handed over" below. This section is the tools half.

### Prediction 1 confirmed: eligibility jumped 9 → 18

Every tool page now carries a `component_level='tool'` component, so all **18** satisfy the ladder's
eligibility predicate (was 9). That is 701's real gift to this lane and it is bigger than the
migration's own notes claim.

### Prediction 2 confirmed: all 8 fences were orphaned, and nobody re-keyed them

The ladder keys on `cc.function` for a tool-level component, so the keys moved `<slug>` →
`tool-<slug>`. The 8 `doc_plans` fences were still at the old keys — **addressing nothing**, with no
error anywhere: Tier 2 writes `needs_criteria`, Tier 4 emits nothing. Exactly what the CONTRIB to the
357 lane warned, including that `tool-simple` — the site's only working verification the day before
— would be the most invisible casualty.

### The effect I did NOT predict, and it is the interesting one

**701 created every adopted component with an instance-scoped template (`{{.InstanceID}}-`) while
preserving the pre-existing rendered bytes.** Those two states agree only until something renders.
This morning's rebuild wave re-rendered **5 of the 10** (08:46–08:49Z) and their served ids changed:
`amt` → `c-tool-simple-amt`. So as of today the site is **half-converted — 5 tools scoped, 5 bare,
all under scoped templates** — and the remaining 5 convert whenever they next render.

| | tools |
|---|---|
| template scoped | **10 of 10** |
| rendering scoped | **5** — equity-release, fee-analyser, rate-forecaster, repayment, simple (re-rendered 09-03 08:46–08:49) |
| rendering still bare | **5** — affordability, bridging-loan, overpayment, portfolio, stamp-duty (untouched since the 701 transaction, all at `2026-09-02 21:06:35`) |

**Nothing is broken by it.** I checked all ten for dangling JS bindings: **0**. The converter
rewrites bindings alongside ids, as designed. **But each of those five re-renders silently
invalidated that tool's acceptance fence.**

⚠ **This is the sharp form of `bugs_open/441`, and it changes what that bug IS.** I had it as a
backlog of stale fences left by an August conversion. It is not a backlog — **it is a live generator
of stale fences.** Every re-render of an adopted-then-scoped tool breaks its fence, and five fired in
four minutes this morning without anyone intending it. Candidate 2 (re-emit the fences) would have to
run after every render, for ever. Only the scope-aware checker is stable.

⚠ **And it retires "bytes unchanged, md5-verified" as a sufficient guarantee for an adoption
migration.** That claim was TRUE at apply time and stopped being true at the next render, because the
migration changed the *template the next render would use*. A byte-equality guard proves the
migration moved nothing; it cannot promise the next render will not. Contributed to
`bugs_closed/357` post-close, not as a reopen.

### What I did: re-addressed all 8 fences, verified at the artefact both sides

Supersede-and-insert, one guarded transaction, `doc_plans`:

- **subject key** `<slug>` → `tool-<slug>` for all 8.
- **selectors re-pointed** with the `c-tool-<slug>-` prefix for the 5 whose pages are already
  scoped (16, 16, 20, 28 and 24 selectors respectively); the 3 bare ones' selectors untouched.
- **no expected value changed** — the arithmetic is the same; only the addresses moved.

Checks run BEFORE writing, each of which could have stopped it:

1. **No key collision** — none of the 8 `tool-<slug>` keys had a current plan
   (`idx_doc_plans_current` is UNIQUE on `(subject_type,subject_key) WHERE is_current`).
2. **No other consumer** — each old key resolved to mortgagecalculator.co.uk and no sibling domain,
   so re-keying moves nothing for another lane. (`doc_plans` keys are fleet-wide; this was the
   caution I gave the 357 lane and then owed myself.)
3. **Every new selector present in the live page** — all 8 fences SATISFIABLE, 0 absent anchors,
   tested with a verbatim reimplementation of `selectorAnchor` + `anchorPresent`.
4. **The control that proves the transform did work:** the OLD fences re-tested against the 5 scoped
   pages come back 4, 4, 5, 7 and 6 anchors **absent**. Had the transform been a no-op, this reads 0.
5. **In-transaction assertion** — `RAISE EXCEPTION` unless exactly 8 new current and 0 old current.
   It returned `OK: 8 new current fences, 0 old remaining`.

⚠ These 8 fences carry **`no_auto_fix: true`**, and it matters here: a failing arithmetic verdict
reaches a human, not `tool-improver`. So re-keying could not dispatch a rewriter at a working
calculator — *"the only way an automated rewriter can turn a red arithmetic fence green is by
changing the numbers"*, as the fence's own reason field puts it. I checked that before dispatching,
not after.

**Result: 13 of 18 tool pages now resolve to a live fence** (6 before). The 5 without one are
`tool-affordability`, `tool-btl-investor`, `tool-credit-health-check`, `tool-portfolio`,
`tool-rate-stress-test` — writing those is real work needing a non-page source per expected value,
and is the next job here.

**8 Tier-4 runs dispatched** (`879ee87a` simple, `076377ba` bridging-loan, `a570b486` rate-forecaster,
`36db5755` repayment, `b1bdb777` stamp-duty, `2efd13f7` fee-analyser, `89a3cc7a` overpayment,
`d9d4dce0` equity-release). Verdicts land as `doc_notes`, not on the work item.

### Imagery handed over

Owner: *"hand that lane the whole job."* Written to
`bugfix_114_imagery_wiring/CONTRIB_2026-09-03_from_mcalc_lane_OWNER_HANDS_YOU_THE_WHOLE_TOOL_IMAGERY_JOB.md`,
with a pointer + correction appended to `bugs_open/114`. **The correction matters:** our 09-02 claim
that ten pages *"have nowhere to put a picture"* is **no longer true** — 701 freed the hero slot, so
the composition change they need is safe as of today and was not yesterday. We keep the tools as
product (fences, Tier-4, `441`/`448`/`449`); they own imagery and the spend.
