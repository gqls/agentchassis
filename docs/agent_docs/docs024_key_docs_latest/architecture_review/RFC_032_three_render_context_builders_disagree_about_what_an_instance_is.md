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
