# BUG 248 — a CTA recompute silently overwrites an AUTHORED `/contact.html` link, because the "keep a sensible authored link" guard refuses every excluded area

**Filed 2026-08-10** by the `bugfix_203_phantom_cta_cleanup` lane, which hit this while
deciding whether the `misdirected_cta` queue could be drained to repair pages.
**Status: OPEN, not started.** Nothing here is fixed.

## Why this is filed on first-hand verification rather than a `090` run

Per CLAUDE.md's 2026-07-31 owner ruling, a `bugs_open/` file asserting a structural cause
must go through the diagnosis loop *or* state plainly why equivalent first-hand
verification was substituted. Substituting, and here is the substitute:

- **The defect is reproduced mechanically**, not inferred — an A/B against the package's
  own existing test, changing exactly one variable (transcript below). A `090` run's
  value is finding a cause you cannot see; this cause is four lines of Go and the repro
  is decisive either way, including a stated refutation branch that did not fire.
- The blast radius is **measured against live data**, not estimated (query below).
- The mechanism was **already half-documented in `bugs_closed/023`** (see "Prior art"),
  so it is not a novel structural claim — it is the destructive direction of a defect
  that file recorded in its benign direction and closed without fixing.

## Symptom

A page whose CTA correctly points at `/contact.html` — with genuine contact copy such as
"Get in Touch" or "Talk to Us" — has that link **replaced by an unrelated tool or hub
page** the next time any `cta_links_stale` rerender touches it. The work item completes
green, the page re-renders successfully, and nothing is logged. The only way to see it is
to diff the CTA url before and after.

## Reproduction (mechanical, 2026-08-10, at commit `3bc0486d7`)

Take the package's own `TestApplyCTARecomputeFallsBackWhenLabelGeneric`, which asserts
that a generic label leaves a stored valid link untouched. Change **one** variable — the
stored URL — from a normal page to `/contact.html`. Same label, same candidates, same
positional target, same `applyCTARecompute` call:

```
CONTROL  stored=/tools/password-entropy.html label="Get in Touch" -> resolved=map[]
CASE     stored=/contact.html                label="Get in Touch" -> resolved=map[
             cta_url:/tools/tool-ai-data-risk-checker.html
             cta_target_title:tool-risk-checker]
```

The control is kept untouched. The case is clobbered. (The repro test was written, run,
and then deleted rather than committed — it asserts a defect, and enshrining current
wrong behaviour as a passing test is how a bug becomes a spec. The fixing thread should
write the *correct* assertion instead; the transcript above is the evidence.)

## Root cause

Two collaborating pieces, both working as written:

1. **`applyCTARecompute`** — `platform/orchestration/actions/rerender_page_sections_action.go:686-691`:

```go
if hasCurrent && current != "" &&
    validPages.Contains(current) &&
    !ctaExcludedDestination(current) &&            // <-- an authored /contact.html fails HERE
    NormalizePagePath(current) != NormalizePagePath(pageURL) {
    return // authored link to a real, sensible destination — keep it
}
```

2. **`areasExcludedFromCTA`** — `platform/orchestration/actions/resolve_internal_links_action.go:72-74`
   `= {about, contact, privacy, terms, legal}`, consumed by `ctaExcludedDestination` (`:468-475`).

So a CTA already pointing into any excluded area **can never take the keep branch**. It
then needs the label-match branch above it to save it — and genuine contact copy is
exactly the copy that matches nothing, because `get`/`in`/`us`/`to`/`we` are all in
`datahelpers.LabelStopwords`: "Get in Touch" reduces to `[touch]`, "Talk to Us" to
`[talk]`, neither of which names a page. Both branches decline, and control falls through
to the positional pick, which overwrites.

**This is the code doing bug 203's job.** 203's original defect was `/contact.html` being
a *fabricated* fallback that needed recomputing away — the exclusion is deliberate and
correct for that case. The problem is that **a fabricated `/contact.html` and an authored
one are byte-identical in `content_data`** — there is no provenance on the field — so the
repair cannot target one without hitting the other.

## Blast radius (measured live, 2026-08-10)

```sql
SELECT count(*) FROM page_components pc
WHERE COALESCE(pc.content_data->>'cta_url', pc.content_data->>'primary_cta_url',
               pc.content_data->>'secondary_cta_url')
      ~ '^/(contact|about|privacy|terms|legal)(\.html|/|$)';
```
→ **24 CTAs fleet-wide** currently sit in the at-risk state (7 of them written during the
2026-08-09→10 window when a separate label-match priority bug was live).

This is a count of components in the *vulnerable* state, **not** a count of damage done —
damage requires a rerender to actually touch them, and with the discovery/improvement
schedulers currently disabled (see below) few will be touched spontaneously. **It is a
measure of what a bulk repair would break**, which is why it matters now.

## Why it is dangerous RIGHT NOW specifically

The `misdirected_cta` queue holds **192 `detected` / 95 `unresolved` / 63 `failed`** items
(2026-08-10). The obvious way to "let pages heal" is to promote them and let the dispatch
loop repair — and `TriageDetectedItemsAction` promotes **every** `detected` row for a site
with **no type filter** (its own file header says so, `triage_detect_items_action.go:1-13`).
So one well-intentioned bulk promotion runs this defect across every affected site at once.
The `bugfix_203` lane came within one command of doing exactly that.

