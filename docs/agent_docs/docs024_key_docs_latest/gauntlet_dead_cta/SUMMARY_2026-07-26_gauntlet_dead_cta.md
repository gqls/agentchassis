# SUMMARY — 2026-07-26 — the backend is live, real, and the build plan is approved

*(New file per the never-overwrite rule; previous read-out: SUMMARY_2026-07-25,
written the morning the machine's pull request opened. Everything below happened
that same day, after the merge.)*

## What we're trying to do

Turn the vonc.com Gauntlet from a decorative mock into a real AI debate
opponent: a visitor files a Position on today's provocation, the AI files a
genuine opposing Position and a challenge, the visitor defends on the clock,
and the AI returns an honest verdict — never simulated, honestly labelled,
degrading honestly if the engine is ever offline. The platform's own
feature-builder was tasked with writing the backend for this — its first-ever
real build — with the owner's review of its pull request as the one hard gate.

## Where we've come from

By the previous morning's read-out, the pull request had just opened: the
machine had produced a complete `tools-api` service — scaffold, Docker image,
kustomize manifests, CORS/rate-limiting middleware, the rounds table, and the
three debate endpoints — after two implementer runs had died mysteriously
mid-build. Those deaths turned out not to be the implementer's fault at all:
a housekeeping cron was deleting the message channels live agents were using,
every ten minutes, because its "is anything running?" check queried a label no
agent has ever carried. Fixed and proven, the implementer completed cleanly on
the very next try.

## What we've done

The owner merged the pull request. Getting the merged code actually running
surfaced a run of genuine, ordinary deploy-time bugs — the kind no amount of
code review catches, only running the thing does:

- The container recipe named the wrong Go version for what the code needed.
- The very first real request revealed that looking up a debate round failed
  every single time on a fresh round — and the failure was silently reported
  to the caller as "round not found," hiding a real defect behind a plausible
  lie.
- A database setup script's safety check wasn't written in valid database
  language at all.
- Cloudflare, sitting in front of the service, was silently replacing our
  honest "the engine is offline" error message with its own bare error page —
  found only by testing from the actual public internet, not from inside the
  cluster.
- Once the owner's real API key was installed, one more bug appeared: the code
  never told itself which environment variable held the key, so every AI call
  failed regardless of whether the key was valid.

Each was found, fixed, and shipped the same day. By mid-afternoon, a complete,
real debate round worked end to end through the public internet: a visitor
files a position, the AI responds with a genuine counter-argument and a
challenge, the visitor defends, and the AI delivers an honest verdict. Two full
rounds were run for real — both judged "opponent wins" against deliberately
thin defences, which is exactly the honest, non-flattering behaviour the
original brief asked for.

With that real evidence in hand, the AI planning council was asked to design
the actual front-end build. Its first attempt ran five genuine rounds of
detailed critique — five different reviewers checking feasibility, honesty,
scope, and data contracts — and hit its round limit without full agreement.
That's a deliberate safety valve, not a failure: rather than loop forever, it
stops and hands back the disagreement for a human (or another pass) to act on.
Reading it, the remaining objections were narrow: we had exact information
the reviewers didn't (the real shape of the API's responses, since we built
and tested it ourselves) plus two small facts about existing code they'd
correctly flagged as unverified. Along the way, that first attempt also
exposed one more genuine platform bug — the review panel's own writing budget
was too small for a large round, causing a verdict to get cut off mid-sentence
— fixed on the spot.

Feeding the missing information back in and running the council once more, it
converged immediately: **the build plan is now formally approved**, with only
minor advisory notes left, none of them blocking.

## Where we are now

A real, live, tested AI debate backend exists on its own dedicated machine,
answering the public internet at `tools.apis.uk`, cleanly separated from the
production cluster. It has been proven end to end with genuine AI-generated
content, not a mock. And for the first time, the platform's own planning
council has fully approved a concrete build plan for turning this into the
actual visitor-facing page — a real milestone for both this specific tool and
for the planning machinery itself, which had never reached full approval on
anything this substantial before.

Two small things remain with the owner: pasting the pubkey for backup access
(done, confirmed working) and creating the dedicated API key (done, confirmed
working) — both already resolved during today's work.

## Where we're going

Next is the actual front-end rebuild: rewriting the gauntlet page against the
now-approved plan, wiring it to the real backend instead of the placeholder
behaviour it has today. After that, a full acceptance run — walking through
the real journey as a visitor would, checking nothing is faked and every
control does something real — closes the loop the owner opened when they
first said the gauntlet didn't work.
