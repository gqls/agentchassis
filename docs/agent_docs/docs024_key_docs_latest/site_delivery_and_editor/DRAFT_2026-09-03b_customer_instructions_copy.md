# DRAFT v2 — the hosting instructions, with the Netlify walkthrough the owner asked for

Supersedes `DRAFT_2026-09-03_customer_instructions_copy.md`. **v1's §1 was wrong** and this file
says so rather than quietly fixing it: see "What v1 got wrong" below.

Written for `bugs_open/475`. **Nothing here is live.** The owner is about to perform §4 himself, which
is what turns it from a guess into instructions.

---

## What v1 got wrong, and how the owner's own request caught it

v1 §1 said: *"Open `index.html` in any browser and the whole site works, offline, exactly as it does
online."*

**That is false for this ZIP.** The owner asked for a section on viewing the site locally, I went to
write it properly, and the check took one command:

```
absolute references ("/assets/…") across the 10 pages : 87
relative references                                    : 0
```

Every page asks the browser for `/assets/css/styles.css`. Opened from a folder, that resolves to
`file:///assets/css/styles.css` — the root of their hard disk — not to the file sitting next to it.
So a double-click gives an unstyled page with no pictures and no working header, footer or news.

I wrote the false version while fixing a bug about a false claim. **Asking for the section is what
found it**, which is the argument for him performing §4 too.

> **This is also a product question, not only a copy question, and it is bigger than this document.**
> A deliverable the customer cannot open is a weaker deliverable. If the framework emitted relative
> paths, the ZIP would work by double-click, offline, for ever. That change touches the live site as
> well, so it is filed rather than done here: `bugs_open/476`.

---

## §1 — What this folder is

Your website, finished, as files. `{{file_count}}` of them.

The pages are the `.html` files. Everything they need — pictures, styling, scripts — sits in
`assets`. Keep the folder as it is, because the pages find each other by where they sit.

**Double-clicking `index.html` will not show you the site properly.** You'll get the words with no
styling and no pictures. Nothing is broken and nothing is missing. A browser opening a file straight
off your disk looks for `assets` at the top of your hard disk instead of inside the folder, and the
site was built to be served from a web address.

Two ways to see it as it really looks. Put it online, which takes about a minute and is §4. Or, if
you're comfortable with a terminal, go into the folder and run `python3 -m http.server 8000`, then
open `http://localhost:8000`. That serves the folder the way a web host does, and everything appears.

## §2 — Where the full instructions live

**{{instructions_url}}**

That page is kept up to date. This file isn't. If they ever disagree, believe the page.

*(§1 and §2 are the whole of `README.txt`. Everything below is page-only.)*

---

## §3 — Your site is already online, for now

It's live at **{{live_site_url}}** until **{{live_until_date}}**. You don't have to do anything today.

## §4 — Putting it on Netlify, step by step

Netlify hosts sites like this free. There's no card and no trial. This takes about a minute.

**1. Unzip the file you downloaded.**
Double-click it. On both Mac and Windows this makes a folder next to the zip, named after it —
something like `{{zip_basename}}`. Open the folder and check you can see `index.html` sitting at the
top, alongside a folder called `assets`. If instead you see one folder and nothing else, go into that
folder — that's the one you want.

**2. Go to `app.netlify.com/drop` in your browser.**
You'll see a large dashed box that says you can drag your site folder in.

**3. Drag the FOLDER onto the box. Not the zip, and not the files inside it.**
Drag the folder from step 1 — the one with `index.html` at its top. This is the step people get
wrong. Dropping the zip, or selecting all 45 files and dragging those, does not give the same result.

**4. Wait.** It uploads and then builds for a few seconds. You'll see it counting files.

**5. You get a web address.**
Something like `curious-pastry-3f8a1c.netlify.app`. Click it. Your site should look exactly as it does
at {{live_site_url}}, with pictures and styling.

**6. Claim it.**
A site dropped this way belongs to nobody until you claim it, and Netlify will say so on the page.
Make a free account when it offers, or the site goes away. You can rename the address afterwards from
`Site settings`, and you can point your own domain at it from `Domain management`.

> ⚠ **OWNER — THE THREE THINGS I AM NOT SURE OF, and your run settles them.** Tell me what you
> actually saw and I'll correct this before it goes near a customer.
>
> 1. **Step 6.** I believe an unclaimed Drop site expires if you don't make an account, but I do not
>    know the current window or the exact wording Netlify uses. If it says something different, the
>    step is wrong.
> 2. **Step 1's folder name.** I know the ZIP is flat — 8 files and 4 folders at the top, no wrapper —
>    so both Mac and Windows should create a folder named after the archive. I have not watched
>    either do it.
> 3. **Step 3.** Whether dragging the 45 selected files works as well as dragging the folder. I've
>    written "don't" because the folder is the reliable form, but if files also work the warning is
>    over-stated and should go.
>
> Everything else in this document I have checked against the actual ZIP.

## §5 — Using your own domain

{{domain_paragraph}}

*Per-site slot. Three cases: bought outright (theirs, point it at the new host); renting at £10 a
month (keeps working, reply to the email to change); neither yet (£59.99 one-off, then theirs).*

## §6 — Changing the site later

These are ordinary HTML files. Any text editor opens them and there's no system to learn.

Changing words is safe. Find the text, type over it, save, reload. Remember §1 — to see the change
properly you need it served, so the quickest loop is to edit, then drag the folder onto Netlify again.

Keep a copy of the original folder before you change anything structural. Then a mistake costs
nothing.

We don't include changes in the price. Editing a site that already works is much easier than starting
from nothing, and that's most of what you've bought.

## §7 — If something looks wrong

Compare it against **{{live_site_url}}** while that's still up. Right there and wrong on your own
hosting means the upload is the problem, not the files. The usual cause is uploading the contents of
the folder instead of the folder itself.

---

## Still not settled

1. **§4 steps 1, 3 and 6** — the owner's run settles these. Until then they are marked above.
2. **`{{live_until_date}}`** — there are three candidate dates and they disagree (presign 7 days,
   email says 30, tokens 42). A page that states a date has to state the right one. See the RUNBOOK's
   three-lifetime finding, under test now.
3. **Whether the page is public or token-addressed.** `/c/` and `/d/` are token-addressed because
   they DO something. A page that only reads could be public and therefore bookmarkable, but public
   also means it names a customer's domain at a guessable address.
4. **Where the page is served.** `links.webdesign.uk` already serves `/c/` and `/d/`, so the seam
   exists. Those pages are generated by Go rather than by the site framework, which the "every site
   goes through the framework" ruling makes worth a second look before adding a third.
5. **Which host to lead with.** Netlify Drop needs no account to try and no command line. That's a
   judgement, and the owner's run is the first evidence either way.
