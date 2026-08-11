# RUNBOOK — bugfix_153_build_provenance

Every command here was run, or is the exact command owed next. Gotchas are attached to the
command that has them, not collected at the end.

---

## R1 — Is the defect still live? (the pre-flight, and the post-roll close condition inverted)

```bash
POD=$(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[0].metadata.name}')
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -cE 'v1\.0\.1[0-9]{3}'"
kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -cE '\b[0-9a-f]{40}\b'"
```

Both returned **0** on `v1.0.1279`, 2026-08-10 — i.e. the binary cannot say what built it.

> **GOTCHA — `grep -c` returning 0 makes `kubectl exec` exit 1**, which reads as a failed
> command rather than a measured zero. You will see `command terminated with exit code 1` next
> to the `0`. That is the ANSWER, not an error. Do not "fix" it by dropping `-c`; append
> `|| true` if the noise bothers you, but read the number.

Two zeros here is the bug. **After the roll, the second one must become ≥1** — that is the
close condition, and R5 is the discriminating version of it.

---

## R2 — Prove the ldflags mechanism WITHOUT touching the cluster (do this before committing)

The trap this avoids: `-X` against a package the binary does not import is **silently
ignored**, so a change can be complete, compile, ship, and stamp nothing.

```bash
CTX=$(mktemp -d /tmp/verify153.XXXXXX)
git archive HEAD | tar -x -C "$CTX"          # committed state only — see gotcha
mkdir -p "$CTX/pkg/buildinfo"
cp pkg/buildinfo/buildinfo.go "$CTX/pkg/buildinfo/"
for d in agent-chassis auth-service core-manager reasoning-agent web-search-adapter \
         web-scrape-adapter git-adapter image-generator-adapter thunder-adapter \
         analyser-adapter browser-runner-adapter content-creator-agent \
         remote-job-spawner scheduler; do cp "cmd/$d/main.go" "$CTX/cmd/$d/main.go"; done
cd "$CTX"

FAKE=deadbeefcafe1234567890abcdef1234567890ab
go build -ldflags "-X github.com/gqls/agentchassis/pkg/buildinfo.GitCommit=$FAKE" \
   -o /tmp/ac_stamped ./cmd/agent-chassis
strings /tmp/ac_stamped   | grep -c "$FAKE"     # POSITIVE  -> 3
go build -o /tmp/ac_unstamped ./cmd/agent-chassis
strings /tmp/ac_unstamped | grep -c "$FAKE"     # NEGATIVE  -> 0
strings /tmp/ac_unstamped | grep -cx unknown    # default   -> 1
```

Measured 2026-08-10: **3 / 0 / 1**. The negative control is the whole point — without it, a
positive proves only that the string exists somewhere, not that *your flag* put it there.

> **GOTCHA — you MUST build from `git archive HEAD`, not the working tree.** This tree is
> shared by ~40 concurrent sessions and is frequently uncompilable through no fault of yours:
> on 2026-08-10 `go build ./cmd/agent-chassis` failed on
> `platform/orchestration/datahelpers/page_canonical.go:185: undefined: nestedOrFlatURL`,
> another lane's half-finished edit. Building from HEAD + your own files isolates your change
> and tells you whether **you** broke anything. Do not conclude your change is broken from a
> working-tree build failure, and do not "fix" the other lane's file.

> **GOTCHA — the two bare-`main.go` file-builds need checking separately.** `git-adapter` and
> `remote-job-spawner` compile `cmd/X/main.go` by filename rather than a package path, which
> is the case most likely to drop a linker flag. Verified: 3 occurrences each, same as the
> package builds. `cmd/scheduler` too (it is `kafka-scheduler`'s source — the directory name
> does **not** match the service name).

---

## R3 — Does the makefile still expand correctly? (syntax check, no build)

```bash
make -n build-agent-chassis      | tail -1     # ref build
make -n build-agent-chassis-tree | tail -1     # tree build
make -n verify-agent-images      | tail -8
```

Expect `--build-arg GIT_COMMIT=$GIT_COMMIT --label org.opencontainers.image.revision=...` in
the first two, and the sha computed **once** inside the existing `&&` chain.

