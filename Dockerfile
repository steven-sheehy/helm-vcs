FROM scratch
ADD bin/linux/helmvcs /helmvcs
CMD ["/helmvcs"]
