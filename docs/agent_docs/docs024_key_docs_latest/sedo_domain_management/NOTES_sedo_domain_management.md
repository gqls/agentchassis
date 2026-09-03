# NOTES — Sedo domain management (append-only, newest at the bottom)

## 2026-09-02 — lane opened (owner request: "set up sedo … so that you can manage my domains")

**Research** (web, sources in PLAN):
- API endpoint `https://api.sedo.com/api/v1/<Function>`, GET or POST,
  XML out; SOAP/WSDL also offered, not needed.
- Auth params on every call: `partnerid` (int), `signkey`, `username`
  (≤25 ch), `password` (≤16 ch — the API doc's own cap).
- Access: email api@sedo.com from the registered address; partner-programme
  registration required first; approval issues Partner ID + SignKey.
- Functions: DomainList/-Extended, DomainStatus, DomainInsert/Edit/Delete,
  DomainSearch, portfolio CRUD, parking setup/keyword/template, blacklist
  check. `results` ≤100/call, `startfrom` pages. currency 0=EUR 1=USD 2=GBP.
- Prior art: NO existing Sedo integration anywhere in the estate — repo
  grep hits were archive.org JS noise, the parking-200s landmine, and a
  stats-export aside in the traffic-probe brief.

**Measured today** (each could have come out otherwise):
- `busybox:1.36` ephemeral pod, in-cluster: pulled, resolved, TLSed to
  api.sedo.com, got `<SEDOFAULT> E7 "Partnerid doesn't exist"` on dummy
  creds. So: public image pull works, egress works, API answers, and
  **faults are in-band XML on a readable body** (the BusyBox
  4xx-body-drop limitation doesn't bite here).
- Same probe printed `wget: note: TLS certificate validation not
  implemented` → BusyBox wget REJECTED as transport for credentialed calls.
- `curlimages/curl:8.10.1` pod: pulled, POST with `--fail-with-body`, same
  E7. This is the pinned transport.
- `kubectl run -i` printed its own warning that commands+output land in
  container logs, and its attach-failure fallback double-printed the
  response → runner design changed to run → wait → logs → delete, creds
  via envFrom only.

**Missteps:**
- First cut of `PARAM_RE` used `\[\]` inside an ERE bracket expression —
  backslash does NOT escape there, so the first `]` closed the class and
  both positive self-tests failed. Fix: `[][A-Za-z0-9_.-]` (the
  `]`-first idiom). Caught by the script's own `--self-test` on first run —
  which is the argument for writing the self-test before the first use.

**Built:** `scripts/domains/sedo-api.sh` (self-test PASS 7/7; probe PASS; secret
missing → clear pointer to RUNBOOK §3). Credentialed path UNEXERCISED —
blocked on owner obtaining the account + SignKey (RUNBOOK §1–§3).
Registered as OPP-012.

## 2026-09-02 (later) — concurrent-lane discovery, script relocated

The first commit's scope report ran while HEAD had moved: 62 seconds
earlier the **domains_cloudflare_rollout** lane committed
`scripts/domains/porkbun.py` — that lane (open since 08-03, active now)
already holds registrar helpers (porkbun/dynadot/epp/spaceship plans) under
`scripts/domains/`. My prior-art grep for "sedo" could not have found it
(different keyword, different service class) — the commit stream found it,
which is the shared-tree working as designed. Moved `sedo-api.sh` →
`scripts/domains/` (git mv; both paths named on the commit per the
LANDMINES entry; verified single copy at HEAD), re-pointed every reference,
added the missing OPP-012 index row the pattern check caught, and wrote the
division of labour into PLAN "Cross-lane constraint".

## 2026-09-02 (evening) — owner supplied the bulk-listing format; DomainInsert mapped

Secret still absent (`--check-secret`, 17:31 BST) — P2 remains with the
owner; self-test still 7/7 at HEAD.

Owner provided Sedo's Domain Importer template
(`~/Downloads/Example_File_Domain_Importer.xlsx`). No xlsx reader installed
(openpyxl/pandas both absent) — unzipped it and parsed the sheet XML +
sharedStrings directly. Seven columns, eleven example rows, no embedded
`dataValidation` elements, authored 2021 in Excel. Decoded content and
column semantics now in RUNBOOK §6.

Fetched the DomainInsert function doc
(`apidocs/v1/Basic/functions/sedoapi_DomainInsert.html` — NB the guessed
URL `Basic/DomainInsert.html` 404s; the index page gives the real paths):
`domainentry` array ≤50/request; per-entry `domain`/`forsale`/`price`/
`minprice`/`fixedprice`/`currency` all required, plus `domainlanguage`
(NOT in the sheet — the web importer defaults it, API callers must pass
it); `category` optional ≤3. Two behaviours worth their lines: insert is
ASYNC (post-hoc checks, failures arrive by email), and EVERY insert
auto-enables parking regardless of `forsale` — recorded in RUNBOOK §6
because it interacts with the cross-lane parking constraint.

Wire shape for the array param is `[INFERRED]` from the doc's own
`http_build_query` example (the literal nested-key URL is never printed);
the planned first one-domain call doubles as its confirmation.

## 2026-09-02 (late evening) — owner: account EXISTS; portfolio sheet built (1,318 domains)

Owner message: has **partnership status** and an account under
**info@designconsultancy.co.uk**; wants most domains added; asked this lane
to (a) build the full-portfolio sheet in Sedo's importer format, (b) ask
the dynadot/porkbun/nominet/spaceship sessions for the domain lists, (c)
correspond with the new **domain valuation** session on Sedo+Afternic
pricing. All six are live peer sessions; the valuation lane had already set
the inbound contract, so this lane consumes their `inbound/` rather than
running a second collection.

**Cross-session outcomes (all same evening):**
- dynadot: 451 complete per API contract (one unpaginated response;
  panel-count cross-check pending); **Dynappraisal fetch running** →
  `dynadot_valuations_2026-09-02.csv`, rate-capped/day (50–300 by tier).
- porkbun: 683 delivered (`02ffa6f40`). **CORRECTION to my framing:** the
  global opt-in gates PER-DOMAIN endpoints (future NS ops), NOT the list —
  I had repeated the valuation lane's "delivery blocked on opt-in" line;
  wrong, and the correct fact is in their message. 13/683 NS cells blank
  (DoH SERVFAIL); 600/683 on afternic-family NS; 0 on Porkbun's own
  marketplace (43,203-listing intersection).
- spaceship: 203 confirmed complete (fetched == API `total`). ⚠ their
  seller-hub listings file is NOT an inventory (36/831 registered there).
- nominet: contract confirmed — one CSV (`domain,expiry_month`) will serve
  valuation + sedo; still owner-gated on the login/walk.
- domain valuation: contract agreed — canonical
  `OUTPUT_prices_<date>.csv` in THEIR lane (columns frozen by them, ~10
  fields incl. category/keep_or_sell/confidence); interim blank-price sheet
  approved with two owner-facing caveats (lowball risk without minimums;
  auto-parking disclosure). Cut confirmed: bottom ~500 BUY_NOW keen,
  categories stay together; £150 webdesign.co.uk transfer-away fee
  (2026-08-17) as de-facto keen floor.

**Built:** `scripts/domains/sedo-importer-xlsx.py` (self-test 9/9; emulates
the template's sharedStrings encoding; refuses non-ACE domains, unknown
selling options/currencies, malformed prices; maps prices by header name;
verifies its own artefact by re-reading the zip). Generated
`outbound/SEDO_IMPORT_2026-09-02_draft1.{xlsx,csv}` + provenance CSV +
`EXCLUDED_live_cloudflare_2026-09-02.txt`:
- **1,337 unique domains** in hand (451+683+203, 0 cross-registrar dupes —
  the earlier `uniq -d` "duplicate" was the two quoted header rows my
  `NR>1` failed to skip; `FNR>1` is the correct guard);
- **19 fenced out** (Cloudflare NS = live estate sites, incl.
  websitedesign.com, boxingonline.com, relojistas.com, vetcomparison.uk);
- **1,318 in the sheet**, all MAKE_OFFER / yes / no price;
- NS provenance: 1,193 afternic-family, 111 other, 13 none, **1 already on
  sedoparking.com** (`officestationery.net` — corroborates the owner's
  pre-existing Sedo relationship).
- Verified at the artefact: 1,319 `<row` occurrences (header + 1,318);
  sharedStrings uniqueCount 1,327 = 1,318 domains + 7 headers + 2 shared
  literals. (`grep -c '<row '` reads 1 — single-line XML; count with
  `grep -o | wc -l`.)

Outstanding: Nominet ~1,500 .uk (owner walk) → draft2; prices import once
the valuation lane freezes/ships `OUTPUT_prices`; owner still owes the §2
API-access email + §3 secret for the API route.

**Reconciliation closed (valuation lane, same evening):** their ingest
matches 1,337/0-dupes exactly; the 19-domain fence is adopted into their
model as `keep_or_sell=KEEP` overrides regardless of rank (relojistas.com
additionally carries the owner-confirmed $12k Afternic floor, 2026-07-28).
They message this session when OUTPUT columns freeze — nothing owed from
this lane until then.

**Dynadot completeness CLOSED + draft1 refreshed (same evening, later):**
owner panel cross-check + fresh re-pull confirm `list_domain` complete;
451 → **453** purely by addition (overhead-cranes.com, paper-cups.com,
added by the owner today; zero drops). Their inbound CSV refreshed in
place (`574a04290`); draft1 regenerated from it — **1,320** rows, CSV diff
exactly the two insertions, fence unchanged at 19. Standing practice per
their note: the owner adds domains on occasion, so **re-pull all registrar
CSVs immediately before any sheet the owner will actually upload**.

**Dynadot follow-up (same evening):** appraisals landed for **300 of 451**
(daily quota fired at exactly 300; remainder resumes tomorrow,
idempotent). And a correction to "0 listings": `isForSale` is unreliable
(their landmine) — a marketplace-dump control found **5 live Dynadot Buy
Nows** (see RUNBOOK §7's new priced-import gotcha for names/prices). All 5
verified present in draft1 as make-offer/no-price — no displayed conflict
today; the price-agreement constraint recorded for every priced draft and
relayed to the valuation lane. 394/451 Dynadot domains on Afternic NS —
existing Afternic asks must join the same sweep once the owner's export
lands.

## 2026-09-03 — Nominet delivered (1,606 .uk); fence WIDENED from 19 to 50; draft2 = 2,895 domains

Nominet CSV: `nominet_domains_2026-09-03.csv` (`f8ca8389d`), 1,606 domains,
`domain,expiry_month`. Their own caveat: the walk had **never succeeded
before tonight** — 3 EPP-client bugs fixed same evening, verified against
the registry's own count — so treat 1,606 as first-proven, not
long-established.

Their message also raised the exact gap this lane's own PLAN had flagged:
the raw walk includes the ~40 domains they cut over to Cloudflare, and
"cross-reference against that list before building the sheet, not after."

**Investigated properly rather than trusting one source:**
- Asked Nominet for their live-cutover list. Their first answer ("~40")
  was already stale by the time it arrived — a live re-query mid-reply
  found 4 more zones cut over in the interim (advertise/designblog/
  seotools/websitepromotion.co.uk) → **46** zones, corrected same message.
- Independently queried `clients_db.sites` (schema read first,
  `\d sites`): `status='deployed'` = **39** rows — corroborates their
  original "~40" almost exactly, good cross-check. Two more rows at
  `status='test'` are real domains (buytoletcalculator.uk,
  indoorplanters.co.uk) — included; the `pool-*`/`system.internal` rows are
  placeholder slots, not real domains, excluded.
- **Nominet's own framing settled the method**: fence a domain if it is in
  EITHER source, not just one — a Cloudflare zone can exist with no site
  row (apis.uk, ugg2.com: zone + worker route, no site behind it) and a
  site row can predate its zone (mid-cutover). Neither list alone is
  sufficient; this is now RUNBOOK §7's documented method.
- Union of {46 zones, 41 sites-table domains, the old 19-item NS fence} =
  **50** unique domains (`EXCLUDED_live_2026-09-03.txt`, superseding the
  09-02 file). 31 were invisible to the NS-based method alone — most
  visibly `adversecreditmortgage.co.uk`: `status='deployed'` in the DB yet
  sitting on MARKETPLACE nameservers (a known state, improvement-loop D2)
  — exactly the kind of case that proves NS-only fencing is not enough,
  since its own delegation reads as "not live."

**Draft2 built and verified:** all four registrar/registry sources (dynadot
451 + porkbun 683 + spaceship 203 + nominet 1,606 = 2,945 raw, **0 dupes**
confirmed by direct grep with the `FNR>1` guard — plain `NR>1` across
multiple awk input files undercounts by treating only the first file's
header as a header, the same misstep as 09-02) minus the 50-domain fence =
**2,895** domains. Verified at the artefact: sheet XML carries 2,896
`<row` occurrences (header + 2,895, `grep -o | wc -l` not `grep -c`); all
four spot-checked fenced domains (idea.uk, webdesign.uk, relojistas.com,
wykefarm.uk) confirmed ABSENT from the output CSV by direct grep.
Files: `outbound/SEDO_IMPORT_2026-09-03_draft2.{xlsx,csv}` +
`_provenance.csv` + `EXCLUDED_live_2026-09-03.txt`.

Outstanding: prices import once the valuation lane freezes/ships
`OUTPUT_prices`; owner still owes the §2 API-access email + §3 secret for
the API route; dynadot's remaining 151 appraisals resume on their own
schedule.

## 2026-09-03 — owner sanity-checked draft2, asked to withdraw the Appleby family; draft3 = 2,888

Owner asked two questions and gave one instruction: (1) confirm no BUY_NOW
prices in the sheet, (2) what minimum-offer amounts were set, (3) "take
out all the appleby domains" before uploading.

**Verified directly against the artefact** (never trust the intent, check
the file): `awk` over every one of the 2,895 rows in draft2 — Selling
Option column has exactly one distinct value, `MAKE_OFFER`; Price,
Minimum Price and Currency columns are empty on all 2,895 rows, zero
exceptions. So: no BUY_NOW rows, and the honest answer to "what minimums"
is **none set** — every domain currently reads as an open make-offer with
no floor, which is the interim shape agreed with the valuation lane
pending their `OUTPUT_prices` (RUNBOOK §7 priced-import section).

**Appleby**: grepped every inbound CSV (not just the sheet) for the
substring — **7** domains actually held: anthonyappleby.com, appleby.cv,
katherineappleby.co.uk, kathyappleby.co.uk, kathyappleby.com,
oliverappleby.co.uk, williamappleby.co.uk. Three more names
(katherineappleby.com, oliverappleby.com, williamappleby.com) appear in
the valuation lane's new `appraisal_queue_proxy_2026-09-04.csv` as
comparables paired against the .co.uk domains, not domains this estate
holds — checked directly against all four registrar/registry CSVs,
confirmed not held, correctly left untouched.

**Misstep, caught before committing**: first instinct was to append the 7
names straight into `EXCLUDED_live_2026-09-03.txt` — wrong, because that
file's own name states a specific reason (live-site protection) and gets
**fully regenerated** from its two sources (Nominet zones + `sites` table)
on every future cut; anything hand-added there would be silently dropped
the next time it's rebuilt from source. Reverted (`git checkout --`)
before it was ever committed. Correct fix, done instead: widened
`sedo-importer-xlsx.py`'s `--exclude-file` to `action="append"` (unions
multiple files; self-test now 10/10, the new check calls the actual
reader function rather than re-implementing the union inline — a test
that reimplements the logic under test proves nothing), then a SEPARATE
file `EXCLUDED_owner_appleby_2026-09-03.txt` (7 names) for this specific,
differently-reasoned request. RUNBOOK §7 now states the rule: one fence
file per REASON, never merge.

**Draft3 built and verified**: 2,895 − 7 = **2,888** domains, confirmed
three ways — the tool's own printed exclusion list (57 = 50 + 7), an
independent re-read of the xlsx (2,889 `<row` = header + 2,888), and a
direct grep of the output CSV for "appleby" (zero hits). Files:
`outbound/SEDO_IMPORT_2026-09-03_draft3.{xlsx,csv}` + `_provenance.csv` +
`EXCLUDED_owner_appleby_2026-09-03.txt` (new); `EXCLUDED_live_2026-09-03.txt`
unchanged and reused as-is.

Owner said he'll upload what's built now — draft3 is current. Prices
remain the outstanding item; nothing else changed this round.

## 2026-09-03 (later) — williama.co.uk + wyke/pastured-egg withdrawn; draft5 = 2,879

**Cross-lane input first**: valuation lane confirmed draft3's 7-name
Appleby withdrawal applied and independently reconciled (same 7, no
more/fewer), stored it as a DISTINCT `keep_override=owner-withdrawn`
value rather than folding it into the live-site KEEP state (their
reasoning: a live-site keep can be revisited if the site comes down, an
owner withdrawal cannot — collapsing the two loses that). They also flagged
a real bug their own message caught, unprompted: 6 domains the estate does
NOT actually own (3 expired, 3 registered elsewhere) had been counted as
sellable stock and priced — now excluded ahead of every other rule on
their side. And they raised a genuine open question rather than guessing:
**williama.co.uk** reads plausibly as the same Appleby family (the estate
also holds williamappleby.co.uk, oliverappleby.co.uk) but was NOT in the
withdrawal set; four more person-name domains with no obvious family link
(ianstirling.com, kapoor.uk, keeler.uk, anne-marie.co.uk) were also still
sitting in the sheet.

**Put to the owner directly** (AskUserQuestion — genuinely his call, not
one to infer): confirmed **williama.co.uk should be withdrawn** (same
Appleby reasoning). The four unrelated names got no answer — **left alone,
open**, not treated as "no" by default.

**Same message, owner also said**: "also remove anything like wyke or
wykefarm or pasteured egg." Grepped ALL FOUR sources fresh rather than
guessing at spelling variants — found **wyke-farm.co.uk / wyke-farm.uk**
(hyphenated, Nominet, NOT previously fenced — distinct from the unhyphenated
`wykefarm.co.uk`/`wykefarm.uk`, which were already out via the live-site
fence, so no duplicate action needed there) and **six** pastured-egg names
(Porkbun: pasturedegg.co.uk, pasturedegg.uk, thepasturedegg.com,
thepasturedeggcompany.{co.uk,com,uk}) — none previously fenced (they sat
on Porkbun's own NS, "OTHER" provenance class, not Cloudflare, so the
live-site check never touched them).

**New reason file** (per RUNBOOK §7's one-file-per-reason rule — this is a
brand/farm family, unrelated to Appleby, so a THIRD file, not appended to
the second):
`EXCLUDED_owner_wykefarm_pasturedegg_2026-09-03.txt` (8 names — the 2
non-hyphenated wykefarm domains omitted since already covered).
`williama.co.uk` went into the EXISTING Appleby file (same
reason/category as that withdrawal, unlike wyke/pastured which is a
different one).

**Draft5 built and verified**: three exclude files now unioned (live=50,
appleby=8, wykefarm/pasturedegg=8) = 66 total minus 2 overlap (wykefarm
already-live) = distinct exclusion count 66 as printed. 2,887 (draft4,
intermediate, not separately shipped) − 8 = **2,879**. Verified at the
artefact: 2,880 `<row` = header + 2,879; grep for wyke/pastur/appleby
across the output CSV returns zero; zero non-MAKE_OFFER or priced rows.
Files: `outbound/SEDO_IMPORT_2026-09-03_draft5.{xlsx,csv}` +
`_provenance.csv` + the new wykefarm/pasturedegg file; draft4 kept on disk
as the intermediate step (williama-only) but draft5 supersedes it.

Notified valuation lane of both new withdrawals for their `keep_override`
model. Prices remain the only outstanding item.

## 2026-09-03 (later still) — cross-lane reconciliation: "all domains will be in afternic", a real landmine, an open check

**copy_quality_two_stage relayed** an owner remark from an UNRELATED
conversation about about-page copy: "All domains will be in afternic" —
raised because it sits on top of `about_page_commercial`'s D1
(2026-07-24), which deferred the multi-route (Sedo/Flippa/direct)
question rather than deciding it. They deliberately did not act on it or
interpret it, and asked this lane + domain_valuation to reconcile.

**Resolved with first-hand evidence from this lane**: the message that
opened THIS thread today said, verbatim, "correspond with them to
establish the right price for listing on sedo AND afternic" — specific,
recent, and about pricing consistency across both, not Afternic-exclusive
routing. Read as: every domain gets an Afternic presence (and the
about-page's enquiry link), which is compatible with ALSO cross-listing
on Sedo — the two questions (which marketplaces list the domain vs. what
the on-site CTA points to) are orthogonal. copy_quality_two_stage agreed;
D1's multi-route deferral stays parked, untouched by this. Not treating
this exchange as a formal ruling — it's a reconciliation between two
sessions' partial evidence, not an owner decision — but no contradiction
survived it.

**A real, measured bug came out of this same thread** (domain_valuation,
reconciling their wyke-farm withdrawal against my fence files): their
joiner globbed `EXCLUDED_live_cloudflare_*`, which matched my OLD
2026-09-02 fence name and silently missed the current
`EXCLUDED_live_2026-09-03.txt` sitting beside it — so their model read a
19-domain fence, not 50, and counted **31 live sites as sellable stock for
most of 2026-09-03** (webdesign.co.uk, idea.uk, mortgagecalculator.co.uk,
loancalculator.co.uk, 27 more). My rename (dropping "cloudflare" from the
name when the method widened beyond NS-only) was reasonable on its own but
broke a downstream consumer I had no way to see. **Filed as a landmine**
(LANDMINES.md, "A downstream consumer that globs a fence file by its FULL
historical name…") — the check: consumers glob the stable stem
(`EXCLUDED_live*`) and UNION every match, never pick one file.
`landmines-verify-dispatch.sh` found nothing pending (another session's
concurrent sync had already consumed the "new" flag — the exact trap the
file itself warns about), so triggered verification directly:
`trigger-landmine-verifier.sh 'LANDMINES.md#a-downstream-consumer...'`,
correlation `a83710be-6f10-4d47-8021-7babe826dae8`. Also confirmed: their
own separate discovery that wykefarm.co.uk/.uk should ALSO be in their
owner-withdrawn set (not just the live-site fence), reasoning a withdrawal
outranks and outlives a fence that could lapse — sound; worth mirroring
into my own wykefarm/pasturedegg file next time it's touched (not done
yet, no output change since they're already excluded via the live fence).

**Open, unverified**: copy_quality_two_stage additionally reports **6
domains the estate does not own** (3 expired, 3 registered to others) are
"already listed at Afternic at $10k–$50k, undeliverable," and asked
whether my sheet shares the same 6. `[UNVERIFIED]` — asked
domain_valuation for the actual 6 names to check by name rather than rely
on the structural argument (my sheet is built only from the four raw
registrar/registry list_domain exports, never their derived
appraisal_queue files, so an unowned domain has no route in) — structural
soundness is not the same as checked. Resolve before the owner's next
upload if the names arrive in time; otherwise flag as still open.

**Resolved same message**: valuation lane checked draft5 directly by name
(all 6 → 0 rows) with real positive/negative controls (`aakn.com` → 1, an
invented domain → 0), not just the structural argument. Both halves now
hold — clean by construction AND clean by measurement, which matters
because construction alone would have survived a fence bug exactly like
today's. They also filed the producer-side landmine half from their end.

**MISSTEP, caught ~30 min later, cost real churn**: every timestamp in
this session from "Nominet delivered" onward — 11 generated files (three
draft sheets × 3 files, two exclusion lists) and several log headings —
was dated **2026-09-04**, which was wrong; the actual date all along was
**2026-09-03**. Traced the origin: I saw `appraisal_queue_proxy_2026-09-04.csv`
in the valuation lane's inbound directory while grepping for Appleby
names, and anchored on that filename as "today" instead of running `date`
myself. Caught only incidentally — checking registrar exports for
past-due expiry dates (prompted by copy_quality_two_stage's sharper
framing of the undeliverable-domains question) needed `date +%F`, which
printed 2026-09-03; cross-checked against fresh `git log` timestamps,
which agreed. **This had already reached a committed LANDMINES.md entry**
(`[MEASURED 2026-09-04]`) and a peer session had verified my sheet by its
exact wrongly-dated filename. Fixed forward: `git mv` all 11 files to the
correct date, corrected every doc heading/reference (one line preserved —
the valuation lane's own genuinely-09-04-named file, not mine to alter),
fixed the LANDMINES.md tag and source line, regenerated the renamed sheet
from the renamed exclude files and diffed it byte-identical against the
pre-rename version before trusting nothing else broke. Notified both
peers of the corrected filenames. Full account, including why none of
this workstream's own dating discipline caught it:
`WRONG_CALLS.md`, "I dated files, log entries, and a committed landmine
'2026-09-04'…". **Lesson for every future date-in-filename: run `date
+%F` yourself; a date glimpsed in someone else's filename is a fact about
their file, not the clock.**

**Root cause traced further, same evening**: the valuation lane's own
`appraisal_queue_*_2026-09-04.csv` was theirs, dated by the window it was
built FOR, while every other file in their inbound directory is dated by
when it was written — an inconsistent convention within one directory,
which is what made reading it as "today" a reasonable inference rather
than a careless one. They fixed it at source (renamed to `_2026-09-03`,
stated the convention explicitly in their own RUNBOOK — file dates are
production dates, never intended-use dates) and confirmed their earlier
verification of draft5 still stands unchanged (byte-identical rename,
zero re-run needed). No further action here; recorded for completeness.

## 2026-09-03 (urgent) — copyonline.co.uk withdrawn: owner told a peer it's a keeper, not stock; draft6 = 2,878

copy_quality_two_stage relayed a direct, time-sensitive owner statement:
"copyonline is my site but could become my wife's if she chooses to do
more copywriting" — a keeper with prospective personal use, must not
carry a for-sale listing anywhere. Acted immediately, before checking
whether their diagnosis of HOW it slipped through was right.

**Measured the actual cause, and it is NOT what they proposed.** They
reasoned the live-site fence's `status IN ('deployed','test')` filter
should have caught it (copyonline is `status=test`) and hypothesised a
definitional gap — "build_status is not ownership intent." Checked
directly: `sites.created_at` for copyonline.co.uk = **09:27:25 UTC**;
`EXCLUDED_live_2026-09-03.txt`'s first commit (`ea3b63b53`) = **08:57:16
UTC**. **The row did not exist when the fence was built — this is
staleness, not a filter gap.** My own RUNBOOK already says "regenerate
the whole union with each sheet"; the actual defect is that I built the
fence once and reused the SAME file across three later regenerations
(draft4, draft5, draft6-before-this-fix) spanning hours, rather than
re-querying each time. Confirmed by re-running the exact same query fresh
just now: it returned copyonline.co.uk PLUS the 41 originals, nothing
else new — so the query and its status filter were always correct; only
the cadence of re-running it was wrong.

Their broader point still stands independently, though: a domain that
has NEVER had a `sites` row created for it (the owner mentions intent,
nobody has typed a command yet) would defeat even a perfectly fresh
re-query, because there is no signal to find. That is a real, harder gap
their message correctly named, distinct from the staleness bug that
actually fired here.

**Fixed**: re-ran the live-site query fresh, diffed against the 09:03
snapshot (exactly the one new row, confirming nothing else was missed),
rebuilt `EXCLUDED_live_2026-09-03.txt` in place (50→51), regenerated as
draft6. Verified at the artefact: 2,879 `<row` = header + 2,878; zero
copyonline hits in the output CSV; diffed whole-file against draft5 —
exactly one row removed, nothing else changed. Files:
`outbound/SEDO_IMPORT_2026-09-03_draft6.{xlsx,csv}` + `_provenance.csv`.

**Process fix, not just this one domain**: regenerating from a REUSED
fence file across multiple same-day drafts is the actual defect class —
added to RUNBOOK §7 as a standing instruction to re-query fresh before
EVERY regenerate, not just the first one of a session.

**copy_quality_two_stage's diagnosis was wrong, and worth recording why
rather than just noting it was corrected**: they later logged this
themselves as a wrong call, with a sharp observation — the two possible
causes (a status-filter gap vs. staleness) "want opposite fixes." Acting
on their invented cause would have added a redundant intent-check while
leaving the real 30-minute staleness window wide open, and would have
LOOKED like a fix while leaving the actual hole live. A correct finding
(copyonline was the sole overlap) is not evidence for whatever mechanism
is asserted alongside it.

## 2026-09-03 (later) — owner relaxed copyonline; deliberate call made; D1's CTA is now RULED, exposing relojistas.com has no Sedo listing to point at

**copyonline.co.uk**: owner told copy_quality_two_stage either outcome is
fine — "we can still sell it, we'll just charge a lot" OR "leave it as
not for sale, that's fine too" — explicitly delegating the call, asking
only that it be deliberate, not left ambiguous. **Decision: keep it
withdrawn.** Reasoning: every other row in the sheet is priced by the
valuation lane's process, not hand-typed; "a lot" has no number attached,
and inventing one would be exactly the ad hoc pricing this lane has
avoided throughout. Withdrawn is also the owner's own stated no-new-
information-needed fallback. **Durability fix applied regardless of which
way the call went**: added `EXCLUDED_owner_copyonline_2026-09-03.txt` (1
domain) alongside the existing live-site fence entry, mirroring the
wykefarm.co.uk lesson — a live-site fence entry can lapse if the site's
`sites` row status changes; an owner withdrawal should not. Verified: with
the new file added, output is byte-identical to draft6 (copyonline was
already excluded via the live fence) — confirmed by full diff, zero
changes, no new draft number needed. If a real BUY_NOW figure is ever
supplied, re-adding it as a priced row is a one-line change.

**D1 is RULED**: owner's exact words, relayed by copy_quality_two_stage,
in direct answer to their flagging D1 as parked: **"Yes, point to Sedo."**
The about-page enquiry CTA destination is decided — Sedo, not Afternic —
independent of whatever marketplace-listing strategy is settled here.

**This exposes a real gap, not a documentation task**: copy_quality_two_stage
asked for the Sedo listing URL pattern for relojistas.com so the CTA can
point at it. Checked: **relojistas.com has NO Sedo listing at all** — it
sits in `EXCLUDED_live_2026-09-03.txt` and has been excluded from every
sheet built in this lane (`grep` confirms zero rows in draft6). There is
no URL because there is nothing to link to; composing one from a guessed
pattern would produce a dead link on a live, Spanish-language, already-
confirmed-for-sale page ($12k Afternic floor since 2026-07-28) — exactly
what they asked NOT to receive. Told them plainly rather than guess.
**No Sedo URL format is documented anywhere in this lane's own research
either** (checked RUNBOOK/API docs) — the correct address for a real
listing has to be captured FROM a real listing (API `DomainStatus`/
`DomainList` response, or the dashboard, once one exists), not composed
from a template.

**Real decision this surfaces, put to the owner**: relojistas.com is
already confirmed for-sale elsewhere (Afternic) and is excluded from
every Sedo sheet only because the fence protects ALL live sites
uniformly, not because there was doubt about ITS for-sale status
specifically. Now that the on-site link is set to point at Sedo, that
link has nothing real to point to until relojistas.com is actually
listed on Sedo — which cannot happen via the credentialed API yet (§2/§3
still owed) but COULD happen today via a small, deliberate, one-domain
web-import upload (exactly the "first writing call, one domain, not a
batch" principle already stated in PLAN P3). Asked the owner directly
rather than deciding unilaterally — adding a live site to any for-sale
listing is not a call this lane makes on its own, even a site already
confirmed for sale elsewhere.
