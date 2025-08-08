# NTP 时间同步工具

一个基于 Go 语言开发的 NTP 时间同步工具，提供现代化的 Web 界面和毫秒级精度的时间显示。

## 功能特性

- 🎯 **毫秒级精度** - 实时显示 NTP 时间，精确到毫秒
- 🌐 **多服务器支持** - 预置 10 个 NTP 服务器快速选择
- 📊 **实时监控** - 显示网络延迟和时间偏移信息
- ⚡ **系统时间同步** - 支持 Windows 系统时间同步（需管理员权限）
- 🚀 **轻量快速** - 单文件运行，启动即用

## 预置 NTP 服务器

- **腾讯云** - time.cloud.tencent.com ⭐ 默认
- **教育网NTP** - time.edu.cn
- **阿里云** - ntp1.aliyun.com
- **中国NTP** - cn.ntp.org.cn
- **Cloudflare** - time.cloudflare.com
- **苹果亚洲** - time.asia.apple.com
- **NTP Pool** - pool.ntp.org
- **微软时间** - time.windows.com
- **谷歌时间** - time.google.com
- **美国NIST** - time.nist.gov

## 快速开始

### 方式一：下载 Release（推荐）
1. 前往 [Releases](../../releases) 页面
2. 下载适合你系统的可执行文件
3. 直接运行 `ntp-timer.exe`

### 方式二：从源码运行
```bash
go run main.go
```

程序启动后会自动打开浏览器访问 http://localhost:8080

## 使用说明

1. **选择 NTP 服务器**：点击预置服务器按钮或手动输入服务器地址
2. **同步时间**：点击"同步时间"获取 NTP 时间和偏移信息
3. **系统时间同步**：点击"同步到系统时间"将系统时间调整为 NTP 时间（需管理员权限）

## 许可证

MIT License


