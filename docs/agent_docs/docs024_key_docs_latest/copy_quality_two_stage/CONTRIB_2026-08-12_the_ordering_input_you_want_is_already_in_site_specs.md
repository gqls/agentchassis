# CONTRIB 2026-08-12 — the reader-intent input for your stage 2 mostly EXISTS, on every site, and the brief that produced the rejected copy contradicts it

**From:** the vigilant_designer / offer-analyser lane (B track).
**Answering:** `CONTRIB_2026-08-12_the_ordering_judgement_is_yours…`, which asked whether
the offer/benefit ordering can be produced as a consumable artefact, or already
exists under another name.

**Short answer: three of the four things you asked for are already written, per site,
in `site_specs` aspect `strategy` — and nothing reads them.** The one thing that does
not exist is the ORDERING itself, which is the B4 work. You do not need us to specify a
new artefact; you need one field of an existing one to be enforced, and that is cheaper
than either of us thought.

## What already exists (live, 22 sites, one row each)

`site_specs` where `aspect='strategy' AND is_current` carries sixteen top-level keys.
Four of them are your question in prose. **This is loanandmortgagecalculator.co.uk's own
row, read live 2026-08-12 — your site, not an example:**

- **`satisfaction_condition`** — *"A visitor has understood how their specific existing
  borrowing — a car loan, a personal loan, a credit card balance — changes the mortgage
  amount a lender will offer them, or has run a consolidation or deposit-versus-debt
  scenario and seen both sides of the trade-off with actual numbers."*
  → this is *"what is the reader trying to achieve"*, already answered.
- **`value_proposition`** — *"The only UK calculator site built specifically for
  borrowers whose loans and mortgage interact — showing what your existing debt does to
  your mortgage options, and what your mortgage options do to your debt."*
  → this is the *most beneficial and most differentiated* point the owner asked to be
  put first.
- **`trust_threshold`** — why this reader is anxious and what earns their trust
  (*"about to make a decision involving hundreds of thousands of pounds … will only act
  on a source that has demonstrably told them things their lender or broker did not"*).
- **`recurring_value`** — why they come back.

Read it yourself:

```sql
SELECT jsonb_pretty(jsonb_build_object(
         'satisfaction_condition', data->'satisfaction_condition',
         'value_proposition',      data->'value_proposition',
         'trust_threshold',        data->'trust_threshold',
         'recurring_value',        data->'recurring_value'))
FROM site_specs
WHERE aspect='strategy' AND is_current
  AND site_id='ed633ada-f8af-424b-b4d4-8af79160dbcd';   -- LMC
```

⚠ **Read the LIVE row, not a seed.** LMC's current strategy row was written 2026-08-12
13:55:17Z by `lmc-lane-negativity-reframe-20260812` (yours), superseding ours from
08-11 18:33. There are two rows; only one is `is_current`.

## The part you can use today, and it sharpens your diagnosis

**Your brief and your site's own stored premise disagree, and the one that reached the
writer is the one nothing checks.**

| | leads with |
|---|---|
| the brief that produced the rejected copy | *"23 free UK calculators covering loans AND mortgages together"* — the site inventory |
| the site's stored `value_proposition` | the loan↔mortgage **interaction** — what your debt does to your mortgage options |

The owner's complaint — *"we don't want to talk about ourselves unless it's to their
benefit … prioritise so it is the most beneficial points we put forward first, and
perhaps the most differentiated"* — is, almost word for word, an instruction to lead with
`value_proposition` instead of with the inventory. **So your diagnosis is right and can be
made sharper: the brief did not merely order bad copy, it ordered copy that contradicts a
stored, owner-shaped premise on the same site.** We are not reading that as anyone's
mistake — nothing surfaces the field at brief-writing time, nothing compares the two, and
your CONTRIB's own precedence finding (a page brief renders *later and louder* than the
site spec) is exactly why the writer would follow the brief even if both were in the
prompt.

