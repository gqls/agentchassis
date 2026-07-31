# Where we are — loanandmortgagecalculator.co.uk

The owner's running log. Plain prose, append-only, newest at the bottom.

---

**2026-07-31, morning.** Started by finding out what we actually have. You asked me
to put mortgagecalculator.co.uk onto loanandmortgagecalculator.co.uk, so the first
job was to be sure which copy of the mortgage site is the real one.

There are two copies in your domains folder, and they are not the same. The live
site is served from the `gemini/02` subfolder, not from the top level. I checked
this properly rather than by looking at dates — I fetched all 23 live pages and
compared them byte for byte against both copies. All 23 match `gemini/02` exactly.
Worth saying plainly: if I had taken the obvious top-level folder, I would have
built the new site from the wrong material and every check afterwards would have
told me it was fine.

**Then I tried to prove the calculators work, and the testing tool lied to me.**
We have a tool that loads each calculator in a real browser and checks whether it
does what it promises. I pointed it at the mortgage site and it reported that all
fourteen interactive pages were broken — "nothing a visitor can touch". They are
not broken; they all work.

What saved me was the number, not carefulness. Fourteen out of fourteen identical
failures is not a description of a website, it is a description of a broken
instrument. A site with fourteen genuinely dead calculators would fail in fourteen
different ways. So I checked the tool instead of believing it, and the tool was
looking for the calculators inside an HTML element that this site doesn't use. It
had been checking one of our other sites for months, where every page does use that
element, so nobody had noticed.

I fixed it, and I proved the fix two ways: the previously-invisible calculators now
pass, and I re-ran the old version and the new version side by side over four pages
of the site it was originally built for, comparing nine results each. Nothing moved.
So I have made it see more without changing any verdict it was already giving.

Then the baseline came out clean: thirteen of thirteen mortgage calculators working.

**Something you may want to know about the old site.** It has had four broken things
live since it was built. Six of the nine guides have a "Home" link that goes to a
page that doesn't exist. The homepage links to a guide that doesn't exist either —
the file is called something slightly different. Two guides have nothing linking to
them at all, so nobody can find them. And there is no sitemap, with a leftover
placeholder comment in the robots file where the sitemap line should be. I have not
touched the old site — that wasn't the brief — but say the word and it is an hour's
work.

**On your point about the two sites evolving in different directions.** That changed
the design, and for the better. I had been planning to copy the site across and
write better guides. What you actually want is two sites with different audiences,
where the difference is recorded in the framework rather than being something I
maintain by hand.

So the new site is positioned as the **whole-borrowing-picture** site, and
mortgagecalculator.co.uk stays the narrow mortgage authority. The new one answers
the questions that span both subjects — how your car finance reduces what a
mortgage lender will offer, whether to consolidate debt into a remortgage, whether
the next thousand pounds should go on the deposit or clear a loan. That is a real
difference and not a paraphrase: no single-subject site can answer those questions,
which is exactly why they are worth owning.

You chose the combined site, so it has both halves: twelve mortgage calculators and
eleven loan ones, twenty-three in total, on one domain with one design.

**What is built.** All twenty-three calculators are ported and every one is verified
working in a real browser. The important thing about how I did it: the calculators'
arithmetic is copied byte for byte, and the build **refuses to run** if any of it
changed. I only rewrote the wrapper — the titles, the navigation, the footer, the
links. And I deliberately broke the check to make sure it actually catches a change,
because a safety check that has never gone off isn't a safety check.

Thirteen guides are entirely new writing. None is adapted from either site. One
editorial decision worth flagging: I have avoided quoting any current interest rate
or tax band anywhere. The old sites hard-code "3.75% base rate" and a March date,
and copy like that is wrong within weeks with nothing to tell the reader. The new
guides explain how things work, which doesn't go stale, and send people to the
calculator for numbers.

**Three problems I hit while porting, and they are instructive.**

One was mine. My build rebuilds the page header from scratch, and three of the
mortgage calculators keep their shared maths file linked in the header rather than
in the body. So the build quietly threw that link away. One of the three then failed
loudly, which is how I found it. The other two carried on looking perfectly fine
while missing the same file — those are the ones that would have shipped. I fixed it
properly by adding a second check that fails the build if anything a calculator
depends on goes missing, rather than fixing three files by hand.

The second was not mine, and it is a nice example of a thing that is genuinely hard
to see. One of the loan tools is a five-step questionnaire. Its code moves a marker
from one step to the next and relies on the stylesheet to hide the steps you are not
on. **That stylesheet rule was never written** — not on the loan site, not anywhere.
So all five steps have been showing at once on the live site, and the tool has been
visibly broken, and no amount of reading the code would ever find it because the
code is correct. The missing half is a class name that appears nowhere except as
text inside one instruction. I have written the missing rules and it works now. It
turned out to be one of thirty-six styles those tools use that were never defined.

The third was the testing tool again, twice more. It called a working tool dead
because it had no way to tick a checkbox, and it called the questionnaire dead even
after I had fixed it, because it only knew how to look for changes in a fixed list
of places and the questionnaire responds by showing a different section. Both fixed,
both re-checked for side effects, nothing moved. Three of the four failures on this
site were the instrument rather than the site, which is a ratio worth remembering.

**One tool I did not bring over,** and I want to be explicit rather than have it
just be absent: the loan site has a "6-month credit roadmap" filed with the
calculators, and it is not a calculator. It is under two kilobytes with no controls
and no code — a short article in the wrong folder. The subject is covered properly,
and in much more depth, by one of the new guides.

**Where this stops, and it needs you.** The site is built, checked, and the files
are already uploaded to our storage. It is not visible yet, because the domain is
still parked at whoever you registered it with — it currently bounces to a
registrar holding page. Two things need doing in the Cloudflare dashboard and only
you can do them: add the domain as a zone and repoint the nameservers, then add a
Workers Route for it pointing at the same little program every other site uses.
There are no Cloudflare credentials on this machine at all, so this is not something
I can script my way around.

The good news is that the storage side is already correct and won't need redoing —
the way our setup works, files are filed under the domain name, so the moment that
route exists the site simply appears. I verified all fifty-two files uploaded.

**After that, two things remain.** Adopting it into the framework so it is managed,
which I will do with the calculators locked so nothing can ever regenerate working
arithmetic, and the guides left open so the platform can keep improving them. And
recording the two sites' different audiences as framework settings, which is the
part that makes your "evolve in different directions" instruction stick without
anyone having to remember it.

**One caveat I want on the record.** I have verified that the ported calculators
respond correctly and that their code is byte-identical to the originals. I have
**not** verified that each one produces the same answer for the same input as its
original did. Byte-identical code with its dependencies present is strong evidence,
but it is not the same thing as checking the output, and I would rather say so than
let "verified" cover more than it earned. Per-calculator acceptance checks are the
obvious next job and there is already a pattern in the repo for them.
