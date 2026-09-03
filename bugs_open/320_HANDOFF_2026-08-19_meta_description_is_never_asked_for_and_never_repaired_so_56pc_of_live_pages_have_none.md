# 320 — `pages.meta_description` is never asked for, never repaired, and silently overwritten: 56% of live pages have none

**Filed** 2026-08-19. **Status: OPEN.** Found while trying to execute `bugs_open/309` §10,
which is blocked on five of these.

> **Bug numbers collide on this tree — resolve this one by SLUG**
> (`meta_description_is_never_asked_for`), and `git log` the FILE PATH, not the number.

---

## 1. What a reader needs first

A **meta description** is the one-sentence summary a page carries in its `<head>`. Google
prints it under the page's title in a search result, and on this estate it is also what
the blog-listing component shows as each card's blurb. It lives in one place: the
`meta_description` column of the `pages` table.

**`[MEASURED 2026-08-19]` 407 of 731 active pages — 55.7% — have none.** 26 of our 27
sites are affected. Three have none on any page at all: `loancalculator.co.uk` (43/43),
`adversecreditmortgage.co.uk` (19/19), `loanzy.uk` (13/13).

> **RE-MEASURED 2026-09-03 — the headline figure above is now badly stale, and the residual
> has changed KIND.** `[MEASURED 2026-09-03]` **37 of 1,201 active pages — 3.1% — have
> none**, across 11 sites. The backfiller has done its job on everything it can see.
>
> **⚠ AND ALL 37 ARE PERMANENTLY OUT OF ITS REACH.** `load_pages_missing_meta` requires
> `page_visible_text_len(p.id) > 200`; **zero** of the 37 clear it. They average **8**
> characters of visible text (max 166). Demand control, so this is not a broken instrument:
> the 1,164 pages that DO have a description average **4,401** characters and **1,137
> (97.7%)** clear the gate.
>
> **So what remains of this bug is a COVERAGE FLOOR, not a writing failure.** The 37 are
> near-empty pages, and arguably should not carry a description until they carry content —
> which makes "is this a bug or the correct refusal?" an owner question rather than a
> backlog. Whoever picks this up: **re-measure before quoting 55.7% anywhere**, and decide
> the floor question before treating the 37 as work.
>
> Found while verifying `bugs_open/338`'s fix after the 2026-09-03 roll — 338 §9 has the
> full table and the reason its own acceptance test can no longer pass.

There are **two independent mechanisms**, and both are still live. Neither is a backlog.

## 2. Mechanism 1 — pages are born empty because the planner is never asked

`build-site-planner`'s prompt gives the model a `Return JSON:` template for each page. It
asks for exactly:

```
name, title, page_type, nav_label, nav_order, in_header, in_footer, sections
```

**There is no description field in it.** `[MEASURED — live `agent_definitions` row,
2026-08-19; not the seed.]`

The code that writes the page then asks the plan for one anyway and accepts the blank:

```go
// platform/orchestration/actions/site_db_actions.go:1131  (func upsertPage)
metaDescription := datahelpers.GetStringField(page, "meta_description", "")
```

So every plan-built page is born with an empty description, and always has been. This is
the dominant mechanism by volume: `content` pages are **81.4%** empty, `landing`
**85.3%**.

⚠ **`default_config::text ILIKE '%meta_description%'` on this agent returns TRUE and
means nothing.** The planner does contain the string — in `load_existing_pages`, a
`query_database` step that SELECTs the column. Read the output schema, not the grep.

## 3. Mechanism 2 — an unguarded upsert blanks descriptions that DO exist

Same function, the conflict clause. Note the asymmetry with the line two above it:

```sql
-- site_db_actions.go:1163-1180
ON CONFLICT (site_id, name) DO UPDATE SET
    ...
    nav_label        = COALESCE(NULLIF(pages.nav_label, ''), EXCLUDED.nav_label),   -- guarded
    ...
    meta_description = EXCLUDED.meta_description,                                    -- NOT guarded
```

`nav_label` protects an existing non-empty value. `meta_description` does not. Combined
with the `""` default above, **any replan or resync pass whose page map omits the key
overwrites a real description with an empty string.**

