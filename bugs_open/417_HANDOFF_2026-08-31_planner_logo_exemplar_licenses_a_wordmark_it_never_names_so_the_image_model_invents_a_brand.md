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

---

## RESIDUAL TAKEN AND CLASS FIX SHIPPED — bugfix 417/420 lane, 2026-08-31 (transfer accepted from the loanzy lane)

**Status: the STRUCTURAL fix is committed (`8bcd4ccae`) and INERT until the next chassis roll.
The data wash (migration 680) IS applied and live. 417 stays OPEN until the roll, per the
fixed-AND-live bar.** Council: `Council-Submitted: bb099a3d-0555-4fcf-b12a-31652b59f8b9`
— verdict NOT yet read, so nothing here should be taken as an approved verdict.

### 1. It had already FIRED a second time, on the first paid site — measured, not inferred

The boxingonline.com session downloaded and LOOKED AT the served asset
(`/assets/images/logo.png`, 400×218, `assets` id `20ce80fb…`, created 12:56:10Z). **It reads
"BOXING NEWS". The site is called Boxing Online** — `company_name = 'Boxing Online'`, the img
alt is "Boxing Online", the order is "Boxing Online", the domain is boxingonline.com. Nothing
anywhere calls it Boxing News. Same mechanism as "Farm Shield Info", live in the header of 19
pages with the favicon and og_card derived from it.

Its `origin_prompt` is the race-tail plan row `b56182fa` verbatim.

### 2. Why 669+670 could not have stopped it — the two class points

- **The race tail.** `b56182fa` was created **12:36:55Z, 41 seconds after 669 applied at
  12:36:14Z**, by a planner run already in flight. **A migration that fixes a prompt SOURCE
  cannot bound prompts already in flight**, and the honest post-fix census is therefore dated
  after the last in-flight plan lands, not at migration commit.
- **The paraphrase.** The row says *"no text **other than** the wordmark itself"*; the exemplar
  said *"**outside**"*. 670's surgical arm keyed on the literal and could never have matched it.
  **The model REWORDS the licence, so no literal match bounds this class** — and this sharpens
  §3's own warning: counting the licence by literal is still a literal.

### 3. The structural root, and why the fix is where it is

**The estate's no-lettering rule was coupled to the prompt's SOURCE, not to the asset's
PURPOSE.** It lived only inside `composeBrandImagePrompt` — the FALLBACK builder, reached only
via `if prompt == ""` in `load_work_item_actions.go:2391-2399` and
`check_placeholder_image_in_use.go:120-145`. Every planner-built site supplies a prompt, so the
rule governed exactly the population that never needed it. This file's own line — *"the ruled
path is the fallback nobody reaches"* — turns out to be the whole diagnosis.

The fix applies the rule at `GenerateImageAction`, after `getImagePromptWithPriority` has
collapsed all three prompt tiers into one value and beside the `kind`-gated block that already
makes logo-specific decisions. That placement is the point: it governs **every producer, present
and future, and work items ALREADY QUEUED with an unwashed prompt** — which is exactly what a
config migration structurally cannot reach. **The guard needs no detection of the licence at
all**, which is why it is a bound where 669/670 are a floor.

### 4. A measured fact that inverts an assumption this file rested on

