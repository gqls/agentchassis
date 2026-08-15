# DRAFT — dartsonline.com privacy & cookies copy, for owner approval

**Status: DRAFT. Not published, not registered, no page created.** Requested by the owner
2026-08-15 ("please add a UK GDPR page so I can apply to the two affiliate programs").

**Why this is a draft for approval rather than generated copy.** The precedent is
`noted.co.uk`, two days ago: privacy copy was drafted, **the owner approved it (with one
edit)**, it was registered in `evidence_base` as `supplied_copy.privacy`, and the page was
built to carry it **VERBATIM** (`1e55dde19` → `fb317132a` → `f6e63e4d7`). That is the
platform's own answer to "the framework writes the content" for legally binding text: the
words are the owner's, and the framework is required to reproduce them exactly rather than
invent them. A privacy policy asserts facts about data processing; an invented one is a
false statement to regulators and to affiliate networks, so it does not go through the
normal writer.

---

## ⚠ THREE MEASURED FINDINGS THE OWNER MUST SEE FIRST

These are not copy suggestions. They are facts about the live site, measured 2026-08-15,
and two of them mean **no wording can make the site compliant on its own.**

### 1. Google Tag Manager is running with NO consent banner `[MEASURED]`

`curl https://dartsonline.com/` → container **`GTM-PQ3WCTBD`** present (2 references).
Search for any consent/cookie banner → **0 matches**.

Under UK PECR (the cookie rules that sit alongside UK GDPR), non-essential cookies —
analytics included — require **consent before they are set**. The site currently sets them
on arrival with no notice and no choice. **A privacy policy does not fix this**; the policy
describes the practice, it does not authorise it. Options: add a consent banner, or
configure the container to load nothing non-essential. This is the one genuine compliance
gap and it is a code/config change, not a copy change.

⚠ Note also that **`GTM-PQ3WCTBD` is the same container as `idea.uk`** `[MEASURED]` — it
appears to be a shared, fleet-wide container. So a change here is not necessarily local to
dartsonline, and whoever owns that container should be asked before it is altered.

### 2. The contact form does not submit anywhere — it opens the visitor's email client `[MEASURED]`

`<form class="contact-form" action="mailto:darts@contactforsales.com?subject=..." method="POST">`

This is genuinely good for privacy (no server-side collection, no database of enquiries)
and the copy below says so honestly. It also means the form is **unreliable in practice** —
`mailto:` form posts fail silently in most modern browsers — which is a separate
functional issue worth its own look, and matters because an affiliate network will try the
contact route when assessing an application.

### 3. The footer still advertises "Shipping & Returns", twice `[MEASURED]`

On a site whose own identity spec says it *"does not hold stock, run a warehouse, or have
a retail premises"*. The 2026-07-29 copy rails (D4) said to park shipping/returns. An
affiliate network's reviewer reads exactly these pages, and a shipping page on a site that
ships nothing is the kind of thing that gets an application queried. **Not in scope of this
draft — flagging it because the owner's stated goal is the application.**

---

## The draft copy

> **Privacy and Cookies**
>
> Darts Online is a UK-based darts publication. We publish buying guides, setup advice and
> news. We are online only: we do not hold stock, take payments, or ship anything, so we
> never handle payment or delivery details.
>
> This page explains what happens to information when you visit the site. It is written to
> be read, not to be waved at a regulator, and we have tried to keep it short and true.
>
> **Who is responsible**
>
> Darts Online is the data controller for this website. You can reach us at
> darts@contactforsales.com or on 07934 524 911.
>
> **When you contact us**
>
> The contact form on this site does not send anything to us directly. It opens your own
> email program with a message addressed to us, which you then send yourself. That means we
> receive an ordinary email, and we hold it in our mailbox for as long as we need it to deal
> with your enquiry. We do not keep a separate database of enquiries, and we do not add you
> to a mailing list.
>
> **Analytics and cookies**
>
> We use Google Tag Manager and Google Analytics to understand how many people visit the
> site and which guides are read. These set cookies in your browser and share information,
> including your IP address, with Google. We use this to decide what to write next. We do
> not use it to identify individuals, and we do not sell or share it with anyone else for
> their own purposes.
>
> You can clear or block cookies in your browser settings at any time, and doing so will not
> stop you reading anything on this site.
>
> **Affiliate links**
>
> Some links to retailers on this site may be affiliate links. If you follow one and buy
> something, we may earn a commission from that retailer at no extra cost to you. The
> retailer may set its own cookie to record that the visit came from us. This never changes
> what we recommend: our guides are written on specifications and how equipment behaves, and
> a product does not appear here because it pays better.
>
> **Your rights**
>
> Under UK data protection law you can ask us what information we hold about you, ask us to
> correct it or delete it, object to how we use it, or ask for a copy of it. Email
> darts@contactforsales.com and we will respond within one month. If you are not satisfied,
> you can complain to the Information Commissioner's Office at ico.org.uk.
>
> **Changes**
>
> If we change how any of this works, we will change this page.

---

## What the copy deliberately does NOT say, and why

- **No postal address**, per the identity truth reset (D4, 2026-07-29). A UK controller is
  normally expected to be identifiable; email and phone are given. **If the owner has a
  registered business name or address he is willing to publish, it belongs here** — that is
  an owner decision, not something to invent.
- **No claim of a consent banner**, because there isn't one (finding 1). The copy describes
  cookies honestly and points to browser controls, which is accurate but is **not a
  substitute for consent**. If a banner is added, this section should be revised to describe it.
- **No claim that affiliate links exist yet** — measured: zero outbound retailer links on the
  homepage today `[MEASURED]`. The wording is deliberately prospective ("may be"), so it is
  true today and stays true on the day the first link ships, which is what an affiliate
  application needs.
- **No invented retention periods, no invented processor list, no "we take your privacy
  seriously"**. Every claim above is checkable against the live site.

## If approved, the route to publish

1. Register the approved wording in `evidence_base` as `supplied_copy.privacy`, deriving the
   new spec row from the live one (`data || {...}`) so the existing banned_claims and facts
   carry across untouched — the `noted_rebuild/apply_privacy_copy.py` pattern, which exists
   and would need parameterising for this site.
2. Create the `pages` row. ⚠ **There is no framework path that adds one content page on
   demand** — established by the noted lane and worth not re-deriving: the planner makes all
   pages at once (unknown blast radius on a re-run), `create_tool_component` is tool-only,
   `create_report_page` forces `rebuild_policy='owned'`, and `needs_content_page` needs a
   page that already exists. The route used successfully on noted was to mirror the
   **adoption** path (`apply_adoption_plan_action.go:541`, `INSERT … ON CONFLICT (site_id,
   name)`), taking every value from the current plan so no identity is hand-rolled
   (`bugs_open/080`).
3. Add the footer link (the footer currently has none `[MEASURED]`), build, and verify the
   served page carries the copy verbatim.
