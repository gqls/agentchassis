# HANDOFF — portfolio_positioning — 2026-08-20. **START HERE IN A FRESH SESSION.**

Supersedes `HANDOFF_2026-08-19_continue_here.md`. Owner read-out:
`SUMMARY_2026-08-19_first_sites_live_and_the_wall_the_fleet_would_have_hit.md`.
Design: `PLAN_2026-08-19_one_flow_three_brief_sources.md` (+ its 08-19b addendum).

**Verified at the top of this file 2026-08-20 — re-check if hours have passed; this tree moves.**
- **Chassis `v1.0.1317`, build point `2d13d530d`** (2026-08-19 22:21Z; pods up 22:26Z).
  ⚠ **Found it by probing candidate commits, because the `build provenance` startup line had
  scrolled past `--tail=200000`.** And note the method: **`git merge-base --is-ancestor <your
  commit> <build point>`**, NOT a grep for your own sha in the binary — the binary carries only
  the ONE commit it was built from, so your commit is *always* absent and that means nothing.
  I made exactly that error today before catching it.
- **Both portfolio sites still LOCKED** under the owner's build halt.
- **Directory: 25 active mortgage lenders, 17 savings providers.**

---

## 1. ⛔ THE HALT and 2. THE OPEN DECISION

`sites.locked_at` holds `adversecreditmortgage.co.uk` (41 items queued) and
`remortgagecalculator.uk`. **Flow A is RULED** (owner 2026-08-19): one flow, brief always
present, three producers — human / generated / third-party — under one contract.

**Still open: the brief-writer itself is unbuilt.** Owner decisions already taken that shape it:

- **He reads EVERY brief**, with a few words of direction to add or change — so this is
  generate-then-EDIT, not approve/reject. `approval_mode` gives the hold; the edit is a
  `mission_brief` revision (`site_specs` already supersedes, so his version becomes current with
  the generated one preserved underneath). **Provenance matters** — a brief he touched should be
  distinguishable from one he waved through.
- **"The spec is aspirational, the plan is achievable."** Briefs may ask for anything; the
  PLANNER degrades explicitly and records a `capability_gap`. ⚠ 42 raised, **41 deferred** —
  triaging that queue is a prerequisite, not a nicety, before 1,500 briefs multiply it.
- **Third-party briefs need security screening** (untrusted text reaching a prompt; injection;
  steering toward a regulated identity) **and reasonability screening** (would we build it).
- **Third-party sites do NOT go on the owner's domains.** The register therefore covers OUR
  estate only. Their domains need a search-and-buy workflow — see §6.
- **The brief-writer should read the register, REPLACING RFC_037's classifier input** — reasoning
  in the addendum §C. The deciding argument: the owner reads briefs, so positioning in the brief
  is positioning he can correct; positioning fed to the classifier is invisible to him. Keep
  `RFC_037` open as the home for a *binding* collision check, which is the only review-sampling
  rule that scales to 1,500.

## 3. ✅ THE REGULATED-IDENTITY GUARD IS LIVE — and partial

**CGV-033 is live on `v1.0.1317`** (proved by ancestry). Six fleet-wide first-person patterns
refuse a site claiming to be an authorised firm; the only exemption is a recorded attestation
(firm, FRN, who checked it, what they saw), validated in full — any missing field means not
attested.

