# FUTURE: Data Graph / Chart Pipeline

Status: **scoping only — not built.** Captured 2026-05-20 during the imagery
provider work. This is a distinct pipeline from image generation and must NOT
reuse the Banana/SDXL image path.

## The hard constraint

Image-generation models (Banana/Gemini-image, SDXL, any diffusion model)
**cannot plot real data.** Asked for "historic oil & gas prices", they produce
something that *looks* like a chart — plausible axes, a trend line — but the
values are fabricated, tick labels are garbled, and the series bears no relation
to reality. This is fundamental, not a quality gap. Diffusion models render the
*appearance* of a chart, not the chart.

Therefore graphs are a **code-rendered** artifact with an optional LLM editorial
layer. The data must come from a real source and be plotted by a real charting
library. The LLM never touches the data values.

## Architecture (three stages)

### 1. Fetch data (code, not generation)
Pull real series from an API or dataset. For energy prices:
- **EIA** (US Energy Information Administration) — free API, historical oil/gas
  prices, production, etc. https://www.eia.gov/opendata/
- **FRED** (Federal Reserve Economic Data) — free API, broad economic series
  incl. commodity prices.
- Generic: any REST/CSV source the site's topic needs.

Output: a clean typed series (dates + values + units + source attribution).
Store the raw series so the chart is reproducible and the source is citable.

### 2. Render chart (code)
Feed the real series to a charting library. Options, in rough order of fit:
- **go-echarts** (Go, in-chassis) — produces real interactive HTML/JS charts
  (Apache ECharts under the hood). Stays in the Go agent stack, no extra runtime.
  Best fit if the deliverable is an embeddable interactive chart.
- **Python + Plotly** (separate step/pod) — richest annotation/interactivity,
  exports static PNG/SVG or interactive HTML. Needs a Python runtime in the
  agent image or a dedicated chart-render pod.
- **Python + matplotlib** — simplest static PNG/SVG output; least interactive.

Recommendation: **go-echarts for interactive embeds, Plotly if static annotated
images are wanted.** Decide based on whether the site embeds live charts (HTML)
or drops in a rendered image (PNG/SVG).

Output: chart artifact (HTML for embed, or PNG/SVG to the assets pipeline — note
PNG/SVG could then reuse the existing asset-deployer/git-commit path that the
image pipeline already uses).

### 3. Editorial annotation layer (LLM — optional, value-add)
This is where Claude (already wired in the planner as claude-opus-4-6) earns its
place. Given the *accurate* chart + the underlying series, the LLM decides the
narrative:
- which points to annotate ("2008 spike", "2020 negative WTI", "2022 war premium")
- callout text and positioning
- title, subtitle, takeaway caption
- what to emphasise for the site's audience

The LLM outputs annotation specs (point coordinates + text); the charting code
draws them. **The LLM proposes the story; the code owns the numbers.** It must
never be asked to produce or adjust the data values.

## Agent shape (when built)

A new agent, e.g. `data-chart-generator` (adapter or orchestrator), with a
workflow roughly:

    fetch_series        (action: fetch from EIA/FRED/CSV → typed series)
      → render_chart    (action: go-echarts/plotly → chart artifact)
      → annotate (opt)  (action: execute_llm_prompt → annotation spec)
      → redraw (opt)    (apply annotations to chart)
      → store_asset     (reuse existing asset pipeline for PNG/SVG output)
      → complete

The planner's imagery block has `kind` ∈ {logo, hero, illustration, icon,
infographic}. A graph is NOT one of these — it would need a new kind
(`chart` or `graph`) AND routing to this new agent, NOT to the image-generator.
The planner prompt would need a new section describing when to request a chart
and what data source/series to specify. The DB CHECK constraint on `kind` would
need the new value added.

## Relationship to "infographic" kind

The existing `infographic` kind goes to Banana and produces *stylised decorative*
infographics — fine for a good-looking process diagram, NOT for data accuracy.
Keep that distinction explicit:
- `infographic` (Banana) = decorative, illustrative, no real data
- `chart`/`graph` (this pipeline) = real data, code-rendered, accurate

Do not let the planner emit data-bearing graphs as `infographic` — they'll be
fabricated. If a planned "infographic" actually needs real numbers, it's a
`chart` and belongs here.

## Why this is deferred

The image pipeline (icons/heroes/illustrations) is the current priority and is
now working. The graph pipeline is net-new (new data-fetch integrations, new
charting runtime decision, new agent, new `kind`, planner changes, CHECK
constraint change). It's a clean future workstream, not an extension of the
imagery work.
