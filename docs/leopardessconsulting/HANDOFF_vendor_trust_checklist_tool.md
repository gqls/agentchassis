# HANDOFF — build the "AI Vendor Trust Checklist" interactive tool

**Purpose.** Start a fresh chat from exactly here. This file is self-contained: you do not
need to have read anything else in this directory to build this tool, though the context
in §1 explains why it exists and §7 links the article it supports.

**Site:** `leopardessconsulting.co.uk` · `site_id = 4851f6fc-71cf-4160-a270-e03d6d3e0732`
**Branch:** `087_towards_multiple_domains` (or whatever `git branch --show-current` says —
check, don't assume; this repo runs many concurrent sessions on one shared tree, see
`CLAUDE.md` at the repo root before touching anything, especially the git-commit rules)
**Written:** 2026-07-30, by the session that published the trust-series articles this tool
supports. Nothing below has been built yet — this is a plan, not a status report.

---

## 1. What this is and why it exists

On 2026-07-29/30, a 4-part researched article series was published on this site about
trusting AI with data (`/blog/can-you-trust-ai-with-your-data.html` + three industry
deep-dives). The pillar article has a section, "What 'Trustworthy' Actually Looks Like,
Concretely," which argues that evaluating an AI vendor's data-handling posture is no longer
a matter of taking their word for it — there is now a real, checkable list (certifications,
retention policy, training-data policy, incident response), illustrated using Anthropic's
own published trust-centre documentation as a worked example.

The owner, reviewing that work, asked whether a tool could be built for this content — the
site already has several: an AI-readiness quiz, an LLM cost calculator, an agent-complexity
estimator, a password-entropy checker. The previous session assessed it as a good, feasible
fit and **deliberately did not build it** — it decided a new interactive feature deserved
its own properly-scoped, browser-tested build rather than being rushed in at the end of an
already large session. This handoff is that scoping.

**The tool: a short, deterministic, no-LLM checklist that scores an AI vendor's data-
handling trustworthiness** against the specific factors the research (and the article)
identified as checkable — and returns a plain-English verdict, not a black-box score.

## 2. Design starting point — adjust freely, this is a proposal not a spec

A set of yes/no toggles across four categories, mirroring the pillar article's own framing
("what's certified, what's retained, what's used for training, what happens if something
goes wrong"):

**Certifications** (does the vendor publish evidence of):
1. SOC 2 Type II attestation
2. ISO/IEC 42001 (AI management systems) or ISO 27001 (information security)
3. A sector-specific certification if relevant to your use case (HIPAA/BAA, PCI-DSS,
   FedRAMP, etc.) — or N/A if not applicable

**Data handling:**
4. Does NOT train on your data by default — opt-in only, not opt-out
5. Offers Zero Data Retention (or equivalent) for regulated/sensitive workloads
6. Publishes a sub-processor list or data-flow diagram
7. Has a clear, published process for deleting your data on request

**Governance & oversight:**
8. A human reviews or can override high-stakes automated decisions
9. Decisions/outputs are auditable — you can see why it did what it did
10. Publishes an incident-notification commitment (what happens, how fast, if something
    goes wrong)

**Transparency:**
11. Discloses to the people affected when AI is being used on/for them (not silent)
12. Publishes any accuracy, bias, or fairness testing it has done

**Scoring:** count of "yes" out of 12 (treat N/A on item 3 as excluded from the denominator,
i.e. score out of 11 if N/A is selected — decide the exact UX for this when building).
Three plain-English verdict tiers, not just a number:
- **9–12 (or ratio ≥ 0.75): "Strong footing — the checkable evidence is there."**
- **5–8 (0.4–0.74): "Reasonable, but ask about the gaps before you commit."**
- **0–4 (< 0.4): "Significant gaps — get answers on these before proceeding."**

Each item, when unchecked, should show a one-line "why this matters" note (short, factual,
non-alarmist) — e.g. for item 4: "Vendors that train on customer data by default have, in
practice, repeatedly had that policy become a point of dispute later. Opt-in is the safer
default." Keep these consistent in tone with the article series (measured, cited where
possible, not scaremongering).

**Do not present the score as a guarantee.** The pillar article's own words on this are the
right framing to reuse or paraphrase: *"None of that guarantees good outcomes on its own —
certifications describe a process, not a promise... it is now realistic to ask any AI
system touching your data for the equivalent list."* The tool should embody that same
honesty — it is a conversation-starter checklist, not a certification.

## 3. The exact technical pattern to follow — a complete, working example already on this site

