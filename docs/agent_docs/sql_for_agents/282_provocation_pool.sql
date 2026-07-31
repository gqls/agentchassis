-- 282_provocation_pool.sql — the provocation pool, and the nine we already have.
--
-- WHY THIS EXISTS
-- The daily provocation on vonc.com has never rotated. The cause was never a
-- broken job: there was no mechanism at all. Phase 0 (2026-07-31) made the
-- builder schedule-driven, but that schedule is a Python literal under docs/
-- which the cluster cannot execute and `make build-*` does not ship. So nothing
-- deployable can choose today's provocation. This table is where the schedule
-- moves so that a scheduled orchestration can read it.
--
-- WHAT SELECTS FROM IT
-- platform/orchestration/actions/provocation_feed_action.go
-- (`render_provocation_feed`), which ports the rules verified across 39 dates by
-- provocation_pipeline/builder/verify_rotation.py.
--
-- THE TWO RULES ENCODED IN THE SHAPE, rather than left to whoever writes the SQL:
--   1. `today` is the latest APPROVED row whose publish_on has arrived; the
--      archive is everything published strictly before it. So an entry is
--      archived exactly when a later one is published and is never in both at
--      once — the owner's rule of 2026-07-31, as a property of the data.
--   2. Two rows sharing a publish date would make "the latest" ambiguous, and an
--      ambiguous daily is how you get a different provocation depending on plan
--      order. The partial unique index below makes that unrepresentable rather
--      than merely discouraged.
--
-- DELIBERATELY NOT STORED: any "published" flag. Whether a row has been
-- published is a fact about its date and today's, so storing it would create a
-- second source of truth that can disagree with the first. `status` carries only
-- the EDITORIAL state, which is not derivable from anything.
--
-- Provenance columns (source/source_ref/gate_verdict/gated_at) are here from the
-- start because the next phase is Grok-generated provocations behind an
-- automated filter with no human approving each one, and that filter is required
-- to log what it rejects. A rejected row must be readable afterwards, which means
-- it has to be a row.

BEGIN;

DO $guard$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables
               WHERE table_schema = 'public' AND table_name = 'provocations') THEN
        RAISE EXCEPTION '282: provocations table already exists — migration already applied';
    END IF;
END
$guard$;

CREATE TABLE provocations (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),

    -- No default. The exporter this feeds refuses to run without an explicit
    -- domain, and a default here would quietly re-introduce the guess.
    domain       text NOT NULL,
    slug         text NOT NULL,

    -- Categories are coming (politics -> pets, each with its own audience), so
    -- content is taggable from today. NOTE the engine currently supports exactly
    -- one live provocation per site: internal/tools-api/handlers/round.go reads a
    -- single `today` key. Until that changes, selection ignores category and the
    -- column is metadata. Making it a selection axis is a change to the ENGINE's
    -- contract, not to this table.
    category     text NOT NULL DEFAULT 'general',

    -- NULL = in the pool, written but not scheduled. Selection requires a date.
    publish_on   date,

    -- Editorial state only. 'approved' is the sole state that can be published;
    -- everything else is invisible to the feed builder, so an unreviewed or
    -- rejected provocation cannot reach the site by any path.
    status       text NOT NULL DEFAULT 'draft'
                 CHECK (status IN ('draft', 'approved', 'rejected', 'retired')),

    -- Archive shape. Always required: an entry with neither cannot render.
    title        text NOT NULL,
    teaser       text NOT NULL,
    card_desc    text,              -- longer lobby-card blurb where one is written
    detail_body  text,              -- the full case; absent = renders non-openable

    -- Today shape. Long-form. Absent for the eight entries authored as archive
    -- rows before rotation existed; the exporter falls back to title/detail_body
    -- exactly as the Python builder does, and that fallback is documented rather
    -- than hidden. Anything added from here on should author both.
    headline     text,
    body         text,

    -- Provenance and the automated gate's decision.
    source       text NOT NULL DEFAULT 'human',
    source_ref   text,
    gate_verdict jsonb,
    gated_at     timestamptz,

    created_at   timestamptz NOT NULL DEFAULT now(),
    updated_at   timestamptz NOT NULL DEFAULT now(),

    UNIQUE (domain, slug)
);

-- Rule 2 above. Scoped to approved rows so that drafts may be parked on a date
-- speculatively; the collision then surfaces at approval time, loudly, instead of
-- silently deciding a day's provocation by plan order.
CREATE UNIQUE INDEX idx_provocations_one_per_day
    ON provocations (domain, publish_on)
    WHERE publish_on IS NOT NULL AND status = 'approved';

