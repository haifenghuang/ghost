package evaluator

import (
	"testing"

	"ghostlang.org/x/ghost/lexer"
	"ghostlang.org/x/ghost/object"
	"ghostlang.org/x/ghost/parser"
	"ghostlang.org/x/ghost/utilities"
	"ghostlang.org/x/ghost/value"
	"github.com/shopspring/decimal"
)

func TestEvalNumberExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"5", 5},
		{"10", 10},
		{"-5", -5},
		{"-10", -10},
		{"5 + 5 + 5 + 5 - 10", 10},
		{"2 * 2 * 2 * 2 * 2", 32},
		{"-50 + 100 + -50", 0},
		{"5 * 2 + 10", 20},
		{"5 + 2 * 10", 25},
		{"20 + 2 * -10", 0},
		{"50 / 2 * 2 + 10", 60},
		{"2 * (5 + 10)", 30},
		{"3 * 3 * 3 + 10", 37},
		{"3 * (3 * 3) + 10", 37},
		{"(5 + 10 * 2 + 15 / 3) * 2 + -10", 50},
		{"index := 0; index++; index", 1},
		{"index := 6; index--; index", 5},
		{"index := 0; index += 10; index", 10},
		{"index := 12; index -= 2; index", 10},
		{"index := 2; index *= 5; index", 10},
		{"index := 100; index /= 10; index", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testNumberObject(t, evaluated, tt.expected)
	}
}

func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"true and true", true},
		{"true and false", false},
		{"true or false", true},
		{"false or true", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		boolean, ok := tt.expected.(bool)

		if ok {
			testBooleanObject(t, evaluated, boolean)
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestDecimal(t *testing.T) {
	dec1 := decimal.NewFromInt(1)
	dec2 := decimal.NewFromInt(1)

	t.Logf("%v", dec1.LessThanOrEqual(dec2))
}

func TestRangeOperators(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		// {`1 .. 0`, []int{}},
		{`-1 .. 0`, []int{-1, 0}},
		{`1 .. 1`, []int{1}},
		{`1 .. 5`, []int{1, 2, 3, 4, 5}},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case []int:
			list, ok := evaluated.(*object.List)

			if !ok {
				t.Errorf("object not List. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if len(list.Elements) != len(expected) {
				t.Errorf("wrong number of elements. want=%d, got=%d", len(expected), len(list.Elements))
				continue
			}
		}
	}
}

func TestAssignStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"x := 10; x", 10},
		{"x := 10; x := 20; x", 20},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int64:
			testNumberObject(t, evaluated, expected)
		case string:
			errObj, ok := evaluated.(*object.Error)

			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}

func TestEvalBooleanExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true", true},
		{"false", false},
		{"1 < 2", true},
		{"1 > 2", false},
		{"1 < 1", false},
		{"1 > 1", false},
		{"1 == 1", true},
		{"1 != 1", false},
		{"1 == 2", false},
		{"1 != 2", true},
		{"true == true", true},
		{"false == false", true},
		{"true == false", false},
		{"true != false", true},
		{"false != true", true},
		{"(1 < 2) == true", true},
		{"(1 < 2) == false", false},
		{"(1 > 2) == true", false},
		{"(1 > 2) == false", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestBangOperator(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"!true", false},
		{"!false", true},
		{"!5", false},
		{"!!true", true},
		{"!!false", false},
		{"!!5", true},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testBooleanObject(t, evaluated, tt.expected)
	}
}

func TestIfElseExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"if (true) { 10 }", 10},
		{"if (false) { 10 }", nil},
		{"if (1) { 10 }", 10},
		{"if (1 < 2) { 10 }", 10},
		{"if (1 > 2) { 10 }", nil},
		{"if (1 > 2) { 10 } else { 20 }", 20},
		{"if (1 < 2) { 10 } else { 20 }", 10},
		{"if (1 < 2) { 10 } else if (1 == 1) { 20 } else { 30 }", 10},
		{"if (1 > 2) { 10 } else if (1 == 1) { 20 } else { 30 }", 20},
		{"if (1 > 2) { 10 } else if (1 == 2) { 20 } else { 30 }", 30},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(int)

		if ok {
			testNumberObject(t, evaluated, int64(number))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestReturnStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"return 10;", 10},
		{"return 10; 9;", 10},
		{"return 2 * 5; 9;", 10},
		{"9; return 2 * 5; 9;", 10},
		{"if (10 > 1) { if (10 > 1) { return 10; } return 1; }", 10},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		testNumberObject(t, evaluated, tt.expected)
	}
}

func TestLetStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"a := 5; a;", 5},
		{"a := 5 * 5; a;", 25},
		{"a := 5; b := a; b;", 5},
		{"a := 5; b := a; c := a + b + 5; c;", 15},
		{"a := 5; a = 10; a;", 10},
	}

	for _, tt := range tests {
		testNumberObject(t, testEval(tt.input), tt.expected)
	}
}

func TestNamedFunctionStatements(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"function five() { return 5 } five()", 5},
		{"function ten() { return 10 } ten()", 10},
		{"function fifteen() { return 15 } fifteen()", 15},
	}

	for _, tt := range tests {
		testNumberObject(t, testEval(tt.input), tt.expected)
	}
}

func TestFunctionObject(t *testing.T) {
	input := "function(x) { x + 2; };"

	evaluated := testEval(input)

	function, ok := evaluated.(*object.Function)

	if !ok {
		t.Fatalf("object is not Function. got=%T (%+v)", evaluated, evaluated)
	}

	if len(function.Parameters) != 1 {
		t.Fatalf("function has wrong parameters. Parameters=%+v", function.Parameters)
	}

	if function.Parameters[0].String() != "x" {
		t.Fatalf("parameter is not 'x'. got=%q", function.Parameters[0])
	}

	expectedBody := "(x + 2)"

	if function.Body.String() != expectedBody {
		t.Fatalf("body is not %q. got=%q", expectedBody, function.Body.String())
	}
}

func TestFunctionApplication(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"identity := function(x) { x; }; identity(5);", 5},
		{"identity := function(x) { return x; }; identity(5);", 5},
		{"double := function(x) { x * 2; }; double(5);", 10},
		{"add := function(x, y) { x + y; }; add(5, 5);", 10},
		{"add := function(x, y) { x + y; }; add(5 + 5, add(5, 5));", 20},
		{"function(x) { x; }(5)", 5},
	}

	for _, tt := range tests {
		testNumberObject(t, testEval(tt.input), tt.expected)
	}
}

