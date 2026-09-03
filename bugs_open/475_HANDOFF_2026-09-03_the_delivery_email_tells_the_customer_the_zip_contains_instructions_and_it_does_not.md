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

**Open question for the owner, and it is the only thing blocking:** where do the instructions live,
and are they the same for every customer or per-site? A generic "how to put a folder of HTML on free
hosting" page is one artefact for everyone and can be written once. Anything naming their domain or
their host is per-site and needs generating.

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
