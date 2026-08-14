# HANDOFF — TRACK B is AUTHORISED. Read `bugs_open/253` before you start it.

**Written 2026-08-11 (evening)**, against chassis **v1.0.1288**.
Entry point for Track B. Track A is finished — `HANDOFF_2026-08-11_after_track_a_decisions_pending.md`
carries its outcomes and the owner's rulings.

---

## 0. ⛔ READ THIS FIRST — the authorisation predates the finding that changes it

The owner authorised Track B. **Four hours later, `bugs_open/253` was found**, and it
changes what Track B *is*, so a session picking this up should put it in front of the
owner rather than treat the go-ahead as settled.

**What was authorised**, on the risk picture as it stood: 22 calculator pages, where
the danger was "the widget gets replaced by prose, or moved to the bottom of the
page". **Both of those are now guarded** — the tool row is locked, and its matching
rule is pinned by `save_sections_positional_tool_slot_test.go`.

**What `253` found, live, on the LMC homepage four hours after Track A decomposed it:**
the generic pipeline rewrote its `prose-0` and kept the words and links while
stripping **every layout component** — `class="card"` 18→0, `tool-grid` 3→0,
`btn-primary` 15→0, `highlight-box` 1→0, `hero` 1→0.

**Applied to Track B that means:** the calculator keeps working and keeps its markup
(it is locked), while the cards, buttons and grid *framing* it are silently flattened
on the next generic rebuild — on 22 live consumer-finance pages. That is a different
risk from the authorised one. It is not catastrophic and it is not silent-forever
(it is visible on the page), but it is real, it is unguarded, and the owner has not
seen it.

**The existing guard does not catch it.** `bugs_open/178`'s shrink floor measures
**text volume**: the 15:24Z attempt was refused at 35% of text kept, the 15:47Z one
passed at 84%. A rewrite can keep 84% of the words and 0% of the components.

**Recommended:** put `253` to the owner, and offer fix candidate 1 (a
component-class floor beside the existing text floor) as a small precondition, before
converting 22 pages into that exposure.

---

## 1. ✅ THE PIN — measured, and it is the thing that would have caused real damage

`HANDOFF_2026-08-10d` §2 warned that Track A's pin is unsafe for Track B. **Measured
2026-08-11, and the warning understates it.**

`decompose_lmc.py` pins `PINNED_REF = "b318a8fad"`. Against the **live DB rows** for
the 22 Track B pages:

| ref | stored row == repo bytes |
|---|---|
| `b318a8fad` (the current pin) | **6 / 22** |
| `origin/master` | **22 / 22** |

**Decomposing Track B from the current pin would write STALE calculator HTML over 16
live calculators** — reverting the `bugs_open/224` zero-rate guards and the
`bugs_open/225` SDLT fix, which was 16 months out of date and under-quoting by
£5,000. That is the single worst thing available to do on this estate, and it is what
the tool does today if you run it unchanged.

**Re-point the pin to a concrete SHA before decomposing anything.** As of writing
`origin/master` is **`e69b5b275`** (08-11 19:23). Do **not** pin to the branch name:
`decompose_lmc.py`'s own docstring says the pin exists "because a baseline that names
a moving thing stops being a control", and rerenders push to that repo continuously —
`origin/master` moved several times during this session alone.

**And re-verify at the moment of use**, because another session's rerender can move a
page between your pinning it and your using it. The check is ~15 seconds:

```python
# stored row md5 vs repo bytes at REF, for the 22 owned+verbatim pages
# (full script: this file's §4; it is the one check that must never be skipped)
```

> **⚠ `b26fdc81b` is NOT a sites-repo commit.** `load_lmc.py`'s baseline is named
> `BASELINE_2026-08-09_stored_md5_at_b26fdc81b.txt`, and `git cat-file -t b26fdc81b`
> fails in `~/projects/sites` — all 22 paths come back missing. The filename names a
> ref from the *other* repo. Do not go looking for it in the sites repo and conclude
> the baseline is corrupt.

