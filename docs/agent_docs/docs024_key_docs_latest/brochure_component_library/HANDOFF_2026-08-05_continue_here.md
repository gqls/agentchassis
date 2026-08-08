# HANDOFF 2026-08-05 — brochure component library / fundamentallyai.com

**Supersedes `HANDOFF_2026-08-03_continue_here.md` and its 08-04 addendum.** Written to
cold-start a fresh session; every liveness claim below was re-verified against the
running system on 2026-08-05 morning, not carried forward.

## 1. Where the lane is, in one paragraph

The looking loop is CLOSED and live end to end: acceptance runs photograph the page a
visitor actually lands on (both camera halves live on **v1.0.1252**, three independent
post-roll proofs), the photographs carry their viewport and state on the note line —
`(desktop 1366x900@1x, landing state)` on a real production note — and a weekly cron
puts the contact sheet in front of the owner. The 151 candidate-3 duplication checker is
enabled fleet-wide: since 08-03 it has swept **7 sites, deleted nothing, and filed 7
flag-only capability_gaps** — which is both the design working and the measured case for
this lane's largest remaining build, **candidate 1**. The site itself is whole:
`/tools.html`, working CTAs, both companion guides live. Two other lanes (bugfix 188,
200) independently verified the camera behaviour; bug 156 (the sibling duplication
class) was closed by its own lane on v1.0.1252.

## 2. What is LIVE, and the proof for each (all re-checked 08-05)

- **Camera, both halves, v1.0.1252.** Adapter markers `profileViewport`→2 and the
  driven-state fallback string→1 (`grep -acF` on `/app/browser-runner-adapter` —
  `strings` is absent from these images); chassis prints `landing state`→1. Three
  behavioural proofs: bugfix_188's run `25c44133` (22/22, populated landing render,
  their pod-grep + eye), bugfix_200's run `b14fee91` (its note carries the full
  new-form line above), and this lane's own fetch-and-look at `b14fee91`'s desktop
  PNG — the simulator shows the DEFAULT preset (70.1% headline, seat list), not the
  post-Clear empty panel that started TL-035 (d). A third manual run I queued before
  finding 188/200's proofs was **cancelled as redundant** — check for other lanes'
  proofs before spending a run.
- **Camera behaviour spec** (what a fresh session must not re-litigate): renders =
  landing state (captured post-settle, pre-`evaluateOnPage`, uploaded only on a full
  pass, `Stage:"landing"`); a failed landing capture falls back to a driven-state
  render with NO stage stamp; failure evidence = driven state always, no stamp;
  stage-less refs render the old line form byte-for-byte. Councils `a18db904`
  (viewport) and `8e35caad` (landing) both APPROVED r1; `2f374cdaf` carries
  `Council-Reviewed: 8e35caad` after acting on two advisories.
- **Checker (`content_duplication`), enabled by seed 296 on
  `completeness-discovery-agent`.** Fleet since enable: capability_gap rows on
  fundamentallyai, leopardess, gamesdesign, gaswholesalers (08-03), idea.uk,
  webdesign, robot-hands (08-04) — all flag-only (`do_not_auto_rewrite`), several
  now status `blocked` (they have no handler BY DESIGN; blocked is idle, not stuck).
  Zero `content_duplication` deletion items anywhere, ever. The would-delete census
  goes stale by design — re-run `gauntlet_dead_cta/scripts/dedup_census_shipped.go`
  before reasoning about it.
- **Weekly contact sheet.** `crontab -l` → Mondays 08:53,
  `scripts/weekly_contact_sheet_refresh.sh`; log `~/acceptance_renders/refresh.log`.
  **First scheduled fire is Monday 2026-08-10** — nobody has watched a scheduled
  (as opposed to hand-run) execution yet; check the log that morning. The claude.ai
  page (`14a45889-e1f0-46e9-969a-08295cc36650`) refreshes only on request in an
  interactive session (headless `claude -p` has NO Artifact tool — measured), and
  the URL is replaceable state: the owner deleted the 08-03 one within a day.
  `contact_sheet.py` captions every image landing/driven off the note-line stage
  token.
