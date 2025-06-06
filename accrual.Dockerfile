FROM ubuntu

WORKDIR /app

COPY "/cmd/accrual/accrual_linux_amd64" /app/accrual_linux_amd64

ENV DSN=""

CMD ["/app/accrual_linux_amd64", "-d", "$DSN"]