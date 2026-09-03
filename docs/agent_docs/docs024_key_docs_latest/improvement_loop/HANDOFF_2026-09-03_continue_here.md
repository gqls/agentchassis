# HANDOFF 2026-09-03 — improvement loop, continue here

**COLD-START: this file, then `SUMMARY_2026-09-02_improvement_loop_ownership.md`.**
Supersedes `HANDOFF_2026-09-02_continue_here.md` (kept; its §1 traps are all still true).
Depth: `PLAN` §5 for the ordered work · `NOTES` for the missteps — §(w), §(z), §(gg), §(ii),
§(oo) are the five worth your time · `RUNBOOK` for commands with their traps attached ·
`README_where_we_are` for the owner's plain-prose log.

---

## 1. The four things that will mislead you in the first ten minutes

Unchanged from 09-02 and all still live:

1. **The written record says the loop is switched off. It is not.** The 2026-07-29 ruling
   ("stopped DELIBERATELY … do not re-enable it") is superseded by migration `389`.
   `[MEASURED 2026-09-03 12:29Z]` `improvement-sweep` enabled, last triggered 12:29:54.
   **Read `scheduled_tasks`, never the ruling.**
2. **`complete_clean` does not mean the site is clean** — it is also the terminus for a
   skipped audit (mig 291's fingerprint gate) and for a site whose whole pile was held
   unroutable. `collected_data->'audit_state'` separates them.
3. **`execution_path` is empty on every improvement-loop row.** Use
   `jsonb_object_keys(collected_data)`.
4. **Never sum `not_promotable` across runs** — it is a per-run count of a standing pile.
   It gave me 3,866 for a backlog of 1,385.

## 2. What shipped, and what it actually did

### 2a. The skip link — LIVE, and DRAINING

`[MEASURED 2026-09-03 12:3xZ]`

| | 2026-09-02 midday | now |
|---|---|---|
| `head_essentials_missing` open | 978 | **704** |
| all flag-only `detected` | 1,385 | **1,162** |
| retracted since the roll (20:56:43Z) | — | **265** |

**The attribution is checked, not assumed** — and I got this wrong once already (NOTES §(oo):
10 retractions on 09-02 looked like mine and were another lane's, at 19:05, an hour and 51
minutes *before* the roll). These 265 are post-roll, span 8+ domains (finetuning 52,
gamesdesign 48, dartsonline 42, relojistas 31, fundamentallyai 26, lendzy 17, vetcomparison
15, garden-tools 14) and every reason reads *"re-probed …: title, skip-link and footer all
present"*. **Anyone quoting a lower number later must check `completed_at` against
2026-09-02 20:56:43Z before crediting it here.**

Survived your roll to `v1.0.1356`: re-probed the new binary for `data-skip-link` — present,
with the present-control (`GROWTH_DOOR_PROBE_FAILED`) hitting and an invented literal missing.
**Do not inherit that probe** — a same-tag rebuild serves a cached image, so re-probe after
any roll before reading the numbers as evidence.

**Expect the drain to plateau at ~10, not 0.** `[MEASURED 2026-09-02]` 968 of 978 rows carry
`spec.assembled=true`; the residual is 10 named pages on 5 sites with no `page_components`
rows (NOTES §(gg)). A plateau near 10 is the floor, not a stall.

### 2b. Migration 722 — new sites born holding growth. APPLIED and APPROVED

`722_new_sites_are_born_holding_growth.sql`, applied 2026-09-03, council
`070347dd-c410-4cf2-b5e6-8c87e568a792` **APPROVED at round 5**. Verified at the artefact:
`pg_trigger` shows `trg_sites_born_holding_growth` enabled; no existing row moved.

It is a **BEFORE INSERT trigger, not a column default** — rounds 1-3 were a default and the
council killed it. Read the migration header for why; the short version is that a default is
bypassed by an INSERT naming `settings`, and 2 of 15 creation paths do.

**OWNER RULINGS, both recorded in the migration header, WDS-020 and §2c below:**
- 2026-09-02: a brand-new site is born `hold`, released by a human.
- 2026-09-03: **adopted sites are held too, until specifically released.** Intended scope.
  **Do not exempt adoption to "fix" a held adopted site.**

### 2c. ⚠ TWO THINGS THAT WILL BITE THE NEXT PERSON

**(i) If you re-run 722, ARM 4 WILL NOW ABORT — and that is the legitimate case, not a bug.**
Arm 4 raises if more than 1 site carries a posture. `[MEASURED 2026-09-03]` **2 do**:
gamedesign.uk (hand-held 09-02 by its own lane) and **apis.uk** (created 2026-08-22, so NOT
born held — hand-held by somebody, not by this trigger). The migration's own comment above
that arm says exactly this: a leak is impossible (a column default and an INSERT-only trigger
cannot touch existing rows), so a fire almost certainly means another lane held a site
deliberately. **Re-read the rows, raise the threshold in a commit that says whose they are,
and never delete a posture to make it pass.**

**(ii) `BEFORE INSERT` does NOT protect existing rows under an UPSERT.** `EXCLUDED` is the
POST-trigger row, so a future `DO UPDATE SET settings = EXCLUDED.settings` on `sites` re-holds
a released site — induced both ways 2026-09-03. None of the four current upserts does it.
Defended in code, not prose: `check_sites_upsert_excluded_settings` (pattern-check.py) refuses
one at commit time in every session; `LANDMINES.md` carries the induction recipe.

## 3. NOT YET PROVEN, and it is the honest gap in 722

**No site has been created since 722 applied, so the trigger has never fired on real traffic.**
Its six verify arms all pass (including the two that catch a wrong-nesting default and an
UPDATE-firing trigger), but every one ran in a rolled-back transaction. **The first real site
creation is the demand test.** When one happens, check it directly:

```sql
SELECT domain, created_at, settings->'maintenance_profile'->>'growth_posture'
  FROM sites ORDER BY created_at DESC LIMIT 3;   -- a new row must read 'hold'
```

Related and also owed: the `gamedesign.uk` lane holds the only site whose *held work items*
can be read. Their next improvement-loop run over that site is the only live evidence the
HOLD ITSELF behaves as designed (items born `deferred`, handler-less, `[growth held]` in the
summary, release recipe on the row). **Ask them for it** — I have, twice.

## 4. Open, in the order I would take them

1. **The "held longer than N days" report. MINE, UNBUILT, and the residual I introduced.**
   A site nobody releases stops growing and nothing errors — the same silent shape as the
   1,385 findings this lane opened on. Named in 722's risks and by the council's guardian seat.
2. **Watch the skip-link drain to its ~10 floor**, then re-census the flag-only pile. The
   remaining ~390 rows of other types are plan item 4's real input and I have never examined
   them.
3. **Plan item 4, reframed — and the reframing is the point.** I assumed nobody had noticed
   the flag-only pile was undrainable. `check_archived_page_still_serving.go:104` quotes
   `bugs_open/083` **at itself** — *"a detector whose output nobody drains is not neutral —
   it is actively misleading"* — names the exact door that makes its own finding
   undispatchable, and ships anyway. Eleven checks made that trade independently. Several say
   *"THIS PASS"* / *"no handler agent in v1"* — **deferred handlers, not human judgements.**
   So the question is not "where do we display 1,385 findings" but which are waiting on a
   handler nobody built and which genuinely need a person; and for the second group, the brief
   that exactly one check writes (`[MEASURED 2026-09-02]` **9 of 1,385** rows carry a
   `triage_hint`). Shared-seam change across eleven producers → its own council round.
4. **`RFC_061`** (head injection has two entry points) — filed at the council's request,
   CANDIDATE, unowned. Its §3 states where my evidence stops: I have **not** established that
   `AssemblePageAction` serves any live page. **That census is task one for whoever takes it**;
   if the second path is dead, the file closes and five documented gaps close with it.

## 5. Owner's, not mine

- **Point `boxingonline.com`** to `alexis.ns.cloudflare.com` / `leah.ns.cloudflare.com`
  (21 pages built and deployed; pointing it makes the site appear).
- **Do NOT point `adversecreditmortgage.co.uk`** — zero pages ever deployed, and its build
  queue is stalled (19 `needs_page`, 22 link items, 0 attempts). Pointing it shows a blank,
  which reads as the pointing having failed. Detail: `POINTING_2026-09-02_domains_to_repoint.md`.

## 6. What this lane learned that generalises

Five council rounds on one small migration, each finding something the change *working* would
never have shown. Worth reading once, in `NOTES`:

- a verify block that was a substring match — passing on exactly the defect class it existed
  to catch;
- a coverage claim ("no creation path names `settings`") drawn from two citations and **false**,
  which invalidated the whole design;
- "`BEFORE INSERT`, therefore existing rows are safe" — **false under an upsert**;
- **my own verify arm writing to a production site**, which `improvement-sweep`'s
  `ORDER BY updated_at ASC` would have turned into "a live site sent to the back of my own
  lane's rotation";
- and, from the skip-link work, a **test suite that passed with the fix deleted** — found by
  mutation, in the exact risk I had already written three paragraphs about.

The pattern across all five: **I wrote the hazard down correctly and did not assert on it.**
Prose about a risk is not a control on it.
