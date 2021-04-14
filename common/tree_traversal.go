// Copyright (c) 2019-2021, AT&T Intellectual Property Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"encoding/xml"

	"github.com/danos/yang/xpath/xutils"
)

// GetFilter
func GetFilter(name string) xutils.XFilter {
	return xutils.NewXFilterConfigOnly(xml.Name{Space: "", Local: name})
}

// GetSingleChildValue - return value of child node, if there's only one.
// Otherwise return false (error).  Wraps logic of getting value of child
// node where we expect only a single child.
func GetSingleChildValue(node xutils.XpathNode, filter xutils.XFilter,
) (string, bool) {
	childNodes := node.XChildren(filter)
	if childNodes == nil || len(childNodes) > 1 {
		return "", false
	}
	return childNodes[0].XValue(), true
}

// GetDescendantNodesFromSingleNode - walk the provided path for the initial
// node to find all descendant nodes that match it.
func GetDescendantNodesFromSingleNode(
	node xutils.XpathNode,
	path []string,
) []xutils.XpathNode {

	nodes := []xutils.XpathNode{node}
	return GetDescendantNodes(nodes, path)
}

// GetDescendantNodes - walk the provided path for all initial nodes to find
// all descendant nodes that match it.
func GetDescendantNodes(
	nodes []xutils.XpathNode,
	path []string,
) []xutils.XpathNode {
	if len(path) == 0 {
		return nodes
	}

	curNodes := nodes
	for _, name := range path {
		filteredNodes := []xutils.XpathNode{}
		filter := GetFilter(name)
		for _, node := range curNodes {
			filteredNodes = append(filteredNodes, node.XChildren(filter)...)
		}
		curNodes = filteredNodes
	}
	return curNodes
}

// GetCountOfChildNodesWithRequiredValues - return the number of child nodes
// that match the required set of {name, value} pairs.
func GetCountOfChildNodesWithRequiredValues(
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
