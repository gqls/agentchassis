# NOTES — tmpfs / scratch exhaustion

Append-only, newest at the bottom. Technical log: what was tried, what the system actually said,
and every misstep. **The missteps are the point.**

---

## 2026-08-24 — session opens, mandate widened from "write it up" to "own the fix"

Owner: *"let this lane be responsible for fixing the persistent /tmp exhaustion problem"*, pointing
at `HANDOFF_2026-08-23_tmp_is_ram_not_disk.md`. So the question is no longer "why does it fill" —
that is diagnosed — but "why is it still filling after we fixed it".

### First measurement: `/tmp` looks fine, which is the trap

```
$ df -h /tmp
tmpfs            16G  4.9G   11G  32% /tmp
```

4.9 G / 32%, matching the handoff's §10 figure exactly. On this evidence the handoff's conclusion
holds and there is nothing to do. **I nearly stopped here.** What stopped me stopping was that the
handoff's §10 correction had already caught itself over-claiming once — a lane that has just been
wrong about a single-cause story is a lane to re-measure rather than re-read.

### The composition of `/tmp` is all old

```
2026-08-22 19:25  1.3G  /tmp/go-build1574765867
2026-08-23 13:10  1.1G  /tmp/go-build952636743
2026-08-23 19:13  453M  /tmp/trio
2026-08-22 19:29  451M  /tmp/headchk
```

Everything ≥24h idle. `trio` at 08-23 19:13 postdates the fix, i.e. a session that launched before
the settings change and kept its old environment — exactly what the handoff §9 predicted. So the
`/tmp` half is behaving as designed.

### Then the disk, and the number was not what I expected `[MEASURED]`

```
$ du -xsh /home/ant/.claude-scratch
147G    /home/ant/.claude-scratch
$ stat -c '%w' /home/ant/.claude-scratch
2026-08-03 22:12:56 +0100
```

**147 GB accumulated in 20.6 days.** Birth date is *exactly* the day of the `CLAUDE_CODE_TMPDIR`
change recorded in handoff §3. So the directory that fixed `/tmp` had been growing, unwatched,
since the hour it was created.

Composition, by shape rather than by name — a directory holding a `go.mod` whose module line is
this repo's, and **no `.git`**:

- **308 bare extracts of this repo, 130 GB** — 88% of the total
- 21 further `go.mod` hits that are *not* extracts: tiny throwaway modules a session wrote
  (`module tmplcheck`, `module authprobe`). **Matching on `go.mod` alone would have deleted
  somebody's scratch program.** The module-line test is not fussiness.
- 73 GB of the 130 was created in the **last 7 days** → **10.4 GB/day**, against **123 GB free**.
  **~12 days to a full root filesystem.**

### The mechanism, and why every copy of the recipe looks self-cleaning

Every pasted variant contains an `rm -rf`, which is why nobody noticed. It is the **setup** half —
it clears the directory the run is *about* to use, so it reclaims a tree of the same name. Each run
picks a new name, so it reclaims nothing:

```
2026-08-24 11:47  headtree     2026-08-24 12:10  headtree3
2026-08-24 12:01  headtree2    2026-08-24 12:16  headfinal
2026-08-24 12:30  ht5          2026-08-24 12:41  ht6
```

Six trees, ~2.8 GB, one session, one morning. **Disconfirmable and it came out right**: the
setup-`rm` reading predicts N directories with drifting names; a working cleanup predicts one,
overwritten. N is what is there.

### Why the 2026-08-23 documentation fix did not take

Two things, and neither is anyone's fault:

1. **The census was wrong by 8×.** Handoff §3 says "9 documents"; 8 were updated. Actual count
   `[MEASURED]` 2026-08-24: **73 documents** carry a hand-rolled `git archive HEAD | tar`, **66
   with no cleanup**. This is CLAUDE.md's own "a census goes stale by ADDITION" rule biting the
   document that was written to fix a recurrence.
2. **The fix went where only the already-informed would look.** `grep -c verify-head-builds
   CLAUDE.md` → **0**. Same for `MEMORY.md`. Eight lane RUNBOOKs were updated; a session does not
   read another lane's RUNBOOK at the moment it types a command. Of the 50 extracts created since
   the script shipped, **none** use its layout.

`scripts/verify-head-builds.sh` itself is fine — read it, it refuses a tmpfs target by filesystem
type and traps EXIT. It is unadopted, not broken. **A silent mechanism is usually undriven.**

### Misstep: my own self-test reported two guards broken, and the guards were fine

Wrote `scratch-janitor.sh --self-test` to plant each hazard and assert the guard fires. First run:

```
  PASS  control: an idle /tmp dir reaches the delete list
  FAIL  guard: a candidate holding a .git was NOT refused
  FAIL  guard: accepted an idle gate under 2h
```

Both "failures" were the harness. The tests were shaped `"$0" … 2>&1 >/dev/null | grep -q '…'`, and
the script is `set -o pipefail`: a *successful* refusal exits 2, so the pipeline returns 2 even
when grep matched. **A correct refusal was being reported as a broken guard.** Fixed by capturing
into a variable and grepping that.

Two things worth keeping from it. First, it failed in the safe direction — a false alarm about a
guard, not a false all-clear — but the *general* form of this bug is not safe: the same
`cmd | grep` shape reports **success** whenever the last stage passes and pipefail is off, which is
the default. Second, I only knew the guards were fine because I ran the refusal by hand and read
it; **a red test is a claim like any other and it needs the same check as a green one.**

### Prior art I should have found sooner

`systemd-tmpfiles-clean.timer` is **active and runs daily**, with `q /tmp 1777 root root 10d`
(`/usr/lib/tmpfiles.d/tmp.conf:11`). So the handoff's §4.3 framing — "nothing reaps idle scratch" —
is **false as stated**. Something does; its age gate is 10 days while `/tmp` went empty-to-full in
about four, so it can never fire before the failure. The correction matters because the two
statements point at different fixes: "nothing reaps" says build a janitor, "nothing reaps in time"
says shorten the existing gate (one root-owned drop-in in `/etc/tmpfiles.d/`). Both are worth
having — the drop-in cannot cover the disk scratch, which is where 88% of the volume now is.

### What shipped this session

- `scripts/scratch-janitor.sh` (`abf9b7485`) — reaps both roots. Disk side by **shape**, because a
  session scratchpad holds the disposable extract and the session's real work in the same
  directory and an age gate cannot tell them apart. `/tmp` side by age with the system directories
  excluded by name. Dry run by default; refuses a gate under 2h; `--self-test` proves the guards.
- Dry run at the shipped gates: **261 directories, 108 GB**. Not applied — that is other sessions'
  data on shared ground and the last cleanup was done on the owner's instruction for that reason.
