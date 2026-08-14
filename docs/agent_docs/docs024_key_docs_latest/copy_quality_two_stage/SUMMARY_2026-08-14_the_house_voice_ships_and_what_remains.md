# SUMMARY 2026-08-14 — the house voice ships, and what remains

Second summary in this lane. The first (08-12) reported why the copy read wrong and what
would fix it. This one records the first delivery. Written to be read aloud.

## What we're trying to do

Have the framework write copy that reads as though an intelligent person wrote it — for
every site, without a human rewriting pages by hand. The lane exists because bad copy was
traced to process, not to the writing model: briefs and specs were ordering the faults,
and corrections had nowhere to land that reached more than one page.

## Where we've come from

The lane opened on the 12th from a rejected homepage. Its first day established that the
rejected copy was the brief faithfully rendered; that the negativity came from the site's
own identity spec, written at adoption and never revisited; and that most of the machinery
we planned to build already existed unused. A day of research found the owner's 9 August
voice decision undelivered, a shared voice carrier built in July with a single consumer,
and the mortgagecalculator thread's evening of corrections — which superseded the voice
the undelivered decision would have shipped.

## What we've done

The owner read two candidate texts for the fleet house voice: one assembled from the
9 August decision, one from the mortgagecalculator corrections (register-matched
contractions, considered British sentences, positive definition, the read-aloud test). He
asked to see real copy under the newer one before choosing; we ran a real writer at a real
page with every section locked, so the proposal was captured and nothing shipped. He
approved the newer text.

It is now live fleet-wide: the seven writing agents had their private, drifted copies of
the old voice deleted and replaced with a reference to the one shared block, and the block
was swapped for the approved text in a single edit. Verification was done at the artefact,
not the configuration — a second locked-page run whose actual received prompt carried the
new voice exactly once and the old rule zero times. Every prior text is backed up; any of
it is restorable with one command.

Along the way: two bugs filed (four auditors filing every finding under one producer name;
a schema dialect declared extinct that is quietly returning), two landmines recorded, the
carrier registered so it cannot be rediscovered from scratch a third time, and the sample
run reproduced the known set-preservation failure a fourth time — the writer added two
links it was told not to add — confirming that axis stays mechanical.

## Where we are now

Every writer in the fleet, and the blog writer, carries one owner-approved house voice
from one place, and the next correction to it is one edit. Sites' own voice specs still
outrank it, so hand-refined sites are untouched. The proof case for stage 2 — six missing
homepage links the owner ruled should stay until stage 2 fixes them — is committed and
waiting. The stage-2 build itself remains deliberately last: its output queues for human
review by the owner's decision, and the queue it would feed has no working surface yet,
so that bug is now this lane's blocking dependency.

## Where we're going

In order: the audit-attribution fix (config work on four agents); the read-only review of
the three lending sites that carry the "authoritative" tone setting; switching on the
claims checker for the calculator site; then the review surface; then stage 2, graded
against its committed proof case. The first real test of the new voice arrives on its own:
the next page the framework builds on a site with no voice spec of its own.
