# DRAFT v3 — the hosting instructions, rewritten after the owner actually performed them

Supersedes `DRAFT_2026-09-03b_customer_instructions_copy.md`. **v2 is kept, not corrected**, because
the gap between what I wrote from knowledge and what happened in a real browser is the most useful
thing in this directory. Read them side by side before writing instructions for anything else.

The owner ran §4 on 2026-09-04 with the real `idea.uk` ZIP and sent ten screenshots and a written
account. Every correction below is his observation, not my inference.

---

## What v2 got wrong, in order of how badly

**1. "No account is needed to try it."** False. The signup demand is immediate, on the drop. There is
no anonymous preview and no unclaimed site to claim later, so v2's step 6 — *"a site dropped this way
belongs to nobody until you claim it"* — described a flow that no longer exists.

**2. IT IS PRIVATE BY DEFAULT, and v2 did not mention privacy at all.** This is the one that would
have cost a customer their site being visible. Netlify now ships a *"🔒 New: Private by default —
build privately, share when you're ready"* behaviour. A dropped site is reachable **only by the
logged-in owner** until they explicitly publish it.

**3. And it looks published when it is not.** The owner opened his own URL and it rendered correctly —
*because he was signed in*. Only a private browsing window revealed **"This site is private. Sign in
with an invited Netlify account to view it."** He had already been shown a "Your project is now
private" screen, chosen public, and been returned to a page **still marked private**; the state only
changed when he pressed a separate **Make public** button and got a *"Your project is public — anyone
can visit your production site"* confirmation.

> **This is the same shape as everything else this week: a check that passes for the person
> performing it and fails for everyone else.** A customer follows the steps, visits their address,
> sees their site, and concludes it worked. It did not. **The instructions must make the signed-out
> check a step, not a suggestion.**

**4. Signup is not one screen.** His actual sequence: sign-up form → *"Please complete the security
check and try again"* **with no security check visible anywhere** → retry → *"Password is too easy to
guess"* on a good 11-character password → a longer password accepted → confirmation email, **slow to
arrive** → verify link → and only then the project screen.

**5. v2's timings are fiction.** "About twenty seconds" and "about a minute" describe the upload. The
owner's elapsed time from drop to a publicly reachable site was **roughly forty minutes**, most of it
waiting on an email.

---

## §4, rewritten. This is what actually happens.

Putting the site online is free and there is no card. Set aside **half an hour** the first time,
because most of it is making an account and waiting for a confirmation email.

**1. Unzip the file.** Double-click it. You get a folder next to the zip. Open it and check
`index.html` is at the top, alongside a folder called `assets`.

**2. Go to `app.netlify.com/drop` and drag the folder onto the dashed box.**
Drag the folder, not the zip and not the loose files inside it.

**3. It will ask you to make an account.** There is no way to skip this and no anonymous preview. You
can use Google, or an email address and password.

**4. Expect the signup to argue with you.** Two things happened to us and both are Netlify's, not
yours. It may say *"Please complete the security check and try again"* **without showing you a
security check** — try again and it passes. It may reject a perfectly good password as *"too easy to
guess"*; a longer one with more words is accepted.

**5. Wait for the confirmation email and click Verify email.** Ours took a while. Check spam.

**6. ⚠ YOUR SITE IS PRIVATE AT THIS POINT. This is the step people miss.**
Netlify makes new sites private by default. You will see a screen saying so. Choosing "public" on
that screen may not be enough — we were returned to a page still marked **Private**.

Find the **Make public** button and press it. You should get a confirmation reading *"Your project is
public — anyone can visit your production site."*

**7. Now check it the only way that tells the truth: open your address in a private or incognito
window** (or on your phone with wifi off).

This matters more than any other step. **Signed in, a private site looks completely normal to you.**
If it is still private you will see *"This site is private — sign in with an invited Netlify account
to view it"*. If you see your site, it is genuinely live for everyone.

**8. Your address is something like `tangerine-valkyrie-b5c0a7.netlify.app`.** You can rename it, and
you can point your own domain at it, from the Domain management panel.

---

## Things the owner found that we should decide about, not just document

**Netlify offers an AI agent that edits the site.** The project screen carries a *"Work with an AI
agent"* panel — *"ask an agent to create a new page"*, with OpenAI Codex among the options. A
customer we hand a folder to is one click from a competitor's agent rebuilding it. Not a problem to
solve in copy; worth the owner knowing it is there.

**Netlify sells domains from the same screen.** "Custom domain — use your own domain, or buy a new
one" with a Go to Domain management button, sitting directly beside our own £10/month rental and
£59.99 buy-out offer in the delivery email.

**Screenshots.** The owner has asked for images on the instructions page, and the ten he captured are
the right ones: the drop box, the security-check message, the password rejection, the verify email,
the "project is private" screen, the still-private project page, the signed-out "This site is
private" wall, and the "Your project is public" confirmation. **The signed-out wall is the most
valuable image on the page**, because it is the thing a customer will otherwise never see.
Source: `/home/ant/Downloads/idea_uk_netlify/`.

---

## Still open

1. **Where the page is served, and whether it is public or token-addressed.** Unchanged from v2.
2. **`{{live_until_date}}`** — unchanged; three candidate dates that disagree.
3. **Whether we recommend Netlify at all**, now that the real cost is an account, a password fight,
   an email wait and a privacy step that silently fails open-to-nobody. The alternative is not another
   host with the same shape. It is us hosting it, or us making the account — which the owner has
   asked to consider separately.
