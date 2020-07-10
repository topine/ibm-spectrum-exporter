FROM scratch

COPY ./ibm-spectrum-exporter /
COPY ./metrics_conf.yaml /
ENTRYPOINT [ "/ibm-spectrum-exporter" ]
