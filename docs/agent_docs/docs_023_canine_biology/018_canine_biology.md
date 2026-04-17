# 018 — Canine Biology

## From Research to Embeddings to Training

---

## 1. What This Project Is

The canine biology knowledge base is the first real content for the RAG pipeline and the veterinary vertical. It produces three things:

1. **A populated knowledge_base** — hundreds of chunked, embedded knowledge entries in the `veterinary` collection that content writers can draw from via `rag_lookup` when building vet/pet sites
2. **Training data for text LoRA** — the research outputs (prompt → structured extraction) become fine-tuning examples for local models that can do the same extractions cheaper and faster
3. **Training data for image LoRA** — biological diagrams, pathway illustrations, and anatomical images become fine-tuning examples for an image model that can generate consistent scientific illustrations

The original 1M-agent design was aspirational. This plan is practical — focused on getting useful knowledge into the system and using it, not on hitting an agent count.

---

## 2. Scope: What to Build First

The full canine biology tree covers 20+ body systems and cross-cutting domains. For the first pass, focus on what directly serves the veterinary vertical's content needs (vetcomparison.uk and similar domains):

### Priority 1 — High content value, directly usable

| Topic Area | Why | Output |
|---|---|---|
| Breed health profiles (top 20 UK breeds) | Most searched vet content. "labrador health problems" etc. | 20 breed profiles with conditions, timelines, screening schedules |
| Common procedures and costs | Second most searched. "how much to neuter a dog" | 30-40 procedure guides with what's involved, recovery, cost ranges |
| Conditions and diseases (top 30) | Deep content for authority. Hip dysplasia, BOAS, IVDD, etc. | 30 condition guides with causes, symptoms, diagnosis, treatment |

### Priority 2 — Builds depth and authority

| Topic Area | Why | Output |
|---|---|---|
| Nutrition fundamentals | Underpins breed-specific feeding advice | Macronutrient needs, life-stage requirements, common dietary issues |
| Vaccination and preventive care | Every puppy owner searches this | Vaccination schedules, parasite prevention, health check timelines |
| Behaviour basics | Supports "choosing the right breed" content | Breed temperament profiles, common behavioural issues, training approaches |

### Priority 3 — Deep science (for authority differentiation)

| Topic Area | Why | Output |
|---|---|---|
| Genetics and genomics | Differentiates from surface-level content | Breed-specific genetic conditions, inheritance patterns, testing available |
| Pharmacology basics | Powers drug/treatment content | Common veterinary drugs, mechanisms, contraindications by breed |
| Anatomy by system | Reference material for all other content | Musculoskeletal, cardiovascular, nervous, digestive, respiratory |

---

## 3. Research Phase — How to Generate the Knowledge

### 3.1 Using LLM-as-Researcher

The most practical approach for the first pass: use Claude (via the Anthropic API, which your agents already call) to generate structured knowledge from its training data, then validate and index the output.

This is not the same as asking an LLM to "write an article about Labrador health." It's structured extraction with specific prompts designed to produce factual, citable knowledge chunks.

**For each breed health profile, the research prompt would be:**

```
You are a veterinary knowledge extraction specialist. Extract a structured 
health profile for the {breed_name} breed.

For each section, provide ONLY information you are confident about. 
Mark uncertain claims with [UNCERTAIN]. Do not invent statistics.

Return JSON with these sections:

1. breed_overview: Size, weight range, lifespan, breed group, origin
2. genetic_predispositions: Array of conditions this breed is predisposed to,
   each with: condition_name, prevalence (if known), age_of_onset, severity,
   inheritance_pattern (if known)
3. screening_schedule: Array of recommended health screenings by life stage
   (puppy 0-1yr, adult 1-7yr, senior 7+yr), each with: test_name, 
   recommended_age, frequency, what_it_detects
4. common_conditions_by_age: Array grouped by life stage, each condition with:
   name, typical_age_of_onset, symptoms, diagnosis_approach, treatment_options,
   prognosis, estimated_uk_treatment_cost_range
5. breed_specific_anaesthesia_notes: Any breed-specific anaesthesia 
   considerations (brachycephalic, sighthound sensitivity, etc.)
6. nutrition_notes: Breed-specific dietary considerations, common allergies,
   weight management notes
7. exercise_requirements: Daily exercise needs, joint protection considerations,
   age-appropriate activity changes

Sources to draw from: BSAVA Manual of Canine and Feline Clinical Pathology,
Kennel Club breed health surveys, published breed-specific studies.
```

