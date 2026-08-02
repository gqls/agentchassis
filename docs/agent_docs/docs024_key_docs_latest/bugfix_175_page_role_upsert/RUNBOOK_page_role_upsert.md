# RUNBOOK — `bugs_open/175` / PBP-027

Every command here had a gotcha attached. The gotcha is the reason the line is
here; the command without it is worth much less.

## Is the bug still there? (re-run before trusting any census)

```bash
grep -rn "ON CONFLICT (site_id, name)" --include=*.go
```

**Gotcha:** 175's census lists six sites; this returns eleven. The extra five are
all in the *opposite* camp (they carry `page_type = EXCLUDED.page_type`), so they
do not change the fix — but a session that trusts the bug file's table will
believe the grep was exhaustive when it was a snapshot.

## Exposure — is a collision reachable, and has one happened?

```sql
-- names each CONSTANT-ROLE arm would claim, held under a different page_type
SELECT 'guide-arm' AS arm, s.domain, p.name, p.page_type, p.build_status
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.name LIKE '%-guide' AND COALESCE(p.page_type,'') <> 'blog-post'
UNION ALL
SELECT 'tool-arm', s.domain, p.name, p.page_type, p.build_status
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.name LIKE 'tool-%' AND COALESCE(p.page_type,'') <> 'tool' AND p.name NOT LIKE '%-guide'
UNION ALL
SELECT 'report-arm', s.domain, p.name, p.page_type, p.build_status
FROM pages p JOIN sites s ON s.id = p.site_id
WHERE p.name LIKE 'report-%' AND COALESCE(p.page_type,'') <> 'report'
ORDER BY 1, 2, 3;
```

**Gotcha:** a hit is a *surface*, not a collision. The arm's name is computed
(`CanonicalisePage` gives tool pages `tool-<slug>`, and the guide is
`<page name>-guide`), so a row only becomes reachable if a tool with that exact
name is deployed. Say which of the two you measured. 2026-08-02: 4 rows, all
`deployed`, none reachable by a tool that exists today.

## Is `build_status` the right test for "this page is live"? (No.)

```sql
SELECT COALESCE(build_status,'(null)') AS bs, count(*),
       count(*) FILTER (WHERE deployed_at IS NOT NULL) AS ever_deployed
FROM pages GROUP BY 1 ORDER BY 2 DESC;
```

**Gotcha, and it is the reusable one:** `needs_rebuild` is not "not yet live" —
**35 of 46 such rows carry a non-null `deployed_at`**, i.e. they shipped and are
still being served while they wait to be rebuilt. Any guard written as
`build_status = 'deployed'` is blind to them, which is `bugs_closed/037` in full.

## Testing when the shared working tree does not compile

Another session's uncommitted WIP can break the package for everybody, and
"fixing" it sweeps their work into your commit. Test against committed HEAD plus
only your own files:

```bash
SP=<your scratchpad>
rm -rf $SP/headtree && mkdir -p $SP/headtree
git archive HEAD | tar -x -C $SP/headtree
for f in <your changed files>; do cp "$f" "$SP/headtree/$f"; done
cd $SP/headtree && go build ./platform/... && go test ./platform/orchestration/actions/...
```

**Gotcha:** the extracted tree is **not a git repo**, so anything using
`git ls-files` there silently returns nothing — including an audit script that
then reports a happy zero. Walk the filesystem instead.

## Proving a guard rather than asserting it

```bash
cd $SP/headtree
cp platform/orchestration/actions/page_role_upsert.go{,.orig}
# break exactly one thing, e.g. ADOPT stops writing page_type:
sed -i 's|columnNames(req.Columns), true)|columnNames(req.Columns), false)|' platform/orchestration/actions/page_role_upsert.go
go test ./platform/orchestration/actions/ -run TestUpsertPageForRole_UnshippedPageOfAnotherRoleIsAdopted   # MUST fail
cp platform/orchestration/actions/page_role_upsert.go{.orig,}
```

