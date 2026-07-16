All four decisions from the coordination handoff are resolved, built, validated against the live cluster, and committed — one commit per deliverable with explicit pathspecs, practising the ruling this thread exists
to implement.

1. Pre-dispatch coverage check — built and proven live (4accef4e3). The 090 needs_diagnosis trigger now refuses to dispatch (unless FORCE=1) when any open work item already touches the target, running before the
   intake insert so a refusal writes nothing. It has three probes: a new PAGES= env matched across all sites (the July 16 incident spanned two sites, so keying to the single RUNTIME_SITE would have missed half of it),
   a seed_scope file-overlap probe — which resolves the handoff's open question: the coverage key for code-only diagnoses is the seed_scope file set, at file granularity — plus a non-blocking advisory when seed files
   are dirty locally, and a site-level fallback that blocks only on open items touched in the last 2 hours (older parked backlog is FYI-only, so operators aren't trained to always FORCE). Semantics are copied
   verbatim from silentCoverageClause, including "complete does not cover". Validation: it reproduces the match on the actual json-leak-fix-retry incident rows, and its first real run refused a dispatch at
   relojistas.com after surfacing two different sessions' open items on one page — a genuine collision-in-waiting found immediately.

2. Commit hygiene — adopted via a new repo-root CLAUDE.md (fc773d084). None existed, so every session now loads the practice at start. One refinement over the handoff: commit with an explicit pathspec (git commit
   <paths> -m …), which ignores whatever other sessions have staged in the shared index — I hit exactly that hazard while committing. Enforcement hooks deliberately not built; habit first, per §6.3.

3. Deploy blast radius — structural fix, announcements dropped (0c7b17616). I ruled that the structural fix dominates the doc_notes channel, for the handoff's own reason: a ref-built image cannot bundle WIP, while
   an announcement only works if it's read. make build-agent-chassis-ref REF=<ref> builds from a git archive context (verified complete — the dockerfile needs nothing gitignored; I didn't run the multi-minute docker
   build itself since the default tag is another session's staged release). The default working-tree target is unchanged in behaviour but now prints exactly what it would sweep in — 40 uncommitted changes at test
   time. With per-task commits, "wait for someone else's release" is obsolete: commit your task, ref-build, deploy.

4. File-claims — deferred, per the handoff's own recommendation.

Fittingly, the problem demonstrated itself throughout: three other-session commits landed mid-build (including the v1.0.1125 tag bump), a file I was testing against got committed under me, and MEMORY.md was
rewritten externally. All recorded in RUNNING_NOTES_2026-07-16_coordination_build.md (ccb434803) alongside the handoff, with small follow-ups noted there (084 bare trigger still unchecked by design; ref-builds
cover agent-chassis only). Memory index updated.

----

