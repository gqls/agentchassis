# HANDOFF — `kind:"hero"` routes to a model that cannot render text; the lane that can was never used

**Filed:** 2026-07-18, from the leopardessconsulting.co.uk rebuild (owner review).
**Severity:** Medium. No code change is strictly required to get good infographics today — the
capability is already wired. The bug is a routing default plus an unused lane.
**Status:** **R1 FIXED, LIVE AND PROVEN END-TO-END 2026-07-20** on `v1.0.1139` — verified in
the running binaries *and* by a real generation: `hero_guides` on dartsonline at 10:41Z came
back `origin_model = banana/gemini-3-pro-image-preview` (with 7 icons behind it, 8/8, no
timeouts). **Cost is answered and the change is APPROVED by the owner** (§6: ~14× per hero,
≈ +$5/month, fleet image bill under $15/month). **R2/R3/R4 remain OPEN, so this file stays in
`/bugs_open/`.** The infographic capability itself is **PROVEN WORKING**. R1 going live arms
`bugs_open/028` on heroes — read §6 before generating any. Half-price batch API deferred to
`features_open/008`.

> ## ⚠️ CORRECTION — read this before anything else
> The first version of this handoff (same number, 2026-07-18 morning) claimed *"generated
> images cannot render readable text"* and recommended building an SVG renderer because
> diffusion models "synthesise glyph-shaped texture, not text". **That claim was wrong**, and
> a thread acting on it would have built the wrong thing.
>
> The owner produced two Gemini infographics with perfectly legible, correctly-spelled text,
> then asked whether we could wire that up. We already had: the deployed
> `BANANA_DEFAULT_MODEL` is **`gemini-3-pro-image-preview`**, and `kind:"infographic"`
> **already routes to it**. Generating through that lane produced a production-quality
> infographic on the first attempt — legible throughout, correct figures, on-brand.
> Evidence: `https://leopardessconsulting.co.uk/assets/images/infographic-what-we-build.jpg`
> (asset `infographic_what_we_build`, `origin_model=banana/gemini-3-pro-image-preview`).
>
> The generalisation "image models can't do text" was true of SDXL and is no longer true of
> the current Gemini image model. Corrected in place rather than deleted, because the wrong
> inference is an easy one to repeat.

---

## 1. What is actually broken

**The garbled homepage hero was a routing accident, not a capability limit.**

`internal/adapters/imagegenerator/dynamic_adapter.go` switches provider on `kind`:

```
icon | logo | illustration | infographic | sprite_sheet | content_hero  → Banana (gemini-3-pro-image-preview)
everything else (including "hero")                                     → Stability (SDXL v1.0)
```

So `kind:"hero"` gets SDXL. SDXL genuinely cannot render text. When a hero prompt implies any
structure ("a diagram of a pipeline…"), SDXL returns a convincing-looking flowchart full of
gibberish words — which is exactly what shipped as this site's homepage hero
(`/assets/images/hero.jpg`, still live as the site-wide fallback and still on how-it-works).

Two consequences worth separating:

1. **Routing default.** For a site whose house style is flat illustration, `hero` is the one
   kind that lands on the photographic model least able to serve it. Heroes on this site only
   became good when explicitly requested as `kind:"illustration"`.
2. **Unused lane.** `infographic` has routed to the capable model all along. Nothing on this
   site had ever used it. The "we can't make infographics" belief was self-inflicted.

## 2. What works today (verified, no code change needed)

- `kind:"infographic"` → Banana → `gemini-3-pro-image-preview`, with a **richly specified
  prompt**, produces publishable infographics with legible, accurate text.
- Prompt specificity is the dominant variable. The successful prompt names the layout, every
  column header, every card's heading and body text verbatim, the exact figures permitted, the
  palette by hex, the icon for each card, and ends with an explicit instruction that all text
  must be correctly spelled and real, and that **no number outside the supplied list may
  appear**. Thin prompts are what produced the earlier rubbish.
- `kind:"illustration"` → Banana with hard no-text constraints produces good text-free heroes
  (three now live on this site).

## 3. Remaining real work

