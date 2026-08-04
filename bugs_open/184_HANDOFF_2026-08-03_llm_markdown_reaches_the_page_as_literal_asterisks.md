# 184 — LLM-emitted markdown reaches the rendered page as literal `**asterisks**`

**Filed 2026-08-03** by the `mortgagecalculator_couk_adoption` lane, found on the
first page it built. **OPEN, unowned. Low severity, but it is live copy on
production sites and it is trivially detectable.**

## Symptom

A content writer emits markdown emphasis inside a text field. The renderer treats
that field as plain text (correctly — it is not a markdown field), so the asterisks
reach the visitor verbatim:

> Banks evaluate your application using a `**Decision Engine**` (an automated
> algorithm that grades your financial history).

Live at `https://mortgagecalculator.co.uk/guides/first-time-buyer/index.html`
(hero slot) as of 2026-08-03.

## Scope — small, and cross-site, which is the point

Three components fleet-wide, on **three unrelated sites and three different slot
types**, so this is not one agent or one template misbehaving:

```sql
SELECT s.domain, p.url, pc.slot_name,
       substring(pc.content_data::text from '\*\*[A-Za-z][^*]{2,40}\*\*') AS sample
  FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
 WHERE pc.content_data::text ~ '\*\*[A-Za-z][^*]{2,40}\*\*';
```

| domain | url | slot | sample |
|---|---|---|---|
| mortgagecalculator.co.uk | /guides/first-time-buyer/index.html | hero | `**Decision Engine**` |
| gaswholesalers.com | /how-pricing-works.html | pricing | `**Recommended next steps:**` |
| webdesign.co.uk | /news/index.html | news-listing | ``**the `animation`**`` |

Note the third: it carries a **backtick code span as well**, so a fix that only
strips `**` leaves that one still wrong.

## Why it is worth a file despite being three rows

It is the cheapest possible class to detect and it is **silent** — every existing
check passes. The page renders, the HTML is valid, the component is structurally
complete, `build_status` reads `deployed`, and nothing in the discovery-check layer
looks at it. The only reason it was found is that a human read the prose.

## Root cause (candidate — NOT yet verified in code)

> `[UNVERIFIED]` I did not trace which writer produced these three, and the three
> come from different agents, so a single culprit is unlikely. The general shape:
> nothing on the write path normalises or rejects markdown syntax in fields that
> are rendered as plain text, and prompts do not forbid it. **Do not quote this
> paragraph as a diagnosis** — it is where to start looking, not a finding.

## Fix candidates, ordered by what closes the door

1. **Detect it.** A discovery check in the `check_*` family, matching
   `\*\*[^*]+\*\*`, `` `[^`]+` `` and `^#{1,6} ` in rendered text slots. Cheap,
   offline, no LLM. This is the one that generalises — it catches the next writer
   that does it, including one that does not exist yet.
2. **Normalise on write** — convert `**x**` → `<strong>x</strong>` for slots whose
   schema says they accept inline HTML, and strip otherwise. Needs care: the
   renderer's escaping rules differ per slot, so this is not a blanket
   `strings.ReplaceAll`, and doing it wrong turns a cosmetic bug into an injection
   surface.
3. **Forbid it in the prompts.** Weakest — it is an instruction, not a control, and
   `LANDMINES`/`WRONG_CALLS` are full of cases where a prompt instruction was
   treated as an enforcement mechanism. Do this *as well as* 1, never instead.

## How to verify a fix

Re-run the query above; expect 0 rows. Then confirm at the **artefact**, not the
DB — `curl` the page and grep the visible text, because `content_data` and
`rendered_html` are separate copies and a repair to one does not imply the other
(see `bugs_open/097`).

## Progress — 2026-08-03/04

**Scope decided**: detect (candidate 1) + prompt hardening (candidate 3).
Normalise-on-write (candidate 2) deliberately deferred — the render path has
zero HTML escaping anywhere (`text/template`, not `html/template`), so a
markdown→HTML converter at write time would be writing into an unescaping
pipe, and mutating the shared `SavePageSectionsAction` choke point changes what
that save guarantees for every writer. Named as future work in the register
entry (CQ-019), not silently dropped.

