package services

import (
	"sort"
	"strings"
	"sync"
)

// TrieNode represents a node in the trie
type TrieNode struct {
	children     map[rune]*TrieNode
	isEnd        bool
	canonicalTerm string // The normalized term to use
	frequency    int     // How many times this term has been seen
}

// TermNormalizer handles intelligent term normalization and canonicalization
type TermNormalizer struct {
	root              *TrieNode
	canonicalTerms    map[string]string   // normalized -> canonical
	termVariations    map[string][]string // canonical -> variations
	fuzzyMatchCache   map[string]string   // cache for fuzzy matches
	mu                sync.RWMutex        // mutex for thread-safe operations
}

func NewTermNormalizer() *TermNormalizer {
	return &TermNormalizer{
		root:            &TrieNode{children: make(map[rune]*TrieNode)},
		canonicalTerms:  make(map[string]string),
		termVariations:  make(map[string][]string),
		fuzzyMatchCache: make(map[string]string),
	}
}

// NormalizeTerm finds the canonical form of a term using intelligent matching
func (tn *TermNormalizer) NormalizeTerm(term string) string {
	// Basic cleaning
	cleaned := strings.ToLower(strings.TrimSpace(term))
	if cleaned == "" {
		return term
	}

	// Check cache first (read lock)
	tn.mu.RLock()
	if canonical, exists := tn.fuzzyMatchCache[cleaned]; exists {
		tn.mu.RUnlock()
		// Update frequency with write lock
		tn.mu.Lock()
		tn.insertIntoTrie(cleaned, canonical)
		tn.mu.Unlock()
		return canonical
	}

	// Try exact match
	if canonical, exists := tn.canonicalTerms[cleaned]; exists {
		tn.mu.RUnlock()
		// Update frequency with write lock
		tn.mu.Lock()
		tn.insertIntoTrie(cleaned, canonical)
		tn.mu.Unlock()
		return canonical
	}
	tn.mu.RUnlock()

	// Acquire write lock for modifications
	tn.mu.Lock()
	defer tn.mu.Unlock()

	// Double-check after acquiring write lock
	if canonical, exists := tn.fuzzyMatchCache[cleaned]; exists {
		tn.insertIntoTrie(cleaned, canonical)
		return canonical
	}

	// Try prefix-based matching (for incomplete words)
	prefixMatch := tn.findLongestPrefixMatch(cleaned)
	if prefixMatch != "" {
		// Check if this is a reasonable prefix match
		if tn.isSimilarByPrefix(cleaned, prefixMatch) {
			tn.fuzzyMatchCache[cleaned] = prefixMatch
			tn.canonicalTerms[cleaned] = prefixMatch
			tn.insertIntoTrie(cleaned, prefixMatch)
			return prefixMatch
		}
	}

	// Try fuzzy matching with existing canonical terms BEFORE registering new term
	bestMatch := tn.findBestFuzzyMatch(cleaned)
	if bestMatch != "" {
		tn.fuzzyMatchCache[cleaned] = bestMatch
		tn.canonicalTerms[cleaned] = bestMatch
		tn.insertIntoTrie(cleaned, bestMatch)
		// Add to variations
		tn.termVariations[bestMatch] = append(tn.termVariations[bestMatch], cleaned)
		return bestMatch
	}

	// This is a new term, make it canonical
	tn.registerCanonicalTerm(cleaned)
	return cleaned
}

// registerCanonicalTerm adds a new canonical term
func (tn *TermNormalizer) registerCanonicalTerm(term string) {
	tn.canonicalTerms[term] = term
	tn.termVariations[term] = []string{term}
	tn.insertIntoTrie(term, term)
}

