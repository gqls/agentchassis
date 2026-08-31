# CONTRIB 2026-08-31 — from the boxingonline.com build: the page DESCRIBES the editorial policy instead of doing it, and the words came from our own research spec

**To:** copy_quality_two_stage (you own the writer prompt, the house voice and the
demonstrations-govern finding).
**From:** the session reviewing boxingonline.com — site `d2aa5206-73bc-4707-a69c-2702c1eb9152`,
order BR-9AUZ59, **the first paid customer build**, planned and built 2026-08-31.
**Owner raised it in chat the same evening.** All figures `[MEASURED 2026-08-31]` against the
live DB and the served site at `https://boxingonline.ugg2.com`. Root cause NOT asserted — see §5.

---

## 1. The complaint, in the owner's words

> "the copy still suffers from telling the user what the site is doing rather than talking to
> the user."

He quoted three passages. Here is the first, served today at `/about.html`:

> **How we cover it**
> We write the way a knowledgeable fan talks: direct, opinionated where an opinion is called
> for, and specific. **A preview that says a fight 'could be great' tells the reader nothing,
> so we'd rather name the styles, the records and the stakes** and let the reader decide for
> themselves. Where we're giving our own view of a matchup or a ranking, we say so plainly…
>
> - **Accuracy on the details:** dates, venues, records and sanctioning body decisions get checked…
> - **Opinion kept separate from fact:** when we rate a fight or call a decision, that's clearly our view of it.

## 2. Where those sentences actually came from — this is the finding

They are a paraphrase of **our own `vertical_landscape` spec**, which is the researcher's
instruction sheet to the build, not material for a page. Side by side:

| served copy (`about.html`) | `site_specs.aspect='vertical_landscape'`, `lessons.avoid[]` |
|---|---|
| "A preview that says a fight 'could be great' tells the reader nothing, so we'd rather name the styles, the records and the stakes" | "Vague fight previews that say 'this could be a great fight' — every preview must contain specific analysis of styles, records, and what's at stake or it adds no value" |
| "Opinion kept separate from fact: when we rate a fight or call a decision, that's clearly our view of it." | "Letting opinion drift into fact — Ring and BoxingScene both label analysis as analysis; Boxing Online must be equally disciplined" |
| "Accuracy on the details: dates, venues, records… get checked, because a wrong date… does the reader more harm than no information at all." | "Stale calendar entries — a wrong fight date actively harms readers and destroys trust faster than any other error" |
| "The whole sport: that includes women's boxing, amateur boxing and the international scene." | "Over-reliance on a single weight class or fighter tier… leaves the majority of the boxing audience underserved" |

Reproduce:
```sql
SELECT jsonb_pretty(data->'lessons') FROM site_specs
 WHERE site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152' AND aspect='vertical_landscape' AND is_current;
SELECT pc.content_data->>'content' FROM page_components pc JOIN pages p ON p.id=pc.page_id
 WHERE p.site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152' AND p.name='about';
```

**This is your headline mechanism wearing a new hat.** You proved that *demonstrations govern
and instructions don't* — the writer copies the shape of what it is shown. Here it was shown a
list of editorial rules written in the second person about itself, and it **rendered the rules
as prose**. The seed for `vertical-exemplar-researcher` even says *"reasons, not copies"* —
that instruction binds the researcher and nothing downstream of it.

It is the same family as the two entries already in the index —
`prompt-text-poisons-its-own-detector` and
`a-quoted-exemplar-in-a-prompt-is-copied-verbatim` — but a third variant worth its own line:
**a rule ABOUT the writing, placed in the writer's context, is emitted AS the writing.**

## 3. It is not confined to the about page — the pattern scales with emptiness

`/articles/index.html` (3,114 chars of body copy) contains **zero articles** and four headed
sections of editorial policy: *"What's in the mix"*, *"Every weight, every corner of the
world"*, *"Keeping it accurate"*, *"Where to go next"*. The owner quoted:

> **Keeping it accurate** — Names, records and dates matter more here than style points. When a
> fighter's record changes or a fight gets moved, we'd rather fix the piece than leave it wrong.

The reason there are no articles is `bugs_open/419` (the planner emitted a zero-section
`blog-post` page, so six brief-promised article slots never built). **But note what the writer
did with the void: it wrote a manifesto.** An index page with nothing to index does not fail —
it produces policy prose in the house voice and passes every check (`valid=true, issues=0` on
all pages). That is worth a demonstration in its own right: *when there is nothing to list, say
so in one line, do not describe the section.*

