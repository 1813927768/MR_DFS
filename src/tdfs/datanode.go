package tdfs

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	// "io/ioutil"

	"os"
	"strings"
)

func (datanode *DataNode) Run() {
	// curl -X POST http://127.0.0.1:11091/upload -F "upload=@/Users/treasersmac/Programming/MilkPrairie/Gou/TinyDFS/Client/chunk-1" -H "Content-Type: multipart/form-data"
	router := gin.Default()
	router.MaxMultipartMemory = 128 << 20 //上传最大文件限制1024MB
	router.POST("/putchunk", func(c *gin.Context) {

		ReplicaNum := c.PostForm("ReplicaNum")
		fmt.Printf("* ReplicaNum= %s\n", ReplicaNum)

		file, header, err := c.Request.FormFile("putchunk")
		if err != nil {
			c.String(http.StatusBadRequest, "XXX Bad request")
			TDFSLogger.Panic("XXX DataNode error: ", err)
			return
		}
		filename := header.Filename
		fmt.Println("****************************************")
		fmt.Println(file, err, filename)
		fmt.Println("****************************************")

		chunkout, err := os.Create(datanode.DATANODE_DIR + "/chunk-" + ReplicaNum) //在服务器本地新建文件进行存储
		if err != nil {
			fmt.Println("XXX DataNode error at Create chunk file", err.Error())
			TDFSLogger.Panic("XXX DataNode error: ", err)
		}
		defer chunkout.Close()
		io.Copy(chunkout, file) //在服务器本地新建文件进行存储

		chunkdata := ReadFileByBytes(datanode.DATANODE_DIR + "/chunk-" + ReplicaNum)

		hashStr := GetHashStr(chunkdata)
		fmt.Println("** chunk hash", ReplicaNum, ": %s", hashStr)
		FastWrite(datanode.DATANODE_DIR+"/achunkhashs/chunkhash-"+ReplicaNum, []byte(hashStr))

		updateDataNodeMetadata(datanode)

		c.String(http.StatusCreated, "PutChunk SUCCESS\n")
	})

	router.GET("/getchunk/:chunknum", func(c *gin.Context) {
		chunknum := c.Param("chunknum")
		num, err := strconv.Atoi(chunknum)
		if err != nil {
			fmt.Println("XXX DataNode error(getchunk) at Atoi parse chunknum to int", err.Error())
			TDFSLogger.Panic("XXX DataNode error: ", err)
		}
		fmt.Println("Parsed num: ", num)

		chunkBytes := ReadFileByBytes(datanode.DATANODE_DIR + "/chunk-" + strconv.Itoa(num))
		chunkHash := ReadFileByBytes(datanode.DATANODE_DIR + "/achunkhashs/chunkhash-" + strconv.Itoa(num))

		/* check hash */
		hashStr := GetHashStr(chunkBytes)
		fmt.Println("*** chunk hash calculated: ", hashStr)
		fmt.Println("*** chunk hash get: ", string(chunkHash))

		if hashStr == string(chunkHash) {
			c.String(http.StatusOK, string(chunkBytes))
		} else {
			c.String(http.StatusExpectationFailed, "")
			fmt.Println("X=X the broken chunk-", num, "'s hash(checksum) is WRONG")
			TDFSLogger.Panic("XXX Broken chunk found at ", num)
			// broken chunk, something could be done to recover it
		}

	})

	router.DELETE("/delchunk/:chunknum", func(c *gin.Context) {
		chunknum := c.Param("chunknum")
		num, err := strconv.Atoi(chunknum)
		if err != nil {
			fmt.Println("XXX DataNode error at Atoi parse chunknum to int", err.Error())
			TDFSLogger.Panic("XXX DataNode error: ", err)
		}
		fmt.Println("Parsed num: ", num)

		CleanFile(datanode.DATANODE_DIR + "/chunk-" + strconv.Itoa(num))
		// CleanFile(datanode.DATANODE_DIR+"/achunkhashs/chunkhash-"+strconv.Itoa(num))
		DeleteFile(datanode.DATANODE_DIR + "/achunkhashs/chunkhash-" + strconv.Itoa(num))

		c.String(http.StatusOK, "delete DataNode{*}/chunk-"+strconv.Itoa(num)+" SUCCESS")
	})

	router.GET("/getmeta", func(c *gin.Context) {
		c.JSON(http.StatusOK, datanode)
	})

	router.Run(":" + strconv.Itoa(datanode.Port))
}

