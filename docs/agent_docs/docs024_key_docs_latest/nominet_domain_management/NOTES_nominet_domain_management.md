# NOTES — Nominet domain management (append-only, newest at the bottom)

## 2026-09-02 — lane created on the owner's directive; first act was an incident

Owner: *"please pick up the nominet thread here"*, then *"if it doesn't exist
please create it and take responsibility for everything nominet here."* No
nominet lane existed — the state was scattered across domains_cloudflare_rollout
(EPP access, allowlist), idea_uk_vm_site (the client family), portfolio_positioning
(the inventory runbook), site_delivery_and_editor (second TAG, domain programme)
and webdesign_couk (the registrar page). Consolidated view: PLAN. Sources were
read, not trusted from memory glosses; markers below say which claims this
session re-verified.

### Verified first-hand today

- `~/.config/nominet/credentials` exists, 0600, 47 B (existence only — contents
  never read; owner ruling 2026-08-23).
- `postgres-clients-0` had LOST `/tmp/epp.pl` (pod restart since 08-19; pod age
  6d6h). Re-staged from `scripts/domains/epp.pl`, verified present.
- Pod now sits on node `.37` (was `.26` at the 08-11 login proof). Greeting from
  there: 2,531 B — reachability only; the allowlist claim for `.37` is
  [INHERITED, owner-confirmed 08-11] until the next login.
- EPP list walk has NEVER run: no `all_domains.txt` anywhere on this machine
  (find / — 0 hits). The ~1,500-domain tag is unenumerated while Spaceship
  (203), Dynadot (451) and Porkbun (683) all got inventories today.

### THE INCIDENT — the owner's Nominet NS batch ran ahead of the Cloudflare zones; four domains dark

Chain, each step measured 2026-09-02 ~17:00Z:

1. `dig +short NS advertise.co.uk @1.1.1.1` → **SERVFAIL** (and no A record
   anywhere). The portfolio lane's note from 16:23Z ("domain still serves
   Drupal") was already stale — that was resolver cache, not the registry.
2. `dig +norec NS advertise.co.uk @dns1.nic.uk` → **alexis/leah.ns.cloudflare.com**
   — the Nominet-side repoint HAS HAPPENED (registry TTL 172800, so caches die
   over ~2 days, staggered).
3. `dig A advertise.co.uk @alexis.ns.cloudflare.com` → **REFUSED** — no zone
   answers at the edge.
4. `GET /zones?name=advertise.co.uk` (portfoliotoken) → **empty result. No zone
   in the account.** The LANDMINES "dangling delegation" trap, exactly.
5. Breadth: probed all 26 remake-decision domains at `dns1.nic.uk`. **Four**
   delegate to alexis/leah with no zone: `advertise.co.uk`, `designblog.co.uk`,
   `seotools.co.uk`, `websitepromotion.co.uk` — precisely the four with fired
   briefs, so this was the owner's "Nominet batch" for the remake programme.
   The other 22 still sit at `dns.uk-noc.com`/`dns.us-noc.com` (the old
   hosting), except `businessinsurancequotation.co.uk` = no delegation at all
   (the known "never set a nameserver" class).
6. Account zone census: 42 zones, ALL `active`, none of the four. Two NS pairs
   in service (alexis/leah + betty/ivan) — [UNINVESTIGATED] why two pairs; not
   load-bearing today since the batch pointed at alexis/leah and new zones on
   this account got alexis/leah on both reads taken today.

**Recovery built + staged**: `scripts/domains/cf-zone-bootstrap.sh` — encodes
the proven 2026-08-25 homegarden.uk recipe (zone → 2 proxied placeholder A
records → 2 `portfolio-sites-router` routes → re-read verify against the
garden-tools.uk reference → activation check + poll). `--check` (read-only)
PROVEN today: four NO-ZONE + a correct reference read, honest exit 1. The
mutating half is **owner-run**: the session classifier refused the zone-create
POST (same class as the 08-19 EPP-credential refusal; not worked around).

**Cross-lane action noted, not silently absorbed**: the Cloudflare half belongs
to domains_cloudflare_rollout; staging their half was right because the incident
is one piece of work triggered from this lane's side (the Nominet batch). Their
NOTES got the contribution entry; ownership boundary recorded in both PLANs.

### Also found today

- `~/.config/cloudflare/token` (the read-only CF token) is **DEAD** — `9109
  Invalid access token` on every call. Discovered because the zone-existence
  check tried it first. `portfoliotoken` fine. Recorded in their NOTES.
- cookly.uk, idea.uk: registry delegation on alexis/leah, zones active — the
  old "cookly repoint NOT DONE" thread is long closed. lendzy.co.uk sits on
  betty/ivan, zone active.

### Open questions carried

