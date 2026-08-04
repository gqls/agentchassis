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
