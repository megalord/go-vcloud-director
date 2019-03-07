/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	. "gopkg.in/check.v1"

	"github.com/vmware/go-vcloud-director/types/v56"
)

// Retrieves an external network and checks that its contents are filled as expected
func (vcd *TestVCD) Test_GetExternalNetwork(check *C) {

	fmt.Printf("Running: %s\n", check.TestName())
	if vcd.skipAdminTests {
		check.Skip(fmt.Sprintf(TestRequiresSysAdminPrivileges, check.TestName()))
	}
	networkName := vcd.config.VCD.ExternalNetwork
	if networkName == "" {
		check.Skip("No external network provided")
	}
	externalNetwork, err := GetExternalNetworkByName(vcd.client, networkName)
	check.Assert(err, IsNil)
	LogExternalNetwork(*externalNetwork)
	check.Assert(externalNetwork.HREF, Not(Equals), "")
	expectedType := "application/vnd.vmware.admin.extension.network+xml"
	check.Assert(externalNetwork.Name, Equals, networkName)
	check.Assert(externalNetwork.Type, Equals, expectedType)
}

func (vcd *TestVCD) Test_CreateExternalNetwork(check *C) {
	if vcd.skipAdminTests {
		check.Skip("Configuration org != 'Sysyem'")
	}
	networkName := vcd.config.VCD.ExternalNetwork
	if networkName == "" {
		check.Skip("No external network provided")
	}

	externalNetworkRef, err := GetExternalNetworkByName(vcd.client, networkName)
	check.Assert(err, IsNil)
	if *externalNetworkRef != (types.ExternalNetworkReference{}) {
		externalNetwork := NewExternalNetwork(&vcd.client.Client)
		externalNetwork.ExternalNetwork = &types.ExternalNetwork{
			HREF: externalNetworkRef.HREF,
		}
		err = externalNetwork.DeleteWait()
		check.Assert(err, IsNil)
	}

	results, err := vcd.client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "virtualCenter",
		"filter": fmt.Sprintf("(name==%s)", vcd.config.VCD.VimServer),
	})
	check.Assert(err, IsNil)
	if len(results.Results.VirtualCenterRecord) == 0 {
		check.Skip(fmt.Sprintf("No vSphere server found with name '%s'", vcd.config.VCD.VimServer))
	}
	vimServerHref := results.Results.VirtualCenterRecord[0].HREF

	externalNetwork := &types.ExternalNetwork{
		Name:        networkName,
		Xmlns:       "http://www.vmware.com/vcloud/extension/v1.5",
		Description: "Test Create External Network",
		Configuration: &types.NetworkConfiguration{
			Xmlns: "http://www.vmware.com/vcloud/v1.5",
			IPScopes: &types.IPScopes{
				IPScope: types.IPScope{
					Gateway: "192.168.201.1",
					Netmask: "255.255.255.0",
					DNS1:    "192.168.202.253",
					DNS2:    "192.168.202.254",
					IPRanges: &types.IPRanges{
						IPRange: []*types.IPRange{
							&types.IPRange{
								StartAddress: "192.168.201.3",
								EndAddress:   "192.168.201.250",
							},
						},
					},
				},
			},
			FenceMode: "isolated",
		},
		VimPortGroupRefs: &types.VimObjectRefs{
			VimObjectRef: []*types.VimObjectRef{
				&types.VimObjectRef{
					VimServerRef: &types.Reference{
						HREF: vimServerHref,
					},
					MoRef:         vcd.config.VCD.ExternalNetworkPortGroup,
					VimObjectType: "DV_PORTGROUP",
				},
			},
		},
	}
	task, err := CreateExternalNetwork(vcd.client, externalNetwork)
	check.Assert(err, IsNil)
	AddToCleanupList(externalNetwork.Name, "externalNetwork", "", "Test_CreateExternalNetwork")
	check.Assert(task, Not(Equals), Task{})

	err = task.WaitTaskCompletion()
	check.Assert(err, IsNil)

	externalNetworkRef, err = GetExternalNetworkByName(vcd.client, networkName)
	check.Assert(err, IsNil)
	check.Assert(externalNetworkRef.Name, Equals, networkName)
}
