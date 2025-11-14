package rpmostree

import (
	"encoding/json"
)

type RPMOstreeStatus struct {
	Deployments []struct {
		Unlocked                           string          `json:"unlocked"`
		RequestedLocalPackages             json.RawMessage `json:"requested-local-packages"`
		BaseCommitMeta                     json.RawMessage `json:"base-commit-meta"`
		BaseRemovals                       json.RawMessage `json:"base-removals"`
		Pinned                             bool            `json:"pinned"`
		Osname                             string          `json:"osname"`
		BaseRemoteReplacements             json.RawMessage `json:"base-remote-replacements"`
		RegenerateInitramfs                bool            `json:"regenerate-initramfs"`
		Checksum                           string          `json:"checksum"`
		ContainerImageReferenceDigest      string          `json:"container-image-reference-digest"`
		RequestedBaseLocalReplacements     json.RawMessage `json:"requested-base-local-replacements"`
		ID                                 string          `json:"id"`
		Version                            string          `json:"version"`
		RequestedLocalFileoverridePackages json.RawMessage `json:"requested-local-fileoverride-packages"`
		RequestedBaseRemovals              json.RawMessage `json:"requested-base-removals"`
		RequestedPackages                  json.RawMessage `json:"requested-packages"`
		Serial                             int             `json:"serial"`
		Timestamp                          int             `json:"timestamp"`
		Staged                             bool            `json:"staged"`
		Booted                             bool            `json:"booted"`
		ContainerImageReference            string          `json:"container-image-reference"`
		Packages                           json.RawMessage `json:"packages"`
		BaseLocalReplacements              json.RawMessage `json:"base-local-replacements"`
	} `json:"deployments"`
	Transaction  json.RawMessage `json:"transaction"`
	CachedUpdate json.RawMessage `json:"cached-update"`
	UpdateDriver struct {
		DriverName   string `json:"driver-name"`
		DriverSdUnit string `json:"driver-sd-unit"`
	} `json:"update-driver"`
}
