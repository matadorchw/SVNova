# API 契约

## 会话与错误

JSON API 前缀为 `/api/v1`，所有接口要求登录。通过网页登录取得会话，变更请求要求匹配配置访问地址的 Origin 和 `X-CSRF-Token`。Cookie 为 HttpOnly、SameSite=Lax，HTTPS 访问地址启用 Secure。密码、盐、服务凭据和 Task.Source 不返回；错误以安全中文文本表示。

未登录为 401，无权限或请求验证失败为 403，缺失资源为 404；显式冲突为 409，限流为 429，不支持方法为 405，其他业务参数错误通常为 400。并非每种失败都使用相同状态码，调用方应读取具体响应。字段名按模型／DTO 保持现有大小写，不作统一重命名。

## 账号、组织与团队

下表省略 `/api/v1` 前缀。Owners 指本组织 Owners 成员或全站管理员。

| 方法与路径 | 权限及契约 |
| ---- | ---- |
| GET /me | 当前用户 |
| GET /users | 启用用户的基础资料，用于成员选择 |
| POST /user/password | 当前启用本地用户；CurrentPassword、NewPassword、ConfirmPassword，成功 changed=true 并撤销全部会话；LDAP 拒绝 |
| GET/POST /organizations | 可访问组织列表；创建仅管理员，Name，创建者加入 Owners，OwnerID 兼容字段仅允许本人 |
| GET/POST /organizations/:id/members | Owners；UserID、Remove，保护最后 Owners |
| GET/POST /organizations/:id/teams | Owners；创建 Name、Description，禁止 Owners／new 保留名 |
| POST /organizations/:id/settings | Owners；FullName、Email、Description、Website、Location，不接受改名／权限变更 |
| POST /organizations/:id/leave | 仅本人退出，移除组织及团队关系，保护最后 Owners |
| POST /organizations/:id/delete | Owners；Name 完整匹配，仅空组织；有仓库 409，成功 ok=true |
| GET/POST /teams/:id/members | Owners；UserID 或添加时 Username、Remove；添加成员自动加入组织 |
| GET/POST /teams/:id/repositories | Owners；RepositoryID、Remove，内置 Owners 不作为普通团队修改授权 |
| POST /teams/:id/leave | 本人退出，保护最后 Owners |
| GET/POST /admin/site | 管理员；SiteName 为去空白后 1–80 字符，无控制字符 |
| GET/POST /admin/users | 管理员列表及创建本地用户 |
| POST /admin/users/:id | 管理员编辑 Active、Admin、Password，撤销目标会话；不转换认证源或编辑邮箱 |
| POST /admin/users/:id/delete | 管理员；Name 完整匹配；仅本地，禁止当前账号、最后活跃管理员及组织最后活跃 Owners；事务清理关系／会话，成功 ok=true |

自行改密的新密码为 12–1024 字节，管理员创建／重置至少 12 字节；自行改密要求确认一致且不同于旧密码，15 分钟最多 10 次提交。用户 Created／LastLogin 为 Unix 秒，0 表示未知，仅成功发放会话才更新 LastLogin。组织资料上限：全名 100、邮箱 254、描述 255、网站 255（HTTP/HTTPS）、地区 50 字符。

## 仓库与任务

所有仓库接口检查当前整仓库访问权限；管理操作另需 Owners，本地迁入／重试本地迁入只允许管理员。

| 方法与路径 | 契约 |
| ---- | ---- |
| GET/POST /repositories | 可读列表／创建迁移任务；成功返回 pending 仓库及站内 Location |
| GET /repositories/:id | 仓库详情，不暴露受管物理路径或密码 |
| DELETE /repositories/:id | Owners；Confirm=仓库名、DeleteFiles 默认 false；拒绝活动任务及不安全路径 |
| GET /repositories/:id/tree | path、revision，指定修订目录项 |
| GET /repositories/:id/file | 文件内容、大小、二进制标识；预览可转码 |
| GET /repositories/:id/download | 单文件原始字节下载 |
| GET /repositories/:id/log | 固定 revision 快照，path、cursor、direction，数组响应，每页 10 条 |
| GET /repositories/:id/tasks | 当前仓库最近 100 个任务，倒序 |
| POST /repositories/:id/sync | Owners，ready 且未 halted 的镜像；queued=true；重复活动任务 409 |
| POST /repositories/:id/retry | Owners；重试最近失败任务，完整性停止不自动解除；本地任务仅管理员 |
| POST /repositories/:id/settings | Owners，镜像配置；全局锁或本仓库活动任务冲突返回明确中文 409 |
| POST /repositories/:id/convert-mirror | Owners，本地 ready 仓库快速转换后台任务 |

