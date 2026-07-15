# Voice rewrite prompt — leopardessconsulting.co.uk

The site copy still reads as LLM-written: fluent, evenly-cadenced, and slightly hollow.
This is the prompt to fix that, page by page. Feed it to whatever writes the copy (an
agent, or a person, or me in a rewrite pass), together with the specific page's current
text and the facts it may draw on.

**Two hard rules before anything else.** (1) Every factual claim must trace to a row in
`AUDIT_verified_facts.md`; if it doesn't, cut it or mark it plainly as something we *could*
do, not something we *have* done. (2) Rewrite the *substance*, not just the surface — if a
sentence has nothing specific to say, delete it rather than repolish it.

---

## The brief for the writer

You are the person who actually built this platform, writing to someone who runs a business
and is tired of being sold AI. They are clever, busy, and sceptical because everything they
have read about AI has been either breathless or vague. You earn their attention by being
specific, by admitting limits, and by sounding like a real person who has done the work —
not by being polished.

Write the way you would explain your work to a sharp friend over coffee who happens to run a
logistics firm: plainly, concretely, and without performing. If a sentence would make that
friend think "that's just words," it fails.

## What "sounds like an LLM" means here — and what to do instead

The current copy has these tells. Hunt each one down.

1. **The neat triad.** "observability, fault isolation, and cost controls." Real speech
   rarely lists three balanced things. Use one concrete example, or two uneven ones. If you
   catch yourself writing a third item to make the rhythm nicer, delete it.

2. **The summarising flourish at the end of a paragraph.** "…and that is the point." "…and
   that is how it should be." "That directness is how we earn the engagements worth taking."
   These add nothing. End on the last real thing you had to say and stop.

3. **The "not X, but Y" strawman.** "Not a framework that works in demos, but a system that
   ships." You are inventing a villain to look good next to. Say what the thing is. Let the
   reader supply the contrast.

4. **Throat-clearing openers.** "In today's landscape…", "When it comes to AI…", "Most
   businesses know that…". Start with the actual sentence.

5. **Even cadence.** Every sentence roughly the same medium length, joined by em-dashes and
   semicolons into a smooth hum. Break it. Use a short sentence. Then a longer one that
   carries a specific detail with a number or a name in it. Vary deliberately.

6. **Abstraction where a concrete noun belongs.** "streamline operational workflows" →
   "take the Tuesday-morning report off whoever currently spends two hours building it."
   Name the actual thing: Companies House, a spreadsheet, an inbox, a registration number in
   a website footer.

7. **Confidence words doing the work of evidence.** "powerful", "robust", "seamless",
   "cutting-edge", "enterprise-grade". Delete them. If the thing is good, show the fact that
   makes it good ("it has run for five months and resumes from the step it failed on").

8. **Hedged bigness.** "helps businesses unlock the potential of AI." Every word is soft.
   Replace with the smallest true statement: "checks 2,767 business records against
   Companies House and asks a person when it can't be sure."

## What good looks like (write toward this)

- **Concrete before abstract.** A reader should be able to picture the job. "A pipeline that
  reads a company's website, finds the registration number in the footer, and confirms it
  against Companies House" beats "automated entity verification."
- **Numbers with a source.** "5,652 news items scored, 4,672 of them for how much the source
  can be trusted" — and say it's from our own database. Small and checkable beats large and
  round.
- **Admit the edge.** "The matcher only looks at companies that are still active, so it skips
  dissolved ones rather than reconciling them. That's a limitation, not a feature." A sentence
  like this buys more trust than a paragraph of confidence.
- **The metaphor earns one sentence, then gets out of the way.** "Leopardess" is the name of
  the practice. Don't hunt, prowl, or pounce. If precision-and-quiet-persistence earns a line,
  let it, then return to substance.
- **Say "could," and mean it.** Where we describe something we haven't built: "We haven't done
  this for a client yet. The nearest thing we've built is X; pointing it at your problem would
  be real work, and we'd rather scope it than promise it."

## Method (do this per page)

1. Read the page's current copy and list the *actual* points it makes. Discard the rest.
2. For each real point, find the most concrete, specific way to say it — a named tool, a real
   number (from the audit), a picturable task.
3. Write it in varied sentences. Read it aloud. Where it hums evenly, break it. Where it ends
   on a flourish, cut the flourish.
4. Check every claim against `AUDIT_verified_facts.md`. Anything unsupported: cut or mark as
   "could."
5. Run the banned-word/tell list (above and in `specs/voice.json`) as a final pass.

## One worked example

**Before (LLM-flavoured):**
> Leopardess Consulting designs and deploys production-grade multi-agent AI systems with
> observability, fault isolation, and cost controls built in from day one. We help businesses
> move beyond prototypes and unlock the full potential of automation — reliably, securely, and
> at scale.

**After (in voice):**
> We build AI systems that do one defined job and keep doing it without someone watching. The
> one we're proudest of checks scraped business records against Companies House — it reads the
> registration number straight out of a website footer where there is one, and when a match is
> genuinely uncertain it stops and asks a person instead of guessing. It has verified 2,767
> records so far. That is the sort of work this is good at: narrow, checkable, and dull in the
> best way.

Note what changed: the balanced triad is gone; the abstractions became a nameable task with a
real number; the "unlock the full potential… reliably, securely, and at scale" flourish is
deleted; and it ends on a plain, slightly self-deprecating true sentence instead of a sell.
