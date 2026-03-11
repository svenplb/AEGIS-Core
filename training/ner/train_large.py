#!/usr/bin/env python
"""
NER Training: xlm-roberta-large auf MultiNERD
==============================================
Für RunPod A100 80GB – alles in einem Rutsch.

Setup auf RunPod:
1. runpod.io → Account erstellen → Credits aufladen ($15)
2. Pods → Deploy → Template: "RunPod Pytorch 2.4.0" (oder ähnlich)
3. GPU: A100 80GB PCIe oder SXM (~$1.19-1.39/h)
4. Volume: 50GB mounten auf /workspace
5. Pod starten → "Connect" → Terminal/Jupyter
6. Dieses Script hochladen und ausführen:
   pip install seqeval accelerate optimum[onnxruntime] datasets transformers
   python train_ner_runpod.py

Geschätzte Laufzeit: ~5-8h auf A100 80GB (Early Stopping stoppt wenn optimal)
Geschätzte Kosten: ~$8-10
"""

import os
import shutil
import numpy as np
from datasets import load_dataset
from transformers import (
    AutoModelForTokenClassification,
    AutoTokenizer,
    DataCollatorForTokenClassification,
    EarlyStoppingCallback,
    Trainer,
    TrainingArguments,
)
from seqeval.metrics import classification_report, f1_score

# ════════════════════════════════════════════════════════════════════
# KONFIGURATION
# ════════════════════════════════════════════════════════════════════
MODEL_NAME = "xlm-roberta-large"
OUTPUT_DIR = "/workspace/ner-model-large"
ONNX_OUTPUT_DIR = "/workspace/ner-output"
MAX_LENGTH = 512

# Daten – voller EU-Datensatz, kein Subsample nötig auf A100!
TRAIN_SAMPLES = None          # None = alles verwenden (~1.3M)
VAL_SAMPLES = None
TEST_SAMPLES = None
EU_LANGS = {"de", "en", "es", "fr", "it", "nl", "pl", "pt"}

# Training – A100 kann viel größere Batches
NUM_EPOCHS = 10               # Mehr Epochen, Early Stopping stoppt automatisch
BATCH_SIZE = 16               # Kleiner wegen MAX_LENGTH=512
GRAD_ACCUM = 2                # Effective batch: 16 × 2 = 32
LEARNING_RATE = 1e-5
EVAL_BATCH_SIZE = 32

# ════════════════════════════════════════════════════════════════════
# LABELS
# ════════════════════════════════════════════════════════════════════
LABEL_LIST = [
    "O",
    "B-PER",  "I-PER",
    "B-ORG",  "I-ORG",
    "B-LOC",  "I-LOC",
    "B-ANIM", "I-ANIM",
    "B-BIO",  "I-BIO",
    "B-CEL",  "I-CEL",
    "B-DIS",  "I-DIS",
    "B-EVE",  "I-EVE",
    "B-FOOD", "I-FOOD",
    "B-INST", "I-INST",
    "B-MEDIA","I-MEDIA",
    "B-MYTH", "I-MYTH",
    "B-PLANT","I-PLANT",
    "B-TIME", "I-TIME",
    "B-VEHI", "I-VEHI",
]
ID2LABEL = {i: l for i, l in enumerate(LABEL_LIST)}
LABEL2ID = {l: i for i, l in enumerate(LABEL_LIST)}

# ════════════════════════════════════════════════════════════════════
# MODELL & TOKENIZER
# ════════════════════════════════════════════════════════════════════
print(f"Lade {MODEL_NAME}...")
tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
model = AutoModelForTokenClassification.from_pretrained(
    MODEL_NAME,
    num_labels=len(LABEL_LIST),
    id2label=ID2LABEL,
    label2id=LABEL2ID,
)

# ════════════════════════════════════════════════════════════════════
# DATENSATZ
# ════════════════════════════════════════════════════════════════════
print("Lade MultiNERD Datensatz...")
dataset = load_dataset("Babelscape/multinerd", verification_mode="no_checks")

if "lang" in dataset["train"].column_names:
    print("Filtere auf EU-Sprachen...")
    dataset = dataset.filter(lambda ex: ex.get("lang", "en") in EU_LANGS)

# Optionales Subsampling (None = alles verwenden)
if TRAIN_SAMPLES:
    dataset["train"] = dataset["train"].shuffle(seed=42).select(
        range(min(TRAIN_SAMPLES, len(dataset["train"])))
    )
if VAL_SAMPLES:
    dataset["validation"] = dataset["validation"].shuffle(seed=42).select(
        range(min(VAL_SAMPLES, len(dataset["validation"])))
    )
if TEST_SAMPLES:
    dataset["test"] = dataset["test"].shuffle(seed=42).select(
        range(min(TEST_SAMPLES, len(dataset["test"])))
    )

print(f"  train: {len(dataset['train']):,}")
print(f"  val:   {len(dataset['validation']):,}")
print(f"  test:  {len(dataset['test']):,}")

