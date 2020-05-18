FROM scratch

COPY ./bin/ibm-spectrum-exporter /
COPY ./metrics_conf.yaml /
ENTRYPOINT [ "/ibm-spectrum-exporter" ]
