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

### Misstep, the larger one: I built a reaper that already existed

Went to write the concept-register entry for `scratch-janitor.sh` and read the neighbouring
entries first, as you do. **OPP-005 — "Session scratchpads: versioned by a snapshot hook, reaped by
content class", deployed 2026-08-03** — ships `scripts/scratch-report.py --reap`: both roots,
marker-verified identification, dry run by default, refuses anything it cannot positively identify.
It is the tool I spent the morning writing, three weeks older, and slightly better designed.

```
$ ./scripts/scratch-report.py --days 2
250 marker-verified extraction dir(s) older than 2.0d = 97.1G
```

**97.1 GB reapable at its own default gate, and no evidence it had ever been run.** So the whole
framing in PLAN §4.1 was wrong. The estate did not lack a reaper; it lacked a *schedule*. The
remaining work is one crontab line, not 250 lines of bash.

CLAUDE.md says to consult the register *"before concluding something does not exist"*. I consulted
it before **announcing** — which is a different and much later moment, and by then the code was
written, tested and committed. The check costs one `grep` of
`docs026_concept_register/register/*.md` and I did not spend it because the diagnosis felt novel.
**Novel symptom, existing machinery** is exactly the case the register is for.

Deleted `scratch-janitor.sh` (`0097d25de`) rather than keeping two reapers, which is the drift
class this estate files bugs about weekly.

### What the fold-in found, and it is the sharper half

`scratch-report.py`'s `ROOTS` lists `~/.claude-scratch` **and** `/tmp`, and OPP-005's own landmine
says: *"Both tools read both roots for this reason; a check that inspects only one will be
confidently wrong for as long as the old sessions live."*

`[MEASURED 2026-08-24]` **it read one.** Every candidate came from `scratch_dirs(root)`, which
yields `<root>/claude-*/<proj>/<uuid>/` — and `/tmp` has never had a `claude-*` directory in it
(handoff §8 recorded exactly that fact on 2026-08-23, as *reassurance* that no scratchpads were in
`/tmp`; the same fact is why the tool could not see `/tmp` at all). Control:

```
$ ls -d /tmp/claude-*
ls: cannot access '/tmp/claude-*': No such file or directory
$ ls -d /home/ant/.claude-scratch/claude-*
/home/ant/.claude-scratch/claude-1000
/home/ant/.claude-scratch/claude-7dd0-cwd
```

**The tell was an absent section header.** The report prints `=== <root> ===` per root; it printed
one. Nothing errored, no figure was wrong, no zero appeared — a heading simply was not there, and
an early `continue` on "no session rows" skipped the root silently. **A missing row is the hardest
refutation to notice because there is nothing to read.** Worth generalising: check for the section
you *expected*, not only the numbers in the sections you got.

Fixed with `loose_reapables()` — a root's top level plus exactly one level below it (holding
directories like `gotmp/` and `adhoc/` are not themselves scratch, but the linker dirs sit inside
them; deeper would start walking real work). Two shapes only, both regenerable by construction:
marker-verified extraction, and `go-build[0-9]+`. `/tmp` now reports **7 dirs, 2.7 GB**.

Also ported the self-test onto it — `scratch-report.py` had none. Six guards, all firing, with the
age-gate case deliberately re-using **the same directory** as the control so that "held back by the
gate" cannot be confused with "never identified in the first place".

### Where the numbers stand at the end of the session

| | |
|---|---|
| reapable, existing tool, 2-day gate | **250 dirs / 97.1 GB** |
| reapable, after the `/tmp` fold-in, 1-day gate | **272 dirs / 106.0 GB** |
| irreplaceable session work (never reaped, and correctly so) | 14.8 GB |
| free on `/` | 120–123 GB |

Nothing deleted. That is the owner's call.

## 2026-08-24 (later) — the reap ran, and what it did and did not move

Owner chose the 2-day gate and a daily dry-run cron. Both done.

**The reap, `scratch-report.py --days 2 --reap`:**

```
254 marker-verified extraction dir(s) older than 2.0d = 98.7G
...
freed 98.7G
```

| | before | after |
|---|---|---|
| `/` | 766 G used, **87%**, 124 G free | 659 G used, **75%**, **231 G free** |
| `~/.claude-scratch` | **147 G** | **38 G** |
| `/tmp` | 4.9 G / 32% | 4.9 G / 32% — unchanged, correctly |
| irreplaceable session work | 14.8 G | **14.8 G — untouched** |
| Swap free | 380 KiB | **380 KiB — unchanged** |

Four things worth reading in that table rather than skimming.

1. **`/tmp` did not move, and that is the right answer.** Its contents are ~1.7 days idle, under a
   2-day gate. A tool that had "helpfully" taken them would have been the wrong tool.
2. **The 14.8 GB of real session work was not touched.** That is the design holding: shape, not
   age. It is also the number I would want re-checked by anyone who changes the identification
   test.
3. **Swap did not budge**, and this is the disconfirmable half. The 2026-08-23 episode freed 11 GB
   of *tmpfs* (RAM-backed) and returned 6.9 GB of swap. Today freed 98.7 GB of *disk* files and
   returned nothing — exactly what handoff §10's correction predicts, and the opposite of what the
   original single-cause story would have predicted. **Memory pressure on this box is the ~50
   concurrent sessions (25 GiB), not scratch.** Nine times the bytes, none of the effect.
