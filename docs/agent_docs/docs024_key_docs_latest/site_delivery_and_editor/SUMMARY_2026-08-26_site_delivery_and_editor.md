# SUMMARY 2026-08-26 — site delivery and customer editor

## What we're trying to do

When a customer buys a £149 website, we need a way to actually hand it over: their site
live at an address, their files as a ZIP to keep, one email carrying everything, and a
quality check by the owner before anything goes out. Later, a customer editor on top —
and since this week, the plan is that the editor's input will be voice.

## Where we've come from

The last summary (18 August) ended at: the ZIP machinery worked and was proven, but there
was no way to send the delivery email, no review step, no download link a customer could
keep, and the whole delivery route into the cluster leaned on one line of configuration
on a box — widen it by a character and the admin system that holds every site's data
would have been on the internet. On 25 August the launch was paused with two product
questions open: should the owner edit sites before customers see them, and should
customers be able to edit during their hosted month.

## What we've done

Both questions were answered by the owner on the 26th. He gives himself an edit pass
before each site goes out — invisible to the customer, who still buys a one-shot product
with no approval stage — and customers get no editing at launch, with voice editing as
the next thing we build after launch. Neither answer moved the terms, the register, or
the copy, which means the launch position from the 25th stands as written.

The delivery machinery is now complete, and it was built defensively. Customer links got
their own door into the cluster, so a mistake on the box can no longer expose anything
but those links — and the software refuses to start if anyone ever wires a customer
route back onto the admin door. The confirmation link became a page with a button, so
email scanners can't press it. The delivery email exists as a mechanism whose wording
lives in editable configuration — the owner can change the words without any deployment —
and it will not send for any site he has not approved, cannot send twice however it is
retried, and refuses to send at all if any link it promises can't be produced. The ZIP
download link lasts the customer's full window even though the storage links behind it
die weekly: a refresher renews them on a six-hour cycle, and if a customer ever catches a
stale one they get an honest "being refreshed, try again shortly" page rather than a
broken download.

An independent adversarial review was run over the finished work at the owner's request,
and it earned its keep: it found a race that could have double-emailed a customer, a gate
that could never have opened, and a pod that could sit looking healthy with its customer
door dead. All fixed and proven, most with the failing test written before the fix. Every
piece then went through the review council; the final two rounds passed first time.

The email account itself went from "not working" to fully proven: the fault was never
the mail server (which was healthy throughout) but where the DNS records had to live, and
by the end of the evening a real test message passed all three authentication checks at
Gmail itself. There is no reason to change email host.

The domain programme also firmed up around us: registering customer domains is now a
severable, opt-in layer per site, the registration and DNS tools exist in dry-run form,
and the contract our email and hosting code will read was reviewed and settled before
anything reads it — including the rule that any confusion falls back to serving the site
at its safe default address, never to taking it down.

## Where we are now

Every piece of code for delivery is written, tested, and council-approved. The database
change is applied. The mail path is proven end to end. Nothing more needs inventing —
what remains is mechanical: the next deployment picks up the new code, two prepared
configuration files are applied after it, and the email account's password goes onto the
cluster as a secret only the owner holds. No customer has been handed anything yet, and
nothing can go out until the owner approves a site through his own review step.

## Where we're going

Immediate: the deployment-and-apply sequence above, then a first rehearsal of the whole
flow — file a review, owner edits and approves, cut the ZIP, send the email — on a site
of our own before any customer sees it. After that, delivery waits only on the shopfront
launch and the first sale. The next build on this lane is the customer voice editor, per
the owner's ruling. In parallel, the domain layer's first real registration is a
deliberate, owner-gated moment (it spends money), and the two Cloudflare facts the DNS
plan needs are still one dashboard look away.
