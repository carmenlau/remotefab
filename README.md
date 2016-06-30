Remote fabric is a simple web server that pull latest code from git and execut fabfile.py for deployment

### Requirements

- Fabric: http://www.fabfile.org/
- Golang 1.6

## Configuration

| Environmental Variables       |                                                 |
|-------------------------------|-------------------------------------------------|
| REMOTE_FABRIC_CLONE_DIR_PATH  | Local path for deployment, default is `/tmp/`   |
| REMOTE_FABRIC_PORT            | Web server port, default is `:8080`             |
| <APP_HASH>_CLONE_URL          | Deployment clone URL                            |
| <APP_HASH>_BRANCH             | Deployment branch                               |
| <APP_HASH>_ROLES              | Deployment roles                                |


Example:
```
export E10ADC3949BA59ABBE56E057F20F883E_CLONE_URL=git@bitbucket.org:xxx/xxx.git
export E10ADC3949BA59ABBE56E057F20F883E_BRANCH=master
export E10ADC3949BA59ABBE56E057F20F883E_ROLES=parse01,parse02
```
