//go:generate python $GOPATH/src/github.com/elliotchance/go-named-params/compile.py $GOFILE
// +build ignore

package main

// This file is used to test the added functionality of named paramaters for
// functions in Go. The individual function names follow a pattern:
//
// anon or named: Whether the function would be a pure (anonymous function) or
//     a function with named parameters.
// test number: A sequential number to indicate different versions of the same
//     group.
// version: Represents the version of test, where:
//     0: no arguments
//     1: a single argument
//     2: multiple arguments
//     3: multiword argument
//     4: return value
//     5: multiple arguments short syntax
//
// The functions themselves do not contain any body becuase if something goes
// wrong with the regular expression, the compiler will throw an error.

// These things are pure Go, they are here as to make sure that the parser
// doesn't mess with any of the existing syntax:
func anon10() {
}
func anon11(name string) {
}
func anon12(a int, b int) {
}
func anon13(c chan int) {
}
func anon14(a int, b int) int {
  return a + b
}
func anon15(a, b int, c string) int {
  return a + b * len(c)
}

// Here are the versions with the named parameters:
func named11(name: string) {
}
func named12(a: int, b: int) {
}
func named13(c: chan int) {
}
func named14(a: int, b: int) int {
  return a + b
}
func named15(a, b: int, c: string) int {
  return a + b * len(c)
}

// Helper functions
func check(result, expectedResult int) {
  if result != expectedResult {
    panic("Failed!")
  }
}

func main() {
  var result int

  // Simply calling them.
  anon10()
  anon11("bob")
  anon12(3, 2)
  anon13(make(chan int))
  result = anon14(3, 5)
  check(result, 8)
  result = anon15(3, 2, "foo")
  check(result, 9)

  named11(name: "bob")
  named12(a: 3, b: 2)
  named13(c: make(chan int))
  result = named14(a: 2, b: 3)
  check(result, 5)
  result = named15(a: 3, b: 2, c: "foo")
  check(result, 9)

  // Different combinations of nesting.
  result = named14(a: named14(a: 7, b: 4), b: 2)
  check(result, 13)
  result = named14(a: anon14(7, 4), b: 2)
  check(result, 13)
  result = anon14(named14(a: 7, b: 4), 2)
  check(result, 13)

  // Grouping brackets should not be affected.
  named12(a: 3, b: (2 + 3))
  result = (5 * 2)
  check(result, 10)
  result = ((((((((1 + 3))))))))
  check(result, 4)
}
