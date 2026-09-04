# PLAN 2026-09-04 — the delivery email's instructions promise, and the vocabulary that lets it drift

**Lane:** `bugfix_475_delivery_email_instructions` — the MECHANISM half of `bugs_open/475`.
**Split agreed with `site_delivery_and_editor` 2026-09-04 (they hold the COPY half):** they own the
words and the owner's performed account; this lane owns the placeholder, the `LinkConfig` entry, the
page's delivery route, the zipper and the guard. They have said they will not write Go or SQL against
475 without telling this lane first.

> **⚠ AUTHORSHIP NOTE.** The owner asked for this plan to be prepared using **Fable**. The Fable
> subagent was launched with the full brief and **terminated on a usage limit** (HTTP 429,
> `claude-fable-5-1`) before returning anything. This plan is therefore **authored by this session
> (Opus)**, not by Fable. It is recorded here rather than quietly substituted, because "who wrote
> this" is exactly the kind of provenance that gets lost. Re-running it on Fable when credits allow
> is cheap and worth doing as an adversarial read of §2.

---

## 1. The bug, and the two things that have changed since it was filed

`bugs_open/475`: the customer delivery email says

> *"Your finished site as a ZIP, yours to keep: `{{zip_link}}`
>  The ZIP comes with instructions that walk you through putting it on free hosting."*

and the ZIP contains no instructions — 45 site files and no document of any kind.

**[MEASURED 2026-09-04 14:40Z] The clause is still live.** Read off the live
`delivery-email-sender` row, `default_config->'workflow'->'steps'->'send_email'->'config'->>'body_template'`.