## 2. State, verified on v1.0.1288 this evening

| check | result |
|---|---|
| pods `596d84f6b-{kmc2t,tb8gd}`, `189`/`204` both replicas | `1 / 1 / 0` |
| Track A pages still served == predicted | **16 / 17** — the 17th is the homepage, legitimately rewritten (`253`) |
| Track B population | 22 pages, `owned` + `["ported-page"]` |
| `loans-consolidation` | the finished shape: `["prose-0","tool-1","prose-2"]`, tool row `lock_type='permanent'` |

**A `predicted/` file is only valid until the framework next writes the page.** It
predicts *assembly*, not *content*. That is why the homepage now "fails" a byte-diff
it should fail. Re-derive, or diff a page the framework has not touched.

## 3. The per-page sequence (from 10c §4, with what is now settled)

1. **Decompose while the page is still `owned`.** Writing rows does not invoke the
   generic pipeline, so this step is safe at any policy.
2. **The tool row is born locked** — `load_lmc.py` does this itself (`lock_type='permanent'`,
   `component_id` NULL). Do not fire `section_data_resolved` at a locked row.
3. ~~Confirm the tool slot is named in the incoming composition~~ — **SETTLED.** A
   positional `tool-1` matches exactly, and again via the normaliser (`tool_1` →
   `tool-1`). The trap needs a composition that **omits** the slot, which is what a
   seeded site plan would produce. **Seeding a plan is the dangerous act; per-page
   rerenders are not.**
4. **Then** flip that page to `'generic'` — as a migration with a `DO`/`RAISE` verify
   block. A verify block of bare `SELECT`s cannot stop the `COMMIT`.
5. **Re-run the arithmetic on that page before starting the next**, with the controls
   in the same session or a green run is not evidence:
   `oracle.py --selftest-parse` and `oracle.py --mutate expectation --tools simple`,
   then `oracle.py --tools <the one you touched>`.

⛔ **Never flip a page that is still a single verbatim `ported-page` row** (10c §4).
`rebuild_policy='owned'` is the only thing standing between a verbatim calculator and
a generic rebuild that replaces it with prose. Unchanged and not affected by anything
settled above.

## 4. Commands

```bash
LANE=docs/agent_docs/docs024_key_docs_latest/loanandmortgagecalculator_couk
export DECOMP_WORK=<your own scratch dir>          # must be YOURS

# 0. AFTER EVERY CHASSIS ROLL — revalidate the mirror on an UNCHANGED page
python3 $LANE/deploy_pages.py --tag <tag> legal
diff <(curl -s -A Mozilla/5.0 https://loanandmortgagecalculator.co.uk/legal.html) \
     $DECOMP_WORK/predicted/legal.html

# 1. RE-POINT THE PIN, then re-verify the 22 at that exact SHA (§1). Non-negotiable.
# 2. build the manifest for ONE page, check tool blocks > 0 (a calculator MUST have one)
python3 $LANE/decompose_lmc.py $DECOMP_WORK/manifest.json --pages <slug> --verbose
# 3. predict, apply, deploy, diff — as Track A
python3 $LANE/load_lmc.py --check <page> && python3 $LANE/load_lmc.py --apply <page>
python3 $LANE/deploy_pages.py --tag <tag> <page>
# 4. arithmetic, controls FIRST, same session
cd $LANE && python3 oracle.py --selftest-parse && python3 oracle.py --mutate expectation --tools simple
cd $LANE && python3 oracle.py --tools <tool>
# rollback (PROVEN 2026-08-11 by round trip against the baseline md5)
python3 $LANE/load_lmc.py --restore <page>
```

**Manifest slug ≠ `pages.name`** (10d §4.1) — `--pages` takes the slug derived from
the file path, and an unmatched `--pages` writes an **empty manifest and exits 0**
printing `pages: 0`. Always read the `pages: N` line. And `--pages` suppresses the
"expected 23 tool pages" assertion, which is how the empty run stays quiet.

