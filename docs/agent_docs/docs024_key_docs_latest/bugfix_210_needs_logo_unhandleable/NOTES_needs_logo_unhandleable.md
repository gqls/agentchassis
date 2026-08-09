# NOTES — bugfix 210 (needs_logo slug)

Append-only, newest at the bottom. Missteps are the point, not an appendix.

---

## 1. 2026-08-09 — validity re-check at HEAD and live, before anything else

The bug was filed the same day, but the tree is shared and files move underneath us, so every
load-bearing claim in §2 was re-read rather than trusted.

- `check_placeholder_image_in_use.go:91-98` — `spec["image_prompts"]` still written **only**
  inside `if prompt != ""`, item filed unconditionally at `:108-122`. **Confirmed.**
- `input_mapping.go:100-137` — allow-list; `isOptional` is a `?` **suffix on the destination
  field** (`:102`), hard error otherwise (`:130-136`). **Confirmed.**
- Live `agent_definitions` (not a seed): `call_logo_gen.config.input_mapping.prompt` =
  `input_data.spec.image_prompts.logo`, `call_hero_gen` = `…hero_home`, neither with `?`.
  **Confirmed.**
- Both failing rows still `failed` with that exact error. **Confirmed.**
- `image-generator.default_config ? 'prompt_template'` → **false**. §4's trap holds.

So the bug is valid. Two claims in it are not, and are corrected below.

## 2. The §6 false-positive guess was right in conclusion, wrong in mechanism

§6 said a *basename* had probably been attributed to a local path, marked `[UNVERIFIED]`. I
went looking for basename matching in `isPathReferencedInPages` and **there is none** — it
matches the full path `/assets/images/logo.png`. For a while I thought §6 was simply wrong.

It is not: the match is real and the attribution is still false, because the predicate is an
**unanchored substring** and the matching HTML is

```
<div class="portfolio-logo"><img src="https://leopardessconsulting.co.uk/assets/images/logo.png
```

A **cross-origin** URL contains the path byte-for-byte. So the check reported fundamentallyai
as serving its own placeholder because it renders a partner's logo from the partner's domain.

Worth recording as a general shape: *§6 reached the right verdict from the wrong mechanism, and
had I trusted the mechanism I would have "refuted" a correct finding.* Checking the conclusion
independently of the stated cause is what saved it.

## 3. MISSTEP — an SQL operator-precedence error that returned a plausible wrong table

Logged in `WRONG_CALLS.md`. I ran, fleet-wide:

```sql
WHERE pc.build_status='deployed' AND pc.locked_at IS NULL
  AND pc.rendered_html LIKE '%logo.png%' OR pc.rendered_html LIKE '%hero.jpg%'
```

`AND` binds tighter than `OR`, so this is `(A AND B AND C) OR D` — every `hero.jpg` match in the
database, including undeployed and locked components. **No error, and the table looked
reasonable.** It contradicted my own earlier, correct single-site query (which found exactly one
match on fundamentallyai) by reporting four — and the contradiction is the only reason I caught
it. Re-run parenthesised in RUNBOOK R5.

The cheap check that would have caught it immediately: the corrected query must **agree with the
single-site query I had already run**. A fleet aggregate that disagrees with a verified point
measurement is wrong until proven otherwise.

## 4. What the DB said that the bug file did not know

- **A third producer.** `WriteBuildItemsAction` (`load_work_item_actions.go:348-364`) files
  `needs_logo`/`needs_hero_image` with `spec.image_prompts = planData["image_prompts"]`
  verbatim, while `needs_logo` is an independent boolean (`:196`) and `image_prompts` is
  defaulted to `{}` (`v3_site_actions.go:3107`). Same defect, on the **primary build path**,
  unfired so far by luck.
- **87% of sites cannot use the recovery branch.** 33 of 39 sites have no current `site_plan`
  spec row at all; 1 has a row with no `image_prompts`; **5** have the object. So
  `loadImagePromptsForSite` returns empty for 34 of 39.
- **Perfect correlation in the census.** Promptless 2/2 failed; keyed 4/4 completed; flat
  `spec.prompt` 9/9 completed.
- **mortgagecalculator is a TRUE positive** — 6 same-origin `url('/assets/images/hero.jpg')`
  CSS references, no active hero asset. Without it this would read as a pure false-positive
  story, and the class defect would look academic. It is not.

## 5. A fourth defect, found by reading the sibling rather than the file under repair

`isPathReferencedInPages` scans `page_components` only. A logo lives in the site **chrome**
(`site_components`). The sibling `check_image_url_404.go:471-482` scans both, with a comment
naming this as `bugs_closed/128` "defect 3" — so the blindness was **found and fixed in the
flag-only check, and left in place in the one that routes a repair**.

That is `LANDMINES.md:1996` firing exactly as written: an overlap landmine is keyed to a PAIR
and is findable only from the half you are not touching. I found it because the LANDMINES grep
for `check_placeholder_image_in_use` returned the pair entry, and the entry told me to go read
the other file.

**I am NOT claiming this is currently costing detections.** Four sites reference the logo path
in chrome and **all four have an active `logo` asset**, so the check's second precondition skips
them regardless `[MEASURED]`. Impact today is nil; the defect is structural. Saying otherwise
would be exactly the overclaim this lane is meant to avoid.

## 6. Why the fix is not the one-liner, restated as a decision

`"prompt"` → `"prompt?"` makes the error stop. It also makes `getImagePromptWithPriority` fall
through to `generic_fallback` — and I confirmed the fall-through is **structural, not
incidental**: exactly one step in the fleet runs `generate_image`, and neither the agent config
nor the step config has a `prompt_template`, so Priorities 2 and 3 of its own documented chain
do not exist here. The chain has two rungs, not four.

So the framework fix is to make the **last** rung refuse (Fix C) rather than to make the first
rung optional. Measured safe: 0 of 344 recorded `origin_prompt`s are the generic string, with a
positive control proving the predicate matches real data, and the string being shorter than the
shortest observed prompt as corroboration. **Blind spot stated:** 55 generated rows have no
recorded prompt and this measurement cannot see them.

## 7. The design choice I am NOT making, and what I found that bears on it

The bug file's option 2 is "the producer synthesises a prompt from the imagery style guide and
`design_intent`". For **logos** that runs into a deliberate, documented exclusion:

- `imagery_style_guide.go:24-25` — *"photographic kinds get medium+mood+palette; icons get
  palette only; **logos get nothing** — the 2026-05-20 contamination lesson"*.
- `generate_image_actions.go:430-433` — *"Logos stay excluded — generated once, human-approved,
  then locked … a locked asset must not acquire a new colour instruction on a re-run."*

So auto-deriving a logo prompt from the site's imagery signal is the thing this codebase already
decided not to do. That does **not** settle the owner's question ("who decides what the logo
looks like?") — a prompt built from brand identity rather than imagery direction is a different
proposition — but it does mean option 2 cannot be adopted for logos as a small change, and it
makes routing to a human the disposition consistent with the existing lifecycle rather than a
capitulation.

Recorded so the owner decides D1 with this in front of them, not after.