> **⚠ CORRECTED 2026-08-19 (council round 2, `editquality`): this file said the new clause "matches the `nav_label` clause". IT IS THE MIRROR IMAGE.**
> ```
> nav_label        = COALESCE(NULLIF(pages.nav_label, ''),         EXCLUDED.nav_label)       -- EXISTING wins
> meta_description = COALESCE(NULLIF(EXCLUDED.meta_description,''), pages.meta_description)  -- INCOMING wins unless blank
> ```
> The code is right and the sentence was wrong. The two policies are deliberately
> OPPOSITE: a description is content the plan owns and must be able to REVISE, so a real
> incoming value has to win and only a blank is refused; a nav label, once set, should not
> be churned by every replan. **Do not "make them consistent"** — that would either freeze
> descriptions for ever or hand nav labels back to the planner. What the two clauses share
> is the NULLIF idiom, not the polarity, and "matches nav_label" papered over the one
> difference that carries the meaning.

**This is not merely representable — it has fired.** `[MEASURED 2026-08-19]` from
`site_snapshots.pages_snapshot`, which carries the column:

| domain | page | snapshot | chars then | chars now | page `updated_at` |
|---|---|---|---|---|---|
| robot-hands.com | `index` | 2026-04-10 | **120** | **0** | 2026-08-19 |
| robot-hands.com | `product-detail` | 2026-04-10 | **97** | **0** | 2026-08-17 |
| robot-hands.com | `tool-gripper-payload-calculator-guide` | 2026-04-10 | **169** | **0** | 2026-08-17 |
| robot-hands.com | `tool-gripper-payload-calculator` | 2026-04-10 | **329** | **0** | 2026-08-17 |

All four are `built_from_plan_version IS NOT NULL` — plan-managed, i.e. reachable by this
exact writer. `index` was updated at 02:40Z on the day of filing, so the site is actively
being replanned.

**The denominator, because 4 alone is not a rate:** the snapshot corpus covers **139**
pages across 7 sites, almost all captured 2026-04-10. Of those, **30** had a description
then; **4 lost it**; 10 gained one.

⚠ **Do NOT extrapolate 4/30 to the fleet.** It is a 139-page, 7-site, four-month-old
sample. It establishes **existence**, not scale. Sizing M2 needs snapshots we do not have.

## 4. Why nothing repairs it

- **No UPDATE path exists.** All seven writers of the column are create-or-upsert:
  `upsertPage`, `create_tool_component_action.go:302,352,504`,
  `deploy_tool_action.go:395,428,589`, `create_report_page_action.go:172`,
  `ApplyAdoptionPlanAction:517,549`, `adopt_verbatim.go:452,474`,
  `cmd/webdesignport/import.go:178`. A multiline-aware scan for `UPDATE pages` blocks
  containing the column returns nothing.
- **No agent config carries one either.** `default_config ~* 'update[[:space:]]+pages'`
  over live, non-snapshot agents returns **0 rows** — so the Go census is not missing an
  SQL-in-config writer.
- **No check detects it.** Of the **58** checks under
  `platform/orchestration/actions/discovery_checks/`, **none names the column.**
- **No writer agent can reach it.** `llm_fields` / `llm_field_specs` are **per-section**
  (`plan_sections_action.go:864-897`), scoped to a component's `input_schema`.
  `meta_description` is a **`pages` column**. This is the `bugs_closed/203` shape — *a
  precedent transfers on mechanism, not on wording; ask who owns the field first.*

## 5. The trap that makes this expensive — a fix that reports success and does nothing

`content_rewrite` items **are** filed about missing meta descriptions and routed to
`page-build-handler`. Two are `complete`. **Neither wrote anything:**

| item | target | completed | column today |
|---|---|---|---|
| `ec701bb3-85e7-40e7-bff1-2ce1ae104861` — *"neither page carries a meta description"* | gaswholesalers.com `about` | 2026-08-15 16:30Z | **0 chars** |
| `13522562-2392-4db9-96b5-204ab67cb999` — *"the home page meta description … leads with a self-description of inventory"* | webdesign.co.uk `index` | 2026-08-15 19:15Z | **0 chars** |

