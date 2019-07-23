package main

import (
	"github.com/ethereum/go-ethereum/core/types"
	lua "github.com/yuin/gopher-lua"
)

func txToLTable(L *lua.LState, tx *types.Transaction) *lua.LTable {
	t := L.NewTable()
	if tx.To() != nil {
		L.SetField(t, "to", lua.LString(tx.To().String()))
	}
	if tx.Value() != nil {
		L.SetField(t, "value", lua.LString(tx.Value().String()))
	}
	if tx.Data() != nil {
		L.SetField(t, "data", lua.LString(string(tx.Data())))
	}
	return t
}

func validate(rules string, tx *types.Transaction) (bool, error) {
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
		txToLTable(L, tx),
	)
	if err != nil {
		return false, err
	}

	ret := L.Get(-1)
	L.Pop(1)

	return ret == lua.LTrue, nil
}