- **Site content.** `/tools.html` live (resolver-fed items — do not hand-tend the
  list); "Explore All Tools" → `/tools.html` on both tool pages; calculator hero →
  `#input-tokens` / its guide; both guides serve 200; decision-record stub archived.
  Simulator probe: 47 checks 0 failed as of 08-03 (the probe grows — trust exit
  code, not remembered counts; re-run after ANY re-render).

## 2b. TL-035 (e) — THE MACHINE EYE, wired 2026-08-05 evening (added after this file was written)

**Owner decision 2026-08-03:** close "nobody looks" with a **machine** eye — a vision
check that raises a work item on suspicion — rather than only the human contact sheet.
**Wired 08-05 by seed `317`** (applied by hand, recorded in the ledger).

```
ensure_site_record -> load_docs -> request_run -> judge -> look -> record_look -> complete
                                                            |       |
                                             error_step ----+-------+--> complete_no_look
```

- `look` = `execute_vision_prompt`, `images_field: "browser_run"` (the resolver descends
  object → `.response` → `.renders`), `output_type: "text"`, `max_images: 4`.
- `record_look` = `append_doc_note`, category **`render-critique`** — deliberately NOT
  `acceptance-run`.

**Four safety properties, each ASSERTED in seed 317's verify block, not left to review:**

1. **`look` runs AFTER `judge`**, so the verdict and acceptance note are already
   persisted. **Proven by accident on the first run:** the vision step failed and the
   12:06:57 acceptance note is still a normal `PASSED` note with a landing render.
2. **Both new steps route `error_step` to `complete_no_look`, a SUCCESS terminal.**
   `execute_vision_prompt` is fail-loud by design and **zero renders is a NORMAL outcome**
   for a run whose profiles all failed — routing that to `complete_error` would turn a
   working acceptance run into a reported failure.
3. **A distinct terminal**, so `current_step` alone says looked / did-not-look.
4. **`render-critique`, never `acceptance-run`** — polluting that category would break
   both the lane re-check AND `contact_sheet.py`, and would recreate the two-producers-
   one-category trap already in LANDMINES against it.

**It raises NO work items.** The findings→work-item drain and the general critic belong
to `vigilant_designer_offer_analysis` (their A2, `design-critique-agent`, still unseeded
as of 08-05). This wires the eye to OUR acceptance runs only. **Do not seed a rival
critic** — the intended end state is ONE critic with TWO image sources, and they own it.

**The prompt encodes `bugs_closed/188`'s findings as explicit NON-defects:** sticky nav
painted mid-page, extreme full-page height, and any blank/mid-interaction image (the rare
driven-state fallback) → say INCONCLUSIVE, do not report.

### The trap this uncovered, which cost the first run

**`params.StorageClient` was nil on the chassis, and every storage CREDENTIAL being set is
what hid it.** First live call died on `no storage client — cannot download screenshots`.
`agentbase/agent.go:308-330` builds the client only when the BUCKET is non-empty, from
`IMAGE_BUCKET` — unset. `execute_vision_prompt` is the **only** chassis action that takes
`params.StorageClient`; every other builds its own (`storage_actions.go:95/:612`, or
`screenshots.go:66` from service config — which is why the browser-runner uploads renders
fine while also having no `IMAGE_BUCKET`). **Fixed** in the chassis overlay (`820a033c0`),
hardcoded not `configMapKeyRef` (a wrong key there is `CreateContainerConfigError` and the
whole chassis stops). **Live from v1.0.1254**; verify at the log line, not `printenv`:

```bash
kubectl logs -n ai-persona-system <chassis-pod> | grep "Storage client initialized"
#   -> bucket=personae-prod-uk001-images   (the failure branch says "not configured")
```

**Transferable, and now in LANDMINES + 016b:** read a register's *"built, no live call
yet"* as **"deployment contract unverified"**, not "unused". Wire-shape tests assert
request bodies and pass happily in a world where the action can never obtain a client.

