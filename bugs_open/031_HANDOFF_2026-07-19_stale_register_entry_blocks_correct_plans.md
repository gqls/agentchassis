# BUG 031 — a wrong entry in the concept register became a HIGH-severity council blocker on a correct plan

**Filed:** 2026-07-19 · travelling-docs thread (from the `bugs_open/024` council rounds)
**Severity:** medium — costs real credits and real time, and misleads any thread reading the register.
**Status:** OPEN. Diagnosed with evidence; register NOT yet corrected.

---

## Symptom

Council-gate submission `7ef4de4e-3930-47fe-8ca6-ba40a2d440cc` (the fix for
`bugs_open/024`) was blocked at **HIGH severity** in round 3 by the
`render_guardian` seat:

> Edit 1 raises the rerender request with `spec.reason='section_data_resolved'`,
> which routes into SCOPED rerender mode. The pipeline's own contract states
> scoped mode *"REGENERATES section HTML from content_data but SKIPS pages whose
> content hash is unchanged."* The bug being fixed here is precisely a case where
> `html_template` changes but `content_data` does NOT change.

If true, the fix would be a guaranteed no-op. It is not true of the code.

## Where the claim comes from

Not from the code — from **our own knowledge base**, quoted as if it were the
implementation's contract:

- `docs/agent_docs/docs026_concept_register/register/styling-render-pipeline.md:389`
- `docs/agent_docs/docs026_concept_register/.buckets/styling-render-pipeline.md:590` and `:1193`

> "page-rerender has two modes: with `spec.reason='section_data_resolved'` (or
> `'image_landed'`) + `spec.page_name` it regenerates section HTML from
> content_data but SKIPS pages whose content hash is unchanged — silently wrong
> for header/footer-only changes; assemble mode (page_id, no reason) re-embeds
> current header/footer unconditionally."

Same claim also in the originating workstream's own docs:
`docs/leopardessconsulting/RUNNING_NOTES.md:681` and
`docs/leopardessconsulting/scripts/reassemble_pages.sh:11`.

## Evidence that it is wrong

1. **No such code exists.** `grep -rn "content_hash" --include=*.go platform/ internal/`
   returns **zero** hits in `rerender_page_sections_action.go`,
   `rerender_single_page_action.go` or `create_rerender_items_action.go`. The only
   `content_hash` users in the repo are `rag_actions.go`, `code_symbols_actions.go`,
   `vet_med_price_scrape_action.go` and a struct field in `site_db_actions.go`.
2. **It never existed.** `git log -S "content_hash" --` over those three files
   returns **no commits**. This is not a stale-but-once-true entry; the mechanism
   was never in that path.
3. **The real skip paths are enumerable and none is hash-based.** The re-render
   loop carries a section (rather than re-rendering) in exactly three cases:
   component not found in `schemas` (~line 229); `planSection` status != `"ready"`
   (~239); empty `html_template` (~250). Anything else is re-rendered from
   `htmlTemplate`.
4. **No live agent config gates on it.** No `agent_definitions` row references
   `content_hash` except `fix-proposer` / `council-gate` themselves (the council
   agents, whose prompts carry reviewer text).
5. **Empirical.** Probe work item `478c44c9` (`reason=section_data_resolved`, on
   the page whose `content_data` had not changed) **did** reach `rerender_sections`
   and process it — `section_count:1`, `escalated:true`. Scoped mode engaged and
   skipped nothing on any hash.

## Root cause (charitable and, I think, correct)

The originating thread almost certainly **observed a real symptom** — pages not
updating after a change — and **inferred a wrong mechanism** for it. From the
outside, the three `carried` paths and the content pre-check escalation are
indistinguishable from "it skipped because nothing changed". That inference was
then written into the register as a *contract*, in the register's confident
"what:" voice, with no file:line and no reproduction.

The register is fed to agents as authoritative context. A reviewer seat quoted it
verbatim, called it "the pipeline's own contract", and blocked a correct plan at
high severity. It cost a full council round (~30–60 min of queueing plus credits)
to disprove with three greps and one probe.

## Why this is worth a bug and not just a doc edit

- The register **is agent-facing input**, not prose for humans. A wrong entry does
  not merely mislead a reader; it is quoted as evidence in machine reviews.
- The wrong claim is **load-bearing and specific** ("skips pages whose content
  hash is unchanged"), which makes it exactly the kind of statement a reviewer
  will treat as decisive.
- It is **replicated in three files** plus two in the originating workstream, so
  correcting one leaves four.
- Anyone touching the page-rerender pipeline will hit it next.

## Fix candidates

1. **Correct the entry** in `register/styling-render-pipeline.md` and both
   `.buckets/` copies: replace the content-hash claim with the three actual carry
   conditions, and say what the originating thread most likely hit
   (`carried` on unresolved plan / missing component / empty template, or the
   content pre-check escalation — see `bugs_open/024`). Cite file:line.
2. **Correct the source docs** in `docs/leopardessconsulting/` so the claim does
   not simply get re-harvested into the register.
3. **Structural — the real ask:** register entries that assert a *code contract*
   should carry a **file:line citation**, and ideally be checkable. An entry with
   no citation is an observation, and should be voiced as one ("we observed X",
   not "the pipeline skips Y"). Worth a convention, because this class of error
   is invisible until it blocks something.
4. **Council-side mitigation:** a reviewer citing a register entry as decisive
   should be expected to name the file:line in the *code*, not the register.
   Consider prompting seats to distinguish "the register says" from "the code
   does". This is the cheapest guard and it generalises.

## How to verify a fix

```bash
# the false claim should be gone from all three register copies
grep -rn "content hash is unchanged" docs/agent_docs/docs026_concept_register/
# and the code should still show no hash logic in the rerender path
grep -rn "content_hash" --include=*.go platform/orchestration/actions/rerender_*.go
```
Then: a council submission touching the rerender path should no longer draw a
content-hash objection.

## Related

- `bugs_open/024` — the plan this blocked; its round-4 submission carries the full
  rebuttal and evidence in `grounded_in`.
- Pattern for 016b §9: **verify a cited "contract" against the code before
  revising a plan around it** — a confident claim in our own knowledge base is not
  evidence, and this one had never been true.
