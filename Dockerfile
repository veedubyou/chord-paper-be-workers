FROM python:3.8-buster as builder

RUN apt-get update
RUN apt install -y ffmpeg

#RUN pip install --no-cache-dir tensorflow==2.3.0
RUN pip install --no-cache-dir spleeter

RUN wget https://golang.org/dl/go1.16.4.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.16.4.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /chord-paper-be-workers

COPY go.mod go.sum ./
COPY pkg/ ./pkg/
COPY src/ ./src/

WORKDIR /chord-paper-be-workers/src

RUN go build -o chord-paper-be-workers
 
#CMD exec /bin/sh -c "trap : TERM INT; sleep 9999999999d & wait"
CMD ["./chord-paper-be-workers"]
