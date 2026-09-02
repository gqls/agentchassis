# HANDOFF — bugs_closed/423, chrome UTF-8 · written 2026-09-02, ~15:5xZ · **CLOSED 16:4xZ**

> ## ✅ THIS LANE IS COMPLETE. Read §0 first; the rest is the trail.
>
> **`bugs_closed/423`** — fixed, live on `v1.0.1354`, and **both casualties repaired at the
> artefact**: garden-tools.uk 16:21:32Z (NULL for ten days → 2,427 B) and boxingonline.com
> 16:27:56Z (2,289 B), footers fleet-wide not `rendered` **2 → 0**, and the webdesign lane's
> pre-delivery probe **passes** (empty `sites.email`, no contact block in the served footer).
> The offending em-dash label renders intact.
>
> **OWNER RULING 2026-09-02 — the escalation is UNGATED.** Council round 4 returned REVISE on
> the gate deletion after rounds 1–3 had pushed the other way; three seats in three directions
> meant the owner broke the tie per CLAUDE.md, not a fifth round. `badff59a9` carries it and
> **ships on the next roll** — that is the one behaviour change still pending. See §6.
>
> **What is genuinely left:** `bugs_open/435` (a residual swallow, deferred on a measurement)
> and a `ConfigKeys` declaration for `render_site_components`. Both are §8, both are optional,
> neither blocks anything.

**Read this first, then `NOTES_chrome_utf8.md` (newest at the bottom) and the bug file.**
Everything below is either measured with its query attached, or marked as unverified.

---

## 1. One paragraph of state

> **UPDATE 16:21Z, after this file was first written: the open verification in §2 PASSED.**
> garden-tools.uk's footer stores again after ten days. Half 2 is proven at the artefact.
> Only boxingonline (§3) stands between this and closure.

The root cause is found, fixed, committed and **LIVE on `agent-chassis:v1.0.1354`** —
proven at the binary with a removed-string control, not inferred from git or the tag.
The council **APPROVED** at round 3 and a round-4 resubmission is in flight because an
advisory made me delete a mechanism the earlier rounds had argued for. **What is NOT yet
proven is the thing that matters most: that a footer now actually stores.** A chrome
re-render for garden-tools.uk was dispatched at ~15:5xZ to settle exactly that, and its
result is the first thing to check.

---

## 2. ~~DO THIS FIRST — the open verification~~ ✅ DONE AND PASSED, 16:21Z

**The verification below RAN and PASSED while this handoff was being written.**
garden-tools.uk's footer — NULL since 2026-08-23 — stored at **16:21:32Z**: `rendered`,
**2,427 bytes**, `digest_ok = true`, header and head stored in the same run. The offending
label renders **intact**, em-dash preserved:
`How We Assess Garden Tools — Our Methodology | Garden Tools UK`.

Fleet census after: footers not `rendered` went **2 → 1** (boxingonline only, untouched by
choice — §3); rows with NULL/empty `rendered_html` went **1 → 0**.

**Half 2 is proven end to end at the artefact.** The only thing left for closure is
boxingonline (§3). The original instructions are kept below because they are the recipe to
re-run if you need to check another site.

### the recipe, retained

A `rerender-chrome` run was published for **garden-tools.uk**, correlation
**`af0857d2-61c5-4cf6-ab82-c7b5001134ad`** (publish receipt asserted, exit 0).

```sql
-- THE close condition for half 2. Was pending / NULL / digest NULL before.
SELECT s.domain, sc.slot_name, sc.build_status, sc.updated_at,
       length(sc.rendered_html) AS len,
       (sc.rendered_html_digest = md5(sc.rendered_html)) AS digest_ok
  FROM site_components sc JOIN sites s ON s.id = sc.site_id
 WHERE s.domain = 'garden-tools.uk' ORDER BY 2;
```

**PASS = `build_status='rendered'`, a non-null length, `digest_ok = true`.**

