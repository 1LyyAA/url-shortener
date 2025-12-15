FROM alpine:latest

RUN apk add --no-cache \
    openssh-server \
    sudo \
    python3 \
    py3-pip \
    sshpass \
    ansible \
    bash \
    && ssh-keygen -A \
    && adduser -D -s /bin/bash admin \
    && echo "admin:securepassword" | chpasswd \
    && echo "admin ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers \
    && mkdir -p /home/admin/.ssh \
    && chown -R admin:admin /home/admin/.ssh \
    && chmod 700 /home/admin/.ssh \
    && sed -i 's/#PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config \
    && sed -i 's/#PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D", "-e"]