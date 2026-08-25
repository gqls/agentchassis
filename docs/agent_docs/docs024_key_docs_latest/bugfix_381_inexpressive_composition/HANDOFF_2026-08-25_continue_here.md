# HANDOFF — `bugs_open/381`, 2026-08-25. **Both halves are LIVE. One thing is left and it needs the owner.**

**Lane:** `docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/`
**Bug:** `bugs_open/381_HANDOFF_2026-08-24_the_planner_composes_pages_from_components_that_cannot_express_the_page_it_planned.md`
**State: THE VALIDATION BUILD IS RUNNING. Everything built is applied, verified and evidenced; the
one open item is now in flight — see §0.**

---

## 0. ⏳ LIVE RIGHT NOW — the validation build (`homegarden.uk`), dispatched 2026-08-25 **10:21Z**

The owner offered a domain list and authorised the build this lane was waiting for.

| | |
|---|---|
| domain | **`homegarden.uk`** |
| site id | `5904bd0f-33fd-4212-9c1b-50b28fe72fdb` |
| correlation | `f20ddbf6-d512-4d55-8b3d-3276717c0c39` |
| orchestration | `0ca5de49-6a47-4d5b-8e08-a742427769ec` |
| dispatched | **2026-08-25 10:21:49Z** (site row `created_at`, UTC) — **LANDED** (receipt-asserted, `bugs_open/327`) |
| first state | submitter `COMPLETED`; site row `active`/`build=pending`; `needs_domain_research=triaged` |

> **⚠ CORRECTED 2026-08-25 by the `loanzy_uk_example_site` lane: this section first said "~11:15Z",
> which was BST READ AS UTC.** The site row is `created_at 2026-08-25 10:21:49.579398+00`; the shell
> reported `date -u = 10:26Z` against `date local = 11:26+0100` in the same breath. Nothing about the
> build changed — but **every elapsed-time figure derived from the wrong stamp would have been an hour
> out and would have read as a stall that never happened.** That lane's own 08-24 handoff §3 carries
> this exact trap: **kubectl/klog lines are LOCAL, the DB is UTC — stamp `date -u` in the same command
> as any time you record.**
>
> **⚠ AND "LANDED" COVERS THE DISPATCH, NOT THE BUILD** (same lane, fair pushback, accepted).
> The receipt asserts the Kafka message was published and consumed. It says nothing about pages.
> **Nothing about `homegarden.uk` is proven until pages serve.**

**Why this domain:** the only one on the offered list that plausibly exercises all three new
components at once (the gardening year → `period-calendar`; what-to-check → `checklist`; choosing
between options → `comparison-table`), and the **nearest neighbour to `garden-tools.uk`**, so it is a
genuine before/after on the same kind of subject. Insurance and health domains were deliberately
declined: `comparison-table`'s claims exposure is real and unmitigated on a site with no evidence
base, and a new site starts with none — a wrong claim about compost is not a wrong claim about cover.
`indoorplanters.co.uk` was declined too — it already exists with 0 pages and would confound.

**The brief was deliberately NOT rigged.** It describes the subject, audience and tone and names no
component and no calendar; a brief that asked for a month-by-month section would have tested only
whether the framework obeys me. It is at `<scratch>/mission_homegarden.txt`.

⚠ **The dispatch script is NOT executable** (`-rw-rw-r--`, unlike `097_TRIGGER…`). `./082_…` fails
with *Permission denied*. **Use `bash <script>`; do not `chmod`.**

### ▶ JUST RUN THIS — the four reads are a script, already dry-run

```bash
bash docs/agent_docs/docs024_key_docs_latest/bugfix_381_inexpressive_composition/ACCEPTANCE_homegarden.sh
```
Written and **dry-run 2026-08-25 10:47Z while the build was still at hop two**, deliberately, so it
is not being debugged at the moment it matters — Q4 reads a table with a ~25h retention and a SQL
error then would cost the only chance to see what the planner was told. **Every query reports
"not yet" rather than an empty result**, because on this lane an ambiguous zero has been misread
three times in two days. It prints how to read itself, including the two different meanings of a
zero in Q1.

### WHAT TO DO NEXT — in this order, and step 2 EXPIRES

Reference timing from `garden-tools.uk`: submission 17:17 → pages 20:15, **about 3 hours**.

1. **Is it progressing?**
   ```sql
   SELECT item_type, wi.status, handler_agent FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
    WHERE s.domain='homegarden.uk' ORDER BY wi.created_at;
   ```
