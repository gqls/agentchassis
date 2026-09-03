# HANDOFF — components lane, 2026-09-02 (evening)

**Read this first, then `bugs_open/425`.** Everything below is either measured with its
control stated, or explicitly marked as not-established. Six of my own claims turned over
during this session, every one because I read a column instead of the artefact — so treat
any unmarked assertion here as owing you a re-run, and prefer the artefact queries.

---

## 0a. POST-ROLL RESULTS — token refreshed 2026-09-02 ~21:15Z, §0 items 1 and 2 DONE

**The defect SURVIVES v1.0.1355.** A `template_changed` rerender of boxingonline `/index.html`
filed 21:21:06Z, completed 21:22:43Z, **wrote** (`page_components.updated_at` 21:22:14) and
produced the **old shape**: no `excerpt` key, title still suffixed. So the four other commits in
that roll touching `rerender_page_sections_action.go` / `plan_sections_action.go` /
`queryresolve/` (444, 443, 137-residue, 427) did **not** fix it. That was a real measurement, not
a confirmation — the `site_delivery_and_editor` lane corrected its own "should not change the
answer" to an inference from commit subjects before I ran it.

**§0 item 1 done, and it closes the hole I could not close before.** Both pods on the current
build, both controls:

| pod (started) | `ListItemExcerpt` | `resolvePagesWhereType` (+control) | invented symbol (−control) |
|---|---|---|---|
| `8ddbf8958-cd2h9` (20:56:43Z) | PRESENT | PRESENT | absent |
| `8ddbf8958-vppjz` (20:57:10Z) | PRESENT | PRESENT | absent |

**So the current finding is not a pod-variance artefact**: two pods, both carrying the fix, and a
rerender still produces the old shape. (This does not retroactively close the 13:5x-era
one-pod-of-two hole — those pods are gone — but it makes that hole irrelevant to the live defect.)

**Next: the ⭐ experiment in §2.** It is now the only cheap thing left that discriminates.

## 0d. MIGRATION 721 — working mechanically, ONE page improved so far, and my own count went stale in 12 hours

`[MEASURED 2026-09-03 ~08:50Z]` instances of the six hero components carrying
`content_data.background_image`:

| era | rows | window |
|---|---|---|
| **before 721** | **5** | 08-25 20:18 → 09-02 14:00 |
| **after 721** (applied 20:15:47Z) | **23** | 09-02 20:20 → 09-03 05:50 |

**So the field resolves and lands.** 721 is not inert — 23 instances have gained
`background_image` through resolution since it applied, which is the mechanism working.

**But the visible improvement is ONE page.** Of those 23: **22 have no page-scope hero asset**, so
they take the declared fallback or the site hero — correct behaviour, no change to what a reader
sees. **1 renders its own hero.** Only a page that BOTH has its own asset AND has re-rendered
since 20:15Z can improve, and most overnight re-renders were on pages without one.

**The designblog lane predicted exactly this and was right**: the fix is *necessary and not
sufficient*; closing the class needs re-renders carrying a re-resolving reason across the affected
pages, then verification at the SERVED bytes, filename-anchored. The component row now reads
correct everywhere while the pages still serve the old image.

> ⚠ **AND MY OWN "5 of 158" WENT STALE IN TWELVE HOURS — by ADDITION, caused by my own change.**
> I published that count in migration 721's header and in its council submission, and used it to
> reject a proposed cheaper route. It was correct when measured (2026-09-02, pre-migration) and is
> now 28. Nothing about it was wrong; it simply counted a population my own edit then grew. This
> is precisely why CLAUDE.md requires a count to carry the date it was counted — the header says
> `[MEASURED 2026-09-02]`, so the staleness is visible rather than silent, and
> `--since 2026-09-02` re-derives it. **A census does not go wrong; it goes stale, and yours can
> be stale because of you.**

## 0g. WHERE THE SESSION ENDS — two experiments queued, and the queue is getting WORSE

| experiment | batch | reason | state |
|---|---|---|---|
| **deck discriminator** — `guides-index`, new-shape baseline | `…000688` | `template_changed` | **triaged**, filed 08:26 |
| **hero test** — `remortgagecalculator.uk/about`, unpinned | `…000690` | `image_landed` | **triaged**, filed 08:55 |
| ~~advertise.co.uk/about~~ | `…000689` | — | **cancelled for a WRONG reason** (see §0f); the row carries a dated correction |

