
# [中文](README-CN.md) | [English](README.md)

## About DocStack 
There will be regular sharing and training within the company, often with meeting notes and related documents, although there are many cloud notes that can be used,
but the company's information is still in the hope of being able to control it, so DocStack is born in the document management needs inside of the company. 

DocStack is used for knowledge sharing and notes,DocStack is developed on the basis of [BookStack] (https://github.com/TruthHun/BookStack), which is born for document management and knowledge sharing.

### The docker repo：

https://hub.docker.com/r/jermine/docstack/

### The Github repo is：
https://github.com/JermineHu/DocStack


## RoadMap

## V0.2 feature
- [ ] 增加docker-compose解决编排和依赖的问题;
- [ ] 增加注册开关，使其注册功能可配置;
- [ ] 将配置改为环境变量配置，取消配置文件的参与，主要考虑集群部署的方便性;
- [ ] 支持主流平台的分享功能，让知识传播的更远；
- [ ] 支持打赏功能，提高大家分享的积极性；
- [ ] 支持onlyoffice的文档协作功能，实现在线word、excel、ppt的编辑和预览功能;

## V0.1 在BookStack的基础上完善的功能
- [x] 增加dockerfile解决了环境依赖的问题;
- [x] 实现了DocStack的docker自动构建功能，只要提交代码编译通过即可生成最新镜像;
- [x] 增加了govendor的支持，解决当前go项目中包依赖问题;
- [x] 增加了中英文的README文档，让DocStack支持国际化，让全球的开发者都能用DocStack;
- [x] 将DocStack的版权改为了宽松的MIT协议;

## How to use it ?

```
# get the code from repo
git clone https://github.com/JermineHu/DocStack.git

cd DocStack

# build dokcer images
docker build -t jermine/docstack .

# run Docstack by docker
docker run -d --restart=always -p 8081:8181 -v ~/DocStack/conf:/app/conf  -v ~/DocStack/dictionary:/app/dictionary -v ~/DocStack/logs:/app/logs -v ~/DocStack/store:/app/store -v ~/DocStack/uploads:/app/uploads jermine/docstack
```

## Donation

### If it helps you, buy me a cup of coffee :)

#####  AliPay

![donation-alipay](https://raw.githubusercontent.com/JermineHu/docker-frp/master/img/alipay.png)

#####  Wechat Pay

![donation-wechatpay](https://raw.githubusercontent.com/JermineHu/docker-frp/master/img/wechat.png)

#####  Paypal

Donate money by [paypal](https://paypal.me/jerminehu) to my account **jermine.hu@qq.com**.
