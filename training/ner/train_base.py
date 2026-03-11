"""
Fine-tune XLM-RoBERTa on MultiNERD for multilingual Named Entity Recognition.

Usage:
    pip install -r ../requirements.txt
    python train_base.py --output_dir ./ner-model --epochs 5

Base model: xlm-roberta-base (100+ languages, MIT license)
Dataset: MultiNERD (Tedeschi & Navigli, 2022) — 164K sentences, 10 languages
"""

import argparse
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


# MultiNERD BIO label scheme.
LABEL_LIST = [
    "O",
    "B-PER", "I-PER",
    "B-ORG", "I-ORG",
    "B-LOC", "I-LOC",
    "B-ANIM", "I-ANIM",
    "B-BIO", "I-BIO",
    "B-CEL", "I-CEL",
    "B-DIS", "I-DIS",
    "B-EVE", "I-EVE",
    "B-FOOD", "I-FOOD",
    "B-INST", "I-INST",
    "B-MEDIA", "I-MEDIA",
    "B-MYTH", "I-MYTH",
    "B-PLANT", "I-PLANT",
    "B-TIME", "I-TIME",
    "B-VEHI", "I-VEHI",
]
ID2LABEL = {i: l for i, l in enumerate(LABEL_LIST)}
LABEL2ID = {l: i for i, l in enumerate(LABEL_LIST)}

MODEL_NAME = "xlm-roberta-base"


def parse_args():
    parser = argparse.ArgumentParser(description="Fine-tune XLM-RoBERTa on MultiNERD")
    parser.add_argument("--output_dir", type=str, default="./ner-model")
    parser.add_argument("--epochs", type=int, default=5)
    parser.add_argument("--batch_size", type=int, default=16)
    parser.add_argument("--learning_rate", type=float, default=2e-5)
    parser.add_argument("--max_length", type=int, default=512)
    parser.add_argument("--seed", type=int, default=42)
    return parser.parse_args()


def tokenize_and_align(examples, tokenizer, max_length):
    """Tokenize text and align NER labels to subword tokens."""
    tokenized = tokenizer(
        examples["tokens"],
        is_split_into_words=True,
        truncation=True,
        max_length=max_length,
        padding=False,
    )

    all_labels = []
    for i, labels in enumerate(examples["ner_tags"]):
        word_ids = tokenized.word_ids(batch_index=i)
        label_ids = []
        previous_word_id = None

        for word_id in word_ids:
            if word_id is None:
                # Special tokens get -100 (ignored in loss).
                label_ids.append(-100)
            elif word_id != previous_word_id:
                # First subword token of a word gets the actual label.
                label_ids.append(labels[word_id])
            else:
                # Subsequent subword tokens: use I- variant if B-.
                label = labels[word_id]
                label_name = LABEL_LIST[label] if label < len(LABEL_LIST) else "O"
                if label_name.startswith("B-"):
                    entity_type = label_name[2:]
                    i_label = LABEL2ID.get(f"I-{entity_type}", label)
                    label_ids.append(i_label)
                else:
                    label_ids.append(label)
            previous_word_id = word_id

        all_labels.append(label_ids)

    tokenized["labels"] = all_labels
    return tokenized


def compute_metrics(eval_preds):
    """Compute seqeval F1 score."""
    logits, labels = eval_preds
    predictions = np.argmax(logits, axis=-1)

    true_labels = []
    true_predictions = []

    for pred_seq, label_seq in zip(predictions, labels):
        true_label_seq = []
        true_pred_seq = []

        for p, l in zip(pred_seq, label_seq):
            if l == -100:
                continue
            true_label_seq.append(ID2LABEL.get(l, "O"))
            true_pred_seq.append(ID2LABEL.get(p, "O"))

        true_labels.append(true_label_seq)
        true_predictions.append(true_pred_seq)

    f1 = f1_score(true_labels, true_predictions, average="micro")
    report = classification_report(true_labels, true_predictions)
    print(report)

    return {"f1": f1}


def main():
    args = parse_args()

    print(f"Loading model: {MODEL_NAME}")
    tokenizer = AutoTokenizer.from_pretrained(MODEL_NAME)
    model = AutoModelForTokenClassification.from_pretrained(
        MODEL_NAME,
        num_labels=len(LABEL_LIST),
        id2label=ID2LABEL,
        label2id=LABEL2ID,
    )

    print("Loading MultiNERD dataset...")
    dataset = load_dataset("Babelscape/multinerd", verification_mode="no_checks")

    # Filter to EU languages.
    eu_langs = {"de", "en", "es", "fr", "it", "nl", "pl", "pt"}

    def filter_eu(example):
        return example.get("lang", "en") in eu_langs

    if "lang" in dataset["train"].column_names:
        dataset = dataset.filter(filter_eu)

    print("Tokenizing dataset...")
    tokenized_dataset = dataset.map(
        lambda ex: tokenize_and_align(ex, tokenizer, args.max_length),
        batched=True,
        remove_columns=dataset["train"].column_names,
    )

    data_collator = DataCollatorForTokenClassification(tokenizer)

    training_args = TrainingArguments(
        output_dir=args.output_dir,
        num_train_epochs=args.epochs,
        per_device_train_batch_size=args.batch_size,
        per_device_eval_batch_size=args.batch_size * 2,
        gradient_accumulation_steps=2,
        learning_rate=args.learning_rate,
        weight_decay=0.01,
        warmup_ratio=0.1,
        eval_strategy="epoch",
        save_strategy="epoch",
        load_best_model_at_end=True,
        metric_for_best_model="f1",
        greater_is_better=True,
        seed=args.seed,
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
        data_collator=data_collator,
        compute_metrics=compute_metrics,
    )

    print("Starting training...")
    trainer.train()

    print(f"Saving model to {args.output_dir}")
    trainer.save_model(args.output_dir)
    tokenizer.save_pretrained(args.output_dir)

    print("Evaluating on test set...")
    results = trainer.evaluate(tokenized_dataset["test"])
    print(f"Test F1: {results['eval_f1']:.4f}")

    print("Done!")


if __name__ == "__main__":
    main()
