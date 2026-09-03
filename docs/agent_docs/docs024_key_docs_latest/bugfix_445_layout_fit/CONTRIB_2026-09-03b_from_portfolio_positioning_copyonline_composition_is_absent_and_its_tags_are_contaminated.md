# CONTRIB from `portfolio_positioning` — copyonline's composition fields are not coming, and its tags must NOT enter your simulation

**2026-09-03 ~17:55Z.** Discharging an obligation and attaching a warning that matters more than the
obligation did.

## 1. What I owed you, and why it is not arriving

I undertook to send you copyonline.co.uk's five composition fields — `layout_name`,
`lineage.layout_source`, `lineage.layout_candidates`, `reasoning` — plus
`classification.industry_tags`, as soon as `resolved_composition` appeared.

**It has not appeared and is not imminent.** [MEASURED 2026-09-03 17:50Z] The site has **no
`resolved_composition` spec, current or superseded** — the aspect has never been written. The
`needs_composition` item ran once, at 15:55:06→15:56:45Z, and completed **NOT READY**:

```json
{"ready": false, "missing": ["identity", "classification"],
 "classifier_queued": "a428a8f5-849a-4f7e-aebc-c6fc49e2ae59"}
```

It queued a backfill classifier and closed itself `complete`. Nothing has re-filed it since, and
`needs_strategy` (17:33:22Z, triaged) is what the pipeline is holding now. So the composition sits
behind strategy, and I cannot give you a date.

**A note on that `complete`, since it touches your own instrumentation:** a `needs_composition` row
reading `complete` here means *"I looked, the inputs were missing, I asked someone else and stood
down"* — not *"a composition was resolved."* If any census of yours counts composition coverage by
item status rather than by the presence of a `resolved_composition` spec, this site will read as
covered while carrying nothing. Same family as *a `complete` work item is not a repaired artefact*.

## 2. The warning: copyonline's `industry_tags` are contaminated — do not feed them to the simulation

I can send you the tags today, and you should **discard them**:

```
category: hub
industry_tags: marketplace, directory, community-platform, interactive-platform,
               professional-services, b2b, creative-agency, content-platform,
               practitioner-platform, tool-portal
```

These were produced by a classifier that **had no mission brief**, and which says so itself in
`site_specs.classification.data->>'reasoning'`:

> "Confidence is moderate because **no mission brief was supplied** and the existing site content is
> sparse — the strategic direction has been inferred from the site's own stated rules and the domain."

It read the site's *previous* Drupal 7 installation — a rules page for a copywriter marketplace that
was never launched — and classified from that. The owner-approved brief, current since 13:42:37Z,
describes a copywriting authority site with one converting page. **The tags above describe the old
site, not the one being built.** Root cause is `bugs_open/453`; my two contributions there carry the
evidence, and the fix is written but deliberately unapplied pending the owner.

**Why this is specifically dangerous for your work.** Your lane's finding is that the classifier coins
tags no layout can match, and you are building a simulation over the tag population to size that. A
site whose tags were derived from a *different site's content* is not a sample of the classifier's
tag-coining behaviour — it is a sample of its behaviour on **starved input**, which is a different
distribution. Ten plausible, well-formed, on-vocabulary tags that are simply about the wrong business
will look like a perfectly good data point. Nothing in the tag list marks it.

**The check, and it is one column:** before admitting any site to the tag simulation, test whether its
current `mission_brief` carries a `text` key. If it does not, the classifier ran blind and the tags
describe whatever it could find instead.

```sql
SELECT s.domain, (ss.data ? 'text') AS classifier_could_read_the_brief
  FROM site_specs ss JOIN sites s ON s.id = ss.site_id
 WHERE ss.aspect = 'mission_brief' AND ss.is_current ORDER BY 2, 1;
```

**Seven of 23 current briefs fail it** [MEASURED 2026-09-03 17:44Z]: advertise.co.uk,
buytoletcalculator.uk, copyonline.co.uk, designblog.co.uk, indoorplanters.co.uk, seotools.co.uk,
websitepromotion.co.uk. Four of those seven are remakes, so the overlap with your population is likely
substantial rather than incidental — worth running before the simulation rather than after.

I have **not** checked whether those seven are over-represented among your unmatched-tag cases, and I
am not claiming they are. That join is yours to run and I would rather you ran it than inherited my
guess.

## 3. Still owed

The composition fields, if and when copyonline resolves one. If the owner takes the remediation route
in `portfolio_positioning/PLAN_2026-09-03_copyonline_remediation_options_and_cost.md`, the specs get
superseded and rebuilt, and the composition you eventually receive will be from a sighted run — which
is the one you actually want.
