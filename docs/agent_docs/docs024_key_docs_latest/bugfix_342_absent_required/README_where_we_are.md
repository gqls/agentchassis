# README — where we are (plain prose, append-only, newest at the bottom)

## 2026-08-22 — picking the bug up, and what turned out to already be done

Bug 342 is about a quiet way pages lose content: if the writer never supplies a piece of
text the template needs (say a headline that the component's contract marks as required),
the page doesn't fail — it just renders that spot as blank, and then the assembly step
throws the visually-empty section away. The reader never sees an error and neither do we.

The previous lane fixed the biggest part of this: every render now *notices* and *says*
when a required field came up empty, and the two riskiest editing routes (the ones that
write straight onto a live page) also file a work item about it. We checked today and all
of that is genuinely running in production — including a piece the bug file still said was
waiting to ship. The bug file was out of date on that point and we are correcting it.

What is still missing, and what this lane is doing:

1. **The editors still ship the blank.** They notice, they file the note — and then they
   write the damaged section to the live page anyway. We are adding the refusal: if the
   edit would blank out a required field, the edit is declined and the live page keeps what
   it had. This is switched off by default and turned on deliberately per agent, which is
   the house rule for giving the system new powers.
2. **A safety switch that was built but never turned on.** The site-chrome route (headers
   and footers) can file the same work item, but the switch has been off everywhere. We
   measured: turning it on today fires on nothing at all (no header or footer in use even
   has required fields), so it is free to arm now and it protects us the day someone adopts
   a chrome component that does have them.
3. **The "hard part" shrank.** The bug file worried about ~75 components with no contract
   at all. It turns out almost all of those are self-contained tools that don't need one by
   design. Only five real ones remain, each used on exactly one page. That's a small
   tidy-up job for a content lane, not a platform change, and we've written it down rather
   than done it here.

Next: put the code change through the review council, commit it for the next build, and
apply the two small config changes (one now, one held until the new code has rolled).

## 2026-08-22 midday — the review pushed back, and it was right about the important one

We put the change through the review council. It came back "revise" with several
objections, and two of them were worth the round.

The first was a factual one I should have checked instead of asserting. Our config-change
script updates the section-editor's settings by name — and the system has a known trap
where a few agents exist twice, with only the newer copy actually running, so an update by
name can quietly patch the copy nobody uses. I had written a guard against exactly that but
had never confirmed which agents are affected. Checked now: the section editor is not one of
them (four others are), and the guard now names them so a future reader knows what it is
protecting against.

The second changed the design, and it used our own bug against us. My plan turned on
*detection* for the site headers and footers but only added *protection* to the page editor.
That is the same mistake this bug is about: noticing a problem and then letting it through
anyway, on whichever path nobody got round to. So the headers-and-footers path now has the
same protection, sharing the identical decision code — but switched off, deliberately. Right
now nothing can trigger it (no header or footer in use has a required field), so switching it
on would arm something that can never fire; leaving the capability out entirely would mean
the first site that needs it waits for a code change and a deployment. This way it is one
setting away, and we have written down the signal that says it is time: the first alert of
this kind about a header or footer.

Two smaller ones were also fair. Our check that the config change had been applied everywhere
only looked at the top level of each agent's workflow, so it would have reported "all done"
while missing anything nested one level deeper — it now looks in both places. And the
verification we plan to run after deployment now checks three things rather than one: that
the live page is untouched, that the alert was filed, and what the job's own status ended up
saying (we expect it to say "complete" wrongly until a separate known bug is fixed, and it is
better to see that written down than to be surprised by it).

Resubmitted. The code is committed and will ride the next build.

## 2026-08-22 evening — it works. A real edit was refused and the page was left alone.

The new build went out, so we could finish the job. First we checked the running software
actually contained the new code — not by trusting the version number, but by looking inside the
running program on both servers for text that only exists in this change. It does. (One of our
three checks turned out to be worthless: the phrase we looked for already existed elsewhere in
the codebase, so it would have said "yes" no matter what shipped. We caught it, threw that check
away, and wrote the mistake down. The other two were sound.)

Then we switched the protection on, and tested it properly — the whole point being that a setting
which is switched on but does not work looks exactly like one that does.

We picked a page section that genuinely has this problem (three pieces of required text missing)
and deliberately chose one that is not published, so that if the protection failed nothing live
would be harmed. Then we tried to edit it. The result:

- **The edit was refused**, with a message naming exactly which pieces of text were missing.
- **The stored page section was untouched** — byte-for-byte identical, and its "last modified"
  date still reads 17 July. Nothing was written at all.
- **The problem was still recorded** on the work queue, naming the missing fields. This mattered
  to us: refusing must never become the reason a fault goes unnoticed.

And the control: we made a *clean* edit to a different section, one with all its required text
present. That one went through and saved normally. This is the check people skip, and it is the
one that distinguishes a working safeguard from a broken system that has simply stopped accepting
edits.

One nice surprise. We expected three missing fields to be named; only two were. The third gets
filled in automatically from a site-wide default further down the process, so it never actually
renders blank. We had predicted this behaviour when the code was written and captured it in a
test, but this is the first time we have seen it happen for real — and it confirms the safeguard
is correctly conservative: it only complains about text that would genuinely vanish from the page.

One thing we did **not** manage to test, and are saying so rather than letting it look covered:
when an edit is refused, the job record that drove it may wrongly report "completed" because of a
separate known bug. Our test was run by hand rather than from the work queue, so there was no job
record to inspect. That check is still outstanding and needs a queue-driven edit.

Everything is committed. What is left on this bug is small and specific: five components that
have no content contract at all (a content job, not a software one), the header-and-footer
version of the protection which is built but deliberately switched off until something can
actually trigger it, and that one outstanding check above.