**R1 — fix the hero routing default. ✅ FIXED IN CODE 2026-07-18 — see §6 for what shipped
and what is still owed.** Choose the provider from the site's
`design_intent.imagery_direction` (or an explicit per-site provider preference), not from the
kind string alone. A site declaring a flat-illustration house style should never have its
heroes sent to the photographic model. Low risk, fleet-wide benefit.

> **Route not taken, and why.** The obvious reading of R1 — infer the provider by keyword-
> matching the free-text `design_intent.imagery_direction` — was tried against all 11 live
> values first and **rejected**: it misfires on at least three. Site `9ec3b9ee` reads
> *"Minimal photography. Prefer abstract geometric constructions…"* and `1244516d` reads
> *"Photography and illustration should be minimal…"* — both contain "photography" while
> intending the opposite, so a substring match would misroute them **silently**. An explicit
> per-site field plus a better default beats a fuzzy guess over prose the planner never wrote
> for this purpose.

**R2 — a text-legibility guard before publish.** Even the good model is not perfect: the
owner's own Gemini map rendered "REPRETITIVE" for "REPETITIVE". A typo in a generated
infographic is a real defect on a professional site, and no pipeline signal catches it —
generation reports success. Add an OCR/vision pass after generation that (a) extracts the
rendered text and (b) flags misspellings and any number not present in the request. Route
findings to human review; never auto-publish an image whose text failed the check. This is the
same check→work-item→HITL shape as the claims and voice gates.

**R3 — numbers must come from the evidence base.** The generated infographic is accurate
because the prompt carried audited figures and forbade any others. That should be structural,
not a matter of prompt discipline: build infographic prompts from
`site_specs.evidence_base` facts so an infographic cannot state an unverified number. Ties
directly into the claims-verification layer.

**R4 — keep code-rendered SVG for exact data.** Generated infographics are now good enough for
*explanatory* graphics. They are still the wrong tool for a chart whose values must be exactly
right, selectable, translatable and screen-reader accessible. The L7 chart component (Go emits
SVG from real values) remains worth building for data; it is no longer needed for explanation.

## 4. Blast radius

Fleet-wide. Every site's heroes route by kind, so every site with an illustration house style
has the same mismatch, and no site has used the infographic lane. Any already-deployed image
generated as `hero` with a structural prompt is a candidate for the R2 sweep once the detector
exists.

## 6. R1 — what shipped, 2026-07-18 (and the one thing still owed)

Fixed as one commit on `085_debug_and_feature_loops`. Council gate correlation
`e996bf0a-4cdd-40fa-8ff0-1f1a76c3d181` — **three rounds, final verdict REVISE, not APPROVED**;
what remains open is recorded below rather than quietly dropped.

**The bug was bigger than "hero is on the wrong model".** The real defect is the *mechanism*:
provider selection was a hand-maintained `switch` whose `default:` branch routed to Stability
**silently**. `content_hero` fell through it and shipped mis-routed; `hero` fell through it and
shipped a gibberish diagram as a client homepage. Both were found months later by a human
looking at an image, because generation reports success and dropped brand anchors say nothing.
Adding `hero` to the switch's list would have fixed instance three and left instance four to be
found the same way — this is the council's `bug_historian` objection, and it was right.

What landed:

1. **`internal/adapters/imagegenerator/routing.go` (new).** The switch is now an enumerable
   table `kindProviderRouting` plus a pure `routeProvider` function. Because the routed set is
   data the code can interrogate, a non-empty kind *absent* from it is **detected**: the adapter
   logs `UNROUTED KIND` naming the kind and listing the valid set, instead of quietly serving
   from the weaker provider. Adding a kind is still a code change; **forgetting to is no longer
   silent.** An empty kind deliberately does *not* warn — legacy callers predating the field are
   a documented Stability path, and a warning that fires constantly is one nobody reads.
2. **`hero` joined the routed set** — the last kind left behind, by omission, and the largest:
   **84 of 155** `site_plan_imagery` rows.