-- The exporter's read path: approved, dated, for one domain, in date order.
CREATE INDEX idx_provocations_schedule
    ON provocations (domain, publish_on)
    WHERE status = 'approved' AND publish_on IS NOT NULL;

COMMENT ON TABLE provocations IS
    'Pool + schedule for daily provocations. Read by render_provocation_feed. '
    'today = latest approved row with publish_on <= current_date; archive = '
    'everything approved and strictly earlier.';

-- ---------------------------------------------------------------------------
-- The nine we already have.
--
-- GENERATED from provocation_pipeline/builder/build_provocations.py's SCHEDULE,
-- not transcribed — nine long bodies carrying curly quotes, em-dashes and
-- apostrophes are precisely where a hand-copy goes wrong in a way that reads
-- fine. Dollar-quoted for the same reason.
--
-- All nine are 'approved': they are already live on the site, so calling them
-- anything else would make the migration's own state disagree with what is
-- being served.
-- ---------------------------------------------------------------------------

INSERT INTO provocations
    (domain, slug, category, publish_on, status, title, teaser, card_desc,
     headline, body, detail_body, source)
VALUES
    ('vonc.com', $prov$group-chats-replaced-friendship$prov$, 'general', DATE $prov$2026-06-28$prov$, 'approved',
     $prov$Group chats replaced friendship maintenance$prov$,
     $prov$Presence without effort. The bar has never been lower.$prov$,
     NULL,
     NULL,
     NULL,
     NULL,
     'human'),
    ('vonc.com', $prov$nobody-reads-terms-of-service$prov$, 'general', DATE $prov$2026-06-29$prov$, 'approved',
     $prov$Nobody actually reads terms of service — and that's rational$prov$,
     $prov$The cost of reading outweighs the power to change anything.$prov$,
     NULL,
     NULL,
     NULL,
     $prov$Reading takes an hour. Understanding takes a lawyer. Refusing takes the service away. Given those three prices, not reading is the correct decision, and every study that frames it as apathy has mistaken a rational calculation for a character flaw.

Which moves the burden somewhere else entirely. If consent is only ever given unread, then consent is not the thing doing the work, and we should stop pretending that it is.$prov$,
     'human'),
    ('vonc.com', $prov$four-day-week-productivity-myth$prov$, 'general', DATE $prov$2026-06-30$prov$, 'approved',
     $prov$The four-day week is a productivity myth$prov$,
     $prov$The pilots that prove it were self-selected true believers.$prov$,
     NULL,
     NULL,
     NULL,
     $prov$The pilots recruit organisations that already believed, run them for six months with everyone watching, and measure self-reported output. That is a design which cannot return a negative result. It tells you what motivated people do under observation, not what a four-day week does.

The counter is that the effect may well be real regardless, and demanding a hostile trial of something people obviously want is its own motivated reasoning. Possibly. Run it on a sceptical workforce for two years and the argument ends.$prov$,
     'human'),
    ('vonc.com', $prov$fiction-makes-you-worse-at-facts$prov$, 'general', DATE $prov$2026-07-01$prov$, 'approved',
     $prov$Reading fiction makes you worse at facts$prov$,
     $prov$Narrative trains you to want a tidy arc. Reality doesn't have one.$prov$,
     NULL,
     NULL,
     NULL,
     $prov$A novel teaches you to expect that events connect, that behaviour has motive, and that the ending explains the beginning. None of that is true of a pandemic, an election or a market. The better you get at narrative, the more confidently you impose one.

Against that: fiction is the main way most people practise holding a mind that is not their own, which is hardly nothing when the facts in dispute are about other people. Perhaps the trade is worth making. But it is a trade, and it is almost always sold as a free gain.$prov$,
     'human'),
    ('vonc.com', $prov$data-driven-decisions-arent$prov$, 'general', DATE $prov$2026-07-02$prov$, 'approved',
     $prov$Most 'data-driven' decisions aren't$prov$,
     $prov$The numbers get picked after the gut already chose.$prov$,
     NULL,
     NULL,
     NULL,
     $prov$Watch the sequence. Someone forms a view, then commissions the analysis, then reads the analysis for the part that agrees. The dashboard is not an input to the decision. It is the receipt.

The defence is that this still beats nothing — that even a motivated search for evidence occasionally turns up the number that stops you. Fair enough. But then say that is what the dashboard is for, and stop calling the output data-driven.$prov$,
     'human'),
    ('vonc.com', $prov$privacy-is-already-over$prov$, 'general', DATE $prov$2026-07-03$prov$, 'approved',
     $prov$Privacy is already over$prov$,
     $prov$You traded it years ago. The fight now is who profits.$prov$,
     NULL,
     NULL,
     NULL,
     $prov$You cannot claw back a decade of location history, contact graphs and purchase records by changing a setting. The data exists, it has been copied, and the copies are the asset. Every privacy control shipped since governs what happens next, never what already happened.

So the honest question stops being whether privacy survives and becomes who is permitted to profit from its absence. That is a distribution argument rather than a technical one, and it has an entirely different set of winners.$prov$,
     'human'),
    ('vonc.com', $prov$remote-work-killed-mentorship$prov$, 'general', DATE $prov$2026-07-04$prov$, 'approved',
     $prov$Remote work killed mentorship$prov$,
     $prov$You can't absorb judgement over a video call.$prov$,
     NULL,
     NULL,
     NULL,
     $prov$Judgement is not transferred in meetings. It is absorbed in the two minutes after one — the aside, the raised eyebrow, the way someone rewrites your paragraph while you watch. None of those moments has an agenda item, so no scheduled call contains them.

The rebuttal is that this was always a story senior people told about their own value. Plenty of people learned their craft alone, from documents, badly lit, and turned out fine. So which is it: a genuine transmission loss, or nostalgia for the office as a stage?$prov$,
     'human'),
    ('vonc.com', $prov$ai-never-funny-on-purpose$prov$, 'general', DATE $prov$2026-07-05$prov$, 'approved',
     $prov$AI will never be funny on purpose$prov$,
     $prov$The machine can recombine a million jokes and still not know why any land.$prov$,
     NULL,
     NULL,
     NULL,
     $prov$A model can hold every joke ever written and still not know which one to tell. Humour is a social risk instrument: it needs a target, a shared assumption to break, and a real chance of the room going cold. A system tuned never to offend and never to fail has removed all three ingredients before it starts.

The counter-case is that funniness is only a pattern in the data, and the machine is a better pattern-finder than you are. If that holds, the failure is temporary and the punchlines improve. If it does not, then everything an AI has ever produced that made you laugh was written by a person it read.$prov$,
     'human'),
    ('vonc.com', $prov$nobody-wants-personalised-internet$prov$, 'general', DATE $prov$2026-07-26$prov$, 'approved',
     $prov$Nobody actually wants a personalised internet$prov$,
     $prov$What gets sold as personalisation is the quiet removal of what you'd have shared with a stranger.$prov$,
     $prov$What gets sold as personalisation is mostly the quiet removal of whatever you'd have had in common with a stranger.$prov$,
     $prov$Nobody actually <em>wants</em> a personalised internet.$prov$,
     $prov$Every feed is tuned to one person, and every conversation now opens with “have you seen” and closes with a shrug. What gets sold as personalisation is mostly the quiet removal of whatever you would have had in common with a stranger. The engine is not serving you — it is dividing the room so each half can be sold separately.$prov$,
     NULL,
     'human')