This site's tools are `content_components` rows rendered as a page section, with the
interactive logic in a separate static JS file. **`tool-ai-agent-roi-estimator` is the
cleanest existing example and should be your copy-paste starting point.** Fetch both parts
directly from the live DB before writing anything:

```sql
-- the HTML/CSS template (a Go text/template — {{.field}} placeholders resolve from
-- page_components.content_data, NOT from the tool's live interaction state)
SELECT html_template FROM content_components WHERE name='tool-ai-agent-roi-estimator';
```
```bash
# the client-side JS — this is the ENTIRE interactive logic, no server round-trip
curl -s https://leopardessconsulting.co.uk/tools/assets/tool-ai-agent-roi-estimator.js
```

Structurally, what you'll find (already verified working, live, in a browser-servable
state):
- A `<style>` block scoped under a single top-level class
  (`.tool-ai-agent-roi-estimator-section`), using the site's real CSS custom properties
  (`--color-primary`, `--color-accent`, `--section-*`) with literal fallbacks — follow this
  exactly, it's how the tool inherits the site's theme without hardcoding colours.
- A `<section class="..." data-component="tool-<name>">` wrapper. **The JS finds its own
  section via `document.querySelector('[data-component="tool-<name>"]')`** — this is the
  hook that lets the same page host multiple independent tool sections without collision.
- Template fields (`{{.headline}}`, `{{.slider_employees_label}}`, etc.) are all either
  `source:"llm"` (copy written once at build time, stored in `page_components.content_data`
  — see §5, you'll write these by hand, not via an LLM call, matching this session's
  established no-fabrication practice for this site) or `source:"static"` with a
  `fallback:"..."` (a value that never changes) or `source:"site_specs.cta.primary_url"`
  (resolved from the site's own spec — copy this exact pattern for a CTA link if you want
  one).
