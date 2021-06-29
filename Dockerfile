FROM golang

COPY ./initial/client/initial_client /app/initial/
COPY ./initial/server/initial_server /app/initial/
# CMD ["/app/initial"]