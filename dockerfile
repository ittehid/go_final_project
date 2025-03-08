FROM ubuntu:latest

WORKDIR /app

COPY go_final_project /app/go_final_project
COPY web /app/web

ENV TODO_PORT=7540
ENV TODO_DBFILE=/app/data/scheduler.db
ENV TODO_PASSWORD=12345

EXPOSE 7540

CMD ["./go_final_project"]