FROM ubuntu:20.04
WORKDIR /
RUN apt update
RUN apt install -y wget net-tools iproute2 iputils-ping
COPY ./target/simlet /simlet
COPY ./config.yaml /config.yaml
COPY ./trace/tasks_stream.log ./trace/tasks_stream.log 
CMD /simlet 
