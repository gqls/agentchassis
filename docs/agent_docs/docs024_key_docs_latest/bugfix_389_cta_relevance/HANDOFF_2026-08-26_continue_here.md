# HANDOFF — 2026-08-26. **START HERE.** `bugs_open/391` — step 2 DONE and proven; retirement is half-applied and safe; **one decision blocks the rest**

**Supersedes `HANDOFF_2026-08-25c_continue_here.md`** (keep it for the reasoning trail). Read this,
then `bugs_open/391` from the bottom.

> **Deploy facts have a shelf life of hours — re-probe, do not quote.** Chassis at handoff:
> **`v1.0.1345`** (`kubectl -n ai-persona-system get deploy agent-chassis -o jsonpath=…`, matching the
> `uk_001` overlay `newTag` at HEAD). GitHub Actions: **All Systems Operational** as of 20:35Z, after
> a `major_outage` this afternoon.
> ⚠ **Do NOT verify the chassis sha with a binary grep** — the all-zeros control MATCHES on this
> image (`grep -aq "000…0" /proc/1/exe` succeeds), so both the positive and the negative answer are
> worthless. Read the image tag, or the `build provenance` startup log line while it is still in range.

---

## 0. THE BUG IN ONE PARAGRAPH (unchanged)

`chooseCTATargets` (`resolve_internal_links_action.go:651`) picks a site's primary CTA by sorting
every `tool`/`game` page on `COALESCE(nav_order,100)` then `name` and taking `[0]` — **no topic, tag
or semantic input at all.** A password-strength toy carried the fossil value `nav_order = 1` on three
sites and so won the primary button on every page. The framework then writes button copy **naming**
whatever it picked, so a wrong pick **locks itself in**.

## 1. STATE — verified 2026-08-26 ~20:3xZ

| # | work | state |
|---|---|---|
| 1 | `nav_order` demoted 1 → 900 on three sites | ✅ done, holding |
| 2 | the 12 label-locked pages | ✅ **COMPLETE** — the lock query returns **0 rows fleet-wide** (was 20 of 80) |
| 3a | **archive** the three tool pages | ⏳ **2 of 3** — finetuning.uk ✅, leopardess ✅, **ai-agent-orchestration.com NOT YET** |
| 3b | **re-resolve** the inbound references | ⏳ **2 canaries proven**, 41 fields remain — **BLOCKED ON THE DECISION IN §2** |
| 3c | **retract** the deployments | ⛔ not started; refuses until 3b clears the inbound links |
| 4 | the platform lever (owner decision 3) | not started |

**All three tool pages still return 200.** Archiving freezes a page and keeps serving it, so the
half-applied state is coherent — nothing is dead, nothing is half-broken, and every step so far is
reversible by setting `status` back to `'active'`.

## 2. ⛔ THE DECISION THAT BLOCKS STEP 3b — read this before dispatching anything

**Clearing every `password-entropy` reference would satisfy this lane's stated success criterion
while leaving the pages wrong.**

The re-resolution sends a field to whichever tool ranks first. That is right when the button's copy
is about a tool. It is **wrong when the copy asks the reader to get in touch** — you get a button
reading *"Write to leopardess@contactforsales.com"* that opens an ROI estimator. This is
`bugs_open/248`'s shape and the KEEP #1 comment in `rerender_page_sections_action.go` describes the
mechanism exactly.

**It is not a corner case.** `[MEASURED 2026-08-26 20:3xZ]` of the **41** fields still pointing at
the tool on active pages, **23 carry contact-intent labels** (`get in touch|contact|write to|email|
call us|book a|talk to|speak to|discovery call`) and 3 have no label at all. **A majority comes out
wrong if step 3b runs as originally planned.**

**Both canaries are the evidence, and they disagree in exactly the informative way:**

| page | labels | result | verdict |
|---|---|---|---|
| `finetuning.uk/about.html` | tool-intent (*"Check Your AI Data Risk"*) | → `tool-ai-data-risk-checker` | ✅ correct in kind |
| `leopardess/careers.html` | contact-intent (*"Write to leopardess@…"*) | → `tool-ai-agent-roi-estimator` | ❌ wrong in kind |

**⇒ THE ASK:** add a second clause to the success criterion — **every rewritten CTA's label and
destination must agree in KIND** — and write it as a gate against `content_data` that runs *before*
dispatch, not as a per-page discovery afterwards. The 23 contact-intent fields need a contact
destination (`/contact.html`) or a copy rewrite, not the positional pick. **That design decision is
the next piece of work and it is the owner's call how the 23 are handled.**

## 3. THE SEQUENCE — three steps, not two, and PROVEN end-to-end

The old two-step order (*retire → then re-resolve*) **deadlocks**: the retraction refuses while
anything editorial links in, and the re-resolution is blocked by KEEP #2 while the destination is
still valid. Each is the other's precondition.