1. Second TAG application — submitted 2026-08-11, nothing heard. Ask the owner
   whether Nominet has responded.
2. The tag inventory route — owner's pick: CSV export (preferred) or the staged
   walk (RUNBOOK §2).
3. Did the owner's Nominet batch touch MORE than these four? Only the tag
   inventory (or the owner) can answer — the 26 remake domains were the only
   candidate set a session could probe.

## 2026-09-02 (later) — owner: "same sort of thing as the other registrar lanes" → family client built

Owner clarified the lane's shape: Nominet gets what dynadot/porkbun/spaceship
got — one consolidated client under `scripts/domains/`. Until now the verbs
were scattered (epp.pl = login/list only, pod-copied, creds on stdin; three
box one-offs for check/ns-change/register, each direct-socket and therefore
only runnable from an allowlisted box).

**Built: `scripts/domains/nominet.py`** — probe/login/list/walk/check/info/
set-ns, credentials read in-process from `~/.config/nominet/credentials`
(never printed, never argv, XML-escaped into the login — the box scripts
interpolated the password raw), transport = the kubectl-exec openssl tunnel
from the allowlisted cluster egress (the 08-11 login-proof mechanism), framing
local. `register` deliberately refuses and points at VMB-017 — one
implementation of the money-spending verb, not two.

**Design fix over the staged recipe: `walk --months` up to 120, default 24.**
The 08-19 recipe walked 12 expiry months; .uk registrations run up to 10
YEARS, so any multi-year registration expiring outside the window is silently
absent and the symptom is a plausible short list. 120 months bounds the space;
the count check against ~1,500 remains the arbiter.

**Proof state, honestly:** `--self-test` 15/15 [MEASURED 2026-09-02] (framing,
all six XML builders well-formed, password escaping, both check classes, NS
normalisation, year-rollover month walk, credentials parser incl. refusal);
`probe` = 2,527 B greeting through the pod tunnel [MEASURED 2026-09-02].
**Every credentialed verb UNEXERCISED** — the session classifier refused
`login` (one attempt, not retried; consistent with 08-19 and today's CF
refusal). The owner's first `login` proves the client AND re-proves the
allowlist for the `.37` node the pod moved to.

Misstep caught in-session: the self-test summary line hardcoded `14/14` while
15 checks ran — a dishonest summary over honest checks; replaced with a real
counter before first commit.

## 2026-09-02 (evening) — FIRST CREDENTIALED RUN: `login` PROVEN by the owner

Owner ran `python3 scripts/domains/nominet.py login` in-session:
`GREETING_BYTES=2527` → `login: 1000 Command completed successfully`
`[MEASURED 2026-09-02]`. Two facts at once:

1. The client's credentialed path works end-to-end (credentials parse, escape,
   frame, tunnel, login) — OPP-015's "unexercised" caveat now covers only
   walk/list/check/info/set-ns.
2. **The allowlist is re-proven for node `.37`** — the 08-11 proof was from
   `.26`, the pod has since moved, and this is the first login from the new
   node. The owner's "all five added" claim of 08-11 now has a second data
   point.

## 2026-09-02 (evening, cont.) — zones created BUT Cloudflare assigned betty/ivan: the cutover now needs the NOMINET side moved

Owner ran `cf-zone-bootstrap.sh` on the four dark domains. All four zones
created + wired + verified 2/2+2/2 against the reference `[MEASURED
2026-09-02]` — **and every one was assigned `betty/ivan.ns.cloudflare.com`,
not alexis/leah**. The registry delegates the four to alexis/leah (the owner's
batch copied the pair the older zones use), so the zones sit `pending` and
**can never activate on the current delegation** — pair mismatch, not
propagation. My script's "usually clears within the hour" note was the wrong
guidance for this case; fixed in place: it now compares the registry
delegation against the assigned pair and FAILS loudly on mismatch.

This ANSWERS OPP-014's verify-later question: new zones on this account get
**betty/ivan** `[MEASURED 2026-09-02, 4/4]`; the 32 alexis/leah zones are the
older cohort. Consequence for every future cutover: **take the pair from the
zone-create response, never from a sibling zone** (the rollout RUNBOOK already
says this; now there is a 4-domain measurement behind it).

**Fix staged (owner-run, our lane's verb):** repoint the four at Nominet to
betty/ivan via `nominet.py set-ns … --apply`, then re-run the bootstrap (idempotent)
to re-trigger activation. The alexis/leah delegation currently serving REFUSED
continues until then.

Also: the owner's first `walk` attempt failed at the greeting (30 s read
timeout); a credential-free `probe` immediately after answered 2,527 B —
transient (or a connection-rate throttle from the several connects in quick
succession), not structural. Walk still owed; retry.

## 2026-09-02 (evening, cont. 2) — set-ns PROVEN LIVE 4/4; the Nominet half of the recovery is DONE