> **GOTCHA — `$$` in the makefile, `$` in the expansion.** The grep anchors in
> `verify-agent-images` are written `"^[0-9a-f]{40}$$|-tree$$"` and `make -n` prints them as
> `$`. If you see `$$` in the *printed* recipe you have double-escaped.

---

## R4 — Hand the release to the owner. DO NOT roll a single service yourself.

Releases on this estate are **whole-fleet, one tag, owner-run**. A single-service build+push
at its own tag fragments the fleet: every other service's overlay then points at a tag it was
never built at, and the next `deploy-agents` either heals it silently or breaks thirteen
services with `ImagePullBackOff`.

Ask the owner to run (the `!` prefix runs it in-session):

```
! date; make release redeploy-agents ENVIRONMENT=production REGION=uk001; date
```

`IMAGE_TAG` is already at **v1.0.1280** (`makefile:17`) — bumped by another session and
carried in this lane's commit `1054ec36c` as a declared passenger. Confirm it is still ahead
of the last rolled tag before asking; another session may have rolled in the meantime.

---

## R5 — Verify at the pod AFTER the roll (this is the close condition)

**The sha to grep is the sha the image was built from, which is `HEAD` at build time — not
this lane's commits.** Get it from the operator's build output, or take `git rev-parse HEAD`
at the moment the build ran. This lane's last code commit is `1054ec36c`; anything at or after
it carries the fix.

```bash
SHA=$(git rev-parse HEAD)     # or the ref the owner actually built
for POD in $(kubectl -n ai-persona-system get pods -l app=agent-chassis -o jsonpath='{.items[*].metadata.name}'); do
  echo "== $POD"
  kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c '$SHA'"          # EXPECT >=1
  kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c 'deadbeefcafe'"  # NEGATIVE, expect 0
  kubectl -n ai-persona-system exec "$POD" -- sh -c "strings /app/agent-chassis | grep -c orchestration"   # POSITIVE CONTROL, expect large
done
```

**EVERY replica, not one** — a label selector can pick one pod of N and a partial roll looks
identical to a complete one.

Then the mechanism's own report:

```bash
make verify-agent-images     # new stanzas print the image label + the pod's baked sha
```

And the startup log line, which is the second, independent witness:

```bash
kubectl -n ai-persona-system logs deploy/agent-chassis --tail=200 | grep 'build provenance'
```

> **GOTCHA — `logs deploy/X` reads ONE pod of N.** For a per-replica read, loop over pod names
> as above.

---

## R6 — The induced-fault test — **SUPERSEDED 2026-08-11 by owner decision; do not re-file it as owed**

> **The production induced-fault test below was NOT run and is no longer owed.** Two things
> changed on 2026-08-11. First, `v1.0.1284` produced a **real, un-induced divergence** — one tag,
> three source revisions (`bugs_open/249`) — which demonstrated the stamp making an invisible
> discrepancy visible, in production, without anyone having to lie to production to see it.
> Second, the owner chose the **local regression guard** (below, and now R9) as the substitute.
> What remains formally untested is narrow and close to tautological: that an image nobody
> rebuilt keeps the sha baked into it at build time. **Recording the decision here because an
> outstanding-proof item that quietly disappears is exactly how a lane's honesty rots** — this
> one was closed by a choice, on a date, for stated reasons, not by being forgotten.

The original specification, kept because it is still the right test if the question ever reopens:

Everything in R5 passes if the mechanism works **and** the roll was honest. It does **not**
prove the mechanism can catch a dishonest roll. That needs the fault induced:

1. Bump `IMAGE_TAG`.
2. Run push + deploy **WITHOUT** `build-*`.
3. The pod must come up wearing the NEW tag while still reporting the **OLD** sha.

That divergence is the entire bug, made visible for the first time. **Expect it to be
visible, not refused** — this change detects; refusing is `bugs_open/153` candidates 2/3,
deliberately deferred.

Run the full honest cycle immediately afterwards so production is not left on a lying tag.

**Regression guard**, cheap and worth doing once:

