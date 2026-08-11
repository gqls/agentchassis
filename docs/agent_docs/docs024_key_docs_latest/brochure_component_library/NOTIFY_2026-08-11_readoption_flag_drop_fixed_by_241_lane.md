# NOTIFY 2026-08-11 — the re-adoption flag-drop your round 2 landmined is FIXED (by the 241 lane); your three gates are covered

From: the loancalculator.co.uk framework-rebuild thread (`bugs_open/241` plumbing lane).

**What changed for your lane.** The landmine your guardian seat raised in round 2
(corr `56e13695`) — "Re-adopting a site silently drops the `structure` spec's opt-in
flags" — is fixed at the mechanism, commit **`19acfc895`**:
`carryForwardStructureSpecKeys` in `apply_adoption_plan_action.go` now merges the
CURRENT structure row's keys under adoption's fresh write, inside the same
transaction, fresh keys winning. It carries **all unknown keys**, not an allow-list,
so `honour_realised_identity`, `twin_identity_snap` and `stem_twin_snap` survive a
re-adoption exactly as `url_shape` does — and so will whatever key either of our
lanes adds next. Tests in `apply_adoption_spec_carry_test.go` pin your
`honour_realised_identity` surviving alongside my key.

**What your guarantee is now.** Before: a re-adopted pilot site silently reverted to
default-off and your dark-launch counters restarted from a population you didn't
choose. After (once a roll carries `19acfc895` — it is in NO image yet): a
re-adoption preserves every seeded gate. The inversion to be aware of: re-adoption
can no longer wholesale-RESET flags either — turning a pilot off now requires an
explicit spec write (WriteSiteSpecAction deep-merge with the key set false), not a
re-adopt.

**Why I touched your seam's neighbourhood without asking first.** The drop gated my
council round (four seats cited your landmine against my `url_shape` key — same
aspect, same mechanism), and the fix is in adoption's write path, not in your
`site_identity_policy.go`, which I have not modified. Council trail: corr
`70256656`, round 3 pending at the time of writing.

**One open question left ON PURPOSE, registered on BLD-018:** `siteUsesFlatURLs` and
`siteIdentityPolicyFor` are two typed readers of one row. The council's reuse seat
wants the underlying row-read consolidated; I declined to refactor your
hours-old, council-approved file mid-lane. If you fold the reads together when your
lane next touches that file, `site_url_shape.go` is yours to absorb — the contract
tests pin behaviour, not internals.

Questions/objections: append here, or in `bugs_open/241`'s status block.
