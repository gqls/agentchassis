# SUMMARY 2026-08-24 — the check now looks at the tool

## What we're trying to do

Every site we run keeps a register of checked facts: a figure, where it came from, and a
link to the source, re-verified every night. It governs what a page is allowed to **say**.
It has never governed what a **calculator works out**. That gap is not theoretical — a
stamp duty tool on one of our sites used a tax threshold that had expired sixteen months
earlier, while the correct figure sat in the register beside it, re-checked that morning,
and every check we own passed the page.

## Where we've come from

In mid-August we built the first half of the answer. A calculator can declare which
registered facts it uses, and the nightly sweep tells it when one of those figures moves.
That went through three rounds of review, two of which found real defects, and it is
proven working.

Then the lane went quiet for a week.

## What we've done

We picked it up and re-counted first, which turned out to matter more than anything else
we did. When the bug was written there were 143 facts across 12 sites; there are now 294
across 15. The register has more than doubled in eight days. In the same period, the
number of calculators that declare anything went from zero to **one** — and that one only
because we asked another team to add it by hand. There are 178 calculator pages on sites
that have a register.

So the machinery worked and almost nothing was plugged into it, and the gap was widening.
Worse: on the single calculator that had adopted it, the sweep had filed thirteen requests
asking a person to confirm the figures by hand, and a week later all thirteen were still
sitting untouched. We had built something that asks a question nobody answers.

We built four things.

**The declaration can no longer fail silently.** Two safeguards that the written record
said were in place turned out not to be: a validation rule that had never actually run on
these documents in the place they are written, and a parser that responded to a broken
declaration by doing nothing at all — *including* skipping the warning whose only job was
to say it had done nothing. Both are fixed, and a broken declaration now leaves a note
attached to the calculator it belongs to.

**The one tool that reads a calculator's raw code can now be aimed at the facts that
matter.** It could not be before: the code skipped past it for exactly the kind of fact
that carries a legislated figure. Its address was also a reference that dies whenever a
page is rebuilt — the original bug's own reference is already dead.

**A check that looks inside the calculator's code for the registered figure.** It only
reports for now; it changes no decisions. It runs for a month, we measure how often it is
right, and only then do we let it act on anything.

**And the adoption piece.** Rather than asking people to hand-write 178 declarations, the
sweep now proposes them, with a ready-to-paste answer. Fifteen such bindings exist today
across three calculators — one of which is our **second** stamp duty calculator, currently
protected by nothing.

## Where we are now

All four are committed and reviewed. The first went through the review council and was
approved first time. The third and fourth came back for revision, and the reviewer was
right: we had made a subtle version of the very mistake the work exists to prevent, which
we then fixed and proved.

The single most useful thing we learned is worth stating plainly, because it nearly cost
us the whole exercise. The obvious way to check a calculator is to search the page for the
registered figure. **That would not work, and it would look like it was working.** Our own
system writes the registered figure into the page's *wording* — that is what the register
is for. So on the original bug's page the text says the correct figure while the code
underneath says the old one. A check that searches the page finds it in the prose and
declares the calculator healthy. It would have passed the exact bug it was built to catch,
every day, for sixteen months. The check therefore reads only the code and ignores the
prose entirely.

None of it is live. It is all code, and code here does nothing until the next fleet build.

## Where we're going

The next build, then a proof at the running system rather than at the version number.
Then one real fact retyped so the raw-code checker finally has something to check — it has
none today. Then a month of the new check reporting without acting, so that when we do let
it act we are deciding on measured numbers rather than on confidence.

What none of this does, and we should keep saying so: nothing here can tell a correct
figure from a confidently wrong one. Everything assumes the register is right. If the
register and the calculator are wrong in the same direction they agree, and every check we
have stays silent. That is a separate piece of work and it needs its own design review
before anyone builds it.
