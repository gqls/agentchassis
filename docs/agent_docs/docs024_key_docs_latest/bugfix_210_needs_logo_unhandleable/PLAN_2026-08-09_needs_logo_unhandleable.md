# PLAN — `bugs_open/210` (needs_logo slug): promptless image items are unhandleable

**Lane opened 2026-08-09.** Bug file:
`bugs_open/210_HANDOFF_2026-08-09_needs_logo_items_cannot_be_handled_and_the_comment_promises_a_fallback_that_does_not_exist.md`
(**resolve this number by SLUG** — `210` also names
`…_a_content_failed_page_build_is_stamped_deployed…`, a different, closed-in-substance case
owned by another lane).

Ownership checked before starting: `scripts/who-owns.py needs_logo` returns the
`bugfix_210_content_failed_build_stamped_deployed` lane as "active", but that is the **other**
210 — its commits are all about `update_page_status`/`DEPLOY_STAMP_REFUSED_ON_SKIP`. Live
`.jsonl` transcripts were grepped for `check_placeholder_image_in_use.go` (the lagging-check
remedy from memory): the only substantive hits are the session that FILED this bug and the
`bugfix_209` session that appended the 2026-08-09 contribution about `logo.jpg`. **No session
is fixing it.**

---

## 1. What the bug file got right, and the two things it did not know

The filed mechanism is **confirmed at HEAD and live** (§2 of the bug file, re-verified
2026-08-09 — see NOTES §1 for every citation):

- `image-build-handler`'s `call_logo_gen` still maps `"prompt": "input_data.spec.image_prompts.logo"`,
  `call_hero_gen` maps `hero_home`; both **required** (no `?` suffix), read from the live
  `agent_definitions` row, not from a seed.
- `check_placeholder_image_in_use.go:91-98` writes `spec.image_prompts` **only** when a prompt
  is found, and files the item regardless.
- The two failing rows are still `failed` with exactly that error.

Two things it did not know change the shape of the fix.

### 1a. The motivating item was a FALSE POSITIVE, and the mechanism is CROSS-ORIGIN, not basename

§6 guessed "a basename appears to have been attributed to a local missing path" and marked it
`[UNVERIFIED]`. The real mechanism is narrower and worse.
`isPathReferencedInPages` (`check_placeholder_image_in_use.go:132-148`) matches with an
**unanchored substring**:

```sql
AND pc.rendered_html LIKE '%' || $2 || '%'      -- $2 = '/assets/images/logo.png'
```

A **cross-origin** URL contains that substring. Measured 2026-08-09 `[MEASURED]` — the only
`logo.png` match on fundamentallyai, and the only one **anywhere in the fleet**:

```
name  | context
index | iv class="portfolio-logo"><img src="https://leopardessconsulting.co.uk/assets/images/logo.png
```

So the check reported fundamentallyai as serving its own placeholder logo because it renders a
**partner's** logo from the partner's own domain. It is not a basename bug; the path is
byte-identical and the origin is what differs. Same family as `bugs_closed/128` (a comparison
that ignores where a file actually comes from), one hop over.

### 1b. The check is BLIND to the surface a logo actually lives on

`isPathReferencedInPages` scans `page_components` only. A logo is rendered by the site
**chrome** (`site_components`). `bugs_closed/128` identified this exact blindness as its
"defect 3" and **fixed it in the sibling check** — `check_image_url_404.go:471-482` scans both
surfaces, with a comment explaining why — while leaving the check that actually *routes* a
regeneration blind. That is the LANDMINE at `LANDMINES.md:1996` firing verbatim: *"an overlap
landmine is keyed to a PAIR, and is often findable only by the half you are not touching."*

**Honest impact today: nil.** Four sites render `/assets/images/logo.png` in chrome
(finetuning.uk, gaswholesalers.com, idea.uk, leopardessconsulting.co.uk) and **all four have an
active `logo` asset** `[MEASURED]`, so the check's second precondition would skip them anyway.
The blindness costs a missed detection only when one of them loses its asset. It is recorded
because it is structural, not because it is biting.

---

## 2. The class defect, restated with its real magnitude

The bug file says "not site-specific". It is broader than that in two ways it did not measure.

**Three producers file these items; none guarantees the key.**

