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

## 2026-09-03 (urgent) — a structural blind spot found: a WHOLE HOSTING STACK was invisible to the fence; 39 domains pulled, incl. the owner's own email + a client relationship

**domain_valuation's Nominet nameserver sweep found something bigger than
a missing name**: my live-site fence is derived entirely from the
framework's OWN sites/zone list (Cloudflare delegation + `sites` table).
The estate has a SECOND hosting stack — Clook (`dns*.uk-noc.com`), the
older sites that predate the Cloudflare rollout — and the fence is
structurally blind to it, not just incomplete. Their probe found **33
Clook-hosted domains serving real HTTP 200 content**, zero of them in any
fence file, all 33 currently sitting in draft6's no-price sheet. Six more
are Clook-delegated but currently serve nothing (ambiguous — ownership
intent unclear from an HTTP probe alone).

**This is not a "re-query more often" fix** — re-querying `sites`/zones
faster would never have caught this, because the source itself never
covered this stack. The fix has to be a second, independent source
(delegation sweep across BOTH hosting stacks), which domain_valuation has
offered to own and derive fresh each time — accepted (see reply).

**Two of the 33 are not "portfolio stock with a site" at all**:
`wpx.uk` carries the owner's own registered email address, and
`designconsultancy.co.uk` is the company behind the very Nominet tag this
whole estate is registered under — the operator's own infrastructure, not
inventory. Two more, `leopardess.co.uk`/`leopardess.uk`, sit directly
adjacent to `leopardessconsulting.co.uk` — already fenced specifically
because it is the client relationship copy_quality_two_stage named as the
worked example for "buy this site" being a relationship breach (D4). These
four did not go anywhere near the sheet without a second look — see the
question put to the owner.

**Immediate action, before any of that gets resolved**: fenced all 39 (33
confirmed-live + 6 ambiguous, leaning cautious on the ambiguous set —
"serves nothing today" is not the same as "abandoned") into
`EXCLUDED_live_clook_2026-09-03.txt`, regenerated as **draft7 = 2,839**.
Verified: 2,840 `<row` = header + 2,839; the four sensitive domains
individually confirmed absent by direct grep.

## 2026-09-03 (later still) — OWNER RULING: list live sites too, priced high — a new track, not a blank-price bulk add

Put to the owner directly: relojistas.com has no Sedo listing, does he
want it listed now. **His answer went further than the question**: "yes,
list all live sites, we can price them quite high e.g. webdesign.uk can
be over 1 million because it's probably going to be worth that in a year
or so." This reverses the fence's original premise — live sites are not
protected-by-default any more, they are a HIGH-VALUE TIER that has been
sitting unpriced this whole session, not excluded stock.

**This is NOT being implemented as "unfence everything and ship at
blank/no-price"** — that would misrepresent domains the owner values at
seven figures as a plain make-offer with nothing behind it, which is a
worse error than the withholding this replaces. Live sites need their
OWN pricing track from the valuation lane before anything ships — relayed
to them as a new, distinct workstream (see reply), not folded into the
existing MAKE_OFFER/no-price portfolio sheet.

**The four sensitive domains are a separate, narrower question, held out
even under this ruling pending explicit confirmation**: "list all live
sites" almost certainly was not meant to include the owner's own email
domain (wpx.uk), his own company's domain (designconsultancy.co.uk), or
an active client relationship (leopardess.co.uk/.uk) — asking rather than
assuming, per the same principle applied to copyonline.co.uk and every
other edge case this session. copy_quality_two_stage's D4 rule (for-sale
needs per-site confirmation, precisely to prevent a client relationship
breach) is their lane's mechanism to protect exactly this case; this
lane's job is to not route around it by accident.

## 2026-09-03 (later still) — a NAMED, DOCUMENTED owner precedent found: leopardessconsulting.co.uk must not be swept into "list all live sites"; webdesign.uk ≠ webdesign.co.uk

