// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xpathtest"
	"github.com/danos/yang/xpath/xutils"
)

type interfaceLeafrefTest struct {
	name,
	ref,
	startPath string
	expAll,
	expL3,
	expOrig bool
}

func TestInterfaceLeafrefMustReplacement(t *testing.T) {

	// NB: while the original 'interface must' allows VIFs using both 'tagnode'
	//     and 'ifname', all VIF interfaces use the same YANG that has
	//     'tagnode', so we don't test for 'ifname' in VIF here.
	tests := []interfaceLeafrefTest{
		{
			name:      "Valid tagnode",
			ref:       "dp0s2",
			startPath: "/feature/intf-ref",
			expAll:    true,
			expL3:     true,
			expOrig:   true,
		},
		{
			name:      "Invalid tagnode",
			ref:       "dp0s999",
			startPath: "/feature/intf-ref",
			expAll:    false,
			expL3:     false,
			expOrig:   false,
		},
		{
			name:      "Valid ifname",
			ref:       "erspan4",
			startPath: "/feature/intf-ref",
			expAll:    true,
			expL3:     true,
			expOrig:   true,
		},
		{
			name:      "Invalid ifname",
			ref:       "erspan999",
			startPath: "/feature/intf-ref",
			expAll:    false,
			expL3:     false,
			expOrig:   false,
		},
		{
			name:      "Valid tagnode / VIF (tagnode)",
			ref:       "dp0s1.1",
			startPath: "/feature/intf-ref",
			expAll:    true,
			expL3:     true,
			expOrig:   true,
		},
		{
			name:      "Valid switch (neither tagnode nor ifname)",
			ref:       "sw1",
			startPath: "/feature/intf-ref",
			expAll:    true,
			expL3:     false,
			expOrig:   false,
		},
		{
			name:      "Valid neither tagnode nor ifname / VIF (tagnode)",
			ref:       "sw2.1",
			startPath: "/feature/intf-ref",
			expAll:    true,
			expL3:     true,
			expOrig:   true,
		},
		{
			name:      "Invalid VIF (tagnode) for tagnode interface",
			ref:       "dp0s1.2",
			startPath: "/feature/intf-ref",
			expAll:    false,
			expL3:     false,
			expOrig:   false,
		},
		{
			name:      "Valid vhost",
			ref:       "vhost3",
			startPath: "/feature/intf-ref",
			expAll:    true,
			expL3:     true,  // false with must, now true
			expOrig:   false, // false with must, still false here.
		},
		{
			name:      "Valid backplane",
			ref:       "sw1",
			startPath: "/feature/intf-ref",
			expAll:    true,
			expL3:     false,
			expOrig:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.ref == "" {
				t.Fatalf("Interface reference not specified.")
			}
			testTree := xpathtest.CreateTree(t,
				[]xutils.PathType{
					// dataplane tagnode
					{"interfaces", "dataplane/tagnode+dp0s1", "address@1111"},
					{"interfaces", "dataplane/tagnode+dp0s2", "address@2222"},
					{"interfaces", "dataplane/tagnode+dp0s2", "address@2223"},
					{"interfaces", "dataplane/tagnode+dp0s3"},
					// dataplane tagnode + vif ...
					{"interfaces", "dataplane/tagnode+dp0s1",
						"vif/tagnode+1"},
					// switch (not L3)
					{"interfaces", "switch/name+sw1"},
					// switch vif
					{"interfaces", "switch/name+sw2", "vif/tagnode+1"},
					// erspan with ifname
					{"interfaces", "erspan/ifname+erspan4"},
					// backplane (excluded from L3)
					{"interfaces", "backplane/name+bp1"},
					// vhost
					{"interfaces", "vhost/name+vhost3"},
					// test leafref
					{"feature", "intf-ref+" + test.ref},
				})

			testNode := testTree.FindFirstNode(
				xutils.NewPathType(test.startPath))
			ns := xpath.NewNodesetDatum([]xutils.XpathNode{testNode})

			actAll := isInterfaceLeafref([]xpath.Datum{ns})
			if actAll.Boolean("(not used)") != test.expAll {
				t.Fatalf("Leafref failed on all interface match.")
			}

			actL3 := isL3InterfaceLeafref([]xpath.Datum{ns})
			if actL3.Boolean("(not used)") != test.expL3 {
				t.Fatalf("Leafref failed on L3 interface match.")
			}

			actOrig := isInterfaceLeafrefOriginal([]xpath.Datum{ns})
			if actOrig.Boolean("(not used)") != test.expOrig {
				t.Fatalf("Leafref failed on original interface must match.")
			}
		})
	}
}
