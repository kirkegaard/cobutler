package models

import "strings"

// Templates for common code constructs in different programming languages
var codeTemplates = map[string]map[string]string{
	"go": {
		"function":  "func NAME(PARAMS) RETURNS {\n\t\n}",
		"main":      "func main() {\n\t\n}",
		"if":        "if CONDITION {\n\t\n}",
		"for":       "for i := 0; i < N; i++ {\n\t\n}",
		"range":     "for _, item := range items {\n\t\n}",
		"struct":    "type NAME struct {\n\t\n}",
		"interface": "type NAME interface {\n\t\n}",
		"switch":    "switch VALUE {\ncase CASE:\n\t\ndefault:\n\t\n}",
		"import":    "import (\n\t\"package\"\n)",
		"error":     "if err != nil {\n\treturn err\n}",
	},
	"javascript": {
		"function": "function NAME(PARAMS) {\n\t\n}",
		"arrow":    "(PARAMS) => {\n\t\n}",
		"if":       "if (CONDITION) {\n\t\n}",
		"for":      "for (let i = 0; i < N; i++) {\n\t\n}",
		"foreach":  "items.forEach(item => {\n\t\n});",
		"class":    "class NAME {\n\tconstructor(PARAMS) {\n\t\t\n\t}\n}",
		"import":   "import { MODULE } from 'package';",
		"try":      "try {\n\t\n} catch (error) {\n\t\n}",
		"const":    "const NAME = VALUE;",
		"let":      "let NAME = VALUE;",
	},
	"python": {
		"function": "def NAME(PARAMS):\n\t",
		"class":    "class NAME:\n\tdef __init__(self, PARAMS):\n\t\t",
		"if":       "if CONDITION:\n\t",
		"for":      "for i in range(N):\n\t",
		"foreach":  "for item in items:\n\t",
		"with":     "with EXPRESSION as VAR:\n\t",
		"try":      "try:\n\t\nExcept Exception as e:\n\t",
		"import":   "import MODULE",
		"from":     "from MODULE import SUBMODULE",
		"main":     "if __name__ == \"__main__\":\n\t",
	},
	"ruby": {
		"function": "def NAME(PARAMS)\n\t\nend",
		"class":    "class NAME\n\tdef initialize(PARAMS)\n\t\t\n\tend\nend",
		"if":       "if CONDITION\n\t\nend",
		"for":      "for i in 0..N\n\t\nend",
		"foreach":  "items.each do |item|\n\t\nend",
		"begin":    "begin\n\t\nrescue Exception => e\n\t\nend",
		"module":   "module NAME\n\t\nend",
		"require":  "require 'MODULE'",
	},
	"java": {
		"class":     "public class NAME {\n\t\n}",
		"method":    "public TYPE NAME(PARAMS) {\n\t\n}",
		"main":      "public static void main(String[] args) {\n\t\n}",
		"if":        "if (CONDITION) {\n\t\n}",
		"for":       "for (int i = 0; i < N; i++) {\n\t\n}",
		"foreach":   "for (Type item : items) {\n\t\n}",
		"try":       "try {\n\t\n} catch (Exception e) {\n\t\n}",
		"switch":    "switch (VALUE) {\n\tcase CASE:\n\t\tbreak;\n\tdefault:\n\t\tbreak;\n}",
		"import":    "import PACKAGE;",
		"interface": "public interface NAME {\n\t\n}",
	},
	"c": {
		"function": "TYPE NAME(PARAMS) {\n\t\n}",
		"main":     "int main(int argc, char *argv[]) {\n\t\n\treturn 0;\n}",
		"if":       "if (CONDITION) {\n\t\n}",
		"for":      "for (int i = 0; i < N; i++) {\n\t\n}",
		"while":    "while (CONDITION) {\n\t\n}",
		"struct":   "struct NAME {\n\t\n};",
		"switch":   "switch (VALUE) {\n\tcase CASE:\n\t\tbreak;\n\tdefault:\n\t\tbreak;\n}",
		"include":  "#include <HEADER>",
		"define":   "#define NAME VALUE",
	},
	"cpp": {
		"class":     "class NAME {\nprivate:\n\t\npublic:\n\t\n};",
		"function":  "TYPE NAME(PARAMS) {\n\t\n}",
		"main":      "int main(int argc, char *argv[]) {\n\t\n\treturn 0;\n}",
		"if":        "if (CONDITION) {\n\t\n}",
		"for":       "for (int i = 0; i < N; i++) {\n\t\n}",
		"foreach":   "for (const auto& item : items) {\n\t\n}",
		"include":   "#include <HEADER>",
		"namespace": "namespace NAME {\n\t\n}",
		"template":  "template<typename T>\nclass NAME {\n\t\n};",
	},
	"rust": {
		"function": "fn NAME(PARAMS) -> RETURNS {\n\t\n}",
		"main":     "fn main() {\n\t\n}",
		"if":       "if CONDITION {\n\t\n}",
		"for":      "for i in 0..N {\n\t\n}",
		"foreach":  "for item in items {\n\t\n}",
		"struct":   "struct NAME {\n\t\n}",
		"impl":     "impl NAME {\n\t\n}",
		"match":    "match VALUE {\n\tPATTERN => {\n\t\t\n\t},\n\t_ => {\n\t\t\n\t},\n}",
		"use":      "use PACKAGE;",
		"trait":    "trait NAME {\n\t\n}",
	},
	"lua": {
		"function":       "function NAME(PARAMS)\n\t\nend",
		"local_function": "local function NAME(PARAMS)\n\t\nend",
		"if":             "if CONDITION then\n\t\nend",
		"for":            "for i = 1, N do\n\t\nend",
		"foreach":        "for _, item in ipairs(items) do\n\t\nend",
		"while":          "while CONDITION do\n\t\nend",
		"table":          "local NAME = {\n\t\n}",
		"require":        "local MODULE = require('PACKAGE')",
	},
}

// FindTemplate looks for a suitable template based on a prefix and language
func FindTemplate(prefix string, language string) (string, bool) {
	// Default to "go" if language not specified
	if language == "" {
		language = "go"
	}

	// Get templates for this language
	langTemplates, exists := codeTemplates[language]
	if !exists {
		return "", false
	}

	// Check for direct matches with template keys
	for pattern, template := range langTemplates {
		if strings.HasPrefix(prefix, pattern) {
			return template, true
		}
	}

	// Look for more fuzzy matches
	for pattern, template := range langTemplates {
		if strings.Contains(prefix, pattern) {
			return template, true
		}
	}

	return "", false
}
