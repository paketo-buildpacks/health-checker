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

package hc_test

import (
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/health-checker/hc"
	"github.com/sclevine/spec"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		build hc.Build
		ctx   libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Application.Path = t.TempDir()
		Expect(err).NotTo(HaveOccurred())

		ctx.Plan.Entries = append(ctx.Plan.Entries, libcnb.BuildpackPlanEntry{Name: "hc"})
		ctx.Buildpack.Metadata = map[string]interface{}{
			"dependencies": []map[string]interface{}{
				{
					"id":      "thc",
					"version": "0.1.0",
					"stacks":  []interface{}{"test-stack-id"},
				},
			},
		}
		ctx.StackID = "test-stack-id"
	})

	it("contributes tiny health checker", func() {
		ctx.Buildpack.Metadata = map[string]interface{}{
			"configurations": []map[string]interface{}{
				{
					"name":    "BP_HEALTH_CHECKER_DEPENDENCY",
					"default": "thc",
				},
			},
			"dependencies": []map[string]interface{}{
				{
					"id":      "thc",
					"version": "0.1.0",
					"stacks":  []interface{}{"test-stack-id"},
					"cpes":    []string{"cpe:2.3:a:tiny-health-checker:tiny-health-checker:0.1.0:*:*:*:*:*:*:*"},
					"purl":    "pkg:generic/tiny-health-checker@0.1.0?arch=amd64",
				},
			},
		}
		result, err := build.Build(ctx)
		Expect(err).NotTo(HaveOccurred())

		Expect(result.Layers).To(HaveLen(1))
		Expect(result.Layers[0].Name()).To(Equal("thc"))
		Expect(result.Processes[0].Type).To(Equal("health-check"))
		Expect(result.Processes[0].Command).To(Equal("thc"))
		Expect(result.Processes[0].Arguments).To(HaveLen(0))
		Expect(result.Processes[0].Default).To(BeFalse())
		Expect(result.Processes[0].Direct).To(BeTrue())
	})
}
