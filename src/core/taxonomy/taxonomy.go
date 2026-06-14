package taxonomy

import "strings"

var canonicalCategories = []string{
	"artificial-intelligence",
	"developer-tools",
	"infrastructure",
	"kubernetes",
	"security",
	"database",
	"networking",
	"web-development",
	"data-engineering",
	"observability",
	"automation",
	"bots",
	"productivity",
	"media",
	"gaming",
	"blockchain",
	"iot",
	"privacy",
	"communications",
	"systems",
	"documentation",
	"ui-components",
	"science",
	"uncategorized",
}

var directAliases = map[string]string{
	"":                        "uncategorized",
	"ai":                      "artificial-intelligence",
	"ai-agent":                "artificial-intelligence",
	"ai-agents":               "artificial-intelligence",
	"ai-automation":           "artificial-intelligence",
	"ai-coding-agents":        "artificial-intelligence",
	"ai-dev-tools":            "artificial-intelligence",
	"ai-engine":               "artificial-intelligence",
	"ai-gateway":              "artificial-intelligence",
	"ai-memory-server":        "artificial-intelligence",
	"ai-model":                "artificial-intelligence",
	"ai-platform":             "artificial-intelligence",
	"ai-proxy":                "artificial-intelligence",
	"ai-research":             "artificial-intelligence",
	"ai-routing":              "artificial-intelligence",
	"ai-runtime":              "artificial-intelligence",
	"ai-sandbox":              "artificial-intelligence",
	"ai-security":             "artificial-intelligence",
	"ai-simulation":           "artificial-intelligence",
	"ai-tool":                 "artificial-intelligence",
	"ai-tooling":              "artificial-intelligence",
	"ai-toolkit":              "artificial-intelligence",
	"ai-tools":                "artificial-intelligence",
	"ai-workflow":             "artificial-intelligence",
	"ai-workflows":            "artificial-intelligence",
	"artificial-intelligence": "artificial-intelligence",
	"computer-vision":         "artificial-intelligence",
	"deep-learning":           "artificial-intelligence",
	"llm-security":            "artificial-intelligence",
	"machine-learning":        "artificial-intelligence",
	"nlp-framework":           "artificial-intelligence",

	"accessibility-tool":     "developer-tools",
	"build-tools":            "developer-tools",
	"code-quality":           "developer-tools",
	"code-search":            "developer-tools",
	"coding-tool":            "developer-tools",
	"date-time-utils":        "developer-tools",
	"dev-tool":               "developer-tools",
	"dev-tools":              "developer-tools",
	"developer-tools":        "developer-tools",
	"development-tool":       "developer-tools",
	"driver-build-tools":     "developer-tools",
	"github-tools":           "developer-tools",
	"go":                     "developer-tools",
	"go-framework":           "developer-tools",
	"installer-sdk":          "developer-tools",
	"issue-tracking":         "developer-tools",
	"osx-development":        "developer-tools",
	"parser-generator":       "developer-tools",
	"perl-cpan":              "developer-tools",
	"programming-language":   "developer-tools",
	"programming-languages":  "developer-tools",
	"rust":                   "developer-tools",
	"testing-framework":      "developer-tools",
	"testing-tool":           "developer-tools",
	"text-processing":        "developer-tools",
	"typescript":             "developer-tools",
	"workflow-orchestration": "developer-tools",

	"ci-cd":                     "infrastructure",
	"cloud-cost-management":     "infrastructure",
	"cloud-infrastructure":      "infrastructure",
	"cloud-query-language":      "infrastructure",
	"cloud-storage-cli":         "infrastructure",
	"cloud-utils":               "infrastructure",
	"container-management":      "infrastructure",
	"core-service":              "infrastructure",
	"desktop-service":           "infrastructure",
	"dev-ops":                   "infrastructure",
	"devops":                    "infrastructure",
	"devops-tooling":            "infrastructure",
	"devops-tools":              "infrastructure",
	"distributed-systems":       "infrastructure",
	"homelab-management":        "infrastructure",
	"infrastructure-as-code":    "infrastructure",
	"infrastructure-automation": "infrastructure",
	"serverless-framework":      "infrastructure",
	"virtualization-management": "infrastructure",

	"k8s-security":             "kubernetes",
	"kubernetes-tooling":       "kubernetes",
	"kubernetes-tools":         "kubernetes",
	"kubernetes-observability": "kubernetes",
	"kubernetes-operator":      "kubernetes",

	"application-security": "security",
	"auth-library":         "security",
	"authz-cache":          "security",
	"cloud-security":       "security",
	"cyber-security":       "security",
	"cybersecurity":        "security",
	"iam-integration":      "security",
	"identity-protocol":    "security",
	"linux-security":       "security",
	"security-tool":        "security",
	"security-tools":       "security",
	"windows-security":     "security",

	"database-drivers":    "database",
	"database-library":    "database",
	"database-management": "database",
	"database-orm":        "database",
	"database-security":   "database",
	"database-system":     "database",
	"database-tooling":    "database",
	"database-tools":      "database",
	"data-storage":        "database",
	"metadata-management": "database",
	"sql-parser":          "database",

	"network-api":          "networking",
	"network-automation":   "networking",
	"network-monitoring":   "networking",
	"network-optimization": "networking",
	"network-protocol":     "networking",
	"network-security":     "networking",
	"network-server":       "networking",
	"network-utility":      "networking",
	"networking-config":    "networking",
	"networking-tool":      "networking",
	"networking-tools":     "networking",
	"radio-communication":  "networking",
	"routeros-script":      "networking",

	"admin-backend":         "web-development",
	"documentation-website": "web-development",
	"fullstack-boilerplate": "web-development",
	"nextjs-boilerplate":    "web-development",
	"web-crawling":          "web-development",
	"web-framework":         "web-development",
	"web-frameworks":        "web-development",
	"web-hosting":           "web-development",
	"web-scraping":          "web-development",
	"web-servers":           "web-development",
	"webassembly":           "web-development",
	"webgl-graphics":        "web-development",

	"data-collection":          "data-engineering",
	"data-repository":          "data-engineering",
	"data-scraping":            "data-engineering",
	"pdf-creation":             "data-engineering",
	"cloud-monitoring":         "observability",
	"ci-metrics":               "observability",
	"monitoring-tool":          "observability",
	"monitoring-tools":         "observability",
	"monitoring":               "observability",
	"monitoring-observability": "observability",
	"security-monitoring":      "observability",

	"chatops-agents":       "automation",
	"home-automation":      "automation",
	"industrial-iot":       "iot",
	"iot-dashboard":        "iot",
	"iot-management":       "iot",
	"robotics-competition": "iot",

	"discord-bot":      "bots",
	"e-commerce-bot":   "bots",
	"social-media-bot": "bots",
	"telegram-bot":     "bots",
	"terminal-agent":   "bots",

	"e-commerce-sdk":     "productivity",
	"obsidian-plugin":    "productivity",
	"productivity-app":   "productivity",
	"productivity-tool":  "productivity",
	"project-management": "productivity",

	"audio-engine":        "media",
	"audio-utility":       "media",
	"document-processing": "media",
	"file-sharing":        "media",
	"iptv-streaming":      "media",
	"media-streaming":     "media",
	"media-tools":         "media",
	"music-server":        "media",
	"streaming-api":       "media",
	"video-streaming":     "media",
	"visual-novel-patch":  "media",

	"emulation-engine":  "gaming",
	"game-development":  "gaming",
	"game-engine":       "gaming",
	"game-modification": "gaming",
	"game-revival":      "gaming",
	"gaming-ai":         "gaming",
	"gaming-client":     "gaming",
	"minecraft-plugin":  "gaming",

	"blockchain-curation":       "blockchain",
	"blockchain-infrastructure": "blockchain",
	"blockchain-tools":          "blockchain",
	"crypto-wallet":             "blockchain",

	"ad-blocking":         "privacy",
	"adblock-filterlists": "privacy",
	"privacy-tools":       "privacy",
	"proxy-configs":       "privacy",
	"proxy-management":    "privacy",
	"proxy-subscription":  "privacy",
	"reverse-proxy":       "privacy",
	"vpn-client":          "privacy",
	"vpn-server":          "privacy",
	"vpn-tools":           "privacy",

	"aprs-stack":           "communications",
	"discussion-platforms": "communications",
	"messaging-bridge":     "communications",
	"push-notifications":   "communications",
	"realtime-comm":        "communications",

	"concurrent-programming": "systems",
	"gpu-framework":          "systems",
	"hardware-hacking":       "systems",
	"hpc-libraries":          "systems",
	"linux-config":           "systems",

	"documentation": "documentation",
	"ui-components": "ui-components",

	"mathematical-research": "science",
	"medical-software":      "science",
	"scientific-simulation": "science",
	"weather-forecast":      "science",
}

