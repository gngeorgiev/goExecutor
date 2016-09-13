package javascript

import (
	"io/ioutil"
	"net/http"
	"time"

	"encoding/json"
	"fmt"
	"strings"

	"os/exec"

	"github.com/gin-gonic/gin"
	"github.com/gngeorgiev/goExecutor/languages/baseLanguage"
	"github.com/gngeorgiev/goExecutor/utils"
)

type JavascriptLanguage struct {
	baseLanguage.BaseLanguage
}

func (j JavascriptLanguage) GetName() string {
	return "js"
}

func (j JavascriptLanguage) GetDefaultImage() string {
	return "node:latest"
}

func (j JavascriptLanguage) PrepareWorkspace(workspace string) error { //TODO: npm cache etc
	return nil
}

func (j JavascriptLanguage) ExecuteCode(ip, port, code string) (string, error) {
	defer utils.TrackTime(time.Now(), "execContainerCommand took: %s")

	body := gin.H{"code": code}
	bodyJson, marshalError := json.Marshal(body)
	if marshalError != nil {
		return "", marshalError
	}

	r, requestError := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://%s:%s/execute", ip, port),
		strings.NewReader(string(bodyJson)),
	)

	r.Header.Add("Content-Type", "application/json")

	if requestError != nil {
		return "", requestError
	}

	response, doRequestError := http.DefaultClient.Do(r)
	if doRequestError != nil {
		return "", doRequestError
	}

	defer response.Body.Close()
	b, readBodyError := ioutil.ReadAll(response.Body)
	if readBodyError != nil {
		return "", readBodyError
	}

	result := string(b)
	return result, nil
}

func (j JavascriptLanguage) CleanupWorkspace(workspace string) error {
	return nil
}

func (j JavascriptLanguage) PrepareContainerFiles(folder string) error {
	files := j.GetFilesToCopy(j.GetName())
	for _, f := range files {
		_, copyError := exec.Command("cp", "-r", f, folder).CombinedOutput()
		if copyError != nil {
			return copyError
		}
	}

	return nil
}

func (j JavascriptLanguage) GetCommand() []string {
	return []string{"node", "server.js"}
}
