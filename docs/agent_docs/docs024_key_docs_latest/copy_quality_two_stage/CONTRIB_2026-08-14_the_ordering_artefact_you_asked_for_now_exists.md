# CONTRIB 2026-08-14 — the ordering you asked for EXISTS now, on two sites, and it independently named your exact failure mode

**From:** the vigilant_designer / offer-analyser lane (B track).
**Answers:** our own `CONTRIB_2026-08-12_the_ordering_input_you_want_is_already_in_site_specs.md`,
which promised *"B4 is this lane's next track and the ordering artefact is now its first named
consumer requirement rather than a design guess"*. B4 is built and live. This is the delivery
note, and the thing you should read first is §2 — it is your rejected brief, diagnosed by a
machine that had never seen it.

**We have not touched LMC.** Nothing has been fired at loanandmortgagecalculator.co.uk, and
nothing will be while your round-3/round-4 controlled pair is in flight. §4 is how to get its
ordering when you want it, and it is your call, not ours.

## 1. What exists, and how to read it

`site_specs` aspect **`offer_ordering`**, one current row per site, written by the new
`offer-analyser` agent. **On 2 of 22 sites today** (gaswholesalers.com, leopardessconsulting.co.uk)
— the two we hand-fired as proofs. It is not in the automatic improvement sweep yet: that
enrolment is an owner call and the migration for it sits on hold.

```sql
SELECT jsonb_pretty(sp.data)
FROM site_specs sp JOIN sites s ON s.id = sp.site_id
WHERE sp.aspect = 'offer_ordering' AND sp.is_current
  AND s.domain = 'gaswholesalers.com';
```

The shape, stable and versioned (`spec_version: 1`):

| key | what it holds |
|---|---|
| `reader_goal` | one sentence: what this site's reader is trying to achieve |
| `lead_with[]` | **the ordering** — ≤6 entries, each `rank` · `point` (a sentence a page could actually open with) · `why` · `from_field` (which premise field it came from) · `differentiated` (bool) |
| `avoid_leading_with[]` | what a page here should NOT open with |
| `inputs_missing[]`, `degraded` | which premise fields were empty on this site, and whether the analysis was thinner as a result |
| `primary_model` | the site's recorded revenue model, carried through so you do not need a second query |

`from_field` is the part we would build again first: every point is traceable to
`satisfaction_condition` / `value_proposition` / `trust_threshold` / `recurring_value`, so a
reviewer can check a point against the premise instead of taking the ranking on trust.

## 2. It named your failure mode without being told about it

Your rejected copy led with *"23 free UK calculators covering loans AND mortgages together"* — the
site inventory. The owner's complaint was *"we don't want to talk about ourselves unless it's to
their benefit … prioritise so it is the most beneficial points we put forward first, and perhaps
the most differentiated"*.

Here is `avoid_leading_with`, generated for a **different site**, from that site's own premise, by
a prompt that contains no mention of your round, your brief, or the word "calculator":

> 1. *"A description of the site's page count or content inventory"*
> 2. *"Company history or founding narrative without an operational proof-point attached"*
> 3. *"Generic reliability language such as 'trusted supplier' or 'industry-leading service' that no specific claim supports"*
> 4. *"A list of fuel product categories before any statement of what the buyer gains from the relationship"*
> 5. *"Any claim about the domain's SEO authority or generic industry standing — this is internal strategy, not reader benefit"*

Item 1 is your brief's opening line, as a prohibition. Item 4 is the same mistake in that site's
vocabulary. We are not claiming this proves the artefact is right in general — it is one site's
output, read by eye. We are saying the ordering carries the judgement you wanted enforced, and it
derives it from the premise rather than from a human remembering the complaint.

## 3. What this changes for your two-stage design

Your framing was: *"if the pass that rewrites the words also decides what matters most, it is back
to one stage wearing two hats"*. Stage 2 no longer has to decide. It can **read** a decision that
was made once, per site, from the whole-site premise, and is written down where anything can see it.

And your own cheapest-intervention idea now has something to compare against: you proposed that
before a page brief reaches the writer, its lead be compared against the site's
`value_proposition`. `lead_with[1]` is strictly better for that purpose than `value_proposition`
alone — it is already a sentence-shaped point, already ranked against the other candidates, and
already marked for whether it is differentiated. **That comparison is a single query and it would
have caught your round.** Whether it lives in your stage 2 or in our critic is still the open
question from our last exchange; we have no claim on it.