- A trailing `<script src="/tools/assets/tool-<name>.js"></script>` tag. **The filename
  here MUST exactly match the deployed file's path** — `llm-cost-calculator` (one of this
  site's 5 tools) is currently broken precisely because its template references the wrong
  JS filename (`bayesian-ranking-hero-tool.js`, a different tool's file). Do not repeat
  this — verify the exact path you commit in §5 matches the exact path in the template you
  write in §4, character for character, before you consider this done.
- The JS file itself: a single IIFE, no dependencies, no framework, no LLM call, no fetch.
  It queries its DOM elements by id (scoped inside the section it found via
  `data-component`), attaches `input`/`change`/`click` listeners, computes, and writes
  results back into the DOM with `.textContent`. For a checkbox-based checklist, the same
  shape applies: query all `input[type=checkbox]` inside the section, listen for `change`,
  recompute the score and verdict tier, update a results area. This is genuinely simple —
  the ROI estimator's `calculate()` function (59 lines total) is a complete template for
  the shape of the logic you need, just replace sliders with checkboxes and a dollar
  formula with a count/tier lookup.

## 4. Registering the component and page (DB side)

Follow the exact shape of the existing `ai-agent-roi-estimator` page. Read these two rows
first and mirror their structure precisely rather than guessing field names:

```sql
SELECT * FROM pages WHERE site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND name='ai-agent-roi-estimator';

SELECT pc.slot_name, pc.position, cc.name, cc.function, cc.render_mode, cc.category,
       jsonb_pretty(pc.content_data)
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN content_components cc ON cc.id=pc.component_id
WHERE p.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732' AND p.name='ai-agent-roi-estimator';
```

You will need, in order:
1. A new `content_components` row (e.g. `name='tool-ai-vendor-trust-checklist'`,
   `function` matching `name`, `render_mode='template'`, `category='general'`) with your
   `html_template` and an `input_schema` mirroring the ROI estimator's field-sourcing
   pattern (§3).
2. A new `pages` row: `page_type='tool'`, `url='/tools/ai-vendor-trust-checklist.html'`,
   `sections=["tool-ai-vendor-trust-checklist","tool-cta"]` (the `tool-cta` slot is a
   shared, already-working component that pulls contact details from `site_specs`
   automatically — include it exactly as the ROI estimator does, no changes needed there).
   Decide `nav_order`/`in_footer` by looking at the current values across all 5 tool pages
   and slotting in consistently (query in §4 above the sections query shows you the
   pattern — `nav_order` 1/6/7/200/200 currently, `in_footer` mixed).
3. `page_components` rows for both slots, with `content_data` filled in completely — **run
   the same two escalation-guard checks this session used for every one of the 4 published
   articles before ever firing a render** (see §6.1 below), or a render can silently
   escalate to the automated content-writer instead of using your hand-written copy.

**Use dollar-quoted SQL string literals for anything containing real HTML/JS content**,
e.g. `$MYTAG$...content...$MYTAG$`, not single-quoted strings — the moment your content
contains an apostrophe (and copy like "don't" and "vendor's" will), a naive single-quoted
literal breaks. Pick a tag string not present in your content and assert that
programmatically before writing the SQL (this session's Python scripts did exactly this;
copy the pattern from `docs/leopardessconsulting/scripts/L8_article_trust_pillar.sql` if
useful as a reference for the INSERT shape, even though that one is for a blog post, not a
tool).

## 5. Deploying the JS file — the exact working recipe

The JS file is a **static asset committed directly to the site's git repo**, not something
that goes through `page_components` rendering. The proven, already-working mechanism for
this (used for this site's logo/favicon/OG-card, and generalises to any file) is a direct
message to the platform's git-adapter — **read
`docs/leopardessconsulting/scripts/commit_brand_assets.sh` in full before writing anything
new**; it is a complete, working, documented example. The shape:

```python
msg = {
    "headers": {
        "correlation_id": "<uuid>", "orchestration_id": "<uuid>", "request_id": "<uuid>",
        "client_id": "demo_client", "step_name": "commit_tool_asset",
        "message_type": "request", "sender_agent_type": "user",
        "sender_agent_id": "<uuid>", "sender_pod_name": "cli",
        "responses_topic": "system.agent.generic.responses",
    },
    "body": {
        "action": "commit",
        "data": {
            "repo_name": "sites",
            "domain": "leopardessconsulting.co.uk",
            "files": {
                "tools/assets/tool-ai-vendor-trust-checklist.js": {
                    "content": "<base64 of your JS file>",
                    "encoding": "base64",
                }
            },
            "commit_message": "feat: AI vendor trust checklist tool JS",
        },
    },
}
```
Sent as a single line (no embedded newlines) to Kafka topic
`system.adapter.git.requests` via `kcat`, exactly as `commit_brand_assets.sh` does it.
**Verify by curling the deployed path afterward** — `https://leopardessconsulting.co.uk
/tools/assets/tool-ai-vendor-trust-checklist.js` should return your JS with HTTP 200 —
never trust the git-adapter's own logs or a "sent" message as proof it landed (see §6.2 on
propagation delay: this session hit a consistent pattern of a fresh deploy 404ing or
serving stale content on the first curl, then resolving cleanly on retry a few seconds
later — retry once before concluding something failed).

## 6. Landmines from the session that published the article series — all directly relevant here

### 6.1 — The escalation guard (do this before every render, no exceptions)

A section rerender in `section_data_resolved` mode escalates the WHOLE PAGE to an
automated content-writer (which can fabricate — this happened to this exact site before)
if content_data for ANY section on the page fails either of two checks:
```sql
-- branch A: content_data must be a non-empty JSON OBJECT (not null, not '{}', and
-- critically NOT a valid JSON array/scalar either -- those unmarshal to a nil map and
-- silently trip this same check)
SELECT slot_name FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.name='<your-page>' AND p.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732'
  AND (pc.content_data IS NULL OR jsonb_typeof(pc.content_data::jsonb) <> 'object'
       OR pc.content_data::jsonb = '{}'::jsonb);
```
```sql
-- branch B: every field the component's schema marks required+source:"llm" must be a KEY
-- present in content_data (even if you're supplying the value by hand, not via an LLM)
SELECT pc.slot_name,
  (SELECT array_agg(k) FROM jsonb_object_keys(cc.input_schema->'fields') k
     WHERE (cc.input_schema->'fields'->k->>'source')='llm'
       AND COALESCE((cc.input_schema->'fields'->k->>'required')::boolean,false)=true
       AND NOT (pc.content_data ? k)
  ) AS missing_required_llm_fields
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN content_components cc ON cc.id=pc.component_id
WHERE p.name='<your-page>' AND p.site_id='4851f6fc-71cf-4160-a270-e03d6d3e0732';
```
Both must come back clean (no rows / all NULL) before you fire a render. This is not
optional caution — it is the difference between your hand-written, accurate copy shipping,
and an LLM silently rewriting the whole page.

### 6.2 — Propagation delay is normal, not failure

A fresh deploy (page render, image, or JS asset) will sometimes 404 or serve stale content
on the very first check immediately after the orchestration reports COMPLETED, then
resolve cleanly seconds later. This happened repeatedly and consistently in the session
that built the article series. **Retry once with a cache-busting query string before
concluding something is broken.** Do not add a blocking `sleep` to work around it — just
do something else for a beat (another check, another file write) and re-curl.

### 6.3 — Chassis pod age, before firing anything

No orchestration dispatch within ~300 seconds of a chassis pod (re)start — the spawn is
silently dropped. Check first:
```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis -o custom-columns='NAME:.metadata.name,START:.status.startTime'
```
and compare against current time.

### 6.4 — Cross-links to a not-yet-existing page get silently stripped

If this tool's page links to the pillar article (recommended — link back to
`/blog/can-you-trust-ai-with-your-data.html`, and consider adding a link to the tool FROM
the pillar article's "What Trustworthy Actually Looks Like" section once the tool exists),
know that a `<a href>` pointing at a page that doesn't exist yet at render time gets
silently stripped from `rendered_html` (confirmed this session: `content_data` keeps it,
`rendered_html` does not — mechanism not diagnosed further). Since the pillar article
already exists, linking TO it from the new tool page should be fine on the first render.
Linking FROM the pillar article back to the new tool will need the pillar article
re-rendered once the tool page exists — the same one-extra-step this session needed
between its own four pages. Use `scripts/rerender_pages.sh <site> <domain>
can-you-trust-ai-with-your-data` for that (regenerates from `content_data`, which will
still hold the link even if a previous render stripped it).

### 6.5 — Don't fire the full `rerender-pages` workflow for a small, contained change

For deploying just this one new page, use `scripts/rerender_pages.sh <site> <domain>
ai-vendor-trust-checklist` (or whatever `name` you choose) — this drives `page-rerender`
directly for one page. Do **not** fire the full `rerender-pages` agent workflow expecting
it to "pick up" the new page — that workflow's `create_rerender_items`/
`get_pages_for_rerender` steps operate at a different scope and this site has a history of
a wide rebuild clobbering hand-fixed content on other pages (historically `bugs_open/001`,
now closed — see `HANDOFF.md` §10 in this same directory — but the general principle of
"use the narrowest tool for the job" still holds regardless of that bug's status).

### 6.6 — `%` in a SQL string literal is not a printf escape

If you write your SQL via an f-string/format-string in Python (as this session did
throughout), a literal `%` in your copy (e.g. "70% of...") does NOT need doubling to
`%%` — that convention is for printf-style or `psql -c` format strings, not for the
content of a plain single- or dollar-quoted SQL literal read from stdin. This session
wrote `%%` out of habit in one meta_description and had to fix it afterward (it landed in
the DB as a literal double-percent). Write `%` as `%` in your content.

## 7. Where this fits — the published context

- Pillar article, live now: `/blog/can-you-trust-ai-with-your-data.html` — its "What
  'Trustworthy' Actually Looks Like, Concretely" section is the direct inspiration and
  should read as a natural companion to this tool once both exist. Consider a one-line
  cross-link once the tool ships (see §6.4 for the mechanics).
- Full research + citation trail for anything you want to reference in the tool's
  "why this matters" notes: `RUNNING_NOTES.md` in this directory, 2026-07-29 entries.
- Existing tools to look at for pattern/precedent, in order of usefulness for THIS build:
  `ai-agent-roi-estimator` (best match — sliders + live results, use as the template),
  `password-entropy` (simpler, fully self-contained, good if you want an even leaner
  starting point), `tool-agent-complexity-estimator` (a different, unusual pattern —
  `sections: []`, likely served as a static file outside the normal render pipeline; not
  recommended as your template, mentioned only so you don't mistake it for the norm).

## 8. Acceptance criteria before calling this done

- [ ] Page loads in a real browser (not just `curl`), at
      `/tools/ai-vendor-trust-checklist.html` (or your chosen slug), on both desktop and a
      mobile viewport width.
- [ ] Every checkbox/toggle updates the score and verdict tier live, client-side, with no
      page reload and no network request.
- [ ] The verdict tiers and "why this matters" notes are readable, non-alarmist, and
      consistent in tone with the published article series.
- [ ] The JS file is committed and its path matches the template's `<script src>` exactly
      (§3 — this is precisely how `llm-cost-calculator` is currently broken; don't repeat it).
- [ ] Linked from wherever this site's other tools are linked from (footer utility nav —
      check the current pattern across the 5 existing tool pages before wiring).
- [ ] Full-site image/link sweep re-run after publishing (this session's practice
      throughout: refetch the sitemap, check every page for anything broken) to confirm
      nothing regressed elsewhere.
