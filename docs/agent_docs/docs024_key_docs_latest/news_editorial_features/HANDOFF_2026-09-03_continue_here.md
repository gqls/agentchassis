# HANDOFF — news_editorial_features, 2026-09-03. START HERE.

Supersedes `HANDOFF_2026-08-25_continue_here.md`. **Its predecessors are not dead and
this file does not restate them:**

- **08-21 §3 (the page recipe, proven three times) and §9 (the ten traps)** still
  govern before shipping a feature page.
- **08-24 §8** is the instance-scope evidence trail; **08-25 §1 is DONE** (the three
  SQL scripts ran, the directory is deleted — see that file's banner).
- **The companion lane's `editorial_design_uplift/HANDOFF_2026-09-02_continue_here.md`
  §9 is the post-roll verification for this lane's code.** Read it rather than
  re-probing; its probe hazards are inherited below.

Everything is measured 2026-09-03 unless marked. `IMAGE_TAG` is `v1.0.1357`.

---

## 1. THE HEADLINE — P1's code SHIPPED, and P1 is NOT DONE

**`035 P1` is in `v1.0.1355`** (companion lane's fully-controlled binary probe,
controls on opposite sides). `rerenderFlatSections` and `hierarchyChildrenOf` are
PRESENT in the binary.

**⚠ `recomposeAncestors` is ABSENT from the binary, and that is CORRECT.** It has **no
caller** — every source reference is inside its own definition, so the linker
eliminates it. The commit (`3fd617ef6`) shipped; the capability is unreachable.
**An absent capability literal means "not reachable in this build", never "the commit
did not ship".** Verified again 2026-09-03: `grep recomposeAncestors(` outside its own
definition returns nothing.

**So the true state is: the READ path is live and inert; DIRECTION 2 IS WRITTEN AND
UNWIRED.** That is the single most important fact in this file.

| piece | state |
|---|---|
| walk + `deriveRenderMode` third value | shipped, inert (`1f745e730`) |
| `loadStoredSections` reads `id`, `parent_instance_id` | shipped (`bd811fa93`) |
| membership helpers | shipped, called (`bc8167100`) |
| `recomposeAncestors` | shipped, **NO CALLER** (`3fd617ef6`, `fdb03cbc1`) |
| direction-1 guards, both paths | shipped and wired (`028c3e112`) |
| the extraction + decomposition | shipped (`2a0bdb001`, `94f81cc60`, `22ed53ee7`) |
| 137 residue fix, both twins | shipped (`3b1389ca0`) |
| **the walk BRANCH in the render loop** | **NOT WRITTEN** |
| **byte-identity tests + mutation check** | **NOT WRITTEN** |
| **migration 643** | **VOID — see §3** |

Composition remains unreachable: **0 of 3,035 `page_components` carry a
`parent_instance_id`; 0 of 386 `content_components` declare a `slots` block.**

## 2. WHY THE METHOD CHANGED — plans to code

**The council said REVISE three times on correlation
`53d71504-8cd1-49bc-8e2d-d1465ba65103`, and the GATING objection was identical every
round: a sketch calling symbols no edit defines.** Round 1 `recomposeParent`
undefined; round 2 renamed, still undefined; round 3 defined, its three helpers
undefined. One level deeper each time.

That is not a documentation slip. It is what authoring code in prose produces: every
sketch bottoms out in new symbols and nothing stops it. **Round 4 must be real diffs
from compiling, tested code.** Do not submit another sketch.

**Writing it found five things three rounds of review could not.** Each is recorded
where it lives; they are listed here so the next session does not re-derive them.

1. **The `db`/`tx` split is FORCED.** `buildRenderContextFromDB` and
   `DeriveAndBindInstanceToken` take `*sql.DB` and cannot take a `*sql.Tx`.
2. **Two estate guards caught the new `rendered_html` writer within minutes** — the
   floors coverage test and `pattern-check`'s `unrepaired-component-write`. Neither
   was taken as an exemption; both are now called.
3. **The two direction-1 guards are DIFFERENT PREDICATES.** `apply_section_edit`
   holds a ROW → `hierarchyChildrenOf`. `RenderComponentAction` holds a component
   DEFINITION, has no row, and RETURNS html → `hierarchySlotsFromSchema`. The
   submitted plan used the row predicate for both and **would not have compiled**.
4. **The loop runs TWO counters with DIFFERENT advance rules.** `sectionOccurrences`
   advances for EVERY row including carries (per-section imagery); the instance token
   counter advances only for RESOLVED rows. The walk must thread both, differently,
   and getting either wrong is silent.
5. **§3 below.**

## 3. ⚠ NOTHING READS `render_mode` — P1's routing story has no landing place

`[MEASURED 2026-09-02]`, the whole chain:

- **WRITTEN** by `deriveRenderMode` (`store_generated_component_action.go:732` INSERT,
  `:822` UPDATE)
- **SCANNED** into `Component.RenderMode` (`load_component_library_actions.go:228,275`)
- **READ** by nothing. `.RenderMode` has no other use in `platform/` or `internal/`.

And 035 D3's claim that `check_render_mode` "routes sections to LLM generation when
`render_mode == 'agent'`" is **false three ways**: it is not a top-level step (it sits
under `{workflow,steps,process_sections_loop,config,sub_workflow,steps}`); it is a
**binary conditional** (`condition`/`then_step`/`else_step`), not a `conditions` array;
and it branches on `current_section.llm_field_specs`, **not on `render_mode` at all**.
The step containing "rerender_mode" is `check_rerender_mode` on the **page-rerender**
agent, branching on `input_data.spec.reason` — a different step on a different agent.

