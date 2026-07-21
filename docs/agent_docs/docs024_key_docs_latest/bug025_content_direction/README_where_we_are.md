# Where we are — bug 025, the dead content_direction column

Plain-prose log, append-only, newest at the bottom.

## 2026-07-21

Bug 025 was a "false affordance": the `pages.content_direction` column had a comment
promising you could steer a single page's copy by setting it, and the same promise was
written into the main contracts doc — but no code actually read the column. Anyone who
believed the comment and planned around it lost time (the relojistas thread did, which
is how it got filed).

You had three options: delete the dead column, actually build the feature it promised,
or just fix the misleading comment. You chose to **build it** — make the code match
what the docs already say.

I checked the whole chain against the live system first, and it's a good job I did:
the old seed file said the page builder gets its page data from one place, but the
*live* builder had been changed to load the page fresh from the database at build time.
If I'd trusted the seed I'd have wired the wrong loader and "verified" a fix that did
nothing on the real rebuild path. So I wired both loaders that matter.

State now:
- The code change (two small Go loaders now read the column) is written, compiles, and
  is committed — but Go changes only take effect when a new chassis image is built and
  rolled out, so it is not live yet.
- The prompt change (the writer now has a "Page-Specific Content Direction" section
  that only appears when a page actually has one set) is live in the database already.
  I tested the writer's template against the real live prompt for three cases (none /
  full / partial) and it renders correctly and adds no stray blanks.
- I corrected the two docs that carried the false promise.

What's left to fully close it: build and roll the chassis image, then the real proof —
set a direction on one live page, rebuild it, and read the saved section to confirm the
instruction actually changed the copy. Until that's done and live, the bug stays open.
