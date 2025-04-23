FROM golang:1.24.2-bookworm

RUN mkdir -p /opt/bos/services/common/
COPY ../device/drivers/terabee/go.mod /opt/bos/services/
COPY ../device/drivers/terabee/go.sum /opt/bos/services/
COPY ../common/ /opt/bos/services/common/

WORKDIR /opt/bos/services/common/
RUN go install .

RUN mkdir -p /opt/bos/services/common/
COPY ../device/drivers/terabee /opt/bos/services/device/drivers/terabee/
WORKDIR /opt/bos/services/device/drivers/terabee/

WORKDIR /opt/bos/services/device/drivers/terabee/cmd/
CMD [ "go","run","server.go" ]


