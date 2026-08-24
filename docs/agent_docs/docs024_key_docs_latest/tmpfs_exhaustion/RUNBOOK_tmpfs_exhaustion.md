# RUNBOOK — tmpfs / scratch exhaustion

Every command here was got right the hard way. The gotcha is attached to the command, not filed
separately. **When one changes, change it HERE.**

---

## The two-line health check

```bash
df -h /tmp /                     # tmpfs (RAM) and root (disk) — BOTH, always
du -xsh "${CLAUDE_CODE_TMPDIR:-$HOME/.claude-scratch}"
```

> **⚠ Looking at `/tmp` alone is how this lane got the wrong answer on 2026-08-23.** `/tmp` was
> fine (+0.5 GB/day) while the disk scratch was taking 10.4 GB/day. The producer had moved, not
> stopped. **Check both or check neither.**

## The reaper — `scripts/scratch-report.py` (OPP-005)

```bash
scripts/scratch-report.py --days 2              # DRY RUN — what would go, and how much
scripts/scratch-report.py --days 2 --reap       # delete
scripts/scratch-report.py --self-test           # prove the guards still fire
scripts/scratch-report.py --days 7              # more cautious gate
```

Dry run by default. It deletes **only** marker-verified repo extractions and `go-build*` linker
dirs; never a loose file, never a session directory, never anything it cannot positively identify.

> **⚠ THIS TOOL ALREADY EXISTED AND HAD NEVER BEEN RUN.** Deployed 2026-08-03, correct, and
> 97.1 GB was sitting reapable at its own default gate on 2026-08-24. **Check the concept register
> before building a reaper** — this lane built a duplicate first (`scratch-janitor.sh`, deleted the
> same day). *A silent mechanism is usually undriven, not missing.*

> **⚠ Its `/tmp` arm was INERT until 2026-08-24, while looking covered.** `ROOTS` listed `/tmp` and
> the register entry claimed both roots were read. Every candidate came from `scratch_dirs()`,
> which requires a `<root>/claude-*/<proj>/<uuid>` layout — and `/tmp` has never had a `claude-*`
> directory. **The tell was a missing section header**: the report printed no `=== /tmp ===` at
> all. Fixed by `loose_reapables()`. If you change the roots, check the header prints.

> **⚠ Run `--self-test` after editing it, and read the CONTROL line first.** The opening PASS
> asserts a real extraction *reaches* the reap list; the age-gate case then holds back **the same
> directory**. Without that pairing every refusal is vacuous — a guard that never sees a candidate
> "passes" while protecting nothing.

> **⚠ A dry run takes ~2 min** — it walks ~750 session directories computing sizes. Not a hang.

## Measuring the actual producer (the numbers in PLAN §1)

Find bare `git archive` extracts of **this repo** by shape — module line present, no `.git`:

```bash
SCRATCH="${CLAUDE_CODE_TMPDIR:-$HOME/.claude-scratch}"
MOD="$(grep -m1 '^module ' go.mod)"
OUT="$SCRATCH/gotmp/extracts.nul"            # NB: on DISK. A list is small; a 450MB tree is not
find "$SCRATCH" -maxdepth 6 -name go.mod -type f 2>/dev/null | while read -r g; do
    d="$(dirname "$g")"
    [ -e "$d/.git" ] && continue
    [ "$(grep -m1 '^module ' "$g")" = "$MOD" ] && printf '%s\0' "$d"
done > "$OUT"
tr -cd '\0' < "$OUT" | wc -c                        # count
du -xc --files0-from="$OUT" -sh | tail -1           # total
```

> **⚠ Match the MODULE LINE, not just the presence of `go.mod`.** A plain `-name go.mod` search
> returns 329 hits here, of which **21 are tiny throwaway test modules** a session wrote
> (`module tmplcheck`, `module authprobe`, …). Only **308** are 450 MB extracts of this repo.
> Deleting by "has a go.mod" would take somebody's scratch program.

> **⚠ `du -xsh` on the parent tells you nothing about the cause.** It said 145 GB for one directory.
> The composition is the finding, and you only get it by shape.

Creation rate over the last 7 days (this is the number that matters, not the running total —
the total is dominated by history and understates a rising rate):

```bash
tr '\0' '\n' < "$OUT" | while read -r d; do find "$d" -maxdepth 0 -mtime -7 -print0; done > "$SCRATCH/gotmp/last7.nul"
du -xc --files0-from="$SCRATCH/gotmp/last7.nul" -sh | tail -1
```

Per-day histogram, which is what shows whether a fix took:

```bash
tr '\0' '\n' < "$OUT" | xargs -d'\n' stat -c '%y' | cut -c1-10 | sort | uniq -c
```

## Has a fix actually reached the fleet?

```bash
grep -c 'verify-head-builds' CLAUDE.md          # the file EVERY session loads. 0 = unadopted
grep -rl 'git archive HEAD *| *tar' docs/ *.md | wc -l   # 73 as of 2026-08-24
```

> **⚠ Count the documents before believing a "we fixed the N copies" claim.** The 2026-08-23 work
> fixed 8 documents against a stated census of 9. The real census was **73**. A census does not go
> wrong, it goes **stale by addition** — so date it, and re-run it rather than quoting it.

## Is anything already reaping `/tmp`?

```bash
systemctl list-timers --all | grep tmpfiles
grep '^[a-z]' /usr/lib/tmpfiles.d/tmp.conf      # q /tmp 1777 root root 10d
```

> **⚠ There IS a janitor and it has never helped.** `systemd-tmpfiles-clean` runs daily with a
> **10-day** age gate, while `/tmp` historically went from empty to 100% in about **four**. It
> cannot fire before the failure. "Nothing reaps `/tmp`" is false; "nothing reaps it in time" is
> the true statement, and only one of those suggests the right fix (shorten the gate, root
> required, `/etc/tmpfiles.d/` drop-in).

## Which sessions are producing, right now

```bash
SCRATCH="${CLAUDE_CODE_TMPDIR:-$HOME/.claude-scratch}"
find "$SCRATCH" -maxdepth 6 -name go.mod -mmin -1440 -printf '%TY-%Tm-%Td %TH:%TM  %h\n' | sort
```

Useful when the rate jumps: it names the session directory, and the directory *names* tell you
whether the session used `verify-head-builds.sh` (layout `head-verify/<pid>/tree`) or hand-rolled
one (anything else).

## Environment — is the redirection reaching THIS session?

```bash
echo "TMPDIR=$TMPDIR  GOTMPDIR=$GOTMPDIR  CLAUDE_CODE_TMPDIR=$CLAUDE_CODE_TMPDIR"
go env GOTMPDIR
```

> **⚠ NEW SESSIONS ONLY.** A running session keeps the environment it launched with. A session
> started before the 2026-08-23 settings change is still writing to `/tmp` and always will be —
> which is why `/tmp` kept gaining after the fix, and why that gain is **not** evidence the fix
> failed. Check the session, not the machine.

> **⚠ `go env GOTMPDIR` empty means Go falls back to `TMPDIR`**, i.e. `/tmp`. `CLAUDE_CODE_TMPDIR`
> means nothing to the Go toolchain. This is what makes a full tmpfs present as
> `link: mapping output file failed: no space left on device`, which reads like a compiler fault.
