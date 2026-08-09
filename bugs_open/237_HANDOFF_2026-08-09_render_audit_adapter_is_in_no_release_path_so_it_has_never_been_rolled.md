# 237 — `render-audit-adapter` is in no release path, so it has never been rolled

**Status: OPEN. One-line fix known (below); not applied — releases are
whole-fleet and the owner runs them.**

Found 2026-08-09 while verifying `bugs_open/233`'s fix at the pod: every
storage-touching service had picked up the fix on v1.0.1274 except this one,
which is still serving the pre-fix binary and still has a B2 credential in its
live log buffer.

## Mechanism

`render-audit-adapter` deliberately has no binary of its own — it runs the
**browser-runner** image with a different topic and consumer group
(`deployments/kustomize/services/render-audit-adapter/overlays/production/uk_001/`).
Its overlay pins the tag and its own comment says:

```yaml
# Pin the image tag for this cluster. Keep exactly one tag-pin line — the
# release tooling seds it.
```

**The release tooling does not sed it.** There are two tag-update mechanisms in
the makefile and this service is in neither:

1. Per-service hardcoded lines (`makefile:952-1031+`), one
   `sed -i.bak 's/newTag:.*/…/' …/services/<name>/overlays/…` per service —
   `agent-chassis`, `reasoning-agent`, `web-search-adapter`, `web-scrape-adapter`,
   `git-adapter`, `image-generator-adapter`, `thunder-adapter`, `analyser-adapter`,
   `browser-runner-adapter`, `content-creator-agent`, `remote-job-spawner`,
   `kafka-scheduler`, … **no `render-audit-adapter` line.**
2. `update-kustomization-images` (`makefile:918`), a `for agent in …` loop over
   another **hardcoded list of 11** — also no `render-audit-adapter`.

`grep -n "render-audit" makefile` → **zero hits anywhere in the file.** It has
no build target either; it does not need one (it reuses the browser-runner
image), but that is exactly why nothing ever bumps its tag.

Consequence: the overlay has been untouched since the pod was created
(`git log … -- <overlay>` → single commit `0143a693e`, "feat(render-audit):
give the audit its own pod, own logs, own failure state"), so it is frozen at
**v1.0.1194 while the fleet is on v1.0.1274 — 80 tags behind.**

**This is not only about credentials.** Every browser-runner fix since 1194 —
every render, screenshot and check-vocabulary change — has been reaching the
browser-runner pod and *not* the render-audit pod, which runs the same binary
against a different topic. Any lane that verified a browser-runner fix at the
browser-runner pod and assumed the audit path had it too was wrong.

## Census (with a control, because the first attempt was inert)

All 30 production overlays, tag extracted and compared to the fleet tag:

| service | pinned tag | note |
|---|---|---|
| `render-audit-adapter` | **v1.0.1194** | 80 behind; runs browser-runner binary |
| `component-render-check` | v1.0.1258 | own image; links storage, never calls it |
| `shared-output-fields-check` | v1.0.1265 | own image |
| `github-actions-runner` | **v1.0.948** | huge drift; own image, own concern |
| `github-actions-runner-vmsites` | v1.0.1126 | shares the runner image |
| `ollama-adapter`, `ollama-eval` | `latest` | third-party image, unpinned |
| everything else (22) | v1.0.1274 | current |

> **MISSTEP, recorded because the check was inert, not merely wrong.** My first
> census used `grep -A1 "images:" | grep newTag` and printed `<none>` for **all
> 30 services** — including `render-audit-adapter`, which I had just read as
> `newTag: v1.0.1194`. The `newTag` line is two lines below `images:` (the
> `name:` line sits between), so `-A1` could never reach it: the census would
> have reported "no tags anywhere" whether or not any tag existed. It was caught
> only because I happened to hold a known-positive. **A census needs a row whose
> value you already know, or it cannot come out false.**

## Fix candidates, ordered by what closes the door

1. **Make the tag-update mechanism enumerate the filesystem, not a hand-list.**
   `update-kustomization-images` already globs nothing — replace both hardcoded
   lists with a loop over `deployments/kustomize/services/*/overlays/$(OVERLAY_PATH)/kustomization.yaml`,
   skipping overlays whose image is third-party (`ollama/ollama`). Then a new
   service is covered by existing, not by remembering. This is the only
   candidate that makes the bad state unrepresentable; the others rely on
   someone remembering.
2. **Add the missing `sed` line** for `render-audit-adapter` (and decide
   `github-actions-runner*` separately — 948 is drift of a different order and
   may be deliberate). One line, immediate, but leaves the class open: the next
   image-sharing service repeats this exactly.
3. **Fail loudly instead**: a check that greps every overlay's pinned tag and
   errors when one is more than N tags behind the fleet. Detection, not
   prevention — but it would have caught this months ago, and the estate already
   runs several such CronJob checks.

Do **not** "fix" this by `kubectl delete pod` on render-audit — the overlay pins
v1.0.1194, so it returns on the same stale image (and, per `bugs_open/233`,
writes a fresh plaintext credential line when it restarts).

## Why no 090 run (declared per the OWNER RULING 2026-07-31)

The claim is structural — "a service is in no release path" is a statement
about shared tooling — so the ruling applies. Substituting first-hand
verification, stated plainly rather than omitted:

- `grep -n "render-audit" makefile` → 0 hits (absence proven for that spelling;
  the service dir is spelled `render-audit-adapter` everywhere else, checked).
- Both tag-update mechanisms read directly, top to bottom; both are literal
  enumerations, quoted above with line numbers.
- `git log -- <overlay>` → one commit ever, at creation.
- The running pod confirms the consequence, not merely the config:
  `browser-runner-adapter:v1.0.1194` while the fleet is 1274.
- The census reproduces a known-positive row (see the misstep note).

That is direct reading of the deciding code plus the live artefact, which is
what the loop would have done here; there is no non-obvious cause to hunt.

## Verify a fix

```
grep -c "render-audit" makefile                       # expect >0 after fix 1 or 2
kubectl get pod -n ai-persona-system -l app=render-audit-adapter \
  -o jsonpath='{.items[*].spec.containers[*].image}'   # expect >= v1.0.1274
```
Then re-run `bugs_open/233`'s pod-grep against it: `access_key_present` = 1,
`B2_APPLICATION_KEY from env` = 0. That image has no `strings` — use
`grep -ac <str> /app/browser-runner-adapter`, with `NewS3Client` as the
positive control.

## Relations

- `bugs_open/233` — the credential leak this was found by; carries the
  **rotation-ordering constraint** that makes this bug time-sensitive: roll
  render-audit BEFORE rotating B2 keys, or the new key is logged in plaintext.
