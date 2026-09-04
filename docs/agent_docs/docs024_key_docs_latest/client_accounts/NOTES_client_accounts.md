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