## 5. Suggested order

`loans-consolidation` is already done and is the reference. Start with a **class A/B
tool already covered by `oracle.py`** so the post-change arithmetic check is one
command — `compare-loans`, `interest-rate-stress-test`, `loan-vs-savings`,
`settlement-calculator` are the four that migration 377 moved here. Leave the class C
tools (`damage-checker`, `fact-finder`, and the four other fence-eligible pages) until
last: they have no external right answer and want `invariants.py`, not arithmetic.

## 6. Everything else outstanding

- `bugs_open/253` — **read before starting**; governs this whole track.
- `bugs_open/251` → then `252`. Order matters: `og:url` must agree with the canonical.
- `bugs_open/250` — one item left: round-trip the loancalculator `--restore`.
- The **live LMC homepage** is serving the flattened version. Decision owed (`253`).
  `load_lmc.py --apply index` will **REFUSE** — its baseline guard sees the row has
  moved, correctly. Any repair must be deliberate and targeted.
- `loancash_couk_fca_validation/` — caps verified current; the worry moved to
  `complaint-deadline-calculator.html`, still unchecked.
- Track C (loancash decomposition) needs `decompose_lmc.py` ported, not run.

---

# ADDENDUM 2026-08-12 — the component floor is BUILT (not rolled), and §0 needs correcting

**Owner rulings taken this session:** *fix the component floor before Track B*, and
*the live homepage can wait until 253 is fixed*. The first is done. **The second was
already moot** — see below.

## What changed since this file was written

| | state |
|---|---|
| chassis | **v1.0.1291** (pods `6588556967-*`, up 14:55Z). `189`/`204` = `1 / 1 / 0` on both replicas |
| the component floor | **COMMITTED `0c8e08ccb`**, `Council-Submitted: b30ac52c`. **NOT ROLLED** — committed is not running |
| the flattened homepage | **already repaired**, 08-12 16:03, by the lane session seeding a `content_direction`. `card` 12, `tool-grid` 2, `btn-primary` 12, `hero` 1 |
| Track B | one page converted by the lane session (`owned decomposed` 1 → 2, `owned verbatim` 22 → 21) |

## ⛔ Three corrections to what this file says above

1. **§0's "read 253 first" still stands, but its risk is now guarded** —
   `save_sections_component_floor.go` refuses a save that keeps a slot's words and
   strips its layout. **It is not rolled.** Track B is not protected by it until a
   pod is serving it, and this lane proves that at the artefact, never at the tag.

2. **`253` is an AMBIGUOUS NUMBER.** It names two unrelated bugs. The other is
   `..._label_match_overlap_count_ties_on_incidental_nav_label_words`, and **its** fix
   commits (`c6dcbcaa8`, `6ea633cea`, `9b7811d4b`) are what a `git log` by number
   returns. Refer to this one by slug — `framework_rewrite_of_a_prose_block`.

3. **The code fix is the SAFETY NET, not the remedy.** The page was repaired by
   telling the writer its component vocabulary in `content_direction`, not by any
   platform change. **The writer was uninstructed, not malfunctioning:** handed markup
   with no description of what the markup means, it produced clean prose; told the
   vocabulary, it kept it. So for each Track B page, the *first* move is a
   `content_direction` that names the calculator page's vocabulary — the floor is
   what catches the page where nobody thought to write one.

## The floor, in one paragraph, so you can judge it

Refuses a save where a same-named slot loses more than 50% of its **class
attributes**. Calibrated on the real before/after — prose-0 went **43 → 1** on the
flattening (0.02) and **43 → 31** on the good rewrite (0.72), so the two are **35×
apart** and 0.25/0.34/0.50 all separate them. Slots under **10** class attributes are
out of scope (fleet median is 5, p90 35, over 1,422 unlocked slots; ~31% in scope).
Counts attributes not tokens; blind to *which* classes on purpose. Default **ON** at
0.5, `section_component_floor` in step config tunes it, 0 disables. Fails closed,
refusal writes nothing and raises a queue item.

