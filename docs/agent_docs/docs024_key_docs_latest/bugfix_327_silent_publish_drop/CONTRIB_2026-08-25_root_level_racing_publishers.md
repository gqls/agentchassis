# CONTRIB — the census was sliced by directory, and the repo ROOT is in no slice

**Written 2026-08-25 by a session arriving at `HANDOFF_2026-08-25_open_decisions.md`, while the
closing `git mv` was in another session's working tree.** Filed as a separate file precisely so it
does not collide with that close. Nothing here reopens the bug; one live gap is fixed, one
remaining item needs the owner.

## What I was doing

Re-deriving the closing audit's state rather than trusting it — the audit is dated today, but this
tree turns over fast. Two of the three conditions re-derived exactly as claimed: the detector is in
place and in the `CHECKS` roster, and `scripts/` carries **zero** racing publishers touched in the
last 30 days.

## What the slices could not see

The scope table in the bug file (`## Scope of what remains`) partitions the class by directory:
`docs/` (86, out of scope — lane one-offs), `scripts/` **and** touched within 30d (11, the real
queue), plus dormant / duplicate / comment-only. **The repo root belongs to none of those
prefixes**, so a file sitting there is not in the queue, not out of scope, and not counted as
dormant. It is simply absent — which reads as nothing to report.

`[MEASURED 2026-08-25]` repo-wide, comments stripped and required to parse: **144** runnable
racing publishers (the lane quoted 160 on 08-24; the difference is migrations landing). Of those,
**42** were touched within 30 days. **Exactly two of the 42 are outside `docs/` — and both are at
the repo root:**

| file | last commit | in 30d window | status |
|---|---|---|---|
| `run_improvement_sweep_once.sh` | 2026-08-04 (`95aff7691`) | **yes** | **FIXED by this contrib** |
| `082_submit_domain_unified.sh` | 2026-07-30 (`95639d4f6`) | yes | **stale twin — needs the owner, see below** |
| `084_TRIGGER_diagnose_v1.sh` | 2026-07-16 (`63d51441e`) | no (dormant) | left as documented residual |

Reproduce the whole thing:

```bash
grep -rl "kcat -P" --include="*.sh" . \
  | while read f; do sed 's/#.*//' "$f" | grep -q "run -i" && bash -n "$f" 2>/dev/null && echo "$f"; done \
  | while read f; do d=$(git log -1 --since='30 days ago' --format=%ad --date=short -- "$f"); \
      [ -n "$d" ] && echo "$d  $f"; done | sort | grep -v '  docs/'
```

> ⚠ **The trap that nearly hid this from me too.** My first pass filtered with `grep '^\./scripts/'`
> and got **0**, which agreed with the audit and would have ended the check. `grep -rl … .` on this
> machine emits paths **without** the `./` prefix, so the anchor matched nothing — a false zero that
> confirmed what I already expected. The tell was that the same file reported 55 hits for
> `scripts/initial_messages` one command later. **An anchored path filter needs a control that must
> match**, exactly like every other measurement here.

## What I fixed: `run_improvement_sweep_once.sh`

The textbook form — `printf '%s\n' "$PAYLOAD" | kubectl -n kafka run -i --rm … kcat -P` — with the
`SAVE: SWEEP_CORR=` line printing unconditionally afterwards. Migrated to `kafka_publish_checked`,
with the reassuring line moved behind the receipt and a failure path that names what did not happen.

**Two things make this file worse than a generic instance, and both are now in its header comment:**

1. **`set -euo pipefail` was already at the top of this script and never helped.** The failure mode
   is a **zero** exit, which is the one thing `set -e` cannot see. A reader who checks for `set -e`
   and moves on gets the wrong answer.
2. **Nothing else would ever re-fire it.** This is the manual trigger for a scheduled task the owner
   has deliberately left `enabled=false`. A dropped publish here is not a retried publish — it is a
   sweep that silently never happens.

Because the blast radius of a *duplicate* sweep is real (full audit chain, LLM spend, live page
changes via promoted work items), the failure path distinguishes the library's code 10 (nothing
landed, retry is safe) from code 11 (receipt indeterminate — resolve with `kafka_verify_landing`
before re-firing, or you fire two sweeps).

### Induced-failure test — run with a fake `kubectl` on `PATH`, zero cluster contact

