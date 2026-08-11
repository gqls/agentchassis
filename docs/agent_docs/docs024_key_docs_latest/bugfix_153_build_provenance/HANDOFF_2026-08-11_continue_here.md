# HANDOFF — `bugfix_153_build_provenance`, 2026-08-11 evening · **start here**

Supersedes `HANDOFF_2026-08-10_continue_here.md` (banner added there). Read this one.

## 1. Headline — the lane is DONE except for one conditional check

`bugs_open/153` (no build provenance) is **fixed and live on every backend service**, proven
across three separate rolls now. Its follow-on `bugs_open/249` (one tag, several revisions) is
**fixed, committed and exercised**, and closes the moment a release straddles a commit.

```
153  BLD-019  stamp in every backend binary + image     LIVE, 14/14, three rolls
249  BLD-020  release pins ONE commit for the sweep     COMMITTED, exercised, not yet PROVEN
```

Both files stay in `bugs_open/` (owner ruling 2026-08-06).

## 2. THE ONE THING STILL OWED — and it costs nothing, you just have to wait for it

**After the next `make release`, run RUNBOOK R9b + R9b(ii).** That is the entire remaining task.

`v1.0.1286` was the first release under the pin and came out **coherent** — one revision
(`c3b424c8e`) across 14/14, confirmed at the image labels and at the pods. **It is NOT proof.**
Zero commits landed inside its 5m52s build window (the nearest was 10 seconds *before* the first
build, which is what the release pinned to), so the old unpinned code would have produced the
same result. The measurement could not have come out otherwise.

**The close condition:** a release whose build window CONTAINS at least one commit, still shipping
one revision. Then `bugs_open/249` closes and BLD-020 goes from *exercised* to *proven*.

```bash
# R9b — one revision across the 14 backend images?
for S in auth-service core-manager agent-chassis reasoning-agent web-search-adapter \
         web-scrape-adapter git-adapter image-generator-adapter thunder-adapter \
         analyser-adapter browser-runner-adapter content-creator-agent \
         remote-job-spawner kafka-scheduler; do
  docker image inspect docker.io/aqls/$S:$IMAGE_TAG \
    --format '{{index .Config.Labels "org.opencontainers.image.created"}} '"$S"' {{index .Config.Labels "org.opencontainers.image.revision"}}'
done | sort                                    # revisions must be identical

# R9b(ii) — COULD it have come out otherwise? Take the window from the labels above.
git log --format='%h %ad %s' --date=format:'%H:%M:%S' --since='<first created>' --until='<last created>'
#   commits inside + one revision  -> PROVEN. Close 249.
#   no commits inside              -> coherent only. Say so. Wait for the next one.
```

> **Timezone trap:** image labels are UTC (`…Z`); `git log %ad` prints local (BST = UTC+1).
> Convert, or the window will look an hour off and you will find "no commits" every time.

Expect this to resolve quickly — the busiest 7-minute window on 2026-08-11 held **13** commits,
and a release is ~6 minutes.

## 3. What was decided, so nobody re-opens it

Four owner decisions, 2026-08-11:

1. **Pin the ref** — built as `pinned_sweep` in the makefile, BLD-020.
2. **Local regression guard instead of the production induced-fault test** — run and passed
   (RUNBOOK R9a). **R6 is closed by decision, not forgotten**; its section says so. Do not
   re-file it as an outstanding proof.
3. **CLAUDE.md's build section rewritten** — marker-hunting and `strings` retired, replaced with
   "ask the service what it is running" + `git merge-base --is-ancestor`.
