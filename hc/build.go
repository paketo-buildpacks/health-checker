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

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type ProcessContributor interface {
	ContributeProcesses() []libcnb.Process
}

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	result := libcnb.NewBuildResult()

	pr := libpak.PlanEntryResolver{Plan: context.Plan}
	if _, ok, err := pr.Resolve(PlanEntryHealthChecker); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to resolve health checker plan entry\n%w", err)
	} else if ok {
		b.Logger.Title(context.Buildpack)

		cr, err := libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
		}

		dc, err := libpak.NewDependencyCache(context)
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency cache\n%w", err)
		}
		dc.Logger = b.Logger

		dr, err := libpak.NewDependencyResolver(context)
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to create dependency resolver\n%w", err)
		}

		depId, _ := cr.Resolve("BP_HEALTH_CHECKER_DEPENDENCY")
		hcDependency, err := dr.Resolve(depId, "")
		if err != nil {
			return libcnb.BuildResult{}, fmt.Errorf("unable to find dependency\n%w", err)
		}

		hcLayercontributor := NewTinyHealthChecker(hcDependency, dc, cr)
		hcLayercontributor.Logger = b.Logger

		result.Layers = append(result.Layers, hcLayercontributor)
		result.Processes = hcLayercontributor.ContributeProcesses()
	}

	return result, nil
}
