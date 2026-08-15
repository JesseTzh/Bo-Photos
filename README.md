# BoPhoto

BoPhoto 是一个自托管摄影作品集和图库管理系统。它提供公开作品展示、相册、图文指南、后台管理、图片上传处理和访问统计，适合个人摄影站点或小型作品集站点部署。

## 主要功能

- 公开首页、图库、相册、封面、图片与视频预览和关于页面。
- 后台管理图片、视频、相册、标签、站点设置和账户密码。
- 上传图片后自动提取信息并生成预览图和缩略图；视频支持 MP4、WebM 和 MOV。
- 序章封面可指定图片或自动播放视频，并可独立关闭封面文字。
- 记录公开访问统计，并在后台展示概览和分析数据。
- 数据和媒体文件都保存在本地目录，部署简单。

## 本地开发

准备环境：

- Go 1.24 或更高版本。
- Node.js 20 或更高版本。
- pnpm 11.9.0。
- 本地图片处理需要安装 `vipsthumbnail` 和 `exiftool`。

安装前端依赖：

```bash
cd frontend
pnpm install
```

启动后端：

```bash
BOPHOTOS_DATA_DIR="$PWD/data" \
BOPHOTOS_INITIAL_PASSWORD="change-this-to-a-strong-password" \
make dev-backend
```

另开一个终端启动前端：

```bash
make dev-frontend
```

访问 <http://localhost:5173>。首次启动空数据目录时，会使用 `BOPHOTOS_INITIAL_PASSWORD` 创建管理员密码。

运行检查：

```bash
make check
```

## Docker 部署

设置初始管理员密码并启动：

```bash
BOPHOTOS_INITIAL_PASSWORD="change-this-to-a-strong-password" \
docker compose up -d --build
```

也可以在 `.env` 中设置：

```text
BOPHOTOS_INITIAL_PASSWORD=change-this-to-a-strong-password
```

然后执行：

```bash
docker compose up -d --build
```

服务启动后访问 <http://localhost:8080/login> 登录后台。

持久化数据保存在宿主机 `./data`，容器内路径为 `/data`。首次创建管理员后，后续启动不会用环境变量覆盖已有密码。

HTTPS 反向代理部署时，请设置：

```yaml
environment:
  BOPHOTOS_COOKIE_SECURE: "true"
```

## 常用配置

| 环境变量 | 默认值 | 说明 |
|---|---|---|
| `BOPHOTOS_ADDR` | `:8080` | 服务监听地址 |
| `BOPHOTOS_DATA_DIR` | `/data` | 数据和媒体文件目录 |
| `BOPHOTOS_FRONTEND_DIR` | `frontend/dist` | 前端构建产物目录 |
| `BOPHOTOS_COOKIE_SECURE` | `false` | HTTPS 部署时设为 `true` |
| `BOPHOTOS_MAX_UPLOAD_BYTES` | `2147483648` | 单文件上传大小上限 |
| `BOPHOTOS_INITIAL_PASSWORD` | 无 | 空数据目录首次启动时的管理员密码 |

`BOPHOTOS_INITIAL_PASSWORD` 在需要时必须为 12 到 128 个字符。

## 更多文档

- [系统架构](docs/architecture/system-architecture.md)

## 来源

BoPhoto 重构自 [sourcexu7/XPhotos](https://github.com/sourcexu7/XPhotos)。

## 许可证

本项目基于 [MIT](LICENSE) 许可证开源。
