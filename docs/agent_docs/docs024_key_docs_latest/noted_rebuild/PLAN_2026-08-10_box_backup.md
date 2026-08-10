# PLAN — off-box backup for the noted database

2026-08-10. Everything in §3 was **tested against real B2**, not read from
documentation; the probe key and its test file have been deleted.

---

## 1. What this is protecting against, and what it is not

Local dumps already exist (`/var/backups/noted/`, nightly, 14 days, restore
proven). They cover a bad migration or a `DELETE` without a `WHERE`. This plan
covers the three they cannot:

| Threat | Covered by |
|---|---|
| The box is lost, wiped, or the provider terminates it | the copy being off-box |
| Someone gets root on the box and destroys the backups | a key that cannot delete + Object Lock |
| Someone gets root and wants to **read everyone's notes** | encryption *before* upload |

The third is the one most easily forgotten and the most damaging here. This
product's entire remaining privacy position, once sign-in ships, is "your notes
are on our server; we could read them; we don't". Shipping nightly unencrypted
dumps of every user's private writing to a third-party bucket would make that
sentence materially worse — and it would be true whether or not anyone ever
looked. **So the dumps are encrypted on the box, to a key the box does not
hold.**

## 2. Design

```
pg_dump -Fc  ─►  age -e -R /etc/noted/backup-recipients.txt  ─►  b2 file upload
   (as today)      (public key only — the box CANNOT decrypt)     (write-only key)
                                                                        │
                                                          personae-noted-backups
                                                          Object Lock, governance 30d
```

**Encryption: `age`** (in Ubuntu 24.04 apt as `1.1.1-1ubuntu0.24.04.3`; not yet
installed). Chosen over gpg because it has no keyring, no agent, no trust model
to misconfigure — one recipient file, one flag.

The box holds **only the public key**. It can encrypt and cannot decrypt, so
rooting the box gets you the ability to write new backups and no ability to read
old ones. The private identity never goes on the box.

> **THE RISK THIS CREATES, STATED PLAINLY: lose the age identity and every
> off-box backup is permanently unreadable.** No support call recovers it. It
> must exist in at least two places the owner controls (password manager + one
> offline copy), and the restore drill in §6 must be run once *from those copies*
> before anyone relies on this. A backup you cannot decrypt is not a backup, and
> this failure mode is silent until the day you need it.

**Naming:** `noted/pg/noted-YYYYMMDDTHHMMSSZ.dump.age`. Prefix `noted/pg/` so
one key can later be scoped to it if the bucket ever holds anything else.

## 3. The B2 key — every setting, and what each one buys

### 3a. The bucket (already created)

```
personae-noted-backups   allPrivate   isFileLockEnabled: true
defaultRetention: governance, 30 days
```

- **`allPrivate`** — no public reads, ever.
- **`--file-lock-enabled`** — set **at creation**. It *can* be retrofitted with
  `b2 bucket update --file-lock-enabled`, but it **cannot be reverted**, so
  creating the bucket with it on is the clean path.
- **Governance, not compliance.** Compliance retention cannot be overridden by
  *anyone*, including the account owner, until it expires — the CLI's own help
  warns it "may lead to high storage costs". Governance gives the same
  protection against the backup key while leaving the owner an escape hatch
  (`--bypass-governance`). **Tested:** deleting a retained file failed even with
  the full admin key, and succeeded only with `--bypass-governance` added.

### 3b. The key

```bash
b2 key create --bucket personae-noted-backups noted-box-backup listBuckets,writeFiles
```

| Setting | Why |
|---|---|
| `--bucket personae-noted-backups` | Restricts the key to this bucket alone. **Tested:** `b2 ls b2://portfolio-sites/` → `ERROR: Application key is restricted to buckets: ['personae-noted-backups']` |
| `writeFiles` | Upload. The only thing the backup job needs. |
| `listBuckets` | **Required by the b2 CLI, not by the task.** Without it, `b2 account authorize` refuses outright: *"application key has no listBuckets capability, which is required for the b2 command-line tool"*. On a bucket-restricted key it reveals only that one bucket, so it leaks nothing. |
| *no* `readFiles` | **The important omission.** A stolen key cannot download a single past backup. **Tested:** download → `unauthorized`. |
| *no* `listFiles` | Cannot even enumerate what backups exist. **Tested:** `b2 ls` → `unauthorized`. |
| *no* `deleteFiles` | Cannot destroy history. **Tested:** `b2 rm` → `unauthorized for application key with capabilities 'writeFiles,listBuckets'`. |
| *no* `--duration` | A key that silently expires stops the backups silently. Rotate deliberately instead (§7). |

### 3c. The gap this probe found — `writeFiles` includes **hide**

