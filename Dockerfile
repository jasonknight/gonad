FROM debian:9.4-slim
MAINTAINER Jason Martin <jason.martin83@protonmail.com>
COPY ./gonad /gonad
EXPOSE 601/tcp

ENTRYPOINT ["/gonad"]

