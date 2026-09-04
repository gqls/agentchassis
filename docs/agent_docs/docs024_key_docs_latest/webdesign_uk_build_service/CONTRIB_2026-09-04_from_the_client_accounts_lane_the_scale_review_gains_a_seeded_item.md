# CONTRIB from the `client_accounts` lane, 2026-09-04 — one seeded item for the parked scale review, and one measurement it should not re-derive

`PLAN_2026-08-21_todo_from_here.md` parks *"whole-architecture scale review incl. own cluster(s)"* as
backlog item (b), **AFTER the working site**. This note adds to that agenda and **does not un-park
it** — the owner raised the question today and the parking is his.

**The owner, 2026-09-04, to the `client_accounts` lane:** *"we could have a separate cluster for
them altogether."*

Answered in full at
`docs/agent_docs/docs024_key_docs_latest/client_accounts/PLAN_2026-09-04b_client_accounts_design.md`
(§ OWNER SUGGESTION). The short version, and the one measurement worth carrying:

**The compute half is further along than any status line suggests.**
`[MEASURED 2026-09-04]` **`remote-job-spawner` is LIVE in the production cluster and idle** — 1/1,
187 days old, startup line `cluster_id: uk_001`, consumer group `remote-job-spawner-uk_001`,
consuming `system.dispatch.requests`, provenance `239ab3626`. `dispatch_agent` is a registered
action. `kubectl config get-contexts` holds **one** cluster; MCL-002's `va001` is not in this
kubeconfig. **So the receiving half of multi-cluster dispatch is deployed with nothing to do: what is
missing is a second cluster and a reason, not a mechanism.**

**And the gate the scale review should cost first, because it is bigger than the thing it gates:**
register **MCL-008** — the live Kafka cluster has **no `spec.kafka.authorization`**, so everything
connects as `User:ANONYMOUS` with full access. Adding SCRAM to an external listener gates the
*connection* and leaves an authenticated user unrestricted. **A customer-facing satellite is not
safe without this, and turning cluster-wide authorization on is risky precisely because every
in-cluster app currently connects anonymously.** (Source: `multicluster/HANDOFF_multi_cluster_dispatch.md`
§3, "Authorization caveat"; that lane is dormant — no substantive commits in months — so this is
unlikely to reach the review any other way.)

**Two framings worth keeping when it is costed:**

- The isolation risk that bites FIRST is not compute, it is the **single Cloudflare account** —
  already named and costed as an A → B → C trajectory in
  `../site_delivery_and_editor/PLAN_2026-08-17_delivery_architecture_decisions.md` §4.4, with the
  trigger around ~500–1,000 domains. A separate compute cluster does not touch it: sites serve from
  B2 through the Worker and never run on the cluster.
- **BIZ-014**, quoted: the unit of blast-radius isolation is distinct from the unit of
  separability-for-sale, and *"operating thousands of domains does not require thousands of
  clusters."*

**What the `client_accounts` lane is doing with it meanwhile:** costing it as a fourth line in the
hosting costing the owner asked for (money · obligation · stop-serving · isolation), and **not**
reordering its own plan — because you cannot shard a relationship you have not recorded, and
`[MEASURED 2026-09-04]` the estate has one network, 42 of 60 sites funnelling through it, and one
customer row reachable from no site.