func TestStringLiteral(t *testing.T) {
	input := `"Hello World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)

	if !ok {
		t.Fatalf("object is not string. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. expected=%q, got=%q", "Hello World!", str.Value)
	}
}

func TestStringConcatenation(t *testing.T) {
	input := `"Hello" + " " + "World!"`

	evaluated := testEval(input)
	str, ok := evaluated.(*object.String)

	if !ok {
		t.Fatalf("object is not String. got=%T (%+v)", evaluated, evaluated)
	}

	if str.Value != "Hello World!" {
		t.Errorf("String has wrong value. got=%q, expected=Hello World!", str.Value)
	}
}

func TestListLiterals(t *testing.T) {
	input := "[1, 2 * 2, 3 + 3]"

	evaluated := testEval(input)
	result, ok := evaluated.(*object.List)

	if !ok {
		t.Fatalf("object is not List. got=%T (%+v)", evaluated, evaluated)
	}

	if len(result.Elements) != 3 {
		t.Fatalf("list has wrong number of elements. got=%d, expected=3", len(result.Elements))
	}

	testNumberObject(t, result.Elements[0], 1)
	testNumberObject(t, result.Elements[1], 4)
	testNumberObject(t, result.Elements[2], 6)
}

func TestListIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			"[1, 2, 3][0]",
			1,
		},
		{
			"[1, 2, 3][1]",
			2,
		},
		{
			"[1, 2, 3][2]",
			3,
		},
		{
			"i := 0; [1][i];",
			1,
		},
		{
			"[1, 2, 3][1 + 1]",
			3,
		},
		{
			"myList := [1, 2, 3]; myList[2];",
			3,
		},
		{
			"myList := [1, 2, 3]; myList[0] + myList[1] + myList[2];",
			6,
		},
		{
			"myList := [1, 2, 3]; i := myList[0]; myList[i]",
			2,
		},
		{
			"[1, 2, 3][3]",
			nil,
		},
		{
			"[1, 2, 3][-1]",
			nil,
		},
		{
			"myList := []; myList[0] := 5; myList[0]",
			5,
		},
		{
			"grid := []; grid[0] := []; grid[0][0] := 10; grid[0][0]",
			10,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(int)

		if ok {
			testNumberObject(t, evaluated, int64(number))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestMapLiterals(t *testing.T) {
	input := `two := "two";
	{
		"one": 10 - 9,
		two: 1 + 1,
		"thr" + "ee": 6 / 2,
		4: 4,
		true: 5,
		false: 6
	}`

	evaluated := testEval(input)
	result, ok := evaluated.(*object.Map)

	if !ok {
		t.Fatalf("object is not Map. got=%T (%+v)", evaluated, evaluated)
	}

	expected := map[object.MapKey]int64{
		(&object.String{Value: "one"}).MapKey():             1,
		(&object.String{Value: "two"}).MapKey():             2,
		(&object.String{Value: "three"}).MapKey():           3,
		(&object.Number{Value: decimal.New(4, 0)}).MapKey(): 4,
		value.TRUE.MapKey():                                 5,
		value.FALSE.MapKey():                                6,
	}

	if len(result.Pairs) != len(expected) {
		t.Fatalf("map has wrong number of pairs. got=%d", len(result.Pairs))
	}

	for expectedKey, expectedValue := range expected {
		pair, ok := result.Pairs[expectedKey]

		if !ok {
			t.Errorf("no pair for given key in Pairs")
		}

		testNumberObject(t, pair.Value, expectedValue)
	}
}

func TestMapIndexExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}["foo"]`,
			5,
		},
		{
			`{"foo": 5}["bar"]`,
			nil,
		},
		{
			`key := "foo"; {"foo": 5}[key]`,
			5,
		},
		{
			`{}["foo"]`,
			nil,
		},
		{
			`{5: 5}[5]`,
			5,
		},
		{
			`{true: 5}[true]`,
			5,
		},
		{
			`{false: 5}[false]`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		integer, ok := tt.expected.(int)

		if ok {
			testNumberObject(t, evaluated, int64(integer))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestMapDotNotationExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`{"foo": 5}.foo`,
			5,
		},
		{
			`{"foo": 5}.bar`,
			nil,
		},
		{
			`{}.foo`,
			nil,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(int)

		if ok {
			testNumberObject(t, evaluated, int64(number))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestWhileExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{"while (false) { }", nil},
		{"n := 0; while (n < 10) { n = n + 1 }; n", 10},
		{"n := 10; while (n > 0) { n = n - 1 }; n", 0},
		{"n := 0; while (n < 10) { n = n + 1 }", nil},
		{"n := 10; while (n > 0) { n = n - 1 }", nil},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(int)

		if ok {
			testNumberObject(t, evaluated, int64(number))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestForExpressions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`x := 1; for (x := 0; x < 10; x := x + 1) { x }; x;`, 1},
		{`for (i := 0; i < 10; i := i + 1) { i };`, nil},
		{`y := []; for (x in 1 .. 10) { push(y, x) }; length(y)`, 10},
		{`y := []; x := 100 for (x in 1 .. 10) { x := x + 1 }; x`, 100},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(int)

		if ok {
			testNumberObject(t, evaluated, int64(number))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestBuiltinFunctions(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`length("")`, 0},
		{`length("four")`, 4},
		{`length("hello world")`, 11},
		{`length(1)`, "argument to `length` not supported, got NUMBER"},
		{`length("one", "two")`, "wrong number of arguments. got=2, expected=1"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int64:
			testNumberObject(t, evaluated, expected)
		case string:
			errObj, ok := evaluated.(*object.Error)

			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}

func TestMathModule(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`Math.abs(123)`, 123},
		{`Math.abs(-123)`, 123},
		{`Math.abs("foo")`, "argument to `Math.abs` must be NUMBER, got STRING"},
		{`Math.abs()`, "wrong number of arguments. got=0, expected=1"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)

		switch expected := tt.expected.(type) {
		case int64:
			testNumberObject(t, evaluated, expected)
		case string:
			errObj, ok := evaluated.(*object.Error)

			if !ok {
				t.Errorf("object is not Error. got=%T (%+v)", evaluated, evaluated)
				continue
			}

			if errObj.Message != expected {
				t.Errorf("wrong error message. expected=%q, got=%q", expected, errObj.Message)
			}
		}
	}
}

func TestImportExpression(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`module := import("../stubs/module"); module.A`,
			5,
		},
		// {
		// 	`module := import("../stubs/module"); module.Sum(2, 3)`,
		// 	5,
		// },
		// {
		// 	`module := import("../stubs/module"); module.a`,
		// 	nil,
		// },
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, ok := tt.expected.(int)

		if ok {
			testNumberObject(t, evaluated, int64(number))
		} else {
			testNullObject(t, evaluated)
		}
	}
}