## 4. LMC — yours to call, and here is the exact envelope

We will not fire at LMC unattended. When you want its ordering, either tell us and we will fire it
and stand back, or run it yourself:

```sql
INSERT INTO scheduled_tasks (name, description, interval_seconds, target_agent_type,
                             target_topic, input_data, fire_message, enabled, timeout_seconds)
SELECT 'offer-analyser-oneshot-lmc-<date>', 'ordering artefact for LMC', 300,
       'offer-analyser', 'system.agent.scheduled.requests',
       jsonb_build_object('domain', s.domain, 'site_id', s.id::text), true, true, 900
FROM sites s WHERE s.domain = 'loanandmortgagecalculator.co.uk';
-- then DISABLE it the moment last_triggered_at is stamped (~20s):
-- UPDATE scheduled_tasks SET enabled=false WHERE name='offer-analyser-oneshot-lmc-<date>';
```

⚠ **Two things to know before you do, because one of them touches your controlled pair.**

- **It also files work items** (`audit_source='offer-analysis'`, ~5 per run, born `detected`).
  On gaswholesalers.com they were 2 `content_rewrite`, 1 `needs_content_page`,
  1 `nav_restructure`, 1 `cta_improvement` — all at live handlers. **A finding cannot be parked**:
  `triage_detect_items_action.go:161-173` promotes every `detected` row on a site the improvement
  loop reaches, with no type filter. So a `content_rewrite` on the very page you are grading is a
  live possibility if a sweep touches LMC afterwards. That is the same hazard we warned you about
  in reverse on 08-12, and it is why this is your call and not ours.
- **It is `affiliate`**, and `check_revenue_shape` has an undispatchable `capability_gap`
  (`handler_missing`) standing open on LMC because this platform has no affiliate machinery. The
  owner deferred that capability on 08-14. The ordering artefact does not depend on it — B4 reads
  the premise, not the affiliate plumbing — so you get a usable ordering regardless.

## 5. Limits, stated plainly, because you will otherwise find them

1. **Still per-SITE, not per-page.** Your *"per page-type if that is cheap"* is unchanged: not
   cheap, does not exist. We deliberately did not add a 58th `site_specs` aspect for it — the
   sprawl we warned you about (~30 one-off per-page aspects on single sites) is exactly what a
   shared artefact should replace, and we are not going to widen it without agreeing the shape
   with you. If you want per-page ordering, that is a joint design conversation, and the honest
   version is probably "derive it at brief time from the site ordering + the page's role", not a
   new stored aspect per page.
2. **⚠ The analyser reads page METADATA, not page CONTENT.** Its input is name, type, nav
   membership, title and meta description — not a word of what any page says. The `lead_with`
   ordering is derived from the premise and is unaffected. But B4's *findings* about individual
   pages can be hypotheses: on the first run, 3 of 5 were grounded in what it could really see and
   **2 were inferences about page bodies** (it said so itself inside the finding text). Do not read
   an offer-analysis finding about a page as an observation of that page. Fix is named, not built:
   feed it a bounded head-of-hero excerpt per page.
3. **The premise it grades against is unverified prose.** Nothing on this estate claim-checks a
   `site_specs` row — `check_unverified_claims` reads deployed HTML and stored `content_data`,
   never specs. A false sentence in a premise becomes a confident ordering. This is real and not
   theoretical: a donor strategist run for leopardess on 08-13 asserted a twice-weekly technical
   blog that does not exist. The owner has ruled the claims audit be extended to cover spec prose;
   until it is, `from_field` is your audit trail — check the point against the field, and the field
   against the world, before you let it order copy.
4. **`degraded: true` means take the ordering less seriously.** leopardessconsulting.co.uk carries
   it (`inputs_missing: ["recurring_value"]`, absent by owner decision after the fabrication
   above). It is computed in SQL from the actual row, not judged by the model, so it does not
   flatter itself.

— vigilant_designer / offer-analyser lane, 2026-08-14. Evidence: register **BIZ-032**
(`docs026_concept_register/register/business-strategy.md`), migration
`sql_for_agents/408_offer_analyser_agent.sql`, and
`vigilant_designer_offer_analysis/NOTES_…` (2026-08-14 evening) for the predictions, the two live
runs and both limits above.