copy_quality_two_stage caught something more severe than the two adjacent
`leopardess.*` domains from the Clook batch: **the client site itself,
`leopardessconsulting.co.uk`**, is a paying client's site, and the owner
has an EXISTING, DOCUMENTED, NAMED ruling about it —
`about_page_commercial/PLAN_2026-07-24_about_page_commercial.md`, D4,
verbatim: *"'buy this site'/ads/built-by on a paying client's site
(leopardess) is a relationship breach"* — the exact case D4 exists to
prevent, and the owner's own words name this domain specifically.

**Verified**: `leopardessconsulting.co.uk` is `status=deployed`, 55
pages — squarely inside today's "list all live sites" instruction if
read literally, and it is currently correctly fenced (in
`EXCLUDED_live_2026-09-03.txt`, absent from draft7). The risk is not
today's sheet — it's the FUTURE live-sites track: nothing currently
distinguishes "fenced because nobody's decided yet" from "fenced because
listing this one for sale would breach a client relationship and the
owner already said so in writing." Those need different treatment when
the live-sites track gets built, or this one quietly rides along with
the rest once prices exist.

**Reasoning for treating this differently from wpx.uk/
designconsultancy.co.uk/leopardess.co.uk/.uk**: those four are held
pending a first confirmation; this one has an EXISTING, prior, specific,
documented owner ruling that predates and outranks today's broad
instruction (D4, 2026-07-24, six weeks old, names the site by name). A
broad "yes, all of them" answered in a conversation about relojistas.com
should not be read as silently overriding a specific, written, named
precedent about a different domain — the peer's own framing: "that
deserves an explicit yes, not an inferred one." Applying the same
standard I've used all session: when in doubt, hold out and ask by name.

**Action**: recording `leopardessconsulting.co.uk` as PERMANENTLY EXCLUDED
from any live-sites-for-sale effort unless the owner reconfirms BY NAME,
citing D4, not folded into a general "all live sites" sweep. This is
distinct from the live-site fence (which could in principle lapse or be
regenerated) — like the copyonline.co.uk and wykefarm.co.uk precedents,
a client-protection decision needs to survive independent of build
status. No new exclude file needed today (already fenced, not in any
sheet) — the durable record is this NOTES entry plus the sharpened
question put to the owner.

**Second catch, smaller but real**: `webdesign.uk` (18 pages) and
`webdesign.co.uk` (155 pages, the actual shopfront/store-window for the
webdesign business) are two DIFFERENT domains. The owner's 7-figure
example was specifically `webdesign.uk`. Checked: both are tracked as
distinct entries throughout this lane's fence files and sheets — no
conflation found — but the owner's own intent needs the same precision:
his example named one, and it matters which one a live-sites pricing
effort actually prices and lists. Flagging to the valuation lane and the
owner rather than assuming his answer covers both identically.

## 2026-09-03 (later still) — owner ruled per-domain on the 39-domain Clook batch; draft8 = 2,860; one domain unaddressed; a real cost basis surfaced

Owner sent explicit exclude/release lists covering the 39-domain Clook
batch. **Reconciled by diff, not by eye**: his exclude list (17, after
dedup — `designconsultancy.co.uk` and `dsgn.co.uk` were each named
twice in his message) plus his release list (21) = 38. **One of the 39
was in neither list: `2v.uk`.** Not assumed either way — kept in the
fence pending explicit word, same principle as every other edge case
this session.

**Owner's 17 confirmed exclusions**: websy.uk, workdomain.co.uk, wpx.uk,
leopardess.co.uk, leopardess.uk, healthcare.uk, designconsultancy.co.uk,
dsgn.co.uk, 5un.co.uk, onpointcopy.co.uk, managementemail.co.uk/.uk,
minisitemaker.co.uk, emailsecurity.uk, vectordb.uk, email-account.co.uk/.uk.
Notably these read as operational/product/service names (email, design
consultancy, minisite maker, vector db) rather than generic keyword
domains — consistent with the pattern already flagged for wpx.uk/
designconsultancy.co.uk. `EXCLUDED_live_clook_2026-09-03.txt` rebuilt to
these 17 + 2v.uk = 18 (down from 39).

