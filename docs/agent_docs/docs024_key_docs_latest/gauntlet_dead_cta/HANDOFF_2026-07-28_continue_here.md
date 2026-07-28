# HANDOFF — gauntlet_dead_cta, 2026-07-28

**Supersedes `HANDOFF_2026-07-26b_after_p4_live.md`.** Keep that one only for the
API contract (§6) and its landmine list; its step list is done.

## 0. Read this first

Multiple sessions share this repo and cluster. Commit per task by explicit
pathspec, forward-only. **Re-verify anything load-bearing before planning around
it** — this file's predecessors had premises that decayed within 90 minutes, and
one of them (§9 of `bugs_open/083`, "wait 24–48h for real traffic") was falsified
by the very logging it recommended.

**A chassis roll does NOT touch this workstream's engine.** Verified across
v1.0.1172, 1174, 1175, 1180 and 1182. `tools-api` runs under docker compose on an
island VM; the cluster is deliberately absent from its network.

## 1. State in one paragraph

**vonc.com is honest and works.** The Gauntlet plays end to end against a real AI
opponent and now survives a page reload. The Arena's fabricated community is gone.
`claimscan` returns 0 findings across 49 components. The engine's failures are
diagnosable for the first time and **nothing has failed in ~49 calls over three
days**. The site was restyled today at the owner's request. The one thing nobody
can fix by engineering: **the Gauntlet has had no visitors** — every request in
its log is ours.

## 2. What is live (all verified on the served artefact, not the DB row)

| piece | state |
|---|---|
| Gauntlet page + engine | LIVE, `tools-api:v1.0.1178` on the island |
| **Round survives a reload** | LIVE today — sessionStorage resume, verified desktop + mobile |
| Arena | LIVE, scoped down, zero fabrication |
| Provocations archive + feed | LIVE, deep links, no invented stats |
| vonc.com restyle | LIVE — text +25%, content +25%, purple split for contrast |
| `bugs_open/083` fix | LIVE and council-APPROVED (`e004fd81`), trailer on `9474e6b68` |

## 3. Next actions, highest value first

1. **`bugs_open/083` — DO NOT speculatively fix. Wait for a real fault.**
   The log is armed and empty. Candidate 3 (retry) is unjustified by evidence;
   candidate 4 (`max_tokens`) is **refuted so far** — the TRUNCATED branch is live
   and has never fired, exactly as §2 predicted. **If it still has not fired after
   the next burst, close candidate 4 as refuted rather than leaving it open.**
   Trigger, not a date:
   ```bash
   ssh root@toolsapisuk.vs.mythic-beasts.com \
     'cd /opt/island && docker compose logs tools-api | grep -E "gauntlet/(round|position|defend): "'
   ```
2. **Candidate 2 (HTTP client timeout) is FLEET-WIDE, not island-only.**
   `&http.Client{}` at `platform/aiservice/anthropic.go:63`, referenced by 17 Go
   files. A genuine latent defect for every agent. Argue it on fleet grounds with
   its own council round — do not slip it in as a fix for a burst that stopped.
3. **P5 acceptance is still blocked on the harness, not the page.**
   `browserrunner/run_checks_action.go:199` waits `stepDelay = 300ms`; the AI calls
   take 8–23s, so the plan's own checks would FAIL a correct page. Fix the harness.
   **Never** make the page paint placeholder text — that would pass with the engine
   switched off.
4. **`bugs_open/103`** — 16 tool pages on 6 sites publish their build brief as the
   public meta description. vonc's is corrected; the code fix and the other 15 are
   open. Root cause cited in the file.
5. **The Gauntlet has no visitors.** Owner decision, not an engineering task.

## 4. Landmines earned the hard way (the RUNBOOK has the full list)

- **`083` is an AMBIGUOUS NUMBER.** A different open case shares it, and *almost
  every commit message saying "083" refers to that one*. Resolve by SLUG; run
  `git log` against the FILE PATH, not the number.
- **Before editing any `content_components` row, count its `page_components`.**
  `hero` is rendered on **182 pages across the fleet**. Nearly moved 181 other
  people's pages to widen one homepage.
- **A component's inline `<style>` beats a linked stylesheet on ORDER.** Override
  on SPECIFICITY, and *measure that it applied* — `body .pc-container` (0,1,1)
  loses to `.provocation-card-section .pc-container` (0,2,0); class count is
  compared before element count.
- **`rem` resolves against the ROOT.** `font-size` was on `body`, so changing
  `--font-size-base` alone moves almost nothing.
- **One colour cannot serve white-text-on-it AND text-on-dark.** White needs
  L≤0.183, a dark page needs L≥0.195. Split the variable; measure, don't eyeball.
- **The binary is at `/tools-api`, not `/app/tools-api`.** My first runbook draft
  had `/app/` and every check would have read as a failed deploy.
- **`Fingerprint` is a FALSE-POSITIVE chassis pod-grep marker** (21 unrelated
  hits). `TopLevelJSONObjects` discriminates. Always pair with a POSITIVE control.
- **A missing orchestration row: check CONSUMER LAG, not elapsed time.** Non-zero
  lag with a consumer attached = queued, do not re-fire.
- **`check_endpoint_health` COMPLETED rows say nothing about the generic lane** —
  it is a ~90s cron on its own lane. Read what the rows ARE, not how many.
- **Cloudflare 403s a non-browser UA** (`error code: 1010`) on vonc.com itself,
  not just the API. Send a browser User-Agent from every script, curl included.

## 5. The rail to protect

Nothing on vonc.com claims a number that is not true by construction, and no
control changes state except as the consequence of a real API response.
`input_schema.llm_guidance` says so at every field — that is the only durable
defence against a re-render reintroducing a win-rate or a leaderboard. **If one
reappears, look there first.**

**One deliberate exception, so nobody "fixes" it as a regression:** the Gauntlet
now writes to `sessionStorage`. That is NOT the fabrication pattern deleted from
the Arena — that faked a submission which went nowhere. This resumes a REAL
server-side round by its real id, and stores nothing that is not already true.
The reasoning is written at the code.

## 6. Where things are

- Sources, harnesses, backups: `p4_sources/` (all committed)
- Commands + every gotcha: `RUNBOOK_gauntlet_dead_cta.md` §§8–10
- Island: `infra/island/RUNBOOK_island.md` (incl. verify-against-the-running-container)
- Latest read-out: `SUMMARY_2026-07-27b_gauntlet_dead_cta.md`
- Owner's plain-prose log: `README_where_we_are.md` (append only, newest at bottom)

## 7. Open question for the owner

Whether the snippet decision needs revisiting is **closed** (fingerprint shipped).
What is open: **this site works and has no audience.** Everything since 07-22 has
been about making it honest and reliable, which was the right order. No further
engineering produces a visitor.
