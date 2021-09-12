FROM centos
WORKDIR /qiang
EXPOSE 5900
COPY ./char /qiang
CMD ["./char"]