from transformers import AutoTokenizer, AutoModelForCausalLM
from peft import PeftModel

BASE_MODEL = "microsoft/Phi-3-mini-4k-instruct"
ADAPTER_PATH = "./copywriter-phi3-lora"

# Load base model and tokenizer
tokenizer = AutoTokenizer.from_pretrained(BASE_MODEL)
base_model = AutoModelForCausalLM.from_pretrained(
    BASE_MODEL,
    device_map="auto",
    load_in_4bit=True
)

# Load LoRA adapter
model = PeftModel.from_pretrained(base_model, ADAPTER_PATH)
model.eval()

# Example prompt
instruction = "Write a fun ad for the ACEBOTT 4 DOF Robot Hand Kit aimed at kids who love building robots."
input_text = "Key features: 4 DOF arm, 260° clamp, 180° rotations, servo claw, ESP32 programming."

prompt = f"Instruction: {instruction}\nInput: {input_text}\n\nResponse:"

inputs = tokenizer(prompt, return_tensors="pt").to("cuda")
outputs = model.generate(
    **inputs,
    max_new_tokens=200,
    temperature=0.9,
    top_p=0.95,
    do_sample=True
)

print("\n🧠 Model output:\n")
print(tokenizer.decode(outputs[0], skip_special_tokens=True))
