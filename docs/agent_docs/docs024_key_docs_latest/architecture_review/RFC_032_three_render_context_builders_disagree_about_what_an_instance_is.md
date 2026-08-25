# RFC_032 — Three render context-builders disagree about what an "instance" is, and `ComponentID` means two things

**Status:** **RULED 2026-08-22 (owner) — converge on `{{.InstanceID}}`; see §8 for the ruling,
the evidence pack §4 asked for, and a dated correction to §2's own table.** Filed 2026-08-16.
**Raised by the council gate twice**, on the same lane, in two
consecutive rounds — by the `architecture` seat in round 1 and by the `reuse_agent` seat in round 2
(correlation `07635a2f-3605-4e67-9a6d-7636b07f16ca`).

**This file exists because the deferral had no locator.** Round 2's submission said the unification
was "filed as the follow-up the architecture seat asked for". The `reuse_agent` seat objected, at
medium severity, that no work item, doc_plan or artifact id was given anywhere — *"a vague 'filed'
claim with no locator is not evidence a reuse-unification is actually tracked; it reads as the same
deferral that let the tool-creation split persist."* It was right: nothing had been filed. This is
the thing that was claimed to exist.

---

## 1. What the thing is, in plain terms

When the platform turns a stored component into HTML for a page, it first builds a **render
context** — a bag of values the template can reference, like `{{.Title}}` or `{{.ComponentID}}`.
Three different pieces of code build that bag, for three different situations: assembling a page
from scratch, re-rendering a page's sections, and rendering a single section on its own.

An **instance** is one placement of a component on a page. The same calculator placed twice on one
page is two instances of one component.

The problem is that the three builders do not agree on what identifies an instance, and one
placeholder name — `{{.ComponentID}}` — resolves to two genuinely different things depending on
which builder ran.

## 2. The defect

| builder | `ComponentID` resolves to | is it per-instance? |
|---|---|---|
| `assemble_from_library.go` | `component_<function>_<idx>` — built from the loop index | ~~**yes**~~ **NO — see §8.2, corrected 2026-08-22: the substitution is unreachable and 0 of 270 live placements carry this shape** |
| `rerender_page_sections_action.go` | `comp.ID` — the `content_components` row id | **no** — same for every instance |
| `v3_site_actions.go` (`RenderComponentAction`) | `comp.ID` — the same row id | **no** |

They are not even the same *shape*: one is a `component_<function>_<n>` string, the other a UUID.

A template that writes `id="{{.ComponentID}}-loanAmount"` is therefore namespaced on one path and
not on the other, **and on a page with one instance the two are indistinguishable** — so a canary
passes either way, and the defect only appears at the second instance, which is the case the author
was trying to enable. Five live components use `{{.ComponentID}}` today (`faq`,
`generic-text-block`, `mechanism-flow`, `evidence-timeseries`, `pricing`).

## 3. What `bugs_open/283` did instead, and why that is not the fix

283 added a **third** per-instance concept, `{{.InstanceID}}`, with one rule (component function +
occurrence on the page) and one derivation, bound on every render path. It is measured, tested,
council-approved and live (chassis `v1.0.1304`).

It deliberately did **not** touch `ComponentID`, because re-pointing it changes the served element
ids of those five live components — a change to what a shared mechanism *guarantees*, which is the
architecture-scope trigger under the owner ruling of 2026-07-29 §1.

So the estate now has **two names for adjacent guarantees** where it previously had one name for
two guarantees. That is a real improvement — the ambiguity is now visible in the vocabulary rather
than hidden inside it — but it is not resolution, and the `reuse_agent` seat is right that leaving
both live without a tracked unification is how the tool-creation novel/fork split persisted.

## 4. The question for architecture review

**Should `ComponentID` be re-pointed to the canonical per-instance identity, or formally re-named to
what it actually is on two of three paths (the component row id), or left as-is with the ambiguity
documented?**

Sub-questions the round should answer:

1. **What breaks if `ComponentID` becomes per-instance everywhere?** The five components' served
   element ids change. Measure which pages carry them and what addresses those ids —
   `oracle.py`-style checks, CSS, JavaScript, anchor links, and any external deep link.
