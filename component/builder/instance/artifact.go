// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import "github.com/hashicorp/packer-plugin-sdk/packer"

var _ packer.Artifact = (*Artifact)(nil)

// Artifact represents the Oxide image created by the builder. This artifact
// contains the image ID and name that can be used to launch new instances.
type Artifact struct {
	// Unique identifier of the created image.
	ImageID string
	// Name of the created image.
	ImageName string
	// Additional state data associated with the build.
	StateData map[string]any
}

// BuilderId returns the builder ID used to create this artifact.
func (*Artifact) BuilderId() string {
	return BuilderID
}

// Files returns the files associated with the artifact.
func (a *Artifact) Files() []string {
	return nil
}

// Id returns the unique identifier for this artifact.
func (a *Artifact) Id() string {
	return a.ImageID
}

// String returns a description of the artifact.
func (a *Artifact) String() string {
	return a.ImageName
}

// State returns builder state related to the artifact.
func (a *Artifact) State(name string) any {
	return a.StateData[name]
}

// Destroy deletes the artifact when it is determined to no longer be needed.
func (a *Artifact) Destroy() error {
	return nil
}
