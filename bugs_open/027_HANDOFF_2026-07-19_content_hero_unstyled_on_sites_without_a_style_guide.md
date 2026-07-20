# HANDOFF — `content_hero` generates unstyled on every site without an imagery style guide

**Filed:** 2026-07-19, from the imagery workstream (checking whether the funded tool
rollout was safe to let drain).
**Severity:** Medium — no data loss, but it spends real image-API credits producing
output of the exact class that failed the D13 owner gate, and it is armed on three
sites right now.
**Status:** OPEN — **§4b's CODE FIX APPLIED 2026-07-20 (`1191cecdb`), INERT UNTIL AN
IMAGE ROLL.** Stays OPEN because the /bugs_closed/ bar is fixed AND live. §1's
missing-style-guide half was mitigated in config 2026-07-19 (three sites seeded,
live immediately); the fleet-default-direction candidate (§5(a)) remains unbuilt.

> **Fix trail (2026-07-20).** Seven council-gate rounds on correlation `0a07f5ed`
> (one round voided by bugs 019), final tally **11 approve / 2 object / 4 abstain**
> — REVISE, not APPROVED, so the commit carries **no `Council-Reviewed:` trailer**;
> this note is the honest record of that gap (precedent: 011 R1, same residual
> class, also shipped on a REVISE). What shipped: palette composed FIRST; the
> truncation backoff changed first-'. ' → LAST-'. ' within the cap (**a latent bug
> my own reorder would have armed — my round-2 computation was wrong, no council
> seat caught it, and the owner's "think hard before committing" prompt is what
> caught it**: under the old backoff robot-hands would have kept ONLY its palette);
> `composeImagePromptWithDirection` returns a `truncated` bool and the call site
> WARNs `Imagery direction TRUNCATED before generation` (the deploy marker);
> `datahelpers.SafeCut` is the one rune-safe cut, with `TruncateString` (31 sites)
> and helpers.go's two truncators delegating to it. Tests: survival fixture that
> FAILS under the old backoff, palette-first order pin, rune-safety edges — all
> green against `git archive HEAD` + fix overlaid (the working tree carries
> another session's broken WIP test, tool_acceptance_convergence_test.go).
> **Unresolved objections, recorded as follow-ups, not smuggled in:** persist
> truncation signals beyond a log line (bug_historian; same class as 011's
> UnmigratedKind residual); loud-cap treatment for the remaining 30 TruncateString
> sites (the 016b §9 "silent cap" entry names the sweep). **Landing gate before
> §4b closes:** regenerate robot-hands' 3 ARTICLE content heroes vs the D13
> criteria (its 3 TOOL heroes wait out the bugs 020 owner hold); verify the deploy
> via the WARN log literal, not a symbol grep (A6.3); ≥5 observed generations —
> 028 is live in v1.0.1140 and precedes this, so post-roll deltas attribute here.
> **Base-voice trade, decided and stated:** all four base voices (304–398 chars)
> cannot fit palette AND prose in 200; they flip from prose-without-colours to
> colours-without-prose. Remedy is config (shorten palette glosses — the WARN now
> names every over-cap site); the three guides I authored get a backed-up
> needle-gate migration after the roll; robot-hands' base guide is the owner's.

---

## 1. The defect

D14 (2026-07-18) made `content_hero` its own imagery KIND: flat duotone editorial
illustration, routed to Banana, with its style supplied by a per-kind override map in
the site's `imagery_style_guide` spec (`kinds.content_hero`).

The style-guide side of that was wired correctly. The **free-text fallback gate was
not**. In `platform/orchestration/actions/generate_image_actions.go:407-433`:

```go
direction := styleGuide.directionForKind(kind)      // "" when the site has no guide
if direction == "" && directionAppliesToKind(kind) {
    direction = getImageryDirectionForSite(...)     // design_intent.imagery_direction
}
```

`directionAppliesToKind` (`generate_image_actions.go:1109`) excludes only
`icon`, `logo`, `sprite_sheet`; **`content_hero` falls to the `default: return true`
branch.** So on a site with no `kinds.content_hero` override, a kind defined as flat
illustration receives the site's free-text imagery direction — which is written to
describe the site's *photographic* house style.

That is the contamination class the function's own doc comment exists to prevent:

> "prepending a photography directive to an icon prompt makes the model render a
> photograph with the icon composited into a corner (observed on icon_cycle_time,
> 2026-05-20)"

`referenceKeysForKind` has the same shape but is harmless here: with no guide it
returns nil, so the generation is simply unanchored rather than wrongly anchored.

**The five-place checklist for a new kind (HANDOFF_imagery_best_in_class.md,
"Mechanisms") says a new kind must be added to BOTH gating functions.** D14 added
`content_hero` to `directionForKind` (via the `Kinds` override map) but not to
`directionAppliesToKind`. It reads as done because the only site exercising it —
robot-hands.com — has the override, so its output is correct and proved nothing about
anyone else.

## 2. Who it bites, verified live 2026-07-19

Only robot-hands.com has an `imagery_style_guide` row at all:

```sql
SELECT s.domain, (sd.id IS NOT NULL) AS has_guide,
       (sd.data->'kinds' ? 'content_hero') AS content_hero_override
  FROM sites s LEFT JOIN site_specs sd
    ON sd.site_id=s.id AND sd.aspect='imagery_style_guide' AND sd.is_current=true;
```
| domain | has_guide | content_hero_override |
|---|---|---|
| robot-hands.com | t | t |
| finetuning.uk | f | — |
| gamesdesign.co.uk | f | — |
| leopardessconsulting.co.uk | f | — |

Their `design_intent.imagery_direction` — what would be prepended instead — varies in
how badly it fits a flat-illustration kind:
- **gamesdesign.co.uk**: *"Minimal — primarily icons and emoji-style glyphs as section
  markers; no photography…"* — compatible, arguably fine.
- **finetuning.uk**: *"Abstract, geometric, and atmospheric. Network patterns, data
  flow visualisations…"* — broadly compatible.
- **leopardessconsulting.co.uk**: *"Two families, kept apart on purpose. Explanatory
  images — how a pipeline flows, what a wor…"* — a direction that describes **two**
  styles and relies on a human choosing between them. Prepended wholesale to a single
  prompt, it is incoherent by construction.

So the harm is not uniform, and this handoff does **not** claim all 19 images would be
unusable. What it claims, and what is verified, is that **the kind's style is a lottery
per site instead of a defined default** — and style consistency is precisely the axis
on which the D13 gate failed on 2026-07-17.

## 3. Why it is armed right now

`content_image_missing`'s F3 surface table (`check_content_image_missing.go:129`) added
the `tool` surface. It gates on a site having a component that consumes
`query.pages_where_type:tool`, and takes deployed tool pages. Live counts:

| domain | tool-list consumers | deployed tool pages | current imagery state |
|---|---|---|---|
| gamesdesign.co.uk | 2 | 9 | none — all 9 would emit `generate` |
| finetuning.uk | 1 | 5 | none — all 5 would emit `generate` |
| leopardessconsulting.co.uk | 2 | 5 | none — all 5 would emit `generate` |
| robot-hands.com | 1 | 3 | fulfilled (hero + fresh card) — silent |
| idea.uk | 2 | 1 | **`deployed_at IS NULL` → excluded, emits nothing** |

**19 generations across three unstyled sites**, at 10 per site per pass, on the next
discovery pass on any of them. The check is registered on `design-discovery-agent`
itself (`SQL_2026-07-16_register_content_image_missing.sql` patches the agent
definition by `type`), so it runs on **every** site's discovery pass — no per-site
opt-in stands between this and a routine sweep.

**It is not on a timer.** `scheduled_tasks` has no discovery/improvement-loop entry
(the 12 enabled tasks are health checks, reapers, build-pipeline-trigger and feed
refresh). Discovery passes are fired by hand. So this trips when **any concurrent
session runs a routine discovery pass on one of those three sites** — leopardess had
123 discovery items in the last two days, so that is a live possibility, not a
theoretical one.

Their last discovery items all predate the F3 tool-surface deploy, which is why it has
not fired yet:
```
gamesdesign.co.uk           2026-07-13 12:05
finetuning.uk               2026-07-17 13:03
leopardessconsulting.co.uk  2026-07-17 19:43
robot-hands.com             2026-07-19 10:30   ← the only one post-F3
```

## 4. This also corrects the imagery handoff's rollout figures

`HANDOFF_imagery_best_in_class.md` (B16.2 / next-actions §2) says the funded rollout
drains via *"gamesdesign.co.uk (9 tool pages) and idea.uk (1) … the other 7 sites with
tool pages have no `tool-list`, so the consumer gate spends nothing on them."*

Both halves are wrong against the live DB:
- **finetuning.uk and leopardessconsulting.co.uk DO have tool-list consumers** (1 and 2)
  and 5 deployed tool pages each → +10 generations nobody has counted.
- **idea.uk spends nothing**: its single tool page has `deployed_at IS NULL`, and the
  tool surface uses `DeployedPageEligibilitySQL`.

Net: pending exposure is **19 generations across 3 sites**, not 10 across 2. That is a
material input to the B16.2 volume sign-off, which was sized on the smaller number.

## 4b. A SECOND, SHARPER DEFECT found by piloting the fix (added 2026-07-19)

Writing the three sites a style guide (§5(b)) was applied — and the pilot **failed**,
which exposed a defect that bites **every site including robot-hands**.

Two pilot content heroes on gamesdesign.co.uk came back with the correct near-black
GROUND but an **invented accent** — orange/navy on one, a teal field on the other, no
cyan anywhere — and inconsistent with each other. `assets.origin_prompt` says why. The
prompt actually sent ended:

> `... colour palette: near-black ground (#121212). Header image for a web-based tool
> representing: Effective Health (EHP) Calculator ...`

**The cyan is simply gone.** Mechanism, both halves verified in code:
- `composeDirection` (`imagery_style_guide.go:136`) joins `medium` → `mood` →
  `"colour palette: "+palette` — **the palette is always LAST**;
- `composeImagePromptWithDirection` truncates the whole direction at
  **`maxImageryDirectionInPrompt = 200`** (`generate_image_actions.go:1037`).

So a verbose `medium`+`mood` silently eats the colour instruction — the single most
brand-identifying part of the guide — and the model invents an accent.

**This is why robot-hands looked fine and gamesdesign did not.** robot-hands'
content_hero direction composes to **233 chars** — also over the cap — but its cut
lands *after* `electric blue (#0080FF)`, so it loses only "light grey secondary
accents only". The same config shape works or fails on whether the accent colour
happens to fall before character 200. Nothing warns; the truncation is silent and the
generated image looks deliberate.

Check any site with:
```sql
SELECT s.domain, length((d->>'medium')||'. '||(d->>'mood')||'. colour palette: '||(d->>'palette')) AS composed_len
  FROM site_specs sd JOIN sites s ON s.id=sd.site_id,
       LATERAL (SELECT sd.data->'kinds'->'content_hero') AS k(d)
 WHERE sd.aspect='imagery_style_guide' AND sd.is_current=true;
```

**Mitigated in config** (`SQL_2026-07-19_style_guides_terse_directions.sql`, applied):
terse medium+mood and the accent FIRST inside the palette string, bringing the three
sites to 139–147 chars. **robot-hands is still at 233 and has NOT been changed** —
its output is currently acceptable and it is another gate's testbed; changing it
unprompted would invalidate a passed gate.

**A third symptom the config fix does NOT address:** both pilot images contained
lettering ("HP", "EHP", "ARMOR"; "g", "v") despite `avoid` listing text/lettering AND
the positive prompt ending "no text or lettering in the image". Banana/gemini renders
text readily (that is precisely the capability `bugs_open/011` celebrates for
infographics). Whether the Banana path sends `avoid` as a negative prompt at all is
UNVERIFIED and is the next thing to check — if it does not, every `avoid` list in the
fleet is inert for Banana-routed kinds, which is all the flat kinds.

## 5. Fix candidates

Two independent fixes; they are not alternatives, they address different halves.

**(c) NEW, and the highest-value one — compose the palette FIRST, or exempt structured
guides from the 200-char cap** (see §4b). Truncating a structured, brand-approved guide
at a fixed character count, palette last, silently discards the most important field.
Code, fleet-wide, wants the council gate.

**(a) Structural — give `content_hero` a defined default (code, inert until an image
roll).** Either add `content_hero` to `directionAppliesToKind`'s exclusion list (so an
override-less site gets *no* direction: unstyled but uncontaminated), or — better and
matching D14's intent — give the kind a **fleet default flat-illustration direction**
used when no site override exists. The second is more code but removes the lottery
instead of just muting it. This wants the council gate: it changes generation
behaviour fleet-wide.