**For each procedure/condition, a similar structured extraction prompt.**

The key difference from generic content generation: the output is structured JSON with confidence markers, not flowing prose. The knowledge base stores the structured data; the content writers later turn it into readable pages with appropriate tone and context.

### 3.2 Augmenting with Web Research

For data the LLM's training doesn't cover well (current UK vet pricing, recent breed health survey results, specific RCVS guidelines), use the existing web scrape pipeline:

1. Research agent searches for authoritative sources (Kennel Club breed pages, BSAVA guidelines, RCVS practice standards)
2. Scrape agent fetches the content
3. LLM extraction agent processes the raw content into structured knowledge
4. `rag_index` stores the results

This is the same pipeline the vertical research orchestrator will use later — you're just running it manually for the first batch.

### 3.3 Validation

LLM-extracted veterinary knowledge needs validation before it enters the knowledge base as authoritative content. For the first pass:

- **Self-consistency check**: Generate the same extraction twice with different prompts. Compare outputs. Flag discrepancies.
- **Cross-reference check**: For numerical claims (prevalence rates, cost ranges), search the web for corroboration. If a claim can't be verified, mark it `source_authority: 2` (lower confidence).
- **Structural check**: Does the JSON parse? Are all required fields present? Are cost ranges plausible (not £5 for surgery or £50,000 for a vaccination)?

Later, expert review from a vet would raise the authority level of verified content to 4-5. For now, LLM-generated content starts at authority 3 ("industry body level" — reasonable but not externally validated).

---

## 4. Indexing Phase — Getting Knowledge into the RAG Pipeline

### 4.1 Preparing Content for rag_index

The structured JSON from the research phase needs to be converted into text chunks suitable for embedding and retrieval. Each chunk should be:

- **Self-contained**: Makes sense without context (a reader who sees just this chunk understands what it says)
- **200-500 words**: Long enough to contain useful information, short enough for the embedding model to capture the meaning
- **Tagged with metadata**: Collection, topic, sub-topic, breed (if applicable), source authority, knowledge type

**Example: Converting a breed profile to chunks**

The Labrador breed profile JSON might produce these chunks:

```
Chunk 1 (collection: veterinary, knowledge_type: factual, topic: breed_health, breed: labrador):
"Labrador Retrievers are large breed dogs weighing 25-36kg with a typical 
lifespan of 10-14 years. They are the most popular breed in the UK. Common 
genetic predispositions include hip dysplasia (affecting approximately 12-15% 
of the breed), elbow dysplasia, progressive retinal atrophy (PRA), exercise-
induced collapse (EIC), and centronuclear myopathy (CNM). The Kennel Club 
recommends hip scoring, elbow scoring, and PRA/CNM DNA testing before breeding."

Chunk 2 (collection: veterinary, knowledge_type: procedural, topic: screening, breed: labrador):
"Recommended screening schedule for Labrador Retrievers: Puppy (0-1yr) — 
initial health check, vaccination course, microchipping, first worming 
programme. Adult (1-7yr) — annual health check including weight assessment 
(Labradors are prone to obesity), annual booster vaccinations, dental check, 
hip and joint assessment from age 2. Senior (7+yr) — biannual health checks, 
blood panel including thyroid function (hypothyroidism common in older Labs), 
joint assessment, eye examination for cataracts, cardiac auscultation."

Chunk 3 (collection: veterinary, knowledge_type: pricing, topic: procedure_costs, breed: labrador):
"Common procedure costs for Labradors (UK estimates, 2025): Neutering 
£200-400 (males cheaper than females due to less invasive procedure). Hip 
replacement £4,000-6,000 per hip. Cruciate ligament surgery (TPLO) £2,500-
4,500. PRA DNA test £50-80. Annual vaccination booster £50-80. Dental 
cleaning under anaesthetic £250-600 depending on extractions needed. Lump 
removal (lipoma) £300-800 depending on size and location."
```

