# NOTES — client accounts

Running technical record. Append only, newest at the bottom. Missteps included on purpose.

---

## 2026-09-04 — survey session, evidence log

**Ask:** *"search the docs and discussions and code and find out everything we've planned and
discussed about setting up client accounts. This lane can have the responsibility for it."*

### Search misstep worth recording
`grep -ril "client account"` over the whole repo returns **zero** hits. The phrase the owner
used is not the phrase the estate uses. The corpus says *customer* (customer account, customer
portal, customer login, account hub, customer session), and the one place "client" is load-bearing
is the **`clients` table**, which is a different (and older) concept. Anyone repeating this search
should start from `customer`, `account hub`, `editor login`, `magic link`, `handed_over_at`.

### Live measurements, all `[MEASURED 2026-09-04]`, via
`kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db`

- `clients` = **3** rows (`Default Client`, `System Scheduler`, `Boxing Online` — the last created
  2026-08-27 14:29Z, `email = aaa@designconsultancy.co.uk`, `tier`/`customer_status` both NULL).
- `networks` = **1** row, `00000000-0000-0000-0000-000000000002`, `client_id = …0001` (Default Client).
- `sites` = **60**, of which **42** carry a `network_id`; boxingonline.com and idea.uk both point at
  the single default network. **No site is joined to the `Boxing Online` client row.**
- `billing_orders` = 1, status `paid`. `vouchers` = 1.
- `customer_access_tokens` = 2 rows: one `zip_download`, one `confirm_transfer` (the 2026-09-03
  idea.uk rehearsal). Live CHECK: `purpose = ANY (ARRAY['zip_download','confirm_transfer'])` —
  `editor_session` is **not** admitted yet.
- `sites` with `handed_over_at IS NOT NULL` = **1** (idea.uk, 2026-09-03 19:30:31Z,
  `live_link_expires_at` 2026-10-15, `transfer_confirmed_at` NULL).
- `kubectl get ingress -n ai-persona-system` → **No resources found.** Only non-ClusterIP service
  in the namespace is `wireguard` (NodePort). No public route into the cluster.

### Code facts, each with its file:line
- `platform/orchestration/actions/site_db_actions.go:1085` — `createDefaultNetwork`, the ONLY
  `INSERT INTO networks` in the tree (`grep -rn "INSERT INTO networks" --include=*.go .` → 1 hit).
  It inserts `slug='default'` with `ON CONFLICT (slug) DO UPDATE`, so it can only ever produce one
  row. **Nothing creates a per-customer network.**
- `internal/core-manager/admin/customer_handlers.go:190` — the only `INSERT INTO clients`. Admin
  CRUD (ADM-011); reads sites via `JOIN networks n ON s.network_id = n.id WHERE n.client_id = c.id`,
  which today can only ever return the default network's sites for the default client.
- `platform/orchestration/actions/seed_build_queue_action.go:317` — `seedCustomerIdentity`. Writes
  `sites.email`/`company_name` **only** from `direction.published_contact` (bugs_open/420); the
  payer's address is delivery-only and stays in `build_queue.direction.customer_email`. It does
  **not** touch `clients`.
- `platform/delivery/handover.go:104` — `const LiveLinkWindow = 6*7*24*time.Hour` (42 days);
  `:106` `AdvertisedLiveWindowDays` (30) is a **named deliberate margin**, not a limit.
- `cmd/auth-service/main.go:185–240` — auth-service already exposes register/login/refresh, a users
  surface and a subscriptions surface. Unreachable from outside (no ingress) and rejected for
  customer auth by the Phase 5 design on the grounds that it is platform-user space.
- `frontends/user-portal/src/App.tsx` — **0 bytes.** A persona-era stub. Not a starting point;
  say so before anyone cites its existence as prior art.

### Register entries that bear on this
`payments.md` PAY-001…PAY-010 · `admin-dashboard-and-api.md` ADM-007 (public API, aspirational),
ADM-008 (`site_ownership` junction, **abandoned**), ADM-011 (customer CRUD), ADM-012 (approval
fan-out) · `business-strategy.md` BIZ-014 · `site-chatbot.md` CHAT-011 (the box brief store and
the `BR-` reference).

⚠ `payments.md` carries `covers-through: 2026-07-13` at its head — the extraction freeze. PAY-009
and PAY-010 postdate it and were added by hand. Absence from that file is not absence from the
platform (`bugs_open/106`).

### Unverified / not yet checked, marked so nobody reads them as findings
- `[UNMEASURED]` whether any **other** lane's product (finetuning.uk, noted.co.uk) would want to
  share one account system. `finetuning_uk_service/DESIGN_2026-09-03_examples_catalogue_shape.md` §3
  says outright *"The estate does not have a customer account system today for framework sites …
  So this is the first real one, and it is the largest piece here"* — i.e. a second lane is about to
  need exactly this. Not raised with them yet.
- `[UNMEASURED]` the free-tier ceilings in OPTIONS §B (Netlify team limits) — the OPTIONS doc says
  itself they have not been measured.
- `[INFERRED]` that the `Boxing Online` client row was created by hand or via `POST /admin/customers`.
  Nothing in the build path writes `clients`, so it cannot have come from the pipeline — but I did
  not find the actual call.

---

## 2026-09-04 (later) — the delivery lane had already written a pre-plan, and RFC_058 had already ruled

**I opened this lane without knowing either existed.** Two things I should have found in the survey
and did not:

1. `site_delivery_and_editor/PLAN_2026-09-04_preliminary_customer_accounts_for_the_client_accounts_thread.md`
   — a pre-plan written *for* this thread, the same day. The owner pointed me at it.
