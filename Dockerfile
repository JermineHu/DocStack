FROM golang:alpine
MAINTAINER Jermine.hu@qq.com
#RUN apk add --no-cache git go ;\
#    go get github.com/TruthHun/BookStack ;\
#    go get github.com/TruthHun/converter ;\
#    go get github.com/TruthHun/gotil ;\
#    go get github.com/TruthHun/html2article ;\
#    go get github.com/TruthHun/html2md ;\
#    go get github.com/alexcesaro/mail ;\
#    go get github.com/aliyun/aliyun-oss-go-sdk ;\
#    go get github.com/astaxie/beego ;\
#    go get github.com/huichen/sego ;\
#    go get github.com/kataras/iris 
#ENV APP_HOME /go/src/github.com/TruthHun/BookStack
#WORKDIR $APP_HOME
#RUN go env && go build -v -o BookStack && chmod +x BookStack
FROM ubuntu
MAINTAINER Jermine.hu@qq.com
ENV CALIBRE_VERSION 3.19.0
RUN apt update -y && apt install -y --no-install-recommends \
    libgl1-mesa-dev ttf-wqy-zenhei fonts-wqy-microhei wget chromium-browser xdg-utils xz-utils ;\
    mkdir -p /opt/calibre && rm -rf /opt/calibre/* && cd /opt/calibre ;\
    wget --no-check-certificate -O  /tmp/calibre-${CALIBRE_VERSION}-x86_64.txz  https://download.calibre-ebook.com/${CALIBRE_VERSION}/calibre-${CALIBRE_VERSION}-x86_64.txz ;\
    tar xvf /tmp/calibre-${CALIBRE_VERSION}-x86_64.txz -C /opt/calibre ;\
    mkdir /usr/share/desktop-directories/ ;\
    /opt/calibre/calibre_postinstall
WORKDIR /app
# Get a file from first floor image
#COPY --from=0 /go/src/github.com/TruthHun/BookStack/BookStack .
#COPY --from=0 /go/src/github.com/TruthHun/BookStack/conf conf
CMD	 ["./BookStack","install"]
