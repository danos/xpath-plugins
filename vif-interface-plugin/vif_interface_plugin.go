// Copyright (c) 2021, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/danos/xpath-plugins/common"
	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

var RegistrationData = []xpath.CustomFunctionInfo{
	{
		Name:          "parent-interface-string-length",
		FnPtr:         parentInterfaceStringLength,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsNumber,
		DefaultRetVal: xpath.NewNumDatum(0),
	},
	{
		Name:          "validate-vif-vlan-settings",
		FnPtr:         validateVifVlanSettings,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
	{
		Name:          "check-vlan-values-do-not-conflict",
		FnPtr:         checkVlanValuesDoNotConflict,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
	{
		Name:          "check-implicit-vlan-id-unique",
		FnPtr:         checkImplicitVlanIdUnique,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
}

// Filters used to find required nodes. Values never change, so create once
// and reuse.
var ifnameFilter = common.GetFilter("ifname")
var innerVlanFilter = common.GetFilter("inner-vlan")
var nameFilter = common.GetFilter("name")
var tagnodeFilter = common.GetFilter("tagnode")
var vifFilter = common.GetFilter("vif")
var vlanFilter = common.GetFilter("vlan")

// parentInterfacesStringLength
//
// Full must statement:
//
//   must "(string-length(../*[local-name(.) = 'tagnode' or local-name(.) = 'ifname'
//         or local-name(.) = 'name']) + string-length(tagnode)) < 15";
//
// Implemented with plugin as:
//
//   configd:must "parent-interface-string-length(.) + string-length(tagnode) < 15"
//
func parentInterfaceStringLength(
	args []xpath.Datum,
) (retNum xpath.Datum) {

	// Function has flexibility to be applied to '.' or any other node,
	// rather than just the current node, but should only be applied to
	// a single node.  So, if nodeset is empty or has multiple entries,
	// return false.
	ns0 := args[0].Nodeset("parent-interface-string-length()")
	if len(ns0) != 1 {
		return xpath.NewNumDatum(0)
	}
	srcNode := ns0[0]
	parent := srcNode.XParent()

	tagnode, ifname, name, ok := "", "", "", true

	if tagnode, ok = common.GetSingleChildValue(parent, tagnodeFilter); ok {
		return xpath.NewNumDatum(float64(len(tagnode)))
	}
	if ifname, ok = common.GetSingleChildValue(parent, ifnameFilter); ok {
		return xpath.NewNumDatum(float64(len(ifname)))
	}
	if name, ok = common.GetSingleChildValue(parent, nameFilter); ok {
		return xpath.NewNumDatum(float64(len(name)))
	}

	return xpath.NewNumDatum(0)
}

// validateVifVlanSettings
//
// Full must statements, to be run on each VIF on the interface node provided.
//
//   must "not(vlan) or (count(../vif[vlan=current()/vlan]) = 1) or " +
//	      "(count(../vif[vlan=current()/vlan]/inner-vlan) = " +
//	      "count(../vif[vlan=current()/vlan]))";
//
//   AND
//
//   must "vlan or inner-vlan or not(../vif[vlan=current()/tagnode])";
//
// The expensive part of this follows the first 'or', as we need to loop
// through ALL configured VIFs multiple times. We reimplement as this:
//
//   configd:must "validate-vif-vlan-settings(.)";
//
type vifData struct {
	vif       string
	vlan      string
	innerVlan string
}

func getVifData(intfNode xutils.XpathNode) map[string]vifData {

	vifNodes := common.GetDescendantNodesFromSingleNode(
		intfNode, []xutils.XFilter{vifFilter})

	vifs := make(map[string]vifData, len(vifNodes))
	vifId, vlan, innerVlan := "", "", ""
	for _, vifNode := range vifNodes {
		vifId, _ = common.GetSingleChildValue(vifNode, tagnodeFilter)
		vlan, _ = common.GetSingleChildValue(vifNode, vlanFilter)
		innerVlan, _ = common.GetSingleChildValue(vifNode, innerVlanFilter)

		vifs[vifId] = vifData{
			vif:       vifId,
			vlan:      vlan,
			innerVlan: innerVlan,
		}
	}

	return vifs
}

func validateVifVlanSettings(
	args []xpath.Datum,
) (retBool xpath.Datum) {

	// Function has flexibility to be applied to '.' or any other node,
	// rather than just the current node, but should only be applied to
	// a single node.  So, if nodeset is empty or has multiple entries,
	// return false.
	ns0 := args[0].Nodeset("validate-vif-vlan-settings()")
	if len(ns0) != 1 {
		return xpath.NewBoolDatum(false)
	}
	intfNode := ns0[0]
	vifs := getVifData(intfNode)

	for _, vif := range vifs {
		if !checkVlanValuesDoNotConflictInternal(vif, vifs) {
			return xpath.NewBoolDatum(false)
		}

		if !checkImplicitVlanIdUniqueInternal(vif, vifs) {
			return xpath.NewBoolDatum(false)
		}
	}

	return xpath.NewBoolDatum(true)

	// Get VLAN ID and InnerVlan ID for each VIF
}

// checkVlanValuesDoNotConflictInternal
//
// Full must statement:
//
//   must "not(vlan) or (count(../vif[vlan=current()/vlan]) = 1) or " +
//	      "(count(../vif[vlan=current()/vlan]/inner-vlan) = " +
//	      "count(../vif[vlan=current()/vlan]))";
//
// The expensive part of this follows the first 'or', as we need to loop
// through ALL configured VIFs multiple times. We reimplement as this:
//
//   configd:must "check-vlan-values-do-not-conflict(.)";
//
func checkVlanValuesDoNotConflictInternal(
	currentVif vifData,
	vifs map[string]vifData,
) bool {

	// 'not(vlan)'
	currentVifVlanId := currentVif.vlan
	if currentVifVlanId == "" {
		return true
	}

	// 'or (count(../vif[vlan=current()/vlan]) = 1)'
	// 'or (count(../vif[vlan=current()/vlan]/inner-vlan) =
	//	    count(../vif[vlan=current()/vlan]))'
	matchingVlanCount := 0
	innerVlanCount := 0
	for _, vif := range vifs {
		if vif.vlan == currentVifVlanId {
			matchingVlanCount++
		}
		if vif.innerVlan != "" {
			innerVlanCount++
		}
	}
	if matchingVlanCount == 1 {
		return true
	}
	if matchingVlanCount == innerVlanCount {
		return true
	}

	return false
}

// checkVlanValuesDoNotConflict
//
// Full must statement:
//
//   must "not(vlan) or (count(../vif[vlan=current()/vlan]) = 1) or " +
//	      "(count(../vif[vlan=current()/vlan]/inner-vlan) = " +
//	      "count(../vif[vlan=current()/vlan]))";
//
// The expensive part of this follows the first 'or', as we need to loop
// through ALL configured VIFs multiple times. We reimplement as this:
//
//   configd:must "check-vlan-values-do-not-conflict(.)";
//
func checkVlanValuesDoNotConflict(
	args []xpath.Datum,
) (retBool xpath.Datum) {

	// Function has flexibility to be applied to '.' or any other node,
	// rather than just the current node, but should only be applied to
	// a single node.  So, if nodeset is empty or has multiple entries,
	// return false.
	ns0 := args[0].Nodeset("check-vlan-values-do-not-conflict()")
	if len(ns0) != 1 {
		return xpath.NewBoolDatum(false)
	}
	vifNode := ns0[0]

	// 'not(vlan)'
	// This check is redone by checkVlanValuesDoNotConflictInternal BUT by
	// doing it here we avoid a potentially costly call to getVifData().
	if _, ok := common.GetSingleChildValue(
		vifNode, vlanFilter); !ok {
		return xpath.NewBoolDatum(true)
	}

	// 'or (count(../vif[vlan=current()/vlan]) = 1)'
	// 'or (count(../vif[vlan=current()/vlan]/inner-vlan) =
	//	    count(../vif[vlan=current()/vlan]))'	intfNode := vifNode.XParent()
	vifs := getVifData(vifNode.XParent())

	currentVifId, ok := "", false
	if currentVifId, ok = common.GetSingleChildValue(
		vifNode, tagnodeFilter); !ok {
		return xpath.NewBoolDatum(true)
	}

	if checkVlanValuesDoNotConflictInternal(vifs[currentVifId], vifs) {
		return xpath.NewBoolDatum(true)
	}
	return xpath.NewBoolDatum(false)
}

// checkImplicitVlanIdUniqueInternal
//
// Full must statement:
//
//   must "vlan or inner-vlan or not(../vif[vlan=current()/tagnode])";
//
// This is checking that a VIF with an implicit VLAN Id (equal to the VIF ID
// when no explicit vlan or inner-vlan value is set) is not the explicit VLAN
// ID of any other VIF on this interface. New configd:must is:
//
//   configd:must "check-implicit-vlan-id-unique(.)"
//
func checkImplicitVlanIdUniqueInternal(
	currentVif vifData,
	vifs map[string]vifData,
) bool {

	// 'vlan'
	if currentVif.vlan != "" {
		return true
	}

	// 'or inner-vlan'
	if currentVif.innerVlan != "" {
		return true
	}

	// 'or not(../vif[vlan=current()/tagnode])'
	currentVifId := currentVif.vif
	for _, vif := range vifs {
		if vif.vlan == currentVifId {
			return false
		}
	}

	return true
}

// checkImplicitVlanIdUnique
//
// Full must statement:
//
//   must "vlan or inner-vlan or not(../vif[vlan=current()/tagnode])";
//
// This is checking that a VIF with an implicit VLAN Id (equal to the VIF ID
// when no explicit vlan or inner-vlan value is set) is not the explicit VLAN
// ID of any other VIF on this interface. New configd:must is:
//
//   configd:must "check-implicit-vlan-id-unique(.)"
//
func checkImplicitVlanIdUnique(
	args []xpath.Datum,
) (retBool xpath.Datum) {

	// Function has flexibility to be applied to '.' or any other node,
	// rather than just the current node, but should only be applied to
	// a single node.  So, if nodeset is empty or has multiple entries,
	// return false.
	ns0 := args[0].Nodeset("check-implicit-vlan-id-unique()")
	if len(ns0) != 1 {
		return xpath.NewBoolDatum(false)
	}
	vifNode := ns0[0]

	// 'vlan'
	if _, ok := common.GetSingleChildValue(vifNode, vlanFilter); ok {
		return xpath.NewBoolDatum(true)
	}

	// 'or inner-vlan'
	if _, ok := common.GetSingleChildValue(vifNode, innerVlanFilter); ok {
		return xpath.NewBoolDatum(true)
	}

	// 'or not(../vif[vlan=current()/tagnode])'
	vifs := getVifData(vifNode.XParent())

	currentVifId, ok := "", false
	if currentVifId, ok = common.GetSingleChildValue(
		vifNode, tagnodeFilter); !ok {
		return xpath.NewBoolDatum(true)
	}

	if checkImplicitVlanIdUniqueInternal(vifs[currentVifId], vifs) {
		return xpath.NewBoolDatum(true)
	}
	return xpath.NewBoolDatum(false)
}
