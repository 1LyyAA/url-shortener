FROM jenkins/jenkins:lts
USER root

RUN apt-get update && apt-get install -y golang ansible openssh-client sshpass rsync

CMD ["/usr/local/bin/jenkins.sh"]