**Owner's 21 released** (agentcoordinator.uk, aiartgallery.uk,
businesschristmasgifts.co.uk/.uk, businessinsurancequotation.co.uk,
cartoon.co.uk, catalogues.co.uk, conferences.co.uk, fatherchristmas.uk,
felines.co.uk/.uk, fridge-magnets.co.uk, pelletburners.co.uk,
personalgift.co.uk, personae.uk, santaclaus.uk, seduce.co.uk,
soyrocks.co.uk, uniquedirectory.co.uk, vinrose.uk, writesy.uk) — released
back into the ORDINARY portfolio sheet, not the high-value live-sites
track, since "don't need to exclude" reads as "treat as normal stock,"
not "list at a premium." Draft8 = 2,839 + 21 = **2,860**, verified: 2,861
`<row` = header + 2,860; cartoon.co.uk confirmed present as MAKE_OFFER/
blank; the four permanently-held domains (wpx.uk, leopardess.co.uk,
leopardessconsulting.co.uk, 2v.uk) confirmed still absent.

**A real cost basis surfaced, same message**: owner said "I paid over
5000 pounds for cartoon.co.uk so don't underprice that one." This is now
a genuine floor — same class as relojistas.com's owner-set $12k Afternic
floor — and it must reach the valuation lane's pricing pass BEFORE
cartoon.co.uk gets a number, not discovered after. Relayed (see reply).
Worth noting for future releases from this batch: if one domain carried
an unstated £5k+ cost, others in the released-21 may too — the owner
volunteered this one unprompted, nothing here rules out more.

**Cross-check against domain_valuation's parallel finding**: their
Dynappraisal tool priced relojistas.com at $1,490 against the owner's own
$12k floor (8×) and webdesign.uk at $1,144 against his stated
">£1,000,000" (~1,100×) — the algorithmic appraiser is structurally blind
to exact-match/category-leadership value and cannot anchor the
high-value tier at all; it stays valid for ordinary keyword-domain stock.
Their principle — carry the owner's own figure labelled AS HIS ESTIMATE,
never silently substitute the appraisal for it — applies equally to
cartoon.co.uk's £5k floor. wpx.uk confirmed never-list by both lanes
independently, for the same reason (owner's own email address).

## 2026-09-03 (later still) — leopardessconsulting.co.uk CLOSED: owner's explicit word, via a different thread

copy_quality_two_stage relayed the owner's direct answer to the
sharpened question this lane raised: **"no leopardessconsulting need not
be listed."** The permanent exclusion (D4 precedent) is now confirmed by
his explicit word, not merely inferred from the documented ruling —
nothing to wait on further here. **Worth recording the mechanism, not
just the result**: he answered a different thread than the one that
asked, same shape as the copyonline.co.uk latitude earlier — his replies
land wherever he happens to be, not necessarily the lane that posed the
question. Practical consequence: don't assume silence on a question this
lane raised means it's unanswered; check whether a peer thread has
already relayed the answer before re-asking or continuing to wait.

Still genuinely open, both confirmed unaddressed by any thread as of
this entry: **2v.uk** (Clook batch, neither excluded nor released) and
**which webdesign** (webdesign.uk vs webdesign.co.uk) the owner meant —
copy_quality_two_stage deliberately left the second one to this lane
rather than duplicate the question.

## 2026-09-03 (later still) — both open questions CLOSED: 2v.uk never-sell; both webdesigns are in scope, will converge to one endpoint

Owner: "both webdesign's they will be the same endpoint one day, don't
sell 2v.uk."

**2v.uk**: CONFIRMED permanent exclusion, no longer pending — was
already sitting in `EXCLUDED_live_clook_2026-09-03.txt` as a cautious
hold, now that hold is upgraded to a decided "never sell," same standing
as the other Clook exclusions. No sheet change (already absent from
draft8) — this is a documentation-only closure, recorded so a future
regeneration never treats it as still-open.

**webdesign.uk / webdesign.co.uk**: BOTH are in scope for the high-value
live-sites tier, resolving the earlier ambiguity — his 7-figure remark
was not exclusive to the 18-page one. New fact worth carrying forward
carefully: **"they will be the same endpoint one day"** — a stated future
consolidation/redirect plan for the two domains. This has a real
implication beyond pricing: whichever pricing or listing work happens
now should not assume today's two-separate-sites structure is permanent,
and the about-page-commercial lane's CTA work may care about which
domain becomes canonical and when. Relayed to both peer lanes — this
lane does not act on the consolidation itself, only flags it exists.

## 2026-09-03 (later still) — webdesign shopfront claim CORRECTED (by a peer, verified before propagating); merge-vs-sale coupling RULED: quote as a pair

**Correction, caught by copy_quality_two_stage and independently
verified before accepting it**: this lane's own RUNBOOK claimed
`webdesign.co.uk` (155 pages) was the shopfront — backwards. Checked
`CLAUDE.md:716` directly rather than trusting the correction on say-so:
it quotes the owner's own words, "the webdesign.uk shopfront." So
`webdesign.uk` (18 pages) is both the shopfront AND the owner's
seven-figure example — the same domain, not two separate facts. The
original wrong claim was an inference from page count and domain
pattern, never checked — same failure class this session has caught
twice already (the date mixup, the git-mv duplication). RUNBOOK
corrected with a visible `> **CORRECTED**` marker rather than silently
edited.

**The coupling copy_quality_two_stage raised — two domains destined to
merge, both in scope for independent sale — put to the owner directly**
(AskUserQuestion): quote as a pair / list separately and accept the risk
/ hold both off the market. **RULED: quote as a pair.** Matches the
valuation lane's own recommendation exactly (their reasoning: two
independent prices would systematically under-price the pair and
over-price the part — consolidation makes one canonical and the other a
redirect, so selling either alone hands away brand adjacency the
business needs). Relayed to the valuation lane as the ruling, not a
recommendation to weigh — their pricing work on this pair should proceed
on a combined basis, not two independent figures.

