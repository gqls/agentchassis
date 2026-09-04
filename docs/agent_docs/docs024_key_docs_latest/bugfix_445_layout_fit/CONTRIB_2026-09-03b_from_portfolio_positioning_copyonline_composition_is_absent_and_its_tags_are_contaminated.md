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

---

## ADDENDUM ~18:05Z — the tag warning STANDS; one sentence of context around it does not

**§2 is unchanged and you should still act on it.** copyonline's `industry_tags` were written by a
classifier that had no brief and said so, they describe the previous Drupal site, and they must not
enter the tag simulation. The admission check in §2 and the seven failing sites are all measured and
unaffected by what follows.

**What I need to withdraw is the surrounding picture.** §2 says the site "is being designed as a
copywriter marketplace". That was true of the classification and is no longer true of the site.
Eleven minutes after I wrote it, `domain-strategist` produced `site_specs.strategy` with
`site_type: "authority-portal"`, reproducing the brief in detail. It reads the brief through a
whole-blob render (`{{.site_specs}}`) where the classifier and planner reach for
`{{.site_specs.specs.mission_brief.text}}` and get nothing. Full account in `bugs_open/453` CONTRIB (3).

**Why this is worth your attention rather than just my correction.** It makes copyonline a **cleaner**
case for you, not a spoiled one:

- The **tags** are a pure sample of classifier-on-starved-input. Still contaminated, still excluded.
- The **site's direction** was recovered downstream by a different agent. So if you ever want to ask
  whether a bad classification actually determines a site's shape, this site is a worked negative — it
  did not, because a later agent read the brief generously.
- If a `resolved_composition` eventually appears here it will have been composed against a **correct**
  strategy and an **incorrect** classification. Whichever your composition path leans on will be
  visible in the result, which is a natural experiment I could not have arranged deliberately.

I have not measured which of the two the composition path weights, and I am not implying it. Flagging
the shape because it is yours to exploit if it is useful.

---

## ADDENDUM 21:30Z — copyonline's tags are now SIGHTED; the contamination warning applies only to the superseded row

Migration 764 (classifier + planner render a `text`-less brief as its object) went live at 20:55:27Z on
the owner's word and was proven at the artefact at **21:25:18Z**: copyonline's classification was re-run
against its brief and the new current row reads `category=editorial`, tags
`editorial-guides, content-hub, tool-portal, interactive-tools, guides, directory, knowledge-base,
long-form-content, practitioner-platform, founder-tools`, with a reasoning that names the brief's four
tools and its lead route. **These tags ARE a sample of the classifier reading a brief** — admit them.

The 16:57:10Z row (`hub`; marketplace / community-platform / …) is superseded in place and remains the
starved-input sample; exclude it by `created_at`, not by site. The admission check in §2 still stands
for the other six sites until they are re-classified, and after 764 the check becomes "was the
classification row written after 2026-09-03 20:55:27Z" for any site whose brief lacks `text`.

---

## ADDENDUM 2026-09-04 ~07:50Z — copyonline's composition EXISTS, resolved against SIGHTED tags: the five fields I owed you

`site_specs.resolved_composition` written **2026-09-04 02:07Z** by the composition step that re-filed
itself overnight (the `MissingStyleCollectionCheck` route this lane predicted; `sites.style_collection_id`
now `88e3cfb9…`). It was resolved against the classification of 2026-09-03 21:25:18Z — the first one
that could read the brief — so this is a clean sample of the resolver on a sighted tag set.

| field | value |
|---|---|
| `layout_name` | **`content-hub-tools`** |
| `lineage.layout_source` | `library_match` |
| `lineage.layout_fit.score` / `margin` / `tag_coverage` | 18.37 / 8.18 / 0.565 |
| `lineage.layout_fit.matched_terms` | content-hub, editorial-guides, editorial-publication, guides, interactive-tools, long-form-content |
| `lineage.layout_fit.unmatched_terms` | directory, founder-tools, knowledge-base, pra… (truncated in my read — pull the row) |
| `lineage.layout_fit.runner_up` / `runner_up_score` | `tool-portal-light` / 10.19 |
| `classification.industry_tags` (21:25:18Z) | editorial-guides, content-hub, tool-portal, interactive-tools, guides, directory, knowledge-base, long-form-content, practitioner-platform, founder-tools |
| `reasoning` | present on the row; not quoted here |

Two things you will want to weigh rather than take from me: (1) `content-hub-tools` is the archetype
your bug title says the library lacked — if your `736` archetype is what it matched, this is its first
sighted hit and the score/margin are your calibration data; (2) `directory` is UNMATCHED, and the
directory is a core page of this brief — whether that is a library gap or correct subordination is
your call. The earlier §2 warning stands only for the superseded 16:57 classification row.
