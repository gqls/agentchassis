# 271 — `spec.content_guidance` has no reader: every `content_rewrite` brief in the fleet is silently ignored

**Filed:** 2026-08-13 · **Lane:** `ai_site_selling_automation` · **Severity: high** —
the field every operator and the platform's own gap-planner use to tell the
writer WHAT TO CHANGE is write-only. The work happens anyway, steered by
nothing but `writer_block` and the existing page, and reports `complete`.
**Class:** structural (a dead channel that looks like the steering wheel).

> **STATUS: OPEN.** Verification: two independent code reads eleven days apart
> plus rendered-prompt probes across six LLM calls (§3) — stated per the
> 2026-07-31 ruling as the substitute for a 090 run; a 090 remains cheap to add
> if a fixing thread wants the independent pass.

## 1. The symptom, as an operator experiences it

You file a `content_rewrite` item with careful `content_guidance` ("state X,
remove Y, add Z"). The item completes; the page is rewritten fluently; the
specific asks are randomly honoured or not. Whether they are honoured depends
entirely on whether `evidence_base.writer_block` happens to say the same thing
— which is why the failure is intermittent enough to have survived: briefs
that RESTATE the register appear to work.

Measured cost in one lane, 2026-08-12/13: **four LLM rewrite rounds** burned
trying to get six service names into one FAQ answer, three of them spent
"fixing" real-but-incidental register defects the rounds appeared to expose.
The guidance never reached any writer prompt in any round.

## 2. The mechanism

`content_guidance` is written into `site_work_items.spec` by four emitters —
`apply_gap_plan_action.go:209` (the platform's own content-gap pipeline),
`tool_content_item.go:170`, `create_tool_component_action.go:448`,
`deploy_tool_action.go:575` — and **read by nothing on the work-item path**.
The only read of the string anywhere (`apply_gap_plan_action.go:178`) takes it
from a gap PLAN, not from an item spec, and only to copy it into the next
spec. `page-build-handler → plan_sections → generate_content` builds the
writer's prompt from `writer_block` + the section's stored content + the
section brief; the item's spec never contributes prose.

## 3. Evidence

- **Grep at HEAD (2026-08-13):** `grep -rn content_guidance --include=*.go
  platform/ internal/` → the four writers above, zero readers. (Spelling
  caveat acknowledged; closed by the probes below.)
- **Independent prior read, 2026-08-02:** `bugs_closed/177` close-out footnote:
  *"written by all four tool emit sites and read by NO handler on the
  work-item path... dead weight even on satisfiable items."* Same conclusion,
  different session, different motivation — and it sank as a footnote.
- **Rendered-prompt probes (the artefact, not the code):** rounds 3 and 4 of
  the webdesign.uk FAQ job, six `page-content-writer` calls
  (`llm_call_log` 1e65e476/9c7735b8/089e055f and f0fd40b4/b45ed6de/df9f43b5):
  six distinctive guidance phrases ("One job", "Change no other answer",
  "grouped by purpose", "Extend the FAQ answer", …) — **absent from every
  prompt**, while `writer_block` content is present verbatim in all of them.
- **The discriminating case:** round 5 moved the same instruction INTO
  `writer_block` and changed nothing else about the item (its guidance says
  only "this field has no reader"). See item `gapfill5_faq` for the outcome.

## 4. Blast radius

Every `content_rewrite` filed with a brief, fleet-wide, forever — including
every `add_to_page` item the platform's own `content-gap-planner` emits (its
LLM-composed `content_guidance` is the entire point of that pipeline). The
gap pipeline "works" to the extent the gap planner's asks coincide with what
the register and the section structure already imply.

## 5. Fix candidates, ordered by what closes the door

1. **Wire the field in**: `plan_sections` (or the generate_content prompt
   assembly) injects `spec.content_guidance` into the section brief for every
   section the item touches. One insertion point; makes four existing writers
   and every historical brief meaningful. Council: it changes what a shared
   pipeline GUARANTEES about items already in flight — submit before/alongside.
2. **Kill the field**: delete it from all four writers, document
   `writer_block` / section data as the only steering channels. Honest, but
   forfeits per-item steering entirely — the gap pipeline loses its purpose.
3. Do nothing but document — rejected by writing this file: a steering wheel
   that is connected only sometimes is worse than either connecting it or
   removing it.

## 6. Verification for any fix

File a `content_rewrite` whose guidance demands a sentinel phrase that appears
NOWHERE in the register or existing copy, on a canary page. Grep the rendered
prompt (`llm_call_log.prompt_rendered`) for the sentinel, then the artefact.
Today both greps return nothing; after fix 1 both must return it. Run the
negative control too: an empty-guidance item must not regress.

## 7. Relations

`bugs_closed/177` (the footnote that saw it first) · `bugs_open/268` §the
four rounds (the incident that measured the cost) · the memory lessons
`writes-the-field-is-not-reads-the-field` and `a-doc-comment-is-not-an-
enforcement-mechanism` (this is both at once: a field whose NAME is the
documentation) · `webdesign_uk_build_service` NOTES 2026-08-09 ("facts[] is
bookkeeping; writer_block is the wire") — the same law, one layer down.


---

## 8. THE DISCRIMINATING CASE RAN — the fix-1 direction is CONFIRMED at the artefact (2026-08-13 22:1xZ)

`gapfill5_faq`: identical item shape, `content_guidance` deliberately inert
("this field has no reader"), the instruction moved into `writer_block` as an
imperative (`SQL_2026-08-13c`). **Outcome: all six names present in the
component and on the served page** (deployed 22:19:40Z; verified cache-busted
the next morning, buttons and terms intact) — after four rounds where the same
ask, carried only by `content_guidance`, produced nothing.

One channel dead, the other alive, same writer, same page, same day. §6's
verification recipe stands for whoever wires the field in; note the item's own
status reads `failed` while the page deployed fine (the spawn→call handshake
race, fourth sighting in this lane — verify at the artefact, not the status).
