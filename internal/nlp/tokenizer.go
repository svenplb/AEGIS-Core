package nlp

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"unicode/utf8"
)

// Tokenizer loads a HuggingFace tokenizer.json and tokenizes text
// using the Unigram (SentencePiece) algorithm. Pure Go, no CGO.
type Tokenizer struct {
	vocab    map[string]int     // token string → token ID
	scores   map[string]float64 // token string → log-probability
	unkID    int                // ID of the <unk> token
	bosID    int                // ID of the <s> token (beginning of sequence)
	eosID    int                // ID of the </s> token (end of sequence)
	padID    int                // ID of the <pad> token
	maxLen   int                // maximum sequence length
	replacer string             // SentencePiece space marker (▁)
}

// TokenizedOutput holds the result of tokenization.
type TokenizedOutput struct {
	InputIDs      []int64
	AttentionMask []int64
	Offsets       []TokenOffset // one per token (excluding special tokens)
}

// TokenOffset maps a token back to its position in the original text.
type TokenOffset struct {
	Start int // character offset in original text
	End   int // character offset in original text (exclusive)
}

// tokenizerJSON matches the HuggingFace tokenizer.json schema.
type tokenizerJSON struct {
	Model struct {
		Type  string          `json:"type"`
		UNKId int             `json:"unk_id"`
		Vocab json.RawMessage `json:"vocab"`
	} `json:"model"`
	AddedTokens []struct {
		ID      int    `json:"id"`
		Content string `json:"content"`
		Special bool   `json:"special"`
	} `json:"added_tokens"`
	PreTokenizer *struct {
		Type        string `json:"type"`
		Replacement string `json:"replacement"`
	} `json:"pre_tokenizer"`
}

// LoadTokenizer reads a HuggingFace tokenizer.json file.
func LoadTokenizer(path string, maxLen int) (*Tokenizer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("nlp: read tokenizer %s: %w", path, err)
	}

	var tj tokenizerJSON
	if err := json.Unmarshal(data, &tj); err != nil {
		return nil, fmt.Errorf("nlp: parse tokenizer %s: %w", path, err)
	}

	// Parse vocabulary: array of [token_string, score] pairs.
	var rawVocab [][]json.RawMessage
	if err := json.Unmarshal(tj.Model.Vocab, &rawVocab); err != nil {
		return nil, fmt.Errorf("nlp: parse vocab: %w", err)
	}

	tok := &Tokenizer{
		vocab:    make(map[string]int, len(rawVocab)),
		scores:   make(map[string]float64, len(rawVocab)),
		unkID:    tj.Model.UNKId,
		maxLen:   maxLen,
		replacer: "▁", // U+2581 LOWER ONE EIGHTH BLOCK
	}

	if tj.PreTokenizer != nil && tj.PreTokenizer.Replacement != "" {
		tok.replacer = tj.PreTokenizer.Replacement
	}

	for i, entry := range rawVocab {
		if len(entry) < 2 {
			continue
		}
		var token string
		var score float64
		if err := json.Unmarshal(entry[0], &token); err != nil {
			continue
		}
		if err := json.Unmarshal(entry[1], &score); err != nil {
			continue
		}
		tok.vocab[token] = i
		tok.scores[token] = score
	}

	// Map special tokens from added_tokens.
	for _, at := range tj.AddedTokens {
		tok.vocab[at.Content] = at.ID
		switch at.Content {
		case "<s>":
			tok.bosID = at.ID
		case "</s>":
			tok.eosID = at.ID
		case "<pad>":
			tok.padID = at.ID
		case "<unk>":
			tok.unkID = at.ID
		}
	}

	return tok, nil
}

// Encode tokenizes text and returns input IDs, attention mask, and offsets.
// Adds <s> at the beginning and </s> at the end. Pads/truncates to maxLen.
func (t *Tokenizer) Encode(text string) TokenizedOutput {
	words := t.preTokenize(text)

	var tokens []int
	var offsets []TokenOffset

	for _, w := range words {
		subTokens, subOffsets := t.tokenizeWord(w.text, w.charStart)
		tokens = append(tokens, subTokens...)
		offsets = append(offsets, subOffsets...)
	}

	// Truncate to maxLen - 2 (reserve space for BOS/EOS).
	maxTokens := t.maxLen - 2
	if len(tokens) > maxTokens {
		tokens = tokens[:maxTokens]
		offsets = offsets[:maxTokens]
	}

	// Build final sequences with special tokens.
	seqLen := len(tokens) + 2
	inputIDs := make([]int64, t.maxLen)
	attentionMask := make([]int64, t.maxLen)
	finalOffsets := make([]TokenOffset, seqLen)

	// BOS token.
	inputIDs[0] = int64(t.bosID)
	attentionMask[0] = 1
	finalOffsets[0] = TokenOffset{Start: 0, End: 0}

	// Content tokens.
	for i, tid := range tokens {
		inputIDs[i+1] = int64(tid)
		attentionMask[i+1] = 1
		finalOffsets[i+1] = offsets[i]
	}

	// EOS token.
	inputIDs[seqLen-1] = int64(t.eosID)
	attentionMask[seqLen-1] = 1
	finalOffsets[seqLen-1] = TokenOffset{Start: 0, End: 0}

	// Remaining positions are padded (zeros).

	return TokenizedOutput{
		InputIDs:      inputIDs,
		AttentionMask: attentionMask,
		Offsets:       finalOffsets,
	}
}

type wordSpan struct {
	text      string // word with ▁ prefix
	charStart int    // character offset of the word in original text
	charEnd   int    // character offset (exclusive)
}