### 4.2 Running rag_index

You can index content in three ways:

**Option A: Direct SQL insert (fastest for initial batch)**

```sql
INSERT INTO knowledge_base (
    collection, industry, title, content, content_hash,
    source_type, source_authority, metadata
) VALUES (
    'veterinary', 'veterinary',
    'Labrador Retriever — Breed Health Profile',
    'Labrador Retrievers are large breed dogs weighing 25-36kg...',
    encode(sha256('Labrador Retrievers are large breed dogs...'::bytea), 'hex'),
    'llm_extraction', 3,
    '{"breed": "labrador", "topic": "breed_health", "knowledge_type": "factual"}'
);
```

This skips embedding — you'd need to embed these chunks separately (see 4.3). But it gets content into the table immediately and it's searchable via trigram fallback.

**Option B: Via rag_index action in a workflow**

Create a simple "knowledge-seeder" agent with a workflow that takes prepared content and runs it through `rag_index`. This handles chunking, hashing, embedding, and storage automatically.

**Option C: Script that calls Ollama directly**

A Python or Go script that reads prepared content files, calls Ollama for embeddings, and inserts into Postgres. Most flexible for batch operations.

### 4.3 Embedding Existing Content

If you insert content via SQL (Option A), you need to generate embeddings separately. A script to embed all un-embedded chunks:

```python
import psycopg2
import requests
import json

conn = psycopg2.connect("dbname=clients_db user=clients_user")
cur = conn.cursor()

# Find chunks without embeddings
cur.execute("""
    SELECT id, content FROM knowledge_base 
    WHERE embedding IS NULL 
    AND collection = 'veterinary'
    LIMIT 100
""")

for row_id, content in cur.fetchall():
    # Call Ollama for embedding
    resp = requests.post(
        "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434/api/embeddings",
        json={"model": "nomic-embed-text", "prompt": content}
    )
    embedding = resp.json()["embedding"]
    
    # Update the row
    embedding_str = "[" + ",".join(str(f) for f in embedding) + "]"
    cur.execute(
        "UPDATE knowledge_base SET embedding = %s::vector WHERE id = %s",
        (embedding_str, row_id)
    )
    conn.commit()

conn.close()
```

Run this from inside the cluster (or port-forward Ollama). After embedding, vector similarity search will work for these chunks.

### 4.4 Verifying the Knowledge Base

After indexing:

```sql
-- How much is in the veterinary collection?
SELECT * FROM knowledge_base_stats WHERE collection = 'veterinary';

-- Check embedding coverage
SELECT 
    COUNT(*) as total,
    COUNT(embedding) as embedded,
    COUNT(*) - COUNT(embedding) as needs_embedding
FROM knowledge_base WHERE collection = 'veterinary';

-- Test a similarity search (requires at least one embedded chunk)
-- This uses a raw text query via trigram as a quick test
SELECT title, LEFT(content, 100), 
       similarity(content, 'labrador hip dysplasia') as sim
FROM knowledge_base
WHERE collection = 'veterinary'
  AND content % 'labrador hip dysplasia'
ORDER BY sim DESC
LIMIT 5;
```

---

## 5. Using the Knowledge — RAG in Content Generation

Once the knowledge base has content, wire it into the content generation pipeline.

### 5.1 Add rag_lookup to page-content-writer

Add a step before content generation in the page-content-writer workflow:

```json
"lookup_knowledge": {
    "action": "rag_lookup",
    "config": {
        "query_field": "current_page.rag_query",
        "collection_field": "current_page.rag_collection",
        "top_k": 5,
        "embedding_service": {
            "provider": "ollama",
            "model": "nomic-embed-text",
            "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
        }
    },
    "next_step": "generate_content",
    "output_field": "knowledge_context"
}
```

