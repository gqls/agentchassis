# RUNBOOK — designblog.co.uk lane

## Fetch the served pages (the artefact, not the DB)

```bash
for p in "" "inspiration/index.html" "the-design-feed/index.html" \
         "tools/smart-contrast/index.html" "uk-studios-directory/index.html" \
         "glossary.html"; do
  curl -s -o "dl_$(echo "$p" | tr '/' '_').html" \
       -w "%{http_code} %{size_download} /$p\n" "https://designblog.co.uk/$p"
done
```

Gotcha (fleet-wide, from the mortgagecalculator lane): bare directory URLs 404
except root — always fetch the full `/index.html` path.

## Check a listing page actually lists something

Count content items, not prose. A page can carry 60KB of copy about its subject
and zero entries of it:

```bash
python3 - dl_glossary.html <<'EOF'
import sys, re
html = open(sys.argv[1]).read()
print("dt:", len(re.findall(r'<dt', html)))
print("article:", len(re.findall(r'<article', html, re.I)))
print("h3:", [re.sub(r'<[^>]+>|\s+',' ',x).strip()[:60]
              for x in re.findall(r'<h3[^>]*>(.*?)</h3>', html, re.S)])
print("img:", len(re.findall(r'<img\b', html)))
EOF
```

Gotcha: content `<h3>`s that are meta-headings ("What's covered", "How to use
a showcase") are the brief-echo signature — headings about the content the page
intends to have. Read them, don't just count them.

## Nav check

```bash
python3 - dl_.html <<'EOF'
import re
m = re.search(r'<nav[^>]*>.*?</nav>', open('dl_.html').read(), re.S)
print(re.findall(r'href="([^"]*)"', m.group(0)))
EOF
```

2026-09-02 baseline: 6 links (Home, The Feed, Criticism, Inspiration, Studios
Directory, Glossary), no Tools.
