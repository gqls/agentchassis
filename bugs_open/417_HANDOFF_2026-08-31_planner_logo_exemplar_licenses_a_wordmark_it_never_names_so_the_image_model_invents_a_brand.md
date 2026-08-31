# 417 — the site planner's logo exemplar licenses a wordmark it never names, so the image model invents a brand ("Farm Shield Info" on farmerinsurance.uk)

**Filed 2026-08-31 by the loanzy_uk_example_site lane, from the owner's review of
farmerinsurance.uk.** Diagnosis loop NOT run — substituted equivalent first-hand
verification, stated per the 2026-07-31 owner ruling: every link in the causal chain below
is a read of the live row/artefact, each quoted, and the chain has no inferential hop.
(Owner-visible harm: a served UK insurance site whose logo carries **someone else's brand
name** — a trademark problem, not a cosmetic one.)

## The chain, each link verified 2026-08-31

1. **The exemplar** — `agent_definitions` row `build-site-planner` (live, is_active), worked
   example in `default_config`:
   `"prompt": "A precise, technical logomark — geometric, restrained, no human figures, no text outside the wordmark itself"`
   It PERMITS a wordmark and names no brand. A quoted exemplar ships verbatim-in-shape
   (memory class: *a QUOTED exemplar in a prompt is copied verbatim*).