### 5.2 Update the content writer prompt

Add a conditional knowledge injection block:

```
{{if .knowledge_context.rag_context}}
## Authoritative Domain Knowledge

The following is verified knowledge relevant to this page. Use it to inform 
your content. Prefer this information over general knowledge. Include specific 
figures (costs, prevalence rates, timelines) where provided.

{{.knowledge_context.rag_context}}
{{end}}
```

### 5.3 Test the difference

Build the same page twice — once without RAG (empty collection), once with. Compare:
- Does the RAG version include specific figures the non-RAG version lacks?
- Does it reference specific conditions, procedures, or costs?
- Does it sound more authoritative and specific?

This is the validation that the entire vertical concept works.

---

## 6. Text LoRA Fine-Tuning

### 6.1 What to Fine-Tune and Why

Two candidates for text LoRA from this project:

**Candidate A: Veterinary knowledge extractor**

The LLM calls that extract structured knowledge from raw text (breed profiles, procedure details, condition guides). This is a well-defined task with consistent input/output format — ideal for fine-tuning.

- Input: Raw veterinary text (from scrapes or reference material)
- Output: Structured JSON with breed health data, procedure details, cost ranges
- Why fine-tune: This extraction runs repeatedly across breeds and conditions. A fine-tuned 7B model at ~10-30s per call on CPU costs nothing compared to Claude API calls at ~£0.01-0.05 each. At 500+ extractions, the savings justify the fine-tuning investment.

**Candidate B: Site classifier with vertical awareness**

The site-classifier agent that determines what vertical a domain belongs to. Short structured output, runs every build, high volume.

- Input: Domain name + objective
- Output: JSON with site_type, vertical_slug, disposition
- Why fine-tune: Runs on every domain in the portfolio. Currently uses Claude Sonnet 4.6 (upgraded this session). A fine-tuned 7B model that does this classification reliably saves API costs across hundreds of domains.

### 6.2 Collecting Training Data

The `llm_call_log` table is already collecting every LLM call. For fine-tuning, you need 200+ successful examples for the target task.

**Check readiness:**

```sql
SELECT agent_type, step_name, COUNT(*) as examples
FROM llm_call_log
WHERE success = true
GROUP BY agent_type, step_name
ORDER BY examples DESC;
```

**For the knowledge extractor**, you'll accumulate examples during the research phase itself. Every successful breed profile extraction is a training example. Process 50 breeds and you have 50 examples. Process 30 conditions and 40 procedures and you're at 120. You need 200+ for a reasonable fine-tune.

**For the site classifier**, examples accumulate from normal pipeline operation. Every domain that goes through classification is an example.

### 6.3 The Fine-Tuning Process

**Hardware needed**: A machine with a GPU — 16GB+ VRAM. Options:
- RTX 3090/4090 (24GB) — can fine-tune 7B-8B models comfortably with QLoRA
- Cloud GPU rental (RunPod, Lambda, Vast.ai) — ~$0.50-1.50/hour for a suitable instance
- Google Colab Pro — T4 GPU is tight but works for small models with QLoRA

**Framework**: Unsloth — 2x faster than vanilla HuggingFace, 70% less VRAM. Handles QLoRA, GGUF export, and Ollama integration.

**Step-by-step process:**

**Step 1: Export training data**

```sql
-- Export as JSONL (one example per line)
COPY (
    SELECT json_build_object(
        'instruction', 'Extract a structured veterinary knowledge profile from the following text.',
        'input', prompt_rendered,
        'output', response_text
    )
    FROM llm_call_log
    WHERE agent_type = 'knowledge-extractor'
      AND success = true
      AND response_text IS NOT NULL
      AND LENGTH(response_text) > 100
) TO '/tmp/vet_extractor_training.jsonl';
```

Transfer the file to your GPU machine.