**Its stated weakness, which a reviewer should press on:** the safety evidence is
**one** good rewrite. If Track B throws false refusals, that is the thing that was
thin, and the escape hatch is the config key — not deleting the guard.

## Still true, still owed

- **The pin.** Unchanged and still the thing that would do real damage:
  `PINNED_REF b318a8fad` matches only **6 of 22** Track B pages.
  `decompose_lmc.py` now **REFUSES** on a stale pin (`feeb85acf`), so the tool will
  stop you — but re-point it to a concrete SHA and re-verify at the moment of use.
- `bugs_open/251` → then `252` (`og:url` must agree with the canonical).
- `bugs_open/250`: round-trip the loancalculator `--restore`.
- **Read the council verdict for `b30ac52c`** before anyone writes
  `Council-Reviewed:` on the floor. `098` credits it automatically once approved.

---

# ADDENDUM 2026-08-13 — floor is LIVE, council said REVISE, and Track B is nearly done

## Verified this session, on v1.0.1295 (pods `68ddcf9655-*`, up 13:53Z)

| check | result |
|---|---|
| component floor **in the running binary** | `section_component_floor` **3** on both replicas; positive control `SECTION SHRINK REFUSED` **1**, negative control **0** |
| `189`/`204` | `1 / 1` both replicas |
| Track B | **owned decomposed 18, owned verbatim 5** — the lane session has converted 17 of 22 |
| floor refusals, fleet, 24h | **0** (shrink refusals also 0 — quiet, not proven working in production) |

**So the owner's precondition is met: the floor is live, and Track B is unblocked
on that ground.** It is NOT proven in production — zero refusals is consistent both
with "nothing tried to flatten" and with "it is not being reached", and the second is
now known to be true for most writers (below).

## ⛔ COUNCIL VERDICT: REVISE — and it found a real defect

`b30ac52c`, round 1, 11 reviewers, gating objection from `bug_historian` (HIGH).
Objecting seats: `bug_historian`, `guardian`, `debug_historian`, `constitution`,
`architecture`. Full working in `bugs_open/253_..._framework_rewrite...`.

**The finding: both floors guard 1 door of 9.** Nine Go writers touch
`page_components.rendered_html`; only `save_page_sections` is guarded. The one that
matters is **`ApplySectionEditAction`** — live via the `section-editor` agent, three
direct `UPDATE … SET rendered_html` sites, and **it is the per-component edit path
decomposition exists to enable**. The guard covers the door the incident came
through and misses the one the design steers people toward. `rerender-pages`,
`report-builder` and `tool-generator` are also live and unguarded.

**And it is inherited**: the `bugs_open/178` TEXT floor has the same single-call-site
wiring, so both floors have protected one door since 2026-08-02.

### The revision (NOT done — this is the next task)

1. Extend **both** floors to `ApplySectionEditAction` first. It updates a single row
   by id, so the comparison is simpler than the save path's — read that row's
   existing `rendered_html`, compare, refuse on the same terms.
2. Then decide, per remaining writer, whether it can replace an existing prose slot
   at all (`adopt_verbatim`, `create_report_page`, `create_tool_component`,
   `deploy_tool`, `fix_*_colours`, `rebuild_blog_listing`). Several plausibly cannot;
   say which and why rather than wiring all nine reflexively.
3. Resubmit on the SAME correlation: `RESUBMIT_CORR=b30ac52c` so the trail accumulates.
4. Also answer the low-severity objections, both fair:
   - `editquality`: does the guard issue one query per save or per slot? **Per save**
     — one `page_components` read, so the single added `ExpectQuery` is correct. Say
     so; it was a reasonable question the plan did not answer.
   - `bug_historian`: `minComponentGuardClasses=10` leaves a flattening just under
     the threshold silently unrefused. State it as residual exposure.

