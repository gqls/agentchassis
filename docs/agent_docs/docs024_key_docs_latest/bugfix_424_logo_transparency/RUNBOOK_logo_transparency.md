# RUNBOOK — bugfix_424_logo_transparency

## Run the tests

```bash
go test ./internal/adapters/imagegenerator/... ./platform/orchestration/actions/... \
  ./platform/orchestration/actions/discovery_checks/... -v
```

Just the new coverage:

```bash
go test ./internal/adapters/imagegenerator/... -run TestKeyOutBackground -v
go test ./platform/orchestration/actions/... ./platform/orchestration/actions/discovery_checks/... \
  -run "LogoBackground|BackgroundKey" -v
```

`gofmt -l` the touched files before committing — a `var` block with mixed comment widths reformats
silently and will otherwise ship unformatted.

## Verifying the served asset — the check that actually catches this bug

A viewer cannot tell a painted checkerboard from real alpha by looking (a viewer draws a
checkerboard FOR real transparency too). Check the PNG bytes, not a screenshot, and check BOTH
signals — colour type AND `tRNS` — never one alone (a palette-transparent PNG carries `tRNS` at
colour type 3, which a colour-type-6-only check would misreport as still broken):

```python
import struct
with open('logo.png', 'rb') as f:
    data = f.read()
assert data[:8] == b'\x89PNG\r\n\x1a\n'
w, h, bitdepth, colourtype = struct.unpack('>IIBB', data[16:26])
has_trns = b'tRNS' in data
print(f"{w}x{h} depth={bitdepth} colour_type={colourtype} tRNS={has_trns}")
# FIXED: colour_type in (4, 6) OR has_trns True.
# STILL BROKEN (this bug's own shape): colour_type == 2, has_trns False.
```

**Probe the correct slug, not the domain in the bug title.** boxingonline.com is a PARKED
CATCH-ALL that 200s on any path (`[MEASURED 2026-09-02]`, boxingonline session) — the site actually
serves at `boxingonline.ugg2.com` (`sites.publish_target='b2worker'`). Check
`sites.publish_project` for the site in question before curling anything.

## Live baseline (interim fix, pre-this-fix), for comparison

`[MEASURED 2026-09-02]`, boxingonline.com logo asset: 139,777 bytes, 400×218, **bit depth 16**,
colour type 2, no `tRNS`. The bit depth matters — a 16-bit source decodes to `*image.RGBA64`, not
the 8-bit `*image.NRGBA` most of the unit suite uses — and is now covered directly:
`TestKeyOutBackground_16BitSourceDepth` builds a genuine `*image.RGBA64` test image and passes.
Still worth confirming once against the real regenerated asset rather than trusting the synthetic
case alone, since a real generation's 16-bit values won't be exact `×0x101` multiples of a known
8-bit palette the way the test's are.

## DB queries used during this workstream

```sql
-- The asset row this bug is about, and its history-that-isn't (UPSERT in place).
SELECT a.id, a.purpose, a.mime_type, a.url, a.origin_model, a.created_at, a.updated_at,
       left(a.origin_prompt, 200) AS prompt_head
FROM assets a JOIN sites s ON s.id = a.site_id
WHERE s.domain ILIKE '%boxingonline%' AND a.purpose = 'logo'
ORDER BY a.updated_at DESC LIMIT 3;

-- The mime_type gap (bugs_open/433), fleet-wide:
SELECT coalesce(nullif(mime_type,''),'(EMPTY/NULL)'), count(*) FROM assets GROUP BY 1;
```

## Council submission

Submitted 2026-09-02 (`council_submission_424_logo_transparency.json`, 7 edits, 54,082 bytes,
`DRY_RUN=1` admission passed first):

```
SUBMISSION_CORR=d018a48f-bd76-420a-8530-4491681d3bd4
RUN_ORCH_ID=8bb44322-42f1-43bb-aac7-a15c113548e1
```

Verdict:
```sql
SELECT created_at, metadata->>'decision' FROM diagnosis_artifacts
WHERE correlation_id='d018a48f-bd76-420a-8530-4491681d3bd4' AND kind='council_report' ORDER BY created_at;
```

Committing before it lands — use `Council-Submitted: d018a48f-bd76-420a-8530-4491681d3bd4`. If
APPROVED, a later reference should use `Council-Reviewed: <that id>` — do not write that trailer
without having read an approved verdict.

**Round 2** (the `BorderKeyed` measures-the-wrong-thing bug, found by live production testing):
`council_submission_424_round2_borderkeyed.json`. `SUBMISSION_CORR=52bd50a1-3783-4801-868a-31a0ee599e60`.
**Verdict: APPROVED, all reviewers, no objections** (read 21:21:07Z). The fix commit (`fcbe6071c`)
was made BEFORE this submission, so it carries neither trailer — the correlation lives here and in
the HANDOFF instead. If you're closing this lane out and want the historical record clean, a small
follow-up commit noting `Council-Reviewed: 52bd50a1-3783-4801-868a-31a0ee599e60` in its own message
(not amending `fcbe6071c` — forward-only) would let `098`'s report find it by grep even without
file-overlap resolution.

## Build + roll

Affects the `image-generator-adapter` service (`internal/adapters/imagegenerator/`) AND
`agent-chassis` (the prompt-policy half lives in `platform/orchestration/actions`, compiled into
`agent-chassis`) — **BOTH need a rebuild + roll** for the full mechanism to be live at any given
commit. Bump `IMAGE_TAG` for both. Verify via build provenance / binary probe per CLAUDE.md, never
by tag or a "roll happened" assumption.

**2026-09-02 status: a fresh build IS live (tag `v1.0.1354`) and DOES carry the original matting
fix (`6440ec968`), verified at the artefact on both services — but it does NOT carry the
magenta-contradiction fix (`b2322a203`), because that commit postdates the running build.** Full
verification commands, so the next session can re-run them rather than trust this note:

```bash
# image-generator-adapter — build provenance (works; not a busy service)
kubectl -n ai-persona-system logs -l app=image-generator-adapter --tail=3000 | grep -m1 'build provenance'
# -> read the git_commit value, then:
git merge-base --is-ancestor <this-fix-commit> <that-git_commit> && echo LIVE || echo NOT-LIVE

# agent-chassis — build provenance usually scrolls out of range (busy service); fall back to the
# binary probe, ALWAYS with a positive control (a long-merged symbol) and a negative control
# (a made-up symbol) in the same breath, never just the target alone:
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- grep -aq "applyLogoTextPolicy" /proc/1/exe && echo "positive control: PRESENT (expected)"
kubectl -n ai-persona-system exec "$POD" -- grep -aq "applyLogoBackgroundPolicy" /proc/1/exe && echo "424 matting fix: PRESENT" || echo "424 matting fix: ABSENT"
kubectl -n ai-persona-system exec "$POD" -- grep -aq "must use no shade of magenta or pink" /proc/1/exe && echo "magenta fix: PRESENT" || echo "magenta fix: ABSENT"
kubectl -n ai-persona-system exec "$POD" -- grep -aq "applyLogoBackgroundPolicyNOTREAL" /proc/1/exe && echo "unexpected: negative control PRESENT (probe is broken)" || echo "negative control: ABSENT (expected)"
```

**Do not trigger a real `kind=logo` generation against the currently-deployed build** — it will
hit the magenta/background contradiction the council caught (see NOTES, "council verdict read").
Roll again after `b2322a203` (and anything else that lands before the next roll) before testing on
a live asset.