**Built, not yet enabled/applied/live:**
- `platform/orchestration/actions/discovery_checks/check_literal_markdown.go` —
  new discovery check, dual-surface (`content_data` + `rendered_html`, the
  `check_unverified_claims`/093 precedent), letter-guarded regex patterns
  (bold/code-span/heading) that do not fire on `3 * 4`, `#fff`, `#1 rated`,
  JS `` `${x}` ``. Routes to `page-content-writer` for auto-repair (the
  `check_placeholder_contact` precedent — this is a definite mechanical
  defect, not a judgement call). Retracts via `CheckResult.Resolved`
  following `check_required_fields_missing`'s shape (no hand-rolled status
  filter; `resolveWorkItems` alone owns `workItemClosedStatuses`).
  Unit-tested (`check_literal_markdown_test.go`), `go build`/`go vet`/`go test`
  clean for the package (one pre-existing, unrelated failure in the same
  package — `TestRegisteredVerifiersMatchClaimTimeoutExclusion` on
  `page_canonical_collision` — confirmed via `git stash` to predate this
  change and belong to a different concurrent thread's work).
- `docs/agent_docs/sql_for_agents/303_enable_literal_markdown_check.sql` —
  enables the check on `quality-discovery-agent`. **Apply AFTER the image is
  live** (unregistered check names fail loudly since bugs_open/149 B4).
- `docs/agent_docs/sql_for_agents/304_forbid_markdown_in_text_fields.sql` —
  extends live STRICT RULE 9 of `page-content-writer`'s `generate_content`
  prompt in place (scoped `replace()`, fail-loud verification, backup table),
  measured live 2026-08-03 that `content-writer` and
  `simple-content-writer-with-approval` don't carry this prompt block or the
  `save_page_sections` write path, so they are not touched.
- Concept register `CQ-019` added (`content-quality.md`, `000_concept_index.md`).

**Still to do before this bug can close**: submit to the council gate; build +
push + deploy an agent-chassis image carrying the new check (pod-verify the
symbol); apply migration 303 (image first), then 304; let a discovery run fire
on the three founding sites and confirm `page-content-writer` repairs the three
rows; re-run this file's own SQL (expect 0 rows) AND curl the three live URLs
(artefact-level, per the note above — `content_data` and `rendered_html` are
separate copies). Bug stays OPEN until fixed AND live AND the three founding
rows are verified clean at the artefact, not merely at the DB.

## Progress — 2026-08-04 (continued) — everything BUILT is now LIVE; repair cycle not yet run

**Commits**: `0dd08d6a5` (fix: check + tests + both migrations + register +
this file), `de62a2c63` (gofmt follow-up on the test file). Both carry
`Council-Submitted: eb8f9cc0-3a28-437a-8725-1a785f3d12b5`.

**Council**: still `EXECUTING_STEP` at `review_improvement_guardian` as of
2026-08-04 ~09:15 (last checked). Poll:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
 WHERE correlation_id='eb8f9cc0-3a28-437a-8725-1a785f3d12b5' AND kind='council_report';
```
If APPROVED, the trailer is already `Council-Submitted:` so `098` credits it
automatically — no amend needed. If REVISE/REJECTED, read the objections
(`SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY
created_at DESC LIMIT 1`) and judge whether they change anything below; the
code is already on the shared branch and live either way.

**Image: LIVE, verified at the artefact, not just the roll.** A concurrent
whole-fleet `make release` (started 08:53, unrelated to this bug) deployed
`v1.0.1247` at 09:06 — pod-grepped and it carried ZERO occurrences of
`literal_markdown`, because its build read HEAD *before* `0dd08d6a5` landed
(09:05:28) — a live instance of the fleet's own "a roll is not evidence your
fix shipped" landmine, caught rather than assumed. Built `v1.0.1248` from HEAD
myself (`make build-agent-chassis IMAGE_TAG=v1.0.1248` — note another
concurrent session independently built the *same* tag number at the same
time, since "highest seen + 1" is deterministic from shared state; harmless,
just wasted a duplicate build), verified locally
(`docker run --rm --entrypoint sh ... strings /app/agent-chassis | grep -c
literal_markdown` → 11), pushed, bumped the kustomization `newTag`. A further
concurrent deploy cycle (not mine) then moved the fleet to `v1.0.1250`.
**Re-verified 2026-08-04 on the CURRENT pods**
(`agent-chassis-88cf8787-*`, both replicas): `literal_markdown` → 11 on each,
a negative-control nonsense string → 0 on each. **The fix is live fleet-wide,
confirmed at the binary, right now.**

**Migrations 303 and 304: APPLIED and verified live** (2026-08-04):
```sql
-- both return true:
SELECT default_config #> '{workflow,steps,run_checks,config,checks}' ? 'literal_markdown'
  FROM agent_definitions WHERE type='quality-discovery-agent' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
SELECT (default_config #>> '{workflow,steps,process_sections_loop,config,sub_workflow,steps,generate_content,config,prompt_template}') LIKE '%Plain string also means NO markdown syntax%'
  FROM agent_definitions WHERE type='page-content-writer' AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
303's idempotent guard and 304's fail-loud verify both passed cleanly (no
RAISE fired; 304 printed `OK: rule 9 extended, rules 10 and 14 intact`).

**NOT yet done — this is the actual remaining work:**

1. **No discovery run has fired against the three founding sites yet.**
   `SELECT * FROM site_work_items WHERE item_type='literal_markdown'` returns
   **0 rows** — the check is enabled but nothing has invoked
   `quality-discovery-agent` for these sites since. Checked
   `scheduled_tasks`: every `quality-discovery-agent`/`*-discovery-agent` row
   is a **one-shot, already fired and `enabled=false`** (per-site,
   `oneshot-<agent>-<site-slug>-<date>` naming) — there is no live recurring
   schedule to just wait for. Site ids for the three founding sites:
   `mortgagecalculator.co.uk` = `62b5978e-4271-4589-8e00-4baebfc0447c`,
   `gaswholesalers.com` = `5fe15466-4e2e-4ff2-981e-98c1b7074002`,
   `webdesign.co.uk` = `6b49db8e-d447-4467-8277-4f3018af9897`.
   **Next action**: dispatch `quality-discovery-agent` for each site id —
   either insert a one-shot `scheduled_tasks` row (`fire_message=true,
   enabled=true, target_agent_type='quality-discovery-agent', input_data:
   {"site_id": "<id>"}` — **read what actually polls/fires this table and
   at what cadence before assuming "insert = immediate"**, I did not verify
   that path) or fire a direct kafka message to
   `system.agent.generic.requests` (`action=orchestrate`,
   `agent_type=quality-discovery-agent`) mirroring the envelope pattern in
   `scripts/trigger-landmine-verifier.sh` / `097_TRIGGER_council_review_v1.sh`
   (topic and `input_mapping` will differ — read a live `quality-discovery-agent`
   `agent_definitions` row's `workflow.start_step`/expected `input_data` shape
   first, don't guess it from the council script).
2. **Whether `page-content-writer` can actually service a
   `literal_markdown`-shaped work item is UNVERIFIED.** The risk was named at
   plan time (council submission `risks` field) but never checked against the
   live handler's routing table. Before trusting auto-repair, grep how
   `page-content-writer` dispatches by `item_type` today (it already handles
   `placeholder_contact` per that check's precedent — confirm `literal_markdown`
   either falls into the same generic path or needs an explicit addition).
3. Once items are filed and (auto- or manually-) repaired: re-run this file's
   own SQL (§ "Scope" above, expect 0 rows) **and** curl the three live URLs
   for the literal string (artefact-level — `content_data` and `rendered_html`
   are separate copies, bugs_open/097). Confirm the check's `Resolved`
   retraction actually closes the three items on the next discovery pass
   (`SELECT item_key, status, result->>'reason' FROM site_work_items WHERE
   item_type='literal_markdown'`).
4. **Close-out, once fixed AND live AND verified at the artefact**: move this
   file to `bugs_closed/` — **the number 184 is AMBIGUOUS** (shared with
   `bugs_closed/184_..._three_more_detectors_...md`, an unrelated closed
   case) — `git mv` **both** the old and new paths and name **both** on the
   `git commit` pathspec in one go (the LANDMINES entry on this exact trap:
   a pathspec commit after a bare `git mv` can silently ship a COPY, leaving
   the file in both dirs at HEAD — verify with `git ls-tree -r --name-only
   HEAD -- bugs_open/ bugs_closed/ | grep 184` returning exactly the ONE new
   path). Consider whether this defect class earns a 016b §9 entry (a
   transferable pattern: an LLM writer's markdown syntax leaking into a
   plain-text render surface with zero escaping) — not yet written.
5. No `WRONG_CALLS.md` entry needed from this session: the one design
   correction made (the retraction query's shape, initially copied a
   hand-rolled `status NOT IN (...)` filter from the fable-model plan, caught
   by reading `check_required_fields_missing`'s own header warning against
   exactly that) was caught and fixed **before** committing, not after —
   doesn't meet the bar of "a claim that turned out to be false after being
   acted on".

**Handoff note**: this bug's code/build/migration work is DONE; what remains
is dispatch-and-verify, which needs fresh research into the live dispatch
mechanism (item 1 above) rather than more of what's already been established
here. A fresh session can start directly at item 1 without re-deriving
anything above.
