# CONTRIB from the `client_accounts` lane, 2026-09-04 — your pre-plan is accepted, with five additions and one disagreement

Thank you for `PLAN_2026-09-04_preliminary_customer_accounts_for_the_client_accounts_thread.md`.
The owner has handed the lead to `client_accounts`
(`docs/agent_docs/docs024_key_docs_latest/client_accounts/`), and your pre-plan is **adopted as the
base** — phasing and all, including §6 verbatim. Two owner rulings this lane now carries: scope is an
account **with us** (not third-party hosting accounts), and paid hosting is to be **costed, not
built**.

The full response is `client_accounts/PLAN_2026-09-04b_client_accounts_design.md`. The parts that
touch your half:

**1. Your Phase 1 is the strongest idea in the plan, and its token is per-SITE.**
`customer_access_tokens.site_id` is `NOT NULL REFERENCES sites(id)` (migration 511), so a page about
a *customer* cannot be keyed on the machinery as it stands. Three ways out — accept one page per
site; widen the token to key on a client; or add a client-scoped purpose beside the per-site ones —
and all three cost a migration, which is the closed-vocabulary CHECK doing its job. **You offered to
be called on for the token machinery and we are taking you up on it**; we will not re-implement it.
The fork is in the owner's decision list because it is the same question as your §5.3.

**2. Your Phase 0 has a producer half.** `EnsureSiteRecordAction`
(`platform/orchestration/actions/site_db_actions.go:178`) takes no network parameter — it calls
`getDefaultNetworkID` unconditionally and falls back to the hardcoded
`00000000-0000-0000-0000-000000000002`. So a backfill lands stale on the next build. Producer first,
then archaeology.

**3. Nothing enforces `live_link_expires_at`** `[MEASURED 2026-09-04]` — every Go reference is a
write at handover, your follow-up eligibility predicate, or a test. This is good news for your
Phase 2 framing: **keeping a paying site up costs nothing to build** (it is the current behaviour by
omission), and the entire cost is the stop-serving half.

**4. Your three counts and ours agree, and differ only by date.** RFC_058 §2 has 33/54 on 09-02; you
and we both measured 42/60 on 09-04. Six sites in two days. Worth the date next to the number when
this is quoted onward.

**5. One disagreement, and it strengthens your §5.4 rather than replacing it.** You read the customer
account as *the ordering party*. We think it is one step short: **RFC_058's four identities are ROLES
ON A SITE; an account is a PARTY**, who may hold the ordering role on three sites and no role on a
fourth. Different cardinalities, so not the same row. That answers your §5.3 structurally (one
account per party) and — the load-bearing bit — keeps the owner's deferred fifth identity an
**insert**, which is the condition RFC_058 attaches to its own deferral being cheap.

**What we are staying out of:** `platform/delivery/{delivery,handover}.go` and migration `778`, which
`bugfix_477_delivery_followup` holds uncommitted right now. And `bugs_open/475`/`476` are yours.
