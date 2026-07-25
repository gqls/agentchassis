# 022 FEATURE — a rendered-text legibility guard for generated imagery

**Filed:** 2026-07-25, at the closure of `bugs_closed/011` (hero provider routing), where this
was carried as "R2". Split out rather than dropped: 011's own defect — `hero` routing to a
model that cannot render text — is fixed, live and proven, but the *check* that would have
caught its symptom without a human looking at the picture was never built.
**Status:** OPEN — designed, not built. No owner gate on it; it has simply never been picked up.

## The originating observation

The owner's own Gemini infographic — produced by the *capable* model, on the lane we now route
every declared kind to — rendered **"REPRETITIVE"** for "REPETITIVE". One character wrong,
inside an image, on a page that would otherwise have shipped.

That is the whole case in one word. The old failure mode (SDXL producing wholesale gibberish)
was so obviously broken that any human glancing at the page caught it. The new failure mode is
**a single misspelling inside an otherwise excellent graphic**, and it is the harder one,
because everything around it looks right.

## What is non-obvious

**Nothing anywhere in the pipeline reads the text that a generated image actually contains.**
Generation reports success on the HTTP response; the asset row is written `active`; the page
references it. There is no stage between "the model returned bytes" and "a client sees it" that
could possibly notice a typo. `bugs_closed/011` §2 established that prompt specificity is the
dominant variable in getting the text right — but prompt discipline is a *practice*, and every
practice this platform has relied on instead of a check has eventually been skipped.

Second, and worse for a client site: the model can render a **number that was never in the
prompt**. `bugs_open/043` (generated page copy inventing quantitative claims) is exactly this
defect one layer up, in prose, where we do now scan for it. An invented figure inside a JPEG is
invisible to every scanner we have, including that one.

## Shape of the fix (from 011 R2, unchanged)

An OCR/vision pass **after generation, before the asset is usable**, that:

1. extracts the rendered text;
2. flags misspellings;
3. flags any number not present in the generation request;
4. routes findings to a `site_work_items` row for human review;
5. **never auto-publishes an image whose text failed the check.**

This is the same `check → work-item → HITL` shape as the claims gate and the voice gate, which
is the argument for it: the machinery exists and is understood; this is a new detector on an
established rail, not a new mechanism.

## Constraints whoever builds it must respect

- **The image-generator adapter has no database handle.** This is not incidental — it is why
  provider routing was hardcoded in the first place. A detector that needs to *persist* anything
  cannot simply be dropped into the adapter. `bugs_closed/011` §7 established the correct shape
  and it is now live and council-approved: **the adapter reports the condition in its response
  → the chassis persists it** (`reported_conditions` → `agent_error_log`, sender-allowlisted).
  A legibility verdict should travel the same road rather than inventing a second one.
- **Adding a reporting sender IS the review.** `conditionReportingAgentTypes` is load-bearing and
  was imposed by a guardian hard veto: an unconditional parse in shared dispatch is a fleet-wide
  ungoverned reporting bus. If the detector runs somewhere new, that allowlist entry is a
  deliberate decision, not a config tweak.
- **Absent ≠ malformed ≠ partly-dropped ≠ over-cap.** The same four-state distinction the council
  imposed on `reported_conditions` across rounds 5–9 applies to any new verdict channel. "No
  findings" and "the detector broke" must never look the same.

## The legacy sweep — measured, and it is the first job

011 §4 anticipated this: *"Any already-deployed image generated as `hero` with a structural
prompt is a candidate for the R2 sweep once the detector exists."* Measured at closure,
**2026-07-25**:

```sql
SELECT a.site_id, s.domain, count(*) AS active_sdxl
  FROM assets a LEFT JOIN sites s ON s.id = a.site_id
 WHERE a.origin_model ILIKE '%stability%' AND a.status = 'active'
 GROUP BY 1,2 ORDER BY 3 DESC;
```

**60 active SDXL-generated assets across 8 sites** — robot-hands 14, idea.uk 10, dartsonline 9,
vonc 8, gamesdesign 7, leopardessconsulting 5, relojistas 4, vetcomparison 3. Every one predates
the routing fix (**newest 2026-07-17**; zero generated on or after 2026-07-18). They are the
detector's natural first corpus: a known-weak generator, a known-bounded set, and a real answer
at the end of it about how much of the live estate is affected.

**One is confirmed rendering on a live client page right now** (checked 2026-07-25, see
`bugs_closed/011` §8): `leopardessconsulting.co.uk/how-it-works.html` uses asset `hero`
(`/assets/images/hero.jpg`, SDXL, 2026-07-17) as its hero `background-image` — a flowchart whose
every label is gibberish, dimmed behind a 50–60% black gradient. It is the picture that caused
011 to be filed, still live seven days after the routing defect behind it was fixed. Regenerating
it needs no code: the kind now routes to Banana.

## Open questions

- **Which vision model, and what does a check cost?** The generation itself is $0.134; a check
  that costs the same again doubles the image bill (see `features_open/008` for the volumes:
  ~108 images/month, ~$14.50). A cheap OCR pass plus an LLM only on suspicion may be the right
  shape.
- **Where does "publish" actually happen?** Assets land in the bucket and are marked `active`
  before any page references them, so "never auto-publish a failed image" needs a precise
  definition of the gate point — probably `status`, not the upload.
- **False positives.** Brand names, non-English copy, deliberate stylisation and abbreviations
  will all read as misspellings. A guard that cries wolf gets switched off; the escalation
  threshold matters more than the detector.
- **Numbers are the stronger half.** "Any number not in the request" is a *decidable* check with
  almost no false-positive surface, and it is the one that maps onto `bugs_open/043` and the
  claims-verification layer. It may be worth shipping alone, before spell-checking.

## How you would know it was worth building

Run the sweep over the 60 legacy assets first and count how many contain text at all, and how
many of those contain wrong text. If the answer is "most of them" the guard pays for itself on
the backlog alone. If it is "three", the honest conclusion is that prompt discipline plus the
provider fix already did the work, and the numbers-only check is all that is left worth doing.
Either way that count is cheap to get and nobody has it — which is why this file says *measure
first* rather than *build*.
