# HANDOFF — multi-session coordination: work-item collisions, deploy blast radius, shared-file edits, commit hygiene

**Filed:** 2026-07-16, from the "diagnosis fixloop 3" thread, at the owner's request.
**Status:** problem statement + evidence + reusable machinery identified. **No code written.**
**Scope:** a WORKING-PRACTICE and TOOLING problem across concurrent Claude sessions.
It is not a defect in any one workflow. Four related symptoms, one root condition.

**Why this exists as its own thread:** it surfaced inside the diagnosis-loop thread by
costing that thread a real diagnosis run (§2). It is not a diagnosis-loop bug, and
fixing it there would derail that build. Keep it separate.

## Working rules (hold these)
Go, not Python. British English. **Schema first**: read `\d <table>` before SQL, read
the function before changing it. Structural fixes over patches. **Reuse existing
functions** — §6 lists machinery that already exists for three of these four problems.
Go changes are inert until the chassis image is rebuilt; DB config is live immediately.
Verify a deploy by grepping the RUNNING POD's binary, never git, never the tag:
`kubectl exec -n ai-persona-system <pod> -- sh -c 'strings /app/agent-chassis | grep -c "<log string>"'`.
DB: `PSQL="kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db"`.

---

## 1. The root condition

Many Claude sessions work this repo and this cluster **at the same time**, against
**one working tree**, **one branch**, **one image tag sequence**, and **one live
database**. Each session can see its own actions. **No session can see any other
session's actions**, in-flight or committed, except by going and looking — and there is
no agreed place to look.

Every symptom below is that one condition, viewed from a different angle.

## 2. Symptom A — dispatching work at a target another session is already fixing

**This is the one with a measured cost. It happened today.**

The fixloop thread re-verified case `aaa_fails_to_mend/004` against the live pod
(correctly — the standing rule says a filed handoff is a snapshot, re-verify before
dispatch), found 3 article-body pages still broken, and fired a `needs_diagnosis` run
at them. The loop returned **REFUTED / UNVERIFIABLE**.

It refuted because **the evidence was repaired underneath it, mid-run, by another
session**:

| time (UTC, 2026-07-16) | event | actor |
|---|---|---|
| 11:31:13 | `json-leak-fix-retry` needs_page batch created for the 3 broken pages | **another session** |
| 11:36:26 | `needs_diagnosis:envelope-regen-false-complete` dispatched at those same pages | fixloop thread |
| 11:36:50 | gamesdesign `/guides/tool-xp-curve-designer-guide.html` repaired | the retry batch |
| 11:38:05 | diagnosis bundle iteration 1 assembled (reads the rows) | fixloop thread |
| 11:39:02 | finetuning `/guides/llm-cost-calculator-guide.html` repaired | the retry batch |
| 11:39:17 | finetuning `/blog/why-most-ai-projects-fail-…` repaired | the retry batch |
| 11:40:09 | diagnosis bundle iteration 2 assembled | fixloop thread |

