// Copyright (c) 2019-2020, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/xml"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

var RegistrationData = []xpath.CustomFunctionInfo{
	{
		Name:          "verify-queue-id-and-traffic-class",
		FnPtr:         verifyQueueIdAndTrafficClass,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
	{
		Name:          "verify-dscp-group-to-queue-mappings",
		FnPtr:         verifyDscpGroupToQueueMappings,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsNodeset},
		RetType:       xpath.TypeIsBool,
		DefaultRetVal: xpath.NewBoolDatum(false),
	},
}

// Filters used to find required nodes. Values never change, so create once
// and reuse.
var idFilter = xutils.NewXFilterConfigOnly(
	xml.Name{Space: "", Local: "id"})
var groupNameFilter = xutils.NewXFilterConfigOnly(
	xml.Name{Space: "", Local: "group-name"})
var toFilter = xutils.NewXFilterConfigOnly(
	xml.Name{Space: "", Local: "to"})
var trafficClassFilter = xutils.NewXFilterConfigOnly(
	xml.Name{Space: "", Local: "traffic-class"})

func getFilter(name string) xutils.XFilter {
	return xutils.NewXFilterConfigOnly(xml.Name{Space: "", Local: name})
}

// getSingleChildValue - return value of child node, if there's only one.
// Otherwise return false (error).  Wraps logic of getting value of child
// node where we expect only a single child.
func getSingleChildValue(node xutils.XpathNode, filter xutils.XFilter,
) (string, bool) {
	childNodes := node.XChildren(filter)
	if childNodes == nil || len(childNodes) > 1 {
		return "", false
	}
	return childNodes[0].XValue(), true
}

// getDescendantNodesFromSingleNode - walk the provided path for the initial
// node to find all descendant nodes that match it.
func getDescendantNodesFromSingleNode(
	node xutils.XpathNode,
	path []string,
) []xutils.XpathNode {

	nodes := []xutils.XpathNode{node}
	return getDescendantNodes(nodes, path)
}

// getDescendantNodes - walk the provided path for all initial nodes to find
// all descendant nodes that match it.
func getDescendantNodes(
	nodes []xutils.XpathNode,
	path []string,
) []xutils.XpathNode {
	if len(path) == 0 {
		return nodes
	}

	curNodes := nodes
	for _, name := range path {
		filteredNodes := []xutils.XpathNode{}
		filter := getFilter(name)
		for _, node := range curNodes {
			filteredNodes = append(filteredNodes, node.XChildren(filter)...)
		}
		curNodes = filteredNodes
	}
	return curNodes
}

// getCountOfChildNodesWithRequiredValues - return the number of child nodes
// that match the required set of {name, value} pairs.
func getCountOfChildNodesWithRequiredValues(
	nodes []xutils.XpathNode,
	filterValueMap map[xutils.XFilter]string,
) int {
	count := 0
	for _, node := range nodes {
		match := true
		for filter, value := range filterValueMap {
			children := node.XChildren(filter)
			if len(children) != 1 {
				continue
			}
			if children[0].XValue() != value {
				match = false
				break
			}
		}
		if match {
			count++
		}
	}
	return count
}

// verifyQueueIdAndTrafficClass
//
// Implements:
//
//  must "/policy:policy/qos:ingress-map or
//        (count(../../../../qos:name/qos:shaper/qos:profile)
//        + count(../../../../qos:profile)
//        = count(../../../../qos:name/qos:shaper/qos:profile/qos:queue[
//            qos:id=current()/qos:id and
//            qos:traffic-class = current()/qos:traffic-class])
//        + count(../../../../qos:profile/qos:queue[
//            qos:id=current()/qos:id and
//            qos:traffic-class = current()/qos:traffic-class]))"
//
// We can rewrite this with absolute paths and no namespaces as:
//
//  must "/policy/ingress-map or (count(/policy/qos/name/shaper/profile)
//        + count(/policy/qos/profile)
//        = count(/policy/qos/name/shaper/profile/queue[
//            id=current()/id and
//            traffic-class = current()/traffic-class])
//        + count(/policy/qos/profile/queue[
//            id=current()/id and
//            traffic-class = current()/traffic-class]))"
//
// The first part of the statement reflects that this restriction does
// not apply if the new ingress-map style of QoS classification is used.
// So we will first check for the presence of any ingress-map and return
// true if any is present.
// Then if we assume we are 'rebasing' our root to /policy/qos, and we
// replace the [a and b] predicate with [a][b], we get:
//
//  must "count(name/shaper/profile) + count(profile)
//        = count(name/shaper/profile/queue
//            [id=current()/id] [traffic-class = current()/traffic-class])
//        + count(profile/queue
//            [id=current()/id] [traffic-class = current()/traffic-class])"
//
func verifyQueueIdAndTrafficClass(
	args []xpath.Datum,
) (retBool xpath.Datum) {

	// Function has flexibility to be applied to '.' or any other node,
	// rather than just the current node, but should only be applied to
	// a single node.  So, if nodeset is empty or has multiple entries,
	// return false.
	ns0 := args[0].Nodeset("verify-queue-id-and-traffic-class()")
	if len(ns0) != 1 {
		return xpath.NewBoolDatum(false)
	}
	srcNode := ns0[0]
	root := srcNode.XRoot()

	// Return true if we have any ingress-maps configured
	mapNodes := getDescendantNodesFromSingleNode(
		root, []string{"policy", "ingress-map"})
	if mapNodes != nil && len(mapNodes) != 0 {
		return xpath.NewBoolDatum(true)
	}

	// Get current()/id and current()/traffic-class
	id, trafficClass, ok := "", "", true
	if id, ok = getSingleChildValue(srcNode, idFilter); !ok {
		return xpath.NewBoolDatum(false)
	}
	if trafficClass, ok = getSingleChildValue(
		srcNode, trafficClassFilter); !ok {
		return xpath.NewBoolDatum(false)
	}

	var reqValues = map[xutils.XFilter]string{
		getFilter("id"):            id,
		getFilter("traffic-class"): trafficClass,
	}

	// Now look at the entries that need to match id/traffic-class.
	qosNodes := getDescendantNodesFromSingleNode(
		root, []string{"policy", "qos"})
	if qosNodes == nil || len(qosNodes) > 1 {
		return xpath.NewBoolDatum(false)
	}
	qosNode := qosNodes[0]

	// Get local profiles, then queue children with matching required values.
	localProfileNodes := getDescendantNodesFromSingleNode(qosNode,
		[]string{"name", "shaper", "profile"})
	localProfileQueueNodes := getDescendantNodes(
		localProfileNodes, []string{"queue"})
	matchingLPQNodeCount := getCountOfChildNodesWithRequiredValues(
		localProfileQueueNodes, reqValues)

	// Get global profiles, then queue children with matching required values.
	globalProfileNodes := getDescendantNodesFromSingleNode(qosNode,
		[]string{"profile"})
	globalProfileQueueNodes := getDescendantNodes(
		globalProfileNodes, []string{"queue"})
	matchingGPQNodeCount := getCountOfChildNodesWithRequiredValues(
		globalProfileQueueNodes, reqValues)

	// count(name/shaper/profile) + count(profile) =
	// count(n/s/p/q[match id and tc]) + count(profile/queue[match id and tc])
	if (len(localProfileNodes) + len(globalProfileNodes)) ==
		(matchingLPQNodeCount + matchingGPQNodeCount) {
		return xpath.NewBoolDatum(true)
	}
	return xpath.NewBoolDatum(false)
}

