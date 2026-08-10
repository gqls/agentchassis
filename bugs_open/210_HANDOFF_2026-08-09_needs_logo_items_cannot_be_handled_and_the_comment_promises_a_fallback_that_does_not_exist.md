# 210 — `needs_logo` items are unhandleable, and the producer's comment promises a fallback that does not exist

**Filed 2026-08-09** from the fundamentallyai improvement-sweep front. **OPEN.**
~~Status: diagnosed first-hand, **not fixed**~~ — **FIX BUILT + COMMITTED 2026-08-09, council
APPROVED round 1; stays OPEN until it is live on the fleet and pod-verified. See the section
at the foot of this file, which also CORRECTS §6's mechanism (cross-origin, not basename) and
§2's producer count (three, not one).** The one-line fix is a trap (§4) and is not taken; §5's
design choice is still open and unmade, now with new evidence against option 2 for logos.

**Diagnosis norm (CLAUDE.md):** this is a cross-cutting claim about a shared handler, so it
would normally go through `090_TRIGGER_needs_diagnosis_v1.sh` before being asserted. It has
**not** been run. Substituted first-hand verification, declared: the failing item's own error
text names the exact missing path, the live step config was read from `agent_definitions`,
the producer was read at the line that builds the spec, and the callee's `input_contract` and
prompt-resolution function were read end to end. Every claim below cites the artefact it came
from. A 090 run would still be worth its cost on §5's choice.

## 1. Symptom

`site_work_items` row `placeholder_image_in_use:logo` on fundamentallyai.com,
status **`failed`**:

```
step call_logo_gen failed: failed to execute action call_agent:
failed to extract data for agent: step : input_mapping failed:
source path 'input_data.spec.image_prompts.logo' not found for field 'prompt'
```

The error helpfully dumps available paths. `input_data.spec` contains exactly
`check`, `original_pipeline`, `path`, `purpose` — and no `image_prompts`.

## 2. Mechanism, in three reads

**(a) The handler demands the key.** `image-build-handler`, step `call_logo_gen`:
```json
"input_mapping": { "prompt": "input_data.spec.image_prompts.logo", ... }
```
`input_mapping` is a strict allow-list: `input_contracts/input_mapping.go:122-136` errors when
a source path is absent, **unless the destination field name ends in `?`** (`:101-102`,
`isOptional`). `prompt` does not.

**(b) The producer writes the key only sometimes.**
`discovery_checks/check_placeholder_image_in_use.go:76-96` loads the site's planned prompts
and then:
```go
if prompt != "" {
    spec["image_prompts"] = map[string]interface{}{promptKey: prompt}
    spec["prompt_key"] = promptKey
}
```
When `loadImagePromptsForSite` returns nothing for `logo`, the item is filed **anyway**, with
no `image_prompts` — and the handler cannot consume it. The finding even records
`"prompt_known": false`, so the producer *knows* it is filing a promptless item.

**(c) The comment above that code asserts a fallback that does not exist:**
> *"If the site_plan spec doesn't have the prompt either, the handler will fall back to its
> default prompt template — still useful, just less specific."*

It never reaches any fallback. It dies at input extraction, one step earlier.

**Blast radius:** not site-specific. Any site whose logo the placeholder detector flags, and
whose plan carries no `logo` prompt, fails identically. **`call_hero_gen` has the same shape**
(`"prompt": "input_data.spec.image_prompts.hero_home"`), and the same producer maps
`hero → hero_home`, so the hero path fails the same way.

## 3. What is NOT wrong

`image-generator` is innocent and its contract is right: `input_contract.required` is **empty**
and `prompt` is listed **optional**. `getImagePromptWithPriority`
(`generate_image_actions.go:831`) implements a real chain — parent step config → collected
data → `input_data.prompt` → agent `prompt_template` → step `prompt_template` → generic. The
callee was built to tolerate a missing prompt. Only the parent's mapping forbids it.

## 4. Why the one-line fix is a TRAP — do not ship it

The tempting fix is `"prompt"` → `"prompt?"`, making the mapping optional. **Do not**, on its
own. Measured 2026-08-09: `image-generator`'s `default_config` has **no `prompt_template`**
(Priority 2 absent), so with the field skipped the chain falls all the way through to
`generate_image_actions.go:895-897`:

```go
logger.Warn("No prompt found in any tier, using generic fallback")
return "Generate content based on the provided context.", "generic_fallback"
```

That would generate an image from a meaningless sentence and `store_logo_asset` would save it
**as the site's logo**. A loud failure becomes a silently wrong brand asset — strictly worse.
This is the shape where "make it not error" and "make it correct" point in opposite directions.

## 5. The fix is a design choice — options, costed

The question nobody has answered: **when a site's logo is missing and its plan never contained
a logo prompt, who decides what the logo looks like?**

1. **Give `image-generator` a kind-aware default prompt** (DB config, live immediately, no
   roll). Cheapest. Risk: one `prompt_template` is shared across logo/hero/imagery/variant, so
   a single generic string is wrong for at least three of them. Would need per-kind templates,
   which the current Priority-2 lookup does not model.
2. **Make the producer synthesise a real prompt** from `purpose` + the site's imagery style
   guide (`imagery_style_guide.go`) and `design_intent`. Honest, and it makes the existing
   comment true. Go change ⇒ build + roll. Objection to answer: is a discovery check the right
   place to author brand imagery prompts?
3. **Do not file an item the handler cannot consume** — file `needs_human_review` instead when
   `prompt_known` is false. Smallest blast radius and arguably the most honest: a missing logo
   with no planned prompt genuinely is a human decision. Loses automation.

**Recommendation: 3 now, 2 later.** 3 stops the fleet-wide failure immediately and cannot
produce a wrong asset; 2 is the real capability and deserves its own round.

## 6. A second defect in the same row — the finding was probably a FALSE POSITIVE

On fundamentallyai nothing local references `/assets/images/logo.png`. The site's logo is
`/assets/images/logo.jpg` (HTTP 200, `image/jpeg`, 60,897 B). The only `logo.png` in served
HTML is `https://leopardessconsulting.co.uk/assets/images/logo.png` — an **external partner
logo that returns 200**. So a basename appears to have been attributed to a local missing
path: the `bugs_closed/128` family (purpose/basename vs path), resurfacing in a check 128 was
meant to have fixed. `[UNVERIFIED]` — `check_placeholder_image_in_use`'s matching was not read
for this; only its spec construction was.

Its sibling row `image_url_404:logo.png` is **`blocked`** with *"No handler_agent set — item
cannot be routed to any agent"*, so that one can never drain at all.

## 7. How to verify a fix

```sql
-- the failing row
SELECT status, error FROM site_work_items
 WHERE site_id='199733a8-ac9c-4c30-b2ce-65ecdac6f3bd'
   AND item_key='placeholder_image_in_use:logo';

-- fleet-wide exposure: who else has a promptless logo/hero item
SELECT s.domain, w.item_type, w.status, w.spec ? 'image_prompts' AS has_prompt
  FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.item_type IN ('needs_logo','needs_imagery')
   AND w.status NOT IN ('complete','cancelled') ORDER BY 1;
```
A real fix must show a `needs_logo` item reaching `store_logo_asset` **with a prompt whose
source is not `generic_fallback`** — grep the chassis logs for
`"Using prompt from"` and assert the source, because a stored logo alone does not prove the
prompt was sane.

## 8. Related

`bugs_closed/128` (purpose/basename vs path, §6), the `input_mapping` allow-list LANDMINE
(a dispatcher has two gates), and `HANDOFF_2026-08-05b_improvement_sweep.md` §5.1 where this
was first recorded.

---

## CONTRIBUTION 2026-08-09 — §6's `logo.jpg` is NOT that site's normal logo; it is bug 231 firing

From the `bugfix_209_deploy_purpose_keyed_source` lane, in passing — no claim on this
file's own defect or its fix. (Note for searchers: this number is shared with
`210_HANDOFF_2026-08-06_a_content_failed_page_build…`; this contribution is about the
`needs_logo` file only.)