**Step 2: Set up Unsloth environment**

```bash
# On GPU machine
pip install unsloth
# Or use Docker:
docker run -d -p 8888:8888 --gpus all \
    -v $(pwd)/data:/workspace/data \
    unsloth/unsloth
```

**Step 3: Fine-tune with QLoRA**

```python
from unsloth import FastLanguageModel
import torch

# Load base model (4-bit quantized for memory efficiency)
model, tokenizer = FastLanguageModel.from_pretrained(
    model_name="unsloth/llama-3.1-8b-instruct-bnb-4bit",
    max_seq_length=4096,
    load_in_4bit=True,
)

# Add LoRA adapters
model = FastLanguageModel.get_peft_model(
    model,
    r=16,                    # LoRA rank — 16 is good for focused tasks
    target_modules=["q_proj", "k_proj", "v_proj", "o_proj"],
    lora_alpha=16,
    lora_dropout=0,
    bias="none",
    use_gradient_checkpointing="unsloth",  # 30% less memory
)

# Load your training data
from datasets import load_dataset
dataset = load_dataset("json", data_files="vet_extractor_training.jsonl")

# Configure training
from trl import SFTTrainer
from transformers import TrainingArguments

trainer = SFTTrainer(
    model=model,
    tokenizer=tokenizer,
    train_dataset=dataset["train"],
    args=TrainingArguments(
        per_device_train_batch_size=2,
        gradient_accumulation_steps=4,
        num_train_epochs=3,
        learning_rate=2e-4,
        fp16=not torch.cuda.is_bf16_supported(),
        bf16=torch.cuda.is_bf16_supported(),
        output_dir="outputs",
        logging_steps=10,
    ),
    max_seq_length=4096,
)

trainer.train()
```

**Step 4: Export to GGUF for Ollama**

```python
# Save as GGUF (Q4_K_M is good balance of size vs quality)
model.save_pretrained_gguf(
    "vet-extractor-v1",
    tokenizer,
    quantization_method="q4_k_m"
)
```

This produces a ~4GB GGUF file.

**Step 5: Load into Ollama**

```bash
# Copy GGUF to the ollama-adapter PVC
kubectl cp vet-extractor-v1.gguf \
    ai-persona-system/ollama-adapter-xxx:/root/.ollama/models/

# Create Ollama model
kubectl -n ai-persona-system exec -it deploy/ollama-adapter -- \
    ollama create vet-extractor-v1 -f /path/to/Modelfile
```

The Modelfile:
```
FROM /root/.ollama/models/vet-extractor-v1.gguf
PARAMETER temperature 0.3
SYSTEM "You are a veterinary knowledge extraction specialist."
```

**Step 6: Update agent definition to use local model**

```sql
UPDATE agent_definitions
SET default_config = jsonb_set(
    default_config,
    '{ai_service}',
    '{
        "provider": "ollama",
        "model": "vet-extractor-v1",
        "api_url": "http://ollama-adapter.ai-persona-system.svc.cluster.local:11434"
    }'::jsonb
)
WHERE type = 'knowledge-extractor';
```

**Step 7: A/B test**

Run the same extraction with both Claude and the local model. Compare output quality. If the local model produces comparable structured extractions, it saves API costs for all future knowledge base building.

### 6.4 What Good Training Data Looks Like

Quality matters more than quantity. 500 clean examples beat 5,000 noisy ones.

For the vet knowledge extractor, a good training example has:
- **Input**: A clear prompt asking for structured extraction of a specific breed/condition/procedure
- **Output**: Well-structured JSON with specific figures, proper medical terminology, appropriate uncertainty markers
- **No hallucination**: Claims are verifiable or appropriately hedged

Filter your training data:

```sql
-- Remove examples with very short responses (probably errors)
-- Remove examples where JSON parsing failed
-- Remove examples with known hallucination patterns
SELECT COUNT(*) FROM llm_call_log
WHERE agent_type = 'knowledge-extractor'
  AND success = true
  AND LENGTH(response_text) > 500
  AND response_text LIKE '{%'  -- starts with JSON
  AND response_text NOT LIKE '%I cannot%'  -- not a refusal
```

