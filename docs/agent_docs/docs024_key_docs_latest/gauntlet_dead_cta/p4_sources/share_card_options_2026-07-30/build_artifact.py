#!/usr/bin/env python3
"""Build the owner's decision page, embedding every mock as a data URI.

The Artifact CSP blocks external hosts, so the images must be inlined.
"""
import base64, pathlib

def uri(path):
    return "data:image/png;base64," + base64.b64encode(pathlib.Path(path).read_bytes()).decode()

IMG = {k: uri(v) for k, v in {
    "baseline":  "opt_mock0.png",
    "whole":     "opt_mock1_whole_debate.png",
    "exchange":  "opt_mock1b.png",
    "debate":    "opt_mock2b.png",
    "hook":      "opt_mock3a.png",
    "timeline":  "opt_timeline_comparison.png",
    "record":    "opt_mock3b_record.png",
}.items()}

HTML = """<title>The share card and the full debate — a choice</title>
<style>
  :root{
    --ground:#faf9fc; --surface:#ffffff; --raised:#f3f1f8; --line:#ddd8e8;
    --ink:#16131f; --ink-2:#4a4459; --muted:#6f6883;
    --primary:#6d28d9; --primary-ink:#5b21b6; --amber:#b45309; --amber-flat:#f59e0b;
    --good:#047857; --bad:#b91c1c;
    --display:"Syne","Space Grotesk",system-ui,-apple-system,sans-serif;
    --body:"DM Sans","Inter",system-ui,-apple-system,sans-serif;
    --quote:Georgia,"Times New Roman",serif;
    --mono:ui-monospace,"SFMono-Regular",Menlo,Consolas,monospace;
  }
  @media (prefers-color-scheme:dark){
    :root{
      --ground:#0a0a0f; --surface:#13121f; --raised:#1b1930; --line:#2a2640;
      --ink:#f0eeff; --ink-2:#c9c3e0; --muted:#8b85b0;
      --primary:#a78bfa; --primary-ink:#c4b5fd; --amber:#f59e0b;
      --good:#34d399; --bad:#f87171;
    }
  }
  :root[data-theme="dark"]{
    --ground:#0a0a0f; --surface:#13121f; --raised:#1b1930; --line:#2a2640;
    --ink:#f0eeff; --ink-2:#c9c3e0; --muted:#8b85b0;
    --primary:#a78bfa; --primary-ink:#c4b5fd; --amber:#f59e0b;
    --good:#34d399; --bad:#f87171;
  }
  :root[data-theme="light"]{
    --ground:#faf9fc; --surface:#ffffff; --raised:#f3f1f8; --line:#ddd8e8;
    --ink:#16131f; --ink-2:#4a4459; --muted:#6f6883;
    --primary:#6d28d9; --primary-ink:#5b21b6; --amber:#b45309;
    --good:#047857; --bad:#b91c1c;
  }

  *{box-sizing:border-box;min-width:0}
  body{margin:0;background:var(--ground);color:var(--ink);font-family:var(--body);
       font-size:17px;line-height:1.62;-webkit-font-smoothing:antialiased}
  .wrap{max-width:830px;margin:0 auto;padding:0 24px 96px}

  /* ── masthead ── */
  .top{border-bottom:1px solid var(--line);background:var(--surface)}
  .top .wrap{padding:26px 24px 24px;display:flex;flex-direction:column;gap:8px}
  .kicker{font-family:var(--mono);font-size:11.5px;letter-spacing:.16em;
          text-transform:uppercase;color:var(--amber);font-weight:700}
  h1{font-family:var(--display);font-size:clamp(27px,4.6vw,40px);line-height:1.1;
     letter-spacing:-.025em;font-weight:800;margin:0;text-wrap:balance}
  .standfirst{color:var(--ink-2);font-size:18.5px;margin:6px 0 0;max-width:62ch}
  .byline{font-family:var(--mono);font-size:12px;color:var(--muted);
          letter-spacing:.04em;margin-top:6px}

  /* ── structure ── */
  section{margin-top:56px}
  h2{font-family:var(--display);font-size:23px;font-weight:800;letter-spacing:-.015em;
     margin:0 0 4px;text-wrap:balance}
  h2 .n{font-family:var(--mono);font-size:12px;font-weight:700;color:var(--amber);
        letter-spacing:.12em;display:block;margin-bottom:9px;text-transform:uppercase}
  h3{font-family:var(--display);font-size:18px;font-weight:700;margin:0 0 6px;
     letter-spacing:-.01em}
  p{margin:0 0 15px;max-width:66ch}
  p:last-child{margin-bottom:0}
  a{color:var(--primary);text-decoration-thickness:1px;text-underline-offset:2px}
  a:focus-visible{outline:2px solid var(--primary);outline-offset:2px;border-radius:2px}
  strong{color:var(--ink);font-weight:700}
  em.q{font-family:var(--quote);font-style:italic;font-size:18.5px;color:var(--ink)}
  code{font-family:var(--mono);font-size:.87em;background:var(--raised);
       padding:.12em .38em;border-radius:3px;color:var(--ink-2)}
  .lede{font-size:19px;color:var(--ink-2)}

  /* ── the headline number ── */
  .figures{display:grid;grid-template-columns:repeat(auto-fit,minmax(178px,1fr));
           gap:1px;background:var(--line);border:1px solid var(--line);
           border-radius:10px;overflow:hidden;margin:26px 0}
  .fig{background:var(--surface);padding:18px 20px 16px}
  .fig b{display:block;font-family:var(--display);font-size:30px;font-weight:800;
         letter-spacing:-.03em;line-height:1.05;font-variant-numeric:tabular-nums}
  .fig span{display:block;font-size:13.5px;color:var(--muted);margin-top:5px;line-height:1.4}
  .fig.hero b{color:var(--primary)}
  .fig.warn b{color:var(--bad)}

  /* ── table ── */
  .scroll{overflow-x:auto;margin:22px 0;border:1px solid var(--line);border-radius:10px;
          background:var(--surface)}
  table{border-collapse:collapse;width:100%;font-size:14.5px;min-width:520px}
  th,td{text-align:left;padding:11px 16px;border-bottom:1px solid var(--line)}
  th{font-family:var(--mono);font-size:11px;letter-spacing:.09em;text-transform:uppercase;
     color:var(--muted);font-weight:700;background:var(--raised)}
  tbody tr:last-child td{border-bottom:none}
  td.num{font-variant-numeric:tabular-nums;font-family:var(--mono);font-size:13.5px}
  tr.fail td{color:var(--bad)} tr.pass td.verdict{color:var(--good);font-weight:700}

  /* ── mocks ── */
  figure{margin:22px 0 0}
  figure img{display:block;width:100%;height:auto;border:1px solid var(--line);
             border-radius:8px}
  figcaption{font-size:13.5px;color:var(--muted);margin-top:10px;line-height:1.5;
             max-width:70ch}
  figcaption b{color:var(--ink-2)}

  /* ── option blocks ── */
  .opt{border:1px solid var(--line);border-radius:12px;background:var(--surface);
       padding:24px 26px 26px;margin-top:24px}
  .opt.pick{border-color:var(--primary);box-shadow:0 0 0 1px var(--primary) inset}
  .opt-head{display:flex;align-items:baseline;gap:12px;flex-wrap:wrap;margin-bottom:12px}
  .tag{font-family:var(--mono);font-size:11px;font-weight:700;letter-spacing:.1em;
       text-transform:uppercase;padding:4px 9px;border-radius:20px;
       background:var(--raised);color:var(--muted);border:1px solid var(--line)}
  .tag.rec{background:var(--primary);color:#fff;border-color:transparent}
  .opt-head h3{margin:0}
  .ledger{display:grid;gap:1px;background:var(--line);border:1px solid var(--line);
          border-radius:8px;overflow:hidden;margin:18px 0 0}
  .ledger div{background:var(--surface);padding:11px 15px;display:flex;gap:14px;
              font-size:14.5px;align-items:baseline}
  .ledger dt{font-family:var(--mono);font-size:11px;letter-spacing:.08em;
             text-transform:uppercase;color:var(--muted);flex:0 0 118px;font-weight:700}
  .ledger dd{margin:0;color:var(--ink-2)}
  .carries dt{color:var(--good)} .drops dt{color:var(--bad)}

  /* ── callouts ── */
  .rail{border-left:3px solid var(--amber);background:var(--raised);padding:16px 20px;
        border-radius:0 8px 8px 0;margin:24px 0;font-size:15.5px}
  .rail p{max-width:64ch;margin-bottom:10px}
  .rail .lab{font-family:var(--mono);font-size:11px;letter-spacing:.11em;font-weight:700;
             text-transform:uppercase;color:var(--amber);display:block;margin-bottom:7px}
  ul{margin:0 0 15px;padding-left:22px;max-width:66ch}
  li{margin-bottom:9px}
  li::marker{color:var(--muted)}
  .ask{border:1px solid var(--primary);border-radius:12px;padding:22px 26px;
       background:var(--surface);margin-top:22px}
  .ask ol{margin:0;padding-left:22px}
  .ask li{margin-bottom:13px}
  .ask li:last-child{margin-bottom:0}
  .marker{font-family:var(--mono);font-size:11px;font-weight:700;letter-spacing:.06em;
          color:var(--amber);border:1px solid var(--amber);padding:1px 5px;
          border-radius:3px;white-space:nowrap}
  hr{border:none;border-top:1px solid var(--line);margin:52px 0 0}
  @media (max-width:560px){
    .ledger div{flex-direction:column;gap:3px}
    .ledger dt{flex:none}
  }
</style>

<header class="top"><div class="wrap">
  <span class="kicker">vonc.com · the gauntlet · a decision, not a build</span>
  <h1>The share card carries a verdict about a stranger. Here is what it could carry instead.</h1>
  <p class="standfirst">Three options, mocked with one real round. The measurement
     decides more of this than taste does, so it comes first.</p>
  <p class="byline">30 July 2026 · every string in every mock is from round
     39595461, argued that afternoon</p>
</div></header>

<div class="wrap">

<section>
  <h2><span class="n">The problem, at real size</span>What travels today</h2>
  <p>This is the card the button produces right now, reproduced from the shipped
     renderer. It carries the provocation headline, the judge's ruling as two
     words, and the address.</p>
  <figure>
    <img src="__BASELINE__" alt="The current share card: a provocation headline, the words 'opponent wins', and a URL, with the lower half of the card empty.">
    <figcaption><b>The whole argument is the phrase &ldquo;opponent wins&rdquo; — 13 characters.</b>
      The round behind it ran to 3,357. Note also that roughly half the card is
      unused: whatever we do here, we are not fighting for space that is already spent.</figcaption>
  </figure>
</section>

<section>
  <h2><span class="n">The measurement</span>A real debate is 3,109 characters. A legible card holds about 700.</h2>
  <p>Measured on the island database over every complete round — 51 of them, 25 to
     30 July. A round is six pieces of prose: the provocation, the visitor's
     position, vonc's counter-argument, its challenge, the visitor's defence, and
     the judge's reasons.</p>
  <div class="figures">
    <div class="fig hero"><b>3,109</b><span>characters in an average complete round<br>(min 2,396 · max 5,073)</span></div>
    <div class="fig"><b>51</b><span>complete rounds stored, 25&ndash;30 July</span></div>
    <div class="fig warn"><b>~700</b><span>characters that fit legibly on one 1200&times;630 card</span></div>
    <div class="fig warn"><b>23%</b><span>of one round, at best, on a single card</span></div>
  </div>
  <p>The second figure is the one that settles it. A card posted to a timeline is
     displayed small and then scaled down, so type on the card shrinks with it.
     Here is the capacity of a 1200&times;630 card at each body size, measured with
     the real font metrics the renderer uses:</p>
  <div class="scroll"><table>
    <thead><tr><th>Body size on card</th><th>Effective size in a timeline</th>
      <th>Chars per line</th><th>Capacity</th><th>Share of a round</th><th class="verdict">Readable while scrolling?</th></tr></thead>
    <tbody>
      <tr class="fail"><td class="num">11px</td><td class="num">4.6px</td><td class="num">198</td><td class="num">6,336</td><td class="num">189%</td><td class="verdict">No</td></tr>
      <tr class="fail"><td class="num">16px</td><td class="num">6.7px</td><td class="num">140</td><td class="num">3,220</td><td class="num">96%</td><td class="verdict">No</td></tr>
      <tr class="fail"><td class="num">20px</td><td class="num">8.4px</td><td class="num">108</td><td class="num">1,944</td><td class="num">58%</td><td class="verdict">Barely</td></tr>
      <tr><td class="num">24px</td><td class="num">10.1px</td><td class="num">91</td><td class="num">1,365</td><td class="num">41%</td><td class="verdict">Marginal</td></tr>
      <tr class="pass"><td class="num">28px</td><td class="num">11.8px</td><td class="num">76</td><td class="num">988</td><td class="num">29%</td><td class="verdict">Yes</td></tr>
      <tr class="pass"><td class="num">32px</td><td class="num">13.4px</td><td class="num">67</td><td class="num">737</td><td class="num">22%</td><td class="verdict">Yes</td></tr>
      <tr class="pass"><td class="num">40px</td><td class="num">16.8px</td><td class="num">53</td><td class="num">477</td><td class="num">14%</td><td class="verdict">Comfortably</td></tr>
    </tbody>
  </table></div>
  <div class="rail">
    <span class="lab">One assumption, and why it does not change the answer</span>
    <p><span class="marker">ASSUMED</span> I have taken a timeline card as being
       displayed about 504px wide, which is the usual figure for X on desktop. I
       have not measured X itself. It does not matter: even if a card were shown at
       600px, the whole-debate version would render at 5.5px rather than 4.6px.
       The conclusion — <strong>no single card can hold a whole debate legibly</strong>
       — survives any plausible figure.</p>
  </div>
  <p>So every option below except the third is an <strong>excerpting</strong>
     strategy. That is the real choice, and it is not the choice the three options
     appear to offer.</p>
</section>

<section>
  <h2><span class="n">Option 1</span>The whole debate on the card</h2>
  <p>Taken literally, and auto-fitted rather than eyeballed: the largest type at
     which all 3,357 characters of the round actually fit is <strong>11px</strong>.</p>
  <figure>
    <img src="__WHOLE__" alt="A card with the entire debate on it in very small type, six labelled blocks filling the frame.">
    <figcaption>Legible on a desktop monitor at full size; roughly 4.6px once a
      timeline has scaled it down. This is not a card, it is a screenshot of a
      document.</figcaption>
  </figure>
  <p>The workable form of option 1 is therefore an excerpt. And here the handoff's
     own framing needs a correction, because it changes the cost:</p>
  <div class="rail">
    <span class="lab">Correction to the brief</span>
    <p>The brief warns that option 1 &ldquo;likely forces a <em>best exchange</em>
       excerpt, which is an editorial choice a machine will make badly&rdquo;.
       <strong>There is no such choice to make.</strong> A round has exactly one
       exchange by construction — one position, one counter, one challenge, one
       defence, one verdict, each a single column on the row. Nothing is being
       selected from a set. The only decision is which of the six fields to
       include, which is a design decision made once, not per round.</p>
  </div>
  <figure>
    <img src="__EXCHANGE__" alt="A card carrying vonc's challenge, the visitor's defence in serif type, and the ruling.">
    <figcaption><b>Option 1 as it would actually ship: the exchange.</b> vonc's
      challenge, the visitor's answer, the ruling. Auto-fitted to <b>26px</b> and
      comfortably readable. It carries 599 of the round's 3,357 characters — 18%.</figcaption>
  </figure>
  <div class="ledger carries"><div><dt>Carries</dt><dd>The challenge put, the defence given, the ruling</dd></div></div>
  <div class="ledger drops"><div><dt>Drops</dt><dd>vonc's 889-character counter-argument and the judge's 1,396-character reasons — the two longest and most persuasive pieces of the round</dd></div>
    <div><dt>Cost</dt><dd>Hours. One function in the page's JavaScript; no new URL, no new endpoint, no migration.</dd></div></div>
</section>

<section>
  <h2><span class="n">Option 2</span>Two cards</h2>
  <p>A second card gets a second budget, so it can carry the judge's reasoning that
     option 1 has to drop. Auto-fitted, the challenge, defence and reasons together
     land at <strong>16px</strong>.</p>
  <figure>
    <img src="__DEBATE__" alt="A denser second card carrying the challenge, the defence and the judge's full reasons.">
    <figcaption>1,995 characters — 59% of the round, and the most interesting 59%.
      But at 16px it is <b>readable only if someone taps the image open</b>; in the
      timeline it is a grey texture. It also still drops vonc's counter-argument.</figcaption>
  </figure>
  <div class="ledger"><div><dt>Two cards hold</dt><dd>About 1,540 characters between them at readable sizes — <strong>46% of one round</strong>. Two cards still do not hold a debate.</dd></div>
    <div><dt>Cost</dt><dd>Roughly double option 1: a second render path, and both must be kept honest.</dd></div>
    <div><dt>Workflow cost</dt><dd>You choose which card to post, every time. This is the part worth deciding deliberately — it is a recurring cost to you, not a one-off cost to us.</dd></div></div>
</section>

<section>
  <h2><span class="n">Option 3 · recommended</span>The card is a hook; a real page holds the round</h2>
  <p>Stop asking the card to be a container. Give it one job — earn a tap — and put
     the round somewhere that has room for it. The card can then use large type,
     because it is carrying one sentence rather than six paragraphs.</p>
  <figure>
    <img src="__HOOK__" alt="A card with the provocation headline, one quoted sentence from the judge, and a prominent 'Read the whole round' link.">
    <figcaption><b>The card, at 30px — the most legible of every option here.</b> It
      quotes the judge's own opening sentence, which is a fact of the round, and
      spends the space the current card wastes on making the link worth following.</figcaption>
  </figure>
  <figure>
    <img src="__RECORD__" alt="A dark permanent record page showing the provocation, the visitor's position, vonc's answer and challenge, the defence, and the judge's full ruling, ending with a call to argue the same provocation.">
    <figcaption><b>The page behind it, carrying all 3,357 characters without
      strain.</b> The whole round in order, the judgement in full, then an invitation
      to argue the same provocation. This is also the first thing on vonc.com that
      is permanent and linkable per round.</figcaption>
  </figure>
  <div class="ledger carries"><div><dt>Carries</dt><dd>Everything. 100% of the round, at reading size.</dd></div></div>
  <div class="ledger"><div><dt>Cost</dt><dd>Days, not weeks — and less than it looks. The round is <em>already</em> stored complete in one database row, and the read function to fetch it already exists.</dd></div>
    <div><dt>New parts</dt><dd>One read-only endpoint; one static page that fetches by id client-side, exactly as the gauntlet page already does; one opt-in publish flag; the card change.</dd></div>
    <div><dt>Workflow cost</dt><dd>None. One card, one link, nothing to choose when you post.</dd></div></div>
  <div class="rail">
    <span class="lab">One limitation, stated plainly</span>
    <p>A page that fetches its round in the browser cannot show a per-round preview
       image when the <em>link</em> is shared, because crawlers do not run
       JavaScript — the link would preview with the site's generic card. This does
       not affect the plan we actually have, where you post the PNG yourself and it
       travels as an image. Worth knowing before anyone treats the permalink as the
       primary thing to share.</p>
  </div>
</section>

<section>
  <h2><span class="n">All five, as a stranger would meet them</span>Scaled to timeline size, nothing else changed</h2>
  <figure>
    <img src="__TIMELINE__" alt="Five cards shown at 504 pixels wide: today's verdict-only card, the whole-debate card as illegible texture, the exchange card, the dense debate card, and the hook card.">
    <figcaption>The only operation applied here is the downscale a timeline
      performs. Two of these are readable while scrolling, two are not, and one
      says almost nothing.</figcaption>
  </figure>
</section>

<hr>

<section>
  <h2><span class="n">The recommendation</span>Option 3, reached through option 1</h2>
  <p class="lede">They are not rivals, and the ordering is free.</p>
  <p><strong>Ship the exchange card first</strong> — it is hours of work, it needs
     nothing new anywhere, and it replaces &ldquo;opponent wins&rdquo; with an actual
     argument today. <strong>Then build the record page</strong>, and the same card
     becomes the hook by gaining one line. Nothing built in the first step is thrown
     away by the second, and you have something better to post this week either way.</p>
  <p>I would not build option 2. It costs about twice option 1, still cannot hold a
     round, and the thing it adds is a decision you have to make every time you post
     — which is the one cost here that never amortises.</p>
</section>

<section>
  <h2><span class="n">What I need from you</span>Two questions I should not answer myself</h2>
  <div class="ask"><ol>
    <li><strong>Publication must be opt-in — confirm that is what you want.</strong>
       Rounds are written by whoever visits. Publishing one by default would put a
       stranger's writing on a public page without asking them, which I do not think
       you want on this site of all sites. The mock therefore shows a visitor
       pressing publish. It means most rounds will never have a page, and that the
       public record starts empty.</li>
    <li><strong>Should the private ledger link to the public record?</strong> The
       opinion ledger shipped on 29 July keeps your own rounds in your browser.
       If a round has been published, its ledger entry could link to the page. The
       brief raises this and it is genuinely your call — it is the difference
       between a private diary and a portfolio.</li>
  </ol></div>
  <div class="rail">
    <span class="lab">Rail check, and one uncomfortable fact</span>
    <p>Every string in every mock above is a real value from round 39595461. No
       counts, no rates, no leaderboard, no crowd — the fabrication classes stay
       deleted, and a public record with an opt-in flag grows only from real
       published rounds.</p>
    <p>The uncomfortable part: <code>count(DISTINCT client_ip_hash)</code> over all
       95 stored rounds is <strong>1</strong>. Every round in that table was produced
       by our own harnesses behind one proxy address. <strong>No stranger has ever
       argued on this page.</strong> That is a fact about distribution, not about the
       card — it is exactly what the experiment is meant to change — but it is why
       the public record must start empty rather than be seeded with what is already
       stored.</p>
  </div>
</section>

</div>
"""

for key, val in IMG.items():
    HTML = HTML.replace("__" + key.upper() + "__", val)

pathlib.Path("share_card_decision.html").write_text(HTML)
print("share_card_decision.html", len(HTML) / 1024, "KB")
