# DIRECTION LEDGER — the blessed direction documents

The constitution and the mission are the platform's **fixed direction** (owner
ruling, 2026-07-20). This ledger is the drift reference for the three guards:

- **D2** — `.githooks/commit-msg` blocks any commit touching a path listed here
  unless the commit message carries a `Direction-Approved: <name>` trailer. The
  trailer is earned by the owner's explicit word for that specific change; adding
  it without that word defeats the guard and leaves a false record.
- **D3** — `fixloop_eg_dartsonline/100_CHECK_direction_integrity.py` recomputes
  these hashes, compares every sanctioned copy against its canonical, and
  needle-checks the constitution/mission council-seat prompts in BOTH councils.
- Updating this ledger **is** what owner sign-off writes down: change the doc,
  update its hash here, same commit, `Direction-Approved:` trailer on it.
  (This file guards itself: it is on the D2 blessed list.)

## Blessed documents

| doc | canonical path | sha256 (16) | approved | by |
|---|---|---|---|---|
| Constitution (thin slice) | `docs/agent_docs/docs024_key_docs_latest/adoption/docubundle/thin_slice_constitution.md` | `18453e8cac84bdfe` | 2026-07-20 | uk (owner) |
| Platform mission (doc 028) | `docs/agent_docs/docs024_key_docs_latest/028_platform_mission_and_pipeline_direction(2).md` | `c6aa949edab8e44a` | 2026-07-20 | uk (owner) |

## Sanctioned copies (must stay byte-identical to their canonical)

| copy of | path |
|---|---|
| Constitution | `scripts/documentation_project/02/thin_slice_constitution.md` |
| Constitution | `docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/go_files/contextkit/thin_slice_constitution.md` |
| Mission | `docs/agent_docs/docs014_documentation_collection/028_platform_mission_and_pipeline_direction(2).md` |

All copies verified byte-identical to canonical at blessing time (2026-07-20).

## Notes

- **Canonical-choice caveat (honest gap):** `CTXK-004` says the contextkit
  assembler pastes the constitution into every bundle, from `cmd/assembler` —
  but **no `cmd/assembler` exists in this repo** (checked 2026-07-20; the
  contextkit Go files live under `docs019 .../go_files/contextkit/`, which is why
  that copy is sanctioned). The docs024 adoption/docubundle copy is blessed as
  canonical because it is the newest actively-referenced home. If the assembler
  is found to load a different path, re-bless that path here — with the trailer.
- The council-seat prompts (fix-proposer + council-gate: `review_constitution`,
  `review_mission`) carry digests of these documents. They are DB rows, not
  files — the D2 hook cannot see them; `100_CHECK` needle-checks them instead.
- Longer-term (D4, planned): the constitution becomes `standards` rows
  (`scope='constitution'`) seeded FROM the canonical file; this ledger and the
  guards stay, the copies disappear.
