# Pull base image.
FROM dockerfile/ubuntu

# Install Redis.
RUN apt-get install -y software-properties-common
RUN add-apt-repository -y ppa:chris-lea/redis-server
RUN apt-get update
RUN apt-get install -y redis-server

# Expose ports.
EXPOSE 6379

# RUN echo "daemonize yes" > redis.conf
# Define an entry point.
CMD redis-server 
# ENTRYPOINT ["redis-server"]
