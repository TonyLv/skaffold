/*
Copyright 2018 Google LLC

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

package packer

import (
	"path/filepath"
	"strings"

	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha2"
	"github.com/sirupsen/logrus"
)

type PackerDependencyResolver struct{}

const sourceQuery = "kind('source file', deps('%s'))"

func (*PackerDependencyResolver) GetDependencies(a *v1alpha2.Artifact) ([]string, error) {
	// Packer has a full variable system with includes, which can reference
	// either local or absolute paths. Therefore the files to watch must be
	// explicitly provided for Packer builds.

	var deps []string
	for _, l := range a.PackerArtifact.Files {
		dep := ""
		// absolute path
		if strings.HasPrefix(l, "/") {
			dep = l
		// relative path
		} else {
			dep = filepath.Join(a.Workspace, l)
		}
		deps = append(deps, dep)
	}

	logrus.Debugf("Found files to watch: %s", deps)
	return deps, nil
}
