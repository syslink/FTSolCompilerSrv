package main

import (
	"database/sql"
	"fmt"
	"time"
)

type ProjectInfo struct {
	Id uint64 		 			`gorm:"column:id;AUTO_INCREMENT;PRIMARY_KEY"`
	UserId   uint64          	`gorm:"column:userId;unique_index:i_userIdName;not null"`
	Name   string        		`gorm:"column:name;unique_index:i_userIdName;not null"`
	Desc   sql.NullString   	`gorm:"column:desc"`
	CreatedTime  time.Time     `gorm:"column:createdTime;not null"`
}

func (ProjectInfo) TableName() string {
	return "projectInfo"
}

func CreateProjectInfoTable() {
	dbInstance := GetDB()
	if !dbInstance.HasTable(&ProjectInfo{}) {
		dbInstance.CreateTable(&ProjectInfo{})
		fmt.Println("Create table projectInfo.")
	}
}

func AddProject(userId uint64, name, desc string) (bool, uint64) {
	var nsDesc sql.NullString
	if err := nsDesc.Scan(desc); err != nil {
		panic(err)
	}
	projectInfo := ProjectInfo{UserId: userId, Name:name, Desc:nsDesc, CreatedTime:time.Now()}
	dbInstance := GetDB()
	dbInstance.Create(&projectInfo)
	fmt.Println(projectInfo.Id)
	result := dbInstance.NewRecord(projectInfo)
	return !result, projectInfo.Id
}

func DelProject(projectId uint64) bool {
	projectInfo := ProjectInfo{Id: projectId}
	dbInstance := GetDB()
	dbInstance.Delete(&projectInfo)
	result := dbInstance.NewRecord(projectInfo)
	return result
}

func ModifyProjectName(projectId uint64, name string) bool {
	projectInfo := ProjectInfo{Id: projectId}
	dbInstance := GetDB()
	dbInstance.Model(&projectInfo).Update(ProjectInfo{Name:name})
	dbInstance.First(&projectInfo, projectId)
	return projectInfo.Name == name
}

func GetAllProjects() []ProjectInfo {
	projects := []ProjectInfo{}
	dbInstance := GetDB()
	dbInstance.Find(&projects)
	return projects
}

func GetAllProjectsOfUser(userId uint64) []ProjectInfo {
	projects := []ProjectInfo{}
	dbInstance := GetDB()
	dbInstance.Where(&ProjectInfo{UserId: userId}).Find(&projects)
	return projects
}

func GetOneProject(projectId uint64) ProjectInfo {
	project := ProjectInfo{}
	dbInstance := GetDB()
	dbInstance.Find(&project, projectId)
	return project
}