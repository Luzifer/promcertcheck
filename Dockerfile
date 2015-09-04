FROM scratch

ADD https://rootcastore.hub.luzifer.io/v1/store/latest /etc/ssl/ca-bundle.pem
ADD promcertcheck /promcertcheck

EXPOSE 3000
ENTRYPOINT ["/promcertcheck"]
CMD ["--probe=https://www.google.com/", "--probe=https://www.facebook.com/"]
