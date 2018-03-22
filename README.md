
# [中文](README-CN.md) | [English](README.md)

## About DocStack 
There will be regular sharing and training within the company, often with meeting notes and related documents, although there are many cloud notes that can be used,
but the company's information is still in the hope of being able to control it, so DocStack is born in the document management needs inside of the company. 

DocStack is used for knowledge sharing and notes,DocStack is developed on the basis of [BookStack] (https://github.com/TruthHun/BookStack), which is born for document management and knowledge sharing.

## How to use it ?

``
git clone https://github.com/JermineHu/DocStack.git

cd DocStack

docker build -t jermine/docstack .

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
