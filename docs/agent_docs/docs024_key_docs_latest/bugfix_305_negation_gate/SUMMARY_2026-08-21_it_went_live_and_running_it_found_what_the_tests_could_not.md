# SUMMARY 2026-08-21 — it went live, and running it found what the tests could not

Second in this lane's series (08-20 *the rule becomes a check* → this). Written because the previous
entry described something waiting for a build, and everything interesting has happened since it got
one: the mechanism went live, was run on real pages, and was wrong twice in ways no test had caught.

## What we're trying to do

Stop a habit of machine writing reaching our pages: saying what a thing *isn't* in order to say what it
is. The owner read three of our pages, quoted two sentences of exactly that shape, and asked for two
things — fix those pages, and make sure that sort of copy never leaves the framework again.

## Where we've come from

The mannerism is partly ours: the instructions for the complained-of site hand the writer a house
tagline built on it, and that phrase reaches the writer in around thirteen hundred prompts. But the
writer also does it unprompted, so fixing the instructions was never going to be enough. Two previous
attempts to fix it with words had failed, and the team who studied that wrote down why: a rule can name
a shape, but what goes wrong is a habit.

So we built a mechanical check instead — one that counts the mannerism, and, on the writer that
produces almost all our page copy, sends the offending sentences back once and pastes the answers in.
It was reviewed four times, refused once, and approved on the fourth.

## What we've done

**It is live.** Both halves: the daily check on our site instructions has been running since the 20th,
and the writing check went live yesterday morning on the fresh build.

**And running it did what a year of testing would not have.** Within three minutes of going live, the
first page showed the repair was **completely inert** — it could not find a model to call, because this
writer keeps its model setting on a different step from the one we had added. It was detecting
perfectly and repairing nothing. We only saw it because a reviewer had insisted, two days earlier, that
"the machine failed" and "nothing needed changing" must not report the same way. That objection paid
for itself in a day.

**Once fixed, it works, and the edits are the right shape.** On its first real page it found five
instances, left one alone because it was a regulatory statement, allowed two under the per-page
allowance, rewrote two, rejected none — and the rewrites were surgical: *"the result breaks down by
area rather than giving you one verdict:"* became *"the result breaks down by area:"*. Nothing else on
the page moved.

**Then the first busy page found a real bug in our own repair.** Where several of these phrases sit in
one block of text, each rewrite was being applied to the *original* version of that block, so they
overwrote one another: six rewrites accepted, one actually applied — and the report cheerfully claimed
six. Fixed, and the fix is now live too.

**I nearly missed it, and that is the part worth telling.** My first check asked whether each rewrite's
text appeared in the stored page. Five of six said yes, and it was meaningless: these rewrites trim the
*end* of a sentence, so the beginning is identical whether the edit landed or not. The honest question
— is the phrase we removed actually gone? — said three of six were still there. We have now written
that lesson down four times in this area, and I still did it, on my own code, with a result I wanted.

## Where we are now

Every defect we have found is fixed and live, and the mechanism has been watched doing its job on real
pages rather than in tests. One confirmation is still outstanding — a busy page built since the newest
build, to watch the repaired version reach the stored page — and it is waiting on ordinary traffic
rather than on anything being wrong.

**And the thing that has not changed since the 20th: this does not fix the three pages the owner read.**
That tagline is in the site's own instructions, which order it onto four different page types, and the
check deliberately leaves alone anything a site's instructions supplied. It counts it, reports it, and
stops. Nine of our twenty-five sites are in the same position.

## Where we're going

Four things, none of them ours alone. The site teams decide whether those nine sets of instructions get
corrected — that is a positioning judgement, not a technical one. The owner or the architecture track
decides whether the counting should run across the whole estate rather than on one writer, which is
written up with its costs. Someone should decide whether "rather than" — which is in nearly half of
everything we write — is a habit worth removing or simply English. And the bug itself stays open until
those three pages are rebuilt, because the fault is fixed while the damage is not.
