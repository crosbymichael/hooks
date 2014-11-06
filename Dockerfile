FROM busybox

COPY hooks /usr/bin/
ENTRYPOINT ["hooks"]
