Worth caveating: we don't know which versions are pre-installed. They might be more recent than the cu124/torch 2.6 stack we ended up on this morning, which could put us back in the torchao-incompatibility cascade we just escaped. Or they might be older and require updates. Either way, we don't trust the env blindly — first thing on the VM is still:
bashnvidia-smi              # confirm A100 80GB
source ~/unsloth_env/bin/activate 2>/dev/null || ls /opt/unsloth* /opt/venv* ~/  # find their venv path
python -c "
import torch, unsloth, transformers, peft, trl
print(f'torch:        {torch.__version__}')
print(f'unsloth:      {unsloth.__version__}')
print(f'transformers: {transformers.__version__}')
print(f'cuda avail:   {torch.cuda.is_available()}')
print(f'gpu:          {torch.cuda.get_device_name(0)}')
print(f'bf16:         {torch.cuda.is_bf16_supported()}')
"
If those numbers look reasonable (torch ≥2.6, unsloth recent, transformers <5, cuda+bf16 yes), skip the setup script entirely and go straight to upload + run. If anything's off, run a partial setup that fixes the specific issue rather than the whole script.
The big practical win regardless of pre-installed versions: the cached base model. Almost certain the Unsloth template ships with common base models pre-cached. If unsloth/Llama-3.3-70B-Instruct-bnb-4bit is already at ~/.cache/huggingface/, that's another 5-8 min of download skipped per run.
Updated workflow with the Unsloth template
bash# Local
mkdir -p ~/iter0_eval/

my configuration:
╭───────────────────────────╮
│                           │
│  Mode:       Prototyping  │
│  Template:   Unsloth      │
│  GPU Type:   A100 80GB    │
│  GPUs:       1            │
│  vCPUs:      4            │
│  RAM:        32 GB        │
│  Disk Size:  100 GB       │
│  Ephemeral:  0 GB         │
│                           │
╰───────────────────────────╯



# Create with Unsloth template this time
tnr create   # pick Unsloth template
tnr connect 0

# Sanity-check the env BEFORE uploading anything
nvidia-smi
ls ~/                     # likely shows existing venv / cached models
# run the python check above

# Conditional: only run setup script if the env doesn't satisfy our needs
# (probably won't need to)

# Upload (the 791MB adapter is still the slow part regardless of template)
# Run from your laptop in another terminal:
tnr scp held_out_cases_v1.jsonl  0:/home/ubuntu/
tnr scp 04_eval_iter0.py         0:/home/ubuntu/
tnr scp  lora_iter0_full       0:/home/ubuntu/lora_iter0_full
tnr scp 00_vm_setup.sh         0:/home/ubuntu/

[
ant@ant-XPS-15-9500:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning$ cd working/phase2/lora_iter0_full/
ant@ant-XPS-15-9500:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase2/lora_iter0_full$ gunzip adapter_model.safetensors.gz
ant@ant-XPS-15-9500:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase2/lora_iter0_full$ cd ..
ant@ant-XPS-15-9500:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase2$ tnr scp  lora_iter0_full 0:/home/ubuntu/lora_iter0_ful
]

## from tnr
chmod +x ~/00_vm_setup.sh
~/00_vm_setup.sh 2>&1 | tee ~/setup.log

# After setup, activate and run smoke
source ~/unsloth_env/bin/activate
export HF_HUB_ENABLE_HF_TRANSFER=1

# Run
source <whatever_venv_path>
python 04_eval_iter0.py --adapter ~/lora_iter0_full \
--cases ~/held_out_cases_v1.jsonl \
--output ~/iter0_eval_results_v1.jsonl \
--n 20

# Pull + delete
tnr scp 0:/home/ubuntu/iter0_eval_results_v1.jsonl ./
tnr delete 0

----------------------------------------

Updated workflow with the Unsloth template
bash# Local
mkdir -p ~/iter0_eval/

# Create with Unsloth template this time
tnr create   # pick Unsloth template
tnr connect 0

# Sanity-check the env BEFORE uploading anything
nvidia-smi
ls ~/                     # likely shows existing venv / cached models
# run the python check above

# Conditional: only run setup script if the env doesn't satisfy our needs
# (probably won't need to)

