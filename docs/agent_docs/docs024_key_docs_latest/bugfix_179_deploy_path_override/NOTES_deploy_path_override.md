# NOTES — bugs_open/179 finding A, the `deploy_path` override

Append-only, newest at the bottom. Evidence, commands, what the system actually
said, and every misstep.

---

## 2026-08-04 — taking finding A

Session "bugfix 100", second bug of the session (the first, `bugs_open/116`, was
handed back unfixed and owner-gated — see `bugfix_116_link_check_coverage/`).

**Ownership checked three ways before starting**, because `who-owns.py` names the
filing lane (`bugfix_168_deployed_asset_path`, ACTIVE, 22 commits/14d, 25 mentions)
and would read as OWNED:

1. `who-owns.py 179` → filing lane, active.
2. **Their own newest handoff settles it**: `bugfix_168_deployed_asset_path/HANDOFF_2026-08-03b_continue_here.md:34`
   — `| bugs_open/179 | **OPEN**, unowned. Unchanged. |`, and their next work is
   Decision 2's dedup half and RFC_010 Q1, neither of which is this.
3. Live transcript scan across every session touched in the last 150 minutes: no
   session mentioning 179.

So: contribute-don't-compete does not apply — it is genuinely free. Claimed in the
bug file and committed (`d06502e73`) before writing code, so the claim is visible
to `who-owns` and to a transcript grep.

## The finding that changed the fix

The bug file frames the risk as *a caller might set `deploy_path`*, and measures it
empty three ways. **That census cannot see the real exposure.**

`ExtractActionInputs` resolves every declared field by a **depth-20 recursive
search of the whole of `collected_data`**
(`platform/orchestration/datahelpers/unified_extractor.go:440-489`), and
`deploy_path` was a **declared optional input** on the live `asset-deployer` row:

```
 input_fields      ["s3_uri", "deploy_path", "purpose", "domain", "asset_key"]
 contract optional ["deploy_path", "purpose", "asset_key"]
```

So a `deploy_path` key **anywhere** in a deploy orchestration — a nested sub-agent
response, an echoed spec — was hunted out, bound, and used to redirect the git
commit. **The caller never had to ask for it.** A census of *values callers set*
returns zero and stays zero while that is true.

This decided two things:

- **The fix is deletion, not gating.** An input nobody can be trusted *not* to
  supply accidentally is not made safe by a flag.
- **The refusal must NOT be wired to `inputs.Get`.** Refusing on the recursive
  hunt would convert a stray nested key into a **false denial** of a legitimate
  deploy, fleet-wide. Explicit sources only (step config, the deprecated
  `deploy_path_field`, `input_data.deploy_path`); anything only the deep search
  can find is **ignored**, and the derived path wins.

## Why the bug file's own higher-ranked candidate was rejected

Candidate 2 was *"route `deploy_path` through the derivation instead of around it:
keep the override but require it to be recorded where readers can see it"*. The
action already records what it wrote (`deploy_image_asset_action.go:296-325`,
`UPDATE assets SET filename = …, url = …`), conditional on `asset_id`.

**Measured: no reader reads it.** All six resolve `(asset_key, purpose)` through
the derivation — `emit_sprite_css_action.go:136`, `plan_sections_action.go:304-423`,
`derive_card_asset_action.go:204`, `render_site_components_action.go:415`,
`check_image_url_404.go:565`, `queryresolve.go:301-304`. So a *recorded* override
is still a path no reader resolves; recording it changes the audit trail, not the
defect. Making a reader prefer a recorded path is `bugs_open/152`/`155`'s seam and
belongs to those lanes.

## MISSTEP — my own comment broke my own ordering test, and the test was right

The new sensor asserts the refusal precedes `DownloadOptimizeAndPrepare`,
`sendGitCommitRequest` and `storage.DeployedAssetPath(`, anchored on
`strings.Index` — the **first** occurrence of each token.

I wrote a helpful doc comment above the function saying the output path is
`storage.DeployedAssetPath(asset_key, purpose)`. That put the token at line ~52,
before the guard at ~225, so the ordering assertion failed **on a comment**.

Fixed by rewording the comment to name the function without its call syntax, and
by recording the trap in the action's own doc comment and in the test, because the
next person to explain this code will reach for exactly the same sentence.

Then the same shape again, one level on: assertion 4 (*"the guard must not read
`inputs.Get("deploy_path")`"*) fired because the guard's **own comment** names that
call in order to say it is deliberately not used. A sensor that cannot distinguish
code from prose forbids the code explaining itself. Fixed by scanning non-comment
lines only — the same `commentOnly` treatment the sibling derivation sensor already
uses for its anti-vacuity arm.

**Transferable:** a source-scanning test makes comments load-bearing. Both failures
were the test working; neither was a false positive.

## Mutation results — six, each isolating a different sensor

Run one at a time, restoring between (the tree is shared; the window was seconds,
and the files were verified residue-free afterwards).

| mutation | what failed | what it proves |
|---|---|---|
| guard condition never true (`deployPath == "zzz-never-matches"`) | **only** the 3 behavioural subtests | the behavioural test is load-bearing independently of the source scan |
| reason literal renamed (`refused:` → `declined:`) | source scan **and** behavioural | the literal is the anchor for both |
| `_ = storage.DeployedAssetPath("x","y")` inserted **above** the guard | **only** the ordering assertion | ordering is pinned independently of existence |
| guard rewired to `inputs.Get("deploy_path")` | **only** the negative control (+ assertion 4) | the false-denial regression is caught |
| `storage.AssetPaths{}` added to `derive_card_asset_action.go` | **only** the tree-wide storage test | the class ban works on files this lane never touches |
| one refusal arm (`deploy_path_field`) deleted | **only** that subtest | the three arms are pinned separately |

The third is the one worth keeping: it isolates ordering from existence, which a
"delete the guard" mutation cannot — that fails everything at once and proves less.

## Blast radius, measured before submitting

**Value census** (JSON shape `'%"deploy_path":"%'`, *not* the bare word — the bare
word returns 2 definitions and 93 orchestrations, all declarations and this lane's
own council submissions):

| population | values |
|---|---|
| open work-item specs | **0** |
| active agent definitions | **0** |
| all orchestration history | **0** |
| `deploy_path_field` (definitions / orchestrations) | **0 / 0** |

**Declaration census:** only `asset-deployer` declared it.
`image-build-handler`'s bare-word match is its step **description** — its
`input_mapping` passes no `deploy_path`, checked directly rather than inferred.

**Source census:** exactly **one** `storage.AssetPaths{` construction outside
`platform/storage` in the entire tree — the block deleted here. That is what made
the tree-wide ban cheap and total.

## What shipped

- Code + register: `fd0516b18` (`Council-Submitted: 7435c263-…`).
- Config: `f62265138` — seed **307 applied**, snapshot `e9a9bac9` taken first;
  044 (canonical) updated so a fresh apply cannot reintroduce the declaration.
- The migration verifies with `DO`/`RAISE`, not `SELECT`s — a verify block of
  `SELECT`s cannot stop the `COMMIT` (`ON_ERROR_STOP` ignores a non-empty result).
  **And it was induced**: an inverted copy of the block was run against the
  post-apply row and raised as designed, so the silent pass is a measurement rather
  than an artefact of a block that can never fire.