2. **⚠ READ WITHIN ~24h OF THE PLANNER RUNNING — `orchestration_states` is a rolling window.**
   `[MEASURED 2026-08-25 10:32Z]` — I had been repeating this figure from memory; it is now checked,
   and it holds: 7,659 rows, oldest **1 day 00:55** old, **nothing beyond 48h**, and the tail is
   thin (1,129 rows <6h, 6,368 at 6–24h, **162** at 24–48h, 0 older). So the honest instruction is
   **read it the same day; by ~25h it is gone**, and a row still present at 30h is the exception
   rather than the rule. (Checked because the `loanzy_uk_example_site` lane made the point that a
   **caveat is an assertion wearing the grammatical form of modesty** — nobody marks a warning
   `[INFERRED]`, so an inherited caution can sit in the imperative for ever without anyone measuring
   it. This was mine: correct, and unmeasured until now.)
   Pull the rendered `plan_site` prompt and confirm it really carries `[expresses: …]` /
   `[prose only]` per component and rule 19. **This is the only chance to see what the planner was
   actually told**, and it is the difference between "it was offered a calendar and declined" and
   "it never saw one".
3. **THE HEADLINE RESULT:**
   ```sql
   SELECT cc.function, count(*) FROM page_components pc
     JOIN content_components cc ON cc.id=pc.component_id JOIN pages p ON p.id=pc.page_id
    WHERE p.site_id='5904bd0f-33fd-4212-9c1b-50b28fe72fdb'
      AND cc.function IN ('checklist','period-calendar','comparison-table') GROUP BY 1;
   ```
   **Any non-zero closes this lane's open item.**
4. Page structure on THIS site's pages only, plus the loanzy lane's promise-vs-delivery check.

### ⚠ THIS BUILD IS ALSO `bugs_open/206`'s CLOSURE TEST — division of labour agreed

`reconcile_site_plan` runs **inside `build-site-planner`**, last in its order (after `plan_site` /
`write_site_plan` / `sync_pages` / design / imagery / nav). So the step this lane is waiting on to
test 381 is the same step that **mints the `needs_page` rows**, and the owner's 2026-08-25 retraction
moved 206's proof onto "the next greenfield build". **This is that build.** Agreed split with the
`loanzy_uk_example_site` lane so neither of us measures half of it:

- **This lane takes the PROMPT and the CHOICE** — what the planner was shown, and what it composed.
- **That lane takes the MINT** — what `reconcile_site_plan` then filed, contributed to the 206 lane
  rather than written up as theirs. Their assertion, so it can be sanity-checked: a row with
  `page_role='entity-directory'` must carry `handler_agent='directory-build-handler'` and
  `created_by='reconcile_site_plan'`, joined `pages` on `(site_id, page_name)` as the authority —
  **NOT** on `spec->>'page_type'` or `'page_id'`, which are absent on 134 of 134 reconcile-minted
  rows fleet-wide and so return a confident zero for the very population they exist to count.

⚠ **AND IF PAGES FAIL WITH "no sections ready to build", DO NOT LET IT READ AS A 381 FAILURE.**
That is 206's residual class and it cost garden-tools **5 of 12 pages**. Ask which of two it is,
because **the string cannot tell you and the fixes differ**: (a) a type mapped to a wrong or missing
builder, or (b) a type mapped to a builder that cannot fill a missing layout —
`ensure_page_section_layout` exists **only** in `directory-build-handler`'s workflow, so the generic
path has no way to create one.

### ⚠ THE PLANNER PROMPT IS BEING CAPTURED AUTOMATICALLY — do not rely on noticing

`evidence/PLANNER_PROMPT_homegarden_<ts>.txt`, written by a background watcher the moment
`plan_site` appears in `llm_call_log`. **Why it is automatic rather than a step in a checklist:**
the pages show what the planner CHOSE and **never what it was SHOWN**, and if the choice
disappoints those two readings have completely different remedies — one is "it declined", the other
is "the capability never reached it". `prompt_rendered` is durable in `llm_call_log`; the
**orchestration row that proves which run it was is not**, so the capture takes both together.

### How to read the outcome

- **Chosen and filled** → central claim proven; close `bugs_open/381` to `/bugs_closed/`.
- **Chosen but thin** (twelve near-identical months) → the schema weakness named as risk 3 in council
  `c134b0e9`. A real finding and a follow-up, not a failure of the seam.
- **Not chosen** → the interesting negative; step 2 tells you whether it was told and declined.
- ⚠ **NOT a failure of this fix: pages that never build.** `bugs_open/206` is open and cost
  garden-tools **5 of 12 pages**. If it bites here it is that lane's bug in my window — anticipate
  it, do not misattribute it.

---

## 1. What the bug was

