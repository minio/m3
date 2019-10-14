FROM nginx:latest

WORKDIR /etc/nginx

CMD ["nginx"]
EXPOSE 80 443

COPY nginx.conf /etc/nginx/nginx.conf
COPY html /usr/share/nginx/html