**Consequences.** Migration 643 is **VOID — drop it, do not re-derive it.** There is
nothing to add an arm to and adding one routes nothing. `deriveRenderMode`'s third
value is inert in a STRONGER sense than earlier files claim: not "inert until a slots
block exists" but inert even then.

**RECOMMENDED SCOPING, for the owner:** move the routing arm from P1 to **P2**, where
generation happens and a composite would actually need routing. **P1's own acceptance
does not need it** — recompose one page by hand, prove byte-equivalence, rewrite one
child, prove siblings untouched. All hand-made rows and the edit path. That shrinks P1
to exactly what its name says and lets it finish. Not yet ruled on.

Also verified live: `jsonb_set(doc, path, NULL || jsonb_build_array(...))` returns
**NULL for the whole document**. Any future edit to that step must guard the array
shape before writing.

## 4. What to do next, in order

1. **Wire direction 2.** `recomposeAncestors` has no caller — call it from
   `apply_section_edit` after the child write, in the same transaction, and file the
   `needs_rerender` item for each stale ancestor it reports. This is the piece whose
   absence the binary probe is currently reporting.
2. **The walk branch** behind the `hierarchy_walk` step-config flag, threading BOTH
   counters (§2.4).
3. **The byte-identity tests** with a mutation check — and note the guarantee now
   RESTS on them (§5).
4. **Round 4** as real diffs, on the same correlation, with migration 643 dropped and
   §3 explained.
5. Then the live recomposed page, and the register entry saying **"exercised on
   `<page>`" with a date** — never "deployed" (035 §6.8).

## 5. A DESIGN FORK I TOOK — the council was told the older story

Rounds 2 and 3 promised the flag-OFF branch would be the **byte-identical original
loop**. It is not, and cannot be: keeping it would mean duplicating ~200 lines,
because a composition child needs the same carry semantics as a section, and two
copies drift.

**So there is ONE implementation, and the guarantee moved from a structural claim to
the byte-identity TEST.** The flag now guards only whether the walk composes children.
**This makes the test's mutation check load-bearing rather than decorative** — a
byte-identity test never shown to fail proves nothing. Say this plainly in round 4;
the seats were told otherwise.

## 6. G6 — Fable's draft is written and awaits four owner decisions

`features_open/035_G6_DRAFT_stored_pattern_library.md` (`30d161249`), written by Fable
at the owner's direction. 035 itself untouched. **A pattern is a promoted
`content_components` row** — concrete slots, `component_id` per slot, a
`pattern_provenance` jsonb, `created_from='promoted'`; **no new table**, because a
second home for slot declarations is the drift class D3 already retired.

The **loops clause is satisfied compositionally** — parts scored at their grain, the
whole ruled by the experience council because a promotion IS a design decision — with
**no third loop**, and Fable flagged that as a **weakening for the owner to overrule**.

Four decisions in its §7: the clause's meaning (the substantive one), library scope
(global vs new schema for per-site), version binding on application, and whether
auto-apply is ever revisited.

**⚠ 035 hazard 10 blocks G6's gate.** Verified by probe 2026-09-02:
`extractTemplateVariables` returns deduplicated FIELD ROOTS, so
`{{.slots.lead}}{{.slots.quote}}` yields **`[slots]`** — one name for two slots. Every
composite therefore mis-scores in `compute_component_quality`, silently, as a low
score rather than an error. G6's promotion gate requires a non-NULL `quality_score`
per child family, so it would gate on a number computed wrongly. **This is P1/P2's to
fix, not G6's.**

