
# 安装和启动IPFS on macOS

InterPlanetary File System (IPFS) 是一个分布式文件系统，旨在连接所有计算设备，以便形成一个统一的世界文件系统。以下是在macOS上安装和启动IPFS的步骤。

## 前提条件

- macOS操作系统
- Homebrew (macOS的包管理器)

如果您尚未安装Homebrew，请打开终端并运行以下命令：

```sh
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

## 安装IPFS

1. 打开您的macOS终端。

2. 使用Homebrew安装IPFS。在终端中输入以下命令：

```sh
brew install ipfs
```

3. 等待安装完成。Homebrew会自动下载并安装IPFS及其依赖项。

## 初始化IPFS

在安装IPFS之后，您需要初始化IPFS配置文件。这将创建一个新的IPFS节点。

1. 在终端中运行以下命令来初始化IPFS节点：

```sh
ipfs init
```

2. 初始化过程将生成一个新的节点ID，并创建默认的配置文件和存储目录。

## 启动IPFS守护进程

IPFS守护进程是一个长期运行的进程
```sh
ipfs daemon
```


# 运行

## 先决条件
在mac 上启动需要安装 go1.18 和 docker 
如果是m系列芯片需要使用orbstack代替docker https://orbstack.dev/

## 下载
项目需要安装在GOPATH目录下才能生效
```sh
cd $GOPATH/src
git clone https://github.com/cangcang-zcr/gnark_and_fabric.git 
```

## 启动
依赖都在vendor，如果有需要可以利用go mod vendor 重新安装
启动直接使用start.sh
```
chmod +x ./start.sh
./start.sh
```

# 项目说明

## 功能
这个代码是一个区块链项目的示例代码，主要实现了以下功能：

1. 通过 IPFS 存储字符串信息，并获取它的 CID。
2. 生成私钥和公钥，并将公钥进行 base64 编码。
3. 使用 SDK 创建通道并加入，并且创建链码生命周期。
4. 调用链码中的外部服务，将公钥信息存储到区块链上。
5. 查询区块链中存储的公钥信息。
6. 生成零知识参数，并将其存储到区块链上。
7. 获取区块链中存储的零知识参数。
8. 执行零知识证明，生成 proof。
9. 将生成的 proof 和公共 witness 的数据转换为二进制流，并存储到区块链上。
10. 验证区块链中存储的零知识证明。

## web
