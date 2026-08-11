# HANDOFF 2026-08-11 — RFC_015 is COMPLETE, LIVE and PROVEN on both seams. Read this, then §X.43–X.52.

**Written by the "ideauk sec" session at the owner's request (token load).**
Supersedes `HANDOFF_2026-08-03_continue_here.md` for everything after 08-04; that
file still holds the ingress/Cloudflare arc, which is CLOSED and untouched.

Cold-start reading order: **this file → RUNNING_NOTES §X.43–X.52 → RFC_015 →
`README_where_we_are.md` (owner's own log, plain prose)**.

---

## 1. What this arc was, in one paragraph

The owner removed all 35 component locks on idea.uk ("I haven't intentionally
locked anything") and ruled: *the provenance records should be what stops the
planner overwriting and regressing decisions — we want the site to improve, so
changes should be allowed, but not regress.* That became **RFC_015: decision
records with three faces** — WHY (prose), STEER (injected into generation), GUARD
(an outcome assertion checked by the discovery layer), plus a **citation gate** at
the write seams: *you may change anything you can NAME; you may not change what
you did not know existed.* It is now live end to end.

## 2. State: everything below is [VERIFIED] live, do not redo

Running chassis **v1.0.1289**, built from **`f914ec81d`** — resolved via the
BLD-019 provenance stamp (`strings /app/agent-chassis | grep -oE '^[0-9a-f]{40}'`),
then `git merge-base --is-ancestor <my-commit> <build-sha>`. **Use that method, not
marker-hunting**: my last fix added no new string literal, so a grep for it could
never have dated it. And use a real control — a commit made AFTER the build must
test NOT-an-ancestor, or your test cannot come out false.

| piece | state |
|---|---|
| STEER | LIVE on `webdesign-agent` (`load_decisions` step) |
| GUARD (`decision_guards` check) | LIVE + **caught a real regression unprompted** on 08-10 (D-002), and has a completion **verifier** (`VerifyDecisionRegressionResolved`) |
| CITATION GATE seam 1 (`apply_section_edit`) | LIVE, **both directions proven at the artefact** |
| CITATION GATE seam 2 (`save_page_sections`, the rebuild door) | LIVE, **fired 5× on real traffic**, preserves rather than duplicates |
| `build_status='removed'` honoured | LIVE in **both** rerender readers |
| D-004 fence narrowed to the copy slot | LIVE (migration 394), **proven end to end** |
| `component_id`-first matching | LIVE in v1.0.1289 |

**Four decision records** live, all on idea.uk (`doc_notes`, `categories ?
'decision-record'`): D-001 free-beside-paid, D-002 no-tools-directory-on-index,
D-003 logo-reads-IDEA, D-004 guide-copy-hand-authored.

## 3. The five things that will mislead you (read before touching anything)

1. **`categories ? 'decision'` is NOT the enforcement key** — three other lanes'
   rows use it to mean "a note about a decision". Both readers filter
   **`'decision-record'`**. All four real rows carry BOTH tags, so the steer-side
   reader keyed on `'decision'` still works. A row tagged only `'decision'` is
   invisible to enforcement while looking like a decision in every listing.
2. **`page_components` count > `pages.sections` length is no longer a defect
   signal.** Removed rows and gate-preserved rows are both *correct* now. The
   discriminator is **timestamp identity**: rows from one save share a
   `created_at` to the second; real duplication leaves **older** rows beside new
   ones (`count(DISTINCT date_trunc('second', created_at)) > 1`). A census of
   `build_status='removed'` returns **zero fleet-wide whether or not the bug
   exists**, because the bug consumes the row into a `deployed` one.
3. **The rebuild DELETE now carries `AND NOT (id = ANY($2::uuid[]))`.** `id =
   ANY(NULL)` is NULL, so a NULL parameter makes it match **nothing** and every
   rebuild fleet-wide silently stops clearing rows while reporting success. The
   empty case is the NORMAL one (13/14 sites have no decision rows). It is a
   literal `'{}'`, and `TestDecisionProtectedIDArrayLiteral_EmptyIsAnEmptyArrayNotNull`
   fails if that changes. **Re-run the census + timestamp check any time you touch
   that statement.**
4. **The gate cannot tell copy from structure** — it preserves the whole row. So
   the *fence* has to carry that distinction. D-004 originally named 9 pages with
   `"slots":[]` (= every slot) and froze 27 sections, against its own wording.
   If you add a decision, **name the slots**.
5. **A `covers` fence spanning many pages needs a page-keyed item_key.** Ours is
   `decision_regression:<site>:<key>:<page>` and
   `decision_blocked_change:<page>:<slot>` for exactly this reason (dedup drops
   everything after the first on a coarser key).

## 4. Open, deliberately — with the reason

- **`page-content-writer` is unguarded.** This is the council's standing objection
  (`bug_historian`, HIGH, gated rounds 1–3). **Owner ruled: stop at 3** — each
  round names the next seam, and seats disagreed (`architecture` approved r3).
  Prerequisite before anyone builds it: that writer has **no way to cite a
  decision** today, so the citation must be plumbed there first.
- `decision_guards` reads stored assembly via `string_agg(rendered_html)`, which
  does not replicate the assembler's ≤10-visible-char drop / `data-runtime-fill`
  exemption (3 seats). Over-detection — the safe direction.
- `ExtractFencedBlock` is not hardened against a decision note's own prose
  mentioning a ```covers fence (documented landmine for the sibling parser).
- The sibling **lock** path still matches on slot name alone, so it keeps the
  `bugs_open/189` duplication trap that the decision gate no longer has. Not this
  lane's file.
- **Contributed, not ours:** `lock_blocked_change` has **37 live rows** and sits in
  **neither** half of the verifier coverage guard — its completions are taken on
  the handler's word with nothing recording that choice. Noted in
  `verifier_coverage_test.go` beside our own entry.
- Three `decision_blocked_change` items sit at `needs_human_review` by design
  (`index:brief-explanation`, `index:tool-list`,
  `guide-building-it:generic-text-block`). They are records of *prevented*
  overwrites, not defects. Resolution = decide whether to re-dispatch WITH a
  citation.

## 5. Older residuals from this lane (pre-RFC_015, still open)

- **idea.uk's first ORGANIC signed Stripe webhook.** Plumbing proven through the
  proxy (`/stripe/webhook → 400` = reaching the signature check); a synthetically
  signed event was deliberately never fired. After the next real order, check
  `/var/lib/idea/orders.json` moves to paid.
- Owed site work: tools-page card images into `items[].image`, tool-page heroes,
  `derive_brand_head_assets` (favicon/og-card are live 404s), news at
  `/data/latest-news.json`.
- **The empty-kind → SDXL image-routing hole** (`provider_hint` > per-kind routing
  > silent Stability fallback) — diagnosed but never filed as a bug.
- **Ingress (CLOSED, but the landmines outlive it):** grey is NO LONGER a safe
  rollback for idea.uk — `ufw allow 80,443` FIRST, grey second; a timeout curling
  `116.203.204.115` is the firewall WORKING. Liveness = `https://idea.uk` with a
  `cf-ray`, or ssh.

## 6. How to work this lane

- **Everything through the framework** (owner's standing constraint): insert a work
  item at `triaged` with `handler_agent='section-editor'`, or publish a
  `page-rerender` envelope. **Never edit content directly.** For a section removal
  there IS a route — `DELETE /admin/sites/:site_id/pages/:page_name/components/:component_id`
  (`HandleRemoveComponent`), which marks removed, LOCKS, and triggers the rebuild.
  I missed it twice and hand-rolled instead; don't.
- **Verify at the artefact, never at a status.** A `complete` item, a `success:true`
  and a rendered page are three different claims. Check the commit's **stat** — an
  EMPTY commit reports success (bit me: my first "cited edit works" proof wrote a
  value that was already the value, so the commit was 0/0 and proved nothing).
- **Council submissions: derive the edit list from `git diff --name-only`, never
  from memory.** Three rounds running I omitted files I had changed — once the file
  carrying the fleet-wide DELETE. Two HIGH objections were spent on my bookkeeping
  instead of the code.
- **Don't re-file the 090 diagnosis** on the resurrection mechanism: the run FAILED
  on a timeout (5 bundles, no verdict) and the defect is fixed at HEAD anyway. The
  structural claim rests on declared first-hand verification, which the 2026-07-31
  ruling permits provided you say so — and §X.51 does.

## 7. Verification recipes worth keeping

```bash
# did MY commit ship? (BLD-019, not marker-hunting)
kubectl exec -n ai-persona-system <pod> -- sh -c \
  'strings /app/agent-chassis | grep -oE "^[0-9a-f]{40}(-tree)?$" | head -1'
git merge-base --is-ancestor <my-sha> <build-sha> && echo IN || echo NOT-IN
# control: a commit made AFTER the build must print NOT-IN

# fleet safety after touching the rebuild DELETE
#   pages whose rows span >1 save = the real duplication tell
SELECT s.domain||'/'||p.name, count(DISTINCT date_trunc('second', pc.created_at))
FROM sites s JOIN pages p ON p.site_id=s.id JOIN page_components pc ON pc.page_id=p.id
GROUP BY 1 HAVING count(DISTINCT date_trunc('second', pc.created_at)) > 1;

# the two decision item types
SELECT item_type, status, item_key FROM site_work_items
WHERE item_type IN ('decision_regression','decision_blocked_change') ORDER BY created_at;
```

## 8. Post-roll check done for you (v1.0.1289, 2026-08-11 20:38Z)

Matcher fix confirmed live (ancestry + a control that returns NOT-IN). 41 rows
created across 12 pages since the roll. Duplication census reads **11**, and every
one is accounted for: `idea.uk/index` is the gate working (its single `removed`
row), `gamesdesign.co.uk/game-jelly-invaders` predates all of this work, and the
other nine have **all rows from a single save**. **Zero pages show the
old-row-beside-new signature.** The census number drifts as sites rebuild — it is
not a defect trend, which is the whole point of landmine 2 above.
