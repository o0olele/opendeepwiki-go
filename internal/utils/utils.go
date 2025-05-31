package utils

import (
	"fmt"
	"regexp"
	"strings"
)

// ExtractRepoName Get the repository name from a Git URL.
func ExtractRepoName(gitURL string) (string, error) {
	// remove .git suffix
	name := strings.TrimSuffix(gitURL, ".git")

	// handle github.com/xxx/xxx format
	if strings.Contains(name, "github.com") {
		parts := strings.Split(name, "/")
		if len(parts) > 0 {
			return parts[len(parts)-1], nil
		}
	}

	// handle other formats, such as gitlab.com/xxx/xxx or bitbucket.org/xxx/xxx, etc.
	parts := strings.Split(name, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("failed to extract repository name from %s", gitURL)
}

// ExtractTagContent 提取标签内容
func ExtractTagContent(text, tag string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?s)<\s*%s\s*>(.*?)<\s*/%s\s*>`, tag, tag))
	matches := re.FindStringSubmatch(text)
	if len(matches) >= 2 {
		return strings.TrimSpace(matches[1])
	}
	return text
}

func ReplaceTagContent(text, tag, newContent string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?s)<\s*%s\s*>(.*?)<\s*/%s\s*>`, tag, tag))
	return re.ReplaceAllString(text, fmt.Sprintf("<%s>%s</%s>", tag, newContent, tag))
}

// ExtractJSON 从文本中提取 JSON
func ExtractJSON(text string) string {
	// 尝试找到 JSON 对象或数组
	reObject := regexp.MustCompile(`(?s)\{.*\}`)
	reArray := regexp.MustCompile(`(?s)\[.*\]`)

	// 首先尝试提取对象
	matches := reObject.FindString(text)
	if matches != "" {
		return matches
	}

	// 然后尝试提取数组
	matches = reArray.FindString(text)
	if matches != "" {
		return matches
	}

	// 尝试提取代码块中的 JSON
	reCodeBlock := regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")
	codeMatches := reCodeBlock.FindStringSubmatch(text)
	if len(codeMatches) >= 2 {
		return strings.TrimSpace(codeMatches[1])
	}

	return ""
}
