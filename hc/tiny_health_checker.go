/*
 * Copyright 2018-2025 the original author or authors.
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
	"github.com/heroku/color"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
	"github.com/paketo-buildpacks/libpak/crush"
)

type TinyHealthChecker struct {
	ApplicationPath  string
	ConfigResolver   libpak.ConfigurationResolver
	LayerContributor libpak.DependencyLayerContributor
	Logger           bard.Logger
}

func NewTinyHealthChecker(dependency libpak.BuildpackDependency, cache libpak.DependencyCache, cr libpak.ConfigurationResolver, appPath string) TinyHealthChecker {
	contributor := libpak.NewDependencyLayerContributor(dependency, cache, libcnb.LayerTypes{
		Build:  false,
		Cache:  false,
		Launch: true,
	})

	return TinyHealthChecker{ApplicationPath: appPath, LayerContributor: contributor, ConfigResolver: cr}
}

func (t TinyHealthChecker) Contribute(layer libcnb.Layer) (libcnb.Layer, error) {
	t.LayerContributor.Logger = t.Logger

	binDir := filepath.Join(layer.Path, "bin")
	hcBin := filepath.Join(binDir, "thc")
	symlinkPath := filepath.Join(t.ApplicationPath, "health-check")

	newLayer, err := t.LayerContributor.Contribute(layer, func(artifact *os.File) (libcnb.Layer, error) {
		t.Logger.Bodyf("Expanding to %s", layer.Path)
		if err := crush.Extract(artifact, layer.Path, 1); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to expand tiny-health-checker\n%w", err)
		}

		if err := os.MkdirAll(binDir, 0755); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to mkdir\n%w", err)
		}

		if err := os.Symlink(filepath.Join(layer.Path, "thc"), hcBin); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to symlink thc\n%w", err)
		}

		if err := os.Chmod(hcBin, 0775); err != nil {
			return libcnb.Layer{}, fmt.Errorf("unable to `chmod 775` binary %s\n%w", hcBin, err)
		}

		return layer, nil
	})

	// symlink `thc` into `/workspace/health-check` so it's easy to invoke
	t.Logger.Bodyf("The Tiny Health Checker binary is available at %s", hcBin)
	if err := os.Symlink(hcBin, symlinkPath); err != nil && !os.IsExist(err) {
		return libcnb.Layer{}, fmt.Errorf("unable to create symlink\n%s", err)
	} else if err != nil && os.IsExist(err) {
		t.Logger.Bodyf(color.New(color.Faint, color.Bold).Sprintf("WARNING: A file already exists at %s, skipping creation of symlink", symlinkPath))
	} else {
		t.Logger.Bodyf("A symlink is available at %s for your convenience", symlinkPath)
	}

	return newLayer, err
}

func (t TinyHealthChecker) Name() string {
	return t.LayerContributor.LayerName()
}
