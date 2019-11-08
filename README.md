## 整体思维导图

![image-20191107141258379](img/README/image-20191107141258379.png)



## 问题及知识点 

#### 秒传原理

1、计算文件Hash(MD5,SHA1等)，并保存到数据库唯一文件表（tbl_file）中

2、用户上传前计算本地文件hash然后与去文件表中查询是否存在

3、如果存在，说明已上传同样的文件，则直接将唯一文件表中的对应文件信息存储一份到用户文件表（tbl_user_file）中。秒传完毕。



#### 分块上传/断点续传

分块上传：文件切成多块，独立 传输，上传完成后合并

断点续传：基于分块上传，传输暂停或异常中断后，可基于原来进度重传

##### 优点说明：

​	小文件不建议分块上传，适得其反

​	分块上传可并行执行。可无序传输

​	失败重试后可基于已上传成功的进行后续传输，减少流量和时间

##### 流程：

![image-20191107154706416](img/README/image-20191107154706416.png)

#### 

#### 云存储 OSS

​	Object Storage Service 简称OSS，对象存储服务。

##### 为什么要引入云存储

​	原有的存储都是将文件存储在本地。如果本地磁盘坏了 ，这些数据就丢失了。另一个随着用户以及数据量的增加，本地磁盘是完全不够用的，就涉及到扩容，分布式存储等等一系列问题。课程中有一章节专门介绍私有云 Ceph 的搭建和使用。其实大部分中小型公司为了便利以及维护成本多数会选择共有云。

##### 阿里云OSS优点：

- 可靠性：服务可用性，数据持久性
- 资源隔离存储，多用户访问鉴权管理
- 提供标准restful风格API，多语言SDK方便接入	

##### 加入阿里云OSS后的上传流程：

![image-20191107163752960](img/README/image-20191107163752960.png)



#### RabbitMQ

​	RabbitMQ 是一个由 erlang 开发的 AMQP (Advanced Message Queuing Protocol) 的开源实现。

​	AMQP ：高级消息队列协议，是应用层协议的一个开放标准，为面向消息的中间件设计。消息中间件主要用于组件之间的解耦，消息的发送者无需知道消息使用者的存在，反之亦然。 AMQP 的主要特征是面向消息、队列、路由（包括点对点和发布 / 订阅）、可靠性、安全。 RabbitMQ 是一个开源的 AMQP 实现，服务器端用 Erlang 语言编写，支持多种客户端，如：Python、Ruby、.NET、Java、JMS、C、PHP、ActionScript、XMPP、STOMP 等，支持 AJAX。用于在分布式系统中存储转发消息，在易用性、扩展性、高可用性等方面表现不俗。

##### 更多资料

> ###### [RabbitMQ 的应用场景以及基本原理介绍](<https://learnku.com/articles/27446>)
>
> ###### [RabbitMQ系列（二）深入了解RabbitMQ工作原理及简单使用](https://www.cnblogs.com/vipstone/p/9275256.html)
>
> ###### [RabbitMQ系列（三）RabbitMQ交换器Exchange介绍与实践](https://www.cnblogs.com/vipstone/p/9295625.html)
>
> ###### [RabbitMQ系列（四）RabbitMQ事务和Confirm发送方消息确认——深入解读](https://www.cnblogs.com/vipstone/p/9350075.html)
>
> ###### [RabbitMQ系列（五）使用Docker部署RabbitMQ集群](https://www.cnblogs.com/vipstone/p/9362388.html)
>
> ###### [RabbitMQ系列（六）你不知道的RabbitMQ集群架构全解](https://www.cnblogs.com/vipstone/p/9368106.html)

​	

#####  引入目的：

​	这里引入主要是讲是上一步上传服务中的云存储，转为异步处理，优化上传时间、应用解耦。还有一些秒杀类的功能可以提供流量削峰的作用。

![image-20191107172007042](img/README/image-20191107172007042.png)



##### 上传文件架构变迁

![image-20191107172331562](img/README/image-20191107172331562.png)



#### 微服务化

​	课程中讲到的微服务是一中架构风格，是相对于单体应用演变而来的。将一个庞大复杂的应用分散治理的理念。根据实际情况按照业务能力或功能的粒度，拆分成一个一个`小的服务` ，他们是`独立的进程`，且`独立部署`，`无集中式管理`。服务间通过`轻量级的通信`进行业务交互。

​	关于微服务的概念和架构演进，优缺点以及落地架构衍生出来的问题，有很多内容课程由于时间缘故讲解的还是比较简单，可以看下这篇文章的详细讲解 [**一文详解微服务架构**](https://www.cnblogs.com/skabyy/p/11396571.html)



##### gin框架进行服务改造

​	首先使用现在很火go web 框架 gin 对接口进行了改造。为此专门单独去学习了一篇 gin 的使用[**gin-test-project**](<https://github.com/GibsonCool/gin-test-project>)

##### RPC

​	RPC 代指远程过程调用（Remote Procedure Call），它的调用包含了传输协议和编码（对象序列号）协议等等。允许运行于一台计算机的程序调用另一台计算机的子程序，而开发人员无需额外地为这个交互作用编程

![image-20191108140630257](img/README/image-20191108140630257.png)

​	

​	乍一看好像和我们平常的 Restful 的过程好像差不多，而且搜索了挺多的网上的文章还是迷迷糊糊的，后来在回头再看一遍视频作者的解读，大概抓住了要点。与其从他们的概念上去区分，不如从他们的用途，有了 Restful 为什么还需要 rpc （简单、通用、安全、效率）的方向去理解，就更能区别以及对两者有清晰分开的认知。

![image-20191108144122745](img/README/image-20191108144122745.png)



##### gRPC

​	是Google开源的一个 rpc 框架，基于 HTTP/2 协议设计，采用了 Protobuf 作为IDL（Interface description language [**接口描述语言**](<https://zh.wikipedia.org/wiki/%E6%8E%A5%E5%8F%A3%E6%8F%8F%E8%BF%B0%E8%AF%AD%E8%A8%80>)）,

​	详细介绍，可以看看这篇 [4.1 gRPC及相关介绍](<https://eddycjy.gitbook.io/golang/di-4-ke-grpc/install>) 以及后续的 gRPC相关内容，详细且友好

![img](img/README/7dcac5be0a34636c699025368242d3f3.png)

##### go-micro

[Micro中文文档](https://micro.mu/docs/cn/index.html)

##### consul

中文文档[Consul](https://book-consul-guide.vnzmi.com/)

##### 