The brief for this work asserted, from `bugs_closed/028`, that the provider discards negative
clauses. **That is 028's PRE-fix state, and its fix `foldNegativeIntoPrompt` is live.** The
adapter log for the failing generation proves the prohibition was DELIVERED:
`2026-08-31T12:55:50Z banana_provider "folded NegativePrompt into positive prompt as a
prohibition clause"`, kind=logo, negative list including `"text"`, `prompt_len 232 → 407`.
**The model received "no text" and lettered BOXING NEWS anyway, because the positive prompt
licensed a wordmark. A folded negative LOSES to a positive licence in the same prompt.**
So the clause is positive-framed and explicitly voids earlier wording; the negative channel is
belt only. Corollary worth keeping: the rule that stood in `default_brand_prompt.go` was
negative-framed, and therefore **weaker than its own comment claimed, for as long as it stood.**
(Logged in `WRONG_CALLS.md` — I quoted a closed bug's diagnosis and skipped its remedy.)

### 5. The owner's ruling on candidate 1's open question, and the opt-in

Put to the owner directly, 2026-08-31: **logos carry no words by default, and a lettered logo is
allowed only where the EXACT string is named.** So `constraints.wordmark_text` — an opt-in field
per the 2026-08-02 shape, unsafe side OFF, riding the existing `input_data.spec.constraints`
seam — whose **value IS the exact text**, making "a wordmark" with no text named inexpressible.
Validated against the site's own `company_name` / `logo_text` / domain stem; a mismatch
**degrades to text-free with a Warn, never refuses** (bugs_open/210's unhandleable-item lesson).
That grounding check is what closes the door against the field's own producer: a planner LLM can
write this field, so *"the field exists"* cannot be the licence — *"the field names THIS site"*
is. It also gives farmerinsurance.uk's deliberate wordmark a durable, auditable home.

**The vigilant_designer lane's §3 warning was right and is now costed:** the opt-in is REQUIRED,
not polish. ~8 of the 28 current-plan logo prompts name their wordmark **on purpose** (cv1
'CareerPrep', idea.uk, oufe, relojistas, robot-hands, webdesign.uk, lendzy, loanzy), and four of
those never use the word "wordmark" — further live proof of the paraphrase point. An
unconditional text-free clause would have wrecked them on regeneration.

### 6. Census that would DISCONFIRM this — keyed on the concept and the guard's own mark

Post-roll, for every `kind=logo` generation: `assets.origin_prompt` must contain
`LogoTextFreeSentinel` or a wordmark clause. **Any row with neither = the guard was UNREACHED on
some path** (check `kind` arrival first — see the blind spot below). Then the obedience canary:
generate from a FRESH paraphrase no migration ever matched, download the PNG and **look**; 5+
runs per 028 §6. Clause present + lettering still in the image = a prompt-engineering failure,
which is a *different* finding from the guard not firing, and only the origin_prompt mark
separates them. **A literal grep of stored prompts is a floor and is not a pass criterion.**

### 7. Stated blind spot — not discovered later

`site-work-orchestrator` and `pageflow-builder` (step `call_logo_generation`) map **no `kind`**,
so `resolveKind` returns `""` and the kind-gated guard is blind on those paths. `input_mapping`
has no literal syntax, so there is no config-only fix. **Their liveness is [UNVERIFIED]** — the
`orchestration_states` probe returned zero, and so did a known-live control, so it proved
nothing and was discarded rather than reported. `LANDMINES.md` entry added; census
disconfirmation (§6) is the live detector. A Go-side fix is deliberately not designed here.

### 8. Owed, and to whom

- **boxingonline.com's logo must be REGENERATED** — row `b56182fa` is washed (680) so a
  regeneration now produces a text-free mark, but the served asset still says BOXING NEWS.
  Owner ruled the site is fixed before the delivery email; the delivery lane owns the dispatch.
  Favicon + og_card must be re-derived AFTER it lands — presence-based discovery will not refile
  them (this file's §3, and 322).
- **`sites.logo_text` is NULL on boxingonline while `company_name` is 'Boxing Online'.** The
  "text-free mark + brand name in HTML beside it" contract only holds if something renders the
  name — and chrome currently suppresses the visible name when a logo image exists (alt text
  only). Chrome/design-family follow-up; named here so it is not lost.
- **A SECOND, DIFFERENT DEFECT in the same asset, which this fix does NOT address and which
  wants its own file:** the served logo is not a logo, it is a **two-panel design comp** — the
  mark on dark navy left, the mark plus lettering on light grey right, a presentation board
  scaled into a header slot. Even text-free, this asset is unusable. The mechanism is
  *generation output accepted structurally unexamined* (store, deploy, favicon-crop and header
  all took it), which is the trust-the-status shape of 012/028 at the pixel layer — **not** an
  input-licence problem, so 417's guard cannot catch it. Cheapest first check is dimensional (a
  `kind=logo` asset far outside a sane aspect envelope; `kindDefaults` asks 1024×1024, this is
  ~2:1) and needs no vision model; the two-background-fields signature is a classifier and
  inherits classifier gaps, so cost it separately. Found by the boxingonline session.
- **Candidate 2 (pixels vs identity) remains UNBUILT and deferred**, exactly where the
  vigilant_designer lane left it: architecture-scope by the 2026-07-29 test, the owner's call,
  and it would be a guarantee conditional on a classifier, inheriting its gaps.

---

## COUNCIL ROUNDS 2 AND 3 — what the gate found that I did not, and the three axes of this bug

Rounds 1 and 2 both returned REVISE. **Every objection was correct, and two of them found real
defects in code I had already committed.** Recording them here because the pattern is more useful
than the fix: the gate kept finding *this bug's own shape* in my fixes for it.

### The three axes of 417, only the first of which was in the original file

417 is "a logo prompt reaches the image model ungoverned". There turn out to be three distinct
routes, and each of my first two attempts closed one while leaving another open:

1. **The prompt SAYS the wrong thing** (the filed defect — the exemplar licensed an unnamed
   wordmark). Closed by 669/670, bounded by the choke-point guard.
2. **The generation is never IDENTIFIED as a logo**, so no per-kind policy applies (round 1's
   HIGH: the two legacy parents map no `kind`). I had documented this in a risks block.
   `bug_historian`: *"Disclosing the gap in risks does not close the exposure."* Correct — a
   risks block is not a control. Closed in round 2 by resolving intent from every available
   signal including the step name.
3. **The generation is identified WRONGLY** — a caller states a non-logo `kind` for what is
   actually a logo (round 2's MEDIUM). My round-2 fix *believed* any stated kind, so this route
   skipped the policy with no Warn, no note and no error at all. `bug_historian` again:
   *"reproducing bug 417 on a THIRD axis, and this one gets none of the round-2 detection the
   author built for the other two."* Closed in round 3 by recording a `Conflict` when a stated
   kind disagrees with the step's own name. **The stated kind still wins** — a classifier
   overriding a caller is the worse failure — but the disagreement is now filed as
   `image_kind_conflict` rather than swallowed.

### The catch that mattered most: my compensating control was broken on arrival

Round 2 answered the legacy-parent gap partly with a durable detector — "if those parents are
dead the table stays empty, and that empty IS the answer". Three seats (`tooling_provenance`,
`constitution`, `guardian`) converged on the hand-rolled `doc_notes` INSERT behind it.
`[MEASURED 2026-08-31]` `doc_notes_subject_type_check` permits exactly eight values —
`tool, pipeline, experience, action, experience-pattern, landmine, component, decision` — and I
had used **`subject_type='site'`**. Every insert would have failed, and because a note writer is
best-effort, it would have failed **silently**.

The guardian named the consequence precisely: it *"WOULD silently null out the exact detector
round 2 leans on as its liveness measurement."* A later reader would have queried the category,
found zero, and concluded the condition never occurred — when it was never recordable.

**Fixed by REUSE, not by patching the literal**, which is the more durable lesson and was
`reuse_agent`'s separate objection: `recordImagePolicyEvent` now goes through
`LogActionEntryInheritingProvenance` / `agenterrors.Entry`, the estate's action-level recording
family. **A reused writer cannot get its own table's constraints wrong** — so the reuse habit
would have prevented this without anyone knowing the constraint existed. LANDMINES entry added.

### Other objections acted on

- **`reuse_agent` (HIGH, gating round 2):** I wrote a parallel `resolveLogoIntent` beside the
  file's existing `resolveKind`, for the same axis, in the file the diagnosis cites. Deleted;
  `resolveKind` is now the one resolver, extended.
- **`editquality` (MEDIUM):** "identified as non-logo" and "nothing identified it" were
  indistinguishable, so a hero call setting `purpose` but no `kind` would have filed a FALSE
  no-kind note — polluting the very measurement the note exists to provide. `Answered` now
  reports whether any source spoke.
- **`guardian` (MEDIUM):** the step-name heuristic had to be ENUMERATED, not asserted — and the
  enumeration changed the code. `[MEASURED 2026-08-31]` every live step name containing
  logo/hero: `call_logo_gen`, `call_logo_generation`, `generate_logo`, `store_logo_asset`,
  `deploy_logo_image`, `check_logo_or_hero`. **The last names BOTH kinds**, so the hint now
  abstains on ambiguity rather than guessing.
- **`prior_art_librarian` (MEDIUM):** the "`wordmark_text` has zero occurrences" claim is now
  FALSE as stated, because my own round-2 code is in the tree. Restated with its baseline —
  zero *before this change* — which is the only form of an absence claim that survives its own
  fix landing.
- **`debug_historian` (MEDIUM):** my roll-verification instruction (the `build provenance` log
  line) is the method the estate's own landmine calls unreliable. Now a pod-grep of the RUNNING
  binary for this change's literals, with a present-and-absent control pair.

### What is STILL open, stated precisely rather than as "covered"

A caller supplying no `kind`, no `default_kind`, no `purpose`, no step-config kind/purpose, and
whose step name names neither "logo" nor "hero", is **still ungoverned** — it now files
`image_generation_without_kind` instead of passing silently. A caller that MISLABELS is still
obeyed, and files `image_kind_conflict`. **Both are detections, not preventions**, and they are
labelled as such. The honest response to either firing is to fix the caller's `input_mapping`,
not to add more arms to the heuristic — which is the `architecture` seat's standing LOW, accepted.

---

## APPROVED (round 4) AND LIVE — but the census that would DISCONFIRM it has no subject yet

**Council: APPROVED, round 4, `bb099a3d-0555-4fcf-b12a-31652b59f8b9`** — "approved with 2 advisory
objection(s) — none high-severity". ⚠ The two advisories' text is **no longer retrievable**: the
`orchestration_states` row has aged out of the rolling window (0 rows for this correlation as of
2026-09-02) and the surviving `doc_notes` verdict carries only the summary line. Recorded as a
limitation rather than as "there were none" — this is the estate's own
*a closer census cannot see what it succeeded at* shape, arriving on a verdict.

### Live at the artefact, with a control pair — not inferred from the roll

Chassis rolled **2026-09-01 21:00:33Z** (both replicas). Probed the RUNNING binary, per this
lane's own runbook (never the `build provenance` log line — the estate's landmine calls that
unreliable, and the debug_historian seat objected to it in round 3):

| needle | result | what it proves |
|---|---|---|
| `image_kind_conflict` | **PRESENT** | round-3 code is in the binary |
| `Render a text-free mark` | **PRESENT** | the policy clause is live |
| `resolveLogoIntent` | **absent** | **removed-string control** — round 3 DELETED this symbol, so its absence proves the binary is round 3, not round 2 |
| `zzz_needle_that_cannot_exist_417` | absent | the grep can return zero, so the PRESENTs mean something |

The third row is the one that matters: a present-only check could not have distinguished round 2
from round 3, and the deletion gave a free removed-string control.

### The census, stated honestly `[MEASURED 2026-09-02]`

**Disconfirmation A (did the guard REACH every logo generation?) — NO SUBJECT YET.**
Zero logo assets have been generated since the roll. The last `asset_key='logo'` row anywhere is
**2026-08-31 12:56:10** — boxingonline's, the bad one. **A zero here is not a pass**; there is
nothing to have passed.

**Disconfirmation C (obedience) — NO SUBJECT YET**, same reason. It needs a real generation and a
human looking at the PNG.

**Disconfirmation D (not over-applied) — one positive data point.** 7 assets were GENERATED
post-roll (`origin_model` non-empty), so `GenerateImageAction` demonstrably ran; the most recent
is a `content_hero`, and it carries **no** policy sentinel — correct, and the first live evidence
that the guard does not contaminate non-logo prompts.

**The kindless detector — 7 opportunities, 0 fires.** `image_generation_without_kind` has zero
rows while `agent_error_log` took **1,045 rows since 2026-09-01** (demand control: the table is
being written) and `agent_type='unattributed'` reads **0** (the provenance path is healthy). So
every post-roll generation supplied a resolvable kind. **That is evidence the legacy-parent path
is dormant — it is NOT proof those parents are dead**, and it must not be written up as one. It
is, however, strictly more than the liveness probe could give in round 2, which returned zero for
a known-live control and therefore said nothing at all.

### The convergence worth acting on

**boxingonline's logo regeneration is both the customer fix and this fix's canary.** It is still
owed (the served asset still reads "BOXING NEWS"), its plan prompt is washed by migration 680, and
firing it would give disconfirmations A and C their first subject in the same run that repairs the
first paid site. Whoever fires it should download the PNG and **look at it** — and also check
`bugs_open/421`, because a text-free two-panel design comp is still unusable.

---

## CENSUS RUN AGAINST THE FIRST POST-ROLL LOGO GENERATION (2026-09-02) — A PASSES, and it exposes a designed-in weakness the estate has already measured

The boxingonline regeneration fired and completed today (work item
`0aa6cf1d-3ae0-4939-8b58-a7fb6bb9746e`, `needs_imagery`, complete 10:25:39Z), giving
disconfirmation A its first subject. **Read the generation side; the served file is the delivery
lane's half and is still mid-flight.**

⚠ **First, a trap I nearly fell into.** The asset row is `20ce80fb…` — **the SAME row as the bad
08-31 logo**, because the store path upserts (`origin_prompt = COALESCE(EXCLUDED.origin_prompt,
…)`). Its `created_at` still reads `2026-08-31 12:56:10`. **So "when was the last logo generated?"
answered by `max(created_at)` says 08-31 and is WRONG** — the row was updated in place. Anyone
censusing generations by `created_at` on `assets` will miss every regeneration.

### Disconfirmation A — PASSED, and the needle had to be chosen carefully

`origin_prompt` carries my clause. **But my sentinel and migration 680's wash clause differ only
by a capital letter** — 680 appends *"…: render a text-free mark with no lettering or words of any
kind"* (lowercase, after a colon) and the Go clause is *"Render a text-free mark: a single
pictorial symbol…"*. A `LIKE '%Render a text-free mark%'` looks decisive and would have matched
the Go clause only by luck of Postgres's case sensitivity. Settled with two **Go-only** needles
instead:

| needle | present | source |
|---|---|---|
| `Render a text-free mark: a single pictorial symbol` | **true** | Go guard only |
| `overrides any earlier wording in this prompt` | **true** | Go guard only (680 says *"is void"*) |
| `render a text-free mark with no lettering` | true | migration 680's wash |

**So the guard reached a real generation and did its work.** That is the first live evidence for
the whole placement argument — a planned prompt, from a real customer site, governed at the choke
point.

Detectors: `image_generation_without_kind`, `image_kind_conflict` and `logo_wordmark_rejected` all
have **zero rows** — correct for this run (the kind resolved, declarations agreed, no wordmark
opt-in was requested).

### ⚠ But the composed prompt shows the weakness, and the estate measured it two days ago

The full `origin_prompt`, in order: **the original plan text — still containing
*"no text other than the wordmark itself"*, the exact licence that produced "BOXING NEWS"** —
then 680's wash, then the owner's 2026-09-02 transparency ruling, then my clause.

**Four instructions about text are co-present, and mine relies on precedence language to win.**

`bugs_closed/390`, filed 2026-08-31 and already in 016b §9, measured precisely this:

> *"the fixer's prompt carried a computed, case-specific instruction … AND an older general
> instruction …, with prose saying the specific one supersedes. **The model obeyed the general
> one.** Two co-present instructions are adjudicated by the model, not by precedence language —
> the fix is to FENCE them in the template so only one ever renders."*

**My design is the shape 390 says does not reliably work.** 670/680 appended rather than rewrote,
deliberately, so the original wording stayed auditable — and that choice is what leaves the licence
co-present in every washed prompt. The Go clause then adds a fifth voice rather than removing the
first.

This does **not** retract disconfirmation A: the guard demonstrably fires and its text is
demonstrably delivered. It bounds what A can prove. **A proves the instruction ARRIVED. Only the
PNG proves it was OBEYED**, and 390 is direct evidence that arrival does not imply obedience when
a contrary instruction sits in the same prompt.

Note 390's other instruction, which this census followed: *"read `prompt_rendered` … for what was
actually co-present, never the template for what you meant."* Reading the composed `origin_prompt`
rather than my own constant is what made the co-presence visible at all.

### Disconfirmation C is now the binding test, and the instrument is a human eye

The delivery lane downloads and LOOKS when the publish wave lands. Their check list is this file's
plus the new ruling: text-free (417), single composition (421), **transparent background** (owner
2026-09-02 — the 08-31 regen shipped a baked dark ground), and near-square-or-declared.

**What each outcome means for THIS fix, stated before the result so it cannot be rationalised
after:**

- **Text-free →** the override held on this case. Good evidence, *not* proof of the mechanism —
  390's failure was probabilistic, and one clean generation cannot establish a rate (the estate's
  own *two clean runs cannot establish stability*). It would still be worth fencing.