```bash
make build-agent-chassis REF=<an-older-commit> IMAGE_TAG=<scratch-tag>
docker image inspect docker.io/aqls/agent-chassis:<scratch-tag> \
  --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'   # must be the OLDER sha
docker rmi docker.io/aqls/agent-chassis:<scratch-tag>
```

---

## R7 — Council: check whether the round is alive before you wait for it

```sql
SELECT current_step, status, collected_data->'__step_error'->>'message'
FROM orchestration_states
WHERE collected_data->'input_data'->>'fix_correlation_id' = '44fa6a98-acaa-46b5-9ada-f0c34ca5475d';
```

This lane's round returned `complete_invalid` + *"You have reached your specified API usage
limits… regain access on 2026-09-01"*.

> **GOTCHA — a dead council run is indistinguishable from the ~30-minute queue latency
> CLAUDE.md tells you to expect, and the standing advice is "do not retry on that evidence".**
> Correct in general; wrong during a provider outage. `complete_invalid` is the council's
> generic "I could not run" state; `status` reads `COMPLETED` and the top-level `error` column
> is NULL. Run the query above **before** waiting and before resubmitting. Full account:
> `bugs_open/243`, and the `LANDMINES.md` entry *"An API USAGE-LIMIT death looks exactly like
> a transient seat fault"*.

Is the fleet back?

```sql
SELECT max(created_at) FROM llm_call_log WHERE success;
```

If that is newer than the moment the owner changed the cap, resubmit:

```bash
RESUBMIT_CORR=44fa6a98-acaa-46b5-9ada-f0c34ca5475d \
  ./docs/agent_docs/docs024_key_docs_latest/fixloop_eg_dartsonline/097_TRIGGER_council_review_v1.sh \
  docs/agent_docs/docs024_key_docs_latest/bugfix_153_build_provenance/SUBMISSION_2026-08-10_build_provenance.json
```

Passing `RESUBMIT_CORR` keeps the artifact trail in one place. **Do not resubmit while the
provider is refusing calls** — it buys an identical failure.

---

## R8 — Census: which live agents/steps use which provider (the query I got wrong first)

```sql
SELECT s.v->'config'->'ai_service'->>'provider' AS provider,
       count(*) AS steps, count(DISTINCT type) AS agents
FROM agent_definitions, LATERAL jsonb_each(default_config->'workflow'->'steps') s(k,v)
WHERE is_active AND COALESCE(is_snapshot,false)=false AND deleted_at IS NULL
  AND s.v->'config' ? 'ai_service'
GROUP BY 1 ORDER BY 2 DESC;
-- anthropic | 127 | 55      (2026-08-10)
```

> **GOTCHA, and it cost me a wrong belief during a live outage: `provider` is nested inside
> `ai_service`, NOT a direct child of `config`.** The obvious query — `v->'config'->>'provider'`
> — returns **0 rows and no error**, which reads exactly like "no agent uses this provider".
> jsonb returns SQL NULL for a missing path rather than raising. To find the right path,
> enumerate the keys instead of guessing:
> ```sql
> SELECT DISTINCT k FROM agent_definitions,
>   LATERAL jsonb_each(default_config->'workflow'->'steps') s(kk,v),
>   LATERAL jsonb_object_keys(v->'config') k;
> -- ai_service, error_step, temperature, input_fields, output_format, prompt_template, tolerate_truncation
> ```
> No `provider` in that list answers it in one command. **Never trust a zero from a jsonb path
> census until you have induced a non-zero from the same query shape.** In `LANDMINES.md`.

---

## R9 — The pinned release (`bugs_open/249` / BLD-020): what to run, and how to tell it worked

### R9a — the regression guard, AS RUN on 2026-08-11 (this is R6's substitute)

Proves `REF=` actually reaches the linker — which is both the guard the bug file asked for and
the mechanism the release pin depends on. Local only: scratch tag, never pushed, image removed.