// preTokenize splits text into words with ▁ prefix (Metaspace pre-tokenizer).
// Returns word spans with their character offsets in the original text.
func (t *Tokenizer) preTokenize(text string) []wordSpan {
	if text == "" {
		return nil
	}

	var spans []wordSpan
	runes := []rune(text)
	n := len(runes)
	i := 0

	for i < n {
		// Skip leading whitespace.
		for i < n && isWhitespace(runes[i]) {
			i++
		}
		if i >= n {
			break
		}

		// Collect word characters.
		wordStart := i
		for i < n && !isWhitespace(runes[i]) {
			i++
		}

		word := string(runes[wordStart:i])
		// Add ▁ prefix (SentencePiece convention).
		spans = append(spans, wordSpan{
			text:      t.replacer + word,
			charStart: wordStart,
			charEnd:   i,
		})
	}

	return spans
}

// tokenizeWord uses the Viterbi algorithm to find the optimal
// Unigram tokenization for a ▁-prefixed word.
// Returns token IDs and character offsets in original text.
func (t *Tokenizer) tokenizeWord(word string, origCharStart int) ([]int, []TokenOffset) {
	runes := []rune(word)
	n := len(runes)
	if n == 0 {
		return nil, nil
	}

	// Length of the ▁ prefix in runes.
	prefixLen := utf8.RuneCountInString(t.replacer)

	// Viterbi: best[i] = best log-prob to tokenize runes[0:i].
	best := make([]float64, n+1)
	bestEnd := make([]int, n+1) // bestEnd[i] = start of the best token ending at i
	for i := range best {
		best[i] = math.Inf(-1)
	}
	best[0] = 0

	for end := 1; end <= n; end++ {
		for start := 0; start < end; start++ {
			if best[start] == math.Inf(-1) {
				continue
			}
			piece := string(runes[start:end])
			score, exists := t.scores[piece]
			if !exists {
				// Single character fallback.
				if end-start == 1 {
					score = -100.0 // heavy penalty for unknown single chars
				} else {
					continue
				}
			}
			candidate := best[start] + score
			if candidate > best[end] {
				best[end] = candidate
				bestEnd[end] = start
			}
		}
	}

	// Backtrace.
	if best[n] == math.Inf(-1) {
		// Entire word unknown — return <unk>.
		return []int{t.unkID}, []TokenOffset{{Start: origCharStart, End: origCharStart}}
	}

	var pieces []string
	var pieceRanges [][2]int // [start, end] in runes
	pos := n
	for pos > 0 {
		start := bestEnd[pos]
		pieces = append(pieces, string(runes[start:pos]))
		pieceRanges = append(pieceRanges, [2]int{start, pos})
		pos = start
	}

	// Reverse (backtrace gives reverse order).
	for i, j := 0, len(pieces)-1; i < j; i, j = i+1, j-1 {
		pieces[i], pieces[j] = pieces[j], pieces[i]
		pieceRanges[i], pieceRanges[j] = pieceRanges[j], pieceRanges[i]
	}

	// Convert pieces to token IDs and compute character offsets.
	ids := make([]int, len(pieces))
	offs := make([]TokenOffset, len(pieces))

	for i, piece := range pieces {
		id, ok := t.vocab[piece]
		if !ok {
			id = t.unkID
		}
		ids[i] = id

		// Map rune positions back to original text characters.
		// The ▁ prefix maps to no original character.
		runeStart := pieceRanges[i][0]
		runeEnd := pieceRanges[i][1]

		// Adjust for ▁ prefix: characters before prefixLen are the space marker.
		charStart := origCharStart
		charEnd := origCharStart

		if runeEnd > prefixLen {
			// This piece covers some actual characters.
			actualStart := runeStart
			if actualStart < prefixLen {
				actualStart = prefixLen
			}
			charStart = origCharStart + (actualStart - prefixLen)
			charEnd = origCharStart + (runeEnd - prefixLen)
		}

		offs[i] = TokenOffset{Start: charStart, End: charEnd}
	}

	return ids, offs
}

func isWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

// buildCharToByteMap returns a mapping from character (rune) index to byte offset.
func buildCharToByteMap(text string) []int {
	charToByte := make([]int, 0, len(text))
	byteIdx := 0
	for _, r := range text {
		charToByte = append(charToByte, byteIdx)
		byteIdx += utf8.RuneLen(r)
	}
	// Append the final byte offset (length of text) for exclusive end offsets.
	charToByte = append(charToByte, byteIdx)
	return charToByte
}

// CharToByteOffset converts a character offset to a byte offset.
// Returns the byte offset, or len(text) if charIdx is out of range.
func CharToByteOffset(text string, charIdx int) int {
	byteIdx := 0
	i := 0
	for _, r := range text {
		if i == charIdx {
			return byteIdx
		}
		byteIdx += utf8.RuneLen(r)
		i++
	}
	return byteIdx
}

// DecodeTokenOffsets converts character-based offsets to byte-based offsets.
func DecodeTokenOffsets(text string, results []NERResult) []NERResult {
	if len(results) == 0 {
		return results
	}

	charToByte := buildCharToByteMap(text)
	textLen := len(text)
	decoded := make([]NERResult, len(results))
	copy(decoded, results)

	for i := range decoded {
		start := decoded[i].Start
		end := decoded[i].End

		if start < 0 {
			decoded[i].Start = 0
		} else if start >= len(charToByte) {
			decoded[i].Start = textLen
		} else {
			decoded[i].Start = charToByte[start]
		}

		if end < 0 {
			decoded[i].End = 0
		} else if end >= len(charToByte) {
			decoded[i].End = textLen
		} else {
			decoded[i].End = charToByte[end]
		}
	}

	return decoded
}