`[MEASURED 2026-09-03 ~09:0xZ]` **234 `page_rerender` triaged, and only 9 completed in the last 30
minutes.** Earlier in the session it was 192 triaged draining ~29 per 30 minutes. **The backlog is
growing faster than it drains**, so neither experiment is hours away — it may be considerably
longer, and that is worth knowing before anyone re-files thinking the item was lost. **Both items
are queued, not stuck; find them by batch id, do not duplicate them.**

> **THE WATCHER FIX WORKED, which is worth one line.** After the first watcher timed out and
> printed the unchanged baseline under the heading "DISCRIMINATOR RESULT", I added the item's
> terminal status to the same output. The second watcher expired too — and printed
> `item outcome: triaged` beside the reading, so the output labels itself as "did not run" rather
> than presenting a baseline as a finding. **Any query that reads an artefact to judge a dispatch
> should print the dispatch's status in the same breath.**

## 0f. ⚠ A RETRACTED FINDING — "33 of 57 are version-pinned and unreachable" WAS WRONG

`[MEASURED 2026-09-03]` I found that 33 of the 57 repairable hero instances carry a
`component_version_id` pinned to a version predating migration 721, and that those pinned
versions lack the `background_image` field. I reported that as **33 instances unreachable by any
rerender**, proposed it as a distinct repair needing its own owner, and a peer lane was routing an
ownership decision to the owner on it.

**It is wrong. `page_components.component_version_id` is WRITE-ONLY.**
`save_sections_component_version.go:40`, verbatim:

> *"THIS FILE ONLY WRITES. Nothing reads component_version_id, so this change is inert by
> construction: it cannot alter what any page serves."*

Its header adds, measured 2026-08-22: *"0 of 1,930 rows populated, no Go code writing it, no Go
code reading it — dormant machinery."* The only reader is `loadStoredSections`, which SELECTs it
to carry the stamp forward, **not** to resolve a schema. Component resolution keys on
`component_id` against the **live** `content_components` row, so a pinned page resolves exactly
like an unpinned one.

**What that costs and returns:**
- there is **no unreachable category** — the split is 57 repairable, full stop;
- batch `689`'s cancellation was for a wrong reason (the page was always a valid test). The row's
  error now carries a dated correction; `690` supersedes it and is equally valid;
- **the dartsonline result returns to its face value** — a `section_data_resolved` rerender,
  sections path, attributed by `source_item_id`, that did **not** resolve the newly-declared
  field. I had explained it away. It is unexplained again, and it is a **second** data point
  pointing where §2 points, on a *different source type* (`site_assets.*`, not `query.*`).

> **The error, because it is the seventh on this thread and the pattern has not varied:** I found
> a **correlation** — pinned AND missing field — and reported **causation**, without checking that
> the mechanism exists. Two greps would have caught it, and I only ran them because the estate's
> landmine list pointed me at the pin, which made me ask what the pin actually does. **Before
> attributing an outcome to a mechanism, confirm something READS the thing you are blaming.**

## 0e. THE HERO CLASS AND THE DECK CLASS MAY SHARE ONE ROOT CAUSE — one-page test filed

`[MEASURED 2026-09-03]` **57 instances across 24 sites** have their own page-scope hero and are not
rendering it (5 already correct, 104 with no own hero — nothing to repair). Derived independently;
the `bugs_open/114` census says 61, same class.

**The proposed repair — "re-renders carrying a re-resolving reason" — is an UNTESTED premise, and
my one data point contradicts it.** I traced the single page that has visibly improved since 721
(`garden-tools.uk/contact`, 09-02 23:18) via `page_component_history.source_item_id`:

```
item 726aa1e5 · type unbuilt_internal_link · handler page-build-handler · reason (NONE)
```

**The BUILD path.** Not a re-render, and carrying no reason at all. So the route we were about to
dispatch 57 items down has **zero** confirmed successes, while the one success came via the route
nobody proposed. The delivery lane has marked its own relayed condition `[UNTESTED → under test]`
— it was a hypothesis wearing the voice of a fact.

