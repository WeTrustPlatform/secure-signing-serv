package main

import (
	"github.com/ethereum/go-ethereum"
	lua "github.com/yuin/gopher-lua"
)

func callToLTable(L *lua.LState, call ethereum.CallMsg) *lua.LTable {
	t := L.NewTable()
	if call.To != nil {
		L.SetField(t, "to", lua.LString(call.To.String()))
	}
	if call.Value != nil {
		L.SetField(t, "value", lua.LString(call.Value.String()))
	}
	if call.Data != nil {
		L.SetField(t, "data", lua.LString(string(call.Data[:])))
	}
	return t
}

func validate(rules string, call ethereum.CallMsg) (bool, error) {
	L := lua.NewState()
	defer L.Close()

	err := L.DoString(rules)
	if err != nil {
		return false, err
	}

	err = L.CallByParam(
		lua.P{
			Fn:      L.GetGlobal("validate"),
			NRet:    1,
			Protect: true,
		},
		callToLTable(L, call),
	)
	if err != nil {
		return false, err
	}

	ret := L.Get(-1)
	L.Pop(1)

	return ret == lua.LTrue, nil
}
