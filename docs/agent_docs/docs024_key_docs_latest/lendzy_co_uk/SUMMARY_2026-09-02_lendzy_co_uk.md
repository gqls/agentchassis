# SUMMARY — lendzy.co.uk — 2026-09-02

The lane's first read-out, written to be said aloud.

## What we're trying to do

Lendzy is our consumer-credit rights site — price caps, rollover limits, card-payment rules,
complaints. It was the first site the framework ever built end to end, and as of today it has its
own lane. The lane's job is threefold: keep the site healthy, and above all make one sentence true
that was once served falsely — that the site's financial facts are checked against the FCA
Handbook, rule by rule. Not as a slogan: as running machinery, with the sentence itself staying off
the site until the checking is real and regular.

## Where we've come from

The site was built on 2 August as a shadow experiment, and carried two birthmarks from that day.
Three of its nine calculators were stored without a component identity, which meant they served
perfectly while being recorded as never built — unpublishable, missing from the sitemap, and the
target of 47 spurious "broken link" reports. And a planted test sentence claiming FCA diligence
became a served compliance claim for 24 days (the 414 affair, closed in August). Four different
lanes worked pieces of lendzy; nobody held the whole.

## What we've done

Today, in order: opened the lane and its working docs. Traced the three stuck calculators to their
root cause and repaired them by adoption — their own live HTML became their component, byte for
byte, no regeneration — so all three now carry the first publication stamps of their lives, with
their input counts proving no calculator was swapped. Checked every FCA rule the site cites against
the Handbook's own text and found two citations wrong (the rollover limit attributed to a
definitions rule; the card-payment limit attributed to the rollover rule); on the owner's word,
fixed both everywhere they were stored — pages, the writing instructions that would have re-planted
them, the tool template, and a fork of that tool serving on loancash — and verified the correction
in the served bytes of both sites. Built lendzy's evidence register: eight facts, each citing its
rule with a verbatim quote verified through the production matcher. And the method travelled: three
sibling finance sites built their own registers the same day, finding three more wrong claims
between them. Five wrong live claims found in one day is the measured answer to "was this worth
doing".

## Where we are now

The site is fully healthy at the artefact: every page serves, every calculator works and is
properly recorded, no wrong rule number is served anywhere in the fleet. The register migration
(695) is the one unapplied piece — its council review has been killed twice by today's rolling
chassis deploys and waits for a calm window. Two tails will close on their own schedules: the
sitemap picks up the three repaired pages on its next rotation (27 → 30), and the 47 link reports
drain as revalidation re-judges them. The compliance sentence remains off, by the owner's ruling,
until the new rule-span checker — approved today, built by the claims-verification lane — is
running regularly.

## Where we're going

Three threads. First, finish the register: apply 695 when its verdict lands, then watch the daily
checker's first pass over lendzy's facts. Second, the local Handbook mirror — cited chapters first,
tables designed to widen — which waits on pacing work for the shared fetch path, and inherits a
day-one landmine: the FCA's site returns success for every address including invented rules, so
everything we build must prove it can tell a real rule from a fake one, every run. Third, the
sentence itself: once checking is live and proven, the owner decides what the site says about it —
true, dated, and after the machinery, never before it.