**ONE DISPATCH, NOT 57.** Filed batch `00000000-0000-0000-0000-000000000689`:
`advertise.co.uk/about`, `reason='image_landed'`, chosen because that site has **0 open work
items** so attribution cannot be muddied. Baseline recorded before it runs: instance
`4e681d76`, `hero-about`, **no `background_image` key at all**, `updated_at` 09-02 17:02:11.
Its own hero asset is `hero_about`.

> **CHECKED BEFORE READING THE TEST — every post-721 rerender on this class was ASSEMBLE MODE.**
> `[MEASURED 2026-09-03]` **66** completed `page_rerender` items since 20:15:47Z on pages carrying
> one of the six hero components with their own hero asset. **All 66 carry `reason = (none)`.
> Zero qualifying reasons. Not one routed to the sections path.**
>
> So the "nine re-rendered since 721, none recovered" evidence **proves nothing about the sections
> path** — it was never exercised. That makes the one-page test *necessary* rather than
> confirmatory: batch `689` will be the **first** time the sections path runs against this class
> since the field became declarable.
>
> ⚠ **And beware the word RECOVERED in that sweep.** Ten rows read as recovered, all on
> `leopardessconsulting.co.uk` and `garden-tools.uk/contact` — but those are the pages that were
> ALREADY correct before 721 (the pre-existing 5) plus the one the build path fixed. Assemble mode
> preserved bytes that were already right. **"Currently correct" is a STATE; it is not evidence of
> a transition**, and a column labelled RECOVERED invites reading it as one. **No page has been
> fixed by a re-render. Zero.**

| result | conclusion |
|---|---|
| `background_image` appears **and matches `hero_about`** | the wave is viable → prepare the other 56 as a HELD migration in the 683 shape, hand firing to the site owners |
| **no key, or the site hero instead** | **the hero class and the deck class share one root cause — `425` §2** — and the fix is the rerender path itself, not a wave |

**The second outcome is the more valuable one**, because it unifies two classes worked separately
by three lanes, and it means a 57-item wave would have completed, stamped fresh timestamps and
changed nothing.

⚠ Queue: 192 triaged, 164 older than my 08:26 item, draining ~29 per half-hour. Expect hours.

## 0c. A FLEET FINDING made while measuring this bug — 30 rerenders carrying PROSE as their reason

`[MEASURED 2026-09-02]` completed `page_rerender` items over 24 hours, by reason:

| reason | completed |
|---|---|
| **(none — assemble mode)** | **1,418** |
| `section_data_resolved` | 60 |
| `cta_links_stale` | 39 |
| `template_changed` | 17 |
| **a human-readable SENTENCE** | **30** |

Those 30 were filed by two lanes via three migrations — 701 (16), 696 (11), 693 (3) — with
`spec.reason` set to explanations like *"FCA rule citation corrected by migration 696 (owner
decision 2026-09-02)"*. `check_rerender_mode` tests **equality against five literals**, so a
sentence takes `else_step` exactly as NULL does. All 30 completed and re-shipped stored HTML.

**Both lanes told.** The `bugs_open/357` lane then supplied the important qualification: for
*their* pages the two routes converge **byte-identically by construction** (the adopted template
IS the stored bytes), so their artefacts are correct and only their *mechanism* claim was wrong.
My "has not reached those pages" was too strong for their case.

**Recorded in `LANDMINES.md`** as an append to the existing reason entry, which also needed
correcting: its quoted condition listed **three** values where the live config has **five**
(`template_changed` and `literal_markdown` were missing), and its remedy tests for NULL, which
cannot see a set-but-unrecognised value. The check now tests **membership, not presence**.

### ⚠ AND THE INSIGHT THAT BEARS DIRECTLY ON §2

The 357 lane reached, from the opposite direction, the property that has blocked this bug all
evening: **byte-identical output is not evidence about which path ran.** Their md5-unchanged
result proves only that concatenation reproduces its input; my unchanged array proves only that
*something* left the same bytes. Neither of us can read the artefact alone and say which branch
executed.

That is exactly why the §2 ⭐ experiment matters: it does not try to read harder, it **changes the
baseline** so the two hypotheses stop producing the same bytes.

