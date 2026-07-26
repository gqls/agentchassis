# Where we are — the dead links that the platform spotted and published anyway

Plain prose, append-only, newest at the bottom.

---

**2026-07-26 — what the bug actually was**

Every time we build a page, the platform reads the links in it and checks each one against
the list of pages that really exist on that site. It gets this right. It finds the bad ones,
by name. Then it writes "warning" next to them and publishes the page anyway, because
warnings do not count towards the pass/fail decision.

The comment in the code explaining why that was acceptable said the improvement loop would
come along and fix them later. The improvement loop has been switched off since May. So the
excuse had not been true for months, and nobody noticed because the comment reads as though
someone had thought about it.

**What I measured before changing anything**

The bug report was explicit that we should not just make the check stricter without counting
the damage first, because "a page with one bad link" turning into "no page at all" is worse.
That was the right instinct and the numbers backed it up.

Over the last thirteen days, sixteen page builds went through the check. Three of them had
bad links — nineteen per cent. Seventeen bad links between them, fifteen distinct targets,
and every single one of those fifteen was a page that has never existed in any form; not one
was a typo we could fix. All three pages published.

The two pages affected were the **home pages** of oufe.com and webdesign.co.uk. So if I had
simply made bad links fatal, as the obvious fix, two home pages would have failed to publish
at all. That is worse than what we have now, and it is exactly what the bug report warned
about.

**What I did instead**

The page still publishes, but the bad link does not. If the target is real and the writer
just wrote the address slightly wrong — `/contact` where the page is actually
`/contact.html` — we correct the address. Otherwise we take the link off and leave the words
alone, so a sentence like "read our pricing guide for details" keeps its full meaning and
simply stops being clickable. Nothing gets deleted.

I also made it write down what it did, permanently. Before, the only record was a line in a
log that gets thrown away, so a day after a build nobody could tell you which links had been
wrong. Deliberately *not* a job on a to-do queue: I checked, and that queue has never once
been drained — those items have been created twenty-two times and actioned zero times — so
adding to it would have looked like progress and produced none.

**The thing I found that is arguably bigger than the bug**

We do tell the writer which pages exist. There is a step whose entire job is to hand it the
list and say "only link to these". I checked whether it was working. It is not: on the last
twenty page-writing runs it found zero pages every time, so the instruction is dropped from
the prompt entirely and the model is left to guess. That explains why the bad links are all
invented rather than mistyped.

That is a different fault in a different place, so I have written it up on its own rather
than quietly bundling it in. It is the more valuable of the two to fix next, but it changes
how the writer behaves everywhere, and that deserves its own measurement rather than riding
along with this one.

**Where this stands**

The code is written and tested. It is not live yet — this kind of change only takes effect
when a new image is built and rolled out, and I will confirm it against the running system
rather than against the code before I call it done.

---

**2026-07-26, later — a mistake worth recording**

I nearly shipped a test that could not fail.

The habit is to deliberately break your own code and check the tests notice. I did that, and
it reported four failures out of eight. I assumed four of my tests were weak. They were not:
they had never run at all. One test asked for "the first repair" without checking there was
one, so when there were none it crashed the whole test program, and every test written after
it was skipped in silence. The output of a run that did a third of the work looks exactly
like the output of a run that did all of it.

Fixed, and re-checked: all eight now fail when the code is broken, and the four tests whose
job is to confirm nothing changes still pass. Worth knowing generally — a crash in one test
quietly disables the rest of the file.
