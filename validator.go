package main

import (
	"github.com/ethereum/go-ethereum/core/types"
	lua "github.com/yuin/gopher-lua"
)

// Transaction represents an ethereum transaction to be passed to the Lua VM
type Transaction struct {
	To    string
	Value string
}

// Checks whether the first lua argument is a *LUserData with *Transaction and returns this *Transaction.
func checkTransaction(L *lua.LState) *Transaction {
	ud := L.CheckUserData(1)
	if v, ok := ud.Value.(*Transaction); ok {
		return v
	}
	L.ArgError(1, "transaction expected")
	return nil
}

// Getter and setter for the Transaction#To
func transactionGetSetName(L *lua.LState) int {
	p := checkTransaction(L)
	if L.GetTop() == 2 {
		p.To = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(p.To))
	return 1
}

// Getter and setter for the Transaction#Value
func transactionGetSetValue(L *lua.LState) int {
	p := checkTransaction(L)
	if L.GetTop() == 2 {
		p.Value = L.CheckString(2)
		return 0
	}
	L.Push(lua.LString(p.Value))
	return 1
}

// Constructor
func newTransaction(L *lua.LState, transaction *Transaction) *lua.LUserData {
	ud := L.NewUserData()
	ud.Value = transaction
	L.SetMetatable(ud, L.GetTypeMetatable("transaction"))
	L.Push(ud)
	return ud
}

var transactionMethods = map[string]lua.LGFunction{
	"to":    transactionGetSetName,
	"value": transactionGetSetValue,
}

// Registers my transaction type to given L.
func registerTransactionType(L *lua.LState) {
	mt := L.NewTypeMetatable("transaction")
	L.SetGlobal("transaction", mt)
	// methods
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), transactionMethods))
}

func validate(rules string, tx *types.Transaction) (bool, error) {
	L := lua.NewState()
	defer L.Close()

	registerTransactionType(L)

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
		newTransaction(L, &Transaction{
			To:    tx.To().String(),
			Value: tx.Value().String(),
		}),
	)
	if err != nil {
		return false, err
	}

	ret := L.Get(-1)
	L.Pop(1)

	return ret == lua.LTrue, nil
}