3. **Per-site escape hatch.** `imagery_style_guide` gained an optional `provider` field
   (`"banana"` | `"stability"`), guide-level and per-kind, resolved by `providerForKind` —
   mirroring `avoidForKind`'s override-wins-**even-when-empty** contract — and passed to the
   adapter as `provider_hint`. A site wanting SDXL heroes back sets **data, not code**.
   The adapter has no DB handle, which is *why* routing was hardcoded; resolving in the action
   layer and shipping the answer as data is what makes the decision site-owned.
4. **Tests** for the two widest-blast-radius behaviours plus the guard itself
   (`routing_test.go`, `TestProviderForKind`).

**Verified this round** (both were council escalate-conditions, so they are checked, not
asserted): `ImageRequestData{}` is constructed **nowhere** in the repo — the adapter only
unmarshals it from Kafka JSON — and only three files touch the adapter topic
(`topic_manager.go` declares it, `generate_image_actions.go` is the sole producer, the adapter
consumes). **`GenerateImageAction` is the exclusive path, so no caller bypasses the plumbing.**
The `imagery_style_guide` JSON likewise has exactly one reader (`getImageryStyleGuideForSite`)
plus one seed file — no UI, frontend or other service — so the new field is safe to add.

### ~~Still owed~~ — the residual objection (`bug_historian`, medium; `guardian`, low)

> **RESOLVED 2026-07-20 evening — see §7, which supersedes this subsection.** Another thread
> built it in exactly the shape prescribed below (*adapter reports → chassis persists*) and it
> is **live on `v1.0.1140`**, verified in the deployed binaries: `UNROUTED_IMAGE_KIND` /
> `REFERENCE_ANCHORS_DROPPED` present on **both** adapter replicas, and the chassis carries
> `persistReportedConditions` (`coordinator.go:2258`, commit `8ec9e2ab8`) — confirmed by its
> `"reported_conditions from an UNSANCTIONED sender"` and `"…but MALFORMED"` strings. The
> paragraphs below are kept because the **trap** they describe is what the implementation had
> to avoid, and it is still the trap for anyone who touches this again.

**`UnmigratedKind` is a log line, not a record.** It closes the silent-failure gap only for
someone reading logs on the right pod — which this repo's own history says is unreliable. The
platform already has the right shape for this: `agent_error_log(severity, resolved,
work_item_id, context)`, and `site_work_items` for anything that should demand action. A
fleet-wide dashboard should be able to catch the next unmigrated kind; today only a human
tailing logs can.

**Why it was not done in this commit, and the trap for whoever does it.** The image-generator
adapter has **no database handle at all** (`grep sql.DB|pgxpool` in `dynamic_adapter.go`
returns nothing) — persisting from there means giving an adapter service a DB dependency, which
is an architectural change, not a small one. The tempting shortcut is to detect it in the
**action** layer instead, which does have a DB — **do not do this**: the action and the adapter
are *separate services on separate images*, so the action would be predicting a routing table
that may not match the one actually deployed in the adapter. That is the
dedup-index↔Go-list drift class this platform has already been bitten by. The structurally
correct shape is **adapter reports the condition in its response → the action, which has the DB
and the orchestration context, persists it**. That is a coherent task of its own.

### Cost — ANSWERED 2026-07-20, and the change is APPROVED by the owner

The routing change is **approved**. The cost question that was left open at sign-off has been
costed from list prices against real generation volume. **Headline: ~14× more per hero image,
but the absolute amounts are trivial — about +$5/month at current volume, and the fleet's
entire image bill is under $15/month.**

| | per image | source |
|---|---|---|
| Gemini 3 Pro Image, 1K/2K (`gemini-3-pro-image-preview`) | **$0.134** | Google's official pricing page (1120 tokens @ $120/1M) |
| Gemini, **Batch API** (async, ≤24h) | **$0.067** | same page — exactly 50% |
| SDXL 1024×1024 @ 30 steps (our exact `hero` config) | **~$0.0094** | getimg.ai's rate for that identical config — see caveat |

**Real volume** (`assets`, excluding `derived-*`):

