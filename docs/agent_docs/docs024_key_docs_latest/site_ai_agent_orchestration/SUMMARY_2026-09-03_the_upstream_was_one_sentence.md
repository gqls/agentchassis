# SUMMARY — 3 September 2026: the upstream turned out to be one missing sentence

## What we are trying to do

Make ai-agent-orchestration.com a site where every piece of text can actually be read. That sounds
small. In practice it has been the hardest recurring problem on this site, because the failure is
invisible in exactly the way that matters: the page loads, nothing errors, and a button label is
simply the same colour as the button.

## Where we have come from

We have fixed this four times. Four separate repairs, each one correct, each one shipped — and each
time the fault came back on pages that had not existed the week before. By the start of this week we
had a site with forty-five pages, four known failures, and a strong suspicion that we were mopping a
floor with the tap running.

The note left for this session said as much: stop fixing instances, go and find where they come
from.

## What we have done

We found the tap. The system already works out a correct, readable colour for text sitting on a
coloured button, and publishes it on every site including this one. Two of the prompts that write
our pages know about it. **The prompt that writes our interactive tools was never told it exists.**
So when that prompt needs a text colour, it guesses, and its guess is the sensible-looking one — use
the page's own background. On most of our sites that is right. On this one it is invisible, because
our brand colour and our page colour are nearly the same near-black.

The evidence is the part worth keeping. If the prompt is the cause, the fault should sit in things
built by *that* prompt and not in things built by the others. Of 151 ordinary components, **none**
has it. Of 261 tool components, **148** do. That test could easily have come back mixed. It did not.

We have written the missing sentence into both tool-writing prompts, fixed an audit that was
reporting the correct repair as an unknown mistake, and — this is the part we did not originally
plan — **built a detector**. A reviewer asked the question we should have asked ourselves: a prompt
is *taught*, not *obeyed*, so what would tell us if it were ignored? Nothing would have. Now
something does, and its first run found twenty-five new instances in seven days across five other
sites, six of them created that day.

## Where we are now

The diagnosis is written up, the code is committed, and the detector is live and reporting. **The
prompt change itself is deliberately not switched on yet** — it was still in review when this was
written, and that review has already earned its keep twice.

Because it is worth being straight about: the review caught two real faults in our own work. One
was a change of ours that left a test failing for two hours in a part of the system we had not
thought to run. The other was subtler and more useful — the obvious way to fix that test would have
switched on a warning across the whole estate that was wrong six times out of seven. **A passing
test would have been the failure.** We only avoided it by going and measuring the real sites instead
of doing what the test asked.

And the honest limit: **none of this repairs a page that is broken today.** It stops new ones
arriving. The four known failures on this site are still there and still need a rebuild pass. If
someone looks tomorrow and sees them, that is expected.

## Where we are going

Switch the prompt change on, wait two days, and run the detector: if new tool pages stop arriving
with the fault, the tap is off. Then go back and repair the pages already broken — starting with the
nine colour schemes across the estate where this fault is actually visible, rather than all of them.

After that, two older items that have been waiting: a shared contact form used on twenty sites, and
two pages on this site that our measuring tool still cannot read at all.

The wider point, which is not really about colour: a rule that lives in one prompt and not its two
siblings is invisible to everything we have. Nobody would have found this by looking at the site. We
found it by asking why the same repair kept being needed, which is a question worth asking more
often.