- **Any lettering →** the precedence language lost, exactly as 390 predicts, and the fix is a
  FENCE: the guard must remove or neutralise the licence in the composed prompt rather than
  argue with it. That is a real design change and it re-opens this file.

**I am deliberately not writing the fence now.** The artefact decides which fix is needed, it is
hours away, and building the more invasive version first would be choosing before the evidence —
which is the habit this whole bug is about.

---

## DISCONFIRMATION C — the PNG was read, and the answer is TEXT-FREE (2026-09-02)

The `site_delivery_and_editor` lane downloaded the served asset and looked at it. **Zero
lettering.** Served file sha256 `1abcf69c08ab4462`. Single composition too, so `bugs_open/421` is
clear on this asset as well.

**Both halves agree**: this lane's census showed the guard's clause reached the generation
(disconfirmation A), and their eye shows the model obeyed it. That is the first end-to-end
evidence for the fix.

**Held to the pre-registered reading, and it is worth restating because the temptation runs the
other way:** this is **good evidence the override won THIS case, not proof of the mechanism.**
`bugs_closed/390`'s failure was probabilistic, and the estate's own rule is that *two clean runs
cannot establish stability* — this is one. **The fence is still worth building**, because the
original licence (*"no text other than the wordmark itself"*) remains co-present in the composed
prompt and my clause still wins only by precedence language, which is the shape 390 measured
losing.

