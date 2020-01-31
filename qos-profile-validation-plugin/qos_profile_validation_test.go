// Copyright (c) 2019-2020, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xpathtest"
	"github.com/danos/yang/xpath/xutils"
)

type qosProfileTestSpec struct {
	name      string
	config    []xutils.PathType
	startPath string
	expResult bool
}

// Use these in 'config' entries below so it's obvious what is being changed
// from the 'default' / working case.
const (
	DSCP_GRP_HIGH   = "dscp-group/group-name+high"
	DSCP_GRP_MED    = "dscp-group/group-name+med"
	DSCP_GRP_LOW    = "dscp-group/group-name+low"
	QUEUE_ID_1      = "queue/id+1"
	QUEUE_ID_2      = "queue/id+2"
	TO_3            = "to+3"
	TO_4            = "to+4"
	TRAFFIC_CLASS_1 = "traffic-class+tc1"
	TRAFFIC_CLASS_2 = "traffic-class+tc2"
)

func TestQueueMatch(t *testing.T) {

	tests := []qosProfileTestSpec{
		{
			name: "Ingress map and match id and traffic-class - PASS",
			config: []xutils.PathType{
				{"policy", "ingress-map"},
				{"policy", "qos", "profile/name+prof1", "queue/id+1",
					"traffic-class+tc1"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+1", "traffic-class+tc1"},
			},
			startPath: "/policy/qos/profile/queue",
			expResult: true,
		},
		{
			name: "Ingress map and mismatched id and traffic-class - PASS",
			config: []xutils.PathType{
				{"policy", "ingress-map"},
				{"policy", "qos", "profile/name+prof1", "queue/id+1",
					"traffic-class+tc1"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+1", TRAFFIC_CLASS_2},
			},
			startPath: "/policy/qos/profile/queue",
			expResult: true,
		},
		{
			name: "Global profile match id and traffic-class - PASS",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "queue/id+1",
					"traffic-class+tc1"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+1", "traffic-class+tc1"},
			},
			startPath: "/policy/qos/profile/queue",
			expResult: true,
		},
		{
			name: "Global profile, mismatched traffic-class - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "queue/id+1",
					TRAFFIC_CLASS_2},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+1", TRAFFIC_CLASS_1},
			},
			startPath: "/policy/qos/profile/queue",
			expResult: false,
		},
		{
			name: "Global profile, mismatched id - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", QUEUE_ID_2,
					"traffic-class+tc1"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", QUEUE_ID_1, "traffic-class+tc1"},
			},
			startPath: "/policy/qos/profile/queue",
			expResult: false,
		},
		{
			name: "Local profile match id and traffic-class - PASS",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "queue/id+1",
					"traffic-class+tc1"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+1", "traffic-class+tc1"},
			},
			startPath: "/policy/qos/name/shaper/profile/queue",
			expResult: true,
		},
		{
			name: "Local profile, mismatched traffic-class - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "queue/id+1",
					TRAFFIC_CLASS_2},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+1", TRAFFIC_CLASS_1},
			},
			startPath: "/policy/qos/name/shaper/profile/queue",
			expResult: false,
		},
		{
			name: "Local profile, mismatched id - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", QUEUE_ID_2,
					"traffic-class+tc1"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", QUEUE_ID_1, "traffic-class+tc1"},
			},
			startPath: "/policy/qos/name/shaper/profile/queue",
			expResult: false,
		},
		{
			name: "Local profile against multiple profiles - PASS",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "queue/id+3",
					"traffic-class+tc99"},
				{"policy", "qos", "profile/name+prof2", "queue/id+3",
					"traffic-class+tc99"},
				{"policy", "qos", "profile/name+prof3", "queue/id+3",
					"traffic-class+tc99"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+3", "traffic-class+tc99"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "queue/id+3", "traffic-class+tc99"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profC", "queue/id+3", "traffic-class+tc99"},
			},
			startPath: "/policy/qos/name/shaper/profile/queue",
			expResult: true,
		},
		{
			name: "Local profile against multiple profiles - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "queue/id+3",
					"traffic-class+tc99"},
				{"policy", "qos", "profile/name+prof2", QUEUE_ID_1,
					"traffic-class+tc99"},
				{"policy", "qos", "profile/name+prof3", "queue/id+3",
					"traffic-class+tc99"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+3", "traffic-class+tc99"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "queue/id+3", "traffic-class+tc99"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profC", "queue/id+3", "traffic-class+tc99"},
			},
			startPath: "/policy/qos/name/shaper/profile/queue",
			expResult: false,
		},
		{
			name: "Local profiles with multiple queues - PASS",
			config: []xutils.PathType{
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+3", "traffic-class+tc99"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "queue/id+4", "traffic-class+tc98"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "queue/id+3", "traffic-class+tc99"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "queue/id+4", "traffic-class+tc98"},
			},
			startPath: "/policy/qos/name/shaper/profile/queue",
			expResult: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testTree := xpathtest.CreateTree(t, test.config)

			testNode := testTree.FindFirstNode(
				xutils.NewPathType(test.startPath))
			ns := xpath.NewNodesetDatum([]xutils.XpathNode{testNode})

			actResult :=
				verifyQueueIdAndTrafficClass([]xpath.Datum{ns}).Boolean(
					"(unused value)")
			if test.expResult != actResult {
				t.Fatalf("Unexpected result for %s: exp %t, got %t\n",
					test.name, test.expResult, actResult)
			}
		})
	}
}

