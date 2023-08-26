// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fingerprint

import (
	"fmt"
    "strconv"
    "os/exec"
    "bytes"
	"strings"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/nomad/nomad/structs"
)

const bytesInMB int64 = 1024 * 1024

// MemoryFingerprint is used to fingerprint the available memory on the node
type MemoryFingerprint struct {
	StaticFingerprinter
	logger log.Logger
}

// NewMemoryFingerprint is used to create a Memory fingerprint
func NewMemoryFingerprint(logger log.Logger) Fingerprint {
	f := &MemoryFingerprint{
		logger: logger.Named("memory"),
	}
	return f
}

func strToUint64(str string) uint64 {
    ui64, err := strconv.ParseUint(str, 10, 64)
    if err != nil {
        panic(err)
    }

    return ui64
}

func getTotalPhysicalMem() uint64 {
    cmd := exec.Command("sysctl", "hw.usermem64")

    var out bytes.Buffer
    cmd.Stdout = &out

    err := cmd.Run()

    if err != nil {
        panic(err)
    }

    outStr := out.String()
    parts := strings.Split(outStr, " ")
    if len(parts) != 3 {
        panic("malformed sysctl hw.usermem64 output")
    }

    return strToUint64(strings.TrimSpace(parts[2]))
}

func (f *MemoryFingerprint) Fingerprint(req *FingerprintRequest, resp *FingerprintResponse) error {
	var totalMemory int64
	cfg := req.Config
	if cfg.MemoryMB != 0 {
		totalMemory = int64(cfg.MemoryMB) * bytesInMB
	} else {
	    totalMemory = int64(getTotalPhysicalMem())
	}

	if totalMemory > 0 {
		resp.AddAttribute("memory.totalbytes", fmt.Sprintf("%d", totalMemory))

		memoryMB := totalMemory / bytesInMB

		// COMPAT(0.10): Unused since 0.9.
		resp.Resources = &structs.Resources{
			MemoryMB: int(memoryMB),
		}

		resp.NodeResources = &structs.NodeResources{
			Memory: structs.NodeMemoryResources{
				MemoryMB: memoryMB,
			},
		}
	}

	return nil
}