| month | heroes generated | all images |
|---|---|---|
| 2026-05 | 8 | 22 |
| 2026-06 | 15 | 46 |
| 2026-07 (to 20th) | **40** | **108** |

- **Delta from this change:** heroes only. At July's rate, 40 × ($0.134 − $0.0094) ≈ **+$5.00/month**.
- **Fleet total after the change** (every declared kind is now Banana): 108 × $0.134 ≈ **$14.50/month**,
  against ~$8.50 for the mixed routing it replaces.
- **One-off backlog:** `site_plan_imagery` holds **89 planned heroes** not yet generated. If a
  sweep drains them: 89 × $0.134 ≈ **$11.93**, against ~$0.84 on SDXL.

**The lever, if volume grows.** The Batch API halves the price and our image pipeline is
**already fully asynchronous** — Kafka, work items, hand-fired sweeps; nothing waits on an image
interactively — so the ≤24h turnaround costs us nothing we are using. That would take the fleet
to ~$7.25/month. **Not verified:** whether `banana/api` and our provider wrapper can submit batch
jobs at all; that is a code question nobody has looked at.

**Caveats, so the figures are not over-trusted.**
- The SDXL number is a **proxy**: Stability no longer publishes a legacy v1 REST rate card for
  `stable-diffusion-xl-1024-v1-0` (1 credit = $0.01 is confirmed; the legacy engine is listed
  only as "varies by resolution/steps"). Published estimates span **$0.002–$0.009**, so the
  multiple is somewhere between **~14× and ~65×** — but since the base is fractions of a cent,
  the absolute delta barely moves either way.
- These are **list prices**. No invoice was consulted, and **the platform records no cost data
  at all** — `llm_call_log` covers text calls only (`provider` ∈ {anthropic, ollama}); image
  generations write nothing anywhere. Per-image spend is currently unknowable from our own data.
- **What to watch is the growth, not the total**: heroes went 8 → 15 → 40 per month. At 10× this
  volume the fleet bill is ~$145/month and batch stops being optional.
- **Latency: the timeout risk did not materialise on first contact, but is not measured.**
  The adapter's HTTP timeout (120s) was tuned around SDXL's 30–60s generation, and the worry
  was that a slower provider would surface as timeouts under load. The first 8 post-roll
  generations (dartsonline, 2026-07-20 10:41–10:56Z: 1 `hero` + 7 `icon`) **all succeeded,
  8/8, with no timeouts and no retries** — so the 120s ceiling holds for real Gemini calls at
  1024². What is still *not* known is the per-call generation time: the ~2-minute spacing
  between those rows is orchestration cadence, not latency, and the adapter's own
  `TimeSpent` is logged but never persisted. Anyone wanting a real number should read
  `duration` off the adapter log rather than infer it from `assets.created_at`.

Reversible per-site as data (`provider:"stability"`) and fleet-wide by one line in
`kindProviderRouting`.

**Content risk:** photographic-house-style sites (`00ff3af5` robot-hands, `5fe8785b` darts,
`ecf15e75` relojistas) will get *new* heroes from a different model than their existing ones,
so a page can mix two visual languages until regenerated.

**Deploy state: ✅ LIVE 2026-07-20 on `v1.0.1139`** (both services; pods started 07:35).
Verified against the **running binaries**, never the tag — and using log-message strings,
because the Docker build does not retain `case` values and a miss on those reads exactly
like a stale deploy:

| service | pod | check | result |
|---|---|---|---|
| image-generator-adapter | `…-764d758d5c-lmp5j` | `strings /app/image-generator-adapter \| grep -c "UNROUTED KIND"` | 1 |
| image-generator-adapter | `…-764d758d5c-lmp5j` | `… grep -c "routed_kinds"` | 1 |
| agent-chassis | `…-645674b498-rndg9` | `strings /app/agent-chassis \| grep -c "site provider preference applied"` | 1 |
| agent-chassis | `…-645674b498-rndg9` | `… grep -c "provider_hint"` | 1 |

