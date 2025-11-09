from transformers import AutoModelForCausalLM

model_id = "microsoft/Phi-3-mini-4k-instruct"
model = AutoModelForCausalLM.from_pretrained(
    model_id,
    device_map="auto",
)

for name, module in model.named_modules():
    if any(x in name.lower() for x in ["attn", "proj", "q", "v", "gate"]):
        print(name)