**(b) Per-site — write the three sites an `imagery_style_guide` with a
`kinds.content_hero` override (config, live immediately, no image roll).** This is what
robot-hands has and what made its cards consistent. Cheap and reversible, but it is a
brand decision per site, so it needs the owner — and it does not stop the next new site
hitting the same hole.

**Containment if a pass must run before either lands:** the emitted items are
`needs_imagery` at status `detected`; they do not spend until triaged and dispatched.
Deleting or cancelling them before triage costs nothing.

## 6. How to verify a fix

- Unit: `directionAppliesToKind("content_hero")` / whatever the chosen default is,
  asserted alongside the existing icon/logo/sprite_sheet cases.
- Live, the honest test: force one generation on **leopardessconsulting.co.uk** (the
  worst-fit direction) via A6.5 + A6.1, then read the adapter log line
  `Prepended imagery direction` and check `source` — `+style_guide` is right,
  `+imagery_direction` means the fallback still wins — and eyeball the image.
- Do **not** verify by grepping the pod binary for `content_hero`: that marker is not
  retained by the Dockerfile build and a miss proves nothing (016b §9, RUNBOOK A6.3).

## 7. What is NOT claimed

I have not run a generation on any of the three sites, so the *output* harm is inferred
from the code path plus the D13 gate precedent, not observed. The code path itself, the
absence of the three style guides, the 19-page exposure and the corrected rollout
figures are all verified live and quoted above.
