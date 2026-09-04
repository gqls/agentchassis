# RUNBOOK — tool fabrication fence (bugs_open/482)

Commands that were hard to get right, with the gotcha attached. Fix them HERE, not in scrollback.

---

## Probe the live fabrication gate against a real component

**The point of this recipe:** `DetectToolFabrication` is the exported *pure core* of
`check_tool_fabrication_action.go` — no DB, no `ActionParams` — which is why it can be driven
from a throwaway test. Everything else in that action needs an orchestration.

```bash
S=<scratchpad>
# 1. pull the real template bytes (NOT rendered_html — the template is what the gate sees at birth)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -tAc "SELECT html_template FROM content_components WHERE name='<component name>';" > $S/x.html

# 2. temporary probe in the actions package (it must be IN the package: the core is exported,
#    but the package is internal to the module, so an external test binary cannot import it)
cat > platform/orchestration/actions/zz482probe_test.go <<'EOF'
package actions
import ("os"; "testing")
func TestZZ482Probe(t *testing.T) {
    b, _ := os.ReadFile(os.Getenv("PROBE_DIR") + "/x.html")
    r := DetectToolFabrication(string(b), "", false)          // birth: no original, no analysis flag
    t.Logf("BIRTH fabricated=%v tier=%q signals=%v", r.Fabricated, r.Tier, r.Signals)
    // POSITIVE CONTROL — must convict, or the run proves nothing
    c := DetectToolFabrication(`<script>// we generate a large, realistic, deterministic dataset
    function makePostcode(){}</script>`, "", false)
    t.Logf("CONTROL fabricated=%v tier=%q", c.Fabricated, c.Tier)
}
EOF
PROBE_DIR=$S go test ./platform/orchestration/actions/ -run TestZZ482Probe -v
rm platform/orchestration/actions/zz482probe_test.go     # ⚠ shared tree — do not leave it
```

⚠ **The control is not optional and it is not ceremony.** Every one of the seven components I
probed returned `fabricated=false`. Without a control in the *same run* that result is
indistinguishable from a harness that is not calling the detector at all — and "all clean" is
exactly the answer a broken probe gives. The control convicting is what makes the zeros evidence.

⚠ **Read `Signals` even when `Fabricated` is false.** This is the whole finding of this lane:
three components return `Fabricated=false` **with `Signals` populated** (`large literal record
array (~30 entity objects)`). A probe that only prints the boolean reports "the gate is blind"
when the truth is "the gate detected it and discarded the verdict". Print both, always.

⚠ **Feed it `html_template`, not `rendered_html`.** The birth gate inspects the generated HTML
before it is stored and rendered. Probing `rendered_html` measures a later artefact and can
differ (instance-scoping rewrites ids).

⚠ **`zz` prefix and a unique name.** Many sessions share this tree. A probe called
`fabrication_test.go` will collide with somebody's real work; `zz482probe_test.go` will not.
Delete it before you commit anything — `git status` the directory and expect to see only other
sessions' files.

---

## Census the tool corpus WITHOUT encoding your own instance

⚠ **Read the WRONG_CALLS entry for 2026-09-04 before writing a census predicate here.** Keying on
date shapes returns **1 of 335** and reads as "first occurrence"; the worst real instance
(`tool-vet-comparison-vetcomparison-uk`, 30 invented practices) **contains no date at all**.

The population:
```sql
SELECT count(*) FROM content_components WHERE is_active AND component_level='tool';
-- 335  [MEASURED 2026-09-04]
```

The calibration denominator (all dataset-bearing tools, record counts + key sets), maintained by
the 427 lane — **use the file, not a remembered figure**; its author corrected 133 → 134 within
the hour after a case-folding difference:
```
docs/agent_docs/docs024_key_docs_latest/bugfix_427_event_render/CENSUS_2026-09-04_tool_embedded_datasets_calibration_set.txt
```
⚠ Neither column in that file is a verdict. The wide count is mostly legitimate UI vocabulary
(glossary terms, option labels). **It is a false-positive denominator, not a finding.**

---

## Establish whether a component is actually SERVING (not just active)

`is_active` on `content_components` is not deployment. Join through to the page:

```sql
SELECT s.domain, p.url, p.deployed_at, pc.build_status
FROM content_components cc
JOIN page_components pc ON pc.component_id = cc.id
JOIN pages p            ON p.id  = pc.page_id
JOIN sites s            ON s.id  = p.site_id
WHERE cc.name = '<component name>';
```
⚠ `build_status='deployed'` + a `deployed_at` is what makes a finding *live damage* rather than a
latent one. The vetcomparison hit is `/index.html` — a homepage, not a buried tool page — and
that is the difference between "file it" and "message the lane now".

---

## Which agents wire a given action

```sql
SELECT type FROM agent_definitions
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND default_config::text ILIKE '%check_tool_fabrication%';
-- tool-recreation-handler   ← the ONLY one  [MEASURED 2026-09-04]
```
⚠ Always carry the three predicates `is_active AND NOT is_snapshot AND deleted_at IS NULL`.
Snapshot rows are historical copies; counting them turns "wired on one agent" into "wired on
several" and inverts the finding.

To list an agent's workflow steps without dumping a 5 KB prompt into your context:
```sql
SELECT jsonb_object_keys(default_config->'workflow'->'steps')
FROM agent_definitions WHERE type='tool-generator'
  AND is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL;
```
⚠ A `SELECT` of the whole `default_config` on these rows is large enough to be unhelpful; and
`collected_data::text ILIKE '%…%'` over `orchestration_states` **times out** (>120 s) — it is a
full scan of a very large JSONB column. Query the component and the agent definition instead;
they answer the same question in milliseconds.

---

## Which write paths can change a tool's template

```bash
grep -rln "content_components" platform/orchestration/actions/*.go \
  | xargs grep -ln "INSERT INTO content_components\|UPDATE content_components"
```
`[MEASURED 2026-09-04]` the authorship paths are:
- `create_tool_component_action.go` — **birth and regeneration** (`regenerateToolComponentInPlace`
  is called from *inside* it at `:307`, i.e. **after** its `HasToolDocHeader` `:127` and
  `componentTemplateValid` `:160` gates, so both arms are covered by anything added before `:307`);
- `update_component_html_action.go` — tool-improver; calls `sharedComponentWriteCheck`;
- `tool-recreation-handler`'s workflow — the one path that *does* have the fabrication gate;
- `deploy_tool_action.go` — forks an existing row to another site (**propagation, not authorship**
  — but note it propagates a fabrication to a second site without re-inspecting it).

⚠ The existing coverage ratchet for the *other* fence is
`component_template_writer_coverage_test.go`. Read its header before mirroring it: it reads
SOURCE, so it proves a call **exists**, not that it **executes**, and it strips comments so that
naming the fence in a doc comment cannot satisfy it (`LANDMINES.md`: "a source-scanning test makes
your COMMENTS load-bearing").