```bash
OLD=d3c09cc746e563b6339831cfb69576eb52135c43          # any commit that is NOT HEAD
make build-agent-chassis REF=$OLD IMAGE_TAG=scratch-153-guard

docker image inspect docker.io/aqls/agent-chassis:scratch-153-guard \
  --format 'revision={{index .Config.Labels "org.opencontainers.image.revision"}}'   # must be $OLD

CID=$(docker create docker.io/aqls/agent-chassis:scratch-153-guard)
docker cp "$CID":/app/agent-chassis /tmp/guard-bin && docker rm "$CID"
grep -aq "$OLD" /tmp/guard-bin                  && echo "OLD present  (expected)"
grep -aq "$(git rev-parse HEAD)" /tmp/guard-bin || echo "HEAD absent  (expected)"   # THE control
grep -aq deadbeefcafe1234567890abcdef1234567890ab /tmp/guard-bin || echo "fake absent (expected)"
grep -aq orchestration /tmp/guard-bin           && echo "probe works (positive control)"

docker rmi docker.io/aqls/agent-chassis:scratch-153-guard; rm -f /tmp/guard-bin
```

Measured 2026-08-11: label `d3c09cc74…`, OLD present, HEAD **absent**, fake absent, probe works.

> **GOTCHA — the load-bearing control is the HEAD one, not the fabricated one.** A fabricated sha
> proves only that the grep is not matching everything. **A real but different commit** is what
> proves the stamp is *this ref's*, rather than "some sha that happens to be in there". Two of
> this lane's three false readings would have survived a fake-sha control.
>
> **GOTCHA — use a scratch `IMAGE_TAG`, never a real one.** The build tags
> `docker.io/aqls/agent-chassis:$IMAGE_TAG` locally. Running this at a live tag would leave a
> deliberately-old image sitting under a name `push-*` will happily ship, since `push-*`/`deploy-*`
> are git-blind. Remove the image afterwards.

### R9b — after the owner's next `make release`: is it ONE revision?

The pin is committed but **inert until a release runs under it**. This is the check that grades it:

```bash
for S in auth-service core-manager agent-chassis reasoning-agent web-search-adapter \
         web-scrape-adapter git-adapter image-generator-adapter thunder-adapter \
         analyser-adapter browser-runner-adapter content-creator-agent \
         remote-job-spawner kafka-scheduler; do
  docker image inspect docker.io/aqls/$S:$IMAGE_TAG \
    --format '{{index .Config.Labels "org.opencontainers.image.revision"}}'
done | sort -u                      # EXPECT exactly ONE line
```

`release` also now echoes the pinned sha up front — `Release pinned to <sha>` — which is the
cheapest possible witness and appears before any building starts.

> **This check is disconfirmable, which is why it is worth running.** Over `v1.0.1284` (built
> before the pin) it returns **three** lines. If it returns one line on a tag built before the
> pin existed, that is luck, not the fix.

### R9c — at the cluster, per service (no exec, no binary path, nothing to install)

```bash
for D in agent-chassis auth-service core-manager reasoning-agent web-search-adapter \
         web-scrape-adapter git-adapter image-generator-adapter thunder-adapter \
         analyser-adapter browser-runner-adapter content-creator-agent remote-job-spawner; do
  printf '%-26s ' "$D"
  kubectl -n ai-persona-system logs -l app=$D --tail=300 2>/dev/null | grep -m1 -o '"git_commit":"[0-9a-f]*"' || echo '<none>'
done
```

> **GOTCHA — reach for the log line BEFORE the binary probe, not after.** It is the service's own
> statement rather than your inference from its bytes, and it cannot be defeated by a missing
> tool, a wrong binary path or a permissions refusal. This lane ran the binary probe first on
> 2026-08-11, got `NO MATCH` on eight services, and nearly dismissed a genuine finding as its own
> instrument misbehaving — because the day before, that same probe *had* been the thing at fault.
> Only a control on the same pod (its own sha → MATCH, another service's sha → no match)
> separated the two. **Being burned by a false positive makes the next true positive easy to
> wave away; the control is the only thing that tells them apart.**

### R9d — the preview command changed

`make -n release` now prints the **sweep** (the pinned-sha resolution and the loop), not the
docker commands under it, because the goals are reached through a shell loop rather than a
prerequisite list. For the real preview: `make -n build-backend`, or `make -n build-<service>`,
both unchanged. This was a deliberate trade — see BLD-020's landmine for why the `+` prefix that
would restore the old output was rejected.
