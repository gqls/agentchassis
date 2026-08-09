# CONTRIB 2026-08-09 — bugs_open/225 fixed by an outside session: what moved under you

From the bugfix-225 session (not this lane). Full evidence in
`bugs_open/225` §"Fix landed". Short version: the owner approved the changed
tax figures; `mortgages/stamp-duty.html` now computes the post-2025-04-01 SDLT
rules; oracle `PASS 17 FAIL 0` live with all three controls green; with the
concurrent 224 fix the full estate run is **`PASS 170 FAIL 0 CONVENTION 6`**.

## What you need to know before your next move

1. **The page's bytes moved.** sites `9d1a17202`;
   sha256 `c82013b8…` → `817d80c7…` (md5 `99628a97…` → new). The DB row
   `55682bc8-0113-4bf1-a10b-08aff6e8ea22` was repaired repo→DB with your own
   `gate_component_bytes.py --repair` (re-run shows GATE PASSES; sha256 stamp
   re-stamped and verified). Three-way identity (repo = DB = wire) holds again.
2. **`BASELINE_2026-08-05_stored_md5_at_b318a8fad.txt` was deliberately NOT
   edited** — it is a dated snapshot at a named ref; rewriting it would
   falsify history. Consequence: `load_lmc.py`'s pre-write guard will now
   correctly REFUSE on stamp-duty. That refusal is the guard working; when you
   next target the page, re-baseline consciously.
3. **Move `decompose_lmc.py`'s `PINNED_REF=b318a8fad` past sites `9d1a17202`
   before decomposing stamp-duty** — the pin predates the fix, and
   `repo_file()` reads `git show <PINNED_REF>:…`, so decomposing today would
   re-freeze the expired-rule bytes into a permanently locked tool row.
4. **The gemini/02 MORT_SRC copy is fixed too** (domains repo `c463764`), and
   `build_site.py`'s port of stamp-duty was proven byte-identical to the fixed
   sites file — a future build reproduces the fix exactly. NOTE:
   `build_site.py --check` currently crashes on the LOANS side
   (`loancalculator.co.uk/tools/standard-calc.html` no longer exists at
   LOAN_SRC) — pre-existing, unrelated to this fix, not touched; I ported
   stamp-duty alone through `port()` to get the assertion.
   **The domains repo now backs up to GitHub** (`gqls/domains`, private) after
   a log-stripped mirror rewrite — it had been 56 commits ahead of a codeberg
   remote nothing on this box can authenticate to, and 19 of its history blobs
   exceeded GitHub's 100MB limit. Read the `LANDMINES.md` entry before your
   next domains commit: **the local repo's SHAs no longer match the remote**,
   so a plain `git push` from `~/projects/domains` still fails.
5. **Your golden stays valid.** The three stamp-duty vectors sit outside both
   defect bands by design; the post-fix `golden_compare_post.py` output is
   byte-identical to the recorded pre-fix run (the 9 content-span lines on
   this non-decomposed page are the script's documented behaviour, not a
   regression).
6. **Your `--emit-criteria` gate is now OPEN**: your PLAN held emission until
   224 and 225 were both fixed. Both are live (224 by a concurrent session,
   `71ba7bb76`) and the full-estate oracle is green — emitting Tier-4
   `computed_values` fences no longer pins a wrong answer.
7. One-line fix applied to `oracles.py:176`: the comment read "(from
   2026-04-01)"; the rule is in force from **2025**-04-01 (value was always
   correct). Your own bug file warned against copying that typo into the JS —
   it is now gone at source.
