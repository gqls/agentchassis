# NOTES — bugs_open/125 (append-only, newest at the bottom)

## 2026-07-31 (1) — how this bug was chosen, and the bug I chose FIRST and had to abandon

I was asked to take "the next bug in `bugs_open/` that no other thread is on". My first
pick was **143** (`derive_card_asset` commits before its lock check). Every ownership
check said unowned — `who-owns.py` → "(none identified)", the bug file quiet since
07-29, `git status` clean on `derive_card_asset_action.go`, and the one lane citing 143
had written in its own notes that it is "related, not duplicate" to their work.

**All four readings were correct and another session was 20 minutes into the same fix.**
We wrote a shared asset-lock helper into the same Go package in the same minute. I found
out only because the Edit tool told me the file had changed under me. I deleted my
helper, reverted my hunk, confirmed `go build ./platform/...` still clean on their code,
and contributed my three extra findings into `bugs_open/143` instead of competing.
Logged in `WRONG_CALLS.md`.

**The lesson that changed how I picked 125:** every ownership signal CLAUDE.md lists is
*lagging* — commits, bug-file history, dirty files. None of them exists in the window
between another session choosing a bug and its first Write, which for a research-first
session is 20+ minutes. The one leading signal is the other sessions' live transcripts,
and the discriminating form is to grep for the target's **code symbols**, not its bug
number — a bug number appears in every session that ran `ls bugs_open/`. I had actually
run that grep, seen hits, and dismissed them as filename noise without checking.

So for 125 I grepped `resolveFilePath` / `ensureHTMLExtension` / `git_deployer_actions`
across all live `.jsonl` transcripts: **one hit, my own session.** That is what "nobody
is on it" looks like when you ask the question properly.

## 2026-07-31 (2) — the bug is still live, and bigger than filed

Filed 07-28 as 280/431. Re-measured today: **316 / 472**. Grounding figures rather than
carrying them forward was worth it here — the denominator moved by 41 pages in three days.

## 2026-07-31 (3) — my first wrong call on this bug: the function name

The bug file says the defect is in `resolveFilePath` at `git_deployer_actions.go:414-445`.
**There is no `resolveFilePath` in this repo.** The function is `determinePageFilename`
at `:374`. The sweep handoff has the right name; the bug file does not. I nearly grepped
myself into believing the file had been refactored. Corrected in the bug file.

## 2026-07-31 (4) — the pre-work's "strip the fragment" is wrong, and dangerously so

The handoff's pre-work says the fix "must strip `#…`" for the one fragment-bearing URL.
I nearly implemented exactly that. Checking what the stripped value would BE stopped me:

```
idea.uk | tool-audience-check | /tools.html#audience-check
idea.uk | tools               | /tools.html            ← a DIFFERENT page owns it
```

Stripping the fragment makes a rebuild of `tool-audience-check` overwrite the `tools`
page's file. A URL with a fragment points *into* a page; it does not name one. So the
helper **declines** it and the caller falls back — sanitising a value into a valid-looking
path is how you aim a write at someone else's file.

**Transferable:** when a sanitiser's output lands in a namespace someone else occupies,
"make the input valid" and "make the input correct" are different operations, and only
the second is safe. Check what the sanitised value collides with before writing it.

## 2026-07-31 (5) — the leading slash the pre-work did not mention

`pages.url` is site-absolute on 472/472 rows. `CommitToRepo` does
`prefixedPath := data.Domain + "/" + path` (`internal/adapters/git/github_client.go:69`),
so passing the URL through unchanged gives `example.com//tools/x.html` — a `//` and an
empty segment in a GitHub tree path. Every existing path in this pipeline is
repo-relative (`assets/css/styles.css`). Found by reading the adapter rather than
assuming the resolver's output was the final word.

## 2026-07-31 (6) — my second wrong call: "determineFilename is dead code"

I grepped `determineFilename` while **excluding its own file** (I was filtering out the
definition line) and concluded it was an unreachable correct implementation — a lovely
story, and false. It has two live callers at `file_extractor.go:89` and `:104`
(`ExtractFiles` methods 3 and 4). Caught within a minute by re-running the grep without
the exclusion, before it reached any durable doc other than this one.

**The cheap check:** never let a "does anything call this?" grep exclude a path. Filter
the *output*, not the *input*.

## 2026-07-31 (7) — the finding that changed the fix shape

Grepping the **derivation** instead of the symptom found five places that turn a page
into a deploy path — and four of them already consult `url` first:

| site | url first? |
|---|---|
| `datahelpers/file_extractor.go:194` `determineFilename` | yes — *"Try url field first"* |
| `rerender_single_page_action.go:521` | yes |
| `get_pages_for_rerender_action.go:176` | yes |
| `rerender_pages_actions.go:324` | yes |
| **`git_deployer_actions.go:374` `determinePageFilename`** | **no — the bug** |

