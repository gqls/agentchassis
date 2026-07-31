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