The loop read rows that had *just been fixed*, correctly observed the envelope was gone,
and refuted the hypothesis. **The loop behaved correctly and stopped honestly**
(`stopped_by: scope-not-narrowing`, "Hand to a human with the full trail; do NOT
auto-conclude"). The waste was upstream of it: **one diagnosis run's credits spent on a
bug another session was already fixing, five minutes earlier.**

Artefacts, for anyone reconstructing this:
- correlation `781ea4f7-996d-4b41-be0e-96473c4a7996`
- intake `site_work_items.item_key = 'needs_diagnosis:envelope-regen-false-complete'`
- the other session's batch: `SELECT * FROM site_work_items WHERE created_by='json-leak-fix-retry';`

**The trap is subtle and worth stating plainly:** the fixloop thread *did* re-verify
against the live pod. That was not enough. A live-state check is **also a snapshot** —
it tells you what is true now, not what is *about to become* true because another
session already has work in flight. Checking the pod does not check the queue.

## 3. Symptom B — a deploy ships every session's uncommitted work

The chassis image is built by `docker build` **from the local filesystem** (`makefile`
~line 101), not from a git ref. So an image built by session X bundles **the entire
working tree**, including every other session's uncommitted, untested, mid-edit code.
A session that builds an image is silently deploying other people's work.

This is already understood and already costing caution rather than accidents — the
article-body thread *deliberately declined to deploy its own verified fix* for exactly
this reason:

> "Code (branch `085_debug_and_feature_loops`, **NOT yet deployed** — an image would
> bundle other sessions' uncommitted work; ships on the next normal release)"
> — `NOTES_2026-07-15_article_body_json_envelope_ACTIONED.md` §Changes made this session

So the current mitigation is *"don't deploy; wait for someone else's release."* That is
a real cost: verified fixes sit undeployed, and nobody knows whose code is in flight
when a release does go.

Related, already-earned deploy gotchas (do not relearn):
- **Same-tag rebuild ships a stale binary** — the node reuses its cached image. Bump
  `IMAGE_TAG` (makefile line 16), never rebuild an existing tag.
- **Concurrent sessions shipped v1.0.1119 and v1.0.1121 the same afternoon**; v1.0.1119
  lacked the digest code its own thread expected. Always grep the POD binary for your
  symbol before firing anything that depends on it.
- **Rollout order:** image FIRST, then the seed. A seed naming an unregistered action
  fails at runtime.
- **Rebalance window:** never fire an orchestration within ~300s of a chassis pod
  (re)start — the spawn is silently dropped. Hold rollouts behind a cluster-quiet check
  when orchestrations are `AWAITING_RESPONSES`.

**Live right now, as a worked example:** prod runs `v1.0.1123`; the makefile already
declares `IMAGE_TAG ?= v1.0.1124` (committed, `77c6c2a10`). A build is staged. Nothing
tells any session what is in it, or when it lands, or whose WIP it will carry.

## 4. Symptom C — sessions edit the same files with no announcement

Multiple sessions edit one tree. There is no lock, no announce, no "I am in this file".
Collisions are found by accident, usually as a surprising diff.

Indicative, from this session's opening `git status`: 17 modified `kustomization.yaml`
files across every service overlay, plus untracked WIP docs from at least three other
threads (`empty_sections_loop_integrity/`, `travelling_docs/`, `vonc/`), none of it this
session's work. That is the normal steady state, not an incident.

The fixloop thread's own files show the pattern too: `rerender_page_sections_action.go`
is owned by the empty_sections thread, cited by the imagery thread, and read by the
fixloop thread — concurrently.

## 5. Symptom D — commits bundle unrelated sessions' work (owner's explicit ask)

Because every session shares the tree, a broad `git add -A` / `git commit` sweeps up
whatever every other session has left lying around. The standing rule has been to accept
this:

> "**Git: forward-only, no resets/amends.** Many concurrent sessions commit to the same
> branch; your changes may land in another session's commit — that's fine, nothing is
> lost. Check `git log` before assuming your commit is HEAD."
> — `HANDOFF_diagnosis_fixloop_2.md` §6

The evidence that this is now hurting is in the log itself — commits like
`"go links check files and tool test files for vonc and doc traveller"` and
`"product sources, idea audience check, images docs, relojistas sql"` are four unrelated
threads' work in one commit. Bisecting, reverting, or reviewing any single task's change
is then impossible.

**The owner has now ruled against the bundling** (2026-07-16): commit **per task**,
do not bundle everything into one commit for the owner to untangle.

**Practice this implies (the cheap half — adopt immediately, no tooling needed):**
- `git add <explicit paths>` **only**. Never `git add -A`, never `git add .`, never
  `git commit -a`. Add the specific files your task touched.
- One commit per task, with a message naming that task.
- `git status` before committing, and read it — if a path you did not touch is staged,
  unstage it.
- Forward-only still holds: no resets, no amends, no rebases. Another session may have
  committed between your add and your commit.

## 6. Machinery that ALREADY EXISTS — reuse before building

**This is the most important section. Three of the four problems have existing parts.**

### 6.1 The work-item coverage check (Symptom A) — already written, just not wired here

`platform/orchestration/actions/diagnose_silent_check_action.go:252-261` defines
`silentCoverageClause`, which answers exactly the question "is any work item already
touching this page?":

```go
// A page is COVERED (and therefore not silent) if any work item references it —
// by page_id column, by spec->>'page_id', or by item_key segment — in any status
// except the closed ones below.
const silentCoverageClause = `
	NOT EXISTS (
		SELECT 1 FROM site_work_items w
		WHERE w.site_id = p.site_id
		  AND w.status NOT IN ('complete','cancelled','rejected')
		  AND (w.page_id = p.id
		       OR w.spec->>'page_id' = p.id::text
		       OR w.item_key LIKE '%:' || p.name
		       OR w.item_key LIKE '%:' || p.name || ':%')
	)`
```

The silent-check already refuses to emit a finding for a page the immune system is
already handling. **The identical rule is not applied to manual intake** — the
`090_TRIGGER_needs_diagnosis_v1.sh` script writes its item and dispatches with no
coverage check at all. That gap is precisely what cost the run in §2.

Note the deliberate semantics, which are *already correct* for our purpose and should be
copied, not redesigned: **`'complete'` does NOT count as covering** — "a completed remedy
with the violation still observable is precisely the remediation-ineffective signal."
That is exactly the false-complete case the fixloop thread was chasing.

Also existing: `idx_swi_dedup (site_id, item_key) WHERE status not terminal` (dedups
intake by key), and `insertWorkItem`'s two-strike recurrence rule.

**Sketch (not built):** a pre-dispatch coverage check in the 090 trigger — resolve the
target page(s), run the coverage clause, and if anything open references them, print
what/who/when and **refuse to dispatch unless forced** (`FORCE=1`). Cheap, deterministic,
no LLM. It would have printed the `json-leak-fix-retry` batch and saved the run.
**Open question for the thread:** the clause is keyed to pages. A code-only diagnosis
(no `runtime_site`) has no page to check — what is the coverage key there? Possibly the
symbol/file set in `seed_scope`, which nothing currently indexes.

### 6.2 `doc_notes` — an existing, unused-for-this announcement channel (Symptoms B, C)

`doc_notes` already exists and is already the awareness surface the digest and triage
write to. Schema (verified):

```
id | subject_type | subject_key | site_id | body | categories jsonb | source
   | source_agent | source_item_id | created_by | created_at
constraint: subject_type IN ('tool','pipeline')
indexes: idx_doc_notes_categories gin (categories),
         idx_doc_notes_subject btree (subject_type, subject_key, created_at DESC)
```

It is queryable, categorised, indexed by recency, and already read by the digest. A
deploy announcement (`categories: ["deploy"]`, body = tag + symbols + what's in it) or a
file-claim note would need **no new table**.
**Caveat, read before designing:** the `subject_type` CHECK allows only `'tool'` and
`'pipeline'` — a deploy/file-claim note must either map onto one of those or the
constraint must change. Do not assume; decide deliberately.

**Honest caveat on the whole idea:** an announcement channel only works if sessions
*read* it. A note nobody reads is worse than nothing, because it looks like coordination.
Whatever is built must be **pulled into the session at the moment of risk** (pre-dispatch,
pre-build, pre-edit), not merely written. Consider whether a hook (`.claude/settings.json`,
see the `update-config` skill) is the right delivery mechanism rather than trusting each
session to remember to look. That is a genuine design question, not a formality.

### 6.3 Commit hygiene (Symptom D) — practice, not tooling

§5's practice half needs no tooling and can be adopted today. A hook could enforce it
(reject `git add -A`), but the practice comes first — do not build the enforcement
before the habit.

## 7. What this thread should decide

Roughly in dependency order; the first is the one with a proven cost.

1. **Pre-dispatch coverage check in the 090 trigger** (§6.1). Highest value, smallest
   build, reuses an existing verified clause. Resolve the code-only coverage-key question.
2. **Commit hygiene** (§5, §6.3). Adopt the practice immediately — it is free. Decide
   separately whether to enforce it with a hook.
3. **Deploy announcement + pre-build check** (§3, §6.2). What goes in a note; whether
   the builder must announce before `docker build`; whether the "image bundles everyone's
   WIP" problem deserves a structural fix (build from a git ref / a clean checkout,
   rather than the working tree) instead of an announcement. **A structural fix here may
   dominate the announcement channel entirely** — an image built from a committed ref
   cannot bundle anyone's WIP, which makes the deploy blast radius disappear rather than
   become visible. Weigh that before building notes.
4. **File-claim announcements** (§4). Weakest case — highest ceremony, least proven cost,
   and the most likely to be ignored in practice. Consider deferring until 1–3 land.

## 8. Landmines specific to THIS problem

- **Do not fix this inside the fixloop thread.** It is a platform practice concern; that
  thread is mid-build on the loop itself.
- **Anything built here is itself subject to the problem it solves** — you will be
  editing shared files (`090_TRIGGER_needs_diagnosis_v1.sh`, `makefile`,
  `.claude/settings.json`) while other sessions edit the same tree. Add explicit paths,
  commit per task, and re-read `git status` before every commit.
- **The coverage clause is load-bearing for silent-check.** If you refactor it into a
  shared helper, `diagnose_silent_check_test.go` covers its current behaviour — run it.
  Changing its semantics (especially the `'complete'` exclusion) silently changes what
  the silent checker emits.
- **`.claude/settings.json` hooks are user-level config**, not platform code — they ship
  by editing the file, not by rebuilding the chassis. Different deploy path, different
  blast radius.

## 9. Cross-references

- `fixloop_eg_dartsonline/HANDOFF_diagnosis_fixloop_2.md` §6 — the gotchas list this
  supersedes on git bundling; §7 — the 090 intake contract.
- `fixloop_eg_dartsonline/NOTES_running_fixloop(10).md` — the v1.0.1119/1121 concurrent-
  deploy incident.
- `NOTES_2026-07-15_article_body_json_envelope_ACTIONED.md` — the "did not deploy because
  an image bundles others' WIP" decision (§Changes made this session).
- `aaa_fails_to_mend/004_HANDOFF_image_landing_blanks_article_body.md` — the case whose
  diagnosis run was lost to §2; its own §3 records the same "case shifted under us" class.
- `platform/orchestration/actions/diagnose_silent_check_action.go:252-261` — the clause
  to reuse.
- `makefile` line 16 (`IMAGE_TAG`), line ~101 (`docker build` for agent-chassis).
</content>
