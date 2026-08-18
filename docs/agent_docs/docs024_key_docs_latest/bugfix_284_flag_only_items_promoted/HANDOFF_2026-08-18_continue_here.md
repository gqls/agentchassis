# HANDOFF 2026-08-18 — the 279/284/290 thread: what is DONE, what is LIVE, what is OPEN

**Read this first if you are picking the thread up cold.** It covers four bugs, one
RFC verification and one site item, across 2026-08-15 → 08-18. Everything below was
verified at the artefact; where a claim is unproven it says so.

## 1. State as at 2026-08-18 15:45 UTC (re-measure before trusting — this decays)

| fact | value | how to re-check |
|---|---|---|
| chassis image | `v1.0.1309`, revision `f0117fb8b93e…` | digest of pod == digest of local image, then read the OCI `revision` label |
| binary probe | new revision PRESENT, previous (`a6d1c53c`) absent, fake sha absent | `kubectl exec <pod> -- grep -aq "<40-hex>" /proc/1/exe`, **always with both controls** |
| handler-less `blocked` rows | **0** (was 60) | `SELECT count(*) FROM site_work_items WHERE status='blocked' AND COALESCE(handler_agent,'')='';` |
| minted `audit_finding_%` since 08-15 | **0** | `… WHERE item_type LIKE 'audit_finding_%' AND created_at > '2026-08-15 18:45+00';` |
| `agent_definitions` NULL descriptions | **0** (column is NOT NULL) | `SELECT count(*) FROM agent_definitions WHERE description IS NULL;` |
| CHECK `swi_no_handlerless_promotable` | present + **validated** | `SELECT convalidated FROM pg_constraint WHERE conname='swi_no_handlerless_promotable';` |
| RFC_022 budget cron | writing a `doc_notes` row daily | `… WHERE body LIKE 'OPTIONAL-KEY BUDGET CHECK%'` — **a MISSING row means the job did not run**, which is not the same as "nothing is wrong" |

## 2. CLOSED, fixed AND live (all in `bugs_closed/`)

- **279** — `classifyFinding`'s fallback minted `audit_finding_<category>`, an item_type
  registered nowhere, so those rows died in `detected`. Now files **`capability_gap`**
  (deferred, empty handler, `spec.builder_needed`), counted in the result map as
  `unrouted_categories` and Warn-logged. Dead `work_item_type` deleted from two prompts
  (mig `416`). Closed-set CI test + a build-time ratchet banning **constructed**
  item_types, at two layers (`pattern-check.py` advisory + a blocking Go test).
  Councils `925d7759` (r1) and `336d1549` (r3) APPROVED.
- **115** — the brief-fidelity auditor that "nobody ran". It now speaks the router's
  category vocabulary (mig `417`) and runs inside every improvement sweep (mig `419`).
  Live proof: 8 findings → 8 routed items, 0 minted.
- **284** — flag-only findings were promoted then stamped `blocked` by the claim path.
  Guard `7027a2801` + 60 rows repaired (mig `442`) + hand-insert path closed by CHECK
  (mig `443`). **Proven with a manufactured demand control**, twice — before and after
  the tie-break refactor: a site with 36 flag-only rows and nothing routable returns
  `promoted: 0, not_promotable: 36`.
- **290** (renumbered from 287 — a concurrent session took 287 by 89 seconds; resolve
  bug numbers by SLUG) — an agent seeded without a `description` could not be spawned,
  resolved or listed; five readers scanned a nullable column into a Go string. Data fix
  (mig `420`), **schema fix (mig `438`, owner-ruled): `NOT NULL DEFAULT ''`**, code fix
  `COALESCE` at all five readers + a module-wide source guard. Council `ad789fe1` APPROVED.

## 3. The owner's three decisions of 2026-08-17 — all executed

1. **Re-file the fundamentallyai.com item** → **NOT re-filed, deliberately.** Checked the
   served pages first: all three asks (Tools in nav; guides link their tool; tools linked
   from body copy) were already true, done by other lanes' rebuilds while the item sat
   blocked. Item closed as satisfied with the measurements inside it. **Filing work that
   is already done burns a pipeline run and leaves a false record.**