# Upload (the 791MB adapter is still the slow part regardless of template)
# Run from your laptop in another terminal:
tnr scp held_out_cases_v1.jsonl  0:/home/ubuntu/
tnr scp 04_eval_iter0.py         0:/home/ubuntu/
tnr scp lora_iter0_full       0:/home/ubuntu/lora_iter0_full

# Run
source <whatever_venv_path>
python 04_eval_iter0.py --adapter ~/lora_iter0_full \
--cases ~/held_out_cases_v1.jsonl \
--output ~/iter0_eval_results_v1.jsonl \
--n 20

# Pull + delete
tnr scp 0:/home/ubuntu/iter0_eval_results_v1.jsonl ./
tnr delete 0

export HF_HUB_ENABLE_HF_TRANSFER=1

ls /home/ubuntu/lora_iter0_full/                      # should show lora_iter0_full/
ls /home/ubuntu/lora_iter0_full/lora_iter0_full/      # should show adapter files

mv /home/ubuntu/lora_iter0_full/lora_iter0_full /home/ubuntu/lora_iter0_full
rmdir /home/ubuntu/lora_iter0_full/lora_iter0_full
ls -lh /home/ubuntu/lora_iter0_full/                 # confirm: README, adapter_config.json, adapter_model.safetensors, etc

# On VM
source ~/unsloth_env/bin/activate
export HF_HUB_ENABLE_HF_TRANSFER=1

# From laptop (other terminal)
tnr scp held_out_cases_v1.jsonl 0:/home/ubuntu/
tnr scp 04_eval_iter0.py        0:/home/ubuntu/


---

claude as judge
# Cd into wherever the eval results live
cd ~/projects/agentchassis/docs/.../eval/iter0_eval/

# 1. Structural (free, immediate)
python3 05_level1.py --results iter0_eval_results_v1.jsonl --output level1_metrics.json

# 2. Judge (~$1, ~5 min)
export ANTHROPIC_API_KEY="sk-ant-..._wAA"
python3 06_level2.py --results iter0_eval_results_v1.jsonl --output level2_judgments.jsonl

# 3. Report
python3 build_report.py \
--results iter0_eval_results_v1.jsonl \
--level1  level1_metrics.json \
--level2  level2_judgments.jsonl \
--output  iter0_evaluation_report.md


------
set up venv

python3 -m venv ~/.venvs/flywheel_d
source ~/.venvs/flywheel_d/bin/activate
pip install anthropic


-----

# Results

Headline: Claude 16, iter_0 4, ties 0 (out of 20). On the surface that's a 4× win rate for Claude, which sounds bad.
But the dimension scores are remarkably tight. Eyeballing the 60 numbers, both models cluster around 4-5 on every dimension. The gap on any individual dimension is mostly 1 point in either direction. iter_0 isn't getting destroyed — it's losing close votes.
Three patterns worth flagging before you run the report builder, because they'll change how you read the output:
1. Identical-score cases all went to Claude. Looking at:

Case 11: iter_0 R5/V5/I5 vs Claude R5/V5/I5 → Claude
Case 16: iter_0 R4/V4/I5 vs Claude R4/V4/I5 → Claude
Case 17: iter_0 R5/V4/I5 vs Claude R5/V4/I5 → Claude
Case 20: iter_0 R5/V4/I5 vs Claude R5/V4/I5 → Claude

Four cases where the rubric scored them identically but the judge still picked Claude — every time. That's a strong signal of self-recognition bias (the judge is Claude Opus 4.7 evaluating an output from a model trained on Claude Sonnet 4.6 — close enough family that stylistic affinity is real). If the rubric can't distinguish but the picker still picks Claude 4/4, the picker is using something the rubric doesn't measure, and that something correlates with Claude-ness.
This means the 16-4 headline overstates the gap. On numbers-the-judge-actually-articulated, it's closer to 12-4 with 4 not-meaningfully-different.
2. Two iter_0 integrity scores of 2. Cases 8 and 9 are the only ones where iter_0 got a 2 (out of 5) on integrity — meaning the judge saw something it considered genuinely problematic, probably fabrication that L1's regex didn't catch. These are the most important cases to read in the report's L3 section. If iter_0 is fabricating in ways the regex missed, that's the real concern from this eval, not the head-to-head record.
3. iter_0 won 4 cases, and three of them (4, 6, 14, 19) had iter_0 strictly beating Claude on at least two dimensions. Not flukes. Worth understanding what made those briefs work for iter_0.
