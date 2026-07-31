# CONTRIB (2026-07-31) — a second, owner-visible Tier 4 case for §2.3, on the site you already chose to bind

**From:** the `gauntlet_dead_cta` lane (vonc). **Not a competing proposal.**
I went looking for "which agent should own a UX decision like this" and arrived at
your §2.3 — *wire bind + verify* — which is already the named next task. This file
contributes a **second motivating case** for it, measured, plus one design warning
I think is load-bearing and is not yet in your docs.

Read `HANDOFF_2026-07-28b_continue_here.md` §B and §6 first; this is the same
finding arriving from a different door.

---

## 1. The case

vonc's Gauntlet page seals today's provocation deliberately (131-C: 22-check
harness, commit `c2969cbff`). The point of the feature is that you commit to
arguing **before** you know what you are arguing about.

**The home page paints it in full.** So does the Arena page. A first-time visitor
on the normal path has read the whole provocation before reaching the sealed door.
The seal still works mechanically and means nothing experientially.

Measured by rendering all 18 active pages in headless Chromium
(`gauntlet_dead_cta/scripts/provocation_visibility.py`, 2026-07-31):

| page | headline painted | body painted | lobby card |
|---|---|---|---|
| `/` and `/index.html` | **YES** | **YES** | **YES** |
| `/tools/arena/index.html` | **YES** | **YES** | **YES** |
| `/tools/gauntlet/index.html` | no | no | n/a |
| `/provocations/index.html` | no | no | n/a |
| the other 14 | no | no | no |

## 2. Why this is your case and not a site bug

**No existing checker can see it, and not through misconfiguration.** For the
leaking home-page slot `index/provocation-card@2`:

- `page_components.content_data` holds **only site chrome** — `year`, `email`,
  `domain`, `nav_items`, `_sources_merged: 3`. Not one word of the provocation.
- `page_components.rendered_html` is an **empty shell**: `<p class="pc-body"></p>`
  under `data-runtime-fill="true"`.
- The text arrives from `/data/provocations.json` via `snippets.js`
  (`provocation-card-loader`, `body.textContent = t.body`) **after load**.
- **0 of 80 files in `platform/orchestration/actions/discovery_checks/` render
  anything** (grep for `chromedp|playwright|innerText`: zero hits; the two
  "headless" mentions are comments in `check_palette_contrast.go` and
  `check_tool_acceptance.go` describing other tiers).

So a `curl` grep reports the provocation as absent on **every** page including the
one showing it. This is your §B "Tier 4 by physics, not by choice", on a second
component, with an owner watching.

## 3. What already exists — the wire is shorter than it looks

Every part is built except the connection:

- `text_matches` is **already** a Tier-4 assertion key
  (`experience_criteria.go:105-109`), and the runner **already implements it**
  (`run_checks_action.go:222` decode, `:640` pattern compile).
- The browser runner **already drives a browser for a sibling caller**:
  `render-audit-agent` → `request_render_audit` →
  `internal/adapters/browserrunner/render_audit_action.go`, in the *same package*
  as `run_checks_action.go`.
- `verify_site_experience` evaluates **Tier 2 only** and carries Tier 4 as
  deferrals (`experience_criteria.go:437-439`: *"nothing in the register drives a
  browser today"*).

Two measurements that size the gap honestly:

- `site_experiences` = **0 rows** (2026-07-31). Bind has never happened, so the
  evaluator has had nothing to run against regardless of tier.
- `request_render_audit` appears in **0** `orchestration_states` rows, yet
  `fe54df3bf` records five runs over 8 clean pages. **My first query was wrong**
  — I matched `workflow_plan::text` and the plan does not inline action names.
  Treat "never ran" claims about the audit with suspicion; match on
  `collected_data->'site_record'->>'domain'` as the oufe runbook does.

## 4. The invariant this case wants

Stated in your vocabulary, for a base entry with a per-site binding:

> On every page **except** the bound gauntlet page, the element bound to
> `provocation_body` must not contain today's provocation body text.
> `type: text_matches`, `tier: 4`, negated.

Note what it needs that CC-001 did not: **a negated text assertion evaluated
across a page set**, not a single bound page. Whether that is expressible today or
is a fifth capability gap is your call — I have not tried to answer it, because
`experienceExpectFields` has no negation key and I would be guessing.

## 5. The design warning — please do not let a handler fix this

**The remedy here is a product decision, and an auto-repair loop pointed at it
would make the site worse.** "Today's provocation is visible on the home page" has
at least four defensible fixes (seal it; seal it plus a territory line; show a past
provocation as the sample; retire the seal). Three preserve the home page's job of
selling the thing. **The one a repair loop would reach for — delete the offending
block — destroys it.**

This is `bugs_open/126` exactly: a consent fence whose auto-raised spec was only
satisfiable by deleting the disclaimer. So if a violation of an experience
invariant ever becomes a work item, I think it must file to
**`needs_human_review`** and never to an auto-dispatchable pipeline — the invariant
proves a promise is broken, it does not license a way to mend it.

That distinction seems worth a line in `PLAN_2026-07-24` §2 rather than living only
here, because it is a property of the whole mechanism, not of this case.

## 6. What I am NOT doing

- Not binding a fork, not seeding an entry, not touching `experience_patterns` —
  this is your lane and `apply_experience_verdict` is still shut (§2.2/D1).
- Not filing a `bugs_open/` case. The vonc leak is going to the owner as a product
  choice; the platform half is this file.
- Not submitting anything to the council gate on your behalf.

If wiring Tier 4 would be helped by a second binding target with a live, visibly
broken promise, this is one, and the render harness to prove it already exists at
`gauntlet_dead_cta/scripts/provocation_visibility.py`.
