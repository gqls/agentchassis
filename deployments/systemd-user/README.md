# systemd *user* units — `scratch-report`

Versioned copies of what is installed at `~/.config/systemd/user/`. Install with:

```bash
cp deployments/systemd-user/scratch-report.{service,timer} ~/.config/systemd/user/
systemctl --user daemon-reload && systemctl --user enable --now scratch-report.timer
systemctl --user start scratch-report.service     # prove the unit runs, don't just schedule it
systemctl --user list-timers scratch-report.timer
```

## Why a timer and not the crontab line this replaced

**This box is a laptop and it sleeps.** `[MEASURED 2026-08-25]` it was suspended
**23:15 → 09:54**, so the `41 6 * * *` crontab entry armed the previous afternoon never fired,
and never would have on any night with the same pattern. Plain `cron` does not replay a window it
slept through — `anacron` covers only `/etc/cron.{daily,weekly,monthly}`, not a user crontab line.

`Persistent=true` runs the job on resume when its window was missed. That is the whole reason the
unit exists, and it is why `systemd-tmpfiles-clean.timer` on this same machine fires reliably while
the crontab entry did not.

> **⚠ `Linger=no` for user `ant` as of 2026-08-25** — a *user* timer runs only while the user has a
> session. It survives suspend, but not logout/reboot-without-login. `loginctl enable-linger ant`
> fixes that and needs root. Until then: a missing log block after a *logout*, as opposed to a
> suspend, is expected rather than a fault.

> **⚠ A MISSING block in `/home/ant/scratch-report.log` means the job did not run.** It must never
> read as "nothing is wrong" — that is exactly how the 2026-08-24 crontab failure was caught.

Full account: `docs/agent_docs/docs024_key_docs_latest/tmpfs_exhaustion/`.