2. **Is the third builder needed at all?** The architecture seat's round-1 words: *"The estate needs
   one canonical per-instance identity, not three ad hoc derivations sharing a name."* 283 supplied
   the canonical identity; unifying the *builders* is the remaining half.
3. **What is the migration shape?** A rename with both keys populated during a transition is
   available and cheap, because a render context is a map — the cost is deciding when to drop the
   old key, which is the part that never happens without a trigger.

## 5. ⚠ The trigger this RFC must not miss

The `architecture` seat approved 283 under the **RFC_022 narrow exception** — an opt-in field whose
unsafe default is OFF and which **no live consumer names**. Measured 2026-08-16: **0 of 243** active
component templates reference `{{.InstanceID}}`.

Its approval note is explicit about when that stops being true:

> *"The moment the 22 templates start consuming `InstanceID`, condition 3 of the exception (zero
> live consumers) stops holding and this becomes a real load-bearing contract across the component
> library. That conversion PR, not this one, is where an RFC or at minimum a fresh architecture pass
> belongs."*

**So the template-conversion work in `bugs_open/283` §9.6 is architecture-scope, and this RFC is its
gate.** The seat also asked for a *written trigger* — a mechanical signal fired the first time a
live template references `{{.InstanceID}}` — rather than a prose reminder. That trigger is **not
built**; see §6.

## 6. The trigger — BUILT AND RUNNING (2026-08-16); the unification is not

> **UPDATED 2026-08-16, hours after filing.** This section said the trigger was not built and
> assigned it to whoever converts the first template. It is built, deployed and has run.

**`instance-token-adoption-check`** — a daily CronJob (07:40 UTC),
`deployments/kustomize/services/instance-token-adoption-check/`. It counts active components whose
`html_template` references `{{.InstanceID}}`; **0 means this exception still holds, non-zero means
it has expired and this RFC is owed a round.** One `doc_notes` row per run
(`subject_key='instance-token-adoption'`) including a quiet one, so a *missing* row means the job
did not run — which is not the same as "the exception still holds".

**Why a CronJob and not the pattern-check finding the seat suggested.** The instinct was right and
that example cannot work: `scripts/pattern-check.py` is a commit-time lint over **repo files**, and
a component's `html_template` is written by the component-creator agent, by hand-authored SQL, by
migrations and by the admin UI — four routes, none of which passes through a commit.

**⚠ Its healthy answer is ZERO, so it carries a demand control.** A broken query, a mis-escaped
`LIKE` or an empty table all return zero too. Every run therefore also counts `{{.ComponentID}}`
through the same `LIKE` in the same statement, and **refuses (exit 2)** if that comes back 0 rather
than reporting a reassuring zero it has not earned. First live run, 2026-08-16 15:29 UTC:
`adopters 0, control 5, active 243` — the control fires, so the zero is evidence.

**⚠ Retire the job once it trips.** A tripwire left failing daily is how a real signal becomes one
people mute. A trip is **not** a defect report — converting templates is 283's intended next phase.

**Still NOT built: the unification itself.** Nothing in 283 moves `ComponentID`. That is §4's
question and it is what this RFC is actually for.

## 7. Sources

- `bugs_open/283` (§4 the two-path measurement; §9 the round-2 record; §10 the approved verdict)
- Council correlation `07635a2f-3605-4e67-9a6d-7636b07f16ca` — round 1 `architecture` seat
  (`ARCHITECTURE_SIGNAL: insufficient`, the "three ad hoc derivations" note), round 2 `reuse_agent`
  seat (medium, "no locator") and `architecture` seat (medium, "recommend a written trigger")
- Register **CLC-014** (the `InstanceID` seam); `LANDMINES.md` § "`{{.ComponentID}}` is the estate's
  per-instance id convention on ONE render path"
- Owner ruling 2026-07-29 §1 (an addition to a shared vocabulary needs an RFC when it changes what
  the shared mechanism GUARANTEES); RFC_022's narrowing (the exception 283 was approved under)

---

## 8. RULED 2026-08-22 (owner) — CONVERGE ON `{{.InstanceID}}`; the evidence pack §4 asked for

Raised again the same day `bugs_open/283` closed (`9223c421d`, 16:09), which lists this RFC as a
deliberate residual. §4's three options were put to the owner with the measurements below.

