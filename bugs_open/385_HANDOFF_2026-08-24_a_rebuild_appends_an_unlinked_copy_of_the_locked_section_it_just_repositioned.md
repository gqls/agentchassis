# 385 — a rebuild can append an UNLINKED copy of the locked section it just repositioned, and the page then renders that tool twice and can never be re-rendered again

**Filed 2026-08-24**, `loancalculator_couk` lane, found by the acceptance harness on its
first working run after being repaired (`0aafce405`). **The live damage is REMEDIATED**
(see §7); ~~the **root cause is OPEN**~~ **ROOT CAUSE ESTABLISHED 2026-08-25 — see §5c.**

> **CORRECTED 2026-08-25: the cause is now established, first-hand, at every joint —
> §5c.** The writer is the **Layer 2 interactive-tool preservation block** in
> `save_page_sections_action.go` (the "re-append dropped interactive tool" arm,
> ~line 585), whose matcher pairs by exact slot-name string only. §5b's suspicion
> ("something between the plan and the save added a sixth entry") was right; its named
> candidate (LOCK-008's merge) was **wrong** — the merge pairs correctly and inserted
> nothing (refuted in §5c). Per the owner ruling of 2026-07-31 the `090` substitution is
> declared in §6; §5c's chain carries its evidence inline, each claim with the query or
> file:line that grounds it.

## 1. The symptom, at the artefact

`https://loancalculator.co.uk/tools/loan-vs-savings.html` renders its calculator
**twice**. Every id-bearing element of the tool appears twice — `loan-rate`, `save-rate`,
`spare-cash`, `tax-bracket`, `results`, `loan-panel`, `save-panel`, `loan-benefit`,
`save-benefit` — as do `function compare` and `function copy`.

**The lower copy is dead.** The script resolves its outputs with `getElementById`, which
returns the FIRST match, so both copies' answers are written into the upper one. A visitor
typing into the lower calculator sees nothing happen.

`[MEASURED 2026-08-24]` a census of duplicate `id="…"` attributes over **all 28** served
pages of this site found duplicates on **exactly this one**.

## 2. The rows

```
pos slot_name     component_id  locked  html md5    content_data md5  created
 2  tool-2        448422ce…     t       be85284e…   f65a0b6e…         08-02 22:51
 6  tool-2        NULL          f       be85284e…   f65a0b6e…         08-23 14:14
```

**Byte-identical, same slot name, one locked and one orphaned.** Written by the
owner-released tool-page rebuild that completed **2026-08-23 14:15:19** — the same wave
that got the other nine pages right.

## 3. The second, worse consequence: the page is STUCK

The orphan row has no component to resolve, so the re-render path refuses the whole page.
This is already in the queue, failed 3/3, and it names the row exactly:

```
step rerender_sections failed: … page "tool-loan-vs-savings": 1 of 6 section(s) could not
resolve a component and were carried unrendered instead — unresolved component
[tool-2 (pos 6)]
```

So the defect is self-perpetuating: **the page cannot be repaired by the framework's own
re-render** while the row is there, and any queued fix aimed at it fails on arrival.

## 4. Blast radius — dated, because a census goes stale by ADDITION

`[MEASURED 2026-08-24]` `page_components` rows with `component_id IS NULL` on **active**
pages: **11 rows across 6 domains**, of which **two were created on 2026-08-23** —
`gamesdesign.co.uk/games/jelly-invaders/index.html` (slot `section`) and this one. Not
loancalculator-only, and not historic. Re-run before quoting:

```sql
SELECT s.domain, p.url, pc.position, pc.slot_name, length(pc.rendered_html),
       to_char(pc.created_at,'MM-DD HH24:MI')
FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
WHERE pc.component_id IS NULL AND p.status='active' ORDER BY pc.created_at DESC;
```

> **NARROWED the same day, and this is the more useful number.** I first wrote that the
> other ten were an unmeasured candidate population. They are not: characterised
> `[MEASURED 2026-08-24]`, **none of the ten is a byte-twin of any other row on its page,
> and none has a locked sibling in the same slot** — so `component_id IS NULL` is a column
> value several unrelated shapes share, and **this duplication is a population of ONE
> observed instance.** The discriminating query, which is the one to re-run rather than the
> bare `IS NULL` count above:
>
> ```sql
> WITH orphans AS (
>   SELECT pc.id, pc.page_id, pc.slot_name, md5(pc.rendered_html) AS h, s.domain, p.url
>   FROM page_components pc JOIN pages p ON p.id=pc.page_id JOIN sites s ON s.id=p.site_id
>   WHERE pc.component_id IS NULL AND p.status='active')
> SELECT o.domain, o.url, o.slot_name,
>        (SELECT count(*) FROM page_components t WHERE t.page_id=o.page_id AND t.id<>o.id
>           AND md5(t.rendered_html)=o.h) AS byte_twins_on_page,
>        (SELECT count(*) FROM page_components t WHERE t.page_id=o.page_id AND t.id<>o.id
>           AND t.slot_name=o.slot_name AND t.locked_at IS NOT NULL) AS locked_same_slot
> FROM orphans o ORDER BY 4 DESC;
> ```
>
> **A non-zero `byte_twins_on_page` is this bug. A bare `component_id IS NULL` count is not**
> — it over-reports by 10 out of 11, which is the difference between "a fleet-wide class" and
> "one page", and I had written the first before running the second.

> **⚠ FORWARD POINTER, added 2026-09-04 by the `bugfix_450_tool_page_shells` lane —
> `bugs_open/479`. The narrowing above is correct and it has a cost, so read it with this
> attached.** It is true of THIS bug (the duplication), and the bare `component_id IS NULL` count
> it discourages is the right census for a DIFFERENT bug living in the same arm: the re-append
> drops the stored row's `component_id` whether or not it also duplicates anything. `[MEASURED
> 2026-09-04]` that count is now **17 rows across 7 sites** (11 here on 08-24), of which **5 are
> tool slots serving working tools** — and **all 17 have `byte_twins_on_page = 0` and
> `locked_same_slot = 0`**, so this file's own discriminator excludes every one of them.
> **Three were created AFTER §7b's fix went live on 2026-08-26**, so the identity drop is a live
> producer even though the matcher is fixed. §5c joint 5 already names the line
> (`ComponentID:'' because the RFC_046 carryStoredIdentity opt-in is OFF`) — it was right, and it
> was read as incidental. Fix committed 2026-09-04 (`reappendedComponentID`); the 17 existing rows
> are NOT repaired, see 479 §6 for why the obvious repair is unsound.
>
> Also worth carrying: **§5c's closing `[MEASURED 2026-08-25] the armed set … is 1 row
> fleet-wide` is the LOCKED subset** (0 today). The orphaning arm needs no lock, and Layer 2's
> actual preload predicate selects **378 rows across 371 pages** `[MEASURED 2026-09-04]`.

## 5. Mechanism — one hypothesis REFUTED, one UNVERIFIED

### 5a. REFUTED: `bugs_closed/189` (positional slot naming)

189 is the obvious story: *"resolving a locked positional slot duplicates it on the page"*,
filed on **this very page**, closed 2026-08-21 as fixed and live. It does not survive one
query. `[MEASURED 2026-08-24]` **all 11** locked tool sections on this site are
positionally named (`tool-1`…`tool-4`) and **none** matches its component's function; ten
went through the same wave; **one** duplicated. Positional naming is not the discriminator.

189's shape also differs in the deciding column: its duplicate carried **the same
`component_id`** as the locked row. This one carries **NULL**.

### 5b. UNVERIFIED: the incoming composition carried the slot twice

`save_page_sections_action.go:1074` is the single INSERT every page-composition path flows
through. It writes `component_id = componentIDPtr`, **which may be nil**, and the guard
immediately above it (`sectionIsUnresolvableStub`, `bugs_open/039`) deliberately refuses
only an *empty generic stub with no component link* — 11,845 bytes of working calculator is
outside that discriminator by design.

Reading the loop's arithmetic against the observed rows: five sections were inserted at
positions 1–5 with position 2 **skipped** by the locked-row match branch (the locked row is
repositioned there and its fresh copy discarded, which is `bugs_closed/058` working), and a
sixth entry — same `slot_name`, no `component_id` — was inserted at 6. `matchLockedRow`
can consume the lock **once**; a second entry naming the same slot finds it already
consumed and is inserted as a new row.

**What refutes the easy version of this:** the plan does NOT list the tool twice.
`[MEASURED]` `site_plan_sections` for `tool-loan-vs-savings` in plan `9463e31d` is exactly
`hero · tool-loan-vs-savings · ported-prose · faq · tool-cta`, byte-for-byte the same shape
as `tool-compare-loans` and `tool-overpayment-calculator`, which both rebuilt correctly.
**So if the composition carried six entries, something between the plan and the save added
the sixth.** LOCK-008's merge of locked rows into `pages.sections` is the obvious candidate
and is **UNREAD** — that is the next session's first move, not a conclusion.

**Also worth knowing before you reason from `updated_at`:** the reposition is
`UPDATE page_components SET position = $2 WHERE id = $1` — it does **not** touch
`updated_at`. So "the locked row's `updated_at` never moved" proves the BYTES were not
rewritten; it does **not** prove the row was not repositioned. The 2026-08-23 canary
evidence should be read with that limit in mind (its md5 evidence is unaffected).

### 5c. ESTABLISHED 2026-08-25: the Layer 2 interactive-preservation re-append, matched by slot-name string alone

**The writer is `save_page_sections_action.go`'s Layer 2 block** ("Preserve interactive
tool sections", `:484-608` at `3973260c5`-era HEAD), and the door it walked through is
its matcher (`:551-558`):

```go
for i := range sections {
    if sections[i].ComponentName == p.slot { matchedIdx = i; break }
}
```

— exact slot-name string only. No identity arm, no kebab-normalisation, although the
preload's own SELECT (`:496-505`) already fetches `pc.component_id` and `cc.function`
for exactly this pairing question. This is the same slot-names-are-not-component-
functions defect that `matchLockedRow` fixed with its identity arm and
`MergeLockedPageSlots` mirrors as arm 3 — in a **third** pairing judgement that was
never given the arm. `bugs_closed/189`'s own "STILL OPEN" section predicted the class on
08-03: *"the save still derives the name from `component_function`, and a build-path run
over a locked positional slot would still rename and duplicate."*

**The verified chain, joint by joint:**

1. The 08-23 14:07–14:12 build (corr `2e68eb79`, the owner-released `needs_page` wave)
   composed the plan's **5** sections. `[MEASURED 2026-08-25]` in `llm_call_log`: the
   generation loop ran iters {0,2,3,4}; iter 1 (the tool, no `llm_field_specs`) took the
   no-LLM `render_from_template` branch. No iter 5 — the composition was five entries.
2. At the save (14:14:24), `enrichSectionsWithComponentIDs` (`:397` — **before** Layer 2)
   resolved the plan's `tool-loan-vs-savings` entry to component `448422ce`. (Proven
   downstream: only the identity arm can have consumed the lock — see joint 6.)
3. The byte-twin dedup (`:472`, `bugs_open/156`) ran — **before** Layer 2, so it
   structurally cannot see what Layer 2 appends. Its own comment asserts *"Nothing below
   can re-introduce a duplicate under this key: the Layer 2 carry-forward appends only
   slots ABSENT from the set"* — true under slot-name identity, and exactly the blind
   spot: Layer 2 re-introduces the same BYTES under a different slot name.
4. The Layer 2 preload (`WHERE build_status='deployed' AND <interactive>`) selected the
   locked row `10be4f71`. `[MEASURED 2026-08-25]` this predicate selects **exactly one
   locked row in the entire fleet — this one**. On this site, 11 of 12 locked rows are
   `build_status='approved'`; the victim's alone is `'deployed'`. **That is the whole
   discriminator** for why one page in a ten-page wave duplicated: the other nine locked
   calculators were never preloaded into Layer 2 at all — their pairing was never even
   attempted, so their cleanliness says nothing about the matcher.
5. The matcher compared `'tool-2'` against `[hero, tool-loan-vs-savings, ported-prose,
   faq, tool-cta]` → no match → the "slot dropped entirely" arm (`:585-605`) **appended**
   `SectionData{ComponentName:'tool-2', HTML:<stored bytes>, ContentData:<stored>,
   ComponentID:''}` — ComponentID empty because the RFC_046 `carryStoredIdentity` opt-in
   is OFF (default). The enricher at `:397` had already run, so nothing ever resolves the
   appended entry. This accounts for **every column of the orphan**: slot `tool-2`,
   `component_id NULL`, byte-identical `rendered_html` AND `content_data` (`be85284e…` /
   `f65a0b6e…`), position 6 (`i+1`, appended last), `build_status='deployed'` (the
   INSERT's literal, `:1075`).
6. Insert loop: the plan's tool entry (id `448422ce`) consumed the lock via
   `matchLockedRow`'s **identity arm** → locked row repositioned to 2, fresh copy
   discarded — `bugs_closed/058` and the 182/204 identity fixes working exactly as
   designed. (Had the identity arm NOT fired, the fresh copy would have been INSERTed
   with its id at position 2 and the locked row exiled to the tail by the unconsumed-tail
   pass — the opposite of what §2 shows.)
7. The appended `tool-2` entry reached the loop with no id; the slot arm found the lock
   **already consumed** → no match → INSERTed as the orphan. The `bugs_open/039` stub
   guard correctly declined to catch it (11,845 B of working calculator is not an empty
   generic stub — §5b already noted this by design).

**Why the rerender arm is clean (the §7a trap, explained):** on `page_rerender` the
incoming set is built from the page's own rows, so its `ComponentName`s ARE slot names —
`'tool-2'` is present, the matcher matches, no append. The defect fires only when the
incoming NAME SPACE is the plan's (function names) while the stored name space is
positional slots — i.e. **only on the build arm, only on a decomposed/positionally-named
page**, which is why the 08-25 rerender wave over all ten pages proved nothing about it.

**The refutation of §5b's named candidate, so nobody re-chases it:** LOCK-008's
`MergeLockedPageSlots` pairs the plan entry to the locked row by arm 3
(`ComponentFunction == name`): `[MEASURED 2026-08-25]` `cc.function` for `448422ce` is
exactly `tool-loan-vs-savings`, and the 08-17 census in the lane NOTES (a join of locked
rows' `content_components.function` against `site_plan_sections` on plan `9463e31d`)
already proved the match held before the failure. The component row's last write is
08-19 21:19 (`component-template-fixer`, a template-repair sweep over 9 tool components
21:18–21:22 — template bytes, not function renames). The merge inserted nothing; the
composition left `load_page_sections_from_spec` and `plan_sections` at five entries
(joint 1's iteration census is the receipt).

**Still armed, and with what population:** `[MEASURED 2026-08-25]` the armed set —
locked AND `build_status='deployed'` AND interactive — is **1 row fleet-wide: this same
row**. The next **build-arm** rebuild of `tool-loan-vs-savings` re-duplicates it;
nothing else in the fleet can currently reproduce this until another locked interactive
row acquires `'deployed'`. `[INFERRED, not proven]` why this one row is `'deployed'`:
the save INSERT hardcodes `build_status='deployed'` (`:1075`), and this row was created
08-02 **22:51** — four hours after its ten siblings (18:46, all `'approved'`, the
adoption path's literal) — so it was most plausibly created by the save path and locked
afterwards. `page_components` has no `updated_at` trigger, so a later silent write
cannot be excluded; the orphan's existence is what proves the predicate held at
2026-08-23 14:14:24 regardless.

## 6. Why there is no `090` verdict, stated rather than omitted

Filed as required: intake `0a53b04e-e06e-48c8-ad11-4845d8ee96d5`, run correlation
`b53c355b-7bfc-4202-b61d-89f16decffe2`. It ran five iterations and returned
**`UNVERIFIABLE — Diagnosis NOT confirmed (stopped: iteration-cap)`**, with **zero**
non-bundle artifacts. This is the known LANDMINE (*"a 090 on a symbol in a large file
returns bundles and NO verdict"*), third route: the iteration cap, with no truncation.
`v3_site_actions.go` is **344,503 bytes** and `save_page_sections_action.go` is **88,798**,
both far over the ~60 KB bar, and the symptom named five symbols across the two.

**If you re-file, name ONE symbol** — the landmine's own advice, and the entry records that
a single-symbol scope still failed twice on other lanes, so budget for the declared
substitute rather than for a verdict.

**2026-08-25: the declared substitute was performed instead of a re-file.** §5c's chain
was established first-hand — every joint grounded in read code (file:line) or a query run
against the live DB that day, with the discriminating census (`build_status` over all 12
locked rows; the armed set fleet-wide) run rather than assumed. A re-filed single-symbol
090 would target `save_page_sections_action.go` (89 KB, over the landmine's ~60 KB bar),
which is the exact shape that returned UNVERIFIABLE twice on other lanes; per the owner
ruling of 2026-07-31 this substitution is declared here rather than silently made.

## 7. What was done to the live page, and what was deliberately NOT done

Owner-approved 2026-08-24. Remediation followed `bugs_closed/189`'s own worked recipe:

1. **Deleted the orphan row** (`3fd2639d…`) inside a transaction whose `WHERE` asserted
   every distinguishing property (`component_id IS NULL AND locked_at IS NULL AND
   position=6 AND slot_name='tool-2' AND md5(rendered_html)='be85284e…'`) so it could not
   reach the locked row, followed by a `DO`/`RAISE` block — **not** a block of `SELECT`s,
   which cannot stop a `COMMIT` — asserting 5 rows remain, the locked row still locked, its
   bytes unchanged and its `updated_at` still `2026-08-02 23:01:02.947526+00`. All passed;
   `DELETE 1`.
2. **The delete is recoverable and that was verified, not assumed:**
   `trg_page_component_artefact_archive_del` wrote the row to `page_component_history`
   (`op=delete`, `source=artefact_archive_trigger`, md5 `be85284e…`).
3. **Checked `pages.sections` BEFORE redeploying** — it is a materialised cache that
   LOCK-008 merges locked rows into, so a stale sixth entry there would have let the
   assemble re-materialise the duplicate and made the repair look done. It held **5**.
4. **Filed an ASSEMBLE-ONLY redeploy** — `page_rerender` with **no `spec.reason`**, which
   takes the `render_page` branch and stitches the stored `rendered_html`. Deliberately not
   `section_data_resolved`: that re-renders every section from `content_data`, which is the
   route `bugs_closed/189` warns reproduces the duplication, and it would rewrite 51 prose
   rows on a decomposed page. Item `98529d02-6e12-47af-968b-47a29d0a3962`, completed
   19:04:33, commit `e1becb2a` to `gqls/sites`.

**VERIFIED AT THE ARTEFACT `[MEASURED 2026-08-24]`** — the damage is gone:

```
served sha256   e3d2da2b… == the committed file, exactly   (was d30d112c…, 57,349 B)
duplicate ids   0                                          (was 11)
harness         react=5  vary=5  12 fields  —  identical to what GOLDEN_2026-08-17
                recorded for this page BEFORE the damage
