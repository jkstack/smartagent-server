# 编排脚本

提供执行yaml编排脚本的能力，对标ansible

## 已支持插件

1. [exec](plugin/exec.md): 执行脚本或命令
2. [file](plugin/file.md): 文件传输

[api接口](../api/yaml/run.md)

## 示例

    name: deploy nginx
    timeout: 600
    tasks:
      - name: flow control
        plugin: exec
        cmd: echo "do even things"
        if: $EVEN = 1
        timeout: 1
      - name: flow control
        plugin: exec
        cmd: echo "do odd things"
        if: $ODD = 1
      - name: check if nginx installed
        plugin: exec
        cmd: apt list --installed | grep nginx | wc -l
        timeout: 300
        output: cnt
        auth: sudo
      - name: uninstall nginx
        plugin: exec
        cmd: apt purge -y nginx
        if: $cnt > 0
        auth: sudo
      - name: install nginx
        plugin: exec
        cmd: apt install -y nginx
        auth: sudo
      - name: add server
        plugin: file
        action: push
        src: example.conf
        dst: /etc/nginx/conf.d/example.conf
        auth: sudo
      - name: mkdir /var/www/example
        plugin: exec
        cmd: mkdir -p /var/www/example
        auth: sudo
      - name: add index.html
        plugin: file
        action: push
        src: index.html
        dst: /var/www/example/index.html
        auth: sudo
      - name: reload service
        plugin: exec
        cmd: systemctl reload nginx
        auth: sudo

## 通用字段

`name`: 任务名称，全局作用域表示总任务名称，tasks中表示子任务名称

`plugin`: 子任务所需的插件名称，详见已支持插件章节

`auth`: 提权方式，仅linux有效

`timeout`: 超时时间，最外层为所有任务总超时时间，内层为单个任务的超时时间

`if`: 条件匹配，仅当给定表达式结果为`true`时执行当前子任务，目前仅支持单条表达式，
      支持的匹配运算符如下：
  - `>`: 左值大于右值为true
  - `<`: 左值小于右值为true
  - `=`: 左值等于右值为true
  - `!=`: 左值不等于右值为true
  - `<=`: 左值小于等于右值为true
  - `>=`: 左值大于等于右值为true
  - 以上比较过程若左值和右值都能转成整数时按照整数进行比较，否则按照字符串进行比较

## 变量

编排文件中的变量以$开头，在脚本运行时会内置以下系统变量，所有系统内置变量名称均为大写：

1. `$ID`: 当前主机ID
2. `$IDX`: 当前主机在列表中的下标
3. `$EVEN`/`$ODD`: 奇偶标志位0或1
4. `$DEADLINE`: 超时时间戳