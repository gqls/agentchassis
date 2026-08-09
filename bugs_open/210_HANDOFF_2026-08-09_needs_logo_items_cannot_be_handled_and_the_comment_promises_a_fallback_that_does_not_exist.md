# 210 — `needs_logo` items are unhandleable, and the producer's comment promises a fallback that does not exist

**Filed 2026-08-09** from the fundamentallyai improvement-sweep front. **OPEN.**
Status: diagnosed first-hand, **not fixed** — the obvious one-line fix is a trap (§4) and
the real fix is a design choice that should not be made unilaterally (§5).

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
