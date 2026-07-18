# idea.uk → VM: status summary

**As of 2026-07-16.** Companion detail: `PLAN`, `RUNBOOK`, `RUNNING_NOTES` in this directory.

---

## In one paragraph

idea.uk exists as two disconnected halves: a chassis-built static marketing site that deploys to
Backblaze B2 (where **nobody sees it** — DNS points at the VM), and the live £29 report tool on a
Hetzner box whose nginx currently serves *only* the tool. The goal is one complete site at idea.uk —
the marketing pages **and** the tool — behind the VM's nginx. This chat completed the groundwork:
the site is now a coherent set of pages, the code that lets a site deploy to the VM instead of B2 is
written and shipped, and the tool's public request form is hardened against the spam it was getting.
The remaining work is the cutover itself, which runs on the live box and needs the owner's hands. A
separate, unplanned security finding — real credentials committed to a public repo — was contained,
and **closed on 2026-07-17**: the owner rotated both exposed credentials and deleted the old AWS
user, so the values still visible in public git history no longer work.

---

## What we did

**Security (unplanned, found while mapping the tool).**
Real AWS SES email credentials and the tool's internal API key had been sitting in a public GitHub
repo since 4 June, inside a file named `…example` (which is why nobody had looked). The Stripe and
Anthropic keys there were only placeholders, so **the payment path was never exposed**. We scrubbed
the file, added an automatic guard that blocks any future secret from being committed, and verified
it. **Rotation completed 2026-07-17**: both real credentials replaced, the old AWS user deleted, and
email sending verified — the values still visible in public history are dead.

**Thread 1 — completing the site.**
The site had three catalogued-but-unbuilt pages whose links 404'd. Building them exposed a
**structural bug in the site planner**: re-planning a site silently discards the composition of
pages that were already built (and can't fill empty ones). We recovered the site cleanly — it is now
a coherent nine pages, all deployed, with two previously-broken navigation pages built and the "Free
Audience Check" now pointing straight at the live tool. The planner bug is written up as its own
fix-handoff so it can be corrected properly; left unfixed it will silently degrade any site that gets
re-planned.

**Thread 1 — moving the deploy target.**
The mechanism to let a specific site deploy to the VM instead of B2 was designed but never wired up.
We wired it (four small code changes), and it has now shipped in the current chassis image. Turning
it on for idea.uk is gated on one safety step (below) and is deliberately not done yet.

**Thread 2 — the tool's spam problem.**
The tool's public request form had no defences and was collecting fake submissions. We added the
standard stack — a honeypot field, a too-fast-submission check, per-address rate limiting, email
validation, and capture of the submitter's address so a block-list can be built later — all covered
by automated tests. It's ready for the owner to deploy to the box. We also corrected an earlier
handoff that had sent this work looking in the wrong place (the tool stores data in a file on the box,
not in the central database).

**Along the way** we set the site's public contact email to `idea.uk@contactforsales.com`, and logged
three infrastructure issues we noticed but didn't chase (a crash-looping deploy-runner replica, a
fleet-wide dead contact-form, and a build-system retry-churn bug) into the shared error-notes folder.

---

## Where we are

| Area | Status |
|---|---|
| Leaked credentials — repo scrubbed + guard installed | **Done** |
| Leaked credentials — **rotate the real values** | **Done 2026-07-17** (old AWS user deleted; email verified) |
| Site completed — 9 coherent pages, nav fixed | **Done** |
| Planner bug that caused a scare | **Recovered; fix handed off** |
| Per-site VM deploy target — code | **Done, shipped in current image** |
| Per-site VM deploy target — turn on for idea.uk | **Done 2026-07-16** (allowlist guard verified live first) |
| Shared deploy repo safety (domain→host allowlist) | **Done + proven live** (idea.uk skipped, relojistas deploys) |
| Deploy repo's Action runner (never existed; image lacked ssh/rsync) | **Fixed: runner live on image v1.0.1126** |
| idea.uk pages + assets in the deploy repo | **Done — seeded from the built artefact** |
| Request-form hardening — code + tests | **Done; ready to deploy to the box** |
| Contact email set | **Done** |
| Pull-sync on the box (site files now ON the server) | **Done 2026-07-18** — 8/8 pages, re-syncs every 5 min |
| **VM cutover — the site is LIVE on the server** | **Done 2026-07-18** — one origin, 16 tool paths proxied, verified |
| Money path proven end-to-end (Stripe test event) | Outstanding — the last cutover check |
| Real-client-IP behind Cloudflare | Not started (needs the proxied/DNS-only answer first) |
| Remove existing spam from the tool | Prepared; needs owner SSH to the box |

---

## Where we're going

**Owner actions (need access we don't have):**
1. ~~Rotate the exposed SES + internal key~~ **done 2026-07-17**.
2. **Deploy the hardened tool** — build the binary, copy to the box, restart. *(the tool has no CI)*
3. **VM cutover** when ready — put the static site behind the box's nginx while the tool keeps its
   own paths; DNS unchanged, rollback is a one-line nginx revert. The runbook lists all 16 tool paths
   that must keep working (the earlier plan listed only 7 — the gap would have silently broken the
   free taster and the operator flow).

**Update 2026-07-16: the safety step is done and VM deploy is ON.** The shared deploy repo now maps
each domain to its own box and skips unmapped domains (idea.uk deploys by pulling, not pushing —
verified live three times). En route we found and fixed the reason the deploy mechanism had *never
actually run*: the repo had no build runner, and the runner image lacked the ssh/rsync tools. The
repo now holds idea.uk's complete built site, updating automatically on every future build. The next
step is the owner's: set up the box's pull-sync, then the nginx cutover — one wrinkle added to the
cutover config: the static site ships its own copies of the terms/refund/privacy pages, so three
redirect lines keep the tool's versions (the ones buyers agree to) canonical.

**Can continue without the owner:** the planner-bug fix (handed off), and repointing the dead
contact-form.

---

## Decisions taken (recorded so they're not re-litigated)

- **Keep the static build; the VM is a second destination, not a second renderer.** Rendering stays
  in the cluster; the box just serves files. This is a strict *improvement* to reliability — today a
  tool crash takes the whole site down.
- **Serve the whole site from one origin (the box), not split static/B2 + tool/VM.** One cert, one
  DNS record, forms and cookies "just work". This stays the exception; B2 remains the default for the
  fleet's thousands of sites.
- **Pull, not push:** each box syncs its own files from the repo, so no single deploy key can reach
  the whole fleet.
- **The "Free Audience Check" links straight to the live tool**, no interstitial page.

## Open decisions (none blocking)

- `/privacy` after cutover — served by the tool or the static site? (default: the tool.)
- `/contact.html`'s form currently posts nowhere useful — repoint to the tool's request flow or a
  `mailto:`?

## Risks flagged (in the shared error-notes folder, `aaa_fails_to_mend/`)

- A deploy-runner replica has been crash-looping for 18 days — redundancy is gone, though deploys
  still work via its healthy twin.
- Generated contact forms across the whole fleet post to a dead endpoint.
- The build system sometimes re-does work it already finished (a retry-churn bug).
