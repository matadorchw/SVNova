# 数据库模型与迁移

## 当前版本与存储

当前 schema 为 **14**，以 migration 表的最高 ID 检查兼容性。存储使用本机 SQLite／xorm，开启 foreign_keys、WAL、5 秒 busy_timeout，连接数限制为 1。启动先检查版本，再 Sync2 补表／补列及执行专用迁移；更高版本数据库会被拒绝。

升级前停止服务，备份数据库、尚未合并的 WAL、原密钥及仓库。直接降级 EXE 不受支持，回退须恢复对应快照。数据库和加密候选配置均为敏感数据，不能提交源码仓库。

## 模型分组

| 表 | 用途与约束 |
| ---- | ---- |
| user、session | 本地／LDAP 用户、启停与管理员标记、邮箱、创建和最后登录；会话与 CSRF；密码／盐／令牌不通过普通 JSON 输出 |
| organization、membership | 组织资料和成员；组织＋用户唯一，旧 Owner 位不作为当前授权来源 |
| team、team_member、team_repository | 内置 Owners、团队成员和整仓库授权；成员及授权组合唯一 |
| repository | 类型、状态、UUID／修订、远端、加密凭据、调度、最近成功、完整性停止、大小缓存；组织＋名称唯一 |
| task、task_execution | 持久任务及每次实际执行；触发方式、阶段／进度、当前尝试时间、前后修订和结果快照 |
| mirror_conversion | 以 repository_id 为主键的转换恢复快照，包含原 r0 属性，不含服务密码 |
| repository_path_move | 路径迁移恢复记录，保存原／新路径、身份及阶段所需标记 |
| ldap_config、ldap_schedule、ldap_sync_run | 唯一 LDAP 配置（固定主键 1，触发器约束）、cron 和同步历史；历史不含源编号／名称 |
| notification_config、notification_delivery、notification_attempt | 单机器人配置、事务 Outbox 和发送尝试；EventKey 唯一，凭据加密且不回显 |
| site_setting | ID=1 的站点设置 |
| revision_event、activity_index | 保留的修订索引兼容表；当前首页不再采集或使用 |
| migration | 已应用 schema ID 与应用时间 |

密码与密钥的处理见[架构与恢复](architecture.md)。Repository.Path、Password 和 Task.Source 不通过普通 JSON 返回。Task.Source 在转换任务中保存候选配置 JSON，其中密码已加密。

## 时间与进度

用户、仓库、task 及配置时间通常为 Unix 秒；task_execution、LDAP 同步历史和通知投递／尝试使用毫秒时间，LDAP 历史另有 DurationMS。API 不可混用单位。旧用户创建／登录为 0 时显示未知，不推测历史；业务页面将有效时间格式化为北京时间。

Task 的 kind 为 local／mirror／sync／conversion，state 为 queued／running／succeeded／failed。task_execution 另有 interrupted，每次执行独立 ID。任务重试保留 TaskID，重新记录实际开始时间并新建执行记录。ActorID／ActorName 用于转换的权限复查与操作者快照。

Verification、VerificationPosition、ProgressStarted、ProgressUpdated 表示校验或属性阶段。位置 -1 表示尚无完成通知；元数据位置不是已完成数量。属性阶段 verification=revprops，total_revision 为阶段开始时本地 HEAD，数量和百分比包含 r0。published 是本地迁入恢复标记，不能被普通进度覆盖。

Repository.Size 为目录普通文件长度之和，SizeUpdated 为最近成功统计时间，SizeChecked 为最近尝试；SizeUpdated=0 为未知，真实 0 字节有非零成功时间。旧列 NULL 启动规范化为 0。统计失败只更新尝试时间，不覆盖旧成功值或业务结果。

## 事务边界

- 创建组织时原子创建成员和内置 Owners；迁移旧 Owner 关系后不再依赖旧角色位。
- 本地用户删除获取写锁后复查操作者、目标及最后活跃管理员／Owners；原子删除 session、team_member、membership、user，保留组织、团队、仓库和历史。
- LDAP 配置清理与保存同事务，确认令牌绑定候选配置和待删关系；成功同步的用户更新、最近结果和历史原子提交。失败保留用户原状，安全失败历史独立保存。
- 普通任务最终结果、task_execution 和通知入队同事务。转换原子提交镜像配置、转换成功记录、首次同步入队及恢复日志删除，避免崩溃后重复初始化。
- 通知发送为事务外网络操作，投递与每次尝试独立持久化。pending／sending／retry／success／failed／cancelled 为投递状态；中断发送按不确定结果处理。

## 兼容迁移

以下保留结构兼容事实，不作为按日期排列的实施日志：

| schema | 兼容作用 |
| ---- | ---- |
| 1–2 | 基础模型及旧组织所有者迁入 Owners，缺少合法继任者时拒绝迁移 |
| 3–4 | 仓库路径迁移保护及本地转镜像恢复记录 |
| 5–9 | 旧 LDAP 多源模型；其迁移实现已移除，升级边界见 schema 15 |
| 10 | 通知配置／Outbox／尝试、任务执行记录 |
| 11 | 全站 LDAP 调度；当前 Time 字段为五字段 cron，默认 0 0 * * *，不接受旧 HH:MM |
| 12 | 每次仓库任务执行的触发方式／错误快照；历史缺失值不伪造 |
| 13 | 持久后台转换和操作者信息，旧程序不能误解释 conversion 任务 |
| 14 | 删除 notification_config.excluded_repositories；不恢复 cancelled、不生成历史通知，保留其他配置／凭据／代次和历史 |
| 15 | LDAP 单配置、用户 ldap_account 布尔标识；移除旧源编号／名称。仅支持旧 LDAP 配置、用户和历史均为空的库升级，否则在修改 schema 前拒绝 |

部分附加字段由 Sync2 增量创建，不独立提高 schema，例如用户时间、组织资料、任务进度及仓库大小。具体迁移顺序以 models 中实现为准，不能手工修改版本号绕过恢复或兼容检查。

任务中心复用 task 主键、repository_id 索引及现有授权关系，不新增列或迁移，不回填任务创建时间；Started 仍表示当前尝试时间。

LDAP 用户以 `ldap_account` 区分本地账号，以 `ldap_key` 绑定稳定身份；不再保存认证源外键。生产升级不删除其他业务数据。

活动与任务结果概要复用 task_execution 的 before_revision、after_revision、revision_known，不新增 schema 或回填旧记录。活动按每次执行读取，任务按最大执行 ID 读取最新尝试；无可靠快照时仅显示通用完成状态。
