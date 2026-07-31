# SUMMARY — 2026-07-31 — the report gate that destroyed a correct report

## What we're trying to do

Keep the gripper-report pipeline trustworthy enough to put in front of a customer. Its last
step before a page is built is a deterministic gate, `verify_report_prose`, which refuses any
report whose prose contains a number or a product name that did not come from the measured
facts. The gate is fail-closed on purpose: refusal means no page at all. That is what makes
the reports safe to publish, and it is not up for negotiation.

## Where we've come from

On 2026-07-31 a fixture regeneration produced no page. The gate had refused a report because
the summary said **"IP54-or-better"**. IP54 was in the fact block — the requirement the
scoring step itself wrote — but the composed phrase was not, and the gate's test was verbatim
containment. So a correct sentence destroyed a correct report, the URL 404'd, and nothing on
any dashboard said why. The gripper-dossier lane filed it as `bugs_open/160` on its way out
and did not take it.

Three properties made it worse than a one-line bug. It is intermittent, because the trigger is
the writer's phrasing — the identical spec had passed four days earlier, so it reads as a
flaky pipeline and a retry "fixes" it. Its error message says the writer named a model-like
token, which reads as *the model hallucinated*, the exact opposite of what happened. And it is
a family, not a phrase: every hyphenated engineering notation a writer might paraphrase is
exposed.

## What we've done

Reproduced it first as a failing unit test, then fixed the classifier: a token now also clears
when it splits into a head that traces to the facts and a tail of qualifier words. The regex
guarantees the digit-bearing part sits in the head, so what was relaxed is which suffixes are
tolerated — never whether the model number itself was published.

The first version of that rule admitted a tail word by **shape**: lower-case, no digits, two
or more letters. It passed every test, including the fabricated siblings the guard exists to
catch. The council refused it at high severity, and correctly: "not" is exactly as lower-case
as "or", so a report could have said **"IP54-not-rated"** and passed a gate whose whole job is
to stop a report claiming what the facts do not say. The fabrication had not been closed off,
only moved — from the model number to the qualifier. The shipped rule is a closed vocabulary
of about thirty words under three stated admission rules: connectives that assert nothing,
strengtheners that ask for at least the stated value, and attributives that add no fact to the
code they attach to. Nothing that negates, inverts, or names a product family.

It is live on chassis v1.0.1222, verified on both running pods, and proven against the actual
incident rather than a fixture: the destroyed run's own prose and scoring, pulled back out of
the database, reproduce the live error byte-for-byte with the fix disabled and produce zero
violations with it in place.

## Where we are now

`bugs_closed/160`. The gate is strictly stronger than it was this morning against fabrication,
and no longer destroys reports over English.

Two things are worth carrying forward more than the fix is. The first is why the round-1
reasoning was wrong: I declined to build the vocabulary because no such failure had been
observed — but it could not have been observed, because the class did not exist until my own
change created it. **A rule that admits by shape has to be argued against the space of strings
it admits, not the strings that have turned up.** The second is that three separate tests in
this lane were green while never reaching the rule they were written for, and nothing but
deliberately breaking the rule could have shown that: a negative test tells you the input was
refused, never *which* rule refused it.

## Where we're going

Nothing is owed on this bug. The one open thread it touched and did not take: the landmine
recorded for this file is flagged `NEEDS_VERIFICATION` and was not dispatched, because
`bugs_open/163` (filed today, unowned) records that the landmine-verifier cannot answer a
path-bearing query — verifying it would have produced a false negative against the corpus.
That belongs to `163`'s fixing lane.
