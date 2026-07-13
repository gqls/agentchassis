# Go batch — slice 4c: flag_page_image_rebuild handles section scope

Base: the on-disk `flag_page_image_rebuild_action.go`. Today the action no-ops for
anything non-page ("No-ops for non-page-scoped" — observed live when the two section
illustrations landed with no rebuild emitted). Section scope_refs carry the page as a
prefix (`index:1`), so the mapping is a prefix-split and a fall-through to the
existing page path — no new emit code.

Import check first: if `strings` is not in the file's import block, add it (a
permitted, noted addition).

OLD (exact, :105–110):
```go
	scope := inputs.Get("scope")
	pageName := inputs.Get("scope_ref")

	// Only page-scoped imagery triggers a page re-render.
	if scope != "page" || pageName == "" {
```
NEW:
```go
	scope := inputs.Get("scope")
	pageName := inputs.Get("scope_ref")

	// Section-scoped imagery (scope_ref = "<page>:<ordinal>") re-renders its page
	// too: plan_sections resolves section assets per page, so a landed section
	// image needs the same needs_page emit as a hero. Derive the page from the
	// scope_ref prefix and fall through to the page path.
	if scope == "section" && strings.Contains(pageName, ":") {
		pageName = strings.SplitN(pageName, ":", 2)[0]
		scope = "page"
		logger.Info("flag_page_image_rebuild: section-scoped imagery mapped to its page",
			zap.String("scope_ref", inputs.Get("scope_ref")),
			zap.String("page", pageName))
	}

	// Only page-scoped imagery triggers a page re-render.
	if scope != "page" || pageName == "" {
```
Cosmetic follow-ups (optional, not required to function): the file's header comment
(:18–19) and the image-build-handler step description in `agent_definitions` both say
page-only — one-line text updates when convenient.

Build: `go build ./...`; rides any image alongside slice 3.
