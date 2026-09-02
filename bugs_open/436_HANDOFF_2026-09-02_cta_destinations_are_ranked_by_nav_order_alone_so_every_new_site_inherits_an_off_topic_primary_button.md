# 436 — CTA destinations are ranked by `nav_order` alone, so every NEW site inherits an off-topic primary button

**Filed 2026-09-02**, spun out of **`bugs_closed/391`** on its closure (owner instruction). 391 fixed
the *damage* on three sites; this is the *cause*, which is untouched and fires on every site the
framework builds. **Status: OPEN. Severity: medium** (no live damage today — 391 cleared it — but the
next build re-creates it).

## The mechanism, read from the live code

`chooseCTATargets` (`platform/orchestration/actions/resolve_internal_links_action.go:651`) picks a
site's primary CTA by sorting every `tool`/`game` page on `COALESCE(nav_order,100)`, then `name`, and
taking `[0]`. **There is no topic, tag, vertical or semantic input at all.** Whatever sorts first wins
the primary button on every page of the site.

**And the wrong pick locks itself in.** `stampCTADestinationGuidance` (`:362`) feeds the chosen
destination's title into the writer's spec for the label field, so the framework writes button copy
*naming* whatever it picked. The next resolve label-matches that copy back to the same page —
`LoadCTALabelUniverse` runs **ahead** of the positional pick — so the row becomes unreachable by any
`nav_order` change. **Measured on 391's population: 20 of 80 fields had reached that state, including
all three the owner reported.**

## Why 391's fix does not close this

391 corrected the *data* (`nav_order` 1 → 900 on three sites) and repaired the *copy* (20 label-locked
fields rewritten, 21 contact-intent fields routed to `/contact.html`). None of that touches the
ranking. A new site with a fossil or unlucky `nav_order` — or simply an alphabetically-early tool —
gets the same off-topic primary button, and the same lock-in on the next content pass.

`[MEASURED 2026-08-25]` the fossil that caused 391 was set **at page creation, 2026-03-13**, on three
sites at once. Nothing prevents the next one.

## Owner decision 3 (2026-08-25) — approved in principle, not built

**Candidate 1 (an explicit `eligible_as_cta_target` opt-out) paired with candidate 4 (a detector for
the anomalous-`nav_order` shape).** Three constraints, all from review and all still binding:

1. **Read at the RANKING, not the loaders.** `render_site_components_action.go:182-190` — the site
   **header** CTA fallback — calls the loaders directly, takes `ordered[0]`, and its output is
   **never persisted** (`site_components` holds 0 `cta_url` keys). A loader-level change moves every
   site's header button with **no `content_data` diff to show it**.
2. **It must also bind `LoadCTALabelUniverse`**, or the opt-out has a hole exactly the shape of this
   bug — the label match runs first, so an opted-out page still wins through its own copy.
3. **Engage RFC_022 and ENUMERATE the consumers before booking a council round.** Asserting the
   opt-in shape without the query is itself the objection (owner ruling 2026-07-29 §3: a shared
   mechanism's other consumers must be *told*, not merely measured).

## Why this is architecture-scope

It adds a field to a shared seam that every site's CTA resolution reads, and it changes what the
ranking **guarantees** — the 2026-07-29 §1 trigger. Expect `needs_rfc`, and note RFC_022's narrowing:
an opt-in field whose unsafe default is OFF and which no live consumer names is **not** architecture-
scope on shape alone, but this one changes the guarantee, so it is.

## How to verify a fix

**Induce, do not wait.** Seed a site with two tools where the alphabetically/`nav_order`-first one is
off-topic for a given page, run the resolve, and assert the primary CTA is **not** it — then flip the
opt-out and assert it **is** eligible again. Both directions in one run. And assert the **header**
fallback separately (constraint 1): it is a different call site whose output is never persisted, so a
`content_data` check cannot see it.

## Relations

- `bugs_closed/391` — the damage this caused, fixed; its lane docs carry the full evidence.
- `bugs_open/248` (`cta_recompute_clobbers_authored_contact_links`) — KEEP #1, the mechanism that
  makes an authored `/contact.html` durable; 391's contact-intent fix rests on it.
- `bugs_open/399` — records label/destination disagreement at write time. ⚠ **Structurally blind to
  this bug**: when the framework picks the destination *and* names it in the copy, the two agree, so
  the judge says nothing. Their own `TestJudgeCTALabelIsBlindToTheLabelLockedDefect` pins it.
- `bugs_open/384` — the stale-listing family; holds 391's retraction residue.
- Lane: `docs/agent_docs/docs024_key_docs_latest/bugfix_389_cta_relevance/`.
