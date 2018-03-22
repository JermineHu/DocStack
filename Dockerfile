FROM golang:alpine
MAINTAINER Jermine.hu@qq.com
ENV APP_HOME /go/src/github.com/JermineHu/DocStack/
WORKDIR $APP_HOME
COPY . $APP_HOME
RUN go env && go build -v -o BookStack && chmod +x BookStack
FROM jermine/docstack:calibre
MAINTAINER Jermine.hu@qq.com
WORKDIR /app
# Get a file from first floor image
COPY --from=0 /go/src/github.com/JermineHu/DocStack/BookStack .
COPY --from=0 /go/src/github.com/JermineHu/DocStack/conf conf
COPY --from=0 /go/src/github.com/JermineHu/DocStack/dictionary dictionary
COPY --from=0 /go/src/github.com/JermineHu/DocStack/static static
COPY --from=0 /go/src/github.com/JermineHu/DocStack/views views
CMD	 ./DocStack install && ./DocStack
