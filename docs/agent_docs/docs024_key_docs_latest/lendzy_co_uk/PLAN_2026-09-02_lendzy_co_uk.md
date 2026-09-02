# PLAN — lendzy.co.uk — 2026-09-02

Design, phasing, decisions **and their reasons**. Corrections live here, marked, never edited away.
Site: `lendzy.co.uk` = `8ff093d5-1f19-453b-9439-a10379bbcd76`.

---

## 0. Why this lane exists

The owner made lendzy its own lane today. Three instructions, in his order:

1. **Keep the tools if they are working, and add them where necessary.**
2. **Chase the root cause** (three tool pages serve 200 while recorded never-deployed).
3. **Build the 47 internal links.**
4. **Make the retracted claim true.** *"There was a statement on the site that said all facts were
   checked against the FCA handbook line by line. I'd like that to be true please... check all
   financial facts against the FCA handbook line by line. If this means we keep a copy locally in a
   separate database and update it regularly with the FCA changes then let's do that. Or if it means
   we can check with their online version each time then we can do that as well or instead
   (probably as well)."*

(4) is the large one and is the lane's real subject. (1)–(3) turn out to be **one defect**, not
three — see Phase A.

---

## PHASE A — the three tool pages (1, 2 and 3 are the same bug)

### A1. What is actually wrong

Root cause established today and recorded with its evidence in `NOTES` (c). In one sentence:
**each of the three pages has a single `page_components` row whose `component_id` is NULL and whose
`slot_name` is `section`, which resolves to no component by either route, so the re-render fails
fatally, the page never reaches `build_status='deployed'`, and therefore never earns a
`deployed_at` stamp — while the artefact deployed on 2026-08-02 keeps serving 200 for ever.**

`[MEASURED 2026-09-02]` Fleet-wide, the active pages on which **no** component row carries a
`component_id` are exactly these three, and all three are `needs_rebuild` and unstamped. Nothing
else in the estate is in this state.

Filed to the diagnosis loop before being asserted durably (owner ruling 2026-07-31): intake
`1ff4c475-6977-4631-b641-993735429186`, run `89a84ad3-5668-44b3-a089-f9d6c0df7cbb`.

### A2. Why (1), (2) and (3) collapse into it

- The **tools work** — all three serve, with the same `<input>` counts in the served page as in the
  stored HTML. So the owner's "keep them" is satisfied by *not regenerating them*.
- The **47 links** are not broken links. Every one of them points at a URL that serves 200. They are
  filed because `unbuilt_internal_link` asks the **record**, and the record says never-deployed.
  Fix the record and all 47 resolve without touching a single link.
- The **sitemap** gains the same three URLs for the same reason (`render_sitemap_action.go:144`
  filters `deployed_at IS NOT NULL`).

> **DECISION: repair by ADOPTION, not regeneration.** `create_tool_component` builds a component
> from *LLM-generated* HTML. Running it here would replace three working calculators with three new
> ones — the opposite of the instruction. Instead the existing stored HTML is promoted to a
> component's `html_template`. This is **lossless**: `[MEASURED 2026-09-02]` all three stored bodies
> contain **zero `{{` template bindings**, so there is nothing to bind and the template *is* the
> rendered form. The reason to record: a conversion that would have been lossy on a data-bound
> component is safe on a self-contained JS calculator, and that distinction is the licence.

### A3. The change

A migration creating one `content_components` row per tool, matching the shape of the six healthy
siblings measured today (`component_level='tool'`, `section_type` NULL, `forked_from` NULL,
`is_active`, `render_mode='template'`, `category='interactive'`, name `<function>-lendzy-co-uk`
per CLC-020), then repointing each `page_components` row's `component_id` and `slot_name`.

`created_from='adopted'` rather than `'generated'` — the siblings say `generated` and would be the
easy thing to copy, but these rows are adopted from live HTML and the column exists to say so.

Guards the migration must carry, because each is a way the apply could be silently wrong:
- **Assert no library tool claims the function** before inserting with `forked_from IS NULL`
  (`idx_cc_tool_function_unique`; RFC_036 §9.3 says such a row must be born a FORK). True today,
  and another session could make it false between now and apply — so assert, do not assume.
- **Assert each target `page_components` row still has `component_id IS NULL`**, so the migration
  cannot overwrite a repair someone else landed first.
- Verification by `DO`/`RAISE`, never a bare `SELECT` — a verify block of `SELECT`s cannot stop the
  `COMMIT` (`ON_ERROR_STOP` ignores a non-empty result). That trap is already in the estate's record.

