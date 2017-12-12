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

package conversionv1v3

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	apiv1 "github.com/projectcalico/libcalico-go/lib/apis/v1"
	"github.com/projectcalico/libcalico-go/lib/apis/v1/unversioned"
	apiv3 "github.com/projectcalico/libcalico-go/lib/apis/v3"
	"github.com/projectcalico/libcalico-go/lib/backend/model"
	"github.com/projectcalico/libcalico-go/lib/ipip"
	cnet "github.com/projectcalico/libcalico-go/lib/net"
)

// IPPool implements the Converter interface.
type IPPool struct{}

// APIV1ToBackendV1 converts v1 IPPool API to v1 IPPool KVPair.
func (_ IPPool) APIV1ToBackendV1(rIn unversioned.Resource) (*model.KVPair, error) {
	p := rIn.(*apiv1.IPPool)

	var ipipInterface string
	var ipipMode ipip.Mode
	if p.Spec.IPIP != nil {
		if p.Spec.IPIP.Enabled {
			ipipInterface = "tunl0"
		} else {
			ipipInterface = ""
		}
		ipipMode = p.Spec.IPIP.Mode
	}

	d := model.KVPair{
		Key: model.IPPoolKey{
			CIDR: p.Metadata.CIDR,
		},
		Value: &model.IPPool{
			CIDR:          p.Metadata.CIDR,
			IPIPInterface: ipipInterface,
			IPIPMode:      ipipMode,
			Masquerade:    p.Spec.NATOutgoing,
			IPAM:          !p.Spec.Disabled,
			Disabled:      p.Spec.Disabled,
		},
	}

	return &d, nil
}

// BackendV1ToAPIV3 converts v1 IPPool KVPair to v3 API.
func (_ IPPool) BackendV1ToAPIV3(kvp *model.KVPair) (Resource, error) {
	pool, ok := kvp.Value.(*model.IPPool)
	if !ok {
		return nil, fmt.Errorf("value is not a valid IPPool resource Value")
	}

	ipp := apiv3.NewIPPool()
	ipp.Name = cidrToName(pool.CIDR)
	ipp.Spec = apiv3.IPPoolSpec{
		CIDR:        pool.CIDR.String(),
		IPIPMode:    convertIPIPMode(pool.IPIPMode, pool.IPIPInterface),
		NATOutgoing: pool.Masquerade,
		Disabled:    pool.Disabled,
	}

	return ipp, nil
}

func convertIPIPMode(mode ipip.Mode, ipipInterface string) apiv3.IPIPMode {
	ipipMode := strings.ToLower(string(mode))

	if ipipInterface == "" {
		return apiv3.IPIPModeNever
	} else if ipipMode == "cross-subnet" {
		return apiv3.IPIPModeCrossSubnet
	}
	return apiv3.IPIPModeAlways
}

func cidrToName(cidr cnet.IPNet) string {
	name := strings.Replace(cidr.String(), ".", "-", 3)
	name = strings.Replace(name, ":", "-", 7)
	name = strings.Replace(name, "/", "-", 1)

	log.WithFields(log.Fields{
		"Name":  name,
		"IPNet": cidr.String(),
	}).Debug("Converted IPNet to resource name")

	return name
}