# CONTRIB 2026-08-20 — from the `bugfix_305_negation_gate` lane: the writer-side gate is being built here, and here is exactly what it will and will NOT guarantee

**Who this is from.** A session picking up `bugs_open/305` (the bug you filed and diagnosed) with the
owner's instruction to fix it robustly at the framework level. **Your diagnosis is the input, not
something I am re-litigating** — the 08-19 evening correction block in `bugs_open/305` is treated as
authoritative, including the part where the saturation counts were withdrawn and the finding narrowed
to *"the brief hands the writer a canonical tagline built on the construction, and supplied phrases
transfer."*

**Why I am not stepping on you.** `scripts/who-owns.py 305` names your lane. I read
`docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/HANDOFF_2026-08-19_continue_here.md`
before starting: your "Next work" list has no 305 fix in flight, item 6 parks the brief detector's
scheduling as an owner/architecture call, and item 3 leaves the briefs to the owning site lanes. So
the writer-side half was open. Everything below goes into `bugs_open/305` too — this is a
contribution into your bug, not a second account of it.

## What is being built (two halves, both platform code, council-submitted)

### 1. A mechanical gate between the writer and the render, per section, with a per-PAGE budget

- `platform/orchestration/datahelpers/negationtells.go` (new, pure): `ScanDefineByNegation` covering
  the **five** shapes — `x_not_y`, `not_x_but_y`, `staccato`, `rather_than`, `negative_reveal`. Your
  `count_negation_tells.py` patterns are the starting point; the Go side had only `not X, but Y`
  (`voicetells.go:151`), which is **1.5%** of sections and cannot match either sentence the owner
  quoted.
- A new action `rewrite_negations`, wired by migration between `generate_content` and
  `render_section` in `page-content-writer`'s `process_sections_loop`. It asks for **sentence
  replacements** (`{"replacements":[{field,from,to}]}`) and splices them in Go — not a re-generation
  of the section.
- Budget: the first **two** non-exempt hits on a page pass untouched (the house voice's own "once or
  twice per page"), every later one is rewritten, and any hit in a headline-class field is rewritten
  regardless. The count rides in `CollectedData` because loop iterations share one workflow state.
- `render_component` gets a **default-on annotation** (`copy_gate_findings`, no behaviour change), so
  the count exists for every LLM-authored section from every writer even where the fixer is not
  wired. That is the "a gate a workflow author can forget to wire is a comment" property
  (`save_page_meta_description_action.go:44-48`).
- Your `strawman` arm and the meta-description gate inherit the missing shapes for free.
  `rather_than` enters `ScanVoice` only as a **density** finding (>2/page), never per-hit — per-hit
  would fire on 43% of sections and drown `check_voice_tells`.

### 2. The brief-side detector, scheduled — the owner has now called item 6

`cmd/brief-negation-check` + a daily CronJob on the `verifier-remit-check`/WFA-013 shape: derives the
writer-visible surface at runtime from the live agent config, measures **only** that surface (never
`data::text` — your withdrawn census is cited in the code comment as the reason), separates the
**supplied** class from instructional prose, and writes one `doc_notes` row per run **including a
clean one**. Your `audit_writer_brief.py` is the specification for it and is credited as such; the
Python stays as the human-run tool.

## THE GUARANTEE, stated precisely — and it is narrower than "never again"

> **The five named forms, beyond two per page and never in a headline-class field, do not leave
> `page-content-writer`. Brief-supplied sentences and regulatory negations are EXEMPT — counted, not
> rewritten.**

It is deliberately **not** "the instinct never leaves the framework". Your
`fleet_copy_quality/SUMMARY_2026-08-08_why_this_is_hard_and_what_we_have_learned_about_rules.md` says
that is unreachable by rule, and I believe it: what a gate can hold is a **form**.

**Three things I got wrong first, that your own documents caught before any code existed** — worth
having on record because two of them are traps for anyone who builds in this area:

1. **"Re-ask for the whole section and keep whichever attempt scores lower" adopts displacement.**
   Measured in the same 1,503-call corpus: "instead of" 5.9%, "isn't just/a" 6.4%, "more than (just)"
   10.8%, em dash 0.5%. A rewrite to *"X instead of Y"* scores **zero** on the five shapes and wins
   the comparison while being the same instinct — your 08-12 CONTRIB's *"a prohibition displaces a
   problem rather than solving it"*, arriving from the measurement side. So each replacement is
   accepted **individually** and only if it introduces no hit in the five shapes AND none in a
   neighbour set, and preserves every digit, URL and proper noun. **Rejections are logged with a
   reason** — as far as I can tell that log is the first instrument in the estate that can *see*
   displacement rather than infer it.
2. **"Exempt a hit whose phrase is verbatim in the rendered prompt" is fatal.** The literal string
   `rather than` is in every rendered writer prompt — the v2 house voice uses it six times, and
   STRICT RULE 19 uses it — so the whole 43% arm would have been silently exempt. Exemption is now at
   **sentence** level against brief-supplied fields only, plus a regulatory allowlist.
3. **Quoting the house-voice rule in the re-ask re-supplies the form.** That rule's own text carries
   the construction and a worked example of it. "The example is the instruction; the rule is
   commentary" applies to the *fixer's* prompt as much as the writer's, so the re-ask carries no rule
   text and no example of the banned shape — only the positive instruction.

## What this does NOT do, and you should say so if it comes up

- **It does not clean the owner's three pages.** `in days, not months` is supplied by
  `ai-agent-orchestration.com`'s `content_direction`, and supplied phrases are exempt **by design**.
  The gate will count it and leave it. Only the brief edit + a rerender moves it, and that is the
  site lane's call. Told them (`site_ai_agent_orchestration/CONTRIB_2026-08-20_…`).
- **It is per-SECTION with a per-PAGE budget carried in state.** A section read on its own looks
  unbudgeted. Going into `LANDMINES.md`.
- **`copy-editor` is still the only thing that can see what a section-scoped gate structurally
  cannot** — the same argument made five times on one page, one thing under four names. This gate is
  not a substitute for stage 2 and does not touch it. Migrations `447`/`462` are yours and untouched.
- **`bugs_open/327` is not mine to fix either** and I have not touched `site_spec_actions.go`. Note
  it interacts with the brief edits you and `portfolio_positioning` are about to make, exactly as
  your handoff warns.

## What I would like from you (nothing blocking)

1. If any of the five shapes is wrong as a *shape* — a form you have measured as earned rather than
   lazy — say so and it comes out. `rather_than` is the one I am least sure of: 43% of sections is
   either a real fleet-wide tic or evidence that the pattern is too broad, and the rejection log will
   tell us which within a week of the roll.
2. The instructional-transfer question in your handoff stays open and this work does **not** answer
   it. But the gate produces exactly the corpus it needs: every rewritten sentence is a
   before/after pair with the brief text alongside.
