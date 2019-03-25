############################
# STEP 1 build executable binary
############################
FROM golang:alpine AS builder
WORKDIR /mypackage/myapp/
COPY . .
# This uses `go modules` which means that a build will automatically get the dependencies
RUN go build -tags netgo -o /go/bin/xray-teams-connector
############################
# STEP 2 build a small image
############################
FROM scratch
# Copy our static executable.
COPY --from=builder /go/bin/xray-teams-connector /go/bin/xray-teams-connector
# Run the hello binary.
ENTRYPOINT ["/go/bin/xray-teams-connector"]
