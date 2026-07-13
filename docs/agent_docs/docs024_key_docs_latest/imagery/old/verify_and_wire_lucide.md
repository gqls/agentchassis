# Verify & wire the Lucide icon validator

## 1. Verify the allowlist against YOUR bundled Lucide version

`lucide_icons.go` ships a curated list of long-stable Lucide names, but Lucide
renames/deprecates icons across versions. Confirm every allowlisted name
actually exists in the version robot-hands.com (and your other sites) load,
then prune any that don't.

First find the version you bundle. The `features` template uses
`<i data-lucide="...">`, so a `lucide` (vanilla JS) script is loaded somewhere —
check the site `<head>`/footer component or your build for the lucide version
(CDN URL pin or npm dependency).

Dump the canonical kebab-case names from that version. Most reliable source is
the `lucide-static` package (ships kebab-named SVGs):

```bash
npm i -D lucide-static@<your-version>   # match the version you actually ship
node -e '
const fs=require("fs"),path=require("path");
const dir=path.dirname(require.resolve("lucide-static/package.json"))+"/icons";
console.log(fs.readdirSync(dir).filter(f=>f.endsWith(".svg")).map(f=>f.replace(/\.svg$/,"")).join("\n"));
' | sort > lucide_valid_names.txt
wc -l lucide_valid_names.txt
```

Then check the Go allowlist against it — list any allowlisted names NOT present
in your bundle (these are the ones to remove or replace):

```bash
# extract the allowlist names from the Go file
grep -oE '"[a-z0-9-]+": \{\}' lucide_icons.go | grep -oE '[a-z0-9-]+' > allowlist_names.txt
sort -u allowlist_names.txt -o allowlist_names.txt

# names in the allowlist that your bundled Lucide does NOT have:
comm -23 allowlist_names.txt lucide_valid_names.txt
```

If `comm` prints nothing, the allowlist is fully valid for your version — done.
If it prints names, delete those lines from `lucideAllowlist` in `lucide_icons.go`
(or swap them for a near-equivalent that does exist). The render-time fallback
covers you either way, but a clean allowlist means the LLM is never steered
toward a dead name.

(Optional) regenerate a broader allowlist straight from `lucide_valid_names.txt`
if you'd rather permit the full set than a curated subset — but a curated list
gives the LLM better, more consistent choices and guarantees visual coherence.

## 2. Wire the validator into content generation

Two placements; do both for belt-and-suspenders.

### (a) Constrain the LLM — prompt amendment
Wherever the `features` (and other icon-bearing) component content is generated
by an LLM filling the component `input_schema`, inject the allowlist into the
prompt so the model can only pick valid names. Add a line like:

> The `icon` field for each feature MUST be one of these exact Lucide icon
> names (kebab-case, choose the closest match by meaning):
> {{ comma-joined output of AllowedLucideIcons() }}
> Do not invent icon names. If none fit well, use "circle".

Build the list at runtime from `AllowedLucideIcons()` so the prompt and the
validator never drift.

### (b) Validate before store — code
After the content LLM returns and before writing to `page_components.content_data`,
sanitize:

```go
replaced := content.SanitizeFeatureIcons(contentData)
if replaced > 0 {
    logger.Warn("Replaced invalid Lucide icon names with fallback",
        zap.Int("count", replaced),
        zap.String("fallback", content.LucideFallback))
}
// ...then persist contentData to page_components as today
```

`SanitizeFeatureIcons` only rewrites a PRESENT-but-INVALID `icon`; a missing
icon is left alone (the template guards with `{{if .icon}}`).

### (c) Optional render-time safety net
If you want defence regardless of which step produced the content, call the same
validation in the page assembler as it expands `{{range .features}}` — validate
`.icon` just before emitting `<i data-lucide="...">`. This catches legacy
content_data already stored with bad names (e.g. rows written before this fix)
without re-running content generation.

## 3. To fix EXISTING stored content (robot-hands.com)

The deployed features content already has valid names (database, sliders, …) so
no back-fix is needed there. But to sweep any site for bad stored names without
re-generating content, you can sanitize in-place in the DB-bound struct on the
next rerender (placement (c) handles this automatically), or run a one-off that
loads each `features` content_data, applies `SanitizeFeatureIcons`, and writes
back only where it changed something. Not needed for robot-hands.com today.

## Where I need input to finalize

I don't have the code/prompt for the step that fills `features` content_data
(it's NOT build-site-planner — that does structure + the imagery block; this is
a separate content-generation step). Point me at it and I'll write the exact
prompt amendment and the `SanitizeFeatureIcons` call site as a diff, rather than
the generic guidance above.
