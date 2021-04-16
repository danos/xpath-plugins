// Copyright (c) 2021, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xpathtest"
	"github.com/danos/yang/xpath/xutils"
)

type vifInterfaceTestSpec struct {
	name          string
	config        []xutils.PathType
	startPath     string
	expBoolResult bool
	expNumResult  float64
}

func TestParentStringLengthMatch(t *testing.T) {

	tests := []vifInterfaceTestSpec{
		{
			name: "Parent string length (tagnode) - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond1", "vif/tagnode+22"},
			},
			startPath:    "/interfaces/bonding/vif",
			expNumResult: 8,
		},
		{
			name: "Parent string length (ifname) - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/ifname+dp0bond1", "vif/tagnode+22"},
			},
			startPath:    "/interfaces/bonding/vif",
			expNumResult: 8,
		},
		{
			name: "Parent string length (name) - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/name+dp0bond1", "vif/tagnode+22"},
			},
			startPath:    "/interfaces/bonding/vif",
			expNumResult: 8,
		},
		{
			name: "Parent string length () - FAIL",
			config: []xutils.PathType{
				{"interfaces", "bonding/notTagnode+dp0bond1", "vif/tagnode+22"},
			},
			startPath:    "/interfaces/bonding/vif",
			expNumResult: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testTree := xpathtest.CreateTree(t, test.config)

			testNode := testTree.FindFirstNode(
				xutils.NewPathType(test.startPath))
			ns := xpath.NewNodesetDatum([]xutils.XpathNode{testNode})

			actResult :=
				parentInterfaceStringLength([]xpath.Datum{ns}).Number(
					"(unused value)")
			if test.expNumResult != actResult {
				t.Fatalf("Unexpected result for %s: exp %v, got %v\n",
					test.name, test.expNumResult, actResult)
			}
		})
	}
}

func TestCheckVlanValuesDoNotConflict(t *testing.T) {

	tests := []vifInterfaceTestSpec{
		{
			name: "Conflicting VLANs: vlan not set - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond1", "vif/tagnode+22"},
			},
			expBoolResult: true,
		},
		{
			name: "Conflicting VLANs: vlan set once, no other vlans - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond2", "vif/tagnode+22",
					"vlan+333"},
			},
			expBoolResult: true,
		},
		{
			name: "Conflicting VLANs: vlan set once, other vlans - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond3", "vif/tagnode+44",
					"vlan+555"},
				{"interfaces", "bonding/tagnode+dp0bond3", "vif/tagnode+66",
					"vlan+666"},
			},
			expBoolResult: true,
		},
		{
			name: "Conflicting VLANs: vlan set twice - FAIL",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond4", "vif/tagnode+44",
					"vlan+777"},
				{"interfaces", "bonding/tagnode+dp0bond4", "vif/tagnode+66",
					"vlan+777"},
			},
			expBoolResult: false,
		},
		{
			name: "Conflicting VLANs: vlan and inner-vlan set twice - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond5", "vif/tagnode+44",
					"vlan+777"},
				{"interfaces", "bonding/tagnode+dp0bond5", "vif/tagnode+44",
					"inner-vlan+888"},
				{"interfaces", "bonding/tagnode+dp0bond5", "vif/tagnode+66",
					"vlan+777"},
				{"interfaces", "bonding/tagnode+dp0bond5", "vif/tagnode+66",
					"inner-vlan+999"},
			},
			expBoolResult: true,
		},
		{
			name: "Conflicting VLANs: vlan set twice, inner once - FAIL",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond6", "vif/tagnode+44",
					"vlan+777"},
				{"interfaces", "bonding/tagnode+dp0bond6", "vif/tagnode+44",
					"inner-vlan+888"},
				{"interfaces", "bonding/tagnode+dp0bond6", "vif/tagnode+66",
					"vlan+777"},
			},
			expBoolResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testTree := xpathtest.CreateTree(t, test.config)

			// Test higher level function on 'interfaces/bonding' node.
			testNode1 := testTree.FindFirstNode(
				xutils.NewPathType("/interfaces/bonding"))
			ns1 := xpath.NewNodesetDatum([]xutils.XpathNode{testNode1})

			actResult :=
				validateVifVlanSettings([]xpath.Datum{ns1}).Boolean(
					"(unused value)")
			if test.expBoolResult != actResult {
				t.Fatalf("Unexpected result (1) for %s: exp %t, got %t\n",
					test.name, test.expBoolResult, actResult)
			}

			// Now test lower level function on 'interfaces/bonding/vif' node.
			testNode2 := testTree.FindFirstNode(
				xutils.NewPathType("/interfaces/bonding/vif"))
			ns2 := xpath.NewNodesetDatum([]xutils.XpathNode{testNode2})
			actResult =
				checkVlanValuesDoNotConflict([]xpath.Datum{ns2}).Boolean(
					"(unused value)")
			if test.expBoolResult != actResult {
				t.Fatalf("Unexpected result (2) for %s: exp %t, got %t\n",
					test.name, test.expBoolResult, actResult)
			}
		})
	}
}

func TestCheckImplicitVlanIdUnique(t *testing.T) {

	tests := []vifInterfaceTestSpec{
		{
			name: "Implicit vlan: vlan set - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond1",
					"vif/tagnode+22", "vlan+222"},
			},
			expBoolResult: true,
		},
		{
			name: "Implicit vlan: inner-vlan set - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond1",
					"vif/tagnode+22", "inner-vlan+222"},
			},
			expBoolResult: true,
		},
		{
			name: "Implicit vlan - PASS",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond1",
					"vif/tagnode+22"},
				{"interfaces", "bonding/tagnode+dp0bond1",
					"vif/tagnode+33", "vlan+333"},
			},
			expBoolResult: true,
		},
		{
			name: "Implicit vlan - FAIL",
			config: []xutils.PathType{
				{"interfaces", "bonding/tagnode+dp0bond1",
					"vif/tagnode+44"},
				{"interfaces", "bonding/tagnode+dp0bond1",
					"vif/tagnode+55", "vlan+44"},
			},
			expBoolResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testTree := xpathtest.CreateTree(t, test.config)

			// First test higher level function on interfaces/bonding node.
			testNode1 := testTree.FindFirstNode(
				xutils.NewPathType("/interfaces/bonding"))
			ns1 := xpath.NewNodesetDatum([]xutils.XpathNode{testNode1})

			actResult :=
				validateVifVlanSettings([]xpath.Datum{ns1}).Boolean(
					"(unused value)")
			if test.expBoolResult != actResult {
				t.Fatalf("Unexpected result (1) for %s: exp %t, got %t\n",
					test.name, test.expBoolResult, actResult)
			}

			// Then test lower level function on interfaces/bonding/vif node.
			testNode2 := testTree.FindFirstNode(
				xutils.NewPathType("/interfaces/bonding/vif"))
			ns2 := xpath.NewNodesetDatum([]xutils.XpathNode{testNode2})

			actResult =
				checkImplicitVlanIdUnique([]xpath.Datum{ns2}).Boolean(
					"(unused value)")
			if test.expBoolResult != actResult {
				t.Fatalf("Unexpected result (2) for %s: exp %t, got %t\n",
					test.name, test.expBoolResult, actResult)
			}
		})
	}
}
