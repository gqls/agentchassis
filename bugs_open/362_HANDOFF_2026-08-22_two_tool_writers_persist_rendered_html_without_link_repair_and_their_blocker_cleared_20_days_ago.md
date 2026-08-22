# 362 — the two tool writers persist `rendered_html` without link repair, and the blocker that justified the delay cleared twenty days ago

**Filed** 2026-08-22 by the session that took the `RFC_008` ruling, executing that RFC's own
**Recommendation 2** verbatim: *"close the two known-unguarded writers
(`create_tool_component_action.go`, `deploy_tool_action.go`) on their own merits, in the tool lane
that owns those files. That is a bug fix, not architecture."*

**Owner** **`webdesign_tool_rebuilds` — TAKEN 2026-08-22**, accepted by that lane within the hour on
the evidence below. (Routing basis: `scripts/who-owns.py` and the commit record both put both files
there — TL-043/044/047/048, `bugs_closed/331`, `bugs_open/283`'s flow half, commits as recent as
2026-08-21.) Two sessions carry that lane name; the accepting one has notified its sibling and will
hand over if the sibling objects, so **the lane holds it either way — do not treat this as unowned,
and do not start a competing fix.** The filing session wrote the evidence and stopped there.

**Status** OPEN, OWNED, fix in progress. Low urgency by measurement (§3: live damage is one instance,
and that one is probably legitimate) — but the *reason* it was still open is the interesting part,
and it is not "nobody got round to it".

> **On the 2026-07-31 owner ruling** (a cross-cutting root-cause claim needs the `090` loop or a
> stated substitute): **`090` not run, deliberately.** This file asserts no new causal theory. The
> mechanism was diagnosed and fixed elsewhere (`bugs_closed/180`), the writer set is a grep restated
> inline, and the damage figure is a single query restated inline. What is new here is *bookkeeping*:
> a blocker that cleared and two documents that did not notice.

## 1. The defect, in plain terms

An AI writing a page's HTML sometimes writes links that point nowhere. There is a shared repair
function that fixes them in the moment before the HTML is saved. Ten pieces of code save that HTML;
**two of them — both tool writers — never call it.** A tool page can therefore be born with dead
internal links that no later pass repairs.

`RFC_008` asked whether *every* writer should be forced through one save function. The owner ruled
on 2026-08-22 that it should not — the estate detects and attributes rather than refuses. **That
ruling makes this file the whole of the remaining work on the two writers**: not a seam, just wiring
two call sites into a repair that already exists and is already used by three others.

## 2. Why it is still open, which is the transferable part

The wiring was **deliberately deferred on 2026-08-02** for an excellent reason, recorded as a
landmine that day:

> *"wiring the seam into the tool-markup writers must wait for `180`, or it will delete working
> buttons from tools."*

`bugs_open/180` was that repair destroying JavaScript that *builds* anchors: against
`'<a href="' + q.link + '">'` the href capture cannot cross the quote, so the repair read an empty
href and **deleted a working link from a running program**, leaving valid JS and readable prose
behind. Tool markup is exactly where such code lives, so the deferral was correct.

**`180` was fixed the same day it was filed** — `07576d4e1`, 2026-08-02, LNK-029's span-aware
`NonMarkupSpans`/`ReplaceAllInMarkup`, measured over all 509 assembled pages with **0 legitimate
repairs lost** — and it is in `bugs_closed/`. The register even records the consequence explicitly
in LNK-029's own relations line: *"the tool-markup writers, whose wiring this UNBLOCKS."*

**Nobody acted on it for twenty days, and two documents are why.** Both said "wait":

- `LANDMINES.md`'s entry still carried the 2026-08-02 sentence above, with no note that `180` had
  closed. A session reading it before touching either file would correctly conclude the wiring was
  premature.
- LNK-029's relations line pointed at **`bugs_open/136`** for the unblocked work — but `136` is
  **closed**, so the pointer led to a finished bug and the work looked done.

Both corrected in the same commit that files this bug. **This is the estate's known
`a-stale-status-line-prevents-the-thing-it-describes` shape, third recorded instance** (the others:
PBP-040's detector sat unarmed nine days after its blocker cleared; RFC_022's "not built" line ran
three days past the build). The generalisable check is in §6.

