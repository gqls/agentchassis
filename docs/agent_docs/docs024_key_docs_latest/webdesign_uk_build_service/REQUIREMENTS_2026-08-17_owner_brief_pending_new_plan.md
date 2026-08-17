# Owner brief, 2026-08-17 — five requirements, ON HOLD pending the new plan

**STATUS: HOLD. Do not start any of this.** The owner is finalising the plan in a
separate session (`webdesign live web builder project`, session id
`d10f1acc-1627-4729-b660-93d6e84911e3`) and said explicitly: *"Don't go ahead until
I've finalised that plan."* This file exists so the requirements survive the wait and
so the next session starts from what he actually said, not from a paraphrase.

His words, and what each one means in this system:

## 1. The Brief Starter sentence links nowhere

> *"The 'Or answer a couple of quick questions first with the Website Brief Starter…'
> text links nowhere."*

**Measured, and it is worse than he thought:** the sentence is wrapped in
`<a href="tel:+44 (0) 7934 524 911">` — it does not link nowhere, it DIALS THE PHONE.
The `tel:` URI is malformed as well (spaces/parens). Full case, evidence, the control
that stops a false pass, and why it is a producer question and not a copy question:
**`bugs_open/299`**. The section was written **2026-08-16 16:12:45**, i.e. AFTER the
268 fleet fix, so a chassis carrying that fix produced it.

## 2. The chat box moves to the HOME page

> *"I'd like the chat on the home page - probably the first thing after the 'One
> website, one price, built once.' block and prices."*

Placement: `index`, immediately after the hero/price block. Today the chat lives on
`contact` only (`page_components` position 3, `lock_type='permanent'`). Note the
interaction: **the lock is what has been protecting it**, and the 285 fix (live,
accepted 2026-08-17) is what makes a rebuild keep it in the section list. Moving it
means a new placement on `index` and a decision about whether `contact` keeps one.

## 3. Stronger copy on what the offer does NOT include — with pride

> *"I would need stronger copy of what it doesn't offer here as well. We can be more
> proud of the positioning."*

This is a **positioning** instruction, not a disclaimer instruction. The register
already attests the exclusions (`no_refund`, `no_changes_included`, `no_lock_in`,
`price_is_total_no_vat`); what he is asking for is that they be stated as
**deliberate choices that make the offer good**, not as small print. The house voice
rules still apply (no em dashes, no agency-marketing weight, plain British English).

## 4. Do it THROUGH THE FRAMEWORK

> *"Please try and do this through adjustments to the framework e.g. spec and planner
> or however a normal edit would come through from a client."*

**This is the binding constraint on HOW, and it matches the standing owner ruling
(2026-08-04): every site goes through the framework, never hand-built.** So: change
the SPEC (`site_specs` — `evidence_base` / `site_plan` / design intent) and let the
planner and writers regenerate, rather than surgically UPDATE-ing `rendered_html`.
The lane's own surgical-SQL habit is the wrong tool for this job by his instruction.
Treat it as a customer edit request arriving at the front door and follow that path.

## 5. The payment sentence is still wrong

> *"Also this text is still wrong: 'You see it first on a private preview link, and
> you pay in full only once you are happy with it.'"*

**The page is FAITHFUL to the register, so the register is what must change first.**
The live `evidence_base` fact `payment_after_approval` reads: *"The customer sees the
finished site on a private preview link and pays after they have approved…"*, and
`no_refund` reads *"No refund is offered. The customer approves the site before
paying, so there is nothing to return."* The page copy is a faithful rendering of
those facts. **Correcting the page alone does nothing** — the writer's instruction set
and the claims gate both read the register, so the next rebuild restores the wording
(this is the `bugs_open/161` mechanism: the register both causes the claim and then
vouches for it). Sequence: owner rules the true terms → supersede the
`evidence_base` fact (never edit in place; inherit `pinned`; claimscan-test against
the live corpus first) → then rewrite the pages.

**What is unclear and must not be guessed:** *why* it is wrong. Candidate readings —
"pay in full" conflicts with a deposit; or "only once you are happy" implies a
satisfaction condition that sits badly with "no refunds, no changes". **Ask; do not
infer.** The exact wording is owner copy.

## 6. The whole site gets rewritten to fit the new plan

> *"We will need to rewrite the site to fit the new plan to be fair."*

So do not spend effort on surgical fixes to copy that is about to be regenerated —
including `bugs_open/299`'s CTA and the "Get in touch" duplication the 2026-08-17
rebuild introduced. **File and wait.** What survives the rewrite is the PRODUCER
questions: does the CTA generator point at tool pages correctly, and does the register
say the true commercial terms.
