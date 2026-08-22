package rpmostree

import (
	"encoding/json/jsontext"
)

type RPMOstreeStatus struct {
	Deployments []struct {
		Unlocked                           string         `json:"unlocked"`
		RequestedLocalPackages             jsontext.Value `json:"requested-local-packages"`
		BaseCommitMeta                     jsontext.Value `json:"base-commit-meta"`
		BaseRemovals                       jsontext.Value `json:"base-removals"`
		Pinned                             bool           `json:"pinned"`
		Osname                             string         `json:"osname"`
		BaseRemoteReplacements             jsontext.Value `json:"base-remote-replacements"`
		RegenerateInitramfs                bool           `json:"regenerate-initramfs"`
		Checksum                           string         `json:"checksum"`
		ContainerImageReferenceDigest      string         `json:"container-image-reference-digest"`
		RequestedBaseLocalReplacements     jsontext.Value `json:"requested-base-local-replacements"`
		ID                                 string         `json:"id"`
		Version                            string         `json:"version"`
		RequestedLocalFileoverridePackages jsontext.Value `json:"requested-local-fileoverride-packages"`
		RequestedBaseRemovals              jsontext.Value `json:"requested-base-removals"`
		RequestedPackages                  jsontext.Value `json:"requested-packages"`
		Serial                             int            `json:"serial"`
		Timestamp                          int            `json:"timestamp"`
		Staged                             bool           `json:"staged"`
		Booted                             bool           `json:"booted"`
		ContainerImageReference            string         `json:"container-image-reference"`
		Packages                           jsontext.Value `json:"packages"`
		BaseLocalReplacements              jsontext.Value `json:"base-local-replacements"`
	} `json:"deployments"`
	Transaction  jsontext.Value `json:"transaction"`
	CachedUpdate jsontext.Value `json:"cached-update"`
	UpdateDriver struct {
		DriverName   string `json:"driver-name"`
		DriverSdUnit string `json:"driver-sd-unit"`
	} `json:"update-driver"`
}