var keywordRules = []struct {
	Needle   string
	Category string
}{
	{"kubernetes", "kubernetes"},
	{"k8s", "kubernetes"},
	{"security", "security"},
	{"auth", "security"},
	{"database", "database"},
	{"sql", "database"},
	{"network", "networking"},
	{"vpn", "privacy"},
	{"proxy", "privacy"},
	{"privacy", "privacy"},
	{"web", "web-development"},
	{"cloud", "infrastructure"},
	{"devops", "infrastructure"},
	{"infrastructure", "infrastructure"},
	{"monitoring", "observability"},
	{"observability", "observability"},
	{"ai", "artificial-intelligence"},
	{"llm", "artificial-intelligence"},
	{"machine-learning", "artificial-intelligence"},
	{"data", "data-engineering"},
	{"bot", "bots"},
	{"automation", "automation"},
	{"iot", "iot"},
	{"media", "media"},
	{"streaming", "media"},
	{"game", "gaming"},
	{"blockchain", "blockchain"},
	{"crypto", "blockchain"},
	{"terminal", "developer-tools"},
	{"tool", "developer-tools"},
	{"framework", "developer-tools"},
	{"sdk", "developer-tools"},
}

func NormalizeCategory(category string) string {
	category = slug(category)
	if alias, ok := directAliases[category]; ok {
		return alias
	}
	if isCanonical(category) {
		return category
	}
	for _, rule := range keywordRules {
		if strings.Contains(category, rule.Needle) {
			return rule.Category
		}
	}
	return "uncategorized"
}

func Categories() []string {
	out := make([]string, len(canonicalCategories))
	copy(out, canonicalCategories)
	return out
}

func slug(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "-")
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.Trim(value, "-")
	return value
}

func isCanonical(category string) bool {
	for _, item := range canonicalCategories {
		if category == item {
			return true
		}
	}
	return false
}
