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

package pcreate

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
	Operation = "create"
)

//PoolCreate defines structure for pool 'Create' operation
type PoolCreate struct {
	// property list
	Property []string
	// file system property list
	FSProperty []string
	// pool name
	Pool string

	// pool type
	Type string

	//vdev list
	Vdev []string

	// force use of vdevs
	Forcefully bool

	// command string
	Command string

	// checks is list of predicate function used for validating object
	checks []PredicateFunc

	// Executor is to execute the commands
	Executor bin.Executor

	// error
	err error
}

// NewPoolCreate returns new instance of object PoolCreate
func NewPoolCreate() *PoolCreate {
	return &PoolCreate{}
}

// WithCheck add given check to checks list
func (p *PoolCreate) WithCheck(check ...PredicateFunc) *PoolCreate {
	p.checks = append(p.checks, check...)
	return p
}

// WithProperty method fills the Property field of PoolCreate object.
func (p *PoolCreate) WithProperty(key, value string) *PoolCreate {
	if value != "" {
		p.Property = append(p.Property, fmt.Sprintf("%s=%s", key, value))
	}
	return p
}

// WithFSProperty method fills the file system Property field of PoolCreate object.
func (p *PoolCreate) WithFSProperty(key, value string) *PoolCreate {
	if value != "" {
		p.FSProperty = append(p.FSProperty, fmt.Sprintf("%s=%s", key, value))
	}
	return p
}

// WithPool method fills the Pool field of PoolCreate object.
func (p *PoolCreate) WithPool(Pool string) *PoolCreate {
	p.Pool = Pool
	return p
}

// WithType method fills the Type field of PoolCreate object.
func (p *PoolCreate) WithType(t string) *PoolCreate {
	p.Type = t
	return p
}

// WithVdev method fills the Vdev field of PoolCreate object.
func (p *PoolCreate) WithVdev(vdev string) *PoolCreate {
	p.Vdev = append(p.Vdev, vdev)
	return p
}

// WithVdevList method fills the Vdev field of PoolCreate object.
func (p *PoolCreate) WithVdevList(vdevlist []string) *PoolCreate {
	p.Vdev = append(p.Vdev, vdevlist...)
	return p
}

// WithForcefully method fills the Forcefully field of PoolCreate object.
func (p *PoolCreate) WithForcefully(Forcefully bool) *PoolCreate {
	p.Forcefully = Forcefully
	return p
}

// WithCommand method fills the Command field of PoolCreate object.
func (p *PoolCreate) WithCommand(Command string) *PoolCreate {
	p.Command = Command
	return p
}

// WithExecutor method fills the Executor field of PoolCreate object.
func (p *PoolCreate) WithExecutor(executor bin.Executor) *PoolCreate {
	p.Executor = executor
	return p
}

// Validate is to validate generated PoolCreate object by builder
func (p *PoolCreate) Validate() *PoolCreate {
	for _, check := range p.checks {
		if !check(p) {
			p.err = errors.Wrapf(p.err, "validation failed {%v}", runtime.FuncForPC(reflect.ValueOf(check).Pointer()).Name())
		}
	}
	return p
}

// Execute is to execute generated PoolCreate object
func (p *PoolCreate) Execute() ([]byte, error) {
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

// Build returns the PoolCreate object generated by builder
func (p *PoolCreate) Build() (*PoolCreate, error) {
	var c strings.Builder
	p = p.Validate()
	p.appendCommand(&c, bin.ZPOOL)
	p.appendCommand(&c, fmt.Sprintf(" %s ", Operation))

	if IsForcefullySet()(p) {
		p.appendCommand(&c, fmt.Sprintf(" -f "))
	}

	if IsPropertySet()(p) {
		for _, v := range p.Property {
			p.appendCommand(&c, fmt.Sprintf(" -o %s ", v))
		}
	}

	if IsFSPropertySet()(p) {
		for _, v := range p.FSProperty {
			p.appendCommand(&c, fmt.Sprintf(" -O %s ", v))
		}
	}

	p.appendCommand(&c, p.Pool)

	if IsTypeSet()(p) {
		p.appendCommand(&c, fmt.Sprintf(" %s ", p.Type))
	}

	for _, v := range p.Vdev {
		p.appendCommand(&c, fmt.Sprintf(" %s ", v))
	}

	p.Command = c.String()
	return p, p.err
}

// appendCommand append string to given string builder
func (p *PoolCreate) appendCommand(c *strings.Builder, cmd string) {
	_, err := c.WriteString(cmd)
	if err != nil {
		p.err = errors.Wrapf(p.err, "Failed to append cmd{%s} : %s", cmd, err.Error())
	}
}
