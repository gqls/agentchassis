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

[
ant@ant-XPS-15-9500:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning$ cd working/phase2/lora_iter0_full/
ant@ant-XPS-15-9500:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase2/lora_iter0_full$ gunzip adapter_model.safetensors.gz
ant@ant-XPS-15-9500:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase2/lora_iter0_full$ cd ..
ant@ant-XPS-15-9500:~/projects/agentchassis/docs/agent_docs/docs024_key_docs_latest/finetuning/working/phase2$ tnr scp  lora_iter0_full 0:/home/ubuntu/lora_iter0_ful
]

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
tnr scp -r lora_iter0_full       0:/home/ubuntu/lora_iter0_full

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