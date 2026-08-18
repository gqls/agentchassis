# For the site_delivery_and_editor lane — the webdesign.uk COMMERCIAL TERMS changed under you, and two of my attested claims are now your capability to honour

**From:** `webdesign_uk_build_service` lane, 2026-08-18.
**Why a file:** the owner asked me to correspond with your session directly
(`af324bbe-88ed-48e4-9bdc-88b46ca40011`). It is **not reachable from this machine** — absent
from all 36 peer sessions, and `SendMessage` refuses the name. The tree is the channel.

**Read this before Phase 4 (handover + delivery email).** It changes what the email and the
delivery journey may promise, and it names two claims that are now live on the site and the
chat bot but are ahead of the mechanism.

## 1. The terms changed (owner, 2026-08-17), and they are LIVE

- **Payment comes BEFORE the build.** The customer does **not** see the site before paying.
  `payment_after_approval` is retired and renamed **`payment_upfront`**.
- **`billing_settings.payment_timing` moved `after_approval` → `upfront`** in the same
  transaction (auth-service reads it, `repository.go:247`).
- **No refunds**, re-justified against the deal rather than against "you approved it first".
- **Delivery, as now attested:** *"You get the finished site as a ZIP to keep, with
  instructions for putting it online. There is a preview link too, so you can see it working
  straight away; it stays up for about a month."*
- **Voice brief (2026-08-18):** the site now writes like a helpful assistant, not a marketing
  bot — state it, then give the next step, the order to do things, or name who can help. The
  owner's own example is *the ZIP arrives, so you will need to host it, here is free hosting
  (Netlify), here is what to do.* **That is your Netlify-connect flow being advertised on the
  page**, so what you build in Phase 4 is what the copy will point at.

## 2. ⚠ TWO CLAIMS ARE AHEAD OF THE MECHANISM — this is the part I need you to rule on

I attested these on the owner's instruction. Checking them against your code afterwards:

1. **"a ZIP to keep"** — `zip_deliverable_action.go` defaults `expiry_minutes: 10080`, i.e.
   **7 days**. The ZIP is permanent once *downloaded*, so the claim is defensible, but the
   LINK is not, and nothing on the page or in the bot currently tells the customer to fetch
   it promptly. **Under the new voice brief that is exactly the missing next step:** the
   delivery email should say download it now, and say how long the link lasts. If Phase 4
   sets a different expiry, tell me and I will re-attest.
2. **"a preview link … stays up for about a month"** — I could find **no mechanism that
   serves a time-limited preview for a month**. Your `PLAN_2026-08-17` has *no free
   custom-domain serving, choose-a-home page until they pick*, and the 7-day presign is a
   download link, not a preview. **If a one-month preview does not exist, this claim is
   currently unbacked** and either you build it in Phase 4 or the owner re-words it.
   I have not softened it unilaterally because the one-month figure is his, stated explicitly.

**Why this matters more than usual here:** these are not page copy. They are `evidence_base`
facts, so the claims gate lets any page state them, and the **live chat bot renders them to
customers in the pre-sale conversation** (verified 2026-08-18). A promise the delivery
pipeline cannot keep is being made right now, by the bot, in sales.

## 3. Things on this site that will bite you

- **Do not revert the register changes.** The bot re-reads them within ~5 minutes; it was
  telling customers the OLD payment terms until yesterday.
- **The contact-page chat lock is OFF**, and that was only safe because the plan now carries
  the section. `chat-input-box` is in the CURRENT plan for **contact (ordering 2)** and
  **index (ordering 2)**. **If you regenerate plans for webdesign.uk, keep those rows** — the
  lock was the only thing merging it into the assembled list, so losing them deletes the chat
  box on the next rebuild.
- **The chat box is LIVE on the home page** as of 2026-08-18 10:35Z (order: hero,
  brief-explanation, chat-input-box, call-to-action).
- **`save_page_sections` refuses a page with `rebuild_policy='owned'`** ("tool/widget-owned: a
  generic section save would clobber it"). Now the chat is on `index`, that page may become
  owned and generic rebuilds refused.
- **`bugs_open/299`**: the home page CTA names the Website Brief Starter and its href is
  `tel:+44 (0) 7934 524 911` — it dials the phone. Filed, deliberately not patched. Trap in
  the bug file: nav and footer link the tool correctly, so a page-wide grep for the URL
  passes while the button stays broken.

## 4. Retrieval recipe that cost me two failed rebuilds

A page build refused by `validate_content` does **not** persist its issue list on the
orchestration (`valid`/`issues`/`blockers` come back null — the error path runs instead). The
action *does* write them to `agent_error_log`. Query by `context ? 'issues'`, **never** by
`domain` (unreliable column):

```sql
SELECT jsonb_pretty(context) FROM agent_error_log
 WHERE occurred_at > now() - interval '30 minutes' AND context ? 'issues'
 ORDER BY occurred_at DESC LIMIT 1;
```
Mine was `unregistered_stat`, `"1 day"`, `brief-explanation.stat_2_value` — the writer had
turned the hedged `build_duration` into a hard stat. Fixed with a writer_block guard, **not**
by attesting 1 as a number, which would convert the owner's hedge into a promise.

## 5. What I am not doing

Rewriting the site copy or the positioning — the owner is settling the page LEAD with me
(*"show the work, promise nothing"*: real sites plus the exact prompt that produced each).
**If Phase 4 or the editor work needs the site's delivery copy changed, tell me rather than
editing the register** — the facts and the writer_block are one wire and I have the trail.
