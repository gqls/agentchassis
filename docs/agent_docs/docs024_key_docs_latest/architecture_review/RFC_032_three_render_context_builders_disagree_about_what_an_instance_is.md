# RFC_032 — Three render context-builders disagree about what an "instance" is, and `ComponentID` means two things

**Status:** OPEN, filed 2026-08-16. **Raised by the council gate twice**, on the same lane, in two
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
| `assemble_from_library.go` | `component_<function>_<idx>` — built from the loop index | **yes** |
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

## 6. What is NOT built, stated so it is not mistaken for done

- **The trigger.** `scripts/pattern-check.py` is a commit-time lint over repo files; component
  templates live in `content_components.html_template` in the database, so a pattern-check cannot
  see one. The trigger needs a DB-side check (the RFC_006 shape — a daily CronJob query), and the
  query itself is one line:
  ```sql
  SELECT count(*) FROM content_components WHERE is_active AND html_template LIKE '%{{.InstanceID}}%';
  ```
  Whoever converts the first template should either build that check or accept that the RFC_022
  exception has silently expired.
- **The unification itself.** Nothing in 283 moves `ComponentID`.

## 7. Sources

- `bugs_open/283` (§4 the two-path measurement; §9 the round-2 record; §10 the approved verdict)
- Council correlation `07635a2f-3605-4e67-9a6d-7636b07f16ca` — round 1 `architecture` seat
  (`ARCHITECTURE_SIGNAL: insufficient`, the "three ad hoc derivations" note), round 2 `reuse_agent`
  seat (medium, "no locator") and `architecture` seat (medium, "recommend a written trigger")
- Register **CLC-014** (the `InstanceID` seam); `LANDMINES.md` § "`{{.ComponentID}}` is the estate's
  per-instance id convention on ONE render path"
- Owner ruling 2026-07-29 §1 (an addition to a shared vocabulary needs an RFC when it changes what
  the shared mechanism GUARANTEES); RFC_022's narrowing (the exception 283 was approved under)
