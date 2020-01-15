package main

import (
	"database/sql"
	"fmt"
	"time"
)

type UserInfo struct {
	Id 		  uint64            `gorm:"column:id;AUTO_INCREMENT;PRIMARY_KEY"`
	ChainName string 		 	`gorm:"column:chainName;unique_index:chainAddr;not null"`
	Address   string        	`gorm:"column:address;unique_index:chainAddr;not null"`
	Account   sql.NullString   `gorm:"column:account;index:i_account"`
	DevInfo   sql.NullString   `gorm:"column:devInfo"`
	SignedInfo   string        `gorm:"column:signedInfo;not null"`
	CreatedTime  time.Time    `gorm:"column:createdTime;not null"`
}

func (UserInfo) TableName() string {
	return "userInfo"
}

func CreateUserInfoTable() {
	dbInstance := GetDB()
	if !dbInstance.HasTable(&UserInfo{}) {
		dbInstance.CreateTable(&UserInfo{})
		fmt.Println("Create table userinfo.")
	}
}

func AddUser(chainName, address, account, devInfo, signedInfo string) (bool, uint64) {
	var nsAccount sql.NullString
	if err := nsAccount.Scan(account); err != nil {
		panic(err)
	}
	var nsDevInfo sql.NullString
	if err := nsDevInfo.Scan(devInfo); err != nil {
		panic(err)
	}
	userInfo := UserInfo{ChainName: chainName, Address:address, Account:nsAccount, DevInfo:nsDevInfo, SignedInfo:signedInfo, CreatedTime:time.Now()}
	dbInstance := GetDB()
	dbInstance.Create(&userInfo)
	fmt.Println(userInfo.Id)
	result := dbInstance.NewRecord(userInfo)
	return !result, userInfo.Id
}

func main1() {
	InitDB("xchainunion", "pirobot001&", "rm-2ze6ix958dd7np1kg2o.mysql.rds.aliyuncs.com", "xcu_codingshare")
	AddUser("ftchain", "0xaaaaaaaaaaaaaaaaa", "testaccount", "hello world", "0xdddddddd")
}