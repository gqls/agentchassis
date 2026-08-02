# 178 — a "add a link to X" work item regenerates the WHOLE section and silently loses most of its prose; the item reports `complete`

**Filed 2026-08-02** (session "bugfix 19"), immediately after the triage in
`sql_for_agents/286` released the item that caused it. **Status: OPEN,
UNDIAGNOSED at the handler.** The effect is measured and certain on one page and
`[UNVERIFIED]` on three others with the same signature. **I have not read the
`page-build-handler` / `save_page_sections` path**, so this file makes no claim
about *why* a link-insertion item triggers a full regeneration. Run `090` before
fixing.

**Nothing is lost permanently — the prior content is in `page_component_history`.**
That is what makes this safe to leave open rather than hot-fix.

## What happened

Work item `93f2a3b7`, `item_type = content_rewrite`, summary:

> **"Add Gripper Safety Factor Calculator tool reference to how-to-specify-a-gripper page"**

It completed successfully in ~2 minutes (`status=complete`, `attempt_count=0`,
`error` NULL) and **it did add the link, correctly** — one contextual anchor to
`/tools/gripper-safety-factor-calculator/index.html` in each of the page's three
slots, with sensible surrounding prose.

It also **rewrote `generic-text-block` from scratch**, cutting it from 7
paragraphs to 4 and changing its heading:

```
content_data length   4439 -> 1806   (-59%)
heading  "Gripper Specification: A Systematic Approach"
      -> "Defining the Specification Parameters"
```

The HTML is well-formed and closes cleanly, so **this is not the `bugs_closed/012`
truncation signature** (a cut completion saved as a fragment). It is a shorter
*regeneration*. Substantive content that existed before and does not now:

- workpiece-first methodology (envelope dimensions, mass, surface condition,
  fragility threshold in Newtons, tolerance ranges);
- kinematic requirements (grip force at contact vs actuator force, stroke range,
  repeatability, with the worked ±0.05 mm / 150 mm-stroke counter-examples);
- cycle-time vs actuation-technology trade-off (pneumatic vs electric, compressor
  infrastructure, programmable force profiles);
- integration parameters (**ISO 9409-1** flange pattern, payload budget, I/O
  count, wrist force-torque sensing, collaborative-cell TCP mass limits);
- the closing synthesis paragraph.

The page is titled *"How to Specify a Gripper | Engineer's Reference"*. The lost
paragraphs were the reference.

## Why this is worse than it looks

The item's own summary says **add a reference**. Nothing in it asks for the
section to be rewritten. So the blast radius of a `content_rewrite` item is not
what its summary describes, and the difference is invisible downstream: the item
reports `complete`, the validator passes, `build_status` stays `deployed`, and no
record anywhere says content was dropped. This is the `WDS-004` family — *"we
finished" treated as "the work is right"* — in a path that 056 did not cover.

**Distinguish from the neighbours before quoting this file:**
- `bugs_closed/012` — tool-improver truncating a component into a fragment.
  Different: that is a CUT completion; this HTML is complete and well-formed.
- `bugs_closed/056` — regeneration dropping content that tripped a validation
  blocker. Different: nothing was blocked here; the rewrite passed validation.
  056 is closed and its fix does not address this.

## Fleet-wide signature `[PARTLY UNVERIFIED]`

Pages whose `content_data` shrank >25% against their most recent
`page_component_history` snapshot, last 7 days:

```
domain               url                                   before   now   slots   pct
fundamentallyai.com  /tools/review-council-simulator.html   35344   2900   3->3   -92
vonc.com             /about.html                            16030   8015  12->6   -50
relojistas.com       /glosario/index.html                    5899   3083   3->2   -48
vetcomparison.uk     /about.html                            13138   6803   4->4   -48
robot-hands.com      /how-to-specify-a-gripper.html          5795   3323   3->3   -43
gamesdesign.co.uk    /tools/bayesian-ranking.html            7926   5751   4->4   -27
```

**Read the slot column — it discriminates.** `vonc.com/about.html` (12→6) is
almost certainly the legitimate duplicate-component removal from
`bugs_closed/156`, and `relojistas` lost a slot too. The four with an **unchanged
slot count and a large content loss** are the ones matching this bug's signature.
Only `robot-hands.com` is CONFIRMED (I read both versions); the other three are
`[UNVERIFIED]` — same shape, cause not established, do not cite them as instances
without reading their before/after.

**Explicitly NOT an instance:** `bugs_closed/154`'s witnessed `tool-improver` run
the same morning. Its component *grew* (`component_versions` 7944 → 9158 →
10520), so that verification stands.

## Recovery (the exact rows)

```sql
-- pre-write snapshot, captured by the writer itself at 10:41:22.93696Z
SELECT content_data FROM page_component_history
WHERE id::text LIKE 'ecb4b420%';        -- the 4439-char generic-text-block

SELECT id, source, length(content_data::text), created_at
FROM page_component_history
WHERE page_id='5a385981-c2fd-4edb-bc4d-927b93177281' ORDER BY created_at DESC;
```

**Restoring is an owner decision, not a mechanical one**, because the naive
restore also removes the tool link the item was legitimately raised to add. The
merge — old prose plus the new anchor — is an editorial act on a live customer
page. Not done by the filing session for that reason.

## Fix candidates, ordered by what closes the door

1. **A link-insertion item should insert a link.** If `content_rewrite` is being
   used as the carrier for "add a crosslink", the crosslink case wants a handler
   that edits rather than regenerates. Makes the loss unrepresentable.
2. **Refuse a regeneration that shrinks a section past a floor, and record it.**
   The `prune_floor.go` / CTXA-025 machinery from `bugs_closed/135` is the
   existing pattern for exactly this shape — a destructive operation that must
   prove it saw the corpus. Reuse before building.
3. **Make the loss visible even when allowed**: the writer already snapshots to
   `page_component_history` immediately before overwriting, so the delta is
   computable at write time and is currently thrown away. Emitting it would have
   surfaced all four fleet cases on the day they happened.
4. Sweep/restore the affected pages. Cleanup, not a fix; needs 1–3 first or it
   recurs.

## How to verify a fix

Raise a crosslink item against a page with a long prose section and assert the
section's `content_data` length is unchanged apart from the inserted anchor:

```sql
SELECT length(content_data::text) FROM page_components
WHERE page_id=<target> AND slot_name='generic-text-block';
-- expect: prior length + ~90 chars, NOT a wholesale replacement
```

Do **not** verify by the work item reaching `complete` — it did that here, and
that is the whole problem.
