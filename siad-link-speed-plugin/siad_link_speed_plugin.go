// Copyright (c) 2020-2021, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"strconv"
	"strings"

	"github.com/danos/xpath-plugins/common"
	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

const (
	INVALID_INTF_ID = -1
	DP0XE_NAME      = "dp0xe"
)

var RegistrationData = []xpath.CustomFunctionInfo{
	{
		Name:  "verify-siad-link-speed",
		FnPtr: verifySiadLinkSpeed,
		Args: []xpath.DatumTypeChecker{
			xpath.TypeIsNumber,
			xpath.TypeIsNumber,
			xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
}

// Filters used to find required nodes. Values never change, so create once
// and reuse.
var intfFilter = common.GetFilter("interfaces")
var dataplaneFilter = common.GetFilter("dataplane")
var tagnodeFilter = common.GetFilter("tagnode")
var disableFilter = common.GetFilter("disable")
var speedFilter = common.GetFilter("speed")

func getIntfNameAndIdForType(
	intfNode xutils.XpathNode,
	intfPrefix string,
) (string, int, bool) {

	intfName, ok := common.GetSingleChildValue(intfNode, tagnodeFilter)
	if !ok {
		// If we can't get interface name, then it's not one we care about.
		return "", INVALID_INTF_ID, false
	}
	if !strings.HasPrefix(intfName, intfPrefix) {
		return "", INVALID_INTF_ID, false
	}
	if len(intfName) <= len(intfPrefix) {
		return "", INVALID_INTF_ID, false
	}
	intfID, err := strconv.Atoi(intfName[len(intfPrefix):])
	if err != nil {
		return "", INVALID_INTF_ID, false
	}

	return intfName, intfID, true
}

// verifySiadLinkSpeed - check link speed on interface meets requirements.
//
// NB: if we get an internal error, we return true on the basis that it's better
// to allow config that might not be ok rather than force config to fail to parse
// even though it might be ok.  In other words, we only fail if we are sure that
// something is wrong; otherwise we give the benefit of the doubt.
//
// Implements the following must statement (albeit generically for a set of
// dp0xe interfaces in range startIntfID to endIntfID).  Example shown for
// dp0xe20-23
//
// must "(   not((../dp:tagnode = 'dp0xe20') or " +
//              "(../dp:tagnode = 'dp0xe21') or " +
//              "(../dp:tagnode = 'dp0xe22') or " +
//              "(../dp:tagnode = 'dp0xe23'))"
//          "or (../dp:disable)" +
//          "or (current() = 'auto')"
//          "or ((current() = '10g') or (current() = '25g')) " +
//             "and " +
//               "(not(../../dp:dataplane[dp:tagnode = 'dp0xe20']) or " +
//               "../../dp:dataplane[dp:tagnode = 'dp0xe20']/dp:disable or " +
//               "(../../dp:dataplane[dp:tagnode = 'dp0xe20']/dp:speed = current()) or " +
//               "(../../dp:dataplane[dp:tagnode = 'dp0xe20']/dp:speed = 'auto')) " +
//             "and " +
//               "(not(../../dp:dataplane[dp:tagnode = 'dp0xe21']) or " +
//               "../../dp:dataplane[dp:tagnode = 'dp0xe21']/dp:disable or " +
//               "(../../dp:dataplane[dp:tagnode = 'dp0xe21']/dp:speed = current()) or " +
//               "(../../dp:dataplane[dp:tagnode = 'dp0xe21']/dp:speed = 'auto')) " +
//             "and " +
//               "(not(../../dp:dataplane[dp:tagnode = 'dp0xe22']) or " +
//               "../../dp:dataplane[dp:tagnode = 'dp0xe22']/dp:disable or " +
//               "(../../dp:dataplane[dp:tagnode = 'dp0xe22']/dp:speed = current()) or " +
//               "(../../dp:dataplane[dp:tagnode = 'dp0xe22']/dp:speed = 'auto')) " +
//             "and " +
//               "(not(../../dp:dataplane[dp:tagnode = 'dp0xe23']) or " +
//               "../../dp:dataplane[dp:tagnode = 'dp0xe23']/dp:disable or " +
//               "(../../dp:dataplane[dp:tagnode = 'dp0xe23']/dp:speed = current()) or " +
//               "(../../dp:dataplane[dp:tagnode = 'dp0xe23']/dp:speed = 'auto'))
//        )"
func verifySiadLinkSpeed(
	args []xpath.Datum,
) (retBool xpath.Datum) {

	startIntfID := int(args[0].Number("verify-siad-link-speed()"))
	endIntfID := int(args[1].Number("verify-siad-link-speed()"))

	// If only one interface in range
	if endIntfID <= startIntfID {
		return xpath.NewBoolDatum(true)
	}

	// Function has flexibility to be applied to '.' or any other node,
	// rather than just the current node, but should only be applied to
	// a single node.  So, if nodeset is empty or has multiple entries,
	// return false.
	ns0 := args[2].Nodeset("verify-siad-link-speed()")
	if len(ns0) != 1 {
		return xpath.NewBoolDatum(false)
	}
	curSpeedNode := ns0[0]
	curDpEntryNode := curSpeedNode.XParent()

	// Return true if not dp0xe<startIntf> to dp0xe<endIntf> inclusive
	_, curIntfID, ok := getIntfNameAndIdForType(
		curDpEntryNode, DP0XE_NAME)
	if !ok {
		return xpath.NewBoolDatum(true)
	}

	if (curIntfID < startIntfID) || (curIntfID > endIntfID) {
		return xpath.NewBoolDatum(true)
	}

	// Return true if interface is disabled.  'disable' node is type empty so
	// if the boolean (2nd) return value is true, that means node is set ie
	// interface is disabled.
	_, disabled := common.GetSingleChildValue(curDpEntryNode, disableFilter)
	if disabled {
		return xpath.NewBoolDatum(true)
	}

	// Return true if interface speed (current node) is auto
	curSpeed := curSpeedNode.XValue()
	if curSpeed == "auto" {
		return xpath.NewBoolDatum(true)
	}

	// Return false if interface speed is not 10g or 25g
	if curSpeed != "10g" && curSpeed != "25g" {
		return xpath.NewBoolDatum(false)
	}

	// Return false if any other interface in range <startIntf> to <endIntf>
	// is not disabled and speed isn't either auto or same as current node.
	intfNodes := curSpeedNode.XRoot().XChildren(intfFilter, xutils.Sorted)
	if intfNodes == nil || len(intfNodes) > 1 {
		return xpath.NewBoolDatum(true)
	}

	dpNodes := intfNodes[0].XChildren(dataplaneFilter, xutils.Sorted)
	for _, otherIntfNode := range dpNodes {
		_, intfID, ok := getIntfNameAndIdForType(
			otherIntfNode, DP0XE_NAME)
		if !ok {
			continue
		}
		if intfID < startIntfID || intfID > endIntfID || intfID == curIntfID {
			continue
		}

		_, disabled := common.GetSingleChildValue(otherIntfNode, disableFilter)
		if disabled {
			return xpath.NewBoolDatum(true)
		}
		otherIntfSpeed, ok := common.GetSingleChildValue(
			otherIntfNode, speedFilter)
		if !ok {
			// Better to allow if node is missing or we risk an unexpected
			// problem making valid configs invalid.
			return xpath.NewBoolDatum(true)
		}
		if otherIntfSpeed != "auto" && otherIntfSpeed != curSpeed {
			return xpath.NewBoolDatum(false)
		}
	}

	// Return true
	return xpath.NewBoolDatum(true)
}