# ════════════════════════════════════════════════════════════════════
# TOKENISIERUNG
# ════════════════════════════════════════════════════════════════════
def tokenize_and_align(examples):
    tokenized = tokenizer(
        examples["tokens"],
        is_split_into_words=True,
        truncation=True,
        max_length=MAX_LENGTH,
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

print("Tokenisiere...")
tokenized_dataset = dataset.map(
    tokenize_and_align,
    batched=True,
    remove_columns=dataset["train"].column_names,
    num_proc=4,                 # Mehr CPUs auf RunPod verfügbar
)

# ════════════════════════════════════════════════════════════════════
# METRIKEN
# ════════════════════════════════════════════════════════════════════
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

# ════════════════════════════════════════════════════════════════════
# TRAINING
# ════════════════════════════════════════════════════════════════════
training_args = TrainingArguments(
    output_dir=OUTPUT_DIR,
    num_train_epochs=NUM_EPOCHS,
    per_device_train_batch_size=BATCH_SIZE,
    per_device_eval_batch_size=EVAL_BATCH_SIZE,
    gradient_accumulation_steps=GRAD_ACCUM,
    learning_rate=LEARNING_RATE,
    lr_scheduler_type="cosine",
    weight_decay=0.01,
    warmup_ratio=0.06,
    label_smoothing_factor=0.1,
    eval_strategy="epoch",
    save_strategy="epoch",
    load_best_model_at_end=True,
    metric_for_best_model="f1",
    greater_is_better=True,
    seed=42,
    bf16=True,                          # A100 hat nativen BF16 Support → stabiler als fp16
    logging_steps=100,
    save_total_limit=2,
    report_to="none",
    dataloader_num_workers=4,
    eval_accumulation_steps=50,         # Höher für large Modell (große Logits)
)

trainer = Trainer(
    model=model,
    args=training_args,
    train_dataset=tokenized_dataset["train"],
    eval_dataset=tokenized_dataset["validation"],
    processing_class=tokenizer,
    data_collator=DataCollatorForTokenClassification(tokenizer),
    compute_metrics=compute_metrics,
    callbacks=[EarlyStoppingCallback(early_stopping_patience=3)],
)

print("=" * 60)
print(f"Training {MODEL_NAME} auf A100")
print(f"  Daten: {len(tokenized_dataset['train']):,} Trainingssamples")
print(f"  Epochen: {NUM_EPOCHS} (Early Stopping patience=3)")
print(f"  Batch: {BATCH_SIZE} × {GRAD_ACCUM} = {BATCH_SIZE * GRAD_ACCUM} effective")
print(f"  LR: {LEARNING_RATE}, Scheduler: cosine")
print(f"  Label Smoothing: 0.1")
print("=" * 60)

trainer.train()

# Speichern
print("\nSpeichere bestes Modell...")
trainer.save_model(OUTPUT_DIR)
tokenizer.save_pretrained(OUTPUT_DIR)

# Test-Evaluation
print("Evaluiere auf Testset...")
results = trainer.evaluate(tokenized_dataset["test"])
print(f"\n{'=' * 60}")
print(f"Test F1: {results['eval_f1']:.4f}")
print(f"{'=' * 60}")

# ════════════════════════════════════════════════════════════════════
# ONNX EXPORT + INT8 QUANTISIERUNG
# ════════════════════════════════════════════════════════════════════
print("\nStarte ONNX Export...")

try:
    from optimum.onnxruntime import ORTModelForTokenClassification, ORTQuantizer
    from optimum.onnxruntime.configuration import AutoQuantizationConfig

    os.makedirs(ONNX_OUTPUT_DIR, exist_ok=True)

    print(f"Exportiere nach ONNX...")
    ort_model = ORTModelForTokenClassification.from_pretrained(OUTPUT_DIR, export=True)
    ort_model.save_pretrained(ONNX_OUTPUT_DIR)

    print("Wende INT8 Dynamic Quantization an...")
    try:
        quantizer = ORTQuantizer.from_pretrained(ONNX_OUTPUT_DIR)
        qconfig = AutoQuantizationConfig.avx2(is_static=False, per_channel=False)
        quantizer.quantize(save_dir=ONNX_OUTPUT_DIR, quantization_config=qconfig)
        print("Quantisierung erfolgreich!")
    except Exception as e:
        print(f"WARNUNG: Quantisierung fehlgeschlagen ({e}), nutze unquantisiertes Modell.")

    tokenizer.save_pretrained(ONNX_OUTPUT_DIR)

    # Umbenennen zu ner.onnx
    onnx_files = [f for f in os.listdir(ONNX_OUTPUT_DIR) if f.endswith(".onnx")]
    if onnx_files:
        quantized = [f for f in onnx_files if "quantized" in f]
        src = quantized[0] if quantized else onnx_files[0]
        dst = os.path.join(ONNX_OUTPUT_DIR, "ner.onnx")
        src_path = os.path.join(ONNX_OUTPUT_DIR, src)
        if src_path != dst:
            shutil.move(src_path, dst)

    print(f"\n{'=' * 60}")
    print("ONNX Export Dateien:")
    for f in sorted(os.listdir(ONNX_OUTPUT_DIR)):
        path = os.path.join(ONNX_OUTPUT_DIR, f)
        if os.path.isfile(path):
            size_mb = os.path.getsize(path) / (1024 * 1024)
            print(f"  {f}: {size_mb:.1f} MB")
    print(f"\nDateien liegen in: {ONNX_OUTPUT_DIR}/")
    print("Herunterladen: runpodctl send oder SCP")
    print(f"{'=' * 60}")

except ImportError:
    print("WARNUNG: optimum nicht installiert.")
    print("Installiere mit: pip install optimum[onnxruntime]")

print("\nFertig! Pod kann jetzt gestoppt werden.")
print("WICHTIG: Erst Dateien herunterladen, dann Pod stoppen!")
