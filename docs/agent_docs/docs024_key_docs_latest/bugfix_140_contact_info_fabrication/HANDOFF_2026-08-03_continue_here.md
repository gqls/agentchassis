# HANDOFF 2026-08-03 — bugfix 140 + RFC_009 · **everything is DONE and LIVE**

**Read this first. There is no work in flight and nothing is half-finished.**
This is a cold-start doc for whoever picks the thread up, not a to-do list.

---

## One-line state

`bugs_open/140` → **CLOSED** (`bugs_closed/140_…`), fixed at source and proven on
all 8 sites stored *and* served. `RFC_009` → **decided**: A not taken, **B and C
both shipped, council-APPROVED, and LIVE on v1.0.1237**. The only open item in the
whole thread is a **judgement call for the owner**, described at the foot.

## What was wrong (30 seconds)

The shared `contact-info` component rendered Email/Phone/Hours cards
unconditionally and, when a site had not supplied the datum, **invented one** —
`+1234567890` in a `tel:` link, `Monday – Friday, 9am – 6pm`, `info@example.com` —
styled identically to real details. **8 live commercial sites** served the invented
hours; `vetcomparison.uk` served the invented phone.

**The organising fact:** the component's own `input_schema` already declared
`"on_missing": "skip_field"` for phone/hours/address. **The template disobeyed its
own published contract**, so this was never a policy question needing an owner —
it was making the template obey a rule it already published.

## What shipped, and where it lives

| | what | where | state |
|---|---|---|---|
| **fix** | migration 287 — gate every card, delete the 4 literals, repair a template/schema desync | `sql_for_agents/287_contact_info_obeys_its_own_schema.sql` | LIVE (config), 8/8 sites clean |
| **detector** | `check_placeholder_contact` taught the dummies our own library ships | `discovery_checks/check_placeholder_contact.go` | LIVE v1.0.1233 |
| **C — lint** | daily CronJob, reads the LIVE library, fact-vs-label | `deployments/kustomize/services/component-fallback-check/` | LIVE, **fired unattended 06:40** |
| **B — gate** | refuses a fabricating template at the write | `platform/orchestration/actions/component_fallback_guard.go` | LIVE v1.0.1237 |

Commits: `673b2556c` `4cc7da377` `8666a83a8` `4e06cf92d` `395246bb5` `3c5cd09ca`
`f2a59047f` `194f0c5f0` `249df5940` `87ea0a5e7` `d3d673856` `471b0f922` `f48bf3e60`.
Councils: **`40de12b0-…` APPROVED r1** (the fix), **`19bee790-…` APPROVED r1** (B).

## The five things worth knowing before you touch any of it

1. **A rerender does NOT regenerate sections unless `spec.reason` is set.**
   `check_rerender_mode` routes `image_landed|section_data_resolved|cta_links_stale`
   → `rerender_page_sections` (regenerates from template); **everything else** →
   `render_page`, which re-staples the page from section HTML **already stored**.
   A reason-less item is the fleet default. **I got this wrong in the middle of
   this work**: I said the pages would self-correct once the stalled queue drained;
   it drained 294→0, six contact pages rerendered, and every one came back with the
   fabrication intact. Fixed by queuing 7 items carrying the reason.
   *Do not read `create_rerender_items_action.go:219`'s `&& componentIDStr != ""`
   as the consumer's rule — it is producer-side; the agent needs only the reason.*

2. **`page_component_history.source` tells you what WROTE a section.** Do not infer
   it from whatever work item completed nearby. vetcomparison corrected via
   `save_page_sections_overwrite` (the content-writer), **not** a rerender, and
   mis-crediting that is what made "the others will follow" sound safe.

3. **The roster-free detector is UNSOUND.** "Flag any rendered contact fact absent
   from `content_data`" looks like the obvious upgrade and over-fires:
   `RenderContext` carries top-level `Email`/`Phone` whose json tags reach the
   template contract, so a component can legitimately render a phone its
   `content_data` lacks. **`idea.uk` is exactly that shape.**

4. **The write path does NOT close the door, and nothing should imply it does.**
   ~10 writers touch `html_template`; **two are gated** (`store_generated_component`
   absolute, `update_component_html` comparative). `create_tool_component`,
   `deploy_tool_action`, four `fix_*` style repairs and the admin handler are not.
   The daily lint is the backstop because it reads the live library.
   **Gate where it is sound, report everywhere.**

5. **The rule has TWO implementations on purpose** (Go gate, Python lint) — the
   drift class the rule itself detects. Pinned to ONE shared fixture, 24 cases,
   10 must-refuse / 14 must-allow, read by `component_fallback_guard_test.go` and
   by `--selftest`, which the CronJob runs **before** `--report`. If you change a
   pattern, change the fixture; one side will fail if you don't.

## How to check it is still healthy (all fast, all read-only)

