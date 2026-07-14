| FIX-001 | Diagnosis→fix loop programme / council loop (F0–F3) | deployed | End-to-end symptom-to-PR pipeline; PR #1 merged 2026-07-13 | fix-loop.md |
| FIX-002 | fix-proposer agent / constrained edit plan (F1.1a) | deployed | Read-only agent drafts ≤8-edit plans from CONFIRMED diagnoses | fix-loop.md |
| FIX-003 | needs_diagnosis intake route (F0.1c) | deployed | Durable site_work_items intake joined to loop by correlation_id | fix-loop.md |
| FIX-004 | Superseded: null-site-allowed intake design | superseded | Null-site design proved schema-impossible, replaced | fix-loop.md |
| FIX-005 | diagnosis_artifacts table (unified egress store) | deployed | Correlation-keyed egress table, kind grows bundle→escalation | fix-loop.md |
| FIX-006 | Retention/expiry knob on diagnosis_artifacts | aspirational | expires_at/pinned columns exist; no sweep job built | fix-loop.md |
| FIX-007 | Known-answer benchmark methodology | deployed | Blind reruns scored against pre-registered rubric | fix-loop.md |
| FIX-008 | Dartsonline guides pilot selection history | deployed | Three candidates dropped before confirmed differential pilot | fix-loop.md |
| FIX-009 | Blinding discipline for benchmark runs | deployed | Diagnose-agent can't read docs; symptom string is the only leak | fix-loop.md |
| FIX-010 | Standing hypothesis refuted (reconcile_site_plan routing table) | superseded | Loop refused to confirm the wrong-file hypothesis | fix-loop.md |
| FIX-011 | Two intake paths disagreement | deployed | WriteBuildItemsAction and reconcile_site_plan skip different page types | fix-loop.md |
| FIX-012 | mark_no_sections — referenced-but-never-built step | abandoned | Remedy step named in a comment, never implemented | fix-loop.md |
| FIX-013 | Plan validation / hard allowlist for edit plans | deployed | Deterministic validator fails closed on malformed plans | fix-loop.md |
| FIX-014 | Two-reviewer council (F2.1) | deployed | edit-quality + guardian reviewers, guardian holds veto | fix-loop.md |
| FIX-015 | Deterministic council decision + hard veto | deployed | Pure Go aggregation of reviewer verdicts | fix-loop.md |
| FIX-016 | Hard-veto flag at multiple scopes — early design | superseded | Multi-scope veto design narrowed to single guardian flag | fix-loop.md |
| FIX-017 | Revise loop (F2.2) | deployed | Capped repropose cycle on revise decisions | fix-loop.md |
| FIX-018 | Decision router (F2.3) | deployed | Full approved/revise/reframe/escalate router | fix-loop.md |
| FIX-019 | Verify step (diagnose_run_checks) | deployed | Reviewer SQL checks run under read-only containment | fix-loop.md |
| FIX-020 | Schema hint for reviewers (F2.3b(a)) | deployed | Live schema hint fixed hallucinated-column check failures | fix-loop.md |
| FIX-021 | Reframe step (post-veto) | partial | Unit-tested reframe path never fired live since v4 | fix-loop.md |
| FIX-022 | Escalation as first-class success terminal | deployed | complete_escalated treats architecture dead-ends as success | fix-loop.md |
| FIX-023 | Write step / fix-implementer agent (F1.1b(c)) | deployed | Plan→branch→PR write organ, proven live 2026-07-13 | fix-loop.md |
| FIX-024 | Hard file allowlist (diagnose_prepare_fix_commit) | deployed | Rejects any file outside the approved plan before git | fix-loop.md |
| FIX-025 | Build gate (diagnose_build_gate) | deployed | Containerized gofmt+build gate; red blocks the PR | fix-loop.md |
| FIX-026 | git_adapter_request generic adapter caller | deployed | One allowlisted-verb caller for all git-adapter ops | fix-loop.md |
| FIX-027 | isRepoCloningAgent spawn gate / token injection | deployed | Read-only GitHub token injected into dedicated pods | fix-loop.md |
| FIX-028 | diagnose_read_repo_files action | deployed | Fetches plan's current file bodies; modify-404 is a hard error | fix-loop.md |
| FIX-029 | fix-implementer-orchestrator (dedicated-pod wrapper) | deployed | Thin wrapper fixed a shared-pod spawn-gate bypass bug | fix-loop.md |
| FIX-030 | Whole-file rewrite strategy | deployed | Complete file bodies only, no diffs; caps near ~41KB | fix-loop.md |
| FIX-031 | PR as human terminal / nothing merges itself | deployed | Merge is permanently human across the whole fix-loop | fix-loop.md |
| FIX-032 | Fork isolation / NO FORK decision | superseded | Fork-isolation proposal raised then explicitly closed | fix-loop.md |
| FIX-033 | Round-counting scope bug (correlation vs orchestration) | superseded | Fixed in source; one-cycle deploy gap via same-tag trap | fix-loop.md |
| FIX-034 | fixloop-digest / awareness surface | partial | 24h activity digest built, awaiting chassis image | fix-loop.md |
| FIX-035 | Owner standing rule: awareness before autonomy | deployed | Awareness must precede any council/roster widening | fix-loop.md |
| FIX-036 | Council roster expansion vision | aspirational | Guidelines/reuse/bug-historian/compliance bench never built | fix-loop.md |
| FIX-037 | Architecture-change visibility (Q-E signals / detector) | partial | No formal detector; guardian's informal judgement substitutes | fix-loop.md |
| FIX-038 | Guardian veto surfacing an architecture-level fix | deployed | Live instance: guardian caught a disguised architecture change | fix-loop.md |
| FIX-039 | Platform-not-site-data fix philosophy | deployed | Owner ruling: fixes must target platform, not one site's data | fix-loop.md |
| FIX-040 | config_change edit operation type | deployed | Config-only edits labelled but left for a human to apply | fix-loop.md |
| FIX-041 | F1.2 deferred work items | aspirational | ref/base as input, fix_pr artifact, diff strategy all deferred | fix-loop.md |
| FIX-042 | F3 learning layer: bug_records + guideline side-tasks | unknown | Taxonomy/amendment mechanism designed, build status unconfirmed | fix-loop.md |
| FIX-043 | Q-G reviewer context (answered narrowly) | partial | Reviewers share one role prompt; no per-reviewer corpora yet | fix-loop.md |
| FIX-044 | Q-H human-facing result package | deployed | PR body carries diagnosis + plan + council decision together | fix-loop.md |
| FIX-045 | SEED_first_writestep_diagnosis / seeded-bug strategy | deployed | Hand-authored CONFIRMED row exercised the real write chain | fix-loop.md |
| FIX-046 | F0.3 per-iteration notes / doc_notes reuse | partial | Terminal note wired; per-iteration rows never fully landed | fix-loop.md |
| FIX-047 | Loop-worthiness test doctrine (five-criteria) | deployed | Pre-registered intake test applied three times in founding thread | fix-loop.md |
| FIX-048 | Hard deterministic gates between every LLM step | deployed | Plain Go gates decide what proceeds; models only propose | fix-loop.md |
| FIX-049 | Fix-loop value proposition: unattended, cited, consistent | deployed | Differentiator is auditable consistency, not superhuman insight | fix-loop.md |
| FIX-050 | Transferable machinery: legacy-migration and feature intakes | aspirational | Same gate/council scaffolding proposed for other intake types | fix-loop.md |
| AGOV-001 | Confirm-not-initiate + single central confirmer | aspirational | One component applies all proposed→active transitions | autonomy-governance.md |
| AGOV-002 | config_work_items contract | aspirational | Tenant-scoped mirror of site_work_items for config proposals | autonomy-governance.md |
| AGOV-003 | Decision log (premise vs rule_trace; inputs_used) | aspirational | Append-only reasoning log for drift detection and audit | autonomy-governance.md |
| AGOV-004 | Trust ledger + bidirectional ratchet | aspirational | Per-tenant-capability trust level with asymmetric mutation | autonomy-governance.md |
| AGOV-005 | Capabilities catalog: ceiling on the capability | aspirational | Blast-radius-capped trust ceiling, never full autonomy for chassis edits | autonomy-governance.md |
| AGOV-006 | Change-layer integration (change_events, in_band) | aspirational | Typed triggers fan out from a change-event table; in_band closes self-mod loop | autonomy-governance.md |
| AGOV-007 | Two gated paths: config changes vs deliverables | aspirational | Config confirmer and ledger ratchet are two distinct gates | autonomy-governance.md |
| AGOV-008 | The outcome-record gap | aspirational | No signal records whether a deliverable actually succeeded | autonomy-governance.md |
| AGOV-009 | Thin vertical slice before six-contract infrastructure | deployed | Minimal bundle harness built and used before any contract shipped | autonomy-governance.md |
| AGOV-010 | External rollback + recursive self-improvement risk | aspirational | Rollback must not depend on the agents it rolls back | autonomy-governance.md |
| AGOV-011 | Morality review as configured, layered standard | aspirational | Operator-chosen base standard, not a baked-in moral view | autonomy-governance.md |
| AGOV-012 | Contributors vs checkers | deployed | Build-path reviewers vs deployed-site monitors, settled distinction | autonomy-governance.md |
| ABO-001 | trust ledger — earlier draft | aspirational | Earlier draft of the same contract now live as AGOV-004 | autonomous-build-operate.md |
| ABO-002 | change-layer integration contract — earlier draft | aspirational | Earlier draft of the same contract now live as AGOV-006 | autonomous-build-operate.md |
| ABO-003 | context substrate model (authored vs derived) | aspirational | Authored=owned/fallible vs derived=no-owner readout framing | autonomous-build-operate.md |
| ABO-004 | mediator model for competing design concerns | aspirational | Requirement-relative balance among fast/secure/simple/etc. | autonomous-build-operate.md |
| ABO-005 | governance/HITL principles | aspirational | Confirm-not-initiate, sealed-ancestor inheritance, one privileged transition | autonomous-build-operate.md |
| ABO-006 | autonomous-system building-block hardening checklist | aspirational | Cycle guards, outbox transactions, bulk confirmation, no self-recovering rollback | autonomous-build-operate.md |
| ATM-001 | Trust ratchet & capability ceiling model | aspirational | Shared preamble framing behind AGOV-004/ABO-001 | autonomy-trust-model.md |
| ATM-002 | Requirement-mediation model ("right" as balance) | aspirational | Same balance framing as ABO-004, from the shared preamble | autonomy-trust-model.md |
| RSN-001 | Chain-of-thought prompt pattern catalog | unknown | Five CoT archetypes; no confirmed wiring into any agent prompt | reasoning.md |
| RSN-002 | Salience over presence (context bundle) | aspirational | Attention follows the concrete, not mere bundle presence | reasoning.md |
| RSN-003 | Four axes governing a development step | aspirational | Purpose/How-well/Where-heading/What-is; trajectory was the gap | reasoning.md |
| RSN-004 | Why-chain (objective-tree traversal) | aspirational | Root-to-node purpose path used as an anti-drift question | reasoning.md |
| RSN-005 | Direction-of-travel (trajectory layer) | aspirational | Fast-churn heading layer, freshness-stamped, human-confirmed | reasoning.md |
| RSN-006 | Step-type-aware prompt composition (altitude-aware) | aspirational | Framing gets full why-chain; generation collapses to a tether | reasoning.md |
| RSN-007 | Checker model (single-axis parallel checkers) | aspirational | Narrow parallel checkers reconciled by singular arbitration | reasoning.md |
| RSN-008 | Multi-author generation | aspirational | Each concern authors a full solution instead of guarding one | reasoning.md |
| RSN-009 | Mediator as multi-objective optimiser | aspirational | Balance point among authored extremes via priority profile | reasoning.md |
| RSN-010 | N-round convergence (author/checker modes) | aspirational | Rounds shrink the active concern set; non-convergence escalates | reasoning.md |
| RSN-011 | N-round candidate ownership | aspirational | Shared seeded candidate, changed only by adjudicated proposals | reasoning.md |
| RSN-012 | Self-development coding pipeline — positions A/B/C | aspirational | Unresolved cross-area coordination model; lean toward spawn-fresh mediator | reasoning.md |
| INVD-001 | Abandoned "no owner" claim (checked and found false) | abandoned | tool-recreation-handler already owned the responsibility claimed missing | investigation-discipline.md |
| INVD-002 | Verify-before-acting investigation discipline | deployed | Don't act on a theory until code-search verifies it | investigation-discipline.md |
| OPD-001 | Standing evidence rules (working-method contract) | deployed | correlation_id reads, snapshot-before-UPDATE, 0-rows skepticism | operating-doctrine.md |
| OPD-002 | Parallel-thread boundary and handoff convention | deployed | Declared ownership boundaries and handoff docs across concurrent threads | operating-doctrine.md |
| OPP-001 | In-chassis replicability requirement for operator work | deployed | Off-platform actions must map to chassis ops or a named gap | operator-practice.md |
| OPP-002 | Operator discipline: verify-by-artifact, dated backups, kcat | deployed | Never trust a status; diff bytes, dated backups, kcat trigger convention | operator-practice.md |