**The ruling: converge, do not re-point.** Leave `ComponentID`'s meaning alone; convert the five
templates that reference it to the already-approved, already-live per-instance seam, through the
framework's own conversion pipeline (RFC_034's route — work items writing `content_components`,
never hand-applied template SQL); then retire the placeholder by deleting its writers. **One
identity, not two.** Options "re-point `ComponentID`" and "leave the ambiguity documented" were
declined.

Why this is NOT architecture-scope under the 2026-07-29 §1 ruling, stated so a reviewer can
check it rather than take it on trust: `{{.InstanceID}}`'s guarantee is unchanged and already
council-approved (CLC-014); the conversion route is the one RFC_034 ruled on and under which 124
rows have already converted, so the five are late arrivals to an existing programme, not a new
guarantee. `ComponentID`'s guarantee is never altered while a consumer exists — its writers are
deleted only *after* the consumer count is measured zero, which is guarantee-neutral by
construction.

### 8.1 §4 sub-question 1 — "what breaks?" — measured, with the absences named

**Nothing in this estate names a section wrapper id by value.** All measured 2026-08-22:

- **0** occurrences of `href="#{{.ComponentID}}"` anywhere; **0** `href="#"` fragments in the live
  component library (its only fragment href is `/services.html#{{this.slug}}`, from a slug).
- **0** UUID-shaped or `component_`-shaped `#id` selectors among the **206** distinct `#id`
  selectors in repo-side acceptance criteria — sections are addressed by class plus
  `data-component`, never by id.
- The LMC arithmetic oracle's ~170 checks all address `#c-<function>-<inner-id>`, the
  already-prefixed inner controls. It never names a wrapper.
- No sitemap, feed, JSON-LD or `cmd/` tool carries a section id; no CSS rule keys on one (all five
  templates style by class). No skip-link or table-of-contents generator exists.
- Inbound external deep links are unobservable from here; the prior is very low, since the id is
  an opaque internal UUID appearing in no anchor, sitemap or feed we publish.

### 8.2 §2's TABLE IS WRONG, and this is the correction it owes

> **CORRECTED 2026-08-22.** §2 lists `assemble_from_library.go` as resolving `ComponentID`
> per-instance and therefore "reuse-safe". **It has never done so on the healthy path.**
> `RenderTemplate` executes first (`assemble_from_library.go:303`); `missingkey=zero`
> (`call_agent.go:1172`) resolves the placeholder to `<no value>`, which
> `component_library.go:1170` strips to `""`. The post-render
> `strings.ReplaceAll(renderedHTML, "{{.ComponentID}}", componentID)` at `:309` therefore never
> matches anything, and `component_<fn>_<idx>` survives only as a `contentRequirements` map key
> and a log field. **Measured: 0 of 270 live placements carry that shape** — the regex was proved
> against a synthetic positive (1 match) before the 0 was believed, then widened
> case-insensitively. So the "one of three paths is safe" framing that shaped this RFC's own
> question was never true; all live paths were unsafe.

### 8.3 A fourth and fifth producer §2 does not list

- **The section editor binds nothing.** `applyContentEdit` (`section_editor_actions.go:1113`) and
  `applyComponentSwap` (`:1249`) build context via `buildRenderContextFromDB`, which never writes
  a `ComponentID` key — so the template renders `<section id="">`. **11 live placements** serve
  that today, and both routes write `page_components.rendered_html` straight to an already-live
  page with no downstream gate.
- **`page_components.content_data`** carries a `ComponentID` key on **10** rows; in **9** of them
  it simply equals `slot_name`. These are inert once no template reads the key.

### 8.4 An attractive alternative, measured and REFUTED

"Bind `ComponentID` from `slot_name`" looks free — slot name is per-placement, readable, stable,
already in `pages.sections`. **It does not work.** `slot_name` is not unique within a page: 1,940
page/slot pairs, 1,911 distinct, **20 pages repeat a slot name** (never NULL, never empty).
Crossed against the duplicate-id pages, **15 of 18 overlap**, so slot-derived ids would fix **3 of
18**. This independently vindicates the `reuse_agent` seat's 2026-08-16 rejection of
`InstanceTokenFromSlot`.

