package nlp

import (
	"math"
	"strings"
)

// DecodeBIO extracts the best label for each token from logits.
// logits has shape [1, seq_len, num_labels] stored flat.
func DecodeBIO(logits []float32, seqLen, numLabels int) ([]string, []float64) {
	labels := make([]string, seqLen)
	scores := make([]float64, seqLen)

	for i := 0; i < seqLen; i++ {
		offset := i * numLabels
		bestLabel := 0
		bestScore := float64(math.Inf(-1))

		for j := 0; j < numLabels && offset+j < len(logits); j++ {
			score := float64(logits[offset+j])
			if score > bestScore {
				bestScore = score
				bestLabel = j
			}
		}

		if bestLabel < len(bioLabels) {
			labels[i] = bioLabels[bestLabel]
		} else {
			labels[i] = "O"
		}

		// Convert logit to probability via softmax normalization.
		scores[i] = SoftmaxScore(logits[offset:offset+numLabels], bestLabel)
	}

	return labels, scores
}

// SoftmaxScore computes the softmax probability for the given index.
func SoftmaxScore(logits []float32, idx int) float64 {
	if len(logits) == 0 {
		return 0
	}

	maxVal := float32(math.Inf(-1))
	for _, v := range logits {
		if v > maxVal {
			maxVal = v
		}
	}

	sumExp := float64(0)
	for _, v := range logits {
		sumExp += math.Exp(float64(v - maxVal))
	}

	return math.Exp(float64(logits[idx]-maxVal)) / sumExp
}

// MergeBIOSpans converts BIO labels + offsets into NERResult entities.
// Skips BOS (index 0) and EOS (last active index) tokens.
func MergeBIOSpans(labels []string, scores []float64, offsets []TokenOffset, text string) []NERResult {
	var results []NERResult
	runes := []rune(text)

	var current *NERResult
	var currentScoreSum float64
	var currentCount int

	// Skip BOS (index 0) and EOS (last active token).
	start := 1
	end := len(labels) - 1
	if end <= start {
		return nil
	}

	for i := start; i < end; i++ {
		label := labels[i]

		if strings.HasPrefix(label, "B-") {
			// Emit any in-progress entity.
			if current != nil {
				current.Score = currentScoreSum / float64(currentCount)
				current.Text = extractText(runes, current.Start, current.End)
				results = append(results, *current)
			}

			entityType := label[2:]
			off := safeOffset(offsets, i)
			current = &NERResult{
				Label: entityType,
				Start: off.Start,
				End:   off.End,
			}
			currentScoreSum = scores[i]
			currentCount = 1

		} else if strings.HasPrefix(label, "I-") && current != nil {
			entityType := label[2:]
			if entityType == current.Label {
				// Continue current entity.
				off := safeOffset(offsets, i)
				if off.End > current.End {
					current.End = off.End
				}
				currentScoreSum += scores[i]
				currentCount++
			} else {
				// Different type — emit current and start new.
				current.Score = currentScoreSum / float64(currentCount)
				current.Text = extractText(runes, current.Start, current.End)
				results = append(results, *current)

				off := safeOffset(offsets, i)
				current = &NERResult{
					Label: entityType,
					Start: off.Start,
					End:   off.End,
				}
				currentScoreSum = scores[i]
				currentCount = 1
			}

		} else {
			// "O" label or I- without a B- — emit any in-progress entity.
			if current != nil {
				current.Score = currentScoreSum / float64(currentCount)
				current.Text = extractText(runes, current.Start, current.End)
				results = append(results, *current)
				current = nil
			}
		}
	}

	// Emit final entity.
	if current != nil {
		current.Score = currentScoreSum / float64(currentCount)
		current.Text = extractText(runes, current.Start, current.End)
		results = append(results, *current)
	}

	return results
}

func safeOffset(offsets []TokenOffset, i int) TokenOffset {
	if i < len(offsets) {
		return offsets[i]
	}
	return TokenOffset{}
}

func extractText(runes []rune, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	if start >= end {
		return ""
	}
	return string(runes[start:end])
}