---

## 7. Image LoRA Fine-Tuning

### 7.1 What to Train and Why

Scientific and biological illustrations have a distinctive visual style that general-purpose image generators handle poorly. Stable Diffusion and FLUX produce photorealistic images or artistic styles well, but struggle with:

- Clean anatomical diagrams with labelled structures
- Biochemical pathway diagrams
- Breed comparison illustrations
- Veterinary procedure step illustrations
- Consistent colour coding across a series of related diagrams

A LoRA trained on veterinary/biological illustration style could produce consistent, professional diagrams for every breed health page, every procedure guide, and every condition explanation.

### 7.2 Collecting Training Images

You need 20-100 images of the style you want to reproduce. For scientific/veterinary illustrations:

**Sources for training images:**

- **Open-access textbooks**: OpenStax anatomy textbooks, Wikimedia Commons anatomical illustrations (many are CC-licensed)
- **Generated baseline**: Use a capable image model (DALL-E, Midjourney, or FLUX) to generate 50-100 biological diagrams in your target style, then curate the best ones as training data. This is "style distillation" — teaching a smaller model to reproduce a style you've defined.
- **Your own image generator output**: If your current image-generator agent produces some good biological diagrams, save and curate them
- **Licensed illustration libraries**: Medical/veterinary illustration collections (may require licensing for training use)

**What the training set should contain:**

| Category | Examples | Count |
|---|---|---|
| Anatomical cross-sections | Heart, hip joint, spine, eye, ear | 10-15 |
| Breed silhouettes/profiles | Side view, proportions marked, size comparison | 10-15 |
| Pathway diagrams | Drug metabolism, immune response, genetic inheritance | 10-15 |
| Procedure illustrations | Neutering steps, dental procedure, surgery approach | 10-15 |
| Condition visualisations | Hip dysplasia vs normal, BOAS airway, IVDD disc | 10-15 |
| Infographic style | Cost comparisons, screening timelines, breed statistics | 10-15 |

Total: 60-90 images, curated for consistent style (same colour palette, line weight, label style, background treatment).

### 7.3 Image Captioning

Every training image needs a text caption describing what it shows. This teaches the model the relationship between text descriptions and visual output.

Caption format:
```
"A clean veterinary anatomical diagram showing a cross-section of a canine 
hip joint, with the femoral head, acetabulum, and joint capsule clearly 
labelled. Professional medical illustration style with blue and grey colour 
scheme, white background, clean lines, and sans-serif labels."
```

You can caption manually (most accurate for 60-90 images) or use a vision model (GPT-4V, Claude with vision) to generate captions and then review them.

### 7.4 The Image Fine-Tuning Process

**Model choice:**

- **FLUX.2 Dev** (32B parameter, open weights) — best current open model for this. Supports LoRA fine-tuning. Needs ~24GB VRAM for training. Produces very high quality output.
- **Stable Diffusion XL** — lighter alternative, needs less VRAM (~12GB), huge ecosystem of existing LoRA models. Quality is lower than FLUX but may be sufficient for diagrams.
- **PixArt-Sigma** — reportedly works well for vector-style graphics and diagrams. Smaller model, faster training.

For scientific diagrams, SDXL or PixArt may actually work better than FLUX, which excels at photorealism but can overdo detail in diagram contexts.

**Hardware:**

- RTX 3090/4090 (24GB VRAM) — works for SDXL LoRA, tight for FLUX
- Cloud A100 40GB — comfortable for any model. ~$1-2/hour rental.
- H100 — fast but expensive, only needed for FLUX full training

**Training with diffusers (HuggingFace):**

