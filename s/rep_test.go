package s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	var env = NewEnv()
	env.Init()

	var evalTests = []struct {
		root     Item
		expected Item
	}{
		{ // 1 => 1
			root:     Integer{Value: 1},
			expected: Integer{Value: 1},
		},
		{ // "1" => "1"
			root:     String{Value: "1"},
			expected: String{Value: "1"},
		},
		{ // :key => :key
			root:     Keyword{Value: "key"},
			expected: Keyword{Value: "key"},
		},
		{ // + => Func
			root: Symbol{Value: "+"},
			expected: func() Item {
				i, _ := env.Get("+")
				return i
			}(),
		},
		{ // () => ()
			root:     List{},
			expected: List{},
		},
		{ // (+ 1 1) => 2
			root: List{Value: []Item{
				Symbol{Value: "+"},
				Integer{Value: 1},
				Integer{Value: 1},
			}},
			expected: Integer{Value: 2},
		},
		{ // ((fn [a b] (+ a b)) 1 1) => 2
			root: List{Value: []Item{
				List{Value: []Item{
					Symbol{Value: "fn"},
					Vector{Value: []Item{
						Symbol{Value: "a"},
						Symbol{Value: "b"},
					}},
					List{Value: []Item{
						Symbol{Value: "+"},
						Symbol{Value: "a"},
						Symbol{Value: "b"},
					}},
				}},
				Integer{Value: 1},
				Integer{Value: 1},
			}},
			expected: Integer{Value: 2},
		},
	}

	for _, test := range evalTests {
		actual, err := Eval(test.root, env)

		assert.NoError(t, err)
		assert.Equal(t, test.expected, actual)
	}
}

var repTestcases = map[string]string{
	"(+ 1 2)":                            "3",
	"(+ 5 (* 2 3))":                      "11",
	"(- (+ 5 (* 2 3)) 3)":                "8",
	"(/ (- (+ 5 (* 2 3)) 3) 4)":          "2",
	"(/ (- (+ 515 (* 222 311)) 302) 27)": "2565",
}

func TestRep(t *testing.T) {
	for exp, out := range repTestcases {
		result, err := Rep(exp)

		assert.NoError(t, err)
		assert.Equal(t, out, result)
	}
}

func TestRep_Set(t *testing.T) {
	res1, err1 := Rep(`(set x 2)`)
	assert.NoError(t, err1)
	assert.Equal(t, "2", res1)

	res2, err2 := Rep("x")
	assert.NoError(t, err2)
	assert.Equal(t, "2", res2)

	res3, err3 := Rep("(+ 2 x)")
	assert.NoError(t, err3)
	assert.Equal(t, "4", res3)

	res4, err4 := Rep("(set y (+ 1 7))")
	assert.NoError(t, err4)
	assert.Equal(t, "8", res4)
}

func TestRep_Let(t *testing.T) {
	res1, err1 := Rep(`(let {z 9} z)`)
	assert.NoError(t, err1)
	assert.Equal(t, "9", res1)

	res2, err2 := Rep(`(let {z (+ 2 3)} (+ 1 z))`)
	assert.NoError(t, err2)
	assert.Equal(t, "6", res2)

	res3, err3 := Rep(`(let {p (+ 2 3) q (+ 2 p)} (+ p q))`)
	assert.NoError(t, err3)
	assert.Equal(t, "12", res3)
}

func TestRep_Outer(t *testing.T) {
	res1, err1 := Rep(`(set a 4)`)
	assert.NoError(t, err1)
	assert.Equal(t, "4", res1)

	res2, err2 := Rep(`(let {q 9} q)`)
	assert.NoError(t, err2)
	assert.Equal(t, "9", res2)

	res3, err3 := Rep(`(let {q 9} a)`)
	assert.NoError(t, err3)
	assert.Equal(t, "4", res3)

	res4, err4 := Rep(`(let {z 2} (let {q 9} a))`)
	assert.NoError(t, err4)
	assert.Equal(t, "4", res4)

	res5, err5 := Rep(`(let {z a} z)`)
	assert.NoError(t, err5)
	assert.Equal(t, "4", res5)
}

var listCases = map[string]string{
	"(list)":               "()",
	"(list? (list))":       "true",
	"(empty? (list))":      "true",
	"(empty? (list 1))":    "false",
	"(list 1 2 3)":         "(1 2 3)",
	"(count (list 1 2 3))": "3",
	"(count (list))":       "0",
	"(count nil)":          "0",
}

