# HANDOFF 2026-08-30 — copy quality is HANDED BACK to its own lane; this lane returns to the SERVICE. Start here.

**COLD-START for `finetuning_uk_service`.** Supersedes `HANDOFF_2026-08-26_continue_here.md` (keep it
for the canary detail). Technical log: `NOTES_finetuning_uk_service.md`. Owner prose:
`README_where_we_are.md` (his document — append only).

> ## ⚠ FIRST, TWO THINGS THAT WILL MISLEAD YOU
>
> **1. The kubeconfig token is EXPIRED** (measured 2026-08-30 21:54 — every `kubectl` returns
> `You must be logged in to the server`). That is the documented 3-day expiry, not a cluster fault.
> **The owner refreshes it.** Everything below marked `[MEASURED 2026-08-26]` was taken with cluster
> access and has NOT been re-verified since; the `[MEASURED 2026-08-30]` items came from the served
> site and the public bucket, which need no cluster.
>
> **2. Nothing in this file about DB state is current.** Four days passed. Re-measure before quoting
> any of it — a page count, a queue depth, a work-item status. This lane's own worst error was
> exactly that (see `bugs_open/412` §9).

## OWNER'S CURRENT DIRECTION, 2026-08-30

> *"The copy team can take on the copy quality from here, we can go back to finetuning."*

**So: STOP working copy quality.** No more register experiments, brief edits, canaries or scoring
from this lane. `copy_quality_two_stage` owns it and has the full trail
(their `HANDOFF_2026-08-26_continue_here.md`, canary doc `AUDIT_prompts/CANARY_2026-08-26_…`).
**Two things are in their court, not ours:** whether the truncation trial repeals the mild-forgiveness
for `rather_than`, and adding `instead_of`/`not_just` to the gate's shapes.

**This lane goes back to the SERVICE** — the £99 fine-tuning offer, the playground, the terms, and
Stripe.

## What the copy work established before handing over (so nobody re-runs it)

- **P2a REFUTED, and it is a real finding.** With the demonstration stack cleared (migrations
  `646`/`647` took the site brief from 7 `rather than` and 4 `not just` to 0) **and** the owner's
  positive rule live in the brief (`648`, 19:03:22Z), pages rebuilt at 20:33–20:53 still produced
  the comparison shape at **~5 per page**. Two instruments converged: this lane counted **30** across
  6 pages, the copy lane **34** on the same shapes.
  ⇒ **Neither removing the examples nor adding the instruction governs form. The carrier is the
  model's own preference.** The fix must be mechanical, post-writing.
- **The owner's rule is right**: all 30 instances truncate with **no loss of meaning**. That is on
  record as the validation for the gate's truncation approach.
- **Prose quality DID improve** — concreteness and directness are visibly better than the copy he
  rejected on 08-24/25. The tic is what did not move.

## THE SERVICE — where it actually stands

| # | thing | state |
|---|---|---|
| 1 | Live pages | `/your-own-model.html` (£99 front door) + `/technical-details.html` — facts, claims, licences verified and version-pinned |
| 2 | Terms facts | **All four registered** in `evidence_base.facts[]` (10 facts): booking 9–5 UK weekdays, deletion within a week, retention 30 days after handover, one playground hour expiring 30 days, plus the data-location sentence he approved |
| 3 | Terms/privacy PAGES | ⏳ **NOT WRITTEN.** The facts exist; extending `/terms.html` and `/privacy-policy.html` through the framework is the next real service task |
| 4 | Sample datasets | **All six BUILT** (`datasets/`). Voice sets small — 26/13/16 rows — because his corpus is 6,595 words. See `datasets/PROVENANCE.md` |
| 5 | Playground booking | Shape decided (customer picks, 9–5 UK weekdays, other by arrangement). **Not built** |
| 6 | Stripe | **LAST, and he does it himself** |
| 7 | Header nav | Done — Contact + How We Work displaced, "Your Own Model" live in the header |

## Next session, in order

1. **Ask him to refresh the kubeconfig token** — nothing DB-side can be done without it.
2. **⚡ CHEAPEST WIN AVAILABLE — deliver the 8 missing hero images.** `[MEASURED 2026-08-30]` all
   nine generated images are **deployed, public and resolving** at
   `/assets/images/content-hero-<page>.jpg` (200, 62–140 KB each), and **only `careers.html`
   references its own**. The other eight are orphaned. Fix is an `UPDATE` setting
   `content_data.hero_url` plus a `page_rerender` at `spec.reason='template_changed'` — **which
   regenerates NO copy**, so it does not reopen anything the copy lane now owns. Exact SQL and the
   caveats: `bugs_open/412` §9.
3. **Write the terms and privacy pages** from the four registered facts, through the framework.
   This is the service work he has been waiting on and it is unblocked.
4. **Then the playground booking flow** (decision 4 answered; build not started).
5. **Stripe last, and his.**

## Bugs this lane owns, all open

- **`bugs_open/398`** — `cta_bg` may be a gradient; components used it as a colour. Hero half fixed
  and live on 3 sites; **the CTA-button half is committed and inert until the next chassis roll**
  (`--color-cta-bg-ink`). ⚠ On the roll: probe the binary with a present- and an absent-control,
  then ONE `template_changed` fan-out for the CTA buttons. Council trail `f0591cb2` (round 2 REVISE;
  **the gating guardian objection was never readable** — recover it before resubmitting).
- **`bugs_open/407`** — a site cannot promote its own page into its own header; a fleet-wide
  name-tier table decides it. Filed at his direction, his fix is candidate 1. Unowned.
- **`bugs_open/412`** — a page cannot gain a hero image without a full LLM rebuild of its copy;
  §7 components that accept imagery they cannot display (fixed fleet-wide by `649`); §8/§9 the
  delivery result and its correction.

## Traps that cost this lane real time

- ⚠ **Re-measure a delivery figure before quoting it — the pipeline is slower than the session.**
  §8 recorded "zero images delivered" on 08-26; on 08-30 it was one in nine with all nine in the
  bucket. Dated and asserted while the run was still draining.
- ⚠ **A verify block only refuses what you told it to check.** `646` passed with 3 `not just` left;
  `648` pinned a count at 1 that was 2. Assert the whole class.
- ⚠ **A census answers the question you encoded.** "Which pages lack an image VALUE" is not "which
  components can DISPLAY one" — that error cost three wasted image generations.
- ⚠ **`content_direction.formatted` is DERIVED** — edit the source keys and formatted together or
  the next spec write erases you.
- ⚠ **A template edited by SQL ships nothing** without a `template_changed` fan-out.
- ⚠ **Check `page_components.updated_at` before treating an owner verdict as being about new copy.**
