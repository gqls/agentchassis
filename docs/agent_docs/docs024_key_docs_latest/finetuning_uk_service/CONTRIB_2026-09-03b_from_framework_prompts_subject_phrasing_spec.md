# CONTRIB 2026-09-03, from the framework-prompts lane: the subject phrasing spec you are holding for

This is the spec promised when the owner dropped the frame sentence. It is a copy of
`docs/agent_docs/docs024_key_docs_latest/framework_prompts_positive_voice/SPEC_2026-09-03_section_subject_phrasing.md`;
that file is the original if the two ever disagree.

**For this lane specifically.** Your three arrays (playground, your-own-model, technical-details) are valid data in the wrong register, so it is a re-authoring, not a rebuild of the plan. technical-details keeps its six topics under the rewritten brief, so that one is register-only as you said. Note rule 5: the subject now surfaces in the opening line by design, which is the thing to write into Stage B's assertions so it reads as intended behaviour rather than leakage.
---

Path: `docs/agent_docs/docs024_key_docs_latest/framework_prompts_positive_voice/SPEC_2026-09-03_section_subject_phrasing.md`.
Authority: the owner, 2026-09-03, deciding the 641 block: *"drop it and use fuller sentences/phrases"*.
Supersedes the rule in the 443 lane's CONTRIB that a subject must complete "You'll want to know ___".
That rule existed only to serve the frame sentence, and the frame is gone.

## What a subject is now

**The line the section opens on**, written to the reader, in the site's own voice.

It is no longer a label naming a topic. Under the previous block a subject was dropped into the middle
of a fixed sentence, so it had to be a short lower-case fragment. Under A4 it is printed verbatim on its
own line and the writer continues from it.

The owner's own example is the shape:

> If you'd like to prepare in advance of your hour, you might want to get these things ready.

## Where it is seen, which is the part that constrains it

Each subject appears **twice**:

1. As its own section's opening line, printed alone.
2. As one bullet in the list of the page's other sections, shown in **every other section's prompt**.

So a subject is read by the writer of every section on the page, not only its own. That is what the list
is for: it is the positive form of "do not restate what the others cover".

## The rules

1. **Say what the reader gets here, addressed to them.** Not "preparation" and not "what to bring", but
   the sentence you would actually open with.
2. **One sentence, or a phrase that reads as one.** It sits in a list of five or six in every sibling's
   prompt. A long one crowds them, and six long ones crowd the page's own brief.
3. **Give every section on a page a different one.** Two sections sharing a subject string makes both
   drop out of each other's list, because exclusion is by subject equality. Rule 17 already forbids it;
   this is what it costs if it happens anyway.
4. **No em dashes.** House rule, and it now prints into copy.
5. **Write only what you would publish.** The subject will be reproduced closely in the opening line.
   Under this decision that is the intent, not leakage, so a subject that overclaims ships an overclaim.
   Facts still come from the Verified Facts block; a subject is not a route around it.
6. **Sentence case and ordinary punctuation.** A full stop is fine now. A capital is fine now.

## What this does NOT change

The subject is still one authored string per section, carried on `site_plan_sections.subject` (planner)
or `pages.section_subjects[]` (hand-authored fallback). No new field, no schema change, no roll.
The literal `current_section.subject` is unchanged, so any gate keying on it still fits.

## For the build-site-planner nudge (rule 17)

The planner writes subjects unattended, and today writes capitalised noun phrases containing em dashes
(real, gamedesign.uk 2026-09-02). Those now print verbatim as a section's opening line. Replacement text
for the sentence migration 640 inserted, keeping 640's anchor `Any object entry may also carry a "subject"`:

> Any object entry may also carry a "subject": the line this section opens on, written to the reader in
> the site's own voice, as a sentence or a short phrase that reads as one. It says what the reader gets
> from this section rather than naming its topic. Keep it to one sentence, because it is shown to every
> other section on the page as well and a long one crowds their briefs. Give every section on a page a
> different one. A "subject" is REQUIRED on every entry whose component name appears more than once on
> the same page, because repeated components given the same brief write the same section.

Example entries, **deliberately generic**: an example in a prompt ships verbatim into live pages
(memory `a-quoted-exemplar-in-a-prompt-is-copied-verbatim`), and the risk is higher now that the slot
asks for prose rather than a label. Keep them abstract, and carry no numbers, since a copied figure
would be a fabricated claim:

    "sections": [{"name": "hero", "facts": []},
                 {"name": "features", "facts": ["F1-example-id"], "subject": "Here is what the service does day to day."},
                 {"name": "features", "facts": [], "subject": "Here is how a team starts using it."},
                 {"name": "call-to-action", "facts": []}]

Before swapping them, grep for other readers of 640's applied example strings ("What the platform does",
"How a team adopts it"); the apis.uk lane holds that due-diligence step.
