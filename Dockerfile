FROM golang:alpine
MAINTAINER Jermine.hu@qq.com
ENV APP_HOME /go/src/github.com/JermineHu/DocStack/
WORKDIR $APP_HOME
COPY . $APP_HOME
RUN go env && go build -v -o BookStack && chmod +x BookStack
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
COPY --from=0 /go/src/github.com/JermineHu/DocStack/BookStack .
COPY --from=0 /go/src/github.com/JermineHu/DocStack/conf conf
COPY --from=0 /go/src/github.com/JermineHu/DocStack/dictionary dictionary
COPY --from=0 /go/src/github.com/JermineHu/DocStack/static static
COPY --from=0 /go/src/github.com/JermineHu/DocStack/views views
CMD	 ./DocStack install && ./DocStack
