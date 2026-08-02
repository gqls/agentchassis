# HANDOFF — 2026-08-03 — `bugfix_136_sibling_link_repair` lane · continue here

**This supersedes `HANDOFF_2026-08-02_continue_here.md`.** Read this file, then
`bugs_closed/180`. Everything below is committed; this lane leaves nothing dirty in the tree.

---

## 1. State in one paragraph

**`bugs_open/180` is CLOSED, LIVE on `v1.0.1233`, pod-verified on both replicas, and INDUCED
on the page it was damaging** — the anchor it had deleted is back on the wire. It is in
`bugs_closed/`, registered as **LNK-029**. **The block the last handoff shouted about is
GONE:** wiring the repair into the tool-markup writers was unsafe *because of 180*, and 180 is
fixed. That is now the ranked next step, as originally planned.

The one thing genuinely outstanding is a **council verdict nobody has read** — see §4.

## 2. What shipped (live on v1.0.1233)

| thing | where |
|---|---|
| `NonMarkupSpans` — the byte ranges a browser never parses as markup (raw-text elements + comments, whole) | `platform/orchestration/datahelpers/markup_spans.go` |
| `MarkupMatches` / `ReplaceAllInMarkup` — drop-ins for `FindAllStringSubmatchIndex` / `ReplaceAllString` for any regex that REWRITES markup | same file |
| `scanSpans` — ONE tag walk answering both span questions; `RuntimeFillSpans` becomes a caller | `runtime_fill.go` |
| wired into `RepairPageLinks` (one line — fixes all three live callers) and `DropDeadURLControls` | `link_repair.go`, `drop_dead_url_controls.go` |

Commits: `07576d4e1` (fix + register), `c734dbc98` (docs + 016b + WRONG_CALLS + LANDMINES),
`414a3ecfd` (ticket + summary), `c06d7c817` (close + induction SQL).

## 3. ⚠ USE THE HELPER — do not re-derive it, and do not hand-roll a span skip

If you write a regex that **rewrites** HTML, call `datahelpers.MarkupMatches` /
`ReplaceAllInMarkup`. Two traps are already paid for, and both are in `markup_spans.go`'s
header, `LANDMINES.md` and 016b §9:

1. **Mask before matching; do not filter matches afterwards.** With a decoy anchor inside a
   `<script>` followed by a real phantom, the non-greedy `</a>` makes ONE match span both —
   dropping it drops the genuine defect too, and `FindAll` never revisits those bytes.
2. **The filler byte is NUL, not a space.** Whitespace MANUFACTURES matches for a pattern that
   opens with `\s`, and that match begins OUTSIDE the span where the offset filter cannot see it.

## 4. What is still open, ranked

1. **READ THE COUNCIL VERDICT — it is owed and unread.** `Council-Submitted:
   ba199c35-516f-44be-a210-9fd982425eb7`. The first run **stalled at `review_constitution`,
   `updated_at` frozen 21:38:10, one minute before the `v1.0.1231` roll replaced the pods**
   (a roll kills an in-flight council). Resubmitted on the same trail as run
   `612bc0f9-de7e-4627-a712-a3b226694677`. **Act on a REVISE — the code is already live.**
   ```sql
   SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
   WHERE correlation_id='ba199c35-516f-44be-a210-9fd982425eb7' AND kind='council_report' ORDER BY created_at;
   ```
   Then `Council-Reviewed:` may be written on a LATER commit — never retro-fitted, never on an
   unread verdict.
2. **The tool-markup writers — now UNBLOCKED.** `create_tool_component_action.go` and
   `deploy_tool_action.go` should receive `repairComponentHTMLBeforePersist` (LNK-027). 7 of
   the 35 census hrefs sit in tool-shaped slots, and they are deliberately NOT allow-listed in
   `check_unrepaired_component_write`, so it keeps firing on them. Do not silence it.
   ⚠ **Both files were being edited by another session on 2026-08-02** (unrelated hunks, no
   link-repair changes) — re-check ownership before starting.
3. **The DETECTORS still have 180's blind spot, deliberately.** `links.go`'s phantom scan,
   `check_dead_controls`, `check_phantom_internal_links` read a JS-built anchor as a finding.
   A false finding costs attention, a false repair costs content — different fail-safe
   directions, and narrowing a judge is `bugs_open/137`'s separate ruling. Belongs to the
   detection lane (`bugs_open/097`, `bugs_open/116`); adoption is one line each.
4. **`architecture_review/RFC_008`** — the mandatory write seam. Unchanged: four seats
   converged on "advisory is the wrong ceiling", and the measurement that would settle it
   (does anyone read advisory `pattern-check` findings?) has still never been taken.
5. **The standing stock is untouched** — 18 unlinkable hrefs are live 404s today. Detection is
   `bugs_open/116`.

## 5. Verification recipes

- **RUNBOOK §8** — measure a markup writer's blast radius by RUNNING it over the real corpus.
  Assemble by `position` (the live caller sees the ASSEMBLED page), dump **per site** (a
  whole-fleet `COPY` truncated at ~2.8MB and exited 1), and print three numbers — of which
  **"legit matches LOST" is the one that matters** and must be 0.
- **RUNBOOK §9** — proving a guard by breaking it, and what to do when the mutation PASSES.
- **§7 the pod-grep triple**, plus the lesson from this round: **prefer a control that MOVED**
  over an invented absent string. `RuntimeFillSpans` 2 → 1 discriminated because the refactor
  caused it; a newly-added symbol cannot do that.

## 6. Things this round got wrong, so you do not repeat them

1. **A mutation passed that I had already written a comment predicting would fail.** Two
   guards in SERIES; the stronger absorbed the weaker's mutation. The previous session's
   handoff had recorded that exact trap and reading it did not stop me.
   `WRONG_CALLS.md` 2026-08-02. **Never write "change X and this fails" until you have.**
2. **A `complete` work item plus an unchanged page looked exactly like a failed fix.** It was
   Cloudflare (`max-age=3600`). Check `last-modified`/`age` before concluding anything from a
   fetch.
3. **The first post-mutation replacement test stopped discriminating for a THIRD reason** —
   the scanner swallowed the `<style>` into an enclosing tag, so nothing was masked and the
   test passed while measuring nothing. Tests that depend on a precondition should assert it.

## 7. Where everything lives

- Standing five: this directory (PLAN, RUNBOOK, NOTES, README_where_we_are,
  `SUMMARY_2026-08-03_the_repair_learns_what_is_not_markup.md`, and this handoff)
- Closed cases: `bugs_closed/136_…section_editor…`, `bugs_closed/180_…javascript…`
- Register: **LNK-027** and **LNK-029** in `docs026_concept_register/register/link-management.md`
- Landmine: "Any regex you write against HTML will also match inside script/style/…"
- 016b §9: the 180 entry, with the FIX's two lessons appended