### 8.5 The live cost, and the residual this unblocks

**18 pages carry a repeated `{{.ComponentID}}` component, 27 redundant placements (as of
2026-08-22)** — 13 pages ×2, three ×3, one ×4, one ×6, all `generic-text-block`. Live-verified at
the artefact: `apis.uk/index.html` (HTTP 200) serves **six** `<section
id="8d81e665-3ee0-443d-a873-690268c15fbb">`, re-confirmed cache-busted after 283 closed.
Single-instance pages were read as a control and show 1 id, 0 duplicates — the check discriminates.

`component_instance_scope.go:~268` names, as the **first** of two reasons `enforce_instance_scope`
ships default-OFF, exactly these pages: *"defaulting to refuse would fail their next re-render."*
Converging removes that reason, so **arming the rerender path becomes a config migration rather
than new code**. The second reason — the detector is a regex that errs toward reporting — still
stands, so this removes one of two, not both.

### 8.6 Feasibility finding the implementer must not skip

The existing converter **cannot** convert these five. `ConvertTemplateToInstanceScope`
(`component_instance_conversion.go:89`) harvests ids with `\sid="([^"{}]+)"` — the class excludes
braces, so `id="{{.ComponentID}}"` never matches and it refuses with *"template declares no
literal element ids"*. Filing the five as conversion items against today's binary would produce
five polite no-op completions. What IS reusable unchanged is `GateConvertedTemplate`, which
requires `{{.InstanceID}}` present and renders the **doubled** template through the real renderer.

A latent half-state in the same regex, worth closing while there: a template mixing literal ids
**and** `{{.ComponentID}}` would convert "successfully" with the templated id silently ignored —
and the gate could not catch the residue, because `reElementID` (`component_instance_scope.go:215`)
requires one or more non-brace characters and so cannot see duplicate `id=""`. **Measured
2026-08-22: 0 active templates are in that mixed state** (control: 87 active templates carry a
literal id at all), so it is latent rather than live — but it is the shape a future
component-creator generation could produce.

### 8.7 Provenance

Not run through the `090` diagnosis loop; per CLAUDE.md's 2026-07-31 ruling, what was substituted:
both code paths read end to end; the consequence measured at the artefact with controls that could
have come out otherwise; the same functions read independently by a second reader, which reached
the same conclusion and additionally found that `RenderTemplate` already logs
`Warn "fields rendered empty" fields=[ComponentID]` (`component_library.go:1244`) while both
section-editor call sites discard the report; and four affected pages plus two controls fetched
live.

## 9. STEP 2 DONE, AND IT EXPOSED THE REAL SHAPE OF STEP 3 (2026-08-23, `bugs_open/283` lane)

§8's ruling is executed as far as the corpus goes, and the execution turned one of this RFC's
deferred worries into a **measured, reproduced defect with a live victim**. Recording it here
because it changes what the remaining work IS, not merely when it happens.

### 9a. What is done

Pass 0 shipped (`67d34e6c1`, council `cd6a5ef6` r1+r2 APPROVED, both verdicts read) and is LIVE
on chassis **v1.0.1328** — binary-probed with controls, not inferred from the tag. Four of the
five templates are converted through the fixer: `generic-text-block` (179 placements / 152 pages
/ 21 sites), `faq` (82/82/15), `mechanism-flow` (6/6/3), `evidence-timeseries` (3/3/3), all
**as of 2026-08-23**. The conversion self-propagates: the fixer filed **219 `page_rerender`
items**, and those produce correct, distinct per-instance ids (`c-generic-text-block`,
`-2`, `-3`, `-4` — verified in stored HTML on webdesign.co.uk/domains and
gaswholesalers.com/service-areas).

