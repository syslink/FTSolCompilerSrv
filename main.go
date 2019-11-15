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
	"bytes"
	"os/exec"
	"time"
	"flag"
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

const rootDir = "./data/"
const libDir = "/usr/local/lib/solidity/"
const sampleDir = libDir + "samples/"

func main() {
	var port int
	flag.IntVar(&port, "p", 8888, "端口号，默认为8888")
	http.HandleFunc("/solidity/", processSol)
	http.HandleFunc("/sampleCodeList/", querySampleCode)
	http.HandleFunc("/libsList/", queryLibs)
	portStr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(portStr, nil)
}

func querySolFile(dir string) (error, map[string]string) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err, nil
	}
	solFileMap := make(map[string]string)
	for _, f := range files {
		bSolFile := strings.HasSuffix(f.Name(), ".sol")
		if bSolFile {
			fmt.Printf("sol file: %s", f.Name())
			fileContent, err := ioutil.ReadFile(libDir + f.Name())
			if err != nil {
				fmt.Printf(string(err.Error()))
				continue
			}
			fileContentStr := string(fileContent)
			solFileMap[f.Name()] = fileContentStr
		}
	}
	return nil, solFileMap
}

func querySampleCode(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("query libs")
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	var formatter render.Render
	err, fileInfoMap := querySolFile(sampleDir)
	if err != nil {
		responseErr(w, err.Error())
		return
	}
	formatter.JSON(w, http.StatusOK, fileInfoMap)
}

func queryLibs(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("query libs")
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	var formatter render.Render
	err, fileInfoMap := querySolFile(libDir)
	if err != nil {
		responseErr(w, err.Error())
		return
	}
	formatter.JSON(w, http.StatusOK, fileInfoMap)
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
		json.Unmarshal([]byte(result), &solInfo)
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

func createSolFile(folderPath string, fileName string) (err error) {
	if _, err = os.Stat(folderPath); os.IsNotExist(err) {
		// 必须分成两步：先创建文件夹、再修改权限
		err = os.MkdirAll(folderPath, 0777) //0777也可以os.ModePerm
		if err != nil {
			return err
		}
		err = os.Chmod(folderPath, 0777)
		if err != nil {
			return err
		}
	}
	_, err = os.Create(folderPath + "/" + fileName)
	if err != nil{
		return err
	}
	return nil
}

func updateSolHandler(w http.ResponseWriter, accountName string, solFileName string, solFileContent string) {
	var formatter render.Render
	filePath := rootDir + accountName + "/" + solFileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		err = createSolFile(rootDir + accountName, solFileName)
		if err != nil {
			responseErr(w, err.Error())
			return
		}
	}
	file, err := os.OpenFile(filePath, os.O_RDWR | os.O_TRUNC,0777)
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
	var formatter render.Render
	fileNameList := make([]string, 0)
	hash := sha256.New()
	files, err := ioutil.ReadDir(rootDir + accountName)
	if err != nil {
		if os.IsNotExist(err) {
			formatter.JSON(w, http.StatusOK, struct {
				Result []string `json:"result"`
			}{Result: fileNameList})
		} else {
			responseErr(w, err.Error())
		}
		return
	}
	for _, f := range files {
		bSolFile := strings.HasSuffix(f.Name(), ".sol")
		if bSolFile {
			hashedFileName := hex.EncodeToString(hash.Sum([]byte(f.Name())))
			fileNameList = append(fileNameList, hashedFileName)
		}
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

type ContractInfo struct {
	Name string `json:"name"`
	Abi string `json:"abi"`
	Bin string `json:"bin"`
}

func compileSolHandler(w http.ResponseWriter, accountName string, solFileName string) {
	now := time.Now().Unix()
	cmd := exec.Command("solc", "/libs/=" + libDir, "--abi", "--bin", "-o", rootDir + accountName, "--overwrite", rootDir + accountName + "/" + solFileName)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		responseErr(w, stderr.String())
	} else {
		files, err := ioutil.ReadDir(rootDir + accountName)
		if err != nil {
			responseErr(w, err.Error())
			return
		}
		contractInfoMap := make(map[string]ContractInfo)
		for _, file := range files {
			if !file.IsDir() {
				fileName := file.Name()
				fileTime := file.ModTime().Unix()
				fileContent, err := ioutil.ReadFile(rootDir + accountName + "/" + fileName)
				if err != nil {
					responseErr(w, err.Error())
					return
				}
				if fileTime >= now {
					if strings.HasSuffix(fileName, ".bin") {
						contractName := fileName[0:len(fileName) - 4]
						if _, ok := contractInfoMap[contractName]; ok {
							contractInfo := contractInfoMap[contractName]
							contractInfo.Bin = string(fileContent)
							contractInfoMap[contractName] = contractInfo
						} else {
							contractInfo := ContractInfo{Name: contractName, Bin: string(fileContent), Abi: ""}
							contractInfoMap[contractName] = contractInfo
						}
					}
					if strings.HasSuffix(fileName, ".abi") {
						contractName := fileName[0:len(fileName) - 4]
						if _, ok := contractInfoMap[contractName]; ok {
							contractInfo := contractInfoMap[contractName]
							contractInfo.Abi = string(fileContent)
							contractInfoMap[contractName] = contractInfo
						} else {
							contractInfo := ContractInfo{Name: contractName, Bin: "", Abi: string(fileContent)}
							contractInfoMap[contractName] = contractInfo
						}
					}
				}
			}
		}
		json, _ := json.Marshal(contractInfoMap)
		fmt.Printf(string(json))
		var formatter render.Render
		formatter.JSON(w, http.StatusOK, contractInfoMap)
	}

}

func responseErr(w http.ResponseWriter, errInfo string) {
	var formatter render.Render
	formatter.JSON(w, http.StatusOK, struct {
		Result bool `json:"result"`
		Err string `json:"err"`
	}{Result: false, Err: errInfo})
}