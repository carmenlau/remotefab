package config

import (
  "os"
  "strings"
)

// AppSetting - Per app settig from enviroment variable
type AppSetting struct {
	hash string
}

// NewAppSetting - New AppSetting instance
func NewAppSetting(hash string) (AppSetting) {
  return AppSetting{hash: strings.ToUpper(hash)}
}

// IsVaild - check if all required config is set
func (a *AppSetting) IsVaild() bool {
  return a.GetCloneURL() != "" && a.GetBranch() != ""
}

// GetHash - application hash
func (a *AppSetting) GetHash() string {
  return a.hash
}

// GetCloneURL - get clone url from env based on application hash
func (a *AppSetting) GetCloneURL() string {
  return os.Getenv(strings.Join([]string{a.hash, "_CLONE_URL"}, ""))
}

// GetBranch - get deployment branch from env based on application hash
func (a *AppSetting) GetBranch() string {
  return os.Getenv(strings.Join([]string{a.hash, "_BRANCH"}, ""))
}

// GetRoles - get ROLES for fabric deployment
func (a *AppSetting) GetRoles() string {
  return os.Getenv(strings.Join([]string{a.hash, "_ROLES"}, ""))
}

// GetWorkingDir - get app working dir
func (a *AppSetting) GetWorkingDir(workingDir string) string {
  // exmaple: /tmp/fe24118a6918073fb5408055df38279a/
  return strings.Join([]string{workingDir, a.hash}, "")
}

// GetCheckDir - get app checkout dir
func (a *AppSetting) GetCheckoutDir(workingDir string) string {
  // exmaple: /tmp/FE24118A6918073FB5408055DF38279A/checkout/
  return strings.Join([]string{a.GetWorkingDir(workingDir), "/checkout"}, "")
}
