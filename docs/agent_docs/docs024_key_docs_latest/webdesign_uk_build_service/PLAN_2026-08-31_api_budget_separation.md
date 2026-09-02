# PLAN 2026-08-31 — separating the chat's API budget, and how far to take workload separation

**Status: DISCUSSION-STARTER, owner-requested.** Nothing here is implemented. Written for
the owner to take to a separate thread; the one action that needs no discussion (§2) is
his to run whenever he likes. Context: the shared Anthropic spend ceiling silenced the
customer-facing chat twice in four days (2026-08-27 and 2026-08-30) because the
background fleet and the shopfront bill the same account.

## 1. The problem, precisely

One organisation's usage limit covers every Claude call the estate makes: the ~46-pod
chassis fleet (councils, tool-improver, rebuilds, discovery checks), the delivery
machinery, and the webdesign.uk sales chat. The fleet is the overwhelming majority of
spend and runs unattended; the chat is the only surface a live customer watches fail.
When the ceiling is hit, the chat fails closed to contact details — honest, but a
just-launched shop with a silent assistant, twice in its first week. The owner now holds
**a second API key on a separate account** and asks how hard it is to use.

## 2. The straightforward answer: the chat is ALREADY separable — one env line

> **✅ DONE 2026-09-02 17:13Z.** The chat is on the owner's separate key. Fingerprints
> `c3358af6406c` → **`cd3e51a196a7`**; the fleet stays on `79eafe5d414e`. Verified at the
> artefact, not the config: the RUNNING process reports the new fingerprint, the facts
> relay still fetches 27 facts (so the rest of the env file survived the one-line
> rewrite), zero `claude call failed` since, and a live question through the public edge
> got a real answer — *"Usually three or four days…"* — not the contact line. Backup at
> `/etc/webdesign-chat.env.bak-20260902T171340Z`.
>
> **The residual is the owner's and cannot be checked from here:** that the new key's
> account is genuinely separate. The preflight proves the key WORKS and is not capped; it
> cannot see whose billing it is. The Console's usage against the new account after this
> is the artefact that settles it.

> **CORRECTED 2026-08-31 (budget-separation thread), on three measurements against the
> live box and cluster. The conclusion of this section survives; two of its premises do
> not, and one of them would have made the swap look pointless.**
>
> 1. **The chat ALREADY HAS ITS OWN KEY. That is why "give it a separate key" is not, on
>    its own, the fix.** `[MEASURED 2026-08-31]` box `/etc/webdesign-chat.env` key
>    fingerprint `c3358af6406c`; cluster fleet key (`personae-default-secrets` →
>    `ANTHROPIC_API_KEY`) `79eafe5d414e`. Different keys — and both outages hit anyway,
>    because **the usage limit is a property of the ORGANISATION, not of the key.** The
>    RUNBOOK's own note that this is a "scoped Workspace key" is consistent with that and
>    must not be read as a separate budget. So the swap is only worth doing if the new
>    key is on a **different ACCOUNT** (which is what the owner holds). A second key on
>    the same account would change nothing at all, and would look like a fix.
> 2. **There is no second chat instance on the box today.** `[MEASURED 2026-08-31]`
>    `/etc/sitechat/` is EMPTY and `systemctl list-units 'sitechat@*' --all` returns
>    nothing; `webdesign-chat.service` is the only chat unit. The per-unit isolation this
>    section describes is real as a MECHANISM (the template unit and the runbook recipe
>    exist) but noted.co.uk is not running one, so the swap touches exactly one file and
>    inherits nothing. ⚠ The provisioning recipe in `RUNBOOK_webdesign_uk_build_service.md`
>    COPIES the key out of `/etc/webdesign-chat.env`, so after the swap every new site's
>    chat silently joins the NEW account's budget — correct here, but it is a decision
>    being made by a `grep`, so say which budget you are joining when you use it.
> 3. **The cap has lifted; the chat is answering again.** `[MEASURED 2026-08-31]` a
>    one-token preflight with the live key returned **HTTP 200**. Last failure was
>    2026-08-30 20:46Z. So this is no longer an outage being fought — it is a defence
>    being built before the next one.
>
> **The one-env-line claim itself is CONFIRMED** — the env file carries exactly one
> `ANTHROPIC_API_KEY` line, `claude.go` reads it per call from the process environment,
> and nothing else on the box or in the cluster reads it.

It is genuinely easy, because the chat was built separate: `sitechat` is its own binary
on the box (not in the cluster), reads its key from its own env file, calls
api.anthropic.com directly, and carries its own guards (a $10/day spend ceiling in its
store, per-IP limiting, a 20-turn cap). An API key is self-contained — a key from a
different account works in that env file with no other change anywhere.

The swap is the OWNER's to run (a key value must never pass through a session — standing
rule 2026-08-23), in his own terminal:

```bash
docs/agent_docs/docs024_key_docs_latest/webdesign_uk_build_service/box/swap-chat-api-key.sh
```

> **BUILT 2026-08-31, replacing the hand recipe that stood here.** The hand version was
> `ssh` in, edit the file, restart, read the journal — and each of those three steps
> fails SILENTLY. A mistyped key does not stop the service (`main.go` checks only that
> the key is non-empty), so systemd says `active`, `/health` says 200, and every visitor
> gets the fail-closed contact line: **the same symptom as the outage this whole plan is
> about, from a different cause.** Forgetting the restart changes nothing while every
> file-based check agrees with the file. And a restart that fails leaves the shopfront
> with no chat at all.
>
> The script preflights the new key against the real API and **writes nothing unless it
> returns 200** (proven 2026-08-31: an invalid key drew `401 authentication_error` and
> the env file was byte-identical afterwards, no backup taken); backs up and restores on
> a failed restart; rewrites exactly one line or refuses; and reports the fingerprint the
> RUNNING PROCESS holds, not the file's. `--status` is read-only and needs no key — any
> session can run it. `--check` tests a key and writes nothing.

