# Where we are — the silent render fallback (bugs_open/260)

Plain-prose running log, append-only, newest at the bottom.

---

## 2026-08-19 — what this bug actually is, and why the "no damage" line was hiding the cost

Picked this up on your instruction. Short version: the bug is real, it is getting worse rather
than settling, and the reason it looked harmless is that its damage is invisible to the obvious
way of looking for it.

**What goes wrong.** Our page sections are built from reusable templates — HTML with
instructions in it, like "if there is a heading, print it here" and "for each step in this list,
print one of these". Those instructions are meant to be carried out and then vanish, leaving
only finished HTML.

When the proper template engine hits a problem, our code does not stop. It quietly hands the job
to a much older, simpler piece of code and carries on. That older code speaks a **different
dialect** of instruction from the one every one of our templates is written in. So it fills in
the individual words it recognises and leaves every instruction it does not understand sitting
in the page. Nothing logs an error. The output is well-formed HTML, so nothing notices until a
final check refuses the page for containing template gibberish.

**What makes the engine stumble is almost always one field of the wrong shape.** A component has
a list — say, a set of steps, each with alternatives. The field is defined as a *list*. The
writer writes it as a *sentence*. Asking the engine to go through each item of a sentence is
meaningless, so it gives up, and the fallback then wrecks the whole section rather than that one
field.

**Why "no live damage" was the wrong summary, and I have corrected it.** The original
measurement was sound and I re-ran it today: nothing broken has ever reached stored content.
The final check refuses the page first, every time — I confirmed zero damaged sections out of
1,789 stored, and zero out of 72 stored headers and footers. That part holds.

But that phrase got compressed into "no harm done", and it is not. **A page this bug refuses is
a page that never exists — while the pages that did build carry on linking to it.** Two other
lanes found this independently while I was reading: loanzy.uk serves a 404 where its
consumer-rights page should be, and remortgagecalculator.uk has two dead links in the navigation
of every page it serves. A count of stored rows reports this bug as harmless **precisely because
its whole effect is that nothing gets stored.** That is survivorship, not safety. The headline
now says so.

**How often it fires.** Twenty-six occurrences across seven domains between 11 and 18 August,
twenty-five separate pieces of work, twenty-four of them still sitting in the human-review queue.
The file's last count was eleven across four domains three days ago. The worst single day so far
was the 18th — nine occurrences across three sites. It is accelerating, and it fires on brand-new
sites built from nothing as readily as on established ones.

**The two things nobody had measured, and they settle the argument.** The proposed fix is to
delete the fallback, so a template failure stops the build with a clear error instead of quietly
producing a broken page. The obvious objection is "what breaks if we remove it?" — so I measured
it two ways, each with a deliberate wrong answer built in to prove the test could fail.

- Not one of the 251 live component templates fails to load. So the fallback is never reached
  because a template is malformed — it is only ever reached because the *data* was the wrong
  shape.
- Not one of the 1,778 stored page sections fails to render against its own stored content. So
  nothing currently working depends on the fallback to keep working.

Add the earlier finding that none of our 253 components is even written in the dialect the
fallback speaks, and the conclusion is clean: **deleting it changes the behaviour of nothing that
works today.** Its only observable effect is to turn a clear error into a broken page.

**One thing improved while nobody was looking.** The bug file argued that the complementary fix
— checking the writer's output against each component's declared field shapes before rendering —
would be nearly useless, because only four components used the schema format it would understand.
That has changed: those four have been converted or retired, and today **107 of the 110
components at risk carry a schema the check can read.** So the cheaper, complementary fix is now
worth having, where a week ago it would have shipped and done almost nothing. I have corrected
that in the file, because a future session would otherwise have read the old warning and dropped
the idea.

**One constraint I would not have found in the code.** You have ruled that every site should be
able to have tools. Tool pages legitimately contain template-looking text in their copy — a page
explaining template syntax, or a prompt library, will have braces in it on purpose. One of my
twenty-six occurrences is exactly that: harmless content that merely looks like the bug. So
whatever we build has to tell "the renderer failed" apart from "this page is about templates",
and the test for it has to include a good tool page that must still pass. A check tested only
against broken pages cannot notice it has started refusing working ones.

