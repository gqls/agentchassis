# BUG 031 — a wrong entry in the concept register became a HIGH-severity council blocker on a correct plan

**Filed:** 2026-07-19 · travelling-docs thread (from the `bugs_open/024` council rounds)
**Severity:** medium — costs real credits and real time, and misleads any thread reading the register.
**Status:** OPEN. Diagnosed with evidence; register NOT yet corrected.
> **OWNED BY ANOTHER THREAD as of 2026-07-19 (owner-confirmed).** The
> travelling-docs thread filed this case and is **not** working it. Do not start
> a parallel fix — the correction spans five files (three register copies + two
> in `docs/leopardessconsulting/`) and two threads editing them would collide.
> The filing thread's only remaining interest is that the fix candidates below
> stay accurate; correct them in place if the owning thread finds otherwise.

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

---

## RESOLUTION — FIXED & LIVE 2026-07-20 (bugfix-031 thread)

**Status: CLOSED.** Every surface asserting the claim is corrected; the live
council rows are patched and verified. Nothing is inert — no image roll involved.

**Re-verified before fixing** (evidence items 1–3 re-run against HEAD `867049c04`):
zero `content_hash` in the rerender actions; the three section-level carries at
`rerender_page_sections_action.go:229/:239/:251`; plus the two **page-level**
bail-outs the original enumeration under-counted — `skipped` (no stored
components, `:157`) and `escalated` (incomplete content_data, `:186`), neither of
which writes or deploys. Those two are the likely true mechanism behind the
originating observation, and they are what the corrections now say.

**The replication was wider than filed** — six occurrences in five docs026 files
(not three): `register/styling-render-pipeline.md` STY-048, `.buckets/` ×2,
`.clusters/styling-nav-links.md` ×2, `extractions/U25_leopardess_social.md` — plus
`PILOT_final_four_reviewers.md` (seat #9 description), the docs014 mirror of the
leopardess RUNNING_NOTES, **and the live seat prompts themselves**: the claim was
embedded verbatim in `agent_definitions` rows for `fix-proposer` AND
`council-gate` (`review_render_guardian.config.prompt_template`), seeded by
`0NN_fix_proposer_v16_render.sql:47`. The register fix alone would have left the
blocking reviewer quoting the false contract indefinitely.

**What was done (all fix candidates):**
1. Register + all copies corrected with visible `CORRECTED 2026-07-20` markers
   citing the code (candidate 1). Remaining greps for the phrase hit only
   correction/refutation records that quote it to name what was wrong.
2. Source docs corrected (candidate 2): `docs/leopardessconsulting/RUNNING_NOTES.md`
   Turn 14 (visible correction block, owner's practical conclusion preserved —
   assemble mode IS right for chrome changes, for the bail-out reasons, not a
   hash), its docs014 mirror, and `scripts/reassemble_pages.sh` header comment.
3. Convention added (candidate 3): docs026 `README.md` § "Contract claims need a
   citation" — a **what:** asserting a code contract carries `file:line` or is
   voiced as an observation.
4. Council-side guard (candidate 4): the corrected seat bullet now ends
   "cite the code path, not a register entry — the register is documentation,
   not the implementation."
   **Live fix**: `PATCH_render_guardian_031_content_hash.sql` (snapshot +
   surgical two-substring replace, LIKE-guarded/idempotent) applied to
   `fix-proposer` (`UPDATE 1`, false-claim position 0, correction at 702), then
   `099_SYNC_gate_roster.py --apply` mirrored to `council-gate` (dry-run drift:
   `review_render_guardian` only; 15 seats stable; gate transform intact —
   `input_data.rationale` present). v16 seed corrected so a replay cannot
   resurrect the claim.

**Verification run** (2026-07-20): both live rows show
`position('content hash is unchanged')=0` across the whole config and the
corrected text present; `grep content_hash` over the rerender actions still
returns nothing. The remaining test — a council submission touching the rerender
path drawing no content-hash objection — will be proven by the next real
submission; the text that produced the objection no longer exists anywhere the
seat can read.