```bash
# For SDXL LoRA:
accelerate launch train_dreambooth_lora_sdxl.py \
    --pretrained_model_name_or_path="stabilityai/stable-diffusion-xl-base-1.0" \
    --instance_data_dir="./training_images" \
    --output_dir="./vet-diagram-lora" \
    --instance_prompt="a veterinary scientific diagram" \
    --resolution=1024 \
    --train_batch_size=1 \
    --gradient_accumulation_steps=4 \
    --learning_rate=1e-4 \
    --lr_scheduler="constant" \
    --lr_warmup_steps=0 \
    --max_train_steps=1000 \
    --rank=16
```

For FLUX LoRA, use the `diffusers` FLUX LoRA training script or SimpleTuner.

**Output:**

A LoRA weights file (~3-50MB depending on rank) that can be loaded alongside the base model. When generating images, you activate the LoRA and include your trigger phrase (e.g., "a veterinary scientific diagram") in the prompt.

### 7.5 Integrating Trained Image LoRA

The image-generator adapter in your cluster currently generates images via an external service or local model. To use a LoRA-trained model:

**Option A: Run locally via ComfyUI or diffusers**

Deploy a local inference pod with the base model + LoRA weights. The image-generator adapter routes requests to this pod instead of an external API. This gives full control but needs a GPU node in the cluster.

**Option B: Use Replicate or RunPod Serverless**

Upload the LoRA weights to a serverless inference platform. Your image-generator adapter calls their API with the LoRA model specified. Pay per image. No GPU infrastructure needed.

**Option C: Use Ollama (future)**

Ollama is adding image generation support. When available, you could load image models the same way you load text models. This is the simplest long-term path but not yet production-ready.

For the immediate term, Option B is most practical — train the LoRA on a rented GPU, upload to a serverless platform, update the image-generator adapter to reference the LoRA model.

---

## 8. Implementation Order

### Phase A: Generate knowledge (weeks 1-2)

- [ ] Write the breed health profile extraction prompt (section 3.1)
- [ ] Run extractions for top 20 UK breeds via Claude API (use the existing `execute_llm_prompt` action or a dedicated knowledge-extractor agent)
- [ ] Write the procedure/condition extraction prompts
- [ ] Run extractions for top 30 conditions and 40 procedures
- [ ] Run self-consistency validation on extractions (section 3.3)
- [ ] Prepare text chunks from the structured JSON (section 4.1)
- [ ] Total output: ~300-500 knowledge chunks

### Phase B: Index and embed (week 2-3)

- [ ] Insert chunks into knowledge_base (SQL or via rag_index action)
- [ ] Run embedding script to embed all chunks via Ollama (section 4.3)
- [ ] Verify with knowledge_base_stats and test similarity searches (section 4.4)
- [ ] Test rag_lookup returns relevant results for veterinary queries

### Phase C: Wire into content generation (week 3)

- [ ] Add rag_lookup step to page-content-writer workflow (section 5.1)
- [ ] Update content writer prompt with knowledge injection block (section 5.2)
- [ ] Build a test page with RAG and compare quality (section 5.3)
- [ ] If quality improvement is clear, wire into the planner so new vet-vertical pages include rag_collection and rag_query in their specs

### Phase D: Collect training data (weeks 3-6, runs in background)

- [ ] Monitor llm_call_log for extraction examples accumulating
- [ ] Monitor for site-classifier examples accumulating from normal pipeline use
- [ ] At 200+ examples per agent type, export training data (section 6.3)
- [ ] Review and clean training data (section 6.4)

### Phase E: Text LoRA fine-tuning (week 6-7)

- [ ] Rent GPU or use own hardware
- [ ] Set up Unsloth environment
- [ ] Fine-tune knowledge extractor model (section 6.3, steps 2-4)
- [ ] Export to GGUF, load into Ollama (steps 5-6)
- [ ] A/B test against Claude for extraction quality (step 7)
- [ ] If comparable, switch knowledge extraction to local model

### Phase F: Image LoRA fine-tuning (week 7-8)

- [ ] Curate 60-90 training images in target style (section 7.2)
- [ ] Caption all images (section 7.3)
- [ ] Choose base model (SDXL recommended for diagrams)
- [ ] Fine-tune LoRA on rented GPU (section 7.4)
- [ ] Test generated diagrams for consistency and quality
- [ ] Integrate into image-generator pipeline (section 7.5)

