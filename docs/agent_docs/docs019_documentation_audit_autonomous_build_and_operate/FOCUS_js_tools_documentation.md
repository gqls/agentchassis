# FOCUS — Documentation & provenance for the existing JavaScript tools

Status: flagged 2026-06-09. Not started. Owner: TBD.

## Why this exists

The platform already has tools written in JavaScript, and they are becoming more
functional and complex. They will need **prose documentation** and **provenance**
in their own right — and they do not have either yet. The only documentation that
exists for them today is indirect: the **site spec** and **plan spec** they were
generated from, plus the build history. That is provenance-of-origin, not
documentation, and it was not written to be read as a reference.

This is tech debt, not a future plan: the tools exist now, so the gap is current.
It is worth a specific, bounded effort rather than folding it into the code-index
work, because the two are different problems with different mechanisms.

## The distinction that shapes the work

Three needs are often conflated under "the JS tools need docs". They are separate:

- **Prose documentation** — human-readable description of what each tool does, how
  to use it, its inputs/outputs and caveats. This is *prose*, so it is handled by
  the documentation-indexing path (`rag_index` into a `knowledge_base` collection,
  retrieved by `rag_lookup`) — which is **language-agnostic**: it indexes text,
  not code, so no JS parser is involved. This is the main gap.
- **Code-symbol provenance** — the symbols (functions/classes) of the JS tools,
  indexed for retrieval and carrying their commit/path provenance. This needs a
  **JS parser**, which is exactly the drop-in the analyser adapter is being built
  for (Go now, JS behind the same `Analyse` seam later). It records provenance via
  `code_symbols` (commit_sha, path, symbol) once JS parsing lands.
- **Origin history** — the site/plan spec the tool was built from. This already
  exists and is a *useful starting point* for the prose docs, but it is not a
  substitute for them.

## What the focus needs to settle

1. **Sufficiency.** Assess whether the existing spec + build history is enough to
   serve as documentation for each tool, or where dedicated prose is needed. Likely
   the spec covers intent but not usage/contract/caveats.
2. **Authoring.** Where it is not sufficient, produce/curate prose docs — possibly
   seeded from the spec/history, then completed by hand or with assistance.
3. **Home.** The docs must live in a **git repo** (the editable source of truth),
   like the guideline docs, so the in-cluster indexer can check them out and index
   them. Decide which repo and what layout (per-tool docs alongside the tool, or a
   docs tree).
4. **Compatibility with what we are building.** Ensure the docs fit the indexing
   scheme: which `collection` they sit in, how they chunk (prose chunker is fine),
   and how they are retrieved into a build's context. And ensure the provenance
   side is compatible with `code_symbols` + the analyser adapter's JS parser when
   that lands (so a tool's docs and its symbols line up under the same identity).
5. **Coverage signal.** A way to tell which tools lack adequate docs (so the gap is
   visible and closeable, not a vague backlog).

## Relationship to the rest of the work

- Reuses the documentation-indexing path (same `rag_index`/`knowledge_base` as the
  guideline docs — see the migration plan's "Documentation indexing" section). No
  new infrastructure for the prose side.
- The provenance side waits on the analyser adapter's JS parser (the polyglot
  drop-in). The adapter is being built Go-first; JS is the next language behind the
  same seam.
- Per the "separate policy, share mechanism" line (plan B4b): JS docs are prose, so
  they share the prose embedding model and the existing path; only the code-symbol
  side needs the JS-specific parser.

## Open questions

- Which repo(s) hold the JS tools and where would their docs live?
- Is there a per-tool manifest/spec already that could be the seed, and is it
  structured enough to generate first-draft prose from?
- Do the JS tools need provenance *before* prose, or together? (Provenance waits on
  the JS parser; prose can start now via the docs path.)
- Should the prose docs and the eventual `code_symbols` rows share a tool identity
  key so retrieval can join "what this tool is" with "its symbols"?