Notes that matter:
- **Per-unit isolation exists as a mechanism**, ~~and `noted.co.uk`'s chat runs the same
  binary under its own unit with its own env~~ — **CORRECTED 2026-08-31: no second
  instance exists on the box** (see the correction above). Each site's chat *can* sit on
  whichever budget the owner chooses, one env file each; today there is one env file.
- The chat's own $10/day ceiling stays as the per-day brake **inside** whatever the new
  account's limit is; set the new account's limit modestly and the chat can never spend
  past either.
- **Residual after this swap**: the fleet can still exhaust its own ceiling and pause
  webdesign BUILD work (tool-improver, rebuilds, customer site builds, the delivery
  email's compose step). That work is batch-shaped and delay-tolerant — a paused build
  resumes; a silent sales chat loses a customer. The swap fixes the half that hurts.

## 3. Lighter than a second cluster: budget partitioning options in ascending weight

1. **(Done by §2)** Customer-facing chat on its own account/key. Solves the observed
   failure completely.
2. **Workspaces / multiple keys within one account** — the Anthropic Console supports
   partitioning an organisation into workspaces with their own keys and limits
   [verify the current Console feature set when this is decided; the owner's separate
   ACCOUNT achieves the same isolation with a second billing relationship]. Candidate
   split if wanted later: fleet (big, capped hard) · customer-facing (small, protected)
   · delivery/critical (tiny, protected). The cluster side would take per-service key
   env — a small terraform 047 change, same pattern as the Stripe pair — no code change,
   since every consumer reads its key from env already.
3. **Separate namespace / config for webdesign workloads in the SAME cluster** — buys
   config and key isolation without duplicating infrastructure. Most of what "own
   configurations" means is already per-agent/per-site rows in `agent_definitions` /
   `site_specs` / `billing_settings`; a namespace split would add deploy and env
   isolation on top.

## 4. The heavy option the owner named: webdesign on its own framework/cluster

**What it would buy**: total blast-radius isolation — deploys, schema migrations, spend,
config, incident surface. A fleet mishap could never touch the shop; the shop's rolls
could never straddle fleet commits.

**What it would cost, honestly**: the whole platform exists once — one DB, one Kafka,
one chassis image, one migrations discipline, one concept register, one landmines file,
one multi-session tree. A second cluster duplicates ALL of that, and the duplication
itself becomes the new failure class: config drift between clusters, double migration
runs, a tool library and evidence machinery that either fork or need cross-cluster
plumbing, and every "which cluster am I on" mistake this estate's history predicts.
The multi-session coordination costs this repo documents would roughly double.

**What it would NOT solve better than §2+§3**: the budget problem (already solved by key
separation — money is partitioned by KEY, not by cluster), and config isolation
(already per-site/per-agent in the data model). The honest trigger for revisiting C is
not budget but DIVERGENCE: if webdesign's load (many concurrent paying customers'
builds), its risk profile, or its release cadence stops fitting the shared fleet's,
that's when its own deployment earns its cost. Nothing observed so far points there.

## 5. Recommendation

Do §2 now (one env line, owner-run, five minutes). Consider §3.2's key split for the
fleet-vs-delivery boundary at leisure — it's a terraform change, not architecture. Hold
§4 unless a non-budget reason emerges; if the other thread wants to pursue it, the
first question to cost is the DB: shared clients_db across clusters is most of the
pain, and a webdesign-only DB means forking the build pipeline's whole data model.

## 6. Facts the other thread will want

> **STATE AS OF 2026-08-31, measured by the budget-separation thread.** The chat is on
> key `c3358af6406c`, the fleet on `79eafe5d414e`, and **both bill the same account** —
> which is the whole of the problem. The env file has been untouched since 2026-08-26
> 22:01 and the service has been running commit `160546543` since 2026-08-27 14:05, so
> **no swap has happened yet**. The cap is currently lifted (preflight 200). Guard
> figures below CONFIRMED: the env file sets no `DAILY_SPEND_CEILING_USD` or
> `MAX_TURNS_PER_CONVERSATION`, so the code defaults ($10.00/day, 20 turns) are what is
> actually in force.

- Two limit outages: 2026-08-27 ~11:39Z (raised same day) and 2026-08-30 ~20:46Z
  (limit reached again; access self-restores 2026-09-01 00:00 UTC; budget since raised).
- The chat's own guards: $10/day ceiling (persisted, survives restarts), 5 new
  conversations/hour/IP, 20 turns/conversation — all independent of the account limit.
- Key locations today: box chat = `/etc/webdesign-chat.env`; cluster fleet = terraform
  `047-base-configs` → `personae-platform-secrets`; the fleet key is NOT on the default
  console org (standing memory — check keys' "Last used" when auditing spend).
- Fleet spend shape: council prompts were ~85% of fleet LLM spend before migration 377's
  prefix caching cut 68% of it — the fleet is still the dominant spender by far.
- Chat services: `webdesign-chat.service` (webdesign.uk) + `sitechat@<domain>.service`
  template (other sites), one binary, per-unit env — the natural per-site budget seam.
