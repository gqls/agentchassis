# NOTES — apis.uk bees home page

Running record, append-only, newest at the bottom. Includes the wrong turns —
those are the point, not an appendix.

---

## 2026-08-22 — session start: the constraint turned out to be the interesting part

Owner ask: *"build a page about bees for the apis.uk home page but without
affecting the dns for the tools-api that runs on that same domain."*

Took the constraint first, before designing anything, because it is the half that
can break something live.

**What the standing documents said.** Three of them agree, and all three are now
out of date:
- `features_open/020_FEATURE_apis_uk_traffic_probe.md`: *"apex will be repointed
  at the owner's planned bees homepage (separate thread) when that exists — one
  record swap, wildcard/probe unaffected."*
- `gauntlet_dead_cta/infra/island/RUNBOOK_island.md:76`: *"repoint ONLY the apex
  record at its hosting."*
- `gauntlet_dead_cta/NOTES_gauntlet_dead_cta.md:551`: *"apex rides the probe 404
  until then; swap is one DNS record."*

So the expected shape of this task was: build the site, then swap one apex DNS
record from the Cloudflare tunnel to the portfolio hosting.

**What is actually true [MEASURED 2026-08-22].** The apex is *already* served by
the portfolio worker, and no DNS change is needed at all. The zone carries **4**
DNS records and **2** worker routes as of 2026-08-22:

```
A      www.apis.uk   -> 192.0.2.1                      proxied
CNAME  *.apis.uk     -> f917c7c1-….cfargotunnel.com     proxied
CNAME  apis.uk       -> f917c7c1-….cfargotunnel.com     proxied
CNAME  tools.apis.uk -> f917c7c1-….cfargotunnel.com     proxied
routes: apis.uk/* -> portfolio-sites-router ; www.apis.uk/* -> portfolio-sites-router
```

A worker route intercepts at the edge *ahead of* the origin, so the apex CNAME to
the tunnel is vestigial. **Reading the DNS alone gives exactly the wrong answer**,
which is presumably how the three documents above came to agree with each other.

**Not inferred from config — the hostnames were asked, and they identify
themselves by their failure modes.** Four names, three distinct behaviours:

| hostname | response | therefore |
|---|---|---|
| `apis.uk` | 404, body `Not found` | the worker (`worker.js:91` returns that exact string) |
| `www.apis.uk` | 301 → apex | the worker (`worker.js:23` www→apex branch) |
| `zzqq-probe-test.apis.uk` | 404, **0 bytes** | island probe vhost :8082 |
| `tools.apis.uk` | 404, **0 bytes** | island Caddy :8081 → tools-api |

The body-length difference is what separates worker from island; the status code
alone cannot, since everything here 404s.

**The real hazard is not DNS at all — it is the worker ROUTE PATTERN.** DNS
records are per-name and independent, so the apex and `tools` never interfered.
But a route `*.apis.uk/*` would match `tools.apis.uk`, intercept the live API at
the edge, look up a B2 object that does not exist, and serve 404 — with no DNS
record touched and nothing looking wrong. `scripts/cloudflare/add_www_redirect.sh`
records that **24** zones already carry exactly that wildcard route as of its own
2026-08-18 measurement. apis.uk is deliberately not one of them. Filed as a
landmine.

**tools-api liveness control, before doing anything** [MEASURED 2026-08-22]:
`POST https://tools.apis.uk/api/v1/tools/gauntlet/round` with `Origin: vonc.com`
→ **200**. Root `GET /` → 404, which is the documented Caddy arm and NOT a
liveness signal (`WRONG_CALLS.md` records a session getting this exactly wrong).

## 2026-08-22 — scope put to the owner rather than guessed

The framework's fresh-domain pipeline builds a *whole site*; the ask was for a
home page. Rather than assume, asked. Owner chose **home page only** and a
**personal / enthusiast** angle — not beekeeping instruction, not conservation
campaigning.

Mechanism for binding that: the `roadmap_brief` site_spec. **Grounded in the live
agent definition, not in the oufe runbook that recommended it** — `build-site-planner`'s
prompt reads `{{.site_specs.specs.roadmap_brief.text}}` and states *"ROADMAP
OVERRIDES THE COMPONENT LIST. Build ONLY the pages listed in the current phase
below. … Do NOT invent additional pages. The roadmap is the authority for this
site."*

