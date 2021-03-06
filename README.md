# Datahub-Client
----------------
###简介
Datahub-Client实现了DataHub CLI和DataHub daemon两部分功能。Datahub daemon是常驻进程，接收DataHub CLI发送的命令，完成命令的后台执行。


### 开始

在安装GO(需要go1.4以上版本)语言和设置了[GOPATH](http://golang.org/doc/code.html#GOPATH)环境变量之后，安装DataHub daemon：

```shell
go get github.com/asiainfoLDP/datahub
```

启动datahub daemon服务:
```shell
sudo $GOPATH/bin/datahub --daemon --token xxxxxxxxxxx
```

### Docker方式启动

```shell
docker build -t datahub .
docker run -d -e "DAEMON_TOKEN=xxxxxxxxxxx" -e "DAEMON_ENTRYPOINT=http://XXXXXXXXXX:YYYY" -p xxxx:35800 datahub
```

### 运行Datahub CLI

Datahub CLI是datahub的命令行客户端，用来输入datahub相关命令。

- dp
    - Datapool管理
- subs
    - Subscrption管理
- login
    - 登录到hub.dataos.io
- pull
    - 下载数据
- pub
    - 发布数据
- repo
    - Repository管理
- job
    - 显示任务列表
- ep
    - 设置Entrypoint
- logout
    - 登出
- help
    - 帮助命令

### Datahub CLI命令行使用说明
---
#### NOTE：
- 如果没有额外说明，所有的命令在没有错误发生时，不在终端输出任何信息，只记录到日志中。错误信息会打印到终端。
- 所有的命令执行都会记录到日志中，日志级别分[TRACE] [INFO] [WARNNING] [ERROR] [FATAL]。
- 参数支持全名和简称两种形式，例如--type等同于-t。详情见命令帮助。
- 参数赋值支持空格和等号两种形式，例如--type=file等同于--type file。

#### 1. Datapool相关命令

##### 1.1. 列出所有命令池

```shell
datahub dp
```
输出
```shell
{%DPNAME    %DPTYPE}
```
例子
```shell
$ datahub dp
DATAPOOL            TYPE
------------------------
dp1                 file 
dp2                 db2
dphdfs              hdfs
dps3                s3
$
```

##### 1.2. 列出Datapool详情

```shell
datahub dp $DPNAME
```
输出
```shell
DATAPOOL:%DPNAME                  %DPTYPE         %DPCONN
{%REPOSITORY/%DATAITEM:%TAG       %LOCAL_TIME     %T}
```
例子
```shell
$ datahub dp dp1
DATAPOOL:dp1            file                      /var/lib/datahub/dp1
repo1/item1:tag1        2015-10-23 03:57:42       pub		repo1_item1       tag1.txt    		Size:232.00KB
repo1/item1:tag2        2015-10-23 03:59:49       pub		repo1_item1	  tag2
repo1/item2:jinrong-40  2015-10-23 04:01:22       pull		item2location	  jinrong_40.txt	金融信息
cmcc/beijing:jiangsu-lac-ci     2015-11-19 10:57:21       pull		cmcc_beijing	jiangsu-lac-ci.txt   位置区编码
$ 
```
说明：cmcc_beijing为DataItem beijing在Datapool dp1中的位置， jiangsu-lac-ci.txt为tag存储到dp1中的文件名，“位置区编码”为详细信息。

##### 1.3. 创建数据池

- 目前只支持本地目录形式的数据池创建。

```shell
datahub dp create $DPNAME [[file://][ABSOLUTE PATH]] | [[s3://][BUCKET]] | [[hdfs://][USERNAME:PASSWORD@HOST:PORT]]
```
输出
```
%msg
```
例子 1
```
$ datahub dp create testdp file:///var/lib/datahub/testdp
DataHub : Datapool has been created successfully. Name:testdp Type:file Path:/var/lib/datahub/testdp.
$
```

例子 2
```
$ datahub dp create s3dp s3://mybucket
DataHub : s3dp already exists, please change another name.
$
```
说明：mybucket是s3上已存在的bucket。另外，需要在启动daemon的系统中设置环境变量：AWS_SECRET_ACCESS_KEY， AWS_ACCESS_KEY_ID， AWS_REGION。

例子 3
```
$ datahub dp create hdfsdp hdfs://user123:admin123@x.x.x.x:9000
```
说明：“hdfs://”后需要接hdfs的连接串。

##### 1.4. 删除数据池

- 删除数据池不会删除目标数据池已保存的数据。该dp有发布的数据项时，不能被删除。删除是在sqlite中标记状态，不真实删除。

```shell
datahub dp rm $DPNAME
```
输出
```
%msg
```
例子
```
$ datahub dp rm testdp
DataHub : Datapool testdp removed successfully!
$
```

#### 2. subs相关命令

##### 2.1. 列出所有已订阅项

```shell
datahub subs 
```
输出
```
REPOSITORY/DATAITEM               TYPE    STATUS
{%REPOSITORY/%DATAITEM            %TYPE   online/offline}
```
例子
```
$ datahub subs
REPOSITORY/DATAITEM     TYPE    STATUS
cmcc/beijing        file    online
repo1/testing       hdfs    online
$
```

##### 2.2. 列出用户在某个Repository下已订阅的DataItem

```shell
datahub subs %REPOSITORY
```
输出
```
REPOSITORY/DATAITEM           TYPE      STATUS
%REPOSITORY/%DATAITEM         %TYPE     %STATUS

```
例子
```
$ datahub subs cmcc
REPOSITORY/DATAITEM    TYPE      STATUS
cmcc/beijing       file      online
cmcc/Shanghai      file      offline
$
```
##### 2.3. 列出已订阅DataItem详情
```shell
datahub subs %REPOSITORY/%DATAITEM
```
输出
```
REPOSITORY/DATAITEM:TAG            UPDATETIME      COMMENT      STATUS
%REPOSITORY/%DATAITEM:%TAGNAME     %UPDATE_TIME    %COMMENT     %STATUS
```
例子
```
$ datahub subs cmcc/beijing
REPOSITORY/DATAITEM:TAG      UPDATETIME              COMMENT      STATUS
cmcc/beijing:chaoyang       15:34 Oct 12 2015       600M         NORMAL
cmcc/beijing:daxing         16:40 Oct 13 2015       435M         NORMAL
cmcc/beijing:shunyi         16:40 Oct 14 2015       324M         NORMAL
cmcc/beijing:haidian        16:40 Oct 15 2015       988M         NORMAL
$
```

#### 3. pull命令

##### 3.1. 拉取某个DataItem的Tag
- pull一个tag，需指定`$DATAPOOL`, 可再指定`$DATAPOOL`下的子目录`$LOCATION`，默认下载到`$DATAPOOL://$REPOSITORY_$DATAITEM`. 
可选参数:
[--destname, -d]命名下载的Tag
[--automatic, -a]自动下载已订阅的DataItem新增的Tag
[--cancel, -c]取消自动下载Tag

```shell
datahub pull %REPOSITORY/%DATAITEM:$TAG $DATAPOOL[://$LOCATION] [--destname，-d]
```
输出
```
%msg
```
例子
```
$ datahub pull cmcc/beijing:chaoyang dp1://cmccbj
OK.
$
```

#### 4. login命令

- login命令支持被动调用，用于DataHub client与DataHub server交互时作认证。并将认证信息保存到环境变量，免去后续指令重复输入认证信息。

##### 4.1. 登录到hub.dataos.io

```shell
datahub login URL
```
输出
```
%msg
```
例子
```
$ datahub login https://hub.dataos.io
login as: datahub@gmail.com
password: *******
Error : login failed.
$
```

#### 5. pub相关命令
- pub分为发布一个DataItem和发布一个Tag。
- 发布DataItem必须指定`$DATAPOOL`和`$DATAPOOL`下的子路径`$LOCATION` ,  可选参数 --accesstype, -t= 指定DataItem属性：public, private, 默认private
- 发布Tag必须指定TAGDETAIL , 用来指定Tag对应文件名，该文件必须存在于$DATAPOOL://$LOCATION内
- 可选参数--comment, -m= ,描述DataItem或者Tag

##### 5.1. 发布一个DataItem

```shell
datahub pub $REPOSITORY/$DATAITEM $DATAPOOL://$LOCATION --accesstype=public [private]  [--comment, -m]
```
输出
```
Pub success,  OK
```
例子
```
$./datahub pub music_1/migu mydp://dirmigu --accesstype=public --comment="migu music desc"
Pub success,  OK
```

##### 5.2.发布一个Tag

```shell
datahub pub %REPOSITORY/%DATAITEM:$Tag $TAGDETAIL --comment=" "
```
说明：如果DataItem已经在网页上发布过了，那么在发布Tag的时候需要指定`$DATAPOOL`和`$DATAPOOL`下的子路径`$LOCATION`
```shell
datahub pub %REPOSITORY/%DATAITEM:$Tag $TAGDETAIL $DATAPOOL://$LOCATION --comment=" "
```
输出
```
Pub success, OK
```
例子
```
$ datahub pub music_1/migu:migu_user_info migu_user_info.txt
Pub success, OK

$ datahub pub music_1/migu:migu_user_info migu_user_info.txt mydp://dirmigu
Pub success, OK
```

#### 6. repo命令

##### 6.1. 查询自己创建的和具有写权限的所有Repository

```shell
datahub repo 
```
输出
```
REPOSITORY
--------------------
Location_information
Internet_stats
Base_station_location
```

##### 6.2. 查询Repository的详情

```shell
datahub repo Internet_stats
```
输出
```
REPOSITORY/DATAITEM
--------------------------------
Internet_stats/Music
Internet_stats/Books
Internet_stats/Cars
Internet_stats/Ecommerce_goods
Internet_stats/Film_and_television
```
##### 6.3. 查询DataItem的详情

```shell
datahub repo Internet_stats/Music
```
输出
```
REPOSITORY/DATAITEM:TAG	                    UPDATETIME	                COMMENT
------------------------------------------------------------------------------------
Internet_stats/Music:music_baidumusic_6008	2016-03-04 09:15:18|6天前	百度音乐
Internet_stats/Music:music_qqmusic_6001		2016-02-03 09:23:30|1个月前	QQ音乐
Internet_stats/Music:music_kuwomusic_6005	2016-01-06 09:35:44|2个月前	酷我音乐
```

##### 6.4. 删除自己创建的DataItem

```shell
datahub repo rm myrepo/myitem
```
输出
```
Datahub : After you delete the DataItem, data could not be recovery, and all tags would be deleted either.
Are you sure to delete the current DataItem?[Y or N]:Y
DataHub : OK
```
说明：当此DataItem下有正在生效的订购计划时，会提示资费回退规则。

##### 6.5. 删除自己创建的Tag

```shell
datahub repo rm FavouriteMusic/MusicItem:bingyu
```
输出
```
DataHub : After you delete the Tag, data could not be recovery.
Are you sure to delete the current Tag?[Y or N]:y
DataHub : OK
```

#### 7. job命令

##### 7.1. job查看所有任务列表，包括数据下载和发送的任务

```shell
datahub job
```
##### 7.2. job查看某个任务Id对应的信息

```shell
datahub job &JOBID
```

##### 7.3. job rm删除某个job

```shell
datahub job rm &JOBID
```

#### 8. ep命令

- 设置Datahub daemon的Entrypoint，作为数据提供方，需要提供可访问的url，供需求方访问，并下载数据。
- 此命令也可以用来查看是否设置了Entrypoint。

#### 9. logout命令

- 登出hub.dataos.io

#### 10. help命令

- help提供datahub所有命令的帮助信息。

##### 10.1. 列出帮助

```shell
datahub help [$CMD] [$SUBCMD]
```
输出
```
Usage of %CMD %SUBCMD
{  %OPTION=%DEFAULT_VALUE     %OPTION_DESCRIPTION}
```
例子
```
$datahub help dp
Usage: 
  datahub dp [DATAPOOL]
List all the datapools or one datapool

Usage of datahub dp create:
  datahub dp create DATAPOOL [file://][ABSOLUTE PATH]
  e.g. datahub dp create dptest file:///home/user/test
       datahub dp create dptest /home/user/test
Create a datapool

Usage of datahub dp rm:
  datahub dp rm DATAPOOL
Remove a datapool
$
```

