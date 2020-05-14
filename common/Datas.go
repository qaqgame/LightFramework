package common

type ServerModuleInfo struct {
	Id   int     `json:"Id"`
	Name string  `json:"Name"`
	Port int     `json:"Port"`
}

type ConfigUnMarshal struct {
	ServerModuleInfos  []ServerModuleInfo     `json:"ServerModuleInfos"`
}
