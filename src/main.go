package main

import (
	"net/http"
	"io/ioutil"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/unrolled/render"
	"os"
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
	ListSharedAccount OpSolType = 6
	GetSharedSol OpSolType = 7
)

type SolInfo struct {
	Type OpSolType `json:"type"`
	ChainName   string `json:"chainName"`
	AccountName string `json:"accountName"`
	SharedAccountName string `json:"sharedAccountName"`
	SolFileName string `json:"solFileName"`
	NewSolFileName string `json:"newSolFileName"`
	SolFileContent string `json:"solFileContent"`
}

const rootDir = "./data/"
const libDir = "/usr/local/lib/solidity/"

func main() {
	var port int
	flag.IntVar(&port, "p", 8888, "端口号，默认为8888")
	http.HandleFunc("/solidity/", processSol)
	http.HandleFunc("/sampleCodeList/", querySampleCode)
	http.HandleFunc("/libsList/", queryLibs)
	http.HandleFunc("/statInfo/", queryStatInfo)
	portStr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(portStr, nil)
	InitDB("xchainunion", "pirobot001&", "rm-2ze6ix958dd7np1kg2o.mysql.rds.aliyuncs.com", "xcu_codingshare")
	CreateUserInfoTable()
	CreateProjectInfoTable()
	CreateFileInfoTable()
	CreateFileSnapshotInfoTable()

	//rand.Seed(time.Now().UnixNano())
	//index := rand.Intn(10000)
	//
	//success, userId := AddUser("ftchain" + strconv.Itoa(index), "0xaaaaaaaaaaaaaaaaa", "testaccount", "hello world", "0xdddddddd")
	//if success {
	//	success, projectId := AddProject(userId, "testaccount" + strconv.Itoa(index), "helloworld")
	//	if success {
	//		fmt.Println("projectId = ", projectId)
	//		projects := GetAllProjects()
	//		printJSONArr(projects)
	//		success, fileId := AddFile(projectId, "testfile.sol")
	//		if success {
	//			fmt.Println("fileId = ", fileId)
	//			files := GetAllFileOfProject(projectId)
	//			printJSONArr(files)
	//			success, fileSnapshotId := AddFileSnapshotInfo(fileId, (uint64)(time.Now().UnixNano() / 1e6),"file content", "{added info}")
	//			if success {
	//				fmt.Println("fileSnapshotId = ", fileSnapshotId)
	//				fileSnapshots := GetAllFileSnapshotFromTime(fileId, 1577385089530, 0)
	//				printJSONArr(fileSnapshots)
	//			} else {
	//				panic("Fail to add fileSnapshot.")
	//			}
	//		} else {
	//			panic("Fail to add file.")
	//		}
	//	} else {
	//
	//	}
	//} else {
	//	panic("Fail to add user.")
	//}
}

func printJSONArr(v interface{}) {
	jsonObj, err := json.Marshal(v)
	if err == nil {
		fmt.Println(string(jsonObj))
	} else {
		fmt.Println(err.Error())
	}
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
			fmt.Println("sol file: ", f.Name())
			fileContent, err := ioutil.ReadFile(dir + f.Name())
			if err != nil {
				fmt.Println(string(err.Error()))
				continue
			}
			fileContentStr := string(fileContent)
			solFileMap[f.Name()] = fileContentStr
		}
	}
	return nil, solFileMap
}

func querySampleCode(w http.ResponseWriter, r *http.Request) {
	fmt.Println("query samples")
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	var formatter render.Render
	err, fileInfoMap := querySolFile(libDir + r.URL.RawQuery + "/samples")
	if err != nil {
		responseErr(w, err.Error())
		return
	}
	formatter.JSON(w, http.StatusOK, fileInfoMap)
}
type StatInfo struct {
	AccountNum int `json:"accountNum"`
	SolFileNum int `json:"solFileNum"`
}

func queryStatInfo(w http.ResponseWriter, r *http.Request) {
	var formatter render.Render
	files, err := ioutil.ReadDir(rootDir)
	if err != nil {
		if os.IsNotExist(err) {
			formatter.JSON(w, http.StatusOK, struct {
				Result string `json:"result"`
			}{Result: "{accountNum: 0, solFileNum: 0}"})
		} else {
			responseErr(w, err.Error())
		}
		return
	}
	accountNum := 0
	solFileNum := 0
	for _, file := range files {
		if file.IsDir() {
			accountNum++
			dirPath := rootDir + file.Name()
			files, _ := ioutil.ReadDir(dirPath)
			for _, f := range files {
				bSolFile := strings.HasSuffix(f.Name(), ".sol")
				if bSolFile {
					solFileNum++
				}
			}
		}
	}
	statInfo := StatInfo{}
	statInfo.AccountNum = accountNum
	statInfo.SolFileNum = solFileNum
	statInfoJson, _ := json.Marshal(statInfo)
	formatter.JSON(w, http.StatusOK, struct {
		Result string `json:"result"`
	}{Result: string(statInfoJson)})
}