## 0b. The block that was here (RESOLVED — kept for the recipe)

`kubectl` returns **`You must be logged in to the server (Unauthorized)`** fleet-wide
(checked `get ns` and `get pods -n ai-persona-system`, both fail). That is the 3-day
kubeconfig token expiry; **the owner refreshes it.**

**A fresh chassis build was deployed at ~21:00Z.** `[RELAYED, not verified by me]` the
`site_delivery_and_editor` lane reports chassis pods `8ddbf8958-cd2h9`/`vppjz` started
20:56:43/20:57:10Z on **v1.0.1355**, provenance **`0d2feee2f`** read from the pod by that lane's
previous session, and that `f57f5ad1f` is an ancestor of both v1.0.1354 and v1.0.1355 — so the
§2 A/B already had the producer fix aboard and v1.0.1355 should not change its answer. **I could
not verify any of that**: the token expired ~21:08Z, before I could probe. Treat it as a peer's
reading until item 1 below is run. The first three things to do when the token is back:

1. **Probe EVERY pod, with controls** — not one, because two pods of one ReplicaSet are not
   guaranteed identical (same-tag cached image), and probing one of two left a hole in this
   bug's evidence that could not be closed later because the unprobed pod was gone:
   ```bash
   for pod in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
     for sym in ListItemExcerpt resolvePagesWhereType ListItemTitleXYZNOTREAL; do
       kubectl -n ai-persona-system exec "$pod" -- grep -aqs -- "$sym" /proc/1/exe \
         && echo "$pod $sym PRESENT" || echo "$pod $sym absent"
     done
   done
   ```
   `ListItemExcerpt` must be PRESENT; `resolvePagesWhereType` is the positive control (a
   probe that finds nothing must not read as an answer); the invented symbol must be absent,
   and it *contains* `ListItemTitle`, so its absence also proves the grep is not matching
   loosely. **Never `git merge-base` for this** — a same-tag rebuild serves a cached image
   while merge-base still says yes.
2. **Re-run the §2 discriminator.** The new build may have changed the answer.
3. **Check `721`'s effect** (§3) — it was applied before the roll and needs a re-render to show.

---

## 1. What is DONE and verified

### `bugs_open/425` — card components rendering empty slots (the owner's "cards need better designs")

**Root cause:** two producers write the same standard list-item shape for one shared
component and disagreed. The deck was *already in the row* under `meta_description` while
the template read `.excerpt`; the headline carried the site-name suffix one producer
stripped and the other did not.

| half | state |
|---|---|
| **template** — migration `682` guards four per-item slots + the section header | **DELIVERED on 10 of 14 pages.** Verified at the served markup *and* corroborated by `page_component_history` at the exact rerender minutes. Two independent instruments agree |
| **producer** — `queryresolve.ListItemTitle` / `ListItemExcerpt`, both producers sharing one spelling | **IN THE BINARY, NOT EXECUTING on the rerender path.** See §2 |
| **detectors** — `check_card_slot_guards.py` (new), `check_list_empty_states.py` (widened) | LIVE, advisory, `--self-test` passes |

**4 of the 14 pages were REFUSED** by the section COMPONENT floor (`bugs_open/253`) and
**cancelled** with the reason written into each row. This fix trips that floor *by design* —
every collapsed empty element carries layout classes, so the least-fed sites flatten hardest.
⚠ **NARROWED 2026-09-03: a SCOPED override is possible and my earlier "fleet-wide" was too
strong.** Each agent has its own `save_sections` step config — `[MEASURED]`
`page-build-handler` carries `section_shrink_floor = 0.1` (owner-ruled, migration 725) while
`page-rerender`'s is unset, and they do not touch each other. The standing caution is narrower:
setting it on **`page-rerender`'s** step reaches the fleet's highest-volume pipeline, and even a
per-agent override is per-pipeline-for-the-duration rather than per-page — pair it with a
monitored rollback, as that lane did.

⚠ **Do not set `section_component_floor` casually** as its error invites: it is read from
`params.StepConfig.Config`, i.e. the `page-rerender` **step**, so setting it weakens the guard
for every rerender in the fleet to land one page.

### Migration `721` — six page-scope hero components declare their image field

