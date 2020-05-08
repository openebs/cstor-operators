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

package pdestroy

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
	Operation = "destroy"
)

//PoolDestroy defines structure for pool 'Destroy' operation
type PoolDestroy struct {
	//pool name
	Pool string

	//forcefully destroy
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

// NewPoolDestroy returns new instance of object PoolDestroy
func NewPoolDestroy() *PoolDestroy {
	return &PoolDestroy{}
}

// WithCheck add given check to checks list
func (p *PoolDestroy) WithCheck(check ...PredicateFunc) *PoolDestroy {
	p.checks = append(p.checks, check...)
	return p
}

// WithPool method fills the Pool field of PoolDestroy object.
func (p *PoolDestroy) WithPool(Pool string) *PoolDestroy {
	p.Pool = Pool
	return p
}

// WithForcefully method fills the Forcefully field of PoolDestroy object.
func (p *PoolDestroy) WithForcefully(Forcefully bool) *PoolDestroy {
	p.Forcefully = Forcefully
	return p
}

// WithCommand method fills the Command field of PoolDestroy object.
func (p *PoolDestroy) WithCommand(Command string) *PoolDestroy {
	p.Command = Command
	return p
}

// WithExecutor method fills the Executor field of PoolDestroy object.
func (p *PoolDestroy) WithExecutor(executor bin.Executor) *PoolDestroy {
	p.Executor = executor
	return p
}

// Validate is to validate generated PoolDestroy object by builder
func (p *PoolDestroy) Validate() *PoolDestroy {
	for _, check := range p.checks {
		if !check(p) {
			p.err = errors.Wrapf(p.err, "validation failed {%v}", runtime.FuncForPC(reflect.ValueOf(check).Pointer()).Name())
		}
	}
	return p
}

// Execute is to execute generated PoolCreate object
func (p *PoolDestroy) Execute() ([]byte, error) {
	var ret []byte
	var err error
	p, err = p.Build()
	if err != nil {
		return nil, err
	}
	if IsExecutorSet()(p) {
		ret, err = p.Executor.Execute(p.Command)
	} else {
		// execute command here
		// #nosec
		ret, err = exec.Command(bin.BASH, "-c", p.Command).CombinedOutput()
	}
	if err != nil {
		return ret, errors.Errorf("error %s %s", err.Error(), string(ret))
	}
	return ret, nil
}

// Build returns the PoolDestroy object generated by builder
func (p *PoolDestroy) Build() (*PoolDestroy, error) {
	var c strings.Builder
	p = p.Validate()
	p.appendCommand(&c, bin.ZPOOL)
	p.appendCommand(&c, fmt.Sprintf(" %s ", Operation))

	if IsForcefullySet()(p) {
		p.appendCommand(&c, fmt.Sprintf(" -f "))
	}

	p.appendCommand(&c, p.Pool)

	p.Command = c.String()
	return p, p.err
}

// appendCommand append string to given string builder
func (p *PoolDestroy) appendCommand(c *strings.Builder, cmd string) {
	_, err := c.WriteString(cmd)
	if err != nil {
		p.err = errors.Wrapf(p.err, "Failed to append cmd{%s} : %s", cmd, err.Error())
	}
}
