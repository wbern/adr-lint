package syntheticdiff

import (
	"fmt"
	"strings"
)

func GenerateSyntheticDiff(filePath, content string) string {
	if content == "" {
		return fmt.Sprintf("diff --git a/%s b/%s\nnew file mode 100644\n--- /dev/null\n+++ b/%s",
			filePath, filePath, filePath)
	}
	lines := strings.Split(content, "\n")
	added := make([]string, len(lines))
	for i, l := range lines {
		added[i] = "+" + l
	}
	return fmt.Sprintf("diff --git a/%s b/%s\nnew file mode 100644\n--- /dev/null\n+++ b/%s\n@@ -0,0 +1,%d @@\n%s",
		filePath, filePath, filePath, len(lines), strings.Join(added, "\n"))
}
