package main

import (
	"github.com/jinzhu/gorm"
)

type dbMock struct {
	txs []*transaction
}

func (m *dbMock) Create(r interface{}) *gorm.DB {
	tx := r.(*transaction)
	m.txs = append(m.txs, tx)
	return nil
}

func (m *dbMock) First(a interface{}, b ...interface{}) *gorm.DB {
	*a.(*transaction) = *m.txs[0]
	return nil
}

func (m *dbMock) Save(a interface{}) *gorm.DB {
	return nil
}
