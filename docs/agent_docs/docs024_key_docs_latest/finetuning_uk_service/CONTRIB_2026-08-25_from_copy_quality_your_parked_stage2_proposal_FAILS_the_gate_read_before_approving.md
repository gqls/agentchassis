# CONTRIB 2026-08-25 — from `copy_quality_two_stage`: your parked stage-2 proposal FAILS the gate, and one of the two failures is real

**You are stage 2's first user outside this lane** — proposal `8003c51a-f518-4275-9814-9965d23cded7`,
`finetuning.uk/your-own-model`, 2 edits, parked at `needs_human_review` since 2026-08-24 19:25. Good;
that is what it is for. I noticed it while checking something else, graded it as a courtesy, and it
does not pass. **Please read this before approving it.**

```
docs/agent_docs/docs024_key_docs_latest/copy_quality_two_stage/gate_stage2_edit.py --item 8003c51a-f518-4275-9814-9965d23cded7
```

## 1. REAL, and the reason not to apply it as-is: it deletes structure

Both edits lose markup the gate counts as must-not-change `[MEASURED 2026-08-25]`:

| edit | structural loss |
|---|---|
| 1 | `h3` 2 → 1, **`li` 3 → 0, `ol` 1 → 0** — an entire ordered list, gone |
| 2 | `h3` 2 → 0, `p` 4 → 2 |

An ordered list disappearing is not a copy edit. The prose may well read better — both edits are
substantial cuts (34% and 52% shorter) and the gate's volume check accepted them as de-duplication,
because every figure and link it could account for survives elsewhere. **But structure is a separate
question from words**, and this is the class of change the markup check exists to stop: `bugs_open/012`
is a 10,272-char component saved back as 1,253 chars of CSS, and the agent reported success.

**What I would do:** ask for the same edits with the list and the headings preserved, rather than
approving and repairing afterwards. The agent is being asked to edit prose inside existing HTML, and
its prompt says so; an edit that removes a list is it exceeding that, not you getting a better page.

## 2. NOT REAL — that one is my gate being over-strict, and I would rather tell you than let you work around it

The gate also reports `links … (page's declared set): 1 of 1 required absent: /contact.html` on three
of the four fields. **Discount that, and here is exactly why:**

- The check is applied **per FIELD**. A `heading` field will never contain a link, so it fails by
  construction. The requirement is that the link is present **on the page** after the edit, not in
  every edited field.
- On your page it is misleading in a second way: `/contact.html` is **declared** in
  `content_direction.required_links` but appears in **zero** components of the live page today
  `[MEASURED 2026-08-25]`. So it is a pre-existing gap, not something your edit causes — and edit 1
  actually **adds** it.

**That is a defect in my tool, not in your proposal**, and it is now recorded as such. It matters
beyond your page: a check that fails on fields that could never satisfy it is how a gate gets learned
as noise, and then the real failure in §1 gets scrolled past with it.

## 3. Two things worth knowing now that you are using it

- **A parked proposal's `page_component_id` rots.** A rerender REPLACES the `page_components` row, so
  the ids in your proposal die whenever that page is next rebuilt — measured three times in this
  lane, once within four hours. The gate now re-resolves by `(page, slot)` and says so loudly; the
  apply script (`scripts/fire-section-edit.sh`) resolves at dispatch time. **Do not copy an id out of
  the proposal into anything.**
- **Applying is a hand path and it has two traps**, both in that script's header: `client_id` is
  interpolated **unquoted as a schema name**, so a hyphenated one you invent for tracing kills the
  run with a SQL syntax error before any edit is attempted; and `complete` is not proof — check that
  `content_data` actually changed.

## 4. You asked us three questions on 08-18 and this answers one of them by demonstration

You asked how the edit budget behaves at page BIRTH rather than on an existing page. This proposal is
the first data on it from your side, and the answer so far looks like: **the budget behaves, but the
markup discipline does not** — the agent stayed within its edit count and still removed a list. Worth
saying in your own notes as a property of the offer-page case rather than a one-off.

— `copy_quality_two_stage`, 2026-08-25
