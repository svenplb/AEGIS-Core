import numpy as np
from datasets import load_dataset
from transformers import (
    AutoModelForTokenClassification,
    AutoTokenizer,
    DataCollatorForTokenClassification,
    Trainer,
    TrainingArguments,
)
from seqeval.metrics import classification_report, f1_score

LABEL_LIST = [
    "O",
    "B-PER", "I-PER", "B-ORG", "I-ORG", "B-LOC", "I-LOC",
    "B-ANIM", "I-ANIM", "B-BIO", "I-BIO", "B-CEL", "I-CEL",
    "B-DIS", "I-DIS", "B-EVE", "I-EVE", "B-FOOD", "I-FOOD",
    "B-INST", "I-INST", "B-MEDIA", "I-MEDIA", "B-MYTH", "I-MYTH",
    "B-PLANT", "I-PLANT", "B-TIME", "I-TIME", "B-VEHI", "I-VEHI",
]
ID2LABEL = {i: l for i, l in enumerate(LABEL_LIST)}
LABEL2ID = {l: i for i, l in enumerate(LABEL_LIST)}

MODEL_NAME = "xlm-roberta-base"

print("Loading model...")
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
model = AutoModelForTokenClassification.from_pretrained(
    MODEL_NAME, num_labels=len(LABEL_LIST), id2label=ID2LABEL, label2id=LABEL2ID,
)

print("Loading dataset...")
dataset = load_dataset("Babelscape/multinerd", verification_mode="no_checks")

eu_langs = {"de", "en", "es", "fr", "it", "nl", "pl", "pt"}
if "lang" in dataset["train"].column_names:
    dataset = dataset.filter(lambda ex: ex.get("lang", "en") in eu_langs)


def tokenize_and_align(examples):
    tokenized = tokenizer(
        examples["tokens"],
        is_split_into_words=True,
        truncation=True,
        max_length=512,
        padding=False,
    )
    all_labels = []
    for i, labels in enumerate(examples["ner_tags"]):
        word_ids = tokenized.word_ids(batch_index=i)
        label_ids = []
        prev = None
        for wid in word_ids:
            if wid is None:
                label_ids.append(-100)
            elif wid != prev:
                label_ids.append(labels[wid])
            else:
                lab = labels[wid]
                name = LABEL_LIST[lab] if lab < len(LABEL_LIST) else "O"
                if name.startswith("B-"):
                    label_ids.append(LABEL2ID.get(f"I-{name[2:]}", lab))
                else:
                    label_ids.append(lab)
            prev = wid
        all_labels.append(label_ids)
    tokenized["labels"] = all_labels
    return tokenized


print("Tokenizing...")
tokenized_dataset = dataset.map(
    tokenize_and_align,
    batched=True,
    remove_columns=dataset["train"].column_names,
)


def compute_metrics(eval_preds):
    logits, labels = eval_preds
    preds = np.argmax(logits, axis=-1)
    true_l, true_p = [], []
    for ps, ls in zip(preds, labels):
        tl, tp = [], []
        for p, l in zip(ps, ls):
            if l == -100:
                continue
            tl.append(ID2LABEL.get(l, "O"))
            tp.append(ID2LABEL.get(p, "O"))
        true_l.append(tl)
        true_p.append(tp)
    print(classification_report(true_l, true_p))
    return {"f1": f1_score(true_l, true_p, average="micro")}


training_args = TrainingArguments(
    output_dir="/kaggle/working/ner-model",
    num_train_epochs=5,
    per_device_train_batch_size=16,
    per_device_eval_batch_size=32,
    gradient_accumulation_steps=2,
    learning_rate=2e-5,
    weight_decay=0.01,
    warmup_ratio=0.1,
    eval_strategy="epoch",
    save_strategy="epoch",
    load_best_model_at_end=True,
    metric_for_best_model="f1",
    greater_is_better=True,
    seed=42,
    fp16=True,
    logging_steps=100,
    save_total_limit=2,
    report_to="none",
)

trainer = Trainer(
    model=model,
    args=training_args,
    train_dataset=tokenized_dataset["train"],
    eval_dataset=tokenized_dataset["validation"],
    processing_class=tokenizer,
    data_collator=DataCollatorForTokenClassification(tokenizer),
    compute_metrics=compute_metrics,
)

print("Starting training...")
trainer.train()
trainer.save_model("/kaggle/working/ner-model")
tokenizer.save_pretrained("/kaggle/working/ner-model")
results = trainer.evaluate(tokenized_dataset["test"])
print(f"Test F1: {results['eval_f1']:.4f}")
print("Done!")
