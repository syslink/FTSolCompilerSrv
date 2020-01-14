package dbEntity

import (
	"database/sql"
	"fmt"
	"crypto/sha256"
)

type FileSnapshotInfo struct {
	Id   			uint64          	  	   `gorm:"column:id;AUTO_INCREMENT;PRIMARY_KEY"`
	FileId    		uint64          		   `gorm:"column:fileId;unique_index:i_fileHeight;not null"`
	Height          uint64					   `gorm:"column:height;unique_index:i_fileHeight;not null"`  // 文件快照的高度，每个文件会按序递增
	UploadTime 		uint64 		 		       `gorm:"column:uploadTime;not null"`
	BStoreContent   bool                       `gorm:"column:bStoreContent;not null"`
	Content   		sql.NullString             `gorm:"column:content"`          // 初始快照内容，由于内容可以为空，因此无法以内容是否为空来判断文件内容是否被保存，需要通过bStoreContent来判断
	ChangedEvents   sql.NullString             `gorm:"column:changedEvents"`    // 变化事件集合，通过Content和ChangedEvents便可推导出最后的内容

	// 初始内容hash，如果此hash同本文件上一个快照的EndContentHash相等，Content字段便可以从上一个快照推导出来，为了降低文件保存量和网络传输数据量，可以每隔一定时间（如五分钟）保存一次完整快照
	BeginContentHash   string				   `gorm:"column:beginContentHash;index"`
	EndContentHash     string				   `gorm:"column:endContentHash;index"`     // 变化后内容的hash，可以通过客户端上传，无需推导出内容后计算出来
}

var fileCurHeight map[uint64]uint64

func (FileSnapshotInfo) TableName() string {
	return "fileSnapshotInfo"
}

func CreateFileSnapshotInfoTable() {
	dbInstance := GetDB()
	if !dbInstance.HasTable(&FileSnapshotInfo{}) {
		dbInstance.CreateTable(&FileSnapshotInfo{})
		fmt.Println("Create table fileSnapshotInfo.")
	}
}

func AddFileSnapshotInfo(fileId, uploadTime uint64, content, changedEvents, endContentHash string) (bool, uint64) {
	var nsContent sql.NullString
	if err := nsContent.Scan(content); err != nil {
		panic(err)
	}
	var nsChangeEvents sql.NullString
	if err := nsChangeEvents.Scan(changedEvents); err != nil {
		panic(err)
	}
	height, exist := fileCurHeight[fileId]
	if !exist {
		fileHeight, err := GetMaxHeightOfFile(fileId)
		if err != nil {
			return false, 0
		}
		fileCurHeight[fileId] = fileHeight
		height = fileHeight
	}
	curHeight := height + 1

	bStoreContent := false
	lastStoredHeight, err := GetMaxHeightOfContentStoredFile(fileId)
	if err != nil {
		return false, 0
	}

	h := sha256.New()
	h.Write([]byte(content))
	beginContentHash := string(h.Sum(nil))

	if needStoreContent(fileId, curHeight, lastStoredHeight, beginContentHash) {
		bStoreContent = true
		nsContent.Scan(nil)
	}

	fileSnapshotInfo := FileSnapshotInfo{FileId: fileId, Height:curHeight, UploadTime:uploadTime, BStoreContent:bStoreContent,
										 Content:nsContent, ChangedEvents:nsChangeEvents, BeginContentHash:beginContentHash, EndContentHash:endContentHash}
	dbInstance := GetDB()
	dbInstance.Create(&fileSnapshotInfo)
	fmt.Println(fileSnapshotInfo.Id)
	result := dbInstance.NewRecord(fileSnapshotInfo)
	return !result, fileSnapshotInfo.Id
}

// 如果跟上一次快照超过5分钟，则需要保存内容
// 如果本次内容的hash跟上一个快照的结束内容hash不一致，也要保存内容
func needStoreContent(fileId, curHeight, lastStoredHeight uint64, beginContentHash string) bool {
	timeOut := curHeight - lastStoredHeight >= 5 * 60 / 3     // 即5分钟保存一次内容，快照3秒钟一次，每过100个高度保存一次内容
	if timeOut {
		return true
	}
	fileSnapshotInfos := []FileSnapshotInfo{}
	dbInstance := GetDB()
	dbInstance.Where("fileId = ? And height = ? And endContentHash = ?", fileId, curHeight - 1, beginContentHash).Find(&fileSnapshotInfos)
	return len(fileSnapshotInfos) == 0
}

func GetAllFileSnapshotByTime(fileId, fromTime, toTime uint64) []FileSnapshotInfo {
	fileSnapshotInfos := []FileSnapshotInfo{}
	dbInstance := GetDB()
	if toTime == 0 {
		dbInstance.Where("fileId = ? And uploadTime >= ?", fileId, fromTime).Find(&fileSnapshotInfos)
	} else if toTime > fromTime {
		dbInstance.Where("fileId = ? And uploadTime >= ? And uploadTime < toTime", fileId, fromTime, toTime).Find(&fileSnapshotInfos)
	}
	return fileSnapshotInfos
}

func GetAllFileSnapshotByHeight(fileId, fromHeight, heightSpan uint64) []FileSnapshotInfo {
	fileSnapshotInfos := []FileSnapshotInfo{}
	dbInstance := GetDB()
	dbInstance.Where("fileId = ? And height BETWEEN ? And ?", fileId, fromHeight, fromHeight + heightSpan).Find(&fileSnapshotInfos)
	return fileSnapshotInfos
}

// 获取某文件最近一次快照高度
func GetMaxHeightOfFile(fileId uint64) (uint64, error) {
	dbInstance := GetDB()
	var maxHeight uint64
	row := dbInstance.Table("fileSnapshotInfo").Select("MAX(height)").Where(" fileId = ? ", fileId).Row()
	err := row.Scan(&maxHeight)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}
	return  maxHeight, nil
}

// 获取最近一次保存某个文件内容的快照高度
func GetMaxHeightOfContentStoredFile(fileId uint64) (uint64, error) {
	dbInstance := GetDB()
	var maxHeight uint64
	row := dbInstance.Table("fileSnapshotInfo").Select("MAX(height)").Where(" fileId = ? And bStoreContent=true ", fileId).Row()
	err := row.Scan(&maxHeight)
	if err != nil {
		fmt.Println(err.Error())
		return 0, err
	}
	return  maxHeight, nil
}