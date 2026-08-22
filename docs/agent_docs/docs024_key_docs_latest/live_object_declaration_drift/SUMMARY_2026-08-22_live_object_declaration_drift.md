# Summary — live-object declaration drift, 2026-08-22

## What we're trying to do

Some of this platform's behaviour lives in the database rather than in code: a scheduled job's query,
a database trigger, a rule about which values a column accepts, an agent's workflow. Ordinary Go
tests cannot see any of it, because there is no database when tests run. We want a way for the code
and that live behaviour to be held in step that actually works — one where the test can fail when
they disagree.

## Where we've come from

This lane did not start as a project. It started as a request to fix bug 317, which turned out to
have been fixed and closed the day before. Checking that properly — reading the live database rather
than trusting the closure note — is what turned up the real problem.

The way we had been bridging the code/database gap was to have a test read the **migration file**
that once created the object, and check the code agrees with that file. But a migration is a receipt,
not a description. Our own rule, correctly, is that a migration must never be edited after it has run,
because the system stores a fingerprint of it. So the file is frozen the day it applies, while the
live object carries on being changed by every later migration.

The consequence is worse than "the file might be out of date". Three of these tests were asserting
that a frozen file still contains a particular line — an event our own rules have already made
impossible. They could not fail. Two more quietly passed if their file was merely renamed.

## What we've done

Measured it, first. Seven tests across four kinds of database object, and all seven live objects
checked directly. **They all still agree** — there is no damage, and we filed on the strength of the
open door rather than a fault, the same basis bug 317 itself used.

Then found the evidence that makes it more than theory: the exclusion list at the heart of bug 317
has been amended by three separate migrations whose filenames the guarding test cannot even see, and
the original migration was edited nine times after it applied, back when it *was* the living
description. When the freeze rule took that role away, nothing replaced it — which is why the last
author had to write the description into a **comment** for the test to read.

Filed as `bugs_open/363`. Put through the review council: rejected at round one on a fair objection
about the order I had planned the work in, revised, and approved at round two. The first half is
built and committed: a package that holds the current description in a file that is allowed to
change, four tests converted to use it, both silent-pass bugs removed, and an automatic check that
stops the pattern coming back.

We also ran the independent diagnosis loop. It came back inconclusive, and that is recorded as
inconclusive — it never reached a verdict because I dispatched it without telling it where to look.
It did, however, quote the live database back at me, which is how we found the sharpest thing in the
whole file.

## Where we are now

The first half is live in the code and proven: the new check was seen failing on exactly the four
files it should, then passing once they were fixed, and six deliberate breakages were each caught.

The sharpest finding is one that constrains what comes next. The live setting for the claimed-item
timeout **still describes itself incorrectly** — it states a rule that was superseded three days ago
and names a test that has been deleted. That sentence is the original cause of bug 317. So the
comfortable framing, "the old files are stale but the live system is the truth", is not right either:
the live system is carrying its own out-of-date account of itself, and a checker that compared the
live values against our written-down values would sail straight past it, because the values match and
it is the prose that lies.

Bug 363 remains open, deliberately.

## Where we're going

The second half: a daily job that reads the live objects and compares them to what we have written
down. That is a separate piece of work with its own review, because it means a new deployed service.
It was split out precisely so the first half could land cleanly.

Until it exists there is a new, honest gap: someone changing one of these live settings must remember
to update the written description, and nothing will tell them if they don't. That is written up as a
trap in the shared file rather than left implied. It is a smaller gap than the one we started with —
tests that could not fail — but it is a real one, and it is why the bug stays open.
