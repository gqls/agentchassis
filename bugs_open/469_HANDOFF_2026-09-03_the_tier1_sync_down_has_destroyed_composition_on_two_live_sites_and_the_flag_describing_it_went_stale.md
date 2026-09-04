# 469 — the tier-1 sync-down has already DESTROYED page composition on two live sites, and the detector's own flag went stale so nobody could tell

Filed 2026-09-03 by the `bugs_open/427` lane (session "427") while triaging the
`section_source_drift` backlog. **Status: OPEN, unowned.**

**Severity: MEDIUM.** The damage is done and is not spreading fast (two confirmed
instances in five weeks), but it is *silent*, it destroyed deliberate human work, and
the one mechanism built to warn about it reported the damage and then went quiet in a
way that reads as "resolved".

## 0. First-hand verification, per CLAUDE.md's 2026-07-31 ruling

No `090` run. Substituting first-hand verification, all `[MEASURED 2026-09-03]` against
the live `clients_db` in this session: the three stores and the live `page_components`
rows were queried directly for every page named below, and the reverting code path was
read end-to-end at source (`load_page_sections_from_spec_action.go`) rather than
inferred from behaviour. The mechanism is not a hypothesis — it is documented in
migration `154`'s own header from 2026-07-15, and this file's contribution is the
evidence that it has since **fired again and completed**.

## 1. The mechanism (established, not new)

A page's section list lives in three stores, resolved in priority order by
`load_page_sections_from_spec_action.go` (the page-build step `load_spec_sections`):

1. `site_plan_sections` for the current plan — **authoritative** (`:142-148`)
2. `site_specs.site_plan` aspect
3. `pages.sections` — the materialised cache

and the winner is **synced DOWN over `pages.sections`** (`:558-570`).

So an edit made only to the cache is destroyed by the next page **build**. No re-plan is
required. (A re-plan is in fact *safe*: `reconcilePlanWithRealised`,
`v3_site_actions.go:7701-7724`, snaps a `deployed`/`needs_rebuild` page's realised
`pages.sections` back **onto** the plan proposal first. `bugs_open/427` §19.2 has this
backwards and is corrected there.)

## 2. What is NEW here: the loss has completed, twice

`check_section_source_drift` correctly flagged both of these at the time. Both items
then sat open at `needs_human_review` — the check is deliberately flag-only
(`HandlerAgent: ""`) — until this session closed them.

| site / page | item filed | what the cache held at filing | what all three stores say today |
|---|---|---|---|
| `robot-hands.com` / `gripper-catalog` | 2026-07-28 | `[hero, generic-text-block, **gripper-spec-sheet**, info-card-grid, call-to-action]` | `[hero, generic-text-block, info-card-grid, call-to-action]` |
| `idea.uk` / `guides-index` | 2026-08-04 | `[hero, **guide-list**]` | `[hero, content-listing]` |

`[MEASURED 2026-09-03]` in both cases the live `page_components` rows agree with the
authority too — so this is not a stale cache, it is the **component genuinely gone from
the page**.

**`gripper-spec-sheet` is the exact component migration `154` was written to rescue on
2026-07-15**, after `153` swapped it in by writing only `pages.sections` + the aspect.
It was rescued, and then lost again. That is the finding: the remedy did not hold, or a
later write re-introduced the divergence and the sync-down won the second time.

> **CORRECTED 2026-09-03 (later) — `idea.uk/guides-index` is NOT a loss, on the owning
> lane's own first-hand check.** The `idea.uk` lane replied to my notification, verified at
> the artefact, and reported: `/guides/index.html` serves 200 / 90,042 B with `hero` +
> `content-listing` both deployed (rebuilt 2026-09-03), 19 card elements, **all 10 guides
> linked**; and their lane records back to July carry no memory of a distinct `guide-list`
> section that differed from the current listing. So on that site the
> `guide-list` → `content-listing` change reads as a **one-for-one rename**, and nothing a
> visitor would miss is gone. The direction test correctly said "the authority won"; it
> cannot say whether what the authority won *with* was equivalent — which is precisely why
> `authority_won` must route to a human and never auto-close as resolved. **So this file's
> confirmed-loss count is ONE (`robot-hands.com`), not two**, with `idea.uk` reclassified as
> a benign rename and the question still in front of its owner. Caught by notifying the lane
> rather than by any check of mine.

