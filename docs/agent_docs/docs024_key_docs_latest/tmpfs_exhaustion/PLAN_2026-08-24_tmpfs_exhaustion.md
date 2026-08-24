# PLAN 2026-08-24 — bounding scratch, on BOTH filesystems

**Lane mandate (owner, 2026-08-24): this lane owns the persistent `/tmp` exhaustion problem.**
Prior work is `HANDOFF_2026-08-23_tmp_is_ram_not_disk.md`, which diagnosed it and shipped a
documentation fix plus an environment change. This plan starts from a re-measurement of that fix
and it is not the result the handoff expected.

---

## 1. The correction that resizes this lane

The 2026-08-23 work concluded (§10 of the handoff) that the recipe fix had worked: `/tmp` was
refilling at **+0.5 GB/day** against a prior ~3 GB/day, and on that basis the janitor was
"genuinely low priority".

**That measurement was correct and the conclusion drawn from it was wrong, because it only looked
at `/tmp`.** The recipe did not stop. It **relocated** — onto the disk, where the 2026-08-03
`CLAUDE_CODE_TMPDIR` change had pointed session scratch, and where nobody was watching.

`[MEASURED]` 2026-08-24 ~12:00 UTC, all figures re-runnable from `RUNBOOK_tmpfs_exhaustion.md`:

| | |
|---|---|
| `/home/ant/.claude-scratch` | **147 GB** |
| of which bare `git archive HEAD` extracts of this repo | **130 GB in 308 directories** |
| created in the last 7 days | **73 GB** → **10.4 GB/day** |
| free on `/` (`/dev/nvme0n1p2`, 937 GB, 87% used) | **123 GB** |
| **time to a full root filesystem at that rate** | **~12 days** |
| `/tmp` today | 4.9 G / 32% — the handoff's figure, unchanged |

So the lane's own remedy moved a bounded failure (a 16 GB tmpfs; `ENOSPC`; annoying, recoverable)
into an unbounded one (the root filesystem; breaks git, builds, and every session on the box), and
the new one now fills **sooner** than the old one — ~12 days against ~22.

**The generalisable form, which is the finding:** *a bigger container is not a bound.* Redirecting
an unreaped producer to somewhere roomier buys time proportional to the size ratio and changes
nothing else — and it costs you the symptom that was telling you the producer existed.

## 2. Why the pasted recipe never reclaims anything `[MEASURED]`

Every copy of the HEAD-check recipe contains an `rm -rf`, which is why it reads as self-cleaning.
**It is the SETUP half** — it clears the directory the run is about to use:

```bash
rm -rf $SP/headtree && mkdir -p $SP/headtree && git archive HEAD | tar -x -C $SP/headtree
```

It reclaims a tree of **the same name**. In practice every run picks a new name, so it reclaims
nothing. One session on the morning of 2026-08-24 left `headtree`, `headtree2`, `headtree3`,
`headfinal`, `ht5`, `ht6` — **six live copies, ~2.8 GB, in one morning**, each one's setup `rm`
faithfully clearing only itself.

This is disconfirmable and it came out the right way: if the setup-`rm` reading were wrong you
would see **one** directory per session, repeatedly overwritten. You see N, with drifting names.

## 3. Why the 2026-08-23 documentation fix did not take

`scripts/verify-head-builds.sh` (OPP-008) is sound — it writes to disk, refuses a tmpfs target by
filesystem type, and deletes its tree on exit. It is also **essentially unadopted**: of the 50
extracts created since it shipped, **none** use its layout; all are hand-rolled.

Two reasons, both structural rather than anybody's fault:

1. **The census was wrong by 8×.** The handoff says the recipe is in "9 documents" and 8 were
   updated. `[MEASURED]` 2026-08-24: **73 documents** carry a hand-rolled `git archive HEAD | tar`,
   and **66 of them have no cleanup**. Fixing 8 of 73 leaves the practice intact.
2. **The fix was placed only where the lanes that already knew would look.** `CLAUDE.md` — the one
   file every session loads unasked — mentions neither the script nor the trap (**0 hits**,
   2026-08-24). So does `MEMORY.md` (**0 hits**). A session does not consult another lane's RUNBOOK
   at the moment it types a command.

## 4. What this lane will do, in order of leverage

