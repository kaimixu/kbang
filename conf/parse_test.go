package conf

import (
	"testing"
)

func TestBadSyntax(t *testing.T) {
	conf := NewConf()
	err := conf.LoadFile("kbang-bad-item.conf")
	if err == nil {
		t.Error("TestBadSyntax: no detection bad item")
	}

	err = conf.LoadFile("kbang-bad-section.conf")
	if err == nil {
		t.Error("TestBadSection: no detection bad section")
	}
}

func TestGoodConf(t *testing.T) {
	type Section struct {
		Ip 	string	`ip`
		Weight 	int	`weight`
	}
	type Config struct {
		Item		string		`item`
		Request		[2]Section	`request`
	}
	var cfg Config
	conf := NewConf()
	err := conf.LoadFile("kbang.conf")
	if err != nil {
		t.Error("TestGoodConf: load file failed")
	}

	err = conf.Parse(&cfg)
	if err != nil {
		t.Error("TestGoodConf: parse failed")
	}

	if cfg.Item != "" {
		t.Error("TestGoodConf: unkonw item")
	}
	if  cfg.Request[0].Ip != "" || cfg.Request[0].Weight == 0 {
		t.Error("TestGoodConf: unkonw item")
	}
}