// verifyDscpGroupToQueueMappings
//
// This must statement is similar, but with subtle differences.  We need
// to ensure the DSCP-group to queue mappings are identical everywhere, so
// on each local and global dscp-group that the must is called on, we check
// that there is an equivalent entry on all local and global profile maps.
//
// NB: namespaces removed, and relative paths converted to absolute, with
//     /policy/qos/ prefix removed as implicit.
//
// must "count(name/shaper/profile/map)
//       + count(profile/map)
//       = count(name/shaper/profile/map/dscp-group
//             [group-name = current()/group-name and to = current()/to])
//       + count(profile/map/dscp-group
//             [group-name = current()/group-name and to = current()/to])"
//
func verifyDscpGroupToQueueMappings(
	args []xpath.Datum,
) (retBool xpath.Datum) {

	// Function has flexibility to be applied to '.' or any other node,
	// rather than just the current node, but should only be applied to
	// a single node.  So, if nodeset is empty or has multiple entries,
	// return false.
	ns0 := args[0].Nodeset("verify-dscp-group-to-queue-mappings")
	if len(ns0) != 1 {
		return xpath.NewBoolDatum(false)
	}

	srcNode := ns0[0]
	root := srcNode.XRoot()

	// Get current()/group-name and current()/to
	groupName, to, ok := "", "", true
	if groupName, ok = getSingleChildValue(srcNode, groupNameFilter); !ok {
		return xpath.NewBoolDatum(false)
	}
	if to, ok = getSingleChildValue(srcNode, toFilter); !ok {
		return xpath.NewBoolDatum(false)
	}

	var reqValues = map[xutils.XFilter]string{
		getFilter("to"):         to,
		getFilter("group-name"): groupName,
	}

	// Local and global profiles live under same root, so get that once.
	qosNodes := getDescendantNodesFromSingleNode(
		root, []string{"policy", "qos"})
	if qosNodes == nil || len(qosNodes) > 1 {
		return xpath.NewBoolDatum(false)
	}
	qosNode := qosNodes[0]

	// Get local maps, and dscp-group children with required values.
	localMapNodes := getDescendantNodesFromSingleNode(qosNode,
		[]string{"name", "shaper", "profile", "map"})
	localMapDscpGroupNodes := getDescendantNodes(
		localMapNodes, []string{"dscp-group"})
	matchingLMDGNodeCount := getCountOfChildNodesWithRequiredValues(
		localMapDscpGroupNodes, reqValues)

	// Get global maps, and dscp-group children with required values.
	globalMapNodes := getDescendantNodesFromSingleNode(qosNode,
		[]string{"profile", "map"})
	globalMapDscpGroupNodes := getDescendantNodes(
		globalMapNodes, []string{"dscp-group"})
	matchingGMDGNodeCount := getCountOfChildNodesWithRequiredValues(
		globalMapDscpGroupNodes, reqValues)

	// count(name/shaper/profile/map) + count(profile/map) =
	// count(n/s/p/m/dscp-group[match group-name and to]) +
	// count(profile/map/dscp-group[match group-name and to])
	if (len(localMapNodes) + len(globalMapNodes)) ==
		(matchingLMDGNodeCount + matchingGMDGNodeCount) {
		return xpath.NewBoolDatum(true)
	}
	return xpath.NewBoolDatum(false)
}