> **⚠ CORRECTED 2026-08-25 — §9a's "converted" COUNTS THE TEMPLATE, AND A READER WILL TAKE IT AS
> THE PLACEMENTS. Re-derived from `page_components`: 48 of 437 placements carrying a token-bearing
> template still have NO `c-` token in stored `rendered_html`, 26 of them locked (as of
> 2026-08-25).** Per function, the unconverted counts are `generic-text-block` **26 of 188**,
> `evidence-timeseries` **3 of 3**, `tool-loan-repayment` **2 of 2**, `faq` 1 of 88,
> `mechanism-flow` 1 of 6, plus 14 tool functions with one each.
>
> So the line above — *"`evidence-timeseries` (3/3/3)"* — is **exactly wrong for the placements**:
> all three still serve their pre-conversion literal ids (`evidence-timeseries-leakage`, `-ifr`,
> `-pdc-calendar`; the oufe row last written **2026-07-29**, i.e. never touched by the 08-23 batch).
> Confirmed at the artefact: `https://oufe.com/cases/thames-water.html` serves
> `id="evidence-timeseries-leakage"`.
>
> **THE MEASUREMENT LESSON, which is why this is a correction and not just a number.** Found and
> reported by the `news_editorial_features` lane, 2026-08-25, and it is a blind spot no count of
> ours could have seen: **a locked instance that receives NO delivery attempt produces NO SIGNAL
> AT ALL, and is therefore indistinguishable from an instance that needed no delivery.** Two of
> the three `evidence-timeseries` instances are visible *because* delivery was attempted and the
> lock gate refused, filing `lock_blocked_change`. The third got a whole-page `page_rerender` at
> 12:32:25Z — **before** the template conversion at 12:33:33Z — and nothing after, so no gate
> fired, no row was filed, and its owning lane was never told. A coverage measure keyed on
> `lock_blocked_change` reads that as **2 blocked / 1 fine** when the truth is **3 unconverted**.
> Re-derive coverage from `page_components` against the live template, never from the refusals.
>
> **AND WHAT THE 48 ACTUALLY COST, measured the same day — because "48 unconverted" will be
> quoted as damage and it is not.** A literal element id is only a defect where the same component
> appears **twice on one page**; anywhere else it is untidy and inert. Narrowing it:
> **48 unconverted → 8 sitting on a multi-instance (page, function) pair → 1 page carrying an
> actually duplicated element id → 0 reaching a visitor.**
>
> That last page is `webdesign.uk/index.html`, whose two `generic-text-block` rows (written
> **2026-08-05**) both carry `id="8d81e665-3ee0-443d-a873-690268c15fbb"` — a **UUID**, i.e. the
> retired `{{.ComponentID}}` binding frozen in stored bytes that never re-rendered. It serves to
> nobody: **`webdesign.uk` 302-redirects to `webdesign.co.uk`**, which is a **separate `sites`
> row** with its own `page_components`; the redirect target serves four unrelated ids and **zero**
> occurrences of that UUID. Controls run in the same breath: an invented path on `webdesign.uk`
> also 302s (so the domain serves nothing of its own — the parked-domain trap in its redirect
> variant), and the UUID grep against the followed target returns 0.
>
> **So the residual is a consistency debt, not live damage**, and it should be prioritised as
> such. The honest one-line version: *48 placements still carry pre-conversion ids as of
> 2026-08-25; none of them collides on a page a visitor can reach.* ⚠ That is a **census, and it
> goes stale by ADDITION** — re-run the query above before quoting either number, and re-check the
> redirect, because a domain that starts serving its own content converts this from debt to damage
> with no other change.

> **A positive control worth keeping, from the same lane:** across the 08-23 batch, **253**
> instances re-rendered after the conversion (`generic-text-block` 161, `faq` 87, `mechanism-flow`
> 5) and **ZERO** produced an empty id — real demand behind the zero, which is what makes it
> evidence rather than silence.

`pricing` (row `6175e049`) is NOT converted: active, same placeholder, **zero placements**, and
`site_work_items.site_id` is NOT NULL with the site only reachable through a placement, so there
is no honest site to file it against. **It is a precondition of §8's second half** — retire the
`{{.ComponentID}}` bindings while that row still spells the placeholder and its first placement
renders `id=""`.

### 9b. The finding: a FULL PAGE BUILD re-collides every multi-instance page, deterministically

`BindSingleSectionInstanceToken` supplies **occurrence 0** to the two paths that render one
section at a time (`RenderComponentAction` at `v3_site_actions.go:2404`, the section editor at
`section_editor_actions.go:1104,1275`). Its doc comment states the licence plainly: right
"whenever the component appears once on the page, which is every interactive component on every
live page today (measured 2026-08-15)". **That measurement was about `getElementById`
components. It has never been true of the templates in THIS RFC** — they are the ones that
repeat, up to ×6.

