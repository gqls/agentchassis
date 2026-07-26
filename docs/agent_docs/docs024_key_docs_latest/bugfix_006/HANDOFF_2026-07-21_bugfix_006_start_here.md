> **CLOSED 2026-07-26 — this handoff is now HISTORY, not an entry point.**
> All three sections are answered and the case has moved to
> `bugs_closed/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md`, which carries the closure evidence
> and the two named residuals. Current read-out:
> `SUMMARY_2026-07-26_bugfix_006_closed.md`. Design behind §C:
> `PLAN_2026-07-26_claim_timeout_generic_evidence.md`. Commands: `RUNBOOK_bugfix_006.md`.
> Missteps: `NOTES_bugfix_006.md`.
>
> **What in the text below has gone stale:** the "one thing to check first" (that council verdict
> landed — REVISE, resolved, and a later submission was APPROVED as `8bfcbc68`); B's three
> follow-ups are all done; **A resolved on its own**; and **C is fixed and live** via migration
> `220`, so its "residual worth fixing" is no longer open. The §C figures quoted below (42
> timeouts / 0 auto-completions) are the 07-20 measurement — re-measured 07-26 as 27/0 for
> `needs_page` and 84/14 fleet-wide. Left unedited: it is the record of what was believed then.

# HANDOFF — bugfix 006 (start a new chat here)

**Session:** "bugfix 006", 2026-07-20 → 2026-07-21. **Cold-start entry point.**
Bug file (the technical record, kept current): `bugs_open/006_HANDOFF_2026-07-16_idea_uk_infra_errors.md`.
Read that for full evidence; this is the orientation + what-to-do-next.

`006` is three INDEPENDENT errors (A/B/C) filed together, "route each to its own chat". This session
verified all three against the live cluster, fixed **B**, and found a fourth bug while verifying C.

---

## TL;DR — where we are right now

| item | state | who acts next |
|---|---|---|
| **B** — contact forms deliver nothing, fleet-wide | **FIX LIVE IN POD** (v-tag carrying it verified by pod-grep). Effect NOT yet realised (pages need re-render; check not enabled). **Council review PENDING.** | this workstream: watch council verdict, then the 3 follow-ups below |
| **A** — GitHub Actions runner replica crash-looping | CONFIRMED, worse (6365 restarts/23d). Node-level infra. | **owner** (cordon/drain or fix node containerd SystemdCgroup) — not a code change |
| **C** — claim-timeout churn / stall | Open question ANSWERED (sweep = `claimed-item-timeout` scheduled_task; 15-min evidence branch, 40-min reset). Residual: evidence branch covers 3 of 18 item types. | a fixing chat: make completion atomic/idempotent, OR accept per-type gaps |
| **048** (NEW) — no-op pre-query starves its concurrency group | FILED `bugs_open/048`. 4 maintenance tasks dead 79 days. | a fixing chat: `cmd/scheduler/main.go` — release slot on early exit + stamp last_triggered_at |

**The one thing to check first in a new chat: the council verdict for B.**
```
SUBMISSION_CORR = c75718c1-c6e1-45b8-bb4d-f66a28759b5c   (submitted 2026-07-21 ~12:45 UTC)
RUN_ORCH_ID     = c15279a7-c487-4337-93b2-ec1424172ca8   (orchestration name council-gate-124527)
```
```sql
-- verdict (keyed on the submission correlation — the key artifacts are written under):
SELECT created_at, metadata->>'decision'
FROM diagnosis_artifacts
WHERE correlation_id='c75718c1-c6e1-45b8-bb4d-f66a28759b5c' AND kind='council_report'
ORDER BY created_at;
-- human-readable note:
SELECT body FROM doc_notes WHERE categories ? 'council-gate' ORDER BY created_at DESC LIMIT 1;
-- run progress (if no row yet, it is QUEUED, ~30 min under load — NOT dropped, do not resubmit):
SELECT current_step, status FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = 'c75718c1-c6e1-45b8-bb4d-f66a28759b5c';
```
- **APPROVED** → the code is already committed and live; add the trailer to a follow-up commit or
  note it in the bug file: `Council-Reviewed: c75718c1-c6e1-45b8-bb4d-f66a28759b5c`.
