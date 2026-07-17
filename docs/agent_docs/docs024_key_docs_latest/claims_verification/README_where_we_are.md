Git (commit-per-task): A fresh git status showed the multi-session reality the file warns about had already played out — all my claims-verification code had been swept into concurrent sessions' commits (core under
87d13b864 "claims verification further development", the rest under d076c3c8e/f51a7accc). Forward-only, nothing lost — and it had a silver lining: it's why the deployed image already carries my final code including
the precision fix. The one remaining piece of my task (the deployment-verification notes entry) I committed narrowly by pathspec as ab0b1e452 — HEAD moved again seconds before, and the pathspec commit excluded
everything else exactly as the rule intends. I did not touch SPEC_voice_tells_check.md / datahelpers/voicetells.go — that's another session's in-flight WIP in the same directory (looks like a sibling check being
developed; V3-adjacent territory).

Production verification (completing the deploy follow-up):
- Verified against the running pod, never git: strings /app/agent-chassis contains all the claims symbols — the V1a gate is live for any evidence-base site.
- unverified_claims is enabled on quality-discovery-agent (alongside its sibling placeholder_contact); workflows are re-read per message so it took effect immediately.
- Two production runs on leopardess: the improvement-loop run executed the check (40 log mentions in the discovery pod) with zero claims_unverified items — the correct result on the cleaned site; the loop's later
  failure was a pre-existing call_site_review timeout, downstream of and unrelated to discovery. A second single-agent discovery run COMPLETED with zero items of any type. One behaviour worth knowing: a clean check
  leaves no positive trace by design (checks emit findings only), so "ran clean" is evidenced by COMPLETED + zero items — now recorded in the notes.

V1 is done end to end: evidence base live → gate in the build path → post-deploy check in the discovery cycle → HITL routing → clean production baseline. Next phases when you want them: V2 (writer whitelist
injection — the direct answer to the spec-instructed leaks), V3 (LLM claims auditor for prose assertions), V4 (freshness), and vetcomparison as the second evidence-base site.

----