Reproduced 2026-08-23 on `apis.uk/index.html`, six instances of `generic-text-block`:

1. A page rerender hit a **pre-existing** 7-of-8 plan shortfall, so `UpdatePageStatusAction`
   refused the deployed stamp and set `build_status='needs_rebuild'` (`v3_site_actions.go:976`).
2. A `build-pipeline-trigger` picked that up 6 minutes later and rebuilt the page.
3. The rebuild rendered each section individually → occurrence 0 for **all six** → the stored
   HTML now carries `<section id="c-generic-text-block">` six times, and the page redeployed
   with it.

**Severity, stated precisely so it is neither over- nor under-sold: this is NOT a regression.**
Before conversion the same page served the component ROW ID six times; after a rebuild it serves
occurrence 0 six times. Identical collision count, different string. **What it means is that the
conversion does not STICK**: any page that rebuilds returns to a colliding state, so §8's ruling
cannot be completed by converting templates alone. The page-rerender path is correct; the
build path is not.

It is also **worse than a silent risk in one specific way**: `c-generic-text-block` *looks*
converted. Any check that greps for the `c-` prefix, or that asks whether the template carries
`{{.InstanceID}}`, reports success on a page serving six identical ids.

### 9c. What step 3 therefore is

Not "unify the builders" in the abstract — **derive the real occurrence where the page's rows
exist, and fall back to 0 only where they genuinely do not** (a build in progress, before
`page_components` is written). `InstanceCounter` already is the canonical rule; the
single-section paths need its INPUT, not a second rule. The honest constraint is that
`RenderComponentAction` renders during a build when the rows may not exist yet, so the fallback
must stay — which means the fix is a lookup plus a documented fallback, and the fallback's
remaining blind spot should be stated rather than assumed away.

Whoever takes it: `apis.uk/index.html` is the standing repro, and it holds `needs_rebuild` from
an unrelated 7-of-8 shortfall, so it rebuilds and re-collides on its own — a free test case that
regenerates itself.

### 9d. CORRECTED SAME DAY — the trigger is not "a rebuild", it is ANY per-section render, and it is firing in production now

§9b named the trigger as a full page build, because that is the path I happened to reproduce.
**That was too narrow, and the narrower version is the more comfortable one, which is exactly why
it should be distrusted.** Measured a few hours later, at the end of the conversion:

Of the 12 pages that were serving duplicate ids this morning, **9 are fixed and 3 are colliding
again** — and all three carry the same signature, two sections both stamped the occurrence-0
token `c-generic-text-block`. The 9 fixed pages had their sections written at 12:50–13:24 by the
page-rerender queue (which uses `InstanceCounter` and gets it right). The 3 colliding pages were
rewritten at **17:41–17:51**, hours after that queue drained, by:

```
item_type        handler_agent        created_by     completed_at
content_rewrite  page-build-handler   backfill-353   2026-08-23 17:41:50
content_rewrite  page-build-handler   backfill-353   2026-08-23 17:44:02
```

**An unrelated lane's content backfill.** Not a rebuild I triggered, not a section edit, not
anything to do with this work. It renders sections one at a time, so it takes the occurrence-0
path, and it silently re-collided pages that had been correct for four hours. All three pages are
`rebuild_policy='generic'`, so this is not the owned-page `section_edit` delivery route either.

**So the correct statement of the defect is: ANY path that renders a section on its own
re-collides every multi-instance page it touches.** Page builds, content rewrites, section edits,
and whatever is added next. These are ordinary, high-volume operations, which means:

- The conversion **does not hold**. A page fixed today is fixed until the next content operation
  touches it, and nothing warns anyone when it flips back.
- **Counting converted templates, or converted stored HTML, measures the wrong thing.** All four
  templates are converted and stay converted; 244 of 275 placements carry a `c-` prefix. Both
  numbers stay healthy while pages re-collide, because `c-generic-text-block` twice *is* a `c-`
  prefix twice. The only measurement that sees this is DISTINCTNESS per page, at the artefact.
