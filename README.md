# StyleTailor MCP Server + 家庭穿搭智能推荐系统

这是一个基于 PocketBase + Golang + Bailian CLI 的项目，融合了 StyleTailor 虚拟试穿 MCP 能力和家庭穿搭自动推荐系统。

## 核心功能

- **虚拟试穿 MCP**：对外暴露 `styletailor_generate_look`、`styletailor_feedback`、`styletailor_get_result` 等 MCP 工具，支持个性化穿搭生成和负面反馈迭代。
- **家庭成员管理**：管理家庭成员的身材、风格偏好、颜色偏好、虚拟形象描述等。
- **智能穿搭推荐**：每周六早上 7 点自动为家庭成员生成适合季节和场景的穿搭方案。
- **多维度考评**：季节适配、场合适配、色彩协调、风格一致、搭配完整度评分。
- **虚拟试穿效果图**：为每个通过的方案调用 Bailian 生成试穿图。
- **电视大屏展示**：内置 `/tv.html` 全屏展示页，无需翻页，适合家庭电视。
- **中文文档自动生成**：每次生成穿搭后自动写入 `docs/` 目录。

## 项目结构

```
styletailor/
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
├── 20260811_create_styletailor_tables.go
├── 20260811_import_seed_products.go
└── 20260812_add_family_outfit_tables.go
```

## 构建

```bash
source /home/cuizheng/.local/go_env.sh
cd /home/cuizheng/Projects/styletailor-mcp
go build -o /tmp/styletailor-mcp ./styletailor
```

## 运行

```bash
/tmp/styletailor-mcp serve --http=0.0.0.0:8090
```

## 访问

- 管理后台：`http://<服务器>:8090/index.html`
- 电视展示页：`http://<服务器>:8090/tv.html`

## MCP 测试

```bash
curl -X POST http://127.0.0.1:8090/mcp/tools/list
```

## 定时任务

系统内置定时调度，每周六 07:00 自动生成本周穿搭方案。也可以手动触发或通过 cron 调用：

```bash
/tmp/styletailor-mcp generate-weekly
```

## 技术栈

- **后端**：PocketBase (Go)
- **LLM / 图像生成**：阿里云 Bailian CLI (`bl`)
- **前端**：原生 HTML/CSS/JS
- **数据库**：SQLite (PocketBase 内置)

## 作者

charlescui
