# POC Dockerfile to add staticly linked ffmpeg to serverless container
FROM ubuntu:20.04 AS ffmpeg
WORKDIR /usr
RUN apt-get update &&  \
    apt-get install -y wget && \
    apt-get install -y xz-utils && \
    wget https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz -O ffmpeg && \
    tar -xf ffmpeg -C . --strip-components=1

FROM openfunctiondev/buildpacks-run-go:v2.5.0-1.17
USER root
COPY --from=ffmpeg /usr/ffmpeg /workspace/ffmpeg
USER cnb