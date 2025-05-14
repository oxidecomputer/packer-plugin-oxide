// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package instance

type Artifact struct {
	StateData map[string]any
}

func (*Artifact) BuilderId() string {
	return "oxide.builder"
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (*Artifact) Id() string {
	return ""
}

func (a *Artifact) String() string {
	return ""
}

func (a *Artifact) State(name string) any {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return nil
}
