package main

import (
  "log"
  "fmt"
  "strings"
  "sync"
  "bytes"
  "os"
  "os/exec"
  "net/http"
  "io/ioutil"

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
  http.HandleFunc("/deploy/", c.deploymentHandler)
  http.HandleFunc("/output/", c.outputLogHandler)
  http.HandleFunc("/error/", c.errLogHandler)
  http.ListenAndServe(c.Config.Port, nil)
}


func (c *Context)deploymentHandler(w http.ResponseWriter, r *http.Request) {
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
    workingDir := a.GetWorkingDir(c.Config.WorkingDirPath)
    if _, err := os.Stat(checkoutPath); os.IsNotExist(err) {
      log.Printf("[%s] Cloning project...", deployid)
      executeCommand(workingDir, deployid, "git", "clone", a.GetCloneURL(), checkoutPath)
    }
    log.Printf("[%s] Fetching the latest code...", deployid)
    executeCommandWihDir(workingDir, deployid, checkoutPath, "git", "fetch")
    executeCommandWihDir(workingDir, deployid, checkoutPath, "git", "reset", "--hard", "origin/" + a.GetBranch())
    executeCommandWihDir(workingDir, deployid, checkoutPath, "git", "submodule", "update", "--init", "--recursive")

    log.Printf("[%s] Start deploying...", deployid)
    executeCommandWihDir(workingDir, deployid, checkoutPath, "git", "rev-parse", "HEAD")

    log.Printf("[%s] Executing fabric...", deployid)
    deployTask := fmt.Sprintf("deploy:branch_name=%s", a.GetBranch())
    executeCommandWihDir(workingDir, deployid, checkoutPath, "fab", "-R", a.GetRoles(), deployTask)

    currentDeployment.ResetDeploymentID(a.GetHash())
    log.Printf("[%s] Deployment Done...", deployid)
  }()
}

func (c *Context)outputLogHandler(w http.ResponseWriter, r *http.Request) {
  c.logFileHandler(w, r, "/output/", "stdout")
}

func (c *Context)errLogHandler(w http.ResponseWriter, r *http.Request) {
  c.logFileHandler(w, r, "/error/", "stderr")
}

func (c *Context)logFileHandler(w http.ResponseWriter, r *http.Request, route string, logType string) {
  s := strings.TrimPrefix(r.URL.Path, route)
  routes := strings.Split(s, "/")
  if len(routes) != 2 {
    errorHandler(w, http.StatusNotFound, "404 page not found")
    return
  }
  a := config.NewAppSetting(routes[0])
  logFilebytes, err := ioutil.ReadFile(a.GetWorkingDir(c.Config.WorkingDirPath) + "/" + logType + "." + routes[1] + ".log")
  if err != nil {
    fmt.Fprintf(w, "Deployment not found: %s", routes[1])
    return
  }
  b := bytes.NewBuffer(logFilebytes)
  w.Header().Set("Content-type", "text/plain")
  if _, err := b.WriteTo(w); err != nil {
    fmt.Fprintf(w, "Unable to open log file: %s", routes[1])
  }
}

func executeCommand(workingDir string, deployid string, name string, arg ...string)  {
  stdoutfile, err := os.Create(workingDir + "/stdout." + deployid + ".log")
  if err != nil { panic(err) }
  defer stdoutfile.Close()
  stderrfile, err := os.Create(workingDir + "/stderr." + deployid + ".log")
  if err != nil { panic(err) }
  defer stderrfile.Close()
  cmd := exec.Command(name, arg...)
  cmd.Stdout = stdoutfile
  cmd.Stderr = stderrfile
  cmd.Run()
}

func executeCommandWihDir(workingDir string, deployid string, dir string, name string, arg ...string)  {
  stdoutfile, err := os.Create(workingDir + "/stdout." + deployid + ".log")
  if err != nil { panic(err) }
  defer stdoutfile.Close()
  stderrfile, err := os.Create(workingDir + "/stderr." + deployid + ".log")
  if err != nil { panic(err) }
  defer stderrfile.Close()
  cmd := exec.Command(name, arg...)
  cmd.Stdout = stdoutfile
  cmd.Stderr = stderrfile
  cmd.Dir = dir
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