Applied, ledgered, verified at the artefact, committed `d9e18e2e1`, council `cf3a052c`
**submitted, verdict not yet read**.

Six components render `{{or .hero_url .background_image}}` while declaring no image field, so
their generated page hero was orphaned and the page wore the site-wide homepage image —
passing every "has an image" check while doing it. `hero-tool` 77 instances, `about-hero` 43,
`contact-hero` 25, `services-hero` 6, `case-studies-hero` 5, `use-cases-hero` 2 = **158 live**.

⚠ **The `type` is the fix.** `sectionHasImageField` (`plan_sections_action.go:2936`) gates the
authoritative hero aliasing on `type == "image" || "image_url"`. As `"text"` the gate stays
false, the aliasing never runs, `hero_url` still wins, and the migration applies cleanly and
changes nothing.

**Owed:** `gamedesign.uk` is verifying at its own artefact and will report the `url()`s back.
Nothing has re-rendered yet as far as I know.

---

## 2. THE OPEN DEFECT — the producer fix does not execute on the rerender path

**This is the thing to pick up.** It is well-posed, reproducible, and three diagnosis runs
have failed on it.

### The A/B that establishes it

Same site, same component, same binary, three minutes apart:

| path | result on `content_data.articles[0]` |
|---|---|
| **build** (`needs_page` rebuild, 17:23:02Z) | `excerpt` key **PRESENT**, title suffix **STRIPPED** |
| **rerender** (`page_rerender`, `reason='template_changed'`, 17:26:52Z) | **no** `excerpt` key, suffix **INTACT** |

Reproduced three times (13:59, 17:26, 17:32), all wrote, all old shape.

**Controls, all clean:** the rerender *wrote* (`updated_at` + 8 `page_component_history` rows
at the rerender minute); **0** floor refusals fleet-wide in the window; the slot is bound to
the **canonical** component, not a fork; the markup carries no `article-card__category`, so the
post-682 template ran.

### Why the key's ABSENCE is categorical

The fixed projection writes `"excerpt"` into the item map **unconditionally** — present
whatever the value, including empty. So an absent key cannot be thin data; it can only be code
that did not run, or ran and was discarded. That is the sharpest instrument this bug produced.

### Branches ELIMINATED by reading them — do not re-walk these

`mergedContent = stored ⊕ plan.ResolvedData` with **ResolvedData last**
(`rerender_page_sections_action.go:1709-1716`), so the old shape means the key was **missing
from `ResolvedData`**, not that it lost a merge.

| branch | why it is not this |
|---|---|
| literal-markdown **strip** (~:1600) | gated on `reason == 'literal_markdown'`; dispatch used `template_changed`. Mutates values, never deletes keys |
| **`queryListBelowContract`** → `handleMissingField()` → key never written | the base returns **6** items for this site under its own `listedOnly` floor, and the field declares **no `min_items`** |
| **whole-page escalation** to the writer (STY-048) | **0** `needs_page` items created |
| **`save_page_sections` refusing** | row written *and* archived; 0 floor refusals in the window |
| **stale/cached binary** | killed by the A/B — one binary, two paths, opposite results, minutes apart. **Re-confirmed post-roll:** both pods on v1.0.1355 probed with both controls, and a rerender still produces the old shape |
| **the rerender REASON** (`template_changed` vs `section_data_resolved`) | **[MEASURED 2026-09-02 21:30Z]** a `section_data_resolved` rerender of the same page produced the **same old shape** — no `excerpt`, suffix intact. The reason is not the variable, matching the code: the only reason-gated behaviour inside `rerender_page_sections_action.go` is the literal-markdown strip, and `check_rerender_mode` routes all five recognised reasons to the same step |
| **version pinning** | the pinned version postdates the template update and carries the guard |
| a **third producer** of `articles` | only `rebuild_blog_listing` writes `"articles"` literally; the generic path writes `resolvedData[fieldName]`. Both fixed |

### ⚠ THE RERENDER PATH DELETE/re-INSERTs THE ROW — three consequences for how you verify

`[MEASURED 2026-09-02]` two rerenders of boxingonline `/index.html` produced two **new
`page_components.id`s**: `7dead3e5` at 21:22:26Z, `9e643633` at 21:30:07Z. The path replaces the
row; it does not `UPDATE` it.

