# SVNova

采用 Gitea 技术栈与界面习惯的 SVN 仓库浏览及镜像管理应用。面向 Windows 10/11、Windows Server 2016 及以上 x64，单 EXE 交付，中文界面。

## 主要功能

- 本地账号、LDAP 登录及用户同步，管理员后台和组织 Owners／团队整仓库授权。
- 本地 SVN 仓库迁入、HTTP/HTTPS 远端镜像、手动及定时同步、持久任务与恢复保护。
- 指定修订的目录与文件浏览、修订历史、单文件下载及仓库大小统计。
- 迁移／同步活动、日活动趋势、全站钉钉通知及失败投递重发。

SVNova 不提供 SVN 协议读写服务、Git 托管、匿名访问、路径级权限、目录打包下载或独立归档备份系统。本地迁入会移动源仓库，操作前须停止原 SVN 服务及其他写入。镜像源账号必须能读取完整仓库与全部历史。

## 首次运行

将 `SVNova.exe` 放入固定程序目录，在 PowerShell 中运行：

```powershell
.\SVNova.exe --work-path D:\SVNova\work
```

访问 [首次安装页面](http://127.0.0.1:3000)，配置管理员、访问地址、数据库和仓库路径。默认仅监听本机；需要内网安装时可用 `--install-listen 0.0.0.0:3001`，并在安装页设置正确的访问地址。安装完成后，监听地址以 `custom/conf/app.ini` 为准。

工作目录包含数据库、仓库及加密密钥，必须位于本机磁盘并限制 Windows ACL。升级时停止程序、保留完整工作目录，仅替换 EXE。

## Windows 服务

停止控制台实例后，以管理员身份在 EXE 所在目录运行：

```powershell
.\SVNova.exe --work-path D:\SVNova\work service install
.\SVNova.exe --work-path D:\SVNova\work service start
```

服务名为 `SVNova`，自动启动，默认账户为 LocalSystem。完整安装、端口配置和升级步骤见[部署指南](docs/deployment.md)。

## 构建与测试

需要 Go 1.26.4、Node.js >=22.18、pnpm 11，使用仓库锁定的依赖和 SVN 工具包。

```powershell
pnpm install --frozen-lockfile
$env:SVN_BIN = 'D:\SVN\bin'
.\scripts\quality.ps1
.\scripts\build.ps1
```

`SVN_BIN` 指向包含 `svn`、`svnadmin`、`svnlook`、`svnsync` 等工具的目录；未设置时，真实 SVN 集成测试会明确跳过。构建结果为 `dist/SVNova.exe`，内嵌前端、模板、固定 SVN 工具包和第三方许可证。Go 不在 PATH 时，脚本支持 `-Go <绝对路径>`。更多说明见[贡献指南](CONTRIBUTING.md)。

## 文档

- [文档导航](docs/README.md)：功能与权限、架构、运维及接口。
- [验收与限制](docs/acceptance.md)：自动化覆盖、实际环境验证及暂缓项目。
- [待办](docs/todo.md)：尚未完成的工作。
- [上游来源与许可证](docs/upstream/README.md)：固定基线、复用范围和第三方归属。

项目使用 [MIT 许可证](LICENSE)；第三方组件保留各自的版权与许可证。