2. **Unify the third copy of the routability predicate** → done, council `79505ac5`
   APPROVED, live on `v1.0.1307`, and proven end-to-end through the refactored path (not
   just by a string-equality test). The renderer moved DOWN into `discovery_checks`
   (import direction forces it); `actions` delegates in one line so the claim path stays
   byte-identical. **It was FIVE renderings, not three** — the two extra
   (`core-manager/admin/agent_handlers.go:101,:724`) are deliberately excluded because
   they ask a different question (admin CRUD existence, not routability).
3. **Build the RFC_022 counter** → **nothing to build; it already existed** (counter
   2026-08-13, N=10 ruled 08-14, daily cron live since 08-14). CLAUDE.md's clause was
   three days stale and said otherwise — now corrected. What I found and fixed instead:
   `check.py`'s literal counted `retract_asset_files` (4 keys) and `publish_site` (3) as
   **ZERO**, so they were invisible to the accumulation check; the parity test that
   catches this was FAILING at HEAD and nobody had run it.

## 4. OPEN — what the next session can pick up

### Needs an OWNER decision (do not act unilaterally)
- **Two fundamentallyai.com findings**, surfaced while verifying, NOT actioned: (a) three
  tools appear to have **two guides each** at two path conventions (`/blog/<tool>-guide`
  and `/guides/tool-<tool>-guide`) — possible duplicate content; (b) the **Platform Log
  index no longer links to any guide at all** — it links straight to tools, so the writing
  may be orphaned from its own section index. Evidence is in the closed work item's
  `result` and in the lane README.

### Actionable without a decision
- **`go test ./cmd/config-key-audit/` runs nowhere automatic.** The parity guard worked
  and still sat failing for days because that package is untouched by ordinary work. A
  pre-commit or CI hook that runs it when `check.py` or any `ActionInputSpec` changes
  would close it. (LANDMINE written 08-17.)
- **`bugs_open/083` observation, not chased:** `placeholder_contact` → `page-build-handler`
  stands at **0 complete / 4 failed** — a handler that has never once succeeded at an item
  type it is named for. It therefore can never pass the scheduled promoter's
  "has succeeded before" gate. Belongs to that lane.
- **The scheduled promoter has never written through CHECK `443`.** It ran clean after the
  constraint landed but promoted nothing (its own precedent gate excluded the only
  candidates), so compatibility is proven *by construction and a clean run*, not by an
  observed promotion. Notice the first real one.

## 5. Traps this thread hit — do not re-pay for them

- **A "fresh build" can ship NO new code.** A same-tag rebuild serves the node's cached
  image: pods look new, binary is unchanged. Hit on 2026-08-17 (237 commits stranded).
  **The one-command proof is the DIGEST**: pod `imageID` vs local `RepoDigests`. Then the
  binary probe — and the negative control must be *capable* of being absent (40 zeros
  matches every binary; use a plausible fake sha).
- **A producer's status literal may have been changed BY THE FIX you are evaluating.**
  Both `capability_gap` producers read `deferred` at HEAD, which looks like it refutes the
  whole 284 diagnosis; `git log -S` shows `deferred` arrived in 284's own fix commit.
- **A zero from a guessed URL is a false zero.** Six "guides have no tool links" readings
  were 404 pages. Control it: does the page have a body at all?
- **Read the owning lane's DIRECTORY, not just the bug file.** I wrote migration `442`
  while an equivalent repair already sat in this lane (now bannered SUPERSEDED). Logged in
  `WRONG_CALLS.md`.
- **Dates: anchor to the clock, not to the deploy.** A day of work was dated 08-16 because
  I anchored to the roll. Corrected visibly; the two recorded migrations keep their
  wrong prose dates because editing a recorded file drifts its ledger checksum.

## 6. Commands you will want

