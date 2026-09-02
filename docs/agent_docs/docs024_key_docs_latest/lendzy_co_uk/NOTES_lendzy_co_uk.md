# NOTES — lendzy.co.uk lane

Running record, append-only, **newest at the bottom**. Missteps are the point, not an appendix.

Site: `lendzy.co.uk` = `8ff093d5-1f19-453b-9439-a10379bbcd76`.
**Counts carry the date they were counted** (owner ruling 2026-08-22).

---

## 2026-09-02 (a) — the lane is created, and what it inherited

Created today at the owner's instruction ("this is lendzy's own lane now"). Until now lendzy
had **no lane**: it was built by `portfolio_positioning` as the framework's first end-to-end
shadow build (seeded 2026-08-02), and four lanes have since worked pieces of it and handed
them off:

- `bugfix_414_planted_marker_as_claim` — the planted FCA marker. **CLOSED 08-31** (`de99599fb`),
  fixed and live and verified. Re-checked today: **0** components carry the phrase.
- `copy_quality_two_stage` — holds lendzy's copy. Their standing read: the adversarial-adjacent
  frame is EARNED, leave it. Two residuals of theirs: `key_differentiators` shared verbatim with
  `loancash.co.uk` (a copy-paste at adoption), and 5 `voice_tells` items filed 09-01, unreviewed.
- `dispatch_throughput` — used lendzy as the fleet's worst-starvation baseline (55 eligible,
  oldest 10.6h, pinned rank 44). Migrations 657/658 fixed the ordering; lendzy drained 46 → 15.
  Measurement subject, not a defect. Closed for us.
- `architecture_review/RFC_060` — filed by the 414 lane 09-02, owner-decided the same morning.
  Lendzy is the worked example of the top `relied_upon` rung.

## 2026-09-02 (b) — three tool pages serve 200 and are recorded as never built

Found while grounding the lane. `[MEASURED 2026-09-02]`

| page | serves | `build_status` | `deployed_at` |
|---|---|---|---|
| `tool-price-cap-checker` | **200**, 65,356 B, 3 `<input>` | `needs_rebuild` | **NULL** |
| `tool-true-cost-calculator` | **200**, 63,999 B, 1 `<input>` | `needs_rebuild` | **NULL** |
| `tool-complaint-deadline-calculator` | **200**, 63,997 B, 2 `<input>` | `needs_rebuild` | **NULL** |

Six sibling tool pages are `deployed` with stamps written 09-01 10:58–11:03Z. All nine serve;
all five probed carry a correct `rel=canonical`. **Invented-URL control 404s (9 B)**, so the
200s are real pages and not a parked-domain catch-all.

Two measured consequences, both from `deployed_at IS NULL`:
- **Missing from the sitemap.** 30 active pages, **27** `<loc>`. The three missing are exactly
  these. `render_sitemap_action.go:144` filters `deployed_at IS NOT NULL`.
- **47 `unbuilt_internal_link` items** at `needs_human_review`, all filed 2026-09-01, every one
  naming one of these three URLs, every one reading "points at a page that has never been
  deployed". Nothing is queued to rebuild them, so the queue cannot drain itself.

## 2026-09-02 (c) — the root cause, and the two hypotheses that died first

**Dead hypothesis 1: the deploy-skip guard refused the stamp.** `refuseDeployStampOnSkip`
(`page_build_failure_guard.go:78`) flips a page to `needs_rebuild` when the deploy step reports
`skipped`, which fitted the symptom well. **Refuted two ways.** (i) `agent_error_log` holds
**zero** `DEPLOY_STAMP_REFUSED_ON_SKIP` rows for this site — lendzy's only rows since 09-01 are
three `CTA_LABEL_MISMATCH`. (ii) The guard is opt-in on `deploy_result_field` and its own header
says the unarmed path is "every live step today". It never ran. *The lesson I nearly took: the
code that best fits a symptom is not evidence that it executed. Ask the log whether it fired.*

**Dead hypothesis 2: NULL `component_id` is the discriminator, fleet-wide.** True on lendzy
(3/3 stuck pages have one, 0/6 working ones do) and **false as stated**: the fleet carries
**16** such rows across **7** sites and **10** pages, and outside lendzy **none** is stuck.
Correlation inside one site is not a mechanism. The corrected predicate is the one below, and
it was chosen because it could have come out otherwise.

