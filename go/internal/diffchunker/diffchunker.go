package diffchunker

import "strings"

const criticalReminder = "\n⚠️ REMINDER: Only check ADDED lines for violations ⚠️\n" +
	"- Lines prefixed with `+` (e.g., `+ const x = 1`) → ADDED code — check these\n" +
	"- Lines prefixed with `+++` or `---` → file headers — IGNORE\n" +
	"- Lines with no prefix → existing context — IGNORE\n" +
	"- Each file must independently comply. If ANY single file violates the requirement, the overall result is FAIL.\n"

func ChunkDiffByFile(diff string, maxTokensPerChunk int, insertReminders bool) []string {
	fileDiffs := splitByFile(diff)
	if insertReminders {
		for i, fd := range fileDiffs {
			fileDiffs[i] = criticalReminder + fd
		}
	}
	if maxTokensPerChunk <= 0 {
		return fileDiffs
	}
	return chunkByTokenLimit(fileDiffs, maxTokensPerChunk)
}

func estimateTokenCount(text string) int {
	n := len(text) / 4
	if len(text)%4 != 0 {
		n++
	}
	return n
}

func chunkByTokenLimit(fileDiffs []string, maxTokensPerChunk int) []string {
	var chunks []string
	var current []string
	currentTokens := 0
	for _, fd := range fileDiffs {
		ft := estimateTokenCount(fd)
		if currentTokens+ft > maxTokensPerChunk && len(current) > 0 {
			chunks = append(chunks, strings.Join(current, "\n"))
			current = nil
			currentTokens = 0
		}
		current = append(current, fd)
		currentTokens += ft
	}
	if len(current) > 0 {
		chunks = append(chunks, strings.Join(current, "\n"))
	}
	return chunks
}

func splitByFile(diff string) []string {
	if strings.TrimSpace(diff) == "" {
		return nil
	}
	var chunks []string
	var current []string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "diff --git") && len(current) > 0 {
			chunks = append(chunks, strings.Join(current, "\n"))
			current = nil
		}
		current = append(current, line)
	}
	if len(current) > 0 {
		chunks = append(chunks, strings.Join(current, "\n"))
	}
	return chunks
}
