FROM	golang:latest AS builder
WORKDIR	/app
COPY	close_milestones.go	./
RUN	go run ./close_milestones.go
