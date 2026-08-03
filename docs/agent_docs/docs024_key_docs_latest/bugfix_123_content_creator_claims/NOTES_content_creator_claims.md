# NOTES — bugs_open/123, content-creator's copy cannot reach the claims assessor

Append-only, newest at the bottom. Technical log: evidence, commands, what the
system actually said, and every misstep.

---

## 2026-08-03 — session start: bug selection, and why 123

Ran the sweep the owner asked for: pick the next `bugs_open/` case no other thread
is working. Two other sessions are running the **identical** prompt right now
(their first user message is byte-similar to mine), so "unowned" had to be
established against live sessions, not only against git.

**How ownership was actually measured**, because `scripts/who-owns.py` alone was
not enough:

1. `who-owns.py` returned **OWNED or recently active** for *every* candidate I
   tried (033, 066, 071, 084, 085, 093, 096, 107, 113–116, 120–123). Its verdict
   fires on any commit touching the bug file in 14 days, and a triage sweep on
   2026-07-27 touched most of them at once. **[MISSTEP, mine]** I ran it 16 times
   before reading its source and noticing the verdict could not discriminate.
   The useful part of its output is the *owning workstream* block, not the verdict
   line.
2. `git log` per bug file + commits whose SUBJECT names the number in the last 4
   days. That separates "was swept" from "is being worked".
3. **Live session transcripts**, because every git-based check is lagging by
   construction (`~/.claude/projects/-home-ant-projects-agentchassis/*.jsonl`,
   modified in the last 6h). Counting `bugs_open/NNN` per transcript names what
   each live session is actually on: 182, 091, 149, 010, 183, 098, 175, 117, 151,
   168, 139, 172, 140, 153, 115, 178, 138, 155, 159, 157, 128, 150.
   **123 appears in none of them as a subject.**

**[MISSTEP, mine — the measurement answered the question I encoded]** My first
validity check on a different candidate (`bugs_open/116`, "the link-integrity
checks have never run") queried `site_work_items` for item types
`phantom_internal_links`, `dead_controls`, `misdirected_cta` and got zero rows,
which looked like confirmation. Those are **check names**. The **item types** are
singular — `phantom_internal_link` (25 rows, latest 2026-08-03), `dead_control`
(2), `cta_names_unknown_destination` (40). 116's central claim is **stale**: the
agent runs and the checks do produce findings. The cheap check I skipped: list
the distinct `item_type` values the agent actually produced before asserting an
absence, instead of filtering on names I had read in prose.

```sql
-- the query that corrected it
SELECT item_type, count(*), max(created_at)::date
FROM site_work_items WHERE created_by='completeness-discovery-agent'
GROUP BY 1 ORDER BY 2 DESC;
```

### 123 re-verified against the live system before taking it on

| claim in the bug file | check run 2026-08-03 | verdict |
|---|---|---|
| content-creator has no validation of any kind | `grep -rniE "validate\|claims\|evidence\|fabricat\|banned" internal/agents/contentcreator/` → **0 matches** (828 lines, one file) | STILL TRUE |
| it writes nothing to `llm_call_log`, so no per-agent usage query sees it | 14-day `llm_call_log` group-by `agent_type ILIKE '%content%'` → `content-quality-auditor`, `page-content-writer`, `content-reviewer`, `content-gap-planner`. **No `content-creator`** | STILL TRUE |
| the claims machinery cannot be pointed at site-less text | `save_sections_claims_guard.go:81` says so in terms: *"bugs_open/123's content-creator path has no site and no page row, so this seam cannot reach it at all"* | STILL TRUE, and now stated by the platform itself |
| "which workflows call content-creator is untraced" | **NOW MEASURED** — see the correction below | CORRECTED |

### CORRECTION to 123's severity, measured not assumed

123 is filed **HIGH** ("HIGH if any content-creator output is published") and asks
for an owner call "before blog or social output is published anywhere". The
untraced half is now traced, and it lowers the severity:

```sql
SELECT type, is_active, COALESCE(is_snapshot,false) snap, deleted_at IS NOT NULL del
FROM agent_definitions WHERE default_config::text ILIKE '%content-creator%';
--  website-builder           | f | f | t
--  multipage-website-builder | f | f | t
--  multipage-website-builder | f | f | t
```

**All three rows that reference it are deleted and inactive.** No live agent
definition dispatches content-creator. It is reachable only by a direct Kafka
publish to `system.agent.content-creator.requests` — which is exactly how the
2026-07-27 fabrication was produced.

So the honest reading is **latent, not biting** — the same shape as
`bugs_open/134`. It is still worth fixing and the fix is still cheap, but the
"owner call before publishing" framing overstates what is live today. The service
itself is deployed and running (`content-creator-agent-8576b699d4-rgmq2`,
1/1 Running), so a guard placed in it is not a mechanism rotting unexercised — it
is on the path any direct dispatch takes.

### What the platform has grown since 123 was filed (2026-07-27 → 2026-08-03)

This is the part that changes the fix, and it is why the case is worth taking now
rather than when it was written:

- `datahelpers.ScanAllBannedClaims(blocks, eb)` (`claims_global.go:248`) is
  **nil-safe by design** — *"eb may be nil, and a nil eb means 'this site has no
  register', not 'do not scan'"*. That is precisely the site-less entry point
  123's fix candidate 1 asks someone to build. It already exists, shipped by
  `bugs_open/104`'s lane on 2026-07-28, the day AFTER 123 was filed.
- The **claims floor** (`save_sections_claims_guard.go`, `bugs_open/149` C1)
  established the estate's severity precedent: a banned claim REFUSES the write; an
  unregistered number is RECORDED and allowed, because "number extraction has false
  positives by design".
- `platform/voicestyle/` is the working precedent for one shared package consumed
  by both the chassis (`*sql.DB`) and content-creator (`pgxpool.Pool`).

So the fix is now mostly *wiring*, not building — which is the outcome CLAUDE.md's
"reuse existing machinery before building new" is aiming at.

**Still an open gap, and it is the half that matches the observed damage:** the
fleet-wide banned set is about **self-accuracy overclaims** ("guaranteed accurate",
"independently verified", "you can rely on us"). Read all ten patterns at
`claims_global.go:111-215`. **None of them matches "Industry data shows … between
3% and 10%"** — an attributed-but-uncited statistic. That is 123's fix candidate 3
and it needs a new detector.
