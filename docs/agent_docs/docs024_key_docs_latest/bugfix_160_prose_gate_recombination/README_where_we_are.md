# Where we are — the report gate that destroyed a good report

**2026-07-31, evening.** Picked this up straight after finishing the writer-link-constraints
bug. It is a small one with a nasty shape.

The gripper report pipeline has a deliberate safety gate at the end. Before a report is turned
into a page, a checker reads the prose the model wrote and refuses it if it contains any number
or any product name that did not come from the measured facts. That gate is the reason we can
put these reports in front of a customer at all — it is what stops the model quietly inventing
a model number that sounds right. When it refuses, nothing is published: no page, and the
address 404s. That is the correct design and it is staying.

What went wrong is subtler. The checker asks "does this exact string appear in the facts?" But
the writer does what a good writer does — it *combines* things. The facts said the enclosure
must be **IP54**. The writer wrote "IP54-or-better". That phrase, as a whole string, appears
nowhere in the facts, so the checker decided the model had invented a part number, and the
entire report was destroyed. A correct sentence killed a correct report.

Three things make it worse than it looks:

- It is **intermittent**, because it depends on how the writer happens to phrase things that
  run. The identical request passed the same gate four days earlier. So it looks like a flaky
  system rather than a rule, and re-running "fixes" it — which is the worst possible signal,
  because it teaches you nothing is wrong.
- The error message says the writer "names a model-like token not in the candidate set", which
  reads as *the model hallucinated* — the exact opposite of what happened.
- It is a whole family, not one phrase. Anything hyphenated and technical is exposed: IP
  ratings, thread sizes, flange codes. We found four more by test.

**What I changed.** The gate now allows a token when the *technical part* still traces to the
facts and only ordinary English has been attached to it. "IP54-or-better" is allowed because
IP54 is in the facts and "or" and "better" are English words. An invented model number is still
refused, because the part carrying the digits still has to appear in the facts.

The obvious version of that fix is a trap, and the bug report itself warned about it: if you
allow any "non-numeric" piece to be attached, you also allow someone to invent "2F-85-X" or
"2F-85-XL" out of a real "2F-85", which is precisely the fabrication the gate exists to stop.
So the rule is narrower: what you attach must be a lower-case word of at least two letters.
Each of those three conditions is there to block a specific fake, and I tested them one at a
time by deliberately breaking each condition and checking the right test failed.

**That last check earned its keep.** One of the three did *not* fail when I broke it — which
meant my test for it was worthless. I had tested "no single letters allowed" using an
upper-case X, which a different condition already blocked. So the condition I thought I had
proved was never exercised. Changed the test to a lower-case letter and it now fails properly
when broken. Worth saying plainly: the check that looks like box-ticking is the one that caught
me being wrong about my own work.

**Where it stands.** Fix written, nine assertions across two new tests, whole package green.
Submitted to the reviewer council. Committed. It is not fixed *in production* until the chassis
image is rebuilt and rolled — the code is inert until then — so the bug stays open until I have
grepped the running pod and re-run a report end to end.

Nothing needed from you on this one.

---

**2026-07-31, later.** The reviewer council sent it back, and it was right to.

My rule allowed a word to be attached to a real rating if it *looked* like an English word —
lower case, at least two letters. The compliance reviewer pointed out that "not" looks exactly
like "or". So a report could have said **"IP54-not-rated"** and sailed through a gate whose
entire job is to stop a report claiming something the facts do not say. I had moved the
problem rather than fixed it: instead of inventing a model number, the writer could now invert
a specification. A second reviewer found the same hole from another side — "eighty-five" is
also just a lower-case English word.

I had considered a fixed list of permitted words earlier and talked myself out of it, on the
grounds that no such failure had ever been seen. That reasoning was wrong in a way worth
remembering: **the failure had never been seen because it was not possible until my change
made it possible.** You cannot judge a rule by the examples that have already turned up when
the rule is what decides which examples can turn up.

So the fix now uses a fixed list. A word may be attached only if it is one of about thirty
permitted ones — connectives that claim nothing ("or", "and"), or words that restate the
requirement or ask for at least it ("rated", "compliant", "better", "minimum"). Nothing that
negates, reverses or replaces, and no number-words. If a legitimate word is ever missing, the
cost is one rejected report and a one-line addition; the cost of the alternative was a report
that contradicts its own specification.

**The uncomfortable part, and the reason I am writing it down.** I test these rules by
deliberately breaking them and checking the right test fails. Three separate times today that
check caught a test of mine that was passing for a reason I had not noticed — the last one
being the worst: the whole set of negation tests never reached the rule they were written for,
because the test data did not contain the rating in the first place, so every case was being
refused earlier for a completely different reason. Green tests, right assertions, proving
nothing. That is not a tooling problem, it is what a negative test is: it tells you the input
was refused, never *which* rule refused it.

Resubmitted, committed, image built and waiting. Still nothing needed from you.

---

**2026-07-31, done.** Approved on the second round and live.

The reviewers had one more thing, and it was a fair catch: I had written that the permitted-word
list followed two rules, and the list actually contained a third kind of word. That matters more
than it sounds — a list whose stated rule does not describe its contents cannot really be
reviewed by anyone, including me. So I wrote the third rule down honestly and, having done that,
removed three words that did not fit any of them ("series", "style", "type" all name a product
family beyond the thing they are attached to, which is exactly what the rule excludes).

Then I proved it on the real thing rather than on test data. The report that was destroyed this
morning still had its actual text stored in the database, so I pulled it out and ran it through
the gate both ways: with the fix switched off it fails with the identical error, word for word;
with the fix on it passes cleanly, and the phrase that caused all this is still in the summary.

It is live on the running system, checked on both machines rather than assumed from the deploy.
The bug is closed.

One deliberate omission, so it is not mistaken for an oversight: I did not generate a fresh
report end to end. It would not have proved anything — whether the problem appears depends on
how the writer happens to phrase that particular run — and it would have published a page on the
live site that I have no way to remove. The test above is stronger evidence and costs nothing.