迁移字段：OrgID、Name、Kind（local／mirror）、Source。local 要求管理员和 SourceStopped=true，源为服务器本地已停止仓库；mirror 使用 HTTP/HTTPS 根地址、Username、Password、SkipTLS、Interval（秒）。界面默认 Interval=28800，0 关闭自动同步，正值至少 600；同组织重名 409。

镜像设置字段：Remote、Username、ReplacePassword、Password、SkipTLS、Interval。ReplacePassword=false 保留密码；同地址修改凭据可离线加密保存，更换地址需在线检查。UUID 变化、修订回退和已有完整性停止不会因保存自动解除。

快速转换字段：Remote、Username、Password、SkipTLS、Interval、FullAccess。FullAccess 必须确认远端完整读取和停止外部写入。响应为 `{taskID,queued:true,taskURL}`；凭据加密后写入任务候选，不返回或写日志。不提供模式字段；检查仓库身份与镜像配置，跳过逐修订内容比对。执行及提交前复查权限／状态，恢复失败保留日志，重启后的中断转换需手动重试。

历史 cursor 为包含边界的修订号，direction=older／newer（默认 older）。响应头 X-Revision 为快照，X-Older-Cursor／X-Newer-Cursor 为下一游标，0 表示无该方向页。后续请求同时保留 revision 和 path；不能把游标当偏移或假设路径修订连续。修订链接进入对应版本浏览，旧 diff API 和页面返回 404。

仓库 JSON 的 Size 为完整受管目录普通文件长度之和，SizeUpdated／SizeChecked 为成功／尝试统计的 Unix 秒，0 成功时间表示未知。页面仅读缓存；tree 的 Entry.Size 仍为精确字节，目录没有递归大小。文本预览支持 UTF-8、GBK／GB18030、带 BOM 的 UTF-16，下载不转码；命令输出有 16 MiB 上限。

### 状态轮询

GET `/{org}/{repo}/-/migrate/status`（保留名组织可用 `/repositories/:id/-/migrate/status`），逐次检查 CanRead，返回 status（1 准备／2 执行或排队／3 失败／4 完成）、message、tasks、repositoryState、canRetry 和 progress。POST `/{org}/{repo}/settings/migrate/retry` 使用现有重试授权。

progress=-1 为不定进度，0–100 为可计算阶段进度，不是所有迁移阶段统一的总百分比。verify 数量包含 r0；元数据阶段仅报告位置／耗时，不伪造比例。修订属性阶段使用 verification=revprops 和 verification_position（-1 未收到输出），按 `(position+1)/(total_revision+1)` 计算；显示当前阶段与本次执行耗时。原始命令输出和凭据不返回。

## LDAP 管理

仅管理员。规范路径为 `/api/v1/admin/ldap`，只保存一份配置；旧 `/admin/auths` 系列接口已移除，不提供源编号或名称。

| 方法与路径 | 契约 |
| ---- | ---- |
| GET/POST /admin/ldap | 查询／保存唯一配置；未配置时 GET 返回默认值 |
| POST /admin/ldap/preview | 预览配置变更清理 |
| POST /admin/ldap/sync | 手动同步；认证和用户同步均须启用 |
| GET /admin/ldap/sync-history | page（固定每页 20 条；兼容正整数 size，但不改变大小），全部 LDAP 历史倒序 |
| GET/POST /admin/ldap-schedule | 独立保存同步计划，Time 为北京时间五字段 cron |

配置包含 Enabled、URL、BindDN、Password、ReplacePassword、BaseDN、Filter、UsernameAttribute、KeyAttribute、NameAttribute、EmailAttribute、SkipVerify、UsePagedSearch、SearchPageSize、SyncEnabled、AllowDeactivateAll。密码加密且不回显；编辑不替换保留。初始配置默认验证证书、分页 500、同步开启、空结果保护、邮箱 mail；分页大小 1–1000。只保存一个配置，LDAP 账号使用明确标识并绑定稳定身份。同名本地账号同步跳过，登录不能接管。

