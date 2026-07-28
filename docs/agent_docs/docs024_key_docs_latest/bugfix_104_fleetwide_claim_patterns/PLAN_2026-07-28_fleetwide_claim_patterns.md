# PLAN — `bugs_open/104`, fleet-wide banned-claim patterns

**Opened** 2026-07-28 (session "bugsearch 6"), picking up `104` as the natural
follow-on named in `HANDOFF_2026-07-28_bugsearch2_session.md` §4.

## What this workstream is for

`104` says the claims engine is general and sound but pointed at almost nothing:
`banned_claims` is keyed per site, so a lesson learned on one site cannot reach
another and every new site is born unarmed. Its recommended fix (candidate 1) is
to mirror `globalTellPhrases()` from `voicetells.go` into the claims parser — a
universal set unioned with each site's own.

The same decision is also **oufe decision O11**
(`oufe/DECISIONS_2026-07-26_oufe.md:231`), where it is explicitly routed to the
owner *because it changes behaviour beyond one site*. So this workstream does not
own the decision. It owns the measurement the decision needs.

## The one thing that had to be measured first

`CLAUDE.md` § "Platform seams and the ordering exemption": *"Measure the
blast-radius claim before you submit; do not ask the reviewer to."* `104` itself
asks for a dry-run count, but attaches it only to candidate 2. That was the wrong
place to attach it — see below.

The measurement is not a new tool. `cmd/claimscan` already runs **the same shared
engine** as the deploy gate (`validate_page_content` check 8) and the post-deploy
audit, over exported component HTML. Commands in the RUNBOOK.

## What the measurement found (2026-07-28, evidence in NOTES)

Running the 10 tested patterns from `sql_for_agents/226` against the stored
`rendered_html` of **all 15 live sites (908 components)**:

- **7 findings on 3 sites / 6 surfaces** — leopardessconsulting (1),
  robot-hands (4), vonc (2).
- **4 of the 7 are false positives**, and all four fire on a *negated* sentence:
  "Where manufacturer data has **not** been independently verified, that is
  stated explicitly"; "When a figure **cannot** be independently verified, it is
  marked as unverified"; "are Spark's own assessment, **not** independently
  verified" (×2).
- One pattern — `(fully|independently|externally|properly)
  (verified|audited|fact.?checked)` — accounts for **6 of the 7 hits and all 4
  false positives**.
- **Nothing is biting today**: every armed site scores **0** against its own
  register.

So candidate 1 as written would make three live sites' pages unbuildable, and the
majority of what it blocked would be **honest disclosure** — the precise
behaviour this layer exists to encourage. Severity is `blocker`, so a false
positive fails a whole page build.

Two supporting facts that make this certain rather than likely:

1. Candidate 1 is gated on `ParseEvidenceBase` returning non-nil, which needs
   `facts[]` **or** patterns. robot-hands has `facts=5, banned=0` and
   gamesdesign `facts=4, banned=0` — both non-nil. **Candidate 1 does reach the
   sites where the false positives are.** The residual in `104` ("still gated by
   `eb != nil`, so sites with no register remain uncovered") is true, but it
   quietly implies the gate keeps unarmed sites *out*, and for these two it does
   not.
2. There is **no negation-guard prior art anywhere in the estate**, and Go's RE2
   has no lookbehind — so this cannot be fixed inside the pattern string. It
   needs code, in the shape of `isExcludedNumber`.

## The design point this exposes

`ScanBannedClaims` has **no** false-positive apparatus, and that is deliberate
and documented: *"Every match is a KNOWN falsehood for this site (each pattern
was audited out by a human) — callers treat findings as blockers"*
(`claims.go:439-441`). Its sibling `ScanUnregisteredNumbers` has an elaborate one
(`businessClaimContextRe`, `isExcludedNumber`, written-date and unit exclusions)
because it is *not* human-audited, and its own comment says why: *"Noise is not
harmless in a checker: a scanner that always reports something is one people stop
reading."*

**Making the patterns fleet-wide removes the premise that justified the absence,
while keeping blocker severity.** Nobody audited the oufe set against the other
fourteen sites' copy — this session is the first time anyone did, and it found
four false positives in the first ten patterns. That is the real content of the
O11 decision, and it was not visible in either the bug file or the decision doc.

## Phasing

1. **Measure** — done, 2026-07-28. Fleet dry run + per-site self-scan + a
   positive control in both directions.
2. **Record** — this dir, plus a triage section appended to `bugs_open/104` (the
   shared account; `who-owns.py` says contribute into it, do not fork), plus the
   transferable pattern in `016b` §9.
3. **Put the owner call, costed** — O11 is the owner's. It now has numbers, and
   the recommendation has changed: candidate 1 is *not* "small, precedented, one
   roll" until the pattern set is negation-safe.
4. **Only then build.** Whatever is chosen touches a shared scanner used by the
   build gate on every site, so it is architecture-scope by the 2026-07-28
   ruling: council round, and registered in the concept register in the commit
   that ships it.

## Decisions and corrections

- **CORRECTION to `104` § "Fix candidates" (this session).** The dry-run count is
  attached to candidate 2 only, on the reasoning that candidate 2 "changes what
  every site's build gate can block" whereas candidate 1 is contained. Measured,
  candidate 1 changes what **9 of 15** sites' gates can block (7 with patterns +
  robot-hands and gamesdesign, which have facts and therefore parse non-nil), and
  it is where the false positives land. **Both candidates needed the count.**
- **Not filed as a new bug.** The false-positive class is a property of the
  *proposed* fix, not a live defect — nothing bites today. It belongs in `104`.
