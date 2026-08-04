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
