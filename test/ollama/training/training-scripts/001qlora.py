from transformers import AutoModelForCausalLM, AutoTokenizer, TrainingArguments, Trainer
from peft import LoraConfig, get_peft_model, prepare_model_for_kbit_training
from datasets import load_dataset

# ---- Config ----
BASE_MODEL = "microsoft/Phi-3-mini-4k-instruct"
DATA_FILE = "training.json"
OUTPUT_DIR = "./copywriter-phi3-lora"

# ---- Load base model ----
model = AutoModelForCausalLM.from_pretrained(
    BASE_MODEL,
    load_in_4bit=True,
    device_map="auto"
)
tokenizer = AutoTokenizer.from_pretrained(BASE_MODEL)
model = prepare_model_for_kbit_training(model)

# ---- Apply LoRA adapter ----
lora_config = LoraConfig(
    r=8,
    lora_alpha=32,
    target_modules=["qkv_proj", "o_proj"],
    lora_dropout=0.05,
    bias="none",
    task_type="CAUSAL_LM"
)
model = get_peft_model(model, lora_config)

# ---- Load and tokenize dataset ----
dataset = load_dataset("json", data_files=DATA_FILE)

# Format each example into a single instruction prompt
def format_example(example):
    # Combine the fields into a single text prompt (the model learns from this)
    text = (
        f"Instruction: {example['instruction']}\n"
        f"Input: {example['input']}\n\n"
        f"Response: {example['output']}"
    )
    tokenized = tokenizer(
        text,
        truncation=True,
        padding="max_length",
        max_length=512,
    )
    # Labels are required for computing loss
    tokenized["labels"] = tokenized["input_ids"].copy()
    return tokenized

tokenized = dataset.map(format_example)

# ---- Training setup ----
training_args = TrainingArguments(
    output_dir=OUTPUT_DIR,
    per_device_train_batch_size=1,
    gradient_accumulation_steps=4,
    num_train_epochs=1,  # just a quick test
    learning_rate=2e-4,
    fp16=True,
    logging_steps=5,
    save_strategy="no",
    report_to="none"
)

trainer = Trainer(
    model=model,
    args=training_args,
    train_dataset=tokenized["train"]
)

trainer.train()

model.save_pretrained(OUTPUT_DIR)
tokenizer.save_pretrained(OUTPUT_DIR)
print(f"✅ Training complete! LoRA adapter saved to {OUTPUT_DIR}")