The gaswholesalers page's own `updated_at` is **19:46Z that day**, after the item closed:
the handler ran, touched the page, and the column stayed empty. **So filing
`content_rewrite` items is not a workaround** — it is the expensive version of doing
nothing, and it leaves a green record saying the opposite.

## 6. Blast radius

Every affected page is served to search engines with no description. Beyond SEO, the
column is read as **display content** in at least these places, so an empty value
propagates into rendered pages:

- `rebuild_blog_listing_action.go:99` — each blog card's excerpt (**this is what blocks
  `bugs_open/309`**).
- `prepare_link_context_action.go:344`, `maintenance_actions.go:1045`,
  `rerender_single_page_action.go:529`, `rerender_pages_actions.go:273`,
  `html_actions.go:646`, `multipage_actions.go:154`.

`rerender_single_page_action.go:1009-1020` (`reEmptyMetaDescription`) deliberately strips
`<meta name="description" content="">` rather than serving an empty tag — which is correct
behaviour and also **why this is invisible in the served HTML**: there is no empty
attribute to notice, just an absent tag.

## 7. How to verify a fix

```sql
-- must fall, and the denominator must be quoted beside it
SELECT count(*) FILTER (WHERE COALESCE(meta_description,'')='') AS empty,
       count(*) AS active
FROM pages WHERE status='active';

-- M1: no NEW page is born empty (the honest test is by creation date, forward only)
SELECT count(*) FILTER (WHERE COALESCE(meta_description,'')='') AS born_empty, count(*)
FROM pages WHERE status='active' AND created_at > '<fix date>';

-- M2: induce it. Replan a site with a populated description and assert it SURVIVES.
--     A test that only replans an empty page cannot fail.
```

⚠ **`COALESCE(...)=''`, not `IS NULL`** — the column is nullable *and* routinely holds
`''`, because `GetStringField` defaults to the empty string.

⚠ **Verify at the served page with a control**: a page that MUST come out non-empty, and
print the byte count. A zero from a guessed URL is a 404 (`309` §C paid for that once).

## 8. Fix candidates, ordered by what makes the bad state unrepresentable

**None of these should be actioned without an owner decision** — see §9.

1. **Guard the overwrite** (M2). Change one clause to
   `meta_description = COALESCE(NULLIF(EXCLUDED.meta_description,''), pages.meta_description)`,
   borrowing `nav_label`'s NULLIF idiom but with the OPPOSITE polarity (see the
   correction in §3). Makes "a replan silently destroys published
   copy" **unrepresentable**, is ~1 line, and needs no new authority. Does not fill
   anything. **Cheapest and most defensible; it is also the only candidate that stops
   active damage.**
2. **Add the field to the planner's `Return JSON:` template** (M1). Config-only, live
   immediately, no build. Fixes new pages, and — because the upsert overwrites on
   conflict — a subsequent replan would also populate existing plan-managed pages. Note
   that ordering matters: doing (2) before (1) is fine, but doing (2) *without* (1)
   leaves the destruction path open for any pass that omits the key.
3. **A discovery check** `meta_description_missing`. **Ranked below the fixes on purpose.**
   Detection with no handler reproduces the state this estate is already in — 606
   `head_essentials_missing` rows detected and unactionable. A check is only worth adding
   with a handler that can act, which is candidate 4.
4. **A backfill producer** for the 407 existing pages. This is the one that does not
   exist and cannot be improvised: it writes public copy under the owner's name across 26
   sites. New authority on a shared seam ⇒ 2026-08-02 §2 (opt-in, unsafe default OFF) and
   an owner decision.

## 9. What is NOT to be done, and why

- **Do not hand-write the descriptions.** Owner ruling 2026-08-06: the framework writes
  the content. That ruling's own escape clause is the instruction here — *"if the
  generator does not exist yet, that is the finding to report, not a gap to fill
  personally."* It does not exist.