4. **System directories all survived**: 5 of 5 named ones plus all 11 `systemd-private-*`. The
   exclusion list earned its place for the second time.

**The estate's own build check, run afterwards with no `TMPDIR` override:**
`scripts/verify-head-builds.sh ./cmd/config-key-audit/` → `OK — HEAD 85c258c95 builds.`

**The cron entry**, foreground-tested in a stripped cron environment (`env -i`, cron's `PATH`,
`/bin/sh`) **before** arming it — exit 0, log block written. Crontab backed up first; both
pre-existing entries verified present afterwards (3 of 3), because `crontab -` replaces the whole
file and a careless read-modify-write takes another lane's job with it.

### A signal I created and then had to remove

I wrote up "an absent `=== /tmp ===` header is the tell for a broken root arm" as the finding of
the morning — into LANDMINES, WRONG_CALLS, the register and this file. Then the reap ran, `/tmp`
legitimately dropped below the gate, and the header vanished for an entirely innocent reason. **The
same silence now meant both "clean" and "broken".** A tell that fires on two opposite conditions is
not a tell.

Fixed properly rather than re-documented: every root now prints a header unconditionally, with an
explicit line when it has nothing. The generalisable version is the one I had already written down
and then walked straight into — **never leave a missing row as a signal**; if absence carries
meaning, emit something.

## 2026-08-25 — asked "is /tmp full again?", and the answer moved the goalposts twice

**No.** `/tmp` = **4.1 G / 27%**, *lower* than yesterday's 4.9 G / 32%. Scratch = **25 G**, down
from 33 G. Both halves of this lane's fix are holding.

### But the disk went 74% → 85% overnight, and neither cause is scratch

| | 2026-08-24 13:04 | 2026-08-25 12:48 |
|---|---|---|
| `/` used | 656 G, **74%** | 753 G, **85%** |
| `/` free | 234 G | **137 G** |
| `~/.claude-scratch` | 33 G | **25 G — went DOWN** |

97 GB gone in 25 h while the thing I had been watching *shrank*. Two consumers, found by following
the arithmetic rather than the suspicion:

**1. Go's build cache — `~/.cache/go-build`, 117 GB, `[MEASURED]` +50.5 GB / 116,321 files in 25 h.**
Go's own trimmer is working (`trim.txt` stamped ~2 h before I looked) but only removes entries
unused for **5+ days**, against ~50 GB/day of churn from ~50 sessions compiling. Same shape as the
`systemd-tmpfiles` finding: a reaper that exists, runs, and cannot fire fast enough to matter.
Note the 2026-08-23 handoff called this cache out by name as *"the Go cache that actually matters …
untouched by any of this"* — correct at the time, and it is now the second-largest thing on the box.

**2. Docker's build cache — 539 GB, of which 437.9 GB is reclaimable. Plus 103.8 GB of images
(1,034 images, exactly ONE in use).**

### The methodological finding, and it is the third repeat of the same error this lane has made

**`du` could not see it, and said so quietly.** `du -xsb /` totals **226 GB** against `df`'s
**754 GB used** — a **528 GB gap**. Everything in this lane's accounting, yesterday's included,
was built on `du` as user `ant`, and `/var/lib/docker` is root-owned: `du -xsh /var/lib/docker`
returns **`4.0K`**, which is not an error, not a zero, and not a refusal — it is a plausible small
number for an empty directory. `2>/dev/null` then hides the one line that would have said
otherwise.

The check that finds it is free and I should have run it on day one:

```bash
du -xsb / 2>/dev/null | awk '{print $1}'   # vs
df -B1 --output=used /  | tail -1          # a large gap = you cannot see the disk
```

**Reconcile `du` against `df` before believing any disk census.** Deleted-but-open files were the
obvious suspect and were ruled out first (`lsof | grep deleted` → 0.9 GB, not 528).

### And the nightly job I armed yesterday never fired

The log held exactly one block — my own foreground test. `[MEASURED]` the machine was **suspended
2026-08-24 23:15:13 → 2026-08-25 09:54:17**; the 06:41 crontab entry fell inside it, and plain
`cron` does not replay a window it slept through (`anacron` covers only
`/etc/cron.{daily,weekly,monthly}`, not a user crontab line). **It would have missed every night
with that sleep pattern, silently.**

Replaced with a systemd **user timer**, `Persistent=true`, which runs on resume when its window was
missed — the same mechanism that makes `systemd-tmpfiles-clean.timer` on this box fire reliably
while my crontab line did not. Proved the unit **runs** (`systemctl --user start` → `Result=success`,
correct block appended, **both** root headers present), not merely that it is scheduled. Crontab
entry removed so there is one scheduler; the other two lanes' jobs verified present afterwards.
⚠ `Linger=no`, so it still stops at logout — needs root, recorded rather than left as a surprise.

**What caught it was the warning I wrote into the crontab comment and the RUNBOOK** — *"a MISSING
block means the job did not run; it must NOT read as nothing is wrong."* That is the one thing
today that worked exactly as designed, and it is the reason I opened the log instead of glancing at
`df` and calling it fine.

### Three for three, and the pattern is now the finding

Every time this lane has measured, it has measured the container it already knew about:
`/tmp` while the producer moved to disk · the scratch tree while `du` could not see Docker ·
the crontab while the machine was asleep. **Each answer was true and none of them answered the
question.** The generalisable check is the same in all three: name the thing that would look
identical whether you were right or wrong, and go and measure *that*.