- **REVISE** → objections come with the reviewers' checks answered; the fix is live, so treat
  revisions as follow-up commits, resubmit with `RESUBMIT_CORR=c75718c1-...`. Update the **sketch**
  fields, not just prose (reviewers judge the sketch — documented trap).
- **REJECTED** → guardian veto names a contained alternative; the fix is already live, so this would
  mean a revert-or-amend decision — bring it to the owner.

---

## B — the fix, in full (bugs_open/006 §B has the evidence)

**What was actually wrong** (the filed cause was stale): the fleet contact-form template renders
`action="{{.form_action}}"`, `form_action` comes from per-component `content_data` written by the
content LLM, and **no Go code ever set or validated it**. So live values were whatever the model
emitted: `#contact` (8 sites), `""` (3), one hand-fixed `mailto:` (idea.uk). All of `#contact`/`""`
POST to the current URL → 405/404 on a static host → message silently lost. **10 of 11 live sites.**
Nothing has posted to `/contact` (the originally-filed cause) in a long time.

**The fix** (owner's own 2026-07-17 decision "convert form → mailto"):
- `sanitiseFormAction` / `sanitiseFormActionStrings` at the render seam replace a non-delivering
  action with `mailto:<site email>?subject=<domain> enquiry`.
- **Refuses to fabricate an address.** 4 sites (robot-hands, relojistas, vetcomparison, vonc) have no
  contact address; a mailto nobody reads makes the form *look* fixed while still losing the message
  and removes the only outward signal. Those keep their visible breakage; the new check raises them.
- Leaves real destinations alone (existing `mailto:`, live handlers like idea.uk's `/request`).
- `contact_form_undeliverable` discovery check = the backstop for the address-less remainder.

**Two near-misses, both caught by the repo's own machinery (not by me):**
1. A base-map default would have HALF-worked — `ContentData` merges *over* the defaults, so it would
   have fixed the 3 empty sites and left the 8 `#contact` sites broken *looking fixed*. Hence
   post-merge sanitisation, guarded by `TestFormActionSurvivesContentDataMerge`.
2. The fix first covered only ONE of two render branches — the pre-commit pattern check flagged
   `contextToInterfaceMap` changed without its twin `contextToMap` (016b §9 untouched-twin). The
   regex fallback path merges `ContentData` too. Fixed in `c419b6f34`.
   Both tests were **fault-injected and watched to fail** before being trusted.

**Commits (all in history, verified live in the pod):**
- `22678a74b` — sanitiseFormAction, Go-template path
- `c419b6f34` — sanitiseFormActionStrings, regex fallback path
- `3913a0adf` — check_contact_form_undeliverable.go
- `d4e3b42be` — bug-file fix-state record
- (context: `86e581368` verification pass, `ad70e2dda` idea.uk correction)

**Live-in-pod proof (2026-07-21):** `strings /app/agent-chassis | grep -c` in
`agent-chassis-59c675c4f-pxr9f`: `contact_form_undeliverable`=5, `nonexistent_contact_endpoint`=1
(a literal unique to my new file), control `placeholder_contact`=4. Per CLAUDE.md the pod is ground
truth — the code is deployed.

### B — the 3 things still to do (NONE done here)
1. **Enable the check.** Add `contact_form_undeliverable` to a discovery agent's `checks` array
   (a DB config change — live immediately, no image). Sibling checks (`dead_controls`) show the
   convention. Do this AFTER confirming the image carries the file (it does).
2. **Remediate the 10 already-deployed components.** The Go fix only affects NEW renders. A rebuild
   of an already-`deployed` page bounces to `needs_human_review` at attempt 0, so this is a separate
   costed step: a `content_data` migration over the affected components + a render through the review
   gate. Not automatic.
3. **The 4 address-less sites need a real contact address**, or the check keeps raising them
   (correctly). And **`sites.content_data.email` can be stale** — idea.uk holds the old
   `idea-uk@leopardess.uk` while its identity `site_spec` holds current `idea.uk@contactforsales.com`.
   **[UNVERIFIED]** which source populates `RenderContext.Email` on each render path — check before
   remediating, or stale addresses bake in. (Grep `\.Email = ` in `platform/orchestration/actions/`;
   there are ≥8 assignment sites.)

---

## A — confirmed, owner action (no code)
Runner replica `github-actions-runner-5c44ddb44d-lhg9l`: `0/1 CrashLoopBackOff`, **6365 restarts /
23d** (was 4906/18d at filing). Node-level cgroup-driver mismatch (kubelet vs containerd
systemd/cgroupfs) — StartError exitCode 128. Single-point-of-failure on the fleet deploy path; the
healthy sibling `-5pqdv` is carrying it. Fix = reschedule off the bad node or fix that node's
containerd `SystemdCgroup`. A third runner deployment `-vmsites` (healthy) now exists.

## C — open question answered; a real residual remains
The claim-timeout sweep is **not missing and not in Go** — it is the `claimed-item-timeout`
`scheduled_tasks` row (interval 120s). Two branches: **auto-complete-on-evidence at 15 min**, **reset
at 40 min**. That exactly explains the 16-min non-revert in the 07-20 addendum (nothing was due at 16
min; reset is 40). **Residual worth fixing:** the evidence branch tests only 3 item types
(`needs_content_page`, `page_rerender`, `needs_design`) of 18. `needs_page` — C's own symptom — has
42 timeouts / 0 auto-completions over 14 days. Structural fix = mark the item complete atomically
with the last deploy step, or retry the completion write idempotently (removes the need for
per-type evidence branches). See `bugs_open/006` §C.

## 048 — NEW, found while verifying C
`bugs_open/048`. A `scheduled_tasks` row whose pre-query returns no rows takes its concurrency
group's only slot **before** running the query (`cmd/scheduler/main.go:171-184`), then bails on the
no-rows `continue` (`:195-199`) **without** stamping `last_triggered_at` — and the due-query sorts
`last_triggered_at ASC` (`:270`), so it pins itself at the head of the group forever. 4 maintenance
tasks dead 79 days while `enabled=true` (`feasibility-recheck`, `database-cleanup`,
`stale-work-item-reaper`, `work-item-archiver`); measured 11-item backlog. **Not** 029 (that's hung
orchestrations holding real slots). Slot-leak-in-DB hypothesis was tested and REFUTED — the leak is
in-memory. Fix candidates 1+2 in the file. **NB: `cmd/scheduler` is its own binary** — needs a
scheduler build+roll, not chassis.

---

## Landmines this session hit (so you don't)
- **A `bk_*.sql` file is a backup DUMP of a table, not source.** 006 §B cited one as the emitter and
  the wrong cause survived 4 days. Before citing a `bk_` path as a cause, grep the Go tree for the
  field — if nothing sets it, the value is *content*, and the fix is a default + validation, not a
  template edit. (Logged as WRONG_CALLS #9.)
- **The shared tree may not compile — it isn't yours.** `check_news_feed.go` had 135 uncommitted
  lines from another session that broke the package build. Verify your change with
  `git archive HEAD | tar -x` into a clean dir + your files overlaid, then `go build`/`go test` there.
- **Verify a detector by inducing the fault.** Both new tests were confirmed to FAIL with the fault
  present before being trusted; the check's SQL was tested against live data and narrowed (first
  draft matched 16 rows incl. 6 working tool calculators → scoped to `data-component="contact-form"`
  → 10 rows, one per site).
- **Council: budget ~30 min, find your run by payload not the printed id, a missing row is QUEUED not
  dropped — do not resubmit** (costs a duplicate round).

## Commands worth keeping
```bash
# pod-grep proof (ground truth, not git/tag):
kubectl -n ai-persona-system exec agent-chassis-<pod> -- sh -c 'strings /app/agent-chassis | grep -c "nonexistent_contact_endpoint"'
```
```sql
-- live form audit (which sites, which action):
SELECT s.domain, p.name, COALESCE(substring(pc.rendered_html from 'action="([^"]*)"'),'(none)')
FROM page_components pc JOIN pages p ON pc.page_id=p.id JOIN sites s ON s.id=p.site_id
WHERE pc.rendered_html ~* 'data-component="contact-form"' ORDER BY 1;
-- the 4 address-less sites (real email resolvable?):
SELECT s.domain,
       COALESCE(NULLIF(s.content_data->>'email',''), NULLIF(ss.data->>'email',''), '(none)')
FROM sites s LEFT JOIN site_specs ss ON ss.site_id=s.id AND ss.aspect='identity' AND ss.is_current
WHERE s.domain IN ('robot-hands.com','relojistas.com','vetcomparison.uk','vonc.com');
```