func (datanode *DataNode) SetConfig(location string, storageTotal int) {
	temp := strings.Split(location, ":")
	res, err := strconv.Atoi(temp[2])
	if err != nil {
		fmt.Println("XXX DataNode error at Atoi parse Port", err.Error())
		TDFSLogger.Panic("XXX DataNode error: ", err)
	}
	datanode.Port = res
	datanode.Location = location
	datanode.StorageTotal = storageTotal
	datanode.StorageAvail = datanode.StorageTotal

	datanode.ChunkAvail = append(datanode.ChunkAvail, 0)
	for i := 1; i < datanode.StorageAvail; i++ {
		datanode.ChunkAvail = append(datanode.ChunkAvail, datanode.StorageTotal-i)
	}

	datanode.LastEdit = time.Now().Unix()
	for num := 0; num < datanode.StorageTotal; num++ {
		CreateFile(datanode.DATANODE_DIR + "/chunk-" + strconv.Itoa(num))
	}
	fmt.Println("************************************************************")
	fmt.Println("************************************************************")
	fmt.Printf("*** Successfully Set Config data for a datanode\n")
	datanode.ShowInfo()
	fmt.Println("************************************************************")
	fmt.Println("************************************************************")
}

func (datanode *DataNode) Reset() {
	var i int = 0
	for i < datanode.StorageTotal {
		CleanFile("TinyDFS/DataNode1/chunk-" + strconv.Itoa(i))
		i++
	}

	exist, err := PathExists(datanode.DATANODE_DIR + "/achunkhashs")
	if err != nil {
		fmt.Println("XXX DataNode error at Get Dir chunkhashs", err.Error())
		TDFSLogger.Panic("XXX DataNode error: ", err)
	}
	if !exist {
		err = os.MkdirAll(datanode.DATANODE_DIR+"/achunkhashs", os.ModePerm)
		if err != nil {
			fmt.Println("XXX DataNode error at MkdirAll chunkhashs", err.Error())
			TDFSLogger.Panic("XXX DataNode error: ", err)
		}
	} else {
		err := os.RemoveAll(datanode.DATANODE_DIR + "/achunkhashs")
		if err != nil {
			fmt.Println("XXX DataNode error at RemoveAll file hash data", err.Error())
			TDFSLogger.Panic("XXX DataNode error: ", err)
		}

		err = os.MkdirAll(datanode.DATANODE_DIR+"/achunkhashs", os.ModePerm)
		if err != nil {
			fmt.Println("XXX DataNode error at MkdirAll chunkhashs", err.Error())
			TDFSLogger.Panic("XXX DataNode error: ", err)
		}
	}
}

func (datanode *DataNode) ShowInfo() {
	fmt.Printf("Location: %s\n", datanode.Location)
	fmt.Printf("DATANODE_DIR: %s\n", datanode.DATANODE_DIR)
	fmt.Printf("Port: %d\n", datanode.Port)
	fmt.Printf("StorageTotal: %d\n", datanode.StorageTotal)
	fmt.Printf("StorageAvail: %d\n", datanode.StorageAvail)
	fmt.Printf("ChunkAvail: %d\n", datanode.ChunkAvail)
	fmt.Printf("LastEdit: %d\n", datanode.LastEdit)
}

func (datanode *DataNode) RecvChunkAndStore(ReplicaList []ReplicaLocation, chunkData ChunkUnit) {
	var i int = 0
	for i < len(ReplicaList) {
		if ReplicaList[i].ServerLocation == datanode.Location {
			break
		}
		i++
	}
	chunkFileName := "TinyDFS/DataNode1/chunk-" + strconv.Itoa(ReplicaList[i].ReplicaNum) //datanode.chunkAvail[0]
	datanode.ChunkAvail = datanode.ChunkAvail[1:]
	FastWrite(chunkFileName, chunkData)
	fmt.Printf("> Replica data finish stored in %s.\n", chunkFileName)
	if i+1 < len(ReplicaList) {
		fmt.Printf("> Next, replica will send to %s\n", ReplicaList[i+1].ServerLocation)
	}
}