> **UPDATED 2026-07-20 ~11:50Z — now PROVEN end-to-end.** The paragraph below recorded
> the morning's state (zero post-roll generations); the proof arrived the same day.
> Eight assets generated on dartsonline.com (`5fe8785b`), 10:41–10:56Z, the first since
> the roll: **1 `hero` + 7 `icon`, every one
> `origin_model = 'banana/gemini-3-pro-image-preview'`**. The adapter's own decision is
> in its log:
> ```
> {"ts":"2026-07-20T10:41:08.308Z","caller":"imagegenerator/dynamic_adapter.go:569",
>  "msg":"generateImage: dispatching to provider","kind":"hero","provider":"banana",
>  "aspect_ratio":"16:9","prompt_len":519,"has_negative_prompt":true}
> ```
> No `UNROUTED KIND` line on either replica. **Trap found while checking: the adapter
> runs TWO replicas** (`…-lmp5j`, `…-pl6jc`); all traffic hit `-pl6jc`, and grepping only
> the first replica returns nothing — which reads exactly like "no generation happened".
> Grep both. One observation for `bugs_open/028`, not this bug: the hero dispatch carried
> `has_negative_prompt:true`, and the running adapter predates `32f2d51e2` (pods started
> 07:35, no restart since), so that negative prompt still reached Banana's discard path.
>
> > **CLOSED 2026-07-20 18:58 BST.** A fresh build shipped both services to **`v1.0.1140`**,
> > which carries `32f2d51e2`: the string `"folded NegativePrompt into positive prompt as a
> > prohibition clause"` is present on **both** adapter replicas
> > (`…-6df8q`, `…-drwlg`). So from 17:58Z the discard path is gone and **`avoid` lists —
> > including heroes' — reach the model again**, folded into the positive prompt. The
> > 10:41Z hero above was generated in the ~7-hour window where it did not, and is the last
> > one that will be.

**But NOT yet observed end-to-end** *(state as of the roll, 07:35Z — superseded above)*.
Zero assets have been generated since the roll
(`SELECT … FROM assets WHERE created_at > '2026-07-20 07:35'` → 0 rows), consistent with
the owner's tool-imagery HOLD. The code is live and the binaries carry it; **no hero has
actually been generated through the new path.** Per "trust the rendered artefact, not the
status", R1 is not fully proven until one has — the first hero generated after the hold
lifts is that proof, and its `assets.origin_model` should read `banana/…`.

### What R1 going live arms — read `bugs_open/028` before generating heroes

Routing `hero` to Banana means **`hero`'s negative prompt is now inert**: Banana discards
negative prompts outright (`banana/provider.go` header: *"NegativePrompt … is ignored here
(Gemini has no negative-prompt concept)"*), so `kindDefaults["hero"].NegativePrompt`
(*"text, watermark, signature, low quality, blurry, distorted"*) and any style-guide
`avoid` for heroes now reach nothing. **R1 did not cause that defect — `bugs_open/028` is
a pre-existing fleet-wide bug — but R1 extended it to the fleet's largest kind**, which is
recorded in 028 with the verification above. Its fix candidate 1 (fold `avoid` into the
POSITIVE prompt) is the right shape; **do not "fix" it by routing heroes back to SDXL** —
that trades brand anchoring and legible text for a negative prompt, which is the wrong way
round. That bug has an active thread; this one does not touch it.

Second consequence, for whoever takes 028 candidate 1: `maxImageryDirectionInPrompt = 200`
carries an in-code note that it is sized for *"the only generation backend (Stability
hosted SDXL)"* and its 77-token CLIP wall, listing Banana at *"~1000+ char effective"*,
explicitly deferred *"until provider routing lands"*. **Provider routing has now landed**,
so that deferral has come due — the cap is calibrated for a provider no declared kind
uses. See `bugs_open/027` §4b, which found the same cap truncating palettes.

## 5. Key files

- `internal/adapters/imagegenerator/dynamic_adapter.go` — the provider switch (the routing fix)
- `internal/adapters/imagegenerator/banana/provider.go` — Banana/Gemini provider; model from `BANANA_DEFAULT_MODEL`
- `platform/orchestration/actions/generate_image_actions.go` — prompt assembly, `constraints` → negative prompt, per-kind defaults
- `docs/leopardessconsulting/PLAN_imagery_and_design_2026-07-18.md` — the site-side plan this came from
- Working example prompt + result: scratchpad `infographic.json`; live asset `infographic-what-we-build.jpg`

