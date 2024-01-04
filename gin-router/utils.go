// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ginrouter

import (
	"path"
	"reflect"
	"runtime"

	"github.com/gin-gonic/gin"
)

func nameOfHandlers(handlers gin.HandlersChain) []string {
	size := len(handlers)
	names := make([]string, 0, size)
	if size == 0 {
		return names
	}

	for _, h := range handlers {
		names = append(names, nameOfFunction(h))
	}

	return names
}

// region origin

func lastChar(str string) uint8 {
	if str == "" {
		panic("The length of the string can't be 0")
	}
	return str[len(str)-1]
}

func nameOfFunction(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func joinPaths(absolutePath, relativePath string) string {
	if relativePath == "" {
		return absolutePath
	}

	finalPath := path.Join(absolutePath, relativePath)
	if lastChar(relativePath) == '/' && lastChar(finalPath) != '/' {
		return finalPath + "/"
	}
	return finalPath
}

// endregion