```bash
# Is my code actually live? (per SERVICE, never per fleet)
P=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
IMG=$(kubectl -n ai-persona-system get pod $P -o jsonpath='{.spec.containers[0].image}')
kubectl -n ai-persona-system get pod $P -o jsonpath='{.status.containerStatuses[0].imageID}'   # must equal:
docker image inspect "$IMG" --format '{{range .RepoDigests}}{{.}}{{end}}'
docker image inspect "$IMG" --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
git merge-base --is-ancestor <your-commit> <that-revision> && echo LIVE

# Drive the 284 guard on demand (a real demand control, no side effects on a site with
# flag-only rows and nothing routable — leopardessconsulting.co.uk was that site)
#   one-step workflow calling action "triage_detected_items" with site_id;
#   expect promoted: 0, not_promotable: N. Publish with payload-in-COMMAND (kcat landmine).

# RFC_022 counter, on demand
./scripts/audit-optional-key-budget.sh 10          # human;  --json for the full census
go test ./cmd/config-key-audit/                     # the parity guard nobody runs
```

---

> **NOTE ADDED 2026-08-18 by the bugfix-284 session, and it is an apology, not content.**
> I overwrote this file with a shell redirect (`cat >`) while writing a handoff of my own,
> having assumed the name was free. It is restored here **byte-for-byte** from commit
> `1173f49f6` — verified with `git diff --stat 1173f49f6 -- <path>` returning empty — and
> my narrower lane handoff now lives beside it as
> `HANDOFF_2026-08-18b_bugfix_284_lane.md`. Nothing of yours was lost; git had it.
> The habit that would have prevented it is CLAUDE.md's own: *read before write on any
> file you did not create, and prefer a tool that refuses an unread file over a redirect
> that does not*. Logged in WRONG_CALLS.

---

# ADDENDUM — 2026-08-18, later: four owner-directed tasks, and one premise that did not survive measurement

Chassis at this point: **`v1.0.1309`**, revision `f0117fb8b93e…`, digest-matched, probe
discriminating (new revision PRESENT, previous `a6d1c53c` absent, fake sha absent).
Everything this thread built is an ancestor of it. Nothing of ours awaits a roll.

## A. Contributed to `bugs_open/083` — narrow, because the lane already had it

`placeholder_contact → page-build-handler` was already in 083 §3, measured better than
mine (3 held, 0/6 lifetime, with their own correction for pruned rows). So I added only
the two things that were NOT there, both of which bear on whether to canary the pair:

1. **The handler is healthy — the failure is TYPE-specific.** Same handler:
   `needs_page` 179/317, `content_rewrite` 133/277, `needs_content_page` 84/149,
   `tone_shift` 3/4. "Never completed one" does not distinguish a broken handler from a
   broken pair; this does.
2. **3 of the 4 surviving failures are one afternoon's dead AI endpoint** (all
   2026-08-14, identical `execute_llm_prompt: AI end…` error), 1 is a section-shrink
   guard, and **6 `unresolved` rows carry an EMPTY error column** — the silent class,
   and the part I would look at first. ⚠ Marked in the file: the live table holds 4
   failed rows against their lifetime 6, so 2 are pruned and uncharacterised.

**Not done, deliberately:** no rows touched, no canary run. §4's one-way door is the
`453`/083 author's call.

## B. The "duplicate guides" — THE PREMISE WAS WRONG, so nothing was cleaned up

The owner asked me to investigate on the basis that three tools had two guides each and
that duplication was unintended. **They are not duplicates.** Each pair is two DIFFERENT
articles about the same tool — a usage guide and a conceptual/decision guide:

| tool | `/blog/…` | `/guides/tool-…` |
|---|---|---|
| Automation Savings Estimator | "How to Use the Automation Savings Estimator" | "How the AI Automation Time Savings Estimator Works" |
| Model Approach Selector | "How the Model Approach Selector weighs fine-tuning, RAG…" | "Prompting, RAG, or fine-tuning: a decision guide…" |

**Archiving either would have destroyed real content, not removed a copy** — so I filed
no de-duplication work despite being asked to clean it up. Inventory 2026-08-18: 7 active
guides (4 `/blog/`, 3 `/guides/`) + 1 archived (`ai-readiness-checker-guide`, whose live
sibling is the `/guides/` one). **What remains is a content-strategy question for the
owner** — does the site want two pieces per tool? — not a defect, and it is NOT filed as one.

## C. The index → filed as `bugs_open/309`, and it is bigger than "it omits the guides"

