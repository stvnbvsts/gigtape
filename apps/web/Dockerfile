# syntax=docker/dockerfile:1.7

FROM node:20-bookworm-slim AS build

WORKDIR /app

COPY apps/web/package*.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci

COPY apps/web ./
RUN npm run build

FROM nginx:1.27-alpine

COPY apps/web/nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist /usr/share/nginx/html
RUN sed -i '/^user /d; s#pid .*nginx.pid;#pid /tmp/nginx.pid;#' /etc/nginx/nginx.conf \
    && chown -R nginx:nginx /etc/nginx/conf.d /usr/share/nginx/html /var/cache/nginx /var/run /var/log/nginx

USER nginx

EXPOSE 8080
