FROM alpine:latest

RUN apk add curl 
RUN mkdir /youtube-dl
RUN curl -L https://yt-dl.org/downloads/latest/youtube-dl -o /home/youtube-dl
RUN chmod +x /home/youtube-dl
