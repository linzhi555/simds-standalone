FROM  ubuntu:22.04
ARG Config="please_tell_the_config_path"
WORKDIR /
RUN apt update
RUN apt install -y wget net-tools iproute2 iputils-ping
COPY ${Config} /config.yaml
COPY ./google_trace/tasks_stream.log ./google_trace/tasks_stream.log 
COPY ./target/simlet /simlet
CMD /simlet 
