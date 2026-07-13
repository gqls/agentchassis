# Guide — using a context pack to resume and ship a project

Plain walk-through: how to take one of the context packs, start a fresh chat, gather what the task needs, do the work, and ship it. The packs, your script, and our tools can blur together, so this starts by separating them.

---

## The big picture

For any task there are three separate things. Keep them apart and the rest is simple:

1. **The context pack is a recipe, not a tool.** It's a short doc that says, for one task: where it stands, the exact next thing to do, which code and docs to attach, and which live data to pull. It doesn't *collect* anything — it tells you what to collect.
2. **Gathering is collecting that code and data** into one place so you can paste it into the fresh chat. You already have a good way to do this — your `package_*.sh` script. Our analyser/assembler are an *alternative* way, for when you want it leaner. (More below — this is your actual question.)
3. **Shipping is putting the finished change live** — running the SQL, rebuilding the chassis image, swapping the idea.uk binary. Different per project.

So the loop is: **open a chat with the pack → gather the context → do the work → ship → check it worked.**

---

## Your main tool for gathering: the `package_*.sh` script

`package_page_build_debug.sh` already does the whole gathering step for a chassis task. It:

- concatenates a hand-picked set of directories and files (the "page-build-debug" subsystem) into one text file, and
- appends a **read-only live capture** — the `\d` schemas, the decisive skinner-box queries, the agent-definition workflows, and pod/cronjob state.

That is the bundle. For the skinner-box task you ran it, got the 1.7 MB file, and that file already contains everything the chat needs. **So for resuming skinner-box, the script is all you need — you don't have to touch the analyser or assembler.**

Its one trade-off is that it grabs **whole directories**, so it over-includes: test files, an example doc, and unrelated utilities ride along, which is most of why it's 1.7 MB (110 files when the task really turns on about 14). But note the flip side, which is real and in your favour: because it deliberately keeps the shared and registration layers for reuse-discovery, it **caught `registry.go` and the page-build action neighbours** — files the narrower context pack named only by concept. Broad gathering is wasteful but thorough.

---

## When would I use the analyser and assembler instead?

They are the **leaner alternative to the script's code section** — they pull only the functions you name plus the code those functions actually call (the call-graph neighbourhood), instead of whole directories. You'd reach for them when the script's output is too big or noisy and you want a tight, focused bundle.

How they run — two steps:

```
# Step 1 — once per repo. Read the whole tree into a structured index.
go run analyser.go /path/to/chassis > analysis.json

# Step 2 — per task. Pull just the in-scope code + its neighbours into a bundle.
go run assembler.go -analysis analysis.json -root /path/to/chassis \
  -constitution thin_slice_constitution.md \
  -task "one line describing the task" \
  -scope platform/orchestration/actions/load_page_sections_from_spec_action.go \
  -scope platform/orchestration/actions/plan_sections_action.go \
  -doc 016_debugging_guide.md > bundle.md
```

The **analyser** is the "read the codebase" step — run it once, re-run when the code changes. The **assembler** is the "pick what this task needs" step — run it per task, naming the files (or `file.go:FunctionName`) in `-scope`.

**For the skinner-box task specifically:** you've already gathered with the script, so you don't need these to carry on. Their only job here would be to *trim* that 1.7 MB — point `-scope` at the handful of files that matter (the page-build-handler workflow, `load_page_sections_from_spec`, `plan_sections`) and you'd get a far smaller bundle.

**But know the gap before you rely on them:** because the assembler follows function *calls*, it would **miss `registry.go`** — registration happens through an init/registry mechanism, not a call from the handler, so the call graph doesn't reach it. Your directory-walking script catches it; the assembler (today) doesn't. So if you use the assembler, add wiring files like `registry.go` with an extra `-scope`, or stick with the script when completeness matters more than size.

In one line: **script = broad, thorough, large; assembler = narrow, lean, but currently blind to wiring files.** First pass on an unfamiliar area → script. Trimming a known area → assembler, plus the wiring files by hand.

---

## Pulling live database state on its own (`dbcontext`)

Your script already captures the live DB state, so for these tasks you mostly don't need this. `dbcontext` is just the tool-native way to do the same thing if you ever gather *without* the script:

```
go run dbcontext.go -psql 'kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db' \
  -schema site_plan_sections,pages,site_work_items > schema.md          # \d for each table
go run dbcontext.go -psql '…same…' -rows "SELECT … WHERE site_id='…'" > rows.md   # rows, with size limiting
```

idea.uk has no database, so skip it there.

---

## Shipping the work: what "go live" means per project

Once the chat has produced a change, here's how it actually ships. Know which target you're touching — the **chassis platform** (the Go agent system, a container image running in Kubernetes) is a different thing from the **websites it builds** (static files sent to Backblaze).

### Chassis tasks (adoption, thunder, imagery) — up to three steps

1. **If it's a database change** — apply the SQL through psql, snapshot first:
   `kubectl exec -n ai-persona-system postgres-clients-0 -- psql -U clients_user -d clients_db -f change.sql`
   (Snapshot first: `CREATE TABLE …_bak_tag AS SELECT …`. Resolve ids fresh — `site_id` changes on every teardown, so use `(SELECT id FROM sites WHERE domain='…')`.)
2. **If it's a Go code change** — it only takes effect once the chassis image is rebuilt and rolled out: build and push the image, bump the image tag (e.g. from `v1.0.1057`) in `kustomization.yaml`, apply, then watch `kubectl -n ai-persona-system rollout status deploy/<name>`. Use the targets in your `makefile.txt` — confirm them there, don't guess. (Bump the tag or the rollout may reuse the old image; re-tar artifacts or a byte-identical file won't ship.)
3. **Then make it run** — either insert a work-item (the dispatch loop picks it up) or fire a kcat orchestrate trigger, depending on the task. (Adoption = a `needs_content_page` work-item; thunder = a `model-trainer` orchestrate trigger.)

The generated **website** then ships by itself: a page that reaches `build_status='deployed'` goes git → GitHub Actions → Backblaze. Check the live page renders.

### idea.uk — one step, on its own box

No database, no Kubernetes, no Backblaze. Build for Linux, copy the binary up, swap it in place, restart:
`GOPROXY=off GOTOOLCHAIN=local GOOS=linux GOARCH=amd64 go build -o idea . && scp … && ssh … 'mv -f … && systemctl restart idea'`
(Use `mv -f`, not `cp` — the running binary is busy; the `&&` chain stops a failed build shipping a stale binary. The page is embedded, so any page edit needs this rebuild.) Email and forwarding are cPanel steps. **The exact commands and the env block are in the idea.uk pack** — follow them there.

---

## Always, whatever the task

- **Confirm the decisive fact before acting.** Each pack names one (are the section rows actually read? does `CurrentStep` hold the loop name? is `service.go` building?). The packs restate older context and can be stale — the fresh capture from your script is the truth.
- **Snapshot before any database change**, and verify a write by re-querying, not by assuming it took.
- **"Complete" does not mean "it worked."** Check the artifact — components present, page deployed, manifest written, email received — not just a status of `complete`. (Skinner-box is exactly this: the work-item shows `complete`, yet there are zero sections and zero components.)
