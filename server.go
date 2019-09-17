package main

import "github.com/gin-gonic/gin"
import "os"
import "os/exec"
import "net/http"
import "bytes"
import "encoding/json"
import "github.com/levigross/grequests"
import "mime/multipart"
import (
  log "github.com/sirupsen/logrus"
)

var callbackUrl string
var debug string

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func compile(c *gin.Context) {
    form, _ := c.MultipartForm()
    var token string
    var files []*multipart.FileHeader
    if val, ok := form.File["upload[]"]; !ok {
        c.AbortWithStatusJSON(400, gin.H{"error": "form does not contain files in upload[]"})
        return
    } else {
        files = val
    }
    if tokens, ok := form.Value["token"]; !ok {
        c.AbortWithStatusJSON(400, gin.H{"error": "form does not contain token"})
        return
    } else {
        token = tokens[0]
    }
    dir   := "/data/"+token
    os.Mkdir(dir, 0777)
    log.Info(dir+" created")
    mainFound := false
    for _, file := range files {
        if file.Filename == "main.tex" {
            mainFound = true
            break
        }
    }
    if mainFound == false {
        c.AbortWithStatusJSON(400, gin.H{"error": "main.tex not found"})
        return
    }
    cmd := "cd "+dir+" && pdflatex -interaction=nonstopmode main.tex && cd /"
    if callbackUrl!="" {
        uris, ok := form.Value["uri"];
        if !ok {
            c.AbortWithStatusJSON(400, gin.H{"error": "form does not contain callback uri"})
            return
        }
        uri := uris[0]
        cCp := c.Copy()
        go func() {
            formCp, _ := cCp.MultipartForm()
		    for _, file := range formCp.File["upload[]"] {
			    cCp.SaveUploadedFile(file, dir+"/"+file.Filename)
		    }
            _, err := exec.Command("bash","-c",cmd).Output()
            if err != nil {
                value := gin.H{"error": "Compilation failed"}
                jsonValue, _ := json.Marshal(value)
                output, err := http.Post(callbackUrl+"/"+uri, "application/json", bytes.NewBuffer(jsonValue))
                if err != nil {
                    log.Error("Compilation failed; Cannot send request: "+err.Error())
                }
                if debug == "true" {
                    log.Info(output)
                }
		defer output.Body.Close()
            } else {
                f, err := grequests.FileUploadFromDisk(dir+"/main.pdf")
                if err != nil {
                    value := gin.H{"error": "Cannot upload file", "reason": err.Error()}
                    jsonValue, _ := json.Marshal(value)
                    output, err := http.Post(callbackUrl+"/"+uri, "application/json", bytes.NewBuffer(jsonValue))
                    if err != nil {
                        log.Error("Cannot upload file; Cannot send request: "+err.Error())
                    }
                    if debug == "true" {
                        log.Info(output)
                    }
		    defer output.Body.Close()
                }
                defer f[0].FileContents.Close()
                f[0].FileMime  = "application/pdf"
                f[0].FieldName = "binary"
                ro := &grequests.RequestOptions{
			        Files: f,
			        Data:  map[string]string{"token": token},
		        }
                output, err := grequests.Post(callbackUrl+"/"+uri, ro)
                if err != nil {
                    log.Error("Cannot send file: "+err.Error())
                }
                if debug == "true" {
                    log.Info(output)
                }
            }
            os.RemoveAll(dir)
        }()
        c.JSON(http.StatusOK, gin.H{"status": "Files are received"})
    } else {
        for _, file := range files {
			c.SaveUploadedFile(file, dir+"/"+file.Filename)
		}
        _, err := exec.Command("bash","-c",cmd).Output()
        if err != nil {
            c.AbortWithStatusJSON(400, gin.H{"error": "Compilation failed"})
            return
        }
        c.File(dir+"/main.pdf")
        os.RemoveAll(dir)
        c.Header("Content-Description", "File Transfer")
        c.Header("Content-Transfer-Encoding", "binary")
        c.Header("Content-Disposition", "attachment; filename=main.pdf")
        c.Header("Content-Type", "application/pdf")
    }
}

func main() {
    debug="true"
    callbackUrl = os.Getenv("CALLBACK_URL")
    loglevel := os.Getenv("DEBUG")
    if loglevel == "info" {
        log.SetLevel(log.InfoLevel)
    } else if loglevel == "warning" {
        log.SetLevel(log.WarnLevel)
    } else if loglevel == "error" {
        log.SetLevel(log.ErrorLevel)
    }
    if callbackUrl != "" {
        log.Info("Async mode. Return data to "+callbackUrl)
    } else {
        log.Info("Sync mode. Return data back to sender")
    }
    debug       = os.Getenv("DEBUG")
	router := gin.Default()
    v1 := router.Group("/v1")
    {
        v1.POST("/compile", compile)
    }
	router.Run() // listen and serve on 0.0.0.0:8080
}
