# SUMMARY 2026-08-17 — bugs_open/297: the tool-recreation analyst now sees the whole site

## What we're trying to do

When someone's existing website is adopted into our system, any interactive thing on it — a
calculator, a quiz, a game — has to be rebuilt from scratch as clean, self-contained code. Before
rebuilding, the system asks an AI to write a specification of what the tool does and how it fits the
site it lives on. To answer the second half of that, the AI is shown a list of the site's other
pages. This work was about that list.

## Where we've come from

A sister piece of work (bug 275) had found a general problem: several places in the system ask the
database a question and quietly keep only the first N answers before handing them to an AI. That is
invisible by construction — the AI writes something plausible whether it saw ten items or a hundred,
so nothing ever looks wrong and no error is raised. That lane fixed its own instance, built a
fleet-wide detector for the class, and filed the two worst remaining cases as tickets. This is the
first of those two, and by the numbers it was worse than the bug that found it.

## What we've done

The list of "other pages on this site" was capped at ten, and which ten was decided by menu order —
so the AI saw whatever sat highest in the navigation, which is a judgement about menus rather than
about what a tool needs to know. Nineteen of our twenty-five sites have more pages than that cap. On
the biggest, the AI was seeing ten pages out of a hundred and seven.

The obvious move was to copy the sister fix: trim the longest field so the whole list fits. We
measured first, and the measurement said the opposite. These rows are one short line each — a page
name, a type and a title — so the entire hundred-and-seven-page list costs about nine thousand
characters, in a prompt that already carries the original page's complete source code. Nothing
needed trimming. The cap simply went, and because nothing was cut there is no truncation to mark.
That also means there is no new number to outgrow in a month, which is what the ticket asked for.

While measuring we found a second, unrelated fault in the same query, live at the time: because of
how the query joined to our research records, a page with two research entries was being listed
twice. One page on one site already had that, so its prompt was showing the same page twice inside
the visible ten. Removing the cap would have made that worse rather than better, so the same change
closed it — each page now contributes exactly one line, using its most recent research entry.

The change went through the review council and was approved on the first round: fourteen reviewers,
four advisory objections, none serious enough to block. The objections were worth their fee. The
strongest asked what stops the prompt growing forever now the cap is gone, and pointed out I had
left that as something to check later. Checking it produced a better answer than the fix had: this
step has run 129 times and has never once had its reply cut off, it would fail loudly rather than
silently if it did, and there is already a standing six-hourly check across the whole fleet that
watches exactly this and lists this step by name — peak usage 96.7% of its allowance, zero
truncations. Two other objections caught genuine sloppiness of mine, both now written up in our
running log of wrong calls.

## Where we are now

The fix is live and verified: the cap is gone from the running configuration, the biggest site's
query now returns all 107 of its pages where it returned 10, and the duplicate is gone. A backup of
the previous configuration was taken automatically before the change, and a tested one-command
revert sits beside it. The migration is recorded in the ledger, the council verdict is recorded
against the commit, and the bug file carries the evidence.

One thing is honestly still outstanding rather than claimed: the final end-to-end proof — seeing the
longer list inside a real recreation's prompt — needs the next real tool recreation to run, and the
last one was on 11 August. What we can assert today is the query-level proof, which is that a page
sitting past position ten could not previously appear and now does.

## Where we're going

The sister ticket, bug 298 (internal links chosen from at most fifteen candidates), is still open and
deliberately filed with a weaker claim than this one — its population is smaller and whether it has
ever affected a real decision is unmeasured. The general detector built by the 275 lane is committed
but not yet running in production; it ships with the next fleet release, and will then answer the
"has this ever bitten" question for the remaining cases without anyone running a census by hand. Two
traps found along the way have been written into the shared landmine file: that removing a row cap
can unmask a duplicate-row fault the cap was hiding, and that the review council refuses
configuration-only changes by default in a way that reads as "this does not need reviewing".
