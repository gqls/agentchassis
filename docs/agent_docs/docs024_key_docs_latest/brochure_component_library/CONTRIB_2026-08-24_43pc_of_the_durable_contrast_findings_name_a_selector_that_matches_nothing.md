# CONTRIB 2026-08-24 — 73 of your 171 durable contrast findings name a selector that matches NOTHING

From the `bugs_open/352` lane (`docs/agent_docs/docs024_key_docs_latest/bugfix_352_invented_selector/`).
Left here rather than sent, because there was no live session on this lane when I measured.

**This bears directly on the open decision in
`DECISION_INPUT_2026-08-18_the_186_durable_contrast_failures.md` §"do we turn the fixer loose on
these 185, and if so on which ones?"** I think it changes the answer to option 1, and it adds a
class your class A/B split does not have a name for. I am not asking you to do anything in my
lane; I am handing over one number and one warning.

## The number [MEASURED 2026-08-24, live `clients_db`]

Your population, re-counted today (it has drifted down as you predicted — 185 → **171**):

```sql
SELECT count(*) AS open_still_failing,
       count(*) FILTER (WHERE spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$') AS invented,
       count(DISTINCT site_id) AS sites,
       count(DISTINCT site_id) FILTER (WHERE spec->>'selector' ~ '^([A-Za-z0-9]+)\.\1$') AS invented_sites
FROM site_work_items
WHERE item_type = 'contrast_failure'
  AND status NOT IN ('complete','verified','rejected','wont_fix','cancelled');
```

| | |
|---|---|
| open, still failing | **171** |
| of which the selector **cannot match any element** | **73 (43%)** |
| sites affected | **13** of your 15 |

Breakdown: `P.P` 30, `H2.H2` 11, `A.A` 8, `H3.H3` 7, `LEGEND.LEGEND` 5, `H1.H1` 4,
`STRONG.STRONG` 2, then one each of `H4/LABEL/CODE/BUTTON/SPAN/EM`.

## Why those selectors are impossible, in one paragraph

The render audit's in-page sweep records an element's class in a field called `Class`, and for an
element with **no class** it substitutes the element's **tag name** instead
(`internal/adapters/browserrunner/render_audit_action.go:202`, the `|| el.tagName` fallback).
Downstream that value is composed into `TAG.Class`, so a class-less `<p>` becomes `p.P` — as CSS,
"a `<p>` carrying `class="P"`", of which there are none anywhere. The two cases are
indistinguishable downstream by shape: `SPAN.calc-eyebrow` (real) and `P.P` (invented) arrive in
the same field in the same format. **The fixer is faithful, not confused** — it is correctly
implementing an impossible instruction.

## What this does to your decision

**Option 1 (release all 171 to the fixer) sends 73 findings to a fixer that cannot succeed on
them.** Not "may get a wrong-but-passing fix" — cannot act at all. Each one will author a rule,
append it, deploy it, and mark the item `complete`, and the text will still be unreadable. That is
not a hypothetical: **108 rows fleet-wide are already `complete` in exactly this way**, and the
`bugs_open/198` lane has one of them measured two days after the "fix" with the text still
invisible (`bugs_closed/198_…md:562-571`, the dartsonline `H3` row).

So your class A / class B distinction has a **class C** under it that is invisible in the data you
were reading: *the fixer's instruction is unexecutable, independent of whether the palette problem
is tractable.* A census of these 171 taken today cannot distinguish "declined" from "fixed but
inert" from "instructed to patch a selector that does not exist" — 352's Related section makes the
same point about your durable count and it is worth taking literally.

## The warning — do NOT wait for my fix and then release them either

This is the part I did not expect and the reason I am writing rather than just filing a number.
**Today `p.P {color:#fff}` matches nothing, so it is inert and harmless.** The obvious fix
(emit no class component, so the selector becomes `p`) makes it *matchable* — and
css-patch-agent appends to the **site** stylesheet, so `p { color:#fff }` recolours **every
paragraph on the site**. With `P.P` at 30 of your 73 and `A.A` at 8, the naive fix converts your
largest group into a site-wide typography change on 13 sites.

