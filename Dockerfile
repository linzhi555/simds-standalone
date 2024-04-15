FROM ubuntu:20.04
WORKDIR /
COPY ./target/simlet /simlet
COPY ./config.yaml /config.yaml
CMD /simlet 
