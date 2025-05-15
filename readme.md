# 秦务员埋点工具

前端传参 ---> 后端接口接收 ---> redis暂存 ---> 1分钟后分类存储进入mysql数据库

![埋点.drawio](./readme.assets/埋点.drawio.svg)

## 前端传入数据类型

传入客户端类型、网站、事件类型、详细事件

```json
{
  "client_type": "web",             // 客户端类型：PC/APP/Miniprogram 等
  "site": "xxx",            		// 当前网站或应用标识
  "event_type": "button_click",      // 事件类型：button_click等
  "event_detail": "bszn", 			// 详细事件信息，结构由事件类型决定
  "user_detail": "xxx",				//用户标识
  "time_stamp": "",
}
```

## 后端 go+gin+redis+mysql

**gin**实现接口

redis暂存接口数据 测试密码：redis

1分钟后异步mysql存入统计结果数据 测试账密：exam/exam123

mysql数据库 **track**

表格：

```mysql
CREATE TABLE event_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    timestamp BIGINT NOT NULL,                    -- 每条记录的后端写入时间（按分钟对齐）
    client_type VARCHAR(20) NOT NULL,
    site VARCHAR(100) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_detail VARCHAR(255) NOT NULL,
    count INT DEFAULT 1
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

```

## 前端传入方式

```json
{
    "data": "AES加密参数，ECB模式，pkcs7padding"
}
```

`nohup ./Tracking_tool-linux-amd64 > ./run.log 2>&1 &`