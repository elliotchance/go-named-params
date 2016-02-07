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
func getParameters(str string) [][]string {
	var params [][]string

	splitRe := regexp.MustCompile(",?\\s*[a-zA-Z0-9_]+\\s*:")
	arguments := SplitOnRegexpIncludingDelimiter(splitRe, str)
	for _, argument := range arguments {
		re := regexp.MustCompile(",?\\s*([a-zA-Z0-9_]+)\\s*:(.*)")
		groups := re.FindAllStringSubmatch(argument, -1)

		if len(groups) > 0 {
			params = append(params, []string{groups[0][1], groups[0][2]})
		}
	}

	return params
}

func replaceFunctionDefinitions(contents string) string {
	search := regexp.MustCompile("func\\s+([a-zA-Z0-9_]+)\\(([a-zA-Z0-9_]+\\s*:.*)\\)")
	return ReplaceAllGroupFunc(search, contents, func(groups []string) string {
		params := getParameters(groups[2])
		definition := "func " + groups[1]
		for _, value := range params {
			definition += "_" + value[0]
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

func prepareBrackets(str string) (string, int) {
	depth := 0
	maxDepth := 0
	out := ""
	for _, c := range str {
		if c == '(' {
			out += fmt.Sprintf("~%d~", depth)
			depth += 1
			if depth > maxDepth {
				maxDepth = depth
			}
		} else if c == ')' {
			depth -= 1
			out += fmt.Sprintf("~%d~", depth)
		} else {
			out += string(c)
		}
	}

	return out, maxDepth
}

func replaceFunctionInvocations(str string) string {
	str, maxDepth := prepareBrackets(str)

	for depth := maxDepth; depth >= 0; depth -= 1 {
		search := regexp.MustCompile(fmt.Sprintf("([a-zA-Z0-9_]*)~%d~(.*)~%d~", depth, depth))
		str = ReplaceAllGroupFunc(search, str, func(groups []string) string {
			params := getParameters(groups[2])
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
