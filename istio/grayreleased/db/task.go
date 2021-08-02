package db

import (
	"log"

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
	conn := GetDb()
	defer conn.Close()

	err := conn.Create(task).Error
	if err != nil {
		log.Print("创建失败")
	}
}

// 删除

func (task *Task) Del() {
	conn := GetDb()
	defer conn.Close()

	err := conn.Delete(task).Error
	if err != nil {
		log.Print("删除失败")
	}
}

// 编辑
func (task *Task) Update() {
	conn := GetDb()
	defer conn.Close()

	err := conn.Model(task).Update(task).Error
	if err != nil {
		log.Print("修改失败")
	}
}

// 查询所有
func (task *Task) FindAll() (taskList []Task) {
	conn := GetDb()
	defer conn.Close()
	conn.Find(&task)
	return
}

//根据id查询
func (task *Task) FindById() (taskEntity Task) {
	conn := GetDb()
	defer conn.Close()
	conn.Find(&task)
	return
}
