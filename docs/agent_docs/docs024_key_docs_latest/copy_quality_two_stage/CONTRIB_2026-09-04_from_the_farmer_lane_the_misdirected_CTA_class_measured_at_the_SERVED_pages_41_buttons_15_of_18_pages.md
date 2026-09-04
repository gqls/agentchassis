# CONTRIB 2026-09-04 — from the farmerinsurance.uk lane: your misdirected-CTA class, measured at the SERVED pages — 41 buttons on 15 of 18 pages

farmerinsurance.uk got its own lane today (owner instruction; `docs024_key_docs_latest/farmerinsurance_uk/`).
Its first artefact census turned up your class, and the served-side number is worth having next
to your stored-field one. **Nothing has been touched** — the parcel is yours, this is evidence.

## What I measured, and how it differs from your 52

`[MEASURED 2026-09-04, curl of all 18 active pages, anchor labels parsed out of the served HTML]`

- **41 anchors on 15 of the 18 pages** carry a label naming one of the seven tools culled on
  08-31, and every one points at a *different, live* page.
- **41 prose mentions across 9 pages** name a culled tool outside any anchor.
- Two tools account for nearly all of it: **Farm Insurance Needs Checker** and **Livestock Value
  Estimator**.

Verbatim examples:

| page | label as served | href as served |
|---|---|---|
| `/contact.html` | "Check what cover your farm might need with the Farm Insurance Needs Checker" | `/blog/farm-machinery-insurance.html` |
| `/blog/crop-insurance.html` | "Try the Farm Insurance Needs Checker" | `/guides/index.html` |
| `/legal.html` | "Try the Livestock Value Estimator" | `/blog/livestock-insurance.html` |
| `/blog/farm-insurance-claims.html` | "See how sums insured are worked out with the Livestock Value Estimator" | `/blog/livestock-insurance.html` |

**Your 52 and my 41 are different populations, not a disagreement.** Yours counts stored CTA
fields; mine counts anchors a visitor can click today. Both are true; the served count is the one
that answers "what does a reader meet", and it also tells you how much of the stored damage has
already reached the artefact.

## The part I think is useful to your ruling rather than just to your backlog

**This class is invisible to every link checker in the estate, by construction.** The cull's CTA
recompute did its job — it re-pointed every href at a page that exists — so all 41 return HTTP
200. My own crawl of the site the same morning found 27 of 27 internal targets healthy and was
blind to all 41. A link checker asks *does this go somewhere*; the defect is *it goes somewhere
that is not what the button says*. So no 404 sweep, no `unbuilt_internal_link` producer and no
`dead_internal_link_live` check will ever raise one of these — which is presumably why it has
survived four days on a site under active repair.

If you want a detector shape out of it: the check is **label-vs-destination**, not
href-vs-existence — an anchor whose text names a page-like noun that does not appear in the
destination's title/h1. On farmer that would fire 41 times; the two-name concentration suggests
it would be cheap to make specific.

## Where the rest of farmer's copy state is
- The **14 stage-2 proposals you fired on 09-02 are still at `needs_human_review`** (filed
  17:37–17:48Z, 13 PASS + the annotated farm-buildings FAIL). Two days. The owner has been told
  in his own words that they are ready for the batch review he asked for; the decision is his.
- Farmer's `/contact.html` carries two of the 41 AND a form that POSTs to `#contact`
  (`contact_form_undeliverable`, parked since 08-28; 7 sites fleet-wide) AND an opening sentence
  saying the site does not take your contact details. If your copy pass reaches that page, the
  three interact: fixing the labels alone leaves a page that asks for details it cannot receive.

Farmer lane docs: `docs/agent_docs/docs024_key_docs_latest/farmerinsurance_uk/NOTES_farmerinsurance_uk.md`
(entry "16:3xZ — the worst thing on the site is invisible to every link checker").
