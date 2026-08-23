# CALIBRATION — 2026-08-23 — what Phase B's widening would actually do to the fleet

**Question.** `bugs_open/308` asks for the two CTA **writers**' label-match candidate universe to
be widened from `candidatesFromHubs` (section-index + tool/game, utility areas dropped) to the
**detector's** universe (`loadCTAMatchIndex`: every page, minus index/home). Before writing that
change: how many live CTA destinations move, and how many of the moves are wrong?

**Answer, in one line.** The widening is **not** a small change — it multiplies the fleet's CTA
rewrite volume by ~13x — and **31% of the writes it would perform are decided by an alphabetical
tie-break**, i.e. by nothing at all. A ranking key cannot fix that (two were measured and both
trade one arbitrary winner for another, exactly as the 2026-08-11 calibration found). **Refusing
the ambiguous match** does: it removes the coin flips, and 308's own repairs survive it.

---

## Method — the same shape as `CALIBRATION_2026-08-11_label_match_identity_report.txt`

- A throwaway harness (`calib308`, scratchpad, not committed — same convention as the deleted
  `cmd/ctacalibrate`) that **imports the real `datahelpers` package**. `BestLabelMatch`,
  `LabelTokens`, `NewLabelMatchCandidate`, `DeriveCTAURLFields`, `ParseInputSchemaValue`,
  `NormalizePagePath` and `PageURLSet` are the shipping ones, never re-implemented.
- Both candidate pools and the validity set are a **line-for-line mirror of the real loaders**,
  read alongside the source: `loadContentHubs` / `loadInteractivePages` / `loadResolverPageSet` /
  `candidatesFromHubs` for the resolver; `loadCTAMatchIndex` for the detector.
- Inputs **frozen once** to JSON via `kubectl exec … psql` (no key material, no port-forward), so
  every variant below is compared against byte-identical data:
  **829 page rows** (764 not deleted/archived) and **667 `ctaFieldNames` page_components rows**,
  pulled **2026-08-23 ~12:35Z**.
- Population examined: **1,266** CTA url fields whose paired label field (schema-derived, exactly
  as `resolve_internal_links_action.go` derives it) is non-empty.

### The control, and it could have failed

