# rock-analyzer-go

rock-go框架系统的日志分析组件，获取到输入数据后，通过lua脚本处理分析。

# 使用说明

## 导入

```go
import analyzer "github.com/rock-go/rock-analyzer-go"
```

## 注册

```go
rock.Inject(xcall.Rock, analyzer.LuaInjectApi)
```

## lua脚本调用

分为两部分，启动和分析脚本。启动脚本用于配置和启动该模块，分析脚本用于该模块调用来处理接收到的数据

### 启动

```lua
-- 分析模块
local log_analyzer = rock.log_analyzer {
    name = "log_analyzer",
    thread = 5,
    input = kafka_consumer_analyzer,
    script = "resource/script/analyzer",
    heartbeat = 10,
}

proc.start(log_analyzer)

-- console命令
-- .start()启动
-- .close()关闭
```

#### 参数说明

- name: 模块名称，用于日志标识和模块标识

- thread: 线程数，默认为1

- input: 数据来源接口，例如kafka消费者，elasticsearch查询接口

- script: 分析脚本的路径，可以为目录或单个脚本

- heartbeat: 健康检查心跳时间，单位为秒

### 分析脚本

模块实现了一些数据处理的函数，主要处理json和字符串格式的数据。<br>
下面的例子解析了一条vpn日志，输出结果为content中的来源IP地址和用户名。

```json
{
  "@version": "1",
  "facility": 23,
  "content": "2021-07-09 09:42:52 USG6650 %%01USERS/4/USRPWDERR(l): id=USG6650 time=\"2021-07-09 17:42:16\" fw=USG6650 pri=4 vsys=root vpn=sslvpn user=\"user1\" src=180.169.1.1 dst=0.0.0.0 duration=36s rcvd=0byte(s) sent=0byte(s) type=vpn service=5 msg=\"Session: user1 failed to login.\"",
  "priority": 188,
  "tag": "",
  "tls_peer": "",
  "severity": 4,
  "timestamp": "2021-07-09T17:44:12+08:00",
  "hostname": "192.168.0.1",
  "@timestamp": "2021-07-09T09:44:09.602Z",
  "client": "192.168.0.1:3456"
}
```

```lua
-- 获取上述json的content字段中的ip和username
function parse()
    -- 引入数据解析模块
    local parser = rock.analyzer.parser
    -- 声明分析函数
    -- 获取数据
    local msg = parser.msg
    -- 判断是否包含某个字符串
    local contain = parser.contain
    -- 将数据解析成json
    local parse_json = parser.parse_json
    -- 获取json字段
    local json_get = parser.json_get
    -- 获取slice中的值
    local slice_get = parser.slice_get
    -- 分割字符串
    local split = parser.split
    -- byte转化为字符串
    local b2s = parser.b2s

    -- 数据处理逻辑
    -- 获取原始数据
    local data = msg()
    -- 判断是否包含某个字符串
    if contain(data, "failed to login") == false and contain(data, "login failed") == false then
        return nil
    end
    -- 解析json中相应的字段，存储为map
    local obj_map = parse_json(data, "content")
    if obj_map == nil then
        return nil
    end
    -- 获取上面解析的字段值
    local content = json_get(obj_map, "content")
    if content == nil then
        return nil
    end
    -- 解析分割上面获取的值，返回数组
    local obj_slice = split(content, " ")
    if obj_slice == nil then
        return nil
    end
    -- 获取slice中的值
    local src_ip_b, user_b
    src_ip_b, user_b = slice_get(obj_slice, 12), slice_get(obj_slice, 20)
    -- 将结果转化为字符串
    print(b2s(src_ip_b), b2s(user_b))
    -- 执行结果为：180.169.1.1，user1
end

-- 注册和回调
rock.analyzer.callback(parse)
```

#### 函数说明

下列函数中，只有b2s()返回的为string类型，其它的均为userdata

|  函数   | 参数  |   返回   |功能|
  |  ----  | ----  | ----    |----|
| msg()  | 无 |类型：userdata<br>go中值为[]byte|获取原始数据|
| contain(param1,param2)  | param1:值为[]byte的userdata<br>param2:字符串 |布尔值|判断param1中是否包含param2|
|parse_json(param1,pram2,param3...)|param1:userdata,值为[]byte<br>param2,param3...:值为字符串，可多个|userdata，值为Parser{}中的chunkMap|获取param1中的param2,param3等的值，存入userdata中|
|json_get(param1,param2)|param1:parse_json函数返回的值<br>param2:字符串，需要返回值的字段|userdata，值为[]byte|从param1中获取param2字段的值|
|split(param1,param2)|param1:userdata,值为[]byte<br>param2:分割符|userdata,值为Parser{}中的chunkSlice|以param2分割param1|
|slice_get(param1,param2)|param1:userdata,值为split()函数的返回值<br>param2:index|userdata，值为[]byte|获取字符串分割后对应的值|
|b2s(param1)|param1:userdata,值为[]byte|string|将最终结果转为string|