⚠ **That read is deliberately timestamped, because migration `776` was applied to the very same
`body_template` at 12:05:25Z today** (the 477 lane's separate false-promise clause, *"so we stop
reminding you"*). A glance at "that template was migrated today" reads as though 475 were fixed.
It is not — 776 removed a different sentence in a different paragraph. My read is 2h35m *after*
776's apply, and the 477 lane independently confirmed the clause with
`body_template LIKE '%ZIP comes with instructions%'`. **Two lanes, two instruments, same answer.**

**Change 1 — there is a deadline, but it is softer than I first wrote it.** The
`site_delivery_and_editor` lane has minted voucher `WD-KN3WU-9PZN4` (£30, valid to 18 September) at
the owner's request, for another site to go through webdesign.uk. If that build reaches delivery
before this work lands, the same false statement goes out again.

> **⚠ CORRECTED same day, and the correction matters to the decision it drove.** I first wrote *"the
> same false statement goes to a second CUSTOMER"* and argued from it that the bug file's
> *"no delivery is imminent"* premise had expired. **The next build is the owner's own trial run** —
> the voucher was minted for him, so the recipient of that sentence would be *him*, not a stranger.
> I had stated a real risk at the wrong severity, and I stated it in the same voice as the measured
> facts around it. The owner has since ruled **"leave it"** with the risk in front of him (§3).
>
> **The deadline that survives is for phases 1–3: the page must exist before the next REAL customer,
> not before the next build.**

**Change 2 — there is now a second customer letter.** The 477 lane has built
`send_followup_email_action.go` and seeded it disabled (migration `775`). It carries **its own
`body_template` in its own agent row**, and it already names an instructions placeholder. So this is
no longer a one-email problem, and a fix scoped to one email would be wrong on the day it shipped.

---

## 2. THE FRAMEWORK DEFECT — and why the placeholder is the small half

The owner asked for a solution *applicable to the framework as a whole in preference to the
individual case*. Here is the general defect, established at the code today.

**Two customer letters share ONE closed placeholder vocabulary and TWO hand-maintained mirror guard
lists.**

| | `send_delivery_email_action.go` | `send_followup_email_action.go` |
|---|---|---|
| fills via | `fillTemplate` `:197-206` | **the same** `fillTemplate`, `:224` |
| guards via | its own slice, `:125-134` (4 entries) | **its own mirrored slice**, `:166-172` (5 entries) |
| backstop | post-fill `strings.Index(body,"{{")` `:168-175` | same, `:233` |

`fillTemplate` is a `strings.NewReplacer` over a **closed vocabulary**: `{{live_site}}`,
`{{confirm_link}}`, `{{zip_link}}`, `{{domain_rent_link}}`, `{{domain_buy_link}}`,
`{{stripe_portal_link}}`, `{{days}}`.

The guards exist for a specific, nasty failure: a named placeholder with no value is replaced by an
**empty string**, so the customer reads *"Your files: "* with nothing after it — and **no post-fill
scan can see it, because the fill succeeded.** That is why the guard has to run *before*
`delivery.Claim` stamps the handover. The `{{` backstop fires *after* the stamp, which is worse.

### The contract is held by a comment, and the estate has already ruled on that

`send_followup_email_action.go:150-164`, verbatim:

> *"The only other placeholder-refusal in the estate is send_delivery_email's, which this copies
> deliberately rather than shares. … **If a shared mechanism is ever built, these two are its first
> two callers.**"*

and

> *"The list is every placeholder fillTemplate knows MINUS the two this action always produces. …
> **If fillTemplate's vocabulary grows, this list must grow with it.**"*

That second sentence is a **contract enforced by a comment on a tree many sessions share** — the
exact shape of OWNER RULING 2026-08-02 §2 (RFC_010):

> *"A comment is not a control on a tree this many sessions share."*

**And we proved it this afternoon, live.** Two sessions, independently correct, working carefully,
were about to ship **two names for one concept**: this lane's `{{instructions_link}}` and the 477
lane's `{{instructions_url}}`. The only thing that caught it was one lane writing that comment and
the other thinking to send a message. **Neither of those is a mechanism.** If the 477 lane had been
half a day ahead, or had not written the comment, the estate would now have two differently-named
placeholders for one page, filled by two different mechanisms, in two letters to the same customer.

> **This is the finding that makes 475 a framework bug rather than a copy bug.** The false sentence
> is the case that found it. The defect is that the vocabulary and its guard are separable.

### 2.1 The design: make the filler and the guard the same table

Make *"the filler knows a placeholder the guard does not"* **unrepresentable**, rather than
discouraged by a comment. New file, `platform/delivery/vocabulary.go`:

> **⚠ CORRECTED 2026-09-04, before any code was written, by the 477 lane — and this is the
> correction that would have cost the most.** My first sketch of this table was shaped
> "placeholder → link". **The seven tokens are not all links and do not all resolve the same way.**
> There are three kinds:
> - **from `Prepared.Links`**, produced by the claim — `{{live_site}}`, `{{confirm_link}}`,
>   `{{zip_link}}`, the three payment links;
> - **from step CONFIG**, known before anything irreversible, which is exactly *why* the guard can
>   run pre-claim — the payment URLs and `instructions_url`;
> - **neither** — `{{days}}` is `AdvertisedWindowDays`, a compile-time constant, **and it is not a
>   link at all**. Its value is deliberately *not* the obvious one: we advertise **30** while
>   `LiveLinkWindow` serves **42**, on purpose, and `handover.go` says never to derive it from the
>   expiry.
>
> A "placeholder → link" table either forces `{{days}}` in or quietly leaves it out, **and left out
> is the dangerous one**, because it is precisely the entry whose correct value is not what a reader
> would guess. The design below survives this — `Availability.Value` is a resolved string and is
> agnostic about provenance — but the *strict coverage check is what actually saves it*: a sender
> cannot omit `{{days}}`, because omitting any vocabulary token is itself the refusal. **Recorded
> because the shape I nearly built was wrong for a reason no test of mine would have caught.**

```go
// Token is one placeholder a customer letter may name. The set is CLOSED: a
// template author cannot invent one, and the estate cannot grow one without
// every sender saying what it does with it (see Fill.Check).
//
// NOT all links, and not one provenance — see the three kinds above. The type
// is deliberately named for what it IS (a token in a letter) rather than for
// what most of them happen to BE (a URL).
type Token string

const (
    TokenLiveSite      Token = "{{live_site}}"
    TokenConfirmLink   Token = "{{confirm_link}}"
    TokenZipLink       Token = "{{zip_link}}"
    TokenInstructions  Token = "{{instructions_link}}"   // NEW — bugs_open/475
    TokenDomainRent    Token = "{{domain_rent_link}}"
    TokenDomainBuy     Token = "{{domain_buy_link}}"
    TokenStripePortal  Token = "{{stripe_portal_link}}"
    TokenDays          Token = "{{days}}"
)

// Vocabulary is the SINGLE SOURCE. Both the fill and every sender's pre-claim
// guard are derived from it, so they cannot disagree.
var Vocabulary = []Token{ /* all of the above */ }

// Availability is what ONE SENDER can say about ONE token on THIS dispatch.
type Availability struct {
    Value       string // resolved value; "" means this dispatch has none
    Source      string // where it should have come from, for the refusal text
    NeverReason string // non-empty: this sender can NEVER produce it, and why
}

// Fill is a sender's complete declaration over the Vocabulary.
type Fill map[Token]Availability

// Check refuses BEFORE anything irreversible. Two refusals, in this order:
//
//  1. COVERAGE. Every Vocabulary token must have a declaration. This is the
//     drift guard: adding a token to the Vocabulary makes every sender that
//     has not declared it refuse LOUDLY, before the stamp, instead of mailing
//     an empty string. It does not depend on the template naming the token.
//
//  2. AVAILABILITY. If the template names a token whose Value is empty (or
//     whose NeverReason is set), refuse and say which and why.
func (f Fill) Check(template string) error

// Apply substitutes. Derived from the SAME map Check read, which is the whole
// point: there is no second list to fall out of step.
func (f Fill) Apply(template string) string
```

**Why coverage is checked unconditionally rather than only when a template names the token.** The
weaker version — error only if the template names an undeclared token — is still a real improvement,
and it fires before the stamp. But it leaves the estate in a state where a sender is *already* wrong
and nothing says so until somebody edits a template, possibly months later and in a different lane.
The strict version converts a latent drift into an immediate, pre-stamp, named error the day the
vocabulary grows. Deliveries are rare and operator-driven (one ever), the refusal is *before* the
stamp and therefore fully recoverable, and CI catches it long before production (§2.2).

**How the per-caller difference survives, because it is real and must not be flattened.** The two
senders genuinely differ on `{{zip_link}}`:

- the delivery sender refuses it **only when empty** — it *can* carry a presign;
- the follow-up sender refuses it **always**, because *"a scheduled follow-up has no step to mint a
  presign"*.

That is a per-caller fact, not duplication, and `NeverReason` carries it as a first-class state
rather than as an empty string that happens to fail. A sender that can never produce a token says so
once, in its own declaration, with its reason — and the refusal message tells the operator the truth
instead of "value was empty".

**The 477 lane sharpened this and the sharpening is load-bearing:** theirs is not *"refuse when
empty"* but *"this caller can never produce this one, at any time, **by construction**"*. A
`{{zip_link}}` in a follow-up template is an **author error to be caught at dispatch**, not a missing
value that might one day be supplied. Flattening it to "empty ⇒ refuse" would behave identically
today **and would lose the reason — and the reason is the only thing stopping a later session
"fixing" it by wiring a presign into the scheduled follow-up.** `NeverReason` exists to carry that
sentence to whoever tries.

### 2.2 The test that makes drift fail in CI, not in a customer's inbox

A shared helper both senders' tests call:

```go
// AssertCoversVocabulary fails if a sender's Fill omits or invents a token.
// Both senders call it. Adding a Token without teaching a sender turns RED here.
func AssertCoversVocabulary(t *testing.T, f Fill)
```

Plus one test per sender asserting the refusal *text* names the token and the reason — the estate's
own lesson that a receipt nobody asserts on is a log line.

⚠ **Mutation check owed before claiming this works** (MEMORY: *a mock's own bookkeeping cannot
assert a negative*; *mutate the code to prove the guard*): add a token to `Vocabulary`, do **not**
teach the senders, and confirm both test suites go red and both `Check` calls refuse. A guard that
passes because something *else* in series refused proves nothing.

### 2.3 What this does NOT cover, stated so it is not mistaken for covered

**The guard checks PLACEHOLDERS. It cannot check PROSE.** The sentence *"The ZIP comes with
instructions"* carries no placeholder, so it was invisible to the existing guard and will be
invisible to this one. Nothing here detects a false English claim.

What the design does instead is **remove the reason to write one**: the customer-facing thing becomes
a link, links are placeholders, and placeholders are checked. That is a narrowing of the class, not
an elimination of it, and §6 says so plainly rather than letting a future reader infer more.

---

## 3. Phasing, and the ordering rule that is easy to get backwards

**The ordering constraint is not preference — it is a live failure mode.** Copy is DB config and goes
live the moment a migration applies; Go is inert until an image is built and rolled. So:

> ⚠ **THE TEMPLATE MUST NOT NAME `{{instructions_link}}` UNTIL THE BINARY THAT KNOWS IT HAS
> ROLLED.** A template naming a token the running binary's vocabulary lacks leaves the literal
> `{{instructions_link}}` in the body, which trips the post-fill `{{` scan — **and that scan fires
> AFTER `delivery.Claim` has stamped the handover.** The result is a stamped, undeliverable
> handover needing the operator re-mint recipe. This is CLAUDE.md's "image first, then seeds" rule,
> and here the cost of getting it backwards is a customer's delivery wedged mid-claim.

| phase | what | needs a roll? | gate |
|---|---|---|---|
| ~~**0**~~ | ~~Honest interim: the false clause out of the live template~~ **CANCELLED — see below** | — | — |
| **1** | The shared vocabulary + derived guard (§2), no new token | Yes | council gate |
| **2** | `{{instructions_link}}` into the Vocabulary + `LinkConfig`/`Links`; the 477 lane drops its pre-substitution | Yes, **same roll as 1** | council gate |
| **3** | The instructions page, through the framework | No Go | owner's content + screenshots |
| **4** | Migration naming `{{instructions_link}}` in **both** templates | No — but **only after phase 2 has rolled** | verify the stamp by ancestry first |
| **5** | `README.txt` in the ZIP | Yes | council gate |

### Phase 0 — proposed, put to the owner, and RULED AGAINST. It is cancelled.

> **⚠ CORRECTION 2026-09-04, recorded rather than edited away, because the reasoning changed twice.**
>
> **What I argued:** the bug file ruled candidate 4 ("delete the sentence") out because *"no delivery
> is imminent"*. A voucher had since been minted, so I judged that premise expired and asked the copy
> lane to press the owner for one of the two interim wordings they had offered him.
>
> **THE OWNER RULED "LEAVE IT."** No stop-gap. **The false line stays until the page exists.**
>
> **And one fact I did not have makes his ruling sounder than my framing allowed for: the next build
> is his own trial run.** Voucher `WD-KN3WU-9PZN4` was minted for *him*. So the recipient of that
> sentence would be the owner, not a stranger. **My deadline was real but it was not a stranger's
> first impression, and I had stated it as though it were.** The copy lane put the risk to him in
> plain terms — that a second customer could be told the ZIP contains instructions it does not — and
> he chose to wait with that in front of him.
>
> **This is a better outcome for the work than the one I asked for**, and worth noticing why:
> there is now **no competing change queued against the `body_template` jsonb path**, so phase 4 has
> a clear run at it. The thing I wanted to add would have been a third migration against one field in
> two days.
>
> **What survives from the argument:** the deadline is still real for phases 1–3. The page has to
> exist before the next real customer, not before the next build.

---

## 4. The three exits

### 4.1 The page — build it through the framework, on `webdesign.uk` — **OWNER-RULED, 2026-09-04**

> ✅ **RULED. The page is GENERIC, on `webdesign.uk`, built by the framework.** The copy lane put
> this proposal to the owner and he ruled for it, including the consequence below: the per-customer
> content still exists, it lives in the email and the README rather than on the page. The reasoning
> is kept in full because a ruling without its argument cannot be re-derived when it is challenged.

`[MEASURED 2026-09-04]` `webdesign.uk` is itself a framework site — site
`1fcfa4f3-ec80-4010-878b-b971cd46711f`, **18 pages**, including deployed `/guides/*.html` pages built
through the pipeline. So the capability already exists and has been exercised.

⚠ **And it was verified at the artefact, not off the DB row, by the copy lane rather than by me** —
`/guides/tool-css-variables-guide.html` serves **200**, with an **invented `/guides/` path 404ing as
the control**. That control is the whole difference between "the route is real" and "a parked host
200s every path" (MEMORY: *a parked domain 200s every path*). My own claim rested on a page count in
a table; theirs rests on the served bytes with a disconfirming case. **Theirs is the evidence.**

Five reasons, in the order that decides it:

1. **OWNER RULING 2026-08-04 — "EVERY SITE GOES THROUGH THE FRAMEWORK. Never hand-build one."** No
   hand-authored HTML, however small or temporary. `/c/` and `/d/` are Go-rendered, but they are
   *transactional endpoints*; a prose-and-screenshots instructions page is *content*. Hand-building
   it would need a deliberate exception, and the copy lane flagged exactly this. **The proposal
   spends no exception at all.** It is also the positive half of that ruling: a hand-built page
   demonstrates nothing on a site selling framework-built sites.
2. **It must be durable — this is a hard constraint from the 477 lane, not a preference.** Their
   scheduled follow-up sender refuses any placeholder it cannot fill and **has no step that can mint
   a token**. A per-delivery token with a lifetime would be either uncarriable by the follow-up or
   would start refusing when it expired. **This kills the token-addressed option on evidence.**
   Corroborating: `send_followup_email_action.go:48`'s own example value is already
   `https://webdesign.uk/your-site` — the 477 lane had independently assumed the same home.
3. **Correctable**, which is what the owner's "correctable page" requirement actually wants — a
   rebuild. A free host changing its signup becomes one edit for everyone, including customers
   mailed last month.
4. **Bookmarkable and public**, which a customer returning in a year needs.
5. It is the same URL for everybody, so **the follow-up letter and the delivery letter can carry the
   same link** — the drift this whole plan is about, closed at the content level too.

**The consequence, and the one thing needing the copy lane's read:** a framework page is **generic**,
so a customer's own domain cannot be a slot on it. My reading is that this is not a loss but a
better fit for the owner's own rot rule — *anything that can go out of date lives on the page* —
because **a customer's own domain is the one thing in the whole set that never goes out of date**,
and it is already in both the email and their folder. So the per-site slots live in the email and the
`README.txt`; the generic page carries the rot-prone content. It also removes the copy lane's own
objection to a public page: a generic page **cannot** name a customer's domain at a guessable address.

This does **not** re-open the owner's "all three" ruling. All three exits still exist. What is being
proposed is *where the third one lives*. If the copy lane reads his ruling as requiring a
per-customer page, that is their call — they hold his account.

**The screenshots — answered by the copy lane, and it comes with a landmine aimed exactly at me.**

The owner's captures of the Netlify flow are at `/home/ant/Downloads/idea_uk_netlify/`. They are
photographs of a third-party signup, not generated imagery, so they do **not** come from the normal
imagery route — supplied images enter through **`deploy_image_asset`**, which takes an explicit
`s3_uri`.

> ⚠ **`deploy_image_asset` resolves its source by PURPOSE, not by the `asset_id` you pass it — so the
> second same-purpose asset on a site silently deploys as the FIRST one.** There are **ten**
> screenshots and they will almost certainly share a purpose, which is this landmine's worked case
> rather than a near miss.
>
> **To deploy correctly:** supply `spec.s3_uri` explicitly as a genuine `s3://bucket/key`, derived
> from that asset's own **`storage_path`** — **not** its `url` column, which may be a stale presigned
> link or a local path post-deploy (`bugs_open/152`). A non-empty `s3_uri` is consulted *before* the
> buggy site-wide cache.
>
> **And verify at the artefact: `sha256sum` the deployed files and confirm they DIFFER, then open at
> least one.** Do not stop at `success: true` — **ten identical images all reporting success is
> exactly what this bug looks like.** This is the estate's *trust the artefact, not the status* rule
> with a specific, checkable instrument.

**The most valuable image on the page is the signed-out *"This site is private"* wall** — it is the
thing a customer will otherwise never see, and the reason v3 exists at all.

### 4.2 The `README.txt` in the ZIP — synthesise at zip time

⚠ **THE LANDMINE, and it is a one-line trap.** `platform/orchestration/actions/zip_deliverable_action.go`
`composeZip` (`:196`, write loop `:209-231`) iterates **exactly** the S3 listing of
`portfolio-sites/<domain>/`. There is no synthesised entry anywhere in the estate. And
`verifyArchive` (`:250`) asserts at **`:259-261`**:

```go
len(zr.File) == len(files)
```

**Adding a README without teaching that assertion makes the action FAIL.**

**Do not loosen the assertion — make it more precise.** It should assert the synthesised entries **by
name**, not merely admit a bigger number: `len(zr.File) == len(files) + len(synthesised)` *plus* a
check that each synthesised name is present. A count-only relaxation would let a *missing* README and
an *extra* site file cancel out, which is precisely the class of blindness this estate keeps logging.

**Why synthesise at zip time rather than publish the README into the bucket** (the obvious cheaper
route, and it is wrong): objects under `portfolio-sites/<domain>/` are **the served site**. A README
published there would be publicly fetchable at `https://<customer-domain>/README.txt`, crawlable and
indexable, visible to the customer's own visitors, and would enter the population every site-level
check sweeps. It is a private note to the buyer, not site content. It must never be on the site.

**Content rule, from the owner's ruling and not negotiable:** the README carries **what the folder is,
and the page's address** — nothing that can go out of date. No hosting steps, no prices, no free-host
names. A ZIP cannot be edited once a customer has it.

### 4.3 Per-site content

Lives in the email and the `README.txt`, both of which already know the customer's domain. See §4.1.

---

## 5. What I would NOT do, and why

- **Add `{{instructions_link}}` to `fillTemplate` and a fifth line to each of the two guard slices.**
  This is the obvious fix, it is what the bug file's candidate 1 describes, and it works. It is
  rejected because it *grows* the exact structure that is the defect: it leaves two hand-kept mirrors
  and adds a fourth caller-hand-edit to the list of things a future session must remember. It also
  fails the estate's own ranking rule — order candidates by what makes the bad state
  **unrepresentable**, and "the next session must remember to edit both lists" is a defect, not a fix.
- **Delete the sentence and promise nothing.** Removes the false claim and leaves the customer unable
  to do the thing the next paragraph tells them to do. Correct **only** as the phase-0 stop-gap, and
  only because a delivery may now be imminent.
- **Put the instructions inline in the email.** Cheapest, nothing to rot — but it cannot be corrected
  after sending, it lengthens an already-long email, and it is the version a customer is least likely
  to still have when they need it. It also directly contradicts the owner's rot rule, since free
  hosts change their signup flows.
- **Collapse `tokenURL` (`prepare.go:329-332`) and `ConfirmTokenURL` (`handover.go:509-533`).** These
  are **two builders for one customer `/c/` link** — genuinely the same "two mirrors of one concept"
  class as §2, and flagged by three council seats. **Scoped OUT deliberately.** `handover.go` is the
  477 lane's file, they are mid-flight in it (migration 778, `StampHandover`), and the code's own
  comment says the reason they are not yet collapsed is that *"a same-file edit from two lanes is how
  one lane's work gets lost."* Bundling it would demonstrate the failure it fixes. **Referred to the
  477 lane** as their call, with this lane's shared-vocabulary work offered as the precedent.
- **Build a generic "prose claims must be backed by an artefact" checker.** Tempting, and it is the
  bug file's §4 framing. Rejected as designed-for-a-symptom: detecting false English claims is an LLM
  problem with no crisp failure, and the estate already has a large `evidence_base`/banned-claims
  layer aimed at *site copy* that has never covered agent step config. Narrowing the class by making
  customer-facing facts into checked placeholders is the mechanical half, and it is the half that
  actually closes doors. §6 records the residue honestly rather than claiming it is covered.

---

## 6. Risks, blast radius, and what stays open

**Blast radius of §2:** two actions, one new file, no schema change, no config change. Both callers
are delivery-path and neither has ever run at volume (`[MEASURED 2026-09-04]` one delivery ever;
the follow-up sender is seeded **disabled**). The strict coverage check *can* refuse a send — by
design, before the stamp, recoverable.

**Architecture scope.** Under OWNER RULING 2026-07-29 §1, an addition to a shared vocabulary needs an
RFC only when it changes what the shared mechanism **guarantees**. This one *strengthens* the
guarantee (from "a comment asks you to keep two lists in step" to "a sender that has not declared a
token refuses before the stamp") and adds no new authority. Assessment: **normal council gate, not an
RFC.** But the gate must see it, and the submission must name the 477 lane as the other consumer —
OWNER RULING 2026-07-29 §3: *a shared mechanism's other consumers must be TOLD, not merely measured.*
**They have been told, before the submission, and asked for their objection in the round.**

**The honest residue, stated so nobody later reads this plan as covering it:**
- Nothing here detects a false **prose** claim. §2.3.
- The `{{` post-fill backstop still fires **after** the stamp. It is a backstop, not the guard.
- `AdvertisedWindowDays` is 30 while the presign is 7 days and tokens run 42 — the three-lifetime
  finding. **Not this bug**, tracked by the copy lane, and it is the *other* live mismatch between
  what the email says and what the system does.
- `Links.DomainBuy`'s doc comment (`prepare.go:179`) still reads `£200 one-off payment link`.
  Migration 726 changed the price to £59.99 on 2026-09-03. **A stale comment, harmless today**
  because the value comes from config — but it is the same "the copy and the thing disagree" family,
  and it is one line. Fold it into phase 1.

---

## 7. Naming, and two obligations this lane owes other people

### 7.1 The placeholder is `{{instructions_link}}`; its config key stays `instructions_url`

Agreed with the 477 lane, who have already converged (`0949244e8`) — their action, seed and docs.

The reason is stronger than "five of six end `_link`", and it settles a second question I had not
thought to ask. **The estate pairs a `*_url` CONFIG KEY with a `*_link` PLACEHOLDER, three times, in
this lane's own file:**

```
send_delivery_email_action.go:127  {"{{domain_rent_link}}",   stringOr(config, "domain_rent_url"),   …}
                            :128  {"{{domain_buy_link}}",    stringOr(config, "domain_buy_url"),    …}
                            :129  {"{{stripe_portal_link}}", stringOr(config, "stripe_portal_url"), …}
```

So the placeholder is `{{instructions_link}}` and **the config key stays `instructions_url`**.
Renaming the key too would have made it the estate's first `*_link` config key. **Recorded because
"the placeholder and its key have different suffixes" reads like a slip to the next reader** — it is
the convention, and the shared table's entry must carry a comment saying so.

**Why the rename was free, and the window that closes it:** `[VERIFIED by the 477 lane 2026-09-04,
not assumed]` migration `775` is seeded and **NOT applied** — `scheduled_tasks` has **0** rows for it
and `sites.followup_sent_at` **does not exist**. So the rename was a file edit. **After `775` is
applied the identical change costs a migration against the live agent row.**

### 7.2 Two obligations this lane owes, both easy to drop

1. **Delete the 477 lane's coupling landmine in the same commit that lands the shared table.** They
   filed it footprinted on `fillTemplate` and both action files, with the cross-reference at
   `fillTemplate` itself (`d67f08ff4`), precisely because a comment in *their* file protects nobody
   reading *this* one. If the table lands, that trap no longer exists — and **a landmine describing a
   fixed problem is worse than none, because it spends the next reader's attention on nothing.**
2. **Send the 477 lane the council submission before committing.** They have asked to object in-round
   rather than after, and the shared-vocabulary change is their file as much as this lane's. This is
   also OWNER RULING 2026-07-29 §3 — *a shared mechanism's other consumers must be TOLD, not merely
   measured.*

---

## 8. Open questions

**Closed since this plan was first written** (all four by the two peer lanes, within the hour):
~~is v3 stable to build against~~ **yes**, with `{{domain_paragraph}}` now off the page per the
generic-page ruling · ~~where do the screenshots live~~ **answered, §4.1** · ~~the interim wording~~
**ruled: leave it, §3** · ~~does "all three" require a per-customer page~~ **no, ruled, §4.1** ·
~~does `NeverReason` match the 477 lane's intent~~ **yes, and they sharpened it, §2.1**.

**Still open:**

1. **`{{live_until_date}}` on the page — DO NOT WIRE IT YET.** The copy lane flags three candidate
   dates that disagree: the presign's **7 days**, the email's advertised **30**, and the tokens'
   **42**. This is the three-lifetime finding surfacing in a second place. Somebody must settle which
   date a customer is actually owed before any artefact states one. **Wiring it to whichever is
   nearest to hand would be this bug's own root cause committed a third time.**
2. **The ten screenshots have not been deployed or verified.** Until `sha256sum` shows ten *different*
   files served, the page cannot be called done (§4.1).
3. **This lane:** the standing warning *"instructions nobody has followed are a guess with
   formatting"* is **discharged for §4 of the draft** — the owner performed the Netlify flow himself
   on 2026-09-04 and v3 records what he saw. It stays live for anything added afterwards.