## 3. The damage, measured 2026-08-22 — and it is nearly nil

```
tool components (pages named tool-*, rendered_html not null)   597   across 285 pages / 24 sites
  of those, containing a <script> block                        213   ← the 180 risk class
  matching href=""  (UPPER BOUND, see below)                     1
```

The single match is `tool-archetype-taster-quiz`, slot `tool-archetype-taster-quiz`:
`href="" class="result-cta-primary"` — **a quiz result CTA, whose href is almost certainly filled in
by JavaScript when the quiz completes**, i.e. the legitimate runtime-fill case that
`data-runtime-fill` exists to exempt. **Verify before repairing it**: repairing a runtime-fill
anchor is the same class of damage as `180` itself.

⚠ **That figure is an UPPER BOUND and cannot be tightened by a better regex.** The landmine states
the rule: *a regex over `href="…"` counts href-shaped byte sequences, not links* — its own 2026-08-02
census had one JS fragment in 35. So the honest reading is **"no confirmed live damage"**, not
"exactly one".

**Why fix it at all, then?** Because the exposure is prospective, and cheap to close: 597 components
and 24 sites are one careless generation away from the defect, the repair already exists, three
sibling writers already call it, and the reason for the delay is gone. It is prophylactic wiring, and
should be sized and scheduled as such — not as an incident.

## 4. What the fix is

Two call sites, in the shape the three existing callers already use
(`create_report_page_action.go:202`, `section_editor_actions.go:493`):

```go
sectionHTML = repairComponentHTMLBeforePersist(ctx, params, siteID, /* … */)
```

- `platform/orchestration/actions/create_tool_component_action.go` — before the INSERT at `:497`
- `platform/orchestration/actions/deploy_tool_action.go` — before the INSERT at `:500`

Then **remove nothing from the allow-list.** `scripts/pattern-check.py`'s `COMPONENT_WRITE_ALLOWED`
deliberately omits both files, and its own header says why: *"an allow-list that absorbs the open
cases silences the detector on exactly what it was written to catch."* Once wired, the check passes
because the writers repair — which is the honest way for it to go quiet.

## 5. Acceptance

| claim | test | the control that makes it falsifiable |
|---|---|---|
| the repair runs at both writers | a test per writer that mutates the call away → the named test fails | mutation-proved: a mock's own bookkeeping cannot assert a negative |
| **it does not destroy JS-built anchors** | run the REAL `RepairPageLinks` over real tool bytes containing `'<a href="' + x + '">'` and assert byte-identical output | **this is the landmine's own prescribed check** — three lines against `datahelpers`, decisive where reading the regex is not. Do this even though LNK-029 fixed it: the fix is the reason you may wire, not a reason to skip the probe |
| no live regression | re-run §3's census after the roll | it must not rise; if it falls, say which component and verify at the served page, not the row |
| the check goes quiet honestly | `pattern-check.py` stops firing on both files | it must go quiet because the writers repair, never because they were allow-listed |

## 6. The transferable check (also going to 016b §9)

**When you close a bug, grep for documents that told people to WAIT for it.** A blocker's closure is
not self-announcing: the bug moves to `bugs_closed/` and every landmine, register relation, plan and
handoff that deferred work "until `NNN`" keeps saying wait, indefinitely, in a voice that sounds
current. The cheap version, at close time:

```bash
grep -rn "wait for 180\|blocked by 180\|until 180" docs/ bugs_open/ *.md
grep -rln "bugs_open/180" docs/agent_docs/docs026_concept_register/ docs/agent_docs/docs024_key_docs_latest/LANDMINES.md
```

The second is the sharper one: **a `bugs_open/NNN` reference in any document, where `NNN` now lives
in `bugs_closed/`, is a stale-status suspect** — and it is mechanically detectable across the whole
corpus, which makes it a candidate for `scripts/pattern-check.py` rather than a habit.

