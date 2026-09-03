# Which domains are NOT in your registrar lists — and are they expired?

Owner's question, 2026-09-03: *"please list the domains that are not listed in
my registrars, they must have expired"*.

**Short answer: almost none of them expired. Of 692 candidates, 683 were simply
sitting at Nominet, whose list did not exist when the question was asked. Nine
were genuinely unaccounted for, and only three of those are actually gone.**

## How the candidates were found

Every domain appearing in a marketplace/listing source (Afternic export
2026-09-03, Spaceship SellerHub, Dynadot marketplace) that appears in NO
registrar inventory. Before the Nominet walk landed that was **692 domains**.
The Nominet list (1,606 domains, delivered 2026-09-03) accounts for **683** of
them. Full estate: **2,945** domains (1,339 retail + 1,606 Nominet, zero overlap).

## Method — and why DNS could not answer this

Registration status came from **RDAP, the registry's own record**, not DNS. A
domain registered but never delegated answers NXDOMAIN, which is
indistinguishable by DNS from one that lapsed — and this estate has many
never-delegated names (owner, 2026-08-19: *"No nameserver usually means I never
set a nameserver"*). Controls both directions on every run: a name nobody owns
must return 404 and a known-owned name must return 200, else the run aborts.
Nominet throttling produced connection failures which were **retried, never
recorded as answers** — an unretried failure would have invented an expired
domain.

## The nine, individually checked

| domain | registry says | reading |
|---|---|---|
| chicklets.co.uk | **404 — registered to nobody** | EXPIRED and dropped |
| demisexual.uk | **404 — registered to nobody** | EXPIRED and dropped |
| protecty.co.uk | **404 — registered to nobody** | EXPIRED and dropped |
| cheapbuild.co.uk | live, exp 2027-03-30, registrar **Voove Limited** | registered, NOT on the DESIGNCONSULT tag |
| enables.co.uk | live, exp 2027-03-09, registrar **123-Reg** | registered, NOT on the DESIGNCONSULT tag |
| pocketvaginas.com | live, exp 2027-06-09, registrar **Dynadot** | at Dynadot but absent from the Dynadot inventory |
| qlp.us | live, exp 2027-01-19, registrar **NameSilo** | a registrar we hold no account for |
| healthinsuranceconsultant.co | **UNDETERMINED** | .co publishes no RDAP service; whois egress blocked here |
| studentloandebtsettlement.co | **UNDETERMINED** | same |

## What needs the owner's eye

1. **Three names are gone** and are still listed for sale at Afternic. Those
   listings should come down — they advertise domains that cannot be delivered.
2. **Two are registered elsewhere** (Voove, 123-Reg). Either they were allowed
   to lapse and were caught by someone else, or they are held in an account
   nobody has told us about. Worth a look: did you ever use either registrar?
3. **Two sit at registrars outside the four we enumerate** — one at Dynadot but
   missing from Dynadot's own account listing (a second Dynadot account?), one
   at NameSilo. **If a NameSilo account exists, it holds domains this valuation
   cannot see.**
4. **Two .co names cannot be checked from here.** Low value, but they can be
   resolved by looking them up in a browser if it matters.

## Marked

- [MEASURED 2026-09-03] every registry status above, individually re-queried.
- [MEASURED 2026-09-03] 683-of-692 reconciliation against the Nominet CSV.
- [UNDETERMINED] the two .co names — stated, not guessed.
- The tag check reads the RDAP registrar field; on Nominet that resolves to the
  owner's own name via the DESIGNCONSULT tag, which is what makes "not his"
  visible at all.
