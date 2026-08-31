# 423 — a chrome slot whose STORE fails is reported as a reasonless `false`: boxingonline's footer renders, Postgres refuses the bytes (invalid UTF-8, 0x80), and the action's own reason fields all stay empty

**Filed 2026-08-31 (~18:2xZ)** by the delivery-lane session. **Mechanism CAPTURED LIVE
at the artefact**, not inferred: a dedicated repro dispatch (rerender-pages, corr
`7fb750a3-f804-481e-8e8a-656c3290ecd9`) with a log monitor on every chassis pod. This
supersedes the iteration-capped 090 run `387c0a2d` (UNVERIFIABLE) on the same symptom;
the first-hand capture below is the substitute verification per the 2026-07-31 ruling.

## The captured mechanism (pod agent-chassis-6d6856d8d5-vswxv, 18:14:26Z)

```
error  render_site_components_action.go:1338  "Failed to store rendered component"
       slot=footer  error=ERROR: invalid byte sequence for encoding "UTF8": 0x80 (SQLSTATE 22021)
info   render_site_components_action.go:355   "RenderSiteComponentsAction: Complete"
       rendered={"footer":false,"head":true,...}
```

Two defects, one incident:

### Half 1 — the OBSERVABILITY defect (settled at the code + the artefact)

`renderAndStoreSiteComponent`'s store-failure branch
(render_site_components_action.go:1338-1343) logs the error and returns
`false, false, degraded, nil` — the **nil error** means the caller's
`chrome_render_failed` map never hears about it, `ineligible_chrome` and
`locked_slots_preserved` stay empty, and the step completes SUCCESS with
`rendered.footer=false` as the only trace. Measured consequence: THREE runs tonight
(15:39 wave `3f604312`, 17:47 nav-updater `07fed163`, 18:14 repro `7fb750a3`) each
declined the footer this way; two sessions spent ~an hour eliminating locks,
component eligibility, content_data emptiness (fleet control: 31/33 empty, 30
render — boxingonline session) and slot-set membership before a live log capture
named it. The action already has the exact surface for this (`chrome_render_failed`,
built for bugs_open/260) — the store-failure branch just doesn't use it.

### Half 2 — the DATA defect (mechanism named, source OPEN)

The bind that Postgres rejects is `renderedHTML` ($1 of the UPDATE at :1330), so the
invalid byte exists in the GO STRING the template pipeline produced. Every DB-sourced
input is valid UTF-8 by construction (Postgres cannot hold otherwise), so **0x80 — a
bare CONTINUATION byte, e.g. the middle of an em-dash E2 80 94 — is introduced
between template execution and the bind**: the classic signature of a byte-indexed
truncation or splice cutting a multi-byte character. `[UNVERIFIED]` which transform:
candidates include template helper funcs that slice by byte, DropDeadURLControls,
and the favicon/OG injection — none read, none accused. Site-specific: the same
statement stored 30 estate footers cleanly within the last week, and boxingonline's
own footer stored fine at its 13:31 first bake; failures began with the 15:39 wave,
i.e. after the day's data changes (email scrub, nav changes) altered the render
inputs — the corrupting input is in whatever changed for THIS site.

## Why this matters beyond one footer

- It is the third instance tonight of the 420-family shape: **a removal/refresh that
  reads as done** — here the enabling defect is half 1, which converts every store
  failure into silence.
- Consequence on the paid site: the served footer is a 16:05 hand-patch that is the
  ONLY definition of the site's footer (content_data empty by fleet norm), and the
  pipeline cannot replace it until half 2 is fixed. Recorded on the boxingonline
  pre-delivery list.

## Fix candidates

1. Half 1, small and surgical: the store-failure branch populates
   `chrome_render_failed[slot]` (return the error as `renderErr` like the execution-
   failure branch ~90 lines above — same disposition, same surface). Makes every
   future instance of half 2 loud.
2. Half 2: find the byte-slicer (start from the diff between this site's render
   inputs at 13:31 vs 15:39; or bisect by rendering the footer template against the
   live context in a test and hex-scanning the output), then make it rune-safe.
   Postgres's error does not say WHERE in the string the byte sits — a test harness
   hex-scan does.

## How to verify

Re-run render_site_components for the site (the repro dispatch shape is in the
webdesign NOTES 2026-08-31 ~18:1x entry): half 1 fixed ⇒ the failure appears in
`chrome_render_failed` in collected_data; half 2 fixed ⇒ `rendered.footer=true`, the
footer row's updated_at moves, `rendered_html_digest = md5(rendered_html)`, and the
served footer still carries NO contact block (sites.email empty gates it —
component_library.go:1988) — that last probe is the pre-delivery check the
boxingonline session defined.
