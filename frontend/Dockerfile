# morningstarl2504/balkanid_repo/BalkanID_repo-f1fc3ed153144eb6d79e3c90f73a0f3d312b9c79/frontend/Dockerfile
# Stage 1: Build the React application
FROM node:18-alpine as builder

# Declare a build-time argument from docker-compose
ARG REACT_APP_API_URL

# Set the environment variable for the build process
ENV REACT_APP_API_URL=$REACT_APP_API_URL

# Set the working directory
WORKDIR /app

# Copy package.json and package-lock.json
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application source code
COPY . .

# Build the application (it will use the REACT_APP_API_URL passed from docker-compose)
RUN npm run build

# Stage 2: Serve the application using Nginx
FROM nginx:stable-alpine

# Copy the build output from the builder stage
COPY --from=builder /app/build /usr/share/nginx/html

# Expose port 80 to the outside world
EXPOSE 80

# Start Nginx when the container launches
CMD ["nginx", "-g", "daemon off;"]