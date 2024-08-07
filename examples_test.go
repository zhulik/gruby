package gruby_test

import (
	"fmt"

	"github.com/zhulik/gruby"
)

func ExampleGRuby_DefineClass() {
	grb := gruby.New()
	defer grb.Close()

	// Our custom function we'll expose to Ruby
	addFunc := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		args := grb.GetArgs()
		return gruby.ToRuby(grb, gruby.ToGo[int](args[0])+gruby.ToGo[int](args[1])), nil
	}

	// Lets define a custom class and a class method we can call.
	class := grb.DefineClass("Example", nil)
	class.DefineClassMethod("add", addFunc, gruby.ArgsReq(2))

	// Let's call it and inspect the result
	result, err := grb.LoadString(`Example.add(12, 30)`)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Result: %s\n", result.String())
	// Output:
	// Result: 42
}

func ExampleDecode() {
	grb := gruby.New()
	defer grb.Close()

	// Our custom function we'll expose to Ruby
	var logData interface{}
	logFunc := func(grb *gruby.GRuby, self gruby.Value) (gruby.Value, gruby.Value) {
		args := grb.GetArgs()
		if err := gruby.Decode(&logData, args[0]); err != nil {
			panic(err)
		}

		return nil, nil
	}

	// Lets define a custom class and a class method we can call.
	class := grb.DefineClass("Example", nil)
	class.DefineClassMethod("log", logFunc, gruby.ArgsReq(1))

	// Let's call it and inspect the result
	if _, err := grb.LoadString(`Example.log({"foo" => "bar"})`); err != nil {
		panic(err.Error())
	}

	fmt.Printf("Result: %v\n", logData)
	// Output:
	// Result: map[foo:bar]
}

func ExampleCompileContext() {
	grb := gruby.New()
	defer grb.Close()

	ctx1 := gruby.NewCompileContext(grb)
	defer ctx1.Close()
	ctx1.SetFilename("foo.rb")

	ctx2 := gruby.NewCompileContext(grb)
	defer ctx2.Close()
	ctx2.SetFilename("bar.rb")

	parser := gruby.NewParser(grb)
	defer parser.Close()

	if _, err := parser.Parse("def foo; bar; end", ctx1); err != nil {
		panic(err.Error())
	}
	code1 := parser.GenerateCode()

	if _, err := parser.Parse("def bar; 42; end", ctx2); err != nil {
		panic(err.Error())
	}
	code2 := parser.GenerateCode()

	if _, err := grb.Run(code1, nil); err != nil {
		panic(err.Error())
	}

	if _, err := grb.Run(code2, nil); err != nil {
		panic(err.Error())
	}

	result, err := grb.LoadString("foo")
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Result: %s\n", result)
	// Output:
	// Result: 42
}