A third, `oufe.com/contact`, also resolved authority-wards (the cache had lost
`contact-info`; the authority restored it). Recorded for completeness, but the direction
test cannot distinguish "a deliberate removal was destroyed" from "an accidental
omission was repaired", and this one looks like the latter. **Do not count it as a third
loss without looking.**

## 3. Why nobody noticed — the part worth fixing

Three properties compound:

> **CORRECTED 2026-09-03 (later) — "nothing ever closes an item" is FALSE, and the way I got
> it wrong is the more useful finding.** `[MEASURED 2026-09-03]` **two** `section_source_drift`
> items were closed: `section_source_drift:contact` (filed 2026-07-16) and
> `:who-we-help` (filed 2026-07-17). Both carry hand-written narrative receipts.
>
> ⚠ **CORRECTED AGAIN, same day, by the `bugs_open/469` lane — and my first correction was
> itself wrong in the same shape.** They were NOT "closed on the day they were filed by two
> people". Both were closed at **2026-07-19 13:48:11**, *83 milliseconds apart*, by **one**
> thread (`handled_by = 'bugfix thread (bugs_open/002 C)'`) — two and three days after filing.
> I read **`updated_at`**, which on these rows was never bumped on close and still equals
> `created_at`. The closure timestamp is **`completed_at`**. So `updated_at` is not a closure
> time, and it is unreliable as one *because different writers treat it differently* — my own
> migration `753` sets `updated_at=NOW()` when it closes a row, so on rows I closed it does
> track closure. **Read `completed_at`.**
>
> **I could not see them because they are in `site_work_items_archive`, and my census queried
> only the live table** — which returns 6. A closed row leaves the table you are counting in,
> so *a census of open items cannot observe the successes*, and the absence of closures is
> exactly what it will report. That is MEMORY's own *"a closer census cannot see what it
> SUCCEEDED at"*, which was in my index while I wrote the opposite.
>
> **The accurate claim is narrower and still damning:** there is **no automated closer and no
> handler** — `CheckResult.Resolved` is never populated — so closure depends entirely on a
> human noticing *at the time*. Two people did, in July, same-day. The six that accumulated
> afterwards did not get that attention, and the oldest sat 37 days. The defect is not that
> closure is impossible; it is that it is **manual, undriven, and invisible once it happens**.
> Query `site_work_items` **UNION** its archive, or you will re-derive my mistake.

- **The check is flag-only and nothing ever closes an item.** Six were open when this
  session looked, the oldest 37 days.
- **The item's `spec` is frozen at filing time.** `spec.authoritative` and
  `spec.pages_sections` are a snapshot, and they read as current. A reader triaging the
  backlog by reading the items learns nothing about today.
- **An open item SUPPRESSES re-filing.** `idx_swi_dedup` is `UNIQUE (site_id, item_key)`
  over non-terminal statuses, so while a stale item sits open, genuinely new drift on
  that same page cannot be filed. The detector blinds itself on exactly the pages it has
  already flagged.

So the sequence is: drift is detected → flagged → nobody has a handler → the build wins
→ the stores agree again → **the item now describes a state that no longer exists and
looks like it might be resolved** → and any fresh drift on that page is invisible behind
it.

## 4. What this session already did

- Migration `753` closes only items whose stores agree again, re-derived at apply time,
  and records a `direction` (`cache_held` / `authority_won` / `third_list`) in each
  receipt so a close cannot silently ratify a loss. Four closed; three are
  `authority_won`. `apis.uk/index` — a live divergence owned by another lane — was
  excluded by the data and left open.
- Migration `750` corrected `boxingonline.com/tool-fight-calendar`'s authority to match
  its live page, so `bugs_open/427`'s fix stops being one build away from destruction.

Neither addresses **this** bug, which is about the two pages that already lost their
composition and about the blindness that let it happen unremarked.

## 5. What is actually open

