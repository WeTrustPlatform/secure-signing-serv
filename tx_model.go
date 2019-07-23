package main

import "github.com/jinzhu/gorm"

type transaction struct {
	Nonce    uint64
	To       string
	Value    string
	Gas      uint64
	GasPrice string
	Data     string
	Hash     string `gorm:"unique_index"`

	gorm.Model
}
