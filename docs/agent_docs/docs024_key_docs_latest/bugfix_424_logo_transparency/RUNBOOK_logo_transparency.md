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

## Build + roll (not yet done)

Affects the `image-generator-adapter` service (`internal/adapters/imagegenerator/`) — that is the
service to rebuild, not `agent-chassis` (the prompt-policy half lives in `platform/orchestration/
actions`, which is compiled into `agent-chassis`, so BOTH services need a rebuild + roll for the
full mechanism to be live). Bump `IMAGE_TAG` for both. Verify via build provenance / binary probe
per CLAUDE.md, not a same-tag rebuild assumption.
