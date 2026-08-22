# PLAN — outward-facing site builds on their own cluster ("the customer satellite")

**Status: INITIAL PLAN, deliberately unfinished.** The owner asked for a starting point to
beef up in a separate thread (2026-08-22). Everything below is either cited prior art or a
live measurement taken today; the open decisions are listed as open, not guessed.

**Owner's words:** *"I'd like to have the outward facing site builds on a different cluster
to the one I use now."*

---

## 0. The headline: you have discussed this, at length, and it is mostly decided

You asked me to search for previous conversations. There are several, they converge, and
**one of them already reached your exact conclusion** — but on a different argument from the
one that is currently in favour, which is the single most important thing in this document.

| where | what it says | register |
|---|---|---|
| `PLAN_isolated_chat_environment(5).md` §12 (2026-05-27) | *"The build must run on the satellite, not core… the satellite becomes a **second, customer-facing instance of the entire platform**"* | `SAAS-001` |
| `RESEARCH_2026-08-18_throughput_to_thousands_of_domains.md` §0 | *"Nor is a second cluster: **it multiplies capacity we are not using**"* | — |
| same, §3 | when a cluster genuinely fills, the sanctioned shape is the **satellite**: self-contained cluster + Postgres + Kafka + scheduler per shard, *"**never** a second cluster reaching into uk_001's in-cluster singletons"* | `BIZ-014` |
| same, §9 "Closed options" | *"Coupled multi-cluster (shared Kafka/Postgres across clusters) as an isolation mechanism"* — **RULED OUT** | `MCL-001` |
| same, §10 **owner ruling D13, 2026-08-21** | *"first split = a **client satellite**; five seams near-term"* | — |

### The distinction that decides the whole plan

**A second cluster for CAPACITY is ruled out. A second cluster for ISOLATION is not.**

Three days ago the throughput review concluded a second cluster buys nothing, and it is
right: what binds today is the dispatch turn ceiling (~83 items/hour), the LLM spend cap,
the deploy fan-out and the Cloudflare zone cap — **none of which a second cluster relieves.**

Your reason is different. Outward-facing builds are *anonymous, internet-triggered and
token-spending*, and the May plan already put the argument better than I can:

> *"Building customer sites on core would re-introduce the exact load/hack/bug vectors
> isolation exists to remove (anonymous, internet-triggered, cost-bearing builds competing
> with the portfolio)."*

That is a **blast-radius and saleability** argument, and it survives the throughput
refutation completely. **So the plan must be justified on isolation, and must not claim a
throughput benefit** — the moment it does, it collides with a three-day-old measured
finding and the next reviewer will be right to reject it.

### ⚠ And one thing to settle with yourself first

Owner ruling **D13 (2026-08-21)** says the first split is a **client** satellite. Today you
asked for the **customer/outward-facing** builds to move. Those are different populations
and possibly different first steps. Either is defensible; they are not the same plan, and
whichever goes first pays the setup cost for the other. **This is decision D-1 below.**

---

## 1. What already exists, measured today (2026-08-22)

Good news and bad news, and the bad news is the useful kind.

| thing | state | evidence |
|---|---|---|
| `remote-job-spawner` | **deployed and running**, 174d | `kubectl get deploy` → `1/1` |
| `dispatch_agent` action | registered in the codebase | `MCL-001`, `platform/orchestration/actions/dispatch_actions.go` |
| **workflows using `dispatch_agent`** | **ZERO** | `SELECT count(*) FROM agent_definitions WHERE default_config::text LIKE '%dispatch_agent%'` → **0** |
| `agent_dispatch_log` (failure visibility) | **does not exist** | `to_regclass('public.agent_dispatch_log')` → NULL |
| second cluster `va001` | register says it once ran; **not verifiable from here** (one kubeconfig) | `MCL-002`, status *aspirational* |
| makefile `ENVIRONMENT`/`REGION` | already parameterised | throughput research §3: *"satellite wiring is cheap; the state layer is the design work"* |
| terraform envs | `production/uk001`, `development/uk_dev` | `deployments/terraform/environments/` |
| kustomize overlays | single `uk_001` per service (18 services) | `overlays/production/uk_001` |

**The mechanism for multi-cluster dispatch is built, deployed, and has never carried a single
real job.** That is not an argument against the plan — but it means every claim about it
working is untested, and `MCL-005` (Gap C) already says so: *"Before adding a second cluster,
prove the round-trip works with `target_cluster: 'default'` against the local spawner."*

**Note the shape though:** `dispatch_agent` is the *coupled* model — a remote agent reaching
back to core's Kafka and Postgres. That is the model §9 **rules out for isolation**. So the
existing machinery is probably **not** what this plan uses. It is prior art to learn from and
a tempting wrong turn, in equal measure.

---

## 2. The shape this plan proposes

A **satellite**: a self-contained second cluster that runs the outward-facing build product
end to end and never reaches back into core.

```
  CORE (uk_001, today)                    SATELLITE (new)
  ├─ portfolio sites                      ├─ its own Postgres
  ├─ client work                          ├─ its own Kafka
  ├─ admin + council + fix loop           ├─ its own chassis (same image)
  └─ the component/tool library ──────────┤   curated agent_definitions seed
       one-directional publish  ─────────►├─ its own object storage + B2 key
       (library, not site data)           ├─ its own LLM key (own spend cap)
                                          └─ customer sites are BORN here
```

**The four rules that make it a satellite and not a second cluster:**

