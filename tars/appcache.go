package tars

import "github.com/TarsCloud/TarsGo/tars/protocol/res/endpointf"

var (
	appCache AppCache
)

type AppCache struct {
	TarsVersion string
	ModifyTime  string
	LogLevel    string
	ObjCaches   []ObjCache
}

type ObjCache struct {
	Name    string
	Locator string

	Endpoints         []endpointf.EndpointF
	InactiveEndpoints []endpointf.EndpointF
}

func GetAppCache() AppCache {
	return appCache
}
