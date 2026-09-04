# Where we are — bug 469, the detector that noticed a page losing a section and then went quiet

Plain-prose running log for the owner. Append only, newest at the bottom. No jargon where a
plain word will do.

---

## 2026-09-03, evening — picking this up

A page on robot-hands.com lost one of its sections some time after 28 July, and nobody was
told. That is the whole of bug 469, and it is worth saying slowly because the mechanism is
genuinely confusing.

**The set-up.** A page's list of sections — what blocks appear on it, in what order — is
stored in three different places at once. There is a plan (the authority), an older-style
plan document that only some sites still have, and a copy on the page's own row. That third
one is just a photocopy: every time the page is rebuilt, whichever of the first two knows
about the page is copied down over it.

**The trap.** If you correct a page's composition by editing only the photocopy, the next
rebuild undoes you. Not the next re-plan — the next *rebuild*, which happens all the time
and nothing warns you. Nothing errors; the page simply comes back with its old sections.

**What happened.** On 24 July someone deliberately added a spec-sheet block to the gripper
catalogue page, in preference to a product grid whose price and rating fields would have
been empty and would have invited the writer to invent numbers. That was an owner-backed
call, and it was written to the photocopy. Four days later the estate's own detector noticed
the two copies disagreed and raised a flag. Then nothing happened for thirty-seven days, the
page rebuilt at some point, and the spec sheet was gone.

**Why nobody noticed is the part actually worth fixing.** Three things compound:

- the flag has no handler, and nothing on the estate ever closes one of these flags;
- the flag records what it saw *on the day it was raised*, and that record never updates —
  so a month later it reads as though it were describing today;
- and while a flag is open, the detector cannot raise a new one about the same page. It goes
  blind on exactly the pages it has already flagged.

So the sequence is: problem detected → flagged → nobody owns it → the rebuild wins → the two
copies agree again → and the flag now describes something that no longer exists and looks
like it might have been resolved.

## What I have actually checked today, so nobody has to take it on trust

**The problem is real and still there.** The spec sheet is genuinely absent from the gripper
catalogue page — I checked all three stores *and* the live blocks on the page, and all four
agree it is not there. It is alive and well on two other pages, so the block itself was never
lost; only its membership of this page.

**The page is still being served to visitors.** This surprised me. The page is marked
"archived" in the database, which sounds like it is out of use — but I fetched the actual URL
and it returns a normal page. So this is not a filing-cabinet problem; a visitor going to
that address today sees a page with a section missing.

**Nothing else on the estate is currently drifting.** I compared the plan against the
photocopy for every page that has both — 398 pages, plus 34 more on the older-style plan —
and they all agree. The lane that filed the bug cleaned up the backlog earlier today. So
this is not a fire; it is a gap in the safety net that will catch the next one.

**And this is not one detector's problem.** Of the 71 detectors the estate runs, only 19 can
ever close their own findings. Ten raise flags that no handler picks up and that nothing ever
closes. The oldest such flag on the estate is 172 days old. So the accumulation is a pattern,
not a quirk.

## The one thing I nearly got wrong

I was about to propose building a new archive so that a destroyed section list could be
recovered. Before designing it I spent one query checking whether we already record this —
and we partly do. Since early August, every deleted block is archived with its name,
position and full HTML. So the *content* of a lost section survives; what does not survive is
the *list* — the fact that the page used to have five blocks in that order. That is a much
smaller gap than I thought, and it means pointing at what we already have rather than
building a second copy of it.

I have written that up as a near-miss rather than quietly moving on, because "reuse what
exists before building new" is a standing rule here and I was one step from breaking it.

## Where I need a decision from you

**Should the spec sheet go back on the gripper catalogue page?** The lane that owns that site
says yes, and has the receipt: it was a deliberate July decision, with a stated reason, never
reversed.

**But I have stopped short of doing it, and I want to be straight about why.** Putting it
back is harder than the two similar repairs done earlier today. Those aligned a plan to a
page that was already correct. This one has to insert a block back into the middle of the
plan and shift the others down — and then, because the page is marked as already built from
the current plan, the system will decide it has nothing to do and never actually rebuild it.
The correction would sit in the database and never reach the visitor.

Getting past that means withdrawing the page's "already built" stamp, and **whether we are
allowed to do that is precisely the open question in RFC_064**, which another lane put in
front of you today. So rather than force it, I have written the repair out in full, with all
its safety checks, and held it. Say the word on RFC_064 and it becomes a one-command job.

There is also a smaller question sitting behind it: this page is marked archived, and I do
not know whether an archived page is even reachable by the rebuild machinery. If it is not,
the repair needs a different last step. I have flagged that rather than assume.

## 2026-09-03, later the same evening — the fix is built, and one thing it refuses to do

The safety net is written and committed. It is inert until the next fleet rebuild, and then
it starts working on its own.

