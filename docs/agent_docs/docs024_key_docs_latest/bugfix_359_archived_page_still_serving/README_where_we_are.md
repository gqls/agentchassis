
## 2026-08-26 (evening) — the meter is built, and what it refuses to do is the interesting part

The detector is written, tested and committed. It is deliberately switched off until the code
is actually running on the servers — turning it on early would break two other working checks,
for reasons I'll come back to.

Here is what it does, in plain terms. Every site already gets a routine health sweep, roughly
every four hours. This adds one question to that sweep: for each page we have retired, go and
ask the internet whether it still answers. If it does, raise a flag a person can see.

**It raises a flag and nothing else, and that was a deliberate choice.** I could have made it
delete the page automatically — we already have the machinery to do that. I didn't, because
there is no way to tell, from the outside, the difference between a page that was retired on
purpose and is wrongly still live, and a page that was retired *by accident* and is quite
correctly still live. Both look identical: a working page at a URL we've marked "retired". The
first needs deleting. The second needs un-retiring. Getting that backwards means quietly
deleting a good page off a customer's website, which is worse than the problem I'm fixing. So
the flag says both things and a person decides.

**The part I want to explain properly is what the check does when it can't tell.**

Every other check of this kind we have reports a *problem* — a broken link, a missing file, a
site that's down. If one of those checks goes blind, it under-reports: it finds fewer problems
than there are. Annoying, but the shape of the failure is honest.

This check is the other way round. Its "problem" is a page that *works*. So if the whole site
happens to be down when the check runs, every retired page looks correctly gone, and the check
reports **nothing wrong**. And "nothing wrong" is exactly what it reports when everything
genuinely is fine. A broken instrument and a clean bill of health would look the same in the
records — which would make the check worse than useless, because people would trust it.

So before it judges anything, it proves its own instrument twice, on each site. It asks for a
page that definitely does not exist and requires a "not found" — because some domains answer
*everything*, and on one of those every retired page would look alive. And it asks for a page
we know is live and requires a "here it is" — because if that fails, the site is down and
nothing else it sees means anything. **If either check fails, it refuses to judge, says so in
a permanent record, and — this is the important bit — is structurally prevented from marking
any existing flag as resolved.** A blind run cannot quietly tidy away real problems.

**One thing I got wrong, because it is the useful part.** Test files here carry a table saying
"break this line, and that test will fail" — the evidence that each safety catch really is
load-bearing. I wrote mine from the design, since I knew which test was aimed at which catch.
Then I actually ran it: eleven of thirteen behaved. **The two that didn't were the two
instrument checks I've just spent three paragraphs explaining.** I could delete either of them
and the test still passed.

Neither test was wrong about what it was checking. Both were passing for the wrong reason: with
the safety catch removed, the run simply tripped over the *next* thing and produced an error
anyway, and the test only asked "was there an error?". So it saw one, and was happy. Both are
fixed — they now set the scene so the damage genuinely could happen, and they insist on knowing
*which* safety catch stopped it. And I've written the mistake down, because a table like that,
written from intent rather than from a result, is a design document dressed up as proof, and
the next person to read it would have believed it.

I've also put the whole thing through the reviewer council, and there is one more thing to do
after the next server update: switch it on, and check it flags the pages we already know are
live and does *not* flag the ones we know are gone. Until it has done both, a quiet result from
it means nothing.