**The admit half is built too** (owner: *"a client who proves they are regulated… that is a
prime customer"*):

- `scripts/regulated/record_attestation.py` — records a proven attestation, validation mirroring
  the Go guard exactly, refuses to overwrite without `--replace`, supersedes rather than mutates,
  and writes the FRN as a **citable fact** so the permitted claim is also checkable.
- `cmd/regcheck` — tells an operator whether a site is attested and **which field is wrong**.

**⚠ THREE THINGS IT DOES NOT COVER, stated so nobody assumes otherwise:**

1. **The section-editor arm is NOT live.** `ae7a8d739` (2026-08-20) is not in `2d13d530d`. Until
   the next roll, a regulated claim inserted by `apply_section_edit` still bypasses the family.
2. **Chrome is not covered at all.** `render_site_components_action.go` / `chrome_link_policy.go`
   scan no claims; only the post-deploy audit does, which is detection not a gate. **The footer
   is where an "authorised and regulated… FRN" line traditionally goes.**
3. **UNPROVEN at the artefact.** No refusal has been observed in production, and with both sites
   locked there has been no demand. **A zero here is no-demand, not a working guard.** The first
   real build after the halt lifts is the test.

**Council: rounds 1 and 2 both REVISE; round 3 submitted 2026-08-20** (trail correlation
`aac38d5b-61cd-4cd7-b1e1-90d41c67a96f`). Both completed rounds found REAL defects, which is the
argument for submitting rather than defending:
- **round 1 (guardian, high):** relaxing `ParseEvidenceBase`'s nil-return could arm a documented
  landmine — parsing `evidence_base` through the typed struct and writing it back destroys
  citations and writer blocks. All five callers are read-only, but a caller list goes stale, so a
  test now **pins the loss** instead of claiming absence.
- **round 1 (edit-quality):** my calibration implied the negation guard handled negated forms. It
  does not — it is **never reached**. `ScanBannedClaimsIgnoringNegation` returns zero on all five,
  because the patterns need adjacency and an interposed "not" breaks the match first. Safer, but
  a maintenance hazard now pinned: loosen a pattern and it starts depending on a backward scan
  with a documented defect.
- **round 2 (edit-quality, high):** the claim that the scan is "the single function every
  enforcement surface calls" was **false** — the section editor ran no claims guard at all. Fixed
  in code (§3 item 1), not argued.

⚠ **Correction to this file's own first draft:** it said round 3 was already submitted. It was
not — only two runs existed. Checked and submitted properly. *Claiming a submission is the
cheapest possible false claim to make and the easiest to check: count the runs for the
correlation.*

## 3b. ✅ THE BRIEF WRITER IS BUILT AND LIVE — register **BLD-024**, migration `510`

`brief-writer`, 8 steps: research the subject from the domain name, write a comprehensive
`mission_brief`, then **hold**. Fired by `scripts/fire-brief-writer.sh <domain> [direction]`.

- **It writes the SAME `mission_brief` aspect a hand-written brief uses**, so nothing downstream
  changes — the classifier already reads `site_specs` wholesale.
- **Aspirational by instruction**, per the owner's ruling. Each `content_plan` entry carries
  `priority: core|valuable|aspirational` so the planner's degradation is legible.
- **The hold was PROVEN in both directions** against the dispatcher's own predicate before it was
  built on: `approval_mode='manual'`+`triaged` is withheld, flipping to `approved` releases
  (rolled back, nothing persisted). ⚠ `create_work_item` has **no `approval_mode` key**, so the
  agent uses `status='needs_human_review'` — which the dispatcher also cannot pick up, needs zero
  code, and is the estate's existing HITL idiom. Adding `approval_mode` to `create_work_item` is
  the better long-term shape and is recorded as a residual.
- **Generate-then-EDIT, not approve/reject.** Because it is a normal `site_specs` write, the
  owner's edited brief supersedes the generated one **with the generated one preserved
  underneath** — the diff between what the machine proposed and what he changed survives free.
- **Nothing dispatches it automatically.** `domain-submitter` is untouched, so a hand-written
  mission still wins and no existing flow changed.

**Deliberately not done:** it does not read the register yet (next piece, and the addendum argues
it should REPLACE RFC_037's classifier input); it does not screen third-party briefs (a different
job — this WRITES, that VETS).

**First run:** `indoorplanters.co.uk` (a **locked** test site, so nothing could build even by
accident), orchestration `1ea45228-2135-4754-b3d9-d3d792b87df0`. ⚠ **The thing to check is the
OUTPUT quality** — is the brief specific to the subject, or generic with the nouns swapped? That
is the failure mode the prompt argues against and the only thing that decides whether this is
useful. **A completed run is not a good brief.**

## 4. ✅ THE DOMAIN ESTATE — inventoried and classified

`domain-list-2026-07-30.csv` (owner-supplied, slightly stale): **1,567 domains, all registered,
none suspended.**

| class | n | what it means |
|---|---|---|
| PARKED | 1,247 | marketplace/for-sale DNS (dan.com 998, domain.io 193, namepros 25, …) |
| NO_DELEGATION | 207 | registered, nameservers never set — the owner's own explanation |
| **CLOOK** | **62** | real hosting (`dns.uk-noc.com`/`us-noc.com`) — **these are the live sites** |
| REGISTRAR_DEFAULT | 19 | unconfigured at a registrar |
| OTHER | 18 | genuinely unknown (2 AWS, 1 Hetzner, 8 domainmanage) — **not** asserted as parked |
| CLOUDFLARE | 14 | ours, the built sites |

Probed: the CLOOK ones serve substantial content (`cartoon.co.uk` 133 kB,
`businessinsurancequotation.co.uk` 78 kB, `aiartgallery.uk` 74 kB).

**People's-name domains: 5** — `katherineappleby.co.uk`, `oliverappleby.co.uk`,
`williamappleby.co.uk`, `williama.co.uk`, `billz.uk`.

**50 test domains** — `RESERVED_test_domains.md`. ⚠ **SIMPLIFIED by the owner 2026-08-20: they
need NO special status** — *"they are for us testing here… just keep them out of the list to build
for now."* So there is no `reserved_test` state to build and no register ceremony; they are an
exclusion list for the build queue, nothing more. **Plus one extra for immediate use in another
lane: `garden-tools.uk`** (parked, not in the register, not among the 50).

**The 62 hosted domains are listed for the owner in `HOSTED_domains_for_owner_decision.md`**, in
three decision groups: **25 serving real sites** (`cartoon.co.uk` 133 kB — anything here is a
migration, not a build), **13 empty stubs** (several are mail infrastructure — do not assume
free), **24 on hosting but serving nothing** (the readiest build candidates; note seven
`healthinsurance*` names are already register entries, and `interestrates.co.uk` is finance and
would inherit the compliance machinery).

⚠ **`billz.uk` is NOT a person's name** — owner: it is about *bills*. The extractor's
forename+initial rule is its loosest, so it now carries a small not-a-name exception list;
`alexj.uk` still reads as a person.

Tools (all take any domain list; the classifier prefers a registry CSV to live DNS, because the
export answers for domains that were never delegated where a live lookup returns an ambiguous
nothing): `scripts/domains/classify_nameservers.py --ns-csv <export>` ·
`extract_person_name_domains.py` · `pick_test_domains.py --names-tsv …`

**Nominet EPP works** (`LOGIN_CODE=1000`, owner-run 2026-08-19). The twelve-month expiry walk is
still un-run and would refresh this export; the CSV route needs no credentials and gives a
checkable total. Recipe + traps: `RUNBOOK_domain_inventory_and_classification.md`.

## 4b. ✅ 22 HOSTED SITES CLEARED FOR REMAKE (owner 2026-08-20)

Full decision: `DECISION_2026-08-20_remake_the_hosted_sites.md`.

- **PROTECTED (3): `leopardess.co.uk`, `leopardess.uk`, `cartoon.co.uk`.** The leopardess pair is
  an owner ruling and an active lane here; `cartoon.co.uk` was added the same day — *"off limits
  for now I have ideas for that one."* **Do not pre-empt that.** It is also the biggest single page
  in the batch (133 kB homepage on 4 sitemap pages), so what is there is substantial and
  concentrated on the front page.
- **22 free to overwrite.** All small: biggest 22 pages, median 4, **nine single-pagers**. Page
  counts came from following each WordPress sitemap INDEX down to the real URL lists — reading the
  index itself as a page count was my first and wrong answer (it lists 4 sub-sitemaps).
- **`vinrose.uk` images SALVAGED** to `portfolio_positioning/salvage/vinrose.uk/` before anything
  overwrites them — a 1920×1080 hero + a background, both verified as real JPEG data rather than
  200-with-an-error-page. ⚠ Reuse is conditional on the domain still going to wine, which is
  undecided; it has no register entry.
- ⚠ **Three twin pairs need a per-pair decision before either half is rebuilt** —
  `businesschristmasgifts` .co.uk/.uk and `fatherchristmas.uk`/`santaclaus.uk`. Two rebuilt sites
  on one proposition compete with each other, which is what the register exists to prevent.
  `fatherchristmas`/`santaclaus` may deserve two propositions rather than a redirect: unlike a
  spelling variant they are two genuinely different search phrases.
- **`businessinsurancequotation.co.uk` — the owner's own technique is recorded and is worth more
  than the site:** build content from REAL PAST CLAIMS with interesting stories. A past claim is a
  narrative fact, not a financial promotion — it advises nobody and promotes nothing, so it earns
  interest in a regulated vertical without regulatory exposure. It maps cleanly onto the existing
  claims layer (each story an `evidence_base` fact with a citation) and would suit
  `interestrates.co.uk`, the mortgage family and the health cluster equally. Owner is neutral on
  using it here.
- **Sequencing: do NOT start with the biggest.** `businessinsurancequotation.co.uk` is insurance,
  so it inherits the whole compliance layer including a regulated-identity guard that **has not
  yet refused anything in production**. Start with the single-pagers that have strong generic
  names and no regulated angle (`advertise.co.uk`, `conferences.co.uk`, `catalogues.co.uk`,
  `copyonline.co.uk`, `fridge-magnets.co.uk`).

## 3c. ✅ BOTH BUILT 2026-08-21 — sitemap mechanism + the register-reading brief writer

**(a) `render_sitemap` action** — register **SEO-002** (its own verify-later, now answered).
Emits `sitemap.xml` from the pages table. **ON by default, opt OUT with
`deploy_config.sitemap.enabled=false`** — inverted from `render_rss_feed` deliberately, because
an opt-in sitemap would reproduce exactly the situation the ruling exists to end. Probes every
URL and lists only 2xx; a drop is logged as "not fetchable now" and never fails the action;
refuses to publish an EMPTY sitemap because that misinforms a crawler rather than merely being
useless. Registered in the same change. **Council submitted: `8a004aab-be85-4d6d-bdb1-4fb114f1d64b`.**
⚠ Go, so inert until a roll, **and wired into no workflow yet** — that is the next step.
⚠ **My own test first failed to catch its own mutation**: the source-scan found
`noindex IS NOT TRUE` in the file's HEADER COMMENT, so deleting it from the SQL still passed.
Now extracts the query literal first, and all four visibility clauses are individually
mutation-proved.

**(b) The register is now a DATABASE table** — migrations `511`/`512`/`513`, register **BLD-025**,
**189 rows loaded**, and `brief-writer` gained a `read_register` step. **PROVEN**: it returned
M9's entry for `buytoletcalculator.uk` with `raw_md` and a siblings array carrying M1's
proposition.

⚠ **`raw_md` IS AUTHORITATIVE and the typed columns are only an index.** 49 entries use 18
different field names; `owns:` is labelled in exactly one of them. **A lossy parse of a
hand-written register would be a deletion, not a migration**, and "the DB is the source of truth"
is only safe if nothing is lost getting there — the discipline `EvidenceBase` failed at.

⚠ **`REGISTER_positioning.md`'s fate is UNDECIDED** — generated view, one-time input, or retired.
**Until that is settled, do not edit both**: two hand-maintained copies of one roster is the
`099_SYNC_gate_roster.py` drift class.

**This changes RFC_037's answer.** The reader is in the BRIEF WRITER, not the classifier, because
the owner reads every brief and can correct positioning that lands there — positioning fed to the
classifier is invisible to him. RFC_037 stays open as the home for the **binding** collision
check, which at 1,500 briefs is the only sampling rule that scales.

**Attribution split** (512): `field` 60 · `prose` 78 · `exclusion-only` 51. The labelled fields
alone left **82 of 152 portfolio domains with no row**; sweeping prose recovered them, and the
column records the weaker confidence rather than hiding it. **21 still have no row** — all `.uk`
siblings of registered `.co.uk` names the document itself never names: a register gap, not a
parser gap.

## 4c. ⚠ ALL FUTURE SITES MUST HAVE SITEMAPS (owner 2026-08-20) — 8 of 25 have one

Recorded against register **SEO-002**, whose own `verify-later` asked precisely this and is now
answered — and **BUILT, see §3c(a).** Measured 2026-08-20/21 by fetching every live site: only
**8 of 25 serve a sitemap of OURS** — including nothing on `remortgagecalculator.uk`, built four
days ago with every current guard applied, which is the clearest available statement that a manual
step is not a mechanism.

⚠ **CORRECTION to my own earlier figure of "10 of 25".** A 200 on `/sitemap.xml` is not evidence
the site has YOUR sitemap: `adversecreditmortgage.co.uk` returned 200 carrying a single `<loc>`
for `/lander` — the **parking provider's** file, still served from the old infrastructure, with no
matching `pages` row and `/lander` itself 307ing. Judge the body, not the status code.

**The work:** make it a standing mechanism, shape already proposed in SEO-002 — a Go action beside
`render_rss_feed` (read DB rows, emit a file artefact, gate on `deploy_config`). Two rules to carry
over rather than rediscover: **probe before listing** (a sitemap advertising a 404 is worse than
none) and **read every column that decides whether a page should be found** — `noindex` arrived
after the script and silently contradicted it for weeks.

## 5. 🧱 STILL THE WALL — `bugs_open/311` + `RFC_036`

Unchanged and **still not fixed**. Owner ruling: **one submission covering both writers, and it
is a PRECONDITION for wave 1.** New this round: the planner's `load_components` query includes
tool-level components **only if the site already has `plan_includes_tools` AND the tool is
already on one of its own pages** — so on a greenfield build the planner can see no library tool
at all. That is 311's upstream half, and it corrects my earlier claim that the planner does not
know what can be built: it does, for sections and elements, every run.

## 6. NOT STARTED — domains for third-party customers

Search + buy, on domains that are not the owner's. **Nothing in this estate currently spends
money**; whose account holds the domain is a commercial decision before a technical one. What
exists to build on: the nameserver classifier, the Cloudflare zone+route recipe, Nominet EPP.
What does not: any availability search, any purchase path, any registrar key beyond Nominet.

## 7. TRAPS — each cost a cycle today

- **A binary carries ONE commit stamp.** Grepping it for YOUR sha always fails. Use
  `git merge-base --is-ancestor`. The `build provenance` log line scrolls within hours.
- **One client's view of a CDN is not the site's state.** Repeating a fetch from one machine
  measures consistency, not correctness. Cross-check the body against `Last-Modified`: a fresh
  header on a stale body means the delivery path, not the page. (`WRONG_CALLS.md` 08-19.)
- **A stale `/tmp` file reads as a fresh result.** `curl` that fails leaves the previous body in
  place; `rm` first, and treat `http=000` as "no body", never as the body you just printed.
- **Precedence bugs look like data bugs.** The name extractor was wrong in BOTH directions before
  it was right — forename-first gave 35 names of which 3 were people (`christmasbasket` =
  "chris"+"tmasbasket"); compound-first killed `jamesbrown` because `brown` is a dictionary word.
- **"The first N that pass" is not a sample.** The picker returned 50 a/b domains because the
  list was sorted. Stride instead.
- A parked domain returns 200 on every path — read the BODY.
- Discovery queries must stay **< 200 bytes** or `web_search` drops them and blames config keys.
- A migration NUMBER identifies nothing — two different `471`s exist. Ask the ledger by filename.

## 8. Files of record

**Cold start:** this file → `SUMMARY_2026-08-19_…` →
`PLAN_2026-08-19_one_flow_three_brief_sources.md` → `README_where_we_are.md` (owner's log) →
`NOTES_portfolio_positioning.md` (evidence, newest at bottom).
**Decisions:** `RFC_037` (all four questions ruled) · `REGISTER_positioning.md` (P11 + the
registry rulings) · `DECISION_2026-08-18_two_builder_flows_side_by_side.md`.
**The wall:** `bugs_open/311` · `architecture_review/RFC_036`.
**Regulated guard:** `platform/orchestration/datahelpers/claims_regulated.go` (+ test) ·
`platform/orchestration/actions/section_editor_regulated_guard.go` · `cmd/regcheck` ·
`scripts/regulated/record_attestation.py` · register **CGV-033**.
**Domains:** `RUNBOOK_domain_inventory_and_classification.md` · `RESERVED_test_domains.md` ·
`scripts/domains/`.
