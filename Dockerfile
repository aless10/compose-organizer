# Use a base image, for example, a lightweight Linux distribution
FROM alpine:latest

# Create a new user called "testrunner"
RUN adduser -D testrunner 
# Switch to the "testrunner" user
USER testrunner

WORKDIR /code
# Copy all the content of the "app" folder into the /code folder
COPY --chown=testrunner main.py . 

# CMD ["ls", "-altri"]
