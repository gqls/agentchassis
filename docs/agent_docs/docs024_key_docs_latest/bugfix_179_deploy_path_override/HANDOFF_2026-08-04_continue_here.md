# HANDOFF — 2026-08-04 — session "bugfix 100": 179 CLOSED, 116 handed back owner-gated, 170 contributed

Cold start for a fresh thread. **Everything below is committed. Nothing is
half-applied.** Read this, then `NOTES_deploy_path_override.md` (newest at the
bottom) if you are picking up the deploy-path lane.

---

## 1. What this session did, in one paragraph

Took the next `bugs_open/` bugs no other thread held. **`bugs_open/179` finding A is
FIXED, council-APPROVED, LIVE on `v1.0.1250`, induced on the live fleet, and CLOSED**
(moved to `bugs_closed/`). **`bugs_open/116` was handed back OPEN and unfixed on
purpose** — every one of its four fix candidates is owner-gated or forbidden by
written policy, and that finding is the deliverable. **`bugs_open/170`** (another
lane's) was verified LIVE and contributed into, but deliberately not closed.

## 2. State

| thing | state |
|---|---|
| `bugs_open/179` finding A | **CLOSED.** Fixed, live `v1.0.1250` both replicas, refusal induced, moved to `bugs_closed/`. Council `7435c263-…` APPROVED round 2 |
| `bugs_open/116` | **OPEN, owner-gated.** No code written, deliberately. Needs an owner decision, not a fix |
| `bugs_open/170` | **OPEN**, but now known LIVE — steps 1/3/4 of its own verification PASS; only its behavioural induction (step 2) is owed. Whoever induces it closes it |
| seed 307 | **APPLIED** 2026-08-04, snapshot `e9a9bac9`. `asset-deployer` no longer declares `deploy_path` |
| `IMAGE_TAG` | fleet now on **`v1.0.1251`** (re-verified: both 179 controls still hold — a fresh roll can ship an older commit, so this was checked, not assumed) |
| diagnosis loop (116) | filed, run corr `54bf4506-5192-4528-8395-eb2c636a7fad` — **verdict never read** (see §5) |

## 3. `bugs_open/179` — what shipped, and the one thing NOT proven

The `deploy_path` input used to replace the derived `AssetPaths` outright, publishing
files at paths no reader could derive. **Deleted**, not gated; an *explicit*
`deploy_path` draws a refusal-as-result before the download and the git commit; the
input is undeclared from the spec and pruned from the live row; and
`TestAssetPathsAreOnlyConstructedInStorage` bans hand-built `AssetPaths` tree-wide so
the class cannot return.

**The finding that shaped it:** `ExtractActionInputs` resolves *declared* fields by a
**depth-20 recursive search of the whole of `collected_data`**, so a stray
`deploy_path` in any nested step result armed the override **without a caller asking**.
That is why deletion rather than a flag, and why the refusal is wired to explicit
sources only — refusing on the deep search would be a false denial of legitimate
deploys fleet-wide.

**Proven live** (A/B differing in one variable, both with a bogus `s3_uri` so neither
could commit): A with `deploy_path` → `deployed:false, skipped:true, reason "refused:
deploy_path …"`; B without → `FAILED` at `deploy_asset`, *"storage client not
available"*. B proves the guard is not a blanket refusal **and** that it precedes the
storage resolution. Neither probe committed (both paths 404).

> **[NOT DONE] A successful end-to-end deploy was never induced.** B failed at storage
> rather than deploying, so *"a legitimate deploy still works"* rests on the unchanged
> code path, the unit tests, and B showing the guard is not what stopped it. If you
> want that closed properly it needs a valid `s3_uri` and will commit an image to a
> live site — decide whose site before doing it.

## 4. ⚠ READ THIS BEFORE YOU QUOTE ANY `deploy_path` CENSUS — mine was an artefact

`jsonb::text` renders **with a space after the colon**, so
`LIKE '%"deploy_path":"%'` **cannot match a jsonb column at all**. My
three-population "nobody uses this" census was therefore structurally zero, and it
reached the council submission (APPROVED), the IMG-067 register entry, migration 307
and four commit messages before being caught — by accident, when the same query could
not see a probe I had written minutes earlier.

**The conclusion survives re-measurement** (`site_work_items` 0, active
`agent_definitions` 0, `orchestration_states` 1 = my own probe) with:

```sql
WHERE collected_data::text ~ '"deploy_path"\s*:\s*"[^"]+"'
```

**The rule: induce a non-zero before you trust a zero.** Written up in
`WRONG_CALLS.md` and given a footprinted `LANDMINES.md` entry (2026-08-04), because
it applies to every `jsonb::text LIKE` census in the estate. The broken pattern is
**recommended in writing** in `bugs_closed/179`'s own Evidence section (corrected
there) and copied into `bugfix_168_deployed_asset_path`'s handoff — **that copy is
still uncorrected; correcting it is a loose end.**

## 4b. OWNER DECISIONS 2026-08-04 (evening) — all seven questions answered, and what was executed

