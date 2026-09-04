# Where we are — the delivery email's instructions promise (bug 475, mechanism half)

Plain prose, append-only, newest at the bottom. What I'd say out loud.

---

## 2026-09-04, afternoon

**The problem, in one line.** When we deliver a finished website, the email tells the customer that
the ZIP file "comes with instructions that walk you through putting it on free hosting". It does not.
There are no instructions in it — just the 45 files of their website.

You found this yourself, reading the email as a customer would, which is the one test nothing else we
have performs.

**It is still true today.** I checked the live email text this afternoon, and the sentence is still
there. That check needed more care than it sounds like: another thread had edited *that very same
email* three hours earlier, for a completely different false promise. So "the email was fixed today"
was true and irrelevant, and it would have been very easy to read it as "this is sorted". Two other
threads warned me about that independently before I could get it wrong.

**Was anyone already on it?** Yes and no, and the distinction turned out to matter. The delivery
thread owns the bug and has been writing the customer instructions themselves — three drafts, the
latest after you performed the whole Netlify sign-up yourself and found that a site uploaded there is
**private by default and looks perfectly normal to the person who uploaded it**. But nobody was
building the *machinery*: the link, the page, the file in the ZIP. So I asked them directly rather
than guessing, and we agreed a clean split. **They keep the words. I build the plumbing.** They have
said they won't write code against this without telling me, and I won't touch their copy.

**Then something more interesting turned up, and it's why this got bigger than a sentence.**

We are about to have **two** letters going to customers — the delivery email, and a follow-up email
another thread has just built. Both fill in the same set of blanks (`{{the customer's site address}}`,
`{{the download link}}`, and so on) from one shared piece of code. But each letter keeps its **own
separate hand-written list** of which blanks it's allowed to use — and the only thing keeping those
two lists in step is a **comment in the code asking the next person to remember**.

You ruled on exactly this shape a month ago: *a comment is not a control on a tree this many sessions
share.*

**And we proved it this afternoon, by accident.** Two threads — theirs and mine — independently and
carefully picked a name for the new "instructions" blank, and picked **two different names for the
same thing**. The only reason we caught it is that one of us had written that comment and the other
thought to send a message. Neither of those is a mechanism. Half a day's difference in timing and we'd
have shipped two names for one page, in two letters, to the same customer.

**So the fix I'm proposing isn't "add the missing link".** It's to make the list of blanks and the
check that they're all filled **the same single thing**, so that a blank can't exist in one letter and
be missing from the other's safety check. The missing instructions link then goes in as the *first*
entry of the new system rather than the last hand-copied one. That's the framework fix; your email
sentence is the case that found it.

**Three decisions came back from you via the delivery thread, and all three helped.**

*The page.* I proposed that the instructions live as a normal page on **webdesign.uk**, built by the
framework like every other page — rather than as a hand-built page bolted onto the links server. You
ruled for it. It means we spend no exception against your "every site goes through the framework" rule,
the page can be corrected later for everyone (including people we emailed last month), and it's
bookmarkable. The catch is that a framework page is the same for everybody, so it can't name a
particular customer's own domain — but their domain goes in the email and in the ZIP's readme instead,
which is *better*, because a customer's own domain is the one thing in the whole set that never goes
out of date.

*The stop-gap.* I pushed for changing the false sentence **today**, since a voucher is out for another
build. **You said leave it** — and you were right in a way my argument wasn't: the next build is your
own trial run, so the person who'd read that sentence is you, not a stranger. I'd stated a real risk
at the wrong severity. Recorded as a wrong call.

*The dates.* There's a slot on the page for "your site stays live until —" and I've been told **not to
wire it**, because we currently have three dates that disagree: the download link lasts 7 days, the
email promises 30, and the tokens run 42. Picking whichever is nearest to hand would be committing
this exact bug a third time, so somebody needs to settle what a customer is actually owed.

**Where it stands right now.** Plan written and agreed with both neighbouring threads. Nothing built
yet. The next step is putting the design through the reviewer council before I write the code — and
I've promised the follow-up-email thread that they'll see the submission first, since it's their file
as much as mine.

**One thing I should say plainly:** you asked for the plan to be prepared using Fable, and Fable ran
out of credit part-way through and returned nothing. I wrote the plan myself rather than stop, and
said so at the top of it. It's worth re-running on Fable later as a second opinion, but it didn't seem
right to leave the work sitting still for it.