## 7. R1's council residual — BUILT, LIVE, still under review (2026-07-20 evening)

`bug_historian`'s objection across rounds 1–3 (detection lives only in pod logs) is
**implemented and running in production** on `v1.0.1140`, but the council verdict is
**REVISE** — round 8 was queued at ~18:10Z and is the outstanding item.

**What the mechanism is.** The adapter detects a routing condition against the table
*its own binary ships* and reports it in the success response as `reported_conditions`;
the chassis coordinator persists each to `agent_error_log` (severity `warning`,
`resolved=false`, attributed to the reporting service) at `handleCompleteResponse` —
the sole crossing point for every remote completion, reachable only after the atomic
per-request claim, so redelivery cannot duplicate rows. Three codes:
`UNROUTED_IMAGE_KIND`, `UNRECOGNISED_PROVIDER_HINT`, `REFERENCE_ANCHORS_DROPPED`.
Neither service re-derives the other's config; the coupling is one field name plus one
consumer-owned allowlist.

**Two design constraints came from the council and must not be undone:**
1. **The sender allowlist is load-bearing** (`conditionReportingAgentTypes`, currently
   `{image-generator}`). A guardian **hard veto** (round 5) established that an
   unconditional parse in shared dispatch is a fleet-wide ungoverned reporting bus.
   Adding a sender IS the review. `senderMayReportConditions` is the single swap seam
   if architecture review later prefers declaring via `agent_definitions.output_contract`.
2. **Absent ≠ malformed ≠ partly-dropped.** Absent is silent and healthy; a
   present-but-unusable field warns "the reporting contract broke"; a *mixed* list warns
   with an explicit dropped/parsed count. Collapsing any of these reintroduces the exact
   silence the mechanism exists to cure (`bug_historian`, rounds 5 and 7).

**Live-verified** on both replicas of both services, greping strings the change
*creates* (never `case` values): chassis `UNSANCTIONED sender`=1, `Persisted
adapter-reported conditions`=1, `the reporting contract broke`=1; adapter
`UNROUTED_IMAGE_KIND`=1, `REFERENCE_ANCHORS_DROPPED`=1 on `-6df8q` and `-drwlg`.

**How it reached prod without approval:** an owner sweep commit (`bca5d8255`) took the
then-uncommitted tree into the v1.0.1140 build. **No `Council-Reviewed` trailer is
claimed and none may be** until an APPROVED verdict. Disclosed to the council in the
round-8 rationale.

**Owed, in order:**
1. Round 8 verdict (corr `e996bf0a`, orchestration `49512359`). If APPROVED, commit the
   trailer against the landing commits (`8ec9e2ab8` + the swept `bca5d8255`).
2. **Live-fire proof, deliberately not yet run:** one generation with an unrouted kind,
   then `SELECT error_code, severity, agent_type FROM agent_error_log WHERE
   error_code='UNROUTED_IMAGE_KIND' ORDER BY occurred_at DESC LIMIT 1`. Deferred because
   the design was still under review and it costs a real generation plus a junk asset;
   do it once the verdict settles.
3. A `doc_note` recording the contract against diagnosis item `5db192c5`, which **stays
   open** — it names the wider unmatched-case family (`directionAppliesToKind`, the
   per-kind accessors); this fixed only the routing-observability member.
4. A guideline side-task raised by the `guidelines` seat (twice): DECLARED CONTRACTS has
   no clause for chassis-parsed generic response fields (headers already work this way,
   undeclared). Worth an explicit exemption clause so future reviewers do not relitigate.

Commits: `8ec9e2ab8` (per-entry drop counting) · `bca5d8255` (owner sweep carrying the
rest) · `58a7c7a8d` (`bugs_open/019` reproduction #3 — a council round of this work was
voided by the truncated-reviewer defect, still live pre-roll).
