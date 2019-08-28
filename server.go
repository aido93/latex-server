package main

import "github.com/gin-gonic/gin"
import "log"
import "os"
import "os/exec"
import "net/http"
import "bytes"
import "encoding/json"
import "github.com/levigross/grequests"
import "mime/multipart"

var callbackUrl string

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
        c.AbortWithStatusJSON(400, gin.H{"error": "form does not contain files in upload[]`"})
        return
    } else {
        files = val
    }
    if tokens, ok := form.Value["token"]; !ok {
        c.AbortWithStatusJSON(400, gin.H{"error": "form does not contain token`"})
        return
    } else {
        token = tokens[0]
    }
    dir   := "/data/"+token
    os.Mkdir(dir, 0777)
    log.Println(dir+" created")
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
                _, err := http.Post(callbackUrl, "application/json", bytes.NewBuffer(jsonValue))
                if err != nil {
                    log.Println("Compilation failed; Cannot send request: "+err.Error())
                }
            } else {
                f, err := grequests.FileUploadFromDisk(dir+"/main.pdf")
                if err != nil {
                    value := gin.H{"error": "Cannot upload file", "reason": err.Error()}
                    jsonValue, _ := json.Marshal(value)
                    _, err := http.Post(callbackUrl, "application/json", bytes.NewBuffer(jsonValue))
                    if err != nil {
                        log.Println("Cannot upload file; Cannot send request: "+err.Error())
                    }
                }
                defer f[0].FileContents.Close()
                f[0].FileMime  = "application/pdf"
                f[0].FieldName = "binary"
                ro := &grequests.RequestOptions{
			        Files: f,
			        Data:  map[string]string{"token": token},
		        }
                _, err = grequests.Post(callbackUrl, ro)
                if err != nil {
                    log.Println("Cannot send file: "+err.Error())
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
        c.Header("Content-Disposition", "attachment; filename=businessPlan.pdf")
        c.Header("Content-Type", "application/pdf")
    }
}

func main() {
    callbackUrl = os.Getenv("CALLBACK_URL")
	router := gin.Default()
    v1 := router.Group("/v1")
    {
        v1.POST("/compile", compile)
    }
	router.Run() // listen and serve on 0.0.0.0:8080
}
