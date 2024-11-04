package blocks

import (
	"context"

	"github.com/aarzilli/golua/lua"
	. "github.com/kraftwerk28/gost/core"
)

type LuaConfig struct {
	Script string `yaml:"script"`
}

type LuaBlock struct {
	LuaConfig
}

func NewLuaBlock() I3barBlocklet {
	return &LuaBlock{}
}

func (b *LuaBlock) Run(ch UpdateChan, ctx context.Context) {
	s := lua.NewState()
	s.OpenLibs()
	s.DoFile("")
	s.LoadFile(b.Script)
}

func (b *LuaBlock) GetConfig() interface{} {
	return &b.LuaConfig
}

func (b *LuaBlock) Render(cfg *AppConfig) []I3barBlock {
	return []I3barBlock{{
		FullText: "",
	}}
}

func init() {
	RegisterBlocklet("lua", NewLuaBlock)
}
