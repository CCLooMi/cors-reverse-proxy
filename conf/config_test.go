package conf

import (
	"testing"
)

func TestFileServerConfig(t *testing.T) {
	t.Log(Cfg)
	hostConf := Cfg.HostConf["test.com"]
	t.Log(hostConf)
	hds := hostConf.Header
	t.Log(hds)
}
