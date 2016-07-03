package main

import (
  "log"
  "fmt"
  "strings"
  "sync"
  "os"
  "os/exec"
  "net/http"

  "github.com/google/uuid"
  "github.com/carmenlau/remotefab/config"
)

// Context - Handler Context
type Context struct {
	Config config.Configuration
}

var currentDeployment = &ongoingDeployment{
  deploymentIDMap: make(map[string]interface{}),
}

func main() {
  c := Context{Config: config.NewConfigFromEnv()}
  http.HandleFunc("/deploy/", c.requestHandler)
  http.ListenAndServe(c.Config.Port, nil)
}


func (c *Context)requestHandler(w http.ResponseWriter, r *http.Request) {
  settingHash := strings.TrimPrefix(r.URL.Path, "/deploy/")

  if settingHash == "" {
    errorHandler(w, http.StatusNotFound, "404 page not found")
    return
  }
  a := config.NewAppSetting(settingHash)
  deployid := uuid.Must(uuid.NewRandom()).String()

  if !a.IsVaild() {
    errorHandler(w, http.StatusBadRequest,
      "Missing Configuration, please contact administrator!")
    return
  }

  oldDeployid := currentDeployment.GetDeploymentID(a.GetHash())
  if oldDeployid != "" {
    errorHandler(w, http.StatusBadRequest,
      fmt.Sprintf("Duplicated deployment, ongoing deployment id: %q", oldDeployid))
    return
  }
  currentDeployment.SetDeploymentID(a.GetHash(), deployid)
  fmt.Fprintf(w, "Deployment id: %q", deployid)

  // start deployment
  go func() {
    log.Printf("[%s] Application: %s", deployid, settingHash)

    checkoutPath := a.GetCheckoutDir(c.Config.WorkingDirPath)
    if _, err := os.Stat(checkoutPath); os.IsNotExist(err) {
      log.Printf("[%s] Cloning project...", deployid)
      executeCommand("git", "clone", a.GetCloneURL(), checkoutPath)
    }
    log.Printf("[%s] Fetching the latest code...", deployid)
    executeCommandWihDir(checkoutPath, "git", "fetch")
    executeCommandWihDir(checkoutPath, "git", "reset", "--hard")
    executeCommandWihDir(checkoutPath, "git", "checkout", a.GetBranch())
    executeCommandWihDir(checkoutPath, "git", "pull")
    executeCommandWihDir(checkoutPath, "git", "submodule", "update", "--init", "--recursive")

    log.Printf("[%s] Start deploying...", deployid)
    executeCommandWihDir(checkoutPath, "git", "rev-parse", "HEAD")

    log.Printf("[%s] Executing fabric...", deployid)
    deployTask := fmt.Sprintf("deploy:branch_name=%s", a.GetBranch())
    executeCommandWihDir(checkoutPath, "fab", "-R", a.GetRoles(), deployTask)

    currentDeployment.ResetDeploymentID(a.GetHash())
    log.Printf("[%s] Deployment Done...", deployid)
  }()
}

func executeCommand(name string, arg ...string)  {
  cmd := exec.Command(name, arg...)
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  cmd.Run()
}

func executeCommandWihDir(dir string, name string, arg ...string)  {
  cmd := exec.Command(name, arg...)
  cmd.Dir = dir
  cmd.Stdout = os.Stdout
  cmd.Stderr = os.Stderr
  cmd.Run()
}

func errorHandler(w http.ResponseWriter, status int, msg string) {
  w.WriteHeader(status)
  fmt.Fprint(w, msg)
}

type ongoingDeployment struct {
  sync.RWMutex
  deploymentIDMap map[string]interface{}
}

func (o *ongoingDeployment) SetDeploymentID(hash string, id string) {
  o.Lock()
  o.deploymentIDMap[hash] = id
  o.Unlock()
}

func (o *ongoingDeployment) ResetDeploymentID(hash string) {
  o.Lock()
  delete(o.deploymentIDMap, hash)
  o.Unlock()
}

func (o *ongoingDeployment) GetDeploymentID(hash string) string {
  o.RLock()
  defer o.RUnlock()
  if o.deploymentIDMap[hash] == nil {
    return ""
  }
  return o.deploymentIDMap[hash].(string)
}