2. `architecture_review/RFC_058_…md` — **an OWNER RULING of 2026-09-03** deciding the identity model
   (Option C, four identities; multi-valued contacts NOT deferred; a fifth selling-party identity
   deferred). The survey's §1 cited the 2026-08-10 `clients→networks→sites` ruling as the live one.
   It is live, and it is no longer the whole story: RFC_058 postdates it by three weeks and changes
   the shape.

**What would have caught both, cheaply:** the survey grepped for the *mechanism* (`customer account`,
`account hub`, `magic link`) and never grepped `architecture_review/` for the *subject* (`identity`).
An RFC is where a decided-but-unbuilt thing lives on this estate, and it is exactly the directory a
"has this been decided?" question should hit first. Recorded because the survey's own §7 asked the
owner four questions, and **one of them (identity model) had already been answered by him the day
before.** Asking a question already ruled is not neutral — it invites a re-ruling.

**What survived the collision intact**, which is the useful half: both lanes measured the ownership
chain independently, hours apart, and agreed — the chain is a schema with no data in it. And the
three counts in circulation (33/54 on 09-02, 42/60 twice on 09-04) are all correct and differ only
by date. Six sites were built in between.

### Further code facts established this session, each verified

- **`EnsureSiteRecordAction` takes no network parameter** (`site_db_actions.go:178`) — it calls
  `getDefaultNetworkID` unconditionally and falls back to the literal
  `00000000-0000-0000-0000-000000000002` on failure (`:182`; same literal again at `:1031` in
  `createPlaceholderSiteRecord`). **So attribution cannot be supplied at build time today**, and a
  backfill without a producer change is stale on the next build.
- **`customer_access_tokens.site_id` is `NOT NULL REFERENCES sites(id)`** (migration 511). The token
  machinery is structurally per-SITE. A customer-scoped account page cannot be keyed on it without
  a migration.
- **Nothing enforces `live_link_expires_at`** `[MEASURED 2026-09-04]` — every Go reference is a
  write at handover, a follow-up eligibility predicate, or a test. Serving is unbounded, as
  migration 511's header predicted. So "we host for six weeks" currently fails in the *safe*
  direction, and the whole cost of Option C sits in the stop-serving half.
- **`scheduled_tasks`** has no retraction job: `[MEASURED 2026-09-04]` names matching
  `health|avail|endpoint|zip|deliver|followup|retract|expire` return five rows, of which the
  relevant live ones are `zip-link-refresh` (6h) and `site-discovery-rotation-availability` (5 min,
  `bugs_open/236`, probes least-recently-checked deployed sites for `site_unreachable`).
  **There IS an availability probe** — worth knowing before costing "someone to answer when it
  breaks", because part of the detection already exists.

## 2026-09-04 (later still) — owner rulings 3 and 4, and the separate-cluster suggestion

**Rulings** (full text and consequences in `PLAN_2026-09-04b_…md`): scope = an account WITH US ·
paid hosting COSTED not built · **token page now, login later** (a stated destination, so the token
page is the fallback, not a throwaway — the scope of what an account can read is the durable half and
authentication is the swappable one) · **one account per PERSON**.

Ruling 4 settles the §2b token fork on option (ii): `customer_access_tokens.site_id` becomes
nullable, `client_id` is added, and a **purpose-aware** CHECK requires the right one per purpose.
⚠ Recorded in the plan and repeated here because it is the trap: making `site_id` plainly nullable
would silently weaken a NOT NULL guarantee the two live purposes rely on, and **nothing would fail** —
a `zip_download` row with a NULL site would insert cleanly and only break at redemption. The CHECK is
the load-bearing half and should be **induced** in the verify block, as 511 induced its own two.

### Separate-cluster suggestion — measurements taken before answering

- `[MEASURED 2026-09-04]` `kubectl config get-contexts` → **one** cluster,
  `uk001-prod-agent-chassis-cluster`. MCL-002's `va001` is not in this kubeconfig.
- `[MEASURED 2026-09-04]` **`remote-job-spawner` is live and idle**: 1/1, 187d,
  startup line `cluster_id: uk_001`, `consumer_group: remote-job-spawner-uk_001`,
  `dispatch_topic: system.dispatch.requests`, provenance `239ab3626`. So the *receiving* half of
  multi-cluster dispatch is deployed with nothing to do.
- Register: **SAAS-001** (Y-copy satellite) `aspirational`, nothing stood up; **MCL-001** `partial`;
  gaps **MCL-003** (cluster filter logs at `Debug` — invisible in our logs, so the filter cannot be
  verified), **MCL-004** (no consumer for `system.dispatch.responses` — a failed dispatch is silent
  until timeout), **MCL-008** (Kafka has no `spec.kafka.authorization`; everything is
  `User:ANONYMOUS` with full access). **MCL-008 gates any customer-facing satellite and is bigger
  than the satellite.**
- **BIZ-014**, quoted rather than paraphrased because it is the counterweight: the unit of
  blast-radius isolation is distinct from the unit of separability-for-sale, and *"operating
  thousands of domains does not require thousands of clusters."*

**The reply's spine, so it is not re-derived:** a separate cluster isolates one of three things
(serving / builds / data); serving does not run on the cluster at all so it isolates nothing there;
builds are the real prior proposal and the plumbing is live-and-idle; and for data it currently makes
things worse, because §2d's four unjoined stores would become five. **You cannot shard a relationship
you have not recorded** — which is why the suggestion raises the value of Phase 0 rather than
reordering anything.
