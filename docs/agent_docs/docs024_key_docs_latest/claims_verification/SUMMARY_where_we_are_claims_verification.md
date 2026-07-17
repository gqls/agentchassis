# Claims Verification — what we're doing, where we've got to, where we're going

*Written 2026-07-17, to be read aloud. Detail and evidence live in
`RUNNING_NOTES_claims_verification.md`; the design is `SPEC_claims_verification.md`.*

---

## What we're doing

Our platform writes website copy with an LLM, and until this week nothing anywhere
in the pipeline ever compared a **claim** to **evidence**. Generation had prompt
rules saying "never invent". Validation checked form — placeholders, broken links,
bad emails. Post-deploy checks looked at structure. Truth was checked by one person
reading pages against a markdown file.

That gap had real consequences. The leopardess site shipped invented client case
studies, an invented founder, invented statistics, an invented departmental
taxonomy — all straight through every automated layer. One fabrication was audited
out and came back weeks later on an orphan page. Another site, vetcomparison,
shipped fabricated prices, with legal exposure. The one fabrication class that had
a deterministic checker — email addresses — is the one class that was always
caught. That was the lesson: prompts are leaky, humans don't scale, checkers work.

So we are building a claims-verification layer. Three principles:

1. **Evidence is data.** Each site gets a machine-readable *evidence base*: a
   register of verified facts — each with its value, its source (a live SQL
   query, a code artifact, or a named human attestation), a verified-on date and
   a tolerance — plus a per-site blacklist of its own audited-out fabrications,
   and a whitelist of named entities the copy may claim relationships with.
2. **Deterministic first.** Every check that can be a string or number comparison
   is one, before any LLM gets involved.
3. **Truth decisions are human.** The system flags; a person rules. It never
   rewrites factual content on its own.

## Where we've got to

**All of V0 through V2 is built, deployed, and proven in production.**

- **The evidence base is live** for leopardess: eighteen verified facts and
  nineteen banned patterns, transcribed from the audit and re-verified against
  the live database.

- **Two deterministic checkpoints are live.** At build time, a gate scans
  generated copy: a banned claim is a hard blocker — those are *known* falsehoods
  — and any number asserted about the business that no registered fact supports
  routes the page to human review. After deploy, a discovery check runs the same
  scans over the whole published site, catching drift, hand-edits, and pages
  that predate the gate. Both scan what a reader actually reads — text, not HTML
  attributes or code samples — which killed a long-standing false positive.

- **The writer now gets a whitelist.** The old rule was "never invent statistics"
  — unbounded, and provably leaky. The new rule is bounded: the writer's prompt
  now carries the verified facts as **the only numbers and named entities it may
  assert**. If a fact isn't listed, write the capability without the number.
  This is the same fix that worked for emails: "use only these" beats "don't
  invent". One deliberate subtlety: the **banned claims are kept out of the
  prompt**. Telling the model "never say eight departments" puts "eight
  departments" in its head — don't-think-of-an-elephant. The whitelist goes in;
  the blacklist stays outside, enforced by the deterministic gate instead.

- **History is the benchmark.** Every fabrication that previously shipped is now
  a regression test — the eight-departments claim, the seventy-agents fleet
  claim, the fake case studies, the "2,767 Awards Won" stat, the placeholder
  email in an attribute that must *not* flag. All pass.

- **It found real problems immediately.** The first live scan falsified our own
  "site is clean" assumption: nine banned-claim resurrections across four pages
  — including a guide written five days *after* a cleanup sweep. Digging into
  why exposed the real root cause: **the specs feeding the writer were
  poisoned**. The direction spec literally instructed "secondary emphasis on
  70-plus agents across 8 departments", and one writing rule told the writer to
  cite "least-privilege IAM policies" when mentioning security — a capability we
  never had. The writer was never hallucinating; it was obeying. We cleaned
  every spec, deleted a set of dormant invented team personas — and found the
  invented founder still alive in a briefing spec, and deleted him again.

- **The writer behaves under the new regime.** We rebuilt the worst-offending
  page as a live test: every prompt carried the whitelist, zero blockers, and
  the writer cited a whitelisted figure verbatim *with honest dating,
  unprompted*: "more than 90,790 orchestration state records to date (live
  count, 2026-07-16)". That is exactly the behaviour the layer exists to produce.

- **And it caught its first stranger.** Hours after the post-deploy check went
  live, a separate automated thread wrote a call-to-action citing a "Digital
  Transformation Strategy" service — retired language, and a service that does
  not exist. The check caught it within three hours, parked it for human review,
  changed nothing on its own. A human ruled; the copy now names two real
  services. Drift to detection to ruling to fix, exactly as designed.

## Where we're going

- **V3 — the judgement lane.** A claims-auditor agent: one LLM call per page
  that extracts prose assertions — "we have done X", "clients Y" — and
  classifies each as supported, could-framed, or unsupported against the
  evidence base. This catches what regex can't: the fake-case-study class,
  where every sentence is fluent and no single number is wrong. Its findings
  route to the same human-review queue; it never rewrites.

- **V4 — freshness.** Re-run the SQL-sourced facts on a schedule, update values
  and dates, raise a finding when the live number drifts outside tolerance —
  including *under*claiming, since the site saying 2,767 while the database says
  3,104 is also worth a look. The same mechanism keeps the writer's whitelist
  current, and feeds the planned live-data chart components.

- **Second site: vetcomparison.** Its rebuild already requires claim-licensing —
  every price in copy must trace to a licensed source row. That is this layer's
  deterministic lane applied to the one fabrication class with legal exposure.

- **Fleet-wide, by choice.** Everything is opt-in on evidence-base presence. A
  site without one is untouched. Each site that wants protection gets its own
  facts, its own banned history, its own whitelist.

One sentence for the whole programme: the platform can now notice when it is
about to publish an unsupported claim, refuse outright the lies it has told
before, and put everything else in front of a person — and this week it did all
three on real traffic.
