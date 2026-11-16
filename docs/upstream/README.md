# 上游来源与许可证

## Gitea 基线

来源为 [Gitea](https://github.com/go-gitea/gitea)，固定提交 [`146cc3eec57174711eac0e0a0c7b38670c6e3922`](https://github.com/go-gitea/gitea/tree/146cc3eec57174711eac0e0a0c7b38670c6e3922)，MIT。根目录 [LICENSE](../../LICENSE) 及分发的 Gitea／Gogs 许可证保留原版权。参考源码应在项目外单独检出，不依赖开发者机器路径。

复用范围见[复用清单](reuse-manifest.md)：密码哈希、identicon、主题／CSS、模板布局、下拉模块及局部热力图。复制或适配文件保留其原版权／来源注释；SVNova 的 SVN 服务、任务、授权投影和事务恢复为本项目实现，不提供上游 Git 业务。不能把采用相同布局的模板宣称为原样复制。

## 前端依赖与图片

| 来源 | 固定版本／范围 | 归属 |
| ---- | ---- | ---- |
| @primer/octicons | 19.28.1；Icon.vue 和构建时生成的 SVG sprite | [MIT](octicons-LICENSE.txt)，许可证随资源和 JS 分发 |
| Apache ECharts / ZRender / tslib | ECharts 6.1.0、ZRender 6.1.0、tslib 2.3.0；活动趋势按需加载 | [ECharts Apache-2.0](echarts-LICENSE.txt)、[NOTICE](echarts-NOTICE.txt)、[d3 BSD](echarts-d3-LICENSE.txt)、[ZRender BSD](zrender-LICENSE.txt)、[tslib 0BSD](tslib-LICENSE.txt)，随静态资源分发 |
| jQuery | 运行库 4.0.0，以 package.json／锁文件为准 | [MIT](jquery-LICENSE.txt)，用于下拉组件 |
| Gitea Fomantic 分叉 | 固定 Gitea 基线内的 dropdown／transition／辅助模块，保留 GITEA-PATCH 和原声明 | 源码保留第三方许可声明 |
| Gitea 图片 | repo_default、loading、failed、checkmark 等保留资源 | [Gitea／Gogs MIT](../../static/img/gitea-LICENSE.txt)；当前迁移面板使用 SVNova 标识 |

SVNova Logo／favicon 为项目品牌资源。界面其余图标使用固定 Octicons，不替换为其他图标集。Vue 和构建依赖以 package.json／pnpm-lock.yaml 为准；Go 依赖以 go.mod／go.sum 为准，第三方许可证不得随文档清理删除。

## SVN 工具包

Windows 工具来自 VisualSVN 官方的 [Apache Subversion 1.15.0 x64 再分发包](https://www.visualsvn.com/files/Apache-Subversion-1.15.0-x64.zip)。内嵌文件为 `internal/bundle/payload/svn.zip`，构建脚本和运行释放逻辑核验固定 SHA256：

```text
98bb798dd6c70cc875b9bdacd600594351cf88f0a39a831a5055df061c871bb5
```

包内各组件版权和许可文本保存在 [svn-licenses](svn-licenses/) 中，并随内嵌工具释放；Subversion 及其依赖不统一改为本项目 MIT。工具不作为 SVN 协议服务运行。
