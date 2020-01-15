package main

import (
	"fmt"
)

type FileInfo struct {
	Id        uint64          	`gorm:"column:id;AUTO_INCREMENT;PRIMARY_KEY"`
	ProjectId uint64 		 	`gorm:"column:projectId;unique_index:i_projectIdName;not null"`
	Name      string 		 	`gorm:"column:name;unique_index:i_projectIdName;not null"`
}

func (FileInfo) TableName() string {
	return "fileInfo"
}

func CreateFileInfoTable() {
	dbInstance := GetDB()
	if !dbInstance.HasTable(&FileInfo{}) {
		dbInstance.CreateTable(&FileInfo{})
		fmt.Println("Create table fileInfo.")
	}
}
// 此处只有增加文件的接口，删除和修改文件的操作不更新数据库，当做增量部分的更新
func AddFile(projectId uint64, name string) (bool, uint64) {
	fileInfo := FileInfo{ProjectId: projectId, Name:name}
	dbInstance := GetDB()
	dbInstance.Create(&fileInfo)
	fmt.Println(fileInfo.Id)
	result := dbInstance.NewRecord(fileInfo)
	return !result, fileInfo.Id
}

func GetAllFileOfProject(projectId uint64) []FileInfo {
	files := []FileInfo{}
	dbInstance := GetDB()
	dbInstance.Where(&FileInfo{ProjectId: projectId}).Find(&files)
	return files
}