func TestMapMatch(t *testing.T) {

	tests := []qosProfileTestSpec{
		{
			name: "Global profile match 'group-name' and 'to' - PASS",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", "to+4"},
			},
			startPath: "/policy/qos/profile/map/dscp-group",
			expResult: true,
		},
		{
			name: "Global profile, mismatched 'to' - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+high", TO_4},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", TO_3},
			},
			startPath: "/policy/qos/profile/map/dscp-group",
			expResult: false,
		},
		{
			name: "Global profile, mismatched 'group-name' - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "map",
					DSCP_GRP_HIGH, "to+5"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					DSCP_GRP_LOW, "to+5"},
			},
			startPath: "/policy/qos/profile/map/dscp-group",
			expResult: false,
		},
		{
			name: "Local profile match 'group-name' and 'to' - PASS",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", "to+4"},
			},
			startPath: "/policy/qos/name/shaper/profile/map/dscp-group",
			expResult: true,
		},
		{
			name: "Local profile, mismatched 'to' - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+high", TO_4},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", TO_3},
			},
			startPath: "/policy/qos/name/shaper/profile/map/dscp-group",
			expResult: false,
		},
		{
			name: "Local profile, mismatched 'group-name' - FAIL",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "map",
					DSCP_GRP_HIGH, "to+5"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					DSCP_GRP_LOW, "to+5"},
			},
			startPath: "/policy/qos/name/shaper/profile/map/dscp-group",
			expResult: false,
		},
		{
			name: "Local profile match multiple groups - PASS",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+low", "to+3"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+low", "to+3"},
			},
			startPath: "/policy/qos/name/shaper/profile/map/dscp-group",
			expResult: true,
		},
		{
			name: "Local profile match missing entry - FAIL",
			config: []xutils.PathType{
				// Order matters as we need to pick up the 'right' entry
				// for startPath
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+low", "to+3"},
				// [START PATH]
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+low", "to+3"},
			},
			startPath: "/policy/qos/name/shaper/profile/map/dscp-group",
			expResult: false,
		},
		{
			name: "Local profile mismatch multiple groups - FAIL",
			config: []xutils.PathType{
				// Order matters as we need to pick up the 'right' entry
				// for startPath
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+high", TO_4},
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+low", "to+3"},
				// [START PATH]
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", TO_3},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+low", "to+3"},
			},
			startPath: "/policy/qos/name/shaper/profile/map/dscp-group",
			expResult: false,
		},
		{
			name: "Global profiles with multiple dscp-groups - PASS",
			config: []xutils.PathType{
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+low", "to+3"},
				{"policy", "qos", "profile/name+prof2", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "profile/name+prof2", "map",
					"dscp-group/group-name+low", "to+3"},
			},
			startPath: "/policy/qos/profile/map/dscp-group",
			expResult: true,
		},
		{
			name: "Global profile match multiple profiles - PASS",
			config: []xutils.PathType{
				// Global profile 1 [START PATH]
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+med", "to+5"},
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+low", "to+3"},
				// Global profile 2
				{"policy", "qos", "profile/name+prof2", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "profile/name+prof2", "map",
					"dscp-group/group-name+med", "to+5"},
				{"policy", "qos", "profile/name+prof2", "map",
					"dscp-group/group-name+low", "to+3"},
				// Local profile A
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+med", "to+5"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+low", "to+3"},
				// Local profile B
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "map",
					"dscp-group/group-name+med", "to+5"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "map",
					"dscp-group/group-name+low", "to+3"},
			},
			startPath: "/policy/qos/profile/map/dscp-group",
			expResult: true,
		},
		{
			name: "Global profile mismatch multiple profiles - FAIL",
			config: []xutils.PathType{
				// Order matters as we need to pick up the 'right' entry
				// for startPath
				// Global profile 1 [START PATH]
				{"policy", "qos", "profile/name+prof1", "map",
					DSCP_GRP_MED, "to+4"},
				{"policy", "qos", "profile/name+prof1", "map",
					"dscp-group/group-name+low", "to+3"},
				// Global profile 2
				{"policy", "qos", "profile/name+prof2", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "profile/name+prof2", "map",
					"dscp-group/group-name+low", "to+3"},
				// Local profile A
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profA", "map",
					"dscp-group/group-name+low", "to+3"},
				// Local profile B
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "map",
					"dscp-group/group-name+high", "to+4"},
				{"policy", "qos", "name/name+pol1", "shaper",
					"profile/name+profB", "map",
					"dscp-group/group-name+low", "to+3"},
			},
			startPath: "/policy/qos/profile/map/dscp-group",
			expResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testTree := xpathtest.CreateTree(t, test.config)

			testNode := testTree.FindFirstNode(
				xutils.NewPathType(test.startPath))
			ns := xpath.NewNodesetDatum([]xutils.XpathNode{testNode})

			actResult :=
				verifyDscpGroupToQueueMappings([]xpath.Datum{ns}).Boolean(
					"(unused value)")
			if test.expResult != actResult {
				t.Fatalf("Unexpected result for %s: exp %t, got %t\n",
					test.name, test.expResult, actResult)
			}
		})
	}
}
