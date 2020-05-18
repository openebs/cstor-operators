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

package vget

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

//VolumeGetProperty defines structure for volume 'Property' operation
type VolumeGetProperty struct {
	//list of property
	Proplist []string

	//dataset name
	Dataset string

	// scripted mode. Will not display headers, and separate fields by a single tab instead of arbitrary space
	IsScriptedMode bool

	// Display numbers in parsable (exact) values
	IsParsableMode bool

	// Field list to display
	FieldList []string

	//snapshot name
	Snapshot string

	//command for this structure
	Command string

	// checks is list of predicate function used for validating object
	checks []PredicateFunc

	// Executor is to execute the commands
	Executor bin.Executor

	// error
	err error
}

// NewVolumeGetProperty returns new instance of object VolumeGetProperty
func NewVolumeGetProperty() *VolumeGetProperty {
	return &VolumeGetProperty{}
}

// WithCheck add given check to checks list
func (v *VolumeGetProperty) WithCheck(check ...PredicateFunc) *VolumeGetProperty {
	v.checks = append(v.checks, check...)
	return v
}

// WithProperty method fills the Proplist field of VolumeGetProperty object.
func (v *VolumeGetProperty) WithProperty(key string) *VolumeGetProperty {
	v.Proplist = append(v.Proplist, key)
	return v
}

// WithScriptedMode method update the IsScriptedMode field of VolumeGetProperty object.
func (v *VolumeGetProperty) WithScriptedMode(IsScriptedMode bool) *VolumeGetProperty {
	v.IsScriptedMode = IsScriptedMode
	return v
}

// WithParsableMode method update the IsParsableMode field of VolumeGetProperty object.
func (v *VolumeGetProperty) WithParsableMode(IsParsableMode bool) *VolumeGetProperty {
	v.IsParsableMode = IsParsableMode
	return v
}

// WithField method fills the FieldList field of VolumeGetProperty object.
func (v *VolumeGetProperty) WithField(Field string) *VolumeGetProperty {
	v.FieldList = append(v.FieldList, Field)
	return v
}

// WithPropertyList method fills the PropList field of PoolGProperty object.
func (v *VolumeGetProperty) WithPropertyList(keys []string) *VolumeGetProperty {
	v.Proplist = append(v.Proplist, keys...)
	return v
}

// WithDataset method fills the Dataset field of VolumeGetProperty object.
func (v *VolumeGetProperty) WithDataset(Dataset string) *VolumeGetProperty {
	v.Dataset = Dataset
	return v
}

// WithSnapshot method fills the Snapshot field of VolumeGetProperty object.
func (v *VolumeGetProperty) WithSnapshot(Snapshot string) *VolumeGetProperty {
	v.Snapshot = Snapshot
	return v
}

// WithCommand method fills the Command field of VolumeGetProperty object.
func (v *VolumeGetProperty) WithCommand(Command string) *VolumeGetProperty {
	v.Command = Command
	return v
}

// WithExecutor method fills the Executor field of VolumeGetProperty object.
func (v *VolumeGetProperty) WithExecutor(executor bin.Executor) *VolumeGetProperty {
	v.Executor = executor
	return v
}

// Validate is to validate generated VolumeGetProperty object by builder
func (v *VolumeGetProperty) Validate() *VolumeGetProperty {
	for _, check := range v.checks {
		if !check(v) {
			v.err = errors.Wrapf(v.err, "validation failed {%v}", runtime.FuncForPC(reflect.ValueOf(check).Pointer()).Name())
		}
	}
	return v
}

// Execute is to execute generated VolumeGetProperty object
func (v *VolumeGetProperty) Execute() ([]byte, error) {
	var ret []byte
	var err error
	v, err = v.Build()
	if err != nil {
		return nil, err
	}

	if IsExecutorSet()(v) {
		return v.Executor.Execute(v.Command)
	}

	// execute command here
	// #nosec
	ret, err = exec.Command(bin.BASH, "-c", v.Command).CombinedOutput()
	if err != nil {
		return nil, errors.Wrapf(err, "%s", ret)
	}
	return ret, nil
}

// Build returns the VolumeGetProperty object generated by builder
func (v *VolumeGetProperty) Build() (*VolumeGetProperty, error) {
	var c strings.Builder

	v = v.Validate()
	v.appendCommand(&c, bin.ZFS)

	v.appendCommand(&c, fmt.Sprintf(" %s ", Operation))

	if IsScriptedModeSet()(v) {
		v.appendCommand(&c, fmt.Sprintf(" -H "))
	}

	if IsParsableModeSet()(v) {
		v.appendCommand(&c, fmt.Sprintf(" -p "))
	}

	if IsFieldListSet()(v) {
		v.appendCommand(&c, fmt.Sprintf(" -o "))
		for _, f := range v.FieldList {
			v.appendCommand(&c, fmt.Sprintf("%s,", f))
		}
		// add space to separate next arg
		v.appendCommand(&c, fmt.Sprintf(" "))
	}

	if IsProplistSet()(v) {
		for _, p := range v.Proplist {
			v.appendCommand(&c, fmt.Sprintf("%s,", p))
		}
		// add space to separate next arg
		v.appendCommand(&c, fmt.Sprintf(" "))
	}

	v.appendCommand(&c, v.Dataset)

	if IsSnapshotSet()(v) {
		v.appendCommand(&c, fmt.Sprintf("@%s", v.Snapshot))
	}

	v.Command = c.String()
	return v, v.err
}

// appendCommand append string to given string builder
func (v *VolumeGetProperty) appendCommand(c *strings.Builder, cmd string) {
	_, err := c.WriteString(cmd)
	if err != nil {
		v.err = errors.Wrapf(v.err, "Failed to append cmd{%s} : %s", cmd, err.Error())
	}
}
