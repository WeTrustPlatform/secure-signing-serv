package main

import "github.com/jinzhu/gorm"

type dbMock struct{}

func (m *dbMock) Create(i interface{}) *gorm.DB {
	return nil
}

func (m *dbMock) Where(a interface{}, b ...interface{}) *gorm.DB {
	return nil
}

func (m *dbMock) First(a interface{}, b ...interface{}) *gorm.DB {
	return nil
}

func (m *dbMock) Save(a interface{}) *gorm.DB {
	return nil
}
