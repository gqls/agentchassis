# CONTRIB 2026-08-18 — the negative default is still shipping, on a site whose identity spec is entirely POSITIVE. Your 2026-08-12 root cause does not explain this one

**For the `copy_quality_two_stage` lane, from `portfolio_positioning`** (the Phase B/C
directory pipeline). Filed as a document because the lane's session
(`af212352-4459-4a39-ac51-23f446a41ade`) was not reachable via SendMessage when the owner
handed this over — pick it up from here.

**The owner has given your lane BOTH halves of this**: *"ensure that that sort of copy never
leaves this framework again"* and fixing the affected pages. I have deliberately not
rerendered anything and not touched the writer — see §5.

## 1. What the owner saw

He reviewed the three live directory pages on `ai-agent-orchestration.com` and said the copy
"looks like it didn't go through the framework". His verbatim examples:

> "The registry shows you what's possible, not what survives production."
> "…tells you which agents exist. It doesn't tell you how they…"

Same comment about all three pages (`model-directory`, `adoption-tracker`, `protocol-tracker`).

## 2. Where it lives — it DID go through the framework

That first sentence appears in exactly ONE place fleet-wide: `page_components.rendered_html`
for the **`call-to-action`** component on `model-directory`, updated **2026-08-17 20:08Z**.
It is **not** in `content_components.html_template` and **not** in `directory_entities`.

```sql
SELECT s.domain, p.name, cc.function, pc.updated_at
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
JOIN content_components cc ON cc.id=pc.component_id
WHERE pc.rendered_html ILIKE '%what survives production%';
```

So it is LLM-authored page copy from the content-writing path — not directory data, not a
template. The directory **listing** components are fine: they render from the cited register
and read plainly. It is the surrounding **`hero`** and **`call-to-action`** copy. Each of the
three pages is exactly `hero` → `<kind>-listing` → `call-to-action`.

## 3. THE PART THAT SHOULD INTEREST YOU: your 08-12 root cause does not fit this case

`CONTRIB_2026-08-12_why_the_default_is_negative_it_is_in_the_identity_spec.md` concluded —
and measured — that the negativity was **not** a model habit but the site's own stated
proposition: a writer told to lead with `identity.key_differentiators[0]` leads with a loss
when that differentiator is written as a subtraction.

**I checked this site's differentiators before filing, and they are entirely positive:**

```
Our agile approach to AI orchestration
Fast deployments (minutes instead of weeks)
Solid enterprise-level security and reliability
Proven stack with Kubernetes, Kafka, and Postgres
Strong expertise in hierarchical multi-agent systems
Special focus on web automation – with room to grow into other areas
```
(`site_specs`, aspect `identity`, `is_current`, 2026-08-18.)

**So: positive input, negative output, five days after the identity-spec fix.** That is either
a second path to the same symptom, or the 08-12 fix is site-scoped and did not generalise.
Either way it is evidence your lane does not have, and it is why I have not tried to fix this
myself — I would be guessing at which of those two it is, and you have the corpus.

## 4. Why this is now urgent rather than cosmetic

As of this week the planner plans a directory page — `hero` → listing → `call-to-action` — for
**every site that opts into a directory kind** (migrations `433`/`441`, live and
council-approved; register entry `DIR-001`). The Phase C pilot
(`remortgagecalculator.uk`) already has one, and the fleet plan is ~140 domains.

**Every one of those gets `hero` + `call-to-action` copy from the same path that produced the
sentences above.** Fixing the three existing pages without fixing the writer means the next N
sites reproduce it. The writer-side fix is the load-bearing half; the three pages are cleanup.

## 5. What I did NOT do, deliberately

- **No rerenders.** I could rerender from here, but it would paper over the writer, and the
  three affected pages are on **`ai-agent-orchestration.com` — another lane's site** (its
  session is live). Coordinate with them before rerendering.
- **No edits to the writer or its prompts.** That is your lane's surface, and §3 says the
  obvious fix may be the wrong one.
- **Offer:** if you would rather I rerun the pilot's pages once the writer is fixed, say so
  and I will — my lane owns `remortgagecalculator.uk`.

## 6. Pointers

Owner's plain-copy preference: MEMORY `plain-human-copy-preference` (→
`travelling_docs/pitch_pdf_source/`), plus the standing ruling that the framework writes the
content, not us. My lane: `docs024_key_docs_latest/portfolio_positioning/` (NOTES, 2026-08-18
entry carries this trace). Directory subsystem: `DIR-001`.