> ## ⚠ CORRECTED 2026-08-08 — THE STORAGE FIX DID NOT FIX IT, AND THE OWNER HAS RULED THE WHOLE APPROACH OUT
>
> **Everything above about `820a033c0` making the eye work is WRONG.** The fix went to the
> `agent-chassis` **deployment**. That is not where this agent runs. Agents run in
> **dynamically spawned per-agent pods** — `agent-tool-acceptance-agent-<hash>` — and
> **46 pods** run the chassis image (`agent-build-dispatch-loop-*`,
> `agent-directory-build-handler-*`, `agent-internal-link-resolver-*`, …).
>
> Proof, from the pod that actually processed the run (`processing_node` on
> `orchestration_states` is the column that names it — I should have read it on day one):
> ```
> agent-tool-acceptance-agent-e702cc67-2dnsp
>   IMAGE_BUCKET = <ABSENT>   S3_ENDPOINT = <ABSENT>   B2_APPLICATION_KEY_ID = <ABSENT>
>   agentbase/agent.go:329  "Storage client not configured (IMAGE_BUCKET not set)"
> ```
> That pod has **no S3 credentials at all** — not merely a missing bucket name. The
> 2026-08-08 17:16 run failed with the identical `no storage client` error as the first
> one, three days later, on a fully rolled fleet. **`render-critique` notes: 0.
> `llm_call_log` rows for step `look`: 0. The machine eye has never once seen anything.**
>
> **Why I believed otherwise, and the check that would have caught it:** I verified
> `IMAGE_BUCKET` on the `agent-chassis` pods and at `agent.go:324`'s success line, and both
> were real — but on the wrong pods. *"Prove it at the artefact"* is not enough on its own;
> you must prove it **on the node that ran the work**. `processing_node` names it, and no
> amount of grepping the deployment you assumed would ever contradict you.
>
> **OWNER RULING 2026-08-08: all S3 interaction stays with the client; credentials must not
> be spread across agents.** So the remaining route — injecting S3 credentials into the
> spawn template for every agent that might critique — is **closed**, and rightly: it would
> put bucket credentials into dozens of dynamically spawned pods.
>
> **TL-035 (e) is therefore BLOCKED ON AN ARCHITECTURE DECISION, not on a bug.** The seed
> 317 wiring stays in place and is harmless — its whole point is that a failed look cannot
> hurt the run, and three failures across two dates have demonstrated exactly that. But
> `execute_vision_prompt` needs image BYTES, and under this ruling the agent pod may not
> fetch them itself. See the CONTRIB to `vigilant_designer_offer_analysis` — their A2
> critic hits the identical wall, because it is the same action.

## 3. Open, in the order I would take them

1. **151 candidate 1 — assign facts to sections at plan time.** The lane's largest
   unbuilt piece, now with a measured population: 7 sites' capability_gap rows,
   fundamentallyai's being 9 fact-overlap pairs + 1 near-duplicate (fact pool 15).
   Constraints already established, do not rediscover: any REWRITE path needs the
   claims gate in front of it (a 07-29 LLM rewrite fabricated a human credential,
   `bugs_open/149` §C1); `content_data` is not always section content (two components
   can share a byte-identical site-context blob — the narrowing that saved vonc's
   lobby-grid); the residue gaps are the acceptance population for whatever candidate
   1 becomes. Start by READING: `bugs_open/151` (updated by the gauntlet lane),
   `CONTRIB_2026-07-31_151_candidate_3_is_built.md` (both halves + the update), the
   7 gap specs (`spec->>'check'='content_duplication'`). This is plan-time platform
   design — expect an architecture-aware council submission, and check
   `who-owns.py 151` + live transcripts first: other lanes brushed this space
   overnight (156 closed; PBP-033 dedups at the save choke point — candidate 1 is
   the PLAN-time complement, not a duplicate of it).
2. **Watch the checker as sweeps continue.** A non-zero `content_duplication` item is
   worth reading, not alarming — the guard refuses plan-specified repetition and
   locked rows; read the item's skip reasons before touching anything.
