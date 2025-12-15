FROM alpine:latest

RUN apk add --no-cache \
    openssh-server \
    openssh-client \
    sudo \
    bash \
    python3 \
    && mkdir /var/run/sshd

RUN adduser -D -s /bin/bash admin \
    && echo "admin:securepassword" | chpasswd \
    && echo "admin ALL=(ALL) ALL" >> /etc/sudoers

RUN mkdir -p /home/admin/.ssh \
    && chown -R admin:admin /home/admin/.ssh \
    && chmod 700 /home/admin/.ssh

# Настройка SSH
RUN ssh-keygen -A \
    && sed -i 's/#PermitRootLogin.*/PermitRootLogin no/' /etc/ssh/sshd_config \
    && sed -i 's/#PasswordAuthentication.*/PasswordAuthentication yes/' /etc/ssh/sshd_config \
    && sed -i 's/#PubkeyAuthentication.*/PubkeyAuthentication yes/' /etc/ssh/sshd_config

EXPOSE 22
CMD ["/usr/sbin/sshd", "-D"]