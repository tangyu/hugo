// Copyright 2017 The Hugo Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package os

import (
	"errors"
	"fmt"
	_os "os"

	"github.com/gohugoio/hugo/deps"
	"github.com/spf13/afero"
	"github.com/spf13/cast"
)

// New returns a new instance of the os-namespaced template functions.
func New(deps *deps.Deps) *Namespace {

	// Since Hugo 0.38 we can have multiple content dirs. This can make it hard to
	// reason about where the file is placed relative to the project root.
	// To make the {{ readFile .Filename }} variant just work, we create a composite
	// filesystem that first checks the work dir fs and then the content fs.
	var rfs afero.Fs
	if deps.Fs != nil {
		rfs = deps.Fs.WorkingDir
		if deps.PathSpec != nil && deps.PathSpec.BaseFs != nil {
			rfs = afero.NewReadOnlyFs(afero.NewCopyOnWriteFs(deps.PathSpec.BaseFs.ContentFs, deps.Fs.WorkingDir))
		}
	}

	return &Namespace{
		readFileFs: rfs,
		deps:       deps,
	}
}

// Namespace provides template functions for the "os" namespace.
type Namespace struct {
	readFileFs afero.Fs
	deps       *deps.Deps
}

// Getenv retrieves the value of the environment variable named by the key.
// It returns the value, which will be empty if the variable is not present.
func (ns *Namespace) Getenv(key interface{}) (string, error) {
	skey, err := cast.ToStringE(key)
	if err != nil {
		return "", nil
	}

	return _os.Getenv(skey), nil
}

// readFile reads the file named by filename in the given filesystem
// and returns the contents as a string.
// There is a upper size limit set at 1 megabytes.
func readFile(fs afero.Fs, filename string) (string, error) {
	if filename == "" {
		return "", errors.New("readFile needs a filename")
	}

	if info, err := fs.Stat(filename); err == nil {
		if info.Size() > 1000000 {
			return "", fmt.Errorf("File %q is too big", filename)
		}
	} else {
		return "", err
	}
	b, err := afero.ReadFile(fs, filename)

	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ReadFile reads the file named by filename relative to the configured WorkingDir.
// It returns the contents as a string.
// There is an upper size limit set at 1 megabytes.
func (ns *Namespace) ReadFile(i interface{}) (string, error) {
	s, err := cast.ToStringE(i)
	if err != nil {
		return "", err
	}

	return readFile(ns.readFileFs, s)
}

// ReadDir lists the directory contents relative to the configured WorkingDir.
func (ns *Namespace) ReadDir(i interface{}) ([]_os.FileInfo, error) {
	path, err := cast.ToStringE(i)
	if err != nil {
		return nil, err
	}

	list, err := afero.ReadDir(ns.deps.Fs.WorkingDir, path)
	if err != nil {
		return nil, fmt.Errorf("Failed to read Directory %s with error message %s", path, err)
	}

	return list, nil
}

// FileExists checks whether a file exists under the given path.
func (ns *Namespace) FileExists(i interface{}) (bool, error) {
	path, err := cast.ToStringE(i)
	if err != nil {
		return false, err
	}

	if path == "" {
		return false, errors.New("fileExists needs a path to a file")
	}

	status, err := afero.Exists(ns.readFileFs, path)
	if err != nil {
		return false, err
	}

	return status, nil
}