divergences     8, all the cosmetic c-faq container rename; zero controls, zero numeric
```

Re-baselined afterwards: `GOLDEN_2026-08-24_post_385_repair_tool_values.json`, all 11 URLs,
and proven to reproduce (a fresh `--compare` returns 11/11, exit 0).

⚠ **If you re-verify and see the OLD page, RE-SAMPLE before concluding.** I read one `curl`
taken ~90 s after the B2 sync had logged its upload and told the owner the publish had
failed. It had not. `WRONG_CALLS.md` `## 2026-08-24` has why the false reading was
persuasive — a page differing from its peers mid-sweep looks *skipped* and usually means
*not yet reached*.

**NOT done:** no code change, because no cause is established. **Deleting one row is not a
class fix** — the writer that produced it is still live, and the next rebuild of any locked
tool page can do it again.

## 7a. STATE AT 2026-08-25 (post-roll) — three constraints added, none of them a cause

**Not recurred, anywhere.** `[MEASURED 2026-08-25]` `byte_twins_on_page` is **0** for all 9
remaining `component_id IS NULL` rows fleet-wide. On loancalculator: 28 active pages, 11
locks held, **0 orphan rows**.

**The writer is still live and unmodified.** Chassis `v1.0.1339`, build provenance
`a7459a44b` (asked the service). `save_page_sections_action.go` is **unchanged** between
2026-08-24 and that build — the one commit touching neighbouring files (`c735bfd9c`,
`bugs_open/375`'s verifier gate) has an **empty** diff for this file. Nobody has fixed this
and nobody claims to have.

**All ten locked tool pages rebuilt clean — on the WRONG ARM.** `[MEASURED]` every tool page
on the site rewrote 4 of its 5 rows on 08-25 between 13:04 and 13:34, locked row untouched,
none duplicated — **including `/tools/loan-vs-savings.html` at 13:11, the victim itself.**

> ⚠ **Do not read that as the bug being gone.** That wave was **24 `page_rerender` items,
> `source='side_effect'`, from a `tool-cta` template change** — the **rerender** arm
> (`rerender_sections`). The 08-23 duplication came from a **`needs_page`** item on the
> **build** arm (`page-build-handler` → compile → `save_page_sections`). Two different
> upstreams into the same INSERT. What is now established is that **the rerender arm is
> clean on a locked positional tool page**; the build arm has not run on one since it
> failed, and it is the arm that failed.

**A lead chased and CLOSED negatively — do not re-chase it.** The five
`source='save_page_sections_overwrite'` rows in `page_component_history` for the 08-23
rebuild are a **pre-overwrite snapshot of the rows that already existed**
(`save_page_sections_action.go:830-843`, `SELECT pc.id …`) — they describe the page BEFORE
the rebuild, not the composition it received. Four carry `component_id` NULL; the
calculator's carries `10be4f71`, which is a `page_components` **row id**, not a
`content_components` id. That reads like a type confusion pointing at an unresolvable
component. It is not: **153 of 23,627 markers fleet-wide carry a row id (0.6%), and 101 of
them are on this site across 12 pages between 08-03 and 08-25, against ONE duplication.** It
tracks locked/retained rows. A 101-to-1 ratio is not a cause. (The "code changed underneath
the data" explanation was also checked and refused: `git show ec653247f:` — the commit HEAD
was at when the duplication happened — carries the identical snapshot SQL.)

**Acceptance is green throughout:** all 11 calculators reproduce
`GOLDEN_2026-08-24_post_385_repair` exactly, exit 0, after both the roll and the rerender
wave, with `--selftest` green first.

## 7b. 2026-08-26 — the fix is LIVE, and the damage's own stale detection fired harmlessly

**LIVE, probed at the artefact `[MEASURED 2026-08-26]`:** an overnight roll (pods
`agent-chassis-6dd68888dc-*`, started 2026-08-25 23:11Z, after `b9d0f02be`'s 21:03Z)
carries both new symbols — `matchPreservedSectionIdx` AND `PairStoredToIncoming` present
in `/proc/1/exe` on **both replicas**, present-control (`matchLockedRow`) found,
absent-control (`matchPreservedSectionIdxZZZ`) absent. The armed row can no longer
duplicate; the owner's data-disarm question is now optional belt-and-braces.

**And the queue paid out a lesson the same morning.** `content_duplication:
tool-loan-vs-savings` was filed 2026-08-24 14:49 by a discovery run — a GENUINE
detection of the then-live duplicate, hours before §7's remediation removed it. The
remediation repaired the damage but never retracted the queued detection (the estate's
"closing the damage does not retract the pointers at it" shape). It sat `detected`
until the promoter re-enable on 08-26 dispatched it at 08:30: the `deduplicate-sections`
agent found nothing to delete (its `remove_component_ids` named the row §7 had already
deleted; `page_component_history` shows **zero writes** on the page that morning),
completed its parent, and filed a benign assemble-only `page_rerender` — the same
operation §7 step 4 used. Verified after: 5 rows, locked row untouched, byte-twins 0
fleet-wide, served page zero duplicate ids, and the golden harness reproduces exactly
(selftest green first). **If you repair damage by hand, grep the queue for detections of
that damage filed BEFORE your repair — they outlive it and fire when promotion resumes.**
The triage discriminator, measured on a second site the same day (webdesign.co.uk,
2026-08-26: 138 pre-resume `detected` rows, ALL inert): only a row whose `item_type`
has a known **(item_type, handler_agent) pair** can be auto-dispatched by the promoter —
handler-less types (`capability_gap` and kin) are roadmap rows outside its doors. So
the dangerous subset is `status='detected' AND created_at < <your repair>` **with a
handler**; this item fired because `content_duplication` → `deduplicate-sections` is
such a pair.

**CLOSE CRITERION (what keeps this file in `bugs_open/`):** the fix is live but has not
been exercised on the arm that failed. One clean **build-arm** rebuild of a locked
positional tool page (§9), with byte-twins 0 and the golden green after, closes this to
`bugs_closed/`. The rerender waves now sweeping the site (design rotation re-enabled
2026-08-26 09:20Z; the `bugs_open/397` GTM repair's chrome+44-page wave) run the SAFE
arm and do not qualify.

## 8. Fix candidates, ordered by what closes the door

> **RE-RANKED 2026-08-25, now that §5c names the writer.** The original three candidates
> (kept below) were aimed at the INSERT because the writer was unknown; the defect is in
> fact upstream of the INSERT, in a matcher whose fix is small and mirrors two already-
> shipped fixes of the same shape.

0. **THE fix: give the Layer 2 matcher `matchLockedRow`'s arms.** **BUILT, TESTED,
   COMMITTED, AND COUNCIL-APPROVED 2026-08-25** — round 1 `a799579fd`
   (`matchPreservedSectionIdx` + tests, mutation-verified) drew a REVISE whose gating
   objection (reuse_agent: a third hand-mirrored copy of the arms is the drift that
   minted this bug — unify) was **acted on, not defended**: round 2 `b9d0f02be` (+
   gofmt `3552e674b`) moves the relation into `datahelpers/slot_pairing.go` (register
   **LOCK-009**) and makes all THREE matchers adapters over it, with wiring scans in
   both packages so a re-inlined private copy fails the build. Council corr `ece638fb`
   **round 2 APPROVED** (2 advisory objections, none high; dispositions in LOCK-009's
   entry — including the editquality ordering concern, answered: both old closures were
   already arms-outer, so the ordering is preserved, not imposed).
   **The fix is Go, so it is INERT until an image rolls.** Whether it has shipped is a
   query, not an inference: ask the chassis for its `build provenance` stamp and run
   `git merge-base --is-ancestor a799579fd <stamp>` — do not assume either way, and do
   not re-add the fix. Until that ancestor check passes, the armed row (below) can
   still duplicate on a build-arm rebuild. Identity first
   (`sections[i].ComponentID == p.componentID`, both non-empty — the enricher at `:397`
   has already run), then slot exact (today's behaviour, kept), then kebab-normalised
   slot, then `p.componentFunction`/name against the incoming name (the merge's arm 3 —
   and the arm that decides THIS case even when enrichment fails). Every datum needed is
   already in the preload's `preservedSection`. One preserved row should claim at most
   one incoming section (the consumption rule both sibling matchers already have).
   Verify per §9 **on the build arm** — the motivating case is a locked, positionally-
   named, `build_status='deployed'` interactive row against a function-named composition.
1. **Make the state unrepresentable: a partial unique index on
   `(page_id, slot_name)`**, so a second row for a slot cannot be inserted at all. Turns a
   silent duplicate into a loud INSERT failure at the one statement every composition path
   flows through. Needs a census of legitimately-repeated slot names first — several pages
   carry `ported-prose` more than once, so the index likely has to key on something
   narrower, or on `(page_id, slot_name) WHERE component_id IS NULL`. (Note
   `save_sections_dedup.go`'s header argues against exactly this index; read it first.)
2. **Widen the `sectionIsUnresolvableStub` guard's second arm**: refuse ANY
   `component_id IS NULL` insert whose `slot_name` already exists on the page. Cheaper than
   1 and catches this exact shape; does not catch a duplicate that carries a component_id
   (which is `bugs_closed/189`'s shape, and 189's fix covers that arm).
3. **Make the re-render path self-healing** rather than fatal: `[tool-2 (pos 6)]`
   unresolvable *and* byte-identical to a locked row on the same page and slot is a
   removable duplicate, not a reason to fail the page. Weakest — it repairs the damage
   instead of preventing it — but it is the one that stops a page becoming unrepairable,
   which is the part that turns a cosmetic fault into a stuck one.

**Data-side disarm, available today without a roll:** flip the one armed row's
`build_status` from `'deployed'` to `'approved'` (matching its ten serving locked
siblings, so `'approved'` demonstrably serves). Removes the single armed instance but
not the class; an operator/owner call because it writes a human-locked row.

## 9. How to verify a fix

Do **not** re-induce on this page casually (`bugs_closed/189`'s warning still stands). With
the harness now working, the artefact-level check is one command:

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 toolgolden.py --selftest          # ALWAYS first — green, or nothing below counts
python3 toolgolden.py --compare acceptance/<the current golden> \
        https://loancalculator.co.uk/tools/loan-vs-savings.html
```

**And drive the right ARM.** As of 2026-08-25 the rerender arm is proven clean on all ten
locked tool pages; the **build** arm (`needs_page` → `page-build-handler` → compile →
`save_page_sections`) is the one that failed and has not run on a locked positional tool page
since. A fix verified only through `page_rerender` has not been verified against this bug.
The 08-23 wave's own item shape is the template — `NOTES ## 2026-08-23`, source
`loancalc_owner_release_20260823`.

`react=0` is this bug. A duplicate-id census is the cheaper pre-check:
`curl -s <url> | grep -o 'id="[^"]*"' | sort | uniq -d` — non-empty is the defect.

**Is the fix LIVE? (added 2026-08-25, council round 1's debug_historian ask.)** Two probes,
never an inference from git or a tag:
1. The stamp: chassis `build provenance` log line (or `/proc/1/exe` probe of a KNOWN sha),
   then `git merge-base --is-ancestor a799579fd <stamp>`.
2. The symbol, with BOTH controls in the same breath (never a discovery grep):
   `kubectl -n ai-persona-system exec <chassis-pod> -- grep -aq matchPreservedSectionIdx /proc/1/exe`
   — present-control `matchLockedRow` (in every current binary), absent-control
   `matchPreservedSectionIdxZZZ` (must fail). A probe whose controls don't discriminate is
   telling you about the probe.

**Who dispatches this action** `[MEASURED 2026-08-25]` (council round 1's guardian ask):
10 agent types match `save_page_sections` by config text; per the documented landmine,
`council-gate` and `fix-proposer` match on PROMPT TEXT and carry no save step, so the
dispatching set is **8**: `page-rerender`, `page-build-handler`, `site-work-orchestrator`,
`tool-recreation-handler`, `pageflow-builder`, `page-rebuild`,
`required-fields-missing-handler`, `diagnose-agent` — both arms, with the rerender arm's
name space (slot names) unchanged by the fix.

## 10. Related

- `bugs_closed/189` — same *damage*, different *shape* (same `component_id`, not NULL);
  refuted as the cause here in §5a. **But its "STILL OPEN" section predicted §5c's class
  on 08-03** ("a build-path run over a locked positional slot would still rename and
  duplicate") — the naming-mismatch ROOT is shared; the writer differs (189: the renamed
  fresh insert; 385: the Layer 2 re-append). Resolve 189 by SLUG — the number names two
  unrelated bugs.
- `bugs_open/156` / `save_sections_dedup.go` — the byte-twin dedup runs BEFORE Layer 2
  (§5c joint 3), so it cannot catch this shape; its "nothing below can re-introduce"
  comment needs correcting whenever the fix lands.
- `bugs_open/357` / RFC_046 — `carryStoredIdentity` is the opt-in that would have given
  the appended copy its component id (making it 189's shape instead of an orphan). The
  Layer 2 arms are that lane's active seam: cross-check before editing them.
- `bugs_closed/058` — the lock-preservation guard. It **worked**: the locked row kept its
  id, its bytes and its `locked_at`. It is not built to notice a second row arriving beside
  the one it protected.
- `bugs_open/039` — the unresolvable-stub guard whose narrow discriminator this row passes.
- `LANDMINES.md` — *"a browser harness that is down is probably reading `$TMPDIR`"* (why
  this went unseen for a day) and *"a 090 … returns bundles and NO verdict"* (§6).
- Lane docs: `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/`
  (`NOTES_loancalculator_couk.md` `## 2026-08-24`, `README_where_we_are.md`).
