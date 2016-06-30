package main

import (
  "log"
  "fmt"
  "strings"
  "os"
  "os/exec"
  "net/http"

  "github.com/google/uuid"
  "github.com/carmenlau/remote-fabric/config"
)

// Context - Handler Context
type Context struct {
	Config config.Configuration
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
