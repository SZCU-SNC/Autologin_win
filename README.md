# Autologin_win
校园网自动登录的windows服务端

# 自动登录校园网

## 功能特点

- 支持配置登录信息和自动登录选项
- 支持自动检查是否需要登录，并自动登录
- 支持通过命令行参数启动和停止自动登录功能
- 支持保存配置项到文件

## 使用方法

1. 前往[Releases · SZCU-SNC/Autologin_win (github.com)](https://github.com/SZCU-SNC/Autologin_win/releases/)下载最新版本.exe文件
2. 将`Autologin.exe`移动至任意目录（建议C盘）
3. 创建`Autologin.exe`的快捷方式
4. 资源管理器打开`%APPDATA%\Microsoft\Windows\Start Menu\Programs\Startup`
5. 将快捷方式移动进入打开的文件夹。
6. 手动运行快捷方式。
7. 打开 [登录状态检查](http://localhost:1580/) http://localhost:1580/
8. 填写配置项
9. 选择网卡
10. 保存配置，enjoy it!

## 注意事项

- 程序需要连接校园网才能正常工作
- 账号密码对应校园网账号密码
- 检查间隔意思是隔多少秒检查一次电脑是否联网，然后会自动尝试登录请求
- 如果不勾选**自动登录**那么也不会进行自动登录
- 网卡选择请自行查看当前联网网卡名称，如果是无线网一般选`wifi`，有线网一般为`以太网`
