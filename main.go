package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)


var port int = 9000
var debug bool = false
var baseDir string = "/opt/image_bucket"
//var baseDir string = "/tmp"

func getCurrentTIme() string{
	return time.Now().Format("2006-01-02 15:04:05")
}

func successResponse(w http.ResponseWriter, data interface{}){
	w.Header().Set("Content-Type", "application/json")		
	w.WriteHeader(http.StatusOK)
	resMap := make(map[string]interface{}, 0)	
	resMap["code"] = 0
	resMap["data"] =  data
	resMap["msg"] = ""

	rspStr,err := json.Marshal(resMap)
	if err != nil{
		log.Println(errors.Wrap(err, getCurrentTIme()))
	}
	fmt.Fprintln(w, string(rspStr))
}

func failResponse(w http.ResponseWriter, msg string){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	resMap := make(map[string]interface{}, 0)
	resMap["code"] = -1
	resMap["data"] = "{}"
	resMap["msg"] = msg
	rspStr, err := json.Marshal(resMap)
	if err != nil{
		log.Println(errors.Wrap(err, getCurrentTIme()))
	}
	fmt.Fprintln(w, string(rspStr))
}



func uploadHandler(w http.ResponseWriter, r *http.Request){
	//set cross origin resource sharing
	origin := r.Header.Get("Origin")
	if origin != ""{
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length, Authorization")
	}
	if r.Method == "OPTIONS"{
		return
	}



	// file max size 10 M
	err := r.ParseMultipartForm(10 << 20)
	if err != nil{
		failResponse(w, "exceeded file max size")
		return
	}
	//bucket 英文，不超过10个字符, 不包含特殊字符
	bucket := r.PostFormValue("bucket")
	if bucket == "" {
		failResponse(w, "bucket must be specified")
		return
	}

	file, handler, err := r.FormFile("file_name")
	if err != nil{
		failResponse(w, err.Error())
		return
	}
	defer file.Close()
	if debug{
		log.Printf("upload file name: %s\n", handler.Filename)
		log.Printf("upload file size: %d\n", handler.Size)
		log.Printf("MIME Header: %+v\n", handler.Header)
	}

	// create bucket
	storeDirName := baseDir + "/" + bucket
	_, err = os.Stat(storeDirName)	
	if err != nil{
		if os.IsNotExist(err){
			//create directory
			err := os.MkdirAll(storeDirName, os.ModePerm)
			if err != nil{
				failResponse(w, "create bucket failed")
				return	
			}
		}
	}

	//file name hash
	newFileBase := filepath.Base(handler.Filename) + strconv.FormatInt(time.Now().UnixNano(), 10)
	digest := fmt.Sprintf("%x", md5.Sum([]byte(newFileBase)))
	newFileName := digest + filepath.Ext(handler.Filename)

	dst, err := os.Create(storeDirName + "/" + newFileName)
	if err != nil{
		log.Println(err)
	}

	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil{
		failResponse(w, errors.Wrap(err, "save fail").Error())
		return
	}
	
	resMap := make(map[string]interface{})
	resMap["img_url"] = newFileName
	successResponse(w, resMap)
}

func previewHandler(w http.ResponseWriter, r *http.Request){

	//set cross origin resource sharing
	origin := r.Header.Get("Origin")
	if origin != ""{
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type,Content-Length, Authorization")
	}
	if r.Method == "OPTIONS"{
		return
	}

	// r.ParseForm()
	// bucket := r.Form.Get("bucket")
	param := strings.TrimPrefix(r.URL.Path, "/preview/")
	paramList := strings.Split(param, "/")
	if len(paramList) != 2{
		failResponse(w, "request parameter incorrect")
		return
	}
	bucket := paramList[0]
	image_file := paramList[1]
	if len(bucket) <= 0{
		failResponse(w, "bucket must be specified")
		return
	}
	if len(image_file) <= 0{
		failResponse(w, "no specified image name")
		return
	}
	image_location := baseDir + "/" + bucket + "/" + image_file
	fileByte, err := ioutil.ReadFile(image_location)
	if err != nil {
		failResponse(w, errors.Wrap(err, "read file failed").Error())
		return 
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.WriteHeader(http.StatusOK)
	w.Write(fileByte)
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/preview/", previewHandler)

	addr := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: addr, Handler: mux}
	server.ListenAndServe()
}