The index does NOT omit them. It lists them and renders them **unclickable**:
**6 `<article class="bl-card">` cards, ZERO `<a>` elements in any of them.** Five are the
guides; the sixth is an ordinary blog post, so this is the whole index, not a guide
problem. Nothing anywhere on the site links a guide — checked `/`,
`/platform-log/index.html`, `/tools.html`, `/capabilities.html`.

- **Control that makes it a real finding:** `mortgagecalculator.co.uk/investor/index.html`
  renders 6 cards with **6** anchors. The component is capable; this page's render is not.
- **A second defect behind the first:** card 4 advertises `ai-readiness-checker-guide`,
  which is **archived** — restoring anchors alone would give that card a 404.
- **Root cause NOT diagnosed.** ~~`090` fired: correlation
  `df8ca3a1-9cca-474a-88fb-19577e088080` … a verdict should be waiting for the next
  session — read it first.~~
  > **CORRECTED 2026-08-18 (later): THAT RUN DOES NOT EXIST — do not wait for it.**
  > `diagnosis_artifacts`, `orchestration_states` and `site_work_items` are all empty
  > for that id, and no work item of any type was created in that hour. A printed
  > correlation is not a dispatched run; the intake ROW is
  > (`SELECT status, spec->>'dispatch_correlation_id' FROM site_work_items WHERE
  > item_key='needs_diagnosis:<slug>'`). A real run is now in flight —
  > **run correlation `6e578bf5-778a-4e72-aab2-0531e45c07d8`**, and artefacts are
  > keyed by that, never by the intake id the script prints first. Logged in
  > `WRONG_CALLS.md`; mechanism and corrections are in `bugs_open/309`'s addendum.
- ⚠ **Measurement correction recorded in the bug file:** a first pass on one truncated
  card string reported "1 anchor". Re-running over all six returned 0 each. A regex over
  a `[:900]` slice is not a measurement of the card.

## D. RFC_022's parity guard now runs at commit time

`scripts/check-optional-key-parity.sh`, wired as step 4 of `.githooks/pre-commit`.
**Scoped** (fires only when the staged diff touches the cron literal, the acks file, or a
Go file declaring an `ActionInputSpec`) and **advisory** (only `check-secrets.sh` may
block). A **build failure is deliberately not reported as drift** — this tree often does
not compile because of another session's WIP, and "I could not tell" must not read as "it
has drifted". Tested both ways: silent on irrelevant files; on the exact defect that was
live (`publish_site` removed) it names the action and the re-apply step. The
build-failure branch is written but was NOT exercised.

## What is open after this addendum

1. **`bugs_open/309`** — ~~read the `090` verdict (`df8ca3a1`)~~ **that run never
   existed (see the correction in §C)**; a real one is in flight under
   `6e578bf5-778a-4e72-aab2-0531e45c07d8`. The mechanism is now MEASURED end to end in
   309's addendum: the card anchors are gated on `postN_url`, those seven fields source
   `site_specs.blog.*`, and **the `blog` aspect has never existed on any site**, so
   `on_missing=skip_field` drops the key and the `{{if}}` drops the link — silently, with
   no empty href to see. Second half: `findBlogPage` only selects `blog-index`/`name='blog'`
   pages, so `RebuildBlogListingAction` never rebuilds this `section-index` page. It is a
   **missed migration** — every sibling list component already moved to `query.*`
   collection sources; this one alone kept the numbered-flat dialect, which is also why
   card 4 can name an archived page (its titles are `source: llm`, joined to nothing).
   Fix candidates are ranked in 309 §5. Verify at the SERVED page.
2. **Owner ruling 2026-08-18 — CLOSED, no action:** two or three articles per tool is
   fine, it is not a strict rule. The content-strategy question below is settled; nothing
   to file either way.
3. ~~**Owner content-strategy question** — two articles per tool.~~ **RULED
   2026-08-18: keep them; two or three per tool is fine.** See item 2 above.
3. **`bugs_open/083`** — the canary decision on `placeholder_contact` is that lane's,
   now better informed.
4. **Still true from the main handoff:** the scheduled promoter has never written through
   CHECK `443` (compatible by construction and a clean run, not by an observed promotion).
