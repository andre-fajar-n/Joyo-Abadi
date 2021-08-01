### FINAL CONTAINER
FROM    alpine:3.8

# Add user 
RUN     adduser -h /app -u 1000 -D app && apk add --no-cache dumb-init

# Add other package
RUN apk update && apk add ca-certificates bash jq curl && rm -rf /var/cache/apk/*

#Set Timezone 
ENV TZ=Etc/UTC
RUN apk add -U tzdata

# set working directory 
WORKDIR /app

# create folder log and config 
RUN     mkdir -p /app/log
RUN 	mkdir -p /app/config

# edit permission 
RUN	chown -R app:app . 

# copy binary 
ADD     shop /app

USER    app

EXPOSE  8080
ENTRYPOINT ["./backend","-c"]
# CMD     ["config/app.toml","api"]