FROM squidfunk/mkdocs-material:9 AS base

RUN addgroup -S mkdocs && adduser -S mkdocs -G mkdocs

USER mkdocs

ENV HEALTHCHECK_URL=127.0.0.1:8000
HEALTHCHECK --interval=15s --timeout=2s --start-period=20s --retries=2 CMD wget -q -O- -S ${HEALTHCHECK_URL} > /dev/null

RUN pip install mkdocs-open-in-new-tab