⚠ **Expect a wait, and do NOT re-fire.** `system.agent.generic.requests` is
single-partition/single-consumer; measured publish→run latency has been **25–36 minutes**
under load. A missing orchestration row is queue latency, not a drop — re-firing
duplicates the work. Find the run by payload, never by the printed id:

```sql
SELECT current_step, status FROM orchestration_states
 WHERE collected_data->'input_data'->>'domain' = 'garden-tools.uk'
 ORDER BY created_at DESC LIMIT 3;
```

**If it stored:** half 2 is proven end to end. Then re-run the two-way census in
`RUNBOOK_chrome_utf8.md` — the left column (offending labels) should be unchanged and the
right column (unstored footers) should be down to boxingonline only.

**If it did NOT store:** the fix is live and something else is wrong. The failure is now
LOUD by construction — that is half 1 — so look for it rather than guessing:

```sql
SELECT summary, spec->>'phase', spec->>'still_serving', created_at
  FROM site_work_items WHERE item_type='chrome_render_failed'
 ORDER BY created_at DESC LIMIT 5;
```
`spec.phase` tells you which half failed: `failed to render` (template),
`rendered invalid UTF-8` (a byte-slicer upstream — **the error carries the byte offset and
a printable window**, which is the whole point of the gate), or `rendered but was not
stored` (the database refused for some other reason).

---

## 3. The second casualty, and why I did NOT touch it

**boxingonline.com's footer is still `pending`, and that is deliberate.** It is a **paid,
live site mid-delivery by the `webdesign_uk_build_service` lane**, and what it currently
serves is a **hand-patch** made at 16:05 on 2026-08-31 — the only definition of that
site's footer, because `content_data` is empty by fleet norm. Re-rendering **replaces**
that hand-patch with a machine render.

That is the intended end state and I judged it was not mine to trigger unilaterally on a
paid site during someone else's delivery. **The owning lane has been told** —
`docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/NOTE_2026-09-02_from_bugfix_423_lane_root_cause_found.md`.

When it is re-rendered, that lane's own pre-delivery probe still applies: the served
footer must carry **no contact block**, because `sites.email` is empty and
`component_library.go:1988` gates it.

---

## 4. What shipped, and the proof it is live

**Live on `v1.0.1354`**, probed at the binary 2026-09-02 with four controls behaving
correctly (the startup `build provenance` line had already scrolled — an empty grep there
means "not in range", never "unstamped"):

| probe | expected | got |
|---|---|---|
| `rendered but was not stored` | present | **PRESENT** |
| `a byte-indexed slice upstream cut a multi-byte rune` | present | **PRESENT** |
| `This chrome component's template could not be executed` (deleted by this change) | **absent** | **absent** |
| nonsense string | absent | absent |

⚠ **That build carries commit `cccb5ccd6`, which still contains the escalation gate**
(`escalate_chrome_store_failure` probed PRESENT). The gate's **deletion** is commit
`badff59a9` and is **NOT in this build** — it rides the next roll. The gate defaults OFF,
so today's behaviour is identical either way. See §6.

**The commits, in order:**

| sha | what |
|---|---|
| `3edb30476` | the fix: `UpperFirst` + 8 call sites, half 1, the UTF-8 gate, 3 truncations, the emitter `phase` |
| `fbd802cf2` | STY-059 names its commit + the workstream dir (pattern-check advisories) |
| `0a14e31bd` | round-1 REVISE actioned: the opt-in escalation gate |
| `cccb5ccd6` | round-2 REVISE actioned: sketches corrected, `bugs_open/435` filed, WRONG_CALLS |
| `badff59a9` | round-3 advisory actioned: **the gate DELETED** — not yet rolled |

---

## 5. The mechanism, in case you need to re-derive it

`buildServicesHTML` (`render_site_components_action.go`, called at `:125` for the footer's
services column) title-cased each word with `strings.ToUpper(w[:1]) + w[1:]` — a **byte**
slice. `strings.Fields` makes a standalone em-dash its own word, so it cut a 3-byte rune
after one byte. Run it and you get `ef bf bd 80 94`: **first invalid byte `0x80`**, exactly
the byte in the live Postgres refusal. Postgres refuses invalid UTF-8, so the footer
`UPDATE` died — and the store-failure branch returned a **nil** error, so the step reported
success with `rendered.footer=false` as its only trace.