| # | decision | state |
|---|---|---|
| D1 | **Improvement loop stays OFF for now.** 204 findings stay parked; cadence stays manual | Recorded in `bugs_open/116` |
| D2 | **Per-build checks steer DEFERRED** while D1 stands — sessions should stop attempting it | Recorded in `bugs_open/116` |
| D3 | **Induce a successful deploy** | **DONE.** Probe D (spawn+call, spawned-pod tag v1.0.1251): robot-hands hero redeployed to its own derived path, proven at the artefact (200, new bytes). Evidence in `bugs_closed/179`. Bonus finding: the INLINE orchestrate path has NO storage client — only spawned asset-deployers can deploy; and `deploy_result` is empty even on success (git-adapter response overwrites the output_field) — verify at the served path, never the status |
| D4 | **Build one page on finetuning.uk** to close `bugs_open/170` step 2 — authorised | **NOT YET RUN** — execution plan in that file's CONTRIB + authorisation block. Assert the header source marker takes the pool branch |
| D5 | Cloudflare permissions question | Answered in the session reply; summarised below |
| D6 | **robot-hands copy change AUTHORISED** (`bugs_open/147` candidate 1) | **NOT YET RUN** — exact execution steps now in that file's OWNER DECISION block: edit `content_data` not `rendered_html`, check `input_schema` static fields, rerender, verify ON THE WIRE, re-run claimscan expecting 0 BANNED / 2 negated |
| D7 | **Restart vet collection** | **EXECUTED.** `vet-batch-verify` enabled 2026-08-04 ~20:00Z after verifying preconditions (spawned-pod tags v1.0.1251 ≥ the 1151 floor; the provenance prerequisite IS bug 100's fix, live). `ch-vet-collect` + `vet-sweep-continue` deliberately left off. First tick 20:01:55Z. **Successor: run bug 100's two-column acceptance once fresh `data_observations` rows land** (newest was still 2026-03-18 at 20:05Z), then close 100 |

**D5 in one block, for whoever executes `bugs_open/132`:** a scoped Cloudflare API
token (never the Global API Key), with: **Account → Workers Scripts → Edit** (deploy
the 404 fix; Read alone suffices to EXPORT the live worker's source into the repo,
which is the more urgent half — it is currently under no version control) ·
**Zone → Workers Routes → Edit** (only if the route/binding must change — the worker
exists, so likely not needed) · **Zone → Zone → Read** (zone enumeration). Scope it to
the affected zones or the account as the owner prefers, supply it as
`CLOUDFLARE_API_TOKEN` + the account id, and `wrangler` (or the raw API) can pull and
deploy. First act should be the EXPORT, committed to the repo, before any edit.

## 5. Loose ends, in priority order

1. **`bugs_open/116` needs an owner decision, not code.** All four candidates are
   blocked: per-build detection is forbidden by IMP-016 and by
   `validate_page_content.go:644-650`'s own precedent while nothing drains the
   `detected` queue (**204 detected across 10 sites vs 2 triaged**); seating the
   checks elsewhere is warned against by `bugs_open/149:395-398`; and re-enabling the
   improvement loop reverses the owner ruling of 2026-07-29 (it is **G1**, an explicit
   separate owner go). The real question is the 204 parked findings. Full evidence in
   the bug's STATUS block and `bugfix_116_link_check_coverage/`.
2. ~~The 116 diagnosis-loop verdict was never read~~ **READ 2026-08-04 (evening):
   UNVERIFIABLE** — "NOT confirmed (stopped: scope-not-narrowing) … no fix
   proposed". No refutation, no corroboration; the mechanism claim stands on the
   declared first-hand verification (recorded in the 116 lane NOTES).
3. ~~Correct the broken census pattern in the 168 handoff~~ **DONE 2026-08-04
   (evening)** — dated correction in place, crediting the origin correctly.
4. **`bugs_open/170` step 2** — the behavioural induction. Everything else on its
   checklist passes; evidence is a dated CONTRIB in that file (`a0e011723`).
5. **`bugs_open/147`** was surveyed and deliberately NOT taken: its fix is a copy
   change on `robot-hands.com`, and the filing session declined on the grounds that
   rewriting another lane's site voice is theirs to do. Still true.
6. **`bugs_open/132`** is blocked, not unowned: the deployed Cloudflare worker's
   source is in no repo and there is no deploy path from this tree.

## 6. Traps this session hit, so you do not

- **A source-scanning test makes your COMMENTS load-bearing.** Ordering assertions use
  `strings.Index` — the *first* occurrence. Naming the guarded call in a doc comment
  *above* the guard fails the test on ordering, and the message reads like a
  code-ordering bug. Bit me twice; both were the test working.
- **A disabled `scheduled_tasks` row shows a fresh `last_completed_at`**, because
  `improvement-loop`'s `notify_scheduler` stamps `improvement-sweep` **by name** on
  every completion, however it was dispatched. Read `enabled` + `last_triggered_at`.
  (LANDMINES, 2026-08-04.)
- **Counting findings cannot measure detector coverage** — a site with no findings is
  either clean or unexamined. The record of the *act* is
  `sites.settings->'maintenance_profile'->'last_audit'`. (WRONG_CALLS, 2026-08-04.)
- **A migration verify block of `SELECT`s cannot stop the `COMMIT`.** Use `DO`/`RAISE`
  — and induce it, as 307 did.
- `who-owns.py` reads commits and is blind to a session mid-fix. The decisive check
  for 179 was the *filing lane's own handoff* saying "OPEN, unowned", plus a grep of
  live transcripts.

## 7. Commits from this session

`d06502e73` claim(179) · `fd0516b18` fix(179) code+register · `f62265138` config(179)
seed 307 · `13194d96d` docs(179 lane) · `6f69fd757` IMAGE_TAG 1249 · `aea3a49c7`
approved(179) · `59ae3a59f` **close(179)** · `e2925c9e9` + `3bf81fc00` docs+correction
· `0c633d91f` handback(116) · `a0e011723` contrib(170)
