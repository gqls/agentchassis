# RFC_059 — a pinned palette/typography should structurally survive the LLM overlay

**Status: DRAFT — DO NOT RATIFY AS WRITTEN. Three defects found in review the same
day it was filed (2026-09-02); §§1-7 below are the ORIGINAL text, left unedited so
the objections can be read against them. Start at §0.**

## §0. Review objections (2026-09-02, Fable architecture review, same day as filing)

Recorded in the file per PROCESS_architecture_review.md ("objections and revisions
happen in the file, visibly"). All figures measured against the live DB.

**O1 — the fix pins the WRONG STORE, and would repaint two of this RFC's own named
canaries.** §2.1's code returns `out` after copying `themePalette` — i.e. it serves
`comp.Palette`, the site's **`palettes` row**, never `reference_values`. The prose
throughout says `reference_values` is what gets held. Those two stores drift: of 33
composed sites with a current `design_intent`, **4** disagree on background/primary/
text — `cv1.co.uk`, `finetuning.uk`, `gaswholesalers.com`, `loanzy.uk`. Two of them
are the canaries §5 nominates. Their `palettes` row holds the fleet default while
`reference_values` holds the real served colour, and they render correctly TODAY only
because the LLM overlay carries `reference_values` through the prompt. This fix would
skip that overlay and ship them the generic default — the canary failing in the
opposite direction to the one being watched.
→ When pinned, overlay `reference_values` onto the theme palette
(`buildPaletteMap(comp.Palette, referenceValues)`), do not return the composition row.

**O2 — the discriminator discriminates nothing; the "unchanged" population is empty.**
§2 claims sites without `reference_values` behave exactly as today. Measured: **34 of
34** sites with a current `design_intent` have non-empty `palette.reference_values`
(31 for typography). **25 of the 34** were written by `domain-research-classifier`,
whose output migration 613's own header calls "an unsteered coin-flip that lands dark
about a third of the time". So this ships a fleet-wide palette freeze, mostly onto
machine guesses, and makes the boxingonline contradiction class (§1) permanent rather
than merely undetected.
→ Gate on an explicit opt-in lock INSIDE the spec data — e.g.
`design_intent.palette.locked: true` — written by `apply_theme_kit` and by human pin
migrations, never automatically. That is also the shape OWNER RULING 2026-08-02 §2
prescribes (opt-in field, unsafe default OFF), and under RFC_022 such a field with no
live consumers is not itself architecture-scope.
→ §3's rejection of `site_specs.pinned` is reasoned wrongly: a READER needs one
`SELECT pinned … WHERE aspect='design_intent'`, not changes to the shared writer, and
"no behavioural gain over presence" is backwards (presence separates nothing; `pinned`
is true on 4 rows, all human-set). The real objection to `pinned` — which this RFC did
not make — is that it is **lost on supersede** (both `WriteSiteSpecAction` and
`apply_theme_kit` omit it on INSERT; 2 of the 4 pinned rows are no longer current).
That argues for the in-data key, not for presence.

**O3 — Bar 1 of PROCESS_architecture_review FAILS: no measured instance of the
defect.** The bar asks for a defect the current design cannot express a fix for, with
live figures. §1's robot-hands citation is the *no-`reference_values` → invent* branch,
which DES-052 already closed. The boxingonline case was refuted by a peer session and
is (correctly) disclaimed in §1 — but nothing replaced it. Migration 350's own
correction says the one observed pinned run "reproduced the pinned values exactly".
§7's first acceptance checkbox is "confirm the served palette DOES drift" — i.e. the
defect's existence is listed as future work.
→ Run the induced-fault test, or a `090` diagnosis, BEFORE ratification. If the drift
cannot be reproduced on a pinned site, this RFC has no subject and should be withdrawn
in favour of the much smaller O1 fix (make the two stores agree).

**Lesser points:** Bar 4's "no migration" is false — the §2.3 prompt change is a live
config migration to `agent_definitions`, and needs its `_ROLLBACK.sql`. Bar 3's single
stage is fine but should say "one stage, deliberately" rather than describing itself as
staged. §2's advisory check should read the structured `design_intent.dark_light.scheme`
(present on 9/34) before keyword-scanning `colour_mood`. The "~15-line diff" figure
excludes that check. §2's line "`enforceLayoutScheme` is documented as
`buildPaletteMap`'s only non-test caller" is garbled — `enforceLayoutScheme` is not a
caller; the underlying fact (one call site, `:125`, with a comment warning against a
second) is correct. Consumers of the changed guarantee — the classifier lane and the
adoption lane — must be TOLD, not merely measured (OWNER RULING 2026-07-29 §3).

---

**Status of the original draft below: unedited.**

Filed 2026-09-02, session "themes", as Track B of the theme-kit plan
(`/home/ant/.claude/plans/please-think-hard-about-starry-locket.md` §6) —
prioritized by the owner to run *alongside* the theme-kit registry build
(Phase 1, commit `0902039c0`), not deferred, but routed here rather than
shipped silently because it changes a shared rendering guarantee.

## 1. Problem + evidence

**The mechanism, precisely located.** `design_intent.palette.reference_values`
/ `.typography.reference_values` is correctly consulted by the *deterministic*
composition resolver at install time (`resolve_composition_pallette_action.go`
§2 of its cascade, `resolve_composition_typography_action.go` §1) — a themed
site's first render is already correct. The drift is at *render* time, on any
later `needs_design` pass: `webdesign-agent`'s `analyze_design` LLM step is
told these values are "starting points, not exact targets"
(`docs/agent_docs/sql_for_agents/031_webdesign_agent.sql:2506,2510`), and its
fresh output is merged back in **unconditionally** by
`buildPaletteMap`/`buildTypographyMap`
(`platform/orchestration/actions/render_css_composition_helpers.go:72-117`),
called from `render_css_from_spec_action.go:125-126`:

```go
mergedPalette := buildPaletteMap(comp.Palette, specPalette)
mergedTypo := buildTypographyMap(comp.Typography, specTypo)
```

`buildPaletteMap` lets the spec's core slots (primary/secondary/accent/
background/surface/text/text_muted/border) win whenever the LLM supplies a
non-empty value; `buildTypographyMap` lets the spec win on every key. **There
is no check anywhere in this merge for whether the site's palette/typography
was deliberately pinned.**

