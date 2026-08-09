# PLAN — `bugs_open/232`: published Gauntlet rounds are search-engine indexable

**Started 2026-08-09.** Lane opened by a session picking the bug up from the filing
lane (`provocation_pipeline` / `gauntlet_dead_cta`), which split it out of
`architecture_review/RFC_020` precisely so it could be fixed independently.

## The problem, in one paragraph

`vonc.com/tools/gauntlet/round.html?r=<slug>` serves a visitor's own prose plus an AI
verdict at a permanent public URL. Nothing tells search engines to stay away. Low
severity today; high the moment a stranger publishes a round about a named real person,
because then the words are returned by a search for that person's name. Discoverability
is the harm multiplier — and removing it costs **nothing** in reach, since a shared link
behaves exactly as before. That asymmetry is the whole case for doing it now and cheaply.

## Design, and the two decisions that shaped it

### Decision 1 — the bug file's own preferred fix was not available, and finding that out was most of the work

The file ranks `X-Robots-Tag` on the route first. **For the HTML page that cannot be
done from this repo.** Measured: the page is a **B2 object behind Cloudflare**
(`x-amz-request-id`, `x-amz-version-id`, `server: cloudflare`), and the only Caddy
configs in-tree front **tools-api**, a different service on a different host. A header
on the page would be a Cloudflare dashboard rule — outside the repo, outside review,
unrebuildable, and squarely against the framework rule. So candidate 1 survives only for
the JSON endpoint.

Candidate 2 (a meta tag in the component) is warned against because stored component
content gets regenerated away. Right instinct, wrong mechanism **here**: the
round-record component is a body/style/script **fragment with no `<head>` in it**, so it
cannot hold the tag at all. That left one place the tag can come from: **page assembly**.

### Decision 2 — a typed opt-in column, gated at the call site

Owner ruling 2026-08-02 §2 governs: new authority on a shared seam ships as an **opt-in
field with the unsafe default OFF**. `assemblePage` is exactly such a seam — every page
rerender fleet-wide. So:

- `pages.noindex boolean NOT NULL DEFAULT false`, **not** a `page_spec` jsonb key. Two
  states, the unsafe one unrepresentable by omission; visible to `\d pages`; exactly
  censusable. A jsonb key has three states (absent/null/false), every reader
  re-implements the coalescing, and its spelling is unguessable.
- The gate is `if page.Noindex` **at the call site in `assemblePage`**, not inside the
  helper — so a reviewer of the *caller* sees the decision, which is the entire point of
  the ruling.

**Scope call, stated so it can be disputed:** additive-and-inert, therefore the normal
council gate and **not** an RFC (owner ruling 2026-07-29 §1 — an RFC is for a change to
what a shared mechanism *guarantees*). The capability is unreachable until a row opts
in; 1 of 630 does. Submitted with that argument explicit in `risks` so reviewers could
reject it. `Council-Submitted: 1139cbbe-3173-4886-846b-c25daeeda93c`.

### A judgement inside the helper worth its own line

`injectRobotsNoindex` is idempotent on **its own exact tag**, deliberately not on any
`name="robots"` match. Identical behaviour today (zero robots metas fleet-wide);
different **failure** modes. A loose marker lets a future permissive `index, follow` in
the shared chrome silently disable this — wrong result, looks exactly right. The exact
marker emits alongside instead, and crawlers take the most restrictive of multiple
directives, so coexistence still noindexes. Fails safe rather than fails silent. Pinned
by a test, and mutation-proven: swapping in the loose marker fails that test specifically.

## Phasing

| # | phase | state |
|---|---|---|
| 1 | Migration 352 applied (guard induced first) | **DONE 2026-08-09**, row-verified |
| 2 | Go: `PageInfo.Noindex`, `getPageInfo`, `injectRobotsNoindex`, call-site gate | **DONE**, committed `c3d7841f9` |
| 3 | Tests, mutation-proven both ways | **DONE**, green at committed HEAD |
| 4 | tools-api `X-Robots-Tag` (separate deploy) | code **DONE**; **NOT deployed** |
| 5 | Register SEO-003 + landmine | **DONE**, same commit / synced |
| 6 | Council verdict read + actioned | **OWED** |
| 7 | Chassis roll → pod-verify → re-render → verify at artefact | **OWED** |
| 8 | tools-api island deploy → verify at endpoint | **OWED** (owner-adjacent) |

**Ordering, and why it is not merely tidiness.** The migration went in **before the
commit existed at HEAD**. The new `getPageInfo` names `p.noindex`, so a new binary
against an un-migrated DB fails **every rerender fleet-wide**. On this tree I cannot
hold code back — any session's `make build-*` from HEAD ships it at a time I do not
control — so DB-leads-commit is the only ordering available. The reverse gap (flag set,
old binary) is inert.

## Correction to the originating brief

The plan drafted for this lane proposed placing the new step and noted the comment
scheme as "5a2./5b."-style, to be read at implementation time. It also inherited, from
`inject_canonical_link_test.go`'s header, the belief that `assemblePage` is the **single
live assembly path**. **That is false** — three active agent types dispatch the
`assemble_page` action into the other producer. Corrected before writing any code, and
it changed the shape of the deliverable: what would have been "a fix" is now "a fix plus
a documented boundary on what the fix covers". See NOTES.

## Explicitly out of scope

**Converging the two page-`<head>` producers.** `AssemblePageAction`
(`multipage_actions.go`) honours none of the four head injections. That is architecture
scope, it predates this bug, and folding it in would be exactly the scope creep the
guardian seat vetoes. Registered (SEO-003 open question) and landmined instead.
