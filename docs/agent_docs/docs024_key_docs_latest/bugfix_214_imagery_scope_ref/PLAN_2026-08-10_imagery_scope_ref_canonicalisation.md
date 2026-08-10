# PLAN 2026-08-10 — imagery `scope_ref` is minted raw while its own plan canonicalises page identity

**Bug:** `bugs_open/214_HANDOFF_2026-08-07_imagery_scope_refs_are_llm_minted_and_never_validated_so_orphans_ship_silently.md`
**Lane opened:** 2026-08-10, this session. The filing lane (`bugs_open/151`) explicitly
left it — *"same wire, different field, its fix does not gate"* — so it was unowned.
Ownership checked three ways before starting: `who-owns.py`, `git log` on the file, and a
grep of all 39 live session transcripts for the slug (0 hits outside `ls` output).

---

## 1. What is actually wrong (and how it differs from the filing)

`WriteSitePlanAction` (`platform/orchestration/actions/write_site_plan_action.go`)
canonicalises page identity for two of the three tables it writes, and not for the third —
**inside one function, about sixty lines apart**:

| written at | column | canonicalised? |
|---|---|---|
| `:392` → `:503` | `site_plan_pages.name` | **yes** — `datahelpers.CanonicalisePage` |
| `:392` → `:533` | `site_plan_sections.page_name` | **yes** — same canonical `r.Name` |
| `:455` → `:566` | `site_plan_imagery.scope_ref` | **no** — the planner LLM's raw map key, verbatim |

`CanonicalisePage` collapses spellings onto one identity — the section-index family
(`page_canonical.go:159-173`) turns `about` → `about-index`, `contact` → `contact-index`,
`news` → `news-index`. So the planner emits `imagery.pages["about"]`, the plan writes its
page as `about-index`, and the imagery row now names a page **its own plan does not
contain**. Nothing rejects it: `buildImageryRow` (`:1044`) never inspects `scopeRef`.

**This is a platform defect, not an LLM one.** The filing reads flavour 2 as *"the LLM
keyed a page-name variant"*. It did not — it keyed the name it was given, and the platform
renamed the page underneath it without carrying the reference along.

## 2. Corrections to the bug file, established before designing

> **CORRECTION 1 — the census measures the harmless half and cannot see the harmful half.**
> 214's census filters on **ordinal range** within `site_plan_sections`, and reports
> `orphaned_ordinal = 5`. But **no consumer parses the ordinal**: the section join is
> `spi.scope_ref LIKE $2 || ':%'` (`plan_sections_action.go:377`), keyed by asset key and
> aliased by kind, and `flag_page_image_rebuild_action.go:120` splits on the first colon and
> discards the ordinal. So an out-of-range ordinal on a correct page part is **behaviourally
> inert today**. What is fatal is a wrong **page part** — and the census is blind to it,
> because it never checks that the page resolves, and it excludes `scope='page'` entirely.

> **CORRECTION 2 — the lock-transfer key in §"Who consumes scope_ref" is wrong.**
> 214 states imagery locks carry forward on `(scope, scope_ref, category, subject, ordering)`.
> That key belongs to `transferDirectiveLocks` on `site_plan_directives` (`:786`). Imagery is
> `transferImageryLocks` (`:1123`) on **`(plan_id, scope, scope_ref, key)`**. A fix or test
> built from the file's stated key would aim at the wrong function. Consequence for the
> design: scope_ref **is** in the imagery key, so rewriting it breaks lock carry-forward for
> one plan generation unless that is handled — §4.3.

> **CORRECTION 3 — my own first measurement was wrong, and in the flattering direction.**
> I first measured resolvability against `site_plan_pages.name` and got **22** unresolvable
> rows. Twelve of those resolve perfectly well: every consumer matches against **`pages.name`**
> (the deployed page table), not the plan table. The honest figure is **10**. Logged in
> `WRONG_CALLS.md` — the check that would have caught it is "read the consumer's WHERE clause
> before choosing the table you measure against".

## 3. The live damage, measured 2026-08-10 (current plans only)

176 page+section-scope imagery rows; **10 invisible to every consumer** — their scope_ref
page part matches no `pages.name` row on that site:

| domain | scope | scope_ref | key | plan page candidate | asset already paid for |
|---|---|---|---|---|---|
| fundamentallyai.com | page | `news` | hero_news | `news-index` | yes |
| gamesdesign.co.uk | page | `about` | hero_about | `about-index` | yes |
| gamesdesign.co.uk | section | `about:2` ×4 | icon_* | `about-index` | yes |
| gamesdesign.co.uk | page | `contact` | hero_contact | `contact-index` | yes |
| mortgagecalculator.co.uk | page | `about` | hero_about | `about-index` | no |
| mortgagecalculator.co.uk | page | `contact` | hero_contact | `contact-index` | no |
| mortgagecalculator.co.uk | page | `tools-index` | hero_tools | — (none) | no |

Nine of ten have an exact `<ref>-index` candidate in their own plan. **Eight of ten have an
`assets` row at `status='active'`** — generated, deployed, paid for, and referenced by
nothing. That is `bugs_open/114`'s symptom reached through this cause, which is what makes
this worth fixing at the write path rather than per site.