1. ~~**A janitor that reaps both filesystems.** ✅ SHIPPED — `scripts/scratch-janitor.sh`.~~
   > **⚠ CORRECTED 2026-08-24, hours after writing it. The janitor already existed and I did not
   > check.** `scripts/scratch-report.py --reap` — **OPP-005, deployed 2026-08-03**, in the concept
   > register the whole time: marker-verified, dry-run by default, both roots by design. It had
   > **97.1 GB reapable at its own default gate** and no evidence of ever having been run. My
   > duplicate is deleted (`0097d25de`) and its one genuine gap folded into the original.
   >
   > **So the handoff's §4.3 was not asking for a janitor. It was asking for a SCHEDULE, and did
   > not know it** — because it too concluded from the symptom rather than from the register.
   > *A silent mechanism is usually undriven, not missing.* The remaining work here is a crontab
   > entry, not code.

   **What the fold-in fixed, which is a real defect in OPP-005.** `ROOTS` listed `/tmp`, and the
   register entry claimed *"both tools read BOTH roots … a check that inspects only one will be
   confidently wrong"*. `[MEASURED 2026-08-24]` **it read one.** Every candidate came from
   `scratch_dirs()`, which requires a `<root>/claude-*/<proj>/<uuid>` layout, and `/tmp` has never
   had a `claude-*` directory. Inert for three weeks while looking covered. **The tell was an
   absent section header** — no `=== /tmp ===` in the output — not a wrong figure. `loose_reapables()`
   now scans a root's top level and one level below it for the two regenerable shapes
   (marker-verified extraction, `go-build[0-9]+`), and `/tmp` reports for the first time: 7 dirs,
   2.7 GB.
2. **Put the pointer where sessions actually load it** — `CLAUDE.md`, and the memory index. This is
   the "a silent mechanism is usually UNDRIVEN, not missing" pattern: the mechanism exists and
   nothing drives it.
3. **A `pattern-check.py` rule** so document number 74 cannot spell the recipe out again. Cheap,
   and it is what stops the census going stale by addition a third time.
4. **The 108 GB standing backlog** — the owner's call, not a session's. It is other sessions' data
   on shared ground, and the last cleanup was done on the owner's instruction for exactly that
   reason.
5. **NOT: raise the tmpfs, and NOT: assume disk is free.** §1 of the handoff still stands on the
   first. The second is the mistake this plan is correcting.

## 5. The reaper's design decisions, and why

> **⚠ This section was written about `scratch-janitor.sh`, which no longer exists.** The reasoning
> survives because `scratch-report.py` had independently reached the same three conclusions in
> 2026-08-03 — shape not age, positive identification only, dry run by default. **That agreement is
> the strongest argument that the design is right, and the sharpest evidence that I should have
> read the register before writing code.** Kept as written, with the tool renamed where it is a
> command rather than a claim.

- **The disk side reaps by SHAPE, not by age.** A session scratchpad holds the disposable extract
  *and* the session's real work product — its notes, its analysis files — in the same directory.
  An age gate cannot tell them apart. So the disk side only removes two shapes that are regenerable
  by construction: a bare extract of **this repo** (module line in `go.mod`, and **no `.git`**), and
  `go-build*` linker scratch. Everything else is left alone however old it is.
  **Shape beats name**: a variant nobody has invented yet is still caught, which matters because
  the drifting-name behaviour in §2 is the whole mechanism.
- **`/tmp` keeps an age gate**, because its contents are not session work product — with the system
  directories excluded **by name**. That exclusion is not optional: the first cut of the 2026-08-23
  cleanup was a bare `find /tmp -mmin +1440 -exec rm -rf` and would have taken `.X11-unix`,
  `.ICE-unix`, `snap-private-tmp` and 11 `systemd-private-*`.
- **Gated on idle time, never on ownership.** The polluter and the victim are usually different
  sessions, and a finished session's 1.7 GB looks exactly like a live one's. The script refuses an
  idle gate below 2h rather than trusting the caller.
- **Dry run by default.** It deletes on shared ground; `--apply` is a deliberate act.
- **`--self-test` plants each hazard and asserts the guard fires**, with a control proving the
  candidate list was non-empty first — otherwise every refusal could just mean "nothing was ever a
  candidate". Ported onto `scratch-report.py`, which had no test at all. Six cases; the age-gate
  case deliberately re-uses **the same directory** as the control, so "held back" cannot be
  confused with "never found". See NOTES for what that test caught on its first run, which was
  itself.

## 6. Open questions for the owner

- **The 108 GB.** Delete at the shipped gates, or wait? (`README_where_we_are.md` puts the choice
  in plain terms.)
- **Should the janitor run on a schedule?** A crontab entry is the estate's existing pattern
  (two entries there already). Nothing reaps today, so without a schedule this remains a manual
  tool and the accumulation resumes the moment attention moves.
- **Should `/tmp` be a tmpfs on this box at all** (handoff §4.4)? Still open, still a system-level
  decision. Worth noting `systemd-tmpfiles-clean` **already runs daily** with a `10d` age gate for
  `/tmp` (`/usr/lib/tmpfiles.d/tmp.conf:11`) — a janitor that by construction cannot fire before
  the failure, since `/tmp` historically went from empty to full in about four days. A one-line
  drop-in in `/etc/tmpfiles.d/` would shorten it, and needs root.