**Status: 417's residual on this asset is CLOSED. The file stays OPEN**, on two counts: the fence
(a design change, not a residual), and the bar — a fix is not closed until it is fixed AND live
AND the class is bounded, and one generation does not bound it.

**What would settle the mechanism rather than the case:** the next several logo generations across
different sites, each censused for the clause and eye-checked. If any comes back lettered while
carrying the clause, the fence is required and this file re-opens as a design change. Until then
the honest statement is *"one for one"*, and it should be written that way wherever it is quoted.

**A third defect was found on the same asset and is NOT this bug:** `bugs_open/424` — the
transparency ruling produced a painted checkerboard, because alpha is a file-format capability no
prompt can request. Text-free and single-composition both held; the file is still unusable. Three
independent properties, three separate mechanisms, one image.

---

## THE FENCE — decision recorded rather than taken silently (2026-09-02)

`bugs_closed/390`'s remedy for co-present instructions is *"FENCE them in the template so only one
ever renders"*. My guard does not fence: it APPENDS a clause that claims precedence over the
licence still sitting earlier in the same prompt. So the question is live, and this section exists
so the next session does not have to re-derive the reasoning or guess which way it went.

**Decision: NOT building the fence now.** Stated with its grounds, so it can be overturned by
evidence rather than by preference.

