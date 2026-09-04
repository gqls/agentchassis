# 475 — the delivery email tells the customer the ZIP contains instructions, and it contains none

Filed 2026-09-03 by `site_delivery_and_editor`, from the **first delivery email ever sent**
(owner-authorised rehearsal, `idea.uk` → `aaa@designconsultancy.co.uk`, 19:30:31Z). Found by the
owner reading it as a customer would, which is the one test nothing else in the estate performs.

## The false claim, verbatim

`delivery-email-sender`'s live `send_email.config.body_template`, under **YOUR FILES**:

> Your finished site as a ZIP, yours to keep:
> {{zip_link}}
> **The ZIP comes with instructions that walk you through putting it on free hosting.**

## What the ZIP actually contains

Downloaded from the delivered presigned link and listed:

```
45 files:  31 .jpg   10 .html   3 .js   1 .css
top level: index.html about.html contact.html privacy.html terms.html
           refund-policy.html report.html tools.html
           assets/ (33)  tools/ (2)  news/  guides/
```

**No readme. No instructions. No .md, .txt or .pdf of any kind.** The site, and only the site.

> ⚠ A grep for `readme|instruction|guide|how|start|help` returns three apparent hits —
> `assets/images/hero-guides.jpg`, `assets/images/icon-guides.jpg`, `guides/index.html`. All three
> are the SITE'S OWN "guides" section, not documentation. A needle that matches the subject matter
> as well as the artefact type will find the subject matter. Count by extension instead: there are
> no document files at all.

## Why this is worse than a missing file

1. **It is a false statement to a customer**, in an email we send on their purchase.
2. It sits on the load-bearing line. The email's whole "KEEPING IT ONLINE" story is *"you host it
   yourself. Free hosting works well."* — and the promised instructions are the only thing that
   would tell them how. Remove the promise and that section says nothing actionable; keep it and it
   points at something that does not exist.
3. **The failure mode is the customer blaming the download.** They are told the instructions are in
   the ZIP. They open it, find 45 site files, and reasonably conclude the ZIP is broken or truncated
   — so the first thing this defect produces is a support conversation about a file that is fine.
4. It is invisible to every check we have. The chain reports `sent: true, zip_link: true`; the zip
   verifies as 45 files and a good tree hash; the link fetches 206. **Nothing in the system compares
   what the email SAYS to what the artefact CONTAINS.** This is the estate's "a promise in copy must
   be honoured by something that produces it" shape, aimed at a customer rather than an agent.

## Owner's ruling, 2026-09-03

> *"the zip works but doesn't have the instructions in it and I think they should probably be
> separate anyhow"*

So the instructions are **wanted**, and his lean is that they live **outside** the ZIP. That lean is
not yet a decision about where — see the open question below.

## What is already in our favour

**The email copy is CONFIG, not code** — `send_delivery_email_action.go:8` states it as a property:
*"No copy. The template lives in the STEP CONFIG (DB, owner-editable)"*, and the action refuses to
run without `subject`, `body_template` and `links_host`. So fixing the words needs **no image, no
roll, no release** — a migration, live on apply.

The action also already enforces a related discipline worth not breaking: it **refuses to send** if
the template names a `{{placeholder}}` whose link this claim did not produce, rather than mailing a
blank (`send_delivery_email_action.go:131`, `:173`). That is the same class of guard this bug wants
for prose, and it is why any new instructions LINK must be a real placeholder rather than a hardcoded
URL — a hardcoded one cannot be checked and will rot silently.

## Fix candidates, ordered by what closes the door

1. **Instructions as their own page on the links host, named by a placeholder** —
   `{{instructions_link}}`, filled from `LinkConfig` like the others, so the existing
   refuse-to-send-on-an-empty-link guard covers it automatically. Matches the owner's lean, can be
   corrected after a customer already has the email, and cannot silently rot because the guard fires
   if the link is not produced. **Cost:** a page must exist and be served; `links.webdesign.uk`
   already serves `/c/` and `/d/`, so the seam exists.
2. **Both: a short `README.txt` in the ZIP *and* the page.** The ZIP outlives the email — a customer
   returning in a year has files and no message — so the file is the durable copy and the page is
   the correctable one. Costs a change to the zipper as well.
3. **Instructions inline in the email**, replacing the promise with the content. Cheapest, no new
   surface, nothing to rot — but it makes an already-long email longer, cannot be corrected after
   sending, and is the version a customer is least likely to still have when they need it.
