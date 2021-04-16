// Copyright (c) 2019,2021, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"strings"

	"github.com/danos/xpath-plugins/common"
	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

var RegistrationData = []xpath.CustomFunctionInfo{
	{
		Name:          "is-interface-leafref",
		FnPtr:         isInterfaceLeafref,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
	{
		Name:          "is-l3-interface-leafref",
		FnPtr:         isL3InterfaceLeafref,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
	{
		Name:          "is-interface-leafref-original",
		FnPtr:         isInterfaceLeafrefOriginal,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
}

// Filters used to find required nodes. Values never change, so create once
// and reuse.
var intfFilter = common.GetFilter("interfaces")
var intfTypesFilter = common.GetFilter("*")
var vifFilter = common.GetFilter("vif")

// isInterfaceLeafref - implementation of is-interface-leafref(<nodeset>)
// Matches any interface, including VIFs
func isInterfaceLeafref(
	args []xpath.Datum,
) (retBool xpath.Datum) {
	return isInterfaceLeafrefInternal(args, []string{})
}

// isL3InterfaceLeafref - implementation of is-l3-interface-leafref(<nodeset>)
// Matches all VIF interfaces, and all base interface types except switch,
// and backplane.  NB: while this matches the intention of the original
// 'interface leafref must', it allows VHOST interfaces which the original
// must statement didn't.
func isL3InterfaceLeafref(
	args []xpath.Datum,
) (retBool xpath.Datum) {
	return isInterfaceLeafrefInternal(args,
		[]string{"switch", "backplane"})
}

// isInterfaceLeafrefOriginal - implementation of
// is-interface-leafref-original(<nodeset>)
// Matches all VIF interfaces, and all base interface types except switch,
// VHOST, and backplane.  This matches the original 'interface leafref' must
// statement used in several YANG files.
func isInterfaceLeafrefOriginal(
	args []xpath.Datum,
) (retBool xpath.Datum) {
	return isInterfaceLeafrefInternal(args,
		[]string{"switch", "vhost", "backplane"})

}

func isInterfaceLeafrefInternal(
	args []xpath.Datum,
	interfaceFilter []string,
) (retBool xpath.Datum) {

	// Function has flexibility to be applied to '.' or any other node,
	// rather than just the current node, but should only be applied to
	// a single node.  So, if nodeset is empty or has multiple entries,
	// return false.
	ns0 := args[0].Nodeset("is-interface-leafref()")
	if len(ns0) != 1 {
		return xpath.NewBoolDatum(false)
	}
	srcNode := ns0[0]
	intfVal, isVIF, vifVal := parseInterfaceName(srcNode.XValue())
	root := srcNode.XRoot()

	intfNodes := root.XChildren(intfFilter, xutils.Sorted)
	if intfNodes == nil || len(intfNodes) > 1 {
		return xpath.NewBoolDatum(false)
	}

	// Now get all interface list entries (we get one big list with all
	// interface list entries, XValue() == key).  This means we don't need
	// to care about tagnode / ifname etc.
	intfListEntries := intfNodes[0].XChildren(intfTypesFilter, xutils.Sorted)
	if intfListEntries == nil {
		return xpath.NewBoolDatum(false)
	}

	for _, intf := range intfListEntries {
		if intf.XValue() != intfVal {
			// Base interface name doesn't match. Next!
			continue
		}
		if !isVIF {
			if ignoreInterface(intf.XName(), interfaceFilter) {
				// Used to ignore specific interface types (but not VIFs
				// on these interfaces).  Typical use is to remove L2
				// interfaces (eg switch and backplane).
				continue
			}

			// Interface base name matches, not looking for VIF. Pass.
			return xpath.NewBoolDatum(true)
		}

		// All VIFs are L3.
		vifs := intf.XChildren(vifFilter, xutils.Sorted)
		for _, vif := range vifs {
			if vif.XValue() == vifVal {
				// Base interface and VIF both match. Pass.
				return xpath.NewBoolDatum(true)
			}
		}

		// Matched on base interface name, so if no matching VIF, we're done.
		return xpath.NewBoolDatum(false)
	}

	return xpath.NewBoolDatum(false)
}

func parseInterfaceName(intfName string) (string, bool, string) {

	intfParts := strings.Split(intfName, ".")
	intfVal := intfParts[0]

	vifVal := ""
	if len(intfParts) == 2 {
		vifVal = intfParts[1]
		return intfVal, true, vifVal
	}

	return intfVal, false, "not-VIF-interface"
}

func ignoreInterface(intfType string, filter []string) bool {
	for _, name := range filter {
		if intfType == name {
			return true
		}
	}

	return false
}