**Why not:**
- **The evidence is n=1.** One generation came back text-free. The estate's own rule — *two clean
  runs cannot establish stability* — cuts against acting, and it cuts against complacency equally.
- **The real fence is invasive.** Appending cannot fence; only RECOMPOSING can — presenting the
  plan's text as subject-matter and letting constraints come solely from the policy block. That
  changes how every image prompt is assembled, for every kind, to fix a risk currently measured at
  zero occurrences out of one opportunity.
- **The obvious cheap version is the thing this bug is about.** Stripping the licence by matching
  its literal is detection, and 417's whole finding is that the model rewords the licence, so a
  literal is a floor. A fence built that way would inherit exactly the weakness that made
  migrations 669/670 insufficient.
- **Building the invasive version on one data point is the habit this bug exists to correct** —
  choosing before the evidence.

**Why it might still be right, so this is not a dismissal:** the failure mode is SILENT. A lettered
logo ships, passes every automated check, and is caught only when a human opens the image — which
is how both known instances were found. A risk that is invisible to instrumentation deserves more
caution than its measured rate suggests.

**The trigger, so this is decidable rather than perpetual:** the next several logo generations,
across different sites, each censused for the clause (`origin_prompt` carries it) and eye-checked.
- **Any lettered logo that carried the clause → build the fence.** That is the design change, and
  this file re-opens as one.
