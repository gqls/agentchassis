# HANDOFF — Sedo domain management — read this first, then RUNBOOK, then NOTES

**Read order for a fresh session**: this file (state + next actions) →
`RUNBOOK_sedo_domain_management.md` (mechanics, §1–§9, especially §7–§9)
→ `PLAN_2026-09-02_sedo_domain_management.md` (why, phases) → `NOTES_*.md`
only if you need the full evidentiary trail for something below (it is
long — this file exists so you don't have to read it end to end just to
pick up work).

## Where things stand (2026-09-04)

**Current sheet**: `outbound/SEDO_IMPORT_2026-09-04_draft10.xlsx` —
**2,942 domains**. Owner has NOT confirmed he's uploaded it yet as of
this handoff. Every row is `MAKE_OFFER` / for-sale / **blank price and
blank minimum** — this is deliberate, not a placeholder (see "Floor
policy" below), and it is enforced: the generator hard-refuses to write
a BUY_NOW or priced row without an explicit owner-only flag (RUNBOOK
§8). Verified zero priced rows in every draft built so far (1 through
10).

**What's excluded and why** — six reason-files, unioned at build time,
each named by reason (never merge two reasons into one file — see
RUNBOOK §7):
| file | domains | reason |
|---|---|---|
| `EXCLUDED_live_clook_2026-09-03.txt` | 18 | owner named these specifically (a second hosting stack, Clook, that the framework's own live-site tracking can't see — includes his own email `wpx.uk` and company domain `designconsultancy.co.uk`) |
| `EXCLUDED_owner_appleby_2026-09-03.txt` | 7 | Appleby family name domains |
| `EXCLUDED_owner_wykefarm_pasturedegg_2026-09-03.txt` | 8 | Wyke Farm + pastured-egg brand family |
| `EXCLUDED_owner_copyonline_2026-09-03.txt` | 1 | copyonline.co.uk — personal-use keeper |
| `EXCLUDED_owner_leopardessconsulting_2026-09-03.txt` | 1 | owner's own consultancy site — he's opted to treat it with the same protection as a real client relationship (D4-level), by his own explicit choice, not because a third party is involved (full three-part correction story in NOTES if you need it — the short version above is all that matters going forward) |
| `EXCLUDED_owner_rolex-submariners_2026-09-04.txt` | 1 | rolex-submariners.com, likely trademark exposure |

**The old `EXCLUDED_live_2026-09-03.txt`** (the original 50-domain
"protect every live/deployed site" fence) is **retired, not deleted** —
kept on disk for the historical record but no longer passed to the
build command. It stopped mattering when the owner ruled that built
sites should be listed too (see below); do not resurrect it as an
exclude-file without checking this file first, or you'll silently
re-exclude ~49 domains the owner asked to have listed.

## The big decision this session: built/live sites are now listed too

Early in the session, live sites (built, deployed, actually serving
content) were excluded by default — a safety default this lane set
itself, not an owner instruction. The owner later reversed this
explicitly: **"list all live sites, we can price them quite high."**
Then, when pricing turned out to be unready, he removed even THAT
requirement: **"we'll bear with the low balls for a while"** — confirmed
directly that Sedo's Minimum Offer field should be **completely blank**,
not a small nominal number, for now.

**Standing floor policy (RUNBOOK §9), settled after two refinement
passes — this is the part most likely to matter for whatever you do
next**:
- Sedo's `Minimum Price` field **may** carry a real number whenever the
  owner states one directly, or one is explicitly agreed with him in
  conversation. "Blank for now" is not "blank forever."
- The domain's own live site must **never** display a price/floor —
  confirmed structurally true already (a different lane's concern, not
  this one's to enforce, just don't contradict it).
- **Never derive a Sedo floor from `site_specs.commercial.tier` or an
  automated appraisal.** A floor is a direct-owner number or an
  explicitly agreed one, full stop. The domain_valuation lane's
  appraisal tool is known to be wildly wrong at this end of the market
  (undershot the owner's own $12k relojistas.com floor 8×, and his
  stated >£1M webdesign.uk estimate ~1,100×) — it's a fine instrument
  for ordinary keyword-domain stock, not for this.
- **Known, deliberate exception**: `relojistas.com` and `free.me.uk`
  already carry owner-set minimums elsewhere (Afternic: $12,000 and
  $50,000). The owner was asked directly whether Sedo should match those
  or go blank like everything else — he chose **blank, accepting the
  cross-marketplace inconsistency**. This is a recorded decision, not an
  oversight — do not "fix" it by adding those floors back without him
  saying so again.
- `webdesign.uk` and `webdesign.co.uk` are two different domains (do not
  conflate — `webdesign.uk`, 18 pages, is BOTH the shopfront and the
  owner's own £1M+ example; `webdesign.co.uk`, 155 pages, is a separate
  large site). Owner: they'll converge to one endpoint eventually, and
  he wants them **quoted as a pair**, not two independent prices, when
  real pricing ever happens.

## OPEN — needs the owner's word, not yet asked to conclusion

**Two more possible trademark-risk domains, currently still in the
sheet**: `mieleonline.com` (Miele, live German appliance manufacturer)
and `webuyanycarandvan.com` (reads as a brand extension of
webuyanycar.com, a major UK company — the sharper UDRP risk of the two).
Flagged by the domain_valuation lane the same way rolex-submariners.com
was flagged before its withdrawal — that flag was right, this one is
unverified. **Put both to the owner by name before the next upload**;
do not assume the omission (he named rolex specifically but not these
two) was deliberate — ask, per the pattern used for every other
edge case this session.

**Four person-name domains never got an answer**: `ianstirling.com`,
`kapoor.uk`, `keeler.uk`, `anne-marie.co.uk` — flagged early on as
having no obvious family connection to anything else, owner never
responded either way. Still sitting in the sheet, for sale, blank price.
Low urgency (nothing suggests special risk) but worth a batched question
next time you have his attention rather than leaving indefinitely open.

**Registrar exports have drifted — re-pull before the next regenerate.**
`spaceship_domains_2026-09-03.csv` exists in the valuation lane's
`inbound/` and is **44 domains larger** (247 vs 203) than the
`spaceship_domains_2026-09-02.csv` draft10 was actually built from.
Dynadot/Porkbun/Nominet may have moved too — check dates on all four
before trusting draft10 is complete. RUNBOOK §7/§9 build command needs
the freshest files, and the live-site re-query (`SELECT domain FROM
sites WHERE status IN ('deployed','test')...`) should be re-run fresh
too, not reused — see the copyonline.co.uk incident in NOTES for why
(a fence built once and reused across regenerations missed a
same-morning addition).

**Sedo API credentialed access — still entirely on the owner**: he has
the account + partner status already (confirmed early this session,
`info@designconsultancy.co.uk`), but RUNBOOK §2 (the access-request
email) and §3 (installing the K8s secret) are both still outstanding.
Nothing here is blocked on it — the web-import route (uploading the
xlsx by hand) works today without it — but the credentialed route would
remove the owner as the transport for every future change.

**Pricing for the ORDINARY portfolio** (not the live-sites tier, which
has its own settled policy above): still waiting on the domain_valuation
lane's `OUTPUT_prices_<date>.csv`. Two live constraints on that work,
both already relayed to them, no action needed from this lane: (1)
`cartoon.co.uk` has a real owner-stated floor (paid over £5,000 — "don't
underprice that one") that must be honoured; (2) the valuation lane
discovered NO domain in this estate has its acquisition cost recorded
anywhere, so the "keen bottom-500" pricing plan has an unbounded
underpricing risk for anything bought at real cost, not just
cartoon.co.uk — they're addressing this with the owner directly.

## Mechanics reference (don't re-derive, just use)

- **Generate a sheet**: `scripts/domains/sedo-importer-xlsx.py build
  --out ... --domains <4 registrar CSVs> --exclude-file <the 6 files
  above>` — see RUNBOOK §7 for the exact command and §8/§9 for the
  BUY_NOW gate and floor policy. `--self-test` first if you've touched
  the script (14 checks as of this handoff).
- **Verify any new draft at the artefact, not by trusting the script's
  own printed count**: unzip the xlsx, `grep -o '<row ' … | wc -l`
  (never `grep -c`, it undercounts single-line XML), diff the CSV
  against the prior draft to confirm ONLY the intended domains moved.
- **Dating anything new**: run `date +%F` yourself before writing a date
  into a filename or doc heading. This lane mis-dated a whole session's
  work once already (see WRONG_CALLS.md, 2026-09-03 entry) from
  inferring "today" off a filename glimpsed in another lane's directory
  instead of checking the clock.
- **Peer lanes actively coordinating on this**: `domain valuation`
  (pricing, cross-checks the fence independently, owns acquisition-cost
  risk) and `copy quality two stage` (the about-page CTA that will
  eventually point at Sedo listings — D1 is ruled, points to Sedo,
  blocked only on relojistas.com actually having a real listing/URL to
  point to, which hasn't happened yet since nothing has been uploaded).
  Both have been reliable at cross-checking this lane's own claims
  (caught the git-mv duplication bug, the webdesign shopfront mix-up,
  and the leopardessconsulting mischaracterisation, among others) — read
  their messages carefully rather than assuming your own record is
  complete.