## 4. The best-in-class question — your PLAN_2026-08-25 is still 0-for-N, now including a paying customer

The owner also asked, in the same message, whether the mission and specs knew the site was
supposed to be best in its vertical. Measured on this site:

- **The research knew, in detail and well.** `vertical_landscape` (confidence 0.88) analysed
  ringtv.com, boxingscene.com and skysports.com/boxing and produced a genuinely good
  `differentiation_opportunity`: own the *"what to watch this weekend"* moment — a weekly
  curated viewing guide across every broadcaster, with fight-time conversions and honest
  one-line previews. `strategy` carried it forward faithfully: *"A weekly 'what to watch'
  curated guide article, published every Thursday or Friday, becomes the signature piece — the
  thing readers bookmark the site for"*, plus fighter entity pages and the magazine grid.
- **The plan did not.** The site plan is 6 pages: index, about, contact, articles-index,
  fight-calendar, and one zero-section `article`. No weekly guide. No fighter pages. No
  magazine grid.
- **The phrase never arrives.** `[MEASURED]` **0 of 10** current specs on this site contain
  "best in class"/"best-in-class", and `strategy` has **no `benchmark` key**:
  ```sql
  SELECT aspect, (data::text ILIKE '%best%in class%') FROM site_specs
   WHERE site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152' AND is_current;
  SELECT jsonb_object_keys(data) FROM site_specs WHERE aspect='strategy' AND is_current
    AND site_id='d2aa5206-73bc-4707-a69c-2702c1eb9152';
  ```
  That is your PLAN_2026-08-25 §1 measurement ("0 of 51 sites") holding on **site 54**, six
  days later, on the first one a customer paid for.

**So the answer to the owner is: the research reached best-in-class; the build did not.** Your
propagation plan (§3's carrier-row + `{{.build_standard}}` design) is the right fix and none of
it has shipped. This build is the strongest argument for it yet, and it is worth telling him
that the plan predates the failure rather than being a reaction to it.

**A second, cheaper gap sits alongside it.** `vertical_landscape` is read by exactly **two**
agents fleet-wide:
```
SELECT type FROM agent_definitions WHERE is_active AND COALESCE(is_snapshot,false)=false
  AND deleted_at IS NULL AND default_config::text ILIKE '%vertical_landscape%';
  -->  domain-strategist, vertical-exemplar-researcher
```
In Go it appears only as an **existence check** (`check_build_prerequisites.go:121-124` asserts
the row is present). Nothing reads its contents into the planner, the writer, the designers or
the tool suggester. We research what best-in-class looks like, confirm the row exists, and then
build without it.

## 5. What I am NOT claiming

I have not run `090` on the leak in §2, so **do not quote this file as a mechanism finding.**
What is measured is: the served sentences, the spec text they track, the consumer census, and
the absent phrase. Whether the writer saw `vertical_landscape` directly, saw it via `strategy`,
or converged on the same sentences from the brief is **[UNVERIFIED]** — and it matters, because
the three have different fixes. Your replay harness is better suited to settling it than a
diagnosis run: the pre-registered experiment is *rebuild `about` with the `lessons` block
withheld and see whether the policy prose survives.*

## 6. What I would ask of your lane

1. **A demonstration for the empty-index case** — "nothing to list yet" is one honest line, not
   four headed sections about our standards.
2. **A tell for meta-copy**, in the CQ-032 scanner family: first-person-plural editorial-policy
   sentences on a page whose job is to serve content ("we write…", "we'd rather…", "we cover…",
   "gets checked"). This is mechanically detectable and it is now the owner's stated top
   complaint on a paid build.
3. **Ruling wanted:** should `lessons.adopt/avoid` reach the writer's context at all? It reads
   as guidance and renders as copy. If it should, it needs to arrive as constraints on a brief,
   never as prose in the prompt body.
4. Your best-in-class propagation plan is the answer to the owner's question in §4 — it needs a
   go, and this build is the evidence for it.

Related: `bugs_open/419` (no articles to index), and the experience-loop CONTRIB filed the same
day (the home page's editorial slot lists four tool GUIDES as "Latest from the ring").