1. **Should `gripper-spec-sheet` be back on `robot-hands.com/gripper-catalog`, and
   `guide-list` on `idea.uk/guides-index`?** This needs a *human* who knows what those
   pages are for. A machine cannot tell an intended removal from this bug's completion —
   which is precisely why the revalidator design for this item type must classify such a
   case as `unknown` and never as `resolved`. Owner or the owning lanes
   (`robot_hands_gripper_dossier`, `idea.uk`).
2. **The detector needs a closer.** Any check whose items nothing ever resolves will
   accumulate exactly this failure. `check_section_source_drift` never populates
   `CheckResult.Resolved`.
3. **The framework gap that produced the divergence in the first place** — there is no
   typed way to write a composition correction to all three stores, so every lane
   hand-writes SQL and some write only the cache. Going to architecture review as an RFC
   out of the `427` lane; see `bugs_open/427` and the `bugfix_427_event_render` lane docs.

## 5a. The bytes are recoverable; the LIST is not (contributed by the `bugs_open/469` lane, 2026-09-03)

A correction to a premise §5 leans on. This file implies the destroyed sections are simply
gone. They are not, wholly:

**`page_component_history` carries `slot_name`, `position` and the bytes on every
`artefact_archive_trigger` delete** (`bugs_open/357`'s mechanism), so from roughly
**2026-08-09** onward a dropped section's identity and HTML survive the loss. The 469 lane
counted **24 rows for `gripper-spec-sheet` alone**, across `product-detail` and
`gripper-detail`.

**What does NOT survive is the LIST** — the fact that the page had five sections *in that
order*. Every rebuild deletes every row for the page, so "a section was dropped" is only
derivable by **diffing consecutive builds**, never by reading the current state.

Two consequences: restoring a lost section is much cheaper than §5 implies (the bytes exist);
and the thing genuinely worth a durable receipt is the *composition delta*, not the content —
which is what the 469 lane's `section_composition_lost` receipt is being pointed at, rather
than building a second archive.

**Ownership, recorded so it is not raced:** the detector half of this bug (a direction-aware
`Resolved` arm on `check_section_source_drift`, plus that receipt) is owned by the
`bugs_open/469` session as of 2026-09-03. The WRITE path — `RFC_064`,
`apply_page_composition`, any migration correcting a plan's rows — stays with the
`bugs_open/427` lane. Neither writes to the other's files.

## 6. Cross-references

- `bugs_open/427` §19–§21 — the mechanism, corrected; the migration `750` precedent.
- `docs/agent_docs/sql_for_agents/153_gripper_detail_page_swap.sql` and
  `154_product_detail_plan_sections_fix.sql` — the original 2026-07-15 case on this very
  site, and its hand-written remedy.
- `docs/agent_docs/sql_for_agents/750_…`, `753_…` — this session's two migrations.
- `bugs_open/443` — the adjacent class: pages born with a layout in the cache and
  nothing in the authority (`create_blog_posts`).
- `LANDMINES.md`, "`pages.sections` is a materialised CACHE" — the standing entry.

---

## 7. Session 2 (lane `bugfix_469_drift_closer`), 2026-09-03 evening — what changed

Picked up by a session that read this file's "OPEN, unowned" and confirmed the split with
the `427` session directly: **this lane owns §5.2 (the closer); `427` keeps RFC_064 and the
write path.** Lane docs: `docs/agent_docs/docs024_key_docs_latest/bugfix_469_drift_closer/`.

### 7.1 The bug is still valid, and the backlog is currently empty [MEASURED 2026-09-03]

Fleet-wide drift is **zero**, with demand controls: **398** tier-1 page comparisons all
equal, **34** tier-2 aspect comparisons all equal. All six `section_source_drift` items are
`complete`. The zero is **conservative** — the raw comparison omits the locked-row merge,
which can only turn a mismatch into agreement, never the reverse, so a raw zero implies a
merged zero.

⚠ **Do not read "six items complete" as health** (the `427` lane's caveat, and it is right):
closing them freed their dedup keys, so a still-broken page re-files on its next sweep.
Until each of those six sites has had a discovery pass, part of the quiet is "not yet
re-asked". This does **not** weaken the store measurement above, which never consults the
item table.

So: **the damage is historical, the safety net is still missing.** That is what §5.2 fixes.

### 7.2 §5.1 is ANSWERED on the fact, and BLOCKED on the execution

The `robot hands` lane replied with provenance this file did not have:
`docs/agent_docs/docs024_key_docs_latest/robot_hands/SQL_2026-07-24_r9_gripper_catalog_real_grid.sql`
("R9") put `gripper-spec-sheet` on `gripper-catalog` at position 3 on 2026-07-24,
deliberately **instead of** `product-grid` — whose empty e-commerce fields would "invite
fabrication", where spec cards are "the honest fit". Framed in that lane's HANDOFF as an
owner call. No later reversal exists.

**So it is damage, not an intended removal.** §5.1's first question is closed.

The repair is written and **held**: `docs/agent_docs/sql_for_agents/760_restore_gripper_spec_sheet_to_gripper_catalog_HOLD.sql`
(+ `_ROLLBACK`), dry-run proven against the live database in a rolled-back transaction, with
an induced failure showing the pre-check actually aborts. It is held on **three** owner
questions, and the third would moot it:

1. **RFC_064 §7 q2.** The page is `built_from_plan_version = <current plan>` and
   `build_status = 'deployed'`, so `decideEmit` returns `skip_built` **before** comparing any
   section list. A corrected plan would sit there and **never render**. This is where q2 stops
   being about provenance and becomes load-bearing: without the ruling the repair does not
   merely lose history, *it does not happen*.
2. ~~**Does the build path reach an archived page at all?** Unknown; not assumed either way.~~
   **ANSWERED 2026-09-04 — and it makes q1 insufficient on its own.** The reconciler *does*
   reach it (neither `loadPlanPages` nor `loadRealisedPages` filters `pages.status`; the page
   is in the current plan at `nav_order=2`), **but the deploy is then refused** by
   `archived_page_guard.go` (`580af7ff0`, `bugs_open/266`), which guards `git_commit` AND the
   `update_page_status` stamp on a literal `status == "archived"` test. Live in v1.0.1360
   (`git merge-base --is-ancestor 580af7ff0 239ab3626` -> YES; control -> NO), and **already
   proven against this page**: `[MEASURED 2026-09-04]` **308** `ARCHIVED_PAGE_DEPLOY_REFUSED`
   rows all-time, **3** naming `page_id 64fab29e-5d8a-4a50-ad1b-2f9b0721cef6` (this page,
   joined by id), last 2026-08-23.
3. **Should the page be serving at all?** See 7.3. **This is now PRIOR to q1 and q2, not
   third**: while the page is archived no repair can ever deploy, so the fork is binary —
   un-archive it (q1 becomes operative, `760` applies and renders), or leave it archived
   (`760` is **moot**; the defect is the serving 200, fixed by retraction, which the guard
   deliberately does not block — it dispatches `delete_file`).

Migration `750`'s template does **not** transfer: rename vs INSERT-with-shift; an
already-correct page vs both stores wrong; **1** `site_plans` row vs **5**; artefact unchanged
vs a rebuild required.

### 7.3 The page is `archived` — and it serves 200

```
./scripts/probe-page-url.sh robot-hands.com gripper-catalog gripper-catalog-index
gripper-catalog        /gripper-catalog.html        200 SERVING
gripper-catalog-index  /gripper-catalog/index.html  200 SERVING
controls: invented=404 (want non-200)  sibling=200 (want 200)
```

`pages.status = 'archived'`, last built **2026-08-11**, while every *active* page on that site
rebuilt on 2026-09-03. `gripper-catalog-index` is **not** a replacement — it carries a single
`news-listing`.

This was **already known** and I initially overclaimed it as new: `bugs_closed/359` names this
page by exact byte count on 2026-08-26, and its detector (migration `648`) fired. What nobody
had connected is that **the page carrying 469's composition loss is the same page carrying an
un-triaged serving flag**. Nine `archived_page_still_serving` items exist fleet-wide — eight
filed 2026-08-26/27 in per-site pairs, one on **2026-09-02** — and **all nine are still
`detected`**, none triaged, `handler_agent = ''`.

**Answering q1 alone implicitly un-retires this page.** The three questions go to the owner
together or the serving flag stays orphaned even after a composition fix lands.

### 7.4 A correction to this file's implied premise: the BYTES are not lost

§5.1 leaves the restore decision to a human partly on the implication that the content is
gone. **It is not.** Migration `357`'s trigger pair archives every deleted `page_components`
row with `slot_name`, `position` and `rendered_html` — `SELECT count(*) FROM
page_component_history WHERE slot_name ILIKE '%gripper-spec%'` returns **24** (on
`product-detail` and `gripper-detail`, ~12.4 KB each; it was never lost from those pages).

**What is NOT recoverable is the LIST** — that the page had five sections in that order —
because `DELETE`+`INSERT` is the rebuild lifecycle, so the archive holds a delete for *every*
section on *every* build, and "which one was dropped" is only derivable by diffing consecutive
builds. That distinction is what the closer's receipt has to carry, and it is why this lane
points at the existing rows rather than building a second archive. (I was one query from
proposing exactly that; logged in `WRONG_CALLS.md`.)

### 7.5 The class, measured — and a distinction that will mislead a backlog sweep

71 discovery checks; **19** can retract via `CheckResult.Resolved`; **18** file flag-only
items and **10 of those have no closer at all**. Handler-less open items fleet-wide include
`needs_section_data` at **172 days**.

But "an old flag-only item" is **not one defect**, and the two shapes want opposite remedies:

| | `archived_page_still_serving` | `section_source_drift` (this bug) |
|---|---|---|
| has a `Resolved` arm | **yes** | **no** |
| why still open | the finding is **still true** | nothing could ever close it |
| what it says today | accurate | describes a state that **no longer exists** |
| blocking anything? | no | **yes** — the dedup key |
| remedy | triage / routing | a closer that **cannot ratify the loss it observed** |

A check with a working arm and a real unfixed defect produces the same backlog row as a check
with no arm and a defect that completed weeks ago. **The discriminator is whether the item's
own predicate still holds** — which nothing re-derives today, and which the closer must
re-derive before it touches anything.

### 7.6 §5.2 IS FIXED IN CODE — `fc9cad600`, inert until the next chassis roll

The closer is built, and the part worth reading is what it refuses to do.

**A naive closer would have been worse than none.** For this check, *"the finding
no longer reproduces"* and *"the damage completed"* are **the same observation** — the
stores agree again precisely BECAUSE the sync-down ran. Retracting on agreement would
close, automatically and fleet-wide, exactly the cases that most need a human. So:

- **The test is `lost`, not `direction`.** Migration `753`'s three-way label cannot
  separate `oufe.com/contact` (the authority RESTORED a section the cache had dropped —
  nothing destroyed) from `gripper-catalog` (the authority DELETED one). Both are
  `authority_won`. The retraction computes the loss instead: frozen-cache-only names,
  minus what is on the page today. `direction` is still recorded on every close,
  byte-compatible with `753`.
- **A lossy retraction cannot happen without its receipt.** `ResolvedFinding.Receipt`
  (register **WII-039**) makes the record a **precondition**: `resolveWorkItems` writes
  the `section_composition_lost` item FIRST, in the same transaction, and refuses the
  close if it can neither insert it nor confirm an open row holds its key. The coupling
  is on the SEAM, not in the check, because `resolveWorkItems` has two callers and a
  control in either protects only that one.
- **The receipt COPIES its evidence** (register **WII-040**) — both frozen lists, today's
  list, the lost/gained multisets, and a count of `page_component_history` rows holding
  the destroyed bytes. It points at the archive for the bytes and carries the list itself,
  because §7.4's asymmetry means the list survives nowhere else.
- **Detection-time grading** is the only part that acts BEFORE a loss: `would_drop` /
  `would_add` / `would_drop_present` in the spec, severity `high` only when the sync will
  drop a name the page currently carries, and the summary leads with it.

Proven by **mutation**, run and observed failing rather than asserted: 11 mutations of the
check and 4 of the seam, each killing a named test, source restored and diffed
byte-identical. **One of the four survived first time** — a guard in SERIES was doing the
work — and the test was sharpened rather than the survival read as proof.

### 7.7 What is still open on this bug

1. **§5.1's EXECUTION** — `760_..._HOLD.sql` is written and dry-run proven; it needs an
   owner ruling on RFC_064 §7 q2 (may a non-planner withdraw `built_from_plan_version`?),
   plus the two questions in §7.2/§7.3. **Not a code problem.**
2. **§5.3** — RFC_064, the typed writer, open with the owner, owned by the `427` lane.
3. **The receipt lands in a queue nobody drains.** Stated rather than glossed: the estate
   has 331 open `capability_gap` items at 42 days and nine untriaged
   `archived_page_still_serving` items. This change does not fix that and does not claim
   to — the safety property is not that the receipt is READ, it is that **the close cannot
   happen without it**, so a loss is never laundered into "resolved". The drain is
   `bugs_open/033`.
4. **The other nine flag-only checks with no closer.** The seam is reusable; each
   predicate is its own. Not attempted here.

**This bug should NOT be closed yet** — the fix is committed but inert until the next
chassis roll, and the estate's bar is fixed AND live.

---

## 8. 2026-09-04 — the closer is LIVE. §5.2 is CLOSED; §5.1 and §5.3 are not.

### 8.1 Proven at the artefact, with a control

The fleet rolled to chassis **`239ab3626`** (v1.0.1360), pods up **2026-09-03 22:06Z**.

```
git merge-base --is-ancestor fc9cad600 239ab3626   → YES   (the closer is in the binary)
git merge-base --is-ancestor 2cca1b085 239ab3626   → YES   (the council follow-ups)
git merge-base --is-ancestor 5f676db58 239ab3626   → YES
git merge-base --is-ancestor 152f47b65 239ab3626   → YES
CONTROL: is-ancestor 29d611750 239ab3626           → NO    (today's 11:55 commit — correctly absent)
```

> ⚠ **A `grep` for your own commit sha in `/proc/1/exe` is the WRONG instrument and I ran it
> first.** The binary carries **one** stamp — the commit it was BUILT from — not a list of
> ancestors. So `fc9cad600` reads **absent** in a binary that fully contains it, and that
> absence looks exactly like "not shipped". The stamp probe is for reading *which* commit
> built it; the ancestry test is what answers "did my fix ship". Both are in CLAUDE.md and I
> used them in the wrong order.

### 8.2 Live, error-free — and NOT yet exercised. The distinction matters.

`[MEASURED 2026-09-04]`, ~14 hours after the roll:

| | count | reading |
|---|---|---|
| new `section_source_drift` items since the roll | **0** | no fresh drift |
| `section_composition_lost` receipts | **0** | nothing resolved lossily |
| rows carrying `result->'resolution_evidence'` | **0** | no retraction has run |
| `DISCOVERY_CHECK_ERROR` naming this check | **0** | the new queries have not failed |

**Demand controls, because four zeros with no control are indistinguishable from a check
that never ran:**

- the owning agent **did** run — `completeness-discovery-agent` filed items across **9
  distinct hourly windows** since the roll;
- drift **is** genuinely absent — 398 tier-1 comparisons, all agree;
- the error query **can** find rows — **97** `DISCOVERY_CHECK_ERROR` all-time, **5 since the
  roll** (other checks: `archived_page_still_serving` correctly BLINDED by its own controls,
  `structure_floor`), and **none of them this check**.

**So the closer is deployed, running, and has never fired.** That is the correct behaviour on
an estate with no drift — and it is *not* the same as proven. **The first live retraction is
still owed as evidence**, and the disconfirming signal to watch for is a new
`section_source_drift` row followed by either a clean retraction or a
`section_composition_lost` receipt.

### 8.3 §5.2 — CLOSED

> **§5.2 "The detector needs a closer" is CLOSED as of 2026-09-04.** Fixed
> (`fc9cad600`), council-APPROVED (`009fabca`, round 1), and LIVE on `239ab3626` — the
> estate's bar of *fixed AND live* is met. Its residual is exercise, not correctness, and
> that is tracked in §8.2 rather than as an open bug.

### 8.4 What keeps this bug OPEN

- **§5.1's EXECUTION** — `760_..._HOLD.sql`, re-verified against live state on 2026-09-04
  (pre-checks pass, dry-run rolls back clean, damage unchanged, page still `archived`).
  Blocked on an owner ruling, not on work.
- **§5.3** — `RFC_064`, still `Status: OPEN`, the `427` lane's.
- **`RFC_066`** — new, filed at the council architecture seat's request; still `Status: OPEN`.
