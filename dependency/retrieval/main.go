// Copyright 2018-2026 the original author or authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/paketo-buildpacks/libdependency/retrieve"
	"github.com/paketo-buildpacks/libdependency/upstream"
	"github.com/paketo-buildpacks/libdependency/versionology"
	"github.com/paketo-buildpacks/packit/v2/cargo"
)

const (
	id   = "thc"
	name = "Tiny Health Checker"
	purl = "pkg:generic/thc"

	org  = "dmikusa"
	repo = "tiny-health-checker"
)

// assetNames maps a buildpack architecture to the exact upstream asset name.
var assetNames = map[string]string{
	"amd64": "tiny-health-checker-x86_64-unknown-linux-musl.tar.xz",
	"arm64": "tiny-health-checker-aarch64-unknown-linux-musl.tar.xz",
}

type asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type release struct {
	TagName    string  `json:"tag_name"`
	Prerelease bool    `json:"prerelease"`
	Assets     []asset `json:"assets"`
}

type version struct {
	version *semver.Version
	tag     string
	assets  []asset
}

func (v version) Version() *semver.Version {
	return v.version
}

func main() {
	retrieve.NewMetadataWithPlatforms(id, getAllVersions, generateMetadata)
}

func getAllVersions() (versionology.VersionFetcherArray, error) {
	releases, err := fetchReleases(org, repo)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch releases\n%w", err)
	}

	var versions versionology.VersionFetcherArray
	for _, r := range releases {
		v, err := semver.NewVersion(strings.TrimPrefix(r.TagName, "v"))
		if err != nil {
			fmt.Printf("Skipping %s: unable to parse version\n", r.TagName)
			continue
		}

		versions = append(versions, version{version: v, tag: r.TagName, assets: r.Assets})
	}

	return versions, nil
}

func generateMetadata(versionFetcher versionology.VersionFetcher, platform retrieve.Platform) ([]versionology.Dependency, error) {
	v, ok := versionFetcher.(version)
	if !ok {
		return nil, fmt.Errorf("unexpected version type %T", versionFetcher)
	}

	versionString := v.version.String()

	archive := findAsset(v.assets, assetNames[platform.Arch])
	if archive == nil {
		fmt.Printf("Skipping %s: missing %s/%s asset\n", versionString, platform.OS, platform.Arch)
		return nil, nil
	}

	checksum, err := upstream.GetSHA256OfRemoteFile(archive.BrowserDownloadURL)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", archive.BrowserDownloadURL, err)
	}

	source := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", org, repo, v.tag)
	sourceChecksum, err := upstream.GetSHA256OfRemoteFile(source)
	if err != nil {
		return nil, fmt.Errorf("unable to checksum %s\n%w", source, err)
	}

	dependency := cargo.ConfigMetadataDependency{
		Arch:     platform.Arch,
		Checksum: fmt.Sprintf("sha256:%s", checksum),
		CPE:      fmt.Sprintf("cpe:2.3:a:dmikusa-pivotal:tiny-health-checker:%s:*:*:*:*:*:*:*", versionString),
		ID:       id,
		Licenses: []interface{}{
			map[string]string{
				"type": "Apache-2",
				"uri":  "https://github.com/dmikusa-pivotal/tiny-health-checker/blob/main/LICENSE",
			},
		},
		Name:           name,
		OS:             platform.OS,
		PURL:           fmt.Sprintf("%s@%s?arch=%s", purl, versionString, platform.Arch),
		SHA256:         checksum,
		Source:         source,
		SourceChecksum: fmt.Sprintf("sha256:%s", sourceChecksum),
		Stacks:         []string{"*"},
		URI:            archive.BrowserDownloadURL,
		Version:        versionString,
	}

	d, err := versionology.NewDependency(dependency, fmt.Sprintf("%s-%s", platform.OS, platform.Arch))
	if err != nil {
		return nil, err
	}

	return []versionology.Dependency{d}, nil
}

func findAsset(assets []asset, name string) *asset {
	if name == "" {
		return nil
	}

	for i := range assets {
		if assets[i].Name == name {
			return &assets[i]
		}
	}

	return nil
}

func fetchReleases(owner, repo string) ([]release, error) {
	var all []release
	for page := 1; ; page++ {
		releases, err := fetchPage(owner, repo, page)
		if err != nil {
			return nil, err
		}

		if len(releases) == 0 {
			break
		}

		for _, r := range releases {
			if r.Prerelease {
				continue
			}
			all = append(all, r)
		}
	}

	return all, nil
}

func fetchPage(owner, repo string, page int) ([]release, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases?page=%d&per_page=100", owner, repo, page)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github API returned status %d for %s", resp.StatusCode, url)
	}

	var releases []release
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	return releases, nil
}