2. **The paraphrase** — farmer's `needs_imagery` item (site `99cae989…`, item_key
   `needs_imagery:site:-:logo`, created_by `build-site-planner`): prompt ends *"no
   photorealism, **no text beyond the wordmark**, balanced proportions…"* — the exemplar's
   clause, re-worded, still naming no brand. `identity.company_name` ("Farmer Insurance
   UK") is in the identity spec and is NOT in the prompt.
3. **The generation** — `assets` row `a88c0e99…`, `origin_model
   banana/gemini-3-pro-image-preview`, `origin_prompt` = the paraphrase. Licensed to
   letter a wordmark with no text specified, the model invented one: **"Farm Shield
   Info"** (plausibly compressed from "farm … shield … information site" in the prompt).
4. **The serving** — `/assets/images/logo.png` live in the header (sole brand mark; no
   HTML text wordmark beside it), and the favicon + og_card were DERIVED from it, so the
   invented brand is stamped at three surfaces.

## Why this is a class, not a one-off

- The estate already RULED on this: `platform/orchestration/actions/discovery_checks/
  default_brand_prompt.go:231` — the fallback logo prompt says **"no lettering or
  words"**, with the rationale in-file: *"generated wordmarks reliably produce malformed
  text, and this asset is used at favicon size."* The planner's exemplar contradicts the
  estate's own craft rule; two prompt sources, one rule in one of them. Every planner-built
  site takes the exemplar path; the ruled path is the fallback nobody reaches.
- No seat reads pixels: nothing verifies rendered logo TEXT against
  `identity.company_name` (the acceptance council's gap #2 in the farmer review — see
  OWNER_REVIEW_2026-08-31 in the loanzy lane). The defect is silent until a human looks.

## Fix candidates (ordered by what closes the door)

1. **Fix the exemplar** (one migration on `build-site-planner`): either align with the
   ruled default — "no lettering or words" — or, if the owner wants wordmark logos, make
   the exemplar demand the exact brand string: *"the only text is the exact wordmark
   ‹company_name›"* with the planner instructed to substitute identity.company_name.
   Unnamed-wordmark becomes unrepresentable. **Council scope** (agent config migration).
2. **Belt**: a logo-text check (OCR-shaped, or a cheap model look) comparing rendered
   text to identity — the unowned designer-family gap; routed to the imagery/designer
   threads, not built here.
3. Farmer's instance: regeneration item ALREADY FILED through the framework
   (`3740f5f2-e6ff-4dd3-b60d-8ba502b1c636`, prompt names the exact wordmark, positive-
   framed because `bugs_closed/028` proved banana discards negative clauses). Favicon +
   og_card re-derivation owed AFTER it lands (`derive_brand_head_assets`) — presence-based
   discovery will NOT refile them on its own.

## Verify
- Exemplar: `SELECT substring(default_config::text from position('logomark' in default_config::text) for 120) FROM agent_definitions WHERE type='build-site-planner' AND is_active AND COALESCE(is_snapshot,false)=false;`
- The prompt that shipped: `SELECT origin_prompt FROM assets WHERE id='a88c0e99-6de9-4b7d-996d-3c16d530c8a8';`
- The ruled default: read `default_brand_prompt.go` header + :231.
- Related: 210 (needs_logo unhandleable), 235 (logo stored as hero), 322 (brand-head block page-blind), closed 028 (negative prompts discarded).

---

## CONTRIBUTION — `vigilant_designer_offer_analysis` lane, 2026-08-31: the blast radius, measured, and it is 10 sites carrying the exemplar VERBATIM

*Contributing into this file rather than filing a competing one. The lane was asked whether it wants
fix candidate 2 (pixels-vs-identity); this section answers a different and cheaper question first —
**how far has the exemplar already travelled?** — because it changes how candidate 1 should be
priced.*

### 1. Both halves of the chain re-verified independently, first-hand

**The counter-rule is real.** `discovery_checks/default_brand_prompt.go:234` builds
*"legible at favicon size, centred on a plain background, **no lettering or words**, no photographic
texture, no drop shadows"* — and its own comment at :231 says the rule **is not decoration**:
*"generated wordmarks reliably produce malformed text, and this asset is used at favicon size."*
**So the estate has already learned this lesson once, in code, with the reason written down.**

**And the exemplar really does license the opposite.** Live `build-site-planner` config, inside
`### Worked example`:

> `"prompt": "A precise, technical logomark — geometric, restrained, no human figures, no text outside the wordmark itself"`

⚠ **The phrase LICENSES lettering while reading as a restriction.** *"No text outside the wordmark
itself"* presupposes a wordmark, permits text inside it, and **never says what it should read** — so
the model must invent the words. "Farm Shield Info" is that gap being filled, not a model defect.

### 2. `[MEASURED 2026-08-31]` It has already propagated, and mostly verbatim

Current-plan logo prompts, all 27 sites that have one:

| | count | of 27 |
|---|---|---|
| logo prompts on current plans | **27** | — |
| mention `wordmark` | **19** | 70% |
| **carry the exemplar's phrase `no text outside the wordmark` VERBATIM** | **10** | **37%** |
| forbid text in some form (`no lettering\|no words\|no text`) | 21 | 78% |

**37% of live logo prompts contain the worked example's sentence word for word.** This is not
inspiration; it is transcription. ⚠ **And note the overlap trap in the last row:** the exemplar
phrase *itself* matches `no text`, so 10 of those 21 "forbid text" prompts are the contradictory
ones — a census that only counted "does the prompt forbid text?" would score them as SAFE. **Count
the wordmark licence, not the prohibition.**

### 3. Why this is squarely the estate's known exemplar hazard

This is the `a-quoted-exemplar-in-a-prompt-is-copied-verbatim` shape, and 10 of 27 is the strongest
live evidence for it I have seen: **a quoted exemplar ships as text, not as guidance.** It is the
same mechanism the `copy_quality_two_stage` lane measured from the other end this week — a mandated
phrase chain transfers where style instruction does not. **That cuts both ways, and here it cut the
wrong way.**

**It also raises candidate 1's value above "fix one prompt".** Fixing the exemplar stops the
*next* 27; it does **not** repair the 19 already carrying the licence or the 10 carrying it verbatim.
⚠ **Candidate 1 is necessary and NOT sufficient, and the existing rows are the larger half** — a
fixed exemplar with 19 live prompts still licensing a wordmark reads as solved and is not.
Whoever takes candidate 1 should say explicitly whether the 19 are re-planned, rewritten in place,
or left — and a census AFTER the fix must count the licence, not the prohibition (§2).

### 4. On fix candidate 2 (pixels vs identity) — NOT claimed, and here is the honest reason

The offer here is that this is designer-family work. **This lane is not taking it on a peer relay**,
and there is a substantive reason beyond authority: **candidate 2 is a new capability, not a fix.**
No seat reads pixels today; adding one means an image-reading check, a new failure mode when it is
wrong about a legible mark, and a decision about what it does when it disagrees with the identity
spec. That is architecture-scope by the 2026-07-29 test — it changes what the check fleet
*guarantees* — and it wants the owner, not two lanes agreeing.

**What is worth saying now, so candidate 2 is scoped honestly if it is ever built:** it would be
**a guarantee CONDITIONAL on an OCR/vision classifier, and would inherit that classifier's gaps** —
a stylised or partially-occluded wordmark that the reader misses returns a clean pass, and a clean
pass from a blind check outlives the blindness in every document that later cites it. **Candidate 1
needs no classifier and closes the door for new sites; it should ship first regardless of whether
candidate 2 is ever built.**

**Nothing changed by this lane.** No prompt edited, no migration written, no verdict released.

---

## RESIDUAL + OWNERSHIP TRANSFER, 2026-08-31 (loanzy lane, after 669+670 APPROVED r1 corr 3b666f0f and applied)

**Candidate 1 shipped and verified — and the 420/417 lane's post-fix census found the class
is wider than the fix, exactly as their measurement states (verified here at the row):**
`site_plan_imagery b56182fa` (boxingonline.com, created 12:36:56Z — **42 seconds after 669
applied**, a planner run already in flight carrying the old exemplar) reads *"no text OTHER
THAN the wordmark itself"* — a PARAPHRASE 670's surgical arm keyed on "outside" could not
see, and arm (b) never saw because the row did not exist yet. Two durable lessons, theirs:
1. **"Stops the next sites" has a race tail** — the honest verify is a census dated after
   the last in-flight plan lands, not at migration commit time.
2. **Counting the licence by literal is still a literal.** The model REWORDS the exemplar,
   so the binding census is the CONCEPT: a wordmark permitted without the brand string
   named. (Sharpens this file's own §3 warning one turn further.)

**OWNERSHIP: the residual is the `bugfix 420 and 417` lane's from here** (loanzy filed the
bug and shipped candidate 1; the race-tail row, the concept-census, and close-out are
theirs). Their stated coordination: the boxingonline row is being handled WITH that site's
own session, not over it. Candidate 2 (pixels-vs-identity) stays where the designer lane
left it — architecture-scope, unbuilt, the owner's call. boxingonline's row was NOT
deliberately left by 670 — it is purely the race tail.
