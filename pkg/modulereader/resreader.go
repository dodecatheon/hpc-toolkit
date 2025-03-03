/**
 * Copyright 2022 Google LLC
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package modulereader extracts necessary information from modules
package modulereader

import (
	"embed"
	"fmt"
	"hpc-toolkit/pkg/sourcereader"
	"io/ioutil"
	"log"
	"path"
	"strings"
)

// ModuleFS contains embedded modules (./modules) for use in building
var ModuleFS embed.FS

// VarInfo stores information about a module's input or output variables
type VarInfo struct {
	Name        string
	Type        string
	Description string
	Default     interface{}
	Required    bool
}

// ModuleInfo stores information about a module
type ModuleInfo struct {
	Inputs       []VarInfo
	Outputs      []VarInfo
	RequiredApis []string
}

// GetOutputsAsMap returns the outputs list as a map for quicker access
func (i ModuleInfo) GetOutputsAsMap() map[string]VarInfo {
	outputsMap := make(map[string]VarInfo)
	for _, output := range i.Outputs {
		outputsMap[output.Name] = output
	}
	return outputsMap
}

// GetModuleInfo gathers information about a module at a given source using the
// tfconfig package. For applicable sources, this function also stages the
// module contents in a local temp directory and will add required APIs to be
// enabled for that module.
func GetModuleInfo(source string, kind string) (ModuleInfo, error) {
	var modPath string
	switch {
	case sourcereader.IsGitPath(source):
		tmpDir, err := ioutil.TempDir("", "module-*")
		if err != nil {
			return ModuleInfo{}, err
		}
		modPath = path.Join(tmpDir, "module")
		sourceReader := sourcereader.Factory(source)
		if err = sourceReader.GetModule(source, modPath); err != nil {
			return ModuleInfo{}, fmt.Errorf("failed to clone git module at %s: %v", source, err)
		}

	case sourcereader.IsEmbeddedPath(source) || sourcereader.IsLocalPath(source):
		modPath = source

	default:
		return ModuleInfo{}, fmt.Errorf("Source is not valid: %s", source)
	}

	reader := Factory(kind)
	mi, err := reader.GetInfo(modPath)

	// add APIs required by the module, if known
	if sourcereader.IsEmbeddedPath(source) {
		mi.RequiredApis = defaultAPIList(modPath)
	} else if sourcereader.IsLocalPath(source) {
		if idx := strings.Index(modPath, "/community/modules/"); idx != -1 {
			mi.RequiredApis = defaultAPIList(modPath[idx+1:])
		} else if idx := strings.Index(modPath, "/modules/"); idx != -1 {
			mi.RequiredApis = defaultAPIList(modPath[idx+1:])
		}
	}
	return mi, err
}

// ModReader is a module reader interface
type ModReader interface {
	GetInfo(path string) (ModuleInfo, error)
	SetInfo(path string, modInfo ModuleInfo)
}

var kinds = map[string]ModReader{
	"terraform": TFReader{allModInfo: make(map[string]ModuleInfo)},
	"packer":    PackerReader{allModInfo: make(map[string]ModuleInfo)},
}

// IsValidKind returns true if the kind input is valid
func IsValidKind(input string) bool {
	for k := range kinds {
		if k == input {
			return true
		}
	}
	return false
}

// Factory returns a ModReader of type 'kind'
func Factory(kind string) ModReader {
	for k, v := range kinds {
		if kind == k {
			return v
		}
	}
	log.Fatalf("Invalid request to create a reader of kind %s", kind)
	return nil
}

func defaultAPIList(source string) []string {
	// API lists at
	// https://console.cloud.google.com/apis/dashboard and
	// https://console.cloud.google.com/apis/library
	staticAPIMap := map[string][]string{
		"community/modules/compute/SchedMD-slurm-on-gcp-partition": {
			"compute.googleapis.com",
		},
		"community/modules/compute/htcondor-execute-point": {
			"compute.googleapis.com",
		},
		"community/modules/compute/pbspro-execution": {
			"compute.googleapis.com",
			"storage.googleapis.com",
		},
		"community/modules/compute/schedmd-slurm-gcp-v5-partition": {
			"compute.googleapis.com",
		},
		"community/modules/database/slurm-cloudsql-federation": {
			"bigqueryconnection.googleapis.com",
			"sqladmin.googleapis.com",
		},
		"community/modules/file-system/DDN-EXAScaler": {
			"compute.googleapis.com",
			"deploymentmanager.googleapis.com",
			"iam.googleapis.com",
			"runtimeconfig.googleapis.com",
		},
		"community/modules/file-system/Intel-DAOS": {
			"compute.googleapis.com",
			"iam.googleapis.com",
			"secretmanager.googleapis.com",
		},
		"community/modules/file-system/nfs-server": {
			"compute.googleapis.com",
		},
		"community/modules/project/new-project": {
			"admin.googleapis.com",
			"cloudresourcemanager.googleapis.com",
			"cloudbilling.googleapis.com",
			"iam.googleapis.com",
		},
		"community/modules/project/service-account": {
			"iam.googleapis.com",
		},
		"community/modules/project/service-enablement": {
			"serviceusage.googleapis.com",
		},
		"community/modules/scheduler/SchedMD-slurm-on-gcp-controller": {
			"compute.googleapis.com",
		},
		"community/modules/scheduler/SchedMD-slurm-on-gcp-login-node": {
			"compute.googleapis.com",
		},
		"modules/scheduler/batch-job-template": {
			"batch.googleapis.com",
			"compute.googleapis.com",
		},
		"modules/scheduler/batch-login-node": {
			"batch.googleapis.com",
			"compute.googleapis.com",
			"storage.googleapis.com",
		},
		"community/modules/scheduler/htcondor-configure": {
			"iam.googleapis.com",
			"secretmanager.googleapis.com",
		},
		"community/modules/scheduler/pbspro-client": {
			"compute.googleapis.com",
			"storage.googleapis.com",
		},
		"community/modules/scheduler/pbspro-server": {
			"compute.googleapis.com",
			"storage.googleapis.com",
		},
		"community/modules/scheduler/schedmd-slurm-gcp-v5-controller": {
			"compute.googleapis.com",
			"iam.googleapis.com",
			"pubsub.googleapis.com",
			"secretmanager.googleapis.com",
		},
		"community/modules/scheduler/schedmd-slurm-gcp-v5-hybrid": {
			"compute.googleapis.com",
			"pubsub.googleapis.com",
		},
		"community/modules/scheduler/schedmd-slurm-gcp-v5-login": {
			"compute.googleapis.com",
		},
		"community/modules/scripts/htcondor-install": {},
		"community/modules/scripts/omnia-install":    {},
		"community/modules/scripts/pbspro-preinstall": {
			"iam.googleapis.com",
			"storage.googleapis.com",
		},
		"community/modules/scripts/pbspro-install": {},
		"community/modules/scripts/pbspro-qmgr":    {},
		"community/modules/scripts/spack-install":  {},
		"community/modules/scripts/wait-for-startup": {
			"compute.googleapis.com",
		},
		"modules/compute/vm-instance": {
			"compute.googleapis.com",
		},
		"modules/file-system/filestore": {
			"file.googleapis.com",
		},
		"modules/file-system/cloud-storage-bucket": {
			"storage.googleapis.com",
		},
		"modules/file-system/pre-existing-network-storage": {},
		"modules/monitoring/dashboard": {
			"stackdriver.googleapis.com",
		},
		"modules/network/pre-existing-vpc": {
			"compute.googleapis.com",
		},
		"modules/network/vpc": {
			"compute.googleapis.com",
		},
		"modules/packer/custom-image": {
			"compute.googleapis.com",
			"storage.googleapis.com",
		},
		"modules/scripts/startup-script": {
			"storage.googleapis.com",
		},
	}

	requiredAPIs, found := staticAPIMap[source]
	if !found {
		return []string{}
	}
	return requiredAPIs
}
