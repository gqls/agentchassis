# Probe harnesses for bugs_open/260

Three standalone Go programs that produced this lane's measurements. Each carries controls that
must fire — they `panic` if the control comes out wrong, so a clean run is evidence rather than
an absence of output.

**They need two JSON dumps, not committed here (~13MB). Regenerate:**

```bash
# components.json — every component, active or not
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
  "SELECT COALESCE(jsonb_agg(jsonb_build_object('id',id,'name',name,'function',function,
    'is_active',is_active,'html_template',html_template)),'[]'::jsonb) FROM content_components;" > components.json

# sections.json — every stored section joined to its component template
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -tAc \
  "SELECT COALESCE(jsonb_agg(jsonb_build_object('pc_id',pc.id,'cname',cc.name,'cfunc',cc.function,
    'tmpl',cc.html_template,'cd',pc.content_data)),'[]'::jsonb)
   FROM page_components pc JOIN content_components cc ON cc.id=pc.component_id;" > sections.json
```

⚠ **The ~2MB+ stream through `kubectl exec` truncates intermittently** (the same flake
`cmd/component-render-check` documents). Validate the JSON parses and retry up to 3× — a
truncated dump silently measures a subset.

| probe | question | result 2026-08-19 |
|---|---|---|
| `parseprobe.go components.json` | do any live templates fail `Parse`? | **0 of 304 (251 active)** |
| `execprobe.go sections.json` | would any stored section fail `Execute` on rerender? | **0 of 1,778** |
| `contactprobe.go components.json` | would the 13th seam erase the live contact block? | **latent — 1 active component, renders clean** |

`parseprobe` hardcodes the seven FuncMap names from `executeGoTemplate`. **They were extracted
mechanically, not typed** — an undefined function is itself a parse error, so a wrong name set
would manufacture failures:

```bash
sed -n '/func executeGoTemplate/,/}).Parse(templateStr)/p' platform/orchestration/actions/call_agent.go \
  | grep -oE '^\t\t\t"[a-zA-Z]+":' | tr -d '\t":' | sort
```

**Re-run that before trusting `parseprobe` again** — it asserts a count of 7 and will panic if the
list grows, but a *renamed* function would slip through.

`contactprobe` deliberately uses **no FuncMap and no `missingkey=zero`**, because that is exactly
what `RenderTemplateWithMap` does. Its `{{safe .x}}` control proves the probe is faithful to that
absence — if that control stops failing, the probe is no longer testing the right seam.
