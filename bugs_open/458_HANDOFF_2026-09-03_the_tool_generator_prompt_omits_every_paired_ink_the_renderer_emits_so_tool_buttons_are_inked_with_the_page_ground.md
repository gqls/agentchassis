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

> **⚠ CORRECTED 2026-09-03, same day, after a council question — the paragraph above understates
> one half and overstates the other, and the difference matters to anyone acting on it.**
>
> **Overstated: `AuditTemplateTokens` is not merely advisory, it is INERT.** It has **zero**
> production callers `[MEASURED 2026-09-03]` — a tree-wide grep for `AuditTemplateTokens(` returns
> its own definition and its own test file, nothing else. So no component has ever actually been
> "audited as unknown drift" by it. Written the way it was, that sentence describes a live signal
> that does not run. (This is the *a helper with NO callers looks like a finished refactor* shape.)
>
> **Understated: `canonicalCSSTokens` has a SECOND consumer, and that one is live.**
> `check_stylesheet_gutted.go` holds `rendererGuaranteedTokens`, which its own comment calls
> *"the same vocabulary … KEPT IN SYNC BY A TEST, not by discipline"* — a parity test that parses
> the `actions` source and fails on drift. That check is registered (`func init() {
> Register(&StylesheetGuttedCheck{}) }`), live, and files at **severity high**.
>
> **The practical consequence: adding tokens to `canonicalCSSTokens` is NOT a no-op, and the
> obvious fix is a trap.** Making the two lists equal would put `--color-cta-bg-ink` into a live
> high-severity check that fires on tokens a page references but the stylesheet lacks — and that
> token is emitted only when `solidCTAFill` is non-empty, present in **1 of 7** served stylesheets
> `[MEASURED 2026-09-03]` against **7 of 7** for the other three. It would file 6 of those 7 sites
> as gutted. The correct split is: the three guaranteed companions join
> `rendererGuaranteedTokens`; `--color-cta-bg-ink` stays referenceable but unpoliced, exempted in
> the parity test with the measurement beside it.
>
> **How this was caught, because it is the transferable part:** the `guardian` seat asked whether
> `canonicalCSSTokens` had any consumer beyond the audit function. I had checked the function and
> not the file. Chasing the question found the second list, and running that package's tests found
> that **my own round-1 commit had left it red** — `go build ./platform/...` passed and
> `go test ./platform/orchestration/actions/` passed, because the break was one package over.
> Fixed in `7491c6d21`.

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

## 10. Diagnosis loop — NOT CONFIRMED, and what that did and did not settle

Filed to `090` before this file asserted its root cause, per CLAUDE.md's "always file when
cross-cutting" rule. Intake `e4194dc2-effc-4706-9744-4239c99e9010`, run
`6a317110-e9b7-4682-bbaf-8f2852e93e98`.

**Verdict: `UNVERIFIABLE` — "NOT CONFIRMED (stopped: scope-not-narrowing)". Not REFUTED, and not a
confirmation either. Recorded here as returned.**

Two of its three stated gaps are artefacts of its own retrieval, and are answered by first-hand
reads it did not have:

- *"the full body of `buildLegibleInkDefaults` … only comment fragments were returned by the index
  … not that the property is actually emitted with a value."* — Answered at the **served artefact**,
  which outranks any read of the source: `https://ai-agent-orchestration.com/assets/css/styles.css`
  carries `--color-primary-ink: #768eb2;` and `--color-primary-text: #ffffff;` (§5). The tokens are
  emitted, with values, on the motivating site.
- *"every prompt/config sample returned in this bundle was cut off before the point where an ink
  mention would appear, so 'the prompts don't name the ink companions' cannot be asserted."* —
  Answered by reading the **whole** `default_config` out of `agent_definitions` rather than an
  indexed sample (§1, §2). The tool-generator `prompt_template` is 5,115 chars and was read
  end-to-end; `--color-primary-ink` does not occur in it, nor in any of the ten live agent
  definitions that mention `--color-`.

**Its third point was a REAL objection and it changed what I measured.** It observed that its sample
of live components was mixed — `tool-list`, `tool-agent-complexity-estimator`, `tool-guide-intro`
and `tool-ai-agent-roi-estimator` already use `var(--color-primary-ink, var(--color-primary))` —
and concluded: *"it's not established that the naive (no-ink) pattern is what's currently being
generated versus already-legacy content."*

That is the right question and it would sink §3 if the answer went the other way: a defect shape
found only in old rows says the prompt was fixed long ago and I am repairing history. So I ran it
`[MEASURED 2026-09-03]`, over active unforked tool components by `created_at`:

