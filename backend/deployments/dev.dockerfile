FROM golang@sha256:84a409b4c174965a51e393064e46f6eb32adb80daa6097851268458136fd68b6

ARG ssh_prv_key
ARG ssh_pub_key

RUN apk add git openssh-client ca-certificates

# RUN git config --global --add url."git@github.com:".insteadOf "https://github.com/"

# RUN mkdir -p /root/.ssh && \
#     chmod 0700 /root/.ssh && \
#     ssh-keyscan github.com > /root/.ssh/known_hosts

# RUN echo "$ssh_prv_key" > /root/.ssh/id_rsa && \
#     echo "$ssh_pub_key" > /root/.ssh/id_rsa.pub && \
#     chmod 600 /root/.ssh/id_rsa && \
#     chmod 600 /root/.ssh/id_rsa.pub

# RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config

WORKDIR /app

COPY ./ /app

# RUN GOPRIVATE=github.com/Lionparcel/*- go mod download

RUN go get github.com/githubnemo/CompileDaemon

# Remove SSH keys
# RUN rm -rf /root/.ssh/

ENTRYPOINT CompileDaemon --build="go build main.go" --command="./main"