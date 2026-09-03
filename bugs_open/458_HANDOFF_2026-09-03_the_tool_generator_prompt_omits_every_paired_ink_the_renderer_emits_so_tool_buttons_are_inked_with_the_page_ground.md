# 458 — the tool-generator prompt omits every paired ink the renderer emits, so tool buttons are inked with the page ground

**Filed** 2026-09-03 by the `site_ai_agent_orchestration` lane, which found it going upstream
from four per-page contrast failures that would not stop recurring
(`docs/agent_docs/docs024_key_docs_latest/site_ai_agent_orchestration/HANDOFF_2026-09-02_continue_here.md` §3).

**This is a class fix, not an instance fix.** That lane has shipped four per-page migrations
(`456`, `469`, `625`, `636`) and the defect kept arriving on pages that did not exist the week
before. It arrives because the prompt that writes tool markup is still the only component-writing
prompt in the fleet that was never taught the on-colour pairing rule.

---

## 1. The mechanism, in one paragraph

The renderer computes and emits **paired ink tokens** — `--color-primary-text` (the ink that goes
**ON** a primary fill), `--color-primary-ink` / `--color-accent-ink` (primary/accent themselves,
re-tinted to be legible **as ink on the page**). `buildLegibleInkDefaults`
(`platform/orchestration/actions/palette_specialised_slots.go:670-700`) documents the distinction
in its own header comment: *"TWO DIRECTIONS, AND CONFLATING THEM IS THE MISTAKE THIS COMMENT EXISTS
TO STOP."*

The `tool-generator` prompt names **none of them**. Its whole colour vocabulary is:

> `4. Use CSS custom properties for colours: var(--color-primary), var(--color-secondary), var(--color-accent), var(--color-background), var(--color-surface), var(--color-text), var(--color-text-muted), var(--color-border)`
>
> — live `agent_definitions.default_config`, `type='tool-generator'`,
> `.workflow.steps.generate_tool_html.config.prompt_template`, read 2026-09-03

Eight tokens: fills and page inks only. **No pairing rule and no paired ink.** So an LLM that fills
a button with `var(--color-primary)` has to invent a text colour, and the idiomatic guess — "ink it
with the page's own ground" — is what it writes:

```css
.submit-btn { background: var(--color-primary); color: var(--color-background); }
```

That is correct on a palette whose primary is saturated. It is a **guaranteed failure** on a palette
whose primary sits near its ground.

## 2. The contrast: `component-creator` already got this right

The prompt that writes ordinary (non-tool) components teaches the exact rule, in the two-level form:

> `(b) PALETTE BAND: background: var(--color-primary) and re-export the on-colour family:`
> `--section-text: var(--color-primary-text, var(--color-background));`
>
> and lists `--color-primary, --color-primary-hover, --color-primary-text`
>
> — live `agent_definitions.default_config`, `type='component-creator'`, `.prompt_template`, read 2026-09-03

**So this is not a missing mechanism and not an unsolved design question.** The estate already has
the rule, already has the tokens, and already applies them in the component path. Only the tool path
was left behind.

## 3. The measurement that could have come out otherwise [MEASURED 2026-09-03]

If the prompt is the cause, the defect should sit in the tool-generated population and **not** in the
component-generated one. It does, with no leakage:

```sql
WITH c AS (
  SELECT (category ILIKE '%tool%' OR function ILIKE '%tool%' OR name ILIKE 'tool-%') AS is_tool,
         html_template ~ 'background:\s*var\(--color-primary\)'            AS primary_fill,
         html_template ~ '--color-primary-text'                            AS pairs_correctly,
         html_template ~ 'color:\s*var\(--color-(background|surface)\)'    AS pairs_with_ground
  FROM content_components WHERE is_active AND forked_from IS NULL)
SELECT is_tool, count(*) AS components,
       count(*) FILTER (WHERE primary_fill)                                        AS fills_with_primary,
       count(*) FILTER (WHERE pairs_correctly)                                     AS uses_primary_text,
       count(*) FILTER (WHERE primary_fill AND pairs_with_ground AND NOT pairs_correctly) AS fill_inked_with_ground
FROM c GROUP BY 1;
```

