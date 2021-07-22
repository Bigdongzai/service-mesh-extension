package db

import (
	"github.com/jinzhu/gorm"
)

type Task struct {
	gorm.Model
	Name     string `json:"name"`
	Service  string `json:"service"`
	Version  string `json:"version"`
	Describe string `json:"describe"`
}

// 新增
func (task *Task) Add() {

}

// 删除

// 编辑

// 查询
