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

package plabelclear

import (
	"os/exec"
	"reflect"
	"runtime"
	"strings"

	"github.com/openebs/cstor-operators/pkg/zcmd/bin"
	"github.com/pkg/errors"
)

const (
	// Operation defines type of zfs operation
	Operation = "labelclear"
)

// PoolLabelClear defines structure for pool 'Clear' operation
type PoolLabelClear struct {
	// vdev
	Vdev string

	// Forcefully
	Forcefully bool

	// command string
	Command string

	// checks is list of predicate function used for validating object
	checks []PredicateFunc

	// error
	err error

	// Executor is to execute the commands
	Executor bin.Executor
}

// NewPoolLabelClear returns new instance of object PoolLabelClear
func NewPoolLabelClear() *PoolLabelClear {
	return &PoolLabelClear{}
}

// WithCheck add given check to checks list
func (p *PoolLabelClear) WithCheck(check ...PredicateFunc) *PoolLabelClear {
	p.checks = append(p.checks, check...)
	return p
}

// WithForceFully method fills the ForceFully field of PoolLabelClear object.
func (p *PoolLabelClear) WithForceFully(Forcefully bool) *PoolLabelClear {
	p.Forcefully = Forcefully
	return p
}

// WithVdev method fills the Vdev field of PoolLabelClear object.
func (p *PoolLabelClear) WithVdev(vdev string) *PoolLabelClear {
	p.Vdev = vdev
	return p
}

// WithExecutor method fills the Vdev field of PoolLabelClear object.
func (p *PoolLabelClear) WithExecutor(executor bin.Executor) *PoolLabelClear {
	p.Executor = executor
	return p
}

// WithCommand method fills the Command field of PoolLabelClear object.
func (p *PoolLabelClear) WithCommand(Command string) *PoolLabelClear {
	p.Command = Command
	return p
}

// Validate is to validate generated PoolLabelClear object by builder
func (p *PoolLabelClear) Validate() *PoolLabelClear {
	for _, check := range p.checks {
		if !check(p) {
			p.err = errors.Wrapf(p.err, "validation failed {%v}", runtime.FuncForPC(reflect.ValueOf(check).Pointer()).Name())
		}
	}
	return p
}

// Execute is to execute generated PoolLabelClear object
func (p *PoolLabelClear) Execute() ([]byte, error) {
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

// Build returns the PoolLabelClear object generated by builder
func (p *PoolLabelClear) Build() (*PoolLabelClear, error) {
	var c strings.Builder
	p = p.Validate()
	p.appendCommand(&c, bin.ZPOOL)
	p.appendCommand(&c, " "+Operation+" ")

	if IsForcefullySet()(p) {
		p.appendCommand(&c, " -f ")
	}

	p.appendCommand(&c, " "+p.Vdev+" ")

	p.Command = c.String()
	return p, p.err
}

// appendCommand append string to given string builder
func (p *PoolLabelClear) appendCommand(c *strings.Builder, cmd string) {
	_, err := c.WriteString(cmd)
	if err != nil {
		p.err = errors.Wrapf(p.err, "Failed to append cmd{%s} : %s", cmd, err.Error())
	}
}