| | non-tool | tool |
|---|---|---|
| active unforked components | 151 | 261 |
| fills with `--color-primary` | 31 | 174 |
| uses `--color-primary-text` | 59 | 31 |
| **primary fill inked with the page ground** | **0** | **148** |

**Zero and 148.** A 40/60 split would have refuted the prompt theory; this did not.

## 4. Where it actually breaks [MEASURED 2026-09-03]

The shape only fails where primary is close to the ground. Of **59** palette rows carrying both
keys, **9** score under 3.0:1 and **7** under 1.25:1 — a label that is, in practice, not there:

| palette | primary | background | ratio |
|---|---|---|---|
| `palette-loancash-co-uk` | `#e8f5ee` | `#e8f5ee` | **1.00** |
| `palette-ai-agent-orchestration-com` | `#0D1117` | `#080B10` | **1.04** |
| `palette-agritec-uk` | `#1A1F2E` | `#12151F` | **1.11** |
| `palette-dartsonline-com` | `#1A1F2E` | `#111520` | **1.11** |
| `palette-robot-hands-com` | `#1A1F2E` | `#0F1218` | **1.14** |
| `palette-loanandmortgagecalculator-…` | `#e2e8f0` | `#f8fafc` | **1.18** |
| `palette-oufe-com` | `#1B2A3B` | `#111922` | **1.21** |
| `bold-conversion` | `#f97316` | `#ffffff` | 2.80 |
| `palette-mortgagecalculator-co-uk` | `#b59230` | `#f8fafc` | 2.82 |

⚠ **This is the STORED palette, not the served stylesheet.** The design overlay may legitimately
serve different values (owner ruling 2026-09-02 — `reference_values` is advisory and the machine may
override it), so treat this as the population to go and check, not as nine confirmed sites. It was
confirmed at the served artefact for `ai-agent-orchestration.com` only (§5).

> This also **refreshes a stale census inside the code**. `warnUnusablePrimary`
> (`palette_specialised_slots.go`) carries *"Measured fleet-wide 2026-07-27: 4 of 31 palettes score
> below 1.25:1 … (dartsonline.com, fundamentallyai.com, robot-hands.com, oufe.com)"*. As of
> 2026-09-03 it is **7 of 59**, and the membership has changed — `loancash`,
> `ai-agent-orchestration`, `agritec` and `loanandmortgagecalculator` are new to it. A count that
> grew by addition and read as current for five weeks; the owner's 2026-08-22 dating rule exactly.

## 5. Confirmed at the served artefact — `ai-agent-orchestration.com`

Served `https://ai-agent-orchestration.com/assets/css/styles.css`, read 2026-09-03:

```css
--color-primary:      #0D1117;      --color-surface:    #0D1117;   /* identical */
--color-background:   #080B10;
--color-primary-text: #ffffff;                 /* the repair for a FILL — already served */
--color-primary-ink:  #768eb2;                 /* the repair for an INK — already served */
--color-accent-ink:   #F0A500;
```

The two failing tool components, and what the already-served tokens would give:

| component | declaration | now | with the paired ink |
|---|---|---|---|
| `tool-model-approach-selector` `.submit-btn` | `background: var(--color-primary); color: var(--color-background)` | **1.04:1** | `var(--color-primary-text, #fff)` → **18.92:1** |
| `tool-token-calculator` `.stat-value` | `color: var(--color-primary)` in `.stat-box{background:var(--color-surface)}` | **1.00:1** | `var(--color-primary-ink, var(--color-primary))` → **5.66:1** |
| `tool-model-approach-selector` `.error-msg` | `color: var(--color-primary); background: var(--color-surface)` | **1.00:1** | same as above → **5.66:1** |

