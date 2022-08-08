FROM alpine

ADD stress /

ENTRYPOINT ["/stress"]