;

-- Assert the seed landed as the exporter will read it. A migration that inserts
-- nothing, or inserts a row the selector cannot see, would otherwise look like a
-- success and fail at the next scheduled run.
DO $verify$
DECLARE
    n_total    int;
    n_eligible int;
    latest     text;
BEGIN
    SELECT count(*) INTO n_total FROM provocations WHERE domain = 'vonc.com';
    SELECT count(*) INTO n_eligible FROM provocations
        WHERE domain = 'vonc.com' AND status = 'approved' AND publish_on IS NOT NULL;
    SELECT slug INTO latest FROM provocations
        WHERE domain = 'vonc.com' AND status = 'approved'
          AND publish_on IS NOT NULL AND publish_on <= CURRENT_DATE
        ORDER BY publish_on DESC LIMIT 1;

    IF n_total <> 9 THEN
        RAISE EXCEPTION '282: expected 9 seeded provocations, found %', n_total;
    END IF;
    IF n_eligible <> 9 THEN
        RAISE EXCEPTION '282: expected 9 selectable provocations, found %', n_eligible;
    END IF;
    -- The live feed serves this slug today. If the selector disagrees with the
    -- site, the port is wrong and it is better to find out here than after a
    -- scheduled job has published the disagreement.
    IF latest <> 'nobody-wants-personalised-internet' THEN
        RAISE EXCEPTION '282: selector picks % as today, but the live feed serves nobody-wants-personalised-internet', latest;
    END IF;
END
$verify$;

COMMIT;
