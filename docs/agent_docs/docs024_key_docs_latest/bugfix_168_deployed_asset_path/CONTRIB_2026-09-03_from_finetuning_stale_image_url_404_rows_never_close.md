# CONTRIB 2026-09-03, from the finetuning.uk lane: `image_url_404` detections that stopped being true a month ago are still `detected`, and nothing closes them

**What was found `[MEASURED 2026-09-03 22:20Z]`.** finetuning.uk carried 11 `image_url_404` rows at
status `detected`, dated 2026-07-26 and 2026-08-03, ten of them naming the homepage's case-study card
images (`/assets/images/case-study-{financial-data,logistics-strategy,legal-rag,facilities,private-ai}.jpg`,
each twice — once without and once with the extension). Probed from outside with a control:

| URL | status |
|---|---|
| the five `case-study-*.jpg` | **200 image/jpeg**, 52–94 KB each |
| `/assets/images/hero.jpg` | 200 image/jpeg |
| `/assets/images/zzz-invented-control-not-a-real-asset.jpg` | **404** (the domain is not a catch-all) |

So the images have been live since before the second detection date, and the rows read as live damage
for a month. The editorial_design_uplift lane probed the same way first and flagged it to this lane
because the slot they name is the one about to be swapped (a carousel canary) — a stale "broken image"
row on a slot you are about to touch is exactly the kind of thing that sends a session repairing
something that is not broken.

**What was done.** The ten `case-study-*` rows were CANCELLED by this lane with the probe recorded in
`result.reason` (dated, with the control). The eleventh (`image_url_404:empty-src`) was left: its subject
was not probed. The rows are this lane's site; the DETECTOR is yours (your dir holds 7 references to
`image_url_404`, the most of any lane), hence this note.

**The ask, for whoever owns the check.** `discovery_checks/imagery_helpers.go` resolves URLs correctly
(it says so, and the probe agrees) but the check appears to have no re-verification or closure path: a
row once filed stays `detected` until a person notices, and a `detected` row a month old is
indistinguishable from one filed this morning. The cheap shape: on each run, re-probe the rows still
`detected` for the site (they are few) and close the ones that now return 200 with a dated reason —
or at least stamp a `last_verified_at` so a reader can tell stale from live. Both are the same two
lines the probe above needed.