A page whose own `<h2>` promised *"What your shed needs, month by month"* shipped four prose blocks
with three incidental month names. The cause was composition: the planner picked components that
cannot render a list, and the writer had nowhere to put twelve months.
`[MEASURED 2026-08-24]` fleet-wide: **327 of 741 pages (44%)** contained no list, table or `<strong>`
anywhere in their content; **1,863 of 1,980 section placements (94%)** used a prose-only template.

## 2. What shipped — eight migrations, two council rounds, all APPROVED and APPLIED

| # | what | council |
|---|---|---|
| 591 | `component_expresses(html_template, input_schema)` + build-site-planner menu + rule 19 | `ca400ba6` |
| 592 | site-planner menu | `ca400ba6` |
| 593 | content-gap-planner menu (**the busiest planner: 749 calls/30d vs 27**) | `ca400ba6` |
| 594 | four prose slots get real `llm_guidance`, retyped `text`→`html` | `ca400ba6` |
| 595 | writer RULE 10 re-addressed to `html`; RULE 9 narrowed, 304's markdown ban preserved | `ca400ba6` |
| 604 | `checklist` component | `c134b0e9` |
| 605 | `period-calendar` component | `c134b0e9` |
| 606 | `comparison-table` component | `c134b0e9` |

**Re-verified live 2026-08-25** (by needle in the live text, never by `updated_at` — see §6):
all eight recorded in `schema_migrations`; all three planner menus carry `component_expresses`;
rule 19 present; rule 10 addresses `html`; 304's markdown ban intact; all four slots read back
`html`; the three components express `{items,list}` / `{items,list}` / `{items,table}`.

## 3. ⚠ THE EVIDENCE — the writer arm WORKS, and the page-level number does not show it

**Measured at the writer's own output** `[MEASURED 2026-08-25, llm_call_log]` — the durable corpus,
where the change actually applies:

| | calls | with `<ul>/<ol>` | with `<h3>` | with `<strong>` |
|---|---|---|---|---|
| `generic-text-block` slot, **after** (prompt carries my new guidance) | **29** | **21 (72%)** | **29 (100%)** | **18 (62%)** |
| `generic-text-block`, **baseline before** | 173 | 17 (**10%**) | — | — |

**10% → 72% on lists. 100% now carry a subhead.** 268 of 396 writer calls since the apply carry the
new RULE 10, so the prompt is live and in use, not merely present.

