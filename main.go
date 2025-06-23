// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"
	"github.com/oxidecomputer/packer-plugin-oxide/component/builder/instance"
	"github.com/oxidecomputer/packer-plugin-oxide/component/data-source/image"
)

var (
	// Version is the MAJOR.MINOR.PATCH version of this plugin following semantic
	// versioning rules. Maintainers must update this as they develop this plugin.
	Version = "0.1.0"

	// VersionPreRelease is the pre-release identifier for the version. This is
	// generally modified via ldflags to create both release and pre-release builds.
	VersionPreRelease = "dev"

	// VersionMetadata is the build metadata for the version. This is generally
	// modified via ldflags to add build information to version.
	VersionMetadata = ""
)

func main() {
	pluginSet := plugin.NewSet()
	pluginSet.RegisterBuilder("instance", new(instance.Builder))
	pluginSet.RegisterDatasource("image", new(image.Datasource))
	pluginSet.SetVersion(
		version.NewPluginVersion(Version, VersionPreRelease, VersionMetadata),
	)

	if err := pluginSet.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
