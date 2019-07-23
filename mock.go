package main

import "github.com/jinzhu/gorm"

type dbMock struct{}

func (m *dbMock) Create(i interface{}) *gorm.DB {
	return nil
}