> **Status of that check: PROPOSED, UNOWNED, deliberately not built by the filing session.**
> Endorsed independently by the `webdesign_tool_rebuilds` lane (which took the Go half above) and by
> the `bugs_open/358` lane.
>
> ⚠ **CORRECTED 2026-08-22, hours after filing — it does NOT fold into 358's B2, and the reason is
> worth keeping.** This section first said that if 358's B2 built a general "grep the corpus,
> compare against a live source of truth" helper, this should live inside it as a second rule.
> **B2 is now built (`cmd/config-key-audit --finding-codes`, DBG-075) and is not that shape at all:**
> it is **DB-authoritative** — `SELECT DISTINCT error_code` is its source of truth *precisely
> because* scanning the Go source proved untrustworthy for the job (codes reach the table as
> positional arguments as well as `ErrorCode:` fields, so a structural grep misses them; that lane's
> first two source-based censuses returned test fixtures while missing twelve real constants).
> **So there is no corpus-scanning helper for this to be a second rule inside.** It wants
> `scripts/pattern-check.py` on its own terms, as originally described. Boundary agreed by both
> sessions and recorded on both sides.
>
> Whoever builds it owes a concept-register entry: it is a new reusable mechanism, not a bug fix.

## 7. Not in scope

- **The mandatory write seam.** Ruled against, 2026-08-22 — `RFC_008`'s decision record, with the
  four triggers that would reopen it. Wiring these two writers is explicitly *not* the seam.
- **The detectors that still misread JS-built anchors** (`links.go`'s phantom scan,
  `check_dead_controls`, `check_phantom_internal_links`). LNK-029 was writers-only, deliberately: a
  false finding costs attention, a false repair costs content. That gap belongs to the detection
  lane (`bugs_open/097`, `bugs_open/116`).
- **Repairing the one live instance** until someone has established it is not a runtime fill (§3).

---

## 8. §3's uncertain instance SETTLED 2026-08-22 (owning lane): it is REAL damage, not a runtime fill

Read the component's full bytes (10,742 chars, `vonc.com/tools/archetype-taster-quiz/index.html`):
the CTA `<a href="" class="result-cta-primary">Get Your Full Report</a>` appears once in markup and
**no JavaScript anywhere in the component touches it** — no `.href` assignment, no selector on
`result-cta-primary`, and `data-runtime-fill` count is 0. So it is not the legitimate fill case;
it is a genuinely dead link, and worse than inert: an `href=""` navigates to the CURRENT page, so
clicking "Get Your Full Report" **reloads the quiz and destroys the visitor's just-computed result**.
§3's honest reading moves from "no confirmed live damage" to **"exactly one confirmed dead CTA,
live on vonc.com"** — the upper bound was attained. (It is also another instance of the
"tool asserts something untrue about itself" class the webdesign lane has been logging: a
report-promising button that delivers a page reload.)

Repairing that ROW is content work for vonc.com's owner, not part of this wiring (unlinking would
stop the result-destruction but the page still promises a report it cannot give — that needs a real
destination or the CTA's removal, a content decision). Filed here so the census's one match is never
again read as probably-benign.

## 9. FIXED (wiring) 2026-08-22 by `webdesign_tool_rebuilds` — commit carries the acceptance

Both call sites wired in the three-sibling shape (`create_tool_component` before the
page_components INSERT, after the pages row exists so the tool's own URL is in the index;
`deploy_tool_to_site` before its INSERT). Acceptance, run before commit:
- **JS-built-anchor probe (the landmine's own): PASS** — the REAL seam over `'<a href="' + q.link +
  '">'` bytes returns the script span BYTE-IDENTICAL while a phantom markup anchor beside it IS
  unlinked (text kept) and a valid one untouched (`tool_writer_link_repair_362_test.go`, two-armed
  so a do-nothing seam cannot pass).
- **Mutation-proven wiring test**: deleting the deploy_tool call in a scratch copy fails
  `TestToolWritersCallTheRepairBeforeTheirPersist` naming the file.
- **Allow-list untouched**; the pattern check's quiet is the honest kind (its historical positive
  control: the 2026-08-20 commit `e3dee9243`'s pre-commit output, where it fired on exactly these
  two files).
- **§5's census re-run is OWED post-roll** (the wiring is Go, inert until the next chassis roll);
  whoever rolls next re-runs §3's query and appends the number here.
