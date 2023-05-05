package tars

import "github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"

type AppCache struct {
	TarsVersion string
	ModifyTime  string
	LogLevel    string
	ObjCaches   []ObjCache
}

type ObjCache struct {
	Name    string
	SetID   string
	Locator string

	Endpoints         []endpointf.EndpointF
	InactiveEndpoints []endpointf.EndpointF
}

func GetAppCache() AppCache {
	return defaultApp.AppCache()
}

func (a *application) AppCache() AppCache {
	return a.appCache
}