## What this means for Track B

Track B's prose rows are protected **only** against a flattening arriving via
`save_page_sections`. A flattening arriving via `section-editor` — the tool most
likely to be pointed at a decomposed prose block — is still silent. With 18 pages
already decomposed, that exposure is live now, not hypothetical.

**Recommendation:** finish the revision before the remaining 5 pages, or at minimum
before anyone runs `section-editor` at an LMC prose row.

## Unchanged and still owed

- The pin: `b318a8fad` matched 6 of 22. `decompose_lmc.py` REFUSES on a stale pin
  (`feeb85acf`), so the tool stops you — re-point to a concrete SHA and re-verify at
  the moment of use.
- `bugs_open/251` → then `252`. `bugs_open/250`: round-trip the loancalculator restore.

---

# ADDENDUM 2026-08-14 — the floors now reach the section editor, APPROVED, and what is still NOT rolled

**Council APPROVED `b30ac52c` at round 3** (10 of 11 seats; `bug_historian`, who gated
round 1, approves). Both REVISE rounds found real defects — see `bugs_open/253`'s
table. Commits already carry `Council-Submitted: b30ac52c`, so `098` credits them
automatically; **do not add a `Council-Reviewed:` trailer by amend** (forward-only).

## What is live vs what is only committed — the distinction that matters

| | state |
|---|---|
| component floor on **`save_page_sections`** | **LIVE** since v1.0.1295 (verified in the binary, both replicas, with controls) |
| both floors on **`ApplySectionEditAction`** | **COMMITTED, NOT ROLLED** — needs the next chassis build |
| coverage test (the class fix) | committed; it is a test, so it protects the *repo*, not production |

**So Track B's 18 decomposed pages are still unprotected on the section-editor path
until the next roll.** That is the path most likely to be pointed at a decomposed
prose block. Verify at a pod before relying on it — `strings /app/agent-chassis |
grep -c enforceSingleSlotFloors` with a positive and a negative control, never the tag.

## What the coverage test buys the next person

Every file that `UPDATE`s `page_components.rendered_html` must now either enforce a
floor or sit in `exemptWriters` **with a reason**. A tenth writer fails the test until
its author decides in writing. It already earned this: it caught
`create_report_page_action.go`, which the manual audit had filed as create-only and
which in fact overwrites its own row.

**Its stated weakness** (in its own header): it reads SOURCE, so it proves wiring
EXISTS, not that it EXECUTES. A call in a dead branch would satisfy it.

## Two disclosures

1. **The actions package is currently RED from another session's work** —
   `TestLegacyLogoStep_StaticPurposeIsShadowedByDefault` and
   `TestPurposeFieldBridge_DeadForDefaultedField` in
   `deploy_image_asset_purpose_source_test.go`, last touched by `be1cd6b9d`
   (`test(231/380)`). Unrelated to the floors; my own surface passes. Not fixed —
   it is another session's in-flight work. **Do not read a red package here as
   yours** without checking which tests.
2. **The kubeconfig token expired mid-session** (`Unauthorized` fleet-wide) and
   recovered. Documented 3-day expiry; the owner refreshes it.

## Still owed, unchanged

- **The pin** — `b318a8fad` matched only 6 of 22 Track B pages. `decompose_lmc.py`
  now REFUSES on a stale pin (`feeb85acf`); re-point to a concrete SHA and re-verify
  at the moment of use.
- **The last 5 Track B pages** (18 of 23 decomposed).
- `bugs_open/251` → then `252` (`og:url` must agree with the canonical).
- `bugs_open/250` — round-trip the loancalculator `--restore`.
- Convert the colour fixers' exemption from REASONED to MEASURED: run their
  transform over a fixture carrying class attributes and assert
  `countComponentClasses` is unchanged. The experiment is written down in
  `single_slot_floors.go`.