| producer | writes | guarantee |
|---|---|---|
| `check_unfulfilled_image_prompt.go:106-115` | `spec.prompt` **and** `spec.image_prompts.<key>` | safe — only runs when a prompt exists (`:76-78`) |
| `check_placeholder_image_in_use.go:91-98` | `spec.image_prompts.<key>` only if found | **files anyway when absent** |
| `WriteBuildItemsAction` (`load_work_item_actions.go:348-364`) | `spec.image_prompts = planData["image_prompts"]` **verbatim** | none — `needs_logo` is an independent boolean (`:196`) and `image_prompts` is defaulted to `{}` at `v3_site_actions.go:3107` |

The third is the **primary build path** and has never been named in this bug. An LLM plan
saying `needs_logo: true` with no `image_prompts.logo` files exactly the same unconsumable item
from `site-planner`. It has not fired yet only because the two plans that set the flag happened
to carry prompts.

**The recovery path is unavailable on 87% of the fleet** `[MEASURED]` 2026-08-09.
`loadImagePromptsForSite` reads `site_specs` where `aspect='site_plan' AND is_current`:

```
no_current_site_plan_row | row_but_no_image_prompts | has_image_prompts_obj | sites_total
                      33 |                        1 |                     5 |          39
```

So for **34 of 39 sites** the producer's "recover the planner's intent" branch returns nothing.
The promptless branch is not an edge case — it is the **normal** case. The only reason we see 2
failures rather than 34 is that the *other* precondition (no active asset for the purpose) is
currently rare.

**Current exposure, exactly** `[MEASURED]`: of 18 site/purpose pairs referencing a fallback
path, 2 have no active asset — fundamentallyai (logo) and mortgagecalculator (hero) — and
**both filed an item and both failed**. Of the remaining 16, **13 have no planned prompt**, so
each is one asset deletion away from the same failure.

**And mortgagecalculator is a TRUE positive**, which is what stops this being purely a
false-positive story: 6 same-origin CSS references, `url('/assets/images/hero.jpg')`, no active
hero asset. The class really bites.

**Census** `[MEASURED]`: promptless items **2 of 2 failed**; items carrying the key **4 of 4
completed**; items using the flat `spec.prompt` **9 of 9 completed**.

### The callee genuinely cannot save itself — Priorities 2 and 3 are structurally absent

§4's trap is confirmed and is stronger than stated. Exactly **one step in the entire fleet**
runs the `generate_image` action — `image-generator`'s `generate` — and **neither** the agent's
`default_config` **nor** that step's config carries a `prompt_template` `[MEASURED]`. So
`getImagePromptWithPriority` (`generate_image_actions.go:831-898`) has only Priority 1 or the
generic fallback; the middle of its own documented chain does not exist on this fleet.

---

## 3. The fix, ordered by what closes the door

Explicitly following `order-fix-candidates-by-what-closes-the-door`: rank by what makes the bad
state unrepresentable, not by what makes the symptom go away.

### Fix A — the producer must not count a cross-origin reference (code)

Anchor the match so only a **same-origin** reference counts. A quoted/`url(`-delimited path
starting `/assets/images/…` is the site's own; `https://other.example/assets/images/…` is not.

Closes: the only live false positive, and the whole class of foreign-URL misattribution — which
will recur, because rendering partner/portfolio logos is a normal thing for these sites to do.

### Fix B — the producer must scan the chrome surface (code)

Add `site_components` to `isPathReferencedInPages`, matching the sibling's contract
(`check_image_url_404.go:471-482`: every unlocked row, no `deployed` filter — chrome has no
comparable status; `bugs_open/117`).

Closes: a structural blindness on the exact surface the logo case lives on.

### Fix C — no caller can turn a missing prompt into a brand asset (code) ← **the framework fix**

`generate_image` **refuses** when the resolved prompt source is `generic_fallback`, returning an
error naming the caller and kind, instead of generating from *"Generate content based on the
provided context."*

This is the door-closer. It does not care which producer dropped the prompt, or whether a
fourth one is written next month: **the bad state — a meaningless prompt becomes a stored brand
asset — stops being representable at the one place all four branches pass through.** It also
makes the §4 trap permanently safe to disarm, which is what unblocks any future decision to
relax the mapping.

Measured safe before proposing, per the "no collision is possible is a query, not an argument"
rule `[MEASURED]` 2026-08-09:

- **0 of 344** assets with a recorded `origin_prompt` carry the generic string (exact and
  `ILIKE '%provided context%'`). Positive control: the predicate matches real data — 344 rows,
  25 mention "logo", lengths 147–3,882. The generic string is 46 characters, **below the
  minimum observed length**, a second corroboration.
