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

---

## Addendum, 2026-09-03 evening, from the first live build under A4 (finetuning.uk/technical-details)

Two things the first real run showed, both of which bind anyone writing subjects.

**7. A short subject invites a short section, and the shrink floor reads short as truncated.** On
`your-own-model` the writer followed a 46-character hero subject and produced a 212-character hero
against an existing 429-character one, so `save_page_sections` refused the WHOLE PAGE at 49% kept
against the 50% floor (`bugs_open/178`). The floor exists to catch a truncated rewrite and cannot
tell that apart from a deliberately tighter opening. **So when re-authoring subjects for a page that
already has long copy, either keep the subject substantial enough to carry a comparable section, or
expect the floor to refuse and plan for it.** This is a real interaction between A4 and the floor,
not a fault in either.

**8. The subject is NOT the loudest instruction in the prompt, and on a page with several same-typed
sections it currently loses.** Measured on the first build: the A4 block rendered correctly in all six
prompts, each carrying its own subject and its five siblings, and three of the generic-text-blocks
still opened on the same topic. The reason is upstream of anything in this spec: the page's whole brief
is rendered into EVERY section's prompt under "## Rewrite Guidance (IMPORTANT: incorporate this into
the content)", byte-identical across the six (md5 `b4fd73f0…`, 3,295 chars each), and it is 2.6 KB
later in the prompt than the subject block. Each section is therefore told, in the prompt's most
emphatic register, to incorporate material the same brief assigns to its siblings. Filed as a diagnosis
run rather than patched here, because the cause is not in the block and a second edit to the block
would paper over it. **Until that is resolved, a well-written subject is necessary and not sufficient**,
and a page whose sections converge is not evidence that its subjects were written badly.
