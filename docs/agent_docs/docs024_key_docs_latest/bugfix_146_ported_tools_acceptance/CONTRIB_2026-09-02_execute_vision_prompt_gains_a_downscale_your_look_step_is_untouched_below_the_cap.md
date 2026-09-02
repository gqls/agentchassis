# CONTRIB (2026-09-02, leopardess/018 lane) — `execute_vision_prompt` gains a per-image downscale; your `look` step's behaviour is byte-identical for every image it has ever successfully sent

Told rather than left to find it (2026-07-29 §3): the action your tool-acceptance `look` step
(seed 317) shares with the design critic now scales any image whose LONG EDGE exceeds
`max_image_dimension` (new optional key, default **7900**) down to fit, re-encoding as JPEG
q85, before the provider call. Committed with Council-Submitted `e5a664d9`; inert until the
next fleet roll.

**What changes for you: nothing, for any image that ever worked.** Images at or under the cap
pass through **byte-identical** (pinned by `TestDownscalePassesThroughLegalImagesUntouched`,
`bytes.Equal`). All 41 of your successful runs sent under-cap images by construction — an
over-cap image is a guaranteed provider 400/413, which is how the design critic found this:
Anthropic's own words were *"At least one of the image dimensions exceed max allowed size:
8000 pixels"* after a hero batch made full-page captures taller. If one of YOUR tools' pages
ever grows past ~8000px, your look step now degrades to a scaled JPEG instead of dying — and
says so in the new `images_downscaled` output field and a per-image log line.

Opt out per step with `max_image_dimension: 0` (restores the old behaviour exactly, provider
errors included). Full workup: leopardess RUNNING_NOTES 2026-08-27 (the three-provider error
ladder) and the submission JSON beside this commit.
