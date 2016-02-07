package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// A helper method to panic on error.
func check(err error) {
	if err != nil {
		panic(err)
	}
}

// The the top of each file will be the "go:generate" and ""+build ignore"
// directives. They both need to be removed in the generated version of the
// source code.
func stripGoGenerateAndBuildIgnore(str string) string {
	lines := strings.Split(str, "\n")
	for lineNumber := range lines {
		line := lines[lineNumber]
		if strings.HasPrefix(line, "//go:generate") ||
			strings.HasPrefix(line, "// +build") {
			// It is important we don't remove these so we keep the original lines
			// numbers the same.
			lines[lineNumber] = "//"
		}
	}

	return strings.Join(lines, "\n")
}

// Disect the parameter names and values. For a string like "a: int, b: int" the
// result will be:
//
//     [][]string{[]string{"a", " int"}, []string{"b", " int"}}
//
// We cannot use a map becuase the order of the keys are not preserved in Go.
func getParameters(str string, isDefinition bool) [][]string {
	var params [][]string
	var re *regexp.Regexp

	if isDefinition {
		re = regexp.MustCompile(",?\\s*([a-zA-Z0-9_\\s,]+):([^,]+)")
	} else {
		re = regexp.MustCompile(",?\\s*([a-zA-Z0-9_]+)\\s*:")
	}

	arguments := SplitOnRegexpIncludingDelimiter(re, str)
	for _, argument := range arguments {
		if !isDefinition {
			re = regexp.MustCompile(",?\\s*([a-zA-Z0-9_]+)\\s*:(.*)")
		}
		groups := re.FindAllStringSubmatch(argument, -1)

		if len(groups) > 0 {
			if isDefinition {
				individualParams := strings.Split(groups[0][1], ",")
				for _, individualParam := range individualParams {
					params = append(params, []string{strings.TrimSpace(individualParam), groups[0][2]})
				}
			} else {
				params = append(params, []string{groups[0][1], groups[0][2]})
			}
		}
	}

	return params
}

func replaceFunctionDefinitions(contents string) string {
	search := regexp.MustCompile("func\\s+([a-zA-Z0-9_]+)\\(([a-zA-Z0-9_\\s,]+:.*)\\)")
	return ReplaceAllGroupFunc(search, contents, func(groups []string) string {
		params := getParameters(groups[2], true)
		definition := "func " + groups[1]
		for _, value := range params {
			individualValues := strings.Split(value[0], ",")
			for _, individualValue := range individualValues {
				definition += "_" + individualValue
			}
		}
		definition += "("
		first := true
		for _, value := range params {
			if !first {
				definition += ", "
			}
			first = false
			definition += value[0] + " " + strings.TrimSpace(value[1])
		}
		return definition + ")"
	})
}

func skipPrefixUntilSuffix(str, out, prefix, suffix string, i int) (int, string) {
	if i < len(str) - len(prefix) && str[i:i + len(prefix)] == prefix {
		for str[i:i + len(suffix)] != suffix {
			out += string(str[i])
			i++
		}
	}

	return i, out
}

func skipStrings(str, out string, i int) (int, string) {
	var c byte
	if str[i] == '"' || str[i] == '\'' {
		c = str[i]
	} else {
		return i, out
	}

	i++
	out += string(c)
	for {
		if str[i] == '\\' {
			out += string(str[i:i + 2])
			i++
		} else if str[i] == c {
			break
		} else {
			out += string(str[i])
		}
		i++
	}
	i++
	out += string(c)

	return i, out
}

// To allow the regular expressions to work recursively we need to replace all
// the opening and closing parenthesis with a depth indicator then apply the
// regular expression starting with the deepest first.
//
// The opening and closing parenthesis use the same syntax to indicate the
// depth: `~depth~` where `((1 + 2) * 3)` would translate to
// `~0~~1~1 + 2~1~ * 3~0~`.
//
// It is also context aware in that it will not replace brackets that are found
// in comments and strings.
func prepareBrackets(str string) (string, int) {
	depth := 0
	maxDepth := 0
	out := ""
	for i := 0; i < len(str); i++ {
		// Skip comments.
		i, out = skipPrefixUntilSuffix(str, out, "//", "\n", i)
		i, out = skipPrefixUntilSuffix(str, out, "/*", "*/", i)

		// Skip strings and characters.
		i, out = skipStrings(str, out, i)

		// Everything else.
		if str[i] == '(' {
			out += fmt.Sprintf("~%d~", depth)
			depth += 1
			if depth > maxDepth {
				maxDepth = depth
			}
		} else if str[i] == ')' {
			depth -= 1
			out += fmt.Sprintf("~%d~", depth)
		} else {
			out += string(str[i])
		}
	}

	return out, maxDepth
}

func replaceFunctionInvocations(str string) string {
	str, maxDepth := prepareBrackets(str)

	for depth := maxDepth; depth >= 0; depth -= 1 {
		search := regexp.MustCompile(fmt.Sprintf("([a-zA-Z0-9_]*)~%d~(.*)~%d~", depth, depth))
		str = ReplaceAllGroupFunc(search, str, func(groups []string) string {
			params := getParameters(groups[2], false)
			definition := groups[1]
			for _, value := range params {
				definition += "_" + value[0]
			}
			definition += "("

			if len(params) > 0 {
				first := true
				for _, value := range params {
					if !first {
						definition += ", "
					}
					first = false
					definition += strings.TrimSpace(value[1])
				}
			} else {
				definition += groups[2]
			}

			return definition + ")"
		})
	}

	return str
}

// This works just like regexp.Split() except that the regexp that is the
// delimiter is included in the results. For example splitting "foo-bar-baz" on
// "-" would return []string{"foo", "-bar", "-baz"}
func SplitOnRegexpIncludingDelimiter(re *regexp.Regexp, str string) []string {
	result := []string{}
	lastIndex := -1
	for _, slice := range re.FindAllStringIndex(str, -1) {
		if lastIndex < 0 {
			lastIndex = 0
		} else {
			result = append(result, str[lastIndex:slice[0]])
			lastIndex = slice[0]
		}
	}

	if lastIndex < 0 {
		lastIndex = 0
	}

	result = append(result, str[lastIndex:])

	return result
}

// https://gist.github.com/elliotchance/d419395aa776d632d897
func ReplaceAllGroupFunc(re *regexp.Regexp, str string, repl func([]string) string) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			group := ""
			if v[i] >= 0 {
				group = str[v[i]:v[i+1]]
			}
			groups = append(groups, group)
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func main() {
	if len(os.Args) < 2 {
		panic("No files specified.")
	}

	raw, err := ioutil.ReadFile(os.Args[1])
	check(err)
	contents := string(raw)

	contents = stripGoGenerateAndBuildIgnore(contents)
	contents = replaceFunctionDefinitions(contents)
	contents = replaceFunctionInvocations(contents)

	//fmt.Printf("%s", contents)

	err = ioutil.WriteFile(os.Args[1] + ".go", []byte(contents), 0644)
  check(err)
}
