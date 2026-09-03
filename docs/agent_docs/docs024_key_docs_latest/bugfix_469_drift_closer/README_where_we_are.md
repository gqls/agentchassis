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