## 2026-09-03 (later still) — owner asked: double-check no BUY_NOW prices exist, and ban it except when he does it himself

**Verified directly, every draft, not just the latest**: checked all 8
generated sheets (draft1 through draft8) with a direct `awk` scan for
`selling_option=BUY_NOW` or any non-empty price/min_price/currency —
**zero** in all 8. (First check accidentally globbed the `_provenance`
CSV instead of the sheet CSV via `ls -t` and read the wrong columns —
caught before reporting, redone against the correct file explicitly.)

**Built a hard technical gate**, `scripts/domains/sedo-importer-xlsx.py`
§8: `enforce_buy_now_gate()` refuses (SystemExit, prints every blocked
domain+price) any BUY_NOW selling option or non-empty `price` unless
`--owner-authorized-buy-now-prices` is passed; `min_price` alone is
never gated (it's a protective floor, the opposite of what this guards
against). Checked twice — once against the in-memory rows before
writing, once against the WRITTEN artefact after (defense in depth
against a future code path bypassing the first check). Self-test proves
BOTH directions of the gate (blocks without the flag, allows with it),
not just that the normal path still works — a test that only exercises
the happy path cannot prove a guard exists. Verified live from the real
CLI too, not just self-test: built a real priced sheet, confirmed
refusal without the flag (exit 1, every domain listed) and success with
it (exit 0, loud banner naming what shipped).

**The flag is not automation-facing** — it's named and documented as
something the OWNER uses himself, deliberately, per run; a session
reaching for it because "the price data looks ready" is exactly the
failure mode it exists to stop. This is the same shape as this
project's other owner-authority mechanisms (opt-in field, unsafe default
OFF) rather than a config toggle that could quietly stay on.

**Confirmed the guard doesn't disturb the current sheet**: regenerated
draft8's inputs through the gated script — zero prices found (correctly
a no-op), output otherwise structurally identical except **+34 new
domains** the registrar CSVs have picked up since draft8 was built
earlier today (other lanes refresh those files independently; not a
gate issue, not acted on — draft8 stays the official current sheet,
noted for whoever next regenerates from fresh exports).

