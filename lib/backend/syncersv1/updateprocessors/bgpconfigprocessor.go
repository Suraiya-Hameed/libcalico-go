// Copyright (c) 2017 Tigera, Inc. All rights reserved.

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package updateprocessors

import (
	"encoding/json"
	"reflect"
	"strings"

	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	"github.com/projectcalico/libcalico-go/lib/backend/watchersyncer"
	log "github.com/sirupsen/logrus"
)

// Create a new SyncerUpdateProcessor to sync BGPConfiguration data in v1 format for
// consumption by the BGP daemon.
func NewBGPConfigUpdateProcessor() watchersyncer.SyncerUpdateProcessor {
	return NewConfigUpdateProcessor(
		reflect.TypeOf(apiv3.BGPConfigurationSpec{}),
		AllowAnnotations,
		func(node, name string) model.Key { return model.NodeBGPConfigKey{Nodename: node, Name: name} },
		func(name string) model.Key { return model.GlobalBGPConfigKey{Name: name} },
		map[string]ConfigFieldValueToV1ModelValue{
			"loglevel":              logLevelToBirdLogLevel,
			"node_mesh":             nodeMeshToString,
			"svc_external_ips":      svcExternalIpsToString,
			"svc_cluster_ips":       svcClusterIpsToString,
			"communities":           communitiesToString,
			"prefix_advertisements": prefixAdvertisementsToString,
			"listen_port":           listenPortToString,
		},
	)
}

// Bird log level currently only supports granularity of none, debug and info.  Debug/Info are
// left unchanged, all others treated as none.
var logLevelToBirdLogLevel = func(value interface{}) interface{} {
	l := strings.ToLower(value.(string))
	switch l {
	case "", "debug", "info":
	default:
		l = "none"
	}
	return l
}

var nodeToNodeMeshEnabled = "{\"enabled\":true}"
var nodeToNodeMeshDisabled = "{\"enabled\":false}"

// In v1, the node mesh enabled field was wrapped up in some JSON - wrap up the value to
// return via the syncer.
var nodeMeshToString = func(value interface{}) interface{} {
	enabled := value.(bool)
	if enabled {
		return nodeToNodeMeshEnabled
	}
	return nodeToNodeMeshDisabled
}

// We wrap each Service external IP in a ServiceExternalIPBlock struct to
// achieve the desired API structure. This unpacks that.
var svcExternalIpsToString = func(value interface{}) interface{} {
	ipBlocks := value.([]apiv3.ServiceExternalIPBlock)

	// Processor expects all empty fields to be nil.
	if len(ipBlocks) == 0 {
		return nil
	}

	ipCidrs := make([]string, 0)
	for _, ipBlock := range ipBlocks {
		ipCidrs = append(ipCidrs, ipBlock.CIDR)
	}

	return strings.Join(ipCidrs, ",")
}

// We wrap each Service Cluster IP in a ServiceClusterIPBlock to
// achieve the desired API structure. This unpacks that.
var svcClusterIpsToString = func(value interface{}) interface{} {
	ipBlocks := value.([]apiv3.ServiceClusterIPBlock)

	// Processor expects all empty fields to be nil.
	if len(ipBlocks) == 0 {
		return nil
	}

	ipCidrs := make([]string, 0)
	for _, ipBlock := range ipBlocks {
		ipCidrs = append(ipCidrs, ipBlock.CIDR)
	}

	return strings.Join(ipCidrs, ",")
}

// return JSON encoded string of CommunityKVPair
var communitiesToString = func(value interface{}) interface{} {
	communities := value.([]apiv3.CommunityKVPair)
	if len(communities) == 0 {
		return nil
	}
	communitiesStr, err := json.Marshal(communities)
	if err != nil {
		log.Errorf("Error converting []apiv3.CommunityKVPair to string %+v", err)
		return nil
	}
	return communitiesStr
}

// return JSON encoded string of PrefixAdvertisements
var prefixAdvertisementsToString = func(value interface{}) interface{} {
	pa := value.([]apiv3.PrefixAdvertisements)
	if len(pa) == 0 {
		return nil
	}
	paStr, err := json.Marshal(pa)
	if err != nil {
		log.Errorf("Error converting []apiv3.PrefixAdvertisements to string %+v", err)
		return nil
	}
	return paStr
}

var listenPortToString = func(value interface{}) interface{} {
	listenPort := value.(uint16)
	if listenPort == 0 {
		return nil
	}
	return listenPort
}
