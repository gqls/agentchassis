# 471 — the dead-form detector matches a LIST of known-bad actions, so a plausible-looking one is never probed — and there is at least one live instance

**Filed** 2026-09-03 by lane `static_site_form_endpoint`, found while censusing fleet form shapes
for the form-endpoint build. Class: **a detector whose predicate is an enumeration of known-bad
values, so the defect it exists to catch is invisible whenever it wears an unfamiliar name.**
Sibling of `bugs_open/228` (a form that reports success over a `setTimeout`) — same visitor-facing
outcome, different mechanism, and neither is caught by the other's check.

## Symptom, at the artefact

`https://gamesdesign.co.uk/premium.html` serves a form whose action is `/request`. Nothing answers
it. A visitor fills it in, presses send, and the message is gone.

```
### gamesdesign.co.uk — NOT a vm-sites row, so no application exists that could answer
  /premium.html            -> 200   control B (known-good sibling; the site is up)
  /invented-1788448171     -> 404   control A (not a catch-all; a 404 means something)
  /request                 -> 404   TARGET — no handler
```

**The controls are what make the 404 mean "no handler" rather than "site down" or "everything
404s".** And the method is proven discriminating on the same day, against two forms that *are*
wired — both on `vm-sites` rows, where an application does exist:

| domain | vm-sites? | target | result |
|---|---|---|---|
| `idea.uk` | yes | `/request` | **405** — a real POST handler |
| `relojistas.com` | yes | `/intent` | **405** — a real POST handler |
| `gamesdesign.co.uk` | **no** | `/request` | **404** — nothing |

A 405 is Method Not Allowed: the route exists and rejects GET. That is the positive control. Recipe
with both controls: `RUNBOOK_static_site_form_endpoint.md` §2.

## Root cause `[VERIFIED — read at the two deciding arms, quoted below]`

Two mirrored enumerations decide what "delivers" means, and **neither ever asks whether anything
answers.**

**Arm 1 — the render seam.** `platform/orchestration/actions/component_library.go:1495`:

```go
func deliverableFormAction(current string, ctx *RenderContext) (string, bool) {
	if !nonDeliveringFormActions[strings.ToLower(strings.TrimSpace(current))] {
		return "", false // already points somewhere real (a mailto:, a live handler)
	}
```

against the map at `:1448`, whose entire contents are `""`, `"#"`, `"#contact"`, `"#contact-form"`,
`"/contact"`. An action absent from that map returns on the first line. **The comment is the bug**:
"already points somewhere real" is asserted, never tested, and `/request` is exactly a value that
looks real and is not.

**Arm 2 — the detector.** `discovery_checks/check_contact_form_undeliverable.go:99-104` filters in
SQL, before any Go sees a row:

```sql
AND ( pc.rendered_html !~* 'action="[^"]+"'      -- no action at all
   OR pc.rendered_html ~* 'action="\s*"'         -- empty
   OR pc.rendered_html ~* 'action="#[^"]*"'      -- self-anchor
   OR pc.rendered_html ~* 'action="/contact"' )  -- never existed
```

So a `/request` form never enters `findings`. `resolveSiteContact` / `contactAddressResolvable`
run **after** `if len(findings) == 0 { return }` and only choose the *routing* of rows that already
matched — they test the SITE's address, not the action's target. They are not a second chance.

**The two lists must agree, and their agreement is the trap.** The detector cannot see what the
seam declares deliverable, because both are the same enumeration written twice. Adding `/request`
to both fixes this one site and reproduces the class at the next unfamiliar spelling.

## Scope `[MEASURED 2026-09-03 — a count, so it carries its date]`

Fleet-wide, by served form shape (`page_components.rendered_html`, any `<form>`):

| shape | components | sites |
|---|---|---|
| form with **no action attribute at all** | 145 | 37 |
| `mailto:` | 26 | 21 |
| self-anchor `#` | 6 | 6 |
| **other action value** | **5** | **3** |

