# Example

```go
import "fmt"

func add(x: int, y: int) {
	return x + y
}

func main() {
	fmt.Printf("%d", add(x: 3, y: 7))
}
```

# Installation

```bash
go get github.com/elliotchance/go-named-params
```

# Usage

You need to include two lines at the top of each file:

```go
//go:generate python $GOPATH/src/github.com/elliotchance/go-named-params/compile.py $GOFILE
// +build ignore
```

Then use `go generate` to process the files:

```bash
go generate
go build
```

# Description

Using functions with named parameters makes code much easier to read. Consider
the following code:

```go
sayHello("Elliot", false)
```

Can you guess what the `false` is for? You have to find the function definition
for `sayHello` and look at the parameters name which may be located in a file
that's hard to find if your not using an IDE:

```go
func sayHello(name string, alreadyGreeted bool)
```

---

Now imagine the same invocation written with named arguments:

```go
sayHello(name: "Elliot", alreadyGreeted: false)
```

Much more clear! You don't have to goto the function definition to see what the
arguments are, but if you did you would see it is a very similar syntax just
adding `:` for each argument:

```go
func sayHello(name: string, alreadyGreeted: bool)
```

---

There are some important things to know about this implementation:

1. Most notably this is not native Go syntax. You will have to run `go generate`
   before you can build or run the files.

2. Go does not support method overloading (where multiple methods can be defined
	 with the same name but different signatures) for good reason. However, when
	 using named parameters they become part of the method name to make it unique.
	 Here is what it looks like when it is translated:

    ```go
func sayHello(name: string, alreadyGreeted: bool)
//func sayHello_name_alreadyGreeted(name string, alreadyGreeted bool)
    ```

3. The function invocation does not need to know about the definition because
   they both contain the argument names. This is very important in being able to
	 translate the called without having to pass through any external source, or
	 even pass through an AST. It is literally a regular expressions replace.

4. All code generated is *undoable*. If you need to remove the named arguments
   from production code so that it becomes native Go you can parse the files and
	 operate on the results. If you relly want to strip out all traces you can use
	 `gofmt -r 'sayHello_name_alreadyGreeted -> sayHello' -d ./`

5. Since the parameters become part of the method name, you can use this for
   appropriate overloading:

    ```go
func writeToFile(path: string, fromInt: int)
func writeToFile(path: string, fromString: string)
    ```
