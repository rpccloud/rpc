package internal

import (
	"fmt"
	"reflect"
)

type Assert interface {
	Fail(reason string)
	Equals(args ...interface{})
	IsNil()
	IsNotNil()
	IsTrue()
	IsFalse()
}

type testReporter interface {
	Fail()
}

// NewAssert create new assert class
func NewAssert(t testReporter) func(args ...interface{}) Assert {
	return func(args ...interface{}) Assert {
		return &rpcAssert{
			t:    t,
			args: args,
		}
	}
}

// rpcAssert ...
type rpcAssert struct {
	t    testReporter
	args []interface{}
}

func (p *rpcAssert) fail(reason string) {
	fmt.Println("\t", GetCodePosition(reason, 2))
	p.t.Fail()
}

// Fail ...
func (p *rpcAssert) Fail(reason string) {
	p.fail(reason)
}

// Equals ...
func (p *rpcAssert) Equals(args ...interface{}) {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else if len(p.args) != len(args) {
		p.fail("arguments length not match")
	} else {
		for i := 0; i < len(p.args); i++ {
			if !reflect.DeepEqual(p.args[i], args[i]) {
				p.fail(fmt.Sprintf(
					"%s argment is not equal, want %v, got %v",
					ConvertOrdinalToString(uint(i+1)),
					args[i],
					p.args[i],
				))
			}
		}
	}
}

// IsNil ...
func (p *rpcAssert) IsNil() {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else {
		for i := 0; i < len(p.args); i++ {
			if !isNil(p.args[i]) {
				p.fail(fmt.Sprintf(
					"%s argment is not nil",
					ConvertOrdinalToString(uint(i+1)),
				))
			}
		}
	}
}

// IsNotNil ...
func (p *rpcAssert) IsNotNil() {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else {
		for i := 0; i < len(p.args); i++ {
			if isNil(p.args[i]) {
				p.fail(fmt.Sprintf(
					"%s argment is nil",
					ConvertOrdinalToString(uint(i+1)),
				))
			}
		}
	}
}

// IsTrue ...
func (p *rpcAssert) IsTrue() {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else {
		for i := 0; i < len(p.args); i++ {
			if p.args[i] != true {
				p.fail(fmt.Sprintf(
					"%s argment is not true",
					ConvertOrdinalToString(uint(i+1)),
				))
			}
		}
	}
}

// IsFalse ...
func (p *rpcAssert) IsFalse() {
	if len(p.args) < 1 {
		p.fail("arguments is empty")
	} else {
		for i := 0; i < len(p.args); i++ {
			if p.args[i] != false {
				p.fail(fmt.Sprintf(
					"%s argment is not false",
					ConvertOrdinalToString(uint(i+1)),
				))
			}
		}
	}
}
