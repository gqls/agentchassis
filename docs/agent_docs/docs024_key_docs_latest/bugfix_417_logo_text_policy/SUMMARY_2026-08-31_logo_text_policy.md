# SUMMARY — 2026-08-31 — the logo that named itself

## What we're trying to do

Stop the platform putting invented brand names on customers' logos. When the site planner writes
the instruction for a logo, it must never be possible to say "put a wordmark on it" without
saying what the wordmark should read — because an image model given that permission fills the gap
by making a name up.

## Where we've come from

The estate learned this lesson once already, months ago, and wrote it into the code with its
reason attached: generated wordmarks come out malformed, and a logo gets re-cut into a favicon,
so logos carry no lettering. That rule has been sitting in the codebase, correct, the whole time.

This morning the loanzy lane found the planner's worked example contradicting it — a phrase that
permitted lettering and never said what it should say — traced it to "Farm Shield Info" appearing
on farmerinsurance.uk, and shipped two migrations: one fixing the example, one washing the
instructions that had already copied it. Both were right, both passed council review, both were
applied and verified.

## What we've done

We found that those two fixes, correct as they were, could not bound the problem — and the two
reasons why are the substance of this work.

The first is timing. Boxingonline's logo instruction was written **41 seconds after** the planner
was fixed, by a job already in flight. Fixing a source cannot reach what has already left it.

The second is language. The wash searched for the exact sentence; boxingonline's copy had been
*reworded* — "other than" where the original said "outside". Identical meaning, invisible to the
search. **The model paraphrases what it is given, so a literal search will always find some and
never all** — including the search you write afterwards to check the fix worked.

That pointed at the real fault, which turned out to be written down in the bug file all along
without anyone reading it as the diagnosis: **the no-lettering rule was attached to where the
instruction came from, not to what was being made.** It lived inside the fallback that runs only
when a plan supplies no logo instruction — and every planner-built site supplies one. The rule
was protecting exactly the sites that never needed it.

So we moved it. It now applies at the single point every logo request passes through on its way
to the image model, overriding whatever the instruction says rather than trying to recognise the
bad wording. That covers reworded instructions, future producers, and requests already sitting in
the queue — none of which a migration can reach.

We also measured something that corrects a belief we had been carrying: the note in our files
saying the image provider ignores negative instructions is that bug's *pre-fix* state, and we
fixed it long ago. The proof was in the adapter log for the failing generation — the model was
explicitly told "no text" and lettered BOXING NEWS anyway, because the same prompt also permitted
a wordmark. **A permission beats a prohibition in the same prompt.** The fix did not change, but
the reason did, and the old reason pointed at a cheaper repair that could never have worked.

The owner ruled that logos carry no words, with one deliberate exception: lettering is allowed
only where someone names the exact string, and the system checks that string really is that
site's own name. That keeps farmerinsurance's requested wordmark legal and repeatable, and makes
"a wordmark, unspecified" impossible to express. It was also necessary rather than optional —
eight live sites word their logos on purpose, and four of them never use the word "wordmark",
which is the paraphrase finding proving itself a second time.

## Where we are now

The wash for boxingonline's instruction is applied and live; a regeneration will now produce a
text-free mark. The structural fix is committed and with the council, and sleeps until the next
chassis build. Tests pin it, and every mutation we could think of — including re-introducing the
original defect — was run alone and broke its named test.

One blind spot is stated rather than left to be discovered: two older parent workflows don't tell
the image action what kind of asset they're asking for, so the new guard would not see them. We
could not establish whether they still run — the probe returned nothing, and returned nothing for
a known-live control too, so it settled nothing and we threw it away rather than reporting it.

And the first paid customer's logo is still wrong, in two separate ways. It says BOXING NEWS on a
site called Boxing Online — and it is not a logo at all, but a two-panel presentation board with
the mark shown twice on different backgrounds, squeezed into the header. That second fault is a
different mechanism, filed separately as 421.

## Where we're going

Read the council verdict and act on it. After the chassis rolls, run the census that could
disconfirm the fix — checking the recorded prompt of every logo generated, because that is where
the guard leaves its mark, and a generation carrying no mark means the guard never ran. Then the
honest test: generate a logo from a fresh paraphrase no migration ever matched, download it, and
**look at it**. Everything else in this story was caught by a human opening an image.

The delivery lane regenerates boxingonline's logo before handover, and the favicon and social card
are re-cut after it lands.