**Where it stands.** I have claimed the fix — five lanes have contributed to this bug and every
one of them explicitly said it was not fixing it, so it was genuinely unowned. The design is
being drafted now and will go to the review council before it ships. Two other lanes are holding
reproductions for me: one has a site locked in the failing state, and the other will run a
complete build from scratch once the fix is live and report back either way. **The fix is code,
so nothing changes until an image is rebuilt and rolled** — I have told both lanes that, since
one is sequencing an end-to-end test around it.

**One mistake of mine, recorded.** I told two lanes that some of the parked work items were this
bug being counted twice. They were not — I reasoned from the name of the item type while the
evidence that contradicted it was already on my screen. One lane said it would have written my
version into its own notes as fact. Corrected with both, and logged in the fleet-wide
wrong-calls file.

---

## 2026-08-20 — the fix is written, reviewed once, and committed; it is not yet live

**What we changed, in plain terms.** When the proper template engine hit a problem, our code
quietly handed the job to a much older renderer that speaks a different dialect of instruction.
That older code filled in the words it recognised and left every instruction it did not
understand sitting in the page. It is now deleted. The renderer either carries out the
instructions or it says it could not — there is no third outcome any more.

That change breaks, deliberately, every piece of code that asked the renderer for a page and had
no way to be told "I could not render it". There were fifteen of those. Each one now makes a
decision you can read: the page builder **stops and names the field**; the repair path **keeps
the good page it already has** and asks the writer to fix the content; the site header/footer
code **falls back to a plain version** rather than shipping gibberish; the two paths that edit a
page that is already live **refuse the edit and leave the live page alone**. That last pair is
the one worth knowing about: they write straight to a published page with no check in between,
and the guard they did have could never have caught this problem.

**The review found three things I had genuinely got wrong**, which is the best argument for
sending work through it. It caught that I had left three of the fifteen sites unwritten while
claiming they were done; that I had described a known-unsafe piece of Go behaviour as if it were
a safety feature; and that a "nobody had ever noticed this" claim of mine was simply not true —
three earlier pieces of work had noticed it, one of them had even half-fixed it. All three are
corrected. It cost fifteen minutes.

**And writing the deployment file caught a mistake nothing else would have.** Before turning on
the optional early check, I asked what it would refuse if we switched it on today. The answer was
"nothing, except five items on one page" — a live, perfectly healthy page on fundamentallyai.com,
where those five items are simply blank. My check was treating "blank" as "the wrong type". Had we
switched it on, that page would have stopped rebuilding, and it is the only page on the estate
shaped that way, so no test I would have thought to write would have found it. It now shares one
definition of "blank" with the check that already existed, so the two cannot disagree.

**One rough edge, and it cut both ways this morning.** We all share one working copy of the code.
Another session's commit accidentally took two lines of my work with it, which left the shared
code unable to compile for anyone — nothing in the usual status view would tell you why. My fix
for that then took two lines of *their* work in the same way, and briefly broke it again
differently. Both are repaired, both are recorded, and I have written down the ninety-second check
that catches it (build the committed code in a separate scratch copy, not in the shared one). It
is not a fault in anyone's care; it is what happens when two people edit one file and either of
them commits.

**Where this leaves things.** The code is committed and reviewed once; a second review round is
running. It does nothing until the next fleet release builds and ships it — Go changes are inert
until then, so nothing has changed on any live site today. After the release, the loanzy lane runs
a clean build from scratch and reports either way; that is the real test. The optional early check
stays switched off until then, and its switch-on file is written and deliberately held back.

**What to expect when it does ship, stated so it is not misread as a regression:** builds that
would previously have failed late with twenty confusing "blockers" will now fail **early, naming
one field**. The twenty-four pieces of work sitting in the review queue still hold content of the
wrong shape — making that content correct is the writer's job, not this change's. What this buys
is that the failure is honest, immediate, names the field, and can no longer reach a published
page through the two unguarded routes.
