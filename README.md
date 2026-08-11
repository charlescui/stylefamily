# 风格家 StyleFamily — 家庭穿搭智能推荐系统

基于 PocketBase + Golang + Bailian CLI 的家庭穿搭智能推荐系统，为家庭每个成员每周自动生成适合季节和场景的穿搭方案，并通过多维度考评、虚拟试穿、电视大屏展示完成完整闭环。

## 核心功能

- **家庭成员管理**：通过 Web UI 管理每位家庭成员的身材、风格偏好、颜色偏好等信息，并支持更新。
- **智能穿搭推荐**：每周六早上 7 点自动为家庭成员生成适合季节和场景的穿搭方案。
- **多维度考评**：从季节适配、场合适配、色彩协调、风格一致、搭配完整度五个维度评分。
- **虚拟试穿效果**：调用 Bailian 图像生成能力，为每个通过的方案生成试穿效果图。
- **电视大屏展示**：内置 `/tv.html` 全屏展示页，无需翻页，适合家庭电视。
- **MCP 工具接口**：对外暴露 `stylefamily_generate_look`、`stylefamily_feedback`、`stylefamily_get_result` 等 MCP 工具。
- **中文文档自动生成**：每次生成穿搭后自动写入 `docs/` 目录。

## 项目结构

```
stylefamily/
├── main.go                         # 应用入口
├── pkg/
│   ├── bailian/                    # Bailian CLI 封装
│   ├── mcp/                        # MCP 工具接口
│   └── familyoutfit/               # 家庭穿搭核心模块
│       ├── repository.go           # PocketBase 数据访问
│       ├── generator.go            # 穿搭生成与考评
│       ├── tryon.go                # 试穿图生成
│       ├── scheduler.go            # 定时任务
│       └── web.go                  # Web API + Demo 数据
pb_public/                          # 前端静态页面
├── index.html                      # 家庭成员管理
├── tv.html                         # 电视端展示
├── style.css                       # 样式
└── app.js                          # 交互脚本
migrations/                         # PocketBase 数据迁移
```

## 构建

```bash
cd /home/cuizheng/Projects/stylefamily
source /home/cuizheng/.local/go_env.sh
go build -o /tmp/stylefamily ./stylefamily
```

## 运行

```bash
/tmp/stylefamily serve --http=0.0.0.0:8090
```

## 访问

- 管理页：`http://<服务器>:8090/index.html`
- 电视展示页：`http://<服务器>:8090/tv.html`

## GitHub Pages

项目展示页：https://charlescui.github.io/stylefamily/

## MCP 测试

```bash
curl -X POST http://127.0.0.1:8090/mcp/tools/list
```

## 定时任务

系统内置定时调度，每周六 07:00 自动生成本周穿搭方案。也可以手动触发：

```bash
/tmp/stylefamily generate-weekly
```

## 技术栈

- **后端**：PocketBase (Go)
- **LLM / 图像生成**：阿里云 Bailian CLI (`bl`)
- **前端**：原生 HTML/CSS/JS
- **数据库**：SQLite (PocketBase 内置)

## 品牌

- **中文**：风格家
- **英文**：StyleFamily

## 作者

charlescui
