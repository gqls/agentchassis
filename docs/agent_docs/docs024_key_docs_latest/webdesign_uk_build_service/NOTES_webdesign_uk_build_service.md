# NOTES — webdesign.uk build service

Running record, append-only, **newest at the bottom**. Evidence, commands, what
the system actually said, and every misstep.

---

## 2026-07-28 — session 1: grounding the plan

Owner asked for thinking and planning only, and for prior discussion to be found
first. Both done before a line of the plan was written.

### Prior art found (the plan is not new — it was recorded)

`webdesign_couk/PLAN_2026-07-27b_buying_design.md` §8 already records this
direction in the owner's own words, including *"we stand up a copy chassis in its
own cluster with its own database"*, and explicitly says **recorded, not started**.
It also already did two of the checks I was about to repeat: that
`cmd/tools-api`/`internal/tools-api` exist with exactly one live endpoint, and
that `idea_uk_vm_site/` is the precedent that has run to a completed paid
transaction.

**Misstep avoided by looking:** I would otherwise have written §8's content again
as though it were a finding. The prior-art search cost about four minutes.

Also relevant and read: `vm_estate/PLAN_2026-07-25_framework_controlled_vm_estate.md`
(the estate this box would join, and the owner's pull-only ruling),
`webdesign_couk/PLAN_2026-07-27_phase2_buyer_track.md` (superseded, but its
no-figures rail survives), `stripe/PLAN_stripe_billing_integration.md` (May, and
superseded in practice by idea.uk's working `billing.go`).

### What is actually built — checked, not assumed

**tools-api holds no chassis coupling.** `cmd/tools-api/main.go` is config → DB
pool → gin router. Grep for `kafka|Kafka` across `cmd/tools-api/` and
`internal/tools-api/`: **no matches.** Routes are `/health` plus
`/api/v1/tools/gauntlet/{round,position,defend}` with CORS → rate-limit →
input-cap (`internal/tools-api/api/server.go:32-58`).

**core-manager does have an HTTP → pipeline seam**, and it is admin-only:

```
internal/core-manager/api/server.go:108   apiV1.Use(middleware.AuthMiddleware(authConfig))
internal/core-manager/api/server.go:134   adminGroup.Use(middleware.AdminOnly())
internal/core-manager/api/server.go:227   pipelineGroup.POST("/:name/trigger", ...)
```

So "we have no system to directly call the chassis" is right **for untrusted
callers**, and slightly wrong in general — the seam exists for the admin
dashboard. Worth knowing, because P4 may be able to reuse its handler rather than
invent one.

**Site creation is a human writing SQL.** oufe.com was created by
`oufe/SEED_2026-07-25_oufe_site_and_specs.sql` (site row + specs, applied out of
band with `psql -f`, deliberately *not* through the migration runner). Live check:

```sql
SELECT aspect, is_current, created_at::date FROM site_specs
 WHERE site_id=(SELECT id FROM sites WHERE domain='oufe.com') ORDER BY created_at;
-- 12 aspects seeded 2026-07-25: design_intent, mission_brief, identity,
-- roadmap_brief, vertical_landscape, strategy, briefing, resolved_composition,
-- imagery_style_guide, classification, content_direction, submission
```

**Build throughput, for the "next day or so" promise:**

```sql
SELECT s.domain, count(p.id) pages, min(p.created_at)::date, max(p.created_at)::date
  FROM sites s LEFT JOIN pages p ON p.site_id=s.id
 WHERE s.domain IN ('webdesign.co.uk','oufe.com','fundamentallyai.com','relojistas.com','idea.uk')
 GROUP BY 1;
-- oufe.com             8 pages  2026-07-25 → 2026-07-28
-- fundamentallyai.com 13 pages  2026-07-20 → 2026-07-25
-- relojistas.com      20 pages  2026-07-16 → 2026-07-28
-- idea.uk             21 pages  2026-06-21 → 2026-07-28
-- webdesign.co.uk     99 pages  2026-07-25 → 2026-07-27   (PORTED, not generated)
```

**[Caveat, and it matters]** those spans are *elapsed calendar days on a
human-driven build*, not machine time. They are evidence that "next day or so" is
achievable with a human in the loop; they are **not** a measurement of how long
the pipeline takes. Do not quote them as build times.

**Deploy path.** Default repo `sites` → GitHub Action → B2; per-site override to
`vm-sites` resolved by `resolveGitRepoNameDB` (`git_repo_resolution_test.go`
documents the failure it prevents: a VM site's artefacts silently landing in the
default repo). Only idea.uk and relojistas carry `github_repo='vm-sites'`; only
relojistas has a non-empty `deploy_config`.

**Cost is measurable.** `llm_call_log`: 45,205 rows, 2026-03-25 → 2026-07-28. So
the cost of one full site build is a query, not an estimate. This is the single
cheapest thing that could be done before pricing anything.

### Fleet size — the finding I did not go looking for

```sql
SELECT count(*) total, count(*) FILTER (WHERE status='deployed') deployed,
       count(*) FILTER (WHERE domain LIKE 'pool-%') pools FROM sites;
-- total 32 | deployed 14 | pools 17
```

The buyer-track positioning says *"about a thousand sites"*
(`webdesign_couk/README_where_we_are.md:405`;
`SUMMARY_2026-07-28_what_the_news_feed_taught_us.md:21`). Traced back, ~1,000
enters the record as a **target** in the scale arguments
(`OPEN_THREADS_RESTART_LIST.md:344`, `robot_hands_gripper_dossier/NOTES…:636` —
*"9 of the 296 exist for 2 of ~1,000 sites"*, *"at 1,600 domains"*), where it is
used legitimately to argue about per-site Go actions. It has drifted from target
to present-tense claim in outward copy.

It may be true of **domains owned** rather than **sites built**. Written up as
PLAN §12 for an owner decision, because it is already in outward-facing prose and
webdesign.uk would be selling on it.

### webdesign.uk DNS

```
dig +short webdesign.uk A      → (nothing)
dig +short webdesign.uk NS     → (nothing)
dig +short webdesign.co.uk A   → 172.67.192.20 / 104.21.92.109  (Cloudflare)
```

Owner confirmed mid-session: **"I haven't pointed the dns yet."** So the empty
result is expected, not a sign the domain is unregistered.

### An asset with no documentation

The cluster runs a **`wireguard` deployment** (`linuxserver/wireguard:latest`,
manifests under `deployments/kustomize/services/wireguard/`). Grep across
`docs024_key_docs_latest/` returns **no mention of it in any workstream**.
**[UNVERIFIED — I do not know what it is for or whether it is in use.]** If it is
a working private transport to the VM estate it could change the P4 design. Noted,
not pursued; it is a P4 question, not a P0 one.

### Two dead ends, recorded

- **`internal/auth-service/subscription/`** looked like reusable billing — it has
  `StripeCustomerID`, `StripeSubscriptionID`, tiers, quotas. It is a **dormant
  skeleton for the original per-seat SaaS**, not for selling sites:
  `SELECT count(*) FROM subscriptions;` → `ERROR: relation "subscriptions" does
  not exist` in `clients_db`. Do not plan reuse around it. idea.uk's `billing.go`
  is the live path and it is one-off payments, which is what this product needs.
- **`pool-*.internal` sites** (17 rows, `status='pool'`) looked like they might be
  pre-built inventory a customer could be assigned. `content_data` on
  `pool-web-tech.internal` is `{}` — they are empty shells. Not investigated
  further; not load-bearing for this plan.

### Misstep

Queried `SELECT type, name FROM agent_definitions` → `ERROR: column "name" does
not exist`. The column is `display_name`. Cost one round trip; the fix is in the
RUNBOOK so the next thread does not repeat it. This is the "schema first: `\d
<table>` before writing SQL" rule in CLAUDE.md, skipped because the query felt
too small to check.