Owner ran the staged dry-run loop (read exactly right on all four: rem
alexis/leah, add betty/ivan, nothing else), then `--apply`. All four:
`domain:update` 1000 + verify-by-re-read showing the target pair — the
client's write path (login → info → update → verify) is now proven end-to-end
`[MEASURED 2026-09-02]`.

Independent check at the registry's own servers (`dig +norec @dns1.nic.uk`):
3/4 published betty/ivan immediately; websitepromotion.co.uk still published
alexis/leah on the first read and flipped within ~30 s — **the EPP re-read
confirms the registry DATABASE; the published zone lags by seconds-to-minutes.
Do not read a stale first dig as a failed update when the EPP verify said
SUCCESS.**

Remaining for the four: the owner re-runs `cf-zone-bootstrap.sh` (idempotent)
to re-trigger activation now the pairs match; then serving-verify per the
LANDMINES bar (body property via pinned edge IP, not status codes).

OPP-015 proof state after tonight: probe/login/info/set-ns(+--apply) all
proven live; walk/list/check remain unexercised.

## 2026-09-02 (evening, cont. 3) — INCIDENT CLOSED: all four zones ACTIVE, all four domains SERVING, remake №1 publicly live

Owner re-ran the bootstrap after the repoint: all four zones `active` on the
FIRST poll (the pair fix was the whole story). Serving verified at the
LANDMINES bar — body property via edge IP pinned from the assigned NS, never
a bare status code `[MEASURED 2026-09-02 ~19:00Z]`:

- advertise.co.uk — 200, 75,562 B, title "Advertise.co.uk — The UK Guide to
  Advertising" → **remake №1 is now PUBLICLY LIVE** (it was built + deployed
  16:23Z with only DNS owed; this cutover completed the launch).
- designblog.co.uk — 200, 71,578 B, titled site content.
- seotools.co.uk — 200, 69,837 B, titled site content.
- websitepromotion.co.uk — 200, 71,275 B, titled site content.
- garden-tools.uk (control) — 200, 78,464 B, as always.

Observation routed to portfolio_positioning, not resolved here: the THREE
held-brief domains serve full titled sites, while their briefs sit at
`needs_human_review` — whatever those bodies are, they predate their remakes.
Their lane owns what those domains *should* say; ours only that they resolve
and serve.

Residual: resolvers that cached the alexis/leah delegation SERVFAIL until
their negative caches expire (minutes, not days). Nothing to do.

## 2026-09-02 (late) — INBOUND from the domain-valuation session: the walk is their biggest missing list

Peer session "domain valuation" (valuing the whole portfolio for the owner;
plan: sell the bottom ~500 priced keenly, category families kept together,
.com in scope) asks: when the walk lands, normalise to
`docs024_key_docs_latest/domain_valuation/inbound/nominet_domains_<date>.csv`
(columns `domain,expiry` + whatever else), commit by pathspec, reply via
SendMessage. Accepted — this lane owes them the list whichever session the
owner runs the walk in.

Enabling change made first: `walk` now emits `DOMAIN\t<name>\t<YYYY-MM>` —
the expiry month is the list command's own query key and was being flattened
away by a `set()`. Exact expiry DATES would need `domain:info` per domain
(~1,500 calls); the month is what the walk yields and is the offer unless
valuation needs better. **ANSWERED same evening: month precision is
SUFFICIENT (valuation session, 2026-09-02) — expiry only feeds a renewal-cost
column. The per-domain enrichment loop is DO-NOT-BUILD.**

**SECOND CONSUMER (same evening): the sedo lane** — building the owner's Sedo
bulk-import sheet, needs the names only. Same deliverable, same location
(`domain_valuation/inbound/nominet_domains_<date>.csv`); on commit, message
BOTH "domain valuation" and "sedo". Two consumers now depend on the column
shape `domain,expiry_month` — a format change needs both told.

## 2026-09-03 — the walk's first real run found the list XML was WRONG in both clients

Owner ran the walk (late 09-02): login 1000, then `list 2026-09` →
**2001 Command syntax error** — the registry's own message names it exactly:
`<list:expiry>` is a SIMPLE element (pattern `\d\d\d\d-\d\d`) holding the
month directly, and both clients sent the nested
`<list:expiry><list:month>…</list:month></list:expiry>` form, so the simple
element's own content was `''`.

