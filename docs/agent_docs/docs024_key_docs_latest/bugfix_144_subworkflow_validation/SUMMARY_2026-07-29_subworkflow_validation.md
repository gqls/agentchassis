# SUMMARY — 2026-07-29 — sub-workflow validation (bug 144)

Written to be read aloud. Current state only; the chronology is in
`README_where_we_are.md` and the evidence in `NOTES`.

---

## What we are trying to do

Make sure that when one of our agents runs a workflow, *all* of it gets checked before
it runs — including the little workflows that live inside loops. And make sure the
offline report that answers "what do our live workflows actually contain?" looks in the
same places the checker does, so the two cannot quietly disagree.

## Where we have come from

The check has always run once, over the top level of a workflow. But most of our
substantial workflows contain a loop, and a loop carries its own small workflow inside
it — "for each page, do these five things". Those inner steps were extracted straight
out of configuration and executed, having been checked by nothing at all: eighty-five
of them, live, across eighteen agents.

It stayed hidden because our second, offline report was written to look in exactly the
same place. Two checks blind in the same direction agree with each other, and that
agreement reads exactly like correctness. It was found sideways, by a reviewer objecting
to a completely different claim that had been measured with the same kind of query.

## What we have done

The checker now goes into the loops. The offline report no longer walks the workflow
itself — it calls the same piece of code the checker uses, so if one goes blind the
other stops compiling.

Before writing any of it, we pulled every live definition out of the database and ran
the proposed new checker over all 178 of them: **nothing new is refused**. That harness
ships with the change and re-runs in seconds. It validates each workflow twice — once
whole, once with the loops removed — and reports only the difference, so somebody else's
pre-existing problem can never be mistaken for damage from your patch.

Two decisions came from reading the code that actually runs the loops, rather than from
the bug report's suggested fix, and both would otherwise have been wrong in a way no
measurement would have caught: a step inside a loop is *allowed* to point at a step
outside it, and the loop runner silently ignores several fields you might write on a
nested step — so rather than pretend to enforce those, the checker now tells you they
are being thrown away.

The review council approved it first time, with four advisory objections. Two were
right and were answered with code before committing: one reviewer noticed we had
written a second copy of the loop runner's decoder and pinned the two together with a
test — which is precisely the defect this bug *is*, recreated inside its own fix — so
there is now one decoder with two callers. Another noticed we had assumed a database
flag bounded the blast radius without checking; it does, and the answer is the same
either way.

## Where we are now

Committed and council-approved. **Not yet live** — Go code does nothing until a chassis
image carrying it is built and rolled, so the bug stays open until it is running and
verified on the pod. The check to run after the next roll is written down, and it is
deliberately a check for a string this change *deleted*, because that cannot be
satisfied by a stale binary.

Two things are waiting on a human, not on us:

* **The venue question.** This changes what "validated" guarantees for every pipeline,
  and our own rules say a change of that kind may belong in an architecture review
  rather than a bug fix. The guardian reviewer asked for a person to settle it. We have
  recorded it as open rather than argued it away.
* **A separate finding.** The new harness turned up three live agent definitions that
  are refused by the checker *today*, before any of our changes — they name actions that
  no longer exist anywhere in the code, and three live builders still try to call two of
  them. Filed as bug 148 rather than fixed here.

## Where we are going

Wait for the next chassis roll; verify on the pod; close the bug into `/bugs_closed/`
and update the register entry from "built" to "deployed". Then bug 148 is the natural
next thing, and it is small — the traversal and the registry lookup both already exist,
so it is a report rather than a mechanism.
