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

## ADDENDA 2026-08-31 (~19:0xZ) — graders and a latent sibling, from the boxingonline session (attributed; evidentiary status preserved)

**Grader 1 — "the footer contains multi-byte characters" is NOT an explanation.**
32 of 32 stored footers fleet-wide contain multi-byte characters, 31 contain an
em-dash — all stored fine. What distinguishes this site is something that CUTS at a
byte offset, not something that contains one. Reject any proposed cause of that
shape unless it beats this census.

**Grader 2 — "empty sites.email" is NOT sufficient either** (and it killed that
session's own best timing-fit hypothesis, offered here so nobody re-derives it):
13 sites have an empty email and 12 have a rendered footer, one as recently as
today. The DropDeadURLControls-on-a-dead-mailto theory fits this site's TIMING
(failures start with the first render after the ~15:37 email scrub; the 13:31 bake
predates it) but is unsupported by timing alone. The sharp revival test, unrun:
whether those 12 sites' footer templates emit a mailto control at all (this site's
did; theirs may gate it out before it can go dead).

**Code read (theirs) — the obvious alignment bug is NOT there**: renderedHTML
reaches the bind unsliced; the only surgery between RenderTemplate (:1075) and the
store is DropDeadURLControls (:1227) and injectBrandHeadTags (:1261).
`ReplaceAllInMarkup` slices by offsets from a regex over a masked copy — the classic
mid-rune cut — but `maskNonMarkup` (markup_spans.go:108) masks PER BYTE, preserving
length and offsets. **One reading NOT discharged (a reading, not a finding — read it
properly, quote it never):** the span loop `for i := s.Start; i < s.End` can mask
PART of a multi-byte character if a span boundary lands mid-rune, putting a bare
continuation byte into the masked (matching-only) copy — it would have to move a
match boundary to matter.

**Latent sibling, fold into half 1's fix** — same file, same class, NOT this
footer's cause: lines 1439, 1689 and 1779 each do
`if len(summary) > 250 { summary = summary[:247] + "..." }` — a byte slice that cuts
mid-rune exactly as this file describes. They write work-item summaries — and
**:1689's summary is `fmt.Sprintf("Chrome %s failed to render: %v", slot, renderErr)`,
an arbitrary error string — so the surface that REPORTS a chrome failure can itself
mint invalid UTF-8 and fail its own insert.** Whoever makes the failure path
load-bearing (fix candidate 1) must make these three slices rune-safe in the same
pass, or the new reporting can die of the same disease it reports.

## How to verify

Re-run render_site_components for the site (the repro dispatch shape is in the
webdesign NOTES 2026-08-31 ~18:1x entry): half 1 fixed ⇒ the failure appears in
`chrome_render_failed` in collected_data; half 2 fixed ⇒ `rendered.footer=true`, the
footer row's updated_at moves, `rendered_html_digest = md5(rendered_html)`, and the
served footer still carries NO contact block (sites.email empty gates it —
component_library.go:1988) — that last probe is the pre-delivery check the
boxingonline session defined.