§6 reads fundamentallyai's `/assets/images/logo.jpg` (60,897 B) as the site's logo and
concludes the missing-`logo.png` finding was probably a false positive. The first half
needs correcting: **`logo.jpg` is itself the defect.** The correct artefact for purpose
`logo` is a **400×400 PNG** (`storage.ImagePurposes["logo"] = {400,400,90,"png"}`,
`url_helpers.go:364`); fundamentallyai's file is a **1408×768 JPEG**, and its producing
commit subject is "Deploy **hero** image for fundamentallyai.com" — a subject built from
the resolved purpose at `deploy_image_asset_action.go:579`. Eleven sites are in this
state; the census, the three corroborating signals and the `[UNVERIFIED]` producer
question are in `bugs_open/231`'s 2026-08-09 afternoon contribution.

This does **not** resurrect the §6 false-positive call — a checker flagging `logo.png`
by basename while the page references `logo.jpg` is still making the wrong comparison,
and §6's `[UNVERIFIED]` note on `check_placeholder_image_in_use` stands. It means the
underlying site is genuinely wrong too, so the check was **right for the wrong reason**;
"the site's logo is logo.jpg" should not be left in the record as the innocent
explanation.

---

## FIX BUILT + COMMITTED 2026-08-09 — by the `bugfix_210_needs_logo_unhandleable` lane