The gamesdesign four are the sharpest case: `about:2` → `about-index:2`, and `about-index`
has exactly three sections (ordinals 0–2), so **the ordinal was right all along**. Only the
page name was renamed out from under it.

## 4. The fix

**The guarantee:** *an imagery `scope_ref` written by this action either names a page the
plan itself contains, or leaves a durable record saying it does not.* Nothing is ever
silently dropped.

### 4.1 Resolve the page part through the plan's own raw→canonical map

`planRows` is fully built, canonicalised **and deduped** before `flattenImageryBlock` is
called, and `planPageRow` already carries both spellings (`RawName`, `Name`). So the map is
free — no second canonicaliser, and none is possible anyway: **`CanonicalisePage` requires a
`Role`, and imagery refs carry none**, so re-deriving from the ref alone cannot work.

Precedent followed: `buildPageFeatureMap` (`apply_adoption_plan_action.go:776-834`), same
package, same class of bug, same raw-fallback rule — upgraded with the durable log it lacks.

- page scope: `scope_ref` → canonical name.
- section scope: split on the **first** colon (the split every consumer uses), rewrite the
  page part, reassemble with the ordinal byte-identical. The colon is preserved by
  construction, so `chk_scope_ref_consistency` holds without a special case.
- **unresolvable → keep verbatim.** Identical to today's behaviour, so no row can regress;
  plus a durable `IMAGERY_SCOPE_REF_UNRESOLVED` row.
- an alias that would map to two different pages is dropped from the map, not guessed.

### 4.2 The ordinal: validate and record, never rewrite

Justified from the consumer census, not from the filing's pre-census candidate list:

1. no consumer parses it, so a rewrite buys no behaviour;
2. a rewrite risks 23505 on `idx_site_plan_imagery_unique` and breaks lock carry-forward for
   a second generation;
3. the correct target is **unknowable at this seam** — ordinal shift happens in
   `ValidateSitePlanAction`, a different action, and the pre-drop array is gone by here.

So out-of-range/malformed ordinals ship unchanged with an `IMAGERY_SCOPE_REF_ORDINAL_ANOMALY`
row. That converts the filing's flavours 1 and 3 from silent to observable, which is the
honest ceiling for this seam. The real fix for ordinal semantics is 214's candidate 2
(imagery inside the section entry, RFC_016 §1) — **architecture-scope, flagged, not taken**.

### 4.3 Two guards the filing did not know it needed

- **`dedupeImageryRows`** — the imagery analogue of `dedupePlanPageRows`. Collapsing two raw
  spellings onto one canonical name can produce a duplicate `(plan_id, scope, scope_ref, key)`,
  which aborts **the whole plan write** with 23505. That is `bugs_open/215` verbatim on the
  neighbouring table. Zero collisions exist live today, so this is a safety guard, not a live
  fix — built anyway, because the fix is what creates the possibility.
- **`transferImageryLocks` canonical fallback** — on the first post-fix replan the old plan's
  locked row carries `about:2` while the new row carries `about-index:2`, so the exact-match
  UPDATE hits 0 rows and the lock is silently lost. Fallback runs **only** after the exact
  match finds nothing, so it cannot regress a lock that works today, and it self-retires from
  the second replan onward.

## 5. Blast radius, measured before submission rather than asked of reviewers

- **`needs_imagery` item_key** embeds scope_ref (`imageryplan.ItemKey`). Open items on
  affected refs: **3**. Left untouched — the asset is stored under `asset_key`, which the
  rewrite never touches, so a landed asset is immediately reachable through the repaired row.
- **Unique-index collisions on current plans: 0** (measured, all sites).
- **`imageryplan.Classify`/`BrandUpdate`** compare `*scopeRef == "index"`. Strictly improves:
  a planner ref `home` now canonicalises to `index`, so a homepage hero previously
  misclassified lands at the brand-update priority.
- **Rows that work today are untouched by construction** — the map's identity pass sends an
  already-canonical ref to itself, which takes the no-op branch.

## 6. Verification (disconfirmable — what a failure would look like)

1. **Unwired detector.** `TestWriteSitePlan_ImageryRefCanonicalisationIsWired` runs the whole
   action under sqlmock and pins the `INSERT INTO site_plan_imagery` scope_ref bind to
   `about-index`. Revert only the wiring block and it **must fail**. If it still passes, the
   fix is decorative. (This lane's own WRONG_CALLS entry from 08-10 is exactly this failure
   mode on another lane: twelve passing tests around a fix that could be unwired without one
   turning red.)
2. **Backfill.** Invisible-row census must go **10 → exactly 1**. Disconfirmed by 10 (inert)
   *or* by 0 (overreach — it touched the row it was told to leave).
3. **Write path, live, after the next replan.** Any row whose page part is a raw alias of a
   page that same plan contains ⇒ the rewrite is not running. Any unresolved ref with **no**
   same-day `agent_error_log` row ⇒ the durable-record half is not running.

## 7. Route

Council gate, not an RFC. Under the owner ruling of 2026-07-29 §1, an RFC is owed when a
change alters what a shared mechanism **guarantees**. This does change a guarantee — so the
seam is registered in the same commit (condition 2 of the ordering exemption, which is now
the whole of the requirement), the other consumers are **named and told** rather than merely
measured (ruling §3), and it goes to the council before/alongside the commit.