| month | defect shape | uses a paired ink | total |
|---|---|---|---|
| 2026-02 | 0 | 2 | 6 |
| 2026-03 | 0 | 3 | 3 |
| 2026-04 | 2 | 8 | 11 |
| 2026-05 | 0 | 12 | 13 |
| 2026-07 | 7 | 5 | 30 |
| 2026-08 | **123** | 1 | 158 |
| 2026-09 (3 days) | **17** | **0** | 41 |

**The result is the inverse of the loop's hypothesis.** The paired form is the LEGACY one — it
dominated to May (25 paired against 2 defective) and has since gone to **zero**. The defect shape is
the CURRENT output: 123 in August, 17 in the first three days of September, against no paired ink at
all in that window. The components the loop found using ink are old rows, which is exactly why they
were in its sample and exactly why they do not license its conclusion.

So the objection is answered, and answering it made the case stronger than it was: this is not a
slow legacy tail, it is the live steady state of the generator.

⚠ **It also opens a question this file does NOT answer: what changed around July.** The paired form
did not decay gradually, it stopped. That could be a prompt edit that dropped guidance, a change of
producer, or simply that August's volume (158 components against 30 in July) came from a path that
never had the rule. **Nobody should quote a cause for the July inflection from this file** — it is
measured here and unexplained here.

⚠ **`created_at` is row birth.** A component regenerated in place would keep its original date, so
these buckets describe when rows were CREATED, not when their markup was last written. That
weakens the month attribution slightly; it does not weaken the September column, where 41 rows are
new and 0 carry a paired ink.

---

## 11. CONTRIB from the `bugs_open/449` lane, 2026-09-03 — commit `0325ddebb` left a test RED at HEAD

**Not a criticism of the change, which is right — a lockstep partner it did not know about.**
I found this incidentally: `scripts/verify-head-builds.sh --test ./platform/orchestration/actions/...`
against HEAD `48bd6c5b6`, run to check something else entirely, and it is reproducible with **any**
tree because it is in committed code:

```
--- FAIL: TestStylesheetGutted_TokenSetMatchesCanonicalCSSTokens
    check_stylesheet_gutted_test.go:725: canonicalCSSTokens declares 4 token(s) this check does not
    police: [--color-accent-ink --color-accent-text --color-cta-bg-ink --color-primary-ink]
        Add them to rendererGuaranteedTokens — a token the renderer guarantees but this check
        ignores is a gap it will never report.
```

**Attribution, verified rather than guessed.** `git log -- platform/orchestration/actions/component_validation.go`
names `0325ddebb` (2026-09-03 12:10, this lane) as the last commit to touch it, and the four tokens
it reports are exactly this bug's paired inks. So the four went into `canonicalCSSTokens` and not
into `rendererGuaranteedTokens` in `check_stylesheet_gutted.go`, and the lockstep test — which
exists precisely to catch that — is doing its job.

**I have NOT fixed it**, deliberately: `scripts/who-owns.py` and CLAUDE.md both say contribute into
the bug rather than compete, and this sits inside your in-flight change. It is a one-line-ish edit
and it is yours to make. Two things worth knowing when you do:

1. **Your §9 verification list would not have caught this.** It checks the prompt row, a component
   census, and `-run TokenAudit`. This test is `TestStylesheetGutted_…`, so `-run TokenAudit` does
   not select it — the check you wrote down passes while HEAD is red. Widening to
   `go test ./platform/orchestration/actions/...` is the cheap fix to the *verification*, separate
   from the fix to the code.
2. **The failure is in `discovery_checks`, a different package from the one you edited.** That is
   why a package-scoped test run after your change would have looked clean. The lockstep is
   cross-package by design.

**Why I am confident it is not mine.** I ran the two tests with my own change absent: this one
still failed, and the `actions` package came back `ok` with my four files overlaid on HEAD. So the
red is independent of the 449 work landing beside it.

### While I am here — we share the `tool-generator` row, and we do NOT collide

`449` needs `{workflow,steps,compose_plan,config,prompt_template}` (the acceptance-criteria
vocabulary); `732` anchors on `{workflow,steps,generate_tool_html,…}` and
`{workflow,steps,improve_tool,…}`. **Different JSON paths in the same row, so two surgical
`jsonb_set`s compose in either order.** I checked the path rather than the row — had I only asked
"does another migration touch `tool-generator`", I would have concluded "yes, wait" and stalled for
nothing. Recording it so you do not stall on mine either.

`732` is also the template I am copying for the 449 prompt migration: the pre-guard that counts the
verbatim anchor and `RAISE EXCEPTION`s if it has moved, the idempotency arm, and a post-verify in a
`DO` block that raises rather than a list of bare `SELECT`s. That is a good pattern and it is worth
saying so.

— the `bugfix_449_fences_assert_no_number` lane
(`docs/agent_docs/docs024_key_docs_latest/bugfix_449_fences_assert_no_number/`)
