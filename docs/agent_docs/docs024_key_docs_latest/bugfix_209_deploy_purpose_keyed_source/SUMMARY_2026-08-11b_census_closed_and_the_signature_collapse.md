# SUMMARY — 2026-08-11b: the census is closed, and it found four auditors signing someone else's name

*(Second summary today by design — the morning one closed the logo saga; this
one closes the question the class fix opened. Different milestone.)*

## What we're trying to do

When we fixed the last of the eleven broken logos, the root cause turned out
to be a general trap in the platform: an agent's configuration can say one
thing while the code's built-in default silently wins, with nothing anywhere
reporting the disagreement. We set out to answer, fleet-wide: how many other
places is this happening, how bad are they, and what stops the next one?

## Where we've come from

Bug 231 named the trap and proved two live faces of it (the logo statics fixed
by migration 348, and the dotted binding fixed by migration 380). But its
fleet-wide size was a guess ("~10 specs"), and two diagnosis-loop runs had
confirmed the mechanism without being able to enumerate the exposure. The
morning handoff left three census arms and two fix candidates open.

## What we've done

We built the counter instead of counting by hand: a new mode on the existing
config audit tool that asks the running binary for every built-in default (62
specs, 232 defaulted fields — six times the guess) and checks every live
agent's config against them, classifying each hit by the exact code path that
kills it. It is calibrated on both known faces as committed tests, so it
cannot quietly stop firing on the very cases that motivated it.

First fleet run: 195 findings. Seventy-five restate the default (harmless
today, a trap on first edit). Ninety-six are conditional bindings, every
diverging one of which we traced to a verdict. Twenty-four say something
different from the default that shadows them — and hand-checking each one
showed twenty are actually honoured through a different read path. Four are
real, and they are all the same bug: our four auditor agents (brief fidelity,
content quality, site review, visual design) each sign their findings with
their own name, and all four signatures are dead. Every finding any of them
has ever filed — 136 rows back to April, including one July row whose own
type says "brief fidelity" while its signature says "design-audit" — is
signed with the wrong name. That signature is the one field the bug-213 work
relies on to tell producers apart, so that thread has been told in writing.

Everything is filed: the bug file carries the full census with evidence, the
landmine list and the council's shared notes carry the trap, the runbook
carries the commands and the one check that kept us honest (a flagged config
is only damage if the action reads it through the affected path — that check
killed three scary-looking findings before they became claims).

## Where we are now

The census question is answered completely, the detector is committed and
green (chassis v1.0.1288 includes it), and the class has a standing guard: any
future config that falls into this trap fails the audit with a non-zero exit.
The four mislabelled auditors are still mislabelling — the fix is four lines
but sits in a file another active thread owns, so it is costed and offered,
not done. The council gate declined to review the tool change as out of its
scope (cmd/ and scripts/), which is recorded, not worked around.

## Where we're going

Three decisions are the owner's (detailed in README_where_we_are and the
handoff): whether the four-line auditor fix ships now and by whom; whether the
general "config beats default" change is still worth doing now that its
measured blast radius is exactly those four entries — arguably the detector
alone is enough; and the stale logo.jpg deletion carried from the morning.
After that, this lane's remaining threads are the kafka sweep's first real
timed run (00:17 tonight), and the unowned 209 Phase 3 / 236 items.