**Is there a real `pinned` mechanism already?** Yes, but not this one.
`site_specs.pinned` is a real, schema-level, ENFORCED boolean column — but
only for `aspect='evidence_base'`, gating exactly two hand-written call sites
(`evidence_citations.go:216-219,374`; `refresh_evidence_base_action.go:313-317,
1422`). `grep -c pinned platform/orchestration/actions/site_spec_actions.go`
returns 0 — the general spec read/write path, and therefore `design_intent`,
has never been wired to it. A prior migration
(`docs/agent_docs/sql_for_agents/350_pin_design_intent_palette_for_the_three_
unpinned_sites.sql`) used "pin" as a folk term for "populate
`reference_values`" and later added its own correction, in the file, after
applying: *"THIS FILE'S HEADER AND COMMIT MESSAGE OVERSTATE WHAT IT DOES...
`design_intent.palette.reference_values` is ADVISORY BY CONSTRUCTION."*

**The motivating case is `check_generic_theme.go`'s documented history, not a
single site.** The discovery check's own code comment (`platform/
orchestration/actions/discovery_checks/check_generic_theme.go`, and the
"webdesign colour-churn landmine" this session's memory already carried) names
the mechanism directly: testing only `site_specs.aspect='webdesign'` (a
contract nothing has ever written) made the check fire on every themed site,
each pass dispatching `webdesign-agent`, whose `analyze_design` re-rolls the
palette — *"robot-hands R1, 2026-07-17: four CSS rewrites in a day, one rolled
a light background onto a dark site."* That fix (testing the value in
`sites.content_data.color_scheme` too) stopped the check firing spuriously; it
did not stop a *legitimately* re-dispatched `needs_design` pass (a real
content change, a re-classification, an operator-requested redesign) from
still re-rolling a site's colours even when `reference_values` says otherwise.
That residual gap is this RFC.

**One case this RFC does NOT explain, and must not be cited as evidence for
it**: the boxingonline.com site's palette/prose mismatch this session
initially considered as a motivating case. Corrected by `site_delivery_and_
editor` (peer session, 2026-09-02): that site's `palettes` row was
byte-identical to its `design_intent.palette.reference_values` across all
eight slots — `reference_values` was consulted and honoured perfectly, no
later pass overwrote it. The actual defect there was upstream and different:
the same `design_intent` row's `colour_mood` PROSE asked for near-black/
red/gold while its `reference_values` encoded a light cream theme — one row
contradicting itself. This RFC's fix would not have changed that site's
outcome (see §3, "prose/values contradiction" — the fix could make a case
*like* that one WORSE, not better, which is exactly why §2 includes an
advisory check for it).

## 2. Design

**Direct, already-shipped precedent for the shape of this fix.**
`bugs_closed/022` added `enforceLayoutScheme` (`render_css_from_spec_action.go:
390-445`, called at line 134, immediately after the merge) — it makes the
*layout's* declared scheme structurally override a contradicting LLM
background. Its own doc comment states the exact philosophy this RFC
generalizes: *"The layout's scheme is a user decision; the spec's palette is
a per-run LLM guess — on contradiction the layout wins."* `enforceLayoutScheme`
is documented as `buildPaletteMap`'s only non-test caller today — this RFC's
change is the second caller-adjacent site, not a second, independent
mechanism.

**The change** (~15-line diff across two files):

1. `render_css_composition_helpers.go` — `buildPaletteMap`/`buildTypographyMap`
   each take a new `pinned bool` parameter. When true, return the theme's
   values unchanged; skip the spec overlay for that dimension entirely.

   ```go
   func buildPaletteMap(themePalette map[string]string, specPalette map[string]interface{}, pinned bool) map[string]string {
       out := make(map[string]string, len(themePalette)+len(specPalette))
       for k, v := range themePalette {
           if v != "" { out[k] = v }
       }
       if pinned {
           return out // reference_values is a hard constraint — the LLM's colour_scheme never lands
       }
       // ... existing core-slot overlay loop, unchanged
   }
   ```
   Same shape for `buildTypographyMap` (currently spec-wins on every key,
   unconditionally; `pinned` short-circuits before that loop).

2. `render_css_from_spec_action.go` — compute `paletteIsPinned`/
   `typographyIsPinned` from the same presence check the prompt template
   already performs (`site_specs.specs.design_intent.{palette,typography}.
   reference_values` non-empty), reading `params.CollectedData` — no new DB
   read, the data is already there via the workflow's `read_site_specs` step.
   Pass through at the existing single call site (line 125-126).

3. `031_webdesign_agent.sql` — change the `analyze_design` prompt's
   conditional wording (already gated on the same presence check) from
   *"starting points, not exact targets"* to the DES-052-proven pattern:
   *"these values are FIXED and will be used verbatim by the renderer
   regardless of what you output — do not attempt to adjust them."* Not
   load-bearing for correctness once (1)-(2) ship (a structural guard, not
   prompt discipline, is what actually holds); worth doing in the same
   change so the LLM isn't wasting effort or misleading a human reading the
   prompt.

**Known risk this design must not paper over — prose/values contradiction**
(credit: `site_delivery_and_editor`, boxingonline.com finding, §1). Pinning
makes `reference_values` a *hard* constraint, including when it contradicts
its own `colour_mood` prose written by the same classifier pass — today an
LLM occasionally notices and corrects such a contradiction on a later render;
after this fix, it structurally cannot. **This RFC ships a cheap advisory
check alongside the hard pin, not as a separate future ticket**: on merge,
log (not block — this is advisory, matching the existing `claimed_but_
ignored` diagnostic pattern already in this file) when `design_intent.
colour_mood`'s stated polarity (light/dark, warm/cool keyword scan) disagrees
with `reference_values`'s actual luminance. Same shape as
`enforceLayoutScheme`'s own scheme check, reused rather than reinvented where
possible. This does not block the RFC's core fix; it is scoped as part of the
same commit because shipping the hard pin without it makes exactly
boxingonline.com's failure class *permanent* rather than merely undetected.

**What this deliberately does not change:**
- Sites with no `reference_values` (fresh builds, or an adopted site whose
  classifier wrote only prose `colour_mood`) behave exactly as today —
  `pinned=false`, DES-052's "creative freedom" licence is untouched.
- Specialised palette slots (heading/hero_title/cta_bg/etc.) are already
  always theme-wins (`isCorePaletteKey` returns false for them) — no change.
- `resolve_composition_pallette_action.go`/`_typography_action.go` (the
  composition-time resolvers) are untouched — the deterministic resolution
  cascade is already correct; the defect is purely in the render-time
  re-merge.
- `bugs_open/396` (css-patch-agent's appended contrast repairs clobbered by
  the next full render) is a **sibling, different** bug in the same pipeline
  family — different mechanism (a full-row overwrite vs. a permissive merge),
  different fix, not addressed here. Do not cross-credit or conflate the two
  when either is closed.

## 3. Alternatives considered

- **Do nothing; keep relying on prompt wording alone (the status quo).**
  Ruled out by the evidence in §1: the DES-052 prompt-only fix already tried
  this for the "invents from nothing" failure mode and it worked for that
  case; it demonstrably does not stop the LLM from *adjusting* a value it is
  told is only a starting point, and even if a model perfectly obeyed its
  system prompt today, "prompt discipline" is not a control this platform
  trusts elsewhere on a shared render path (see the general "a doc comment
  enforces nothing" pattern already recorded in this codebase's own
  practice). No code-level backstop exists today.
- **Extend `site_specs.pinned` to the `design_intent` aspect instead of a new
  render-merge parameter.** Considered and rejected for this RFC: `pinned` is
  currently a per-ASPECT-ROW flag consulted by exactly two hand-written call
  sites for one aspect; wiring it into the generic `site_specs` read/write
  path would itself be a second, larger shared-contract change (touching
  `WriteSiteSpecAction`/`ReadSiteSpecAction`, used by every spec aspect in
  the platform), a bigger blast radius for no behavioural gain over reading
  `reference_values` presence directly — which is exactly the signal the
  prompt template already keys off, so the two can never disagree.
- **Route through a full re-architected "decision record" (RFC_015's GUARD
  face).** RFC_015 (`decision_records_allow_change_forbid_regression.md`,
  owner-ratified 2026-08-08) already names this exact gap — *"the palette pin
  genuinely works... but `pinned` is inert... Neither [steer nor guard] alone
  suffices"* — but its shipped implementation targeted a different write
  seam (`apply_section_edit`/`save_page_sections`, page-content regression),
  with a citation-gate mechanism (`acknowledges_decision`/`supersedes_
  decision`) that presumes an editor per edit, not a per-run LLM merge. Fully
  generalizing RFC_015's GUARD face to palette/typography would be a larger,
  independently-justifiable RFC; this RFC takes the smaller,
  `enforceLayoutScheme`-shaped step now and leaves that generalization open.

## 4. Blast radius, named

Mechanically derived, `go list -deps ./cmd/<target>/... | grep platform/
orchestration/actions$`, run 2026-09-02: `agent-chassis`, `backfill-tool-
crosslinks`, `component-render-check`, `config-key-audit`, `core-manager`,
`instanceaudit`, `test-spawning` all import the changed package.

**Behaviour actually changes** only where `RenderCSSFromSpecAction` executes
— the orchestration engines that run registered workflow actions
(`agent-chassis`, `core-manager`). The other five import the package
(registry/spec introspection, tooling) but do not themselves invoke this
action in their own entry points; they merely relink. This distinction should
be re-verified by whoever implements the RFC before merge, not taken on this
draft's word alone.

Within the changed package: two functions
(`buildPaletteMap`/`buildTypographyMap`) and their one call site. No other
caller of either function exists outside their own tests (confirmed by the
`enforceLayoutScheme` placement comment already in the file, which documents
and warns against a second caller appearing unnoticed).

## 5. Staged rollout plan

1. **Land the code + prompt change together** (small enough that splitting
   buys nothing — the prompt change alone is inert without the structural
   guard, and the structural guard without the prompt change just makes the
   LLM's now-ignored output a wasted generation, not a hazard).
2. **Canary**: re-run `needs_design` against 2-3 already-pinned sites
   (candidates: the three sites migration 350 pinned; any site with a current
   `theme_kit_adoption` spec once theme-kit Phase 1 is live) and diff the
   served palette before/after. Confirm an UNPINNED site's `needs_design` run
   is byte-for-byte unaffected (a synthetic control: same site, same run,
   compare against a build from the pre-RFC binary).
3. **Fleet**: no special gating beyond the normal image build/roll — this is
   a pure code change with no schema migration, so it activates on the next
   `webdesign-agent` deploy for every site whose `reference_values` happen to
   be populated (existing sites; no backfill needed or wanted — an unpinned
   site's behaviour is deliberately unchanged).

## 6. Rollback plan

Image rollback only — no migration, no schema change, `pinned` is computed
from existing data at read time, not stored. The previous binary tolerates
the (unchanged) data shape entirely; rollback is "redeploy the prior image,"
nothing else. Named loudly because the template requires it, even though
there is genuinely nothing else to say here.

## 7. Acceptance evidence (to be filled in as the RFC implements)

- [ ] Pod-grep / behavioural probe: on a pinned site, trigger a `needs_design`
      pass pre-fix (confirm the served palette DOES drift from
      `reference_values` — the induced-fault half, not just a happy-path
      grep) and post-fix (confirm it does not).
- [ ] Same probe on an unpinned site, both before and after, confirming no
      behavioural change.
- [ ] The prose/values advisory check's log line observed firing on at least
      one real contradictory `design_intent` row (a positive control it can
      actually detect something, not just that it compiles).
- [ ] A week-later census: no NEW colour-churn-shaped finding
      (`check_generic_theme` or an operator report) on a site that was pinned
      before this shipped.
