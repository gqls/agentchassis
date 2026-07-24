# SUMMARY — the Gauntlet becomes real: AI debate opponent, machine-built backend (2026-07-24)

## What we're trying to do

Make the vonc.com Gauntlet a genuinely working tool instead of a decorative mock — and
do it in a way that fixes the platform, not just the page. The owner's design: a
**debate opponent**. You read the day's provocation and file your Position; a real AI
opponent files an opposing Position and challenges yours; you defend on the clock; the
AI judges honestly whether your take held up, with reasons. It plainly labels itself an
AI competitor while the site has no human traffic — a real feature, not a simulated
crowd. The backend that powers it is being **built by the platform's own
feature-builder** (its first-ever real build), and the whole experience is governed by
the experience loop, so "the button does what it promises" is checked by machinery,
not by hope.

## Where we've come from

The Gauntlet shipped as a facade: two dead buttons, invented stats, a leaderboard of
made-up names. A first fix (07-22) removed the dead links and fabrications but was
itself a lesson — it wired the buttons to effects nobody could perceive, and the owner
caught it by simply using the page. That correction is on the record (verify what the
*user experiences*, not whether the handler fired) and it reset the direction: build
the real thing. Since then: the experience loop was unstuck (its contracts reviewer
used to block any plan proposing new code; the rule now distinguishes new code —
allowed if the plan pins the exact interface and an acceptance test would fail were it
never built — from contradictions of existing code, which still block hard). The
owner's ruling travelled into the planner's standing instructions, a formal capability
request produced a council-approved six-stage build plan for the new `tools-api`
service, and the never-before-fired implementer began building it.

## What we've done

- **Decisions locked** (all owner's): debate opponent; feature-builder builds the
  backend; engine lives in the cluster; public access via a shared hostname on
  `apis.uk` behind Cloudflare, a tunnel, and a dedicated bastion host; sites stay
  static. Security shell drafted (tunnel config, bastion proxy allowlist, cluster
  network policy) awaiting three owner infra tasks.
- **The implementer's first fires found five real defects, none of them wasted:** a
  token-cap blowout that exposed both a needless whole-makefile rewrite and a missing
  dockerfile in the approved plan; a genuine platform bug — the Go formatter and the
  commit packager disagreed about a data shape, each half unit-tested alone, never run
  together until now — fixed, regression-tested against the real chain, shipped live
  in chassis v1.0.1155 and **approved by the council gate**; a prompt gap that let
  the model invent a module name; a path deviation my own corrective rule had
  accidentally suggested (fixed, plan-paths-are-law); and repeated silent message
  drops on the dispatch bus (the known bug-003 class), now countered by
  fire-with-ingest-confirmation.
- **Every failure was caught deterministically before bad code could land** — no
  truncated file was ever committed, no PR opened on a red gate. The cage held all
  five times.

## Where we are now

The build is mid-flight: the approved six-stage plan stands, the platform fixes it
needed are live, and the implementer is being re-fired with ingest confirmation (the
dispatch bus dropped two fires today — the third attempt is in flight as this is
written). The experience re-plan is deliberately parked: its council refused to
approve a plan that promises an API that doesn't yet answer — the correct order — and
it re-fires with liveness evidence once the backend is deployed. Nothing customer-
facing has changed since 07-22; the live page is honest but still shallow, by design,
until the real engine lands.

## How the council is working

Three separate councils have been exercised hard these two days — the **council gate**
(reviews platform code changes), the **experience-loop council** (five seats judging
experience plans), and the **feature-builder's design council** (judges staged build
plans). Same pattern in all three: independent reviewer seats, converge-or-escalate,
verdicts recorded durably. The honest read-out:

**The judgement is good — repeatedly better than mine.** The design council approved a
sound plan in four minutes, then in the next round caught that two stages silently
assumed a database connection no stage created. The experience council's feasibility
seat **vetoed** the gauntlet plan because it gated the core journey on an API whose
existence nothing could confirm — exactly right, and its prescription (ship the
static steps now, gate the API-dependent step on a live smoke test) is the sequencing
a careful engineer would demand. The honesty seat called the plan "unusually
disciplined" and approved — the owner's no-placeholders directive is now visibly
enforced by machinery. The contracts seat, after our rule fix, did precisely what the
fix intended: it approved every genuinely-new interface that was pinned and
test-backed, and objected only where it could *quote existing source contradicting
the plan* — both objections real. And the council gate approved the formatter fix on
its merits after reading the evidence trail.

**The delivery around the councils is the weak part.** Verdicts are sound; getting a
run to the verdict is flaky: one spawn silently swallowed by the dispatch bus, one
review seat killed by an upstream API timeout (the run filed as "invalid" though
nothing was judged), one revise loop that hung for hours and died without an error.
None of these corrupted a decision — they cost time and re-fires. The pattern to keep:
**treat a missing verdict as infrastructure noise and re-fire; treat an actual verdict
as signal and obey it.** Advisory status is working as designed: nothing blocks a
commit, but approvals are now reachable (~80% platform-wide since the decision-rule
fix), so the norm of submitting platform changes has real teeth, and every verdict in
this workstream is on the durable record with its correlation id.

## Where we're going

1. Implementer completes the six stages → **a pull request lands for the owner's
   review — the hard gate.** Merge, roll the `tools-api` image, apply the rounds
   migration.
2. Owner infra: name the `apis.uk` subdomain, provision the bastion, approve the
   WireGuard peering; then wire the tunnel and network policy and smoke-test from
   outside.
3. Re-fire the experience plan with API-liveness evidence → approved plan → rebuild
   the gauntlet front-end through the sanctioned owned-page path, objectives bound to
   real events.
4. Tier-4 journey acceptance against the live page — the "a person can feel it work"
   bar this workstream once failed, now enforced by machinery.
5. Later, owner-gated: real human leaderboard once there is traffic; more tools onto
   the same backend pattern.
