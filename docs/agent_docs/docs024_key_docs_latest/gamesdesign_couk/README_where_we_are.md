# gamesdesign.co.uk — where we are (append-only, newest at the bottom)

## 2026-09-02 — the lane opens, and the site gets its own name back

This directory is the home for work on gamesdesign.co.uk, which until today had no
thread of its own. The first job arrived within minutes of the lane opening.

The short story: when this site was first built back in June, it was created by
copying an existing site — gamedesign.uk. That copy carried across the other site's
brand name, so for three months gamesdesign.co.uk has been calling itself
"GameDesign.uk" on nearly half its pages — a name that belongs to a different domain
we also own, which is now being rebuilt as a separate site. You ruled today that
this has to stop, and chose the new name "GamesDesign.co.uk" — the site simply calls
itself by its own address, the same convention as Advertise.co.uk.

What we did: found every place the old name was stored (there were five different
layers — the site's identity documents, the site plan, the page titles, the page
content, and the already-rendered HTML), swapped the name in all of them in one
careful pass with backups, and then asked the platform to re-publish the 32 affected
pages. We were careful to leave one thing untouched: the site's own record that it
was originally adopted from gamedesign.uk, which is history, not branding.

One surprise along the way: a guide page here had a button linking to a simulator
hosted on the OLD gamedesign.uk site — and that old site was cleared out this very
afternoon by the team rebuilding it, so the button pointed at a dead page. Luckily
this site has its own copy of the very same simulator, so the button now points
home instead.

The deeper cause — the site-copying machinery blindly carrying the source site's
name onto a new domain — is now written up as a platform bug (bugs_open/439, filed
by the gamedesign.uk thread) so it gets fixed at the source and can't happen to the
next site we adopt.

Still to confirm as of this entry: that all 32 re-published pages actually show the
new name on the live site — being checked page by page now.
