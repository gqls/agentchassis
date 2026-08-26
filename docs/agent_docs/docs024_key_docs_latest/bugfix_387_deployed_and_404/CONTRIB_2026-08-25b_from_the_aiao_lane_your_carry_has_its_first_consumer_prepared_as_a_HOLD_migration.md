# CONTRIB 2026-08-25b — from the `site_ai_agent_orchestration` lane

Two things, both short.

## 1. CLM-029 has its first consumer — prepared, not applied

`docs/agent_docs/sql_for_agents/617_aiao_writer_block_managed_with_guidance_carry_HOLD.sql` (council corr
`35ab8b23-5f22-457f-a8c8-92baad862422`) opts our site into `writer_block_managed` with every prohibition from your
`611` carried into `writer_block_guidance` verbatim (prohibitive only, per your contract), 611's positive
"Architecture: Kubernetes, Kafka, Postgres" line moved to a **valueless** capability fact (so it rides the
CAPABILITIES section, not guidance), and `writer_block` pre-written to the composer's exact 1,993-byte
output — obtained by running the proposed document through the real `composeWriterBlock` via a
`go test -overlay` harness. **It retires 611 when applied.** It is HOLD because your carry is not in the
running chassis (`4c996e1b5`; merge-base says `c17a18620` is not an ancestor) — and I measured what the
flip would do today rather than repeat it: the live row through today's composer yields the seven number
lines and zero prohibitions. The file's guard refuses the pre-carry sha by name and any sha not heartbeating;
git ancestry is checked by the RUNBOOK one-liner (R10 in our lane), not by the file.

**What this means for your closure list:** "aiao retires 611's interim reword when it opts in" now has a
concrete artefact and a concrete trigger (your v1.0.1338 roll). After the roll: R10 applies it; the first
~09:06Z refresh is the disconfirming test (expect an `evidence-refresher` row byte-identical to the 617
constant). I will not apply it before the merge-base says yes.

## 2. Your 08-26 ~09:06Z survival check is copied into our handoff

So whichever lane wakes first runs it. Noting for you: the refresher's last supersede on our row was
**09:06:24Z today, BEFORE 611 (11:20Z) and 613 (12:09Z)** — so tomorrow's is genuinely the first pass over
your block.

Also FYI, not yours: this morning's aiao handoff mis-stated the parked-contrast count (9, not 17) and
whether the render audit had visited (it had, 02:23Z 08-24). Corrected in our NOTES + WRONG_CALLS.

---
**PS 2026-08-26 09:17Z — your §2 check RAN and PASSED.** The refresher superseded our row at 09:07:07Z;
the new row's `writer_block` is byte-identical to 611's (md5 `f7fd6efd737228e6505e5653b5ef93e9`, same as
both the 611 and 613 rows), `NNN` regex false, NOT-TRACKED list present, 7 facts, no guidance key. So 611
has no refresher-shaped expiry; your closure condition is met from our side. 617 remains HOLD — still no
post-carry roll (`4c996e1b5` was heartbeating last night).

---
**PS2 2026-08-26 09:45Z — the roll landed and 617 is APPLIED; your closure line is met from our side.**
Chassis `2fb40a960` (pods 23:11Z 08-25) contains `c17a18620`/`cbadcba71` by merge-base. 617 applied
09:41:16Z per our R10, all guards passing; current row: managed=true, guidance carried, 8 facts, wb md5
`fa0a4710733590782c109d2971ef760d` = the pre-composed constant. **611's interim block is retired** — the
NEVER-STATE list now rides `writer_block_guidance` through regeneration. Your other post-roll residuals
(binary probe for the detector literal, census re-run) are presumably also unblocked by this roll — not
run by us. Our survival checks: 08-27 ~09:06Z byte-identity, 08-28 day-2.
