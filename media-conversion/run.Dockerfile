# POC Dockerfile to add staticly linked ffmpeg to serverless container
FROM ubuntu:20.04 AS prepare
WORKDIR /usr
RUN apt-get update &&  \
    apt-get install -y wget xz-utils ca-certificates libc6 && \
    wget https://johnvansickle.com/ffmpeg/builds/ffmpeg-git-amd64-static.tar.xz -O ffmpeg && \
    tar -xf ffmpeg -C . --strip-components=1 && \
    wget https://storage.googleapis.com/downloads.webmproject.org/releases/webp/libwebp-1.2.0-linux-x86-64.tar.gz -O libwebp && \
    tar -xf libwebp -C .

FROM openfunctiondev/buildpacks-run-go:v2.5.0-1.17
USER root
COPY --from=prepare --chown=cnb:cnb /usr/ffmpeg /workspace/ffmpeg
COPY --from=prepare --chown=cnb:cnb /usr/libwebp-1.2.0-linux-x86-64/bin/cwebp /workspace/cwebp
COPY --from=prepare --chown=cnb:cnb /usr/libwebp-1.2.0-linux-x86-64/bin/gif2webp /workspace/gif2webp
# Libs needed for cwebp
COPY --from=prepare /lib/x86_64-linux-gnu/libpthread.so.0 /lib/libpthread.so.0
COPY --from=prepare /lib/x86_64-linux-gnu/libm.so.6 /lib/libm.so.6
COPY --from=prepare /lib/x86_64-linux-gnu/libc.so.6 /lib/libc.so.6
COPY --from=prepare /lib/x86_64-linux-gnu/ld-linux-x86-64.so.2 /lib64/ld-linux-x86-64.so.2

ENV LD_LIBRARY_PATH=/lib
USER cnb