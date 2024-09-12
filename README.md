
# logv 日志查看工具

github.com/ssgo/log 输出的日志格式为json，为了方便查看可以使用 logv

### 安装

```shell
go install github.com/ssgo/tool/logv@latest
```

或直接下载对应操作系统的二进制程序：

#### Linux (amd64):

```shell
curl -o logv https://apigo.cc/tool/logv.linux.amd64 && chmod +x logv
```

#### Linux (arm64):

```shell
curl -o logv https://apigo.cc/tool/logv.linux.amd64 && chmod +x logv
```

#### Mac (Intel):

```shell
curl -o logv https://apigo.cc/tool/logv.darwin.amd64 && chmod +x logv
```

#### Mac (Apple):

```shell
curl -o logv https://apigo.cc/tool/logv.darwin.arm64 && chmod +x logv
```

Windows:

https://apigo.cc/tool/logv.windows.amd64.exe

https://apigo.cc/tool/logv.windows.arm64.exe


### Usage

```shell
Usage:
	logv [-j] [-s] [file]
	-j	josn output
	-s	show full time

Samples:
	tail ***.log | logv
	logv ***.log
	tail ***.log | logv -j -f
```


# sskey 密钥管理工具

支持AES和国密SM4

可以通过生成go语言代码来混淆密钥，增加反编译难度

### 安装

```shell
go install github.com/ssgo/tool/sskey@latest
```

或直接下载对应操作系统的二进制程序：

#### Linux (amd64):

```shell
curl -o sskey https://apigo.cc/tool/sskey.linux.amd64 && chmod +x sskey
```

#### Linux (arm64):

```shell
curl -o sskey https://apigo.cc/tool/sskey.linux.amd64 && chmod +x sskey
```

#### Mac (Intel):

```shell
curl -o sskey https://apigo.cc/tool/sskey.darwin.amd64 && chmod +x sskey
```

#### Mac (Apple):

```shell
curl -o sskey https://apigo.cc/tool/sskey.darwin.arm64 && chmod +x sskey
```

Windows:

https://apigo.cc/tool/sskey.windows.amd64.exe

https://apigo.cc/tool/sskey.windows.arm64.exe

### Usage

```shell
sskey

Usage:
	-l		        List all saved keys
	-c keyName	        Create a new key and save it
	-t keyName	        Test key
	-e [keyName] data	Encrypt data by specified key or default key
	-d [keyName] data	Decrypt data by specified key or default key
	-e4 [keyName] data	Encrypt data by specified key or default key with SM4
	-d4 [keyName] data	Decrypt data by specified key or default key with SM4
	-php keyName	        Output php code
	-java keyName	        Output java code
	-go keyName	        Output go code
	-o keyName	        Output key&iv by default key
	-o [byKeyName] keyName	Output key&iv by specified key)
	-sync keyNames	        Synchronization of keys to another machine from url

Samples:
	sskey -l
	sskey -c aaa
	sskey -t aaa
	sskey -e 123456
	sskey -d xxxxxx
	sskey -e aaa 123456
	sskey -d aaa xxxxxx
	sskey -php aaa
	sskey -java aaa
	sskey -go aaa
	sskey -o aaa
	sskey -o bbb aaa
	sskey -sync aaa,bbb,ccc http://xxxxxx
```


# gowatch 监视代码，自动测试或运行

### 安装

```shell
go install github.com/ssgo/tool/gowatch@latest
```


或直接下载对应操作系统的二进制程序：

#### Linux (amd64):

```shell
curl -o gowatch https://apigo.cc/tool/gowatch.linux.amd64 && chmod +x gowatch
```

#### Linux (arm64):

```shell
curl -o gowatch https://apigo.cc/tool/gowatch.linux.amd64 && chmod +x gowatch
```

#### Mac (Intel):

```shell
curl -o gowatch https://apigo.cc/tool/gowatch.darwin.amd64 && chmod +x gowatch
```

#### Mac (Apple):

```shell
curl -o gowatch https://apigo.cc/tool/gowatch.darwin.arm64 && chmod +x gowatch
```

Windows:

https://apigo.cc/tool/gowatch.windows.amd64.exe

https://apigo.cc/tool/gowatch.windows.arm64.exe

### Usage

```shell
Usage:
	gowatch [-p paths] [-pt types] [-ig ignores] [-t] [-b] [...]
	-p	指定监视的路径，默认为 ./，支持逗号隔开的多个路径，以*结尾代表监听该文件夹下所有类型的文件
	-pt	指定监视的文件类型，默认为 .go,.yml 支持逗号隔开的多个类型
	-ig	排除指定的文件夹，默认从 .gitignore 中找到 / 开头的项目进行排除
	-sh	指定执行的命令，默认为 go
	-r	执行当前目录中的程序，相当于 go run *.go
	-t	执行测试用例，相当于 go test ./tests 或 go test ./tests（自动识别是否存在tests文件夹）
	-b	执行性能测试，相当于 go -bench .*，需要额外指定 -t 或 test 参数
	...	可以使用除 run 外的 go 命令的参数

Samples:
	gowatch -r
	gowatch -t
	gowatch -t -b
	gowatch -p ../ -t
	gowatch test
	gowatch test ./testcase
```