Compounding it: that queue is **substantially false positives** — the 2026-08-07 finding
in `bugfix_203_phantom_cta_cleanup/NOTES` is that it flags *correct* "Get in Touch" →
`/contact.html` buttons as misdirected. Those false positives are precisely the population
this defect destroys. The earlier note blamed the detector's *suggestions*; the damage
actually comes from the recompute's *fallback*, which fires regardless of what the
suggestion said. That is a sharper and more actionable statement of the same worry.

## Fix candidates, ordered by what closes the door

1. **Give the field provenance** — the real fix. If `content_data` recorded whether a CTA
   url was authored or resolver-derived, the keep-guard could preserve authored links
   into excluded areas and still recompute fabricated ones. Closes the class: the bad
   state stops being representable. Largest change; touches the write path and needs a
   backfill story for existing rows (which have no provenance and would have to default
   to one side — probably "derived", preserving today's behaviour until re-authored).
2. **Narrow the exclusion to the fallback path only.** `areasExcludedFromCTA` exists to
   stop the *positional picker* choosing contact as a destination. Using the same set to
   judge an *already-stored* link conflates "don't newly send people here" with "this
   existing link is untrustworthy". Splitting those two uses is small, local, and removes
   the clobber without needing provenance — a stored contact link would be kept, while a
   fresh pick still never lands on contact. **Cheapest fix that actually closes the
   symptom; recommended starting point.**
3. **Exempt a CTA whose label is contact-intent copy.** Add a small intent set
   (`touch`, `talk`, `contact`, `enquire`, `enquiry`, `quote`, …) that positively
   licenses an excluded-area destination. Narrow, but it is a keyword list against
   LLM-authored copy, so it will always have gaps — and per the standing landmine,
   narrowing a rule past invented false positives is how a rule goes inert.
4. **Do nothing to the code; gate the queue operationally.** Rejected as a fix — it
   leaves a live trap for every future session and depends on everyone reading a doc.
   Recorded only because it is what is protecting us today, by accident, since the
   schedulers are off.

## Related state that matters to whoever picks this up

- **The discovery and improvement schedulers are currently DISABLED**
  (`improvement-sweep`, `site-discovery-rotation-{quality,design,completeness}` all
  `enabled=f`, measured 2026-08-10). Detection and triage are both dead; only
  `build-pipeline-trigger` (120s) still dispatches. So this defect is mostly dormant —
  and will re-arm the moment those are switched back on. **Do not switch them on before
  this is fixed** without first accepting the 24-component exposure.
- Registered as a LANDMINE the same day (footprint `applyCTARecompute` /
  `site_work_items` `misdirected_cta` / `areasExcludedFromCTA`), synced to `doc_notes`.

## Prior art — this was seen once and closed without fixing

`bugs_closed/023_HANDOFF_2026-07-19_cta_label_url_pairing_unchecked.md:405-410` recorded
the **benign direction** of the same mechanism:

> **`areasExcludedFromCTA` (`:72-74`) = `{about, contact, privacy, terms, legal}` makes
> some correct pairings unreachable.** A button reading "Request Integration Support"
> *should* go to `/contact.html`, and the resolver will never allow it. That exclusion is
> sensible as a default for a generated CTA and wrong as an absolute — it needs an
> authored-intent escape hatch.

That file identified the right fix ("an authored-intent escape hatch") and closed anyway,
because in the forward direction the cost is only a *missing* link. **The destructive
direction — an existing correct link being destroyed — was not noticed**, and it is the
one with a live blast radius. Fix candidate 1 above is 023's escape hatch, generalised.

## How to verify a fix

1. Re-run the A/B repro at the top: the `/contact.html` case must join the control in
   being left untouched, with the same generic label.
2. It must NOT overcorrect: a genuinely fabricated contact fallback whose label DOES name
   a real page (e.g. "Run the Risk Checker" → `/contact.html`) must still be recomputed to
   the risk checker. The existing
   `TestApplyCTARecomputeOverridesValidButMisdirectedLink` covers this direction — it must
   keep passing.
3. Then, and only then, a bulk promotion of the `misdirected_cta` queue becomes safe to
   consider — separately, and still subject to that queue's own false-positive problem.


## Second-site observation — leopardessconsulting.co.uk /services.html (2026-08-14, services-restore session)

The authored CTA on `/services.html` (authored 2026-07-31: primary "Get in touch" →
`/contact.html`; secondary → `/tools/ai-agent-roi-estimator.html` with a label naming the
tool) was found on 08-12 replaced by: primary "Book an architecture conversation" →
`/tools/tool-agent-complexity-estimator.html`, secondary → the time-savings estimator. An
authored `/contact.html` primary replaced by a tool link is this file's mechanism, observed
on a second site.

Attribution `[UNVERIFIED which side]`: the `page_rerender` work item that triggered the
08-11 18:15 regeneration of this page was itself *"1 misdirected CTA(s) on services"* — so
either the recompute created the misdirection it was dispatched to fix, or the regeneration
recreated it and the recompute failed to correct it. One more data point either way: the
`call-to-action` slot was touched again 2026-08-12 20:49, ADDING
`primary_cta_target_title`/`secondary_cta_target_title` keys that name the tool targets —
something annotated the mismatched state without correcting it. (Checked against
`bugs_open/268` before filing here: NOT that class — the URL keys were present throughout,
rewritten rather than dropped.)

Repaired by hand today (content_data restore + `section_data_resolved` rerender, verified
at the served page ~18:45Z, real-click probe + no-init mutant). The restored keys include
`*_target_title` values matching `pages.title` for both targets, so a title-comparing
checker should now read them as consistent. Survival past the next regeneration is
unverified — the leopardess RUNNING_NOTES entry of this date carries the re-check.
