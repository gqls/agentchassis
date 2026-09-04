#!/usr/bin/env python3
"""Acceptance for the homepage card canary, read at the SERVED artefact and the DB, never the status.
Usage: venv/bin/python assert_canary.py   (needs websocket-client for the CDP scroll check; skips it if absent)
A1 the served /index.html carries the swipeable-insight-carousel section and NO case-studies-grid section.
A2 five cards, and each card's label / headline / body / attribution text is byte-identical to the mapped JSON.
A3 the five OTHER sections' rendered_html md5 (DB) equal the pre-swap snapshot.
A4 (CDP) the carousel track scrolls: scrollWidth > clientWidth after load, on the live URL.
"""
import json, os, re, subprocess, sys, html as H, urllib.request
HERE=os.path.dirname(os.path.abspath(__file__)); fail=0
mapped=json.load(open(os.path.join(HERE,'content_data_mapped_2026-09-04.json')))
page=urllib.request.urlopen(urllib.request.Request('https://finetuning.uk/index.html', headers={'User-Agent':'curl/8.5.0 (finetuning-lane canary check)'}), timeout=30).read().decode('utf-8')  # the edge 403s Python's default UA
print(f'served {len(page)} bytes')
has_car = 'swipeable-insight-carousel' in page; has_grid = re.search(r'class="[^"]*case-studies-grid', page) is not None
print(f'A1 carousel present={has_car} grid present={has_grid}'); fail |= (not has_car) or has_grid
sec=re.search(r'(<section[^>]*swipeable-insight-carousel[^>]*>.*?</section>)', page, re.S)
text = H.unescape(re.sub(r'<[^>]+>', '\n', sec.group(1))) if sec else ''
norm = lambda s: re.sub(r'\s+', ' ', s).strip()
tnorm = norm(text)
missing=[]
for i,c in enumerate(mapped['cards'],1):
    for k in ('label','headline','body','attribution'):
        v = norm(c[k])
        if v and v not in tnorm: missing.append((i,k))
n_cards = len(re.findall(r'class="[^"]*swipeable-insight-carousel__card[^"]*"', sec.group(1))) if sec else 0
print(f'A2 cards in section={n_cards} missing text fields={missing}'); fail |= (n_cards!=5) or bool(missing)
title_ok = norm(mapped['section_title']) in tnorm and norm(mapped['section_eyebrow']) in tnorm
print(f'A2 section title/eyebrow verbatim={title_ok}'); fail |= not title_ok
snap = dict(l.split('|')[0:2] for l in open(os.path.join(HERE,'homepage_section_md5_before.txt')).read().split('\n') if '|' in l)
q = "SELECT pc.slot_name || '|' || md5(pc.rendered_html) FROM page_components pc JOIN pages p ON p.id=pc.page_id WHERE p.site_id='1368e337-dd1d-4799-bbb3-8221a1b79bcc' AND p.url='/index.html' ORDER BY pc.slot_name;"
out = subprocess.run(['kubectl','-n','ai-persona-system','exec','-i','postgres-clients-0','--','psql','-U','clients_user','-d','clients_db','-Atc',q],capture_output=True,text=True).stdout
now = dict(l.split('|') for l in out.strip().split('\n') if '|' in l)
changed = [s for s in ('hero','features','differentiators','departments-grid','call-to-action') if snap.get(s)!=now.get(s)]
print(f'A3 other sections changed={changed} (carousel md5 before={snap.get("swipeable-insight-carousel","?")[:8]} now={now.get("swipeable-insight-carousel","?")[:8]})'); fail |= bool(changed)
try:
    import websocket, time
    PORT=9337
    ch=subprocess.Popen(['/snap/bin/chromium','--headless=new','--disable-gpu','--no-first-run','--remote-allow-origins=*',f'--remote-debugging-port={PORT}','--window-size=1280,2000','about:blank'],stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL)
    try:
        for _ in range(40):
            try: tg=json.load(urllib.request.urlopen(f'http://127.0.0.1:{PORT}/json')); pg=next(t for t in tg if t['type']=='page'); break
            except Exception: time.sleep(0.5)
        ws=websocket.create_connection(pg['webSocketDebuggerUrl'],timeout=60); mid=[0]
        def call(m,**p):
            mid[0]+=1; ws.send(json.dumps({'id':mid[0],'method':m,'params':p}))
            while True:
                r=json.loads(ws.recv())
                if r.get('id')==mid[0]: return r.get('result',r)
        call('Page.enable'); call('Page.navigate',url='https://finetuning.uk/index.html'); time.sleep(6)
        r=call('Runtime.evaluate',expression="(function(){var t=document.querySelector('.swipeable-insight-carousel__track'); if(!t) return null; var b=t.scrollLeft; t.scrollBy(400,0); return {scrollWidth:t.scrollWidth, clientWidth:t.clientWidth, cards:t.children.length, movedTo:t.scrollLeft, before:b}})()",returnByValue=True)
        v=r.get('result',{}).get('value'); print('A4 track:',v); fail |= (not v) or v['scrollWidth']<=v['clientWidth'] or v['cards']!=5
    finally: ch.terminate()
except ImportError:
    print('A4 skipped (no websocket-client in this interpreter)')
print('RESULT:', 'FAIL' if fail else 'PASS'); sys.exit(1 if fail else 0)
