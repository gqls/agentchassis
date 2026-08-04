# PLAN — bugs_open/179 finding A, the `deploy_path` override

**Opened** 2026-08-04 by session "bugfix 100". Design, phasing, decisions **and
their reasons**.

> **Process correction, recorded rather than tidied away:** this file was written
> *after* the code, not at the start. The lane's directory was created at claim
> time but the plan lived in a subagent's returned plan and in my head until the
> fix was committed. The standing-five rule exists because a doc written at the end
> loses the wrong turns — and it nearly did here: the two comment-versus-sensor
> missteps in NOTES were reconstructed from the test output rather than logged as
> they happened. NOTES was written contemporaneously; this was not.

---

## The problem

`deploy_image_asset` resolved the shared derivation
`storage.DeployedAssetPath(asset_key, purpose)` — the one function the writer and
all six readers go through since `bugs_closed/168` — and then let a `deploy_path`
input replace the result outright. Readers only ever see `(asset_key, purpose)`, so
a caller setting `deploy_path` published a file at a path **no reader can derive**:
168's writer/reader drift, reintroduced through a supported input.

## The decision that shaped everything: what the census could not see

`bugs_open/179` measured the risk set empty three ways and framed the risk as *a
caller might set this*. **The real exposure was larger and of a different kind.**

`ExtractActionInputs` resolves every **declared** field by a depth-20 recursive
search of the whole of `collected_data`, and `deploy_path` was declared optional on
the live `asset-deployer` row. So a `deploy_path` key anywhere in a deploy
orchestration was hunted out and bound — **the caller never had to ask.** A census
of values-callers-set returns zero and stays zero while that is true.

Two consequences, and they are the plan:

1. **Delete, don't gate.** An input that can be supplied by accident is not made
   safe by a flag guarding deliberate use.
2. **Refuse explicit intent only.** Wiring the refusal to `inputs.Get` would turn a
   stray nested key into a **false denial** of a legitimate deploy, fleet-wide. An
   over-strict guard is the worse of the two bugs. Explicit sources are the step's
   own config keys and `input_data.deploy_path`; anything only the deep search can
   find is ignored, and the derived path wins.

## Candidates, ranked by what makes the bad state unrepresentable

| # | candidate | verdict |
|---|---|---|
| 1 | **Delete the override + refuse explicit intent** | **CHOSEN.** No code path can commit to a caller-chosen path; pinned tree-wide by banning hand-built `AssetPaths` outside `platform/storage`. Explicit intent gets a signal rather than silence. |
| 2 | Delete silently | Same unrepresentability, no signal: the caller's page references *their* path, the file lands at the derived one, the deploy reports success. Rejected on signal, not safety. |
| 3 | Ignore with a warning | A log line nobody reads. Strictly dominated by 1. |
| 4 | *Record* the override so readers can see it (the file's own candidate 2) | **Rejected on a measurement the file did not have:** all six readers DERIVE the path and none reads the recorded asset row, so a recorded override is still a path no reader resolves. Making a reader prefer a recorded path is `bugs_open/152`/`155`'s seam. |
| 5 | Opt-in field, default OFF | See below — the RFC_010 ruling does not license it here. |
| 6 | Status quo | A measured-empty risk set with nothing keeping it empty, plus the accidental-arming path above. |

## The two owner rulings, decided explicitly rather than cited

- **RFC_010 (2026-08-02), opt-in fields.** It governs a seam whose widest branch is
  safe *iff callers behave* — "callers must all be X" becomes a field with the
  unsafe default OFF. **This branch is unsafe however disciplined the caller is**,
  because readers structurally cannot see the caller's choice, so no caller-side
  field can license it. A default-OFF gate here would also be born with zero
  exercisers — the "mechanism rotting unexercised" cost the owner named on
  2026-07-29 when declining to *require* such switches. The ruling's spirit (remove
  decisions a reviewer cannot see) points at deletion.
- **RFC vs council gate (2026-07-29 ruling 1).** This *completes* a guarantee
  IMG-067 already declares as intended and records as the known open gap; it does
  not change what the shared mechanism promises. Finding B — the same shape, same
  action — shipped through the council gate. Council gate it is, with the register
  corrected in the same commit (condition 2) and consumers named and told
  (ruling 3). **No ordering constraint claimed** — condition (1) is retired, and
  none exists: the input carries no value anywhere, so seed/image order is free.

## Phasing

1. Claim in the bug file, commit — makes ownership visible to `who-owns` and greps. ✅
2. Measure blast radius (values, declarations, tree-wide source). ✅
3. Code: delete override, add refusal, undeclare input, correct `url_helpers.go`'s
   "NOT CLOSED" note. ✅
4. Tests: source sensor, behavioural table + negative control, tree-wide class ban
   with its own anti-vacuity test. Each mutation-proven. ✅
5. Council submit **before/alongside** the commit; commit with
   `Council-Submitted:`. ✅ (`7435c263-…`)
6. Config: seed 307 applied + canonical 044 updated. ✅
7. **Roll, pod-verify with positive + negative controls, induce the refusal and the
   healthy control, then close.** ← outstanding

## What this deliberately does not do

- No reader is taught to prefer a recorded path (`152`/`155`).
- `resolveStorageURIFromAsset`'s purpose-not-`asset_id` defect is `bugs_open/155`.
- The `assets` row UPDATE at `:296-325` is untouched.
- No asset-lock check here — LANDMINES rules this action must not get one.
- No queue sweep: the census is zero, so there is nothing to repair.