## 2026-09-03 (later still) — CORRECTION: leopardessconsulting.co.uk is the owner's OWN business, not a third-party client

> **CORRECTED 2026-09-03**, caught by the valuation lane, not by this
> lane: every earlier entry in this file describing
> `leopardessconsulting.co.uk` as "a paying client's site" is WRONG. It
> is the owner's own consultancy — its live copy reads *"Leopardess
> Consulting | AI systems that do one defined job… We run 22 of our own
> sites on Kubernetes, Kafka and Postgres"* (this platform, sold as a
> service), and the owner confirmed directly, describing it as "a
> representation of my own services" while asking for a second domain
> for the same purpose. Neither this lane nor copy_quality_two_stage
> verified the "client" framing against the live site before repeating
> it — it was taken from D4's own wording and never checked, exactly the
> failure class this session has caught in itself twice already (the
> date mixup, the webdesign shopfront claim).

**What this changes**: the "extra scrutiny on any future yes" reasoning
recorded earlier was premised on third-party relationship harm — that
premise is false. The correct standard is the ordinary live-site one
(same as webdesign.co.uk, idea.uk): his call alone, no special
third-party-harm weight to it.

**What this does NOT change**: the exclusion itself. Independent of the
reasoning, the owner ALREADY gave an explicit "no, leopardessconsulting
need not be listed" (relayed 2026-09-03, logged above) — that answer
stands on its own regardless of which justification was attached to the
question that produced it. No sheet action needed.

**Not this lane's fix**: whether D4's own text (`about_page_commercial`
PLAN) actually says "client" or whether that framing was added
downstream is for the owning lane to check — domain_valuation flagged it
to them directly. This lane's job was only to correct its own record,
which is done.

## 2026-09-03 (later still) — leopardessconsulting.co.uk, third movement: extra scrutiny REINSTATED, by the owner's own ruling this time

Full arc, worth keeping visible rather than collapsed into one final
answer, per copy_quality_two_stage's own reasoning: an inherited gloss
(client site, taken from a D4 quote unverified) → refuted from the live
site (it's the owner's own consultancy) → **reinstated by explicit
ruling**: "we can work as if leopardess is a paying client if that
helps, I do pay through the nose for these tokens."

**Both things are true at once, and neither contradicts the other**:
the FACT (his own consultancy, no external client harmed) stands as
corrected above; the RULING (treat it with the D4 client-protection
standard regardless) is separate and controlling. He is opting IN to the
stricter standard by his own choice, not because a third party would be
harmed — the earlier correction downgraded to "ordinary live-site
standard" on the assumption the stricter standard only made sense if a
real client existed; that assumption was wrong. **Restoring: any future
"yes, list it" needs the same extra scrutiny (his explicit word, by
name, not inferred from a broader instruction) as before** — the
standard, not the original justification for it, is what this lane
actually needs to enforce. The exclusion itself was never in question
through any of the three movements; only the correct standard for
handling a future reversal was.

No sheet action — draft8 already excludes this domain, unaffected by
any of the three corrections.

## 2026-09-03 (later still) — the pricing gate that was blocking the built-sites listing is REMOVED; draft9 folds live sites into the main sheet; floor policy clarified in two passes

**First pass**: owner, relayed via copy_quality_two_stage — "we can't do
accurate minimum offers yet and probably don't want to because it sets
an expectation of ballpark in the buyers mind. We'll just have to bear
with the low balls for a while." This removed the reason the whole
"live sites, priced high" track had been held separate from the
ordinary portfolio: it was waiting on real prices before listing;
real prices are now explicitly not required. Put the exact mechanics to
the owner directly (AskUserQuestion: blank Minimum Offer vs a small
nominal one) — **answer: leave it completely blank, revisit later.**

