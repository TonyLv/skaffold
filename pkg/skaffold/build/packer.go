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

package build

import (
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"github.com/GoogleContainerTools/skaffold/pkg/skaffold/schema/v1alpha2"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type PackerBuild struct {
	Name string `json:"name"`
	BuilderType string `json:"builder_type"`
	BuildTime int `json:"build_time"`
	Files []string `json:"files"`
	ArtifactID string `json:"artifact_id"`
	PackerRunUUID string `json:"packer_run_uuid"`
}

type PackerManifest struct {
	PackerBuilds []PackerBuild `json:"builds"`
	LastRunUUID string `json:"last_run_uuid"`
}

func (l *LocalBuilder) buildPacker(ctx context.Context, out io.Writer, a *v1alpha2.Artifact) (string, error) {
	logrus.Debugf("Starting packer build with template '%s'", a.PackerArtifact.Template)
	cmd := exec.Command("packer", "build", a.PackerArtifact.Template)
	cmd.Dir = a.Workspace
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return "", errors.Wrap(err, "running command")
	}
		
	logrus.Debugf("Reading packer manifest '%s'", a.PackerArtifact.Manifest)
	packerManifestBytes, err := ioutil.ReadFile(filepath.Join(a.Workspace, a.PackerArtifact.Manifest))
	if err != nil {
		return "", errors.Wrap(err, "reading manifest")
	}

	packerManifest := PackerManifest{}
	err = json.Unmarshal(packerManifestBytes, &packerManifest)
	if err != nil {
		return "", errors.Wrap(err, "parsing manifest json")
	}
	
	lastBuild := packerManifest.PackerBuilds[len(packerManifest.PackerBuilds)-1]
	runID := lastBuild.PackerRunUUID
	if runID != packerManifest.LastRunUUID {
		return "", errors.Wrap(err, "last build was not the last run")
	}
	if "docker" != lastBuild.BuilderType {
		return "", errors.Wrap(err, "last build was not a packer build")
	}

	imageTag := lastBuild.ArtifactID
	logrus.Debugf("Found Packer-built Docker image with digest: %s", imageTag)

	return imageTag, nil
}
