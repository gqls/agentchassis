# Summary — the page saver that threw away the half it could not see

**2026-08-04** · `bugs_closed/194` · council `b6023fc1` APPROVED round 1

## What we're trying to do

Every page this platform builds is stored twice: once as finished HTML, which is what a
visitor sees, and once as structured content — the headline, the paragraphs, the button
labels, as data rather than markup. The second copy is the one that makes a page cheap to
maintain. When a price changes or an image lands, the platform re-renders from the
structured copy in seconds instead of paying a language model to rewrite the page. Keeping
both halves in step is the whole point.

## Where we've come from

The code that saves a page had to be *told*, in configuration, where the structured half
was. That instruction was added in February and copied by hand to each new caller
afterwards. Six parts of the platform now call it. Four had never been told.

Those four saved the HTML perfectly and dropped the structured half on the floor — writing
an empty value, reporting success, and serving a page that looked entirely correct. Nothing
complained, because nothing was watching. It ran that way for about six months.

It surfaced yesterday morning only by accident: a neighbouring thread's own test run
rebuilt a page correctly and stripped the structured content off all three of its sections
while doing it. That thread fixed one of the four callers and filed the rest.

## What we've done

**Fixed the three that were left, and found that one of them was not broken.** Two were
missing the instruction and now have it — a configuration change, live immediately. The
third turned out to be correct as it was: it rebuilds interactive tools as a single page of
HTML and has no structured content to keep, which the re-render side of the platform
already knew and quietly allowed for. The filing thread had flagged that one as unverified
and was right to; copying the instruction onto it would have been a guess.

**Then removed the need to be told.** The saver now looks for the structured content itself,
using the same address the page-validation code already uses — borrowed from it rather than
copied, so the two cannot drift apart. A caller that genuinely has no structured content
can now say so in its own configuration instead of leaving a silence indistinguishable from
a mistake, and all six now do say which case they are. A caller that wants a missing one to
be a hard failure can ask for that, though nobody is switched on: adding a new way for the
busiest path in the fleet to fail, on a prediction, is how safety checks get deleted a
fortnight later.

**And made the silence impossible.** When a page that had structured content is saved
without any, the platform now records it. That was the real defect — not that four callers
forgot a line, but that forgetting produced no evidence.

**What it cost, and what the evidence is.** Three configuration seeds, one code change,
seven tests. Because the two callers we re-routed are dormant — neither has run in the nine
days our records cover — we cannot prove them on live traffic, so we broke the new code
four different ways on purpose and each break was caught by the test written for it. The
reviewer council approved it first time, with four advisory objections; three we answered
with measurements, and one we recorded as an open question rather than pretending it was
settled.

## Where we are now

The bug as filed is fixed and live: no caller in the fleet can now silently lose structured
content. The structural half — the part that stops a *future* caller from reintroducing the
same mistake — is written, reviewed and approved, but does nothing until the next release
build, as is normal here. The checks that will prove it after that build are written down in
advance, including what would count as failure.

One thing stays genuinely open, and it is the sharpest point the reviewers made: for the
callers running today, a loss is now *recorded* rather than *prevented*. That is deliberate
— the preventing version is switched off by design until we have a week of recordings to
justify switching it on — but it is a deferral, not a fix, and whether that is good enough
is a judgement rather than a deduction.

## Where we're going

Three things, in order. After the next release: confirm the code actually shipped by
checking the running system rather than the release tag, then trigger one of the two
dormant callers by hand, since one of them can be driven directly. After a day of
production traffic: read the new records. If they are silent for the two busy callers, the
preventing version can be switched on caller by caller, which turns "we think this is safe"
into something measured. If they are *not* silent, the recording rule itself is wrong and
we stop and understand that before going further.

Separately, and not part of this: about 13% of stored page sections have no structured
content. Some of that is this bug; some of it is simply pages older than the feature. Each
one turns a cheap refresh into an expensive rebuild — we counted 44 such escalations across
8 sites since mid-July, 13 of which failed outright on 3 August. Repairing them means
rebuilding those pages, not restoring the old data, and it is a separate piece of work with
its own cost.