### Phase G: Expand to other verticals (week 8+)

- [ ] Apply the same research → index → embed → RAG → fine-tune pipeline to the energy vertical
- [ ] Then mortgage vertical
- [ ] Each vertical reuses the same infrastructure but with different research prompts, source material, and knowledge types

---

## 9. Costs and Resources

| Activity | Cost Estimate | Notes |
|---|---|---|
| Knowledge extraction (Claude API) | £15-40 | ~500 extractions at ~£0.03-0.08 each |
| Ollama embedding (running in cluster) | £0 | Already deployed, CPU inference |
| GPU rental for text LoRA | £5-15 | 2-4 hours on RTX 3090/4090 via RunPod |
| GPU rental for image LoRA | £10-30 | 4-8 hours on A100 for SDXL training |
| Image captioning (if using vision model) | £5-10 | 60-90 images via Claude Vision |
| Serverless image inference (ongoing) | £0.01-0.05/image | If using Replicate/RunPod for serving |

Total first-pass investment: roughly £35-95 for the complete pipeline from research through to trained models. The ongoing cost drops to near zero once local models are serving.

---

## 10. What Success Looks Like

After completing this plan:

1. **Knowledge base has 300-500 veterinary chunks** with embeddings, searchable via vector similarity
2. **Content writers produce measurably better veterinary pages** — with specific figures, proper medical terminology, and genuine depth that generic AI content lacks
3. **A fine-tuned 7B model** can extract structured veterinary knowledge locally at zero API cost per call
4. **A LoRA-trained image model** produces consistent veterinary diagrams that look professional and match across a series
5. **The pipeline for doing this in other verticals is proven** — the same steps work for energy, mortgage, gifts, or any new vertical you add

The canine biology project stops being aspirational and becomes the working proof that deep knowledge makes content sites genuinely competitive.


====================================
====================================

The plan covers the full path in 10 sections. Here's what's in each and some of the thinking behind the choices:
Scope (section 2) is deliberately practical rather than aspirational. The 1M-agent version from the earlier session was a demo concept. This plan produces 300-500 knowledge chunks across breed profiles, procedures, and conditions — the content that actually gets used by the veterinary vertical when building sites like vetcomparison.uk. The priorities are ordered by what people actually search for (breed health problems first, then procedure costs, then conditions).
Research (section 3) uses Claude via the existing API as the primary researcher, with structured extraction prompts that produce JSON rather than prose. The distinction matters — structured JSON gets chunked and indexed as facts the RAG system can retrieve. Prose would just be another article. Web scraping augments for current data (2025 pricing, recent breed survey results) that the LLM's training data doesn't cover well.
Embedding (section 4) gives three options for getting content into the knowledge base. Direct SQL is fastest for the initial batch. The Python embedding script handles the gap between "content in table" and "content with vector embedding." Once embedded, vector similarity search works.
Text LoRA (section 6) targets two candidates. The knowledge extractor is the better first target — well-defined input/output, consistent task, and the training data accumulates naturally during the research phase itself. Every breed profile extraction is a training example. By the time you've processed 50 breeds and 30 conditions, you're close to the 200-example threshold. The fine-tuning process itself is a day's work on rented hardware — Unsloth on an RTX 3090 via RunPod at about £5-15 total.
Image LoRA (section 7) is a different beast. You need 60-90 curated training images in a consistent style, each captioned. For scientific diagrams, SDXL is the recommended base over FLUX — FLUX excels at photorealism but tends to overdetail what should be clean line diagrams. The training images can come from open-source anatomy illustrations, style-distilled generations from a capable model, or a combination. The LoRA output is small (3-50MB) and can be served via serverless platforms without running a GPU in your cluster.
The total cost estimate is £35-95 for the complete first pass from research through trained models. The ongoing cost drops to near zero once local models are serving the extractions and Ollama handles the embeddings.

