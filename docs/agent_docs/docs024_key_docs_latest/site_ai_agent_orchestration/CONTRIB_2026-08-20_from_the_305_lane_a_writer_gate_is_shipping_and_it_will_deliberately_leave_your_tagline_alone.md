# CONTRIB 2026-08-20 — from the `bugfix_305_negation_gate` lane: a writer-side gate is shipping, and it will deliberately leave YOUR tagline alone

**Who this is from.** A session building the platform fix for `bugs_open/305` — the owner's complaint
about `model-directory`, `adoption-tracker` and `protocol-tracker` reading like they *"didn't go
through the framework"*. Your site is the one that drew the complaint. **I have not touched your
site, your specs, or your pages**, and I am not asking to.

## What is shipping, in one paragraph

A mechanical check between `page-content-writer`'s section generation and the render, plus a one-shot
sentence-level rewrite: the define-by-negation family ("X, not Y", "not X but Y", "Not a demo. Not a
proof of concept.", "rather than", and the two-sentence negative reveal) is counted per section
against a per-PAGE budget of two, and any hit in a headline-class field is rewritten. Detail:
`docs/agent_docs/docs024_key_docs_latest/bugfix_305_negation_gate/PLAN_2026-08-20_negation_gate.md`.

## The part that concerns you: it will NOT fix your three pages

The gate **exempts any sentence the brief supplied**, because a phrase the brief hands the writer is
the brief's decision, not the writer's mistake — and rewriting it would put the platform in the
position of overruling a site's own voice specification, which the house voice explicitly forbids
(*"A site's own voice specification outranks these rules"*).

Your `content_direction` supplies this, verbatim `[MEASURED 2026-08-19]`:

> *"Multi-agent systems deployed to production **in days, not months** — on Kubernetes, Kafka, and
> Postgres"*

and `content_direction.emphasis` **orders** it into *"the homepage hero, services page hero, site
footer, and meta descriptions"*. That string reaches the writer in **1,369 rendered prompts** and
comes back in **409 responses** — it is the single highest-leverage string on the site, and it is the
owner's own quoted hero sentence. `adoption-tracker`'s hero carries it live today.

**So: the gate will count it and leave it.** The only thing that changes those pages is editing the
brief and rerendering, and that is yours.

## Two traps on that edit, both already recorded by other lanes

1. **`bugs_open/327`** — `formatted` is computed from the INCOMING partial before the merge, so a
   careful narrow correction to one key **deletes the rest of the brief from the writer's view**.
   Your site is one of the three already serving a fragment (**5 of 18 keys since April**). Write the
   whole `content_direction` object, not a patch, until 327 is fixed.
2. **Do not verify by diffing the brief.** `formatted` is regenerated in random key order on every
   write (Go map iteration), so a text diff reports ~100% changed whether or not anything did. Verify
   by label presence and phrase position (`LANDMINES.md`, the two `formatted` entries).

## Verify at the artefact, not at a status — the queries

Note `\y` for word boundaries: Postgres has no `\b` (there it is a *backspace* character and matches
nothing — `LANDMINES.md:4219`, and it cost me a wrong census in this very session).

```sql
-- 1. Does the brief still supply it?
SELECT position('in days, not months' in sp.data->>'formatted') > 0 AS supplied,
       length(sp.data->>'formatted') AS visible_chars
  FROM site_specs sp JOIN sites s ON s.id = sp.site_id
 WHERE s.domain = 'ai-agent-orchestration.com'
   AND sp.aspect = 'content_direction' AND sp.is_current;

-- 2. Which of the three pages still SERVE the construction (the artefact, not the log)?
SELECT p.url, pc.slot_name,
       pc.content_data::text ~* '\w,\s+(not|never)\s+\w'  AS x_not_y,
       pc.content_data::text ~* '\yrather than\y'          AS rather_than
  FROM page_components pc
  JOIN pages p ON p.id = pc.page_id
  JOIN sites s ON s.id = p.site_id
 WHERE s.domain = 'ai-agent-orchestration.com'
   AND p.url ~ '(model-directory|adoption-tracker|protocol-tracker)'
 ORDER BY 1, 2;
```

As of 2026-08-19 that second query returns `x_not_y = true` on six of nine components, including all
three heroes and `model-directory`'s call-to-action. **Do not read `page_components.updated_at` as
"regenerated"** — a rerender bumps it without rewriting anything, which is what made the copy look
five days newer than it was (`bugs_open/305 §3`).

## One thing I noticed and did not act on

`model-directory`'s `cta_url` and `primary_cta_url` both point at `/tools/password-entropy.html`
while the button reads *"Book a Technical Discovery Call"*, and the site carries 27 open
`cta_names_unknown_destination` items. Already flagged in `bugs_open/305 §6`; still true. Yours.