- **Do not file `content_rewrite` items for them.** §5: measured, completes, does nothing.
- **Do not lower `section_shrink_floor`** to push `309`'s page through. It is read from
  step config only (`save_sections_shrink_guard.go:80`), so it is fleet-wide, on the
  highest-volume pipeline there is. (`309` §9 reached the same conclusion independently.)

## 10. Provenance of this file

Diagnosis loop run correlation **`7375631a-dfa1-4145-9a07-d13586f1a7cf`** returned
**`UNVERIFIABLE`** (stopped on the **iteration cap**), **not CONFIRMED** — it is cited
here as a contributor, not as ratification. Its value was refusing the "frozen at
creation" framing, supplying M2, and naming the `site_snapshots` test that settled M2's
occurrence; that test is §3 and I ran it. Everything in §2-§6 is first-hand and carries
its query or its `file:line`.

An earlier run (`d7a9ab39-4917-4479-b1d5-68aae982f79c`) **failed** with
`diagnose_assemble_bundle: no scope` because the symptom named files rather than symbols
and no `SEED_SCOPE` was passed. ⚠ Its terminal orchestration row still reads
`current_step=complete, status=COMPLETED` — the honest check is
`count(*) FROM diagnosis_artifacts`, and the error text sits in
`collected_data->>'__step_error'` on the **`complete`** row, not on the FAILED ones.

Working docs, including the full misstep log:
`docs/agent_docs/docs024_key_docs_latest/meta_description_never_backfilled/`.

---

## 11. OWNER RULING 2026-08-19 (given in the bugfix-309 session, recorded verbatim in intent)

The owner was walked through §8's candidates as four decisions and ruled on each:

1. **Guard the overwrite (M2): YES.** ("Stop the ongoing damage.")
2. **Ask the planner for descriptions (M1): YES.**
3. **The backfill producer: YES — GO AHEAD, and the owner does NOT require a
   review pass of the output.** ("I don't need a review.") This supersedes the
   one-site-at-a-time-with-someone-reading caution in this lane's README §"I have
   also not put it on a schedule" — the mechanism's kill-switch/opt-in SHAPE should
   stay (2026-08-02 §2 is about how authority ships, and it survives the owner
   waiving the review), but the rollout is authorised fleet-wide without a
   read-first gate.
4. **The 309 page: WAIT for the writer/replan** — no regeneration of the five
   articles, no shrink-floor change. (Consistent with what this lane already
   concluded; now ruled, not just reasoned.)

**Plus one NEW requirement, given in the same exchange (2026-08-19, mid-morning):**

> "please make sure the summaries go through the copy guidance and checks so they
> don't sound like AI"

