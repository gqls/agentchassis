# Paired provocation — prototype

Owner asked for this on 2026-07-31: *"a person chooses a set of people that will
be part of the provocation and they set their own provocation (the person in
charge) and the team get to reply that day or over several days until they've all
committed (choices given to the person setting it up). The results for these ones
probably wouldn't be published publicly but distributed to the team directly."*

```bash
cd docs/agent_docs/docs024_key_docs_latest/provocation_pipeline/prototype
go test ./...     # 14 tests, all of which try to break the seal
go run .          # http://localhost:8099
```

Open the create page, make a session, then open each participant's link in a
different browser tab (or a private window) and play both sides.

---

## What this is for

To feel the shape and to force the design questions into the open before anyone
builds it properly. It is **not** a step towards production — see "What it
deliberately fakes" below.

It is a nested Go module on purpose. A nested `go.mod` is excluded from the root
module's `./...`, so this throwaway cannot break `go build ./...` on a shared HEAD
that fourteen services build from. Four other prototypes under `docs/` already do
this.

## What it demonstrates, and what that cost

**The seal is enforced by the type system, not by a check.** A pre-reveal response
is a `SealedView`, whose only description of another participant is
`SealedPeer{Name, Committed}` — a type with nowhere to put their words. A
post-reveal response is a separate `RevealedView`. `ViewFor` returns exactly one of
them. The handler cannot leak a position by forgetting a condition, because it is
never handed one.

This matters more here than anywhere in the public product. If Alice can read Bob's
position before committing hers, the exercise stops measuring what the team thinks
and starts measuring who read the room first — which is the exact failure a
team-deliberation tool exists to prevent.

**Four decisions the prototype had to take, all of them arguable:**

| decision | what it does | why |
|---|---|---|
| **The organiser cannot read positions either** | organiser page shows who has answered, never what | the most tempting feature in the design, and it would destroy the product. A facilitator who has read the answers cannot run the session neutrally, and participants who suspect they have will hedge |
| **Non-responders do not receive the reveal** | stay silent, and you do not get to read the room | without it, the optimal play under a deadline rule is to say nothing and wait — the seal's own failure mode, re-entering through the back door |
| **Commit is final** | no edits once committed | "until they've all committed" needs *committed* to be a state you cannot leave, or the reveal condition keeps un-satisfying itself. Sealed-bid auctions do allow revision, so this is genuinely arguable — flagged rather than silently decided |
| **Reveal is one assignment to one session field** | everyone opens at the same instant | if it opened per participant, whoever polled last would read everyone else's answer before writing their own |

**Three reveal rules, which are the owner's "choices given to the person setting it
up":** when everyone has committed; at a quorum of N; at a deadline. Plus a
`Force reveal now` button, the escape hatch for the non-responder who never will.

## The tests try to break it

`paired_test.go` is written to *break* the seal, not to demonstrate it working —
a comment claiming the seal holds is worth nothing.

**They were mutation-tested**, because a passing test can pass for the wrong
reason:

| mutation | caught? |
|---|---|
| give `SealedPeer` a `Position` field and populate it | **yes** — 3 tests fail |
| let non-responders receive the reveal | **yes** — 2 tests fail |
| stamp `RevealedAt` per read instead of once per session | **no, at first** |

The third one is the useful entry. `TestRevealIsAtomic` originally read all three
participants' views at the *same* `now`, so a per-read timestamp and a session-wide
one produced identical output and the test could not tell them apart. Fixed by
reading at three different times and asserting the stamp equals the moment the last
person committed. **The test was green against a broken implementation until the
mutation exposed it** — which is the whole argument for mutation-testing a guard
rather than trusting that it passed.

## Verified end to end, not just built

Driven against the running server, 2026-07-31:

1. create a 3-person session → organiser page lists three distinct participant links
2. Alice and Bob commit → **Carol's page contains 0 occurrences** of either of their
   positions, while correctly showing "2 of 3 committed"
3. **the organiser's page contains 0 occurrences** of either position
4. Carol commits → all three participants' pages show all three positions

## What it deliberately fakes

Read this before treating any of it as a design.

- **Nothing is persisted.** Restart the process and every session is gone. That is
  a feature, not a shortcut: the moment we store named colleagues' opinions on
  contested topics we have taken on a real confidentiality duty, and that should
  begin with a deliberate design, not with a prototype quietly filling a table.
- **A token in a URL is not access control.** It is the *only* thing standing
  between a stranger and a private session here. Unguessable is not the same as
  authenticated: links leak into chat logs, browser history, and shoulder-surfing.
  Real identity is the single biggest prerequisite for building this properly, and
  the platform currently has none at all — the public Gauntlet keys rounds on a
  hash of the client IP, which `bugs_open/139` measured as a *constant* across all
  83 rows, so it has never distinguished anybody.
- **There is no AI verdict.** Left out on purpose. An AI ruling on a named
  colleague's argument, circulated to their team, is a performance review nobody
  agreed to — worth deciding deliberately (share positions with the group, keep
  each verdict private to its author?) rather than inheriting from the public
  product.
- **No email, no invitations, no reminders.** The organiser copies links by hand.
- **No rate limiting, no CSRF, no audit log.** Prototype.

## Open questions this raised

1. Should a participant be able to revise before the reveal? (Sealed-bid auctions
   say yes; "all committed" as a reveal condition says no.)
2. Should the organiser be a participant too? Currently they are not — they set
   the provocation and never argue it.
3. What does the team *receive*? Right now it is a web page per person. The owner
   said "distributed to the team directly", which sounds like email, and email
   changes the privacy calculus again.
4. Does a paired provocation ever become public, with consent? That is the bridge
   between this and the public daily — and it is how paired mode would feed the
   engagement data that makes "interesting" and "relevant" measurable.
