# CONTRIB 2026-08-05 — 198's fix candidate 1 is LIVE; your next css-patch dispatch is safe to fire

From a session the owner pointed directly at `bugs_open/198` (with a note that Gemini
is currently unavailable — measured irrelevant to this chain, see below). Your handoff's
next-session line — "198's fix candidate 1 before any css-patch dispatch, then A2" —
the first half is done; this note is so you don't redo it.

## What is live right now (config, applied + recorded 2026-08-05)

Migration `sql_for_agents/318_css_patch_agent_appends_never_round_trips.sql`
(commit `c48c773c1`, `Council-Submitted: 5249320e-1d6e-4bb4-94d3-89bd802bd8a4`,
snapshot `661a27b9` taken pre-update):

- `plan_css_fix` → returns ONLY `css_added`; prompt instructs cascade-override patches
  and (your 07-30 landmine) forbids `var(--x)` names not defined in the loaded sheet.
- `save_css_to_db` → server-side APPEND with dated provenance comment; shrink is
  unrepresentable (SQL concatenation); 1..8192-char guard refuses a whole-document echo.
- NEW `check_saved` → zero-row refusal fails LOUD to `complete_error` (would otherwise
  ride git_commit's "no files → skipped, Success:true" path).
- `deploy_css` → commits `css_saved.css_content`, the row the UPDATE returned; DB and
  repo cannot diverge through this workflow.

Proven on the real relojistas theme row in a rolled-back transaction: 26,152 → 26,240
chars, v6 → v7, commit_msg composed; 9,000-char oversize probe matched 0 rows. The live
row is untouched at v6 (your restore, preserved).

## What is committed but INERT until the next chassis roll

`commit_message_field` (opt-in) on `GitCommitAction` fixes the `CSS fix: <no value>`
audit-trail defect — the template context is a fixed `{domain, file_count, filename}`
map, so the message now composes in the save step's RETURNING instead. Until the roll,
commits fall back to `CSS patch: {{.filename}} ({{.domain}})`. Registered DGH-007;
LANDMINES has the template-context trap; four tests, field-wins proven by mutation.

## What this deliberately did NOT do

- **No dispatch fired.** The witnessed end-to-end run (new finding → promote → append →
  next audit stops re-filing) is yours to sequence — attribution on a mid-change site
  was your stated reason to order these carefully, and the four incident pairings are
  already fixed in v6 so the next audit should file nothing for them.
- **198 left OPEN** — its own verify bar (the witnessed run) is unmet; closure call is
  yours.
- Candidate 2 as a *generic* fleet-wide shrink guard on the writers was not built: on
  this workflow it is subsumed (append cannot shrink; deploy writes the append's own
  return value). The 012 class elsewhere remains 012's.

## Gemini (the owner's note, measured)

`llm_call_log`, 7 days: css-patch-agent = anthropic/claude-sonnet-4-6 only;
render-audit-agent logs no gemini rows; the fleet's gemini traffic is
page-content-writer's `generate_content` steps (last call 11:05Z today). A Gemini
outage does not block the 198 chain — it will bite page rerenders/content generation
instead, which matters to your cascade work, not to this fix.

## Council

Verdict pending at write time — correlation `5249320e-1d6e-4bb4-94d3-89bd802bd8a4`.
Find the run by payload, not by printed id:
`SELECT current_step, status FROM orchestration_states WHERE
 collected_data->'input_data'->>'fix_correlation_id' = '5249320e-1d6e-4bb4-94d3-89bd802bd8a4';`
If it comes back REVISE/REJECTED, the objections land on whoever reads this first —
the code is already on the shared branch (committing-is-shipping).
