# Build the manager binary
FROM golang:1.17 as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go
COPY api/ api/
COPY controllers/ controllers/
COPY core/ core/
COPY pkg/ pkg/

# Build

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o manager main.go

FROM registry.access.redhat.com/ubi8/ubi-minimal@sha256:21504085e8d740e62b52573fe9a1a0d58a3e7dba589cac69734ad2fa81d66635

RUN microdnf install yum \
    && yum -y update-minimal --security --sec-severity=Important --sec-severity=Critical \
    && yum clean all \
    && rpm -e yum \
    && microdnf clean all
    
ENV USER_UID=1001 \
    USER_NAME=dell-csi-operator \
    X_CSI_OPERATOR_CONFIG_DIR="/etc/config/dell-csi-operator"
WORKDIR /
COPY --from=builder /workspace/manager .
COPY driverconfig/ /etc/config/local/dell-csi-operator
LABEL vendor="Dell Inc." \
      name="dell-csi-operator" \
      summary="Operator for installing Dell EMC CSI drivers" \
      description="Common Operator for installing various Dell EMC CSI drivers" \
      version="1.7.0" \
      license="Dell CSI Operator EULA"
# copy the licenses folder
COPY licenses /licenses

ENTRYPOINT ["/manager"]
USER ${USER_UID}
