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

## 8. Postgres, not Go, is what proves the anchoring — and both arms are needed

The Go test on `sameOriginPathPattern` is real evidence but not sufficient: `sqlmock` does not
execute SQL (it returns the rows you hand it), so a mocked "cross-origin in, no item out" would
assert the mock's own bookkeeping. The predicate was therefore run **by Postgres over live
`rendered_html`**, comparing old and new side by side:

| domain | old logo match | new logo match | old hero match | new hero match |
|---|---|---|---|---|
| fundamentallyai.com | **1** | **0** | 4 | 4 |
| 16 other sites | 0 | 0 | 137 | 137 |

The false positive goes 1 → 0; **all 141** legitimate same-origin matches survive. Both arms
matter: a pattern that silenced everything would pass the first one alone, and silence is the
failure mode that looks like success here, because these checks have a near-zero natural hit
rate.

## 9. The tests caught a real defect in my own plan, before the council did

I had written `HandlerAgent: "human-review"`. `handler_coverage_test.go` failed with: *"routes
work at handler agent 'human-review', which is not a known agent … Every item it files will be
marked 'blocked' at claim time after occupying a dedup slot."*

Checked rather than argued: `human-review` is **not an `agent_definitions.type`** at all. And
the live convention is unambiguous — **433** rows at `needs_human_review` carry an **empty**
handler, against **12** naming `human-review`. The empty string is the canonical no-agent
spelling (migration 217; the test's own comment says so). Changed to `""` in both producers.

Two council seats (`prior_art_librarian`, `guidelines`) raised exactly this independently, and
both are dispositioned as **already fixed before the verdict was read** — the test found it
first. Worth recording as the useful direction of that relationship: the council is not a
substitute for running the package's own guards.

## 10. Council verdict — APPROVED round 1, and the two objections that changed the work

Corr `c40c9483-5afd-478b-91ca-7e4db505ed0d`. 12 seats; `editquality`, `reuse_agent`,
`guardian`, `prior_art_librarian` objected (advisory); none high. All dispositioned:

- **`editquality`, edit 4, medium — MY METHOD WAS WRONG AND THE ANSWER WAS RIGHT.** My "exactly
  one call site" census used `jsonb_each(default_config->'workflow'->'steps')` — the
  **top-level-only** shape LANDMINES warns about, which cannot see steps nested in
  `sub_workflow`/loops. Re-run recursively with `default_config @? '$.** ? (@.action == "…")'`:
  **still exactly one agent**. And the recursive form is proven to see further — the same
  predicate finds **7 agents the top-level scan misses** for `call_agent` (41 vs 34), so the
  re-measurement could have come out otherwise and didn't. This is the single most valuable
  thing the council did: the claim was load-bearing for a fleet-wide refusal.
- **`guardian`, edit 4, medium — characterise the 55 unrecorded rows.** Done, and it dissolves
  the blind spot: **47 are derivations** (`derived-from-hero` 25 cards, `derived-from-logo` 11
  og_card + 11 favicon) which never call `generate_image`, and the other **8 predate
  2026-03-05**. Zero are recent generations.
- **`editquality`, edit 8, medium** — "assert NO adapter request was published" is the vacuous
  mock-expectation anti-pattern. Already resolved during implementation: the shipped test
  asserts the **error return** and is **mutation-proven**; the non-publication assertion was
  dropped rather than left to pass vacuously.
- **`reuse_agent`, edit 5, medium** — `imagePromptFromPlan` as a second prompt-extraction path
  beside `getImagePromptWithPriority`. Answered by layer: the latter is the **runtime tier
  chain** inside the generator (step config → collected data → agent config → …); the former
  reads **one key out of a plan document** at filing time, in a different package, with no tier
  semantics. Merging them would couple a producer to the generator's runtime chain. Recorded
  rather than dismissed.
- **`reuse_agent` / `bug_historian`, edit 3, low** — the human-review disposition is written
  twice (once per producer) with no shared helper, and the dedup slot is reused across two
  dispositions. Both true. The two producers live in **different packages**
  (`discovery_checks` vs `actions`) and a shared helper would be a new cross-package export —
  the kind of small seam the platform-seams ruling asks you not to add casually. The slot reuse
  is deliberate (one open question per site+purpose) but **nothing closes the parked row if a
  prompt later appears** — that is a real gap, left open and named here rather than papered over.
- **`tooling_provenance`, medium; `architecture`, low; `debug_historian`, medium** — record the
  contract change and the deploy-verification step. Both discharged in the **IMG-069** register
  entry, which states the guarantee change (always-an-image → may-refuse) and carries the
  pod-grep marker: the **log literal**, not the identifier, since string literals survive the
  build and identifiers may not.

## 11. Misstep — I nearly popped another branch's stash into a shared tree

To test whether a failing test was pre-existing I ran `git stash push -q <tracked> <untracked>`.
It **failed** (untracked file in the pathspec) and therefore created **nothing** — then my
`git stash pop -q` reached for `stash@{0}`, which was a five-branch-old WIP from
`066_hitl_questionnaire…` carrying 51 lines of `platform/orchestration/coordinator.go`.

It aborted **only because** another session's uncommitted edit to that same file collided.
Verified no damage (stash count still 7, `platform/awaitedrequests/` absent). Full write-up in
`WRONG_CALLS.md`; a LANDMINES entry now warns that the stash stack is a **third piece of shared
mutable state** that none of the multi-session rules mention. The right move — which I took
afterwards and should have taken first — is a `git archive` tree.

## 12. Misstep — my "baseline" and my "changed" tree were different commits

Both were built with `git archive HEAD`, minutes apart, and **HEAD moved in between** (another
session committed a `registry_parity_test.go` fix). So the baseline passed, the other failed,
and I spent time diagnosing *their already-fixed* failure as mine. Caught by the reported error
line landing inside a comment block in the other tree. **Pin the sha once**
(`SHA=$(git rev-parse HEAD)`) and use it for both arms. Also in `WRONG_CALLS.md`.

Final state at pinned HEAD `65ab866ca`: `platform/orchestration/actions` **ok**;
`discovery_checks` fails only `TestEveryCheckProducedItemTypeIsClassified`
(`decision_regression`, another session's committed work, failing identically at HEAD without
any change of mine).

## 13. A small one against myself: I deleted two lines from an append-only file

Removing the spent `_(to be appended)_` placeholder from `README_where_we_are.md` tripped the
`readme-not-appended` pattern check (2 lines removed from the owner's log). The lines were
**mine**, written earlier the same session, not the owner's — so the spirit of the rule was not
broken. But the rule as written is "never rewrite or reorder; add a dated correction below",
and a check that only fires when the owner's words go would be a check that cannot protect
them, because nothing in the file marks whose a line is. Recorded rather than argued away:
forward-only, the commit stands, and the honest lesson is **don't write a placeholder into an
append-only file in the first place** — write the section when you have it.

## 14. 2026-08-11 — THE LOOP CLOSED: the first defaulted brand image exists, and both halves of the verdict are in

The one-off run produced the lane's outstanding behavioural proof end to end:
item filed 10:09 with `prompt_source='default_from_brand_identity'` → promoted by id →
claimed 10:34:58 → **complete 10:36:20**, no error. The asset's `origin_prompt` is the site's
own imagery direction ("Minimal — logo and icons only. No photography or lifestyle imagery…")
composed IN FRONT of the default prompt, all of it verbatim — the framework's
direction-composition working exactly as documented. `origin_model`
stability/stable-diffusion-xl. The refusal guard was never touched: the default kept it
unreachable, which is the two mechanisms behaving as designed together.

**The image itself (looked at, per the owner's decision 3):** a flat-illustration COLLAGE —
a 3×4 grid of houses, couples and money bags on an amber ground in roughly the site's blues.
Honest verdict: **not shippable as a hero.** (a) It is a grid, not a scene — there is no
"clear space for overlaid headline text" although the prompt asked for it; (b) it contains
**dollar signs, twice**, on a UK mortgage site; (c) it is wall-to-wall happy-couples-and-houses,
which the site's own direction explicitly banned for photography and the model reproduced as
illustration. The composition tension is legible in the output: site direction said "minimal —
logo and icons only", my hero clause said "photographic or illustrative… clear space", and the
model averaged them into an icon sheet.

**Prompt-tuning candidates for the owner (one string, `composeBrandImagePrompt`):**
- "a single cohesive scene or composition, never a grid, collage or sprite sheet";
- extend "no embedded words or lettering" to "no currency symbols or glyphs" (or "UK context,
  £ only" if symbols are wanted);
- suppress the `(domain)` parenthetical when the site name IS the domain — "for
  mortgagecalculator.co.uk (mortgagecalculator.co.uk)" reads as a stutter;
- consider whether the hero clause should DEFER to a site imagery direction when one exists,
  rather than compose with it — the composition is what produced the average.

**And the deploy half hit a KNOWN bug, not a new one:** the file landed at
`/assets/images/input-data.asset-key.jpg` (200, 109,803 B) instead of `hero.jpg` —
`bugs_open/248`'s placeholder-filename defect, filed yesterday, diagnosis-CONFIRMED, owned by
`staged_component_build`. Contributed the fresh instance to that file (different dispatch path,
which widens its producer set). NOT hand-renamed: the framework rule, and nothing regressed —
the pages 404 on `hero.jpg` exactly as before.

**So the lane's own mechanisms are all PROVEN; what stands between the image and the page is
248.** Once 248 fixes the filename, either this asset gets redeployed correctly or the next
generation does.

## 15. 2026-08-11 (afternoon) — the Banana retry: same prompt, right model, unrecognisable result

Owner rejected the SDXL image (malformed figures, collage, dollar glyphs) and asked for a
stronger model + the deletion. Deletion done at BOTH stores (bucket + the gqls/sites repo —
which is what surfaced the repo-replay landmine, §WRONG_CALLS 08-11). Asset row `rejected`.

**Root cause of the wrong model, fixed as migration 390 (APPLIED, verified live):** the adapter
has routed hero/logo → Banana (Nano Banana Pro) since 6896ce22e — provenance-checked in the
running binary — but `resolveKind` reads kind from input_data and the legacy branches never
forwarded it; their `default_kind` sits in parent step CONFIG, invisible to the callee (the
bugs_open/231 dead-value class). Empty kind → adapter fallback → stability. 390 maps
`kind?: input_data.spec.purpose` on both legacy branches; the routing table finally applies.

**The owner's truncation hypothesis: CORRECT.** SDXL's CLIP encoders read ~77 tokens (~55–60
words); the composed prompt was ~115 words, our stability provider does no truncation handling
(silent clip), and CLIP treats negation as attraction — "No stock images of happy couples
outside houses" sat inside the window and produced exactly that. Gemini-family Banana reads the
full prompt; both defects dissolve together.

**Retry, same prompt, item `c5a211d8`: complete, `origin_model=banana/gemini-3-pro-image-preview`,
1408×768.** The output is flat icon-row on a light ground — a £-marked house device plus four
icons — no figures, no artefacts, correct currency, large clear space. It follows the site's own
"Minimal — logo and icons only" direction almost literally. Deployed again to the bug-248
placeholder path (`input-data.asset-key.jpg`); the six pages keep their fallback until 248's
filename fix, which is owned elsewhere. Image KEPT pending the owner's verdict.

The staleness report line now reads total=2 defaulted items — the counter counting.

## 16. 2026-08-11 — owner prefers the EXISTING site branding; the generated hero superseded, not deployed

Owner: "the new logo is fine. Can we use the old one now anyway, I prefer it. The one that is
already on the site." Investigated before acting, because "logo" is ambiguous here — this item
was `needs_hero_image`, not `needs_logo`, and the two slots are different:

- **The header logo** (`images/full-logo.png`, 812x844 PNG, gold roundel/key-and-house mark) is
  a legacy static file — **not in the `assets` table at all**, untracked by the pipeline,
  already in the sites repo (source of truth), already serving 200. Nothing to do; left alone.
- **The generated image** was for the site's HERO slot (6 guide/tool pages reference
  `/assets/images/hero.jpg`, 404). It visually reads as a logo (a monogram + icon row), which is
  almost certainly why the owner called it that.

Owner's stated preference ("prefer it") is not a quality rejection, so `assets.status='rejected'`
would misrepresent it — used **`superseded`** instead (a value the schema already carries).
Asset KEPT in storage (not deleted, unlike the SDXL reject), just no longer `active`, which
matters mechanically: `hasActiveAssetForPurpose` only counts `active`, so leaving it `active`
would have made `placeholder_image_in_use` silently stop flagging the hero gap even though
nothing on the live pages changed. `superseded` re-arms correct detection.

**Left genuinely open, not decided for the owner:** whether the 6-page hero.jpg gap should keep
being pursued at all now that he's expressed a preference for the site's existing look. Asked
rather than assumed — a design-taste call, not an engineering one.

## 17. 2026-08-12 — the loop fired AUTONOMOUSLY (proving the design), and the fleet's own provenance recipe failed on the first try

**The mechanism worked completely unforced.** At 19:06–19:08 on 08-11 — hours after we stopped
touching this site — a `placeholder_image_in_use:hero` item filed by `design-discovery-agent`
(NOT our paused rota; something else dispatched that agent for its own reasons, most likely the
sibling lane's own audit/staleness work, given its commit trail from that window) ran through
the exact path this lane built: `prompt_source=default_from_brand_identity`, routed through
`image-build-handler`, generated on `banana/gemini-3-pro-image-preview` (mig 390's fix), stored.
This is the "next design-discovery pass" the owner asked to wait for — it just arrived via
someone else's dispatch rather than the 7-day rota tick, and the fix held up under a trigger we
did not control or anticipate.

**Verifying the fresh chassis build actually carries the fix hit the fleet's OWN documented
landmine on the first attempt.** `kubectl logs -l app=agent-chassis --tail=300 | grep -m1
'build provenance'` — the method CLAUDE.md itself now recommends — returned **nothing**, on a
pod running ~15 hours. Before concluding "not shipped", checked: this exact failure is already
in `LANDMINES.md` (added 08-11, `bugfix_234_dead_spec_key` lane) — the stamp logs once at
startup and the chassis is noisy enough to rotate it out within minutes. Followed the entry's
own recommended fallback (pull the image, read the OCI label, don't grep `/proc/1/exe` for your
own commit — the landmine's "trap inside the trap" warns that only works for the exact build
commit, not any ancestor):

```
docker pull docker.io/aqls/agent-chassis:v1.0.1290
docker image inspect ... --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
  → fa078ab3d4e8282f7f5f4b1f042515f1f96d2bd3
git merge-base --is-ancestor ca6a03cce fa078ab3d... && echo shipped   # YES
git merge-base --is-ancestor ebaf72729 fa078ab3d... && echo shipped   # YES
git merge-base --is-ancestor 9425615bb fa078ab3d... && echo shipped   # YES
# negative control: a commit made AFTER the stamp must NOT be an ancestor — confirmed absent
```
**Confirmed, authoritatively, not inferred.** All three of this lane's fix commits are ancestors
of the exact commit the running image was built from. The landmine's remedy works as written.

**Bug 248 struck again, on the SECOND independent hero generation, at a path already deleted
from both stores that afternoon** — confirming the placeholder is *reconstructed* fresh by the
deploy step, not a stale leftover. Contributed as fresh frequency evidence to `bugs_open/248`.

**A NEW hero image exists, generated and stored, deployed to the (wrong) placeholder path, and
the owner has NOT seen it** — a genuinely different generation from the one already superseded
by owner preference on 08-11 (68,984 B vs 82,753 B; different content). Looked at it myself
first, per the established discipline: dark navy/geometric background, five clean monochrome
icons (savings, rate-protection shield, house+%+calculator, key, search), no artefacts, correct
currency handling (uses % not a currency glyph, since this one is rate-themed rather than
price-themed), reasonable clear space in the upper half. My own read: shippable, better composed
than either prior attempt. **Not decided for the owner — presented, not superseded, pending his
verdict**, since his 08-11 preference was stated about a specific prior image, not declared as
standing policy against all future generations.

**Two smaller variant assets also appeared** (`hero_about`, `hero_contact` — the page-named
variant path, `check_unfulfilled_image_prompt`'s classification, not this lane's canonical-hero
check) alongside `about-index` now built (4 components, deployed) and `contact-index` (2
components, needs_rebuild) — the sibling lane has clearly been building out pages the
decomposition plan listed as "never-built". Not reviewed individually; noted for the port plan's
context, since the page inventory it was written against has already moved.

**One isolated oddity, not chased:** this item's own `result` column holds unrelated
content-planner JSON rather than the image outcome. Checked against 8 other same-night
`image-build-handler` completions — all 8 clean. Isolated to this one row; cause unknown; noted
in the 248 contribution in case anyone else reads this row's `result` as evidence.
