FROM grafana/k6:latest

WORKDIR /test

COPY test/load_test/load_test.js .

CMD ["run", "/test/load_test.js"]