- **A run of clean generations → the override is holding**, and the honest write-up is the RATE,
  never "it works". Name the failure rate the sample could have detected.

**A cheap intermediate someone may want instead of the fence:** count how often we are relying on
adjudication at all — logo prompts whose `origin_prompt` contains BOTH a wordmark licence and the
policy clause. ⚠ That is literal-matching, and it is legitimate HERE precisely because it is a
MEASUREMENT and not a bound: a floor is a perfectly good instrument, and only fails when mistaken
for a guarantee. It would turn "we do not know the exposure" into a number, for a few lines.

**Not built here either** — it is a new check, and the class fix this file was opened for is done,
approved and live. Recorded so the option is visible to whoever picks the trigger up.

---

## THE FENCE TRIGGER — first readings from the FIXED pipeline (2026-09-03, both fixes live)

The trigger above asks for *"the next several logo generations, across different sites, each
censused for the clause and eye-checked"*, with the outcome written **as a rate, never as
"it works"**. Here are the first generations produced with BOTH 417's prompt fix (`b2322a203`) and
424's guard fix (`fcbe6071c`) live together (`v1.0.1356`, rolled 08:57Z; re-confirmed as ancestors of
the later `v1.0.1358` adapter stamp `d0252fd4d`, with a negative control).

| site | generated | key date dir | prompt carried licence + override | lettering? | verdict |
|---|---|---|---|---|---|
| seotools.co.uk | 09:30Z | `20260903/` ✅ new | **both** (`wordmark=t`, `text_free=t`) | **none** | clean |
| gamedesign.uk | 11:41Z (attempt 3) | `20260903/` ✅ new | **both** | **none** | clean |
| designblog.co.uk | — | `20260902/` (unchanged) | n/a | **no artefact** | see below |