**Gotcha:** do this in the scratch tree, never in the working tree — a mutation
left behind on a shared branch is a live defect, and another session's `git add -A`
will commit it for you.

## Running the new pattern-check rule over the whole tree

```bash
python3 - <<'EOF'
import importlib.util, os, sys
spec = importlib.util.spec_from_file_location("pc", "scripts/pattern-check.py")
pc = importlib.util.module_from_spec(spec); spec.loader.exec_module(pc)
files = [os.path.join(d,f)[2:] for d,_,fs in os.walk(".") if "/.git" not in d
         for f in fs if f.endswith(".go")]
findings = []; pc.check_partial_page_upsert(sorted(files), None, findings)
print(len(findings), "finding(s)"); [print(" ", w) for _,w,_,_ in findings]
EOF
```

**Gotcha — this is the one that cost me an entry in `WRONG_CALLS.md`:** run it
against a tree that still contains the defect *first*. A detector measured only
after its own fix reports 0 and tells you nothing; 0 has two causes and they have
opposite remedies. Expected: **4 at pristine HEAD, 0 after.**

## Verifying the fix is actually in the running pod

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.spec.containers[0].image}{"\n"}{end}'

for p in <every replica>; do
  kubectl exec -n ai-persona-system $p -- sh -c \
    'strings /app/agent-chassis | grep -c "UpsertPageForRole: refused"'        # positive: expect >0
  kubectl exec -n ai-persona-system $p -- sh -c \
    'strings /app/agent-chassis | grep -c "ON CONFLICT (site_id, name) DO UPDATE SET"'  # negative control
done
```

**Gotcha:** the negative control is NOT expected to be 0 — five other arms
legitimately keep that statement (`page_type = EXCLUDED.page_type`). The honest
negative control for THIS change is the removed guide statement:

```bash
strings /app/agent-chassis | grep -c "DO UPDATE SET$(printf '\\n')\s*title = EXCLUDED.title"
```

which is fragile across binaries — so prefer checking the image tag is one built
**after** the fix commit, and the positive control on **every replica**: a roll
can leave two ReplicaSets running (there were four chassis pods on two different
tags at 22:30 on 2026-08-02).

## Post-roll verification for the RFC_010 round (owed — the fix is committed, not live)

Round 1 is live on `v1.0.1233`. **The RFC_010 round and the predicate correction are
not** — they ship on the next chassis roll, and until then `v1.0.1233` carries the
11-row spurious-refusal surface. Run this on **every replica** once a build later than
commit `4ee695cc1` rolls:

```bash
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o jsonpath='{range .items[*]}{.metadata.name}{"  "}{.spec.containers[0].image}{"\n"}{end}'

for p in <every replica>; do
  kubectl exec -n ai-persona-system $p -- sh -c '
    echo -n "ADDED opt-in refusal   = "; strings /app/agent-chassis | grep -c "did not set AdoptUnshippedRows"
    echo -n "ADDED shared predicate = "; strings /app/agent-chassis | grep -c "deployed_at IS NULL AND COALESCE(build_status"
    echo -n "POSITIVE CONTROL       = "; strings /app/agent-chassis | grep -c "UpsertPageForRole: refused"
    echo -n "REMOVED (expect 0)     = "; strings /app/agent-chassis | grep -c "adopted an unshipped page into this role.*needs_rebuild"'
done
```

**Gotchas, both learned the hard way:**

- **The generic `ON CONFLICT (site_id, name) DO UPDATE SET` grep is NOT a negative
  control** — it returns 4 and that is correct, because five arms keep the statement
  deliberately. A negative control must name a string only the removed code had.
- **`grep -c` exits 1 on zero matches**, so a `set -e` loop dies on the very result you
  are trying to record. The loop above tolerates it; if you rewrite it, keep that.
- The predicate string is a compile-time constant embedded in the binary, so it greps
  cleanly — that is why it is the check for "did the correction ship", rather than
  inferring from the image tag (`bugs_open/153`).
