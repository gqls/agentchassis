# Claims Verification — the full journey, to be read aloud

*2026-07-18. Supersedes nothing: `SUMMARY_where_we_are_claims_verification.md` is
the shorter mid-week version. Evidence and dated detail:
`NOTES_claims_verification.md`. Design: `SPEC_claims_verification.md`.*

---

## Where we came from

The platform writes website copy with a language model. Until this week, nothing
anywhere in it ever compared a **claim** to **evidence**.

There were layers, and they all worked — on the wrong problem. Generation had
prompt rules saying "never invent statistics". Build-time validation checked
*form*: placeholder text, unrendered templates, broken links, email addresses.
Post-deploy checks looked at *structure*: empty sections, links to pages that
don't exist. Every one of those is a check that the page is well-made. None of
them is a check that the page is **true**.

The cost of that gap was not theoretical. Our own consultancy site shipped five
invented client case studies with invented results. It shipped a named founder
who does not exist, with a photograph that 404s. It shipped "2,767 Awards Won" —
a real number, wearing a false label. It described an eight-department agent
taxonomy that exists nowhere in the codebase, and a seventy-plus agent fleet
that has never run. A separate site, vetcomparison, shipped fabricated
veterinary prices — that one carries legal exposure and had to be stripped and
recorded. And when a fabrication was audited out by hand, it came back weeks
later, alive on an orphan page, because nothing in the system knew it was on a
banned list.

One detail decided the whole design. The single fabrication class that had a
deterministic checker — email addresses, compared against the site's real
contact — is the single class that was ever caught. Every time. The lesson was
not "write better prompts". It was: **prompts leak, humans don't scale, checkers
work. So build the checker.**

## What we set out to do

Three principles, chosen before any code:

1. **Evidence is data.** Every site that wants protection gets a machine-readable
   evidence base: verified facts, each with its value, its source — a live SQL
   query, a code artifact, or a named human attestation — a verified-on date and
   a tolerance; plus that site's own list of audited-out fabrications; plus the
   named entities its copy may claim relationships with.
2. **Deterministic first.** Anything that can be a string or number comparison
   is one, before any language model is involved.
3. **Truth decisions are human.** The system flags; a person rules. It never
   rewrites a factual claim on its own. That is a rule about power, not
   convenience, and there is no code path that breaks it.

## What we've done

**Five components, built in order, all live, all proven on real traffic.**

**The evidence base.** Eighteen verified facts for our consultancy site,
transcribed from the hand audit and re-verified against the live database on the
day. Nineteen banned patterns — the site's own history of fabrication, turned
into a regression suite.

**The build-time gate.** Generated copy is scanned before it can be saved. A
banned claim is a hard blocker: those are *known* falsehoods for that site,
placed on the list by a human. A number asserted about the business that no
registered fact supports routes the page to human review rather than blocking —
because number extraction can be wrong, and a human should see it either way.

**The post-deploy check.** The same scans run across the whole published site on
the improvement cycle, catching drift, hand-edits, and pages built before the
gate existed. Both layers read what a *reader* reads — text, not HTML attributes
or code samples. That distinction killed a long-standing false positive that had
once blocked every build of every page using the shared contact block.

**The writer's whitelist.** The old instruction was "never invent statistics" —
unbounded, and demonstrably leaky. The new one is bounded: the writer's prompt
now carries the verified facts as **the only numbers and named entities it may
assert**, with the instruction that if a fact isn't listed, write the capability
without the number. Same shape as the fix that worked for emails: "use only
these" beats "don't invent".

One deliberate subtlety worth stating plainly, because it looks like an
omission: **the banned claims are kept out of the prompt.** Telling a model
"never say eight departments" puts "eight departments" into its context — the
don't-think-of-an-elephant problem. So the whitelist goes in, the blacklist
stays out, and the blacklist is enforced by the deterministic gate instead.
Each mechanism does the job it is actually good at.

**The claims auditor.** For what regex cannot catch — fluent, unsupported prose
where no single number is wrong. One model pass per site classifies each
assertion of fact as supported, honestly-framed-as-possible, or unsupported,
against the register. Its findings go to the same human queue. It has no
code path that edits content.

**And the results.**

- **History became the test suite.** Every fabrication that previously shipped
  is now a regression test that must be caught — and each one is.

- **The first live scan falsified our own assumption.** We expected a clean
  site. We found nine banned-claim resurrections across four pages — including
  one written *five days after* a manual cleanup sweep.

- **That led to the real root cause: the specs were poisoned.** The direction
  spec that feeds the writer literally instructed "secondary emphasis on
  70-plus agents across 8 departments". Another rule told the writer to cite
  "least-privilege IAM policies" when discussing security — a capability we have
  never had. The writer was never hallucinating. It was obeying. We cleaned every
  spec, and found the invented founder still alive in one of them, and deleted
  him again.

- **The writer behaves under the new regime.** We rebuilt the worst-offending
  page as a live test: every prompt carried the whitelist, nothing was blocked,
  and the writer cited a whitelisted figure verbatim *with honest dating,
  unprompted* — "more than 90,790 orchestration state records to date, live
  count 2026-07-16". That is precisely the behaviour the layer exists to produce.

- **It caught a stranger within hours.** A different automated thread wrote a
  call-to-action citing a "Digital Transformation Strategy" service — retired
  language, and a service that does not exist. The check caught it, changed
  nothing, and parked it for a human. A person ruled; the copy now names two
  real services. Drift, detection, ruling, fix — the whole cycle, on traffic
  nobody staged.

- **And it fixed something in the platform on the way past.** The documented
  human-review checkpoint action turned out never to have been registered — any
  workflow using it would have failed. Nothing had ever used it. That is now
  fixed.

## Where we are now

Everything above is live and verified against running pods, not against git.
Both behaviours of the new auditor are proven: on a clean site it returns an
empty result and creates no queue noise; on a site with no evidence base it
completes without making a single model call. Nothing is imposed fleet-wide —
every part of this is opt-in on whether a site has an evidence base, so a site
that hasn't asked for it is untouched.

Two honest boundaries. First: the auditor's *catch* has not yet been
demonstrated, because we will not plant a fabrication on a live site to prove a
point — it will prove itself on the first real drift, exactly as the
deterministic check did. Second: the auditor is not yet on a schedule; that is a
cost decision, one model call per site per pass, and it is yours to make.

## Where we're going

**Freshness — the last piece, in build now.** Facts that come from live queries
go stale. A site saying 2,767 when the database says 3,104 is not lying, but it
is out of date, and under-claiming is still worth knowing about. So: a scheduled
pass re-runs every SQL-backed fact, updates its value and its verified-on date,
raises a flagged item when a live number drifts outside its stated tolerance —
and, importantly, regenerates the writer's whitelist automatically, so the
numbers the writer is allowed to use can never quietly rot.

**Then the second site: vetcomparison.** Its rebuild already requires
claim-licensing — every price in copy must trace to a licensed source row. That
is this layer's deterministic lane applied to the one fabrication class that has
already cost us legally.

**Then the fleet, by invitation.** Each site that wants protection gets its own
facts, its own history of fabrication, its own whitelist. Nothing is forced on a
site that hasn't asked.

---

One sentence to close on: the platform can now refuse outright the lies it has
told before, restrict itself to the numbers it can prove, notice a false claim
that arrives from anywhere else, and put every judgement call in front of a
person — and this week, on real traffic, it did all four.