**Both repairs use tokens this site already serves.** Nothing needs to be built and no palette needs
to be changed — which matters, because changing the palette is the withdrawn `RFC_059` and is not
available (see §7).

⚠ **`.error-msg` is a third instance the page audit never reported**, because the audit can only
measure the states the page renders. Error, hover and `:checked` states carry instances that no
render-time contrast probe will ever see. **Do not size this class from audit findings.**

## 6. The second half — the validator calls the correct repair "unknown"

`canonicalCSSTokens` (`platform/orchestration/actions/component_validation.go:175-191`) is the
allow-list `AuditTemplateTokens` checks templates against. It contains `--hero-ink`. It does **not**
contain `--color-primary-ink`, `--color-accent-ink` or `--color-cta-bg-ink` — all three of which the
renderer emits.

```
--color-primary-ink   0
--color-accent-ink    0
--color-cta-bg-ink    0
--hero-ink            1
```

So a component that opts into the documented repair is audited as referencing an unknown custom
property. It is advisory — `AuditTemplateTokens` only logs a Warn and *"the template is always
allowed to persist"* — so this blocks nothing today. **It is filed because it points the signal the
wrong way**: the mechanism is discouraged at the consuming end while being unmentioned at the
producing end, which is a fair part of why adoption is **15 of 412** active unforked components.

## 7. What the fix may NOT be

**Do not repair this by separating primary from surface in the palette.** `RFC_059` proposed making
`reference_values` binding and the **owner withdrew it** on 2026-09-02: *"by default it can start
with a theme and change it if it wishes, but it must have full authority to ignore our set of themes
if it chooses."* A fix that pins the palette is re-proposing a withdrawn RFC. The lever is what the
generator is told, never what the overlay is allowed to choose — which is also why the paired-ink
route is the right one: it is **palette-agnostic** and stays correct whatever the overlay picks.

## 8. Fix candidates, ranked by what closes the door

1. **Teach `tool-generator` (and `tool-improver`) the pairing rule that `component-creator`
   already carries.** Migration against `agent_definitions.default_config`. Highest value: it stops
   new instances at the only place they are minted, and it copies proven in-production wording
   rather than inventing a contract. Does not repair the 148 existing components.
2. **Add the three renderer-emitted ink companions to `canonicalCSSTokens`.** One-line Go change;
   makes the audit signal point the right way and lets adoption be measured honestly.
3. **Re-generate or migrate the 148 existing tool components.** Large, and only 9 palettes make it
   visible — so scope it to the affected sites rather than the fleet.
4. *(rejected — see §7)* pin the palette.

⚠ **Candidates 1 and 2 do not repair a single existing page.** They stop the inflow. The lane that
picks this up should expect the 09-02 page-level failures to persist until candidate 3, and should
not read a still-failing page as evidence that 1 and 2 did not land.

## 9. How to verify

- **That the prompt changed:** read the live row, not the seed —
  `SELECT default_config::text LIKE '%--color-primary-text%' FROM agent_definitions WHERE type='tool-generator' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;`
- **That it took effect:** generate one tool on a site in the §4 list and census the new component
  with the §3 query — `fill_inked_with_ground` must be 0 for it. A prompt change is inert until an
  agent actually runs; DB config is live immediately but nothing regenerates on its own.
- **That the allow-list changed:** `go test ./platform/orchestration/actions/ -run TokenAudit`.
- **At the artefact, never at the stylesheet:** `getComputedStyle`, per the aiao lane's R2 — the
  served site stylesheet is routinely not the winning declaration.

## 10. Diagnosis loop

Filed to `090` before this file asserted its root cause, per CLAUDE.md's "always file when
cross-cutting" rule. Intake correlation `e4194dc2-effc-4706-9744-4239c99e9010`, run correlation
`6a317110-e9b7-4682-bbaf-8f2852e93e98`. **Verdict not yet returned at filing time** — this section
must be updated with it, including if it REFUTES the above.
