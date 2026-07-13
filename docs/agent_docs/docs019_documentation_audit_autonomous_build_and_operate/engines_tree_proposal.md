# Proposed tree — analyser / contextkit "engines" docs section

Grounded in what this project actually contains (the contextkit CLI, the
analyser adapter + code_symbols, and the docs that describe them) and in the
clean-up goal: one home per engine, current docs separated from archived
copies, code and its doc adjacent. This is a TARGET to migrate toward with the
dedup tool + manual moves — not a big-bang restructure.

## The principle

Three kinds of thing, kept apart:
- **engine code** — the Go that runs (lives in the module, not under docs/);
- **engine docs** — the current, canonical prose for each engine (one file,
  pointed-to not duplicated);
- **archive** — superseded copies, out of the way and out of the index.

The `docs/.../go_files_old/`, `docubundle/`, `(N).go` sprawl is the third kind
leaking into the first two. The fix is a clear archive root + the dedup tool
keeping it that way.

## Target layout

```
docs/agent_docs/docs019_documentation_audit_autonomous_build_and_operate/
│
├── README.md                      what this area is; index of the engines below
│
├── engines/                       ← NEW: one dir per engine, current docs only
│   ├── README.md                  the engine map (what each does, how they chain)
│   │
│   ├── analyser/
│   │   ├── analyser.md            what it parses, the Output contract, Go-only note
│   │   ├── code_symbols.md        table shape, owner/repo label convention, lookup/index
│   │   └── adapter.md             in-cluster adapter: envelope, topic, deploy → points at 035
│   │
│   ├── contextkit-cli/
│   │   ├── pipeline.md            the harness pipeline (analyse → resolve → embed → fuse → assemble)
│   │   ├── commands.md            one line per cmd: analyser/embed/resolve_targets/fuse/
│   │   │                          eval_targets/assembler/dbcontext/dedup
│   │   └── B4a_embedding_quality.md   the measurement method + decision rule
│   │
│   └── tool-docs/                 (if you keep the tools thread's docs here)
│       └── tool_doc_header.md     the contract → points at 019 canonical
│
├── runbooks/                      ← NEW: the "how to run/deploy it" docs
│   ├── thin_slice.md              (the current RUNBOOK_thin_slice.md)
│   └── analyser_deploy.md         (the in-cluster deploy sequence, if split out)
│
├── go_files/                      ← the LIVE harness source only (already exists)
│   └── contextkit/                the module — analyser/embed/.../dedup + internal/
│
└── _archive/                      ← NEW: everything superseded, NEVER indexed
    ├── README.md                  "archived copies; not canonical; safe to ignore"
    ├── go_files_old/              (moved here from the docs root)
    ├── docubundle/                (moved here)
    ├── thin_slice_run/            (the older assembler.go/analyser.go copies)
    └── <dated-snapshots>/         older doc versions, by date
```

## Why this shape

- **`engines/` with one dir per engine** — the thing the docs were missing: a
  single obvious home for "everything current about the analyser", separate
  from "everything current about the CLI". Each engine doc POINTS at the
  canonical source (035 for the adapter contract, 019 for the tool-doc header)
  rather than restating it — the anti-rot rule from item 24.
- **`_archive/` as the one graveyard** — replaces the scattered `go_files_old/`,
  `docubundle/`, `(N).go` copies with a single excluded root. The dedup tool's
  default archive is exactly this dir; the analyser's `-exclude _archive/` (one
  entry) then keeps the index clean instead of the six-pattern exclude list.
- **`go_files/contextkit/` stays the live module** — it already exists and the
  README map points there; don't move it, just keep its archived copies out.
- **`runbooks/` separate from `engines/`** — "how it works" (engine docs) vs
  "how to run/deploy it" (runbooks) are different audiences and rot at
  different rates; splitting them stops a deploy-step change from touching the
  conceptual doc.

## Migration order (with the tools you now have)

1. **Report first** — `dedup docs/.../docs019… -near -ext .go,.md` (report only)
   to see the real duplicate landscape before moving anything.
2. **Archive the duplicates** — `dedup … -move` (exact first; review `-near`
   groups, then `-move -near`). Everything lands in `_archive/` with a manifest.
3. **Create `engines/` + `runbooks/`** and move the current canonical docs in
   by hand (this is editorial — which doc is canonical is your call, not a
   script's).
4. **Re-point** — fix internal links to the moved files; update the README map.
5. **Re-index** — rebuild `chassis.json` with `-exclude _archive/` (now a single
   exclude), confirm the duplicate-survivor check reports zero.

Steps 1–2 are mechanical and reversible (the dedup tool). Step 3 is the
human-led consolidation — and per the earlier discussion, do it as
selective-carry-with-the-LLM-as-assistant, never a generative merge.
```
