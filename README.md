# ServerFramework

## 简介
此项目是我们一款叫 Checkmate 游戏的服务端的框架部分。框架项目内容是我们自开发的， 游戏项目能在我们的 Github 上继续了解。

## 功能介绍
### 提供网关框架
此项目提供了一个进行收发消息的网关-`/Server/Gateway.go`，我们的服务端后续开发基于我们这个框架的网关，进行与客户端的数据通讯。  
网关会接受并处理来自客户端的消息，然后调用服务端操作，与客户端之间建立连接。

### 提供 RPC 调用框架
我们自定义了一套 RPC 调用过程。因为我们服务端和客户端使用了不同的开发语言。所以，客户端需要 RPC 调用服务端函数，我们就自定义了一套 RPC过程。  
代码位于`/Network/`文件夹下，通过`Protobuf`包来将消息编码为二进制。随后，客户端如果通过 RPC 方式调用服务端函数，随后，服务端网关会接收到客户端的 RPC 调用消息，随后网关会调用服务端提供的函数对二进制消息进行解码，然后通过反射机制，调用服务端上对应的函数，实现 RPC 调用。  
这个过程是我们已经在这个框架中实现了，所以，我们之后的服务端代码基于此框架开发，服务端逻辑开发时，就不需要再关心这些底层细节问题

### 提供一个帧同步框架
代码位于`/fsplite/`文件夹下，我们使用帧同步的方式来进行客户端之间数据的同步。

### 提供一个服务端中Server的结构和框架
代码位于`/ServerManager/`内  
```go
// /ServerManager/ServerModule.go
type Server interface {
	GetId() int
	Create()
	GetStatus() int
	Release()
	Start(servern Server)
	Stop()
	Tick()
	GetModuleInfo() *common.ServerModuleInfo
}
```
服务端新建的 Server 实现这些接口即可成为一个可适用 Server .框架提供了一个 Server 的默认行为。我们后续在框架基础上开发服务端的时候，可以通过实现接口中的函数，来实现修改行为，定义我们需要的 Server 的行为

### 框架还提供其他细节功能
框架还提供了一个服务端上运行在不同端口的 Server 之间的一般性通信方法。我们在代码中是通过RPC的方式，实现这种通信。使用RPC的原因是，服务端Server可能之后会运行在不同的机器上，这样，通过我们框架提供的RPC的方式，就不论它们在不在一台机器上，都能实现通信了。(服务端 Server 运行在不同的机器上是我们的构想，由于现实原因约束，我们服务端的 Server 只能在同一台 Server 上测试)

## 客户端和基于此框架的服务端内容，可参阅我们Github代码仓库
* 客户端代码仓库：https://github.com/qaqgame/Checkmate
* 基于此框架的服务端代码仓库：https://github.com/qaqgame/FinalCheckmateServer