**`b2 file hide` SUCCEEDS with `writeFiles` alone.** A hidden file vanishes from
normal listings and downloads: after hiding the probe, `b2 ls --recursive`
returned **nothing**, and only `--versions` revealed the truth (a 0-byte hide
marker over the real 43-byte object).

So the accurate claim is **not** "a compromised key cannot destroy the backups".
It is:

> A compromised key cannot **delete** or **read** the backups, but it **can hide
> them**, which looks exactly like deletion to anyone using ordinary tooling.

**The data survives and recovery is one command** — tested end to end:

```bash
b2 file unhide b2://personae-noted-backups/<path>     # then download normally
```

Two consequences that must be designed for, not just noted:

1. **Do not set a short `daysFromHidingToDeleting` lifecycle rule.** Hide is
   reachable by the backup key; permanent deletion N days after hiding would turn
   a reachable action into real destruction on a timer. Object Lock *should*
   block that within the retention window, but **I have not tested the
   lifecycle/Object-Lock interaction** and will not assert it. Keep
   `daysFromHidingToDeleting` comfortably longer than the retention period, or
   omit lifecycle entirely and prune from the admin side.
2. **Monitoring must look for absence.** A hidden backup is invisible to exactly
   the check most people write ("is today's file there?" — it isn't, and neither
   is anything else). See §5.

## 4. Retention

- **Object Lock: governance, 30 days** (set). Nothing can delete a dump inside
  30 days without an explicit `--bypass-governance` from an admin key.
- **Bucket lifecycle: not set yet, deliberately.** Proposal once real data
  exists: `daysFromUploadingToHiding: 90`, `daysFromHidingToDeleting: 45` — the
  second comfortably exceeds the 30-day lock so the two mechanisms cannot race.
- **Local: unchanged**, 14 days on the box.

Sizing is unknown until there is data. An encrypted dump of the empty database is
~2–4 KB; the honest position is that this is untested at scale and should be
re-measured once real notes exist, because audio and photos in Postgres would
change it by orders of magnitude — which is the other reason §3 of the main plan
puts media in B2 rather than in the database.

## 5. Monitoring — the part that makes it a backup rather than a hope

An unmonitored backup job is a job that has already silently stopped. Minimum:

- The nightly unit **fails loudly** on a short dump (already implemented — the
  300-byte floor) and on a non-zero upload exit.
- A **weekly check from OFF the box**, using an admin key, that asserts a dump
  exists for each of the last 7 days **and lists with `--versions`** so a hidden
  file is distinguishable from a missing one. Absence and invisibility must not
  look the same.
- The check must run somewhere other than the box, or a dead box takes the
  monitoring with it.

## 6. The restore drill — run BEFORE relying on this, then quarterly

Not "does the file exist" but "can we get the notes back from nothing but the
bucket and the identity":

```bash
b2 file download b2://personae-noted-backups/noted/pg/<file>.dump.age ./r.age   # admin key
age -d -i <path-to-identity> -o r.dump r.age                                    # OFF-box identity
createdb noted_restore_drill && pg_restore -d noted_restore_drill r.dump
psql -d noted_restore_drill -c '\dt'
```

**Do it from the owner's stored copy of the identity, not from a copy still
lying around from setup.** The thing being tested is the recovery path as it will
actually exist at 3am, and the step that fails is always the key nobody could
find.

## 7. Key rotation

No expiry is set, so rotation is deliberate: create the new key, deploy it to
`/etc/noted/b2.env`, run the backup once, confirm the upload, *then*
`b2 key delete` the old id. Rotate on any suspicion, and whenever someone with
box access leaves.

## 8. What the owner needs to decide or do

1. **Generate the age identity** (`age-keygen`), store the private half in two
   places off-box, give me the public half for the box. **This is the blocking
   step** — everything else is built around it.
2. **Approve the key creation** exactly as §3b, or tell me to change the
   capability set.
3. **Confirm retention**: 30-day lock + 90-day lifecycle, or different numbers.
4. **Decide whether `webdesign-chat`'s own data joins this.** Nothing currently
   backs up its `state.json` / `transcripts.jsonl`, and it is on the same box
   with the same "lost box, lost data" exposure. Not this workstream's, but the
   machinery would be the same and it is cheap to extend.

## 9. Status

| Piece | State |
|---|---|
| Bucket `personae-noted-backups` | **created**, allPrivate, Object Lock on, governance 30d |
| Capability set | **probed and verified**; probe key and test file deleted |
| `age` on the box | not installed |
| Age identity | **not generated — owner action, blocking** |
| Production key | **not created — awaiting approval** |
| Upload step in the backup script | not written |
| Off-box weekly monitor | not written |
| Restore drill | not run |