4. **Builds and deploys stay manual (owner-driven).** So: no refusal mechanism in `push-*` /
   `deploy-*` (153's candidates 2 and 3 stay unbuilt), and **this lane does not run releases.**

Also settled earlier: 153's council round was REJECTED on **scope**, the owner overruled it, and
**no `Council-Reviewed:` trailer exists or may be added** on any of its commits.

## 4. ⚠ Traps this lane paid for — read before verifying anything

- **Never `strings` in a container probe.** Absent from the debian-slim image
  (`browser-runner-adapter`); behind the customary `2>/dev/null` its failure is identical to
  "not stamped". Never a *discovery* grep either — it matches Go's internal digit table and
  returns the same wrong answer on every service. Verify a **known** value:
  `kubectl -n ai-persona-system exec <pod> -- grep -aq "<sha>" /proc/1/exe`.
- **The control that discriminates is a REAL BUT DIFFERENT commit**, not a fabricated sha. A fake
  sha proves only that your grep is not promiscuous. Two of this lane's three false readings
  would have survived a fake-sha control.
- **The startup provenance line SCROLLS.** It is emitted once at boot; on `agent-chassis` it is
  already past `--tail=3000` within hours. An empty result there means *not in range*, not
  *unstamped* — fall back to the binary probe, which has no shelf life.
- **A labelled image is not a stamped binary.** `ref_build` applies the OCI label at the docker
  CLI, so *every* image it builds carries provenance even when the dockerfile was never edited.
  Three images are labelled-but-unstamped today: `component-render-check`,
  `shared-output-fields-check`, and (new, 2026-08-11) `removed-config-keys-check`.
- **Once a probe has fooled you, its next TRUE negative looks like more blindness.** Eight
  services correctly read `NO MATCH` on 08-11 and were nearly dismissed as instrument failure,
  because the same instrument had genuinely been at fault the day before. Only a control on the
  same target tells "absent" from "I could not look".
- **`grep -c` exits 1 on zero matches**, so `cmd && grep -c X f || echo "missing file"` prints
  *"missing file"* for a file that exists. This lane wrote that bug into its own one-liner one
  day after filing the landmine about the same shape.

All of these are in `LANDMINES.md` (synced to `doc_notes`) and `WRONG_CALLS.md`.

## 5. Where everything lives

| what | where |
|---|---|
| the two bugs | `bugs_open/153_…image_tag_does_not_imply_a_rebuild.md`, `bugs_open/249_…one_release_tag_ships_three_source_revisions.md` |
| the mechanisms | register `build-pipeline.md` **BLD-019** (stamp), **BLD-020** (pin) + index rows |
| the commands | `RUNBOOK_build_provenance.md` — **R9a** guard (run), **R9b/R9b(ii)** the owed check, R6 superseded |
| the technical log | `NOTES_build_provenance.md` (append-only; five missteps recorded) |
| the owner's log | `README_where_we_are.md` (plain prose, append-only) |
| the refused submission | `SUBMISSION_2026-08-11_release_ref_pin.json` — **written and REFUSED, never judged** |

**Commits:** `4a116e094` (file 249) · `21b9772a9` (the pin + CLAUDE.md) · `8cc6f649c` (index row) ·
`e1bb7c9eb` (the 1286 verification). Earlier 153 work: `e743e6cfc`, `e5f31dcdb`, `1054ec36c`,
`8d270c68a`, `c4a932680`, `060bfae62`.

## 6. Open threads that are NOT this lane's to close

- **The council gate does not review the makefile** — scope is `platform/`, `internal/`, `pkg/`
  (owner ruling 2026-07-17), so the release path, arguably the most shared mechanism on the
  estate, is outside review while a one-line `pkg/` change draws a full round. **Observed, not
  proposed.** One instance is not a rate; do not open an RFC on this alone.
- **`removed-config-keys-check` is labelled but unstamped** (§4). Another lane owns it and was
  mid-council-round on 2026-08-11 (`a9237f0c9`). Contribute into their work; do not edit their
  dockerfile from here.
- **153's candidates 2 and 3** (make push/deploy *refuse* an unbuilt or mismatched tag) remain
  deliberately unbuilt — they change the push/deploy contract fleet-wide and the owner has chosen
  manual control. BLD-019 and BLD-020 are their prerequisite if they are ever revisited.
- **BLD-020's companion assertion** (warn when images at one tag disagree on revision) was offered
  and not taken. Recorded as a decision in BLD-020 so its absence does not read as an oversight.
