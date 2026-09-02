# Afternic — where we are (append-only, newest at the bottom)

## 2026-09-02 — lane opened

You asked for the Afternic API to be set up so sessions can manage your
domains there. The first thing I found is that Afternic doesn't offer an
API to people selling domains — their real APIs are only for registrars in
their partner programme, and what they built for sellers instead is a chat
assistant inside their own website. The people who do automate Afternic do
it by borrowing their browser login session, which works but is unofficial
and breaks whenever the login expires.

I put three routes to you: borrow the login session, ask Afternic for
proper access, or a no-credential loop where you export your portfolio as
a file and upload their bulk-change spreadsheet back. You chose the
no-credential loop, covering listings and prices, sales and leads, and
which domains are verified or parked.

What's in place now: a parser that reads your portfolio export safely —
it matches columns by their names, refuses anything ambiguous outright,
and checks itself against a value you quote from your own screen. That
last part is a direct fix for the July mistake where I misread your pasted
dashboard and invented a zero minimum offer. Each export is kept as a
dated snapshot and compared with the previous one, so sales, price changes
and new listings show up as a short diff. For "which domains are parked at
Afternic", the nameserver classifier another lane built already answers
it, so nothing new was needed there.

**What I need from you (two small things, one minute each):**
1. From your Afternic portfolio page, export the domain list as a CSV and
   drop it in this folder's `inbound/` directory — then tell me one value
   you can see on screen (for example a floor price) so I can prove the
   parse against it.
2. From the bulk upload page, download their template spreadsheet
   (`bulk_upload_sample_v3.xlsx`) and put it in this folder — that's what
   I'll fill in when you want prices or listings changed in bulk.

Also worth checking while you're in there: whether the dashboard has a
separate sales or leads export. If it does, grab one and I'll build the
reader for it.