- The fix in §9c is therefore not a tidy-up to schedule; it is what makes the ruling's first half
  durable. Until it lands, RFC_032 §8 is achieved at the corpus and only intermittently at the
  page.

Standing repro, no setup required: re-fetch `gaswholesalers.com/pricing-transparency.html` and
`vetcomparison.uk/how-it-works.html` and count distinct `<section id>`.

## 10. OWNER RULED 2026-08-24: BUILD §9c NOW, with the detector/empty-id fix IN THE SAME CHANGE — initial plan exists

Decisions of 2026-08-24: (1) the occurrence-derivation fix is built now, not deferred; (2) the
detector's `id=""` blindness and the render-time fail-loud (bug_historian's HIGH on council
`e8c7414c`) go into the same change; (3) the four idea.uk empty-id pages are repaired on their
own lane (CONTRIB filed in `idea_uk_section_data_missing/`, 2026-08-24) — even at the cost of
content generation, though their intact `content_data` suggests a plain rerender suffices.

**The initial plan is at
`docs024_key_docs_latest/bugfix_283_component_instance_scope/PLAN_2026-08-24_occurrence_derivation_and_empty_id_detector.md`**
(authored by a Fable 5 Plan agent at the owner's direction; its load-bearing measurements
independently re-verified; to be built up by a dedicated thread). Headlines: occurrence counted
from `page_components` under the canonical walk's exact key equality, position-exact on the
editor path and slot+rank on the build path (16 of 30 multi-instance pairs repeat a slot_name,
killing slot-only keys); constant-0 is the universal fallback so no branch is ever worse than
today; `id=""` becomes its own detector class rather than a widened `reElementID`; the
render-time refusal arms an Error log that already exists at `component_library.go:1103`,
conditional on a measured-zero log census per the 2026-08-02 §2 ruling.

New evidence folded in: the formal empty-id census is **6 rows on 6 pages across 2 sites (as of
2026-08-24)** — the 4 idea.uk pages (retired-placeholder cause) plus 2 dartsonline pages whose
cause is a THIRD shape: `category-listing` declares `id="{{.category_slug}}"`, a content field,
which rendered empty. Any unbound field in an id attribute produces this class; the fail-loud
half is general for exactly that reason.

### 10a. The INTERIM EXPOSURE WINDOW, recorded formally (2026-08-24, per council `e8c7414c` round 3)

Round 3 on the retirement correlation returned REVISE (the second), gating seat `bug_historian`,
whose position is accepted rather than argued with: **a plan document is not code**, and the
round's own census proved the silent-empty-id mechanism is broader than the retired placeholder
(the `category_slug` third cause). The seat named two acceptable exits: ship a minimal guard
with the retirement, or *"explicitly accept and record the interim exposure window as a known,
monitored risk with an owner and expiry — not just a risk-section sentence."* This section is
that record:

- **The exposure**: until the §10 build ships, any template field used as an id attribute's
  whole value that resolves unbound renders `id=""` silently; `DetectInstanceCollisions` cannot
  see it (`reElementID` requires a non-empty id) and no render path refuses it. The class is
  general — three distinct causes are already on record (retired placeholder ×4 pages,
  section-editor-era unbound renders, `category_slug` ×2 pages).
- **Known damage, bounded by measurement**: 6 stored rows / 6 pages / 2 sites (formal census
  2026-08-24, query in the round-3 submission's grounded_in). The idea.uk 4 are owned by their
  lane (owner ruling, CONTRIB 2026-08-24); the dartsonline 2 wait on the build's detector, which
  will surface them at their next rerender.
- **OWNER of the window**: the §10 building thread (its starting document is the committed
  PLAN). **EXPIRY**: the build's image roll plus its config activation — the plan's rollout
  steps 3–4. The plan's Open Question 6 already requires re-running the empty-id census on
  build day, which is the check that the window did not widen while open.
- **What is deliberately NOT happening meanwhile**: no round 4 on `e8c7414c`. Two REVISE
  verdicts have said plan-is-not-code; the next submission on that correlation is the build
  itself, which closes the gating objection with edits rather than framing. The correlation's
  arc (REJECTED → REVISE → REVISE) is honest history, not a failure to be tidied.