⚠ **THE PAGE-LEVEL MEASURE READS FLAT AND IS THE WRONG INSTRUMENT — do not quote it as failure.**
`generic-text-block` page rows since the apply: 48, only 5 with a list. That is **not** the writer's
rate. Two dilutions, and a confound in the obvious probe:
1. **Most of those rows are RE-RENDERS**, regenerating from `content_data` authored under the OLD
   prompt. A rerender never calls the writer (this lane's own memory: *"a repro is destroyed by the
   render"*).
2. **Most writer output never reaches a page in the window** — sampled 12 list-bearing responses:
   **3 reached a page and ALL 3 pages carry the list**; 9 were not findable.
3. ⚠ **The correlation probe is itself confounded**: `rewrite_negations` (the copy gate) can REWRITE
   a sentence between the writer's response and the persisted row, so a phrase match can fail on
   content that did land. **A "not found" from that probe is not evidence of loss.**

**Honest conclusion: where the content lands, the structure lands (3/3). The writer's behaviour
changed decisively (10%→72%). The end-to-end page rate is unmeasured, not disproven.**

## 4. ⚠ THE THREE NEW COMPONENTS ARE LIVE AND HAVE NEVER BEEN USED — **this is the open item**

`[MEASURED 2026-08-25]` **0 placements** of `checklist`, `period-calendar` or `comparison-table`.
Not because anything is broken — all three appear in the **live `build-site-planner` menu for
`garden-tools.uk`** (verified 3 of 3) — but because **`build-site-planner` has run ZERO times since
the menus changed.** Only `content-gap-planner` has run (12 calls), and it fills gaps in existing
pages rather than composing new ones.

**So the planner arm and the components are unexercised. That is the whole of what is left.**

## 5. WHAT THE OWNER NEEDS TO BUILD — one greenfield domain

A **fresh site build** is the only thing that exercises `build-site-planner`, and therefore the only
thing that puts the new components in front of it. Per CLAUDE.md, **every site goes through the
framework — never hand-build one**:

```bash
./scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh <domain> \
    --email <contact@domain> \
    --mission-file <path to a mission brief>
```
That entry point drives the whole cascade: `needs_domain_research` → classifier → strategist →
briefing → **`build-site-planner`** (the step that now sees `[expresses: …]` and rule 19) →
composition → design → `needs_content_page` × N → **`page-content-writer`** (the new RULE 10) →
rerender.

**Choose a domain whose content is genuinely structured**, or the test proves little — the point is
a page that *promises* something enumerable. A buying-guide, a how-to, a seasonal or
process-heavy subject will exercise all three components; a two-page brochure will exercise none.

**What to read afterwards, in this order** (recipes in `RUNBOOK_inexpressive_composition.md` §8–9):
1. Did the planner *choose* one? `SELECT ... FROM page_components pc JOIN content_components cc ...
   WHERE cc.function IN ('checklist','period-calendar','comparison-table')` — **any non-zero is the
   headline result.**
2. Did the rendered `plan_site` prompt actually show the capability? Read `orchestration_states`
   **within ~24h** — the rows are a rolling window and this build's will be gone by tomorrow.
3. Page structure share, and the loanzy lane's **promise-vs-delivery** check
   (`loanzy_uk_example_site` HANDOFF 24 §3) — it fires on a page that promises what it does not
   deliver, which is the bug's own shape.

⚠ **`garden-tools.uk` is still off-limits** — the owner's standing instruction, and it is
`bugs_open/206`'s closure canary. If the owner now wants it rebuilt, that is a decision to take
*with* the 206 lane, not around it.

## 6. Traps this lane paid for — read before touching any of it

- ⚠ **`agent_definitions.updated_at` is DEGENERATE.** `[MEASURED 2026-08-24]` 199 of 200 live rows
  share ONE microsecond; a bulk write touches the whole table (twice in 5h20m, rewriting the
  previous value), takes no snapshot and leaves no backup row. **Verify a config apply by your own
  NEEDLE in the live text, never by the timestamp.** `LANDMINES.md`.
- ⚠ **A PostgreSQL regex `\b` is BACKSPACE**, not a word boundary. It matches nothing, raises
  nothing, and returns a clean zero. Use `[\s>]`, and **put a positive control in the same query**.
- ⚠ **To date a deployed commit, ask the binary's own record** — `SELECT git_commit FROM
  service_binary_capabilities WHERE service='agent-chassis' AND kind='build' ORDER BY last_seen_at
  DESC LIMIT 1` (currently `635f2d32f`), then `git merge-base --is-ancestor`. **NOT**
  `grep -a <sha> /proc/1/exe` (GitCommit is one string, not an ancestry, and the 40-zeros control
  matches Go's digit table so it can never fail), and **NOT** pod `.status.startTime` (dates the
  roll, not the image).
- ⚠ **`run-migrations.sh` has NO directory or file scope** — `--apply` takes EVERY pending file,
  including other lanes'. Apply by hand with `psql -v ON_ERROR_STOP=1`, then `--record-only`.
- ⚠ **`missingkey=zero` renders an absent map key as `<no value>`** — but **`RenderTemplate` STRIPS
  it** (`component_library.go:1258`) and `missingBareFields` reports the empty fields by name. This
  lane filed a landmine claiming otherwise and **RETRACTED it the same day**; the retraction is kept
  in place deliberately.

## 7. What is left, precisely

| item | state | owner |
|---|---|---|
| Planner sees capability (591–593) | **LIVE**, unexercised — 0 `build-site-planner` runs since | needs a build |
| Writer emits structure (594/595) | **LIVE and EVIDENCED** — 10%→72% at the writer | done |
| Three generic components (604–606) | **LIVE**, 0 placements | needs a build |
| End-to-end page-level proof | **UNMEASURED** (see §3) | needs a build |
| `comparison-table` claims risk | **STATED, NOT SOLVED** — 29 of 48 sites have neither evidence base nor banned-claims; gate exists, deliberately not applied, one line to reverse (606 header) | `bugs_open/380` |
| §7 card-wall complaint (owner's third point) | **NOT ADDRESSED** — needs a page-composition decision, was never filed as a bug | open |
| `period-calendar` ↔ editorial lane's Phase E timeline boundary | **derived from their description, NOT confirmed** — they were asked 2026-08-24, session had restarted, no reply | awaiting them |

## 8. Can the lane be closed?

**No — but it is one build away, and the bug file's own bar is why.** `/bugs_open/`'s closure test is
**fixed AND live**. Both arms are live and the writer arm is evidenced. What is missing is that the
**planner arm and the three components have never once run**, so the central claim — *"a planner that
can see capability will compose a page that keeps its promise"* — is **built and unproven**.

**Close it when:** one greenfield build places at least one of the three components, or demonstrably
declines to and the reason is understood. **Do not close on the writer evidence alone** — that is
half the bug.

**If the owner would rather not commission a build:** the honest alternative is to leave the bug open
with §4 as its status and wait for the next build that happens anyway. That costs nothing and stays
true. What must NOT happen is closing it on the strength of the config being applied — this lane has
spent the whole of two days on the difference between *shipped* and *working*.
