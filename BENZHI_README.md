# AquaFarm - RAS Monitor

循环水养殖系统 (Recirculating Aquaculture System) 管理服务。

## 功能
- 鱼池管理（创建/查询/更新/删除）
- 水质传感器数据采集（温度/pH/溶氧/氨氮/亚硝酸盐）
- 阈值告警（自动检测越限并生成告警）
- 投喂计划（自动/手动投喂 + 投喂日志）
- 设备管理（泵/过滤器/增氧机/加热器 + 维护任务）
- 鱼苗批次管理（批次入库/死亡记录/存活率）
- 换水记录
- 系统监控面板（前端页面）

## 技术栈
- Go 1.22 + Gin
- SQLite (modernc.org/sqlite, 纯 Go 驱动)
- 前端: 原生 HTML/JS/CSS

## 运行
```bash
go build -o aquafarm .
./aquafarm
# 默认监听 :8585
```

## Docker 构建
```bash
./build_benzhi_docker.sh aquafarm linux/amd64
./build_benzhi_docker.sh aquafarm linux/arm64
```