So the backfiller's generation is NOT bare "write a meta description": it must
carry the house copy-style guidance
(`docs024_key_docs_latest/travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`
— the plain-human copy preference, owner 2026-08-06 family), and its output must
pass the voice/claims controls before being SAVED, not after being published —
the define-by-negation suppression (`bugs_open/305`'s subject), the banned-claims
sweep, and whatever `check_voice_tells` asserts. A description that fails reads
as AI boilerplate on 407 pages under the owner's name, which is the exact class
of damage the review pass would have caught — the checks are what replace it.

Recorded by the bugfix-309 session; the live 320 lane (session "bugfix 284") was
messaged the same content at the time of this commit. Division: the 320 lane owns
execution; this section is the durable record of the authority it acts under.


---

## 12. LIVE AS OF 2026-08-19 EVENING — what is fixed, what it measured, what is still open

All four commits shipped in chassis **`v1.0.1315`** (pod `imageID` == local
`RepoDigests`, revision `590ca3a20cca…`, all four ancestors, `HEAD` correctly not an
ancestor as a control, binary probe with both controls).

| piece | state |
|---|---|
| M2 guard, all **four** write paths | **LIVE** |
| M1 — planner asked (mig `485`) | **LIVE** since 08-19, config |
| `save_page_meta_description` (SEO-004) + copy gates | **LIVE** |
| `meta-description-backfiller` (mig `488`, fixed by `493`) | **LIVE and PROVEN** |

### Measured after the first real runs

| | before | after |
|---|---|---|
| fleet active pages with no description | 407 / 731 (55.7%) | **381 / 736 (51.8%)** |
| loanzy.uk | 13 of 13 empty | **11 filled**; the 2 left have ZERO components |
| fundamentallyai.com | 20 of 25 empty | **20 of 25 now HAVE one** |
| the five pages blocking `bugs_open/309` | all empty | **all five filled** |

Idempotence proven at row level: a repeat run touched only the still-blank page and left
the others' `updated_at` untouched, so `overwrite_existing=false` holds in production.

### The first canary run found two defects that a status check would have missed

It reported `COMPLETED` with no error and wrote **nothing**.

1. **`output_format: "array"` returns a bare array, so `.count` never resolved** and the
   gate silently routed past the only LLM step — with 11 rows sitting in the array. This
   is **`bugs_open/313`**, and it arrived here by copying `internal-linker`, the agent
   `313` was filed against. Fixed by `output_format: "object"` (mig `493`).
2. **`content_sample` was raw markup**, so the writer would have described tool pages from
   their CSS. Now visible text, floor 400→200 chosen from the distribution.

### ⚠ STILL OPEN, and the first one is load-bearing for `309`

1. **Descriptions come in SHORT: mean 102 chars (65-177) against a prompt asking 120-155.**
   `309` §9's shrink-guard arithmetic assumed ~157 each and projected the rebuilt slot at
   ~1,818 chars against a 1,239 floor. At ~102 that projection is materially lower.
   **Whether the guard now passes must be MEASURED by dispatching the rerender — do not
   read "the blocker is cleared" as "309 will now render".**
2. **355 pages still empty fleet-wide.** Only two sites have been run. The rest are a
   `./scripts/backfill-meta-descriptions.sh <domain>` each.
3. **Pages with zero components can never be filled from their own content** — 43 of the
   original 407. They need content before they need a description.
4. **`bugs_open/313`'s wider sweep is untouched**: ≥8 other live conditions use `X.count`
   and whether each is broken depends on its own step's `output_format`. Noted there, not
   silently widened from here.


---

## 13. SCHEDULED 2026-08-20, and what the first scheduled run taught

**Owner instruction 2026-08-20: "we can put the backfiller on a separate schedule."** Done
as migration `498` — hourly, its **own** concurrency group at `max_concurrent=1`.

It is safe to leave enabled for ever: a `pre_query` returning no rows is a clean no-op
that stamps the timestamps (`cmd/scheduler/main.go:205-220`), so when every fillable page
has a description the task costs one cheap SELECT an hour, and it wakes by itself when
new pages appear. **It does not need switching off when the backlog drains.**

### The first scheduled run proved the owner's condition is not decorative

Fired 06:52:34Z, targeted `leopardessconsulting.co.uk`. The writer produced:

> "Read evidence-based articles on AI adoption, **trust gaps**, and governance across
> healthcare, finance, hiring and more."

and `save_page_meta_description` **refused it**, quoting the site's own rule:

```
reason: voice_tell
detail: banned_phrase: owner 2026-07-18: overused;
        say what is checked/verified instead ("trust")
```

`updated: false`, with a named reason. **The gates the owner made a condition of waiving
the review pass are real, and they caught a violation of one of his own rulings.**

⚠ **And I only know that because I checked the artefact.** `last_triggered_at` was set,
the orchestration read `COMPLETED`, the scheduled task looked healthy — and the page was
still blank. A fired schedule is not a completed job.

### Three fixes came out of it

- **`499`** — a repeatedly-refused page on an hourly schedule is a permanent hourly LLM
  bill that never fills, and nothing about it looks like a failure. So the writer is now
  told the site's banned phrases **before** it writes. **9 sites** carry an enabled gate
  (14 phrases on leopardess, 10 on oufe, 1 each on seven more); on the other ~18 it is a
  no-op.
- **`500`** — ⚠ **`499` shipped a defect and broke the live agent.** `write_descriptions`
  went `FAILED`, because **`query_database` stringifies a jsonb column**: `jsonb_agg(...)
  AS rules` reached the template as a JSON *string*, and `{{range}}` over a string is not
  iteration. Diagnosed at the run's own `collected_data` (`rules_type: string`), not
  guessed. Fixed by changing the SHAPE — one row per rule, iterating
  `output_format: "object"`'s own `rows`. Landmine written.
- **`501`** — asks for ≤20 words. Better copy on its own merits, **and** a workaround for
  a defect stated plainly rather than dressed up (below).

**Proven end to end afterwards on the site that had been refused:** 14 rules loaded, the
writer avoided the banned phrase (*"what builds and breaks confidence"* rather than
*"trust gaps"*), and the page was written —
*"Read research-backed articles on AI adoption, governance, and risk across healthcare,
finance, and hiring."* 106 chars, 16 words.

### ⚠ OPEN — a Go fix this lane did NOT make → **now filed as `bugs_open/338`**

`save_page_meta_description` applies the **whole** voice gate to a single sentence. The
gate carries two kinds of rule: **content** (banned phrases — correct anywhere) and
**density/distribution** (mean sentence length, long-sentence share, em-dash per 1000
words, contraction expectation — **statistics over a corpus**). Over one sentence the
second kind degenerates, and it refused a perfectly good 24-word description against a
default trip of 22.

`[MEASURED 2026-08-20]` it bites on only **2 of 27** sites — because **7 of the 9** gated
sites already set `mean_sentence_words: 100000` by hand to switch the length checks off
while keeping the phrase list. *The estate has already voted on this with its config,
one site at a time, without anyone writing it down.*

**The fix is in Go and needs a roll:** in `metaDescriptionFailsCopyGates`, apply the
content findings (`check == "banned_phrase"`) to a single-value field and skip the
distribution ones. `501` mitigates it in the prompt meanwhile. **Do not fix it by raising
a site's thresholds** — that relaxes the rule for the site's pages too, where it works.

### ⚠ A SEPARATE FINDING, not this lane's to fix — **now filed as `bugs_open/339`**

`[MEASURED 2026-08-20]` **11 live pages carry a description of 200-320 characters, and
9 of them are `tool` pages whose text is plainly a BUILD BRIEF**, e.g.
*"Lets designers define a total stat budget for a character or item tier, then…"*,
*"Companion to the Spawn Rate Balancer. Designers input player power growth per wave…"*,
*"Converts between px, rem, em, vw, and vh units given a base font size…"*.

That is `bugs_closed/103`'s exact failure — tool pages publishing their build brief as the
public meta description. **Zero live descriptions exceed 320 characters**, so `103`'s
length guard catches nothing today: the population moved *under* it. Its own
`briefMarkers` regex was meant to be the second signal for exactly this
("length alone would miss a SHORT brief") and does not match these.

Not written by this lane's backfiller — every one of these predates it or comes from the
tool-deploy path. **Flagged, measured, and left for whoever owns `103`.**


---

## 14. 2026-08-21 — a defect in MY OWN floor, and the question it raises about descriptions already written

Two tidy-ups were asked for and both are done (`517`, `518`). The second exposed something
bigger than the tidy-up.

### `517` — the pre-query and the workflow asked different questions

The scheduled task asked *"has this page any rendered component?"*; the workflow asked
*"has it >200 chars of visible text?"*. So the scheduler dispatched an orchestration every
hour that always concluded `complete_nothing_to_do`. Cheap — it stops before the LLM step —
and the third instance in this lane of **a green record over work that never happens**.

Fixed as **one definition** (`page_visible_text_len()`, a `STABLE` SQL function both call)
rather than a second copy of the regex, per `bugs_closed/284`'s rule about renderings of a
shared predicate.

### `518` — ⚠ MY MIGRATION `493` WAS MEASURING THE WRONG THING

I reported `noted.co.uk/index` yesterday as *"197 visible characters, a homepage three
characters under the floor"*, and offered it as a judgement call about the floor.

**The page has 1,205 characters of real text.** `493` stripped `<style>`/`<script>` blocks
**after** `string_agg`, so the match ran across component boundaries and consumed the next
component's prose. Balanced tags — 3 opens, 3 closes — so nothing looked broken.

`[MEASURED 2026-08-21]` across 693 active pages: **349 lose more than half** their visible
text to that formulation, **24** are wrongly judged below the floor, **1** of those was
blank and being declined as too thin. Fixed by stripping per component then joining, with
`ORDER BY` for determinism. That homepage now measures 1,205 and has been backfilled.

### ⚠ THE OPEN QUESTION THIS RAISES — needs an owner decision, not a session's

`content_sample` — the 1,200 characters of page text handed to the **writer** — had the
same flaw. So descriptions written before `518` came from a **possibly-truncated view of
their page**, and the model produced fluent copy from the fragment. **Nothing downstream
could tell**, which is why it went unnoticed.

`[MEASURED 2026-08-21, and the figure MOVES between runs because pages are rerendering
continuously — treat it as an order of magnitude]`: of ~692 pages with a description,
**roughly 270-350 had a degraded sample** when it was written, and **~20-44 severely** (the
old measure kept under 200 chars of a page with more than 200). Around 35 of those carry an
em dash, which this lane's prompt bans, so they were written by another path — the rest are
plausibly the backfiller's.

**Spot-checked three severe cases against what the page actually says. Two are wrong:**

| page | written description | what the page says |
|---|---|---|
| `robot-hands.com/gripper-detail` | *"Full specifications for this gripper, with gripping force, payload, stroke and IP rating…"* | *"Start With the Right Parameters. End With a Shorter Shortlist. Filter the catalog by payload, stroke, actuation type…"* — it is a filterable catalog, not a single-product spec sheet |
| `webdesign.co.uk/news` | *"What is changing in web design — browsers, CSS, accessibility rules and AI tooling, with a UK slant."* | *"New tools and guides, added as they're built. This page lists what has changed across the site."* — a site changelog, not industry news |
| `finetuning.uk/tool-ai-data-risk-checker-guide` | the `composedGuideMetaDescription` template — generic but not inaccurate | (not a backfiller product; em dash gives it away) |

They are plausible *from the page title alone*, which is exactly what the writer had when
the sample was empty.

**Why this is an ASK and not an action.** Regenerating them requires
`overwrite_existing: true` — the unsafe authority that defaults OFF by the owner ruling of
2026-08-02 §2, and the owner's 2026-08-20 authorisation was to **fill blanks** fleet-wide,
not to replace existing copy. So the options are the owner's:

1. **Leave them.** They read well; some are inaccurate. Cheapest, and the inaccuracy is
   invisible to anyone not comparing description to page.
2. **Regenerate only the severe ones** (~20-44). Smallest change that removes the
   demonstrated errors, needs a one-run `overwrite_existing: true` on a scoped set.
3. **Regenerate all degraded** (~270-350). Most thorough, most LLM spend, and it replaces
   copy that is mostly fine.

⚠ Whichever is chosen, **re-run it over what the blind measure already cleared** rather
than only forward — the same rule this lane applied to `339`'s guard.


---

## 15. 2026-08-21 — OWNER RULING: "redo the descriptions through the framework" — DONE

**681 of 704 live descriptions regenerated.** The 23 holdouts are gate refusals, which is
the gate working; they keep their previous copy, so nothing regressed.

### How the authority was handled

Regenerating means REPLACING published copy — `overwrite_existing: true`, the unsafe side
of the opt-in field. It was granted for this act, **not for the standing mechanism**, so:

- the flag was set **INLINE, on a one-off dispatch** (`scripts/regen-meta-descriptions.sh`);
- **the seeded agent was never armed.** Verified after the run:
  `default_config#>'{…save_description,config}' ? 'overwrite_existing'` → **false**. The
  hourly scheduled task remains fill-blanks-only and cannot overwrite anything.

**Fully reversible.** Every pre-regeneration description is in
`meta_description_pre_regen_20260821` (704 rows, count-matched to live before starting):
```sql
UPDATE pages p SET meta_description = b.meta_description
FROM meta_description_pre_regen_20260821 b WHERE b.page_id = p.id;
```

### ⚠ CORRECTION TO §14 — my evidence was weaker than I reported

§14 said *"spot-checked three severe cases; two are wrong"*. **It is one of three.**

`robot-hands.com/gripper-detail` I called wrong twice, on the grounds that its description
said *"full specifications for this gripper"* while the page read *"Filter the catalog
by payload, stroke, actuation type…"*. **I was judging the page from 170 characters of its
middle.** Its actual `h1` is **"Gripper Specifications"**, and its headings run
*"Specification data you can measure"*, *"Every Tool You Need to Specify the Right
Gripper"*. The filtering line is one section inside a specifications page. **Both the old
and the regenerated description are accurate, and my criticism was wrong.**

`webdesign.co.uk/news` **is** genuinely wrong and is now fixed:
- was: *"What is changing in web design — browsers, CSS, accessibility rules and AI tooling, with a UK slant."*
- now: *"Stay updated on new tools, guides, and web standards as they're built and published."*
- the page: `h1` *"News from webdesign.co.uk"*, then *"New tools and guides, added as they're built"* — a site changelog, not industry commentary.

⚠ And I nearly got that one wrong too: my first check of it fetched `/news.html` and got a
**404**, which I would have read as a finding. The real URL is `/news/index.html`. **Third
time this lane has been caught by a guessed URL** — take the URL from `pages.url`.

### The misstep that cost real LLM spend

The first regeneration query was `… WHERE meta_description <> '' … ORDER BY p.name LIMIT 25`.
Over a set that never shrinks, **every run re-picked the same first 25 pages alphabetically
and rewrote them again.** Progress sat at 343 across four dispatch rounds while each run
reported `selected: 25, written: 24`. The runs were working; the *selection* never advanced.

Fixed by joining the backup and selecting only rows that still hold their pre-regeneration
text — which makes the query self-limiting and the whole operation resumable. Proven rather
than assumed: webdesign.co.uk went 105 → 80 still-to-do in one run, exactly 25.

**The tell I should have read sooner:** a counter that does not move while every individual
run reports success. I fired four more rounds before diagnosing.

### The abandoned approach, and why

The first plan targeted only descriptions written while the sample was degraded, using the
OLD concatenate-then-strip measure to identify them. **That predicate is not stable** — it
is the same order-sensitive expression `518` fixed, and it classified `gripper-detail` and
`product-detail` in opposite directions depending on which query evaluated it. A target set
that changes between evaluations is not a target set, so the scope became *every* live
description, which is also the plain reading of the instruction.

### Result

| | |
|---|---|
| regenerated | **681** |
| refused by the copy gates (`voice_tell`), previous copy retained | **23** |
| mean length | 130 → **117** chars (the ≤20-word rule from `501`) |
| seeded agent | **still fill-blanks-only** |

---

## §16 — CONTRIB 2026-09-02 (`routing_capability_guard` lane, closing): **§15's WITHHOLDING IS REVERSED BY THE OWNER**

Pointer, not a restatement. **§15 records the owner granting `overwrite_existing: true` for the
one-off 681-page pass and explicitly withholding it for the standing mechanism.** On 2026-08-26 he
ruled the other way: **an automated finding MAY cause a published description to be rewritten**,
restricted to machine-written ones — *"I haven't yet written any manually"* — and on 2026-09-02
confirmed it should be built.

**The full handover, with the two measurements that change how big this job is, is in this lane's own
directory:**
`docs/agent_docs/docs024_key_docs_latest/meta_description_never_backfilled/CONTRIB_2026-09-02_decision_1_is_ruled_and_re_homed_to_you.md`

The headline of it, because it is the part that resizes the work: **the overwrite authority already
exists.** `save_page_meta_description_action.go:211` is gated by `overwrite_existing` (an opt-in config
field, **default false**, enforced in the WHERE clause) AND by the backfiller's `pre_query` — **two
guards in series, not an unconditional UPDATE.** So what is missing is a work-item-driven ROUTE that
sets the flag, not an overwrite capability. ⚠ Several documents — including a shipped code comment and
`bugs_open/395`'s roster, both now corrected — described that write as *"the only unconditional
UPDATE"*, which oversizes this job.

**The originating lane is CLOSING; do not route questions at it.** Everything is in `bugs_open/395`
and in that lane's handoff (§2, §12, §19, §21), both of which outlive the session.
