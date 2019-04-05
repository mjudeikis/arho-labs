FROM registry.access.redhat.com/rhel7:latest
ADD template template
COPY frontend .
ENTRYPOINT [ "/frontend" ]

