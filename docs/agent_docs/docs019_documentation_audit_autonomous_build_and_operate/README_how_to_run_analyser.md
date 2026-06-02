I put it in:
scripts/documentation_project/01/


The full pipeline works end to end. 
The assembler read the analyser JSON, pulled ResolveQuery in full by its line range, 
rendered the rest of the package as signatures, 
included the constitution and the hand-fed schema, and produced a clean paste-ready bundle.
Presented 3 files
Presented 3 files
That's the whole thin slice working: analyser, constitution, assembler. 
Here's how to run it on your repo. 
Put the three files in a folder together, 
scripts/documentation_project/01/analyser.go, assembler.go
then:

# 1. analyse the repo (you already did this)
go run analyser.go /home/ant/projects/agentchassis/ > analysis.json

# 2. assemble a bundle for one real task
go run assembler.go \
-analysis analysis.json \
-root /home/ant/projects/agentchassis/ \
-constitution thin_slice_constitution.md \
-task "your task in one sentence" \
-step implementation \
-scope internal/some/pkg/file.go:SomeFunc \
-scope internal/some/pkg/other.go \
-schema schema.txt > bundle.md


-scope is repeatable and takes either a whole file (path.go) or one symbol (path.go:Name). -step is framing, implementation, or debug — implementation and debug show in-scope code in full, framing shows it as signatures when you want intent over detail. -schema is optional; hand-feed it for now if the task touches the database (you can paste the relevant \d output into a text file).


--

Looking at the output, I can see the discovery checks live in the z_context package with specific interfaces and structs that match their check_*.go validators. The key challenge is that with 459 files, the assembler can't bundle everything — it needs to intelligently select only the in-scope files plus their dependencies based on the task description. So the assembler will take the analyser JSON, the constitution, a task description, and a schema, then produce a focused text bundle that includes the target files in full along with relevant signatures from the broader codebase. te into a chat.
Inputs the assembler needs:

The analyser JSON (the full repo structure) — produced by the analyser.
The constitution (the flat markdown file).
A task specification: what's being done, and the target — which files (and/or symbols) are in scope. For the thin slice, the user specifies the in-scope files/symbols (since we don't have automatic target resolution yet).
Optionally: schema (hand-fed text for the first runs).
The step_type (framing/implementation/debug) to decide altitude.

What the assembler produces (the bundle):

Metadata (task, target, step_type, generated_at).
The constitution (always-on).
The task description.
In-scope code: the full bodies of the specified files/symbols (pulled from the actual source files, using line ranges from the analyser for symbol-level, or whole files).
Neighbourhood: signatures of related code. For the thin slice, I'll start pragmatically by including signatures from the same packages as the in-scope files, plus optionally signatures matching a name filter, since computing true references requires call-graph analysis.
Schema (if provided).
Pointers: a note of what else exists (counts) so the reader knows the bundle is a selection.

Altitude (step_type):

framing: include constitution + task + neighbourhood signatures of the whole in-scope package(s) + (light) — emphasize structure, less full code. Actually for framing maybe include the package signatures broadly and not full bodies.
implementation: full in-scope code + neighbourhood signatures + schema. (the default rich code context)
debug: full in-scope code + neighbourhood + schema + (a placeholder for runtime evidence — which we don't have in the thin slice; note it).

For the thin slice, I'll implement the altitude as step_type controlling what gets included: implementation and debug modes show full bodies for in-scope symbols, while framing mode uses signatures only to keep things lighter and focus on intent and structure. Everything includes the constitution, task, neighbourhood signatures, and schema, with debug adding a placeholder section for runtime evidence that would normally appear here. That's a clean, useful bundle. Now I'm thinking through how to extract source code for full bodies — the analyzer JSON has line numbers for each symbol, so I can read the file and grab the exact range. For a whole file, I include everything; for a symbol, just those lines should suffice, though adding the package and import context would help readability. For now I'll keep it simple and show just the symbol's lines with a one-line header noting the file and package. The analyzer JSON is quite large with hundreds of files, so the assembler needs to load it efficiently and...
For the scope path, I need to validate it matches something in the analysis JSON and find the corresponding line range for symbols. The neighborhood concept groups files by package and collects their function signatures and type names, excluding the in-scope symbols themselves. Since packages could get large, I'll cap the neighborhood signatures at around 60 per package with a note about remaining items to keep the bundle manageable. Now I'm ready to write the assembler in Go and test it.
I'll use the sample tree from before with sample/resolver.go and sample/sub/s.go, testing the assembler on a scope like sample/resolver.go:ResolveQuery to see the full bundle with the constitution, task, function body, neighborhood signatures from the sample package, and pointers. The assembler should take the analyzer JSON as input rather than re-running the analyzer itself—that keeps the concerns separated and lets the user run analyzer once, then feed the JSON to the assembler.
Since the analyzer JSON already has everything except full function bodies, the assembler will read those bodies directly from source files by line range, so it needs the -root flag pointing to the source directory. I'll define minimal structs in the assembler to unmarshal the analyzer JSON...
For rendering type "signatures" in the neighbourhood context, I'll keep it simple: just show type <Name> <kind> without trying to summarize fields or methods, since that stays concise and readable. Writing the assembler... Writing the assembler code...
