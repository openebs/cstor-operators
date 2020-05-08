/*
Copyright 2019 The OpenEBS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pget

import (
	"fmt"
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	"github.com/openebs/cstor-operators/pkg/zcmd/bin"
	"github.com/pkg/errors"
)

const (
	// Operation defines type of zfs operation
	Operation = "get"
)

//PoolGProperty defines structure for pool 'get Property' operation
type PoolGProperty struct {
	//list of property
	PropList []string

	// scripted mode. Will not display headers, and separate fields by a single tab instead of arbitrary space
	IsScriptedMode bool

	// Display numbers in parsable (exact) values
	IsParsableMode bool

	// Field list to display
	FieldList []string

	//pool name
	Pool string

	//command string
	Command string

	// checks is list of predicate function used for validating object
	checks []PredicateFunc

	// error
	err error

	// Executor is to execute the commands
	Executor bin.Executor
}

// NewPoolGetProperty returns new instance of object PoolGProperty
func NewPoolGetProperty() *PoolGProperty {
	return &PoolGProperty{}
}

// WithCheck add given check to checks list
func (p *PoolGProperty) WithCheck(check ...PredicateFunc) *PoolGProperty {
	p.checks = append(p.checks, check...)
	return p
}

// WithProperty method fills the PropList field of PoolGProperty object.
func (p *PoolGProperty) WithProperty(key string) *PoolGProperty {
	p.PropList = append(p.PropList, key)
	return p
}

// WithPropertyList method fills the PropList field of PoolGProperty object.
func (p *PoolGProperty) WithPropertyList(keys []string) *PoolGProperty {
	p.PropList = append(p.PropList, keys...)
	return p
}

// WithScriptedMode method update the IsScriptedMode field of PoolGProperty object.
func (p *PoolGProperty) WithScriptedMode(IsScriptedMode bool) *PoolGProperty {
	p.IsScriptedMode = IsScriptedMode
	return p
}

// WithParsableMode method update the IsParsableMode field of PoolGProperty object.
func (p *PoolGProperty) WithParsableMode(IsParsableMode bool) *PoolGProperty {
	p.IsParsableMode = IsParsableMode
	return p
}

// WithField method fills the FieldList field of PoolGProperty object.
func (p *PoolGProperty) WithField(Field string) *PoolGProperty {
	p.FieldList = append(p.FieldList, Field)
	return p
}

// WithPool method fills the Pool field of PoolGProperty object.
func (p *PoolGProperty) WithPool(Pool string) *PoolGProperty {
	p.Pool = Pool
	return p
}

// WithCommand method fills the Command field of PoolGProperty object.
func (p *PoolGProperty) WithCommand(Command string) *PoolGProperty {
	p.Command = Command
	return p
}

// WithExecutor method fills the Executor field of PoolDump object.
func (p *PoolGProperty) WithExecutor(executor bin.Executor) *PoolGProperty {
	p.Executor = executor
	return p
}

// Validate is to validate generated PoolGProperty object by builder
func (p *PoolGProperty) Validate() *PoolGProperty {
	for _, check := range p.checks {
		if !check(p) {
			p.err = errors.Wrapf(p.err, "validation failed {%v}", runtime.FuncForPC(reflect.ValueOf(check).Pointer()).Name())
		}
	}
	return p
}

// Execute is to execute generated PoolGProperty object
func (p *PoolGProperty) Execute() ([]byte, error) {
	p, err := p.Build()
	if err != nil {
		return nil, err
	}

	if IsExecutorSet()(p) {
		return p.Executor.Execute(p.Command)
	}

	// execute command here
	// #nosec
	return exec.Command(bin.BASH, "-c", p.Command).CombinedOutput()
}

// Build returns the PoolGProperty object generated by builder
func (p *PoolGProperty) Build() (*PoolGProperty, error) {
	var c strings.Builder
	p = p.Validate()
	p.appendCommand(&c, bin.ZPOOL)

	p.appendCommand(&c, fmt.Sprintf(" %s ", Operation))

	if IsScriptedModeSet()(p) {
		p.appendCommand(&c, fmt.Sprintf(" -H "))
	}

	if IsParsableModeSet()(p) {
		p.appendCommand(&c, fmt.Sprintf(" -p "))
	}

	if IsFieldListSet()(p) {
		p.appendCommand(&c, fmt.Sprintf(" -o "))
		for _, f := range p.FieldList {
			p.appendCommand(&c, fmt.Sprintf("%s,", f))
		}
		// add space to separate next arg
		p.appendCommand(&c, fmt.Sprintf(" "))
	}

	if IsPropListSet()(p) {
		for _, v := range p.PropList {
			p.appendCommand(&c, fmt.Sprintf("%s,", v))
		}
		// add space to separate next arg
		p.appendCommand(&c, fmt.Sprintf(" "))
	}

	if IsPoolSet()(p) {
		p.appendCommand(&c, fmt.Sprintf(" %s ", p.Pool))
	}

	p.Command = c.String()
	return p, p.err
}

// appendCommand append string to given string builder
func (p *PoolGProperty) appendCommand(c *strings.Builder, cmd string) {
	_, err := c.WriteString(cmd)
	if err != nil {
		p.err = errors.Wrapf(p.err, "Failed to append cmd{%s} : %s", cmd, err.Error())
	}
}