3. **Monday 08-10: confirm the first scheduled cron fire** (`refresh.log`). Expected
   failure mode: kubeconfig token expiry → the push says so; that is designed
   behaviour, re-run after the owner refreshes.
4. **Small, unclaimed:** `tool-guide-intro` on the simulator page remains deliberately
   absent (whole-page escalation risk; the guide covers the need; safe route =
   `section_edit` with JSON authored from the guide's copy). The `08-03` handoff §5
   items are otherwise all discharged.

## 4. Traps for a fresh session (beyond CLAUDE.md; each is in LANDMINES/WRONG_CALLS too)

- **A queue `page_rerender` with no `reason` in spec is ASSEMBLE-ONLY** — item
  completes, page deploys, your content_data edit is not served. Use
  `scripts/rerender_page_sections_direct.sh` (proven repeatedly).
- **Static-source `input_schema` fields overwrite authored content_data on every
  resolve** (and `query.*` fields regenerate). Read the component's schema first;
  author only `llm`/unsourced keys.
- **Verify a link TARGET at the artefact before shipping the link** — live copy
  promised a guide that had served 404 for nine days.
- **The work-item schema moves under the docs**: the RUNBOOK's INSERT recipe drifted
  twice in one morning (`category`, `pipeline` — `pipeline` EXISTS again and the
  build trigger selects on it; `category` does not). Copy a live row's shape.
- **This tree is shared and hot**: my own test file gained another lane's edits
  mid-session; LANDMINES/WRONG_CALLS appends travelled as same-file passengers in
  another lane's commit (recorded, nothing lost); `discovery_checks` carried a
  broken WIP for hours — test against `git archive HEAD` + your files only.
- **Hours can pass between turns** — `date` before writing any timestamp claim.
- **look.py's blank lower half** on a short page is its vh-stretch artifact, and
  a full-page capture paints the sticky nav mid-page — capture artifacts, not
  page defects.
- **Cron PATH needs `/snap/bin`** (kubectl is a snap) or auth pre-checks report
  token-expiry for command-not-found.

## 5. Commands a fresh session will want

```bash
# cold start reads (in order): this file, then
#   NOTES_brochure_component_library.md   (tail — the 08-03/04 entries)
#   RUNBOOK_brochure_component_library.md (§republish, §weekly contact sheet, §acceptance by hand)
#   register/tool-lifecycle.md TL-035 + TL-039

# checker state
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
 "SELECT s.domain, wi.item_type, wi.status FROM site_work_items wi JOIN sites s ON s.id=wi.site_id
   WHERE wi.item_type='content_duplication' OR (wi.item_type='capability_gap' AND wi.spec->>'check'='content_duplication')
   ORDER BY wi.created_at DESC;"

# latest acceptance notes (the stage token is the camera's liveness signal)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -t -A -c \
 "SELECT created_at, subject_key, body LIKE '%landing state%' FROM doc_notes
   WHERE categories ? 'acceptance-run' ORDER BY created_at DESC LIMIT 5;"

# contact sheet, on demand
~/.venvs/vonc_pw/bin/python3 docs/agent_docs/docs024_key_docs_latest/brochure_component_library/scripts/contact_sheet.py --limit 8
```

## 6. Commit / council trail (this lane, 08-03 → 08-05)

`30dde02d1` seed 296 · `1f375991f` contact_sheet.py · `d0a873f97` viewport
(council `a18db904` APPROVED r1) · `5c7346b40` docs+TL-039 · `fe51ad611` landing state
(council `8e35caad` APPROVED r1) · `2f374cdaf` advisory fixes (Council-Reviewed
8e35caad) · `948c3d3e4` cadence · `df8f6087a` + this file's commit, docs.
Cross-lane closures against this work: `bugs_closed/188` (0f124a686), bugfix_200's
verification run. Sibling context: `bugs_closed/156` (PBP-033, save-time dedup,
v1.0.1252) — candidate 1 complements it at plan time.