**The part I want to be clear about, because it is the whole design.** The obvious fix here
is: "the two copies agree again, so the warning is resolved — close it." I did not build
that, and building it would have been worse than leaving things as they are. The two copies
agree again *precisely because* the rebuild destroyed somebody's work. A closer like that
would have gone round every site turning silences into certificates, automatically, and
nobody would ever have found out.

So the closer asks a different question: not *do they agree now*, but *what did agreement
cost*. If the page kept what the human wanted, the warning closes quietly. If something was
destroyed, the warning can only close by first filing a permanent record naming exactly what
went — written into the database *before* the warning closes, in the same transaction, so
the two can't come apart. That record copies its evidence rather than pointing at anything,
because the warning it replaces gets archived out of the table a reader would look in.

I also made the original warning lead with the damage: instead of printing two lists and
leaving you to spot the difference, it now says which sections the next rebuild is going to
destroy.

**How I know it works.** I sabotaged my own code fifteen times, each time expecting a
specific test to fail. Fourteen did. One didn't — and that one taught me something. The test
I thought was guarding a particular line was actually being rescued by a different guard
further down the function. The protection was genuinely there; my *proof* of it was not. I
sharpened the test rather than accepting the pass. I've written that up rather than tidying
it away, because "I broke it and the test still passed, so the line must be fine" is exactly
the wrong conclusion.

**Something I nearly built that we already had.** I was one step from proposing a whole new
archive so a lost section list could be recovered. One query showed we already keep the
deleted blocks — names, positions, full contents — and have since early August. What we
don't keep is the *ordering*. Much smaller gap. The fix now points at what exists.

## What I need from you, and it is three questions about one page

They're cheapest answered together, because the third one can make the other two irrelevant.

1. **May we withdraw a page's "already built" stamp** so a composition repair actually
   reaches a visitor? This is the second question in RFC_064, which another lane put in
   front of you today — and on this page it stops being about tidiness. Without it the
   repair doesn't merely lose its history; it doesn't happen at all.
2. **Does the rebuild machinery reach an archived page?** I genuinely don't know. I'd rather
   tell you that than guess.
3. **Should the gripper catalogue page be serving at all?** It's marked archived and it
   serves a normal page. There's already an untriaged warning about that from another
   detector, eight days old, and eight more like it across the estate. If the answer is
   "retire it properly", questions 1 and 2 stop mattering here.

The repair itself is written, with all its safety checks, and rehearsed against the live
database in a transaction I rolled back — including deliberately feeding it wrong data to
confirm it refuses rather than ploughing on. It is sitting held. Say the word and it's a
one-command job.

## Two things I have deliberately not done

**Nine other detectors** raise warnings that nothing ever closes. The mechanism built here
works for them too, but each needs its own judgement about what closing *safely* means —
and doing nine of them quickly is precisely how you end up with nine of the naive closers I
just argued against.

**The record this fix files goes into a queue we don't drain.** There are 331 warnings of
another kind sitting open, the oldest six weeks. I'm not going to pretend this fixes that.
What it guarantees is narrower and, I think, the important half: a page that lost something
can never again be quietly marked "resolved". Whether anyone reads the record is a separate
problem that already has its own bug number.

## 2026-09-03, later still — a correction to something I told you two hours ago

I said nine other detectors raise warnings "that nothing ever closes". **That is wrong, and
how I got it wrong is more useful than the number.**

I counted the warnings sitting open. But closing a warning *removes it from the list I was
counting* — it moves to an archive table. So that query answers "nothing ever closes these"
no matter what the truth is. It cannot produce the other answer.

Counting the archive too: **six of the eight kinds have real closures.** One has 16 out of
61. Another has 5 out of 72. The biggest has **1 out of 334**.

So people do close these — occasionally, by hand, and it leaves no visible trace afterwards.
Including the warnings on this very bug: two were closed in July by one thread, three days
after they were raised, and neither of us could see it because both had left the table we
were looking at.

**This makes the case for the fix stronger, not weaker.** "Nothing ever closes them" invites
the reply "so nobody can" — which was never the problem. One-in-three-hundred-and-thirty-four
says somebody *did*, once, and you would have no way of finding out. That is exactly what a
machine record fixes: not the possibility of resolution, but its invisibility.

It was another lane that caught it, from two archived rows it had told me to read. I had the
general rule written down in my own notes — *"closing a row archives it out of the table you
queried"* — and applied it to somebody else's claim this morning while writing its opposite
into mine this afternoon. It is corrected everywhere it appeared, including in the code
comment, which is the copy a future reader inherits without asking anyone.

## 2026-09-04 — it's live, and it has correctly done nothing

The fleet rebuilt overnight and the safety net is now running. I checked that properly
rather than trusting the version tag: the tag was one I'd already seen used before my work,
so it proves nothing on its own. Asking git whether my change is inside the build that is
actually running says yes, and the control — a commit made after that build — correctly says
no.