Both clean marks were eye-checked at the **served bytes** with a 404 invented-path control:
seotools a magnifying glass over a woven lattice; gamedesign an abstract maze in terracotta and tan.
Single composition each, zero lettering, no invented brand. **421's two-panel shape did not recur.**

**Both were the adjudication case.** Each prompt still contained the wordmark licence *and* the
text-free override, so in both the override had to win on the model's own reading — which is exactly
the situation the fence would remove by construction.
⚠ The clause census was **re-run against the post-regeneration rows**, not carried over from the
morning's. gamedesign's row had been censused before its 11:41Z regeneration; an `origin_prompt` is
replaced by the UPSERT, so the earlier reading described a prompt that no longer existed. Re-checked
at `updated_at = 11:41:02Z` on the `20260903/` key: both clauses present.

### designblog contributes NOTHING to this trigger, and that must not be miscounted

Its three attempts were all **guard refusals** — `border_keyed=0.000, want >= 0.95 — refusing to
store` (item error, 11:36:58Z), correctly storing nothing. **No artefact exists to eye-check**, so it
is neither a clean generation nor a lettered one. Counting a refusal as a clean run would inflate
this trigger's evidence base with runs that never tested the prompt at all.
⚠ Checked explicitly that this was a refusal and not a run killed by the 12:06Z chassis roll: the
refusal is timestamped **before** the roll and carries the guard's own statistic. A killed run leaves
no such line while still consuming an attempt, and the two are otherwise indistinguishable from the
item status alone.

### The rate, and the failure rate this sample could actually detect

**8 clean generations, 0 lettered, as of 2026-09-03** (6 eye-checked pre-fix per the 09-03 handoff,
plus seotools and gamedesign post-fix).

**That bounds the lettering rate at roughly ≤ 31% with 95% confidence** (rule of three: 3/8 ≈ 0.31)
— and that is the honest statement. It is a **weak** bound. It does not license *"the override
works"*, it does not close the trigger, and the failure mode remains **silent**: a lettered logo
passes every automated check and is caught only when a human opens the file.

**Recommendation, unchanged in direction and now better grounded: keep the fence UNBUILT, keep the
trigger OPEN.** Nothing has refuted the override, and 2 of the 2 real post-fix opportunities came
back clean — but 8 runs cannot distinguish "reliable" from "fails one time in five".

### ~~The exposure figure the fence section asked for is now answerable — and it is total~~