**Action taken**: since "blank Minimum Offer, MAKE_OFFER, for-sale" is
literally the SAME row shape the ordinary portfolio sheet already
defaults to, folded the previously-held live sites straight into the
main sheet rather than keeping a separate track that no longer has a
distinct reason to exist. Re-queried `sites` fresh first (42 rows,
diffed against the committed 50-domain fence — the 8 extra entries in
the fence are legitimately from the OTHER union sources, Nominet zones
and NS-based check, not staleness; nothing new to add). Built
`EXCLUDED_owner_leopardessconsulting_2026-09-03.txt` (a dedicated file,
since its exclusion reason — owner ruled client-protection standard,
2026-09-03 — is now wholly separate from "pending pricing," which no
longer applies to anything) and dropped `EXCLUDED_live_2026-09-03.txt`
from the build entirely — its sole purpose (hold live sites off pending
a price) no longer exists. The 18-domain Clook batch (owner-named,
unrelated to pricing) and the appleby/wykefarm/copyonline
owner-withdrawals stay exactly as before — this ruling only released
domains that were held FOR PRICING REASONS.

**Draft9 = 2,943 domains** — the full estate minus only owner-named
holds (36 total: Clook 18, appleby 7, wykefarm/pasturedegg 8,
copyonline 1, leopardessconsulting 1 — some overlap in the 39 Clook
count already resolved). Verified: 2,944 `<row` = header + 2,943;
relojistas.com, webdesign.uk, webdesign.co.uk all confirmed present as
MAKE_OFFER/blank; leopardessconsulting.co.uk confirmed still absent;
zero BUY_NOW/priced rows (the §8 gate was a correct no-op).

**Second pass, same hour, a genuine refinement not a reversal** —
copy_quality_two_stage relayed the owner's fuller position: *"we can set
minimum floors in sedo and not have minimum floors on the sites…
unlink the two."* This does NOT change draft9 (still correctly blank,
since no concrete number has been given), but it corrects the
STANDING POLICY going forward:
- **Sedo's Minimum Offer field MAY carry a real floor** whenever the
  owner gives a number directly or one is agreed with him in
  conversation — "blank for now" was never "blank forever," and this
  lane should set a floor the moment a real number exists, not wait for
  a blanket policy change.
- **The on-site display (about-page CTA etc.) must NEVER show a
  price/floor** — confirmed structurally true already by
  copy_quality_two_stage (no `price`/`floor`/`minimum`/`offer` token in
  the template); that constraint is theirs to hold, not this lane's.
- **A Sedo floor must NEVER be derived from `site_specs.commercial.tier`
  or any automated appraisal** — floors come from the owner directly or
  explicit agreement with him, full stop. This severs the same
  dependency the about_page_commercial and afternic lanes' PLANs had
  wrongly assumed (their PLAN text named "price-by-tier" as a pricing
  source; copy_quality_two_stage is notifying the afternic lane
  directly, not this lane's fix).
- The valuation lane's appraisal/coverage work is off this lane's
  critical path a second, independent way: not needed to list (already
  established) and not needed to price either (a Sedo floor is a
  direct-owner or agreed number, never a derived one). Their work may
  still be useful as an INPUT to a conversation with the owner about
  what number to agree on — just never auto-applied.

## 2026-09-03 (later still) — relojistas.com / free.me.uk: owner accepted the cross-marketplace inconsistency rather than match their existing Afternic floors

domain_valuation raised a sharper case than the general blank-minimum
ruling: `relojistas.com` and `free.me.uk` already carry OWNER-SET
minimums elsewhere ($12,000 and $50,000, Afternic) — blanking the Sedo
listing means a buyer could route around a floor the owner already
decided, on the same domain, via a cheaper door. Different question from
"names with no agreed number" — these two have one, from him, by name.

**Put to the owner directly** (not inherited from the general ruling):
match the existing floors on Sedo, or accept the inconsistency and leave
them blank too. **Answer: leave them blank too.** No sheet change —
draft9 already has both blank; this closes an open question rather than
changing anything. Recorded as his deliberate choice, not an oversight —
worth knowing precisely because a future reader might otherwise "fix"
the inconsistency assuming it was never asked about.