4. **Delete the sentence and promise nothing.** Not a fix — it removes the false claim and leaves the
   customer with no way to do the thing the next paragraph tells them to do. Acceptable only as an
   immediate stop-gap if a delivery were imminent, and one is not: delivery is HELD on boxingonline.

~~**Open question for the owner**~~ **ANSWERED 2026-09-03: "all three I think"** — the page, the
pointer in the ZIP, and per-site content naming their domain.

**That is one body of copy with three exits, not three documents**, and treating it as three is how
they drift apart:

| exit | carries | why this one |
|---|---|---|
| the page | everything, slots filled for that customer | correctable after the email has gone. A free host changing its signup becomes one edit for everyone, including people mailed last month |
| the ZIP's `README.txt` | what the folder is, and the page's address | the ZIP outlives the email. Somebody who keeps their files and returns in a year has the folder and no message |
| per-site | the `{{slots}}` only | the slots are the only part that differs between customers |

**The rule that keeps them honest: anything that can go out of date lives on the page, never in the
README.** Free hosts change their signup flows, our prices change, and a ZIP cannot be edited once a
customer has it.

Draft copy for all three, written to the house plain style and marked with its slots:
`docs024_key_docs_latest/site_delivery_and_editor/DRAFT_2026-09-03_customer_instructions_copy.md`.
It names five things it does NOT settle, and one of them is load-bearing:

> ⚠ **NOBODY HAS EVER PERFORMED THESE INSTRUCTIONS.** The hosting section says to drag the unzipped
> folder onto `app.netlify.com/drop`. That is written from knowledge, not from having done it today
> with this actual ZIP, and free tiers change their signup without notice. **Instructions nobody has
> followed are a guess with formatting** — and shipping them would repeat this bug's own root cause
> one level along: telling a customer something we have not checked. Somebody must do it once, end to
> end, with this ZIP, and write down what they actually saw.

## How to verify the fix

Send a delivery to a test address and check the two halves against each other rather than separately:
every artefact the email NAMES must exist — if it says the ZIP contains something, list the ZIP and
find it; if it names a link, fetch the link. Today's negative control is this delivery: the email
promised instructions in the ZIP, `sent: true` was reported, and the ZIP had none.

## Related

- `bugs_open/474` — the same delivery, the same first-real-use discovery pattern.
- The three-lifetime finding in `RUNBOOK_site_delivery_and_editor.md` (presign 7 days / email says
  30 / tokens 42) — also a mismatch between what the email says and what the system does.
- `send_delivery_email_action.go:131,173` — the existing guard that refuses to send an email naming a
  link it did not produce. Candidate 1 inherits it; a hardcoded URL would not.

---

## UPDATE 2026-09-04 — the mechanism half, from the `bugfix_475_delivery_email_instructions` lane

Split agreed with `site_delivery_and_editor` (who filed this and hold the COPY half): they own the
words and the owner's performed account; this lane owns the placeholder, the page's delivery route,
the zipper and the guard. Working docs:
`docs/agent_docs/docs024_key_docs_latest/bugfix_475_delivery_email_instructions/`.

### Still valid, and the check needs a timestamp to be worth anything

`[MEASURED 2026-09-04 14:40Z]` the live `delivery-email-sender` row still carries the false clause.

⚠ **Migration `776` was applied to the SAME `body_template` jsonb path at 12:05:25Z the same day**,
for `bugs_open/477`'s unrelated *"so we stop reminding you"* clause. So **"that template was migrated
today" is not evidence either way**, and a glance at the commit log reads as though this were fixed.
Two lanes confirmed the clause independently, one by reading the template, one by
`body_template LIKE '%ZIP comes with instructions%'`.

### THE BUG IS BIGGER THAN ITS SENTENCE — and this is the part worth carrying forward

§4 above says *"Nothing in the system compares what the email SAYS to what the artefact CONTAINS."*
True, and there is a **second, mechanical** defect underneath it that the fix candidates did not see:

