# RUNBOOK — bugfix 303 (commands that were hard to get right)

## Export component HTML from the live DB without mangling it

Base64 per row, with `translate(..., E'\n', '')` because Postgres `encode(...,'base64')` inserts a
newline every 76 chars, which shreds line-oriented parsing:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db \
  -t -A -F'|' -c "SELECT id, component_level, is_active, name,
    translate(encode(convert_to(html_template,'UTF8'),'base64'), E'\n', '')
    FROM content_components WHERE html_template IS NOT NULL" > components.psv
```

**Gotcha: a large export dies mid-stream** (`error reading from error stream: unexpected EOF`) and
what you get is a TRUNCATED file that looks complete — count the rows against a `SELECT count(*)`
before trusting it. Fix: gzip inside the pod so the exec stream carries less:

```bash
kubectl -n ai-persona-system exec -i postgres-clients-0 -- sh -c \
  "psql -U clients_user -d clients_db -t -A -F'|' -c \"...\" | gzip -c" > out.psv.gz
```

## Run repo code against exported data without touching the repo

Scratchpad Go module with a `replace` directive — no temp files inside the repo for other sessions
to sweep:

```
module calib
require github.com/gqls/agentchassis v0.0.0
replace github.com/gqls/agentchassis => /home/ant/projects/agentchassis
```

Then import `platform/content` and inline the OLD predicate (10 lines) for the A/B. Harness:
`<scratchpad>/calib/main.go` (session af6f5076); results in the bug file's fix record.

## Prove a commit compiles at HEAD without other sessions' WIP

`git stash` is FORBIDDEN on this tree. Instead:

```bash
git worktree add --detach $W HEAD && cp <your files> $W/... && (cd $W && go build ./platform/...) \
  && git worktree remove --force $W
```

**Gotcha:** `git worktree remove` while `cd`'d inside it leaves the shell with no cwd — run the
remove from outside, or expect a harmless `getcwd` error.
