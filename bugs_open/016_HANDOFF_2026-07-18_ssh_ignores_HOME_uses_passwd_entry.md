# 016 — `ssh` ignores `$HOME` and expands `~` from the passwd entry (FIXED, and my first diagnosis was wrong)

**Found:** 2026-07-18, provisioning the idea.uk box's pull-sync (`box/provision-pullsync.sh`,
RUNBOOK §3a). **Status:** fixed in the scripts; filed because the mechanism is a general trap for
*any* command we run as a service account, and because the first diagnosis was confidently wrong.
**Scope:** operator scripts, not platform Go code — but the pattern applies anywhere we
`sudo -u <svc> env HOME=… <thing that uses ssh>`.

## Symptom

Two different-looking failures, one cause. Running the provisioner as root on the box:

**Run 1** — clone aborted:
```
Cloning into '/var/lib/sitesync/repo'...
Host key verification failed.
fatal: Could not read from remote repository.
```
**Run 2** (interactive, after the "fix" below) — got further, then:
```
The authenticity of host 'github.com (140.82.121.4)' can't be established.
ED25519 key fingerprint is: SHA256:+DiY3wvvV6TuJJhbpZisF/zLDA0zPMSvHdkr4UvCOqU
Are you sure you want to continue connecting? yes
Could not create directory '/var/www/.ssh' (Permission denied).
Failed to add the host to the list of known hosts (/var/www/.ssh/known_hosts).
git@github.com: Permission denied (publickey).
```
Corroborating signal: the GitHub Deploy Key page showed the key as **"Never used"** throughout —
SSH never reached authentication.

## Root cause

**OpenSSH resolves `~` from `getpwuid()` — the passwd database — not from `$HOME`.**

`www-data`'s passwd home on Debian/Ubuntu is `/var/www`. The script ran:
```bash
sudo -u www-data env HOME=/var/lib/sitesync git clone git@github.com:gqls/vm-sites.git …
```
`HOME` correctly configured **git** (which does read `$HOME` for `.gitconfig`), but **ssh** ignored
it and looked in `/var/www/.ssh` for both the identity and `known_hosts`. So:
- the host key the script had written to `/var/lib/sitesync/.ssh/known_hosts` was never consulted
  → *"Host key verification failed"* (run 1);
- the deploy key at `/var/lib/sitesync/.ssh/id_ed25519` was never offered
  → *"Permission denied (publickey)"* (run 2);
- and `www-data` cannot create `/var/www/.ssh` (root-owned web root), so ssh could not even
  self-heal by writing the host key.

One cause, two symptoms that look like unrelated problems (a trust problem, then an auth problem).

## The wrong first diagnosis (recorded deliberately)

From run 1's message alone I concluded the cause was `ssh-keyscan` **exiting 0 having produced
nothing** — with `2>/dev/null` hiding it — leaving an empty `known_hosts`. That is a real latent
flaw and the hardening for it was kept (assert the scan is non-empty; print fingerprints to compare
against GitHub's published list; fail loudly with the SSH-over-443 fallback if port 22 is blocked).
**But it was not the cause**, and the "fix" could not have worked: the file it was carefully
populating was one ssh would never read. Run 2's `/var/www/.ssh` line is what actually identified it.

*Lesson:* "Host key verification failed" says the host key **wasn't found where ssh looked** — it
does not say the file you wrote is empty. I inferred the file's *content* from a message about the
file's *absence*, without checking which path ssh actually used (`ssh -v` prints it). Sibling of the
guide's **"0 rows is not decisive"** and **"a negative inference needs the mechanism checked"**.

## Fix (applied to `docs/…/idea_uk_vm_site/box/`)

Never rely on `HOME` for ssh — name the paths explicitly, and hand git the same command:
```bash
SSH_CMD="ssh -i /var/lib/sitesync/.ssh/id_ed25519 -o IdentitiesOnly=yes \
  -o UserKnownHostsFile=/var/lib/sitesync/.ssh/known_hosts -o StrictHostKeyChecking=yes"

sudo -u www-data env GIT_SSH_COMMAND="$SSH_CMD" git clone …
```
- `provision-pullsync.sh` — `SSH_CMD` defined once beside `SYNC_HOME`; used by the pre-flight auth
  test, the clone and the checkout. Pre-flight now also matches `/var/www/.ssh` explicitly and
  reports *this* bug by number rather than a generic failure.
- `sitesync` — same `GIT_SSH_COMMAND` exported at the top. **This mattered independently:** the
  5-minutely `git fetch` runs as `www-data` under systemd and would have failed the same way on
  every tick, i.e. the sync would have been dead on arrival even if the clone had been done by hand.
- `sitesync.service` — kept `Environment=HOME=…` (git wants it) but corrected the comment, which
  claimed it was what pointed ssh at the key.
- `IdentitiesOnly=yes` so ssh offers only this key rather than walking any agent identities.

## How to verify

```bash
# as the service account, with no HOME games — must greet with the repo name:
sudo -u www-data ssh -i /var/lib/sitesync/.ssh/id_ed25519 -o IdentitiesOnly=yes \
  -o UserKnownHostsFile=/var/lib/sitesync/.ssh/known_hosts -T git@github.com
#   → "Hi gqls/vm-sites! You've successfully authenticated…"   (exits 1 even on success)

# which paths ssh REALLY uses (the diagnostic that would have saved run 1):
sudo -u www-data ssh -v -T git@github.com 2>&1 | grep -iE 'identity file|known hosts'
```
Then the GitHub Deploy Key page's "Never used" flips to a timestamp — that field is the cheapest
proof of whether ssh ever reached authentication at all.

## Transferable pattern (filed to 016b §9)

**`$HOME` does not redirect ssh.** Setting `HOME` for a command run as a service account configures
tools that read `$HOME` (git, most CLIs) but **not** ssh's `~` expansion, which comes from the
passwd entry. Any service account whose passwd home is unwritable (`www-data` → `/var/www`) will
fail with *two* misleading messages — a host-key failure and a publickey refusal — that both point
at the credential rather than at the path. Name `-i` and `UserKnownHostsFile` explicitly (via
`GIT_SSH_COMMAND` for git), and confirm with `ssh -v` which files were actually opened.

Second, smaller pattern, kept from the wrong diagnosis: **a probe command's exit status can lie.**
`ssh-keyscan` exits 0 having written nothing, so `set -e` never fires; with stderr suppressed the
script proceeds on an empty artefact and fails far away. Assert the artefact is non-empty rather
than trusting the exit code. (Sibling of `003`'s "don't swallow kcat stderr".)
