package math

// Used functions and types
func Add(a, b int) int {
	return a + b
}

func Multiply(a, b int) int {
	return a * b
}

type Calculator struct {
	Result int
}

func (c *Calculator) Add(x int) {
	c.Result += x
}

func (c *Calculator) GetResult() int {
	return c.Result
}

// UNUSED functions and types - these should be eliminated
func Subtract(a, b int) int {
	return a - b
}

func Divide(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}

type UnusedStruct struct {
	Value int
	Name  string
}

func (u *UnusedStruct) DoSomething() {
	u.Value = 42
}

func (u *UnusedStruct) GetValue() int {
	return u.Value
}

type AnotherUnusedStruct struct {
	Data []byte
}

func NewUnusedStruct() *UnusedStruct {
	return &UnusedStruct{Value: 0, Name: "unused"}
}

func UnusedGlobalFunction() string {
	return "this should not appear in output"
}

const UnusedConstant = 999

var UnusedVariable = "unused variable"