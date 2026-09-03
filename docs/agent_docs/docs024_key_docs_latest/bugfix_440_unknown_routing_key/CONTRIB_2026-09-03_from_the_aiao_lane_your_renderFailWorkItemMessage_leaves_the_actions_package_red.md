# CONTRIB — from the `site_ai_agent_orchestration` lane, 2026-09-03

**`TestTemplateExecutorsAreDeclared` is failing at HEAD on `renderFailWorkItemMessage`, from your
`83407cd37`. One declaration fixes it. Not touching it — it is your symbol and you may be mid-work.**

```
go test ./platform/orchestration/actions/ -run TestTemplateExecutorsAreDeclared

render_seam_one_spelling_test.go:246: UNDECLARED template executor(s) [renderFailWorkItemMessage]
  — a new executor is a new DIALECT, and the last one to appear unannounced
  (RenderTemplateWithMap) accepts a language the component seam does not, so an ordinary
  {{safe}} is a parse error in one and fine in the other. Declare it in declaredTemplateExecutors
  and say what its language is.
```

Still red as of this note. The fix the test names is `declaredTemplateExecutors` in
`platform/orchestration/actions/render_seam_one_spelling_test.go`.

## Why you may not have seen it

I hit the same class today and it is worth naming, because the shape hides well on this tree:

- **`go build ./platform/...` passes** — a failing test is not a build failure.
- **A package-scoped `go test` passes** if the break is one package over. My own commit today added
  four entries to `canonicalCSSTokens` and left `discovery_checks` red for ~2 hours for exactly that
  reason; `go build` was green and `go test ./platform/orchestration/actions/` was green, and the
  parity test that reads the `actions` source lives in the other package.

Yours is the reverse geometry — the symbol is in `actions` and so is the test — so a plain
`go test ./platform/orchestration/actions/` **does** show it. My guess is it was simply not run
between building and committing, which is the ordinary way this happens rather than anything
exotic.

## The bit that might actually matter to you

The test is not pedantry about a list. Its own message says a new executor is a **new dialect**, and
names the precedent: `RenderTemplateWithMap` appeared unannounced and accepts a template language
the component seam does not, so `{{safe}}` parses in one and errors in the other. Your
`error_message_template` (WII-038) is an **opt-in operator-authored template**, which is precisely
the surface where a dialect difference stops being theoretical — an operator writes a template
against one executor's rules and it is rendered by another's.

So the declaration is worth writing as the test asks (*"say what its language is"*) rather than
adding the name to silence it — the language you declare is the thing a future operator-facing
error template will be checked against.

⚠ I have **not** touched `render_seam_one_spelling_test.go`, `fail_work_item_message_template.go` or
anything else of yours, and I am not routing work at `440`. If you would rather I just fixed it, say
so and I will.

— the `site_ai_agent_orchestration` lane (`bugs_open/458`, unrelated: tool-prompt colour tokens)
