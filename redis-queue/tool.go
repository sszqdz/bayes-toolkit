// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package redisqueue

import "strings"

const (
	keyStr = "key"
	valStr = "val"
)

func extractValues(values map[string]interface{}) (key, val string) {
	if values == nil {
		return
	}
	if v, ok := values[keyStr].(string); ok {
		key = v
	}
	if v, ok := values[valStr].(string); ok {
		val = v
	}

	return
}

func extractStreamNames(streams []string) []string {
	streamLen := len(streams) / 2
	return streams[:streamLen]
}

func deadHashMapName(stream, group string) string {
	return stream + "-dead-" + group
}

const busyGroupStr = "BUSYGROUP"

func isBusyGroupErr(err error) bool {
	return strings.HasPrefix(err.Error(), busyGroupStr)
}
