# NOTES — `bugfix_136_config_key_aliases` (append-only, newest at the bottom)

## 2026-08-08 — session 1: picking the bug

Triage of `bugs_open/` against 32 live `.jsonl` transcripts. The useful discriminator was
**not** `who-owns.py`, which returns "OWNED or recently active" for almost everything on this
estate because almost every workstream directory is active. Two things worked:

1. `git log -1 --format=%cd -- bugs_open/<file>` per bug, sorted — a bug whose file has not
   moved in ≥9 days is a real signal.
2. Grepping the **tail** of each recently-touched transcript (`tail -c 400000`) for
   `bugs_open/NNN` and taking the top few by count. A session's *current* focus is legible
   that way; a whole-file grep is not, because every session that ran `ls bugs_open` matches
   everything.

Rejected after checking: `093` (its own last update says it is no longer a code task — it is
blocked on `083`, and `083` is hot in two sessions) · `211` (`who-owns` OWNED, the 122 lane
committed to it today, and one session had 62 hits on it) · `085`, `181`, `185`, `189`, `203`
(fixed and live; only site-level verification owed) · `114`, `126`, `146` (owning lanes active
within 8 days) · `040` (lane dormant, but the bug is infrastructure and the metrics predate
any claim).

**A trap worth naming**: `b5a58a2b` showed 20 hits on `136`'s vocabulary, which read as a
second session on this bug. Every hit was the auto-memory file
`bugfix-136-domain-pipeline-rename.md` being loaded as context, not work. **Grep the hits,
do not count them** — a memory file mentioning a bug looks exactly like a session working it.

## 2026-08-08 — the bug was still valid, and worse than filed

Re-measured before planning. The three named instances all still present in the live audit
(4 UNKNOWN KEYS, three of them this bug's). Two figures had moved since 2026-07-28:
`target_domain` is carried by **1** step, not 3 (migration 286 removed two under RFC 006), and
`item_domain` by 9, not 7.

Then the finding that changed the priority. §2a's `[MEASURED] Nothing is mislabelled today`
is **false as of 2026-08-04**: `content_duplication` and `page_canonical_collision` — both of
which propagate `dctx.Pipeline` — have joined `completeness-discovery-agent`, and four rows
are filed under `design` on an agent whose config says `content`. The bug file predicted this
exact trigger in the sentence immediately after the measurement. **A figure that depends on
which checks are registered where has a shelf life of days here**, and §2a did not mark it.

## 2026-08-08 — the framework finding, which is the whole shape of the fix

Reading `action_inputs.go` to see whether `ActionInputSpec.Deprecated` could carry these
renames: **it cannot, and would be worse than nothing.** Strategy 3 does
`ExtractNestedField(collectedData, config[oldKey])` — the old key's *value* is a dot-path.
That is a reference alias. A literal setting placed there resolves nothing, takes the
default, and — because `UnknownConfigKeys` recognises `Deprecated` keys on purpose — silences
the detector as well.

This answers the question the bug file asked and could not answer: *"They wrote it on one
action and not on the other two."* There was no honest declaration available, so two of three
authors correctly did nothing. The per-action shim was never the fix; it was the symptom.

## 2026-08-08 — missteps

1. **`grep 'config\["'` to enumerate an action's config reads.** Missed `priority`, which
   arrives via `GetIntField`. Caught by the planner before it reached a doc. Logged in
   `WRONG_CALLS.md`; the check is in the LANDMINE that shipped with the fix. I made this
   bug's own mistake — a key read in a way the obvious search cannot see — while writing its
   fix.
2. **A mutation that did not compile is not a mutation proof.** My first attempt at killing
   `TestResolveConfigSettingPrecedence` (`if canonical == key` → `if false`) left `old`
   unused, so the package failed to build and the test "failed" for the wrong reason. A build
   failure proves nothing about a test: it proves the compiler works. Redone as
   `canonical == key && false`, which compiles and makes the rule inert — and the test then
   failed on exactly the row that exists for the rule.
3. **`go build` at HEAD vs `go test` in the tree is not a comparison.** Chasing whether a
   `thunder` failure was mine, I built the package from `git archive HEAD` and it passed —
   which looked like proof I had broken it. `go build` does not run vet, and the failure was
   a vet diagnostic. `go vet` at HEAD reproduces it identically. Match the instrument before
   attributing the difference.

## 2026-08-08 — what shipped, and what was deliberately left

Committed `3f93456fd`, 13 files, `Council-Submitted: 433de2c0-682f-4d8d-8c48-28637309f1ba`.
Five guards, each killed by a deliberate mutation and restored:

| mutation | test that died |
|---|---|
| alias loop made inert (`&& false`) | `TestResolveConfigSettingPrecedence` — the alias-only row |
| precedence swapped (alias beats canonical) | same test — the both-set row |
| `DeprecatedConfigKeysInUse` gated on `checksConfig()` | `TestDeprecatedConfigKeysInUseIndependentOfOptIn` |
| alias loop removed from `UnknownConfigKeys` | `TestUnknownConfigKeysRecognisesDeprecatedConfigKeys` |
| triage action body reverted to the direct config read, **spec left declaring the alias** | `TestTriageDetectedItemsHonoursTargetDomain` (sqlmock, asserts the value reaches the UPDATE) |

That last one is the one worth keeping: it is the only test in the set that catches a
declaration which has stopped describing the code, which is this bug's own shape.

**Live verification that needed no roll**, run against production `agent_definitions`:
UNKNOWN KEYS **4 → 1**. The survivor, `plan_sections: domain`, is left on purpose —
`page-build-handler` is hot with several sessions today and the UNKNOWN line is the honest
record until its owner deletes the key.

Not done, all recorded on the bug file: the four mislabelled rows are not repaired (two sit
in `detected`, a queue with no consumer — `bugs_open/083`); `summary_template` is still
biting and is **not** an alias case (aliasing it to `summary` would ship a raw
`{{.input_data.topic}}` to a human reviewer); `create_work_item`'s full opt-in; every
definition edit.