**The discriminating census (2026-09-02), which is the thing to re-run rather than trust:**
sites with a services-column label containing a word whose first rune is multi-byte, within
the query's `LIMIT 6` = **exactly 2**; sites whose footer is not `rendered` = **the same 2**.
Full SQL in the RUNBOOK.

Fixed as a **class**: `datahelpers.UpperFirst` + all **8** call sites (census 2026-09-02),
because the estate had already fixed the *truncation* shape of this class on 2026-07-20
(`SafeCut`, `bugs_open/027` §4b) and never went looking for the *casing* shape.

---

## 6. The live judgement call you may want to overturn

Rounds 1–3 argued about whether a store refusal should **fail the build**. The history is
worth reading before you touch it, because I was wrong in an instructive way.

- **Round 1**: I let it fail the step, by analogy with `bugs_open/260`. Guardian objected
  **HIGH**: enumerate the blast radius, don't assert it.
- **Round 2**: enumerated — **7** live workflows dispatch this action and **every one
  declares no `error_step`, no `on_error`, no `continue_on_error`**. So I put the
  escalation behind an opt-in key, default OFF.
- **Round 3**: `bug_historian` showed that was an over-correction, and checking it made it
  worse than the advisory said. **The arm could never gate "store failures" at all** —
  escalation is reached only via `chromeUnserved`, appended to only when
  `!chromeSlotHasStoredHTML`. Its entire reach was *a slot with nothing to serve*, the very
  state `260` and `bugs_closed/054` had already ruled must fail.

> **The error to inherit, not repeat:** the 7-workflows figure bounds **WHO is affected
> when the arm fires**. It says nothing about **HOW OFTEN it can fire**. That second number
> is **ONE row fleet-wide** (`WHERE rendered_html IS NULL OR rendered_html = ''`). I
> answered the second question with the first question's number, and it survived a HIGH
> objection, a redesign and two review rounds **because it felt rigorous**. A count of
> consumers is not a count of occasions.

So `badff59a9` deletes the gate; a store refusal takes `260`'s existing, reviewed
disposition, ungated. **After the next roll, a greenfield build whose chrome cannot be
stored will FAIL rather than ship a footerless site.** That is `260`'s stated intent, and
it is the sharp edge of this change. Expected newly-failing builds today: **zero**.

---

## 7. Council trail

**`SUBMISSION_CORR = dc62975f-9d38-4b3c-9174-330307b9df95`** — use this, not a run id.

Round 1 REVISE (guardian HIGH, blast radius) → round 2 REVISE (editquality HIGH, my sketch
was stale and contradicted its own rationale; prior_art HIGH, the enumeration had no query)
→ round 3 **APPROVED**, 2 advisories, none high → **round 4 IN FLIGHT** (published
~15:5xZ), submitted because the code changed after approval and a trailer must describe the
code that exists.

```sql
SELECT status, current_step FROM orchestration_states
 WHERE collected_data->'input_data'->>'fix_correlation_id' = 'dc62975f-9d38-4b3c-9174-330307b9df95'
 ORDER BY created_at DESC LIMIT 1;
```
Objections live in `diagnosis_artifacts` (`kind='council_report'`, same correlation).

**Commits carry `Council-Submitted:`, never `Council-Reviewed:`** — `098` credits the
correlation automatically once approved, with no amend (forward-only forbids one). **Do not
write `Council-Reviewed:` on a verdict you have not read.**

---

## 8. Open work, honestly scoped

1. **§2's verification.** The only thing between this and closure.
2. **`bugs_open/435`** — the `:1411` "no row matched" swallow, filed rather than fixed.
   Deliberate, on a measurement: **57** sites, only **34** with any `site_components` row,
   **23** missing at least one slot, so reporting there would file up to **~69** findings
   about sites nobody ever built. The real fix distinguishes "never built" from "the row
   vanished"; `trg_site_component_archive` (`sql_for_agents/344`) is the discriminator and
   needs no new state. The council named `bugs_open/034` as the same shape — worth checking
   whether one fix serves both.