1. **No inbound to core.** The only flow is core → satellite, asynchronous, publish-shaped.
2. **No shared state.** Separate Postgres, Kafka, storage bucket, and **a separate LLM key
   with its own spend cap** — the last one matters more than it looks: the account cap was
   hit twice in eleven days at 1/50th of target volume, and a customer build burst must not
   be able to starve client work.
3. **Two `sites` populations, not a copy.** Core owns the portfolio; the satellite owns
   customer sites, born and living there (`PLAN_isolated_chat_environment(5)` §12.3).
4. **The satellite is the blast-radius unit; the domain is the sale unit.** `BIZ-014`:
   *"operating thousands of domains does not require thousands of clusters."*

---

## 3. The prerequisite nobody gets to skip: BIZ-014's five seams

Both the May plan and the August research land on the same precondition, and the research
calls retrofitting it *"a forensic untangling"*:

1. **`owner_id` on site rows** — reuse the existing `clients → networks → sites` chain, which
   the owner already ruled canonical for customer identity (2026-08-10). `ADM-011` built the
   admin CRUD on it.
2. **An entitlement gate** at *both* build-submission and maintenance-run.
3. **A pluggable billing adapter** — never call Stripe directly.
4. **Credential parameterisation everywhere** — the satellite must take its own keys.
5. **A build-tier / cost-profile flag** (`saas_cheap` vs `portfolio`) driving cheaper
   model/batching choices, so a £149 build keeps its margin.

`[UNVERIFIED]` `BIZ-014`'s own `verify-later` asks *"whether an owner_id/entitlement layer or
billing adapter exists anywhere in the schema/codebase today"* — **I have not checked this,
and the beef-up thread should check it first**, because it sizes everything else.

---

## 4. Open decisions — for the other thread, not for me

- **D-1. Client satellite or customer satellite first?** D13 says client; today's ask says
  customer. Pick one, and say why the other waits.
- **D-2. Y-copy or Y-slim?** Y-copy deploys the existing monolithic image against new
  infrastructure — a *config* exercise, no recompile. Y-slim compiles the build actions out:
  smaller attack surface, defensible for an internet-adjacent cluster, but a second artifact
  to maintain for ever. The May plan's lean was **Y-copy first, slim later if wanted**, and
  noted that curating the seed does *not* shrink the binary or the latent attack surface.
- **D-3. Which provider?** The May plan §8: *"a separate provider/account improves credential
  and failure isolation"*, and explicitly **not** `va001` (it runs the coupled spawner).
  Note core runs on Rackspace **Spot** (D14: *"spot OK for now"*) — for a customer-facing
  product, reclaimable nodes are a different risk than for internal work.
- **D-4. What crosses the boundary, exactly?** The component/tool library, style collections
  and knowledge_base patterns are named as one-directional feeds. The publish mechanism is
  not designed.
- **D-5. Does the human review gate (D0b) live on core or the satellite?** You ruled on
  2026-08-21 that there is a human review gate before every release. If customer builds move,
  that gate moves with them, and it is the thing you personally operate.
- **D-6. The admin console** — see the companion plan. If builds move, the console has to
  reach two places, and that shapes both plans.

---

## 5. Suggested phasing (structural first, cheapest reversible thing first)

0. **Check the five seams exist** (§3's `[UNVERIFIED]`). This resizes everything.
1. **Settle D-1 and D-2.** One page. They are commercial/risk calls, not technical ones.
2. **Prove the loopback** — `MCL-005`'s Gap C, dispatch to the *same* cluster first, so a
   later failure is known to be networking rather than the contract. Cheap, and it exercises
   a mechanism that has never run.
3. **Stand up the satellite's state layer** — Postgres, Kafka, storage, its own keys. The
   research calls this *"the design work"*, as against the wiring which is cheap.
4. **Deploy the chassis image against it** with a curated seed (Y-copy).
5. **Move ONE outward-facing build end to end** and compare the artefact against a core-built
   one.
6. **Cut the boundary**: confirm nothing satellite-side can reach a core credential, DSN or
   Kafka address. The May plan §9.7 makes this an explicit verification step, and it is the
   one that decides whether you actually got isolation or just a second cluster.

---

## 6. Falsifiers — check these before trusting this document

- **The five seams.** §3 is `[UNVERIFIED]` and it is load-bearing.
- **`va001`.** `MCL-002` is *aspirational* on the strength of one present-tense sentence in a
  May document. Whether a second cluster exists today is unknown from this tree.
- **D13 vs today's ask** (§0) — one of them is now stale and I do not know which.
- **The throughput research's numbers** (~83 items/hour, the spend cap) are `[MEASURED
  08-19]` over a 7-day window and will drift.
- **This plan claims no throughput benefit.** If a later draft starts claiming one, it has
  drifted into the option that was measured and closed on 2026-08-18.

## 7. Sources

- `docs024_key_docs_latest/tools/tool_widget_clobber/PLAN_isolated_chat_environment(5).md`
  (§5 X-vs-Y, §8 where to run it, §12 build-as-a-service, §13 commercial model)
- `docs024_key_docs_latest/dispatch_throughput/RESEARCH_2026-08-18_throughput_to_thousands_of_domains.md`
  (§0, §3, §9 closed options, §10 owner rulings)
- `docs026_concept_register/register/multicluster.md` (`MCL-001`…`MCL-014`)
- `docs026_concept_register/register/saas-isolation-architecture.md` (`SAAS-001`, `SAAS-002`)
- `docs026_concept_register/register/business-strategy.md` (`BIZ-014`)
- `docs_archive/.../multicluster/FOCUS_multi_cluster_dispatch_mvp(2).md` (§4 cross-cloud,
  §8 Postgres strategy across clusters)