```bash
python3 scripts/check_placeholder_fallbacks.py            # expect: CLEAN — 0 fabricated, 0 ungated (68 -> 0 on 08-03, migration 295)
                                                          # exit 2 = kubectl stream flake, NOT a pass — retry
python3 scripts/check_placeholder_fallbacks.py --selftest  # expect: 10 must-refuse, 14 must-allow
go test ./platform/orchestration/actions/ -run TestFabricatedFallback
kubectl get cronjob component-fallback-check -n ai-persona-system   # LASTSUCCESS should be today
```
```sql
-- the artefact. Expect 0 of 8.
SELECT count(*) FILTER (WHERE pc.rendered_html LIKE '%Monday%9am%6pm%'), count(*)
FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id
WHERE cc.function='contact-info';
-- the daily report, clean or not, one row per run
SELECT created_at, left(body,80) FROM doc_notes
 WHERE source='check_placeholder_fallbacks' ORDER BY created_at DESC LIMIT 3;
```

Full command set with the gotcha attached to each: `RUNBOOK_contact_info_fabrication.md`
(**R4 is SUPERSEDED — use R9**). Missteps: `NOTES_…`. Plain-prose history:
`README_where_we_are.md`. Milestone read-out: `SUMMARY_2026-08-02_…`.

## ~~THE ONE OPEN ITEM~~ — **CLOSED 2026-08-03: the owner asked, and the 68 are gated**

> **SUPERSEDED 2026-08-03 ~12:00Z.** The owner's answer to the judgement call below
> was "please gate the 68 blank-rendering fields". Done and live: **migration 295**
> (`295_twenty_components_gate_their_declared_skip_fields.sql`, commit `f2c8c6b41`).
> The live lint now reports `clean — 173 active components`, where it reported
> `68 ungated`. **The expected lint output in the health-check section above is
> therefore `0 fabricated, 0 ungated`, not `0 fabricated, 68 ungated`.**
>
> **What the next thread most needs to know**, because this doc's own description of
> the 68 was misleading: **62 of the 68 were the ungated PARTNER of a gated field**
> — `{{if .spec_1_name}}<tr><th>{{.spec_1_name}}</th><td>{{.spec_1_value}}</td></tr>{{end}}`,
> where the row is gated on the NAME and the VALUE is what declares `skip_field`. So
> the obvious repair is wrong two ways, **both of which pass the lint**: gating the
> value inside its `<td>` still emits `<td></td>` (a NO-OP that permanently silences
> the finding), and gating the `<td>` emits a 3-cell row in a 4-column table. Four
> treatments were needed; the per-component reasoning is in migration 295's header
> and the prospective check is now in `LANDMINES.md`.
>
> **Two of the 68 were not mild blanks at all**: `featured_article.featured_image`
> rendered `<img src="">` (a broken image — the `bugs_open/018` dead-control class),
> and `hero-tool`'s two CTA labels rendered `<a href="/x"></a>`, an invisible
> unclickable control. Both now gate on url AND label.
>
> **Blast radius, at the artefact rather than in the data**: 20 of the 68 fields have
> ZERO live instances; only **3 stored rows** carried the empty element. The
> "75 field-instances absent from `content_data`" figure is forward-looking RISK, not
> damage — 46 of the 47 `hero.subheadline` rows serve real text from a legacy render
> over an empty `content_data`. The two differ by 25×; do not quote the first as the
> second.
>
> **Proven**: all 20 templates re-fetched from the live library *after* the migration
> and rendered through `actions.RenderTemplate` (the production path, not a replica),
> datum present and absent — 20/20, positive control included.
>
> **NOT council-reviewed, and it could not be**: the trigger refuses a submission
> touching no `platform/`/`internal/`/`pkg/` path, and this is a `docs/` migration.
> (`f2c8c6b41` carries a meaningless `Council-Submitted: pending` trailer — my error,
> logged in `WRONG_CALLS.md`; forward-only forbids the amend.)

**RFC_009's option A is still NOT taken and this does not revive it** — ~90% of
fields (1,938 of 2,163) declare no `on_missing` at all, so a render-time gate would
be inert for nine fields in ten while being the only option that can break a live
page. Revisit if the declaration rate rises, or if a third *fabrication* slips the
lint. Neither is true today.

**The one thing now genuinely open — and it is a smaller call than the last one.**
The lint's ungated class deliberately does not fail the exit code, because "68 of
them predate this check and a permanently-red gate is one everybody learns to
ignore". **That reasoning has expired — the count is 0.** Flipping it to exit 1
closes the door behind this work; the cost is that a legitimately new component from
another lane could turn the daily CronJob red until someone adds one `{{if}}`. Not
done unasked, because it changes what an existing check does to other lanes' work.

## Explicitly NOT owed

- No pod-grep outstanding — done on v1.0.1233 (C's detector) and v1.0.1237 (B's
  guard), both replicas, with positive **and** negative controls each time.
- No council verdict outstanding — both approved at round 1, objections answered
  with checks and recorded in `RFC_009` and the closed bug file.
- No rerenders outstanding — all 8 sites verified clean on the wire.
