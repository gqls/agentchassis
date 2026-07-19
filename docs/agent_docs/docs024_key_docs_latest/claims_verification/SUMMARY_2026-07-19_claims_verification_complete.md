# Claims Verification — the complete story, to be read aloud

*2026-07-19. The current summary; supersedes the 07-17 and 07-18 read-outs, which
remain as dated snapshots. Evidence and turn-by-turn detail:
`NOTES_claims_verification.md`. Original design: `SPEC_claims_verification.md`.*

---

## What we're trying to do

Three principles, fixed before any code was written:

1. **Evidence is data.** Every site that wants protection gets a machine-readable
   evidence base: verified facts, each with its value, its source — a live SQL
   query, a code artifact, or a named human attestation — a verified-on date and a
   tolerance; plus that site's own list of audited-out fabrications; plus the named
   entities its copy may claim relationships with.
2. **Deterministic first.** Anything that can be a string or number comparison is
   one, before any language model is involved.
3. **Truth decisions are human.** The system flags; a person rules. It never
   rewrites a factual claim on its own. That is a rule about power, not
   convenience, and there is no code path that breaks it.

## Where we've come from

The platform writes website copy with a language model. Until this week, nothing
anywhere in it ever compared a **claim** to **evidence**.

There were quality layers, and they all worked — on the wrong problem. Generation
had prompt rules saying "never invent statistics". Build-time validation checked
*form*: placeholder text, unrendered templates, broken links, email addresses.
Post-deploy checks looked at *structure*: empty sections, links to pages that don't
exist. Every one of those asks whether a page is well-made. None of them asks
whether it is **true**.

The cost was not theoretical. Our own consultancy site shipped five invented client
case studies with invented results. It shipped a named founder who does not exist,
with a photograph that returns 404. It shipped "2,767 Awards Won" — a real number
wearing a false label. It described an eight-department agent taxonomy that exists
nowhere in the codebase, and a seventy-plus agent fleet that has never run. A
second site, vetcomparison, shipped fabricated veterinary prices — that one carries
legal exposure and had to be stripped, with a record kept. And when a fabrication
was audited out by hand, it came back weeks later, alive on an orphan page, because
nothing in the system knew it was on a banned list.

One detail decided the design. The single fabrication class that had a
deterministic checker — email addresses, compared against the site's real contact —
is the single class that was ever caught. Every time. The lesson was not "write
better prompts". It was: **prompts leak, humans don't scale, checkers work.**

## What we've done

**Five components, built in order. All of them exist; four are live in production.**

**The evidence base.** Eighteen verified facts for our consultancy site,
transcribed from the hand audit and re-verified against the live database.
Nineteen banned patterns — that site's own history of fabrication, turned into a
regression suite.

**The build-time gate.** Generated copy is scanned before it can be saved. A banned
claim is a hard blocker: those are *known* falsehoods, put on the list by a human. A
number asserted about the business that no registered fact supports routes the page
to human review rather than blocking outright, because number extraction can be
wrong and a person should see it either way.

**The post-deploy check.** The same scans run across the whole published site on the
improvement cycle, catching drift, hand-edits, and pages built before the gate
existed. Both layers read what a *reader* reads — text, not HTML attributes or code
samples. That distinction killed a long-standing false positive that had once
blocked every build of every page using the shared contact block.

**The writer's whitelist.** The old instruction was "never invent statistics" —
unbounded, and demonstrably leaky. The new one is bounded: the writer's prompt now
carries the verified facts as **the only numbers and named entities it may assert**,
with the instruction that if a fact isn't listed, write the capability without the
number. Same shape as the fix that worked for emails: "use only these" beats "don't
invent".

One deliberate subtlety, because it looks like an omission: **the banned claims are
kept out of the prompt.** Telling a model "never say eight departments" puts "eight
departments" into its context — the don't-think-of-an-elephant problem. So the
whitelist goes in, the blacklist stays out, and the blacklist is enforced by the
deterministic gate instead. Each mechanism does the job it is actually good at.

**The claims auditor.** For what pattern-matching cannot catch — fluent, unsupported
prose where no single number is wrong. One model pass per site classifies each
assertion of fact as supported, honestly-framed-as-possible, or unsupported against
the register. Findings go to the same human queue. It has no code path that edits
content.

