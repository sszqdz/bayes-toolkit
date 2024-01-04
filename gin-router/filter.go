// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ginrouter

import (
	"github.com/gin-gonic/gin"
)

type filter struct {
	filterMap     map[string]map[string]struct{} // map[fullPath][method]
	funcName      string
	originHandler gin.HandlerFunc
}

func (f *filter) handler(gtx *gin.Context) {
	if mm, ok := f.filterMap[gtx.FullPath()]; ok {
		if _, ok := mm[gtx.Request.Method]; ok {
			gtx.Next()
			return
		}
	}

	f.originHandler(gtx)
}

func (f *filter) addRoute(httpMethod, fullPath string) {
	if _, ok := f.filterMap[fullPath]; !ok {
		f.filterMap[fullPath] = make(map[string]struct{})
	}
	f.filterMap[fullPath][httpMethod] = struct{}{}
}

func newFilter(handler gin.HandlerFunc) *filter {
	return &filter{
		filterMap:     make(map[string]map[string]struct{}, 0),
		funcName:      nameOfFunction(handler),
		originHandler: handler,
	}
}

func filterWrap(handlers []gin.HandlerFunc) ([]*filter, []gin.HandlerFunc) {
	length := len(handlers)
	filters := make([]*filter, 0, length)
	filtersHandlers := make([]gin.HandlerFunc, 0, length)

	for _, h := range handlers {
		f := newFilter(h)
		filters = append(filters, f)
		filtersHandlers = append(filtersHandlers, f.handler)
	}

	return filters, filtersHandlers
}