**It breaks on `validPages`.** `loadResolverPageSet` (`resolve_internal_links_action.go:964`) selects
`WHERE status NOT IN ('deleted','archived')`. So:

1. **ARCHIVE** (SQL; reversible; page keeps serving) — this alone drops the page out of `validPages`,
   KEEP #2 goes false, and control reaches the positional pick the `nav_order` demotion made correct.
2. **RE-RESOLVE** — `page_rerender` with `spec.reason='cta_links_stale'`, **no LLM**.
3. **RETRACT** — the `page-retraction` agent (`site_id_field`, `page_ids_field`). It also deactivates
   the `site_nav_items` row; editorial inbound is *refused*, structural inbound is *mechanised*,
   newly-stranded outbound is *reported*.

**CONFIRMED end-to-end at the served bytes on two canaries**, both now `0` `password-entropy` hrefs
(`finetuning.uk/about.html`, `leopardess/careers.html`), with `content_data`, `rendered_html` and the
git commit all agreeing.

**Worked SQL, reusable:** `SQL_2026-08-26_archive_password_entropy_canary_finetuning.sql`,
`SQL_2026-08-26_archive_and_canary_leopardess.sql`, `SQL_2026-08-26_canary_relink_after_archive.sql`.

## 4. ⚠ TRAPS THIS LANE PAID FOR — all of these cost hours

- **A `content_rewrite` commissioned for LABELS ONLY rewrites the page BODY.** It destroyed authored
  copy on **2 of 12** pages by overwriting sections with copies of a neighbour. Both restored from
  the offending write's own archive. `LANDMINES.md` + `bugs_open/403`.
- **`grep -c '<p'` is not a prose control.** It read `15 → 15` on a destroyed page (three paragraphs
  replaced by three). **Use `count(*)` vs `count(DISTINCT md5(stripped_text))` per page.** ⚠ Not
  `left(text,80)` — that false-positives on pages whose sections share a heading.
- **The two retraction siblings have OPPOSITE `dry_run` defaults.** `retract_asset_files`: absence
  means dry-run. `retract_page_deployment`: absence means **DELETE** (bool zero value). The live
  `page-retraction` agent passes no `dry_run`. `LANDMINES.md`.
- **There is NO `content_components` row named `tool-password-entropy`.** The active library row is
  **`tool-password-entropy_pre_037`**; the four per-site rows are already inactive. Assert on
  `count(*) WHERE name ILIKE '%password-entropy%' AND is_active = 1`, or your guard falsely aborts.
- **`spec.page_name` is load-bearing** on a rerender — without it the rerender discards its own
  result (`sections_saved: 0, success: true`) and deploys the stale assembly.
- **Item keys dedup in ANY status** (`bugs_open/326`) — a retry needs a fresh `item_key`.
- **The footer is GENERATED, not authored.** Its `content_data` holds no tool links; the served
  footer lists the live tool set in `nav_order` order. Retiring removes the link by construction —
  the risk is the inverse, retiring without refreshing. Recipe: `nav-link-fixer`, then propagate in
  **assemble mode** (`page-rerender` with **no** `spec.reason`). Do **not** use the agent whose name
  says navigation; it deletes every child-path link (NAV-013).
- **`priority` is inert BETWEEN sites** (`bugs_open/413`). Site selection orders by `created_at`;
  priority only orders items *within* a site. An item's wait is set by the age of the oldest item on
  its site — someone else's row.
- **Read `last-modified` BEFORE the body.** A correct commit with a stale served page is a delivery
  lag, not a refutation — and when every check comes back healthy, **ask the CI provider's status
  page** (016b §9).

## 5. AFTER THE DECISION — the rest of the sequence

1. Resolve §2 (the 23 contact-intent fields), then re-resolve the remaining fields per site.
2. **Archive `ai-agent-orchestration.com`** — kept for last deliberately: it is the only site with a
   `site_nav_items` row (`in_header=t`, rendered into the **footer** via `classifyPagesForNav`'s
   utility demotion) and the only one whose footer carries the link.
3. **Retract** all three deployments once editorial inbound is 0.
4. **Final sweep:** zero `password-entropy` at the served bytes on all three sites, footer included.
5. **Owner decision 3** — the platform lever. Read at the RANKING not the loaders
   (`render_site_components_action.go:182-190` takes `ordered[0]` and never persists it); bind
   `LoadCTALabelUniverse` too; engage RFC_022 and **enumerate the consumers** before booking a council
   round.

## 6. Session-start checklist
`git log --oneline -10` · re-read this from disk · `scripts/who-owns.py 391` · **chassis image tag**
(not a binary grep) · `curl -s https://www.githubstatus.com/api/v2/summary.json` if anything looks
undelivered · the §1 state queries · then §2.