The 5 "other" are the population this bug is about, and only they were probed: `idea.uk`
`/audience-check` + `/request` (live), `relojistas.com` `/intent` (live), `gamesdesign.co.uk`
`/request` (**dead**) and one `action=""` on `gamesdesign.co.uk/contact/index.html` (already
covered by the existing check's empty-string arm).

**So the confirmed damage is one page.** `[UNMEASURED]` The 145 no-action forms across 37 sites are
NOT included in that claim: the existing check deliberately scopes to
`data-component="contact-form"` because tool calculators bind submit from an external JS bundle,
and judging them would file work items against working tools. Whether every one of those 145 is a
tool form is not established here, and establishing it is a separate piece of work — the honest
statement is that this bug confirms one instance and shows the predicate cannot bound the class.

## Why this was filed as a structural claim, and what the diagnosis loop said

Filed through `090` first, per the owner ruling of 2026-07-31, because the claim is about a
mechanism rather than a page. Run correlation `ef3049cc-c84d-4857-af8e-9c10d059b1b6`.

**The loop returned `UNVERIFIABLE` — "NOT CONFIRMED (stopped: scope-not-narrowing)" — and it was
right to.** Its own words on what it could not settle:

> none of Run, deliverableFormAction, classifyUndeliverableAction, or contactAddressResolvable's
> bodies are in the bundle, so it is unknown whether an action absent from
> nonDeliveringFormActions … is (a) simply treated as deliverable by falling through the map check
> with no further test, matching the hypothesis, or (b) additionally routed through
> contactAddressResolvable/resolveSiteContact … which would refute the 'never tested' claim.

It named the exact discriminating question and handed it to a human rather than guessing. **This
file is that human answer**: (a) is what happens, established by reading both arms and quoting the
deciding lines above — the fall-through is unconditional, and `contactAddressResolvable` is
unreachable for a non-listed action in the seam and post-filter in the check. An `UNVERIFIABLE` is
not a refutation; it is the loop declining to conclude without the bodies.

## Fix candidates, ordered by what makes the bad state unrepresentable

1. **Probe the action target instead of matching a list.** For any form action that is a relative
   path or an absolute URL, ask whether anything answers it; treat a 404 as undeliverable and a
   405/200/3xx as live. `platform/fetchguard` exists for outbound fetches from the estate and is
   the reuse. **Prefer this**: it is the only candidate where a *new* unfamiliar spelling cannot
   reopen the hole, because nothing depends on having enumerated it.
   Cost to bound honestly: one HTTP request per distinct (domain, action) pair per run, cacheable
   within a run, and it must not fire on `mailto:`/`tel:` (nothing to probe) or on tool forms
   (out of scope by the existing boundary).
2. **Add a "relative action on a site with no application" rule**, using `sites.github_repo` /
   `publish_target` to decide whether a handler could exist at all. Cheap, no network, and catches
   the `gamesdesign` shape exactly — but it is a second enumeration (of hosting modes this time)
   and will not catch a dead path on a site that *does* run an application.
3. **Add `/request` and friends to both lists.** Rejected as the whole fix: it repairs one site and
   re-creates the class at the next spelling. Reasonable only as a same-day stopgap alongside (1).

Whichever is taken, **the two enumerations must stop being two.** If a list survives at all, the
seam and the check should read one exported source, the way the estate already pins
`contactAddressResolvable`'s parity with tests on both sides.

## How to verify a fix

1. Re-run the check against `gamesdesign.co.uk` and confirm `/premium.html` is now reported. It is
   the motivating case, so a fix that does not flag it has not been tested against what prompted it.
2. **Re-run it over what the blind predicate already cleared** — every site the check has passed
   while unable to see this shape. A pass from a blind check outlives the blindness, and closed
   files quote it as "the page is fine".
3. Confirm the two live forms are NOT flagged: `idea.uk/request` and `relojistas.com/intent` must
   stay clean, or the fix has traded a false negative for a false positive against working sites.
4. Confirm no tool-calculator page is flagged — the existing `data-component="contact-form"`
   boundary exists because an earlier draft pulled in 6 tool pages.

## Relationship to the form endpoint

This lane is building a real receiver (`static_site_form_endpoint/PLAN_2026-09-03_form_endpoint_build.md`,
migrations 756 + 757), which gives the seam a third destination and gives address-less sites
somewhere to point. **It does not fix this bug**: an endpoint changes what a form *can* point at,
while this is about the platform being unable to tell whether a form's existing target answers.
The two touch the same function and should not be conflated — the endpoint work will add a branch
to `deliverableFormAction`, and it will sit directly above the unconditional fall-through quoted
above, which will still be there.
