// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

import "github.com/hashicorp/packer-plugin-sdk/packer"

var _ packer.Artifact = (*Artifact)(nil)

// Artifact is the result of this builder.
type Artifact struct {
	ImageID   string
	ImageName string
	StateData map[string]any
}

// BuilderId returns the builder ID used to create this artifact.
func (*Artifact) BuilderId() string {
	return "oxide.instance"
}

// Files returns the files associated with the artifact.
func (a *Artifact) Files() []string {
	return nil
}

// Id returns the unique identifier for this artifact.
func (a *Artifact) Id() string {
	return a.ImageID
}

// String returns a description to describe the artifact.
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
