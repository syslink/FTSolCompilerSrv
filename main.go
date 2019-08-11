package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/unrolled/render"
	"os"
	"crypto/sha256"
	"encoding/hex"
)
type OpSolType int32

const (
	AddSol OpSolType = 0
	DelSol OpSolType = 1
	UpdateSol OpSolType = 2
	ListSol OpSolType = 3
	RenameSol OpSolType = 4
	CompileSol OpSolType = 5
)

type SolInfo struct {
	Type OpSolType `json:"type"`
	AccountName string `json:"accountName"`
	SolFileName string `json:"solFileName"`
	NewSolFileName string `json:"newSolFileName"`
	SolFileContent string `json:"solFileContent"`
}

const rootDir = "./datatype/"

func main() {
	http.HandleFunc("/solidity/", processSol)
	http.ListenAndServe(":8888", nil)
}

func processSol(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	r.ParseForm() //解析参数，默认是不会解析的
	fmt.Printf("request from: %s\n", r.RemoteAddr)
	if r.Method == "POST" {
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Printf("%s\n", result)

		var solInfo SolInfo
		json.Unmarshal([]byte(result), solInfo)
		fmt.Printf("%d %s : %s->%s [%s]\n", solInfo.Type, solInfo.AccountName, solInfo.SolFileName, solInfo.NewSolFileName, solInfo.SolFileContent)

		switch solInfo.Type {
			case AddSol:
				addSolHandler(w, solInfo.AccountName, solInfo.SolFileName)
			case DelSol:
				delSolHandler(w, solInfo.AccountName, solInfo.SolFileName)
			case UpdateSol:
				updateSolHandler(w, solInfo.AccountName, solInfo.SolFileName, solInfo.SolFileContent)
			case ListSol:
				listSolHandler(w, solInfo.AccountName)
			case RenameSol:
				renameSolHandler(w, solInfo.AccountName, solInfo.SolFileName, solInfo.NewSolFileName)
			case CompileSol:
				compileSolHandler(w, solInfo.AccountName, solInfo.SolFileName)
		}
	}
}

func addSolHandler(w http.ResponseWriter, accountName string, solFileName string) {
	var formatter render.Render
	folderPath := rootDir + accountName
	if _, err := os.Stat(folderPath); os.IsNotExist(err) {
		// 必须分成两步：先创建文件夹、再修改权限
		err = os.MkdirAll(folderPath, 0777) //0777也可以os.ModePerm
		if err != nil {
			responseErr(w, err.Error())
			return
		}
		err = os.Chmod(folderPath, 0777)
		if err != nil {
			responseErr(w, err.Error())
			return
		}
	}
	file, err := os.Create(folderPath + "/" + solFileName)
	if err != nil{
		responseErr(w, err.Error())
		return
	}
	defer file.Close()
	formatter.JSON(w, http.StatusOK, struct {
		Result bool `json:"result"`
	}{Result: true})
	return
}

func updateSolHandler(w http.ResponseWriter, accountName string, solFileName string, solFileContent string) {
	var formatter render.Render
	filePath := rootDir + accountName + "/" + solFileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		responseErr(w, err.Error())
		return
	}
	file, err := os.OpenFile(filePath, os.O_RDWR,0777)
	if err != nil{
		responseErr(w, err.Error())
		return
	}
	defer file.Close()
	_, err = file.Write([]byte(solFileContent))
	if err != nil {
		responseErr(w, err.Error())
		return
	}
	formatter.JSON(w, http.StatusOK, struct {
		Result bool `json:"result"`
	}{Result: true})
}

func listSolHandler(w http.ResponseWriter, accountName string) {
	fileNameList := make([]string, 0)
	hash := sha256.New()
	files, _ := ioutil.ReadDir(rootDir + accountName)
	for _, f := range files {
		bSolFile := strings.HasSuffix(f.Name(), ".sol")
		if bSolFile {
			hashedFileName := hex.EncodeToString(hash.Sum([]byte(f.Name())))
			fileNameList = append(fileNameList, hashedFileName)
		}
		var formatter render.Render
		formatter.JSON(w, http.StatusOK, struct {
			Result []string `json:"result"`
		}{Result: fileNameList})
	}
}

func delSolHandler(w http.ResponseWriter, accountName string, solFileName string) {
	var formatter render.Render
	filePath := rootDir + accountName + "/" + solFileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		responseErr(w, err.Error())
		return
	}
	err := os.Remove(rootDir + accountName + "/" + solFileName)
	if err != nil {
		responseErr(w, err.Error())
		return
	}
	formatter.JSON(w, http.StatusOK, struct {
		Result bool `json:"result"`
	}{Result: true})
}

func renameSolHandler(w http.ResponseWriter, accountName string, oldSolFileName string, newSolFileName string) {
	filePath := rootDir + accountName + "/" + oldSolFileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		responseErr(w, err.Error())
		return
	}
	err := os.Rename(rootDir + accountName + "/" + oldSolFileName, rootDir + accountName + "/" + newSolFileName)
	if err != nil {
		responseErr(w, err.Error())
		return
	}
	var formatter render.Render
	formatter.JSON(w, http.StatusOK, struct {
		Result bool `json:"result"`
	}{Result: true})
}

func compileSolHandler(w http.ResponseWriter, accountName string, solFileName string) {

}

func responseErr(w http.ResponseWriter, errInfo string) {
	var formatter render.Render
	formatter.JSON(w, http.StatusOK, struct {
		Result bool `json:"result"`
		Err string `json:"err"`
	}{Result: false, Err: errInfo})
}