/*
 * Copyright 2018-2022 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package hc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/sherpa"
)

type TinyHealthChecker struct {
	ConfigResolver   libpak.ConfigurationResolver
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewTinyHealthChecker(dependency libpak.BuildpackDependency, cache libpak.DependencyCache, cr libpak.ConfigurationResolver) TinyHealthChecker {
	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Build:  false,
		Cache:  false,
		Launch: true,
	})

	return TinyHealthChecker{LayerContributor: contributor, ConfigResolver: cr}
}

func (t TinyHealthChecker) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	t.LayerContributor.Logger = t.Logger

	// if set at build time, we bake them into the image, user can override at runtime
	for _, envVarName := range []string{"THC_PORT", "THC_PATH", "CONN_TIMEOUT", "REQ_TIMEOUT"} {
		if val, found := t.ConfigResolver.Resolve(envVarName); found {
			layer.LaunchEnvironment.ProcessDefault("health-check", envVarName, val)
		}
	}

	return t.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		binDir := filepath.Join(layer.Path, "bin")

		t.Logger.Bodyf("Copying from %s to %s", artifact.Name(), binDir)

		if err := os.MkdirAll(binDir, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to mkdir\n%w", err)
		}

		hcBin := filepath.Join(binDir, "thc")
		if err := sherpa.CopyFile(artifact, hcBin); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to copy tiny health checker\n%w", err)
		}

		if err := os.Chmod(hcBin, 0775); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to `chmod 775` binary %s\n%w", hcBin, err)
		}

		return layer, nil
	})
}

func (t TinyHealthChecker) ContributeProcesses() []libcnb.Process {
	return []libcnb.Process{
		{
			Type:    "health-check",
			Command: "thc",
			Direct:  true,
			Default: false,
		},
	}
}

func (t TinyHealthChecker) Name() string {
	return t.LayerContributor.LayerName()
}
