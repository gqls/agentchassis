https://claude.ai/chat/6cfec7c0-d5b7-472c-b728-ebf1411a3171

# Port-forward
kubectl -n ai-persona-system port-forward svc/ollama-adapter 11434:11434 &

# Test 1: Classification (what site-classifier does)
curl -s http://localhost:11434/api/chat -d '{
"model": "mistral-small3.1",
"stream": false,
"messages": [{"role": "user", "content": "Classify this website project. Domain: vetcomparison.uk. Objective: Compare veterinary practices in the UK. Return ONLY valid JSON: {\"site_type\": \"landing|content|corporate|portfolio|tools\", \"confidence\": 0.0-1.0, \"detected_industry\": \"string\", \"reasoning\": \"brief explanation\"}"}]
}' | python3 -c "import sys,json; r=json.load(sys.stdin); print(r['message']['content'])"

# Test 2: Content generation (what page-content-writer does)
curl -s http://localhost:11434/api/chat -d '{
"model": "mistral-small3.1",
"stream": false,
"messages": [{"role": "user", "content": "Write content for a hero section of a veterinary comparison website called vetcomparison.uk. The tone is professional but approachable. Return ONLY valid JSON with these exact fields: {\"headline\": \"string\", \"subheadline\": \"string\", \"primary_cta\": \"string\", \"primary_cta_url\": \"/contact.html\"}"}]
}' | python3 -c "import sys,json; r=json.load(sys.stdin); print(r['message']['content'])"

# Test 3: Design spec (what webdesign-agent does)
curl -s http://localhost:11434/api/chat -d '{
"model": "mistral-small3.1",
"stream": false,
"messages": [{"role": "user", "content": "You are a web design expert. Create a color scheme and typography for a UK veterinary practice comparison website. Return ONLY valid JSON: {\"color_scheme\": {\"primary\": \"#hex\", \"secondary\": \"#hex\", \"accent\": \"#hex\", \"background\": \"#ffffff\", \"text\": \"#333333\"}, \"typography\": {\"font_family\": \"font stack\", \"heading_font\": \"font\"}, \"design_notes\": \"brief explanation\"}"}]
}' | python3 -c "import sys,json; r=json.load(sys.stdin); print(r['message']['content'])"

kill %1