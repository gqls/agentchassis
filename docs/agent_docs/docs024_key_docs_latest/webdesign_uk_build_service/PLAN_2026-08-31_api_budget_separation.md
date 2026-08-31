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

It is genuinely easy, because the chat was built separate: `sitechat` is its own binary
on the box (not in the cluster), reads its key from its own env file, calls
api.anthropic.com directly, and carries its own guards (a $10/day spend ceiling in its
store, per-IP limiting, a 20-turn cap). An API key is self-contained — a key from a
different account works in that env file with no other change anywhere.

The swap is the OWNER's to run (a key value must never pass through a session — standing
rule 2026-08-23), in his own terminal:

```bash
ssh -i ~/.ssh/webdesign_box_ed25519 root@webdesign.vs.mythic-beasts.com
# edit /etc/webdesign-chat.env: replace the ANTHROPIC_API_KEY=... line with the new key
# then:
systemctl restart webdesign-chat && journalctl -u webdesign-chat -n 5 -o cat
# the startup lines confirm the service; then one chat message from the site proves the key
```

Notes that matter:
- **Per-unit isolation already exists**: `noted.co.uk`'s chat runs the same binary under
  its own unit with its own env (`/etc/sitechat/<domain>.env`) — it keeps whatever key
  it has. Each site's chat can sit on whichever budget the owner chooses, one env file
  each.
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
