# 268 — a `content_rewrite` drops the CTA destination keys, and 214 call-to-action buttons are missing from live pages fleet-wide

**Filed:** 2026-08-12 · **Lane:** `ai_site_selling_automation` · **Severity: high** —
the primary conversion control is absent from 214 components across 19 live
customer-facing sites, and it fails silently in every instrument we have.
**Class:** structural (shared component schema + the regeneration write path).

> **STATUS: OPEN.** webdesign.uk is protected by a site-scoped component lock
> (`SQL_2026-08-12k`), which is a tourniquet and not a fix. The other 18 sites
> are unprotected. The mechanism is NOT established — see §4, including a
> refuted hypothesis of mine.

---

## 1. The symptom

A component that carries a button LABEL but no button URL renders **no anchor at
all**. Both shared templates gate the anchor on the URL rather than the label:

```
hero            {{if and .cta_text .cta_url}}<a href="{{.cta_url}}" …
call-to-action  {{if and .primary_cta .primary_cta_url}}<a href="{{.primary_cta_url}}" …
```

So the failure produces: no error, no missing prose, no shortened byte count, a
clean claims scan, and a page that looks finished. **The call to action is
simply not there.**

## 2. Fleet exposure, measured 2026-08-12 ~20:45Z

```sql
SELECT s.domain,
       count(*) FILTER (WHERE pc.content_data ? 'cta_text' OR pc.content_data ? 'primary_cta') AS has_label,
       count(*) FILTER (WHERE (pc.content_data ? 'cta_text' OR pc.content_data ? 'primary_cta')
                          AND NOT (pc.content_data ? 'cta_url' OR pc.content_data ? 'primary_cta_url')) AS label_no_url
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.slot_name IN ('hero','call-to-action') AND p.status='active'
GROUP BY s.domain ORDER BY 3 DESC;
```

**216 components across 19 sites carry a label with no URL. 214 of them render
zero anchors.** Worst affected: finetuning.uk 31, ai-agent-orchestration.com 29,
gamesdesign.co.uk 21 (of 28), mortgagecalculator.co.uk 21 (of 25),
relojistas.com 17 (of 22), dartsonline.com 15, fundamentallyai.com 15.

The two shared components are `hero` (`23f95f00-…`, 20 sites, 276 instances) and
`call-to-action` (`0197e8d7-…`, 20 sites, 237 instances).

## 3. Reproduction, with a control — this is the load-bearing evidence

A `content_rewrite` (`mode=edit_live`) was dispatched against five webdesign.uk
pages at 20:2xZ. The URL keys were present in `content_data` BEFORE it ran,
because this lane had restored them earlier the same afternoon.

| | before | after |
|---|---|---|
| components carrying `cta_url` / `primary_cta_url` | 7 | **0** |
| site-wide `href="` count | 28 | **13** |

**The control: `contact/hero` was NOT part of the rewrite. It kept its keys
(`cta_url` + `secondary_cta_url`) and both its links.** Every component that WAS
rewritten went to `0|0|0`. Same site, same components, same schema, same run —
the only difference is whether the page was regenerated.

Raw before/after in
`docs024_key_docs_latest/ai_site_selling_automation/` scratch notes and the
NOTES entry of 2026-08-12 (close).

## 4. Mechanism — CANDIDATE ONLY, and one reading already refuted

The fields are declared in `content_components.input_schema` as
`{"type":"url","source":"renderer","required":false,"on_missing":"skip_field"}`.

The candidate reading: `sourceResolver.resolve` short-circuits that source —
`if source == "" || source == "llm" || source == "renderer" || source == "static" { return nil, true }`
— returning **found=true with a nil value**. The field is therefore never
*missing*, `handleMissingField` never runs, and so `carryStored` (the
`bugs_open/238` carry, PBP-039) never runs either: the carry protects fields
that FAIL to resolve, and this class always "succeeds" with nothing.
`plan_sections` then `continue`s on the renderer/static branch writing only a
declared `fallback`, and these declare none. `save_page_sections` replaces
`content_data` wholesale, so the key is gone.

> **⚠ `090` run `97ef39f0-19df-4935-834d-c80514fbc43e` REFUTED this.** Its
> citations are `content_data` rows carrying `"cta_url": "/contact.html"` —
> **the values this lane had restored sixteen minutes before the run started.**
> It measured a repaired system and correctly reported nothing missing. The
> refutation is therefore not decisive either, and the reproduction in §3 (which
> post-dates it, and has a control) is the better evidence. **A re-run is owed,
> authored against `page_component_history` for the 16:37–17:23 window, with the
> symptom stating plainly that the live rows were repaired at 17:23.**
>
> A second reading of mine was refuted outright: I assumed the URL keys were
> **undeclared** in the schema. They are declared, all four, with
> `source: renderer`. A field outside the carry's reach and a field absent from
> the schema look identical from the symptom.

Also note **238 §8 is stale**: it says both fix halves are "inert until the
fleet next rolls". They have rolled — `agent-chassis v1.0.1291` was built from
`da5a7eb8f`, and `git merge-base --is-ancestor d26c26a9a da5a7eb8f` passes with
controls both ways. So the carry is LIVE and this still happened.

## 5. What does NOT fix it

- **Restoring `content_data` alone.** Proven twice on 2026-08-12: a
  `page_rerender` dispatched *after* the keys were restored still rendered no
  buttons.
- **A fallback on the shared schema.** `/contact.html` is wrong for sites with
  no such page; it would ship broken links to 19 other sites.

## 6. Fix candidates, ordered by what closes the door

1. **Take the URL fields out of the renderer short-circuit** so a failed
   resolution reaches `handleMissingField`, and the existing 238 carry protects
   them. Config-only (`content_components.input_schema`), live immediately, no
   roll — but it touches a component used by 20 sites, so it is architecture
   scope and wants the council gate. **Verify the carry actually fires before
   assuming this is sufficient.**
2. **Make `resolve_internal_links` resolve these CTAs**, which is the platform's
   intended mechanism — it already files `unresolved_cta` items saying "no real
   page destination", so it knows it failed. Go change; inert until a roll.
3. **Render-time fallback in the templates** — gate the anchor on the LABEL and
   default the href to the site's contact page. Cheapest, but bakes a
   site-shaped assumption into a shared template.
4. **Component lock** (what webdesign.uk has now) — site-scoped tourniquet.
   Freezes the copy in the locked components; blocked changes are at least
   recorded as `lock_blocked_change` items rather than lost silently.

## 7. How to verify any fix

Diff the invariant as a matched pair, which is the check that caught this:

```sql
SELECT p.name, pc.slot_name,
       (SELECT count(*) FROM regexp_matches(pc.rendered_html,'href="','g')) AS links
FROM page_components pc JOIN pages p ON p.id=pc.page_id
WHERE p.site_id='<site>' AND p.status='active' ORDER BY p.name, pc.position;
```

Take it before the rewrite and after. **Five other checks were green throughout
the original incident** — the claims scan, byte deltas, a retired-term grep, the
served-artefact fetch, and a link gate armed on the one page that happened not
to be affected. None of them can see a missing `href`.

## 8. Relations

`bugs_open/238` (the parent key-loss family; §9 there carries this case and §8
of it is stale) · `bugs_open/229` (what a hand-patched `rendered_html` re-arms —
the repair here had to do exactly that) · `bugs_open/058` (the component lock
this leans on) · `bugs_open/178` (`mode=edit_live`, without which the rewrite
also guts the prose) · `WRONG_CALLS.md` 2026-08-12 (why the five green checks
missed it) · `LANDMINES.md`, the `save_page_sections` REPLACES entry.
