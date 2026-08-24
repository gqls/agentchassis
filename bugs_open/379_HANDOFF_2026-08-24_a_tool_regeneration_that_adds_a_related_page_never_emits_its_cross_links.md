# 379 — a tool REGENERATION that adds a related page never emits cross-links for it: the `replace_existing` arm returns before the emitter, so only a tool's FIRST birth can ever produce cross-links

Filed 2026-08-24 by the staged-component-build lane, **at the direction of `bugs_open/353`'s round-3
council** (`53e3812f`, `bug_historian`, medium): the residual was disclosed only in a Go comment and
in 353's own open list, where it would have closed silently when 353 closed.

**One sentence:** `create_tool_component` reroutes a `replace_existing` request into
`regenerateToolComponentInPlace`, which returns **before** `emitToolCrossLinkItems` is ever reached —
so if a regeneration's spec ADDS a related page that the original birth did not name, no cross-link
is ever emitted for it, by any path.

## 1. Why this is NOT a duplicate of 353

353 is *"cross-links are withheld AT BIRTH and nothing re-emits"* — a guard starving on a gate item
that fix 177 stopped raising. **That is fixed, approved and live** (353 §12.1/§13).

This is a **different defect on the sibling call site**: the regeneration path does not reach the
guard at all. 353's fix cannot help it, because 353's fix is *inside* a function this path never
calls. Filing separately so it is not read as a reopening, and so it does not inherit 353's closure.

## 2. Mechanism

`create_tool_component_action.go` — the `replace_existing` arm returns
`regenerateToolComponentInPlace(...)` before reaching the emitter call (~:559). Checked 2026-08-23:
`create_tool_component_regenerate.go` calls **neither** `emitToolCrossLinkItems` **nor**
`related_pages`.

⚠ **[UNVERIFIED BY AN INDEPENDENT READER]** — the round-3 `prior_art_librarian` seat flagged that a
content search cannot settle this (the code index is declarations-only and its string-literal search
is unreliable), and recommended a human grep the file body directly. The grep above is this lane's
own and was not independently repeated. **Do that first.**

## 3. Blast radius — why it is bounded, and why it is still real

**Bounded:** a regeneration whose related-page set is unchanged loses nothing — the cross-links from
the original birth already exist and are keyed on `tool_crosslink:<function>:<page>:<site>`, which is
idempotent. The damage is confined to the DIFFERENCE: pages named by the regenerated spec that the
original birth did not name.

**Still real:** regenerations are most of the traffic on established tools (353 §3), and a tool's
related-page set is exactly the kind of thing an improvement pass would widen.

**[UNMEASURED] and it is the first task here:** nobody has counted how many regenerations changed
their `related_pages` set. Until that number exists this is a real defect of unknown size — do not
size it by intuition in either direction.

## 4. Fix candidates, ordered by what closes the door

1. **Emit on liveness rather than at any birth** — 353 §6 candidate 1, unchanged and still the
   strongest: cross-links exist iff the page is served, so no call site can miss them. Fixes this
   bug and 353's class together, and makes the withheld state unrepresentable.
2. **Call the emitter from the regeneration path with the opt-in left FALSE.** Cheap and local. The
   default is the safe side (353's 2026-08-02 shared-seam ruling), so a regeneration on an unbuilt
   page still withholds. Does not fix the class.
3. Do nothing and document. **Rejected** — that is what produced this file.

## 5. Verification when fixed

A regeneration whose spec names a page absent from the original birth produces exactly one new
`tool_crosslink:<function>:<newpage>:<site>` item, and re-running it produces no second one.

## 6. Ownership

**UNOWNED.** `bugs_open/353`'s lane (staged-component-build) filed it and is **not** claiming it —
353's own remaining items are (b) and (c′) and neither touches this path. Run `who-owns.py 379`
before starting, announce the claim in this file, and grep 353 §13.3 for the context.

**Adjacent, do not conflate:** `bugs_open/362` also touches `replace_existing` but is about tool
writers persisting `rendered_html` without link repair — a different defect on a nearby path.


## ADDENDUM 2026-08-24 (evening) — the related-pages PICKER does not help this bug, and here is why

By the `staged_component_build` lane. Migration 602 (council `c962abd1`, APPROVED r1) makes both tool
workflows ASK an LLM for `related_pages` when the request names none, and commit `0fb94a7dd` teaches
the emitter to accept that answer as a fallback. **None of it reaches this bug.**

The picker feeds `related_pages_fallback` into `create_tool_component` / `deploy_tool_to_site`, and
both of those reach the emitter. A **regeneration** does not: `replace_existing` returns from
`regenerateToolComponentInPlace` roughly 270 lines before `emitToolCrossLinkItems` is ever called
(`LANDMINES.md`, the `no_related_pages` entry). So a regeneration whose spec ADDS a related page still
emits nothing, and now it will do so on a site where every OTHER tool birth is getting cross-mentions
— which makes the gap harder to notice, not easier.

**This was raised on the record.** The council's `bug_historian` seat objected (advisory, medium) that
the change patches 2 of 3 producers of the same shared mechanism and names the third here rather than
tracking it: *"Disclosed transparently, but disclosure isn't a fix."* That is correct. This addendum is
the tracking it asked for.

**Still UNOWNED.** Its size is `[UNMEASURED]` on purpose — the first task is counting how many
regenerations changed their `related_pages` set, not guessing. Whoever picks it up: the picker
mechanism is now available to reuse (`suggest_related_pages` → `related_pages_fallback?`), so the
regeneration path could gain the same wire cheaply IF someone first decides whether a regeneration
should re-emit at all.