**Go only, no config half. Council `c40c9483-5afd-478b-91ca-7e4db505ed0d` — APPROVED, round 1**
(12 seats, 4 advisory objections, none high; verdict read in full and every objection
dispositioned in the lane NOTES). **STAYS OPEN** until the fix is live on the fleet and
pod-verified — the code is inert until the next chassis image, and a roll is not evidence
(`bugs_open/153`). Lane docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_210_needs_logo_unhandleable/`.
Registered as **IMG-069**; two LANDMINES entries added and the `bugs_closed/128` pair entry
corrected.

### Two corrections to this file, both load-bearing

**§6 reached the right verdict by the wrong mechanism, and the mechanism matters.** It guessed
"a basename appears to have been attributed to a local missing path" and marked it
`[UNVERIFIED]`. There is no basename matching in this check — it compares the full path. The
real defect is that `isPathReferencedInPages` used an **unanchored substring**
(`rendered_html LIKE '%' || $2 || '%'`), and a **cross-origin** URL contains that substring
byte-for-byte. fundamentallyai's only match, and the only `/assets/images/logo.png` match
**anywhere in the fleet**, is:

```
<div class="portfolio-logo"><img src="https://leopardessconsulting.co.uk/assets/images/logo.png"
```

A partner logo, served from the partner's own domain. Had the mechanism been trusted rather
than the conclusion re-checked, a correct finding would have been "refuted". Now fixed and
**verified in Postgres over live rows, both arms**: the false positive goes 1 → 0 while **all
141** legitimate same-origin `hero.jpg` matches across 17 sites survive untouched — a pattern
that silenced everything would have passed the first arm alone.

**§2's "blast radius" understates the producer count by two.** Three producers file these
items and **none** guarantees the key: this check; `WriteBuildItemsAction`
(`load_work_item_actions.go:348-364`), which is the **primary build path** and writes
`spec.image_prompts = planData["image_prompts"]` verbatim while `needs_logo` is an independent
boolean (`:196`) and `image_prompts` is defaulted to `{}` by a separate branch
(`v3_site_actions.go:3107`); and `check_unfulfilled_image_prompt`, which is safe only because
it never runs without a prompt. The second has not fired **by luck**, not by design.

### A fourth defect, and the magnitude nobody had measured

- **The check is blind to the surface a logo lives on.** It scanned `page_components` only; a
  logo is rendered by the site **chrome** (`site_components`). `bugs_closed/128` found this
  exact blindness and fixed it in the **flag-only** sibling (`check_image_url_404.go:471-482`)
  while leaving the check that **routes a regeneration** blind. Now fixed on the sibling's
  contract. **Impact at the time of fixing: nil** — four sites reference the path in chrome and
  all four have an active `logo` asset, so the second precondition skipped them regardless.
  A hole waiting, not a hole leaking.
- **The recovery branch is unavailable on 87% of the fleet** `[MEASURED]`: **33 of 39** sites
  have no current `site_plan` spec row at all, 1 has a row with no `image_prompts`, **5** have
  the object. The promptless branch is the NORMAL case. The only reason 2 items failed rather
  than 34 is that the *other* precondition (no active asset) is currently rare — and **13 of
  the 16 sites that do have an asset carry no planned prompt**, so each is one asset deletion
  away from this failure.
- **mortgagecalculator.co.uk is a TRUE positive** — 6 same-origin CSS `url('/assets/images/hero.jpg')`
  references, no active hero asset, item failed identically. Without it this would read as a
  pure false-positive story. The class really bites.

### What shipped, ordered by what closes the door

1. **`generate_image` REFUSES the generic fallback** — the framework fix, and §4's trap
   disarmed permanently. Every image the platform generates passes through this one function,
   so a meaningless prompt can no longer become a stored brand asset **regardless of which
   producer dropped it, or which producer is written next**. Mutation-proven (guard disabled in
   place, so the package still compiles; the test then fails by running on to the publish path).
2. **Both Go producers stop filing an item they know is unconsumable** — a `needs_human_review`
   item naming the gap instead. `HandlerAgent` is the **empty string**, not `"human-review"`:
   the empty handler is the canonical no-agent spelling (migration 217) and what the fleet
   actually does — **433** live rows at that status carry no handler against **12** naming
   `human-review`, which is **not an `agent_definitions` type at all**. Caught by
   `handler_coverage_test.go`, and independently objected to by two council seats.
3. **Same-origin anchoring** (above) and **the chrome surface** (above).
4. **The lying comment is deleted**, replaced with what is actually true.

### The two council objections that were worth real money

- **`editquality` was right about my method.** My "exactly one call site" claim used the
  `->'workflow'->'steps'` census shape the LANDMINES corpus warns is **top-level only**.
  Re-run recursively (`default_config @? '$.** ? (@.action == "…")'`) the claim **survives** —
  still exactly one agent — and the recursive predicate is **proven capable of finding more**:
  the same form finds **7 agents the top-level scan misses** for `call_agent` (41 vs 34).
- **`guardian` asked me to characterise the 55 assets with no recorded `origin_prompt`** rather
  than leave them as stated residual risk under a fleet-wide refusal. Done: **47 are
  DERIVATIONS** (`derived-from-hero` cards, `derived-from-logo` og_card/favicon) that never call
  `generate_image` at all, and the remaining **8 all predate 2026-03-05**. The blind spot in the
  0-of-344 safety measurement is not a hole.

### §5's design choice is still OPEN, and there is new evidence for the owner

**Option 2 (the producer synthesises a brand prompt) cannot be adopted for LOGOS as a small
change**, because it contradicts a deliberate, documented exclusion this file did not know
about: `imagery_style_guide.go:24-25` — *"photographic kinds get medium+mood+palette; icons get
palette only; **logos get nothing** — the 2026-05-20 contamination lesson"* — and
`generate_image_actions.go:430-433` — *"Logos stay excluded — generated once, human-approved,
then locked."* So routing an unplanned logo to a human is **the lifecycle the code already
implements**, not a capitulation, which is why item 2 above is defensible as an interim.
**It does not answer the owner's question** ("who decides what the logo looks like?") — a prompt
built from brand *identity* is a different proposition from one built from imagery *direction*.
That decision stays open and unmade.

### Also NOT done, deliberately (architecture-scope)

`image-build-handler` carries **two prompt-key conventions** — flat `spec.prompt` on the
Phase-2E branches, nested `spec.image_prompts.<key>` on the legacy two — and
`check_unfulfilled_image_prompt.go:101-105` says the flat key was the intended end state before
the migration stopped. Converging them would make this class **unrepresentable** for future
producers, but it changes what the shared handler guarantees its producers (owner ruling
2026-07-29). Registered as IMG-069's open question. **All three producers now write both key
shapes, so the convergence is config-only whenever someone takes it up.**

### How to verify when it rolls (do not verify at git or the tag)

```bash
kubectl exec -n ai-persona-system <chassis-pod> -- sh -c \
  'strings /app/agent-chassis | grep -c "generate_image REFUSED: no prompt supplied by any caller"'