**Acceptance is at the artefact and the record, not at the migration:** the three pages reach
`build_status='deployed'` with a `deployed_at`; the sitemap serves **30** `<loc>`; the 47
`unbuilt_internal_link` items resolve; and — the control that matters — the three tool URLs still
serve 200 with **the same `<input>` counts as today (3 / 1 / 2)**, proving the repair did not
silently swap the working calculators for something else.

Council gate: migrations are in scope (widened 2026-08-19). Submit before or alongside the commit.

---

## PHASE B — "checked against the FCA handbook, rule by rule", made true

### B0. The standing correspondence

The owner named the **claims-verification** thread as responsible here, and they replied today with
the boundary already drawn. Their answers are the design's foundation and are summarised in `NOTES`
(e) with attribution. Design docs route through that thread or cite `RFC_060`, by their request, so
the estate does not end up with two accounts of one register mechanism.

### B1. What already exists, and must not be rebuilt

- **`evidence_base` + citation facts** — one fact = one URL + one verbatim quote. This *is* the
  "line by line" primitive; a rule citation is exactly its shape.
- **The daily evidence-refresher** (`refresh_evidence_base_action.go` + `evidence_citations.go`,
  CLM-007/008) already re-fetches each citation and re-checks the quote against visible text,
  classifying 403/5xx as *unknown* and a 200-with-quote-gone as `citation_lost`. **That is the
  owner's "check with their online version each time", already built and already daily.**
- `fad209b92` already excludes regulatory citations (CONC, MCOB…) from the business-number scan, so
  lendzy's `0.8% per day under CONC 5A` shape needs nothing new.

### B2. What is genuinely missing (the build)

1. **A local FCA Handbook mirror**, on the Companies House pattern the claims thread pointed at
   (`017_companies_house_enrichment.md`): its own table, its own polite paginated collector,
   deliberately rate-limited well under the published cap, refreshed on a schedule, queried locally
   thereafter. **Not inside `evidence_base`** — that table is built for "this fact cites this quote",
   not for holding a handbook.
2. **Change detection** — which rules exist, and which changed since we cited them. Nothing does this
   today: the refresher only re-checks quotes a human already chose to cite, so a rule that is
   amended in a section we never cited is invisible.
3. **Pacing.** `fetchCitationDocument` is one unthrottled request per fact with no delay and no dedup
   across facts sharing a URL. Harmless at the fleet's current citation count; not harmless when one
   site cites dozens of handbook rules daily. **Pacing must land before the fact count rises**, not
   after something gets blocked.

### B3. ⚠ The trap that shapes the whole design

> **`handbook.fca.org.uk` returns HTTP 200 for every path, including rules that do not exist.**
> `[MEASURED 2026-09-02]` A real section returns a specific title —
> `FCA Handbook - CONC 5A Cost cap for high-cost short-term credit` (477,729 B) — while an invented
> rule path returns 200 with **178,705 B** and the bare title `FCA Handbook`, and a nonsense path
> returns the home shell (165,639 B, `FCA Handbook - Home`). It is an Angular app with a catch-all
> route.
>
> **So HTTP status cannot test whether a rule exists**, and a collector that trusts 200 will mirror
> non-existent rules as though real — which is precisely the failure mode this whole phase exists to
> prevent. The discriminator measured today is the `<title>`. Every collector and every verifier
> gets an **invented-rule control in the same run**, and a landmine entry says so.
>
> The good news, and it is not luck: the existing refresher already degrades safely here, because a
> withdrawn or renumbered rule returns 200 with the quote gone, which it classifies as
> `citation_lost` — drift, correctly. That wants proving, not assuming.

### B4. Phasing (content first, because it needs no code)

`RFC_060` §4 is explicit that populating the register is the highest-value thing, is content work,
and needs no platform change. Lendzy has **no `evidence_base` at all** — it is one of the five
register-less finance sites. So:

- **B-i. Register lendzy's financial facts as FCA citations.** Every figure the site asserts gets a
  fact with the rule URL and the verbatim quote. No code. This alone makes the daily refresher start
  checking lendzy against the live handbook — i.e. it delivers the owner's second option outright.
- **B-ii. The mirror + change detection**, on the CH pattern, with pacing and the invented-rule
  control. Architecture-scope (a new shared mechanism, a new external dependency): RFC first, or an
  extension of RFC_060, routed through the claims thread. Register on ship.
- **B-iii. Only once B-i and B-ii both hold** may any page say so in prose, and then it says what is
  actually true and dated — not a marker. The bug that started all this was a *claim without a
  mechanism*; the fix is a mechanism, and the claim comes last. **Nothing in this lane re-plants a
  compliance sentence on lendzy before B-i and B-ii are live and verified.**

Open, and the owner's call when we get there: whether the site says anything about this at all.
`bugs_closed/414` is the standing evidence that a compliance sentence is a liability the moment it
outruns its mechanism.