func TestImportSearchPaths(t *testing.T) {
	utilities.AddPath("../stubs")

	tests := []struct {
		input    string
		expected interface{}
	}{
		{
			`module := import("../stubs/module"); module.A`,
			5,
		},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		number, _ := tt.expected.(int)

		testNumberObject(t, evaluated, int64(number))
	}
}

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		input           string
		expectedMessage string
	}{
		{"5 + true;", "[1] Type mismatch: NUMBER + BOOLEAN"},
		{"5 + true; 5;", "[1] Type mismatch: NUMBER + BOOLEAN"},
		{"-true", "[1] Unknown operator: -BOOLEAN"},
		{"true + false;", "[1] Unknown operator: BOOLEAN + BOOLEAN"},
		{"5; true + false; 5", "[1] Unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { true + false; }", "[1] Unknown operator: BOOLEAN + BOOLEAN"},
		{"if (10 > 1) { if (10 > 1) { return true + false; } return 1; }", "[1] Unknown operator: BOOLEAN + BOOLEAN"},
		{"foobar", "[1] Identifier not found: foobar"},
		{`"Hello" - "World"`, "[1] Unknown operator: STRING - STRING"},
		{`{"name": "Ghost"}[function(x) { x }]`, "[1] Unusable as map key: FUNCTION"},
	}

	for _, tt := range tests {
		evaluated := testEval(tt.input)
		errorObject, ok := evaluated.(*object.Error)

		if !ok {
			t.Errorf("no error object returned. got=%T (%+v)", evaluated, evaluated)
			continue
		}

		if errorObject.Message != tt.expectedMessage {
			t.Errorf("wrong error message. expected=%q, got=%q", tt.expectedMessage, errorObject.Message)
		}
	}
}

func testEval(input string) object.Object {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	env := object.NewEnvironment()

	return Eval(program, env)
}

func testNumberObject(t *testing.T, obj object.Object, expected int64) bool {
	result, ok := obj.(*object.Number)

	if !ok {
		t.Errorf("object is not Number. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value.IntPart() != expected {
		t.Errorf("object has wrong value. got=%d, expected=%d", result.Value.IntPart(), expected)
		return false
	}

	return true
}

func testBooleanObject(t *testing.T, obj object.Object, expected bool) bool {
	result, ok := obj.(*object.Boolean)

	if !ok {
		t.Errorf("object is not Boolean. got=%T (%+v)", obj, obj)
		return false
	}

	if result.Value != expected {
		t.Errorf("object has wrong value. got=%t, expected=%t", result.Value, expected)
		return false
	}

	return true
}

func testNullObject(t *testing.T, obj object.Object) bool {
	if obj != value.NULL {
		t.Errorf("object is not NULL. got=%T (%+v)", obj, obj)
		return false
	}

	return true
}