- **Blind spot stated:** 55 of the 399 generated rows have no recorded `origin_prompt` at all;
  this measurement cannot see those. It is disconfirmable — a single generic-prompted asset
  would have shown — but it is not exhaustive.
- One call site fleet-wide, so the blast radius is fully enumerable rather than estimated.

### Fix D — the producers must not file an item they know is unconsumable (code)

A shared helper resolves the prompt once; when it cannot, the producer files an item a **human**
can act on (status/handler `needs_human_review`/`human-review`, naming the missing prompt)
rather than a `needs_logo`/`needs_hero_image` aimed at a handler that must fail. Applied to
**both** Go producers, including `WriteBuildItemsAction` — the latent one.

Two things make this the honest disposition rather than a shrug:

- **It is the lifecycle the code already implements.** `generate_image_actions.go:430-433`:
  *"Logos stay excluded — generated once, human-approved, then locked (the 2026-05-20
  contamination lesson)"*, and `imagery_style_guide.go:24-25` gives logos **no** style-guide
  direction, deliberately. A human ruling on an unplanned logo is not a new policy; it is the
  existing one.
- **The queue is not dead**, contrary to what a reader of `bugs_open/033` would assume:
  `evidence_citations.go:426-427` records 428 rows at that status fleet-wide with **86 claimed
  and 55 handled**, newest claim 2026-08-02. `[SECOND-HAND — from a code comment dated
  2026-08-03; re-measure before relying on it]`

### Deferred, deliberately, and NOT decided here

- **D1 — synthesise a real brand-grounded prompt** (bug-file option 2). This is the design
  choice the filing session correctly refused to make alone, and finding 1b/§3 Fix D sharpens
  the objection: for **logos specifically**, synthesising from the imagery signal contradicts a
  documented deliberate exclusion. Belongs to the owner, with that evidence in front of them.
- **D2 — converge the prompt key so all four branches read `spec.prompt`.** The two legacy
  branches use a nested kind-keyed lookup; the two Phase-2E branches use a flat key, and
  `check_unfulfilled_image_prompt.go:101-105` says the switch to `spec.prompt` was the intended
  end state. Converging would make this class unrepresentable for future producers too — but it
  **changes what the shared handler guarantees its producers**, which is architecture-scope
  under the 2026-07-29 ruling. It goes to its own round with its consumers named, not inside a
  bug patch. Registering it here in the same commit satisfies condition (2).

**Not doing** bug-file option 1 (a shared `prompt_template` on `image-generator`): one string is
shared across logo/hero/imagery/variant, so it is wrong for at least three of them, and Fix C
makes the failure loud instead — which is what that option was reaching for.

---

## 4. Why this is not "make the error go away"

The one-line fix (`"prompt"` → `"prompt?"`) is refused, exactly as the bug file demands, and
Fix C is what makes that refusal permanent rather than a note. After A–D:

- the false positive never files (A),
- a real placeholder logo in chrome is finally *seen* (B),
- a promptless item is never routed to a handler that must fail (D),
- and if any path is ever missed, the result is a **loud error**, never a wrong brand asset (C).

## 5. Tests (all new — `grep` finds no existing coverage of these three symbols)

- `TestPlaceholderImageInUse_CrossOriginReferenceIsNotOurPlaceholder` — the fundamentallyai
  HTML verbatim; expects **zero** work items.
- `TestPlaceholderImageInUse_ChromeSurfaceIsScanned` — a logo path present only in
  `site_components`; expects the finding.
- `TestPlaceholderImageInUse_PromptlessDoesNotFileAGenerationItem` — no planned prompt ⇒ no
  `needs_logo` for `image-build-handler`.
- `TestGenerateImage_RefusesGenericFallbackPrompt` — **mutation-proven**: the guard must be
  shown to fail when removed, per `mutate-the-code-to-prove-the-guard`.

## 6. Sequencing

1. Lane docs (this file, NOTES, RUNBOOK, README) — done at the start, not at handoff.
2. Submit A–D to the council gate; commit with `Council-Submitted:` rather than holding code on
   a shared HEAD.
3. Implement + tests; commit by pathspec.
4. Register D2 in the concept register; append the two landmines (cross-origin substring;
   the pair-keyed sibling divergence) and run `landmines-sync.py --apply`.
5. Update the bug file in place with the corrections to §4/§6 — visibly, not silently.
