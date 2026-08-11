# HANDOFF — `bugfix_153_build_provenance`, 2026-08-11 night · **lane closed, nothing owed**

Supersedes `HANDOFF_2026-08-11_continue_here.md`. Read this one.

## 1. Headline — the one thing that was owed is now done

The prior handoff's entire remaining task was: *"after the next `make release`, run RUNBOOK R9b +
R9b(ii)."* That release happened (`v1.0.1287`) and the check ran. It **discriminated** — a real
commit from another session (`d80fbf4bf`) landed inside the 6m39s build window, and all 14 backend
images still came out on one revision (`9b7811d4b`, the commit the pin resolved once at the
start). Under the pre-fix behaviour that would have produced ≥2 revisions, the way `v1.0.1284`
did. It did not.

```
153  BLD-019  stamp in every backend binary + image     LIVE, 14/14, three rolls
249  BLD-020  release pins ONE commit for the sweep      PROVEN, 14/14, discriminating window
```

**`bugs_open/249` is closed** (stays in `bugs_open/` per owner ruling 2026-08-06 — the closure
evidence is a dated banner inside the file, not a move). **BLD-020's register entry now reads
PROVEN**, not *exercised*. Full working: `NOTES_build_provenance.md`, dated entry same night.
Plain-prose account for the owner: `README_where_we_are.md`, same date. Milestone read-out for
the whole lane: `SUMMARY_2026-08-11_build_provenance.md` (new — first one this lane has written).

## 2. There is no section 2

Everything that HANDOFF_2026-08-11 §2–§6 asked for or warned about is either done or still true
as written (the traps in its §4 don't expire — they're about instruments, not about this
measurement). Read that file if you want the full trap list; nothing here contradicts it. This
file exists only to stop a future session re-running R9b/R9b(ii) or re-opening `249` thinking it's
still owed.

## 3. If you land here with a NEW build-provenance question

- The stamp mechanism (BLD-019) and the release pin (BLD-020) are both register entries in
  `docs/agent_docs/docs026_concept_register/register/build-pipeline.md` — read those first, they
  carry the live status.
- The three labelled-but-unstamped images (`component-render-check`, `shared-output-fields-check`,
  `removed-config-keys-check`) are still exactly that — nobody has fixed them, and they belong to
  other lanes. Don't edit their dockerfiles from here.
- The council-scope observation (release path is shared infra but outside `platform/`/`internal/`/
  `pkg/` review scope) is recorded, not actioned. Still just one lane's observation, not a rate.
