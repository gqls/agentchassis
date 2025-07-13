FROM ubuntu:latest
LABEL authors="ant"

ENTRYPOINT ["top", "-b"]