预览返回 requiresConfirmation、deleteUsers、sessions、memberships、teamMemberships、token。清理保存要求 ConfirmCleanup=true、CleanupToken；令牌 10 分钟有效，绑定候选配置和待删关系，过期／伪造／陈旧为 409。保护最后活跃管理员／Owners，清理保留本地用户、组织、团队、仓库和历史。配置未变化不清理。

同步成功返回 ok=true 及 report（matchedUsers／syncedUsers／skippedLocalCount／skippedLocalUsers／skippedDisabledCount／deactivatedUsers）；本地同名摘要最多 20 个。关闭同步返回 409，空结果默认保留用户。历史不含源编号和名称，状态为 running／success／failed，时间使用 Unix 毫秒并有 DurationMS，运行中断记录安全失败分类，不回填旧历史。

调度默认 `0 0 * * *`，支持星号、列表、范围及步长，例如 `30 2 * * *`；不接受 HH:MM。锁忙等待，停用／切换源／修改时间取消尚未开始旧任务；停机不补跑，保存调度不清理用户，无下一次执行时间的表达式不触发。

## 钉钉通知

仅管理员，前缀 `/api/v1/admin/notifications`。

| 方法与后缀 | 契约 |
| ---- | ---- |
| GET/POST 无后缀 | 查询／整体保存配置，响应 config、configured；凭据不回显 |
| POST /test | 主动排队测试消息，总开关关闭仍可用，但需有效机器人配置 |
| GET /deliveries | items、total、page、pages；每页 20 条，查询 q、kind、status、sort |
| GET /deliveries/:id | delivery、attempts、preview 安全详情 |
| POST /deliveries/:id/redeliver | 仅 failed 可复制为新投递，使用当前机器人，保留 ParentID；其他状态 409 |

配置：Enabled 默认 false；SVNEnabled／LDAPEnabled 默认 true；SVNFailureOnly、SVNNoChange、LDAPNoChange 默认 false；RetentionDays=30／90／180／0，默认 90，0 永久；ReplaceWebhook／Webhook、ReplaceSecret／Secret，未替换保留。只接受官方 HTTPS 加签机器人，不支持仓库排除字段。保存不发送，无历史补发；SVNFailureOnly 优先于成功无变化通知。

投递 kind=svn／ldap／test；status=pending／sending／retry／success／failed／cancelled。sort 支持 newest／oldest、created／created-desc、attempts／attempts-desc、next／next-desc、finished／finished-desc，未知回退 newest，相同时 ID 决胜。临时网络／408／429／5xx及限流按 1／5／15 分钟有限重试，最多 4 次；永久错误直接失败。发送至少间隔 3 秒，超时 15 秒，响应最多 64 KiB，拒绝重定向。

关闭通知／分类、切换机器人或改变相关规则会取消受影响的待发投递，已经发出的请求不能撤回。超时和重启可能重复送达，通知失败不改变业务结果。保留清理只删除结束投递及尝试，不清理业务任务或同步历史。

## 活动和头像

GET `/api/v1/activity/heatmap?org=<可选组织ID>` 返回 heatmapData、totalContributions、pending、failed、today，按当前可读仓库迁移／同步执行及北京时间日期统计近一年；没有执行记录的旧任务兜底一次，不虚构历史重试。

GET `/api/v1/activity/trend?org=0&kind=sync&from=YYYY-MM-DD&to=YYYY-MM-DD`：kind=local／mirror／sync，包含首尾日期，范围在今天往前 365 天至今天且 from<=to。非法组织／日期 400，组织越权 403。响应 `{kind,from,to,today,pending,days:[{date,counts:{queued,running,succeeded,failed,interrupted}}]}`，每天补零，状态更新归原执行开始日，无响应缓存。

头像 GET／HEAD：`/avatars/current`、`/avatars/users/:id?org=:orgID`、`/avatars/organizations/:id`。当前用户可读本人，管理员可读用户／组织；其他用户头像要求双方属于指定组织，组织头像要求成员身份。size 白名单为 48／56／64／80／96／128／200／256／512，无效回退默认。先授权再校验 ETag，使用 private,no-cache，匹配 304，撤权仍拒绝。

## 页面入口和列表

