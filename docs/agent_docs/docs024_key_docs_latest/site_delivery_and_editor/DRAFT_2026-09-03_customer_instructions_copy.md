# DRAFT — the hosting instructions, for the owner to correct

Written 2026-09-03 for `bugs_open/475`, after the owner ruled the instructions are wanted and should
live outside the ZIP, then answered "all three" to where.

**Nothing here is live.** This is copy for him to read and mark up. It is written to the house plain
style (`travelling_docs/pitch_pdf_source/REVERSE_ENGINEERED_STYLE_PROMPT_v3.md`): one idea per
sentence, no em dashes, no opening a fact with what isn't true, no words heavier than the fact.

## How "all three" fits together as ONE source, not three

His answer was all three, and they are not three separate documents. They are one body of copy with
three exits:

| exit | what it carries | why this one |
|---|---|---|
| **the page** | everything below, with the `{{slots}}` filled for that customer | correctable after a customer already has the email; a free host changing its signup is a five-minute fix for everyone, including people mailed last month |
| **the ZIP** | `README.txt` — §1 and §2 only, plus the page's address | the ZIP outlives the email. Somebody who keeps their files and comes back in a year has the folder and no message. Keep it SHORT and free of anything that rots |
| **per-site** | the `{{slots}}` below, filled at delivery time | a page naming their own domain is worth more than a general one, and the slots are the only part that differs |

Write it once. The generic page is this text with the slots showing a general form. The per-site page
is this text with the slots filled. The README is the first two sections plus a link.

**The one rule that keeps them honest:** anything that can go out of date lives on the page, never in
the README. Free hosts change their signup flows. Our prices change. The ZIP cannot be edited once a
customer has it.

---

## §1 — What this folder is

Your website, finished, as files.

Open `index.html` in any browser and the whole site works, offline, exactly as it does online. Every
page, every picture, every link between them. Nothing needs installing and nothing phones home.

There are `{{file_count}}` files in here. The pages are the `.html` ones. The pictures and styling
live in `assets`. You don't need to understand the folder layout to use it, and you shouldn't
rearrange it, because the pages find each other by where they sit.

## §2 — Where the full instructions live

The rest of this, including how to put the site online, is at:

**{{instructions_url}}**

That page is kept up to date. This file isn't, so if the two ever disagree, believe the page.

*(§1 and §2 are the whole of `README.txt`. Everything below is page-only.)*

---

## §3 — Your site is already online, for now

It's live at **{{live_site_url}}**, and it stays there until **{{live_until_date}}**.

That gives you time to decide what to do next. You don't have to do anything today.

## §4 — Putting it online yourself, free

A site made of files like these is the cheapest kind to host. Several companies will do it for
nothing, because it costs them almost nothing.

The quickest is Netlify. Go to `app.netlify.com/drop`, drag the whole unzipped folder onto the page,
and wait about twenty seconds. You get a working web address straight away. No account is needed to
try it, though you'll want one to keep it.

Cloudflare Pages and GitHub Pages both do the same job and are also free. They take longer to set up
the first time. If you already use one of them, use that instead.

Whichever you pick, the thing you upload is the folder, not the individual files, and not the ZIP.

> **OWNER: this is the section I am least sure of, and it's the one that rots.** I have not verified
> today that `app.netlify.com/drop` still works without an account, and free tiers change. Before
> this goes near a customer, somebody should do it once, end to end, with this actual ZIP, and write
> down what they actually saw. Everything else in this document I can check; this needs a person.

## §5 — Using your own domain

{{domain_paragraph}}

*Per-site slot. The three cases, and they need his ruling on wording as much as on price:*

- **They bought the domain outright.** It's theirs. They point it at their new host, and the host's
  own instructions cover that better than we can.
- **They're renting it from us at £10 a month.** It keeps working while they rent it. Reply to the
  delivery email to change anything.
- **They have neither yet.** Buying it outright is a one-off £59.99, and it's then theirs to move
  wherever they like.

## §6 — Changing the site later

You can edit these files. They're ordinary HTML, so any text editor opens them, and there's no
system to learn.

Changing words is safe. Find the text in the `.html` file and type over it. Save, and reload the page
in your browser to see it.

Changing layout is harder, and the honest advice is to keep a copy of the original folder before you
start. Then a mistake costs you nothing.

We don't include changes in the price. Editing a site that already works is much easier than starting
from nothing, which is most of what you've bought.

## §7 — If something looks wrong

Compare it against **{{live_site_url}}** while that's still up. If the page looks right there and
wrong on your own hosting, the upload is the problem, not the files. The most common cause is
uploading the contents of the folder instead of the folder.

---

## What this draft does NOT settle, and needs the owner

1. **§4 has never been performed.** See the note in it. Instructions nobody has followed are a guess
   with formatting.
2. **Which host we recommend first.** I've led with Netlify Drop because it needs no account and no
   command line. That's a judgement, not a measurement.
3. **The `{{live_until_date}}` slot exposes the three-lifetime problem** already filed in the RUNBOOK:
   the email says 30 days, the tokens run 42, the download link's signature lasts 7. A page that
   states a date has to state the right one, and right now there are three candidates.
4. **Whether the per-site page is public or behind a token.** `/c/` and `/d/` are token-addressed
   because they do something. A page that only reads could be public, which makes it linkable and
   bookmarkable. Public also means it names a customer's domain at a guessable address.
5. **Where the page is served.** `links.webdesign.uk` already serves `/c/` and `/d/`, so the seam
   exists, and a third route is not new architecture. But the pages there are currently generated by
   Go, not by the site framework, which the "every site goes through the framework" ruling makes
   worth a second look before adding to.