The harness needs a local copy of the ranking so alternative keys can be tried (the real
candidate's token sets are unexported). That copy is only trustworthy if it reproduces the real
function, so every row is scored **both ways** on the real pool:

> **1,266 label/pool pairs compared against `datahelpers.BestLabelMatch`, 0 disagreements.**

A mismatch would have printed per row and invalidated every number below. This is the
disconfirming test; it is not a restatement of the code.

---

## 1. Pool comparison — what widening alone does

| | narrow (today) | wide (detector's universe) |
|---|---|---|
| CTA url fields examined | 1,266 | 1,266 |
| labels that match **anything** | **781** | **1,146** |
| picks that move (narrow → wide) | — | **516** |
| …of which land in a utility area | — | **96** |

**Writes** — a write is a pick that differs from what is stored, i.e. a live button changing
destination on its next build or `cta_links_stale` rerender:

| | narrow (today) | wide |
|---|---|---|
| writes over a **stored, valid** url | 30 | **433** |
| writes over a stored **invalid** url (a repair) | 2 | 2 |
| writes into an **empty** field | — | 189 |
| **writes over a non-empty stored url** | **32** | **435** |

> **CORRECTED 2026-08-23 (later the same day):** the 435 above counts every pick that differs
> from the stored value. **Both writers additionally gate the write on `validPages.Contains`**
> (`setCTAField`, `applyCTARecompute`), and applying that gate — as the shipping code does — the
> figure is **428**, of which **92** land in a utility area (not 96). §4's post-refusal figure is
> **291**, not 298. The corrected numbers are the ones §8 compares. What caught it: re-running the
> same population through a harness path that mirrored the writers' *full* branch condition rather
> than the match alone — i.e. the first version measured the matcher, not the writer.

**The widening multiplies fleet CTA rewrite volume by ~13x (32 → 428).** 308's own population is
147 findings (§4), so **roughly two thirds of what Phase B would rewrite is not what 308 is
about.** That is the number to hold in mind: this change is far wider than the bug that motivates
it, and its safety cannot rest on the bug's own examples looking right.

## 2. 31% of those writes are coin flips

`BestLabelMatch` ranks: identity overlap → total overlap → interactive → **name ascending**. That
last key is arbitrary — it is alphabetical order — and on the wide pool it decides a lot:

| | narrow | wide |
|---|---|---|
| matches whose winner is decided **only** by the alphabetical tie-break | 177 / 781 | **263 / 1,146 (23%)** |
| …of which would **overwrite a stored url** | 19 | **137 of 435 (31%)** |

Two hand-audited families from the live findings, both of which the widening would have
**executed**, and both of which are this defect:

- **finetuning.uk, "how we work"**, stored `/how-we-work.html` — the page the copy names.
  `[work]` ties `how-we-work` (its own name) against `about`, whose *title* is
  "About Finetune | Who We Are and **How We Work**". Alphabetical → `/about.html` wins.
  **13 findings** (2026-08-23) tell the platform to move a correct link to the About page.
- **dartsonline.com, "Read the guides" / "Browse all guides"**, stored `/guides/index.html`.
  `[guides]` ties `guides-index` against `about`, whose title ends "…Spec-First Darts **Guides**".
  Alphabetical → `/about.html`. **6 findings.**

## 3. Two candidate ranking keys — measured, and REJECTED

Both were run over both pools, against the frozen dump.

| key | picks changed vs live (narrow / wide) | verdict |
|---|---|---|
| **NAME first** (label token in the candidate's own `name` outranks a title-only match) | 43 / — | **rejected** |
| **DEPTH before interactive** (an area's own index outranks its children on a tie) | 61 / — | **rejected** |
| NAME + DEPTH (either placement) | 71 / — | **rejected** |

They fix real cases — `NAME` sends gamesdesign.co.uk's five "Browse … Tools" labels to
`/tools/index.html` instead of a random comparator, and repairs the two families in §2 — but each
breaks others, on the same mechanism it was meant to cure:

- `NAME` moves ai-agent-orchestration.com's *"Try the Password Strength Physics tool…"* **off**
  `/tools/password-entropy.html` (correct today) onto an automation-savings estimator, because this
  estate names every tool page `tool-…`, so the token **`tool` sits in every tool page's name** and
  hands them all a name-overlap point.
- `DEPTH` moves loancalculator.co.uk's *"Run the numbers on a loan"* onto `/guides/index.html`,
  because pulling section indexes up demotes the interactive preference wholesale.

**This is the 2026-08-11 finding again** (that report dropped a token-set-size key for the same
reason): *a tie at overlap 1 carries no signal, so any key that breaks it is deciding by an
artefact.* Recording it here so the third session does not try a fourth key.

## 4. The recommendation — REFUSE the ambiguous match

If the winner is tied with a different page on identity overlap, total overlap **and**
interactivity — so only alphabetical order separated them — report **no match**.

For a writer that means the keep branches hold the stored value. For the detector it means no
finding is filed. Measured on the frozen dump:

| | narrow | wide |
|---|---|---|
| matches refused as ambiguous | 177 | 263 |
| writes stopped | 19 | **137** |
| **writes that survive** | 13 | **298** |
| refused picks that were a utility page | 0 | 29 |

**308's own repairs survive**, which is the test that mattered:
`"Book a discovery call" → /contact.html` (28 + 6 + singles, finetuning.uk),
`"Contact our supply team" → /contact.html` (gaswholesalers.com), `"Get in touch" → /contact.html`
(cookly.uk), `"Learn More About Us" → /about.html`. And it stops both §2 families outright.

**Residual it does NOT catch, stated rather than hidden:** a *confident* false match — dartsonline's
*"See how each brand differs, spec by spec"* → `/about.html`, won on `spec` because the About
page's title reads "Spec-First Darts Guides". That is an identity-overlap win, not a tie, so no
tie rule can see it. Stopwords cannot either (`spec` is distinctive). It stays open.

## 5. Scope — how much of 308 the widening can reach AT ALL

[MEASURED 2026-08-23, live]

| | findings |
|---|---|
| `misdirected_cta` findings fleet-wide | **1,855** |
| …sitting on a `ctaFieldNames` component | 771 |
| …**and** whose href is a value in that component's `content_data` | **675 (36%)** |
| findings suggesting a utility destination (this bug's population) | **188** |
| …on a `ctaFieldNames` component | 162 |
| …**repairable by the two writers** | **147 (78%)** |

**41 of this bug's 188 findings can never be fixed by widening the writers' candidate set**, because
the anchor is not a CTA field — it is a link inside prose or in a component outside `ctaFieldNames`
(e.g. slot `ported-page`). Phase B closes 147; the rest need a different mechanism and should not
be counted as closed.

## 6. Two of this lane's own dated claims have expired

- **"200 findings" (2026-08-22) is now 188** (same predicate, re-run 2026-08-23). Split:
  `complete` 63 items/99 findings, `unresolved` 53/86, `cancelled` 2/2, `failed` 1/1. Nothing was
  fixed — the stock moved because items changed status. A stock, not a flow, exactly as the
  runbook warns.
- **"The detector has been silent since 2026-08-19" is FALSE as of 2026-08-22**: 40
  `misdirected_cta` items were filed that day (08-19: 3, 08-18: 128, 08-17: 208). The
  `[INFERRED]` attribution to `bugs_open/230` should not be carried forward without re-checking.

## 7. What this report does NOT establish

- **Precision is judged by hand on samples**, not measured. There is no ground truth for "the label
  names this page"; the §2 families were verified by reading the candidate rows, and the §4
  survivors by reading the labels. A systematic audit would need a human pass over 435 rows.
- The `copy names the current page` proxy in the harness (label tokens ⊆ the stored target's own
  name tokens) fires on only 3 rows fleet-wide. It is a **weak** detector of regressions and is
  reported as such — it is not the basis of the §4 recommendation; the tie census is.
- **The detector-side cost of refusal is a recall trade that has not been audited case by case**:
  263 anchors would stop being classified as "names a page". Two families (19 findings) are known
  false and stop correctly; the rest are unexamined.

---

## 8. OPTION B — "add only the utility pages" — measured, and it is WORSE

The obvious cheaper change is to widen the pool by **only** the utility-area pages
(contact/about/privacy/terms/legal, any `page_type`) and leave every other content page out. 308's
whole population is utility suggestions, so this looks like the minimal sufficient fix, and it
rewrites a third as much.

| pool | matched | writes (validity-gated) | → utility | refused as ambiguous |
|---|---|---|---|---|
| **A** wide (detector's universe) | 1,146 | 428 | 92 | — |
| **A'** wide **+ tie refusal** | 883 | **291** | **66** | 263 |
| **B** utility-only widening | 883 | 136 | 105 | — |
| **B'** utility-only **+ tie refusal** | 702 | **108** | **96** | 181 |

**B' writes FEWER links and MORE of them are wrong.** It makes 26 utility repairs that A' does not
make, and reading them is the argument:

| B'-only "repair" | why it is wrong |
|---|---|
| dartsonline "Check the tungsten percentage guide first" → `/about.html` | the tungsten guide page exists; it is **not in B's pool**, so About wins by default |
| finetuning "How We Work" / "See how we work" → `/about.html` | `/how-we-work.html` is not in B's pool either |
| finetuning "Talk to Us About Your Setup" → `/about.html` | matched on the token **`about`** — the false match this bug file predicted |
| fundamentallyai "Talk to us about a recovery" → `/about.html` | same |

**A' makes zero utility repairs that B' misses** — the wide pool is a superset, so it can only
change a pick by finding something *better*. The mechanism is the point: adding utility pages to a
pool that still excludes the label's real target does not give the matcher a choice, it gives it a
**monopoly**. A narrow widening is not a safer widening; it is a widening with the right answers
withheld.

This also settles the "recalibrate the stopwords" item the bug file carries as a Phase B
constraint: `about` in `LabelStopwords` would suppress the four `Talk to us about …` cases above,
but it would equally suppress *"Learn More About Us"* → `/about.html`, which is a **correct**
repair A' makes. The lever that separates them is not the stopword list — it is having the real
target in the pool, plus refusing the ties. **Do not add `about` to `LabelStopwords`.**
