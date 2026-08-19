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
   matching `nav_label` directly above it. Makes "a replan silently destroys published
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