> **CORRECTED 2026-09-03, same session, ~1 hour after I wrote it. The measurement below was WRONG,
> and wrong in the direction that flattered the conclusion.** What caught it: the 424 lane messaged
> to say it was resetting `designblog` for another retry round, which sent me to read designblog's
> full `origin_prompt` — and the licence-detector fell apart on contact with the actual text.

**What I claimed:** *"Measured 2026-09-03: 5 of 5 sampled sites carry BOTH"* a wordmark licence and
the policy clause (`origin_prompt ILIKE '%wordmark%'` AND `LIKE '%Render a text-free mark%'`), and
therefore *"the override is load-bearing on 100% of logo generations"*.

**Why it is false.** **The override clause itself contains the word.** It ends:
*"…overrides any earlier wording in this prompt that mentions, permits or presupposes a **wordmark**
or any text."* So every prompt carrying the override matches `%wordmark%` **because of the
override**, and my licence-detector was matching the prohibition. The census could not have come out
any other way — which is the estate's own test for whether a `[MEASURED]` figure is evidence at all,
and this one fails it.

⚠ **This is the RUNBOOK's documented trap, in mirror image, and the RUNBOOK is three lines long
about it**: *"⚠ The exemplar phrase itself matches `no text`, so 'does the prompt forbid text?'
scores the contradictory prompts as SAFE. And counting the licence by literal is still a literal."*
I read that warning, wrote the census to avoid counting the prohibition — and then counted the
prohibition, in a new place, because the override *quotes the licence in order to overrule it*.

**Re-measured, stripping the override sentence before looking** `[MEASURED 2026-09-03]`:

| domain | naive `%wordmark%` | licence OUTSIDE the override | `letterform` | `typographic` |
|---|---|---|---|---|
| boxingonline.com | t | **f** | f | f |
| designblog.co.uk | t | **f** | **t** | **t** |
| gamedesign.uk | t | **f** | f | f |
| seotools.co.uk | t | **f** | f | f |
| websitepromotion.co.uk | t | **f** | f | f |

**The real picture is the opposite of what I wrote, and it is more interesting.**

- **0 of 5 carry a wordmark licence outside the override.** Migrations 669/670 worked: the
  exemplar's licence is gone from the composed prompt. The override is redundant on four of five.
- **1 of 5 carries a licence at all — `designblog.co.uk` — and it does NOT use the word "wordmark".**
  Its subject line asks for *"abstract **letterform** or **typographic symbol** suggesting editorial
  authority"*, then forbids *"lettering or words of any kind"* **in the same sentence**.

**That single row is 417's whole thesis reproducing itself.** This file's finding is that *the model
rewords the licence, so a literal is a floor*. The licence did not survive as "wordmark"; it survived
as **"letterform"**, which no literal aimed at 669/670's phrasing would ever catch — including mine,
written by the session that knew about the problem.

### What this does to the rate above — it weakens it badly, and specifically

The 8-clean / 0-lettered bound treats the generations as exchangeable. **They are not.**
`designblog` is the only prompt in the population that actually contains a licence, so it is the only
run that genuinely tests adjudication — **and it has never produced an artefact.** Three attempts,
three transparency refusals, nothing stored.

**So my sample systematically excludes the only case that tests the thing I am measuring.** The
other 8 runs show the model not painting text when nothing asked it to, which is a much weaker claim
than "the override beats a licence". The honest bound on the adjudication case is **n = 0**.

**The recommendation survives but its reasoning is now inverted.** Keep the fence UNBUILT and the
trigger OPEN — but not because the evidence is reassuring. Because **the evidence on the case that
matters does not exist yet**, and the 424 lane's next `designblog` retry is the run that would
produce it. ⚠ **When it lands, eye-check it before anything overwrites it** — a regeneration UPSERTs
the row and the artefact is gone.

**The cheap intermediate the fence section asked for still stands, with a correct detector**: count
prompts carrying a licence *after* removing the override sentence, and do not key on `wordmark`
alone — `letterform`, `typographic`, `monogram`, `initial` and `lettering` are all the same licence
in different clothes. Currently **1 of 5** `[MEASURED 2026-09-03]`.

**See also `bugs_open/462`** — a separate gap found while eye-checking these: the marks are text-free
and correctly matted, and nothing in the estate checks whether one is legible against its header.
