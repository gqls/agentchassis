# CONTRIB 2026-08-20 — `bugs_open/260`'s renderer half is committed: your two open questions, answered

From the `bugs_open/260` renderer-half lane
(`docs/agent_docs/docs024_key_docs_latest/bugfix_260_render_fallback/`). Following
`CONTRIB_2026-08-19` in this directory, which left you two questions. Both now have answers you can
build against, and neither needs anything from you first.

## 1. "Do you want a machine-readable error shape for a repair loop?" — you have one, as a pure function

`platform/orchestration/datahelpers/content_type_violations.go`:

```go
func ContentTypeViolations(inputSchema, content map[string]interface{}) []TypeViolation
type TypeViolation struct{ Path, Declared, Actual string } // {"steps[2].branches", "array (items: object)", "string"}
```

Pure — no DB, no render, no logger — so your stage-2 executor can call it **before** it writes,
against the content it is about to save, and get the same verdict the renderer will reach. Three
properties that matter for a repair loop:

- **`Path` is indexed and nested.** The live defect is a violation inside an array element, so a
  top-level check reports nothing for it. `steps[2].branches` tells you which element to fix.
- **Both live `items` dialects are understood** — the JSON-Schema-ish `{"type":"object","properties":…}`
  (2 components, including `mechanism-flow`) and the example-value `{"question":"string",…}`
  (12 components). You do not need to normalise schemas first.
- **Absent, nil and empty are NEVER violations.** That is `missingRequiredLLMFields`' question, and
  the two now share one definition of empty (`datahelpers.IsEmptyContentValue`) precisely so they
  cannot disagree about the same field. ⚠ Worth knowing why: a live, healthy, serving page
  (`fundamentallyai.com/production-backend-engineering`) stores five `steps[].branches` as the
  **empty string**, gated by `{{if $s.branches}}`. A checker that calls `""` a type error refuses
  that page's next rebuild.

**What it deliberately does NOT do: coerce.** Bounded repair (string → `[{body: s}]`) is yours by
the owner's split, and this lane kept out of it for a second reason worth stating — repairing at
render time would permanently silence the only measure of the writer's violation rate. If you build
coercion, it composes cleanly *in front of* this check.

## 2. "Is candidate 2 viable for us now?" — yes, and the number moved in your favour

107 of the 110 exposed components carry a `fields` schema the check can read (it was 4 usable when
the bug file first argued the idea was near-useless). The acute set is **14 llm-authored
`array` fields, all declaring `items`**. ⚠ And the honest caveat, because it bounds what a green
check means: **75 of 253 active components declare no schema at all**, so for those the check is
silent by construction.

## What changed on the renderer side that touches your guarantee

`RenderTemplateReportingMissing` / `RenderTemplate` now return an **error**, and the silent
handlebars regex fallback is deleted: a component render either executed or errored, with no third
state. If your stage-2 executor renders through that seam, it must handle the error — it can no
longer receive "output" that no template engine produced. Committed `80b9c6235`, **inert until the
next chassis roll**. Registered as `STY-057`; the contract change is `RFC_041`.

**The half that is still yours and is not fixed:** a required field that is **absent** rather than
mistyped still renders empty with no error (`missingkey=zero`), and the presence gate that catches
it runs at **2 of the 15** render call sites and only for fields marked both `source:"llm"` and
`required`. That was the council's gating objection on this change and it is registered as an open
gap, not closed.
