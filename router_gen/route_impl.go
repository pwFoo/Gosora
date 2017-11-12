package main

type RouteImpl struct {
	Name      string
	Path      string
	Vars      []string
	RunBefore []Runnable

	Parent *RouteGroup
}

type Runnable struct {
	Contents string
	Literal  bool
}

func addRoute(route *RouteImpl) {
	routeList = append(routeList, route)
}

func (route *RouteImpl) Before(items ...string) *RouteImpl {
	for _, item := range items {
		route.RunBefore = append(route.RunBefore, Runnable{item, false})
	}
	return route
}

func (route *RouteImpl) LitBefore(items ...string) *RouteImpl {
	for _, item := range items {
		route.RunBefore = append(route.RunBefore, Runnable{item, true})
	}
	return route
}

func (route *RouteImpl) hasBefore(items ...string) bool {
	for _, item := range items {
		if route.hasBeforeItem(item) {
			return true
		}
	}
	return false
}

func (route *RouteImpl) hasBeforeItem(item string) bool {
	for _, before := range route.RunBefore {
		if before.Contents == item {
			return true
		}
	}
	return false
}

func addRouteGroup(routeGroup *RouteGroup) {
	routeGroups = append(routeGroups, routeGroup)
}

func blankRoute() *RouteImpl {
	return &RouteImpl{"", "", []string{}, []Runnable{}, nil}
}

func route(fname string, path string, args ...string) *RouteImpl {
	return &RouteImpl{fname, path, args, []Runnable{}, nil}
}

func View(fname string, path string, args ...string) *RouteImpl {
	return route(fname, path, args...)
}

func MemberView(fname string, path string, args ...string) *RouteImpl {
	route := route(fname, path, args...)
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}

func ModView(fname string, path string, args ...string) *RouteImpl {
	route := route(fname, path, args...)
	if !route.hasBefore("AdminOnly") {
		route.Before("SuperModOnly")
	}
	return route
}

func Action(fname string, path string, args ...string) *RouteImpl {
	route := route(fname, path, args...)
	route.Before("NoSessionMismatch")
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}

func AnonAction(fname string, path string, args ...string) *RouteImpl {
	return route(fname, path, args...).Before("ParseForm")
}

func UploadAction(fname string, path string, args ...string) *RouteImpl {
	route := route(fname, path, args...)
	if !route.hasBefore("SuperModOnly", "AdminOnly") {
		route.Before("MemberOnly")
	}
	return route
}
