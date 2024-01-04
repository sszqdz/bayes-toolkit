// Copyright 2024 Moran. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package ginrouter

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var anyMethods = []string{
	http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodHead, http.MethodOptions, http.MethodDelete, http.MethodConnect,
	http.MethodTrace,
}

type iDecoratorRouter interface {
	iDecoratorRoutes
	Group(string, ...gin.HandlerFunc) *decoratorGroup
}

type iDecoratorRoutes interface {
	iExtendRoutes

	Use(...gin.HandlerFunc) iDecoratorRoutes

	Handle(string, string, ...gin.HandlerFunc) iDecoratorRoutes
	Any(string, ...gin.HandlerFunc) iDecoratorRoutes
	GET(string, ...gin.HandlerFunc) iDecoratorRoutes
	POST(string, ...gin.HandlerFunc) iDecoratorRoutes
	DELETE(string, ...gin.HandlerFunc) iDecoratorRoutes
	PATCH(string, ...gin.HandlerFunc) iDecoratorRoutes
	PUT(string, ...gin.HandlerFunc) iDecoratorRoutes
	OPTIONS(string, ...gin.HandlerFunc) iDecoratorRoutes
	HEAD(string, ...gin.HandlerFunc) iDecoratorRoutes
	Match([]string, string, ...gin.HandlerFunc) iDecoratorRoutes

	StaticFile(string, string) iDecoratorRoutes
	StaticFileFS(string, string, http.FileSystem) iDecoratorRoutes
	Static(string, string) iDecoratorRoutes
	StaticFS(string, http.FileSystem) iDecoratorRoutes
}

type iExtendRoutes interface {
	Unuse(handlers ...gin.HandlerFunc) iDecoratorRoutes
	Skip(handlers ...gin.HandlerFunc) iDecoratorRoutes
}

type decoratorGroup struct {
	decoratorEngine *decoratorEngine
	root            bool

	filters       []*filter
	unusedFilters []*filter

	routerGroup *gin.RouterGroup
}

var _ iDecoratorRouter = (*decoratorGroup)(nil)

func (dg *decoratorGroup) Unuse(middleware ...gin.HandlerFunc) iDecoratorRoutes {
	for _, funcName := range nameOfHandlers(middleware) {
		for _, f := range dg.filters {
			if f.funcName == funcName {
				dg.unusedFilters = append(dg.unusedFilters, f)
			}
		}
	}

	return dg.returnObj()
}

func (dg *decoratorGroup) Skip(middleware ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.Group("").Unuse(middleware...)
}

func (dg *decoratorGroup) combineFilters(filters []*filter) []*filter {
	finalSize := len(dg.filters) + len(filters)
	mergedSlice := make([]*filter, finalSize)
	copy(mergedSlice, dg.filters)
	copy(mergedSlice[len(dg.filters):], filters)

	return mergedSlice
}

func (dg *decoratorGroup) combineUnusedFilters() []*filter {
	finalSize := len(dg.unusedFilters)
	mergedSlice := make([]*filter, finalSize)
	copy(mergedSlice, dg.unusedFilters)

	return mergedSlice
}

// region origin

func (dg *decoratorGroup) Use(middleware ...gin.HandlerFunc) iDecoratorRoutes {
	filters, filtersHandlers := filterWrap(middleware)
	dg.filters = append(dg.filters, filters...)
	dg.routerGroup.Use(filtersHandlers...)

	return dg.returnObj()
}

func (dg *decoratorGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *decoratorGroup {
	filters, filtersHandlers := filterWrap(handlers)

	return &decoratorGroup{
		decoratorEngine: dg.decoratorEngine,
		filters:         dg.combineFilters(filters),
		unusedFilters:   dg.combineUnusedFilters(),
		routerGroup:     dg.routerGroup.Group(relativePath, filtersHandlers...),
	}
}

func (dg *decoratorGroup) BasePath() string {
	return dg.routerGroup.BasePath()
}

func (dg *decoratorGroup) handle(httpMethod, relativePath string, handlers gin.HandlersChain) iDecoratorRoutes {
	for _, f := range dg.unusedFilters {
		f.addRoute(httpMethod, dg.calculateAbsolutePath(relativePath))
	}
	dg.routerGroup.Handle(httpMethod, relativePath, handlers...)

	return dg.returnObj()
}

func (dg *decoratorGroup) Handle(httpMethod, relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.handle(http.MethodGet, relativePath, handlers)
}

func (dg *decoratorGroup) POST(relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.handle(http.MethodPost, relativePath, handlers)
}

func (dg *decoratorGroup) GET(relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.handle(http.MethodGet, relativePath, handlers)
}

func (dg *decoratorGroup) DELETE(relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.handle(http.MethodDelete, relativePath, handlers)
}

func (dg *decoratorGroup) PATCH(relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.handle(http.MethodPatch, relativePath, handlers)
}

func (dg *decoratorGroup) PUT(relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.handle(http.MethodPut, relativePath, handlers)
}

func (dg *decoratorGroup) OPTIONS(relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.handle(http.MethodOptions, relativePath, handlers)
}

func (dg *decoratorGroup) HEAD(relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	return dg.handle(http.MethodHead, relativePath, handlers)
}

func (dg *decoratorGroup) Any(relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	for _, method := range anyMethods {
		dg.handle(method, relativePath, handlers)
	}

	return dg.returnObj()
}

func (dg *decoratorGroup) Match(methods []string, relativePath string, handlers ...gin.HandlerFunc) iDecoratorRoutes {
	for _, method := range methods {
		dg.handle(method, relativePath, handlers)
	}

	return dg.returnObj()
}

func (dg *decoratorGroup) StaticFile(relativePath, filepath string) iDecoratorRoutes {
	return dg.staticFileHandler(relativePath, func(gtx *gin.Context) {
		gtx.File(filepath)
	})
}

func (dg *decoratorGroup) StaticFileFS(relativePath, filepath string, fs http.FileSystem) iDecoratorRoutes {
	return dg.staticFileHandler(relativePath, func(gtx *gin.Context) {
		gtx.FileFromFS(filepath, fs)
	})
}

func (dg *decoratorGroup) staticFileHandler(relativePath string, handler gin.HandlerFunc) iDecoratorRoutes {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static file")
	}
	dg.GET(relativePath, handler)
	dg.HEAD(relativePath, handler)
	return dg.returnObj()
}

// !!! Unable to proxy
func (dg *decoratorGroup) Static(relativePath, root string) iDecoratorRoutes {
	return dg.StaticFS(relativePath, gin.Dir(root, false))
}

// !!! Unable to proxy
func (dg *decoratorGroup) StaticFS(relativePath string, fs http.FileSystem) iDecoratorRoutes {
	panic("URL parameters can not be used when serving a static folder")
}

func (dg *decoratorGroup) calculateAbsolutePath(relativePath string) string {
	return joinPaths(dg.BasePath(), relativePath)
}

func (dg *decoratorGroup) returnObj() iDecoratorRoutes {
	if dg.root {
		return dg.decoratorEngine
	}
	return dg
}

// endregion
