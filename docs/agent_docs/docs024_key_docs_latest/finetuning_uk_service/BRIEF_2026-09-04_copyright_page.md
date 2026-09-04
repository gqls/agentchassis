# BRIEF 2026-09-04 — the copyright page: what we trained on, what we filtered, and what is yours

**For the owner to read before it runs.** Owner, 2026-09-04: *"The examples - we should do a few please
pick say 5 that are safe from copyright and do them. We need a copyright page to explain this too."*

## Why this page has to exist, and why it is not boilerplate

A site that sells fine-tuning, and demonstrates it by imitating named writers, invites exactly one
question from a journalist, a customer's lawyer, or a rival: **whose work did you use, and were you
allowed?** The page answers it before it is asked, in public, in our own words. That is worth more than
a terms clause nobody reads.

It also has to answer a harder second question, and the researcher's sentence is the one to build on:
**"We only used Project Gutenberg" is not a defence.** Every problematic passage in the candidate
corpora — coercion in Pepys, racial caricature in Leacock, antisemitism in Saki — is on Project
Gutenberg, free, and findable in seconds by anyone who reads our launch. **The filter is the
deliverable, not the source.** A page that hides behind "it was free and legal" reads as naive; a page
that says "here is what we removed and why" reads as competent.

## The five sections

1. **hero** — one line: what we trained the demonstration models on, and why we can. No hedging.
2. **generic-text-block — WHAT WE USED.** The demonstration models are trained on writing that is out of
   copyright in the UK, which means the author died more than seventy years ago. Name the writers and
   the editions, and link them. Say plainly that no living writer's work, and no customer's work, is in
   any public demonstration model.
3. **generic-text-block — WHAT WE TOOK OUT, AND WHY.** The honest section, and the one that earns the
   page. Old writing carries attitudes and episodes that a business would not put its name to, and
   "it is in the public domain" is not a reason to reproduce it. Say that we filter each corpus before
   training, that we say what was removed, and give the shape of it: for the Pepys diary, the passages
   describing coerced encounters, identified by the private code he wrote them in, roughly six per cent
   of entries. **Do not moralise about the writers; state what we did.**
4. **generic-text-block — YOUR OWN MODEL IS DIFFERENT.** A customer's model is trained on the customer's
   own documents, used for their model and nothing else, never mixed into a demonstration model, and the
   file is theirs to keep. The base model underneath is open-weight with its licence stated at handover,
   version-pinned. This section is where the £99 offer's promises and this page meet.
5. **faq + call-to-action** — can I use the model commercially (yes, and the base licence is named at
   handover); do you train on my documents for anyone else (no); what if I recognise my own writing
   (contact us and here is how we would handle it); who owns the output of a model I bought.

## Facts discipline

Every licence statement is already a registered fact on this site (`ft-licence-*`, version-pinned) and
the writer may state them only as recorded. **Dates of death, edition numbers and filter percentages
must be registered before they can be published** — same gate as the pricing page. The corpus research
supplies them; they are not facts here yet.

## What this page must not do

No legal advice. No claim that a model's outputs are free of any third party's rights, which is not
ours to assert. No suggestion that public domain means consequence-free, which is the thing the page
exists to disprove.