So this is not a missing feature, it is a **duplicated classifier that drifted**, and
016b §9 already rules on that shape. Two functions eleven characters apart in name, one
right and one wrong, and the wrong one is the one the three build pipelines reach.

That turned candidate 1 ("three lines here") into one shared definition plus five call
sites — the difference between closing the instance and closing the class. The four
"correct" copies were not identical either: none guarded the fragment, none handled a
directory-style URL, three would turn `/foo.php` into `foo.php.html`.

## 2026-07-31 (8) — blast radius measured BEFORE writing the code, not asked of the reviewer

Per the 2026-07-28 ruling ("measure the blast-radius claim before you submit"):

- 471 of 472 live URLs resolve **byte-identically** to today's rerender output.
- 0 pages named `index`/`home` with a non-`/index.html` URL ⇒ dropping that special case
  from the rerender copies is inert.
- 0 URLs with a query string, `..`, `//`, whitespace, or a multi-dot final segment.
- The one fragment row's current rerender output (`tools.html#audience-check.html`) is
  **404 on the live site**, so that copy's defect is latent too.

## 2026-07-31 (9) — submitted to the council gate

`SUBMISSION_CORR = 758f6e62-99b8-4f33-a81b-7143351ecd69`. Two schema notes, because
the RUNBOOK's own warning section is about a *different* shape than the trigger enforces:
`plan` is an **object** (`summary` / `edits` / `grounded_in` / `risks`), not a flat array
with `risks` and `grounded_in` as siblings of `rationale`. My first attempt was refused
client-side with `ERROR: .plan missing` — which is the good outcome (client-side, no
credits) but the error text does not say *what* is missing about it. The authority is the
**097 script header**, lines 22–40, not the RUNBOOK prose. Added to the RUNBOOK here.

## 2026-07-31 (10) — council round 1: REVISE, and the gating objection was right to be raised

12 reviewers, 5 abstained, 0 unreadable. Gating HIGH from `prior_art_librarian`, echoed as
medium by `bug_historian`, `reuse_agent`, `guardian` and `debug_historian` — **five seats,
one question**, and none of them could have known the answer without asking:

> LANDMINES.md names `getPageInfo` in `rerender_single_page_action.go` ("a page whose url is
> `/` deploys a file with NO NAME"), but the plan's edit 5 names `loadPageInfo`. Either they
> are the same function and the plan's description of current behaviour contradicts the
> landmine, or **there is a sixth copy of the classifier in the very file the plan touches**
> — which is exactly the failure mode this plan is fixing.

**They were right to ask and I was wrong in the submission.** `grep -n 'func getPageInfo\|func
loadPageInfo'` returns exactly one function: `getPageInfo` at `:493`. **There is no
`loadPageInfo` in this repo.** I invented the name — I read the file from an offset that
started below the `func` line and typed a plausible name into the `symbol` field without
scrolling up. The edited line, `:526`, is inside `getPageInfo`'s body, so the function the
landmine names *is* the one I changed and there is no sixth copy. But the submission asserted
a symbol I had never verified, in a plan whose entire argument is *"I counted the copies"*.

**The check that would have caught it costs one grep**, and I had run that grep for
*`TrimPrefix`* but never for the *function names I was writing down*. Rule: **the symbol
field of a submission is a claim like any other — grep it.**

### It also produced a correction to the landmine

Checking the landmine's mechanism rather than just its conclusion: it says the guard "keys on
the page being NAMED `index`, so an adopted or hand-made homepage under any other name
(`home`, `landing`, `main`) falls straight through it". **It does not.** The pre-change guard
was `p.URL == "/" || p.URL == "" || p.Name == "index"`, and the *first* clause catches the
root case whatever the page is called; the name clause was a redundant second catch. So the
NO-NAME outcome was not reachable by the route described. Conclusion right, footprint right,
mechanism wrong — corrected in place, crediting the round.

### And it surfaced a real distinction I had glossed

`reuse_agent` asked whether `adopt_verbatim.go`'s `urlToDeployPath` — the landmine's other
footprint — was the missed copy. It is not, and the reason matters: it is the **inverse**
function. It produces the value stored in `pages.url` and deliberately **keeps** the leading
slash (`/index.html`); mine produces a repo path and deliberately **strips** it. Two halves of
one invariant with opposite output contracts. My round-1 grep had classified it as "has a
comment about this" and moved on **without opening the function** — the same not-looking that
016b's 2026-07-19 correction is about. Written into LANDMINES.md so the next person to tidy
"these two URL helpers" does not merge them.

### What round 1 confirmed rather than challenged

Every load-bearing measurement was independently re-run by the seats' own read-only checks and
came back identical: 316/472, 0 index/home rows, the `/tools.html` fragment collision, the
three `page_field` agents, and 0 pages with `url = '/'`. So the round cost one resubmission and
bought: a corrected symbol, a corrected landmine, a documented inverse-function distinction,
four new tests, a specified post-deploy check, and the orphan gap logged on `bugs_open/098`
instead of absorbed. **A REVISE that finds a real error in the submission is the gate working.**
