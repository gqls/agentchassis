# NOTES — bugfix 233, B2 keys and the DB password logged in plaintext

Append-only, newest at the bottom. Missteps are the point.

---

## (a) 2026-09-03 — picked up as an UNOWNED bug, and the first thing found was that its own banner was stale

Taken while looking for unowned work: no lane directory, no live session naming it, untouched 23 d.
The fix itself had shipped 2026-08-09 and been verified at the pod; the file stayed open for one
named reason.

**That reason had evaporated 15 days earlier and the file never said so.** The banner read
*"STILL LEAKING: `render-audit-adapter`, which runs `browser-runner-adapter:v1.0.1194` — 80 tags
behind, pinned since the pod was created"*, and rooted the pin in `bugs_open/237`. **`237` closed
2026-08-19** — the service was unfrozen 08-10, the class fix (BLD-022) shipped 08-17.

⚠ **The cost of a stale banner is not confusion, it is INACTION.** This one carried an explicit
instruction to the owner — *roll that service BEFORE rotating the B2 keys* — so for fifteen days the
correct action looked premature and a plaintext-emitted credential stayed live. Logged in
`WRONG_CALLS.md` as the second recorded instance of `a-closed-blocker-keeps-being-obeyed` (the first
cost 20 days).

**The check belongs to whoever CLOSES a bug, not whoever waits on it:**
`grep -rln "bugs_open/<number you just closed>" bugs_open/ docs/`. Run on `237` it returned **3 open
bugs**, of which only **one** (233) was genuinely gated — `249` cites it as *"adjacent"*, `153`
carries a CONTRIB *from* that lane. **Read each hit for an INSTRUCTION**, not a mention; treating
every citation as a dependency is how a useful check gets abandoned as noisy.

## (b) 2026-09-03 — verifying the leak is gone, at the binary, three ways

The file's own recipe used `strings`, which the estate has since retired (absent from debian-slim
images, and behind `2>/dev/null` its failure is indistinguishable from "not found"). Used `grep -a`
on `/proc/1/exe` instead:

| probe | reads as |
|---|---|
| `NewS3Client` — unrelated to the fix | **control**: the probe mechanism works |
| `access_key_present` — added by the fix | the fix **shipped** |
| `B2_APPLICATION_KEY from env` — removed by the fix | the leak is **gone** |

PRESENT / PRESENT / ABSENT on `render-audit-adapter`. ⚠ **Without the control this is worthless** —
a broken `grep -aq` returns ABSENT for all three and reads as "still leaking".

**Fleet census with its own control** `[MEASURED 2026-09-03]`: **35** images whose version parsed,
**min tag = max tag = `v1.0.1359`**, **0** older than the `v1.0.1274` fix. The **19** unparsed images
are third-party (`postgres:16-alpine`, `ollama/ollama`, `pgbouncer`, `wireguard`, `bitnami/kubectl`)
and are recorded as *unjudged* rather than folded into the zero. Without min/max, "0 older than
v1.0.1274" would read identically if the regex had matched nothing at all.

## (c) 2026-09-03 — a zero I did NOT get to use

Checked whether any retained log buffer still holds the credential. The three pre-fix
`component-render-check` pods (25–27 d old) return **0** leak lines — **out of 0 total lines.**

⚠ **A zero from an empty log is vacuous**, and this file's own earlier check had been careful about
exactly that (*"a non-empty log, so the zero is meaningful"*). Their logs have expired; the question
**cannot now be answered** for them, which is different from answering it clean. The
`render-audit-adapter` pod that did hold one has been replaced, so that buffer is gone with it.

## (d) 2026-09-04 — the residual the fix explicitly did not claim, swept CLEAN

The original file bounded its class by grep and named one gap honestly: *"Not swept: secrets reaching
logs via struct dumps (`zap.Any` of a config object) … NOT claimed here."*

**Two questions, each with a control:**

1. **Do live structs hold a secret value, and are they logged?** Fields named
   `*Password|Secret|Token|Credential|ApiKey|AccessKey` (excluding `*EnvVar`) exist in exactly two
   places in built code — `internal/tools-api/config/config.go:17` (`AnthropicAPIKey`) and the
   `internal/auth-service` request/response DTOs. **Neither is ever logged: 0 hits each.**
   ⚠ CONTROL: the same expression returns **15** repo-wide and finds the known
   `zap.Any("config", params.StepConfig)` dumps at `spawn_actions.go:624/630/638`. Real absence, not
   a broken pattern. Matches under `docs/` are documentation bundles, not built code — excluded
   explicitly rather than silently.
2. **Can a dumped step config carry a secret value?** Those 15 sites dump
   `agent_definitions.default_config`. Text regex over every active agent: **81 agents** match a
   secret-shaped key with a ≥12-char string value (control: **67** carry a `prompt_template`).
   ⚠ **All 81 are FALSE POSITIVES, and the count alone would have read as an emergency.** Extracting
   the matching **key names** — never the values, per the owner ruling — gives exhaustively
   `api_key_env_var` ×160, `secret_key_env_var` ×2, `access_key_env_var` ×2. **Every one holds an
   environment variable NAME**, which is the very convention this bug's fix converged on.

**Second sighting of that false-positive class in this one bug** (the first was
`thunder_ssh_exec_dispatch.go:315`, where `token` was a command-template token NAME).
**A credential-shaped KEY NAME is not a credential**, and the discriminator is to extract names
rather than count matches.

## (e) 2026-09-04 — re-verified on the new build, and what is actually left

`v1.0.1360` fleet-wide (**36** images, min = max = 1360, 0 pre-fix). Binary probe on **both**
`render-audit-adapter` and `agent-chassis`: control PRESENT / `access_key_present` PRESENT /
`B2_APPLICATION_KEY from env` ABSENT.

**Nothing technical remains. One owner decision does.** `git log -S` dates the leak to `9260b86ed`,
**2025-10-28** — so the B2 pair and `CLIENTS_DB_PASSWORD` have been landing in pod logs for **over
ten months** and must be presumed disclosed to anyone with `kubectl logs` access in that window.
**[UNVERIFIED] whether either has been rotated, and this session cannot check it**: the secret's
`creationTimestamp` survives in-place updates, so it cannot distinguish "never rotated" from "rotated
in place", and reading a key value is forbidden (owner ruling 2026-08-23).

⚠ **The code stopped the emission. It did not un-emit ten months of it.** That distinction is why
this file stays open on a fix that is demonstrably live.