## 7. OPEN ARCHITECTURE QUESTION — asked by the owner 2026-09-03, unanswered

> *"Do we need another loop like the experience loop that sorts out this component
> composition? we have theme kits and page composition and experience loop and visual
> designer and designer, might we extend one of their remits or another agent or
> workflow remit or is it a separate responsibility. I think whatever composition we
> decide on it should act like the components — it can have defaults and sites or
> pages can fork as they want."*

**Not yet answered.** The design principle in the last sentence is the operative part
and should constrain whatever is decided: **composition should carry defaults and be
forkable per site or page, exactly as components are** (`content_components` +
`forked_from`).

Live agents in that space, measured 2026-09-03: `brand-designer`,
`design-audit-agent`, `design-critique-agent`, `design-discovery-agent`,
`experience-approval-council`, `experience-planner`, `experience-register-writer`,
`feature-designer`, `reader-experience-auditor`, `site-design-planner`,
`visual-design-auditor`, `visual-designer`, `webdesign-agent`.

Fable's G6 draft already argues AGAINST a third loop (§3 of the draft: an
arrangement-scorer would join §3's dormant mechanisms the day it shipped). That answer
was given for G6's promotion gate specifically and **is not the same question** as the
owner's — his is about the whole composition remit, not just pattern promotion.

## 8. Coordination — four lanes, all answered, nothing owed

- **`bugfix_410` (scan-loss)** — their guard shipped into `loadStoredSections`
  (`7c443aac6`, DBI-027). **Standing constraint on this lane: any column added to
  that projection must be `NOT NULL` or `COALESCE`d** or their guard fires on our
  edit. Ours are fine and round 4 adds none.
- **`inline_guide_imagery` / dartsonline guides** — answered: do NOT wait for P5;
  **the guides are not a P2 consumer** (figure-between-sections suffices, so finer
  FLAT rows give them durability today). Recorded in 035 §8 with their scope caveat.
- **`loanzy.uk` (UK news region)** and **`boxingonline`** — both routed away as
  name-matches on "news"; this lane owns editorial FEATURE PAGES, not feed ingestion.
  The boxingonline handoff carried a verified `evidence_base` finding (2 rows, 3
  facts, 0 `allowed_entities` against sites carrying 111–116 facts per row).
- **`bugfix_357`/RFC_046** — the primitive fix for the provenance stamp is theirs and
  must NOT be bolted into P1.

## 9. Traps from this stretch

1. **A tool's verdict is scoped to its ARGUMENT.** I ran `probe-page-url.sh` against a
   customer domain, got a true CONTROL-FAIL about *that domain*, and restated it as a
   fact about *the site* — then asked a peer to put it in a bug file. The serving host
   discriminates perfectly. **Probe the slug, never the customer domain**
   (`sites.publish_target` names what serves).
2. **`bugs_open/137` is CLOSED** (`bugs_closed/137`, live since 2026-07-31). The
   `pattern-check` advisory's own text still calls it `bugs_open/137`; I repeated the
   tool's string for a day. The gate's message is stale, the bug moved.
3. **Reaching for the built tool is not the end of the discipline.** `probe-page-url.sh`
   exists because four sessions composed URLs by hand; I was occurrence five, six days
   after logging occurrence four myself.
4. **Inherited a denominator, dated it today.** "0 of 1,903" in a council submission:
   numerator mine, denominator four days old. A claim can carry two counts and one
   date, and the composite takes the fresher.
5. **Backticks in `git commit -m` execute.** Cost a word out of a commit message that
   cannot be amended. Use `-F -` with a heredoc.
6. **Probe hazards inherited from the companion lane** (their §9): an expired token
   makes EVERY symbol read absent — so a must-be-present control is mandatory; and
   BusyBox `grep` over `/proc/1/exe` needs 100–120s per grep, where a timeout kill is
   indistinguishable from a negative. Run them singly.

All of the above are in `WRONG_CALLS.md` with their cheap checks.

## 10. What NOT to do

- **Do not submit another council sketch.** Round 4 is real diffs or nothing (§2).
- **Do not re-derive migration 643.** It is void (§3).
- Do not seed any `composite` row or `parent_instance_id` before the walk branch
  ships (035 §9 r1).
- Do not read `recomposeAncestors`' absence from the binary as a missing commit (§1).
- Do not bolt RFC_046's primitive fix into P1 — 357's to route.
- Do not answer 035 §6.1's completeness floor by lowering `prune_floor_ratio`.
- Do not treat the guides as a P2 consumer without NEW evidence — their answer was
  scoped to eleven guides on one site, judged from rendered structure.