**Freshness — the last piece, built this week.** Facts sourced from live queries go
stale: published copy stays frozen while the database moves, and the whitelist then
licenses the writer to assert a number that is no longer true. A scheduled pass now
re-runs every SQL-backed fact, re-syncs its value and date, and raises drift for a
human ruling — including the *under-claiming* case, where the site says 2,767 and the
database says 3,104. Not a lie, but stale, and worth knowing.

It also regenerates the whitelist, with a deliberate ownership split: **humans own
the words, the machine owns the numbers.** Each fact carries a human-authored line
with a placeholder where the figure goes, so caveats like "this is a catalogue
count — never present it as a running fleet" survive regeneration word for word, and
a fact nobody has phrased is simply omitted rather than auto-worded.

Two safety properties in that piece are worth stating, because both were deliberate
and both are the kind of thing that bites later. The SQL comes out of a data column,
so it is treated as untrusted: single statement, must begin with SELECT, no
data-modifying keyword anywhere, executed read-only with a timeout. And because the
pass rewrites a whole human-owned record from a copy read moments earlier, the write
is a compare-and-swap — if a person edited that record in between, the pass writes
nothing and says so. A lost refresh costs one scheduled tick; a lost human edit costs
trust.

### And the results

- **History became the test suite.** Every fabrication that previously shipped is now
  a regression test that must be caught. Each one is.

- **The first live scan falsified our own assumption.** We expected a clean site. We
  found nine banned-claim resurrections across four pages — one of them written
  *five days after* a manual cleanup sweep.

- **That exposed the real root cause: the specs were poisoned.** The direction spec
  feeding the writer literally instructed "secondary emphasis on 70-plus agents
  across 8 departments". Another rule told it to cite "least-privilege IAM policies"
  when discussing security — a capability we have never had. The writer was never
  hallucinating. It was obeying. We cleaned every spec, and found the invented
  founder still alive in one of them, and deleted him again.

- **The writer behaves under the new regime.** We rebuilt the worst-offending page as
  a live test: every prompt carried the whitelist, nothing was blocked, and the
  writer cited a whitelisted figure verbatim *with honest dating, unprompted* —
  "more than 90,790 orchestration state records to date, live count 2026-07-16".

- **It caught a stranger within hours.** A different automated thread wrote a
  call-to-action citing a "Digital Transformation Strategy" service — retired
  language, and a service that does not exist. The check caught it, changed nothing,
  and parked it for a human. A person ruled; the copy now names two real services.
  Drift, detection, ruling, fix — on traffic nobody staged.

- **And it kept finding platform faults on its way past.** A documented
  human-review checkpoint action turned out never to have been registered, so any
  workflow using it would fail — dead since it was written, now fixed. And when the
  freshness work was put through the platform's own review council, the council was
  voided by a defect of its own: one reviewer's answer was cut at its token ceiling,
  and the code that tallies verdicts throws away *every* reviewer's work when one is
  unreadable. Six complete reviews discarded. That is filed as a bug with a fix
  proposed, not worked around.

## Where we are now

Four of the five components are live and verified against running pods. The fifth —
freshness — is written, tested and committed, and is inert until the next image
build, as all code changes here are. Its scheduled task is written and deliberately
not yet switched on, because the house rule is image first, then configuration.

Both behaviours of the auditor are proven: on a clean site it returns an empty result
and creates no queue noise; on a site with no evidence base it completes without
making a single model call. Nothing is imposed fleet-wide — every part of this is
opt-in on whether a site has an evidence base, so a site that hasn't asked for it is
untouched.

Two honest boundaries. The auditor's *catch* has not yet been demonstrated, because
we will not plant a fabrication on a live site to prove a point — it will prove
itself on the first real drift, exactly as the deterministic check did within hours.
And the auditor is not yet on a schedule: that is a cost decision, one model call per
site per pass, and it is yours.

## Where we're going

**Immediately: one image build.** That activates freshness, after which the
scheduled task is switched on and the register keeps itself current.

**Then the second site: vetcomparison.** Its rebuild already requires
claim-licensing — every price in copy must trace to a licensed source row. That is
this layer's deterministic lane applied to the one fabrication class that has already
cost us legally.

**Then the fleet, by invitation.** Each site that wants protection gets its own
facts, its own history of fabrication, its own whitelist. Nothing is forced on a site
that hasn't asked for it.

---

One sentence to close on: the platform can now refuse outright the lies it has told
before, restrict itself to the numbers it can prove, notice a false claim arriving
from anywhere else, keep those numbers from going stale, and put every judgement call
in front of a person — and this week, on real traffic, it did all of it.