**Misstep, corrected in ~2 minutes:** went to `templates_db` for that definition
because the oufe runbook says `agent_definitions` lives there. It returned **0
rows**, which reads exactly like "no such agent" rather than "wrong database".
`agent_definitions` is in **`clients_db`** (**216** rows as of 2026-08-22);
`templates_db` has a same-named table with **8** rows as of 2026-08-22. Recorded
as a correction in the RUNBOOK, because the next person will make the same call
for the same reason.

**Checked that the cascade cannot undo the seed**, since a spec I write is worth
nothing if a later agent overwrites it. Queried every active agent's
`write_site_spec` steps by aspect: only `domain-submitter` writes `roadmap_brief`
(and its step needs `input_data.roadmap_brief`, which the 082 envelope does not
carry, so it errors to its `error_step` and moves on), and **nothing** in the
active set writes `evidence_base` or `imagery_style_guide`. Note that
`write_site_spec` **deep-merges** rather than replaces, so "something writes this
aspect" would not have been fatal — but it would have been a merge nobody
reviewed.

## 2026-08-22 — the evidence base, and three defects in my own seed

`loadEvidenceBase` (`platform/orchestration/actions/validate_page_content.go:1272-1290`)
returns `nil` on `sql.ErrNoRows`, and every claims lane then silently no-ops. So a
site with no evidence base is not "unchecked but fine" — it is **unchecked and
reporting clean**. Seeded before the first page exists.

Bees are an unusually bad subject for this: the field is made almost entirely of
famous repeated numbers (share of food owed to pollinators, flowers per jar, miles
flown, bees per hive, percentage declines, species counts), every one quotable
everywhere and sourced nowhere we hold. Plus one specific famous misattribution —
the "four years left to live" line Einstein never said. So `facts[]` is empty and
the bans target **shapes**, per the oufe precedent. **27** bans and **41** allowed
entities as of 2026-08-22.

**Then tested the ban list against sentences, and it failed three times.** This is
the part worth keeping, because every cheaper check had already passed:

1. **`\\\\.` decodes to a regex meaning "literal backslash", not "decimal point".**
   The seed's own stated safety net — *"an invalid regex degrades to a literal
   substring, so a typo never silently drops a ban"* — **does not fire**, because
   the pattern is *valid*, merely wrong. A valid-but-wrong regex never matches and
   reports clean for ever. `jq -e .` passes it happily; only decoding the string
   and *running* it shows the problem.
2. **`2 million flowers` escaped every pattern.** The digit-adjacent bans require
   the number next to the noun and cannot see a magnitude word between them — and
   "two million flowers to a jar of honey" is the single most repeated bee
   statistic there is. Added digit, spelled-out and bare-plural magnitude bans.
3. **`lives for 6 weeks` escaped**, because the pattern said `live` without the
   optional `s`.

Each was found by asserting on *sentences*, not by inspecting the JSON. The suite
now holds **22** sentences that must be caught and **8** ordinary bee sentences
that must stay clean, as of 2026-08-22 — and it earns its keep precisely because
it *did* come out negative three times.

**Misstep:** the first `cd`-relative fix attempt silently did nothing. The shell's
working directory had already been changed by an earlier call, so the `cd` failed
and `&&` short-circuited — and the "fixed" message never printed but the
re-validation ran anyway and showed the old patterns. Caught it because the
decoded patterns still showed four backslashes. Re-ran with absolute paths.
*A fix you did not see apply has not applied.*

## 2026-08-22 — seed applied, on the second attempt

First apply **failed and rolled back**: the `imagery_style_guide` insert was
missing its `FROM sites WHERE domain='apis.uk'` clause, so `id` was unresolvable
(`ERROR: column "id" does not exist … cannot be referenced from this part of the
query`). The transaction wrapper did its job — `SELECT count(*) FROM sites WHERE
domain='apis.uk'` returned **0** afterwards, i.e. the site row inserted moments
earlier was rolled back too. Fixed the clause, re-applied clean.

Verified after: site row with email present (`bugs_open/063` — the hallucinated-email
check FAILS OPEN with no contact email), three specs `is_current` and `pinned`,
**27** bans / **0** facts / **41** entities.

Submitted 12:18Z. `CORRELATION_ID=ba7a9c24-aea3-4fd0-9def-7e1d6f1cf891`.
Chassis pods were ~3.5 h old, well clear of the ~300s post-restart window in which
spawns are silently dropped.