1. **Never key a before/after on the instance id.** It differs on every rerender whether or not
   anything changed. Key on `page_id` + `slot_name`.
2. **A fresh instance id is NOT evidence of new content.** It is evidence the row was replaced,
   which happens on the runs that reproduce the old shape exactly.
3. **This is why `page_component_history` is reliable here** — the rows come from the **DELETE**
   trigger, not an update trigger — **and why `page_id` is the only sane join**: it survives the
   replacement, and `component_id` is NULL on 44,555 of 45,285 rows.

### STATUS 2026-09-03 ~08:35Z — the experiment is FILED AND QUEUED, not run

Item batch `…000688`, `reason='template_changed'`, on **`guides-index`** (page `2e738efd`) — the
lane's own new-shape baseline, built 17:23:02Z, `articles[0]` carrying `excerpt`, suffix-free,
untouched since. Filed 08:26:24Z, **still `triaged`**.

`[MEASURED 2026-09-03 08:35Z]` it is behind a backlog: **192 `page_rerender` triaged, 164 of them
older than mine**, draining ~29 per 30 minutes. So it is roughly **three hours out** on FIFO. It
is queued, not stuck.

> ⚠ **A TIMED-OUT WATCHER PRINTS THE CURRENT STATE, WHICH READS EXACTLY LIKE A RESULT.** My first
> watcher expired after 8 minutes and printed
> `e6b51597 | t | f | 4 | 17:22:27` under the heading "THE DISCRIMINATOR RESULT". That is the
> **unchanged baseline** — the item had not run. Nothing in the output says so. **Always print the
> terminal status alongside the reading, and treat an unchanged `updated_at` as "did not run"
> rather than "no change".** Here `17:22:27` is the give-away: it predates the dispatch by fifteen
> hours.

**TWO EXPERIMENTS ARE IN FLIGHT ON THE SAME PAGE — both queued, neither read.** Do not re-file
either; find them by these ids:

| what | id | filed | reads out as |
|---|---|---|---|
| **mine** — `template_changed` rerender on the new-shape baseline | batch `00000000-0000-0000-0000-000000000688` | 08:26:24Z | `excerpt` **gone** → the rerender path rebuilds with unfixed code · **survives** → it never touches `articles` |
| **the delivery lane's** — `needs_rerender` / stale_chrome, assemble-only | `ec92320f-3037-448a-bd55-de8385404d92` | 08:35:44Z | tests whether assemble mode really is a byte-for-byte re-ship |

⚠ **The chrome wave's per-page items are CHILDREN.** `rerender-pages` creates them once
`ec92320f` is claimed, so their `source_item_id` chain hangs off that child batch, **not** off
`ec92320f` directly. Find the children by `page_id` + creation time, not by parent id.

⚠ **THE DELIVERY LANE'S PASS CONDITION NEEDS A POSITIVE CONTROL AND I HAVE SENT THEM ONE.**
"No new history row for guides-index" cannot distinguish *assemble preserved the array* from *the
wave never reached this page* — both give zero rows. The control is to confirm a per-page
`_assemble` item for `2e738efd` exists and reached terminal BEFORE reading history as a pass.
Same family as the three absences that cost me yesterday (a 20-second log window, a 98%-NULL
filter column, a timed-out watcher).

⚠ **The two branches want OPPOSITE readings of the same table** — if assemble does not replace
rows, the ABSENCE of a history row is the pass; if it does replace them, the PRESENCE of an
`artefact_archive_trigger` row whose archived `content_data` still carries `excerpt` is the pass.
Decide which branch you are in before reading, not after.

**Attribution does NOT depend on running this before or after anything else** — which is why the
delivery lane's chrome wave was released rather than held for three hours.
`page_component_history` carries `source_item_id` on **1,169 of 1,196** `save_page_sections_overwrite`
rows in 24h, so a sections-path write names the item that caused it. Read the chain, do not police
the order:

```sql
SELECT h.created_at, h.source, h.source_item_id,
       (h.content_data->'articles'->0 ? 'excerpt') AS had_excerpt_BEFORE_this_write
  FROM page_component_history h JOIN pages p ON p.id = h.page_id
 WHERE p.id = '2e738efd-...' AND h.created_at > now() - interval '12 hours'
 ORDER BY h.created_at;
```

