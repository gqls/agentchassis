# HANDOFF — `bugs_open/098` unpublish primitive · **START HERE** · 2026-08-03

Cold-start for the next session. Read this, then `NOTES_unpublish_primitive.md` (the
missteps — they are the expensive part) and `RUNBOOK_unpublish_primitive.md` (every
command, with its gotcha).

---

## STATE IN SIX LINES

| | |
|---|---|
| Bug | `bugs_open/098` — **STILL OPEN**, do not close |
| Resurrection fix | **LIVE**, chassis digest `sha256:5da2888…`, both replicas |
| Unpublish primitive | **BUILT + LIVE + VETOED** — council REJECTED, hard veto from `guardian` |
| Council correlation | `4a7f0877-4149-4431-97d4-318d093570a4` (rounds 1 and 2) |
| Routed to | `architecture_review/RFC_011_git_adapter_action_vocabulary_and_the_unpublish_verb.md` |
| Retraction performed | **NONE.** Population still 13; the live instance still serves 200 |

**Commits (all carry `Council-Submitted: 4a7f0877…`; the verdict is REJECTED, so `098`
will bucket them and that is correct — do NOT write `Council-Reviewed:`):**
`43c259f46` primitive · `5b66615d4` resurrection fix + graph audit · `6d2c3006b` bug
correction + seed · `8f73e7279` cousin fix · `89b6d7963` verdict + RFC 011.

---

## THE ONE THING THAT WILL BITE YOU FIRST

**Do not retract a page.** The owner approved retracting
`robot-hands.com/learning-center/index.html` — *before* the veto existed. That approval
was given on a different premise. Firing a vetoed capability at a live customer site is
the precise thing the veto is about, and the RFC has to be broken by a human first.

If the owner reaffirms after reading RFC 011, the retraction is one dispatch and the
RUNBOOK has it. **Its acceptance is two-part** — 404 immediately, *and* still 404 after the
next news refresh — because a single `curl` passed even before the resurrection fix.

## WHAT THE VETO ACTUALLY SAID

Not "this is wrong". The `architecture` seat endorsed the design in terms worth quoting
back: *"expressing delete-as-null-sha inside the existing CommitToRepo path is the right
reuse, inheriting retry, prefixing and atomicity for free rather than bolting on a
parallel write path. On that axis the plan is sound and I'd carry it."*

It vetoed **how it reached production**: `delete_file` is a new reserved verb on a shared
adapter's vocabulary, added inside a bug fix — `bugs_closed/124`'s shape. The allowlist
edit is the sharp end: `gitAdapterActions` gates what *any* workflow may ask the git
adapter to do, and its own comment used to say destructive verbs were unreachable through
it.

**CLAUDE.md is explicit: a scope veto is NOT answered by resubmitting with better
measurements.** Do not re-run the council on this. RFC 011 costs four options and records
a preference (option B: keep the verb, drop it from the generic allowlist).

## OWED — correctness debts on code that is ALREADY LIVE

These are not packaging arguments; they are defects. All four land on
`retract_page_graph.go`, the part added fastest.

1. **HIGH — the link census does not use `linkablePageStatusPredicate`.** A landmine says
   an offline census over `pages` must, or an archived page makes a correct result look
   wrong. Verify and fix. *(This one is in the auto-memory index; I built the thing it
   warns about.)*
2. **MEDIUM, and the sharpest catch of the round — inbound references also live in
   structured `content_data` fields** (`link_url`, `cta_url`, hero-tool/tool-cta `*_url`),
   not only in `href=` markup. An href-only scan can miss a real inbound link and let a
   retraction proceed — **the exact case the owner's 2026-08-03 directive exists to
   prevent.** Fix before any retraction runs.
3. **MEDIUM — the inbound-source logic is COPIED from `check_orphan_pages.go`** with only a
   comment warning of drift. Extract and share it. (My argument that the two ask different
   questions — what IS unreachable vs what BECOMES unreachable — is in the file header and
   is not wrong, but a comment is not a mechanism.)
