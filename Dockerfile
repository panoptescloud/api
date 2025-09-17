FROM golang:1.25 AS dev

RUN go install github.com/air-verse/air@latest

COPY ./.local/docker/watch.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENV PATH="$PATH:/app/bin"

# This is where you should mount your code
VOLUME /app

# Think this will work on multiple arches
# Don't think i need to setup a multi-arch build as this is dev only
RUN (cd /tmp && curl -Lo ./mockery.tar.gz https://github.com/vektra/mockery/releases/download/v2.46.3/mockery_2.46.3_Linux_$(uname -m).tar.gz \
    && tar -xzvf mockery.tar.gz \
    && chmod +x ./mockery \
    && mv mockery /usr/local/bin \
    && rm -rf /tmp/* \
    )

ENTRYPOINT ["/entrypoint.sh"]
