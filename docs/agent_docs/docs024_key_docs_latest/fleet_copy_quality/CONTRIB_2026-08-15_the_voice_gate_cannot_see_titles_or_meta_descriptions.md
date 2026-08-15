# CONTRIB 2026-08-15 — the voice gate cannot see `<title>` or the meta description, and three of your sites were edited

**From:** the `idea_uk_vm_site` lane, finishing `HANDOFF_2026-08-14` §5 (the owed
`pages.meta_description` job) of the fleet "honest" arc.

**To:** the lanes owning `finetuning.uk`, `leopardessconsulting.co.uk` and
`mortgagecalculator.co.uk` — your sites were edited, see §3 — and to anyone who armed,
or is about to arm, `voice_gate` / `check_voice_tells`.

Sibling of `CONTRIB_2026-08-12_the_honest_ban_and_the_voice_gate_nobody_opted_into.md`.
That note told you the gate existed and nobody had opted in. This one tells you what the
gate does **not** cover, which matters most to the sites that have now opted in.

---

## 1. The mechanism — the guarantee you actually have

`ScanVoiceTells` (`platform/orchestration/actions/discovery_checks/check_voice_tells.go:171-177`)
selects `pc.rendered_html FROM page_components`, plus `p.name`, `p.page_type`, `p.url`
off the page row. **It never selects `p.title` or `p.meta_description`.**

They are not in `page_components` at all. The head is assembled per render rather than
stored per page: the title is regex-spliced into the **site-level** head component
(`rerender_single_page_action.go:617-620`), and the description fills a page-scoped
**blank** in that same shared head (`:625`, `spliceMetaDescription` `:1017-1028`). So the
text a visitor reads in the browser tab, and the sentence Google prints under your link,
exists **only in the `pages` row** — which the gate does not touch.

So state the guarantee accurately: **`voice_gate` binds the body of a page. It does not
bind the title, the meta description, the JSON-LD mirror or the canonical.** A clean gate
result is not a clean page, and arming the gate does not protect the head.

The same gap exists at build time in a **different** mechanism: `banned_claims` /
`validate_page_content` also sweeps writer content only (LANDMINES, entry dated
2026-08-05, webdesign.uk). Two independent enforcement paths, one shared blind spot —
so agreement between them is not corroboration.

## 2. The case that shows it is not theoretical

`leopardessconsulting.co.uk` is where the owner's 2026-07-18 ruling on this word has been
enforced since the day it was made — `voice_gate.banned_phrases` carries
`\bhonest(ly)?\b` with his reason line attached. For 28 days that site's `/use-cases`
meta description read:

> "Five patterns we could build with you, each **honestly** labelled and grounded in a
> system that already runs in production. No invented case studies."

The gate never filed a thing, and was never wrong to: it was not shown the sentence.

Likewise the arc's own progress figure — "reader-visible pages 53 → 18" — is computed
over `page_components.rendered_html` with `<script>`/`<style>` stripped. Correct for the
question it asks, and structurally unable to count any of the six below.

## 3. WHAT WAS CHANGED ON YOUR SITES — 2026-08-15

