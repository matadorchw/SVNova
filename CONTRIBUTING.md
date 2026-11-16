# 贡献指南

## 开发环境

目标平台为 Windows x64，运行与正式构建需 Go 1.26.4、Node.js >=22.18、pnpm 11。按锁文件安装依赖：

```powershell
pnpm install --frozen-lockfile
```

SVN 命令封装在 `services/svn`，使用参数数组、超时、取消及输出上限。不得拼接 shell 命令，不记录凭据，不使用用户仓库作测试。后端必须检查管理员和组织／团队授权，不能依赖界面隐藏操作。

上游基线与可复用文件见[来源说明](docs/upstream/README.md)。参考源码应在仓库外单独检出，不修改参考树或引入 Gitea 的 Git 业务。

## 检查与构建

修改任务触发、调度节奏、锁、队列顺序、状态／时间字段、重试、恢复或停止流程时，须在同一变更中更新[任务调度专题](docs/task-scheduling.md)的说明、Mermaid 图和源码索引，并选择对应回归测试；不要将其当作固定版本的归档快照。

生产 EXE 必须基于已提交的干净工作区构建。`build.ps1` 在开始时和打包 EXE 前检查 Git 状态；已暂存、未暂存或未跟踪文件均会导致报错退出（Git 忽略文件除外）。前端构建若更新了受版本控制的资源，应先检查并提交生成资源，再重新构建。

```powershell
$env:SVN_BIN = 'D:\SVN\bin'
.\scripts\quality.ps1
.\scripts\build.ps1
.\scripts\smoke.ps1 -Port 33003
```

`quality.ps1` 执行前端类型检查、ECharts 趋势、任务中心、仓库任务图表及搜索清除交互测试、完整 Go 测试及 vet；`build.ps1` 还构建前端、核验内嵌 SVN 包摘要并输出 `dist/SVNova-<7位提交号>.exe` 及对应 `.exe.sha256` 校验文件；文件名版本与页脚嵌入版本一致，`smoke.ps1` 默认检查当前提交对应的 EXE。Go 不在 PATH 时传 `-Go <绝对路径>`。底层命令为 `pnpm typecheck`、`pnpm test:activity-trend`、`pnpm test:task-center`、`pnpm test:search-clear`、`pnpm test:repository-tasks`、`pnpm build`、`go test ./...` 和 `go vet ./...`。

`SVN_BIN` 应指向包含测试所需 SVN 工具的目录；未设置时相关真实集成用例明确跳过。`smoke.ps1` 使用仓库下 `.tools/svn/bin` 的工具，需自行准备该隔离工具目录，再使用一次性夹具测试安装、登录、迁入、浏览、配置和重启，不操作已有实例。选择空闲端口，不向真实钉钉群发送自动测试消息。

仅验证一个 Go 用例时使用：

```powershell
go test -run '^TestName$' ./services/app/
```

改动 Go 文件后运行 gofmt；依赖变更后运行 go mod tidy。仓库没有 Vitest、Playwright 或上游 Makefile 的可用测试目标。纯文档更新检查链接和差异；涉及模板、资源或程序行为时运行相应检查。跳过／未执行项目不能写成通过，真实 LDAP、HTTP/HTTPS SVN、Windows 服务和故障验收需分别说明环境与范围。

## 文档和提交

文档用中文，标识符、代码注释和日志用英文。按[文档导航](docs/README.md)维护主题文档，描述当前实现、重要约束与未验证边界；不以日期命名功能文件，不追加 PID、开发目录、临时 EXE 或“本轮未提交”等流水。

API、数据库、权限和部署行为变更同步更新相应文档。未完成事项维护在[待办](docs/todo.md)，验证范围维护在[验收与限制](docs/acceptance.md)。完成事项移出待办，测试日期、业务时间、版权和上游提交不按开发痕迹删除。

保留上游版权和许可证，新增 Go 文件使用当前年份和 SPDX 标识。界面沿用 Gitea 浅色变量和固定 Octicons，只有 SVNova Logo 使用项目标识。

提交使用 Conventional Commits 单行消息，PR 标题同样遵循；不加 Co-Authored-By 或代人签署 Signed-off-by。未经明确请求不提交、推送、amend、重写历史或发布。提交前检查没有数据库、密钥、凭据、日志及运行目录；`.gitignore` 不能保护已被跟踪的敏感文件。