func queryLibs(w http.ResponseWriter, r *http.Request) {
	fmt.Println("query libs")
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")             //返回数据格式是json

	var formatter render.Render
	err, fileInfoMap := querySolFile(libDir + r.URL.RawQuery)
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
	fmt.Println("request from: ", r.RemoteAddr)
	if r.Method == "POST" {
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Println(result)

		var solInfo SolInfo
		json.Unmarshal([]byte(result), &solInfo)
		fmt.Println("%d %s %s %s : %s->%s [%s]\n", solInfo.Type, solInfo.ChainName, solInfo.AccountName, solInfo.SharedAccountName,
			solInfo.SolFileName, solInfo.NewSolFileName, solInfo.SolFileContent)

		switch solInfo.Type {
			case AddSol:
				addSolHandler(w, solInfo.ChainName, solInfo.AccountName, solInfo.SolFileName)
			case DelSol:
				delSolHandler(w, solInfo.ChainName, solInfo.AccountName, solInfo.SolFileName)
			case UpdateSol:
				updateSolHandler(w, solInfo.ChainName, solInfo.AccountName, solInfo.SolFileName, solInfo.SolFileContent)
			case ListSol:
				listSolHandler(w, solInfo.ChainName, solInfo.AccountName)
			case RenameSol:
				renameSolHandler(w, solInfo.ChainName, solInfo.AccountName, solInfo.SolFileName, solInfo.NewSolFileName)
			case CompileSol:
				compileSolHandler(w, solInfo.ChainName, solInfo.AccountName, solInfo.SolFileName)
			case ListSharedAccount:
				listSharedSolHandler(w, solInfo.ChainName, solInfo.AccountName, solInfo.SharedAccountName)
			case GetSharedSol:
				getSharedSolHandler(w, solInfo.ChainName, solInfo.AccountName, solInfo.SharedAccountName, solInfo.SolFileName)
		}
	}
}

func addSolHandler(w http.ResponseWriter, chainName string, accountName string, solFileName string) {
	var formatter render.Render
	accountPath := rootDir + chainName + "/" +  accountName + "/"
	if _, err := os.Stat(accountPath); os.IsNotExist(err) {
		// 必须分成两步：先创建文件夹、再修改权限
		err = os.MkdirAll(accountPath, 0777) //0777也可以os.ModePerm
		if err != nil {
			responseErr(w, err.Error())
			return
		}
		err = os.Chmod(accountPath, 0777)
		if err != nil {
			responseErr(w, err.Error())
			return
		}
	}
	file, err := os.Create(accountPath + solFileName)
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

func updateSolHandler(w http.ResponseWriter, chainName string, accountName string, solFileName string, solFileContent string) {
	var formatter render.Render
	filePath := rootDir + chainName + "/" + accountName + "/" + solFileName
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

func listSolHandler(w http.ResponseWriter, chainName string, accountName string) {
	var formatter render.Render
	fileNameList := make([]string, 0)
	//hash := sha256.New()
	files, err := ioutil.ReadDir(rootDir + chainName + accountName)
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
			//hashedFileName := hex.EncodeToString(hash.Sum([]byte(f.Name())))
			fileNameList = append(fileNameList, f.Name())
		}
		formatter.JSON(w, http.StatusOK, struct {
			Result []string `json:"result"`
		}{Result: fileNameList})
	}
}

func checkAccountBeShared(chainName, accountName string, sharedAccountName string) (bool) {
	return true
}

func listSharedSolHandler(w http.ResponseWriter, chainName string, accountName string, sharedAccountName string) {
	if !checkAccountBeShared(chainName, accountName, sharedAccountName) {
		responseErr(w, "No authority to access.")
		return
	}
	var formatter render.Render
	fileNameList := make([]string, 0)
	files, err := ioutil.ReadDir(rootDir + chainName + "/" + sharedAccountName)
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
		fileNameList = append(fileNameList, f.Name())
	}
	formatter.JSON(w, http.StatusOK, struct {
		Result []string `json:"result"`
	}{Result: fileNameList})
}

func getSharedSolHandler(w http.ResponseWriter, chainName string, accountName string, sharedAccountName string, solFileName string) {
	if !checkAccountBeShared(chainName, accountName, sharedAccountName) {
		responseErr(w, "No authority to access.")
		return
	}
	var formatter render.Render
	filePath := rootDir + chainName + "/" + sharedAccountName + "/" + solFileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		responseErr(w, err.Error())
		return
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		responseErr(w, err.Error())
		return
	}

	formatter.JSON(w, http.StatusOK, struct {
		Result string `json:"result"`
	}{Result: string(data)})
}

func delSolHandler(w http.ResponseWriter, chainName string, accountName string, solFileName string) {
	var formatter render.Render
	filePath := rootDir + chainName + "/" + accountName + "/" + solFileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		responseErr(w, err.Error())
		return
	}
	err := os.Remove(filePath)
	if err != nil {
		responseErr(w, err.Error())
		return
	}
	formatter.JSON(w, http.StatusOK, struct {
		Result bool `json:"result"`
	}{Result: true})
}

func renameSolHandler(w http.ResponseWriter, chainName string, accountName string, oldSolFileName string, newSolFileName string) {
	accountPath := rootDir + chainName + "/" +  accountName + "/"
	filePath := accountPath + oldSolFileName
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		responseErr(w, err.Error())
		return
	}
	err := os.Rename(filePath, accountPath + newSolFileName)
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

func compileSolHandler(w http.ResponseWriter, chainName string, accountName string, solFileName string) {
	accountPath := rootDir + chainName + "/" +  accountName + "/"
	now := time.Now().Unix()
	cmd := exec.Command("solc", "/libs/=" + libDir, "--abi", "--bin", "-o", accountPath, "--overwrite", accountPath + solFileName)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	fmt.Println("Start to compile")
	err := cmd.Run()
	fmt.Println("Finish compiling")
	if err != nil {
		responseErr(w, stderr.String())
	} else {
		files, err := ioutil.ReadDir(accountPath)
		if err != nil {
			responseErr(w, err.Error())
			return
		}
		contractInfoMap := make(map[string]ContractInfo)
		for _, file := range files {
			if !file.IsDir() {
				fileName := file.Name()
				fileTime := file.ModTime().Unix()
				fileContent, err := ioutil.ReadFile(accountPath + fileName)
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
		fmt.Println(string(json))
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