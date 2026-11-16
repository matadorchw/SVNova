# Gitea 复用清单

基线和许可证见[上游来源](README.md)。下列路径为源码路径，用于定位归属，不是开发机器上的绝对目录。原版权、SPDX 和来源说明保留在复制文件中。

## Go 与头像

- `modules/auth/password/hash` 来自上游同名目录，保留密码哈希实现及版权，导入路径适配 svnova；最小日志和参数工具避免引入 Git 依赖。
- `modules/avatar/identicon` 的 identicon、block、colors、polygon 来自上游同名目录，保留原作者和 issue9 来源。头像接口、尺寸选择、私有缓存和访问授权由 SVNova 适配。
- SVNova 用户／组织权限、SQLite 模型、LDAP 清理、SVN 命令、后台任务及恢复实现不能视为 Gitea Git 控制器的直接复制。

## 样式和交互模块

- `web_src/theme-gitea-light.css` 复用浅色主题变量；`web_src/upstream` 下 normalize、button、form、table、header、message、container、label、input、menu、navbar、flexcontainer、segment、dropdown、heatmap 来自上游 CSS 模块，按需要裁剪；base-variables 从基础样式提取。
- navbar.ts 从上游 common-page.ts 的头部菜单切换逻辑提取，适配 aria-expanded。仓库导航采用 secondary pointing menu，保留主题悬停、焦点及窄屏行为。
- `web_src/upstream/fomantic` 保留固定基线的 Fomantic dropdown.js（2.8.7 分叉及 GITEA-PATCH）、transition 和 DOM／无障碍辅助模块，使用固定 jQuery。select-dropdown 为本项目控件初始化适配。
- `web_src/components/ActivityHeatmap.vue` 参考上游同名组件，适配中文、CSP、安全颜色和服务端日期，数据改为仓库任务执行。日活动趋势组件使用 ECharts 堆积柱状图、原生图例和滑块缩放，按 SVN 任务统计的 API 为本项目实现；ECharts、ZRender 和 tslib 许可证及 NOTICE 随静态资源分发。
- Icon.vue 与生成 sprite 均使用固定 Octicons；构建产物保留许可证。Logo 与 favicon 为 SVNova 品牌。

## 模板和功能适配

| 本项目区域 | 上游来源与本地适配 |
| ---- | ---- |
| base、install、user/auth | 上游对应模板片段和登录组合；替换上下文、中文和 CSRF，使用项目会话，不引入 Git 导航 |
| org/header、menu、home、members、teams、settings | 对应组织和团队布局；私有组织、指定仓库和最后 Owners 保护使用本项目后端 |
| user/dashboard、user/settings、用户主页 | 仪表盘／账号／资料布局；活动采用左侧组织头像、正文和右侧任务类型图标布局，来自任务执行，其他用户主页仅管理员查看 |
| admin/auth、admin/user、后台布局 | 单 LDAP 配置表单／用户列表、表单、搜索／过滤／排序、分页及 Octicons；仅呈现已实现的 LDAP／账号字段 |
| repo/header、home、view、settings | secondary-nav、文件列表、左右分栏及设置布局；使用 SVN 修订和路径面包屑，不模拟分支或 Git 数据 |
| repo/migrate | 上游 migrate/git.tmpl、migrating.tmpl 和 repo-migration／repo-migrate 的表单命名与轮询惯例；本地路径、SVN 阶段、权限和持久任务由本项目适配 |
| 通知配置与投递 | 沿用后台附属标题／面板／表格及 Webhook 交互惯例；全站单机器人、事务 Outbox、仅失败重发为本项目业务 |
| 历史及列表分页 | base/paginate 和主题按钮／菜单样式；SVN 历史用修订游标，仓库及首页活动按当前页面约定分页 |

模板按上游目录结构拆分，但移除了不支持的 Git、SSH、2FA、LDAP 组、工单、分支和仓库差异业务。其余表单和 SVN 数据区域属于适配实现，不声称原样复制。页面来源注释中的固定提交用于追溯，不作为需要删除的开发时间痕迹。

## 静态资源与分发

`static/img` 保留的 repo_default、loading、failed、checkmark 图片源于上游 public/assets/img；原默认用户头像已由本地 identicon 替代。当前迁移面板使用 SVNova 标识，不展示 Gitea 茶杯品牌图片；保留的未使用资源不在本次文档整理中删除。

生产构建同步静态图片、固定图标及 Gitea／Gogs、Octicons、jQuery 许可证至 public/assets 并内嵌 EXE。工具包另按各自组件许可证分发。不得为清除过程文档而删除这些许可文本或代码内来源说明。

分页模板复用完整／紧凑两种展示，紧凑布局参考固定基线 DashboardRepoList；首尾跳转由固定 Octicons chevron 路径横向并列组合为双箭头。
