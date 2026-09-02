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