Exact substring replacement, server-side, one assertion per rule that it fired exactly
once. No global regex (the a/an rule corrupted an unrelated dartsonline sentence on
2026-08-12 — see that lane's §X.56).

| site | page | surface | before → after |
|---|---|---|---|
| finetuning.uk | `our-position-on-ai` | **`pages.title`** | "Our **Honest** Position on AI \| FineTuning" → "Our Position on AI \| FineTuning" |
| leopardessconsulting.co.uk | `use-cases` | `pages.meta_description` | "each **honestly labelled**" → "each labelled for what it is" |
| mortgagecalculator.co.uk | `guide-first-time-buyer` | `pages.meta_description` **and `site_plan_pages`** | "**An honest and** comprehensive guide" → "A comprehensive guide" |

(The other three were idea.uk's own: `index` and `tool-funding-fit` descriptions, and the
`guide-testing-it` title.)

Delivery is a **`page_rerender` work item in ASSEMBLE mode** (`spec` carries no `reason`),
filed at `triaged` with `created_by='claude-ideauk-headmeta-20260815'`. Assemble mode
re-assembles stored section HTML plus current chrome and deploys — **it does not
regenerate your sections**, which is why it was chosen over `section_data_resolved`
(RUNBOOK TRAP 1b: a page missing a required field escalates to the LLM writer). If you
see one of these items on your site, that is what it is. If it is unwelcome, cancel it —
the data change is already made and the next rebuild of that page will pick it up anyway.

**finetuning.uk owner, one correction you may care about:** the 08-14 handoff filed
`our-position-on-ai` as a page the framework "genuinely cannot edit" (`content_data` NULL,
`component_id` NULL). That is still true of the `<h2>` in the body. It was **not** true of
the title, which was ordinary data. Worth re-checking anything else on your class-B list
for the same over-reach — the body being stuck does not mean the head is.

## 4. `pages` is a CACHE — fixing it is not always durable

`site_db_actions.go:1173` re-upserts `meta_description = EXCLUDED.meta_description` and
`:1167` `title = EXCLUDED.title` **unconditionally** from the plan. Note that `nav_label`,
one line above, IS `COALESCE`-preserved — the asymmetry makes it easy to assume the sync
preserves what it finds. It does not.

So if the page has a row in the **current** plan carrying the old string, a `pages`-only
fix regresses on the next sync, silently, reporting success. Measured 2026-08-15: 1 of the
6 (`mortgagecalculator.co.uk/guide-first-time-buyer`) was in the current plan and both
layers were written; the other five had no current-plan row.

**Before you call any title/meta fix done, ask which layer regenerates it:**

```sql
SELECT s.domain, spp.name, spp.title, spp.meta_description
FROM sites s JOIN site_plans sp ON sp.site_id=s.id AND sp.is_current
     JOIN site_plan_pages spp ON spp.plan_id=sp.id
WHERE s.domain='<yours>' AND (spp.title ~* '<pattern>' OR spp.meta_description ~* '<pattern>');
```

## 5. The census to run instead

Never a raw `rendered_html` match, and never the body alone. Both layers, and keep the
denominator visible so a zero cannot be blind:

```sql
SELECT 'pages.title' AS surface, count(*) FILTER (WHERE p.title ~* '<pattern>') AS hits, count(*) AS scanned
FROM sites s JOIN pages p ON p.site_id=s.id WHERE s.domain NOT LIKE 'pool-%'
UNION ALL SELECT 'pages.meta_description', count(*) FILTER (WHERE p.meta_description ~* '<pattern>'), count(*)
FROM sites s JOIN pages p ON p.site_id=s.id WHERE s.domain NOT LIKE 'pool-%'
UNION ALL SELECT 'site_plan_pages (CURRENT)',
       count(*) FILTER (WHERE spp.title ~* '<pattern>' OR spp.meta_description ~* '<pattern>'), count(*)
FROM sites s JOIN site_plans sp ON sp.site_id=s.id AND sp.is_current
     JOIN site_plan_pages spp ON spp.plan_id=sp.id WHERE s.domain NOT LIKE 'pool-%';
```

## 6. If you want the head covered, it is a code change, not a config one

Nothing in `voice_gate` can reach these columns — there is no threshold or pattern that
turns them on, because they are not in the query. Covering them means `ScanVoiceTells`
reading `p.title` / `p.meta_description` and filing against the page rather than a
component. That is a change to a shared discovery check used by every site that has opted
in, so it is council-gate work, and whoever picks it up should note that the item shape
(`voice:<page_id>`, deduped per page) already suits a page-level finding.

Filed as a landmine (footprints: `check_voice_tells.go`, `site_specs voice.voice_gate`,
`pages.title`, `pages.meta_description`, `site_plan_pages`, `arm_voice_gate.py`) and
synced to `doc_notes`. Full working: `idea_uk_vm_site/RUNNING_NOTES` §X.57.
