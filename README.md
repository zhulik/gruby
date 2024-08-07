# Gruby - Go bindings to mruby

This is a fork of amazing [gruby](https://github.com/mitchellh/gruby). 
I'm not sure if I want and I can maintain it for a long time, but I updated it 
to support mruby 3.3 and go 1.22.

*The project in being heavily refactored and partially rewritten.*
*Backwards compatibility with go-mruby is broken.*

gruby provides [mruby](https://github.com/mruby/mruby) bindings for
[Go](http://golang.org). This allows Go applications to run a lightweight
embedded Ruby VM. Using the mruby library, Go applications can call Ruby
code, and Ruby code can call Go code (that is properly exposed)!

At the time of writing, this is the most comprehensive mruby library for
Go _by far_. It is also the only mruby library for Go that enables exposing
Go functions to Ruby as well as being able to generically convert complex
Ruby types into Go types. Our goal is to implement all of the mruby API.

**Project Status:** The major portions of the mruby API are implemented,
but the mruby API is huge. If there is something that is missing, please
issue a pull request and I'd be happy to add it! We're also not yet ready
to promise API backwards compatibility on a Go-level, but we're getting there.

## Installation

Installation is a little trickier than a standard Go library, but not
by much. You can't simply `go get` this library, unfortunately. This is
because [mruby](https://github.com/mruby/mruby) must first be built. We
don't ship a pre-built version of mruby because the build step of mruby
is important in customizing what aspects of the standard library you want
available, as well as any other extensions.

To build mruby, we've made it very easy. You will need the following packages
available on your host operating system:

* bison
* flex
* ruby 3.x

Then just type:

```
$ make
```

This will download mruby, compile it, and run the tests for gruby,
verifying that your build is functional. By default, gruby will download
and build a default version of mruby, but this is customizable.

Compiling/installing the gruby library should work on Linux, Mac OS X,
and Windows. On Windows, msys is the only supported build toolchain (same
as Go itself).

**Due to this linking, it is strongly recommended that you vendor this
repository and bake our build system into your process.**

### Customizing the mruby Compilation

You can customize the mruby compilation by setting a couple environmental
variables prior to calling `make`:

  * `MRUBY_COMMIT` is the git ref that will be checked out for gruby. This
    defaults to to a recently tagged version. Many versions before 1.2.0 do not
    work with gruby. It is recommend you explicitly set this to a ref that
    works for you to avoid any changes in this library later.

  * `MRUBY_CONFIG` is the path to a `build_config.rb` file used to configure
    how mruby is built. If this is not set, gruby will use the default
    build config that comes with gruby. You can learn more about configuring
    the mruby build [here](https://github.com/mruby/mruby/tree/master/doc/guides/compile.md).

## Usage

gruby exposes the mruby API in a way that is idiomatic Go, so that it
is comfortable to use by a standard Go programmer without having intimate
knowledge of how mruby works.

For usage examples and documentation, please see the
[gruby GoDoc](http://godoc.org/github.com/zhulik/gruby), which
we keep up to date and full of examples.

For a quick taste of what using gruby looks like, though, we provide
an example below:

```go
package main

import (
	"fmt"
	"github.com/zhulik/gruby"
)

func main() {
	mrb := gruby.NewMrb()
	defer mrb.Close()

	// Our custom function we'll expose to Ruby. The first return
	// value is what to return from the func and the second is an
	// exception to raise (if any).
	addFunc := func(m *mruby.Mrb, self *mruby.MrbValue) (mruby.Value, gruby.Value) {
		args := m.GetArgs()
		return gruby.Int(ToGo[int](args[0]) + ToGo[int](args[1])), nil
	}

	// Lets define a custom class and a class method we can call.
	class := mrb.DefineClass("Example", nil)
	class.DefineClassMethod("add", addFunc, gruby.ArgsReq(2))

	// Let's call it and inspect the result
	result, err := mrb.LoadString(`Example.add(12, 30)`)
	if err != nil {
		panic(err.Error())
	}

	// This will output "Result: 42"
	fmt.Printf("Result: %s\n", result.String())
}
```