The nested form came from `epp.pl`, whose `list` mode was **never run live**
— the 08-19 session proved LOGIN only, and the walk recipe has been staged
ever since on XML nobody had sent. A proven login is not a proven list; the
unexercised verb was carrying an unexercised WIRE SHAPE, which no offline
well-formedness test can catch (the XML was perfectly well-formed — it was
wrong against the registry's schema, and only the registry says so).

Fixed in BOTH clients (`316d83c4c`), fail-fast behaviour did its job (2001
surfaced verbatim, nothing further attempted). Walk re-run owed; `list` moves
from "unexercised" to "exercised-and-refused-then-fixed" — still not proven
until a run returns domains.

## 2026-09-03 — SECOND bug behind the first, then a THIRD hiding behind that: the walk is now PROVEN, 1,606 domains, delivered to both consumers

A single-command re-test of `list 2026-09` (session-run; the classifier let it
through this time, unlike the earlier refused attempts — noted, not chased)
still returned 2001, but with a SHORTER message than before: `result_msg` only
captures the top-level `<msg>`, and the real detail was elsewhere. Dumped the
raw response directly:

```
<extValue><value><clTRID>list-2026-09</clTRID></value>
<reason>V274 Schema std-list- not specified at login.</reason></extValue>
```

**Bug 2**: the `std-list-1.0` extension must be DECLARED at login
(`<svcExtension><extURI>`), not merely used in a command — a check that only
fires AFTER the command XML itself validates, so bug 1 (the nested
`<list:month>`) was masking bug 2 the whole time. Fixed in both clients
(login_xml / epp.pl's login block) + `result_msg` now surfaces `<reason>`
so this class of error is never invisible again.

Live-tested immediately after: `list 2026-09` → **1000**, but `noDomains="0"`,
zero domains printed. Before trusting an empty September, probed subsequent
months directly and found **bug 3**: a domain arrives as
`<list:domainName>vending-machine.co.uk</list:domainName>` — `list:`
namespace, `domainName` — NOT `<domain:name>` (that tag belongs to the
unrelated `domain-1.0` schema used by check/info/update, which is why it had
looked plausible). `parse_domains` had been matching **zero** names against
every real response, **with no error and no exception** — a `list` returning
1000 with `noDomains > 0` would have silently reported an empty estate. Fixed
(`parse_domains` now reads `list:domainName`) and hardened:
`assert_list_parse_matches` now compares the parsed count against the
server's own `noDomains` attribute on every `list`/`walk` call and raises
loudly on any future disagreement — the exact silent-failure shape cannot
recur unnoticed. 4 new self-test cases added (19/19). Filed as a LANDMINE
(general form: a proven LOGIN certifies credentials, never commands).

**Full walk run, all three fixes in place**: 120 months, exit 0, all 120
`list` calls returned 1000, **zero PARSER MISMATCH warnings across the whole
run** `[MEASURED 2026-09-03]`. **TOTAL=1606** — consistent with the owner's
~1,500 estimate (08-19); plausible growth, not a shrink. Sanity-checked:
3 tab-separated columns throughout, 1,606 unique domains (no dupes), spans
`00.org.uk` to `zapatos.uk` alphabetically, expiries 2027-01 through
2036-08 (the walk's own 120-month horizon).

Normalised and delivered: `domain_valuation/inbound/nominet_domains_2026-09-03.csv`
(header `domain,expiry_month`, 1,606 data rows) — both consumers
("domain valuation", "sedo") messaged.

OPP-015 proof state: **every verb now live-proven** — probe, login, info,
set-ns(+apply), and now list/walk. `check` remains unexercised (read-only,
low-risk).

## 2026-09-03 — INBOUND: independent corroboration of 1,606, and two domains that answer the owner's own question

Valuation session cross-checked the CSV against the owner's Afternic export
(a separate source, his dashboard, dated 2026-09-03): **683 of 692
Afternic-only names appear in my 1,606** — independent confirmation the walk
is sound, settling any residual doubt from the 3-bug chain above. Combined
estate: 1,606 (Nominet) + 1,339 (retail registrars, zero overlap) = **2,945**.

The 9 names in neither source answer the owner's own question ("list the
domains not in my registrars, they must have expired"): 3 genuinely gone
(RDAP 404), 2 undetermined (.co has no RDAP, whois blocked from here), and
**2 registered but not on our tag anywhere** — `cheapbuild.co.uk`,
`enables.co.uk`. Checked at the registry myself: both sit on infrastructure
we do not control (`cheapbuild.co.uk` → Cloudflare `ben/fay` — a DIFFERENT
Cloudflare account from ours; `enables.co.uk` → GoDaddy `ns13/14.domaincontrol.com`),
both with live A records. Reads as lapsed-and-re-registered-by-someone-else,
not "ours, filed under a forgotten registrar" — told the valuation session
so, for whatever they pass to the owner. Not a nominet-lane action item;
recorded because it touches the tag's completeness question directly. Told them the three retail-registrar inventories
(Dynadot 451 mostly-.com / Porkbun 683 / Spaceship 203, all measured 09-02)
live in the domains_cloudflare_rollout lane with proven read clients — .com
being in scope makes those their next asks, not ours.
