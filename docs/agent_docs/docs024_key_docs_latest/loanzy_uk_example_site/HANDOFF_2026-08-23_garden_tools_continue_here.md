# HANDOFF 2026-08-23 — `garden-tools.uk` is wired and waiting; the build has NOT been run

**Start here cold.** This lane tests the **one-shot build route** — what the framework produces
from a domain name and nothing else — and feeds what it finds back to the lanes that own the
defects. It began as `loanzy_uk_example_site`; the current subject is `garden-tools.uk`.

**The single most important instruction: do not hint anything.** The owner's ruling is that the
next build gets **no prompt, no mission, no contact details, no seed** — the domain string is
the whole input. If you supply anything else, the experiment is void. Any deviation gets logged
in this lane's NOTES as a deviation, in the session it happens.

---

## 1. State, verified 2026-08-23

| thing | state |
|---|---|
| `garden-tools.uk` DNS | NS `alexis`/`leah.ns.cloudflare.com` — owner set them; propagated |
| Cloudflare zone | `82d90228c20877e2b3fc8470c2bc73d1` — **active** |
| DNS record | one proxied apex `A → 192.0.2.1` (TEST-NET-1; the proxy masks it) |
| Worker route | `garden-tools.uk/*` → `portfolio-sites-router` |
| `www` | record + route added; **301 → `https://garden-tools.uk/`** verified |
| Apex today | **9-byte `Not found`** — route live, bucket empty |
| Database | `sites` **0 rows**, `site_work_items` **0 rows** for the domain |
| Chassis | `v1.0.1330` |

The domain is **deliberately absent** from the positioning register — it is a reserved test
domain (`docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/RESERVED_test_domains.md`),
chosen because it naturally exercises guides, a directory and a calculator with no regulated
angle. **No register entry is owed.**

## 2. The pre-flight, then the one command

```sh
# 1. still clean? (both must be 0 — another session may have touched it)
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT (SELECT count(*) FROM sites WHERE domain='garden-tools.uk') AS sites,
         (SELECT count(*) FROM site_work_items w JOIN sites s ON s.id=w.site_id
           WHERE s.domain='garden-tools.uk') AS items;"

# 2. edge still right? READ THE BODY — a status code cannot fail on a parked domain
curl -s https://garden-tools.uk/ | head -c 40      # expect: Not found

# 3. chassis not just restarted (~300s silent-drop window)
kubectl -n ai-persona-system get pods -l app=agent-chassis \
  -o custom-columns=START:.status.startTime --no-headers

# 4. GO — nothing but the domain
bash scripts/initial_messages/020_build_pipeline/082_submit_domain_unified.sh garden-tools.uk
```

**Then verify it LANDED — the script cannot tell you.** `bugs_open/327`'s underlying drop is
fixed (live on `v1.0.1319`), but **the trigger script is unchanged since 2026-07-30** and still
prints ids and exits 0 regardless. One submission in three vanished this way on 2026-08-18.

```sh
kubectl -n ai-persona-system exec -i postgres-clients-0 -- psql -U clients_user -d clients_db -c \
 "SELECT w.item_type, w.status, w.item_key FROM site_work_items w JOIN sites s ON s.id=w.site_id
   WHERE s.domain='garden-tools.uk';"
# expect a needs_domain_research row, status triaged. No row = re-dispatch.
```

⚠ `bash <script>`, not `./<script>` — the file is mode 644.

## 3. What to check when it finishes — agreed with the owning lanes, not invented here

**(a) `bugs_open/311`'s after-test.** That lane pinned incumbent md5s **before** our run
(`docs/agent_docs/docs024_key_docs_latest/bugfix_311_component_keys/NOTES_311_fix.md`, commit
`527193376`): `824e3309` → `e6ee4b07f11d0b43c1c5a62667f4999f`, `b89f91e1` →
`a2c00f1c66ce6f4ef72b48083f1e3da6`, `7d8b0503` → `5f9534982e7f2bd776605ed78e755010`.
- **Diversion worked**: per tool section, a new `content_components` row
  (`function='<function>-garden-tools-uk'`, `forked_from` NULL, `section_type` = requested name),
  one `COMPONENT_COLLISION_DIVERTED` row in `agent_error_log`, item **complete**.
- **No collateral**: re-read those three md5s — they must be **UNCHANGED**. A run that gets its
  calculator by overwriting `loanandmortgagecalculator.co.uk`'s is the damage the old guard
  existed to prevent. Say so first and loudly if it happens.
