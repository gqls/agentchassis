# 385 — a rebuild can append an UNLINKED copy of the locked section it just repositioned, and the page then renders that tool twice and can never be re-rendered again

**Filed 2026-08-24**, `loancalculator_couk` lane, found by the acceptance harness on its
first working run after being repaired (`0aafce405`). **The live damage is REMEDIATED**
(see §7); the **root cause is OPEN** and is the reason this file exists.

> **ROOT CAUSE NOT ESTABLISHED. Do not read §5 as a finding.** The `090` route is closed
> for this mechanism — see §6 — and the two hypotheses in §5 are one REFUTED and one
> UNVERIFIED. Per the owner ruling of 2026-07-31 I am declaring the substitution
> explicitly rather than omitting it: what is verified first-hand here is the **damage
> and its shape**, measured three independent ways; what is *not* is which writer
> produced the row.

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
3. **Filed an ASSEMBLE-ONLY redeploy** — `page_rerender` with **no `spec.reason`**, which
   takes the `render_page` branch and stitches the stored `rendered_html`. Deliberately not
   `section_data_resolved`: that re-renders every section from `content_data`, which is the
   route `bugs_closed/189` warns reproduces the duplication, and it would rewrite 51 prose
   rows on a decomposed page. Item `98529d02-6e12-47af-968b-47a29d0a3962`.

**NOT done:** no code change, because no cause is established. **Deleting one row is not a
class fix** — the writer that produced it is still live, and the next rebuild of any locked
tool page can do it again.

## 8. Fix candidates, ordered by what closes the door

1. **Make the state unrepresentable: a partial unique index on
   `(page_id, slot_name)`**, so a second row for a slot cannot be inserted at all. Turns a
   silent duplicate into a loud INSERT failure at the one statement every composition path
   flows through. Needs a census of legitimately-repeated slot names first — several pages
   carry `ported-prose` more than once, so the index likely has to key on something
   narrower, or on `(page_id, slot_name) WHERE component_id IS NULL`.
2. **Widen the `sectionIsUnresolvableStub` guard's second arm**: refuse ANY
   `component_id IS NULL` insert whose `slot_name` already exists on the page. Cheaper than
   1 and catches this exact shape; does not catch a duplicate that carries a component_id
   (which is `bugs_closed/189`'s shape, and 189's fix covers that arm).
3. **Make the re-render path self-healing** rather than fatal: `[tool-2 (pos 6)]`
   unresolvable *and* byte-identical to a locked row on the same page and slot is a
   removable duplicate, not a reason to fail the page. Weakest — it repairs the damage
   instead of preventing it — but it is the one that stops a page becoming unrepairable,
   which is the part that turns a cosmetic fault into a stuck one.

## 9. How to verify a fix

Do **not** re-induce on this page casually (`bugs_closed/189`'s warning still stands). With
the harness now working, the artefact-level check is one command:

```bash
cd docs/agent_docs/docs024_key_docs_latest/loancalculator_couk
python3 toolgolden.py --selftest          # ALWAYS first — green, or nothing below counts
python3 toolgolden.py --compare acceptance/<the current golden> \
        https://loancalculator.co.uk/tools/loan-vs-savings.html
```

`react=0` is this bug. A duplicate-id census is the cheaper pre-check:
`curl -s <url> | grep -o 'id="[^"]*"' | sort | uniq -d` — non-empty is the defect.

## 10. Related

- `bugs_closed/189` — same *damage*, different *shape* (same `component_id`, not NULL);
  refuted as the cause here in §5a. Cross-link both ways if this one is ever explained.
- `bugs_closed/058` — the lock-preservation guard. It **worked**: the locked row kept its
  id, its bytes and its `locked_at`. It is not built to notice a second row arriving beside
  the one it protected.
- `bugs_open/039` — the unresolvable-stub guard whose narrow discriminator this row passes.
- `LANDMINES.md` — *"a browser harness that is down is probably reading `$TMPDIR`"* (why
  this went unseen for a day) and *"a 090 … returns bundles and NO verdict"* (§6).
- Lane docs: `docs/agent_docs/docs024_key_docs_latest/loancalculator_couk/`
  (`NOTES_loancalculator_couk.md` `## 2026-08-24`, `README_where_we_are.md`).
