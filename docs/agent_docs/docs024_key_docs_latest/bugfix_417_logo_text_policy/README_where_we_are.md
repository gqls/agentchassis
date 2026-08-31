# Where we are — the logo that named itself

*(Plain-prose log for the owner, append-only, newest at the bottom.)*

## 2026-08-31, evening

The short version: the site planner was telling the image model it could put words in a logo,
but never telling it what the words should say. So the model made a name up. On
farmerinsurance.uk it invented "Farm Shield Info". On boxingonline.com — your first paying
customer — it put **"BOXING NEWS"** on a site called Boxing Online, and that logo has been
sitting in the header of all 19 pages, with the favicon and the social-share card both cut
from the same picture.

Another session had already found this and fixed the planner this morning, and washed the
prompts that had copied it. Both fixes were right. Both were also, by this evening, provably
not enough, and the way they fell short is the interesting part.

**First:** boxingonline's logo instruction was written **41 seconds after** the planner was
fixed — by a job already in mid-air. You cannot fix a source and thereby fix the things that
already left it.

**Second, and worse:** the wash searched for the exact sentence the planner used. Boxingonline's
copy had been *reworded* — "no text other than the wordmark" instead of "no text outside the
wordmark". Same meaning, different words, invisible to the search. The model doesn't copy
instructions, it paraphrases them, so **searching for the exact wording will always find some of
them and never all of them.**

That pushed me to a different kind of fix. Rather than chase the wording, I moved the rule to the
one place every logo request has to pass through on its way to the image model, and made it apply
there regardless of what the instruction says. It doesn't need to *recognise* the bad wording;
it overrides whatever it finds. That also covers requests already sitting in the queue, which no
amount of fixing the planner could have reached.

One thing I got wrong and want on the record. I started from a note in our own files saying the
image provider ignores "don't do X" instructions. It turned out that was true *before* we fixed
it, months ago, and we did fix it. The proof was in the logs: for the boxingonline logo, the
model was **explicitly told "no text"** — and lettered BOXING NEWS anyway, because the same
prompt also said it could have a wordmark. A permission beats a prohibition in the same
sentence. My fix didn't change, but my reason for it did, and the old reason would have pointed
a reviewer at a cheaper repair that could never have worked.

You ruled tonight that logos carry no words. I've built that as the default everywhere, with one
deliberate exception you can use: a logo may carry lettering only if someone says **exactly what
it must say**, and the system checks that text really is that site's own name. That keeps
farmerinsurance's wordmark — which you asked for — legal and repeatable, and it makes "put a
wordmark on it, I don't care what it says" impossible to express, which is the sentence that
caused all this. Worth knowing: eight sites already have deliberately worded logos, so a blanket
no-words rule would have quietly broken them the next time they regenerated.

Where it stands: the wash for boxingonline's instruction is **live now**. The main fix is
committed but sleeps until the next chassis build goes out. It's with the review council.

Two things still owed, neither of them mine to do: boxingonline's logo needs regenerating before
you hand the site over (the delivery thread owns that), and after it lands the favicon and social
card have to be re-cut, because nothing re-does those automatically.

And one thing you should probably see for yourself. Beyond the invented name, that logo file
isn't a logo at all — it's a **two-panel presentation board**: the mark on dark navy on the left,
the mark plus lettering on light grey on the right, both squeezed into the header slot. Even with
the words gone, it's unusable. Nothing between the image model and the live page ever checks that
a logo is *one picture*. That's a separate fault from this one and I've written it up separately
rather than bolting it on here.