4. **MEDIUM — `status='active'` is a third bespoke spelling** rather than a consolidation
   onto a canonical helper. The census half of this was answered after submitting: the only
   other selector with the defect was `queueDirectoryPageRerenders`, fixed in `8f73e7279`.

## WHAT IS SETTLED, SO YOU NEED NOT RE-DERIVE IT

- **The far side already reconciles.** `b2 sync --delete` + a Cloudflare purge per changed
  domain, on every push. Removing a file from the repo genuinely retracts the page. 098's
  "make deploys reconciling" option was half-built already.
- **GitHub's deletion semantics, probed live** (POST `/git/trees` — creates an unreferenced
  object, no ref moves, no workflow fires): null sha on an existing path → **201**; on an
  absent path → **422 `GitRPC::BadObjectState`**. The existence filter is required.
- **Path derivation** (round 1's gating objection, answered by measurement): the real
  `PageFilePathFromURL` over all 13 candidates against **each page's own deploy repo** —
  11 of 12 `sites`-repo pages have a file at exactly the derived path, the 12th is
  genuinely absent and 404s, the 13th deploys to `vm-sites` and is correctly absent there.
  **Zero mismatches.**
- **The primitive works**, proven live: write → delete → idempotent re-delete on a scratch
  path the deploy workflow ignores, no-op logged, branch head unmoved.
- **`page-retraction` agent seeded** in `sql_for_agents/215` — **written but NOT APPLIED**.
  Do not apply it while the veto stands.

## TRAPS THIS SESSION HIT (all in NOTES, worth 30 seconds)

- **Comments are not in the binary.** I pod-verified a chassis image by grepping for
  strings I had added *as Go comments* — read 0 for everything, which looks exactly like
  "the fix did not ship". Use a string literal the change introduced: SQL, a log line, an
  error message.
- **A control is code, and unreviewed code is where errors live.** Three separate
  verification checks were wrong this session while the code was right, including a
  hand-rolled "negative control" regex that fired on the fixed binary.
- **`sites.github_branch` says `main`; `gqls/sites` has no `main`** (it carries `master`,
  its default, and `750start`), and the deploy workflow triggers only on `master`. Never
  pass that column to a commit or retraction.
- **Check each page against ITS OWN deploy repo.** relojistas is `vm-sites`; a sweep
  hardcoding `gqls/sites` reports a stale leftover as the live artefact.
- **A link census that reads one table answers a question about that table.** My first
  orphan check read only `page_components` and confidently reported 10 of 16 pages would be
  stranded. The nav and site chrome are where most links live; the real answer was zero.
- **`go build` cannot parse SQL, and `[]string` is not a driver type here.** PREPARE against
  the live schema (caught `site_components.component_type`, which does not exist — it is
  `slot_name`) and pass arrays via `datahelpers.PGTextArrayLiteral` with `::text[]`.
- **All-zero guard results look identical to a broken query.** Always re-run against a case
  you know is positive: on robot-hands, `/contact.html` gives 4 body / 2 chrome / 1 nav.

## THE 090 DIAGNOSIS RUN RETURNED NOTHING

Filed as required for a structural claim (`5bdec8cf-24cc-419f-8d9d-b3d7a8df6dbb`). It
**completed and wrote three bundles and no verdict** — no confirmation, no refutation. NOTES
records the six-step first-hand chain standing in for it, per the owner ruling of
2026-07-31 that a substitute must be *declared*, not silently omitted. Do not cite that run
as corroboration.

## NEXT ACTIONS, IN ORDER

1. Put **RFC 011** in front of the owner. Nothing else about the primitive moves until it
   is broken.
2. Pay debts **2 then 1** (both make the retraction safe), then 3 and 4.
3. Re-confirm the retraction with the owner *after* the RFC, then run it with the two-part
   acceptance.
4. Only then consider closing 098 — and it closes only when the live instance is retracted
   **and stays retracted**, not when the code exists.