func TestRep_Lists(t *testing.T) {
	for input, output := range listCases {
		res, err := Rep(input)
		assert.NoError(t, err)
		assert.Equal(t, output, res)
	}
}

var ifCases = map[string]string{
	"(if true 7 8)":              "7",
	"(if false 7 8)":             "8",
	"(if true (+ 1 7) (+ 1 8))":  "8",
	"(if false (+ 1 7) (+ 1 8))": "9",
	"(if nil 7 8)":               "8",
	"(if 0 7 8)":                 "7",
	`(if "" 7 8)`:                "7",
	"(if (list) 7 8)":            "7",
	"(if (list 1 2 3) 7 8)":      "7",
	"(if false (+ 1 7))":         "nil",
	"(if true (+ 1 7))":          "8",
}

func TestRep_If(t *testing.T) {
	for input, output := range ifCases {
		res, err := Rep(input)
		assert.NoError(t, err)
		assert.Equal(t, output, res)
	}
}

var condCases = map[string]string{
	"(= 2 1)":                   "false",
	"(= 1 1)":                   "true",
	"(= 1 (+ 1 1))":             "false",
	"(= 2 (+ 1 1))":             "true",
	"(= nil 1)":                 "false",
	"(= nil nil)":               "true",
	"(= 0 0)":                   "true",
	"(= 1 0)":                   "false",
	`(= "" "")`:                 "true",
	`(= "abc" "")`:              "false",
	`(= "" "abc")`:              "false",
	`(= "abc" "def")`:           "false",
	"(= (list) (list))":         "true",
	"(= (list 1 2) (list 1 2))": "true",
	"(= (list 1) (list))":       "false",
	"(= (list) (list 1))":       "false",
	"(= 0 (list))":              "false",
	"(= (list) 0)":              "false",
	`(= (list) "")`:             "false",
	`(= "" (list))`:             "false",

	"(> 2 1)": "true",
	"(> 1 1)": "false",
	"(> 1 2)": "false",

	"(>= 2 1)": "true",
	"(>= 1 1)": "true",
	"(>= 1 2)": "false",

	"(< 2 1)": "false",
	"(< 1 1)": "false",
	"(< 1 2)": "true",

	"(<= 2 1)": "false",
	"(<= 1 1)": "true",
	"(<= 1 2)": "true",
}

func TestRep_Cond(t *testing.T) {
	for input, output := range condCases {
		res, err := Rep(input)
		assert.NoError(t, err)
		assert.Equal(t, output, res, "%s should return %s", input, output)
	}
}

var fnCases = map[string]string{
	"((fn [] 4))":                            "4",
	"((fn [a] (+ 1 a)) 1)":                   "2",
	"((fn [a b] (+ b a)) 3 4)":               "7",
	"( (fn [f x] (f x)) (fn [a] (+ 1 a)) 7)": "8",
}

func TestRep_Func(t *testing.T) {
	for input, output := range fnCases {
		res, err := Rep(input)
		assert.NoError(t, err)
		assert.Equal(t, output, res, "%s should return %s", input, output)
	}
}

func TestRep_Cjojures(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{"( ( (fn [a] (fn [b] (+ a b))) 5) 7)", "12"},

		{"(set gen-plus5 (fn [] (fn [b] (+ 5 b))))", "function"},
		{"(set plus5 (gen-plus5))", "function"},
		{"(plus5 7)", "12"},

		{"(set gen-plusX (fn [x] (fn [b] (+ x b))))", "function"},
		{"(set plus7 (gen-plusX 7))", "function"},
		{"(plus7 8)", "15"},
	}

	for _, c := range cases {
		res, err := Rep(c.input)
		assert.NoError(t, err)
		assert.Equal(t, c.output, res, "%s should return %s", c.input, c.output)
	}
}

func TestRep_Not(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{"(not false)", "true"},
		{"(not true)", "false"},
		{`(not "a")`, "false"},
		{"(not 0)", "false"},
	}

	for _, c := range cases {
		res, err := Rep(c.input)
		assert.NoError(t, err)
		assert.Equal(t, c.output, res, "%s should return %s", c.input, c.output)
	}
}