**The cause, and it is exact.** The discriminating property is not "has a NULL `component_id`"
but "**every** component row on the page has one — so there is nothing resolvable to build
from". Fleet-wide, active pages where no component row carries a `component_id`:

```
 lendzy.co.uk | tool-complaint-deadline-calculator | needs_rebuild | unstamped
 lendzy.co.uk | tool-price-cap-checker             | needs_rebuild | unstamped
 lendzy.co.uk | tool-true-cost-calculator          | needs_rebuild | unstamped
(3 rows)          [MEASURED 2026-09-02]
```

**Three of three, no counter-examples anywhere in the estate.** The query could have returned
healthy pages, or pages on other sites; it returned neither.

The chain, read at the deciding arm rather than inferred:

1. Each page has **one** `page_components` row, written **2026-08-02** (the original shadow
   build), with `component_id = NULL` and `slot_name = 'section'`.
2. `rerender_page_sections_action.go:361` `resolveComponent` resolves a section by
   `component_id` first, then by `slot_name` against a component name/function map. The id is
   empty (`COALESCE(component_id::text,'')`, line 924) and **no component is named, functioned
   or typed `section`** — verified, **0 rows** in `content_components`. Neither route resolves.
3. The slot lands in `UnresolvedSlots`. `rerenderResolution.fatal()` (line 650) counts that as
   fatal, and line 600 returns an error: *"N of N section(s) could not resolve a component"*.
4. So the page never reaches `build_status='deployed'`, and `UpdatePageStatusAction`
   (`v3_site_actions.go:1082`) only stamps `deployed_at` inside its `newStatus == "deployed"`
   branch. No status, no stamp.
5. The 2026-08-02 artefact is still in the bucket, so the URL serves 200 for ever. Every check
   that asks the **artefact** says healthy; every check that asks the **record** says never
   built. Both are reading correctly.
6. It cannot self-heal: `needs_rebuild` re-selects the page, the render fails identically, and
   it re-files. Six `page_rerender` items for these pages since 08-25 — **all `complete`**.

This is the residue of `bugs_closed/182`, not a regression of it: 182's fix made a silent
carry loud, and the loudness has no consumer for this shape.

**Filed to the diagnosis loop before asserting it durably** (owner ruling 2026-07-31, and this
is a cross-cutting mechanism claim): intake `1ff4c475-6977-4631-b641-993735429186`, run
correlation `89a84ad3-5668-44b3-a089-f9d6c0df7cbb`. Verdict to be recorded here when it lands
— **including if it refutes me**, which is the cheapest place to be wrong.

## 2026-09-02 (d) — the FCA ask, and the one thing measured so far

Owner: make the retracted claim TRUE — check all financial facts against the FCA Handbook line
by line, with a local mirrored copy kept current AND live re-checking ("probably as well").

Measured today, nothing else assumed: `https://www.handbook.fca.org.uk/handbook/CONC/5A/` 301s
to `https://handbook.fca.org.uk/handbook/conc5a` and returns **200, 477,729 B**, title
*"FCA Handbook - CONC 5A Cost cap for high-cost short-term credit"*, with rule identifiers in
the markup down to `CONC 5A.1.1`. No auth. So per-rule citation with a verbatim quote looks
mechanisable. `[MEASURED 2026-09-02]` **Everything else about the corpus is `[UNMEASURED]`** —
licensing/terms, rate limits, whether an instrument or release feed exists for change
detection, and how the page markup keys rules to text.

Lendzy has **no `evidence_base` at all** — one of the five register-less finance sites named in
`RFC_060` §1, so its numeric scan never arms. RFC_060 §4: *"If only one thing happens as a
result of this RFC"*, it should be populating those registers.

Wrote to the `claims verification` session (peer `d02867`) before designing anything, since the
owner named them as responsible here: asked what they own, whether the `evidence-refresher` is
theirs and is the right substrate at handbook scale, whether anyone already mirrors an external
regulatory corpus, and where they want the boundary between per-site register work and platform
verification. **Design held until they reply** — building a second spelling of their mechanism
is the failure mode to avoid.