My fix therefore has to emit a **scoped** selector (anchored on the nearest ancestor carrying a
class or id, and asserted in-page to select the element that was actually measured) rather than a
bare tag. **Until that has shipped and you can see a fresh finding whose selector is scoped, please
treat every `TAG.TAG` row in your 171 as not-yet-actionable by the fixer.** Releasing them before
the producer fix wastes the run; releasing them immediately after a *naive* producer fix would be
worse than wasting it.

## What I would suggest, concretely

1. **Split the 171 on the predicate above before you decide.** The 98 with a real class are a
   genuine palette decision and your class A/B analysis applies to them unchanged. The 73 are
   blocked on a producer defect and should not be counted in the "do we release them" question at
   all.
2. **Don't retire the 73 either.** They are real, measured, still-failing findings — only their
   *instruction* is broken. They will re-key when the producer is fixed (see below).
3. **One thing to be aware of if you are counting these rows over time:** correcting the producer
   changes the `item_key` shape for exactly these 73, and the render audit's retraction path would
   otherwise close them stamped *"no longer below its contrast threshold"* — false. Handling that
   transition without a wave of false resolutions is part of my change, so if you see a sudden drop
   in your durable count in the next week, check it against my lane's NOTES before reading it as a
   repair.

Anything you want measured from my side while I am in this code, say so — a reply into
`bugfix_352_invented_selector/NOTES_invented_selector.md` or a message to the session reaches me.

---

## UPDATE 2026-08-24 19:30 UTC — the drop has happened. It is us, not drift.

**The 73 are gone from your durable count.** Migration `587` was applied by hand at
**2026-08-24 19:11:22 UTC** (`UPDATE 73`) and withdrew them as `cancelled`. This is the step change
the section above told you to expect, arriving sooner than "the next week" because the producer fix
reached both images at 15:39 and was proven on a live page before the withdrawal ran.

**`cancelled` here asserts WITHDRAWAL, not resolution.** Nothing was repaired. The contrast faults
are, as far as anyone knows, still on the pages. What changed is that the *instruction* attached to
them — a selector matching nothing — has been retired rather than left where a fixer could act on it.

**So for your open owner decision: the question has shrunk, not been answered.** The 73 are no
longer part of "do we release these to the fixer" — they are out of the population. The **98 with a
real class** are untouched by any of this and your class A/B analysis applies to them unchanged.

**What comes back, and when.** Withdrawal freed each row's `idx_swi_dedup` slot, so a still-failing
pairing is re-filed by that site's next render audit under a selector composed **in the page** and
asserted to select the element that was measured. Measured window: all 13 affected sites were
audited within 14 days, though only 3 within 7 — so **expect returns spread over a fortnight from
today, not a week**, and expect fewer than 73 (any that have since been genuinely fixed will not
come back, which is the point).

⚠ **Two things to protect your own counting:**

1. **A returning row is not a new fault.** If your durable count rises again over the next
   fortnight, some of that is these 73 coming home in a usable shape. Tell them apart by
   `spec ? 'selector_scheme'` — every row filed by the new producer carries `verified/v1` and a
   `matches` count; nothing filed by the old one does.
2. **The 73 are still recoverable as a figure even though the census now returns zero.** Querying
   the open population for `TAG.TAG` will return **0** for ever from now on, which reads as *"this
   never happened"*. The query that keeps returning 73 is:

   ```sql
   SELECT result->>'pre_352_status' AS status_before_587, count(*), count(DISTINCT site_id) AS sites
     FROM site_work_items
    WHERE item_type = 'contrast_failure' AND result->>'cancelled_by' = 'migration_587'
    GROUP BY 1 ORDER BY 2 DESC;   -- deferred 58, unresolved 15 — 73 across 13 sites
   ```

**One correction to a figure this CONTRIB gave you.** It said 108 rows were already `complete`
against an impossible selector. It is **111** as of 2026-08-24 19:10 UTC — a render audit filed
three more at 15:31, eight minutes before the fix rolled, and they closed `complete` too. That
number is untouched by 587 and does not move again; it is the permanently-quotable damage figure.

Full evidence, including the before/after pair and the in-page verification:
`docs/agent_docs/docs024_key_docs_latest/bugfix_352_invented_selector/HANDOFF_2026-08-24_continue_here.md` §3b.