```
≥1 on **every** replica, plus a negative control (a string the change did not add → 0) and a
positive control proving the pipeline. Then the behavioural signal: any `needs_logo`/
`needs_hero_image` failing with `generate_image refused` **names its own producer's bug**, and
that count is the first real measurement of how often an unprompted request is filed — a number
nobody has ever had.

---

## OWNER RULING 2026-08-09 (evening) — §5's design choice is DECIDED, and it reverses the disposition shipped that afternoon

> *"When a site needs a logo it should go to human review for guidance/supply of the logo, but
> for now that human review can default to saying create a logo that suits the mission, target
> market and the domain character. This is because there are 2000 odd domains to populate and I
> won't have time to do or approve that many logos."*

**So: default, do not block.** The human path stays available as an **override** (plan a real
prompt, or lock the asset); it is not a gate. The afternoon's disposition — file a
`needs_human_review` item and stop — is **removed from both producers**, because across ~2,000
domains a blocking queue is one that never drains. That objection was foreseen in the previous
round's own risk 3 ("loses automation") and the owner has now ruled on it.

**Implemented (`ebaf72729`), Go only, inert until the NEXT roll.** New shared helper
`DefaultBrandImagePrompt` (`discovery_checks/default_brand_prompt.go`), called by **both**
producers — which also discharges the previous council round's `reuse_agent` objection that the
two had duplicated the disposition with no shared helper. It composes from site name, domain,
sector, positioning, audience and tone, plus logo craft constraints. Sample output:

```
A simple, distinctive logo mark for Mortgage Calculator UK (mortgagecalculator.co.uk).
Sector: Financial Services / Mortgage Finance (UK). Positioning: The UK's Authority on
Mortgage Finance. Brand character: Direct, authoritative, and no-nonsense... Flat vector
mark, minimal and geometric, a single clear silhouette that stays legible at favicon size,
centred on a plain background, no lettering or words, no photographic texture, no drop shadows.
```

**This does NOT reopen the contamination lesson, and that distinction is what makes it safe to
automate.** The builder reads **brand identity**; it never reads `design_intent.imagery_direction`,
the imagery style guide, or `directionForKind`. Logos stay excluded from the **imagery** axis
exactly as before — a test pins that the logo branch asks for a flat vector mark and does *not*
pick up the hero branch's photographic clause. §5's option 2 was refused for logos on the
imagery axis and is now adopted on the identity axis; those are different things, and the
earlier note in this file that "a prompt built from brand identity is a different proposition"
is the one the owner has taken up.

**The property that keeps IMG-069's refusal unreachable: the default never returns empty for a
site that exists.** The domain alone yields a usable prompt (`robot-hands.com` → "a simple,
distinctive logo mark for robot hands"). That matters because at 2,000 domains a site with
nothing but a domain row is the **common** case — only **21 of 39** current sites carry an
`identity` spec at all. A builder that could return `""` would route straight back into the
refusal, so it is the hardest-tested property.

`spec.prompt_source` now records `default_from_brand_identity` vs `site_plan`:

```sql
SELECT s.domain, w.item_type, w.status, w.spec->>'prompt_source'
  FROM site_work_items w JOIN sites s ON s.id=w.site_id
 WHERE w.spec->>'prompt_source' = 'default_from_brand_identity' ORDER BY 1;
```

### IMG-069's refusal is POD-VERIFIED LIVE (2026-08-09, fresh build)

`agent-chassis-6dc54d77cd-lftkt`: refusal literal **1**, fabricated negative control **0**,
pre-existing positive control **1**. So the floor is live; **the default is not** — it needs the
next roll. Until then a promptless item fails loudly rather than painting, which is the correct
intermediate state.

### What is still open

- **Nobody has looked at a defaulted logo yet.** That prompt text now decides ~2,000 logos and
  no test can tell you it reads well. **Eyeball the first few** and tune the wording — it is a
  string in one function, cheap to change.
- **There is no review SURFACE for defaulted assets.** They are queryable (above), not rendered
  anywhere. If the owner wants a list to skim rather than a query, that is a follow-on and it
  does not exist today. Said plainly so it is not mistaken for delivered.
- Second council round `Council-Submitted: 661557c5-7ae4-43fe-a36d-c0600b54a29c` — submitted
  fresh, not as a resubmission, because it **reverses** a disposition the first round approved.