This script publishes on every invocation, so it must not be run to test it. A stub `kubectl`
lets the **real** library run and exercises the genuine classify logic:

| `kubectl` behaves | library code | script exit | `SAVE:` line |
|---|---|---|---|
| echoes `PUBLISH_OK`, exit 0 | 0 | 0 | printed |
| echoes an error, exit 1 | 10 | 10 | suppressed, "retry is safe" |
| **exit 0, no output — the bug** | 11 | 11 | suppressed, "do not blind-retry" |

**Control, and it is the part that makes the table mean anything:** the pre-fix script under that
same third condition **exits 0 and prints `SAVE: SWEEP_CORR=…`** — a correlation id to watch for a
message that was never sent. The harness is in `RUNBOOK`-able form at the foot of this file.

**Detector check, with a demand control.** `pattern-check.py` reports nothing on the staged fix.
An empty result and a broken invocation look identical, so: `python3 scripts/pattern-check.py
--commit a6a8ad92a` (the commit that *added* this file) fires `kcat-stdin-race` on this exact path.
The clean is a real clean.

## What still needs the owner: the root `082_submit_domain_unified.sh`

**This is a stale duplicate of the very script the bug is named after.**

- `scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh` — **fixed 2026-08-23**
  (`ab8ec8fbf`, this lane's Phase 1). Sources the library, asserts the receipt.
- `082_submit_domain_unified.sh` **at the repo root** — last touched 2026-07-30, **still racing**,
  tracked in git. The two share history to `95639d4f6` and diverge only by this lane's fix.

`[MEASURED 2026-08-25]` **eight documents** give the invocation as `./082_submit_domain_unified.sh`
(`webdesign_couk/RUNBOOK_webdesign_couk.md`, `brochure_component_library/MISSION_BRIEF_fundamentallyai_2026-07-20.md`,
`idea.uk/RUNBOOK_…(25).md` and `running_notes(63).md` plus their `docs014` twins,
`leopardessconsulting/RUNBOOK.md`). Whether that resolves to the root copy depends on the reader's
working directory — **from the repo root, it does.** CLAUDE.md itself points at the full
`scripts/…` path, so the documented path is the fixed one; the hazard is the shorthand.

**I have not touched it, because migrating it is the wrong fix and deleting it is not my call.**
Ordered by what closes the door:

1. **Delete the root copy** — makes the bad state unrepresentable. There is exactly one submit
   script, and `./082…` from the root then fails loudly instead of publishing into a void.
2. Migrate it too — leaves two copies to drift again, and the next divergence will be silent.
3. Leave it — it stays a live racing publisher that eight documents' shorthand can reach.

**`check_untouched_twin` structurally cannot catch this pair**: it is `.go`-only (`if not
path.endswith(".go")`), and this twin is shell. That is why a fix applied to one copy on 08-23 left
the other untouched with nothing firing.

## The transferable bit

**A census that partitions by directory prefix has a hole wherever a file has no directory.** The
arithmetic in the bug file is correct — `scripts/` really does re-derive to zero. The conclusion
drawn from it ("the residual is DATA, not an open defect") was not, because a runnable publisher
touched 21 days earlier was in neither the numerator nor the denominator. Logged in
`WRONG_CALLS.md` 2026-08-25, and the twin hazard in `LANDMINES.md`.

---

### The test harness (reusable for any migrated publisher)

```bash
mkdir -p /tmp/fakebin && cat > /tmp/fakebin/kubectl <<'EOF'
#!/bin/bash
case "$*" in *"SELECT id FROM sites"*) echo "00000000-0000-0000-0000-000000000001"; exit 0 ;;
             *"count(*)"*|*"site_work_items"*) echo 0; exit 0 ;; esac
case "${MODE:-ok}" in
  ok)     echo "PUBLISH_OK"; exit 0 ;;        # receipt seen        -> code 0
  failed) echo "Error from server"; exit 1 ;; # ran and failed      -> code 10
  silent) exit 0 ;;                           # THE BUG: sent nothing, exit 0 -> code 11
esac
EOF
chmod +x /tmp/fakebin/kubectl
for M in ok failed silent; do MODE=$M PATH=/tmp/fakebin:$PATH bash <the-script> <arg>; echo "exit $?"; done
```

⚠ Add the *pre-fix* copy as a control, or the table above proves only that the new code runs.