| 页面 | 路径 |
| ---- | ---- |
| 登录／退出 | /user/login、POST /user/logout |
| 首页／趋势 | /、/activity；不提供 /user/activity |
| 迁移 | /repo/migrate |
| 个人资料／账号／组织 | /user/settings、/user/settings/account、/user/settings/organization |
| 用户主页 | /user/profile?user_id=:id；仅本人或管理员 |
| 用户管理 | /-/admin/users、/new、/:id、/:id/edit |
| LDAP 配置 | /-/admin/ldap、/-/admin/ldap/sync-history |
| 站点／组织／通知后台 | /-/admin/config、/-/admin/orgs、/-/admin/notifications |
| 创建／查看组织 | /org/create、/{org}；保留名组织用 /org/{org} |
| 组织成员／团队／设置 | /org/{org}/members、/teams、/settings；团队详情 /org/{org}/teams/{team}，仓库页追加 /repositories |
| 仓库 | /{org}/{repo}，设置／任务追加 /settings、/tasks；保留名组织使用 /repositories/:id |

部分旧网页路径由服务端兼容跳转，API 前缀不变。HTML 与 JSON 均在后端授权。通知配置页提供 `?tab=deliveries` 记录入口，`&delivery=:id` 详情，无配置／记录标签页；记录每 2 秒局部刷新，对话框／焦点期间暂停。

仓库 HTML 列表每页 10 个，首页用 repo_page，组织／团队／Profile／管理员用户详情用 page；首页活动 page 每页 10 条，两个分页独立保留筛选。列表先权限／搜索／排序再分页，总数和团队候选不截断，仓库 JSON 列表不因此分页。

用户管理列表 q 匹配用户名／姓名／邮箱，active／admin 取空或 0／1，auth 为空／local／ldap；sort 为 alphabetically／reversealphabetically／oldest／newest／lastlogin／reverselastlogin，page 每页 20。组织管理支持 q 及 oldest／newest／name／name-desc／created／created-desc。排序表头和菜单保留筛选，选择后回第一页。业务页面时间为北京时间，原始 SVN 日期字段不变。

管理员用户、通知投递、LDAP 同步历史及仓库任务记录每页 20 条；其余页面分页每页 10 条。共用分页组件提供完整／紧凑模式，LDAP 同步历史使用完整模式，首页仓库使用紧凑模式。LDAP 目录查询批次不受影响。

`GET /api/v1/tasks`：认证后查询任务中心，新增可选参数 q：组织名／仓库名按字面包含匹配，纯数字或 #编号同时精确匹配任务编号，与筛选条件叠加且受仓库权限限制。org、repository 仍兼容正整数 ID 参数，kind 为 local／mirror／sync／conversion，state 为 queued／running／succeeded／failed，page 为页码。每页 20 条，越界回到末页；返回 items、total、page、pages、active 和服务端转义后的 html 列表片段。筛选候选、计数、记录均按当前仓库权限限制，禁止缓存。仓库任务页支持 task_id 定位同仓库历史任务，超过最近 100 条也可访问；选中记录不提供重试按钮。

仓库设置默认 `/settings` 为基本设置，`/settings/mirror` 为镜像设置，`/settings/danger` 为危险操作。保留名组织使用 `/repositories/:id?tab=settings&section=basic|mirror|danger`，管理权限与原设置相同。

任务中心 items 的 Result 为中文结果概要；仓库迁移状态 tasks 的 message 使用同一概要，完成时间改为独立 finished 字段（北京时间字符串，未结束为空）。结果只使用 task_execution 保存的修订快照，不依赖仓库当前修订；排队／运行时概要为“—”。现有授权规则保持不变。

修订历史日志条目的 Paths 增加 CopyFromPath（复制来源路径）和 CopyFromRevision（来源修订），从 SVN XML 读取；无来源时分别为空字符串和 0。原分页与权限契约不变。

`GET /api/v1/repositories/:id/tasks/history?page=1`：按仓库 CanRead 授权，只接受 GET；每页 20 条，页码限制在有效范围。返回 summary（total、successRate 百分比或 null、revisionSpan、added）、points（全历史最新执行点，按开始时间及编号升序）、items（当前页原任务展示字段，新增 finished）、page、pages、active、repositoryState。points 包含 id、kind、state、started／finished（Unix 毫秒，未知为 0）、duration（毫秒或 null）、before／after（可靠成功修订快照或 null）、added（成功同步修订差，否则 0）。重试按任务 ID 合并，只计最新执行；无源路径或凭据字段。成功率分母为成功、失败、中断，空分母为 null；summary 为全历史统计，与分页及客户端图表缩放无关。记录为空时 points／items 为数组，页码及总页数均为 1。