**There are now TWO customer letters sharing ONE closed placeholder vocabulary and TWO hand-kept
mirror guard lists.** `send_followup_email_action.go` (the 477 lane's, seeded disabled by `775`)
calls the *same* `fillTemplate` as the delivery sender and keeps its *own* copy of the guard. Its
comment says so: *"If a shared mechanism is ever built, these two are its first two callers"* and
*"If fillTemplate's vocabulary grows, this list must grow with it."*

That is a contract held by a comment — OWNER RULING 2026-08-02 §2: *"A comment is not a control on a
tree this many sessions share."*

**And it was demonstrated live on 2026-09-04, not argued:** two lanes independently chose **two names
for one concept** — `{{instructions_link}}` and `{{instructions_url}}` — caught only because one lane
had written that comment and the other sent a message. Half a day's difference in timing and the
estate would carry two placeholders for one page, in two letters, to one customer.

**So candidate 1 above ("a placeholder, inheriting the existing guard") is right about the artefact
and wrong about the fix**: adding a fifth line to each of two hand-kept lists grows the defect. The
work in flight single-sources the vocabulary and DERIVES the guard from it, landing
`{{instructions_link}}` as its first single-sourced entry. Council trail
`c8ed56d2-74ea-4bcc-a0a4-73050c436693` (round 1 REVISE, round 2 in flight).

### A SECOND TRAP THIS BUG'S OWN FIX WOULD HAVE SPRUNG — now filed in `LANDMINES.md`

§"What is already in our favour" above says the copy is config and *"fixing the words needs no image,
no roll, no release"*. **True of a word. FALSE the moment the edit adds a `{{placeholder}}`**, because
the token's other half is compiled into the binary.

`fillTemplate` is a closed-vocabulary replacer, so an unknown token is **not substituted** — it
survives, trips the post-fill `strings.Index(body,"{{")` scan, **and that scan runs AFTER the claim**
(`:154` claims, `:162` fills, `:168` scans; the error text itself says *"the handover is now
stamped"*). Result: a stamped, undeliverable handover needing the operator re-mint recipe.

⚠ **Worse in the follow-up sender**, which is the one being enabled next: its scan runs after
`ClaimFollowup` has stamped `followup_sent_at` — **the customer's single follow-up is consumed with
no email sent.** Established by the 477 lane checking their own code against this ordering.

**So: no migration naming `{{instructions_link}}` until a chassis image carrying it has rolled**,
proved by `service_binary_capabilities` + `git merge-base --is-ancestor`. ⚠ And name **the commit that
taught the binary the TOKEN**, not the one that added the action — a binary between the two passes a
naive ancestry check and does not know the token.

The round-2 design closes this by construction: the pre-claim check now also refuses any `{{…}}`
absent from the vocabulary, moving that failure **before** the stamp.

### Owner rulings recorded 2026-09-04 (via the copy lane)

1. **The page is GENERIC, on `webdesign.uk`, BUILT BY THE FRAMEWORK.** Not a Go-rendered page on the
   links host — so no exception is spent against the 2026-08-04 "every site goes through the
   framework" ruling. The per-site content still exists; it lives in the email and the README rather
   than on the page. Verified at the served bytes with a control (a real guide page 200s, an invented
   `/guides/` path 404s).
   ⚠ The link must be **durable — no token, no lifetime.** Hard constraint, not preference: the
   scheduled follow-up sender refuses any placeholder it cannot fill and has **no step that can mint a
   token**.
2. **No interim wording — "leave it"** until the page exists. The bug file's candidate 4 was
   re-proposed by this lane on the grounds that a voucher was out; **the owner ruled against it with
   the risk in front of him**, and the next build is his own trial run rather than a stranger's.
3. **`{{live_until_date}}` must NOT be wired** — three candidate dates disagree (presign 7 / email 30
   / tokens 42). Stating whichever is nearest to hand would commit this bug's own root cause a third
   time.

### Notes for whoever builds the ZIP's `README.txt`

⚠ `zip_deliverable_action.go`'s `composeZip` iterates **exactly** the S3 listing of
`portfolio-sites/<domain>/`, and `verifyArchive` (`:259-261`) asserts `len(zr.File) == len(files)`.
**Adding a README without teaching that assertion makes the action fail** — and teach it the
synthesised entries **by name**, never by loosening the count, or a missing README and an extra site
file cancel out.

**Synthesise at zip time; do NOT publish the README into the bucket.** Objects under
`portfolio-sites/<domain>/` are the SERVED SITE — a README there would be publicly fetchable at
`https://<customer-domain>/README.txt`, crawlable, and visible to the customer's own visitors.