3. **Declare `ConfigKeys` for `render_site_components`** (architecture advisory). Live steps
   carry **six** keys today and `update_page_status_config_contract_test.go` records that a
   wrong list **hard-failed every workflow that stamps the action** — so it earns its own
   change and its own review. Two undeclared sibling keys remain invisible to the RFC_022
   counter.
4. ~~`bugs_open/423` closes once §2 passes and boxingonline is re-rendered.~~ **DONE
   2026-09-02: both happened, and the bug is `bugs_closed/423`** (commit `d0c8ca9c3`,
   moved naming both paths and verified at HEAD with `git ls-tree`, per the `git mv`
   landmine). Nothing here is outstanding.

---

## 9. Traps this lane actually hit (all cost real time)

- **A predecessor's code read carries its own SEARCH BOUNDS.** The 08-31 addendum cleared
  "between `RenderTemplate` and the store" — correct, and useless, because the cut was
  *before* `RenderTemplate`, in an input built at `:125`. Enumerate the region they did not
  name. (`WRONG_CALLS.md`, 2026-09-02.)
- **`gofmt -w` on a shared file steals another lane's line.** `render_site_components_action.go`
  is unformatted **at HEAD** from `effc3a090` (noted_rebuild). Check first, always:
  `git show HEAD:<path> > /tmp/x.go && gofmt -l /tmp/x.go`.
- **`go test` failing in this package is usually NOT your change.** Two other sessions had
  files mid-refactor; **28** peer files needed isolating. The `go test -overlay` recipe is
  in the RUNBOOK. ⚠ A file **you** deleted must map to `""`, not to its HEAD copy, or you
  test a symbol you just removed.
- **`&&`, never `;`, behind anything that tells someone else something.** I published a
  council round in the same shell call whose edit had just failed an assertion; the
  submission asserted a code citation that did not exist for four minutes.
  (`WRONG_CALLS.md`.)
- **The census grep for this defect class matches its own cure.** On a clean tree
  `grep -rn "ToUpper(\w*\[:1\])"` returns **1** — `UpperFirst`'s own doc comment. Use the
  positive control instead: `UpperFirst(` returns **9** (8 call sites + the definition).
  Filed in `LANDMINES.md`.

---

## 10. The files

| | |
|---|---|
| bug (**CLOSED**) | `bugs_closed/423_HANDOFF_2026-08-31_footer_store_fails_on_invalid_utf8_and_the_renderer_reports_the_slot_as_a_reasonless_false.md` — moved from `bugs_open/` in `d0c8ca9c3` |
| spun-out bug | `bugs_open/435_HANDOFF_2026-09-02_chrome_store_that_matches_no_row_is_still_a_silent_success.md` |
| register | `docs/agent_docs/docs026_concept_register/register/styling-render-pipeline.md` → **STY-059** |
| lane | `docs/agent_docs/docs024_key_docs_latest/bugfix_423_chrome_utf8/` — PLAN, RUNBOOK, NOTES, README_where_we_are, SUMMARY, this handoff, the council submission |
| notes to peers | `webdesign_uk_build_service/NOTE_2026-09-02_from_bugfix_423_lane_root_cause_found.md` · `noted_rebuild/NOTE_2026-09-02_from_bugfix_423_lane_shared_file.md` |
| code | `platform/orchestration/datahelpers/data_helpers.go` (`UpperFirst`, `InvalidUTF8At`, beside `SafeCut`) · `platform/orchestration/actions/render_site_components_action.go` |
| tests | `datahelpers/utf8_safety_test.go` · `actions/chrome_failure_report_utf8_test.go` |

**Five tests, every mutation proven red.** Half 1's two dispositions are **not** unit-tested
(reaching them needs the whole render path mocked) — they are verified at the artefact,
which is what §2 is.