That table archives **the state being REPLACED**, so each row says what the page looked like
immediately *before* that write — a per-write before/after chain, already attributed.

⚠ Assemble-path writes may arrive as `artefact_archive_trigger` rows with **no** `source_item_id`.
For the delivery lane's own prediction that is cleaner, not worse: a byte-identical re-ship trips
**neither** trigger (both test `IS DISTINCT FROM`), so **an absent row is their pass and a present
row is their finding**.

### ⭐ THE EXPERIMENT TO RUN FIRST — it breaks the ambiguity that blocked this all evening

Proposed by the `site_delivery_and_editor` lane as a question ("would a rebuild of boxingonline
`/index.html` destroy a repro you still need?"). **It would not — it makes the repro strictly
better**, and it is now the highest-value next action.

**The ambiguity, stated plainly:** today the stored array is the OLD shape, and a rerender
produces the OLD shape. Those two are *byte-identical*, so I cannot tell whether the rerender
**overwrote stored with an old-shape rebuild** or **never touched the key and stored simply
survived**. Every reading this evening ran into that wall.

**A build-path rebuild removes the wall by changing the baseline:**

1. Fire a `needs_page` rebuild of boxingonline `/index.html` (the build path is the known-working
   route — `guides-index` did exactly this at 17:23:02Z and produced `excerpt` present, suffix
   stripped). Confirm the new shape landed: `content_data->'articles'->0 ? 'excerpt'` → true.
2. Then fire a `page_rerender`, `reason='template_changed'`, `spec.page_name='index'`.
3. **Read the key again. The two outcomes now say opposite things:**

| after the rerender | conclusion |
|---|---|
| `excerpt` key **GONE** (reverted to old shape) | the rerender path **actively rebuilds** the array with code that is not the fixed projection — it overwrites good data with old-shape data |
| `excerpt` key **SURVIVES** | the rerender path **never touches** `articles`; `ResolvedData` lacks the key and stored wins by default |

Either answer halves the remaining search space, and neither is available while stored and
produced are identical.

⚠ **Confirm at `page_component_history` keyed on `page_id`** (NOT `component_id`, which is NULL
on 44,555 of 45,285 rows) that the rerender actually wrote — and take a same-window positive
control from another component, or an absence proves nothing.

⚠ **Cost of being wrong:** none to the repro. The defect reproduced on demand three times by
re-filing a rerender, so the old state is one dispatch away whichever result comes back.

### What has NOT been established

- Whether `planSection` is reached for this field on this path at all. `plan_sections_action.go:2695-2709`
  resolves any `query.*` source **unconditionally** — no "only if missing" gate — so the read
  says it should be. The behaviour disagrees. **The behaviour is what ships.**
- An instrumented run (live tails on both pods during a rerender) captured **no**
  `queryresolve:` or `plan_sections:` message — **recorded as INCONCLUSIVE, not as evidence.**
  The pods serve every agent so surrounding lines cannot be attributed without correlation
  filtering, and there is no positive control that the resolver's line would appear if it ran.

### Diagnosis runs — all three failed

`afbf8544` capped UNVERIFIABLE · `fe4b8537` NOT CONFIRMED (0-row data requests; one behind an
`EXPLAIN` erroring `syntax error at or near $`, a parameterised query handed to `EXPLAIN`
unbound — a concrete lead) · `c755b0be` UNVERIFIABLE **even with the four eliminations stated**.
Three failures on a well-posed symptom is a signal about the loop's reach here, not only about
phrasing. **A fourth run in the same shape is probably not the move.**

---

## 3. Council state

| correlation | subject | state |
|---|---|---|
| `84b51f16-086f-493d-b188-49231d0ca907` | 425 (Go + 682 + 683) | **REVISE at round 3.** Two real objections ACTIONED (`c1178442d`, `9f6f91325`): reuse `datahelpers.SafeCut`/`TruncateString` instead of hand-rolled rune slicing; 683's header no longer tells a human to verify a deploy with `git merge-base`. **Round 4 NOT submitted.** |
| `cf3a052c-37e3-4573-ad6c-fdddd83faeda` | 721 (hero image fields) | **Submitted, verdict unread.** |

Round 3's remaining objections were largely **sketch artefacts** — verified: the 682 ROLLBACK
carries the real pre-image (1 occurrence of the markup, 0 of the placeholder), and 682 does wrap
UPDATE+verify in `BEGIN`/`COMMIT` with a drift guard before the UPDATE. If you resubmit, fix the
**sketches**, not the files.

---

## 4. Peers, and who is waiting on what

- **`site_delivery_and_editor`** — owns the boxingonline pipeline. Has the two-probe acceptance
  test (class-count for "did the template render", key-presence for "did the producer run").
  **Waiting on a ping when the producer half is explained.** Verified the 682 delivery at the
  served markup independently.
- **`boxingonline.com`** — raised 425. Holds the before-capture. Contributed the correction that
  **a correct rerender `reason` is necessary and NOT sufficient** (rows completed twice with the
  right reason and produced the old shape).
- **`gamedesign.uk`** — routed the hero defect, handed me the migration, verifying `721` at its
  artefact and will report `url()`s back.
- **`designblog.co.uk`** — routed the owner's "every site looks the same" directive; has my
  census. Routed the counter-instance that stopped me building a two-component version of 721.
- **`theme kits`** — building `page_archetypes`. I gave them a key-collision finding
  (`content_components.function` is not unique after `forked_from IS NULL`: `site-header`,
  `site-footer`, `tool-agent-complexity-estimator` still collide, resolved by `ORDER BY name`).

### The owner's sameness directive — measured, ACKed, NOT actioned

`hero` on 37 sites (634 instances); **74–87% of every site's slots come from the same 10
components**; but **156 shared section components exist, 40 with zero live instances and 61
used on one site only.** The library is not the bottleneck — **selection is**. designblog serves
**0** content images across 50 slots (controls: garden-tools 4/43, boxingonline 5/48, so the
query detects images).

⚠ I told them `page_archetypes` is necessary and probably **not sufficient**, because
`defaultSectionsForPage` is only the **fallback** when the planner returns no sections. **The
cheap measurement nobody has done: what fraction of live pages got their sections from the
planner versus the fallback.** That decides whether the lever is the table or the planner prompt.

---

## 5. Practice notes earned today — the expensive ones

- **Every instrument that survived reads the ARTEFACT; every instrument that fell reads a column,
  or a table about a column.** Six overturned claims, one pattern.
- **A positive control must exercise the same row population as the claim.** I proved
  `page_component_history` works using 389 rows from *other* components, concluded it was blind
  for mine, and told a peer to stop using it. My query was filtering `component_id`, which is
  **NULL on 44,555 of 45,285 rows**. Join on `page_id`.
- **A sound methodological rule reasoning from a false premise produces a confident, quotable,
  wrong conclusion** — and it travelled to a peer's notes inside the hour *because* it was
  well-phrased. Well-phrased + confirming is the accelerant pair.
- **A census offered as evidence about what CODE returns must replicate that code's predicate.**
  I published "the resolver would return 5" from a query missing `ListedPageEligibilitySQL`; a
  peer built a discriminating test on it that would have burned a round proving nothing.
- **Grep before filing.** Three would-be new landmines this session were already recorded
  (`pages.status`'s dead `'deployed'` literal — three entries; the roll-kills-council trap, with
  a *better* check than the one I had; the per-item slot class).
- **The chassis log's readable window is ~SECONDS**, not hours: `--since=3h` returned 318 lines
  spanning 20 seconds. Tail while inducing; never read back.

---

## 6. Files

`bugs_open/425_HANDOFF_2026-09-02_a_card_renders_four_empty_slots_because_two_producers_of_one_item_shape_disagree.md`
— the full case, with **every superseded reading left visible**. The corrections are the record.

Migrations: `682` (+ROLLBACK) applied · `683_..._HOLD` (+ROLLBACK) applied, batch
`…000683`, 10 complete / 4 cancelled · `721` (+ROLLBACK) applied.
Scripts: `scripts/check_card_slot_guards.py`, `scripts/check_list_empty_states.py`,
`scripts/component_template_lib.py`.
Register: **PBP-050**. Landmine + `WRONG_CALLS` entries committed.
