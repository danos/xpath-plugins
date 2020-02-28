// Copyright (c) 2020, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xpathtest"
	"github.com/danos/yang/xpath/xutils"
)

type siadLinkSpeedTestSpec struct {
	name      string
	config    []xutils.PathType
	startPath string
	expResult bool
}

func TestSiadLinkSpeedValidation(t *testing.T) {

	// Tests designed around dp0xe20-23.  The underlying logic is the same for
	// dp0xe24-27, and the function called is generic.
	tests := []siadLinkSpeedTestSpec{
		{
			name: "Non dp0xe interface - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0p1s2", "speed+auto"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "dp0xe interface with no number - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe", "speed+auto"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "dp0xe interface with invalid number - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xeZZ", "speed+auto"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "dp0xe interface outside range of interest (low) - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe1", "speed+auto"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "dp0xe interface outside range of interest (high) - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe24", "speed+auto"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		// From here on, startPath must be for dp0xe20-23 inclusive as we are
		// now looking at tests only carried out on these interfaces.
		{
			name: "dp0xe interface of interest, disabled - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+auto"},
				{"interfaces", "dataplane/tagnode+dp0xe22", "disable%"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "dp0xe interface of interest, enabled, auto - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+auto"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "dp0xe interface of interest, enabled, not 10g or 25g - FAIL",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+100g"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: false,
		},
		// From here, interface must be in range, enabled, 10g or 25g.
		{
			name: "Non dp0xe interface that would fail if in range - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+25g"},
				{"interfaces", "dataplane/tagnode+dp0p1s1", "speed+10g"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "dp0xe interface that would fail if in range - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+25g"},
				{"interfaces", "dataplane/tagnode+dp0xe24", "speed+10g"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "2 dp0xe interfaces in range, mismatched speeds - FAIL",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+25g"},
				{"interfaces", "dataplane/tagnode+dp0xe23", "speed+10g"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: false,
		},
		{
			name: "2 dp0xe interfaces in range, second auto - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+25g"},
				{"interfaces", "dataplane/tagnode+dp0xe23", "speed+auto"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "3 dp0xe interfaces in range, same valid speed - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe20", "speed+10g"},
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+10g"},
				{"interfaces", "dataplane/tagnode+dp0xe23", "speed+10g"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
		{
			name: "3 dp0xe interfaces in range, one with invalid speed - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe20", "speed+10g"},
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+10g"},
				{"interfaces", "dataplane/tagnode+dp0xe23", "speed+25g"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: false,
		},
		{
			name: "3 dp0xe i/fs in range, one invalid speed but disabled - PASS",
			config: []xutils.PathType{
				{"interfaces", "dataplane/tagnode+dp0xe20", "speed+10g"},
				{"interfaces", "dataplane/tagnode+dp0xe22", "speed+10g"},
				{"interfaces", "dataplane/tagnode+dp0xe22", "disable%"},
				{"interfaces", "dataplane/tagnode+dp0xe23", "speed+25g"},
			},
			startPath: "/interfaces/dataplane/speed",
			expResult: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testTree := xpathtest.CreateTree(t, test.config)

			testNode := testTree.FindFirstNode(
				xutils.NewPathType(test.startPath))
			ns := xpath.NewNodesetDatum([]xutils.XpathNode{testNode})
			startIntfId := xpath.NewNumDatum(20)
			endIntfId := xpath.NewNumDatum(23)

			actResult :=
				verifySiadLinkSpeed(
					[]xpath.Datum{startIntfId, endIntfId, ns}).Boolean(
					"(unused value)")
			if test.expResult != actResult {
				t.Fatalf("Unexpected result for %s: exp %t, got %t\n",
					test.name, test.expResult, actResult)
			}
		})
	}
}
