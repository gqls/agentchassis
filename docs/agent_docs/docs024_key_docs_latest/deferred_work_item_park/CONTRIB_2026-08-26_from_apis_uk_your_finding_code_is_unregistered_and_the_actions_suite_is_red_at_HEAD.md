# CONTRIB 2026-08-26 — from the apis.uk session: `WORK_ITEM_STATUS_OVERRIDE_REFUSED` is unregistered, and the whole actions test suite is red at clean HEAD because of it

Found while running `verify-head-builds.sh --test ./platform/orchestration/actions/` for an
unrelated change (per-section subjects), 2026-08-26 morning.

**What fails:** `TestFindingCodeScanEveryWriteIsRegistered` (`findingcodes_scan_test.go:284`), at
COMMITTED HEAD, for every checkout:

> error code "WORK_ITEM_STATUS_OVERRIDE_REFUSED" is written by this package, is not declared in
> docs/agent_docs/docs024_key_docs_latest/architecture_review/finding_code_registry.json, and is
> not in `_scan_baseline` — so it is NEW. Declare it … in the same commit that adds it.

**Whose:** commit `2b46afbe6` (2026-08-25 20:48, "396: an allow-list for status_override") added
`statusOverrideRefusedCode = "WORK_ITEM_STATUS_OVERRIDE_REFUSED"` in `work_items_common.go:202`.
No registry declaration in that commit, in HEAD since, or in the index/working tree when checked
(2026-08-26 ~09:30). The commit-time advisory prints exactly this, but it prints FIRST and a
`| tail` on the commit output eats it — I nearly lost the same advisory on my own commit today.

**What it costs while open:** every session's full actions-package run reads red, which teaches
people to stop believing the suite (my own session had to spend a round proving the red predated
my change). The category call (consumed / instrumented / human-evidence / operational / `unruled`)
is yours — `unruled` is legitimate if genuinely open, and one line ends the fleet-wide red.

Nothing else is owed; the allow-list change itself is not being questioned.
