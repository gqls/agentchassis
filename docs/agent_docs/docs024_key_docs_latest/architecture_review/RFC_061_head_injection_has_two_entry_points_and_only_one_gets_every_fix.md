# RFC_061 — head injection has two entry points, and for the fourth time only one got the fix

**Status: CANDIDATE, nothing built, nothing decided.** Filed 2026-09-02 by the
`improvement_loop` lane at the request of the council gate, which raised it on corr
`3c71ec77-fd15-4aa1-a762-cc36116caca5` (skip link) across **four seats independently** —
`guardian` (medium: *"no follow-up work item or tracking is created for the uncovered
path … same failure mode this council has already seen"*), `bug_historian` (medium),
`reuse_agent` (*"a known second-path duplication-of-concern for follow-up"*) and
`architecture`, which named the shape precisely and declined to fold it into that patch:

> *"the same duplicated-injector-family pattern (assemblePage vs AssemblePageAction)
> recurring for a fourth time is a mild signal that head-injection logic wants
> consolidating behind one entry point eventually — but that's a future RFC candidate, not
> this patch's job, and the author is right not to fold it in here."*

This file exists so the fourth recurrence is tracked rather than re-discovered on the
fifth. **It is a candidate, not a proposal**: I have deliberately NOT established the
claim that would make it a bug, and §3 says exactly where the evidence stops.

Every figure carries the date it was counted.

---

## 1. The shape

`assemblePage` (`rerender_single_page_action.go`) runs a chain of head injectors before it
writes the document. As of 2026-09-02 the chain is five long:

| injector | added for | reaches `AssemblePageAction`? |
|---|---|---|
| `injectPageJSONLD` | zero structured data fleet-wide (2026-07-28) | no |
| `injectCanonicalLink` | zero canonicals fleet-wide (2026-08-02) | no |
| `injectRobotsNoindex` | `bugs_open/232` | no — **stated in a LANDMINE** |
| `injectComponentCSS` | `bugs_open/072` | no |
| `injectSkipLinkCSS` | this lane, 2026-09-02 | no |

`AssemblePageAction` (`multipage_actions.go`) is a second page-building path, live in the
`page-content-writer` agent's config as `compile_page_sections` (`[MEASURED 2026-09-02]` 1
active agent definition references it). It builds a document through `buildPageHTML`
(`v3_site_actions.go:2948`), which emits a minimal head and **no `<main>` at all**, then
optionally injects head/header/footer from the component library.

Each of the five injectors was added by a different lane, each correctly scoped its own
fix, and each documented the gap honestly. **Nobody has been wrong.** The cost is that a
reader of any one injector cannot tell whether the page in front of them has it, and the
answer is now five independent facts rather than one.

## 2. Why it is worth a round rather than a sixth honest risk note

The failure mode is not that the second path is broken. It is that **"the platform emits
X on every page" is no longer answerable by reading the code that emits X** — it depends
on which of two builders last touched the page, which is not recorded on the page. Every
future head-level guarantee inherits that, and the estate keeps adding them: three of the
five arrived in the last five weeks.

The `architecture` seat's own trigger test (owner ruling 2026-07-29 §1) asks whether a
change alters what a shared mechanism GUARANTEES. Individually, none of the five did —
each is opt-in-ish, additive and inert on the path it does not reach. **Collectively they
have changed what "the page shell" means, and no single change was the one that did it.**
That is the accumulation case RFC_022 was written about, arriving in a different
mechanism.

## 3. WHERE MY EVIDENCE STOPS — read this before quoting the file

`[MEASURED 2026-09-02]` What I established:

- The five injectors exist and are called only from `assemblePage`.
- `assemblePage` has exactly 2 callers (`rerender_single_page_action.go:163`,
  `section_editor_actions.go:655`).
- `compile_page_sections` is named by 1 live agent definition (`page-content-writer`).
- Of 978 open `head_essentials_missing` findings, **968 carry `spec.assembled = true`**
  (the page has `page_components` rows) and **10 do not**, on 5 sites.
- Sampling 14 served pages across 12 domains, 13 carry exactly one `<main>`; the one that
  does not is `advertise.co.uk`'s hand-built index.

`[NOT ESTABLISHED]` — and this is the gap that keeps this a candidate:

- **That `AssemblePageAction` currently serves any live page.** The 10 non-assembled rows
  are *consistent* with it and equally consistent with hand-built or adopted pages. I did
  not trace a single served page back to that builder. **A census joining each live page
  to the builder that last wrote it does not exist**, and building one is the first task
  of whoever takes this RFC — not because it is the interesting part, but because the
  whole argument above is worth nothing if the second path is dead.
- Whether consolidation is even the right remedy. "One entry point" is the seat's phrase
  and my instinct too, but the alternative — a recorded `built_by` on the page, making
  the question answerable without unifying anything — is cheaper and was not costed.

## 4. What a round would decide

1. Is `AssemblePageAction` live and serving? (Census first; if no, this file closes and
   the five gaps close with it.)
2. If live: consolidate behind one entry point, or record the builder per page and let the
   two diverge honestly?
3. Either way: does a NEW head-level guarantee need a stated position on both paths before
   it ships — i.e. is this a checklist item on the council gate, or a build-time check?

## 5. Relations

- Council corr `3c71ec77-fd15-4aa1-a762-cc36116caca5` (the round that asked for this).
- `LANDMINES.md`, the `injectRobotsNoindex` entry naming the same split.
- `RFC_022` (accumulation of individually-inert additions; same argument, different seam).
- `docs024_key_docs_latest/improvement_loop/` — the lane that hit it fifth.
