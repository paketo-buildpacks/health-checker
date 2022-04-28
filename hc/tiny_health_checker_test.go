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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/health-checker/hc"
	"github.com/paketo-buildpacks/libpak"
)

func testTinyHealthChecker(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Layers.Path, err = ioutil.TempDir("", "thc-layers")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Layers.Path)).To(Succeed())
	})

	it("contributes health checker", func() {
		dep := libpak.BuildpackDependency{
			URI:    "https://localhost/stub-thc",
			SHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
		}
		dc := libpak.DependencyCache{CachePath: "testdata"}
		cr, err := libpak.NewConfigurationResolver(ctx.Buildpack, nil)
		Expect(err).ToNot(HaveOccurred())

		j := hc.NewTinyHealthChecker(dep, dc, cr)
		layer, err := ctx.Layers.Layer("test-layer")
		Expect(err).NotTo(HaveOccurred())

		layer, err = j.Contribute(layer)
		Expect(err).NotTo(HaveOccurred())

		Expect(layer.LayerTypes.Build).To(BeFalse())
		Expect(layer.LayerTypes.Cache).To(BeFalse())
		Expect(layer.LayerTypes.Launch).To(BeTrue())
		Expect(filepath.Join(layer.Path, "bin", "thc")).To(BeARegularFile())

		finfo, err := os.Stat(filepath.Join(layer.Path, "bin", "thc"))
		Expect(err).ToNot(HaveOccurred())
		Expect(finfo.Mode().Perm().String()).To(Equal("-rwxrwxr-x"))
	})

	context("THC_PORT & THC_PATH are set", func() {
		it.Before(func() {
			Expect(os.Setenv("THC_PORT", "8081")).To(Succeed())
			Expect(os.Setenv("THC_PATH", "/foo")).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("THC_PORT")).To(Succeed())
			Expect(os.Unsetenv("THC_PATH")).To(Succeed())
		})

		it("contributes a health checker with custom port and path", func() {
			dep := libpak.BuildpackDependency{
				URI:    "https://localhost/stub-thc",
				SHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			}
			dc := libpak.DependencyCache{CachePath: "testdata"}
			cr, err := libpak.NewConfigurationResolver(ctx.Buildpack, nil)
			Expect(err).ToNot(HaveOccurred())

			j := hc.NewTinyHealthChecker(dep, dc, cr)
			layer, err := ctx.Layers.Layer("test-layer")
			Expect(err).NotTo(HaveOccurred())

			layer, err = j.Contribute(layer)
			Expect(err).NotTo(HaveOccurred())

			Expect(layer.LayerTypes.Build).To(BeFalse())
			Expect(layer.LayerTypes.Cache).To(BeFalse())
			Expect(layer.LayerTypes.Launch).To(BeTrue())
			Expect(layer.LaunchEnvironment).To(Equal(libcnb.Environment{
				"health-check/THC_PATH.default": "/foo",
				"health-check/THC_PORT.default": "8081",
			}))
			Expect(filepath.Join(layer.Path, "bin", "thc")).To(BeARegularFile())

			finfo, err := os.Stat(filepath.Join(layer.Path, "bin", "thc"))
			Expect(err).ToNot(HaveOccurred())
			Expect(finfo.Mode().Perm().String()).To(Equal("-rwxrwxr-x"))
		})
	})
}