// insertIntoTrie adds a term to the trie
func (tn *TermNormalizer) insertIntoTrie(term string, canonical string) {
	node := tn.root
	for _, ch := range term {
		if node.children[ch] == nil {
			node.children[ch] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.canonicalTerm = canonical
	node.frequency++
}

// findLongestPrefixMatch searches the trie for the longest matching prefix
func (tn *TermNormalizer) findLongestPrefixMatch(term string) string {
	node := tn.root
	lastValidMatch := ""
	
	for _, ch := range term {
		if node.children[ch] == nil {
			break
		}
		node = node.children[ch]
		if node.isEnd {
			lastValidMatch = node.canonicalTerm
		}
	}
	
	return lastValidMatch
}

// isSimilarByPrefix checks if a term is a reasonable prefix match
func (tn *TermNormalizer) isSimilarByPrefix(term string, canonical string) bool {
	// If term is very short, require exact match
	if len(term) < 5 {
		return term == canonical
	}
	
	// Check if term is a prefix or near-prefix of canonical
	commonPrefix := longestCommonPrefix(term, canonical)
	
	// Allow if at least 80% of the shorter term matches
	minLen := len(term)
	if len(canonical) < minLen {
		minLen = len(canonical)
	}
	
	threshold := float64(minLen) * 0.8
	return float64(len(commonPrefix)) >= threshold
}

// findBestFuzzyMatch finds the most similar canonical term
func (tn *TermNormalizer) findBestFuzzyMatch(term string) string {
	bestMatch := ""
	bestScore := 0.0
	minSimilarity := 0.80 // 80% similarity threshold (increased for better accuracy)
	
	// Only compare with canonical forms (keys in canonicalTerms where key == value)
	for key, canonical := range tn.canonicalTerms {
		// Only consider actual canonical terms (not aliases)
		if key != canonical {
			continue
		}
		
		score := calculateSimilarity(term, canonical)
		if score > bestScore && score >= minSimilarity {
			bestScore = score
			bestMatch = canonical
		}
	}
	
	return bestMatch
}

// calculateSimilarity computes similarity score between two terms
func calculateSimilarity(s1, s2 string) float64 {
	// Quick length check - if one term is much shorter, likely not a match
	lenDiff := absTwo(len(s1) - len(s2))
	maxLen := maxTwo(len(s1), len(s2))
	
	if maxLen == 0 {
		return 0.0
	}
	
	// If length difference is more than 20% of max length, reduce base similarity
	lengthPenalty := 1.0
	if float64(lenDiff)/float64(maxLen) > 0.2 {
		lengthPenalty = 0.8
	}
	
	// 1. Longest common subsequence ratio
	lcs := longestCommonSubsequence(s1, s2)
	lcsRatio := float64(lcs) / float64(maxLen)
	
	// 2. Levenshtein distance ratio (higher weight for edit distance)
	distance := levenshteinDistance(s1, s2)
	levRatio := 1.0 - (float64(distance) / float64(maxLen))
	
	// 3. Token overlap (for multi-word terms)
	tokens1 := strings.Fields(s1)
	tokens2 := strings.Fields(s2)
	tokenOverlap := calculateTokenOverlap(tokens1, tokens2)
	
	// 4. Prefix matching bonus (if terms share a long prefix)
	prefix := longestCommonPrefix(s1, s2)
	prefixBonus := 0.0
	if len(prefix) >= 5 && float64(len(prefix))/float64(maxLen) > 0.6 {
		prefixBonus = 0.1
	}
	
	// Weighted combination - give more weight to Levenshtein
	similarity := (lcsRatio * 0.25) + (levRatio * 0.5) + (tokenOverlap * 0.25) + prefixBonus
	
	return similarity * lengthPenalty
}

// longestCommonPrefix finds the longest common prefix
func longestCommonPrefix(s1, s2 string) string {
	minLen := len(s1)
	if len(s2) < minLen {
		minLen = len(s2)
	}
	
	for i := 0; i < minLen; i++ {
		if s1[i] != s2[i] {
			return s1[:i]
		}
	}
	
	return s1[:minLen]
}

// longestCommonSubsequence calculates LCS length
func longestCommonSubsequence(s1, s2 string) int {
	m, n := len(s1), len(s2)
	if m == 0 || n == 0 {
		return 0
	}
	
	// Create DP table
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	
	// Fill DP table
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if s1[i-1] == s2[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else {
				dp[i][j] = maxTwo(dp[i-1][j], dp[i][j-1])
			}
		}
	}
	
	return dp[m][n]
}

// calculateTokenOverlap calculates overlap between token sets
func calculateTokenOverlap(tokens1, tokens2 []string) float64 {
	if len(tokens1) == 0 && len(tokens2) == 0 {
		return 1.0
	}
	if len(tokens1) == 0 || len(tokens2) == 0 {
		return 0.0
	}
	
	set1 := make(map[string]bool)
	for _, t := range tokens1 {
		set1[t] = true
	}
	
	overlap := 0
	for _, t := range tokens2 {
		if set1[t] {
			overlap++
		}
	}
	
	maxSize := maxTwo(len(tokens1), len(tokens2))
	return float64(overlap) / float64(maxSize)
}

// GetCanonicalTerms returns all canonical terms sorted by frequency
func (tn *TermNormalizer) GetCanonicalTerms() []string {
	type termFreq struct {
		term string
		freq int
	}
	
	terms := make([]termFreq, 0, len(tn.canonicalTerms))
	for term := range tn.canonicalTerms {
		freq := tn.getFrequency(term)
		terms = append(terms, termFreq{term, freq})
	}
	
	// Sort by frequency (descending)
	sort.Slice(terms, func(i, j int) bool {
		return terms[i].freq > terms[j].freq
	})
	
	result := make([]string, len(terms))
	for i, tf := range terms {
		result[i] = tf.term
	}
	
	return result
}

// getFrequency returns how many times a term has been seen
func (tn *TermNormalizer) getFrequency(term string) int {
	node := tn.root
	for _, ch := range term {
		if node.children[ch] == nil {
			return 0
		}
		node = node.children[ch]
	}
	if node.isEnd {
		return node.frequency
	}
	return 0
}

// MergeSimilarTerms analyzes all canonical terms and merges very similar ones
func (tn *TermNormalizer) MergeSimilarTerms() {
	canonicals := tn.GetCanonicalTerms()
	
	merged := make(map[string]string) // old -> new canonical
	
	for i := 0; i < len(canonicals); i++ {
		for j := i + 1; j < len(canonicals); j++ {
			term1 := canonicals[i]
			term2 := canonicals[j]
			
			// Skip if already merged
			if _, ok := merged[term1]; ok {
				continue
			}
			if _, ok := merged[term2]; ok {
				continue
			}
			
			similarity := calculateSimilarity(term1, term2)
			if similarity >= 0.85 { // High similarity threshold for merging
				// Keep the more frequent or longer term
				freq1 := tn.getFrequency(term1)
				freq2 := tn.getFrequency(term2)
				
				var keep, merge string
				if freq1 > freq2 || (freq1 == freq2 && len(term1) >= len(term2)) {
					keep = term1
					merge = term2
				} else {
					keep = term2
					merge = term1
				}
				
				merged[merge] = keep
				tn.canonicalTerms[merge] = keep
			}
		}
	}
}

// Helper functions
func absTwo(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func maxTwo(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minTwo(a, b int) int {
	if a < b {
		return a
	}
	return b
}
