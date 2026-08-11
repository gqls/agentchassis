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