- **The artefact, independently**: the fix guarantees *stored and linked*, **not** a good
  calculator. Fetch each tool URL and **count `<input>` elements**. `loanzy.uk` shipped a
  calculator page with **zero inputs** — a stored, linked, selector-visible component that still
  renders no tool is a different failure and must not be reported as success.

**(b) `bugs_closed/260`'s after-test.** That lane asked for a greenfield build as its "after",
and asked to hear the result **either way** — *"if it still fires I want to hear it from the lane
that ran it, not infer it from a quiet queue"*. Contribute as a **two-line pointer**, not a
section: domain · date · work item · anything differing from the known shape. Their fingerprint
is a four-token set (`{{end}}`, `{{if`, `{{.label}}`, `{{range`) and any token count is a
**ceiling**, not a count — `validate_content` caps at 10 per detector.

**(c) The route itself** — this lane's own subject. Everything that goes wrong is material for
`docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/HANDOFF_2026-08-19_fixing_the_one_shot_route.md`.

## 4. Bug state (2026-08-23)

**Closed since the loanzy run:** `260` (render leak), `307` (outage kills items), `311`
(component collision — section half, fixed AND the originating page healed), `317`, `286`,
`331`, `327` (the dispatch drop; fix live on `v1.0.1319`).

**Still open, and what each will do to this run:**
- **`bugs_open/326`** — *a failed build cannot be retried.* Re-submitting returns `COMPLETED`
  and queues nothing; `create_work_item` dedups on `item_key` in **any** status. If this build
  fails partway, recovery is hand-renaming `item_key`s (`… SET item_key = item_key || '_run2'`).
  **Know this before you need it.**
- **`bugs_open/328`** — any page that fails to build **stays linked** from the pages that did,
  so one failure reads as a broken site. Expect dead links if anything fails; they are the
  symptom of that bug, not of the build being wrong.

## 5. What NOT to do

- **Do not steer.** No mission, no contact details, no seed, no mid-flight repairs to make the
  result look better. If you fix something, it stops being a measurement of the route.
- **Do not judge from the DB.** `build_status='deployed'` means *pushed*, not *serving*; a page
  can be `complete` and absent. Verify at the served page, cache-busted.
- **Do not trust a queue you emptied.** Cancelling items does not stop an already-**claimed**
  one, and fleet sweeps file fresh work for any site row that exists (a sweep published a live
  page on `loanzy.uk` **77 minutes** after its build was stopped).
- **Do not conclude from a status code on any domain that might be parked** — a lander returns
  200 on every path. Read the body.

## 6. Falsifiers — what would make this handoff stale

- A newer handoff in this directory.
- `sites`/`site_work_items` non-zero for `garden-tools.uk` (someone started it).
- The apex not returning the 9-byte `Not found` (route or bucket changed).
- `326`/`328` moving to `bugs_closed/`.
- The 311 baselines being re-pinned, or that lane reporting the fix superseded.
- A chassis roll: re-check whether the fixes above are still the live set. **Ask whether the
  CODE was written before probing a binary for it** — a fix that does not exist cannot be in any
  image, and that question is free.

## 7. The docs, by path

- This lane: `docs/agent_docs/docs024_key_docs_latest/loanzy_uk_example_site/`
  — `HANDOFF_2026-08-19_fixing_the_one_shot_route.md` (the route's defects, ordered),
  `SUMMARY_2026-08-18b_the_guard_holds_and_the_site_is_live.md` (what the guard changed),
  `SUMMARY_2026-08-18_the_no_prompt_build_put_a_credit_broker_live.md` (the mistake, in full),
  `NOTES_loanzy_uk_example_site.md`, `RUNBOOK_loanzy_uk_example_site.md` (incl. the
  `garden-tools.uk` zone setup and its measured timings), `EVIDENCE_run1_*` / `EVIDENCE_run2_*`.
- The guard that made the last build safe: **CGV-032** in
  `docs/agent_docs/docs026_concept_register/register/content-governance.md`, shipped by
  `docs/agent_docs/sql_for_agents/464_classifier_regulated_business_needs_a_brief.sql`.
  **One control is still owed**: that a brief which *does* ask for a regulated model still gets
  one. Until then we have proved it declines, not that it still complies.
- DNS recipe: `docs/agent_docs/docs024_key_docs_latest/portfolio_positioning/RUNBOOK_dns_pointing_a_domain_at_the_serving_worker.md`.
