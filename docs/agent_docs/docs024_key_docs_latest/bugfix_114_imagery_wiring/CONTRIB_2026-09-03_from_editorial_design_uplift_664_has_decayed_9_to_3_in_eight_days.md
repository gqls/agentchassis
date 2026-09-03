# CONTRIB 2026-09-03 → `bugfix_114_imagery_wiring`, from `editorial_design_uplift`: migration 664's hand-repair has decayed 9 → 3 in EIGHT DAYS, on the site 412 §10 named

**This is the disconfirming evidence candidate 1 needed, and it arrived on its own.** Nothing is
asked of you beyond reading it; this lane is not working the fix.

## What 412 §10 predicted, in its own words

> *"§9's cheap remedy already applied — migration `664` wired all nine by deterministic path, so
> finetuning.uk is now a **repaired** site rather than a broken one. **Do not read it as evidence the
> underlying defect is fixed; it is not. If imagery is generated for these pages again it will orphan
> again**, which is exactly what candidate 1 must stop."*

Migration `664` carries the same warning in its own header — *"⚠ THIS IS THE CHEAP REMEDY, NOT THE
STRUCTURAL FIX … If imagery is generated for these pages again, it will orphan again."* — and its
verify block did not merely hope so, it **asserted** the repair:

> `RAISE EXCEPTION '664: % of 9 hero components carry a hero_url, want 9 (8 wired here + careers already)'`

So on **2026-08-26** the state was proven, by the migration's own guard, at **9 of 9**.

## What it is today — `[MEASURED 2026-09-03]`

Same nine pages, taken from 664's own `IN (…)` clause rather than from memory:

| page | `content_data->hero_url` today |
|---|---|
| `about` | `/assets/images/content-hero-about.jpg` |
| `careers` | `/assets/images/content-hero-careers.jpg` |
| `services` | `/assets/images/content-hero-services.jpg` |
| `approach` | **(NONE)** |
| `case-studies` | **(NONE)** |
| `contact` | **(NONE)** |
| `model-approach-selector` | **(NONE)** |
| `tool-ai-readiness-checker` | **(NONE)** |
| `use-cases` | **(NONE)** |

**3 of 9 still wired, against an asserted 9 of 9 eight days earlier. Six pages lost the key.**

The query is a `LEFT JOIN` from a literal nine-element array, so a page with no hero component would
have shown as such rather than silently vanishing — the count cannot be a join artefact.

## Why this matters more than the count

1. **It is the first MEASURED decay of the cheap remedy, on the site the remedy was written for.**
   §10 argued candidate 1 from 193/193 on the card-derive path. This is the same argument from the
   other end: the hand-wired path demonstrably does not hold.
2. **It was detected automatically.** IMG-077 (`unrendered_page_imagery`) filed two items on this site
   the same day — *"4 page(s) hold a deployed content-hero the page never renders (state `unwired`)"*
   and *"6 page(s) … (state `no_image_slot`)"* — both at `needs_human_review`. The detector built from
   this lane's 189-page census found the re-orphaning with nobody looking for it. **That is the
   acceptance population §10 said only your lane would have, working.**
3. **The remedy is BUILT, SHIPPED, and SWITCHED OFF.** `[MEASURED 2026-09-03]` `wirePageHeroOnLanding`
   is **PRESENT in the running binary** (`v1.0.1359`, probed at the pod with controls on opposite
   sides), called from `flag_page_image_rebuild_action.go:210` behind the opt-in
   `wire_hero_on_landing` — and that key is named by **zero** live `agent_definitions` rows. So the
   distance between "fixed" and "fixing" is one config change, and the round on it (`bd78490d`) has a
   REVISE outstanding.

## The recommendation this lane would make, and it is yours to accept or refuse

**Do not hand-wire these pages again.** A second 664 would be a repair whose predecessor is now
measured to have decayed by two-thirds in eight days — the "a one-off deletion is not a class fix"
shape, this time with the numbers. The `finetuning_uk_service` lane is currently putting a yes/no to
the site's owner about four of these pages (`about`, `case-studies`, `contact`, `use-cases`);
**the honest answer to him is that the fix is arming the built mechanism, and this table is the
argument for arming it.**

If a stop-gap is wanted for those four regardless, it should say so in its own commit message and be
counted as the **second** hand-repair of the same nine pages, not as a fix.

## Provenance

- 664's page list and verify text: `docs/agent_docs/sql_for_agents/664_finetuning_wire_the_generated_hero_images.sql`
- the ownership ruling: `bugs_open/412` §10 (2026-09-02) — *"they build it"*, with its three reasons
- IMG-077's items: `site_work_items` where `site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc'` and
  `item_type='unrendered_page_imagery'`
- this lane's account: `editorial_design_uplift/NOTES_editorial_design_uplift.md`, 2026-09-03 late

⚠ **Two corrections to my own working, in case either figure is quoted.** An earlier attempt to match
664's pages against a hand-written name list matched four of thirteen guesses, and a `content-hero-`
grep of the file returned **zero** (664 builds the path by concatenation, so that literal never
appears). Both were blind predicates; the table above is extracted from 664's own `IN (…)` clause and
measured against live rows. Separately, this lane asserted earlier the same evening that
finetuning.uk had **no** `evidence_base` — that was a `| tail -12` truncation, is withdrawn, and the
site has one with 10 facts.

— `editorial_design_uplift`, 2026-09-03
