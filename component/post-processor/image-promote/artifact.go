// Copyright (c) Oxide Computer Company
// SPDX-License-Identifier: MPL-2.0

package imagepromote

import (
	"fmt"
)

type Artifact struct {
	ImageID   string
	ImageName string
	Project   string
}

func (a *Artifact) BuilderId() string {
	return BuilderID
}

func (a *Artifact) Files() []string {
	return nil
}

func (a *Artifact) Id() string {
	return a.ImageID
}

func (a *Artifact) String() string {
	if a.Project != "" {
		return fmt.Sprintf("Promoted Oxide image: %s (ID: %s) in project: %s", a.ImageName, a.ImageID, a.Project)
	}
	return fmt.Sprintf("Promoted Oxide image: %s (ID: %s) at silo level", a.ImageName, a.ImageID)
}

func (a *Artifact) State(name string) interface{} {
	switch name {
	case "ImageID":
		return a.ImageID
	case "ImageName":
		return a.ImageName
	case "Project":
		return a.Project
	}
	return nil
}

func (a *Artifact) Destroy() error {
	// We don't destroy promoted images
	return nil
}
