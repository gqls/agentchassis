# NOTES — component CSS coupling (bugs_open/072)

Append-only, newest at the bottom. The missteps are the point.

---

## 2026-07-26 — session 1: diagnosis, fix, commit

### Coverage check first

`scripts/who-owns.py 072` warned on the ambiguous number (two unrelated 072 cases) and
named `model_directory_pipeline` as active. Read their docs: they filed this bug, fixed
their own `model-directory` instance, and **explicitly declined the news half as out of
remit**. The workstream the bug routes the news half at (`news_feed_pooling`) has been
parked behind an owner gate since 07-20. No open `site_work_items` row touches it. So the
lane was genuinely free.

### The symptom reproduced exactly

All four figures from the 07-25 filing reproduced 24 hours later, unchanged. Worth saying
because a symptom that has moved is a different bug.

### MISSTEP 1 — I nearly shipped a root cause the control refutes

My first story was clean and wrong in an interesting way: *the stylesheet is older than
the markup, therefore unstyled*. It fits ai-agent-orchestration (CSS 05-02, markup 07-21)
and relojistas (CSS 07-16, component 07-26). Then the control:

- **robot-hands: CSS 2026-07-20 20:18, markup 2026-07-21 09:04** — CSS written 13 hours
  *before* the markup, and it **is** styled.

So plain date ordering does not decide it. The resolution is that
`loadPagesWithComponents` merges **built** components (`page_components`) with **planned**
ones (`pages.sections`), so robot-hands had news *planned* before its design run and the
snippet matched on the plan. That refinement is [INFERRED] — I could not check it, because
the `orchestration_states` rows for those runs are pruned.

What saved this was running the control at all. The two failing sites alone would have
"confirmed" the wrong mechanism, and I would have written it into a handoff with dates
attached, which is exactly the shape that gets believed.

### MISSTEP 2 — the second story was also incomplete

Next story: the two stylesheets were written while the pre-2026-05-16 empty-list bug was
live. That explains ai-agent-orchestration (CSS 05-02, before the 05-16 fix) but **not**
relojistas (CSS 07-16, after it). Two sites, two different immediate causes, one class.

The lesson is that "I have an explanation that fits" kept being the trap. It fit; it
covered one site. The thing that actually held up is the class-level statement, which does
not depend on which immediate cause applied to which site: **the stylesheet is a frozen
point-in-time artefact and nothing refreshes it.**

### The measurement that reframed the fix

```
in-use component functions with their own <style>:  86
in-use component functions without:                  8
```

I had been treating "components ship their own CSS" as a convention to *introduce*. It is
already the platform's rule at 91% adoption — `latest-news` and `news-listing` are
stragglers. That changed the migration from a design decision into finishing an existing
one, and it is the single most useful number in this investigation.

### The constraint that decided the design

`rerender_single_page_action.go:1` — *"Simple concatenation - no template re-rendering"*.
So a `<style>` added to `html_template` does **not** reach a live page on a rerender; it
needs a full component re-render, which regenerates page content (bugs 038/050). Anything
injected at *assembly* time ships on the cheap assemble-only path. That is why the Go
injection is the class fix and the template `<style>` is the complement, not the reverse.

### MISSTEP 3 — the placement I would have chosen silently destroys the CSS

Appending the `<style>` block at the end of `html_template` is the obvious move and it is
wrong. `saveSectionsExtractFromHTML`'s regex is
`(<section>...</section>)((?:<style>...)*)((?:<script>...)*)` — style blocks are captured
**only ahead of** script blocks. `latest-news` carries its `<script src>` *after*
`</section>`, so a block appended at the end sits behind the script and is dropped.

I did not reason my way out of this; I ran the real regex over both post-migration
templates and over a deliberately-wrong control:

```
ln.html        stored=4566/4598  style_captured=true   css_present=true
ln_wrong.html  stored=1190/4598  style_captured=false  css_present=false
```

3,355 characters of CSS gone, no error, and the component row would have looked perfect in
the DB. This is the `output_tokens == max_tokens` shape all over again: the artefact is
fine, the pipeline eats it, the status says success.

### The tests were checked for vacuity

`TestInjectComponentCSS_*` passed first time, which is exactly when to distrust them.
Induced fault: changed the insertion index from `loc[0]` to `loc[1]` — a one-character
error that puts the block *after* `</head>`. Two tests failed with the right messages;
reverting made them pass. The idempotency test correctly did **not** fire, since it tests
a different property.

### Migration number collided three times in one afternoon

Written as 217; by the time I checked, 217/218/219/220 all existed; took 221; by the time
I saved, another session had taken 221 too. Shipped as 222. The ledger keys on filename so
a collision is survivable, but re-grep immediately before writing, not at planning time.

### Applied by hand, deliberately

`run-migrations.sh --apply` applies **every** pending file — 14 were pending, 13 belonging
to other threads. Applied 222 alone via `psql`, then `--record-only` with a note saying
what was checked. `UPDATE 1` + `UPDATE 1`, both post-write assertions passed,
`component_news_backup_20260726_222` holds the pre-state.

### State at end of session

Committed `7821ad7f5` (5 files, pathspec — the scope block confirmed nothing else rode
along). Council submission `75d1a2af-afb8-492d-9587-4aa13bc440a2`, queued behind 11
messages at submit time; no orchestration row yet, which is queue latency and **not** a
dropped dispatch.

`go build ./platform/... ./internal/...` clean. `go build ./cmd/...` fails in
`cmd/reasoningset/main.go` (declared-and-not-used) — that is another session's
uncommitted WIP, present in the tree at session start, not mine and not touched. Builds
come from `git archive HEAD`, so it cannot reach an image unless they commit it.

**072 stays OPEN.** The Go half is inert until an image roll; the migration is live in the
DB but invisible until sections are re-rendered; no page was re-rendered and no live site
was touched, on the owner's decision.