**It has not fired once, and that is the right answer.** There is no drift anywhere on the
estate at the moment, so there is nothing for it to close. I've been careful to record that
as "running and never exercised" rather than "working", because those are different claims
and only one of them is evidenced. The thing still owed is one real case going through it.

I did check it isn't silently broken, which is the failure that would look identical to
"nothing to do": the detector's owning agent ran nine times overnight, and there is not one
error recorded against this check — against 97 such errors on record overall and five since
the rebuild for *other* checks. So the silence is real silence.

**I also nearly told you the opposite.** My first check was to look inside the running
program for my change's fingerprint. It wasn't there — which looks exactly like "it didn't
ship". It turns out a built program records only the single point it was built from, not
everything that went into it, so my change was fully present and invisible to that test. The
right question was one command away and I asked it second. Written up, because every safety
check I ran on that measurement passed; it was simply measuring the wrong thing.

## So what's actually left

**Nothing to build.** Three decisions, all about the one page, and I'd take them together
because the third can make the first two irrelevant:

1. **May a repair withdraw a page's "already built" mark so it actually renders?** Without
   this the correction sits in the database and never reaches a visitor. It isn't a tidiness
   question here — it's the difference between the repair happening and not.
2. **Does the rebuild machinery even reach an archived page?** I don't know.
3. **Should that page be serving at all?** It's marked archived and serves a normal page.
   There's already an untriaged flag about that, and eight more like it. If the answer is
   "retire it properly", questions 1 and 2 stop mattering.

The repair itself is written, safety-checked, rehearsed against the live database again this
morning, and held.

**Handoff for a fresh session:**
`docs/agent_docs/docs024_key_docs_latest/bugfix_469_drift_closer/HANDOFF_2026-09-04_continue_here.md`

---

## 2026-09-04, later — question 2 is answered, and it changes the shape of the choice

I said above, about the second of the three questions, "I don't know." I do now, and the
answer matters more than I expected — it turns three separate decisions into one.

**The short version: the machinery that rebuilds pages does reach this page. But when it
gets to the last step — actually publishing the page — something deliberately stops it.**

There is a safety catch, added on 12 August after a different investigation, whose whole job
is to make sure a retired page can never be quietly republished. It sits at the exact point
every publishing route has to pass through, and it refuses any page marked "retired". It is
running in the live system right now, and it is not hypothetical: it has turned away **308**
publish attempts since mid-August, and **three of those were this very page**, the last on
23 August. I checked those three are the same page by its database id, not by its name, so
there is no chance I'm looking at a similarly-named page on another site.

So the repair we have written and rehearsed would start, run, and then be blocked one step
from the finish. The page would still look wrong to a visitor. **The first question — may we
withdraw the page's "already built" mark — turns out not to be enough on its own.**

**What this means for you: it's now one decision, not three.** While the page stays marked
retired, nothing we do to its content can ever reach a visitor. So:

- **If the page should be live** — un-retire it. The safety catch then stands down, the
  first question becomes the real one, and the repair we're holding can go in and actually
  show up on the site.
- **If the page should stay retired** — then the repair is pointless and I'd withdraw it.
  The genuine problem is the other one: the page is marked retired and yet is still being
  served to anyone who visits it. Taking it down properly is a different route, and the
  safety catch deliberately doesn't block that one.

There isn't a third option where we fix the page's contents and someone gets to see the fix.

Worth adding: this page is already on a list of nine pages flagged back on 26 August for
exactly this "retired but still serving" problem, and nobody has picked any of them up. So
the second option isn't extra work invented here — it's work already queued and waiting.

**One thing I got half-right and nearly stopped at.** My first check was of the rebuilding
machinery, and it showed no reason the page would be skipped. That's true, and on its own it
would have been a confident, useless answer — I'd have told you "yes, it reaches it", and
the next person would have applied the repair and hit the block. What saved it was noticing
that the neighbouring retired pages on the same site hadn't rebuilt either, which sent me
looking further down the line instead of stopping at the first encouraging result.

**Nothing has been changed on the live system.** The repair is still held, the page is still
as it was. This is the decision I need from you.

**One more thing that bears on your choice.** I checked what actually links to this page. On
the live site, **nothing links to the retired page** — but **19 pages link to a different,
active page with almost the same address** (the "index" version). So taking the retired page
down properly wouldn't break a single link on the site.

That sounds like it settles it, and it doesn't, because of what the other page contains: it
carries only a news list, not the real content. So the honest description of today's state is
**19 links pointing at a thin page, while the substantial one sits retired and linked from
nowhere**. That reads as easily like "the page was retired by mistake" as it does like
"finish retiring it" — which is why I'm bringing it to you rather than picking.

I also nearly reported the "nothing links to it" figure on its own. A zero like that is what
you also get when the question is asked wrongly, so I re-ran it against a page I knew was
linked; that one came back with links, which is what makes the zero worth quoting.