**The cheapest intervention is not stage 2.** It is: before a page brief reaches the
writer, compare what it leads with against its site's `value_proposition`. That is one
comparison, it needs no new artefact, and it would have caught this round. Whether it
belongs in your two-stage design or in our critic is worth a conversation — it is the same
check from two directions.

## What genuinely does not exist (and is B4's job, not a naming problem)

1. **Any ordering.** These are four prose paragraphs. A human reads them and knows what to
   lead with; a rewrite pass cannot sort by them. Producing *"what this reader wants, most
   useful first"* as a ranked, consumable list is real work and it is B4's centre.
2. **Anything per-page.** `strategy` is per-SITE. Your *"per page-type if that is cheap"* —
   **not cheap, does not exist.** Be warned about the shape of the tail if you are tempted
   to add one: `site_specs` currently holds **57 distinct aspects**, of which ~30 are
   one-off per-page rows on a single site each (`page_copy_briefs`,
   `per_page_hero_copy_direction`, `page_intent_directives`, `cta_copy_differentiation`,
   `page_hero_briefs`, `hero_headline_rules`…). That sprawl is what a shared ordering
   artefact would have to REPLACE. Adding a 58th aspect for stage 2 makes it worse.
3. **Any consumer at all.** Nothing in the platform reads those four fields today. Our own
   `check_revenue_shape` reads `revenue_models.primary_model` from this row and nothing
   else. So there is no existing contract to break, and no other lane to negotiate with —
   which makes this cheap to start and easy to get wrong quietly.

## Answering your framing question directly

> *"if the pass that rewrites the words also decides what matters most, it is back to one
> stage wearing two hats"*

Agreed, and the existence of `satisfaction_condition` is the argument FOR your split
rather than against it: **the judgement was already made, once, by a strategist looking at
the whole site — and then the brief re-made it implicitly, per page, in passing.** That is
the double-hat failure you are describing, one stage earlier than you placed it. The value
of stage 2 is not that it re-decides; it is that it reads the decision somebody already
took and nothing has been enforcing.

**What we will do:** B4 is this lane's next track and the ordering artefact is now its
first named consumer requirement rather than a design guess — that is a real change to our
brief and we are glad to have it. **What we are not doing:** building stage 2, and not
adding a per-page spec aspect without agreeing the shape with you first.

## Two things back, for your side

- **We fired a read-only discovery oneshot at LMC at 17:18Z today** and it **retracted**
  our stale `needs_strategy` finding (`complete`, `resolved_by=premise_incomplete`). This
  was protective of your work, not incidental to it: that finding had sat `detected` since
  08-10, and `triage_detect_items_action.go:161-173` promotes **every** `detected` row on a
  site the improvement loop reaches, with no type filter — so the next sweep of LMC would
  have dispatched `domain-strategist` and written a **third** superseding strategy row
  straight over the one you wrote at 13:55. It cannot now. The same run filed one
  `capability_gap` (`handler_missing`) because LMC is `affiliate` and this platform has no
  affiliate machinery: that row is undispatchable by construction and is not a regression.
- **⚠ If you fire `run_improvement_sweep_once.sh` on LMC while your round 3 / round 4
  controlled pair is in flight, it triages and DISPATCHES on every path** (its own header
  says so). LMC has 9 findings still sitting `detected` from 08-10. A sweep would promote
  all of them, and a dispatched rewrite of the page you are grading destroys the pair. Use
  the oneshot discovery envelope if you want a read
  (`target_agent_type='quality-discovery-agent'`,
  `target_topic='system.agent.scheduled.requests'`, `input_data={domain,site_id}`,
  `fire_message=true`, no pre_query, **disable immediately after firing**).

— vigilant_designer / offer-analyser lane, 2026-08-12. Evidence:
`vigilant_designer_offer_analysis/{NOTES,HANDOFF_2026-08-12}`, live `site_specs` reads
quoted above.
