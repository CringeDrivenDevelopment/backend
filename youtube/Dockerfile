FROM oven/bun:1 AS builder

WORKDIR /app

COPY package.json bun.lock ./
RUN bun install

COPY . .

RUN bun build main.ts --compile --outfile server

FROM debian:bookworm-slim

WORKDIR /app

RUN apt update -y && apt install ffmpeg -y

COPY --from=builder /app/server ./server

EXPOSE 3000
CMD ["./server"]