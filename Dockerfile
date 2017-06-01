FROM alpine

RUN mkdir /app
WORKDIR /app

COPY bin/user-graphql .

CMD ./user-graphql
